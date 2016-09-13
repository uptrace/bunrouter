// +build go1.7

package httptreemux

import (
	"context"
	"net/http"
)

type ContextTreeMux struct {
	treemux       *TreeMux // cannot embed because TreeMux embeds Group, which prevents us from "overriding" Handle, which is also the reason we need a ContextTreeMux at all...
	*ContextGroup          // embed because we want method inheritance
}

func (t *TreeMux) UsingContext() *ContextTreeMux {
	return &ContextTreeMux{treemux: t, ContextGroup: &ContextGroup{group: &(t.Group)}}
}

func (ct *ContextTreeMux) TreeMux() *TreeMux {
	return ct.treemux
}

func (ct *ContextTreeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ct.TreeMux().ServeHTTP(w, r)
}

type ContextGroup struct {
	group *Group // cannot embed Group because we want to "steal" its method signatures
}

func (g *Group) UsingContext() *ContextGroup {
	return &ContextGroup{g}
}

func (cg *ContextGroup) NewContextGroup(path string) *ContextGroup {
	return &ContextGroup{group: cg.Group().NewGroup(path)}
}

func (cg *ContextGroup) Group() *Group {
	return cg.group
}

// Handle allows handling HTTP requests via an http.HandlerFunc, as opposed to an httptreemux.HandlerFunc.
func (cg *ContextGroup) Handle(method, path string, handler http.HandlerFunc) {
	cg.Group().Handle(method, path, func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		if params != nil {
			r = r.WithContext(context.WithValue(r.Context(), ParamsContextKey, params))
		}
		handler(w, r)
	})
}

// Syntactic sugar for Handle("GET", path, handler)
func (cg *ContextGroup) GET(path string, handler http.HandlerFunc) {
	cg.Handle("GET", path, handler)
}

// Syntactic sugar for Handle("POST", path, handler)
func (cg *ContextGroup) POST(path string, handler http.HandlerFunc) {
	cg.Handle("POST", path, handler)
}

// Syntactic sugar for Handle("PUT", path, handler)
func (cg *ContextGroup) PUT(path string, handler http.HandlerFunc) {
	cg.Handle("PUT", path, handler)
}

// Syntactic sugar for Handle("DELETE", path, handler)
func (cg *ContextGroup) DELETE(path string, handler http.HandlerFunc) {
	cg.Handle("DELETE", path, handler)
}

// Syntactic sugar for Handle("PATCH", path, handler)
func (cg *ContextGroup) PATCH(path string, handler http.HandlerFunc) {
	cg.Handle("PATCH", path, handler)
}

// Syntactic sugar for Handle("HEAD", path, handler)
func (cg *ContextGroup) HEAD(path string, handler http.HandlerFunc) {
	cg.Handle("HEAD", path, handler)
}

// Syntactic sugar for Handle("OPTIONS", path, handler)
func (cg *ContextGroup) OPTIONS(path string, handler http.HandlerFunc) {
	cg.Handle("OPTIONS", path, handler)
}

// ContextParams returns the params map associated with the given context if one exists. Otherwise, an empty map is returned.
func ContextParams(ctx context.Context) map[string]string {
	if p, ok := ctx.Value(ParamsContextKey).(map[string]string); ok {
		return p
	}
	return map[string]string{}
}

// ParamsContextKey is used to retrieve a path's params map from a request's context.
const ParamsContextKey = "params.context.key"
