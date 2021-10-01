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
	).Verbose()

	router.GET("/", indexHandler)

	router.WithGroup("/api", func(g *bunrouter.VerboseGroup) {
		g.GET("/users/:id", debugHandler)
		g.GET("/images/*path", debugHandler)
		g.GET("/images/my.jpg", debugHandler)
	})

	log.Println("listening on http://localhost:9999")
	log.Println(http.ListenAndServe(":9999", router))
}

func indexHandler(w http.ResponseWriter, req *http.Request, params bunrouter.Params) {
	_ = indexTemplate().Execute(w, nil)
}

func debugHandler(w http.ResponseWriter, req *http.Request, params bunrouter.Params) {
	_ = bunrouter.JSON(w, bunrouter.H{
		"route":  params.Route(),
		"params": params.Map(),
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
