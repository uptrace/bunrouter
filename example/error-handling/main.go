package main

import (
	"errors"
	"log"
	"math/rand"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
)

var (
	err1 = errors.New("error1")
	err2 = errors.New("error2")
)

func main() {
	router := bunrouter.New(
		bunrouter.WithMiddleware(reqlog.NewMiddleware()),
		bunrouter.WithMiddleware(errorHandler),
	)

	router.GET("/", indexHandler)

	log.Println("listening on http://localhost:9999")
	log.Println(http.ListenAndServe(":9999", router))
}

func indexHandler(w http.ResponseWriter, req bunrouter.Request) error {
	if rand.Float64() > 0.5 {
		return err1
	}
	return err2
}

func errorHandler(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		err := next(w, req)

		switch err {
		case nil:
			// ok
		case err1:
			w.WriteHeader(http.StatusBadRequest)
			_ = bunrouter.JSON(w, bunrouter.H{
				"message": "bad request",
				"hint":    "reload to see how error message is changed",
			})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			_ = bunrouter.JSON(w, bunrouter.H{
				"message": err.Error(),
				"hint":    "reload to see how error message is changed",
			})
		}

		return err
	}
}
