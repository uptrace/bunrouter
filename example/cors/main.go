package main

import (
	"errors"
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
		bunrouter.Use(reqlog.NewMiddleware()),
	)

	router.GET("/", indexHandler)

	router.Use(errorMiddleware).
		// Install CORS only for this group.
		Use(newCorsMiddleware([]string{"http://localhost:9999"})).
		WithGroup("/api/v1", func(g *bunrouter.Group) {
			g.GET("/users/:id", userHandler)
			g.GET("/error", failingHandler)
		})

	log.Println("listening on http://localhost:9999")
	log.Println(http.ListenAndServe(":9999", router))
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

func errorMiddleware(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		err := next(w, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return err
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

func failingHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return errors.New("just an error")
}

var indexTmpl = `
<html>
  <h1>Welcome</h1>
  <ul>
    <li><a href="/api/v1/users/123">/api/v1/users/123</a></li>
    <li><a href="/api/v1/error">/api/v1/error</a></li>

    <li><a href="/api/v2/users/123">/api/v2/users/123</a></li>
    <li><a href="/api/v2/error">/api/v2/error</a></li>
  </ul>
</html>
`

func indexTemplate() *template.Template {
	return template.Must(template.New("index").Parse(indexTmpl))
}
