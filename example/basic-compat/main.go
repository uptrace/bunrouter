package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
)

func main() {
	router := bunrouter.New(
		bunrouter.WithMiddleware(reqlog.NewMiddleware()),
	).Compat()

	router.GET("/", indexHandler)

	router.WithGroup("/api", func(g *bunrouter.CompatGroup) {
		g.GET("/users/:id", userHandler)
		g.GET("/images/*path", imageHandler)
		g.GET("/images/my.jpg", customImageHandler)
	})

	log.Println("listening on http://localhost:9999")
	log.Println(http.ListenAndServe(":9999", router))
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	if err := indexTemplate().Execute(w, nil); err != nil {
		panic(err)
	}
}

func userHandler(w http.ResponseWriter, req *http.Request) {
	params := bunrouter.ParamsFromContext(req.Context())

	id, err := params.Uint64("id")
	if err != nil {
		panic(err)
	}

	if err := bunrouter.JSON(w, bunrouter.H{
		"route": params.Route(),
		"id":    id,
	}); err != nil {
		panic(err)
	}
}

func imageHandler(w http.ResponseWriter, req *http.Request) {
	params := bunrouter.ParamsFromContext(req.Context())

	if err := bunrouter.JSON(w, bunrouter.H{
		"route": params.Route(),
		"path":  params.ByName("path"),
	}); err != nil {
		panic(err)
	}
}

func customImageHandler(w http.ResponseWriter, req *http.Request) {
	params := bunrouter.ParamsFromContext(req.Context())

	if err := bunrouter.JSON(w, bunrouter.H{
		"route": params.Route(),
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
    <li><a href="/api/images/my.jpg">/images/my.jpg</a></li>
  </ul>
</html>
`

func indexTemplate() *template.Template {
	return template.Must(template.New("index").Parse(indexTmpl))
}
