package middleware

import (
	"fmt"
	"github.com/cicconee/clox/internal/user"
	"github.com/cicconee/clox/internal/web"
	"github.com/cicconee/clox/internal/web/cookie"
	"github.com/cicconee/clox/internal/web/session"
	"log"
	"net/http"
)

// Registry is a http middleware that can validate if a user is registered or not.
type Registry struct {
	cookies *cookie.Manager
	logger  *log.Logger
}

func NewRegistry(cookies *cookie.Manager, logger *log.Logger) *Registry {
	return &Registry{cookies: cookies, logger: logger}
}

// IsRegistered validates that a user already registered using the current session.
//
// IsRegistered must have a session in the request context. Execute the *Session.Active middleware function before
// calling NotRegistered to set the session. Alternatively, use the session.SetSessionContext to set the session.
func (r *Registry) IsRegistered(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, rq *http.Request) {
		u := session.GetUserContext(rq.Context())

		if u.RegistrationStatus == user.Complete {
			next(w, rq)
			return
		}

		var redirect, flashMessage, flashError string
		switch u.RegistrationStatus {
		case user.Incomplete:
			redirect = web.URLRegister
			flashMessage = fmt.Sprintf("Hello %s! Choose a username.", u.FirstName)
		case user.Blocked:
			redirect = web.URLLogin
			flashError = "You are blocked. Please contact us."
		default:
			redirect = web.URLLogin
			flashError = "Something went wrong. Please try again."
			r.logger.Printf("[WARNING] [%s %s] Unexpected user registration status: %s", rq.Method, rq.URL.Path, u.RegistrationStatus)
		}

		r.cookies.Set(w, cookie.FlashMessage, flashMessage)
		r.cookies.Set(w, cookie.FlashError, flashError)
		http.Redirect(w, rq, redirect, http.StatusFound)
	}
}

// NotRegistered validates that a user has not yet registered using the current session.
//
// NotRegistered must have a session in the request context. Execute the *Session.Active middleware function before
// calling NotRegistered to set the session. Alternatively, use the session.SetSessionContext to set the session.
func (r *Registry) NotRegistered(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, rq *http.Request) {
		u := session.GetUserContext(rq.Context())

		if u.RegistrationStatus == user.Incomplete {
			next(w, rq)
			return
		}

		var redirect, flashMessage, flashError string
		switch u.RegistrationStatus {
		case user.Complete:
			redirect = web.URLDashboard
			flashMessage = fmt.Sprintf("%s, you are already registered.", u.FirstName)
		case user.Blocked:
			redirect = web.URLLogin
			flashError = "You are blocked. Please contact us."
		default:
			redirect = web.URLLogin
			flashError = "Something went wrong. Please try again."
			r.logger.Printf("[WARNING] [%s %s] Unexpected user registration status: %s", rq.Method, rq.URL.Path, u.RegistrationStatus)
		}

		r.cookies.Set(w, cookie.FlashMessage, flashMessage)
		r.cookies.Set(w, cookie.FlashError, flashError)
		http.Redirect(w, rq, redirect, http.StatusFound)
	}
}
