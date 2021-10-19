package bunrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type routeCtxKey struct{}

func ParamsFromContext(ctx context.Context) Params {
	route, _ := ctx.Value(routeCtxKey{}).(Params)
	return route
}

func contextWithParams(ctx context.Context, params Params) context.Context {
	return context.WithValue(ctx, routeCtxKey{}, params)
}

//------------------------------------------------------------------------------

// HTTPHandler converts http.Handler from the stdlib to bunrouter.HandlerFunc.
func HTTPHandler(handler http.Handler) HandlerFunc {
	return func(w http.ResponseWriter, req Request) error {
		ctx := contextWithParams(req.Context(), req.params)
		handler.ServeHTTP(w, req.Request.WithContext(ctx))
		return nil
	}
}

// HTTPHandlerFunc converts http.HandlerFunc from the stdlib to bunrouter.HandlerFunc.
func HTTPHandlerFunc(handler http.HandlerFunc) HandlerFunc {
	return HTTPHandler(http.HandlerFunc(handler))
}

type HandlerFunc func(w http.ResponseWriter, req Request) error

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

//------------------------------------------------------------------------------

type Request struct {
	*http.Request
	params Params
}

func NewRequest(req *http.Request) Request {
	return Request{
		Request: req,
	}
}

func (req Request) WithContext(ctx context.Context) Request {
	return Request{
		Request: req.Request.WithContext(ctx),
		params:  req.params,
	}
}

func (req Request) Route() string {
	return req.params.Route()
}

func (req Request) Params() Params {
	return req.params
}

func (req Request) Param(key string) string {
	return req.params.ByName(key)
}

//------------------------------------------------------------------------------

type Params struct {
	path        string
	node        *node
	wildcardLen uint16
}

func (ps Params) Route() string {
	if ps.node != nil {
		return ps.node.route
	}
	return ""
}

func (ps Params) Get(name string) (string, bool) {
	if ps.node == nil {
		return "", false
	}
	if i, ok := ps.node.params[name]; ok {
		return ps.findParam(i)
	}
	return "", false
}

func (ps *Params) findParam(paramIndex int) (string, bool) {
	path := ps.path
	pathLen := len(path)
	currNode := ps.node
	currParamIndex := len(ps.node.params) - 1

	// Wildcard can be only in the final node.
	if ps.node.part == "*" {
		pathLen -= int(ps.wildcardLen)
		if currParamIndex == paramIndex {
			return path[pathLen:], true
		}

		currParamIndex--
		currNode = currNode.parent
	}

	for currNode != nil {
		if currNode.part[0] != ':' {
			pathLen -= len(currNode.part)
			path = path[:pathLen]

			currNode = currNode.parent
			continue
		}

		i := strings.LastIndexByte(path, '/')
		if i == -1 {
			return "", false
		}
		pathLen = i + 1

		if currParamIndex == paramIndex {
			return path[pathLen:], true
		}

		path = path[:pathLen]

		currParamIndex--
		currNode = currNode.parent
	}

	return "", false
}

func (ps Params) ByName(name string) string {
	s, _ := ps.Get(name)
	return s
}

func (ps Params) Uint32(name string) (uint32, error) {
	n, err := strconv.ParseUint(ps.ByName(name), 10, 32)
	return uint32(n), err
}

func (ps Params) Uint64(name string) (uint64, error) {
	return strconv.ParseUint(ps.ByName(name), 10, 64)
}

func (ps Params) Int32(name string) (int32, error) {
	n, err := strconv.ParseInt(ps.ByName(name), 10, 32)
	return int32(n), err
}

func (ps Params) Int64(name string) (int64, error) {
	return strconv.ParseInt(ps.ByName(name), 10, 64)
}

func (ps Params) Map() map[string]string {
	if ps.node == nil || len(ps.node.params) == 0 {
		return nil
	}
	m := make(map[string]string, len(ps.node.params))
	for param, index := range ps.node.params {
		if value, ok := ps.findParam(index); ok {
			m[param] = value
		}
	}
	return m
}

type Param struct {
	Key   string
	Value string
}

func (ps Params) Slice() []Param {
	if ps.node == nil || len(ps.node.params) == 0 {
		return nil
	}
	slice := make([]Param, len(ps.node.params))
	for param, index := range ps.node.params {
		if value, ok := ps.findParam(index); ok {
			slice[index] = Param{Key: param, Value: value}
		}
	}
	return slice
}

//------------------------------------------------------------------------------

type H map[string]interface{}

// JSON marshals the value as JSON and writes it to the response writer.
//
// Don't hesitate to copy-paste this function to your project and customize it as necessary.
func JSON(w http.ResponseWriter, value interface{}) error {
	if value == nil {
		return nil
	}

	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(value); err != nil {
		return err
	}

	return nil
}
