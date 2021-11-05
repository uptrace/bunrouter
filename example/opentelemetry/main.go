package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	"github.com/klauspost/compress/gzhttp"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/bunrouterotel"
	"github.com/uptrace/bunrouter/extra/reqlog"
	"github.com/uptrace/opentelemetry-go-extra/otelplay"
)

func main() {
	ctx := context.Background()

	shutdown := otelplay.ConfigureOpentelemetry(ctx)
	defer shutdown()

	router := bunrouter.New(
		bunrouter.WithMiddleware(reqlog.NewMiddleware()),
		bunrouter.WithMiddleware(bunrouterotel.NewMiddleware(
			bunrouterotel.WithClientIP(),
		)),
	)

	router.GET("/", indexHandler)

	router.WithGroup("/api", func(g *bunrouter.Group) {
		g.GET("/users/:id", debugHandler)
		g.GET("/users/current", debugHandler)
		g.GET("/users/*path", debugHandler)
	})

	handler := http.Handler(router)
	handler = gzhttp.GzipHandler(handler)
	handler = otelhttp.NewHandler(router, "")

	httpServer := &http.Server{
		Addr:         ":9999",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      handler,
	}

	log.Println("listening on http://localhost:9999")
	log.Println(httpServer.ListenAndServe())
}

func indexHandler(w http.ResponseWriter, req bunrouter.Request) error {
	m := map[string]interface{}{
		"traceURL": otelplay.TraceURL(trace.SpanFromContext(req.Context())),
	}
	return indexTemplate().Execute(w, m)
}

func debugHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return bunrouter.JSON(w, bunrouter.H{
		"route":    req.Route(),
		"params":   req.Params().Map(),
		"traceURL": otelplay.TraceURL(trace.SpanFromContext(req.Context())),
	})
}

var indexTmpl = `
<html>
  <h1>Welcome</h1>
  <ul>
    <li><a href="/api/users/123">/api/users/123</a></li>
    <li><a href="/api/users/current">/api/users/current</a></li>
    <li><a href="/api/users/foo/bar">/api/users/foo/bar</a></li>
  </ul>
  <p><a href="{{ .traceURL }}">{{ .traceURL }}</a></p>
</html>
`

func indexTemplate() *template.Template {
	return template.Must(template.New("index").Parse(indexTmpl))
}
