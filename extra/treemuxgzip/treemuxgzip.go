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

var Middleware = (&Config{}).Middleware

type Config struct {
	CompressionLevel int
	ContentTypes     []string
}

func (cfg *Config) Middleware(next treemux.HandlerFunc) treemux.HandlerFunc {
	var opts []httpgzip.ConfigOption
	if cfg.CompressionLevel != 0 {
		opts = append(opts, httpgzip.CompressionLevel(cfg.CompressionLevel))
	}
	if cfg.ContentTypes != nil {
		opts = append(opts, httpgzip.ContentTypes(cfg.ContentTypes))
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
