package main

import (
	"crypto/subtle"
	"html/template"
	"log"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/basicauth"
	"github.com/uptrace/bunrouter/extra/reqlog"
)

func main() {
	router := bunrouter.New(
		bunrouter.Use(reqlog.NewMiddleware(
			reqlog.FromEnv("BUNDEBUG"),
		)),
	)

	authMiddleware := basicauth.NewMiddleware(authUser, basicauth.WithRealm("test:test"))
	router.GET("/", indexHandler)

	router.Use(authMiddleware).
		WithGroup("/restricted", func(g *bunrouter.Group) {
			g.GET("", debugHandler)
		})

	router.WithGroup("/public", func(g *bunrouter.Group) {
		g.GET("", debugHandler)
	})

	log.Println("listening on http://localhost:9999")
	log.Println(http.ListenAndServe(":9999", router))
}

func indexHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return indexTemplate().Execute(w, nil)
}

func debugHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return bunrouter.JSON(w, bunrouter.H{
		"route":  req.Route(),
		"params": req.Params().Map(),
	})
}

func authUser(req bunrouter.Request) (bool, error) {
	user, pass, ok := req.BasicAuth()
	if !ok {
		return false, nil
	}

	if subtle.ConstantTimeCompare([]byte(user), []byte("test")) == 1 &&
		subtle.ConstantTimeCompare([]byte(pass), []byte("test")) == 1 {
		return true, nil
	}
	return false, nil
}

var indexTmpl = `
<html>
  <h1>Welcome</h1>
  <ul>
    <li><a href="/restricted">/restricted</a></li>
    <li><a href="/public">/public</a></li>
  </ul>
</html>
`

func indexTemplate() *template.Template {
	return template.Must(template.New("index").Parse(indexTmpl))
}
