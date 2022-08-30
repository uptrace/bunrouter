package bunrouter

import (
	"fmt"
	"net/http"
)

// Group is a group of routes and middlewares.
type Group struct {
	router *Router
	path   string
	stack  []MiddlewareFunc
}

// NewGroup adds a sub-group to this group.
func (g *Group) NewGroup(path string, opts ...GroupOption) *Group {
	group := &Group{
		router: g.router,
		path:   joinPath(g.path, path),
		stack:  g.cloneStack(),
	}

	cfg := &config{
		group: group,
	}
	for _, opt := range opts {
		opt.apply(cfg)
	}

	return group
}

func (g *Group) cloneStack() []MiddlewareFunc {
	return g.stack[:len(g.stack):len(g.stack)]
}

func (g *Group) Use(middlewares ...MiddlewareFunc) *Group {
	return g.NewGroup("", Use(middlewares...))
}

func (g *Group) WithMiddleware(middleware MiddlewareFunc) *Group {
	return g.NewGroup("", WithMiddleware(middleware))
}

func (g *Group) WithGroup(path string, fn func(g *Group)) {
	fn(g.NewGroup(path))
}

func (g *Group) Handle(meth string, path string, handler HandlerFunc) {
	g.router.mu.Lock()
	defer g.router.mu.Unlock()

	checkPath(path)
	path = g.path + path
	if path == "" {
		panic("path can't be empty")
	}

	node, params := g.router.tree.addRoute(path)

	if node.handlerMap != nil {
		if h := node.handlerMap.Get(meth); h != nil {
			if node.route == path {
				panic(fmt.Errorf("route %q already handles %s", node.route, meth))
			}
			panic(fmt.Errorf("routes %q and %q can't both handle %s", node.route, path, meth))
		}
	}
	node.setHandler(meth, &routeHandler{
		fn:     g.wrap(handler),
		params: params,
	})

	if node.handlerMap.notAllowed == nil {
		node.handlerMap.notAllowed = &routeHandler{
			fn:     g.wrap(g.router.methodNotAllowedHandler),
			params: params,
		}
	}
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

func (g *Group) wrap(handler HandlerFunc) HandlerFunc {
	for i := len(g.stack) - 1; i >= 0; i-- {
		handler = g.stack[i](handler)
	}
	return handler
}

func (g *Group) Compat() *CompatGroup {
	return &CompatGroup{group: g}
}

func (g *Group) Verbose() *VerboseGroup {
	return &VerboseGroup{group: g}
}

//------------------------------------------------------------------------------

// CompatGroup is like Group, but it works with http.HandlerFunc instead of bunrouter handler.
type CompatGroup struct {
	group *Group
}

func (g CompatGroup) NewGroup(path string, opts ...GroupOption) *CompatGroup {
	return &CompatGroup{group: g.group.NewGroup(path, opts...)}
}

func (g CompatGroup) WithMiddleware(middleware MiddlewareFunc) *CompatGroup {
	return &CompatGroup{group: g.group.WithMiddleware(middleware)}
}

func (g CompatGroup) WithGroup(path string, fn func(g *CompatGroup)) {
	fn(g.NewGroup(path))
}

func (g CompatGroup) Handle(method string, path string, handler http.HandlerFunc) {
	g.group.Handle(method, path, HTTPHandlerFunc(handler))
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

//------------------------------------------------------------------------------

type VerboseHandlerFunc func(w http.ResponseWriter, req *http.Request, ps Params)

// VerboseGroup is like Group, but it works with VerboseHandlerFunc instead of bunrouter handler.
type VerboseGroup struct {
	group *Group
}

func (g VerboseGroup) NewGroup(path string, opts ...GroupOption) *VerboseGroup {
	return &VerboseGroup{group: g.group.NewGroup(path, opts...)}
}

func (g VerboseGroup) WithMiddleware(middleware MiddlewareFunc) *VerboseGroup {
	return &VerboseGroup{group: g.group.WithMiddleware(middleware)}
}

func (g VerboseGroup) WithGroup(path string, fn func(g *VerboseGroup)) {
	fn(g.NewGroup(path))
}

func (g VerboseGroup) Handle(method string, path string, handler VerboseHandlerFunc) {
	g.group.Handle(method, path, func(w http.ResponseWriter, req Request) error {
		handler(w, req.Request, req.Params())
		return nil
	})
}

func (g VerboseGroup) GET(path string, handler VerboseHandlerFunc) {
	g.Handle(http.MethodGet, path, handler)
}

func (g VerboseGroup) POST(path string, handler VerboseHandlerFunc) {
	g.Handle("POST", path, handler)
}

func (g VerboseGroup) PUT(path string, handler VerboseHandlerFunc) {
	g.Handle("PUT", path, handler)
}

func (g VerboseGroup) DELETE(path string, handler VerboseHandlerFunc) {
	g.Handle("DELETE", path, handler)
}

func (g VerboseGroup) PATCH(path string, handler VerboseHandlerFunc) {
	g.Handle("PATCH", path, handler)
}

func (g VerboseGroup) HEAD(path string, handler VerboseHandlerFunc) {
	g.Handle("HEAD", path, handler)
}

func (g VerboseGroup) OPTIONS(path string, handler VerboseHandlerFunc) {
	g.Handle("OPTIONS", path, handler)
}

//------------------------------------------------------------------------------

func joinPath(base, path string) string {
	checkPath(path)
	path = base + path
	// Don't want trailing slash as all sub-paths start with slash
	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return path
}

func checkPath(path string) {
	// All non-empty paths must start with a slash
	if len(path) > 0 && path[0] != '/' {
		panic(fmt.Sprintf("path %s must start with a slash", path))
	}
}
