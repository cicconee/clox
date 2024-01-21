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

// Root returns the root storage path. All paths will be children of
// this path.
func (pm *PathMapper) Root() string {
	return pm.root
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

// FindFile parses a path to a file and returns its ID.
//
// All directories and files in the path must belong to the user and live
// within the users root directory on the server.
func (pm *PathMapper) FindFile(ctx context.Context, q *Query, s DirSearch) (string, error) {
	dirs, file := filepath.Split(s.Path)
	if file == "" {
		return "", app.Wrap(app.WrapParams{
			Err:         errors.New("undefined file name"),
			SafeMessage: "Invalid file path",
			StatusCode:  http.StatusBadRequest,
		})
	}

	directoryID, err := pm.FindDir(ctx, q, DirSearch{
		UserID: s.UserID,
		RootID: s.RootID,
		Path:   dirs,
	})
	if err != nil {
		return "", err
	}

	row, err := q.SelectFileByUserDirName(ctx, s.UserID, directoryID, file)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", app.Wrap(app.WrapParams{
				Err:         fmt.Errorf("file %q does not exist [path: %s]", file, s.Path),
				SafeMessage: fmt.Sprintf("File '%s' does not exist", file),
				StatusCode:  http.StatusBadRequest,
			})
		}

		return "", err
	}

	return row.ID, nil
}

// GetDir returns the name based path to the directory (id). The path will not
// contain a trailing slash unless it is the users root path. Users root path
// will be returned as "/".
func (pm *PathMapper) GetDir(ctx context.Context, q *Query, id string) (string, error) {
	namePath, err := q.SelectDirectoryPath(ctx, id)
	if err != nil {
		return "", err
	}

	// Ignore the "root" level.
	userPath := strings.Join(namePath, "/")
	userPath = strings.TrimPrefix(userPath, "root")
	if userPath == "" {
		userPath = "/"
	}

	return userPath, nil
}

// GetFile returns the name based path to the file.
func (pm *PathMapper) GetFile(ctx context.Context, q *Query, dirID string, filename string) (string, error) {
	dirPath, err := pm.GetDir(ctx, q, dirID)
	if err != nil {
		return "", err
	}

	if dirPath == "/" {
		return dirPath + filename, nil
	}

	return dirPath + "/" + filename, nil
}

// GetDirFS returns the file system path to the directory (id).
func (pm *PathMapper) GetDirFS(ctx context.Context, q *Query, id string) (string, error) {
	// Get the path used on the file system.
	idPath, err := q.SelectDirectoryFSPath(ctx, id)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", pm.root, strings.Join(idPath, "/")), nil
}

// GetFileFS returns the file system path to the file.
func (pm *PathMapper) GetFileFS(ctx context.Context, q *Query, dirID string, fileID string) (string, error) {
	dirIDPath, err := q.SelectDirectoryFSPath(ctx, dirID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/%s", pm.root, strings.Join(dirIDPath, "/"), fileID), nil
}
