package middleware

import (
	"log"
	"net/http"

	"github.com/cicconee/clox/internal/api/auth"
	"github.com/cicconee/clox/internal/app"
)

// Token has middleware functions for handling requests that require API tokens.
type Token struct {
	auth   *auth.Authenticator
	logger *log.Logger
}

// NewToken creates a new Token middleware.
func NewToken(auth *auth.Authenticator, logger *log.Logger) *Token {
	return &Token{auth: auth, logger: logger}
}

// Validate is a http middleware that ensures the request is being made with a valid API token.
// If a token exists and is valid, the user ID of the user making the request is injected into
// the request context.
//
// Validate should wrap all handlers that require an API token.
func (a *Token) Validate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := a.auth.AuthenticateRequest(r)
		if err != nil {
			app.WriteJSONError(w, err)
			a.logger.Printf("[ERROR] [%s %s] Authenticating request: %v\n", r.Method, r.URL.Path, err)
			return
		}

		ctx := auth.SetUserIDContext(r.Context(), userID)
		next(w, r.WithContext(ctx))
	}
}
