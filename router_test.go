package httptreemux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sort"
	"testing"
)

func simpleHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {}

func panicHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	panic("test panic")
}

func newRequest(method, path string, body io.Reader) *http.Request {
	r, _ := http.NewRequest(method, path, body)
	u, _ := url.Parse(path)
	r.URL = u
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

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
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

	notAllowedHandler := func(w http.ResponseWriter, r *http.Request,
		methods map[string]HandlerFunc) {

		calledNotAllowed = true

		expected := []string{"GET", "PUT", "DELETE"}
		allowed := make([]string, 0)
		for m := range methods {
			allowed = append(allowed, m)
		}

		sort.Strings(expected)
		sort.Strings(allowed)

		if !reflect.DeepEqual(expected, allowed) {
			t.Errorf("Custom handler expected map %v, saw %v",
				expected, allowed)
		}
	}

	router := New()
	router.GET("/user/abc", simpleHandler)
	router.PUT("/user/abc", simpleHandler)
	router.DELETE("/user/abc", simpleHandler)

	w := httptest.NewRecorder()
	r := newRequest("POST", "/user/abc", nil)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected error %d from built-in not found handler but saw %d",
			http.StatusMethodNotAllowed, w.Code)
	}

	allowed := w.Header()["Allow"]
	sort.Strings(allowed)
	expected := []string{"DELETE", "GET", "PUT"}
	sort.Strings(expected)

	if !reflect.DeepEqual(allowed, expected) {
		t.Errorf("Expected Allow header %v, saw %v",
			expected, allowed)
	}

	// Now try with a custom handler.
	router.MethodNotAllowedHandler = notAllowedHandler

	router.ServeHTTP(w, r)
	if !calledNotAllowed {
		t.Error("Custom not allowed handler was not called")
	}
}

func TestPanic(t *testing.T) {

	router := New()
	router.PanicHandler = SimplePanicHandler
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
		t.Errorf("/noslash/ was not redirected to /noslash")
	}

	r = newRequest("GET", "//noslash/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != http.StatusMovedPermanently {
		t.Errorf("//noslash/ expected code 301, saw %d", w.Code)
	}
	if w.Header().Get("Location") != "/noslash" {
		t.Errorf("//noslash/ was not redirected to /noslash")
	}

}

func TestSkipRedirect(t *testing.T) {
	router := New()
	router.RedirectTrailingSlash = false
	router.RedirectCleanPath = false
	router.GET("/slash/", simpleHandler)
	router.GET("/noslash", simpleHandler)

	w := httptest.NewRecorder()
	r := newRequest("GET", "/slash", nil)
	router.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Errorf("/slash expected code 404, saw %d", w.Code)
	}

	r = newRequest("GET", "/noslash/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Errorf("/noslash/ expected code 404, saw %d", w.Code)
	}

	r = newRequest("GET", "//noslash", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Errorf("//noslash expected code 404, saw %d", w.Code)
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
	ymHandler := func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		param = params["year"] + " " + params["month"]
	}
	router := New()
	router.GET("/abc/:param", handler)
	router.GET("/year/:year/month/:month", ymHandler)

	r := newRequest("GET", "/abc/de%2ff", nil)
	w := new(mockResponseWriter)
	router.ServeHTTP(w, r)

	if param != "de/f" {
		t.Errorf("Expected param de/f, saw %s", param)
	}

	r = newRequest("GET", "/year/de%2f/month/fg%2f", nil)
	router.ServeHTTP(w, r)

	if param != "de/ fg/" {
		t.Errorf("Expected param de/ fg/, saw %s", param)
	}
}

func TestQueryString(t *testing.T) {
	param := ""
	handler := func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		param = params["param"]
	}
	router := New()
	router.GET("/static", handler)
	router.GET("/wildcard/:param", handler)
	router.GET("/catchall/*param", handler)

	r := newRequest("GET", "/static?abc=def&ghi=jkl", nil)
	w := new(mockResponseWriter)

	param = "nomatch"
	router.ServeHTTP(w, r)
	if param != "" {
		t.Error("No match on", r.RequestURI)
	}

	r = newRequest("GET", "/wildcard/aaa?abc=def", nil)
	router.ServeHTTP(w, r)
	if param != "aaa" {
		t.Error("Expected wildcard to match aaa, saw", param)
	}

	r = newRequest("GET", "/catchall/bbb?abc=def", nil)
	router.ServeHTTP(w, r)
	if param != "bbb" {
		t.Error("Expected wildcard to match bbb, saw", param)
	}
}

func BenchmarkRouterSimple(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/dimfeld", simpleHandler)

	r := newRequest("GET", "/user/dimfeld", nil)

	benchRequest(b, router, r)
}

func BenchmarkRouterRootWithPanicHandler(b *testing.B) {
	router := New()
	router.PanicHandler = SimplePanicHandler

	router.GET("/", simpleHandler)
	router.GET("/user/dimfeld", simpleHandler)

	r := newRequest("GET", "/", nil)

	benchRequest(b, router, r)
}

func BenchmarkRouterRootWithoutPanicHandler(b *testing.B) {
	router := New()
	router.PanicHandler = nil

	router.GET("/", simpleHandler)
	router.GET("/user/dimfeld", simpleHandler)

	r := newRequest("GET", "/", nil)

	benchRequest(b, router, r)
}

func BenchmarkRouterParam(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/:name", simpleHandler)

	r := newRequest("GET", "/user/dimfeld", nil)

	benchRequest(b, router, r)
}
