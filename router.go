package bunrouter

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type Router struct {
	config
	Group

	mu   sync.Mutex
	tree node
}

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

// ServeHTTP implements http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_ = r.ServeHTTPError(w, req)
}

// ServeHTTPError is like ServeHTTP, but it also returns the error returned
// by the matched route handler.
func (r *Router) ServeHTTPError(w http.ResponseWriter, req *http.Request) error {
	handler, params := r.lookup(w, req)
	return handler(w, newRequestParams(req, params))
}

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

type CompatRouter struct {
	*Router
	*CompatGroup
}

func (r *Router) Compat() *CompatRouter {
	return &CompatRouter{
		Router:      r,
		CompatGroup: r.Group.Compat(),
	}
}

type VerboseRouter struct {
	*Router
	*VerboseGroup
}

func (r *Router) Verbose() *VerboseRouter {
	return &VerboseRouter{
		Router:       r,
		VerboseGroup: r.Group.Verbose(),
	}
}

//------------------------------------------------------------------------------

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

func methodNotAllowedHandler(w http.ResponseWriter, r Request) error {
	w.WriteHeader(http.StatusMethodNotAllowed)
	return nil
}

func notFoundHandler(w http.ResponseWriter, req Request) error {
	http.NotFound(w, req.Request)
	return nil
}
