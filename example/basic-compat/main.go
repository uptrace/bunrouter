package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/uptrace/treemux"
	"github.com/uptrace/treemux/extra/reqlog"
)

func main() {
	router := treemux.New(
		treemux.WithMiddleware(reqlog.NewMiddleware()),
	).Compat()

	router.GET("/", indexHandler)

	router.WithGroup("/api", func(g *treemux.CompatGroup) {
		g.GET("/users/:id", userHandler)
		g.GET("/images/*path", imageHandler)
	})

	log.Println("listening on http://localhost:8888")
	log.Println(http.ListenAndServe(":8888", router))
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	if err := indexTemplate().Execute(w, nil); err != nil {
		panic(err)
	}
}

func userHandler(w http.ResponseWriter, req *http.Request) {
	route := treemux.RouteFromContext(req.Context())

	id, err := route.Params().Uint64("id")
	if err != nil {
		panic(err)
	}

	if err := treemux.JSON(w, treemux.H{
		"route": route.Name(),
		"id":    id,
	}); err != nil {
		panic(err)
	}
}

func imageHandler(w http.ResponseWriter, req *http.Request) {
	route := treemux.RouteFromContext(req.Context())

	if err := treemux.JSON(w, treemux.H{
		"route": route.Name(),
		"path":  route.Param("path"),
	}); err != nil {
		panic(err)
	}
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
