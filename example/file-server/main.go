package main

import (
	"embed"
	"log"
	"net/http"
	"text/template"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
)

//go:embed files
var filesFS embed.FS

func main() {
	fileServer := http.FileServer(http.FS(filesFS))

	router := bunrouter.New(
		bunrouter.WithMiddleware(reqlog.NewMiddleware(
			reqlog.FromEnv("BUNDEBUG"),
		)),
	)

	router.GET("/", indexHandler)
	router.GET("/files/", bunrouter.HTTPHandler(fileServer))
	router.GET("/files/*path", bunrouter.HTTPHandler(fileServer))

	log.Println("listening on http://localhost:9999")
	log.Println(http.ListenAndServe(":9999", router))
}

func indexHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return indexTemplate().Execute(w, nil)
}

var indexTmpl = `
<html>
  <h1>Welcome</h1>
  <ul>
    <li><a href="/files/">/files/</a></li>
    <li><a href="/files/hello.txt">/files/hello.txt</a></li>
    <li><a href="/files/world.txt">/files/world.txt</a></li>
  </ul>
</html>
`

func indexTemplate() *template.Template {
	return template.Must(template.New("index").Parse(indexTmpl))
}
