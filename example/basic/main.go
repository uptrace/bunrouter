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
	)

	router.GET("/", indexHandler)

	router.WithGroup("/api", func(g *bunrouter.Group) {
		g.GET("/users/:id", userHandler)
		g.GET("/images/*path", imageHandler)
		g.GET("/images/my.jpg", customImageHandler)
	})

	log.Println("listening on http://localhost:9999")
	log.Println(http.ListenAndServe(":9999", router))
}

func indexHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return indexTemplate().Execute(w, nil)
}

func userHandler(w http.ResponseWriter, req bunrouter.Request) error {
	id, err := req.Params().Uint64("id")
	if err != nil {
		return err
	}

	return bunrouter.JSON(w, bunrouter.H{
		"route": req.Route(),
		"id":    id,
	})
}

func imageHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return bunrouter.JSON(w, bunrouter.H{
		"route": req.Route(),
		"path":  req.Param("path"),
	})
}

func customImageHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return bunrouter.JSON(w, bunrouter.H{
		"route": req.Route(),
	})
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
