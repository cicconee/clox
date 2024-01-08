package user

import (
	"database/sql"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/cicconee/clox/internal/app"
)

type Status string

const (
	Complete   Status = "complete"
	Incomplete Status = "incomplete"
	Blocked    Status = "blocked"
)

type User struct {
	ID                 string
	FirstName          string
	LastName           string
	PictureURL         string
	Email              string
	Username           string
	RegistrationStatus Status
}

func (u *User) Row() Row {
	row := Row{
		ID:             u.ID,
		FirstName:      sql.NullString{String: u.FirstName},
		LastName:       sql.NullString{String: u.LastName},
		PictureURL:     sql.NullString{String: u.PictureURL},
		Email:          u.Email,
		Username:       sql.NullString{String: u.Username},
		RegisterStatus: u.RegistrationStatus,
	}

	if row.FirstName.String != "" {
		row.FirstName.Valid = true
	}

	if row.LastName.String != "" {
		row.LastName.Valid = true
	}

	if row.PictureURL.String != "" {
		row.PictureURL.Valid = true
	}

	if row.Username.String != "" {
		row.Username.Valid = true
	}

	return row
}

// NormalizeUsername normalizes this User's username. Any trailing space will be removed
// and all letters will be lowercase.
//
// The normalized username will overwrite this User's Username field.
func (u *User) NormalizeUsername() {
	username := u.Username
	username = strings.ToLower(username)
	username = strings.TrimSpace(username)
	u.Username = username
}

// Validate validates the fields in this User. If any conditions fail, an error is returned.
// All errors returned by Validate are a app.WrappedSafeError.
//
// NormalizeUsername should be called before calling Validate.
func (u *User) Validate() error {
	if u.Username == "" {
		return app.Wrap(app.WrapParams{
			Err:         errors.New("username is empty"),
			SafeMessage: "Username cannot be empty.",
			StatusCode:  http.StatusBadRequest,
		})
	}

	reg := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	validUsername := reg.MatchString(u.Username)
	if !validUsername {
		return app.Wrap(app.WrapParams{
			Err:         errors.New("username contains unsupported characters"),
			SafeMessage: "Username must only contains letters and numbers.",
			StatusCode:  http.StatusBadRequest,
		})
	}

	return nil
}

// ValidRegistration returns if this user has a valid registration status.
// If the user does not have a valid status, the user should not be allowed
// to use the application.
func (u *User) ValidRegistration() bool {
	return u.RegistrationStatus == Complete || u.RegistrationStatus == Incomplete
}
