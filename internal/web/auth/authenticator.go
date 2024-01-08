package auth

import (
	"fmt"
	"net/http"

	"github.com/cicconee/clox/internal/oauth2"
	"github.com/cicconee/clox/internal/user"
	"github.com/cicconee/clox/internal/web/session"
	"github.com/cicconee/clox/pkg/random"
)

type Generater interface {
	Generate() oauth2.Redirect
}

type Validater interface {
	Validate(r *http.Request, state string) (*oauth2.Token, error)
}

type GenerateValidater interface {
	Generater
	Validater
}

type Authenticator struct {
	oauth2   GenerateValidater
	provider user.Provider
	users    *user.Service
	sessions *session.Manager
}

func NewAuthenticator(oauth2 GenerateValidater, provider user.Provider, users *user.Service, sessions *session.Manager) *Authenticator {
	return &Authenticator{
		oauth2:   oauth2,
		provider: provider,
		users:    users,
		sessions: sessions,
	}
}

func (a *Authenticator) Generate() (url string, state string) {
	redirect := a.oauth2.Generate()
	return redirect.URL, redirect.State
}

// Authenticate authenticates a user against the OAuth2 provider. The state must match the
// state in the request URL query parameters. To get a valid state, call this Authenticator's
// Generate method. This method should be called upon redirection from the URL that is
// returned from the Generate method.
func (a *Authenticator) Authenticate(r *http.Request, state string) (*session.User, error) {
	token, err := a.oauth2.Validate(r, state)
	if err != nil {
		return nil, fmt.Errorf("getting user token: %w", err)
	}

	user, err := a.users.Authenticate(r.Context(), a.provider, token)
	if err != nil {
		return nil, fmt.Errorf("authenticating user: %w", err)
	}

	userSession := session.User{
		SessionID:          random.ID(32),
		UserID:             user.ID,
		FirstName:          user.FirstName,
		LastName:           user.LastName,
		PictureURL:         user.PictureURL,
		Email:              user.Email,
		Username:           user.Username,
		RegistrationStatus: user.RegistrationStatus,
	}
	// If a user has an invalid registration status (blocked or unexpected value), do not
	// set the session in session storage. If the session were to be set in storage, it
	// would cause a redirect loop. Since the session is not being persisted, upon redirect
	// to a endpoint that requires an inactive session, the session middleware will verify
	// that the session is inactive (session key not in storage), and then clear the session
	// key from cookies.
	if !user.ValidRegistration() {
		return &userSession, nil
	}

	err = a.sessions.Set(r.Context(), userSession)
	if err != nil {
		return nil, fmt.Errorf("setting user session: %w", err)
	}

	return &userSession, nil
}
