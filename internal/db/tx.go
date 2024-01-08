package db

import (
	"context"
	"database/sql"
)

type Tx struct {
	conn *sql.Tx
}

func NewTx(tx *sql.Tx) *Tx {
	return &Tx{tx}
}

// QueryRow queries a single row in this Tx.
func (tx *Tx) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return tx.conn.QueryRowContext(ctx, query, args...)
}

// Query executes a query in this Tx.
func (tx *Tx) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return tx.conn.QueryContext(ctx, query, args...)
}

// Exec executes a statement in this Tx.
func (tx *Tx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return tx.conn.ExecContext(ctx, query, args...)
}

// Commit commits this Tx to the database.
func (tx *Tx) Commit() error {
	return tx.conn.Commit()
}

// Rollback rolls back this Tx.
func (tx *Tx) Rollback() error {
	return tx.conn.Rollback()
}
