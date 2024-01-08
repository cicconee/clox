package auth

import (
	"context"
	"fmt"

	"github.com/cicconee/clox/internal/user"
	"github.com/cicconee/clox/internal/web/session"
)

// Registry registers users for Clox.
type Registry struct {
	users    *user.Service
	sessions *session.Manager
}

// NewRegistry creates a new Registry.
func NewRegistry(users *user.Service, sessions *session.Manager) *Registry {
	return &Registry{users: users, sessions: sessions}
}

// Register persists a user. Once registered, the session is updated to reflect the users new state.
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

	return nil
}

// Logout logs out a user. The session is deleted from the cache.
func (r *Registry) Logout(ctx context.Context, user session.User) error {
	return r.sessions.Del(ctx, user)
}
