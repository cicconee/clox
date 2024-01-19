package auth

import (
	"context"
	"fmt"
	"log"

	"github.com/cicconee/clox/internal/cloudstore"
	"github.com/cicconee/clox/internal/user"
	"github.com/cicconee/clox/internal/web/session"
)

// Registry registers users for Clox.
type Registry struct {
	users    *user.Service
	sessions *session.Manager
	dirs     *cloudstore.DirService
	log      *log.Logger
}

// NewRegistry creates a new Registry.
func NewRegistry(users *user.Service, sessions *session.Manager, dirs *cloudstore.DirService, log *log.Logger) *Registry {
	return &Registry{users: users, sessions: sessions, dirs: dirs, log: log}
}

// Register persists a user. Once registered, the session is updated to reflect the users new state.
//
// Upon success, a root storage directory is created for the user. If creating the directory fails
// it will be logged.
func (r *Registry) Register(ctx context.Context, session session.User) error {
	user, err := r.users.Register(ctx, user.Registration{
		ID:         session.UserID,
		FirstName:  session.FirstName,
		LastName:   session.LastName,
		PictureURL: session.PictureURL,
		Email:      session.Email,
		Username:   session.Username,
	})
	if err != nil {
		return fmt.Errorf("registering user: %w", err)
	}

	session.Username = user.Username
	session.RegistrationStatus = user.RegistrationStatus
	err = r.sessions.Set(ctx, session)
	if err != nil {
		return fmt.Errorf("setting user session: %w", err)
	}

	dir, err := r.dirs.NewUserDir(ctx, user.ID)
	if err != nil {
		r.log.Printf("[ERROR] Creating root storage [user: %s]: %v\n", user.ID, err)
	} else {
		r.log.Printf("[INFO] Created root storage [user: %s, dir: %s]\n", user.ID, dir.ID)
	}

	return nil
}

// Logout logs out a user. The session is deleted from the cache.
func (r *Registry) Logout(ctx context.Context, user session.User) error {
	return r.sessions.Del(ctx, user)
}
