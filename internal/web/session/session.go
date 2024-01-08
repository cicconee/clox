package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cicconee/clox/internal/cache"
	"github.com/cicconee/clox/internal/user"
)

var (
	ErrNoSession = errors.New("no session")
)

type Manager struct {
	cache *cache.Redis
}

func NewManager(cache *cache.Redis) *Manager {
	return &Manager{cache: cache}
}

type User struct {
	SessionID          string      `json:"session_id"`
	UserID             string      `json:"user_id"`
	FirstName          string      `json:"first_name"`
	LastName           string      `json:"last_name"`
	PictureURL         string      `json:"picture_url"`
	Email              string      `json:"email"`
	Username           string      `json:"username"`
	RegistrationStatus user.Status `json:"registration_status"`
}

// Set sets the user and the user-to-session mapping in the session storage. Sessions will be created
// with a key formatted as session:{session-id}. The user-to-session mapping will be created with a
// key formatted as user:{user-id}:session.
//
// A user-to-session mapping is used for administrative purposes. For example, if a user needs to be
// blocked, and they have an active session, the session key can be found using this mapping.
func (m *Manager) Set(ctx context.Context, user User) error {
	encodedUser, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("marshalling user to JSON: %w", err)
	}

	sKey := sessionKey(user.SessionID)
	err = m.cache.SetTx(ctx,
		cache.SetTxParams{Key: sKey, Val: encodedUser, Exp: 0},
		cache.SetTxParams{Key: userSessionMappingKey(user.UserID), Val: sKey, Exp: 0})
	if err != nil {
		return fmt.Errorf("setting session in cache: %w", err)
	}

	return nil
}

func (m *Manager) Get(ctx context.Context, sessionID string) (User, error) {
	encodedUser, err := m.cache.Get(ctx, sessionKey(sessionID))
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return User{}, fmt.Errorf("%w [sessionID: %s]", ErrNoSession, sessionID)
		}

		return User{}, fmt.Errorf("getting user session from cache [sessionID: %s]: %w", sessionID, err)
	}

	var user User
	err = json.Unmarshal([]byte(encodedUser), &user)
	if err != nil {
		return User{}, fmt.Errorf("unmarshalling user session [sessionID: %s]: %w", sessionID, err)
	}

	return user, nil
}

// Del deletes user from the session storage. This includes deleting both the user and the user-to-session mapping.
func (m *Manager) Del(ctx context.Context, user User) error {
	if err := m.cache.Del(ctx,
		sessionKey(user.SessionID),
		userSessionMappingKey(user.UserID)); err != nil {
		return fmt.Errorf("deleting user session [sessionID: %s, userID: %s]: %w", user.SessionID, user.UserID, err)
	}

	return nil
}

func sessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

func userSessionMappingKey(userID string) string {
	return fmt.Sprintf("user:%s:session", userID)
}
