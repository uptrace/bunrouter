package treemuxotel

import (
	"net"
	"net/http"

	"github.com/vmihailenco/treemux"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/semconv"
)

var Middleware = New().Middleware

type Config struct {
	clientIP bool
}

type Option func(c *Config)

func WithClientIP(on bool) Option {
	return func(c *Config) {
		c.clientIP = on
	}
}

func New(opts ...Option) *Config {
	c := &Config{
		clientIP: true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (cfg *Config) Middleware(next treemux.HandlerFunc) treemux.HandlerFunc {
	return func(w http.ResponseWriter, req treemux.Request) error {
		span := trace.SpanFromContext(req.Context())
		if !span.IsRecording() {
			return next(w, req)
		}

		attrs := make([]label.KeyValue, 0, 2+len(req.Params))
		attrs = append(attrs, semconv.HTTPRouteKey.String(req.Route()))
		if cfg.clientIP {
			attrs = append(attrs, semconv.HTTPClientIPKey.String(remoteAddr(req.Request)))
		}

		for _, param := range req.Params {
			name := param.Name
			if name == "" {
				name = "*"
			}

			attrs = append(attrs, label.String("http.route.param."+name, param.Value))
		}

		span.SetAttributes(attrs...)

		return next(w, req)
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
