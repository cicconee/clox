package cloudstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/cicconee/clox/internal/app"
)

// PathMapper maps user-facing paths to the path structure on the server.
type PathMapper struct {
	// The path to the root storage. All paths will begin here.
	root string
}

// NewPathMapper creates a new PathMapper.
func NewPathMapper(root string) *PathMapper {
	return &PathMapper{root: root}
}

// DirSearch are the parameters for finding a directory ID based on the
// Path, that belongs to a specific user (UserID), under their root
// directory (RootID).
type DirSearch struct {
	UserID string
	RootID string
	Path   string
}

// FindDir parses a path to a directory and searches for the ID. The directory
// names in the path are mapped to the ID's on the server. If found, the ID
// is returned.
//
// The directory must belong to the user (userID) and live under their
// root directory (rootID).
func (pm *PathMapper) FindDir(ctx context.Context, q *Query, d DirSearch) (string, error) {
	fp := filepath.Clean(d.Path)
	var p string
	if fp == "." || fp == "/" {
		p = "root"
	} else if fp[0] == '/' {
		p = "root" + fp
	} else {
		p = "root/" + fp
	}
	paths := strings.Split(p, "/")

	directoryID := d.RootID
	for i := 1; i < len(paths); i++ {
		pName := paths[i]
		dir, err := q.SelectDirectoryByUserNameParent(ctx, d.UserID, pName, directoryID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", app.Wrap(app.WrapParams{
					Err:         fmt.Errorf("directory %q does not exist [path: %s]", pName, d.Path),
					SafeMessage: fmt.Sprintf("Directory %q does not exist", strings.Join(paths[1:i+1], "/")),
					StatusCode:  http.StatusBadRequest,
				})
			}

			return "", err
		}

		directoryID = dir.ID
	}

	return directoryID, nil
}
