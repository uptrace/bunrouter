package treemuxotel

import (
	"net"
	"net/http"

	"github.com/vmihailenco/treemux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
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

func NewMiddleware(opts ...Option) treemux.MiddlewareFunc {
	c := &config{
		clientIP: true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c.Middleware
}

func (c *config) Middleware(next treemux.HandlerFunc) treemux.HandlerFunc {
	return func(w http.ResponseWriter, req treemux.Request) error {
		span := trace.SpanFromContext(req.Context())
		if !span.IsRecording() {
			return next(w, req)
		}

		attrs := make([]attribute.KeyValue, 0, 2+len(req.Params))
		attrs = append(attrs, semconv.HTTPRouteKey.String(req.Route()))
		if c.clientIP {
			attrs = append(attrs, semconv.HTTPClientIPKey.String(remoteAddr(req.Request)))
		}

		for _, param := range req.Params {
			name := param.Name
			if name == "" {
				name = "*"
			}

			attrs = append(attrs, attribute.String("http.route.param."+name, param.Value))
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
