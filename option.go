package treemux

import "net/http"

type config struct {
	errorHandler            func(w http.ResponseWriter, req Request, err error)
	notFoundHandler         HandlerFunc
	methodNotAllowedHandler HandlerFunc

	headCanUseGet               bool
	redirectCleanPath           bool
	redirectTrailingSlash       bool
	removeCatchAllTrailingSlash bool

	redirectBehavior       RedirectBehavior
	redirectMethodBehavior map[string]RedirectBehavior

	pathSource PathSource

	group *Group
}

type Option func(*config)

// WithErrorHandler handles errors returned from handlers.
func WithErrorHandler(handler func(w http.ResponseWriter, req Request, err error)) Option {
	return func(c *config) {
		c.errorHandler = handler
	}
}

// WithNotFoundHandler is called when there is no a matching pattern.
// The default NotFoundHandler is http.NotFound.
func WithNotFoundHandler(handler HandlerFunc) Option {
	return func(c *config) {
		c.notFoundHandler = handlerWithMiddlewares(handler, c.group.stack)
	}
}

// MethodNotAllowedHandler is called when a pattern matches, but that
// pattern does not have a handler for the requested method. The default
// handler just writes the status code http.StatusMethodNotAllowed and adds
// the required Allowed header.
func WithMethodNotAllowedHandler(handler HandlerFunc) Option {
	return func(c *config) {
		c.methodNotAllowedHandler = handlerWithMiddlewares(handler, c.group.stack)
	}
}

// WithHeadCanUseGet allows the router to use the GET handler to respond to
// HEAD requests if no explicit HEAD handler has been added for the
// matching pattern. This is true by default.
func WithHeadCanUseGet(on bool) Option {
	return func(c *config) {
		c.headCanUseGet = on
	}
}

// WithRedirectCleanPath allows the router to try clean the current request path,
// if no handler is registered for it, using CleanPath from github.com/dimfeld/httppath.
// This is true by default.
func WithRedirectCleanPath(on bool) Option {
	return func(c *config) {
		c.redirectCleanPath = on
	}
}

// WithRedirectTrailingSlash enables automatic redirection in case router doesn't find a matching route
// for the current request path but a handler for the path with or without the trailing
// slash exists. This is true by default.
func WithRedirectTrailingSlash(on bool) Option {
	return func(c *config) {
		c.redirectTrailingSlash = on
	}
}

// WithRemoveCatchAllTrailingSlash removes the trailing slash when a catch-all pattern
// is matched, if set to true. By default, catch-all paths are never redirected.
func WithRemoveCatchAllTrailingSlash(on bool) Option {
	return func(c *config) {
		c.removeCatchAllTrailingSlash = on
	}
}

// WithRedirectBehavior sets the default redirect behavior when RedirectTrailingSlash or
// RedirectCleanPath are true. The default value is Redirect301.
func WithRedirectBehavior(value RedirectBehavior) Option {
	return func(c *config) {
		c.redirectBehavior = value
	}
}

// WithRedirectMethodBehavior overrides the default behavior for a particular HTTP method.
// The key is the method name, and the value is the behavior to use for that method.
func WithRedirectMethodBehavior(value map[string]RedirectBehavior) Option {
	return func(c *config) {
		c.redirectMethodBehavior = value
	}
}

// WithPathSource determines from where the router gets its path to search.
// By default it pulls the data from the RequestURI member, but this can
// be overridden to use URL.Path instead.
//
// There is a small tradeoff here. Using RequestURI allows the router to handle
// encoded slashes (i.e. %2f) in the URL properly, while URL.Path provides
// better compatibility with some utility functions in the http
// library that modify the Request before passing it to the router.
func WithPathSource(value PathSource) Option {
	return func(c *config) {
		c.pathSource = value
	}
}

// WithMiddleware adds a middleware handler to the Group's middleware stack.
func WithMiddleware(fn MiddlewareFunc) Option {
	return func(c *config) {
		c.group.stack = append(c.group.stack, fn)
	}
}

// WithHandler is like WithMiddleware, but it can't modify the request.
func WithHandler(fn HandlerFunc) Option {
	return func(c *config) {
		middleware := func(next HandlerFunc) HandlerFunc {
			return func(w http.ResponseWriter, req Request) error {
				if err := fn(w, req); err != nil {
					return err
				}
				return next(w, req)
			}
		}
		c.group.stack = append(c.group.stack, middleware)
	}
}

// WithGroup calls the fn with the current Group.
func WithGroup(fn func(g *Group)) Option {
	return func(c *config) {
		fn(c.group)
	}
}
