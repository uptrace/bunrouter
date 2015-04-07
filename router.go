package httptreemux

import (
	"fmt"
	"github.com/dimfeld/httppath"
	"net/http"
)

type HandlerFunc func(http.ResponseWriter, *http.Request, map[string]string)
type PanicHandler func(http.ResponseWriter, *http.Request, interface{})

type TreeMux struct {
	root *node

	// The default PanicHandler just returns a 500 code.
	PanicHandler PanicHandler
	// The default NotFoundHandler is http.NotFound.
	NotFoundHandler func(w http.ResponseWriter, r *http.Request)
	// MethodNotAllowedHandler is called when a pattern matches, but that
	// pattern does not have a handler for the requested method. The default
	// handler just writes the status code http.StatusMethodNotAllowed and adds
	// the required Allowed header.
	// The methods parameter contains the map of each method to the corresponding
	// handler function.
	MethodNotAllowedHandler func(w http.ResponseWriter, r *http.Request,
		methods map[string]HandlerFunc)
	// HeadCanUseGet allows the router to use the GET handler to respond to
	// HEAD requests if no explicit HEAD handler has been added for the
	// matching pattern. This is true by default.
	HeadCanUseGet bool

	// RedirectCleanPath allows router to try clean the current request path,
	// if no handler is registered for it.
	// It tries to fix the path using CleanPath from github.com/dimfeld/httppath
	// This is true by default.
	RedirectCleanPath bool

	// This enables automatic redirection in case router doesn't find a matching route
	// for the current request path but a handler for the path with or without the trailing
	// slash exists. This is true by default.
	RedirectTrailingSlash bool
}

// Dump returns a text representation of the routing tree.
func (t *TreeMux) Dump() string {
	return t.root.dumpTree("", "")
}

func (t *TreeMux) Handle(verb, path string, handler HandlerFunc) {
	if path[0] != '/' {
		panic(fmt.Sprintf("Path %s must start with slash", path))
	}

	addSlash := false
	if len(path) > 1 && path[len(path)-1] == '/' && t.RedirectTrailingSlash {
		addSlash = true
		path = path[:len(path)-1]
	}

	node := t.root.addPath(path[1:])
	if addSlash {
		node.addSlash = true
	}
	node.setHandler(verb, handler)
}

func (t *TreeMux) GET(path string, handler HandlerFunc) {
	t.Handle("GET", path, handler)
}

func (t *TreeMux) POST(path string, handler HandlerFunc) {
	t.Handle("POST", path, handler)
}

func (t *TreeMux) PUT(path string, handler HandlerFunc) {
	t.Handle("PUT", path, handler)
}

func (t *TreeMux) DELETE(path string, handler HandlerFunc) {
	t.Handle("DELETE", path, handler)
}

func (t *TreeMux) PATCH(path string, handler HandlerFunc) {
	t.Handle("PATCH", path, handler)
}

func (t *TreeMux) HEAD(path string, handler HandlerFunc) {
	t.Handle("HEAD", path, handler)
}

func (t *TreeMux) OPTIONS(path string, handler HandlerFunc) {
	t.Handle("OPTIONS", path, handler)
}

func (t *TreeMux) serveHTTPPanic(w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		t.PanicHandler(w, r, err)
	}
}

func (t *TreeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if t.PanicHandler != nil {
		defer t.serveHTTPPanic(w, r)
	}

	path := r.RequestURI

	rawQueryLen := len(r.URL.RawQuery)
	pathLen := len(path)
	if rawQueryLen != 0 || path[pathLen-1] == '?' {
		// Remove any query string and the ?.
		path = path[:pathLen-rawQueryLen-1]
		pathLen = len(path)
	}

	trailingSlash := path[pathLen-1] == '/' && pathLen > 1
	if trailingSlash && t.RedirectTrailingSlash {
		path = path[:pathLen-1]
	}
	// params := make(map[string]string)
	var params map[string]string
	n := t.root.search(path[1:], &params)
	if n == nil {
		if t.RedirectCleanPath {
			// Path was not found. Try cleaning it up and search again.
			// TODO Test this
			cleanPath := httppath.Clean(path)
			n = t.root.search(cleanPath[1:], &params)
			if n == nil {
				// Still nothing found.
				t.NotFoundHandler(w, r)
			} else {
				// Redirect to the actual path
				http.Redirect(w, r, cleanPath, http.StatusMovedPermanently)
			}
			return
		} else {
			t.NotFoundHandler(w, r)
			return
		}
	}

	handler, ok := n.leafHandler[r.Method]
	if !ok {
		if r.Method == "HEAD" && t.HeadCanUseGet {
			handler, ok = n.leafHandler["GET"]
		}

		if !ok {
			t.MethodNotAllowedHandler(w, r, n.leafHandler)
			return
		}
	}

	if trailingSlash != n.addSlash && t.RedirectTrailingSlash {
		if n.addSlash {
			// Need to add a slash.
			http.Redirect(w, r, path+"/", http.StatusMovedPermanently)
		} else if path != "/" {
			// We need to remove the slash. This was already done at the
			// beginning of the function.
			http.Redirect(w, r, path, http.StatusMovedPermanently)
		}
		return
	}

	handler(w, r, params)
}

// MethodNotAllowedHandler is the default handler for
// TreeMux.MethodNotAllowedHandler. It writes the status code
// http.StatusMethodNotAllowed, and nothing else.
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request,
	methods map[string]HandlerFunc) {

	for m := range methods {
		w.Header().Add("Allow", m)
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

func New() *TreeMux {
	root := &node{path: "/"}
	return &TreeMux{root: root,
		NotFoundHandler:         http.NotFound,
		MethodNotAllowedHandler: MethodNotAllowedHandler,
		HeadCanUseGet:           true,
		RedirectTrailingSlash:   true,
		RedirectCleanPath:       true,
	}
}
