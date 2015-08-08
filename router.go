// This is inspired by Julien Schmidt's httprouter, in that it uses a patricia tree, but the
// implementation is rather different. Specifically, the routing rules are relaxed so that a
// single path segment may be a wildcard in one route and a static token in another. This gives a
// nice combination of high performance with a lot of convenience in designing the routing patterns.
package httptreemux

import (
	"fmt"
	"github.com/dimfeld/httppath"
	"net/http"
)

// The params argument contains the parameters parsed from wildcards and catch-alls in the URL.
type HandlerFunc func(http.ResponseWriter, *http.Request, map[string]string)
type PanicHandler func(http.ResponseWriter, *http.Request, interface{})

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

	// RedirectCleanPath allows the router to try clean the current request path,
	// if no handler is registered for it, using CleanPath from github.com/dimfeld/httppath.
	// This is true by default.
	RedirectCleanPath bool

	// RedirectTrailingSlash enables automatic redirection in case router doesn't find a matching route
	// for the current request path but a handler for the path with or without the trailing
	// slash exists. This is true by default.
	RedirectTrailingSlash bool

	// RemoveCatchAllTrailingSlash removes the trailing slash when a catch-all pattern
	// is matched, if set to true. By default, catch-all paths are never redirected.
	RemoveCatchAllTrailingSlash bool

	// RedirectBehavior sets the default redirect behavior when RedirectTrailingSlash or
	// RedirectCleanPath are true. The default value is Redirect301.
	RedirectBehavior RedirectBehavior

	// RedirectMethodBehavior overrides the default behavior for a particular HTTP method.
	// The key is the method name, and the value is the behavior to use for that method.
	RedirectMethodBehavior map[string]RedirectBehavior

	// PathSource determines from where the router gets its path to search.
	// By default it pulls the data from the RequestURI member, but this can
	// be overridden to use URL.Path instead.
	//
	// There is a small tradeoff here. Using RequestURI allows the router to handle
	// encoded slashes (i.e. %2f) in the URL properly, while URL.Path provides
	// better compatibility with some utility functions in the http
	// library that modify the Request before passing it to the router.
	PathSource PathSource
}

// Dump returns a text representation of the routing tree.
func (t *TreeMux) Dump() string {
	return t.root.dumpTree("", "")
}

// Path elements starting with : indicate a wildcard in the path. A wildcard will only match on a
// single path segment. That is, the pattern `/post/:postid` will match on `/post/1` or `/post/1/`,
// but not `/post/1/2`.
//
// A path element starting with * is a catch-all, whose value will be a string containing all text
// in the URL matched by the wildcards. For example, with a pattern of `/images/*path` and a
// requested URL `images/abc/def`, path would contain `abc/def`.
//
// # Routing Rule Priority
//
// The priority rules in the router are simple.
//
// 1. Static path segments take the highest priority. If a segment and its subtree are able to match the URL, that match is returned.
//
// 2. Wildcards take second priority. For a particular wildcard to match, that wildcard and its subtree must match the URL.
//
// 3. Finally, a catch-all rule will match when the earlier path segments have matched, and none of the static or wildcard conditions have matched. Catch-all rules must be at the end of a pattern.
//
// So with the following patterns, we'll see certain matches:
//	 router = httptreemux.New()
//	 router.GET("/:page", pageHandler)
//	 router.GET("/:year/:month/:post", postHandler)
//	 router.GET("/:year/:month", archiveHandler)
//	 router.GET("/images/*path", staticHandler)
//	 router.GET("/favicon.ico", staticHandler)
//
//	 /abc will match /:page
//	 /2014/05 will match /:year/:month
//	 /2014/05/really-great-blog-post will match /:year/:month/:post
//	 /images/CoolImage.gif will match /images/*path
//	 /images/2014/05/MayImage.jpg will also match /images/*path, with all the text after /images stored in the variable path.
//	 /favicon.ico will match /favicon.ico
//
// # Trailing Slashes
//
// The router has special handling for paths with trailing slashes. If a pattern is added to the
// router with a trailing slash, any matches on that pattern without a trailing slash will be
// redirected to the version with the slash. If a pattern does not have a trailing slash, matches on
// that pattern with a trailing slash will be redirected to the version without.
//
// The trailing slash flag is only stored once for a pattern. That is, if a pattern is added for a
// method with a trailing slash, all other methods for that pattern will also be considered to have a
// trailing slash, regardless of whether or not it is specified for those methods too.
//
// This behavior can be turned off by setting TreeMux.RedirectTrailingSlash to false. By
// default it is set to true. The specifics of the redirect depend on RedirectBehavior.
//
// One exception to this rule is catch-all patterns. By default, trailing slash redirection is
// disabled on catch-all patterns, since the structure of the entire URL and the desired patterns
// can not be predicted. If trailing slash removal is desired on catch-all patterns, set
// TreeMux.RemoveCatchAllTrailingSlash to true.
//
// 	router = httptreemux.New()
// 	router.GET("/about", pageHandler)
// 	router.GET("/posts/", postIndexHandler)
// 	router.POST("/posts", postFormHandler)
//
// 	GET /about will match normally.
// 	GET /about/ will redirect to /about.
// 	GET /posts will redirect to /posts/.
// 	GET /posts/ will match normally.
// 	POST /posts will redirect to /posts/, because the GET method used a trailing slash.
func (t *TreeMux) Handle(method, path string, handler HandlerFunc) {
	if path[0] != '/' {
		panic(fmt.Sprintf("Path %s must start with slash", path))
	}

	addSlash := false
	if len(path) > 1 && path[len(path)-1] == '/' && t.RedirectTrailingSlash {
		addSlash = true
		path = path[:len(path)-1]
	}

	node := t.root.addPath(path[1:], nil)
	if addSlash {
		node.addSlash = true
	}
	node.setHandler(method, handler)
}

