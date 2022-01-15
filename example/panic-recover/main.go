package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
)

func main() {
	router := bunrouter.New(
		bunrouter.Use(reqlog.NewMiddleware(
			reqlog.FromEnv("BUNDEBUG"),
		)),
	)

	router.GET("/", indexHandler)

	handler := http.Handler(router)
	handler = PanicHandler{Next: handler}

	log.Println("listening on http://localhost:9999")
	log.Println(http.ListenAndServe(":9999", handler))
}

func indexHandler(w http.ResponseWriter, req bunrouter.Request) error {
	panic("oops")
}

type PanicHandler struct {
	Next http.Handler
}

func (h PanicHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			buf := make([]byte, 10<<10)
			n := runtime.Stack(buf, false)
			fmt.Fprintf(os.Stderr, "panic: %v\n\n%s", err, buf[:n])

			// Uncomment to exit instead of recovering.
			// os.Exit(1)

			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		}
	}()

	h.Next.ServeHTTP(w, req)
}
