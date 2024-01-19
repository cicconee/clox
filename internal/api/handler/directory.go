package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/cicconee/clox/internal/api/auth"
	"github.com/cicconee/clox/internal/app"
	"github.com/cicconee/clox/internal/cloudstore"
	"github.com/go-chi/chi/v5"
)

type Directory struct {
	dirs *cloudstore.DirService
	log  *log.Logger
}

func NewDirectory(dirs *cloudstore.DirService, log *log.Logger) *Directory {
	return &Directory{dirs: dirs, log: log}
}

// The request body when creating a new directory.
type newDirRequest struct {
	Name string `json:"name"`
}

// parseNewDirRequest parses the request body into a newDirRequest.
//
// parseNewDirRequest does not close r.Body.
func parseNewDirRequest(r *http.Request) (newDirRequest, error) {
	var request newDirRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return newDirRequest{}, app.Wrap(app.WrapParams{
			Err:         err,
			SafeMessage: "Invalid request body",
			StatusCode:  http.StatusBadRequest,
		})
	}

	return request, nil
}

// The response body when creating a new directory.
type newDirResponse struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"owner_id"`
	DirName   string    `json:"directory_name"`
	DirPath   string    `json:"directory_path"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	LastWrite time.Time `json:"last_write"`
}

// marshalNewDirResponse converts a cloudstore.Dir to a newDirResponse
// and marshals it to json byte slice.
func marshalNewDirResponse(dir cloudstore.Dir) ([]byte, error) {
	return json.Marshal(&newDirResponse{
		ID:        dir.ID,
		OwnerID:   dir.Owner,
		DirName:   dir.Name,
		DirPath:   dir.Path,
		CreatedAt: dir.CreatedAt,
		UpdatedAt: dir.UpdatedAt,
		LastWrite: dir.LastWrite,
	})
}

// New returns a http.HandlerFunc that handles creating a new directory when
// the directories parent ID is apart of the URL path. The name of the directory
// should be specified in a json request body.
//
// New expects the user ID to be in the request context. To set the user ID in the
// request context, use auth.SetUserIDContext.
func (d *Directory) New() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d.new(w, r, func(userID string, request newDirRequest) (cloudstore.Dir, error) {
			return d.dirs.New(r.Context(), userID, request.Name, chi.URLParam(r, "id"))
		})
	}
}

// NewPath returns a http.HandlerFunc that handles creating a new directory
// when the directories path is specified as a URL query parameter with the
// key "path". The name of the directory should be specified in a json request
// body.
//
// NewPath expects the user ID to be in the request context. To set the user
// ID in the request context, use auth.SetUserIDContext.
func (d *Directory) NewPath() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d.new(w, r, func(userID string, request newDirRequest) (cloudstore.Dir, error) {
			return d.dirs.NewPath(r.Context(), userID, request.Name, r.URL.Query().Get("path"))
		})
	}
}

// new is a modified http handler for creating a new directory. The function,
// newDirFunc, will be passed the user ID of the user making the request and
// the request body parsed as a newDirRequest. newDirFunc should create a new
// directory and return it.
func (d *Directory) new(w http.ResponseWriter, r *http.Request, newDirFunc func(string, newDirRequest) (cloudstore.Dir, error)) {
	userID := auth.GetUserIDContext(r.Context())

	request, err := parseNewDirRequest(r)
	if err != nil {
		app.WriteJSONError(w, err)
		d.log.Printf("[ERROR] [%s %s] Failed to decode request body: %v\n", r.Method, r.URL.Path, err)
		return
	}
	defer r.Body.Close()

	dir, err := newDirFunc(userID, request)
	if err != nil {
		app.WriteJSONError(w, err)
		d.log.Printf("[ERROR] [%s %s] Failed creating directory: %v\n", r.Method, r.URL.Path, err)
		return
	}

	resp, err := marshalNewDirResponse(dir)
	if err != nil {
		app.WriteJSONError(w, err)
		d.log.Printf("[ERROR] [%s %s] Failed marshalling response: %v\n", r.Method, r.URL.Path, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
