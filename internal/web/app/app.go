package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cicconee/clox/internal/cloudstore"
	"github.com/cicconee/clox/internal/oauth2"
	"github.com/cicconee/clox/internal/provider/google"
	"github.com/cicconee/clox/internal/server"
	"github.com/cicconee/clox/internal/token"
	"github.com/cicconee/clox/internal/user"
	"github.com/cicconee/clox/internal/web"
	"github.com/cicconee/clox/internal/web/auth"
	"github.com/cicconee/clox/internal/web/cookie"
	"github.com/cicconee/clox/internal/web/handler"
	"github.com/cicconee/clox/internal/web/middleware"
	"github.com/cicconee/clox/internal/web/session"
	"github.com/cicconee/clox/internal/web/template"
)

// App encapsulates the Clox server side application.
type App struct {
	Server       *server.HTTP
	Logger       *log.Logger
	Template     *template.Template
	GoogleOAuth2 *oauth2.Provider
	Cookies      *cookie.Manager
	Sessions     *session.Manager
	Users        *user.Service
	Tokens       *token.Service
	CloudDirs    *cloudstore.DirService

	dashboard *handler.Dashboard
	auth      *handler.Auth
	google    *handler.OAuth2
	tokens    *handler.Token

	sessionMiddleware  *middleware.Session
	flashMiddleware    *middleware.Flash
	registryMiddleware *middleware.Registry
}

// init initializes and validates App. If any required fields in App are not defined an error is returned.
func (a *App) init() error {
	if err := a.Template.Parse(); err != nil {
		return fmt.Errorf("parsing templates: %w", err)
	}

	if err := a.CloudDirs.SetupRoot(); err != nil {
		return fmt.Errorf("setting up root storage directory: %w", err)
	}

	googleAuthenticator := auth.NewAuthenticator(a.GoogleOAuth2, google.New(a.GoogleOAuth2), a.Users, a.Sessions)
	registry := auth.NewRegistry(a.Users, a.Sessions, a.CloudDirs, a.Logger)

	a.dashboard = handler.NewDashboard(a.Template, a.Logger)
	a.auth = handler.NewAuth(registry, a.Cookies, a.Template, a.Logger)
	a.google = handler.NewOAuth2(googleAuthenticator, a.Cookies, a.Logger)
	a.tokens = handler.NewToken(a.Tokens, a.Cookies, a.Template, a.Logger)

	a.sessionMiddleware = middleware.NewSession(a.Sessions, a.Cookies, a.Logger)
	a.flashMiddleware = middleware.NewFlash(a.Cookies)
	a.registryMiddleware = middleware.NewRegistry(a.Cookies, a.Logger)

	return nil
}

// setRoutes sets all the route handlers for App.
func (a *App) setRoutes() {
	a.Server.SetRoute("GET", web.URLDashboard, a.dashboard.Template(),
		a.sessionMiddleware.Active,
		a.registryMiddleware.IsRegistered,
		a.flashMiddleware.Extract)

	a.Server.SetRoute("GET", web.URLLogin, a.auth.TemplateLogin(),
		a.sessionMiddleware.Inactive,
		a.flashMiddleware.Extract)

	a.Server.SetRoute("GET", web.URLGoogleLogin, a.google.Redirect())

	a.Server.SetRoute("GET", web.URLGoogleCallback, a.google.Callback())

	a.Server.SetRoute("GET", web.URLRegister, a.auth.TemplateRegister(),
		a.sessionMiddleware.Active,
		a.registryMiddleware.NotRegistered,
		a.flashMiddleware.Extract)

	a.Server.SetRoute("GET", web.URLTokens, a.tokens.TemplateListing(),
		a.sessionMiddleware.Active,
		a.registryMiddleware.IsRegistered,
		a.flashMiddleware.Extract)

	a.Server.SetRoute("POST", web.URLRegister, a.auth.Register(),
		a.sessionMiddleware.Active,
		a.registryMiddleware.NotRegistered)

	a.Server.SetRoute("POST", web.URLLogout, a.auth.Logout(),
		a.sessionMiddleware.Active)

	a.Server.SetRoute("POST", web.URLTokens, a.tokens.Generate(),
		a.sessionMiddleware.Active,
		a.registryMiddleware.IsRegistered)

	a.Server.SetRoute("DELETE", web.URLTokenResource, a.tokens.Delete(),
		a.sessionMiddleware.Active,
		a.registryMiddleware.IsRegistered)
}

// setStaticAssets sets all the static asset handlers for App.
func (a *App) setStaticAssets() {
	fs := http.FileServer(http.Dir("web/static/"))
	fsHandlerFunc := http.HandlerFunc(http.StripPrefix("/web/static", fs).ServeHTTP)
	a.Server.SetStatic("/web/static/*", fsHandlerFunc)
}

// Start will initialize, set all the routes, and start App.
func (a *App) Start() error {
	if err := a.init(); err != nil {
		return fmt.Errorf("initializing App: %w", err)
	}

	a.setRoutes()
	a.setStaticAssets()

	return a.Server.Start()
}
