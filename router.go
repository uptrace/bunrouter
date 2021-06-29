// This is inspired by Julien Schmidt's httprouter, in that it uses a patricia tree, but the
// implementation is rather different. Specifically, the routing rules are relaxed so that a
// single path segment may be a wildcard in one route and a static token in another. This gives a
// nice combination of high performance with a lot of convenience in designing the routing patterns.
package treemux

import (
	"net/http"
	"net/url"
	"sync"
)

func HTTPHandler(handler http.Handler) HandlerFunc {
	return func(w http.ResponseWriter, req Request) error {
		ctx := contextWithRoute(req.Context(), req.route, req.params)
		handler.ServeHTTP(w, req.Request.WithContext(ctx))
		return nil
	}
}

func HTTPHandlerFunc(handler http.HandlerFunc) HandlerFunc {
	return HTTPHandler(http.HandlerFunc(handler))
}

type HandlerFunc func(http.ResponseWriter, Request) error

// RedirectBehavior sets the behavior when the router redirects the request to the
// canonical version of the requested URL using RedirectTrailingSlash or RedirectClean.
// The default behavior is to return a 301 status, redirecting the browser to the version
// of the URL that matches the given pattern.
//
// On a POST request, most browsers that receive a 301 will submit a GET request to
// the redirected URL, meaning that any data will likely be lost. If you want to handle
// and avoid this behavior, you may use Redirect307, which causes most browsers to
// resubmit the request using the original method and request body.
//
// Since 307 is supposed to be a temporary redirect, the new 308 status code has been
// proposed, which is treated the same, except it indicates correctly that the redirection
// is permanent. The big caveat here is that the RFC is relatively recent, and older
// browsers will not know what to do with it. Therefore its use is not recommended
// unless you really know what you're doing.
//
// Finally, the UseHandler value will simply call the handler function for the pattern.
type RedirectBehavior int

const (
	Redirect301 RedirectBehavior = iota // Return 301 Moved Permanently
	Redirect307                         // Return 307 HTTP/1.1 Temporary Redirect
	Redirect308                         // Return a 308 RFC7538 Permanent Redirect
	UseHandler                          // Just call the handler function
)

type Router struct {
	config
	Group

	mu   sync.Mutex
	root *node
}

func New(opts ...Option) *Router {
	tm := &Router{
		config: config{
			notFoundHandler:         nil,
			methodNotAllowedHandler: nil,

			headCanUseGet:         true,
			redirectTrailingSlash: true,
			redirectCleanPath:     true,

			redirectBehavior:       Redirect301,
			redirectMethodBehavior: make(map[string]RedirectBehavior),
		},

		root: &node{path: "/"},
	}

	tm.Group.mux = tm
	tm.config.group = &tm.Group

	for _, opt := range opts {
		opt(&tm.config)
	}

	if tm.notFoundHandler == nil {
		tm.notFoundHandler = tm.config.wrapHandler(notFoundHandler)
	}
	if tm.methodNotAllowedHandler == nil {
		tm.methodNotAllowedHandler = tm.config.wrapHandler(methodNotAllowedHandler)
	}

	return tm
}

// Dump returns a text representation of the routing tree.
func (t *Router) Dump() string {
	return t.root.dumpTree("", "")
}

func (t *Router) redirectStatusCode(method string) (int, bool) {
	var behavior RedirectBehavior
	var ok bool
	if behavior, ok = t.redirectMethodBehavior[method]; !ok {
		behavior = t.redirectBehavior
	}
	switch behavior {
	case Redirect301:
		return http.StatusMovedPermanently, true
	case Redirect307:
		return http.StatusTemporaryRedirect, true
	case Redirect308:
		// Go doesn't have a constant for this yet. Yet another sign
		// that you probably shouldn't use it.
		return 308, true
	case UseHandler:
		return 0, false
	default:
		return http.StatusMovedPermanently, true
	}
}

func redirectHandler(newPath string, statusCode int) HandlerFunc {
	return func(w http.ResponseWriter, req Request) error {
		newURL := url.URL{
			Path:     newPath,
			RawQuery: req.URL.RawQuery,
			Fragment: req.URL.Fragment,
		}
		http.Redirect(w, req.Request, newURL.String(), statusCode)
		return nil
	}
}

func (t *Router) lookup(w http.ResponseWriter, r *http.Request) (HandlerFunc, string, []Param) {
	path := r.RequestURI
	unescapedPath := r.URL.Path
	pathLen := len(path)

	if pathLen > 0 && !t.useURLPath {
		rawQueryLen := len(r.URL.RawQuery)

		if rawQueryLen != 0 || path[pathLen-1] == '?' {
			// Remove any query string and the ?.
			path = path[:pathLen-rawQueryLen-1]
			pathLen = len(path)
		}
	} else {
		// In testing with http.NewRequest,
		// RequestURI is not set so just grab URL.Path instead.
		path = r.URL.Path
		pathLen = len(path)
	}

	trailingSlash := path[pathLen-1] == '/' && pathLen > 1
	if trailingSlash && t.redirectTrailingSlash {
		path = path[:pathLen-1]
		unescapedPath = unescapedPath[:len(unescapedPath)-1]
	}

	n, handler, params := t.root.search(r.Method, path[1:])
	if n == nil {
		if !t.redirectCleanPath {
			return t.notFoundHandler, "", nil
		}

		// Path was not found. Try cleaning it up and search again.
		// TODO Test this
		cleanPath := Clean(unescapedPath)
		n, handler, params = t.root.search(r.Method, cleanPath[1:])
		if n == nil {
			return t.notFoundHandler, "", nil
		}
		if statusCode, ok := t.redirectStatusCode(r.Method); ok {
			// Redirect to the actual path
			return redirectHandler(cleanPath, statusCode), "", nil
		}
	}

	if handler == nil {
		return t.methodNotAllowedHandler, "", nil
	}

	if !n.isCatchAll || t.removeCatchAllTrailingSlash {
		if trailingSlash != n.addSlash && t.redirectTrailingSlash {
			if statusCode, ok := t.redirectStatusCode(r.Method); ok {
				if n.addSlash {
					// Need to add a slash.
					return redirectHandler(unescapedPath+"/", statusCode), "", nil
				}
				if path != "/" {
					// We need to remove the slash. This was already done at the
					// beginning of the function.
					return redirectHandler(unescapedPath, statusCode), "", nil
				}
			}
		}
	}

	return handler, n.route, params
}

func (t *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler, route, params := t.lookup(w, req)
	reqWrapper := Request{
		ctx:     req.Context(),
		Request: req,
		route:   route,
		params:  params,
	}
	_ = handler(w, reqWrapper)
}

// methodNotAllowedHandler is the default handler for TreeMux.MethodNotAllowedHandler,
// which is called for patterns that match, but do not have a handler installed for the
// requested method. It simply writes the status code http.StatusMethodNotAllowed and fills
// in the `Allow` header value appropriately.
func methodNotAllowedHandler(w http.ResponseWriter, r Request) error {
	w.WriteHeader(http.StatusMethodNotAllowed)
	return nil
}

func notFoundHandler(w http.ResponseWriter, req Request) error {
	http.NotFound(w, req.Request)
	return nil
}
