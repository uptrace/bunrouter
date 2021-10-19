package bunrouter

import (
	"net/http"
	"net/url"
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
		tree: node{route: "/", part: "/"},
	}

	r.Group.router = r
	r.config.group = &r.Group

	for _, opt := range opts {
		opt.apply(&r.config)
	}

	// Do it after processing middlewares from the options.
	if r.notFoundHandler == nil {
		r.notFoundHandler = r.config.wrapHandler(notFoundHandler)
	}
	if r.methodNotAllowedHandler == nil {
		r.methodNotAllowedHandler = r.config.wrapHandler(methodNotAllowedHandler)
	}

	return r
}

func (t *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler, params := t.lookup(w, req)
	reqWrapper := Request{
		Request: req,
		params:  params,
	}
	_ = handler(w, reqWrapper)
}

func (r *Router) lookup(w http.ResponseWriter, req *http.Request) (HandlerFunc, Params) {
	path := req.URL.Path
	unescapedPath := req.URL.Path

	if path == "" {
		return r.notFoundHandler, Params{}
	}

	trailingSlash := len(path) > 1 && path[len(path)-1] == '/'
	if trailingSlash {
		path = path[:len(path)-1]
		unescapedPath = unescapedPath[:len(unescapedPath)-1]
	}

	node, handler, wildcardLen := r.tree.findRoute(req.Method, path[1:])
	if node == nil {
		// Path was not found. Try cleaning it up and search again.
		cleanPath := CleanPath(unescapedPath)

		node, _, _ = r.tree.findRoute(req.Method, cleanPath[1:])
		if node == nil {
			return r.notFoundHandler, Params{}
		}

		return redirectHandler(cleanPath), Params{}
	}

	if handler.fn == nil {
		return r.methodNotAllowedHandler, Params{}
	}

	if wildcardLen == 0 && trailingSlash != handler.slash {
		if handler.slash {
			// Need to add a slash.
			return redirectHandler(unescapedPath + "/"), Params{}
		}
		if path != "/" {
			// We need to remove the slash. This was already done at the
			// beginning of the function.
			return redirectHandler(unescapedPath), Params{}
		}
	}

	return handler.fn, Params{
		node:        node,
		path:        path,
		wildcardLen: uint16(wildcardLen),
	}
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
