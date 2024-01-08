package router

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

// Chi is a router that wraps a *chi.Mux.
type Chi struct {
	*chi.Mux
}

// NewChi returns a Chi router.
func NewChi() *Chi {
	return &Chi{Mux: chi.NewMux()}
}

// SetRoute sets a handler for the specified pattern and method. Supported methods are GET, POST, PUT, and DELETE.
func (c *Chi) SetRoute(method string, pattern string, handler http.HandlerFunc) {
	switch method {
	case "GET":
		c.Mux.Get(pattern, handler)
	case "POST":
		c.Mux.Post(pattern, handler)
	case "PUT":
		c.Mux.Put(pattern, handler)
	case "DELETE":
		c.Mux.Delete(pattern, handler)
	default:
		return
	}
}

// SetStatic sets a GET route for static assets. You may call SetRoute instead, but this method is more explicit.
func (c *Chi) SetStatic(pattern string, handler http.HandlerFunc) {
	c.SetRoute("GET", pattern, handler)
}
