package bunrouterotel

import (
	"net"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/uptrace/bunrouter"
)

type config struct {
	clientIP bool
}

type Option func(c *config)

func WithClientIP(on bool) Option {
	return func(c *config) {
		c.clientIP = on
	}
}

func NewMiddleware(opts ...Option) bunrouter.MiddlewareFunc {
	c := &config{
		clientIP: true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c.Middleware
}

func (c *config) Middleware(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		span := trace.SpanFromContext(req.Context())
		if !span.IsRecording() {
			return next(w, req)
		}

		params := req.Params().Slice()
		attrs := make([]attribute.KeyValue, 0, 2+len(params))
		attrs = append(attrs, semconv.HTTPRouteKey.String(req.Route()))
		if c.clientIP {
			attrs = append(attrs, semconv.HTTPClientIPKey.String(remoteAddr(req.Request)))
		}

		for _, param := range params {
			attrs = append(attrs, attribute.String("http.route.param."+param.Key, param.Value))
		}

		span.SetAttributes(attrs...)

		if err := next(w, req); err != nil {
			span.SetStatus(codes.Error, err.Error())
			return err
		}

		return nil
	}
}

func remoteAddr(req *http.Request) string {
	forwarded := req.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	host, _, _ := net.SplitHostPort(req.RemoteAddr)
	return host
}
