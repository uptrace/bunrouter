package bunrouter

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// Router is the main router structure that implements HTTP request routing.
// It maintains a routing tree and handles incoming HTTP requests.
type Router struct {
	config            // embedded router configuration
	Group             // embedded route group
	mu     sync.Mutex // protects the routing tree
	tree   node       // root node of the routing tree
}

// New creates and returns a new Router instance with the given options.
// Options can include middleware, custom handlers for 404 and 405 responses,
// and other router configurations.
func New(opts ...Option) *Router {
	r := &Router{
		tree: node{
			part: "/",
		},
	}

	r.Group.router = r
	r.config.group = &r.Group
	r.methodNotAllowedHandler = methodNotAllowedHandler

	for _, opt := range opts {
		opt.apply(&r.config)
	}

	// Do it after processing middlewares from the options.
	if r.notFoundHandler == nil {
		r.notFoundHandler = r.group.wrap(notFoundHandler)
	}

	return r
}

var _ http.Handler = (*Router)(nil)

// ServeHTTP implements the http.Handler interface.
// It processes the incoming HTTP request and routes it to the appropriate handler.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_ = r.ServeHTTPError(w, req)
}

// ServeHTTPError is similar to ServeHTTP but also returns any error
// that occurred during request handling.
func (r *Router) ServeHTTPError(w http.ResponseWriter, req *http.Request) error {
	handler, params := r.lookup(w, req)
	return handler(w, newRequestParams(req, params))
}

// lookup finds the appropriate handler and parameters for the given HTTP request.
// It returns the handler function and parsed route parameters.
func (r *Router) lookup(w http.ResponseWriter, req *http.Request) (HandlerFunc, Params) {
	path := req.URL.RawPath
	if path == "" {
		path = req.URL.Path
	}

	node, handler, wildcardLen := r.tree.findRoute(req.Method, path)
	if node == nil {
		if redir := r.redir(req.Method, path); redir != nil {
			return redir, Params{}
		}
		return r.notFoundHandler, Params{}
	}

	if handler == nil {
		if redir := r.redir(req.Method, path); redir != nil {
			return redir, Params{}
		}
		handler = node.handlerMap.notAllowed
	}

	return handler.fn, Params{
		node:        node,
		handler:     handler,
		path:        path,
		wildcardLen: uint16(wildcardLen),
	}
}

// redir handles URL redirects for cleaned paths and trailing slash variations.
// It returns a redirect handler if a redirect is needed, nil otherwise.
func (r *Router) redir(method, path string) HandlerFunc {
	if path == "/" {
		return nil
	}

	// Path was not found. Try cleaning it up and search again.
	if cleanPath := CleanPath(path); cleanPath != path {
		if _, handler, _ := r.tree.findRoute(method, cleanPath); handler != nil {
			return redirectHandler(cleanPath)
		}
	}

	if strings.HasSuffix(path, "/") {
		// Try path without a slash.
		cleanPath := path[:len(path)-1]
		if _, handler, _ := r.tree.findRoute(method, cleanPath); handler != nil {
			return redirectHandler(cleanPath)
		}
		return nil
	}

	// Try path with a slash.
	cleanPath := path + "/"
	if _, handler, _ := r.tree.findRoute(method, cleanPath); handler != nil {
		return redirectHandler(cleanPath)
	}
	return nil
}

//------------------------------------------------------------------------------

// CompatRouter provides compatibility layer for the router.
type CompatRouter struct {
	*Router
	*CompatGroup
}

// Compat returns a new CompatRouter instance that wraps the current router.
func (r *Router) Compat() *CompatRouter {
	return &CompatRouter{
		Router:      r,
		CompatGroup: r.Group.Compat(),
	}
}

// VerboseRouter provides a verbose interface to the router.
type VerboseRouter struct {
	*Router
	*VerboseGroup
}

// Verbose returns a new VerboseRouter instance that wraps the current router.
func (r *Router) Verbose() *VerboseRouter {
	return &VerboseRouter{
		Router:       r,
		VerboseGroup: r.Group.Verbose(),
	}
}

//------------------------------------------------------------------------------

// redirectHandler creates a handler function that performs HTTP redirects
// to the specified new path while preserving query parameters and fragments.
func redirectHandler(newPath string) HandlerFunc {
	return func(w http.ResponseWriter, req Request) error {
		newURL := url.URL{
			Path:     newPath,
			RawQuery: req.URL.RawQuery,
			Fragment: req.URL.Fragment,
		}
		http.Redirect(w, req.Request, newURL.String(), http.StatusMovedPermanently)
		return nil
	}
}

// methodNotAllowedHandler is the default handler for requests with methods
// that are not allowed for the matched route.
func methodNotAllowedHandler(w http.ResponseWriter, r Request) error {
	w.WriteHeader(http.StatusMethodNotAllowed)
	return nil
}

// notFoundHandler is the default handler for requests that don't match any route.
func notFoundHandler(w http.ResponseWriter, req Request) error {
	http.NotFound(w, req.Request)
	return nil
}
