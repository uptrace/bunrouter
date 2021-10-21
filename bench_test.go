package bunrouter

import (
	"net/http"
	"testing"
)

func BenchmarkRouterSimple(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/dimfeld", simpleHandler)

	req, _ := http.NewRequest("GET", "/user/dimfeld", nil)

	benchRequest(b, router, req)
}

func BenchmarkRouterRoot(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/dimfeld", simpleHandler)

	req, _ := http.NewRequest("GET", "/", nil)

	benchRequest(b, router, req)
}

func BenchmarkRouterParam(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/:name", simpleHandler)

	req, _ := http.NewRequest("GET", "/user/dimfeld", nil)

	benchRequest(b, router, req)
}

func BenchmarkRouterLongParams(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/:name/:resource", simpleHandler)

	req, _ := http.NewRequest("GET", "/user/aaaabbbbccccddddeeeeffff/asdfghjkl", nil)

	benchRequest(b, router, req)
}

func BenchmarkRouterFiveColon(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/:a/:b/:c/:d/:e", simpleHandler)

	req, _ := http.NewRequest("GET", "/test/test/test/test/test", nil)

	benchRequest(b, router, req)
}

// This type and the benchRequest function are modified from go-http-routing-benchmark.
type mockResponseWriter struct {
	code        int
	calledWrite bool
}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	m.calledWrite = true
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	m.calledWrite = true
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(code int) {
	m.code = code
}

func benchRequest(b *testing.B, router http.Handler, r *http.Request) {
	w := new(mockResponseWriter)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		router.ServeHTTP(w, r)
	}
}
