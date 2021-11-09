package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/rs/cors"
	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
)

func main() {
	router := bunrouter.New(
		bunrouter.WithMiddleware(reqlog.NewMiddleware()),
	)

	router.GET("/", indexHandler)

	router.NewGroup("/api/v1",
		// Install CORS only for this group.
		bunrouter.WithMiddleware(corsMiddleware),
		bunrouter.WithGroup(func(g *bunrouter.Group) {
			g.GET("/users/:id", userHandler)
		}))

	router.NewGroup("/api/v2",
		// Install CORS only for this group.
		bunrouter.WithMiddleware(newCorsMiddleware([]string{"http://localhost:9999"})),
		bunrouter.WithGroup(func(g *bunrouter.Group) {
			g.GET("/users/:id", userHandler)
		}))

	log.Println("listening on http://localhost:9999")
	log.Println(http.ListenAndServe(":9999", router))
}

// corsMiddleware handles CORS requests with custom middleware.
func corsMiddleware(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
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

// newCorsMiddleware creates CORS middleware using github.com/rs/cors package.
func newCorsMiddleware(allowedOrigins []string) bunrouter.MiddlewareFunc {
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
	})

	return func(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
		return bunrouter.HTTPHandler(corsHandler.Handler(next))
	}
}

//------------------------------------------------------------------------------

func indexHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return indexTemplate().Execute(w, nil)
}

func userHandler(w http.ResponseWriter, req bunrouter.Request) error {
	id, err := req.Params().Uint64("id")
	if err != nil {
		return err
	}

	return bunrouter.JSON(w, bunrouter.H{
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
