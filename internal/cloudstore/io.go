package cloudstore

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"mime/multipart"
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

// NewDirIO is the parameters when creating a new directory.
type NewDirIO struct {
	ID        string
	UserID    string
	Name      string
	ParentID  sql.NullString
	CreatedAt time.Time
	FSDir     string
	FSPerm    fs.FileMode
}

// NewDir writes a directory to the file system and persists its information
// in the database. If the directory is a sub directory, all the ancesteral
// paths are persisted. The directory is returned as a Dir.
func (io *IO) NewDir(ctx context.Context, q *Query, d NewDirIO) (Dir, error) {
	err := q.InsertDirectory(ctx, InsertDirectoryConfig{
		ID:        d.ID,
		UserID:    d.UserID,
		Name:      d.Name,
		ParentID:  d.ParentID,
		CreatedAt: d.CreatedAt,
	})
	if err != nil {
		return Dir{}, err
	}

	err = q.InsertSelfPath(ctx, d.ID)
	if err != nil {
		return Dir{}, err
	}

	if d.ParentID.Valid {
		err = q.InsertParentPaths(ctx, InsertParentPathsConfig{
			ParentID: d.ParentID.String,
			ChildID:  d.ID,
		})
		if err != nil {
			return Dir{}, err
		}
	}

	fsPath, err := io.paths.GetDirFS(ctx, q, d.ID)
	if err != nil {
		return Dir{}, err
	}

	userPath, err := io.paths.GetDir(ctx, q, d.ID)
	if err != nil {
		return Dir{}, err
	}

	if err := io.fs.Mkdir(fsPath, d.FSPerm); err != nil {
		return Dir{}, fmt.Errorf("creating directory [%s]: %w", fsPath, err)
	}

	return Dir{
		ID:        d.ID,
		Owner:     d.UserID,
		ParentID:  d.ParentID.String,
		Name:      d.Name,
		Path:      userPath,
		CreatedAt: d.CreatedAt,
		UpdatedAt: time.Time{},
		LastWrite: time.Time{},
		fsPath:    fsPath,
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
		OwnerID:     f.UserID,
		DirectoryID: f.DirectoryID,
		Name:        f.Name,
		Path:        f.UserPath,
		Size:        f.Size,
		UploadedAt:  f.UploadedAt.UTC(),
		FSPath:      f.FSPath,
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
// persists its information to the database. The file is returned as a FileInfo.
func (io *IO) NewFile(ctx context.Context, q *Query, f NewFileIO) (FileInfo, error) {
	file, err := f.Header.Open()
	if err != nil {
		return FileInfo{}, err
	}

	err = q.InsertFile(ctx, InsertFileConfig{
		ID:          f.ID,
		UserID:      f.UserID,
		DirectoryID: f.DirectoryID,
		Name:        f.Header.Filename,
		UploadedAt:  time.Now().UTC(),
	})
	if err != nil {
		return FileInfo{}, err
	}

	userPath, err := io.paths.GetFile(ctx, q, f.DirectoryID, f.Header.Filename)
	if err != nil {
		return FileInfo{}, err
	}

	fsPath, err := io.paths.GetFileFS(ctx, q, f.DirectoryID, f.ID)
	if err != nil {
		return FileInfo{}, err
	}

	// Create the file and set the file permissions on the file system.
	dst, err := io.fs.Create(fsPath, f.FSPerm)
	if err != nil {
		return FileInfo{}, err
	}

	// Write the file content to the file on the file system.
	_, err = io.fs.Copy(dst, file)
	if err != nil {
		return FileInfo{}, err
	}
	dst.Close()

	return FileInfo{
		ID:          f.ID,
		OwnerID:     f.UserID,
		DirectoryID: f.DirectoryID,
		Name:        f.Header.Filename,
		Path:        userPath,
		Size:        f.Header.Size,
		UploadedAt:  f.UploadedAt.UTC(),
		FSPath:      fsPath,
	}, nil
}

type ReadFileInfoIO struct {
	UserID string
	FileID string
	FSPath string
}

// FileInfo gets the information for a users file. The information is gathered
// from both the database and file system, and returns it as a FileInfo.
//
// The actual file content is not returned by this function.
func (io *IO) ReadFileInfo(ctx context.Context, q *Query, f ReadFileInfoIO) (FileInfo, error) {
	row, err := q.SelectFileByIDUser(ctx, f.FileID, f.UserID)
	if err != nil {
		return FileInfo{}, err
	}

	userPath, err := io.paths.GetFile(ctx, q, row.DirectoryID, row.Name)
	if err != nil {
		return FileInfo{}, err
	}

	fsPath, err := io.paths.GetFileFS(ctx, q, row.DirectoryID, row.ID)
	if err != nil {
		return FileInfo{}, err
	}

	// Get the file size on the file system.
	stat, err := io.fs.Stat(fsPath)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		ID:          row.ID,
		OwnerID:     row.UserID,
		DirectoryID: row.DirectoryID,
		Name:        row.Name,
		Path:        userPath,
		Size:        stat.Size(),
		UploadedAt:  row.UploadedAt.UTC(),
		FSPath:      fsPath,
	}, nil
}
