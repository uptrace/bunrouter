package bunrouter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type routeCtxKey struct{}

// ParamsFromContext retrieves route parameters from the given context.
// It returns an empty Params if no parameters are found.
func ParamsFromContext(ctx context.Context) Params {
	if ctx == nil {
		return Params{}
	}
	route, _ := ctx.Value(routeCtxKey{}).(Params)
	return route
}

// contextWithParams stores route parameters in the context.
func contextWithParams(ctx context.Context, params Params) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, routeCtxKey{}, params)
}

//------------------------------------------------------------------------------

// HTTPHandler converts http.Handler to bunrouter.HandlerFunc.
func HTTPHandler(handler http.Handler) HandlerFunc {
	if handler == nil {
		panic("bunrouter: nil handler")
	}
	return HTTPHandlerFunc(handler.ServeHTTP)
}

// HTTPHandlerFunc converts http.HandlerFunc to bunrouter.HandlerFunc.
func HTTPHandlerFunc(handler http.HandlerFunc) HandlerFunc {
	if handler == nil {
		panic("bunrouter: nil handler")
	}

	return func(w http.ResponseWriter, req Request) (err error) {
		if w == nil {
			return fmt.Errorf("bunrouter: nil response writer")
		}

		ctx := contextWithParams(req.Context(), req.params)

		defer func() {
			if v := recover(); v != nil {
				var ok bool
				err, ok = v.(error)
				if !ok {
					err = fmt.Errorf("bunrouter: panic recovered: %v", v)
				}
			}
		}()

		handler.ServeHTTP(w, req.Request.WithContext(ctx))

		return err
	}
}

// HandlerFunc is a function that handles HTTP requests in bunrouter.
// It returns an error that will be handled by the router.
type HandlerFunc func(w http.ResponseWriter, req Request) error

var _ http.Handler = (*HandlerFunc)(nil)

func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if h == nil {
		http.Error(w, "Handler not found", http.StatusInternalServerError)
		return
	}

	if req == nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h(w, NewRequest(req)); err != nil {
		code := http.StatusInternalServerError
		if httpErr, ok := err.(HTTPError); ok {
			code = httpErr.StatusCode()
		}
		http.Error(w, err.Error(), code)
	}
}

// HTTPError represents an HTTP error with a status code
type HTTPError interface {
	error
	StatusCode() int
}

// MiddlewareFunc is a function that wraps a HandlerFunc to provide middleware functionality.
type MiddlewareFunc func(next HandlerFunc) HandlerFunc

//------------------------------------------------------------------------------

// Request extends http.Request with route parameters.
type Request struct {
	*http.Request
	params Params
}

// NewRequest creates a new Request instance from an http.Request.
func NewRequest(req *http.Request) Request {
	if req == nil {
		req = &http.Request{}
	}
	return Request{
		Request: req,
		params:  ParamsFromContext(req.Context()),
	}
}

func newRequestParams(req *http.Request, params Params) Request {
	if req == nil {
		req = &http.Request{}
	}
	return Request{
		Request: req,
		params:  params,
	}
}

// WithContext returns a new Request with the provided context.
func (req Request) WithContext(ctx context.Context) Request {
	if ctx == nil {
		ctx = context.Background()
	}
	return Request{
		Request: req.Request.WithContext(ctx),
		params:  req.params,
	}
}

// Params returns the route parameters associated with the request.
func (req Request) Params() Params {
	return req.params
}

// Param returns the value of the named parameter or empty string if not found.
func (req Request) Param(key string) string {
	return req.Params().ByName(key)
}

// Route returns the matched route pattern.
func (req Request) Route() string {
	return req.Params().Route()
}

//------------------------------------------------------------------------------

// Params holds route parameters and route information.
type Params struct {
	path        string
	node        *node
	handler     *routeHandler
	wildcardLen uint16
}

// IsZero returns true if Params has no associated route node.
func (ps Params) IsZero() bool {
	return ps.node == nil
}

// Route returns the route pattern that matched the request.
func (ps Params) Route() string {
	if ps.node != nil {
		return ps.node.route
	}
	return ""
}

// Get returns the value of the named parameter and whether it was found.
func (ps Params) Get(name string) (string, bool) {
	if ps.node == nil || ps.handler == nil {
		return "", false
	}
	if i, ok := ps.handler.params[name]; ok {
		return ps.findParam(i)
	}
	return "", false
}

