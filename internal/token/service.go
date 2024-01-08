package token

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cicconee/clox/internal/app"
	"github.com/cicconee/clox/internal/cache"
	"github.com/cicconee/clox/internal/jwt"
	"github.com/cicconee/clox/pkg/random"
)

var ErrTokenName = errors.New("invalid token name")

// Service controls token creation, revocation, and listings.
type Service struct {
	// jwts creates and validates JWT's.
	jwts *jwt.Manager

	// cache caches all the revoked tokens.
	cache *cache.Redis

	// repo executes database queries for tokens.
	repo *Repo
}

// NewService creates a new Service.
func NewService(jwts *jwt.Manager, cache *cache.Redis, repo *Repo) *Service {
	return &Service{jwts: jwts, cache: cache, repo: repo}
}

// New creates a new token and writes it to the database. The token and its relevant data is
// returned as a NewListing.
//
// The token subject is the user ID (uid) and will be valid for the set duration (dur). A random token ID
// (jti) will be generated for the token. All time claims (exp, nbf, iat) and times related to the token
// are UTC times.
//
// The token name is used to identify the token to the user. It is not part of the token.
func (s *Service) New(ctx context.Context, uid string, dur time.Duration, name string) (NewListing, error) {
	// TODO: Trim token space, maybe restrict special characters.

	if name == "" {
		return NewListing{}, app.Wrap(app.WrapParams{
			Err:         fmt.Errorf("%w: token name is empty", ErrTokenName),
			SafeMessage: "Token name cannot be empty.",
			StatusCode:  http.StatusBadRequest,
		})
	}

	now := time.Now().UTC()
	exp := now.Add(dur)
	jti := random.ID(32)

	token, err := s.jwts.New(jwt.NewTokenClaims{
		Sub: uid,
		Exp: exp,
		Nbf: now,
		Iat: now,
		Jti: jti,
	})
	if err != nil {
		return NewListing{}, err
	}

	row := Row{
		ID:        jti,
		Name:      name,
		ExpiresAt: exp,
		IssuedAt:  now,
		LastUsed:  sql.NullTime{Valid: false},
		UserID:    uid,
	}
	if err = s.repo.Insert(ctx, row); err != nil {
		return NewListing{}, fmt.Errorf("inserting token: %w", err)
	}

	return NewListing{
		Token:   token,
		Listing: row.listing(),
	}, nil
}

// List gets all the token listings for a user.
func (s *Service) List(ctx context.Context, uid string) ([]Listing, error) {
	rows, err := s.repo.SelectAll(ctx, uid)
	if err != nil {
		return nil, err
	}

	return rows.listings(), nil
}

// Revoke marks a token as deleted. Tokens that are revoked remain in the database. Only the user
// that created the token may revoke it.
//
// Any JWT that a user revokes will result in in it being invalid. Revoked JWT's will be need to
// be stored on the server for the remainder of its lifespan. Only once a revoked JWT is expired is
// it safe to be deleted.
func (s *Service) Revoke(ctx context.Context, uid string, jti string) error {
	row, err := s.repo.Select(ctx, jti)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return app.Wrap(app.WrapParams{
				Err:         fmt.Errorf("token not found [jti: %s]: %w", jti, err),
				SafeMessage: "Token not found.",
				StatusCode:  http.StatusNotFound,
			})
		}

		return err
	}

	// Token does not belong to user making the request.
	if row.UserID != uid {
		return app.Wrap(app.WrapParams{
			Err:         errors.New("user not authorized"),
			SafeMessage: "You are not the token owner.",
			StatusCode:  http.StatusUnauthorized,
		})
	}

	return s.repo.UpdateDeletedAt(ctx, jti, time.Now().UTC())
}

// Validate validates a JWT and then checks if the token has been revoked. If the JWT is valid
// it will return the user id (sub).
func (s *Service) Validate(ctx context.Context, token string) (string, error) {
	claims, err := s.jwts.Validate(token)
	if err != nil {
		return "", app.Wrap(app.WrapParams{
			Err:         err,
			SafeMessage: "Invalid token",
			StatusCode:  http.StatusUnauthorized,
		})
	}

	row, err := s.repo.Select(ctx, claims.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", app.Wrap(app.WrapParams{
				Err:         fmt.Errorf("token not found [jti: %s]: %w", claims.ID, err),
				SafeMessage: "Token not found",
				StatusCode:  http.StatusNotFound,
			})
		}

		return "", err
	}

	// If token row has a DeletedAt time, it was revoked and is invalid.
	if !row.DeletedAt.Time.IsZero() {
		return "", app.Wrap(app.WrapParams{
			Err:         errors.New("token is revoked"),
			SafeMessage: "Invalid token",
			StatusCode:  http.StatusUnauthorized,
		})
	}

	return claims.Subject, nil
}

// TODO: Implement method to update a tokens last used time.