// Syntactic sugar for Handle("GET", path, handler)
func (t *TreeMux) GET(path string, handler HandlerFunc) {
	t.Handle("GET", path, handler)
}

// Syntactic sugar for Handle("POST", path, handler)
func (t *TreeMux) POST(path string, handler HandlerFunc) {
	t.Handle("POST", path, handler)
}

// Syntactic sugar for Handle("PUT", path, handler)
func (t *TreeMux) PUT(path string, handler HandlerFunc) {
	t.Handle("PUT", path, handler)
}

// Syntactic sugar for Handle("DELETE", path, handler)
func (t *TreeMux) DELETE(path string, handler HandlerFunc) {
	t.Handle("DELETE", path, handler)
}

// Syntactic sugar for Handle("PATCH", path, handler)
func (t *TreeMux) PATCH(path string, handler HandlerFunc) {
	t.Handle("PATCH", path, handler)
}

// Syntactic sugar for Handle("HEAD", path, handler)
func (t *TreeMux) HEAD(path string, handler HandlerFunc) {
	t.Handle("HEAD", path, handler)
}

// Syntactic sugar for Handle("OPTIONS", path, handler)
func (t *TreeMux) OPTIONS(path string, handler HandlerFunc) {
	t.Handle("OPTIONS", path, handler)
}

func (t *TreeMux) serveHTTPPanic(w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		t.PanicHandler(w, r, err)
	}
}

func (t *TreeMux) redirectStatusCode(method string) (int, bool) {
	var behavior RedirectBehavior
	var ok bool
	if behavior, ok = t.RedirectMethodBehavior[method]; !ok {
		behavior = t.RedirectBehavior
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

func (t *TreeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if t.PanicHandler != nil {
		defer t.serveHTTPPanic(w, r)
	}

	path := r.RequestURI
	pathLen := len(path)
	if pathLen > 0 && t.PathSource == RequestURI {
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
	if trailingSlash && t.RedirectTrailingSlash {
		path = path[:pathLen-1]
	}
	n, params := t.root.search(path[1:])
	if n == nil {
		if t.RedirectCleanPath {
			// Path was not found. Try cleaning it up and search again.
			// TODO Test this
			cleanPath := httppath.Clean(path)
			n, params = t.root.search(cleanPath[1:])
			if n == nil {
				// Still nothing found.
				t.NotFoundHandler(w, r)
				return
			} else {
				if statusCode, ok := t.redirectStatusCode(r.Method); ok {
					// Redirect to the actual path
					http.Redirect(w, r, cleanPath, statusCode)
					return
				}
			}
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

	if !n.isCatchAll || t.RemoveCatchAllTrailingSlash {
		if trailingSlash != n.addSlash && t.RedirectTrailingSlash {
			if statusCode, ok := t.redirectStatusCode(r.Method); ok {
				if n.addSlash {
					// Need to add a slash.
					http.Redirect(w, r, path+"/", statusCode)
				} else if path != "/" {
					// We need to remove the slash. This was already done at the
					// beginning of the function.
					http.Redirect(w, r, path, statusCode)
				}
				return
			}
		}
	}

	var paramMap map[string]string
	if len(params) != 0 {
		if len(params) != len(n.leafWildcardNames) {
			// Need better behavior here. Should this be a panic?
			panic(fmt.Sprintf("httptreemux parameter list length mismatch: %v, %v",
				params, n.leafWildcardNames))
		}

		paramMap = make(map[string]string)
		numParams := len(params)
		for index := 0; index < numParams; index++ {
			paramMap[n.leafWildcardNames[numParams-index-1]] = params[index]
		}
	}

	handler(w, r, paramMap)
}

// MethodNotAllowedHandler is the default handler for TreeMux.MethodNotAllowedHandler,
// which is called for patterns that match, but do not have a handler installed for the
// requested method. It simply writes the status code http.StatusMethodNotAllowed.
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
		RedirectBehavior:        Redirect301,
		RedirectMethodBehavior:  make(map[string]RedirectBehavior),
		PathSource:              RequestURI,
	}
}
