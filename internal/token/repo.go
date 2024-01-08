package token

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/cicconee/clox/internal/app"
)

// Repo is the token repository.
type Repo struct {
	// The database connection.
	db app.DB
}

// NewRepo creates a new Repo.
func NewRepo(db app.DB) *Repo {
	return &Repo{db: db}
}

// Row represents a token row in the database.
type Row struct {
	ID        string
	Name      string
	ExpiresAt time.Time
	IssuedAt  time.Time
	LastUsed  sql.NullTime
	UserID    string
	DeletedAt sql.NullTime
}

// listing returns this row as a Listing.
func (r *Row) listing() Listing {
	return Listing{
		TokenID:   r.ID,
		TokenName: r.Name,
		ExpiresAt: r.ExpiresAt,
		IssuedAt:  r.IssuedAt,
		LastUsed:  r.LastUsed.Time,
	}
}

// Rows is a Row slice.
type Rows []Row

// listings returns this Rows as a Listing slice.
func (r Rows) listings() []Listing {
	var listings []Listing

	for _, v := range r {
		listing := v.listing()
		listings = append(listings, listing)
	}

	return listings
}

// Insert inserts a new row into the database.
func (r *Repo) Insert(ctx context.Context, row Row) error {
	query := `INSERT INTO user_tokens(token_id, token_name, expires_at, issued_at, last_used, user_id)
		VALUES($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(ctx, query,
		row.ID,
		row.Name,
		row.ExpiresAt,
		row.IssuedAt,
		row.LastUsed,
		row.UserID)

	return err
}

// SelectAll reads all the tokens from the database that have not been deleted for a specific user id.
func (r *Repo) SelectAll(ctx context.Context, userID string) (Rows, error) {
	query := `SELECT token_id, token_name, expires_at, issued_at, last_used, user_id FROM user_tokens
		WHERE user_id = $1 AND deleted_at IS NULL`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokenRows Rows
	for rows.Next() {
		var row Row

		err := rows.Scan(
			&row.ID,
			&row.Name,
			&row.ExpiresAt,
			&row.IssuedAt,
			&row.LastUsed,
			&row.UserID)
		if err != nil {
			return nil, err
		}

		tokenRows = append(tokenRows, row)
	}

	return tokenRows, nil
}

// Select reads a single token row from the database.
func (r *Repo) Select(ctx context.Context, id string) (Row, error) {
	query := `SELECT token_id, token_name, expires_at, issued_at, last_used, user_id, deleted_at FROM user_tokens
		WHERE token_id = $1`

	var row Row
	err := r.db.QueryRow(ctx, query, id).Scan(
		&row.ID,
		&row.Name,
		&row.ExpiresAt,
		&row.IssuedAt,
		&row.LastUsed,
		&row.UserID,
		&row.DeletedAt,
	)

	return row, err
}

// UpdateDeletedAt sets the deleted_at column with value t where token_id is id.
func (r *Repo) UpdateDeletedAt(ctx context.Context, id string, t time.Time) error {
	// query := `UPDATE user_tokens SET deleted_at = $1 WHERE token_id = $2`
	// _, err := r.db.Exec(ctx, query, t, id)
	// return err
	return r.Update(ctx, id, map[string]interface{}{
		"deleted_at": t,
	})
}

// Update updates a user_tokens row using the updates map where token_id is id.
//
// The updates map key value corresponds to the column names, and the values will be the
// updated values.
func (r *Repo) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	var columns []string
	var args []interface{}
	i := 1

	for col, val := range updates {
		columns = append(columns, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	args = append(args, id)

	query := fmt.Sprintf("UPDATE user_tokens SET %s WHERE token_id = $%d", strings.Join(columns, ","), i)
	_, err := r.db.Exec(ctx, query, args...)
	return err
}
