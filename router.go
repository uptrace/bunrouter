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

type PathSource int

const (
	Redirect301 RedirectBehavior = iota // Return 301 Moved Permanently
	Redirect307                         // Return 307 HTTP/1.1 Temporary Redirect
	Redirect308                         // Return a 308 RFC7538 Permanent Redirect
	UseHandler                          // Just call the handler function

	RequestURI PathSource = iota // Use r.RequestURI
	URLPath                      // Use r.URL.Path
)

// LookupResult contains information about a route lookup, which is returned from Lookup and
// can be passed to ServeLookupResult if the request should be served.
type LookupResult struct {
	// StatusCode informs the caller about the result of the lookup.
	// This will generally be `http.StatusNotFound` or `http.StatusMethodNotAllowed` for an
	// error case. On a normal success, the statusCode will be `http.StatusOK`. A redirect code
	// will also be used in the case
	StatusCode int
	route      string
	handler    HandlerFunc
	params     Params
}

type TreeMux struct {
	config
	Group

	mu   sync.Mutex
	root *node
}

func New(opts ...Option) *TreeMux {
	tm := &TreeMux{
		config: config{
			notFoundHandler:         nil,
			methodNotAllowedHandler: nil,

			headCanUseGet:         true,
			redirectTrailingSlash: true,
			redirectCleanPath:     true,

			redirectBehavior:       Redirect301,
			redirectMethodBehavior: make(map[string]RedirectBehavior),

			pathSource: RequestURI,
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
func (t *TreeMux) Dump() string {
	return t.root.dumpTree("", "")
}

func (t *TreeMux) redirectStatusCode(method string) (int, bool) {
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

func (t *TreeMux) lookup(w http.ResponseWriter, r *http.Request) (LookupResult, bool) {
	path := r.RequestURI
	unescapedPath := r.URL.Path
	pathLen := len(path)
	if pathLen > 0 && t.pathSource == RequestURI {
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
		if t.redirectCleanPath {
			// Path was not found. Try cleaning it up and search again.
			// TODO Test this
			cleanPath := Clean(unescapedPath)
			n, handler, params = t.root.search(r.Method, cleanPath[1:])
			if n == nil {
				return LookupResult{
					StatusCode: http.StatusNotFound,
				}, false
			}
			if statusCode, ok := t.redirectStatusCode(r.Method); ok {
				// Redirect to the actual path
				return LookupResult{
					StatusCode: statusCode,
					handler:    redirectHandler(cleanPath, statusCode),
				}, true
			}
		} else {
			return LookupResult{
				StatusCode: http.StatusNotFound,
			}, false
		}
	}

	if handler == nil {
		return LookupResult{
			StatusCode: http.StatusMethodNotAllowed,
		}, false
	}

	if !n.isCatchAll || t.removeCatchAllTrailingSlash {
		if trailingSlash != n.addSlash && t.redirectTrailingSlash {
			if statusCode, ok := t.redirectStatusCode(r.Method); ok {
				var h HandlerFunc
				if n.addSlash {
					// Need to add a slash.
					h = redirectHandler(unescapedPath+"/", statusCode)
				} else if path != "/" {
					// We need to remove the slash. This was already done at the
					// beginning of the function.
					h = redirectHandler(unescapedPath, statusCode)
				}

				if h != nil {
					return LookupResult{
						StatusCode: statusCode,
						handler:    h,
					}, true
				}
			}
		}
	}

	lr := LookupResult{
		StatusCode: http.StatusOK,
		route:      n.route,
		handler:    handler,
		params:     params,
	}

	return lr, true
}

// Lookup performs a lookup without actually serving the request or mutating the request or response.
// The return values are a LookupResult and a boolean. The boolean will be true when a handler
// was found or the lookup resulted in a redirect which will point to a real handler. It is false
// for requests which would result in a `StatusNotFound` or `StatusMethodNotAllowed`.
//
// Regardless of the returned boolean's value, the LookupResult may be passed to ServeLookupResult
// to be served appropriately.
func (t *TreeMux) Lookup(w http.ResponseWriter, r *http.Request) (LookupResult, bool) {
	return t.lookup(w, r)
}

// ServeLookupResult serves a request, given a lookup result from the Lookup function.
func (t *TreeMux) ServeLookupResult(w http.ResponseWriter, req *http.Request, lr LookupResult) {
	handler := lr.handler

	if handler == nil {
		if lr.StatusCode == http.StatusMethodNotAllowed {
			handler = t.methodNotAllowedHandler
		} else {
			handler = t.notFoundHandler
		}
	}

	reqWrapper := Request{
		ctx:     req.Context(),
		Request: req,
		route:   lr.route,
		Params:  lr.params,
	}
	_ = handler(w, reqWrapper)
}

func (t *TreeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, _ := t.lookup(w, r)
	t.ServeLookupResult(w, r, result)
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
