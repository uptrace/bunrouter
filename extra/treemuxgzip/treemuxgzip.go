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

type config struct {
	compressionLevel int
	contentTypes     []string
}

type Option func(c *config)

func WithCompressionLevel(level int) Option {
	return func(c *config) {
		c.compressionLevel = level
	}
}

func WithContentTypes(contentTypes ...string) Option {
	return func(c *config) {
		c.contentTypes = contentTypes
	}
}

func NewMiddleware(opts ...Option) treemux.MiddlewareFunc {
	c := &config{}
	for _, opt := range opts {
		opt(c)
	}
	return c.Middleware
}

func (c *config) Middleware(next treemux.HandlerFunc) treemux.HandlerFunc {
	var opts []httpgzip.Option
	if c.compressionLevel != 0 {
		opts = append(opts, httpgzip.CompressionLevel(c.compressionLevel))
	}
	if c.contentTypes != nil {
		opts = append(opts, httpgzip.ContentTypes(c.contentTypes))
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
