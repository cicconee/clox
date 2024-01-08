package user

import (
	"context"
	"database/sql"
	"github.com/cicconee/clox/internal/app"
)

type Repo struct {
	db app.DB
}

func NewRepo(db app.DB) *Repo {
	return &Repo{db: db}
}

type Row struct {
	ID             string
	FirstName      sql.NullString
	LastName       sql.NullString
	PictureURL     sql.NullString
	Email          string
	Username       sql.NullString
	RegisterStatus Status
}

func (r *Row) user() *User {
	return &User{
		ID:                 r.ID,
		FirstName:          r.FirstName.String,
		LastName:           r.LastName.String,
		PictureURL:         r.PictureURL.String,
		Email:              r.Email,
		Username:           r.Username.String,
		RegistrationStatus: r.RegisterStatus,
	}
}

func (r *Repo) Insert(ctx context.Context, row Row) error {
	query := `INSERT INTO users(id, first_name, last_name, picture_url, email, username, register_status) 
		 VALUES($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Exec(ctx, query,
		row.ID,
		row.FirstName,
		row.LastName,
		row.PictureURL,
		row.Email,
		row.Username,
		row.RegisterStatus)

	return err
}

func (r *Repo) Select(ctx context.Context, id string) (*Row, error) {
	query := `SELECT id, first_name, last_name, picture_url, email, username, register_status
		FROM users WHERE id = $1`

	var row Row
	err := r.db.QueryRow(ctx, query, id).Scan(
		&row.ID,
		&row.FirstName,
		&row.LastName,
		&row.PictureURL,
		&row.Email,
		&row.Username,
		&row.RegisterStatus)
	if err != nil {
		return nil, err
	}

	return &row, nil
}
