// +build go1.7

package httptreemux

import (
	"context"
	"net/http"
)

// ParamsContextKey is used to retrieve a path's params map from a request's context.
const ParamsContextKey = "params.context.key"

// HandleWithContext is a convenience method for handling HTTP requests via an http.HandlerFunc, as opposed to an httptreemux.HandlerFunc.
func (g *Group) HandleWithContext(method, path string, handler http.HandlerFunc) {
	g.Handle(method, path, func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		if params != nil {
			r = r.WithContext(context.WithValue(r.Context(), ParamsContextKey, params))
		}
		handler(w, r)
	})
}

// ContextParams returns the params map associated with the given context if one exists. Otherwise, an empty map is returned.
func ContextParams(ctx context.Context) map[string]string {
	if p, ok := ctx.Value(ParamsContextKey).(map[string]string); ok {
		return p
	}
	return map[string]string{}
}
