package cloudstore

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"mime/multipart"
	"strings"
	"time"
)

type IO struct {
	fs    *OSFileSystem
	paths *PathMapper
}

func NewIO(fs *OSFileSystem, paths *PathMapper) *IO {
	return &IO{fs: fs, paths: paths}
}

// SetupFSRoot will validate that a directory exists with the value of path.
// If it does not exist, it will be created.
//
// This method should be called once before executing any other methods.
func (io *IO) SetupFSRoot(path string, perm fs.FileMode) error {
	_, err := io.fs.Stat(path)
	if err != nil {
		if io.fs.IsNotExist(err) {
			if err := io.fs.Mkdir(path, perm); err != nil {
				return fmt.Errorf("creating directory [%s]: %w", path, err)
			}

			return nil
		}

		return fmt.Errorf("getting file info [%s]: %w", path, err)
	}

	return nil
}

type DirIO struct {
	ID        string
	UserID    string
	Name      string
	ParentID  string
	CreatedAt time.Time
	UpdatedAt time.Time
	LastWrite time.Time
	FSPath    string
	UserPath  string
}

// NewDirIO is the parameters when creating a new directory.
type NewDirIO struct {
	// The directory ID. ID will also be the file name on the file system.
	ID string

	// The ID of the user the directory belongs to.
	UserID string

	// The name of the directory. This is the name the user will use when
	// interacting with directories via paths.
	Name string

	// The ID of the parent directory. If creating a root directory for a
	// user, ParentID will be empty. All other directories will need a
	// ParentID.
	ParentID sql.NullString

	// The time the directory was created. This should be a UTC time.
	CreatedAt time.Time

	// The path to the central storage directory on the file system.
	FSDir string

	// The directory permissions on the file system.
	FSPerm fs.FileMode
}

// NewDir writes a directory to the file system and persists its information
// in the database. If the directory is a sub directory, all the ancesteral
// paths are persisted. The directory is returned as a DirIO.
func (io *IO) NewDir(ctx context.Context, q *Query, d NewDirIO) (DirIO, error) {
	err := q.InsertDirectory(ctx, InsertDirectoryConfig{
		ID:        d.ID,
		UserID:    d.UserID,
		Name:      d.Name,
		ParentID:  d.ParentID,
		CreatedAt: d.CreatedAt,
	})
	if err != nil {
		return DirIO{}, err
	}

	err = q.InsertSelfPath(ctx, d.ID)
	if err != nil {
		return DirIO{}, err
	}

	if d.ParentID.Valid {
		err = q.InsertParentPaths(ctx, InsertParentPathsConfig{
			ParentID: d.ParentID.String,
			ChildID:  d.ID,
		})
		if err != nil {
			return DirIO{}, err
		}
	}

	fsPath, err := io.paths.GetDirFS(ctx, q, d.ID)
	if err != nil {
		return DirIO{}, err
	}

	userPath, err := io.paths.GetDir(ctx, q, d.ID)
	if err != nil {
		return DirIO{}, err
	}

	if err := io.fs.Mkdir(fsPath, d.FSPerm); err != nil {
		return DirIO{}, fmt.Errorf("creating directory [%s]: %w", fsPath, err)
	}

	return DirIO{
		ID:        d.ID,
		UserID:    d.UserID,
		Name:      d.Name,
		ParentID:  d.ParentID.String,
		CreatedAt: d.CreatedAt.UTC(),
		FSPath:    fsPath,
		UserPath:  userPath,
	}, nil
}

// RemoveFSDir accepts the path to a directory and removes it from the file system.
// All sub directories and files will be removed.
func (io *IO) RemoveFSDir(fsPath string) error {
	return io.fs.RemoveAll(fsPath)
}

// RemoveFS accpets the path to a file or (empty) directory and removes it from the
// file system.
func (io *IO) RemoveFS(fsPath string) error {
	return io.fs.Remove(fsPath)
}

