package cloudstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/cicconee/clox/internal/app"
	"github.com/cicconee/clox/internal/db"
)

var ErrCommitTx = errors.New("failed to commit transaction")

type Store struct {
	db app.DB
}

func NewStore(db app.DB) *Store {
	return &Store{db: db}
}

// Tx begins a transaction and passes the transaction to txFunc. After txFunc
// executes, it will be committed.
//
// If an error occurs, the transaction will be rolled back.
func (s *Store) Tx(ctx context.Context, txFunc func(tx *db.Tx) error) error {
	tx, err := s.db.Tx(ctx, nil)
	if err != nil {
		return err
	}

	err = txFunc(tx)
	if err != nil {
		return s.rollback(tx, err)
	}

	err = tx.Commit()
	if err != nil {
		return s.rollback(tx, fmt.Errorf("%w: %v", ErrCommitTx, err))
	}

	return nil
}

// rollback rolls back the tx. If an error occurs, the original error, err, will
// be wrapped in a error with the error that occured when rolling back.
func (s *Store) rollback(tx *db.Tx, err error) error {
	rbErr := tx.Rollback()
	if rbErr != nil {
		return fmt.Errorf("error: %w, rollback error: %v", err, rbErr)
	}

	return err
}
