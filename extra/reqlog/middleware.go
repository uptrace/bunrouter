package reqlog

import (
	"fmt"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/fatih/color"

	"github.com/uptrace/bunrouter"
)

type middleware struct {
	enabled bool
	verbose bool
}

type Option func(m *middleware)

// WithEnabled enables/disables the middleware.
func WithEnabled(on bool) Option {
	return func(m *middleware) {
		m.enabled = on
	}
}

// WithVerbose configures the middleware to log all requests.
func WithVerbose(on bool) Option {
	return func(m *middleware) {
		m.verbose = on
	}
}

// WithEnv configures the middleware using the environment variable value.
// For example, WithEnv("BUNDEBUG"):
//    - BUNDEBUG=0 - disables the middleware.
//    - BUNDEBUG=1 - enables the middleware.
//    - BUNDEBUG=2 - enables the middleware and verbose mode.
func FromEnv(key string) Option {
	if key == "" {
		key = "BUNDEBUG"
	}
	return func(m *middleware) {
		if env, ok := os.LookupEnv(key); ok {
			m.enabled = env != "" && env != "0"
			m.verbose = env == "2"
		}
	}
}

func NewMiddleware(opts ...Option) bunrouter.MiddlewareFunc {
	c := &middleware{
		enabled: true,
		verbose: true,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c.Middleware
}

func (m *middleware) Middleware(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	if !m.enabled {
		return next
	}

	return func(w http.ResponseWriter, req bunrouter.Request) error {
		rec := &statusCodeRecorder{
			ResponseWriter: w,
			Code:           http.StatusOK,
		}

		now := time.Now()
		err := next(rec, req)
		dur := time.Since(now)

		if !m.verbose && rec.Code >= 200 && rec.Code < 300 && err == nil {
			return nil
		}

		args := []interface{}{
			"[bunrouter]",
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

func (rec *statusCodeRecorder) WriteHeader(statusCode int) {
	rec.Code = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

//------------------------------------------------------------------------------

func formatStatus(code int) string {
	return statusColor(code).Sprintf(" %d ", code)
}

func statusColor(code int) *color.Color {
	switch {
	case code >= 200 && code < 300:
		return color.New(color.BgGreen, color.FgHiWhite)
	case code >= 300 && code < 400:
		return color.New(color.BgWhite, color.FgHiBlack)
	case code >= 400 && code < 500:
		return color.New(color.BgYellow, color.FgHiBlack)
	default:
		return color.New(color.BgRed, color.FgHiWhite)
	}
}

func formatMethod(method string) string {
	return methodColor(method).Sprintf(" %-7s ", method)
}

func methodColor(method string) *color.Color {
	switch method {
	case http.MethodGet:
		return color.New(color.BgBlue, color.FgHiWhite)
	case http.MethodPost:
		return color.New(color.BgGreen, color.FgHiWhite)
	case http.MethodPut:
		return color.New(color.BgYellow, color.FgHiBlack)
	case http.MethodDelete:
		return color.New(color.BgRed, color.FgHiWhite)
	case http.MethodPatch:
		return color.New(color.BgCyan, color.FgHiWhite)
	case http.MethodHead:
		return color.New(color.BgMagenta, color.FgHiWhite)
	default:
		return color.New(color.BgWhite, color.FgHiBlack)
	}
}
