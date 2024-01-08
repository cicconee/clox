package app

import (
	"context"
	"database/sql"
	"github.com/cicconee/clox/internal/db"
)

type Pinger interface {
	Ping() error
}

type DBOpener interface {
	Open(host, port, username, password, dbname string) error
}

type DBOpenPinger interface {
	DBOpener
	Pinger
}

type CacheOpener interface {
	Open(host, port, username, password string)
}

type CacheOpenPinger interface {
	CacheOpener
	Pinger
}

type DB interface {
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Tx(ctx context.Context, opts *sql.TxOptions) (*db.Tx, error)
}
