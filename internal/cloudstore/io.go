package cloudstore

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"mime/multipart"
	"strings"
	"time"
)

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

type IO struct {
	fs *OSFileSystem
}

func NewIO(fs *OSFileSystem) *IO {
	return &IO{fs: fs}
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

	// Get the path used on the file system.
	idPath, err := q.SelectDirectoryFSPath(ctx, d.ID)
	if err != nil {
		return DirIO{}, err
	}

	// Get the path the user will reference.
	namePath, err := q.SelectDirectoryPath(ctx, d.ID)
	if err != nil {
		return DirIO{}, err
	}

	// Ignore the "root" level.
	userPath := strings.Join(namePath, "/")
	userPath = strings.TrimPrefix(userPath, "root")
	if userPath == "" {
		userPath = "/"
	}

	// Ensure writing the directory to the file system is the last operation.
	fsPath := fmt.Sprintf("%s/%s", d.FSDir, strings.Join(idPath, "/"))
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
func (i *IO) NewFile(ctx context.Context, q *Query, f NewFileIO) (FileIO, error) {
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

	// Construct the path the user will reference.
	dirNamePath, err := q.SelectDirectoryPath(ctx, f.DirectoryID)
	if err != nil {
		return FileIO{}, err
	}
	userPath := strings.Join(dirNamePath, "/")
	userPath = strings.TrimPrefix(userPath, "root")
	if userPath == "" {
		userPath = "/" + f.Header.Filename
	} else {
		userPath += "/" + f.Header.Filename
	}

	// Construct the full file system path of the file.
	dirIDPath, err := q.SelectDirectoryFSPath(ctx, f.DirectoryID)
	if err != nil {
		return FileIO{}, err
	}
	fsPath := fmt.Sprintf("%s/%s/%s", f.FSDir, strings.Join(dirIDPath, "/"), f.ID)

	// Create the file and set the file permissions on the file system.
	dst, err := i.fs.Create(fsPath, f.FSPerm)
	if err != nil {
		return FileIO{}, err
	}

	// Write the file content to the file on the file system.
	_, err = io.Copy(dst, file) // TODO: Wrap this call in OSFileSystem.
	if err != nil {
		// TODO: Wrap in specific error signifying this error happened after
		// the file was written to disk.
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
