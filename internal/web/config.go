package web

import (
	"fmt"
	"github.com/cicconee/clox/internal/app"
	"github.com/cicconee/clox/pkg/env"
	"os"
)

// A Config is the web application configuration for the Clox server side app.
type Config struct {
	*app.Config
	GoogleOAuthClientID     string
	GoogleOAuthClientSecret string
}

// LoadConfig will load the application configuration and set the remaining values based on the environment variables.
func LoadConfig(loader *env.Loader) (*Config, error) {
	appConfig, err := app.LoadConfig(loader)
	if err != nil {
		return nil, fmt.Errorf("loading app configuration: %w", err)
	}

	return &Config{
		Config:                  appConfig,
		GoogleOAuthClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		GoogleOAuthClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
	}, nil
}

// OAuthCallbackScheme will return the scheme used by the OAuth2 provider callback functions.
//
// https if running in prod mode. http if running in dev mode.
func (c *Config) OAuthCallbackScheme() string {
	if c.Config.AppEnv() == "prod" {
		return "https"
	}

	return "http"
}

// SecureCookie will return if cookies should be marked as secure.
//
// True if running in prod mode. False if running in dev mode.
func (c *Config) SecureCookie() bool {
	return c.Config.AppEnv() == "prod"
}
