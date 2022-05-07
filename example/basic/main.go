package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
)

func main() {
	router := bunrouter.New(
		bunrouter.Use(reqlog.NewMiddleware(
			reqlog.FromEnv("BUNDEBUG"),
		)),
		bunrouter.WithNotFoundHandler(notFoundHandler),
		bunrouter.WithMethodNotAllowedHandler(methodNotAllowedHandler),
	)

	router.GET("/", indexHandler)
	router.POST("/405", indexHandler) // to test methodNotAllowedHandler

	router.WithGroup("/api", func(g *bunrouter.Group) {
		g.GET("/users/:id", debugHandler)
		g.GET("/users/current", debugHandler)
		g.GET("/users/*path", debugHandler)
	})

	log.Println("listening on http://localhost:9999")
	log.Println(http.ListenAndServe(":9999", router))
}

func indexHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return indexTemplate().Execute(w, nil)
}

func debugHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return bunrouter.JSON(w, bunrouter.H{
		"route":  req.Route(),
		"params": req.Params().Map(),
	})
}

func notFoundHandler(w http.ResponseWriter, req bunrouter.Request) error {
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(
		w,
		"<html>BunRouter can't find a route that matches <strong>%s</strong></html>",
		req.URL.Path,
	)
	return nil
}

func methodNotAllowedHandler(w http.ResponseWriter, req bunrouter.Request) error {
	w.WriteHeader(http.StatusMethodNotAllowed)
	fmt.Fprintf(
		w,
		"<html>BunRouter does have a route that matches <strong>%s</strong>, "+
			"but it does not handle method <strong>%s</strong></html>",
		req.URL.Path, req.Method,
	)
	return nil
}

var indexTmpl = `
<html>
  <h1>Welcome</h1>
  <ul>
    <li><a href="/api/users/123">/api/users/123</a></li>
    <li><a href="/api/users/current">/api/users/current</a></li>
    <li><a href="/api/users/foo/bar">/api/users/foo/bar</a></li>
    <li><a href="/404">/404</a></li>
    <li><a href="/405">/405</a></li>
  </ul>
</html>
`

func indexTemplate() *template.Template {
	return template.Must(template.New("index").Parse(indexTmpl))
}
