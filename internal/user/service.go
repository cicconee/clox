package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/cicconee/clox/internal/app"
	"github.com/cicconee/clox/internal/oauth2"
	"github.com/cicconee/clox/internal/provider"
)

var ErrUserNotFound = errors.New("user not found")

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

type Provider interface {
	UserInfo(context.Context, *oauth2.Token) (provider.User, error)
}

func (s *Service) Authenticate(ctx context.Context, provider Provider, token *oauth2.Token) (*User, error) {
	info, err := provider.UserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("fetching user info from %s: %w", token.Provider(), err)
	}

	row, err := s.get(ctx, info.ID.Encode())
	if err != nil {
		return nil, err
	}
	// User already exists.
	if row != nil {
		return row.user(), nil
	}

	return &User{
		ID:                 info.ID.Encode(),
		FirstName:          info.FirstName,
		LastName:           info.LastName,
		PictureURL:         info.PictureURL,
		Email:              info.Email,
		Username:           "",
		RegistrationStatus: Incomplete,
	}, nil
}

// Get gets a user from the database by id. If a user is not found, a ErrUserNotFound is returned
// within a app.WrappedSafeError.
func (s *Service) Get(ctx context.Context, id string) (*User, error) {
	row, err := s.get(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, app.Wrap(app.WrapParams{
			Err:         fmt.Errorf("selecting user row [id: %s]: %w", id, ErrUserNotFound),
			SafeMessage: "User does not exist",
			StatusCode:  http.StatusNotFound,
		})
	}

	return row.user(), nil
}

func (s *Service) get(ctx context.Context, id string) (*Row, error) {
	row, err := s.repo.Select(ctx, id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("selecting user row [id: %s]: %w", id, err)
	}

	return row, nil
}

type Registration struct {
	ID         string
	FirstName  string
	LastName   string
	PictureURL string
	Email      string
	Username   string
}

// Register will register a user by writing a user to the database.
//
// The Username field value passed in with Registration may be different from Username
// field returned in User. Usernames get normalized before being written to the database
// so this value may change. If you need the up-to-date username, use the username that
// is returned with User.
func (s *Service) Register(ctx context.Context, r Registration) (*User, error) {
	user := User{
		ID:                 r.ID,
		FirstName:          r.FirstName,
		LastName:           r.LastName,
		PictureURL:         r.PictureURL,
		Email:              r.Email,
		Username:           r.Username,
		RegistrationStatus: Complete,
	}

	user.NormalizeUsername()

	if err := user.Validate(); err != nil {
		return nil, err
	}

	err := s.repo.Insert(ctx, user.Row())
	if err != nil {
		return nil, fmt.Errorf("inserting user [id: %s]: %w", user.ID, err)
	}

	return &user, nil
}
