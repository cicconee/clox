package cloudstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

var (
	ErrForeignKeyParentID = fmt.Errorf("%q foreign key violation", "parent_id")
	ErrUniqueNameParentID = fmt.Errorf("%q [columns: %q, %q] unique constraint violation", "unique_directory_name_parent", "name", "parent_id")
	ErrUniqueUserRoot     = fmt.Errorf("%q [columns: %q, %q] unique constraint violation", "unique_user_root_directory", "user_id", "name")
)

type DBTX interface {
	QueryRow(ctx context.Context, query string, args ...any) *sql.Row
	Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Query holds the individual queries for persisting directories.
type Query struct {
	// The connection to the database. This can be a database
	// connection or a database transaction.
	db DBTX
}

// NewQuery creates a new Query with the DBTX.
func NewQuery(db DBTX) *Query {
	return &Query{db: db}
}

type InsertDirectoryConfig struct {
	ID        string
	UserID    string
	Name      string
	ParentID  sql.NullString
	CreatedAt time.Time
}

// InsertDirectory inserts a directory into the directories table.
func (q *Query) InsertDirectory(ctx context.Context, c InsertDirectoryConfig) error {
	query := `INSERT INTO directories (id, user_id, name, parent_id, created_at)
			  VALUES($1, $2, $3, $4, $5)`

	_, err := q.db.Exec(ctx, query,
		c.ID,
		c.UserID,
		c.Name,
		c.ParentID,
		c.CreatedAt.UTC(),
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			// Foreign key constraint violation on the parent_id column.
			if pqErr.Code == "23503" && pqErr.Constraint == "directories_parent_id_fkey" {
				return fmt.Errorf("%w: %v", ErrForeignKeyParentID, err)
			}

			// Unique constraint violation on the parent_id and name. Another
			// directory that is a child of parent_id is using this name.
			if pqErr.Code == "23505" && pqErr.Constraint == "unique_directory_name_parent" {
				return fmt.Errorf("%w: %v", ErrUniqueNameParentID, err)
			}

			// Unique constraint violation on the user_id and name where parent_id
			// is null. Another directory was already created for the users root storage.
			if pqErr.Code == "23505" && pqErr.Constraint == "unique_user_root_directory" {
				return fmt.Errorf("%w: %v", ErrUniqueUserRoot, err)
			}
		}

		return err
	}

	return nil
}

// InsertSelfPath inserts the 0th path into the paths table. This row holds
// a path to itself with depth being 0.
func (q *Query) InsertSelfPath(ctx context.Context, directoryID string) error {
	query := `INSERT INTO paths (parent_id, child_id, depth) 
			  VALUES ($1, $1, $2)`

	_, err := q.db.Exec(ctx, query, directoryID, 0)

	return err
}

type InsertParentPathsConfig struct {
	ParentID string
	ChildID  string
}

// InsertParentPaths inserts all the paths between the child directory and its
// ancestors. It uses all the paths to the childs parent to find the paths and
// then adds 1 to its depth.
func (q *Query) InsertParentPaths(ctx context.Context, c InsertParentPathsConfig) error {
	query := `INSERT INTO paths (parent_id, child_id, depth) 
			  SELECT parent_id, $1, depth + 1 
			  FROM paths 
			  WHERE child_id = $2`

	_, err := q.db.Exec(ctx, query,
		c.ChildID,
		c.ParentID,
	)

	return err
}

func (q *Query) SelectDirectoryFSPath(ctx context.Context, directoryID string) ([]string, error) {
	query := `SELECT d.id
		  	  FROM paths p
		  	  JOIN directories d ON p.parent_id = d.id
		  	  WHERE p.child_id = $1
		  	  ORDER BY p.depth DESC;`

	rows, err := q.db.Query(ctx, query, directoryID)
	if err != nil {
		return nil, err
	}

	var idPath []string
	for rows.Next() {
		var id string

		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		idPath = append(idPath, id)
	}

	return idPath, nil
}

func (q *Query) SelectDirectoryPath(ctx context.Context, directoryID string) ([]string, error) {
	query := `SELECT d.name
		  	  FROM paths p
		  	  JOIN directories d ON p.parent_id = d.id
		  	  WHERE p.child_id = $1
		  	  ORDER BY p.depth DESC;`

	rows, err := q.db.Query(ctx, query, directoryID)
	if err != nil {
		return nil, err
	}

	var path []string
	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			return nil, err
		}

		path = append(path, name)
	}

	return path, nil
}

type DirectoryRow struct {
	ID        string
	UserID    string
	Name      string
	ParentID  sql.NullString
	CreatedAt time.Time
	UpdatedAt sql.NullTime
	LastWrite sql.NullTime
}

func (q *Query) SelectUserRootDirectory(ctx context.Context, userID string) (DirectoryRow, error) {
	query := `SELECT id, user_id, name, parent_id, created_at, updated_at, last_write
			  FROM directories
			  WHERE parent_id IS NULL
			  AND name = 'root'
			  AND user_id = $1`

	var r DirectoryRow
	err := q.db.QueryRow(ctx, query, userID).Scan(
		&r.ID,
		&r.UserID,
		&r.Name,
		&r.ParentID,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.LastWrite,
	)
	if err != nil {
		return DirectoryRow{}, err
	}

	return r, nil
}

// SelectDirectoryByUserNameParent selects a row from the directories table by user_id,
// name, and parent_id.
func (q *Query) SelectDirectoryByUserNameParent(ctx context.Context, userID string, name string, parentID string) (DirectoryRow, error) {
	query := `SELECT id, user_id, name, parent_id, created_at, updated_at, last_write
			  FROM directories
			  WHERE user_id = $1
			  AND name = $2
			  AND parent_id = $3`

	var r DirectoryRow
	err := q.db.QueryRow(ctx, query, userID, name, parentID).Scan(
		&r.ID,
		&r.UserID,
		&r.Name,
		&r.ParentID,
		&r.CreatedAt,
		&r.UpdatedAt,
		&r.LastWrite,
	)
	if err != nil {
		return DirectoryRow{}, err
	}

	return r, nil
}
