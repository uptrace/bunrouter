package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/uptrace/treemux"
	"github.com/uptrace/treemux/extra/reqlog"
)

func main() {
	router := treemux.New(
		treemux.WithMiddleware(reqlog.NewMiddleware()),
	)

	router.GET("/", indexHandler)

	router.NewGroup("/api/v1",
		// Install CORS only for this group.
		treemux.WithMiddleware(corsMiddleware),
		treemux.WithGroup(func(g *treemux.Group) {
			g.GET("/users/:id", userHandler)
		}))

	log.Println("listening on http://localhost:8888")
	log.Println(http.ListenAndServe(":8888", router))
}

func corsMiddleware(next treemux.HandlerFunc) treemux.HandlerFunc {
	return func(w http.ResponseWriter, req treemux.Request) error {
		origin := req.Header.Get("Origin")
		if origin == "" {
			return next(w, req)
		}

		h := w.Header()

		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Access-Control-Allow-Credentials", "true")

		// CORS preflight.
		if req.Method == http.MethodOptions {
			h.Set("Access-Control-Allow-Methods", "GET,PUT,POST,DELETE,HEAD")
			h.Set("Access-Control-Allow-Headers", "authorization,content-type")
			h.Set("Access-Control-Max-Age", "86400")
			return nil
		}

		return next(w, req)
	}
}

//------------------------------------------------------------------------------

func indexHandler(w http.ResponseWriter, req treemux.Request) error {
	return indexTemplate().Execute(w, nil)
}

func userHandler(w http.ResponseWriter, req treemux.Request) error {
	id, err := req.Params().Uint64("id")
	if err != nil {
		return err
	}

	return treemux.JSON(w, treemux.H{
		"url":   fmt.Sprintf("GET /api/v1/%d", id),
		"route": req.Route(),
	})
}

var indexTmpl = `
<html>
  <h1>Welcome</h1>
  <ul>
    <li><a href="/api/v1/users/123">/api/v1/users/123</a></li>
  </ul>
</html>
`

func indexTemplate() *template.Template {
	return template.Must(template.New("index").Parse(indexTmpl))
}
