package oauth2

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"math/rand"
	"net/http"
)

type Provider struct {
	Name   string
	Config *oauth2.Config
}

type Redirect struct {
	State string
	URL   string
}

func (p *Provider) Generate() Redirect {
	state := randomState()
	url := p.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	return Redirect{State: state, URL: url}
}

func (p *Provider) Exchange(ctx context.Context, code string) (*Token, error) {
	oauth2Token, err := p.Config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	return &Token{oauth2Token: oauth2Token, provider: p.Name}, nil
}

func (p *Provider) Validate(r *http.Request, state string) (*Token, error) {
	providerState := r.URL.Query().Get("state")
	if state != providerState {
		return nil, fmt.Errorf("state mismatch [providerState: %s, state: %s]", providerState, state)
	}

	code := r.URL.Query().Get("code")
	token, err := p.Exchange(r.Context(), code)
	if err != nil {
		return nil, fmt.Errorf("exchanging code: %w", err)
	}

	return token, nil
}

// Client returns an HTTP client using the provided token. The token will auto-refresh as necessary.
//
// Due to the token auto-refreshing, if the token is being persisted, it is important to check if the token has
// refreshed before discarding the HTTP client.
func (p *Provider) Client(ctx context.Context, token *Token) *http.Client {
	return p.Config.Client(ctx, token.oauth2Token)
}

func randomState() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@$"
	state := make([]byte, 64)
	for i := range state {
		state[i] = charset[rand.Intn(len(charset))]
	}

	return string(state)
}