func (ps *Params) findParam(paramIndex int) (string, bool) {
	if ps.node == nil || ps.handler == nil {
		return "", false
	}

	path := ps.path
	pathLen := len(path)
	if pathLen == 0 {
		return "", false
	}

	currNode := ps.node
	currParamIndex := len(ps.handler.params) - 1

	if paramIndex < 0 || paramIndex > currParamIndex {
		return "", false
	}

	// Wildcard can be only in the final node.
	if ps.node.isWC {
		if currParamIndex == paramIndex {
			if int(ps.wildcardLen) > pathLen {
				return "", false
			}
			pathLen -= int(ps.wildcardLen)
			return path[pathLen:], true
		}

		currParamIndex--
		if int(ps.wildcardLen) > pathLen {
			return "", false
		}
		pathLen -= int(ps.wildcardLen)
		path = path[:pathLen]
	}

	for currNode != nil {
		if currNode.part[0] != ':' { // static node
			partLen := len(currNode.part)
			if partLen > pathLen {
				return "", false
			}
			pathLen -= partLen
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

		currParamIndex--
		path = path[:pathLen]
		currNode = currNode.parent
	}

	return "", false
}

// ByName returns the value of the named parameter or empty string if not found.
func (ps Params) ByName(name string) string {
	s, _ := ps.Get(name)
	return s
}

// Int parses the named parameter as an integer.
func (ps Params) Int(name string) (int, error) {
	value := ps.ByName(name)
	if value == "" {
		return 0, fmt.Errorf("bunrouter: param '%s' not found", name)
	}
	return strconv.Atoi(value)
}

// Uint32 parses the named parameter as an unsigned 32-bit integer.
func (ps Params) Uint32(name string) (uint32, error) {
	value := ps.ByName(name)
	if value == "" {
		return 0, fmt.Errorf("bunrouter: param '%s' not found", name)
	}
	n, err := strconv.ParseUint(value, 10, 32)
	return uint32(n), err
}

// Uint64 parses the named parameter as an unsigned 64-bit integer.
func (ps Params) Uint64(name string) (uint64, error) {
	value := ps.ByName(name)
	if value == "" {
		return 0, fmt.Errorf("bunrouter: param '%s' not found", name)
	}
	return strconv.ParseUint(value, 10, 64)
}

// Int32 parses the named parameter as a signed 32-bit integer.
func (ps Params) Int32(name string) (int32, error) {
	value := ps.ByName(name)
	if value == "" {
		return 0, fmt.Errorf("bunrouter: param '%s' not found", name)
	}
	n, err := strconv.ParseInt(value, 10, 32)
	return int32(n), err
}

// Int64 parses the named parameter as a signed 64-bit integer.
func (ps Params) Int64(name string) (int64, error) {
	value := ps.ByName(name)
	if value == "" {
		return 0, fmt.Errorf("bunrouter: param '%s' not found", name)
	}
	return strconv.ParseInt(value, 10, 64)
}

// Map returns route parameters as a map[string]string.
func (ps Params) Map() map[string]string {
	if ps.handler == nil || len(ps.handler.params) == 0 {
		return make(map[string]string)
	}
	m := make(map[string]string, len(ps.handler.params))
	for param, index := range ps.handler.params {
		if value, ok := ps.findParam(index); ok {
			m[param] = value
		}
	}
	return m
}

// Param represents a key-value pair of route parameters.
type Param struct {
	Key   string
	Value string
}

// Slice returns route parameters as a slice of Param.
func (ps Params) Slice() []Param {
	if ps.handler == nil || len(ps.handler.params) == 0 {
		return []Param{}
	}
	slice := make([]Param, len(ps.handler.params))
	for param, index := range ps.handler.params {
		if value, ok := ps.findParam(index); ok {
			slice[index] = Param{Key: param, Value: value}
		}
	}
	return slice
}

//------------------------------------------------------------------------------

// H is a shorthand for map[string]interface{}.
type H map[string]interface{}

// JSON marshals the value as JSON and writes it to the response writer.
// It sets the Content-Type header to application/json.
//
// Don't hesitate to copy-paste this function to your project and customize it as necessary.
func JSON(w http.ResponseWriter, value interface{}) error {
	if w == nil {
		return fmt.Errorf("bunrouter: nil response writer")
	}

	w.Header().Set("Content-Type", "application/json")

	if value == nil {
		w.Write([]byte("null"))
		return nil
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(value); err != nil {
		return fmt.Errorf("bunrouter: JSON encoding error: %w", err)
	}

	return nil
}
