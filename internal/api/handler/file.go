package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/cicconee/clox/internal/cloudstore"
	"github.com/go-chi/chi/v5"
)

type File struct {
	cloud *cloudstore.Service
	log   *log.Logger
}

func NewFile(cloud *cloudstore.Service, log *log.Logger) *File {
	return &File{cloud: cloud, log: log}
}

func (f *File) Upload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("{%q: %q,\n%q: %q}", "message", "Uploaded!", "directory_id", chi.URLParam(r, "id"))))
	}
}
