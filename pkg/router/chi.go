package router

import (
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/go-chi/chi"
)

// ChiRouter is an adapter for chi router that implements the Router interface
type ChiRouter struct {
	selectingHost string
	hosts         *map[string]chi.Router
}

// NewChiRouterWithOptions creates a new instance of ChiRouter
// with the provided options
func NewChiRouterWithOptions(options Options) *ChiRouter {
	router := chi.NewRouter()
	router.NotFound(options.NotFoundHandler)

	crouter := &ChiRouter{
		hosts: &map[string]chi.Router{},
	}

	crouter.newHostRouter("*")

	return crouter
}

// NewChiRouter creates a new instance of ChiRouter
func NewChiRouter() *ChiRouter {
	return NewChiRouterWithOptions(DefaultOptions)
}

// ServeHTTP server the HTTP requests
func (r *ChiRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	hostName := requestHost(req)
	logrus.Info("REQUESTING ", hostName)

	if router, ok := (*r.hosts)[hostName]; ok {
		router.ServeHTTP(w, req)
		return
	}

	(*r.hosts)["*"].ServeHTTP(w, req)
}

// Any register a path to all HTTP methods
func (r *ChiRouter) Any(host, path string, handler http.HandlerFunc, handlers ...Constructor) {
	r.with(host, handlers...).Handle(path, handler)
}

// Handle registers a path, method and handlers to the router
func (r *ChiRouter) Handle(host, method string, path string, handler http.HandlerFunc, handlers ...Constructor) {
	switch method {
	case http.MethodGet:
		r.GET(host, path, handler, handlers...)
	case http.MethodPost:
		r.POST(host, path, handler, handlers...)
	case http.MethodPut:
		r.PUT(host, path, handler, handlers...)
	case http.MethodPatch:
		r.PATCH(host, path, handler, handlers...)
	case http.MethodDelete:
		r.DELETE(host, path, handler, handlers...)
	case http.MethodHead:
		r.HEAD(host, path, handler, handlers...)
	case http.MethodOptions:
		r.OPTIONS(host, path, handler, handlers...)
	}
}

// GET registers a HTTP GET path
func (r *ChiRouter) GET(host, path string, handler http.HandlerFunc, handlers ...Constructor) {
	r.with(host, handlers...).Get(path, handler)
}

// POST registers a HTTP POST path
func (r *ChiRouter) POST(host, path string, handler http.HandlerFunc, handlers ...Constructor) {
	r.with(host, handlers...).Post(path, handler)
}

// PUT registers a HTTP PUT path
func (r *ChiRouter) PUT(host, path string, handler http.HandlerFunc, handlers ...Constructor) {
	r.with(host, handlers...).Put(path, handler)
}

// DELETE registers a HTTP DELETE path
func (r *ChiRouter) DELETE(host, path string, handler http.HandlerFunc, handlers ...Constructor) {
	r.with(host, handlers...).Delete(path, handler)
}

// PATCH registers a HTTP PATCH path
func (r *ChiRouter) PATCH(host, path string, handler http.HandlerFunc, handlers ...Constructor) {
	r.with(host, handlers...).Patch(path, handler)
}

// HEAD registers a HTTP HEAD path
func (r *ChiRouter) HEAD(host, path string, handler http.HandlerFunc, handlers ...Constructor) {
	r.with(host, handlers...).Head(path, handler)
}

// OPTIONS registers a HTTP OPTIONS path
func (r *ChiRouter) OPTIONS(host, path string, handler http.HandlerFunc, handlers ...Constructor) {
	r.with(host, handlers...).Options(path, handler)
}

// TRACE registers a HTTP TRACE path
func (r *ChiRouter) TRACE(host, path string, handler http.HandlerFunc, handlers ...Constructor) {
	r.with(host, handlers...).Trace(path, handler)
}

// CONNECT registers a HTTP CONNECT path
func (r *ChiRouter) CONNECT(host, path string, handler http.HandlerFunc, handlers ...Constructor) {
	r.with(host, handlers...).Connect(path, handler)
}

// Group creates a child router for a specific path
func (r *ChiRouter) Group(host, path string, subRoute func(r chi.Router)) Router {
	r.selectRouter(host).Route(path, subRoute)
	return r
}

// Use attaches a middleware to the router
func (r *ChiRouter) Use(host string, handlers ...Constructor) Router {
	r.selectRouter(host).Use(r.wrapConstructor(handlers)...)
	return r
}

// RoutesCount returns number of routes registered
func (r *ChiRouter) RoutesCount() int {
	count := 0

	for _, router := range *r.hosts {
		count = count + r.routesCount(router)
	}

	return r.routesCount((*r.hosts)["*"])
}

func (r *ChiRouter) routesCount(routes chi.Routes) int {
	count := len(routes.Routes())
	for _, route := range routes.Routes() {
		if nil != route.SubRoutes {
			count += r.routesCount(route.SubRoutes)
		}
	}
	return count
}

func (r *ChiRouter) selectRouter(hostName string) chi.Router {

	if _, ok := (*r.hosts)[hostName]; !ok {
		r.newHostRouter(hostName)
	}

	return (*r.hosts)[hostName]
}

func (r *ChiRouter) with(host string, handlers ...Constructor) chi.Router {
	return r.selectRouter(host).With(r.wrapConstructor(handlers)...)
}

func (r *ChiRouter) newHostRouter(hostName string) {
	hostRouter := chi.NewRouter()
	hostRouter.NotFound(DefaultOptions.NotFoundHandler)

	(*r.hosts)[hostName] = hostRouter
}

func (r *ChiRouter) wrapConstructor(handlers []Constructor) []func(http.Handler) http.Handler {
	var cons = make([]func(http.Handler) http.Handler, 0)
	for _, m := range handlers {
		cons = append(cons, (func(http.Handler) http.Handler)(m))
	}
	return cons
}

func requestHost(r *http.Request) (host string) {
	// not standard, but most popular
	host = r.Header.Get("X-Forwarded-Host")
	if host != "" {
		return
	}

	// RFC 7239
	host = r.Header.Get("Forwarded")
	_, _, host = parseForwarded(host)
	if host != "" {
		return
	}

	// if all else fails fall back to request host
	host = r.Host

	return host
}

func parseForwarded(forwarded string) (addr, proto, host string) {
	if forwarded == "" {
		return
	}

	for _, forwardedPair := range strings.Split(forwarded, ";") {
		if tv := strings.SplitN(forwardedPair, "=", 2); len(tv) == 2 {
			token, value := tv[0], tv[1]
			token = strings.TrimSpace(token)
			value = strings.TrimSpace(strings.Trim(value, `"`))
			switch strings.ToLower(token) {
			case "for":
				addr = value
			case "proto":
				proto = value
			case "host":
				host = value
			}

		}
	}

	return addr, proto, host
}
