package cloudstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/cicconee/clox/internal/app"
	"github.com/cicconee/clox/internal/db"
	"github.com/google/uuid"
)

// Service is the business logic for the cloudstore package.
// Interact with the cloud storage capabilities with this struct.
type Service struct {
	path  string
	store *Store
	io    *IO
	log   *log.Logger
}

// NewService created a new Service.
func NewService(path string, store *Store, io *IO, log *log.Logger) *Service {
	return &Service{path: path, store: store, io: io, log: log}
}

// SetupRoot will validate that a directory exists with the value of path
// in this Service. If it does not exist, it will be created.
//
// The directory created will have permissions 0700. Only the system user
// that is running this application can read, write, and execute this
// directory.
//
// This method should be called once before executing any other Service methods.
func (s *Service) SetupRoot() error {
	return s.io.SetupFSRoot(s.path, 0700)
}

type Dir struct {
	ID        string
	Owner     string
	Name      string
	Path      string
	CreatedAt time.Time
	UpdatedAt time.Time
	LastWrite time.Time
}

// NewUserDir creates a new root directory for a user. All sub directories will be
// persisted under this directory. Every users root directory will be named "root". The
// file permissions are set to 0700.
//
// The directory ID and name on the file system will be a randomly generated UUID.
// The path "/" will correspond to this directory.
func (s *Service) NewUserDir(ctx context.Context, userID string) (Dir, error) {
	return s.writeDir(ctx, userID, "root", "")
}

// NewDirPath creates a new directory for a user under the provided path. The file
// permissions are set to 0700. The path is cleaned using the filepath.Clean func.
// An empty path will default to the users root directory.
//
// NewDirPath validates that a users root directory has been created. If it does not exist
// it will create it.
//
// The directory ID and name on the file system will be a randomly generated UUID.
func (s *Service) NewDirPath(ctx context.Context, userID string, name string, pathStr string) (Dir, error) {
	if name == "" {
		return Dir{}, app.Wrap(app.WrapParams{
			Err:         errors.New("empty directory name"),
			SafeMessage: "Directory name cannot be empty",
			StatusCode:  http.StatusBadRequest,
		})
	}

	root, err := s.ValidateUserDir(ctx, userID)
	if err != nil {
		return Dir{}, err
	}

	// Clean and parse the path.
	fp := filepath.Clean(pathStr)
	var p string
	if fp == "." || fp == "/" {
		p = "root"
	} else if fp[0] == '/' {
		p = "root" + fp
	} else {
		p = "root/" + fp
	}
	path := strings.Split(p, "/")

	parentID := root.ID
	for i := 1; i < len(path); i++ {
		pName := path[i]
		dir, err := s.store.SelectDirectoryByUserNameParent(ctx, userID, pName, parentID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return Dir{}, app.Wrap(app.WrapParams{
					Err:         fmt.Errorf("directory %q does not exist [path: %s]", pName, pathStr),
					SafeMessage: fmt.Sprintf("Directory %q does not exist", pName),
					StatusCode:  http.StatusBadRequest,
				})
			}

			return Dir{}, err
		}

		parentID = dir.ID
	}

	return s.writeDir(ctx, userID, name, parentID)
}

// NewDir creates a new directory for a user under a specific parent directory. The
// file permissions are set to 0700. If parentID is empty, it will default to the users
// root directory.
//
// NewDir validates that a users root directory has been created. If it does not exist
// it will create it.
//
// The directory ID and name on the file system will be a randomly generated UUID.
func (s *Service) NewDir(ctx context.Context, userID string, name string, parentID string) (Dir, error) {
	if name == "" {
		return Dir{}, app.Wrap(app.WrapParams{
			Err:         errors.New("empty directory name"),
			SafeMessage: "Directory name cannot be empty",
			StatusCode:  http.StatusBadRequest,
		})
	}

	root, err := s.ValidateUserDir(ctx, userID)
	if err != nil {
		return Dir{}, err
	}

	if parentID == "" {
		parentID = root.ID
	}

	return s.writeDir(ctx, userID, name, parentID)
}

