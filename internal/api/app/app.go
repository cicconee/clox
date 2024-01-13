package app

import (
	"fmt"
	"log"

	"github.com/cicconee/clox/internal/api/auth"
	"github.com/cicconee/clox/internal/api/handler"
	"github.com/cicconee/clox/internal/api/middleware"
	"github.com/cicconee/clox/internal/cloudstore"
	"github.com/cicconee/clox/internal/server"
	"github.com/cicconee/clox/internal/token"
	"github.com/cicconee/clox/internal/user"
)

// App encapsulates the Clox API.
type App struct {
	Server *server.HTTP
	Logger *log.Logger
	Users  *user.Service
	Tokens *token.Service
	Cloud  *cloudstore.Service

	users *handler.User
	cloud *handler.Cloud

	tokenMiddleware *middleware.Token
}

// init initializes and validates App. If any required fields in App are not defined an error is returned.
func (a *App) init() error {
	if err := a.Cloud.SetupRoot(); err != nil {
		return fmt.Errorf("setting up root storage directory: %w", err)
	}

	authenticator := auth.NewAuthenticator(a.Tokens, a.Users)

	a.users = handler.NewUser(a.Users, a.Logger)
	a.cloud = handler.NewCloud(a.Cloud, a.Logger)

	a.tokenMiddleware = middleware.NewToken(authenticator, a.Logger)

	return nil
}

// setRoutes sets all the route handlers for App.
func (a *App) setRoutes() {
	a.Server.SetRoute("GET", "/me", a.users.Me(), a.tokenMiddleware.Validate)
	a.Server.SetRoute("POST", "/api/dir/{id}", a.cloud.NewDir(), a.tokenMiddleware.Validate)
	a.Server.SetRoute("POST", "/api/dir", a.cloud.NewDirPath(), a.tokenMiddleware.Validate)
}

// Start will initialize, set all the routes, and start App.
func (a *App) Start() error {
	if err := a.init(); err != nil {
		return fmt.Errorf("initializing App: %w", err)
	}

	a.setRoutes()

	return a.Server.Start()
}
