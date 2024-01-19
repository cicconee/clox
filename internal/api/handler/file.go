package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cicconee/clox/internal/api/auth"
	"github.com/cicconee/clox/internal/app"
	"github.com/cicconee/clox/internal/cloudstore"
	"github.com/go-chi/chi/v5"
)

type File struct {
	files *cloudstore.FileService
	log   *log.Logger
}

func NewFile(files *cloudstore.FileService, log *log.Logger) *File {
	return &File{files: files, log: log}
}

// uploadFileResponse encapsulates the result of a file upload operation
// in JSON format.
type uploadFileResponse struct {
	ID          string    `json:"id"`
	DirectoryID string    `json:"directory_id"`
	Name        string    `json:"file_name"`
	Path        string    `json:"file_path"`
	Size        int64     `json:"file_size"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

// uploadErrorResponse encapsulates a failed file upload operation in JSON
// format.
type uploadErrorResponse struct {
	FileName string `json:"file_name"`
	Size     int64  `json:"file_size"`
	Error    string `json:"error"`
}

// uploadResponse represents the response body of a batch file upload
// operation in JSON format.
type uploadResponse struct {
	Uploads []uploadFileResponse  `json:"uploads"`
	Errors  []uploadErrorResponse `json:"errors"`
}

// marshalUploadResponse converts a []cloudstore.BatchSave into a
// uploadResponse and marshals it to a []byte to serve as a JSON response
// body.
func marshalUploadResponse(r []cloudstore.BatchSave) ([]byte, error) {
	uploads := []uploadFileResponse{}
	errors := []uploadErrorResponse{}
	for _, b := range r {
		if b.Err != nil {
			errors = append(errors, uploadErrorResponse{
				FileName: b.Name,
				Size:     b.Size,
				Error:    b.Msg(),
			})
		} else {
			uploads = append(uploads, uploadFileResponse{
				ID:          b.ID,
				DirectoryID: b.DirectoryID,
				Name:        b.Name,
				Path:        b.Path,
				Size:        b.Size,
				UploadedAt:  b.UploadedAt.UTC(),
			})
		}
	}

	return json.Marshal(&uploadResponse{
		Uploads: uploads,
		Errors:  errors,
	})
}

// Upload return a http.HandlerFunc that handles uploading 1 or many files to
// a specified directory when the directory ID is apart of the URL path.
//
// Upload expects the user ID to be in the request context. To set the user ID in
// the request context, use auth.SetUserIDContext.
func (f *File) Upload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserIDContext(r.Context())
		directoryID := chi.URLParam(r, "id")

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			if errors.Is(err, http.ErrNotMultipart) {
				err = app.Wrap(app.WrapParams{
					Err:         fmt.Errorf("invalid Content-Type header: %w", err),
					SafeMessage: "Request Content-Type must be multipart/form-data for file uploads",
					StatusCode:  http.StatusBadRequest,
				})
			} else if errors.Is(err, http.ErrMissingBoundary) {
				err = app.Wrap(app.WrapParams{
					Err:         fmt.Errorf("request Content-Type=multipar/formdata missing boundary: %w", err),
					SafeMessage: "Request header Content-Type=multipart/form-data must include a boundary",
					StatusCode:  http.StatusBadRequest,
				})
			} else {
				err = app.Wrap(app.WrapParams{
					Err:         fmt.Errorf("malformed request: %w", err),
					SafeMessage: "Malformed request",
					StatusCode:  http.StatusBadRequest,
				})
			}

			app.WriteJSONError(w, err)
			f.log.Printf("[ERROR] %v\n", err)
			return
		}

		formdata := r.MultipartForm
		files := formdata.File["file_uploads"]

		result, err := f.files.SaveFileBatch(r.Context(), userID, directoryID, files)
		if err != nil {
			app.WriteJSONError(w, err)
			f.log.Printf("[ERROR] [%s %s] Failed to save files: %v\n", r.Method, r.URL.Path, err)
			return
		}

		resp, err := marshalUploadResponse(result)
		if err != nil {
			app.WriteJSONError(w, err)
			f.log.Printf("[ERROR] [%s %s] Failed marshalling response: %v\n", r.Method, r.URL.Path, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

// Download returns a http.HandlerFunc that handles downloading a file when the
// file ID is apart of the URL path.
//
// Download expects the user ID to be in the request context. To set the user ID in
// the request context, use auth.SetUserIDContext.
func (f *File) Download() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserIDContext(r.Context())
		fileID := chi.URLParam(r, "id")

		file, err := f.files.FileInfo(r.Context(), userID, fileID)
		if err != nil {
			app.WriteJSONError(w, err)
			f.log.Printf("[ERROR] [%s %s] Failed getting file info: %v\n", r.Method, r.URL.Path, err)
			return
		}

		w.Header().Set("Content-Disposition", "attachment; filename="+file.Name)
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeFile(w, r, file.FSPath)
	}
}
