package httptreemux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func simpleHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {}

func panicHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	panic("test panic")
}

func newRequest(method, path string, body io.Reader) *http.Request {
	unescaped, _ := url.QueryUnescape(path)
	r, _ := http.NewRequest(method, unescaped, body)
	r.RequestURI = path
	return r
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

	makeHandler := func(method string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			result = method
		}
	}

	router := New()
	router.GET("/user/:param", makeHandler("GET"))
	router.POST("/user/:param", makeHandler("POST"))
	router.PATCH("/user/:param", makeHandler("PATCH"))
	router.PUT("/user/:param", makeHandler("PUT"))
	router.DELETE("/user/:param", makeHandler("DELETE"))

	testMethod := func(method, expect string) {
		result = ""
		w := httptest.NewRecorder()
		r := newRequest(method, "/user/"+method, nil)
		router.ServeHTTP(w, r)
		if expect == "" && w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Method %s not expected to match but saw code %d", w.Code)
		}

		if result != expect {
			t.Errorf("Method %s got result %s", method, result)
		}
	}

	testMethod("GET", "GET")
	testMethod("POST", "POST")
	testMethod("PATCH", "PATCH")
	testMethod("PUT", "PUT")
	testMethod("DELETE", "DELETE")
	t.Log("Test HeadCanUseGet = true")
	testMethod("HEAD", "GET")

	router.HeadCanUseGet = false
	t.Log("Test HeadCanUseGet = false")
	testMethod("HEAD", "")

	router.HEAD("/user/:param", makeHandler("HEAD"))

	t.Log("Test HeadCanUseGet = false with explicit HEAD handler")
	testMethod("HEAD", "HEAD")
	router.HeadCanUseGet = true
	t.Log("Test HeadCanUseGet = true with explicit HEAD handler")
	testMethod("HEAD", "HEAD")
}

func TestNotFound(t *testing.T) {
	calledNotFound := false

	notFoundHandler := func(w http.ResponseWriter, r *http.Request) {
		calledNotFound = true
	}

	router := New()
	router.GET("/user/abc", simpleHandler)

	w := httptest.NewRecorder()
	r := newRequest("GET", "/abc/", nil)
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

func TestMethodNotAllowedHandler(t *testing.T) {
	calledNotAllowed := false

	notAllowedHandler := func(w http.ResponseWriter, r *http.Request) {
		calledNotAllowed = true
	}

	router := New()
	router.GET("/user/abc", simpleHandler)

	w := httptest.NewRecorder()
	r := newRequest("POST", "/user/abc", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected error %d from built-in not found handler but saw %d",
			http.StatusMethodNotAllowed, w.Code)
	}

	// Now try with a custome handler.
	router.MethodNotAllowedHandler = notAllowedHandler

	router.ServeHTTP(w, r)
	if !calledNotAllowed {
		t.Error("Custom not allowed handler was not called")
	}
}

func TestPanic(t *testing.T) {

	router := New()
	router.GET("/abc", panicHandler)
	r := newRequest("GET", "/abc", nil)
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
	w = httptest.NewRecorder()
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
	r := newRequest("GET", "/slash", nil)
	router.ServeHTTP(w, r)
	if w.Code != http.StatusMovedPermanently {
		t.Errorf("/slash expected code 301, saw %d", w.Code)
	}
	if w.Header().Get("Location") != "/slash/" {
		t.Errorf("/slash was not redirected to /slash/")
	}

	r = newRequest("GET", "/noslash/", nil)
	w = httptest.NewRecorder()
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

	r := newRequest("GET", "/", nil)
	w := new(mockResponseWriter)
	router.ServeHTTP(w, r)

	if !handlerCalled {
		t.Error("Handler not called for root path")
	}
}

func TestSlash(t *testing.T) {
	param := ""
	handler := func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		param = params["param"]
	}
	router := New()
	router.GET("/abc/:param", handler)

	r := newRequest("GET", "/abc/de%2ff", nil)
	w := new(mockResponseWriter)
	router.ServeHTTP(w, r)

	if param != "de/f" {
		t.Errorf("Expected param de/f, saw %s", param)
	}
}

func BenchmarkSimple(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/dimfeld", simpleHandler)

	r := newRequest("GET", "/user/dimfeld", nil)

	benchRequest(b, router, r)
}

func BenchmarkRoot(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/dimfeld", simpleHandler)

	r := newRequest("GET", "/", nil)

	benchRequest(b, router, r)
}

func BenchmarkParam(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/:name", simpleHandler)

	r := newRequest("GET", "/user/dimfeld", nil)

	benchRequest(b, router, r)
}
