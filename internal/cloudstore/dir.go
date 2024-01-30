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

// DirService is the business logic for the cloudstore directory
// functionality.
//
// DirService should be created using the NewDirService function.
type DirService struct {
	store   *Store
	io      *IO
	log     *log.Logger
	pathMap *PathMapper
}

// DirServiceConfig is the DirService configuration.
type DirServiceConfig struct {
	Store   *Store
	IO      *IO
	Log     *log.Logger
	PathMap *PathMapper
}

// NewDirService creates a new DirService.
//
// Store and PathMap must be set otherwise it will panic.
//
// If IO is not set, it will default to NewIO(&OSFileSystem{}, c.PathMap).
// If initializing multiple cloudstore services, it is recommended to
// use the same IO with all services. This will avoid initializing
// multiple IO's.
//
// If Log is not set, it will default to log.Default().
func NewDirService(c DirServiceConfig) *DirService {
	if c.Store == nil {
		panic("cloudstore.NewDirService: cannot create DirService with nil Store")
	}

	if c.PathMap == nil {
		panic("cloudstore.NewDirService: cannot create DirService with nil PathMap")
	}

	if c.IO == nil {
		c.IO = NewIO(&OSFileSystem{}, c.PathMap)
	}

	if c.Log == nil {
		c.Log = log.Default()
	}

	return &DirService{
		store:   c.Store,
		io:      c.IO,
		log:     c.Log,
		pathMap: c.PathMap,
	}
}

// SetupRoot will validate that the root storage directory exists. If it
// does not exist, it will be created.
//
// The directory created will have permissions 0700. Only the system user
// that is running this application can read, write, and execute this
// directory.
//
// This method should be called once before executing any other DirService methods.
func (s *DirService) SetupRoot() error {
	return s.io.SetupFSRoot(0700)
}

type Dir struct {
	ID        string
	Owner     string
	ParentID  string
	Name      string
	Path      string
	CreatedAt time.Time
	UpdatedAt time.Time
	LastWrite time.Time
	fsPath    string
}

// NewUser creates a new root directory for a user. All sub directories will be
// persisted under this directory. Every users root directory will be named "root". The
// file permissions are set to 0700.
//
// The directory ID and name on the file system will be a randomly generated UUID.
// The path "/" will correspond to this directory.
func (s *DirService) NewUser(ctx context.Context, userID string) (Dir, error) {
	return s.write(ctx, userID, "root", "")
}

// NewPath creates a new directory for a user under the provided path. The file
// permissions are set to 0700. The path is cleaned using the filepath.Clean func.
// An empty path will default to the users root directory.
//
// NewPath validates that a users root directory has been created. If it does not exist
// it will create it.
//
// The directory ID and name on the file system will be a randomly generated UUID.
func (s *DirService) NewPath(ctx context.Context, userID string, name string, path string) (Dir, error) {
	return s.new(ctx, userID, name, func(rootID string) (string, error) {
		return s.pathMap.FindDir(ctx, s.store.Query, PathSearch{
			UserID: userID,
			RootID: rootID,
			Path:   path,
		})
	})
}

// New creates a new directory for a user under a specific parent directory. The
// file permissions are set to 0700. If parentID is empty, it will default to the users
// root directory.
//
// New validates that a users root directory has been created. If it does not exist
// it will create it.
//
// The directory ID and name on the file system will be a randomly generated UUID.
func (s *DirService) New(ctx context.Context, userID string, name string, parentID string) (Dir, error) {
	return s.new(ctx, userID, name, func(rootID string) (string, error) {
		if parentID == "" {
			return rootID, nil
		}

		return parentID, nil
	})
}

// new creates a new directory with the given name. The root directory is validated
// and then passes the root directory ID to getParentID. This function should return
// the ID of the parent directory that the new directory will be written under.
//
// If getParentID returns an error, new will not modify it and return it as is.
//
// If name is empty an error is returned.
func (s *DirService) new(ctx context.Context, userID string, name string, getParentID idFunc) (Dir, error) {
	if name == "" {
		return Dir{}, app.Wrap(app.WrapParams{
			Err:         errors.New("empty directory name"),
			SafeMessage: "Directory name cannot be empty",
			StatusCode:  http.StatusBadRequest,
		})
	}

	root, err := s.ValidateUser(ctx, userID)
	if err != nil {
		return Dir{}, err
	}

	parentID, err := getParentID(root.ID)
	if err != nil {
		return Dir{}, err
	}

	return s.write(ctx, userID, name, parentID)
}

// write writes and returns a user directory. The location of the directory is defined by
// the parentID. Directories will be a direct child of the parent.
//
// This DirService's io is used to persist the directory to the file system, and store the
// information in the database. The operation is wrapped in a database transaction. If
// commiting the transaction fails, it will attempt to delete the directory from the file
// system.
func (s *DirService) write(ctx context.Context, userID string, name string, parentID string) (Dir, error) {
	var dir Dir

	err := s.store.Tx(ctx, func(tx *db.Tx) error {
		dirIO, err := s.io.NewDir(ctx, NewQuery(tx), NewDirIO{
			ID:        uuid.NewString(),
			UserID:    userID,
			Name:      name,
			ParentID:  sql.NullString{String: parentID, Valid: parentID != ""},
			CreatedAt: time.Now().UTC(),
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
				SafeMessage: fmt.Sprintf("Directory '%s' does not exist", parentID),
				StatusCode:  http.StatusBadRequest,
			})
		case errors.Is(err, ErrUniqueNameParentID):
			err = app.Wrap(app.WrapParams{
				Err:         fmt.Errorf("directory name not available [name: %s, parent_id: %s]: %w", name, parentID, err),
				SafeMessage: fmt.Sprintf("Directory '%s' already exists", name),
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
			go s.Remove(ctx, dir.fsPath)
		}

		return Dir{}, err
	}

	return dir, nil
}

// Remove accepts the path to a directory and removes it from the file system.
// All sub directories and files will be removed. If directory cannot be removed,
// the path will be logged.
//
// Remove only removes the directory from the file system. The database remains
// unchanged.
func (s *DirService) Remove(ctx context.Context, fsPath string) {
	err := s.io.RemoveFSDir(fsPath)
	if err != nil {
		s.log.Printf("[ERROR] Removing directory [path: %s]: %v\n", fsPath, err)
	}
}

// ValidateUser validates that a root directory exists for the user. If it does
// not exist, a root directory will be created for the user. The root directory is
// returned.
//
// If a root directory exists and cannot be created, there is a serious problem with
// the account and/or server. A app.WrappedSafeError will be returned with a message
// stating they need to contact support.
func (s *DirService) ValidateUser(ctx context.Context, userID string) (Dir, error) {
	row, err := s.store.SelectUserRootDirectory(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			rootDir, err := s.write(ctx, userID, "root", "")
			if err != nil {
				return Dir{}, app.Wrap(app.WrapParams{
					Err:         fmt.Errorf("creating users root directory: %w", err),
					SafeMessage: "There is a problem with your accounts root storage. Please contact us.",
					StatusCode:  http.StatusInternalServerError,
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
