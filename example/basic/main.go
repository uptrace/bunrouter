package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/vmihailenco/treemux"
	"github.com/vmihailenco/treemux/extra/reqlog"
)

func main() {
	router := treemux.New(
		treemux.WithMiddleware(reqlog.Middleware),
	)

	router.GET("/", indexHandler)

	router.WithGroup("/api/v1", func(g *treemux.Group) {
		g.GET("/users/:id", userHandler)
	})

	log.Println("listening on http://localhost:8080")
	log.Println(http.ListenAndServe(":8080", router))
}

func indexHandler(w http.ResponseWriter, req treemux.Request) error {
	return indexTemplate().Execute(w, nil)
}

func userHandler(w http.ResponseWriter, req treemux.Request) error {
	id, err := req.Params.Uint64("id")
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
