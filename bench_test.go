package bunrouter

import (
	"net/http"
	"testing"
)

func BenchmarkRouterSimple(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/dimfeld", simpleHandler)

	r, _ := http.NewRequest("GET", "/user/dimfeld", nil)

	benchRequest(b, router, r)
}

func BenchmarkRouterRoot(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/dimfeld", simpleHandler)

	r, _ := http.NewRequest("GET", "/", nil)

	benchRequest(b, router, r)
}

func BenchmarkRouterParam(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/:name", simpleHandler)

	r, _ := http.NewRequest("GET", "/user/dimfeld", nil)

	benchRequest(b, router, r)
}

func BenchmarkRouterLongParams(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/:name/:resource", simpleHandler)

	r, _ := http.NewRequest("GET", "/user/aaaabbbbccccddddeeeeffff/asdfghjkl", nil)

	benchRequest(b, router, r)
}