// writeDir writes and returns a user directory. The location of the directory is defined by
// the parentID. Directories will be a direct child of the parent.
//
// This Service's io is used to persist the directory to the file system, and store the
// information in the database. The operation is wrapped in a database transaction. If
// commiting the transaction fails, it will attempt to delete the directory from the file
// system.
func (s *Service) writeDir(ctx context.Context, userID string, name string, parentID string) (Dir, error) {
	var dir DirIO

	err := s.store.Tx(ctx, func(tx *db.Tx) error {
		dirIO, err := s.io.NewDir(ctx, NewQuery(tx), NewDirIO{
			ID:        uuid.NewString(),
			UserID:    userID,
			Name:      name,
			ParentID:  sql.NullString{String: parentID, Valid: parentID != ""},
			CreatedAt: time.Now().UTC(),
			FSDir:     s.path,
			FSPerm:    0700,
		})
		if err != nil {
			return err
		}

		dir = dirIO
		return nil
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrForeignKeyParentID):
			err = app.Wrap(app.WrapParams{
				Err:         fmt.Errorf("parent directory does not exist [parent_id: %s]: %w", parentID, err),
				SafeMessage: fmt.Sprintf("Directory %q does not exist", parentID),
				StatusCode:  http.StatusBadRequest,
			})
		case errors.Is(err, ErrUniqueNameParentID):
			err = app.Wrap(app.WrapParams{
				Err:         fmt.Errorf("directory name not available [name: %s, parent_id: %s]: %w", name, parentID, err),
				SafeMessage: fmt.Sprintf("Directory %q already exists", name),
				StatusCode:  http.StatusBadRequest,
			})
		case errors.Is(err, ErrSyntaxParentID):
			err = app.Wrap(app.WrapParams{
				Err:         fmt.Errorf("invalid parent directory [parent_id: %s]: %w", parentID, err),
				SafeMessage: fmt.Sprintf("'%s' is a invalid parent ID", parentID),
				StatusCode:  http.StatusBadRequest,
			})
		case errors.Is(err, ErrCommitTx):
			// At this point the file was written to disk, so the application is in a inconsistent state.
			// Remove the directory since the information failed to be commited to the database.
			go s.RemoveDir(ctx, dir.FSPath)
		}

		return Dir{}, err
	}

	return Dir{
		ID:        dir.ID,
		Owner:     dir.UserID,
		Name:      dir.Name,
		Path:      dir.UserPath,
		CreatedAt: dir.CreatedAt,
		UpdatedAt: dir.UpdatedAt,
		LastWrite: dir.LastWrite,
	}, nil
}

// RemoveDir accepts the path to a directory and removes it from the file system.
// All sub directories and files will be removed. If directory cannot be removed,
// the path will be logged.
//
// RemoveDir only removes the directory from the file system. The database remains
// unchanged.
func (s *Service) RemoveDir(ctx context.Context, fsPath string) {
	err := s.io.RemoveFSDir(fsPath)
	if err != nil {
		s.log.Printf("[ERROR] Removing directory [path: %s]: %v\n", fsPath, err)
	}
}

// ValidateUserDir validates that a root directory exists for the user. If it does
// not exist, a root directory will be created for the user. The root directory is
// returned.
//
// If a root directory exists and cannot be created, there is a serious problem with
// the account and/or server. A app.WrappedSafeError will be returned with a message
// stating they need to contact support.
func (s *Service) ValidateUserDir(ctx context.Context, userID string) (Dir, error) {
	row, err := s.store.SelectUserRootDirectory(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			rootDir, err := s.writeDir(ctx, userID, "root", "")
			if err != nil {
				return Dir{}, app.Wrap(app.WrapParams{
					Err:         fmt.Errorf("creating users root directory: %w", err),
					SafeMessage: "There is a problem with your accounts root storage. Please contact us.",
					StatusCode:  http.StatusBadRequest,
				})
			}

			return rootDir, nil
		}

		return Dir{}, err
	}

	return Dir{
		ID:        row.ID,
		Owner:     row.UserID,
		Name:      row.Name,
		Path:      "/",
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt.Time,
		LastWrite: row.LastWrite.Time,
	}, nil
}
