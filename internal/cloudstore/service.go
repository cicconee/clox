package cloudstore

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	path string
}

func NewService(path string) *Service {
	return &Service{path: path}
}

// SetupRoot will validate that a directory exists with the value of path
// in this Service. If it does not exist, it will be created.
//
// The directory created will have permissions 0700. Only the system user
// that is running this application can read, write, and execute this
// directory.
//
// This method should be called once before executing any other methods.
func (s *Service) SetupRoot() error {
	_, err := os.Stat(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.Mkdir(s.path, 0700); err != nil {
				return fmt.Errorf("creating directory [%s]: %w", s.path, err)
			}

			return nil
		}

		return fmt.Errorf("getting file info [%s]: %w", s.path, err)
	}

	return nil
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

// NewDir creates a new directory for a user under the root directory.
//
// The directory ID and name on the file system will be a randomly generated UUID.
// The name passed to this method will act as a label so the user can identify their
// directories.
//
// SetupRoot must be called before calling this method.
func (s *Service) NewDir(ctx context.Context, name string, userID string) (*Dir, error) {
	// TODO:
	//
	// Create the directory under the root path with permissions 0700.
	//
	// Insert the directory data into the database (ID, name, user ID, created at, etc.).
	//
	// Return the directory ID or maybe create a directory struct and return that.

	return &Dir{
		ID:        uuid.NewString(),
		Owner:     userID,
		Path:      "/",
		CreatedAt: time.Now().UTC(),
	}, nil
}
