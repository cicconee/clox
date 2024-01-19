package main

import (
	"fmt"
	"log"
	"os"

	"github.com/cicconee/clox/internal/cache"
	"github.com/cicconee/clox/internal/cloudstore"
	"github.com/cicconee/clox/internal/db"
	"github.com/cicconee/clox/internal/jwt"
	"github.com/cicconee/clox/internal/oauth2"
	"github.com/cicconee/clox/internal/router"
	"github.com/cicconee/clox/internal/server"
	"github.com/cicconee/clox/internal/token"
	"github.com/cicconee/clox/internal/user"
	"github.com/cicconee/clox/internal/web"
	"github.com/cicconee/clox/internal/web/app"
	"github.com/cicconee/clox/internal/web/cookie"
	"github.com/cicconee/clox/internal/web/session"
	"github.com/cicconee/clox/internal/web/template"
	"github.com/cicconee/clox/pkg/env"

	_ "github.com/lib/pq"
)

var envFile = ".env"

func main() {
	logger := log.Default()

	if err := Run(logger); err != nil {
		logger.Printf("[ERROR] Running application: %v", err)
		os.Exit(1)
	}
}

func Run(logger *log.Logger) error {
	loader, err := env.NewFileLoader(envFile)
	if err != nil {
		return fmt.Errorf("creating file loader for file %s: %w", envFile, err)
	}

	config, err := web.LoadConfig(loader)
	if err != nil {
		return fmt.Errorf("loading web configuration: %w", err)
	}

	database := &db.Postgres{}
	if err := config.OpenDB(database); err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer CloseDB(logger, database)

	cache := &cache.Redis{}
	if err := config.OpenCache(cache); err != nil {
		return fmt.Errorf("opening cache: %w", err)
	}
	defer CloseCache(logger, cache)

	googleOAuth2 := oauth2.Google(&oauth2.Config{
		ClientID:          config.GoogleOAuthClientID,
		ClientSecret:      config.GoogleOAuthClientSecret,
		RedirectURLScheme: config.OAuthCallbackScheme(),
		RedirectURLHost:   config.Host,
		RedirectURLPort:   config.Port,
		RedirectURLPath:   "login/google/callback",
	})

	jwts := jwt.NewManager("clox-server-side-app", "clox-api")
	config.SetJWTSecret(jwts)

	webApp := &app.App{
		Server:       server.New(config.Host, config.Port, router.NewChi()),
		Logger:       logger,
		Template:     template.New("clox", "web/templates", logger),
		GoogleOAuth2: googleOAuth2,
		Cookies:      cookie.NewManager(config.SecureCookie(), config.Host),
		Sessions:     session.NewManager(cache),
		Users:        user.NewService(user.NewRepo(database)),
		Tokens:       token.NewService(jwts, cache, token.NewRepo(database)),
		CloudDirs: cloudstore.NewDirService(cloudstore.DirServiceConfig{
			Path:    config.FileStorePath,
			Store:   cloudstore.NewStore(database),
			IO:      cloudstore.NewIO(&cloudstore.OSFileSystem{}),
			Log:     logger,
			PathMap: cloudstore.NewPathMapper(),
		}),
	}

	return webApp.Start()
}

func CloseDB(logger *log.Logger, database *db.Postgres) {
	if err := database.Close(); err != nil {
		logger.Printf("[ERROR] Closing the database connection: %v\n", err)
	}
}

func CloseCache(logger *log.Logger, cache *cache.Redis) {
	if err := cache.Close(); err != nil {
		logger.Printf("[ERROR] Closing the cache connection: %v\n", err)
	}
}
