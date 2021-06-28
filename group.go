package treemux

import (
	"fmt"
	"net/http"
)

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

// Group is a group of routes and middlewares.
type Group struct {
	mux   *Router
	path  string
	stack []MiddlewareFunc
}

// NewGroup adds a sub-group to this group.
func (g *Group) NewGroup(path string, opts ...Option) *Group {
	group := &Group{
		mux:   g.mux,
		path:  joinPath(g.path, path),
		stack: g.cloneStack(),
	}

	cfg := &config{
		group: group,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	return group
}

func (g *Group) cloneStack() []MiddlewareFunc {
	return g.stack[:len(g.stack):len(g.stack)]
}

func (g *Group) Compat() *CompatGroup {
	return &CompatGroup{group: g}
}

func (g *Group) WithMiddleware(middleware MiddlewareFunc) *Group {
	return g.NewGroup("", WithMiddleware(middleware))
}

func (g *Group) WithGroup(path string, fn func(g *Group)) {
	fn(g.NewGroup(path))
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
//	 router = treemux.New()
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
// 	router = treemux.New()
// 	router.GET("/about", pageHandler)
// 	router.GET("/posts/", postIndexHandler)
// 	router.POST("/posts", postFormHandler)
//
// 	GET /about will match normally.
// 	GET /about/ will redirect to /about.
// 	GET /posts will redirect to /posts/.
// 	GET /posts/ will match normally.
// 	POST /posts will redirect to /posts/, because the GET method used a trailing slash.
func (g *Group) Handle(method string, path string, handler HandlerFunc) {
	g.mux.mu.Lock()
	defer g.mux.mu.Unlock()

	if len(g.stack) > 0 {
		handler = g.handlerWithMiddlewares(handler)
	}

	var addSlash bool
	addOne := func(fullPath string) {
		node := g.mux.root.addPath(fullPath[1:], nil, false)
		if node.route == "" {
			node.route = fullPath
		} else if node.route != fullPath {
			panic(fmt.Errorf("%q != %q", node.route, fullPath))
		}
		if addSlash {
			node.addSlash = true
		}
		node.setHandler(method, handler, false)

		if g.mux.headCanUseGet &&
			method == http.MethodGet &&
			node.handlerMap.Get(http.MethodHead) == nil {
			node.setHandler(http.MethodHead, handler, true)
		}
	}

	checkPath(path)
	path = g.path + path
	if path == "" {
		panic("Cannot map an empty path")
	}

	if len(path) > 1 && path[len(path)-1] == '/' && g.mux.redirectTrailingSlash {
		addSlash = true
		path = path[:len(path)-1]
	}

	addOne(path)
}

// Syntactic sugar for Handle("GET", path, handler)
func (g *Group) GET(path string, handler HandlerFunc) {
	g.Handle("GET", path, handler)
}

// Syntactic sugar for Handle("POST", path, handler)
func (g *Group) POST(path string, handler HandlerFunc) {
	g.Handle("POST", path, handler)
}

// Syntactic sugar for Handle("PUT", path, handler)
func (g *Group) PUT(path string, handler HandlerFunc) {
	g.Handle("PUT", path, handler)
}

// Syntactic sugar for Handle("DELETE", path, handler)
func (g *Group) DELETE(path string, handler HandlerFunc) {
	g.Handle("DELETE", path, handler)
}

// Syntactic sugar for Handle("PATCH", path, handler)
func (g *Group) PATCH(path string, handler HandlerFunc) {
	g.Handle("PATCH", path, handler)
}

// Syntactic sugar for Handle("HEAD", path, handler)
func (g *Group) HEAD(path string, handler HandlerFunc) {
	g.Handle("HEAD", path, handler)
}

// Syntactic sugar for Handle("OPTIONS", path, handler)
func (g *Group) OPTIONS(path string, handler HandlerFunc) {
	g.Handle("OPTIONS", path, handler)
}

func (g *Group) handlerWithMiddlewares(handler HandlerFunc) HandlerFunc {
	for i := len(g.stack) - 1; i >= 0; i-- {
		handler = g.stack[i](handler)
	}
	return handler
}

//------------------------------------------------------------------------------

type CompatGroup struct {
	group *Group
}

func (g CompatGroup) NewGroup(path string, opts ...Option) *CompatGroup {
	return &CompatGroup{group: g.group.NewGroup(path, opts...)}
}

func (g CompatGroup) WithMiddleware(middleware MiddlewareFunc) *CompatGroup {
	return &CompatGroup{group: g.group.WithMiddleware(middleware)}
}

func (g CompatGroup) WithGroup(path string, fn func(g *CompatGroup)) {
	fn(g.NewGroup(path))
}

func (g CompatGroup) Handle(method string, path string, handler http.HandlerFunc) {
	g.group.Handle(method, path, HTTPHandler(handler))
}

func (g CompatGroup) GET(path string, handler http.HandlerFunc) {
	g.Handle(http.MethodGet, path, handler)
}

func (g CompatGroup) POST(path string, handler http.HandlerFunc) {
	g.Handle("POST", path, handler)
}

func (g CompatGroup) PUT(path string, handler http.HandlerFunc) {
	g.Handle("PUT", path, handler)
}

func (g CompatGroup) DELETE(path string, handler http.HandlerFunc) {
	g.Handle("DELETE", path, handler)
}

func (g CompatGroup) PATCH(path string, handler http.HandlerFunc) {
	g.Handle("PATCH", path, handler)
}

func (g CompatGroup) HEAD(path string, handler http.HandlerFunc) {
	g.Handle("HEAD", path, handler)
}

func (g CompatGroup) OPTIONS(path string, handler http.HandlerFunc) {
	g.Handle("OPTIONS", path, handler)
}

func (g CompatGroup) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	g.group.mux.ServeHTTP(w, req)
}

//------------------------------------------------------------------------------

func joinPath(base, path string) string {
	checkPath(path)
	path = base + path
	// Don't want trailing slash as all sub-paths start with slash
	if len(path) > 0 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return path
}

func checkPath(path string) {
	// All non-empty paths must start with a slash
	if len(path) > 0 && path[0] != '/' {
		panic(fmt.Sprintf("Path %s must start with slash", path))
	}
}
