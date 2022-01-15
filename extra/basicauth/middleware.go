package basicauth

import (
	"net/http"
	"strconv"

	"github.com/uptrace/bunrouter"
)

type middleware struct {
	realm string
	check func(req bunrouter.Request) (bool, error)
}

type Option func(m *middleware)

func WithRealm(realm string) Option {
	return func(m *middleware) {
		m.realm = strconv.Quote(realm)
	}
}

func NewMiddleware(
	check func(req bunrouter.Request) (bool, error),
	opts ...Option,
) bunrouter.MiddlewareFunc {
	c := &middleware{
		realm: "Restricted",
		check: check,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c.Middleware
}

func (m *middleware) Middleware(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		ok, err := m.check(req)
		if err != nil {
			return err
		}
		if ok {
			return next(w, req)
		}

		return m.basicAuth(w, req)
	}
}

func (m *middleware) basicAuth(w http.ResponseWriter, req bunrouter.Request) error {
	w.Header().Set("WWW-Authenticate", "basic realm="+m.realm)
	w.WriteHeader(http.StatusUnauthorized)
	return nil
}
