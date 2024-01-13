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

type Cloud struct {
	cloud *cloudstore.Service
	log   *log.Logger
}

func NewCloud(cloud *cloudstore.Service, log *log.Logger) *Cloud {
	return &Cloud{cloud: cloud, log: log}
}

func (c *Cloud) NewDir() http.HandlerFunc {
	type requestBody struct {
		Name string `json:"name"`
	}

	type response struct {
		ID        string    `json:"id"`
		OwnerID   string    `json:"owner_id"`
		DirName   string    `json:"directory_name"`
		DirPath   string    `json:"directory_path"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		LastWrite time.Time `json:"last_write"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserIDContext(r.Context())
		parentID := chi.URLParam(r, "id")

		var request requestBody
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			app.WriteJSONError(w, app.Wrap(app.WrapParams{
				Err:         err,
				SafeMessage: "Invalid request body",
				StatusCode:  http.StatusBadRequest,
			}))
			c.log.Printf("[ERROR] [%s %s] Failed to decode request body: %v\n", r.Method, r.URL.Path, err)
			return
		}
		defer r.Body.Close()

		dir, err := c.cloud.NewDir(r.Context(), userID, request.Name, parentID)
		if err != nil {
			app.WriteJSONError(w, err)
			c.log.Printf("[ERROR] [%s %s] Failed creating directory: %v\n", r.Method, r.URL.Path, err)
			return
		}

		resp, err := json.Marshal(&response{
			ID:        dir.ID,
			OwnerID:   dir.Owner,
			DirName:   dir.Name,
			DirPath:   dir.Path,
			CreatedAt: dir.CreatedAt,
			UpdatedAt: dir.UpdatedAt,
			LastWrite: dir.LastWrite,
		})
		if err != nil {
			app.WriteJSONError(w, err)
			c.log.Printf("[ERROR] [%s %s] Failed marshalling response: %v\n", r.Method, r.URL.Path, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}

func (c *Cloud) NewDirPath() http.HandlerFunc {
	type requestBody struct {
		Name string `json:"name"`
	}

	type response struct {
		ID        string    `json:"id"`
		OwnerID   string    `json:"owner_id"`
		DirName   string    `json:"directory_name"`
		DirPath   string    `json:"directory_path"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		LastWrite time.Time `json:"last_write"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		userID := auth.GetUserIDContext(r.Context())
		path := r.URL.Query().Get("path")

		var request requestBody
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			app.WriteJSONError(w, app.Wrap(app.WrapParams{
				Err:         err,
				SafeMessage: "Invalid request body",
				StatusCode:  http.StatusBadRequest,
			}))
			c.log.Printf("[ERROR] [%s %s] Failed to decode request body: %v\n", r.Method, r.URL.Path, err)
			return
		}
		defer r.Body.Close()

		dir, err := c.cloud.NewDirPath(r.Context(), userID, request.Name, path)
		if err != nil {
			app.WriteJSONError(w, err)
			c.log.Printf("[ERROR] [%s %s] Failed creating directory: %v\n", r.Method, r.URL.Path, err)
			return
		}

		resp, err := json.Marshal(&response{
			ID:        dir.ID,
			OwnerID:   dir.Owner,
			DirName:   dir.Name,
			DirPath:   dir.Path,
			CreatedAt: dir.CreatedAt,
			UpdatedAt: dir.UpdatedAt,
			LastWrite: dir.LastWrite,
		})
		if err != nil {
			app.WriteJSONError(w, err)
			c.log.Printf("[ERROR] [%s %s] Failed marshalling response: %v\n", r.Method, r.URL.Path, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}
}
