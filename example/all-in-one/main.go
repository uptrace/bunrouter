package main

import (
	"context"
	"embed"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"

	"github.com/klauspost/compress/gzhttp"
	"github.com/rs/cors"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
)

//go:embed files
var filesFS embed.FS

func main() {
	ctx := context.Background()

	fileServer := http.FileServer(http.FS(filesFS))

	router := bunrouter.New(
		bunrouter.Use(reqlog.NewMiddleware(
			reqlog.FromEnv("BUNDEBUG"),
		)),
	)

	router.GET("/", indexHandler)
	router.GET("/files/*path", bunrouter.HTTPHandler(fileServer))

	httpLn, err := net.Listen("tcp", "localhost:9999")
	if err != nil {
		panic(err)
	}

	handler := http.Handler(router)
	handler = cors.Default().Handler(handler)
	handler = gzhttp.GzipHandler(handler)

	httpServer := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      handler,
	}

	fmt.Println("listening on http://localhost:9999...")
	go func() {
		if err := httpServer.Serve(httpLn); err != nil {
			fmt.Println(err)
		}
	}()

	fmt.Println("Press CTRL+C to exit...")
	fmt.Println(waitExitSignal())

	// Graceful shutdown.
	if err := httpServer.Shutdown(ctx); err != nil {
		fmt.Println(err)
	}
}

func indexHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return indexTemplate().Execute(w, nil)
}

var indexTmpl = `
<html>
  <h1>Welcome</h1>
  <ul>
    <li><a href="/files">/files</a></li>
    <li><a href="/files/">/files/</a></li>
    <li><a href="/files/hello.txt">/files/hello.txt</a></li>
    <li><a href="/files/world.txt">/files/world.txt</a></li>
  </ul>
</html>
`

func indexTemplate() *template.Template {
	return template.Must(template.New("index").Parse(indexTmpl))
}

func waitExitSignal() os.Signal {
	ch := make(chan os.Signal, 3)
	signal.Notify(
		ch,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	return <-ch
}
