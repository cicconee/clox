package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/cicconee/clox/internal/app"
	"github.com/cicconee/clox/internal/token"
	"github.com/cicconee/clox/internal/web"
	"github.com/cicconee/clox/internal/web/cookie"
	"github.com/cicconee/clox/internal/web/session"
	"github.com/cicconee/clox/internal/web/template"
	"github.com/go-chi/chi/v5"
)

// Token encapsulates the handlers for user tokens.
type Token struct {
	tokens  *token.Service
	cookies *cookie.Manager
	tmpl    *template.Template
	log     *log.Logger
}

// NewToken creates a token handler.
func NewToken(tokens *token.Service, cookies *cookie.Manager, tmpl *template.Template, log *log.Logger) *Token {
	return &Token{tokens: tokens, cookies: cookies, tmpl: tmpl, log: log}
}

// TemplateListing executes the tokens template which displays all the active tokens for a user.
//
// TemplateListing expects a registered session.User in the request context.
func (t *Token) TemplateListing() http.HandlerFunc {
	type data struct {
		Listings         []token.Listing
		TokenResourceURL string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		user := session.GetUserContext(r.Context())

		listings, err := t.tokens.List(r.Context(), user.UserID)
		if err != nil {
			t.log.Printf("[ERROR] [%s %s] Getting token list: %v\n", r.Method, r.URL.Path, err)
		}

		t.tmpl.Execute(w, r, "tokens", template.ExecuteParams{
			Title:         "API Tokens",
			PageID:        web.PageTokens,
			NavLinks:      web.NavBarAuthenticated,
			Authenticated: true,
			Data:          data{Listings: listings, TokenResourceURL: web.URLTokenResource},
		})
	}
}

// Generate handles creating a new token.
//
// Generate expects a registered session.User in the request context.
func (t *Token) Generate() http.HandlerFunc {
	type response struct {
		Token     string `json:"token"`
		TokenID   string `json:"token_id"`
		TokenName string `json:"token_name"`
		CreatedAt string `json:"created_at"`
		LastUsed  string `json:"last_used"`
		ExpiresAt string `json:"expires_at"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		user := session.GetUserContext(r.Context())
		tokenName := r.FormValue("tokenName")
		durationStr := r.FormValue("expiration") + "s"

		seconds, err := time.ParseDuration(durationStr)
		if err != nil {
			t.log.Printf("[ERROR] [%s %s] Parsing duration to seconds: %v\n", r.Method, r.URL.Path, err)
			app.WriteJSONError(w, app.Wrap(app.WrapParams{
				Err:         err,
				SafeMessage: "Invalid expires duration",
				StatusCode:  http.StatusBadRequest,
			}))
			return
		}

		newListing, err := t.tokens.New(r.Context(), user.UserID, seconds, tokenName)
		if err != nil {
			t.log.Printf("[ERROR] [%s %s] Creating new token: %v\n", r.Method, r.URL.Path, err)
			app.WriteJSONError(w, err)
			return
		}

		jsonResponse, err := json.Marshal(&response{
			Token:     newListing.Token,
			TokenID:   newListing.Listing.TokenID,
			TokenName: newListing.Listing.TokenName,
			CreatedAt: newListing.IssuedAtString(),
			LastUsed:  newListing.LastUsedString(),
			ExpiresAt: newListing.ExpiresAtString(),
		})
		if err != nil {
			t.log.Printf("[ERROR] [%s %s] Marshalling response: %v\n", r.Method, r.URL.Path, err)
			app.WriteJSONError(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	}
}

// Delete soft deletes a token. The token that is deleted should be specified in the request path.
//
// Delete expects a registered session.User in the request context.
func (t *Token) Delete() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := session.GetUserContext(r.Context())
		tokenID := chi.URLParam(r, "id")

		if err := t.tokens.Revoke(r.Context(), user.UserID, tokenID); err != nil {
			t.log.Printf("[ERROR] [%s %s] Deleting token: %v\n", r.Method, r.URL.Path, err)
			app.WriteJSONError(w, err)
			return
		}

		t.log.Printf("[INFO] [%s %s] [User: %s] Token deleted: %s\n", r.Method, r.URL.Path, user.UserID, tokenID)
		w.WriteHeader(http.StatusOK)
	}
}
