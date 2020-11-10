package reqlog

import (
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/fatih/color"
	"github.com/vmihailenco/treemux"
)

var Middleware = New().Middleware

type Config struct {
	verbose bool
}

type Option func(c *Config)

func WithVerbose(on bool) Option {
	return func(c *Config) {
		c.verbose = on
	}
}

func New(opts ...Option) *Config {
	c := &Config{
		verbose: true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (cfg *Config) Middleware(next treemux.HandlerFunc) treemux.HandlerFunc {
	return func(w http.ResponseWriter, req treemux.Request) error {
		rec := statusCodeRecorder{
			ResponseWriter: w,
			Code:           http.StatusOK,
		}

		now := time.Now()
		err := next(rec, req)
		dur := time.Since(now)

		if !cfg.verbose && rec.Code >= 200 && rec.Code < 300 && err == nil {
			return nil
		}

		args := []interface{}{
			"[treemux]",
			now.Format(" 15:04:05.000 "),
			formatStatus(rec.Code),
			fmt.Sprintf(" %10s ", dur.Round(time.Microsecond)),
			formatMethod(req.Method),
			req.URL.String(),
		}

		if err != nil {
			typ := reflect.TypeOf(err).String()
			args = append(args,
				"\t",
				color.New(color.BgRed).Sprintf(" %s ", typ+": "+err.Error()),
			)
		}

		fmt.Println(args...)

		return err
	}
}

//------------------------------------------------------------------------------

type statusCodeRecorder struct {
	http.ResponseWriter
	Code int
}

func (rec statusCodeRecorder) WriteHeader(statusCode int) {
	rec.Code = statusCode
}

//------------------------------------------------------------------------------

func formatStatus(code int) string {
	return statusColor(code).Sprintf(" %d ", code)
}

func statusColor(code int) *color.Color {
	switch {
	case code >= 200 && code < 300:
		return color.New(color.BgGreen)
	case code >= 300 && code < 400:
		return color.New(color.BgWhite)
	case code >= 400 && code < 500:
		return color.New(color.BgYellow)
	default:
		return color.New(color.BgRed)
	}
}

func formatMethod(method string) string {
	return methodColor(method).Sprintf(" %-7s ", method)
}

func methodColor(method string) *color.Color {
	switch method {
	case http.MethodGet:
		return color.New(color.BgBlue)
	case http.MethodPost:
		return color.New(color.BgCyan)
	case http.MethodPut:
		return color.New(color.BgYellow)
	case http.MethodDelete:
		return color.New(color.BgRed)
	case http.MethodPatch:
		return color.New(color.BgGreen)
	case http.MethodHead:
		return color.New(color.BgMagenta)
	default:
		return color.New(color.BgWhite)
	}
}
