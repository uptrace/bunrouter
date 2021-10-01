package bunrouter

import "net/http"

type config struct {
	notFoundHandler         HandlerFunc
	methodNotAllowedHandler HandlerFunc

	group *Group
}

func (c *config) wrapHandler(handler HandlerFunc) HandlerFunc {
	return c.group.handlerWithMiddlewares(handler)
}

type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

// WithNotFoundHandler is called when there is no a matching pattern.
// The default NotFoundHandler is http.NotFound.
func WithNotFoundHandler(handler HandlerFunc) Option {
	return option(func(c *config) {
		c.notFoundHandler = c.wrapHandler(handler)
	})
}

// MethodNotAllowedHandler is called when a route matches, but that
// route does not have a handler for the requested method. The default
// handler just writes the status code http.StatusMethodNotAllowed.
func WithMethodNotAllowedHandler(handler HandlerFunc) Option {
	return option(func(c *config) {
		c.methodNotAllowedHandler = c.wrapHandler(handler)
	})
}

//------------------------------------------------------------------------------

type GroupOption interface {
	Option
	groupOption()
}

type groupOption func(cfg *config)

func (fn groupOption) apply(cfg *config) {
	fn(cfg)
}

func (fn groupOption) groupOption() {}

// WithGroup calls the fn with the current Group.
func WithGroup(fn func(g *Group)) GroupOption {
	return groupOption(func(c *config) {
		fn(c.group)
	})
}

// WithMiddleware adds a middleware handler to the Group's middleware stack.
func WithMiddleware(fn MiddlewareFunc) GroupOption {
	return groupOption(func(c *config) {
		c.group.stack = append(c.group.stack, fn)
	})
}

// WithHandler is like WithMiddleware, but the handler can't modify the request.
func WithHandler(fn HandlerFunc) GroupOption {
	return groupOption(func(c *config) {
		middleware := func(next HandlerFunc) HandlerFunc {
			return func(w http.ResponseWriter, req Request) error {
				if err := fn(w, req); err != nil {
					return err
				}
				return next(w, req)
			}
		}
		c.group.stack = append(c.group.stack, middleware)
	})
}
