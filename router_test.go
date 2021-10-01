package bunrouter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func simpleHandler(w http.ResponseWriter, r Request) error {
	return nil
}

type TestScenario struct {
	description string
}

var scenarios = []TestScenario{
	{"Test with URL.Path and normal ServeHTTP"},
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

func TestMethods(t *testing.T) {
	for _, scenario := range scenarios {
		t.Log(scenario.description)
		testMethods(t)
	}
}

func testMethods(t *testing.T) {
	var result string

	makeHandler := func(method string) HandlerFunc {
		return func(w http.ResponseWriter, r Request) error {
			result = method
			return nil
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
		r, _ := http.NewRequest(method, "/user/"+method, nil)
		router.ServeHTTP(w, r)

		if expect == "" {
			require.Equal(t, http.StatusMethodNotAllowed, w.Code)
		} else {
			require.Equal(t, expect, result)
		}
	}

	testMethod("GET", "GET")
	testMethod("POST", "POST")
	testMethod("PATCH", "PATCH")
	testMethod("PUT", "PUT")
	testMethod("DELETE", "DELETE")
	testMethod("HEAD", "")

	router.HEAD("/user/:param", makeHandler("HEAD"))
	testMethod("HEAD", "HEAD")
}

func TestNotFound(t *testing.T) {
	calledNotFound := false

	notFoundHandler := func(w http.ResponseWriter, r Request) error {
		calledNotFound = true
		return nil
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
	router = New(WithNotFoundHandler(notFoundHandler))
	router.GET("/user/abc", simpleHandler)

	router.ServeHTTP(w, r)
	if !calledNotFound {
		t.Error("Custom not found handler was not called")
	}
}

func TestRedirect(t *testing.T) {
	for _, scenario := range scenarios {
		t.Log(scenario.description)
		testRedirect(t)
	}
}

func testRedirect(t *testing.T) {
	redirHandler := func(w http.ResponseWriter, r Request) error {
		// Returning this instead of 200 makes it easy to verify that the handler is actually getting called.
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	router := New()
	expectedCodeMap := map[string]int{
		"GET":  http.StatusMovedPermanently,
		"POST": http.StatusMovedPermanently,
		"PUT":  http.StatusMovedPermanently,
	}

	router.GET("/slash/", redirHandler)
	router.GET("/noslash", redirHandler)
	router.POST("/slash/", redirHandler)
	router.POST("/noslash", redirHandler)
	router.PUT("/slash/", redirHandler)
	router.PUT("/noslash", redirHandler)

	for method, expectedCode := range expectedCodeMap {
		t.Logf("Testing method %s, expecting code %d", method, expectedCode)

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(method, "/slash", nil)
		router.ServeHTTP(w, r)
		if w.Code != expectedCode {
			t.Errorf("/slash expected code %d, saw %d", expectedCode, w.Code)
		}
		if expectedCode != http.StatusNoContent && w.Header().Get("Location") != "/slash/" {
			t.Errorf("/slash was not redirected to /slash/")
		}

		r, _ = http.NewRequest(method, "/noslash/", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != expectedCode {
			t.Errorf("/noslash/ expected code %d, saw %d", expectedCode, w.Code)
		}
		if expectedCode != http.StatusNoContent && w.Header().Get("Location") != "/noslash" {
			t.Errorf("/noslash/ was redirected to `%s` instead of /noslash", w.Header().Get("Location"))
		}

		r, _ = http.NewRequest(method, "//noslash/", nil)
		if r.RequestURI == "//noslash/" { // http.NewRequest parses this out differently
			w = httptest.NewRecorder()
			router.ServeHTTP(w, r)
			if w.Code != expectedCode {
				t.Errorf("//noslash/ expected code %d, saw %d", expectedCode, w.Code)
			}
			if expectedCode != http.StatusNoContent && w.Header().Get("Location") != "/noslash" {
				t.Errorf("//noslash/ was redirected to %s, expected /noslash", w.Header().Get("Location"))
			}
		}

		// Test nonredirect cases
		r, _ = http.NewRequest(method, "/noslash", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != http.StatusNoContent {
			t.Errorf("/noslash (non-redirect) expected code %d, saw %d", http.StatusNoContent, w.Code)
		}

		r, _ = http.NewRequest(method, "/noslash?a=1&b=2", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != http.StatusNoContent {
			t.Errorf("/noslash (non-redirect) expected code %d, saw %d", http.StatusNoContent, w.Code)
		}

		r, _ = http.NewRequest(method, "/slash/", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != http.StatusNoContent {
			t.Errorf("/slash/ (non-redirect) expected code %d, saw %d", http.StatusNoContent, w.Code)
		}

		r, _ = http.NewRequest(method, "/slash/?a=1&b=2", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != http.StatusNoContent {
			t.Errorf("/slash/?a=1&b=2 expected code %d, saw %d", http.StatusNoContent, w.Code)
		}

		// Test querystring and fragment cases
		r, _ = http.NewRequest(method, "/slash?a=1&b=2", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != expectedCode {
			t.Errorf("/slash?a=1&b=2 expected code %d, saw %d", expectedCode, w.Code)
		}
		if expectedCode != http.StatusNoContent && w.Header().Get("Location") != "/slash/?a=1&b=2" {
			t.Errorf("/slash?a=1&b=2 was redirected to %s", w.Header().Get("Location"))
		}

		r, _ = http.NewRequest(method, "/noslash/?a=1&b=2", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != expectedCode {
			t.Errorf("/noslash/?a=1&b=2 expected code %d, saw %d", expectedCode, w.Code)
		}
		if expectedCode != http.StatusNoContent && w.Header().Get("Location") != "/noslash?a=1&b=2" {
			t.Errorf("/noslash/?a=1&b=2 was redirected to %s", w.Header().Get("Location"))
		}
	}
}

func TestRedirectClean(t *testing.T) {
	router := New()

	router.GET("/slash/", simpleHandler)
	router.GET("/noslash", simpleHandler)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/slash", nil)
	router.ServeHTTP(w, r)
	require.Equal(t, http.StatusMovedPermanently, w.Code)

	r, _ = http.NewRequest("GET", "/noslash/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	require.Equal(t, http.StatusMovedPermanently, w.Code)

	r, _ = http.NewRequest("GET", "//noslash", nil)
	r.URL.Path = "//noslash"
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	require.Equal(t, http.StatusMovedPermanently, w.Code)
}

func TestRoot(t *testing.T) {
	for _, scenario := range scenarios {
		t.Log(scenario.description)

		var handlerCalled bool

		handler := func(w http.ResponseWriter, r Request) error {
			handlerCalled = true
			return nil
		}

		router := New()
		router.GET("/", handler)

		req, err := http.NewRequest("GET", "/", nil)
		require.NoError(t, err)

		w := new(mockResponseWriter)
		router.ServeHTTP(w, req)

		if !handlerCalled {
			t.Error("Handler not called for root path")
		}
	}
}

func TestWildcardAtSplitNode(t *testing.T) {
	var suppliedParam string
	simpleHandler := func(w http.ResponseWriter, r Request) error {
		t.Log(r.Params().Map())
		suppliedParam, _ = r.Params().Get("slug")
		return nil
	}

	router := New()
	router.GET("/pumpkin", simpleHandler)
	router.GET("/passing", simpleHandler)
	router.GET("/:slug", simpleHandler)
	router.GET("/:slug/abc", simpleHandler)

	r, _ := http.NewRequest("GET", "/patch", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if suppliedParam != "patch" {
		t.Errorf("Expected param patch, saw %s", suppliedParam)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for path /patch, saw %d", w.Code)
	}

	suppliedParam = ""
	r, _ = http.NewRequest("GET", "/patch/abc", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if suppliedParam != "patch" {
		t.Errorf("Expected param patch, saw %s", suppliedParam)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for path /patch/abc, saw %d", w.Code)
	}

	r, _ = http.NewRequest("GET", "/patch/def", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestQueryString(t *testing.T) {
	for _, scenario := range scenarios {
		t.Log(scenario.description)
		param := ""
		handler := func(w http.ResponseWriter, r Request) error {
			param = r.Params().ByName("param")
			return nil
		}
		router := New()
		router.GET("/static", handler)
		router.GET("/wildcard/:param", handler)
		router.GET("/catchall/*param", handler)

		r, _ := http.NewRequest("GET", "/static?abc=def&ghi=jkl", nil)
		w := new(mockResponseWriter)

		param = "nomatch"
		router.ServeHTTP(w, r)
		if param != "" {
			t.Error("No match on", r.RequestURI)
		}

		r, _ = http.NewRequest("GET", "/wildcard/aaa?abc=def", nil)
		router.ServeHTTP(w, r)
		if param != "aaa" {
			t.Error("Expected wildcard to match aaa, saw", param)
		}

		r, _ = http.NewRequest("GET", "/catchall/bbb?abc=def", nil)
		router.ServeHTTP(w, r)
		if param != "bbb" {
			t.Error("Expected wildcard to match bbb, saw", param)
		}
	}
}

func TestRedirectEscapedPath(t *testing.T) {
	router := New()

	testHandler := func(w http.ResponseWriter, r Request) error {
		return nil
	}

	router.GET("/:escaped/", testHandler)

	w := httptest.NewRecorder()
	u, err := url.Parse("/Test P@th")
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", u.String(), nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusMovedPermanently, w.Code)

	location := w.Header().Get("Location")
	require.Equal(t, "/Test%20P@th/", location)
}

func TestMiddleware(t *testing.T) {
	var execLog []string

	record := func(s string) {
		execLog = append(execLog, s)
	}

	newHandler := func(name string) HandlerFunc {
		return func(w http.ResponseWriter, r Request) error {
			record(name)
			return nil
		}
	}

	newMiddleware := func(name string) MiddlewareFunc {
		return func(next HandlerFunc) HandlerFunc {
			return func(w http.ResponseWriter, r Request) error {
				record(name)
				return next(w, r)
			}
		}
	}

	router := New()
	w := httptest.NewRecorder()

	// No middlewares.
	{
		router.GET("/h1", newHandler("h1"))

		req, _ := http.NewRequest("GET", "/h1", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, []string{"h1"}, execLog)
	}

	g := router.NewGroup("", WithMiddleware(newMiddleware("m1")))
	g.GET("/h2", newHandler("h2"))

	// Test route with and without middleware.
	{
		execLog = nil

		req, _ := http.NewRequest("GET", "/h1", nil)
		router.ServeHTTP(w, req)

		req, _ = http.NewRequest("GET", "/h2", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, []string{"h1", "m1", "h2"}, execLog)
	}

	// NewGroup inherits middlewares but has its own stack.
	{
		execLog = nil
		g := g.NewGroup("/g1", WithMiddleware(newMiddleware("m2")))
		g.GET("/h3", newHandler("h3"))

		req, _ := http.NewRequest("GET", "/h2", nil)
		router.ServeHTTP(w, req)

		req, _ = http.NewRequest("GET", "/g1/h3", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, []string{"m1", "h2", "m1", "m2", "h3"}, execLog)
	}

	{
		execLog = nil
		g := g.NewGroup("/g2", WithMiddleware(func(next HandlerFunc) HandlerFunc {
			return func(w http.ResponseWriter, r Request) error {
				record("m4")
				return next(w, r)
			}
		}))
		g.GET("/h6", func(w http.ResponseWriter, r Request) error {
			record("h6")
			return nil
		})

		req, _ := http.NewRequest("GET", "/g2/h6", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, []string{"m1", "m4", "h6"}, execLog)
	}

	// Middleware can serve request without calling next.
	{
		execLog = nil
		g := g.NewGroup("", WithMiddleware(func(_ HandlerFunc) HandlerFunc {
			return func(w http.ResponseWriter, r Request) error {
				record("m3")
				w.WriteHeader(http.StatusBadRequest)
				_, err := w.Write([]byte("pong"))
				return err
			}
		}))
		g.GET("/h5", newHandler("h5"))

		req, _ := http.NewRequest("GET", "/h5", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, []string{"m1", "m3"}, execLog)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("got %d, wanted %d", w.Code, http.StatusBadRequest)
		}
		if w.Body.String() != "pong" {
			t.Fatalf("got %s, wanted %s", w.Body.String(), "pong")
		}
	}
}

// When we find a node with a matching path but no handler for a method,
// we should fall through and continue searching the tree for a less specific
// match, i.e. a wildcard or catchall, that does have a handler for that method.
func TestMethodNotAllowedFallthrough(t *testing.T) {
	var matchedMethod string
	var matchedPath string
	var matchedParams map[string]string

	router := New()

	addRoute := func(method, path string) {
		router.Handle(method, path, func(w http.ResponseWriter, req Request) error {
			matchedMethod = method
			matchedPath = path
			matchedParams = req.Params().Map()
			return nil
		})
	}

	checkRoute := func(method, path, expectedMethod, expectedPath string,
		expectedCode int, expectedParams map[string]string) {
		matchedMethod = ""
		matchedPath = ""
		matchedParams = nil

		w := httptest.NewRecorder()
		r, _ := http.NewRequest(method, path, nil)
		router.ServeHTTP(w, r)
		if expectedCode != w.Code {
			t.Errorf("%s %s expected code %d, saw %d", method, path, expectedCode, w.Code)
		}

		if w.Code == 200 {
			if matchedMethod != method || matchedPath != expectedPath {
				t.Errorf("%s %s expected %s %s, saw %s %s", method, path,
					expectedMethod, expectedPath, matchedMethod, matchedPath)
			}

			if !reflect.DeepEqual(matchedParams, expectedParams) {
				t.Errorf("%s %s expected params %+v, saw %+v", method, path, expectedParams, matchedParams)
			}
		}
	}

	addRoute("GET", "/apple/banana/cat")
	addRoute("GET", "/apple/potato")
	addRoute("POST", "/apple/banana/:abc")
	addRoute("POST", "/apple/ban/def")
	addRoute("DELETE", "/apple/:seed")
	addRoute("DELETE", "/apple/*path")
	addRoute("OPTIONS", "/apple/*path")

	checkRoute("GET", "/apple/banana/cat", "GET", "/apple/banana/cat", 200, nil)
	checkRoute("POST", "/apple/banana/cat", "POST", "/apple/banana/:abc", 200,
		map[string]string{"abc": "cat"})
	checkRoute("POST", "/apple/banana/dog", "POST", "/apple/banana/:abc", 200,
		map[string]string{"abc": "dog"})

	// Wildcards should be checked before catchalls
	checkRoute("DELETE", "/apple/banana", "DELETE", "/apple/:seed", 200,
		map[string]string{"seed": "banana"})
	checkRoute("DELETE", "/apple/banana/cat", "DELETE", "/apple/*path", 200,
		map[string]string{"path": "banana/cat"})

	checkRoute("POST", "/apple/ban/def", "POST", "/apple/ban/def", 200, nil)
	checkRoute("OPTIONS", "/apple/ban/def", "OPTIONS", "/apple/*path", 200,
		map[string]string{"path": "ban/def"})
	checkRoute("GET", "/apple/ban/def", "", "", 405, nil)

	// Always fallback to the matching handler no matter how many other
	// nodes without proper handlers are found on the way.
	checkRoute("OPTIONS", "/apple/banana/cat", "OPTIONS", "/apple/*path", 200,
		map[string]string{"path": "banana/cat"})
	checkRoute("OPTIONS", "/apple/bbbb", "OPTIONS", "/apple/*path", 200,
		map[string]string{"path": "bbbb"})

	// Nothing matches on patch
	checkRoute("PATCH", "/apple/banana/cat", "", "", 405, nil)
	checkRoute("PATCH", "/apple/potato", "", "", 405, nil)

	// And some 404 tests for good measure
	checkRoute("GET", "/abc", "", "", 404, nil)
	checkRoute("OPTIONS", "/apple", "", "", 404, nil)
}
