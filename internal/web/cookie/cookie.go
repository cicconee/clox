package cookie

import (
	"net/http"
	"time"
)

const (
	OAuth2State  string = "oauth2_state"
	Session      string = "session_id"
	FlashMessage string = "flash_message"
	FlashError   string = "flash_error"
)

// Manager is a cookie manager.
type Manager struct {
	secure bool
	host   string
}

// NewManager creates a new cookie manager.
func NewManager(secure bool, host string) *Manager {
	return &Manager{secure: secure, host: host}
}

// Get gets a cookie.
func (m *Manager) Get(r *http.Request, key string) (*http.Cookie, error) {
	return r.Cookie(key)
}

// Set sets a cookie.
func (m *Manager) Set(w http.ResponseWriter, key string, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     key,
		Value:    value,
		HttpOnly: true,
		Secure:   m.secure,
		Domain:   m.host,
		Path:     "/",
	})
}

// SetExpires sets a cookie with an expiration.
func (m *Manager) SetExpires(w http.ResponseWriter, key string, value string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     key,
		Value:    value,
		HttpOnly: true,
		Secure:   m.secure,
		Domain:   m.host,
		Path:     "/",
		Expires:  expires,
	})
}

// Clear clears a cookie.
func (m *Manager) Clear(w http.ResponseWriter, key string) {
	m.SetExpires(w, key, "", time.Now())
}
