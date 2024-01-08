package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/cicconee/clox/internal/app"
	"github.com/cicconee/clox/internal/token"
	"github.com/cicconee/clox/internal/user"
	"golang.org/x/net/context"
)

// Authenticator authenticates API tokens for the Clox API.
type Authenticator struct {
	tokens *token.Service
	users  *user.Service
}

// NewAuthenticator creates a new Authenticator.
func NewAuthenticator(tokens *token.Service, users *user.Service) *Authenticator {
	return &Authenticator{tokens: tokens, users: users}
}

// Authenticate validates token.
func (a *Authenticator) Authenticate(ctx context.Context, token string) (string, error) {
	return a.authenticate(ctx, token)
}

// AuthenticateRequest extracts a Bearer token from the http.Request Authorization header
// and then validates the token.
func (a *Authenticator) AuthenticateRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" {
		return "", app.Wrap(app.WrapParams{
			Err:         errors.New("empty api token"),
			SafeMessage: "No Authorization header provided",
			StatusCode:  http.StatusUnauthorized,
		})
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", app.Wrap(app.WrapParams{
			Err:         errors.New("invalid api token format"),
			SafeMessage: "Invalid Authorization header format",
			StatusCode:  http.StatusUnauthorized,
		})
	}

	return a.authenticate(r.Context(), strings.TrimPrefix(authHeader, "Bearer "))
}

// authenticate validates token and ensures user is not blocked.
func (a *Authenticator) authenticate(ctx context.Context, token string) (string, error) {
	uid, err := a.tokens.Validate(ctx, token)
	if err != nil {
		return "", fmt.Errorf("validating token: %w", err)
	}

	u, err := a.users.Get(ctx, uid)
	if err != nil {
		return "", fmt.Errorf("getting user: %w", err)
	}

	if !u.ValidRegistration() {
		if u.RegistrationStatus == user.Blocked {
			return "", app.Wrap(app.WrapParams{
				Err:         errors.New("blocked user"),
				SafeMessage: "Your account is blocked. Please contact us.",
				StatusCode:  http.StatusUnauthorized,
			})
		}

		return "", app.Wrap(app.WrapParams{
			Err:         fmt.Errorf("unsupported registraton status: %v", u.RegistrationStatus),
			SafeMessage: "Something is wrong with you account. Please contact us.",
			StatusCode:  http.StatusUnauthorized,
		})
	}

	return uid, nil
}
