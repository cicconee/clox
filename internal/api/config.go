package api

import (
	"fmt"
	"os"

	"github.com/cicconee/clox/internal/app"
	"github.com/cicconee/clox/pkg/env"
)

// A Config is the web application configuration for the Clox API.
type Config struct {
	*app.Config
	APIPort string
}

// LoadConfig will load the application configuration and set the remaining values based on the environment variables.
func LoadConfig(loader *env.Loader) (*Config, error) {
	appConfig, err := app.LoadConfig(loader)
	if err != nil {
		return nil, fmt.Errorf("loading app configuration: %w", err)
	}

	return &Config{
		Config:  appConfig,
		APIPort: os.Getenv("API_PORT"),
	}, nil
}