type FileIO struct {
	ID          string
	UserID      string
	DirectoryID string
	Name        string
	UploadedAt  time.Time
	FSPath      string
	UserPath    string
	Size        int64
}

// fileInfo returns this FileIO as a FileInfo.
func (f *FileIO) fileInfo() FileInfo {
	return FileInfo{
		ID:          f.ID,
		DirectoryID: f.DirectoryID,
		Name:        f.Name,
		Path:        f.UserPath,
		FSPath:      f.FSPath,
		Size:        f.Size,
		UploadedAt:  f.UploadedAt.UTC(),
	}
}

type NewFileIO struct {
	ID          string
	UserID      string
	DirectoryID string
	UploadedAt  time.Time
	Header      *multipart.FileHeader
	FSDir       string
	FSPerm      fs.FileMode
}

// NewFile writes a file under a specified directory on the file system and
// persists its information to the database. The file is returned as a FileIO.
func (io *IO) NewFile(ctx context.Context, q *Query, f NewFileIO) (FileIO, error) {
	file, err := f.Header.Open()
	if err != nil {
		return FileIO{}, err
	}

	err = q.InsertFile(ctx, InsertFileConfig{
		ID:          f.ID,
		UserID:      f.UserID,
		DirectoryID: f.DirectoryID,
		Name:        f.Header.Filename,
		UploadedAt:  time.Now().UTC(),
	})
	if err != nil {
		return FileIO{}, err
	}

	userPath, err := io.paths.GetFile(ctx, q, f.DirectoryID, f.Header.Filename)
	if err != nil {
		return FileIO{}, err
	}

	fsPath, err := io.paths.GetFileFS(ctx, q, f.DirectoryID, f.ID)
	if err != nil {
		return FileIO{}, err
	}

	// Create the file and set the file permissions on the file system.
	dst, err := io.fs.Create(fsPath, f.FSPerm)
	if err != nil {
		return FileIO{}, err
	}

	// Write the file content to the file on the file system.
	_, err = io.fs.Copy(dst, file)
	if err != nil {
		return FileIO{}, err
	}
	dst.Close()

	return FileIO{
		ID:          f.ID,
		UserID:      f.UserID,
		DirectoryID: f.DirectoryID,
		Name:        f.Header.Filename,
		UploadedAt:  f.UploadedAt.UTC(),
		FSPath:      fsPath,
		UserPath:    userPath,
		Size:        f.Header.Size,
	}, nil
}

type ReadFileInfoIO struct {
	UserID string
	FileID string
	FSPath string
}

// FileInfo gets the information for a users file. The information is gathered
// from both the database and file system, and returns it as a FileIO.
//
// The actual file content is not returned by this function.
func (io *IO) ReadFileInfo(ctx context.Context, q *Query, f ReadFileInfoIO) (FileIO, error) {
	row, err := q.SelectFileByIDUser(ctx, f.FileID, f.UserID)
	if err != nil {
		return FileIO{}, err
	}

	userPath, err := io.paths.GetFile(ctx, q, row.DirectoryID, row.Name)
	if err != nil {
		return FileIO{}, err
	}

	// Construct the full file system path of the file.
	dirIDPath, err := q.SelectDirectoryFSPath(ctx, row.DirectoryID)
	if err != nil {
		return FileIO{}, err
	}
	fsPath := fmt.Sprintf("%s/%s/%s", f.FSPath, strings.Join(dirIDPath, "/"), row.ID)

	// Get the file size on the file system.
	stat, err := io.fs.Stat(fsPath)
	if err != nil {
		return FileIO{}, err
	}

	return FileIO{
		ID:          row.ID,
		UserID:      row.UserID,
		DirectoryID: row.DirectoryID,
		Name:        row.Name,
		UploadedAt:  row.UploadedAt.UTC(),
		FSPath:      fsPath,
		UserPath:    userPath,
		Size:        stat.Size(),
	}, nil
}
