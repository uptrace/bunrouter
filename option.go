package treemux

type config struct {
	notFoundHandler         HandlerFunc
	methodNotAllowedHandler HandlerFunc

	headCanUseGet               bool
	redirectCleanPath           bool
	redirectTrailingSlash       bool
	removeCatchAllTrailingSlash bool

	redirectBehavior       RedirectBehavior
	redirectMethodBehavior map[string]RedirectBehavior

	useURLPath bool

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

// MethodNotAllowedHandler is called when a pattern matches, but that
// pattern does not have a handler for the requested method. The default
// handler just writes the status code http.StatusMethodNotAllowed.
func WithMethodNotAllowedHandler(handler HandlerFunc) Option {
	return option(func(c *config) {
		c.methodNotAllowedHandler = c.wrapHandler(handler)
	})
}

// WithHeadCanUseGet allows the router to use the GET handler to respond to
// HEAD requests if no explicit HEAD handler has been added for the
// matching pattern. This is true by default.
func WithHeadCanUseGet(on bool) Option {
	return option(func(c *config) {
		c.headCanUseGet = on
	})
}

// WithRedirectCleanPath allows the router to try clean the current request path,
// if no handler is registered for it, using CleanPath from github.com/dimfeld/httppath.
// This is true by default.
func WithRedirectCleanPath(on bool) Option {
	return option(func(c *config) {
		c.redirectCleanPath = on
	})
}

// WithRedirectTrailingSlash enables automatic redirection in case router doesn't find
// a matching route for the current request path but a handler for the path with or
// without the trailing slash exists. This is true by default.
func WithRedirectTrailingSlash(on bool) Option {
	return option(func(c *config) {
		c.redirectTrailingSlash = on
	})
}

// WithRemoveCatchAllTrailingSlash removes the trailing slash when a catch-all pattern
// is matched, if set to true. By default, catch-all paths are never redirected.
func WithRemoveCatchAllTrailingSlash(on bool) Option {
	return option(func(c *config) {
		c.removeCatchAllTrailingSlash = on
	})
}

// WithRedirectBehavior sets the default redirect behavior when RedirectTrailingSlash or
// RedirectCleanPath are true. The default value is Redirect301.
func WithRedirectBehavior(value RedirectBehavior) Option {
	return option(func(c *config) {
		c.redirectBehavior = value
	})
}

// WithRedirectMethodBehavior overrides the default behavior for a particular HTTP method.
// The key is the method name, and the value is the behavior to use for that method.
func WithRedirectMethodBehavior(value map[string]RedirectBehavior) Option {
	return option(func(c *config) {
		c.redirectMethodBehavior = value
	})
}

// UseURLPath determines from where the router gets its path to search.
// By default it pulls the data from the RequestURI member, but this can
// be overridden to use URL.Path instead.
//
// There is a small tradeoff here. Using RequestURI allows the router to handle
// encoded slashes (i.e. %2f) in the URL properly, while URL.Path provides
// better compatibility with some utility functions in the http
// library that modify the Request before passing it to the router.
func UseURLPath() Option {
	return option(func(c *config) {
		c.useURLPath = true
	})
}
