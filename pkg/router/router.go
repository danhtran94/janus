package router

import (
	"net/http"

	"github.com/go-chi/chi"
)

// Constructor for a piece of middleware.
// Some middleware use this constructor out of the box,
// so in most cases you can just pass somepackage.New
type Constructor func(http.Handler) http.Handler

// URLParam returns the url parameter from a http.Request object.
func URLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

// Router defines the basic methods for a router
type Router interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	Handle(host, method string, path string, handler http.HandlerFunc, handlers ...Constructor)
	Any(host, path string, handler http.HandlerFunc, handlers ...Constructor)
	GET(host, path string, handler http.HandlerFunc, handlers ...Constructor)
	POST(host, path string, handler http.HandlerFunc, handlers ...Constructor)
	PUT(host, path string, handler http.HandlerFunc, handlers ...Constructor)
	DELETE(host, path string, handler http.HandlerFunc, handlers ...Constructor)
	PATCH(host, path string, handler http.HandlerFunc, handlers ...Constructor)
	HEAD(host, path string, handler http.HandlerFunc, handlers ...Constructor)
	OPTIONS(host, path string, handler http.HandlerFunc, handlers ...Constructor)
	TRACE(host, path string, handler http.HandlerFunc, handlers ...Constructor)
	CONNECT(host, path string, handler http.HandlerFunc, handlers ...Constructor)
	Group(host, path string, subRoute func(chi.Router)) Router
	Use(host string, handlers ...Constructor) Router

	RoutesCount() int
}

// Options are the HTTPTreeMuxRouter options
type Options struct {
	NotFoundHandler           http.HandlerFunc
	SafeAddRoutesWhileRunning bool
}

// DefaultOptions are the default router options
var DefaultOptions = Options{
	NotFoundHandler:           http.NotFound,
	SafeAddRoutesWhileRunning: true,
}
