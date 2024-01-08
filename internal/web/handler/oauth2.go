package handler

import (
	"fmt"
	"github.com/cicconee/clox/internal/user"
	"github.com/cicconee/clox/internal/web"
	"github.com/cicconee/clox/internal/web/auth"
	"github.com/cicconee/clox/internal/web/cookie"
	"log"
	"net/http"
	"time"
)

// OAuth2 encapsulates the redirect and callback handlers for a oauth2 provider.
type OAuth2 struct {
	auth    *auth.Authenticator
	cookies *cookie.Manager
	logger  *log.Logger
}

// NewOAuth2 creates a new OAuth2 handler.
func NewOAuth2(auth *auth.Authenticator, cookies *cookie.Manager, logger *log.Logger) *OAuth2 {
	return &OAuth2{auth: auth, cookies: cookies, logger: logger}
}

func (o *OAuth2) Redirect() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url, state := o.auth.Generate()
		o.cookies.SetExpires(w, cookie.OAuth2State, state, time.Now().Add(time.Minute*5))
		http.Redirect(w, r, url, http.StatusFound)
	}
}

func (o *OAuth2) Callback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, err := o.cookies.Get(r, cookie.OAuth2State)
		if err != nil {
			o.logger.Printf("[ERROR] [%s %s] Getting oauth2 state cookie: %v\n", r.Method, r.URL.Path, err)
			o.cookies.Set(w, cookie.FlashError, "Something went wrong. Please try again.")
			http.Redirect(w, r, web.URLLogin, http.StatusFound)
			return
		}

		session, err := o.auth.Authenticate(r, state.Value)
		if err != nil {
			o.logger.Printf("[ERROR] [%s %s] Authenticating: %v", r.Method, r.URL.Path, err)
			o.cookies.Set(w, cookie.FlashError, "Authentication failed. Please try again.")
			http.Redirect(w, r, web.URLLogin, http.StatusFound)
			return
		}

		var redirect, flashMessage, flashError string
		switch session.RegistrationStatus {
		case user.Complete:
			// User is registered.
			redirect = web.URLDashboard
			flashMessage = fmt.Sprintf("Welcome back %s!", session.FirstName)
		case user.Incomplete:
			// User authenticated but is not registered.
			redirect = web.URLRegister
			flashMessage = fmt.Sprintf("Welcome, %s! Choose your username.", session.FirstName)
		case user.Blocked:
			// User is blocked.
			redirect = web.URLLogin
			flashError = "You are blocked. Please contact us."
		default:
			// User has a unexpected registration status.
			redirect = web.URLLogin
			flashError = "Something went wrong. Please try again."
			o.logger.Printf("[WARNING] [%s %s] Unexpected user registration status: %s", r.Method, r.URL.Path, session.RegistrationStatus)
		}

		o.cookies.Clear(w, cookie.OAuth2State)
		o.cookies.Set(w, cookie.Session, session.SessionID)
		o.cookies.Set(w, cookie.FlashMessage, flashMessage)
		o.cookies.Set(w, cookie.FlashError, flashError)

		http.Redirect(w, r, redirect, http.StatusFound)
	}
}
