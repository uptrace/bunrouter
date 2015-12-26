package httptreemux

import (
	"fmt"
)

type Group struct {
	path string
	mux  *TreeMux
}

func (t *TreeMux) NewGroup(path string) *Group {
	checkPath(path)
	//Don't want trailing slash as all sub-paths start with slash
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return &Group{path, t}
}

// Add a sub-group to this group
func (g *Group) NewGroup(path string) *Group {
	checkPath(path)
	return g.mux.NewGroup(g.path + path)
}

func (g *Group) Handle(method string, path string, handler HandlerFunc) {
	checkPath(path)
	g.mux.Handle(method, g.path+path, handler)
}

func (g *Group) GET(path string, handler HandlerFunc) {
	g.Handle("GET", path, handler)
}
func (g *Group) POST(path string, handler HandlerFunc) {
	g.Handle("POST", path, handler)
}
func (g *Group) PUT(path string, handler HandlerFunc) {
	g.Handle("PUT", path, handler)
}
func (g *Group) DELETE(path string, handler HandlerFunc) {
	g.Handle("DELETE", path, handler)
}
func (g *Group) PATCH(path string, handler HandlerFunc) {
	g.Handle("PATCH", path, handler)
}
func (g *Group) HEAD(path string, handler HandlerFunc) {
	g.Handle("HEAD", path, handler)
}
func (g *Group) OPTIONS(path string, handler HandlerFunc) {
	g.Handle("OPTIONS", path, handler)
}

func checkPath(path string) {
	if path[0] != '/' {
		panic(fmt.Sprintf("Path %s must start with slash", path))
	}
}
