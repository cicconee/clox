package oauth2

import (
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	ClientID          string
	ClientSecret      string
	RedirectURLScheme string
	RedirectURLHost   string
	RedirectURLPort   string
	RedirectURLPath   string
}

func Google(c *Config) *Provider {
	redirectURL := fmt.Sprintf("%s://%s:%s/%s",
		c.RedirectURLScheme,
		c.RedirectURLHost,
		c.RedirectURLPort,
		c.RedirectURLPath)

	googleConfig := &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}

	return &Provider{Name: "google", Config: googleConfig}
}
