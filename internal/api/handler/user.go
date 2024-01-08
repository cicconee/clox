package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/cicconee/clox/internal/api/auth"
	"github.com/cicconee/clox/internal/app"
	"github.com/cicconee/clox/internal/user"
)

type User struct {
	users *user.Service
	log   *log.Logger
}

func NewUser(users *user.Service, log *log.Logger) *User {
	return &User{
		users: users,
		log:   log,
	}
}

// Me returns a http.HandlerFunc that writes the user information as a JSON response.
//
// The http.HandlerFunc expects a user ID in the request context.
func (u *User) Me() http.HandlerFunc {
	type response struct {
		ID         string `json:"id"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		PictureURL string `json:"picture_url"`
		Email      string `json:"email"`
		Username   string `json:"username"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserIDContext(r.Context())

		user, err := u.users.Get(r.Context(), userID)
		if err != nil {
			app.WriteJSONError(w, err)
			u.log.Printf("[ERROR] [%s %s] Getting user: %v\n", r.Method, r.URL.Path, err)
			return
		}

		resp, err := json.Marshal(&response{
			ID:         user.ID,
			FirstName:  user.FirstName,
			LastName:   user.LastName,
			PictureURL: user.PictureURL,
			Email:      user.Email,
			Username:   user.Username,
		})
		if err != nil {
			app.WriteJSONError(w, err)
			u.log.Printf("[ERROR] [%s %s] Marshalling response: %v\n", r.Method, r.URL.Path, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}
