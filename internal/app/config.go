package app

import (
	"fmt"
	"os"

	"github.com/cicconee/clox/pkg/env"
)

// A Config is the application configuration for Clox. This configuration is considered the base configuration, and it
// will be used by both the Server Side App and the API.
type Config struct {
	appEnv               string
	Host                 string
	Port                 string
	postgresHost         string
	postgresPort         string
	postgresUsername     string
	postgresPassword     string
	postgresDatabaseName string
	redisHost            string
	redisPort            string
	redisUsername        string
	redisPassword        string
	jwtSecretKey         string
	FileStorePath        string
}

// LoadConfig will load the environment variables and create the Config based on these values.
func LoadConfig(loader *env.Loader) (*Config, error) {
	if err := loader.Load(); err != nil {
		return nil, fmt.Errorf("loading environment variables: %w", err)
	}

	config := &Config{
		appEnv:               os.Getenv("APP_ENV"),
		Host:                 os.Getenv("HOST"),
		Port:                 os.Getenv("PORT"),
		postgresHost:         os.Getenv("POSTGRES_HOST"),
		postgresPort:         os.Getenv("POSTGRES_PORT"),
		postgresUsername:     os.Getenv("POSTGRES_USERNAME"),
		postgresPassword:     os.Getenv("POSTGRES_PASSWORD"),
		postgresDatabaseName: os.Getenv("POSTGRES_DBNAME"),
		redisHost:            os.Getenv("REDIS_HOST"),
		redisPort:            os.Getenv("REDIS_PORT"),
		redisUsername:        os.Getenv("REDIS_USERNAME"),
		redisPassword:        os.Getenv("REDIS_PASSWORD"),
		jwtSecretKey:         os.Getenv("JWT_SECRET_KEY"),
		FileStorePath:        os.Getenv("FILE_STORE_PATH"),
	}

	return config, nil
}

// OpenDB will pass the database credentials to a DBOpener to open a database connection. It will ping the database
// to ensure a connection was made.
func (c *Config) OpenDB(opener DBOpenPinger) error {
	err := opener.Open(
		c.postgresHost,
		c.postgresPort,
		c.postgresUsername,
		c.postgresPassword,
		c.postgresDatabaseName)
	if err != nil {
		return err
	}

	if err := opener.Ping(); err != nil {
		return fmt.Errorf("pinging database: %w", err)
	}

	return nil
}

// OpenCache opens the cache. It will ping the cache to ensure a connection was made.
func (c *Config) OpenCache(cache CacheOpenPinger) error {
	cache.Open(
		c.redisHost,
		c.redisPort,
		c.redisUsername,
		c.redisPassword)

	if err := cache.Ping(); err != nil {
		return fmt.Errorf("pinging cache: %w", err)
	}

	return nil
}

// AppEnv returns the application environment.
func (c *Config) AppEnv() string {
	return c.appEnv
}

// SecretSetter is the interface that wraps the SetSecret function.
type SecretSetter interface {
	// SetSecret will set the secret key used to sign JWT's.
	SetSecret(secret string)
}

// SetJWTSecret will set the secret key for SecretSetter to this Config's jwtSecretKey value.
func (c *Config) SetJWTSecret(jwt SecretSetter) {
	jwt.SetSecret(c.jwtSecretKey)
}
