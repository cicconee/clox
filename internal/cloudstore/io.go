package cloudstore

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
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

type NewDirIO struct {
	ID        string
	UserID    string
	Name      string
	ParentID  sql.NullString
	CreatedAt time.Time
	FSDir     string
	FSPerm    fs.FileMode
}

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

	fsPath := fmt.Sprintf("%s/%s", d.FSDir, strings.Join(idPath, "/"))
	if err := io.fs.Mkdir(fsPath, d.FSPerm); err != nil {
		return DirIO{}, fmt.Errorf("creating directory [%s]: %w", fsPath, err)
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
