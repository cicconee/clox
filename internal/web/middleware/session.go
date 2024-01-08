package middleware

import (
	"errors"
	"log"
	"net/http"

	"github.com/cicconee/clox/internal/web"
	"github.com/cicconee/clox/internal/web/cookie"
	"github.com/cicconee/clox/internal/web/session"
)

type Session struct {
	sessions *session.Manager
	cookies  *cookie.Manager
	logger   *log.Logger
}

func NewSession(sessions *session.Manager, cookies *cookie.Manager, logger *log.Logger) *Session {
	return &Session{sessions: sessions, cookies: cookies, logger: logger}
}

// Active is a http middleware that validates a user session. The user session (session.User) is injected into the
// http request context.
//
// Active should wrap http handlers that require an active session.
func (s *Session) Active(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := s.cookies.Get(r, cookie.Session)
		if err != nil {
			http.Redirect(w, r, web.URLLogin, http.StatusFound)
			return
		}

		user, err := s.sessions.Get(r.Context(), sessionCookie.Value)
		if err != nil {
			if errors.Is(err, session.ErrNoSession) {
				s.cookies.Set(w, cookie.FlashError, "Please log in again.")
			} else {
				s.cookies.Set(w, cookie.FlashError, "Something went wrong. Please log in again.")
				s.logger.Printf("[ERROR] [%s %s] Getting user session: %v\n", r.Method, r.URL.Path, err)
			}

			s.cookies.Clear(w, cookie.Session)
			http.Redirect(w, r, web.URLLogin, http.StatusFound)
			return
		}

		ctx := session.SetUserContext(r.Context(), user)
		next(w, r.WithContext(ctx))
	}
}

// Inactive is a http middleware that validates no active session.
//
// Inactive should wrap http handlers that require an inactive session.
func (s *Session) Inactive(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := s.cookies.Get(r, cookie.Session)
		if err != nil {
			// Only possible error is http.ErrNoCookie.
			next(w, r)
			return
		}

		// If there is a session cookie, verify it is not active in sessions.
		_, err = s.sessions.Get(r.Context(), sessionCookie.Value)
		if err != nil {
			if !errors.Is(err, session.ErrNoSession) {
				s.logger.Printf("[ERROR] [%s %s] Getting session: %v\n", r.Method, r.URL.Path, err)
			}

			s.cookies.Clear(w, cookie.Session)
			next(w, r)
			return
		}

		s.cookies.Set(w, cookie.FlashMessage, "You are already logged in.")
		http.Redirect(w, r, web.URLDashboard, http.StatusFound)
	}
}
