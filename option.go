package treemux

import "net/http"

type Option func(*TreeMux)

// WithErrorHandler handles errors returned from handlers.
func WithErrorHandler(handler func(w http.ResponseWriter, req Request, err error)) Option {
	return func(t *TreeMux) {
		t.errorHandler = handler
	}
}

// WithNotFoundHandler is called when there is no a matching pattern.
// The default NotFoundHandler is http.NotFound.
func WithNotFoundHandler(handler HandlerFunc) Option {
	return func(t *TreeMux) {
		t.notFoundHandler = handler
	}
}

// MethodNotAllowedHandler is called when a pattern matches, but that
// pattern does not have a handler for the requested method. The default
// handler just writes the status code http.StatusMethodNotAllowed and adds
// the required Allowed header.
func WithMethodNotAllowedHandler(handler HandlerFunc) Option {
	return func(c *TreeMux) {
		c.methodNotAllowedHandler = handler
	}
}

// WithHeadCanUseGet allows the router to use the GET handler to respond to
// HEAD requests if no explicit HEAD handler has been added for the
// matching pattern. This is true by default.
func WithHeadCanUseGet(on bool) Option {
	return func(t *TreeMux) {
		t.headCanUseGet = on
	}
}

// WithRedirectCleanPath allows the router to try clean the current request path,
// if no handler is registered for it, using CleanPath from github.com/dimfeld/httppath.
// This is true by default.
func WithRedirectCleanPath(on bool) Option {
	return func(t *TreeMux) {
		t.redirectCleanPath = on
	}
}

// WithRedirectTrailingSlash enables automatic redirection in case router doesn't find a matching route
// for the current request path but a handler for the path with or without the trailing
// slash exists. This is true by default.
func WithRedirectTrailingSlash(on bool) Option {
	return func(c *TreeMux) {
		c.redirectTrailingSlash = on
	}
}

// WithRemoveCatchAllTrailingSlash removes the trailing slash when a catch-all pattern
// is matched, if set to true. By default, catch-all paths are never redirected.
func WithRemoveCatchAllTrailingSlash(on bool) Option {
	return func(t *TreeMux) {
		t.removeCatchAllTrailingSlash = on
	}
}

// WithRedirectBehavior sets the default redirect behavior when RedirectTrailingSlash or
// RedirectCleanPath are true. The default value is Redirect301.
func WithRedirectBehavior(value RedirectBehavior) Option {
	return func(t *TreeMux) {
		t.redirectBehavior = value
	}
}

// WithRedirectMethodBehavior overrides the default behavior for a particular HTTP method.
// The key is the method name, and the value is the behavior to use for that method.
func WithRedirectMethodBehavior(value map[string]RedirectBehavior) Option {
	return func(t *TreeMux) {
		t.redirectMethodBehavior = value
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
	return func(t *TreeMux) {
		t.pathSource = value
	}
}
