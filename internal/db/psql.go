package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Postgres wraps a database connection for a PostgreSQL database.
type Postgres struct {
	conn *sql.DB
}

// Open opens this Postgres connection using the options provided.
func (p *Postgres) Open(host, port, username, password, dbname string) error {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, password, dbname))
	if err != nil {
		return err
	}

	p.conn = db

	return nil
}

// Close closes this Postgres connection. Open must be called before calling this function.
func (p *Postgres) Close() error {
	return p.conn.Close()
}

// Ping pings the Postgres database. Open must be called before calling this function.
func (p *Postgres) Ping() error {
	return p.conn.Ping()
}

// QueryRow queries a single row in the Postgres database. Open must be called before calling this function.
func (p *Postgres) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return p.conn.QueryRowContext(ctx, query, args...)
}

// Query executes a query in the Postgres database. Open must be called before calling this function.
func (p *Postgres) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return p.conn.QueryContext(ctx, query, args...)
}

// Exec executes a statement in the Postgres database. Open must be called before calling this function.
func (p *Postgres) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return p.conn.ExecContext(ctx, query, args...)
}

// Tx creates a new transaction in the Postgres database. Open must be called before calling this function.
func (p *Postgres) Tx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := p.conn.BeginTx(ctx, opts)
	return NewTx(tx), err
}
