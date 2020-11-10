package treemuxgzip

import (
	"net/http"

	"github.com/vmihailenco/httpgzip"
	"github.com/vmihailenco/treemux"
)

const (
	vary           = "Vary"
	acceptEncoding = "Accept-Encoding"
)

var Middleware = New().Middleware

type Config struct {
	compressionLevel int
	contentTypes     []string
}

type Option func(c *Config)

func CompressionLevel(level int) Option {
	return func(c *Config) {
		c.compressionLevel = level
	}
}

func ContentTypes(contentTypes ...string) Option {
	return func(c *Config) {
		c.contentTypes = contentTypes
	}
}

func New(opts ...Option) *Config {
	c := &Config{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (cfg *Config) Middleware(next treemux.HandlerFunc) treemux.HandlerFunc {
	var opts []httpgzip.Option
	if cfg.compressionLevel != 0 {
		opts = append(opts, httpgzip.CompressionLevel(cfg.compressionLevel))
	}
	if cfg.contentTypes != nil {
		opts = append(opts, httpgzip.ContentTypes(cfg.contentTypes))
	}
	hgz, err := httpgzip.New(opts...)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, req treemux.Request) error {
		w.Header().Add(vary, acceptEncoding)

		if !hgz.AcceptsGzip(req.Request) {
			return next(w, req)
		}

		gw := hgz.ResponseWriter(w)
		defer gw.Close()

		return next(w, req)
	}
}
