package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-redis/redis_rate/v9"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
)

var ErrRateLimited = errors.New("rate limited")

var limiter *redis_rate.Limiter

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	limiter = redis_rate.NewLimiter(rdb)

	router := bunrouter.New(
		bunrouter.Use(reqlog.NewMiddleware()),
		bunrouter.Use(errorHandler),
		bunrouter.Use(rateLimit),
	)

	router.GET("/", indexHandler)

	log.Println("listening on http://localhost:9999")
	log.Println(http.ListenAndServe(":9999", router))
}

func indexHandler(w http.ResponseWriter, req bunrouter.Request) error {
	_, err := w.Write([]byte("hello world"))
	return err
}

func rateLimit(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		res, err := limiter.Allow(req.Context(), "project:123", redis_rate.PerMinute(10))
		if err != nil {
			return err
		}

		h := w.Header()
		h.Set("RateLimit-Remaining", strconv.Itoa(res.Remaining))

		if res.Allowed == 0 {
			// We are rate limited.

			seconds := int(res.RetryAfter / time.Second)
			h.Set("RateLimit-RetryAfter", strconv.Itoa(seconds))

			// Stop processing and return the error.
			return ErrRateLimited
		}

		// Continue processing as normal.
		return next(w, req)
	}
}

func errorHandler(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
	return func(w http.ResponseWriter, req bunrouter.Request) error {
		err := next(w, req)

		switch err {
		case nil:
			// ok
		case ErrRateLimited:
			w.WriteHeader(http.StatusTooManyRequests)
			_ = bunrouter.JSON(w, bunrouter.H{
				"message": "you are rate limited",
			})
		default:
			w.WriteHeader(http.StatusInternalServerError)
			_ = bunrouter.JSON(w, bunrouter.H{
				"message": err.Error(),
			})
		}

		return err
	}
}
