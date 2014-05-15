package httptreemux

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func simpleHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {}

func panicHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	panic("test panic")
}

// This type and the benchRequest function are taken from go-http-routing-benchmark.
type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(int) {}

func benchRequest(b *testing.B, router http.Handler, r *http.Request) {
	w := new(mockResponseWriter)
	u := r.URL
	rq := u.RawQuery

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		u.RawQuery = rq
		router.ServeHTTP(w, r)
	}
}

func TestMethods(t *testing.T) {
	var result string
	var param string

	makeHandler := func(method string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			result = method
			param = method
		}
	}

	router := New()
	router.GET("/user/:param", makeHandler("GET"))
	router.POST("/user/:param", makeHandler("POST"))
	router.PATCH("/user/:param", makeHandler("PATCH"))
	router.PUT("/user/:param", makeHandler("PUT"))
	router.DELETE("/user/:param", makeHandler("DELETE"))

	testMethod := func(method string) {
		result = ""
		param = ""
		w := new(mockResponseWriter)
		r, _ := http.NewRequest(method, "/user/"+method, nil)
		router.ServeHTTP(w, r)
		if result != method {
			t.Errorf("Method %s got result %s", method, result)
		}

		if param != method {
			t.Errorf("Method %s got result %s", param, result)
		}
	}

	testMethod("GET")
	testMethod("POST")
	testMethod("PATCH")
	testMethod("PUT")
	testMethod("DELETE")
}

func TestNotFound(t *testing.T) {
	calledNotFound := false

	notFoundHandler := func(w http.ResponseWriter, r *http.Request) {
		calledNotFound = true
	}

	router := New()
	router.GET("/user/abc", simpleHandler)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/abc/", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected error 404 from built-in not found handler but saw %d", w.Code)
	}

	// Now try with a custome handler.
	router.NotFoundHandler = notFoundHandler

	router.ServeHTTP(w, r)
	if !calledNotFound {
		t.Error("Custom not found handler was not called")
	}
}

func TestPanic(t *testing.T) {

	router := New()
	router.GET("/abc", panicHandler)
	r, _ := http.NewRequest("GET", "/abc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected code %d from default panic handler, saw %d",
			http.StatusInternalServerError, w.Code)
	}

	sawPanic := false
	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, err interface{}) {
		sawPanic = true
	}

	router.ServeHTTP(w, r)
	if !sawPanic {
		t.Errorf("Custom panic handler was not called")
	}

	// Assume this does the right thing. Just a sanity test.
	router.PanicHandler = ShowErrorsPanicHandler
	router.ServeHTTP(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected code %d from ShowErrorsPanicHandler, saw %d",
			http.StatusInternalServerError, w.Code)
	}
}

func TestRedirect(t *testing.T) {
	router := New()
	router.GET("/slash/", simpleHandler)
	router.GET("/noslash", simpleHandler)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/slash", nil)
	router.ServeHTTP(w, r)
	if w.Code != http.StatusMovedPermanently {
		t.Errorf("/slash expected code 301, saw %d", w.Code)
	}
	if w.Header().Get("Location") != "/slash/" {
		t.Errorf("/slash was not redirected to /slash/")
	}

	r, _ = http.NewRequest("GET", "/noslash/", nil)
	router.ServeHTTP(w, r)
	if w.Code != http.StatusMovedPermanently {
		t.Errorf("/noslash/ expected code 301, saw %d", w.Code)
	}
	if w.Header().Get("Location") != "/noslash" {
		t.Errorf("/noslash/ was not redirected to /noslash/")
	}

}

func TestRoot(t *testing.T) {
	handlerCalled := false
	handler := func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		handlerCalled = true
	}
	router := New()
	router.GET("/", handler)

	r, _ := http.NewRequest("GET", "/", nil)
	w := new(mockResponseWriter)
	router.ServeHTTP(w, r)

	if !handlerCalled {
		t.Error("Handler not called for root path")
	}
}

func BenchmarkSimple(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/gordon", simpleHandler)

	r, _ := http.NewRequest("GET", "/user/dimfeld", nil)

	benchRequest(b, router, r)
}

func BenchmarkParam(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/:name", simpleHandler)

	r, _ := http.NewRequest("GET", "/user/dimfeld", nil)

	benchRequest(b, router, r)
}
