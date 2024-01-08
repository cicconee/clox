package server

import (
	"context"
	"fmt"
	"github.com/cicconee/clox/internal/router"
	"net/http"
)

// HTTP is a http server that will serve the route handler and static assets.
type HTTP struct {
	httpServer *http.Server
	handler    *router.Chi
}

// New will create a HTTP that will serve the handler on the host address and port.
func New(host string, port string, handler *router.Chi) *HTTP {
	return &HTTP{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf("%s:%s", host, port),
			Handler: handler,
		},
		handler: handler,
	}
}

// Middleware is a function type that wraps a http.HandlerFunc around another http.HandlerFunc.
type Middleware func(handlerFunc http.HandlerFunc) http.HandlerFunc

// SetRoute sets this HTTP server to serve the http.HandlerFunc for the specified method and pattern. All the
// middlewares will be applied starting at index 0 to n. For example, middleware[0] will wrap middleware[1], until
// middleware[n], middleware[n] will then wrap the handler.
func (s *HTTP) SetRoute(method string, pattern string, handler http.HandlerFunc, middlewares ...Middleware) {
	for i := len(middlewares) - 1; i >= 0; i-- {
		mw := middlewares[i]
		handler = mw(handler)
	}

	s.handler.SetRoute(method, pattern, handler)
}

// SetStatic sets this HTTP server to serve the static asset handler. You may call SetRoute instead without any
// middlewares, but this method is more explicit.
func (s *HTTP) SetStatic(pattern string, handler http.HandlerFunc) {
	s.handler.SetStatic(pattern, handler)
}

// Start will initiate this HTTP server to listen on its host and port.
func (s *HTTP) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown will stop this HTTP server.
func (s *HTTP) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
