package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/vmihailenco/treemux"
	"github.com/vmihailenco/treemux/extra/reqlog"
)

func main() {
	router := treemux.New(
		treemux.WithMiddleware(reqlog.NewMiddleware()),
	)

	router.GET("/", indexHandler)

	router.WithGroup("/api", func(g *treemux.Group) {
		g.GET("/users/:id", userHandler)
		g.GET("/images/*path", imageHandler)
	})

	log.Println("listening on http://localhost:8888")
	log.Println(http.ListenAndServe(":8888", router))
}

func indexHandler(w http.ResponseWriter, req treemux.Request) error {
	return indexTemplate().Execute(w, nil)
}

func userHandler(w http.ResponseWriter, req treemux.Request) error {
	id, err := req.Params().Uint64("id")
	if err != nil {
		return err
	}

	return treemux.JSON(w, treemux.H{
		"route": req.Route(),
		"id":    id,
	})
}

func imageHandler(w http.ResponseWriter, req treemux.Request) error {
	return treemux.JSON(w, treemux.H{
		"route": req.Route(),
		"path":  req.Param("path"),
	})
}

var indexTmpl = `
<html>
  <h1>Welcome</h1>
  <ul>
    <li><a href="/api/users/123">/api/users/123</a></li>
    <li><a href="/api/images/path/to/image.jpg">/images/path/to/image.jpg</a></li>
  </ul>
</html>
`

func indexTemplate() *template.Template {
	return template.Must(template.New("index").Parse(indexTmpl))
}
