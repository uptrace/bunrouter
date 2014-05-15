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

	PanicHandler    PanicHandler
	NotFoundHandler func(w http.ResponseWriter, r *http.Request)
}

// Dump returns a text representation of the routing tree.
func (t *TreeMux) Dump() string {
	return t.root.dumpTree("")
}

func (t *TreeMux) Handle(verb, path string, handler HandlerFunc) {
	if path[0] != '/' {
		panic(fmt.Sprintf("Path %s must start with slash", path))
	}

	addSlash := false
	if len(path) > 1 && path[len(path)-1] == '/' {
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

func (t *TreeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if t.PanicHandler != nil {
		defer func() {
			if err := recover(); err != nil {
				t.PanicHandler(w, r, err)
			}
		}()
	}

	path := r.URL.Path
	pathLen := len(path)
	trailingSlash := path[pathLen-1] == '/'
	if pathLen > 1 && trailingSlash {
		path = path[:pathLen-1]
	}
	params := make(map[string]string)
	n := t.root.search(path[1:], params)
	if n == nil {
		// Path was not found. Try cleaning it up and search again.
		cleanPath := httppath.Clean(path)
		n := t.root.search(cleanPath[1:], params)
		if n == nil {
			// Still nothing found.
			t.NotFoundHandler(w, r)
			return
		}
	}

	handler, ok := n.leafHandler[r.Method]
	if !ok {
		t.NotFoundHandler(w, r)
		return
	}

	if pathLen > 1 && trailingSlash != n.addSlash {
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

func New() *TreeMux {
	root := &node{path: "/"}
	return &TreeMux{root: root,
		NotFoundHandler: http.NotFound,
		PanicHandler:    SimplePanicHandler}
}