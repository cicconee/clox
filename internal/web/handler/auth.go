package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cicconee/clox/internal/app"
	"github.com/cicconee/clox/internal/web"
	"github.com/cicconee/clox/internal/web/auth"
	"github.com/cicconee/clox/internal/web/cookie"
	"github.com/cicconee/clox/internal/web/session"
	"github.com/cicconee/clox/internal/web/template"
)

// Auth encapsulates the handlers for authenticating with the server side app.
type Auth struct {
	registry *auth.Registry
	cookies  *cookie.Manager
	tmpl     *template.Template
	log      *log.Logger
}

// NewAuth creates a new Auth.
func NewAuth(registry *auth.Registry, cookies *cookie.Manager, tmpl *template.Template, log *log.Logger) *Auth {
	return &Auth{registry: registry, cookies: cookies, tmpl: tmpl, log: log}
}

// TemplateLogin executes the login template.
func (a *Auth) TemplateLogin() http.HandlerFunc {
	type data struct {
		Google web.Link
	}

	return func(w http.ResponseWriter, r *http.Request) {
		d := data{
			Google: web.Link{URL: web.URLGoogleLogin, Value: "Sign in with Google"},
		}

		a.tmpl.Execute(w, r, "login", template.ExecuteParams{
			Title:    "Authenticate",
			PageID:   web.PageLogin,
			NavLinks: []web.NavLink{web.NavLinkLogin},
			Data:     d,
		})
	}
}

// TemplateRegister executes the register template.
func (a *Auth) TemplateRegister() http.HandlerFunc {
	type data struct {
		LinkRegister web.Link
	}

	return func(w http.ResponseWriter, r *http.Request) {
		data := data{
			LinkRegister: web.Link{URL: web.URLRegister, Value: "Register"},
		}

		a.tmpl.Execute(w, r, "register", template.ExecuteParams{
			Title:         "Register",
			PageID:        web.PageRegister,
			Authenticated: true,
			Data:          data,
		})
	}
}

// Register handles post requests to the register endpoint.
func (a *Auth) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := session.GetUserContext(r.Context())

		user.Username = r.FormValue("username")

		err := a.registry.Register(r.Context(), user)
		if err != nil {
			a.log.Printf("[ERROR] [%s %s] Registering user: %v\n", r.Method, r.URL.Path, err)
			app.WriteJSONError(w, err)
			return
		}

		a.cookies.Set(w, cookie.FlashMessage, fmt.Sprintf("Welcome, %s!", user.FirstName))
		http.Redirect(w, r, web.URLDashboard, http.StatusFound)
	}
}

func (a *Auth) Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := session.GetUserContext(r.Context())

		if err := a.registry.Logout(r.Context(), user); err != nil {
			a.log.Printf("[ERROR] [%s %s] Logging out user: %v\n", r.Method, r.URL.Path, err)
		}

		a.cookies.Clear(w, cookie.Session)
		http.Redirect(w, r, web.URLLogin, http.StatusFound)
	}
}
