package cloudstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
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

// NewUserRootDir creates a new root directory for a user. All sub directories will be
// persisted under this directory. Every users root directory will be named "root". The
// file permissions are set to 0700.
//
// This is a shortcut for NewUserDir(ctx, userID, "root", "").
//
// The directory ID and name on the file system will be a randomly generated UUID.
// The path "/" will correspond to this directory.
func (s *Service) NewUserRootDir(ctx context.Context, userID string) (Dir, error) {
	return s.newUserDir(ctx, userID, "root", "")
}

// NewUserDir creates a new directory for a user under a specific parent directory. The
// file permissions are set to 0700.
//
// The directory ID and name on the file system will be a randomly generated UUID.
func (s *Service) NewUserDir(ctx context.Context, userID string, name string, parentID string) (Dir, error) {
	if parentID == "" {
		// TODO: Get the id of the users root directory and use that as the parentID.
	}

	dir, err := s.newUserDir(ctx, userID, name, parentID)
	if err != nil {
		if errors.Is(err, ErrForeignKeyParentID) {
			return Dir{}, app.Wrap(app.WrapParams{
				Err:         fmt.Errorf("parent directory does not exist [parent_id: %s]: %w", parentID, err),
				SafeMessage: fmt.Sprintf("Directory %q does not exist", parentID),
				StatusCode:  http.StatusBadRequest,
			})
		}
	}

	return dir, nil
}

func (s *Service) newUserDir(ctx context.Context, userID string, name string, parentID string) (Dir, error) {
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
		// At this point the file was written to disk, so the application is in a inconsistent state.
		// Remove the directory since the information failed to be commited to the database.
		if errors.Is(err, ErrCommitTx) {
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
