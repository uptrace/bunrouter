package httptreemux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func simpleHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {}

func panicHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	panic("test panic")
}

func newRequest(method, path string, body io.Reader) (*http.Request, error) {
	r, _ := http.NewRequest(method, path, body)
	u, _ := url.Parse(path)
	r.URL = u
	r.RequestURI = path
	return r, nil
}

type RequestCreator func(string, string, io.Reader) (*http.Request, error)
type TestScenario struct {
	RequestCreator RequestCreator
	description    string
}

var scenarios = []TestScenario{
	TestScenario{newRequest, "Test with RequestURI"},
	TestScenario{http.NewRequest, "Test with URL.Path"},
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
	for _, scenario := range scenarios {
		t.Log(scenario.description)
		testMethods(t, scenario.RequestCreator)
	}
}

func testMethods(t *testing.T, newRequest RequestCreator) {
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
		r, _ := newRequest(method, "/user/"+method, nil)
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
	r, _ := newRequest("GET", "/abc/", nil)
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
	r, _ := newRequest("POST", "/user/abc", nil)
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
	r, _ := newRequest("GET", "/abc", nil)
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
	for _, scenario := range scenarios {
		t.Log(scenario.description)
		t.Log("Testing with all 301")
		testRedirect(t, Redirect301, Redirect301, Redirect301, false, scenario.RequestCreator)
		t.Log("Testing with all UseHandler")
		testRedirect(t, UseHandler, UseHandler, UseHandler, false, scenario.RequestCreator)
		t.Log("Testing with default 301, GET 307, POST UseHandler")
		testRedirect(t, Redirect301, Redirect307, UseHandler, true, scenario.RequestCreator)
		t.Log("Testing with default UseHandler, GET 301, POST 308")
		testRedirect(t, UseHandler, Redirect301, Redirect308, true, scenario.RequestCreator)
	}
}

func behaviorToCode(b RedirectBehavior) int {
	switch b {
	case Redirect301:
		return http.StatusMovedPermanently
	case Redirect307:
		return http.StatusTemporaryRedirect
	case Redirect308:
		return 308
	case UseHandler:
		// Not normally, but the handler in the below test returns this.
		return http.StatusNoContent
	}

	panic("Unhandled behavior!")
}

func testRedirect(t *testing.T, defaultBehavior, getBehavior, postBehavior RedirectBehavior, customMethods bool,
	newRequest RequestCreator) {

	var redirHandler = func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		// Returning this instead of 200 makes it easy to verify that the handler is actually getting called.
		w.WriteHeader(http.StatusNoContent)
	}

	router := New()
	router.RedirectBehavior = defaultBehavior

	var expectedCodeMap = map[string]int{"PUT": behaviorToCode(defaultBehavior)}

	if customMethods {
		router.RedirectMethodBehavior["GET"] = getBehavior
		router.RedirectMethodBehavior["POST"] = postBehavior
		expectedCodeMap["GET"] = behaviorToCode(getBehavior)
		expectedCodeMap["POST"] = behaviorToCode(postBehavior)
	} else {
		expectedCodeMap["GET"] = expectedCodeMap["PUT"]
		expectedCodeMap["POST"] = expectedCodeMap["PUT"]
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
		r, _ := newRequest(method, "/slash", nil)
		router.ServeHTTP(w, r)
		if w.Code != expectedCode {
			t.Errorf("/slash expected code %d, saw %d", expectedCode, w.Code)
		}
		if expectedCode != http.StatusNoContent && w.Header().Get("Location") != "/slash/" {
			t.Errorf("/slash was not redirected to /slash/")
		}

		r, _ = newRequest(method, "/noslash/", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != expectedCode {
			t.Errorf("/noslash/ expected code %d, saw %d", expectedCode, w.Code)
		}
		if expectedCode != http.StatusNoContent && w.Header().Get("Location") != "/noslash" {
			t.Errorf("/noslash/ was not redirected to /noslash")
		}

		r, _ = newRequest(method, "//noslash/", nil)
		if r.RequestURI == "//noslash/" { // http.NewRequest parses this out differently
			w = httptest.NewRecorder()
			router.ServeHTTP(w, r)
			if w.Code != expectedCode {
				t.Errorf("//noslash/ expected code %d, saw %d", expectedCode, w.Code)
			}
			if expectedCode != http.StatusNoContent && w.Header().Get("Location") != "/noslash" {
				t.Errorf("//noslash/ was not redirected to /noslash")
			}
		}

		// Test nonredirect cases
		r, _ = newRequest(method, "/noslash", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != http.StatusNoContent {
			t.Errorf("/noslash (non-redirect) expected code %d, saw %d", http.StatusNoContent, w.Code)
		}

		r, _ = newRequest(method, "/slash/", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, r)
		if w.Code != http.StatusNoContent {
			t.Errorf("/slash/ (non-redirect) expected code %d, saw %d", http.StatusNoContent, w.Code)
		}
	}
}

func TestSkipRedirect(t *testing.T) {
	router := New()
	router.RedirectTrailingSlash = false
	router.RedirectCleanPath = false
	router.GET("/slash/", simpleHandler)
	router.GET("/noslash", simpleHandler)

	w := httptest.NewRecorder()
	r, _ := newRequest("GET", "/slash", nil)
	router.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Errorf("/slash expected code 404, saw %d", w.Code)
	}

	r, _ = newRequest("GET", "/noslash/", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Errorf("/noslash/ expected code 404, saw %d", w.Code)
	}

	r, _ = newRequest("GET", "//noslash", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Errorf("//noslash expected code 404, saw %d", w.Code)
	}
}

func TestCatchAllTrailingSlashRedirect(t *testing.T) {
	router := New()
	redirectSettings := []bool{false, true}

	router.GET("/abc/*path", simpleHandler)

	testPath := func(path string) {
		r, _ := newRequest("GET", "/abc/"+path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		endingSlash := strings.HasSuffix(path, "/")

		var expectedCode int
		if endingSlash && router.RedirectTrailingSlash && router.RemoveCatchAllTrailingSlash {
			expectedCode = http.StatusMovedPermanently
		} else {
			expectedCode = http.StatusOK
		}

		if w.Code != expectedCode {
			t.Errorf("Path %s with RedirectTrailingSlash %v, RemoveCatchAllTrailingSlash %v "+
				" expected code %d but saw %d", path,
				router.RedirectTrailingSlash, router.RemoveCatchAllTrailingSlash,
				expectedCode, w.Code)
		}
	}

	for _, redirectSetting := range redirectSettings {
		for _, removeCatchAllSlash := range redirectSettings {
			router.RemoveCatchAllTrailingSlash = removeCatchAllSlash
			router.RedirectTrailingSlash = redirectSetting

			testPath("apples")
			testPath("apples/")
			testPath("apples/bananas")
			testPath("apples/bananas/")
		}
	}

}

func TestRoot(t *testing.T) {
	for _, scenario := range scenarios {
		t.Log(scenario.description)
		handlerCalled := false
		handler := func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			handlerCalled = true
		}
		router := New()
		router.GET("/", handler)

		r, _ := scenario.RequestCreator("GET", "/", nil)
		w := new(mockResponseWriter)
		router.ServeHTTP(w, r)

		if !handlerCalled {
			t.Error("Handler not called for root path")
		}
	}
}

func TestWildcardAtSplitNode(t *testing.T) {
	var suppliedParam string
	simpleHandler := func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		suppliedParam, _ = params["slug"]
	}

	router := New()
	router.GET("/pumpkin", simpleHandler)
	router.GET("/passing", simpleHandler)
	router.GET("/:slug", simpleHandler)
	router.GET("/:slug/abc", simpleHandler)

	r, _ := newRequest("GET", "/patch", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if suppliedParam != "patch" {
		t.Errorf("Expected param patch, saw %s", suppliedParam)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for path /patch, saw %d", w.Code)
	}

	suppliedParam = ""
	r, _ = newRequest("GET", "/patch/abc", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if suppliedParam != "patch" {
		t.Errorf("Expected param patch, saw %s", suppliedParam)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for path /patch/abc, saw %d", w.Code)
	}

	r, _ = newRequest("GET", "/patch/def", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for path /patch/def, saw %d", w.Code)
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

	r, _ := newRequest("GET", "/abc/de%2ff", nil)
	w := new(mockResponseWriter)
	router.ServeHTTP(w, r)

	if param != "de/f" {
		t.Errorf("Expected param de/f, saw %s", param)
	}

	r, _ = newRequest("GET", "/year/de%2f/month/fg%2f", nil)
	router.ServeHTTP(w, r)

	if param != "de/ fg/" {
		t.Errorf("Expected param de/ fg/, saw %s", param)
	}
}

func TestQueryString(t *testing.T) {
	for _, scenario := range scenarios {
		t.Log(scenario.description)
		param := ""
		handler := func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			param = params["param"]
		}
		router := New()
		router.GET("/static", handler)
		router.GET("/wildcard/:param", handler)
		router.GET("/catchall/*param", handler)

		r, _ := scenario.RequestCreator("GET", "/static?abc=def&ghi=jkl", nil)
		w := new(mockResponseWriter)

		param = "nomatch"
		router.ServeHTTP(w, r)
		if param != "" {
			t.Error("No match on", r.RequestURI)
		}

		r, _ = scenario.RequestCreator("GET", "/wildcard/aaa?abc=def", nil)
		router.ServeHTTP(w, r)
		if param != "aaa" {
			t.Error("Expected wildcard to match aaa, saw", param)
		}

		r, _ = scenario.RequestCreator("GET", "/catchall/bbb?abc=def", nil)
		router.ServeHTTP(w, r)
		if param != "bbb" {
			t.Error("Expected wildcard to match bbb, saw", param)
		}
	}
}

func BenchmarkRouterSimple(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/dimfeld", simpleHandler)

	r, _ := newRequest("GET", "/user/dimfeld", nil)

	benchRequest(b, router, r)
}

func BenchmarkRouterRootWithPanicHandler(b *testing.B) {
	router := New()
	router.PanicHandler = SimplePanicHandler

	router.GET("/", simpleHandler)
	router.GET("/user/dimfeld", simpleHandler)

	r, _ := newRequest("GET", "/", nil)

	benchRequest(b, router, r)
}

func BenchmarkRouterRootWithoutPanicHandler(b *testing.B) {
	router := New()
	router.PanicHandler = nil

	router.GET("/", simpleHandler)
	router.GET("/user/dimfeld", simpleHandler)

	r, _ := newRequest("GET", "/", nil)

	benchRequest(b, router, r)
}

func BenchmarkRouterParam(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/:name", simpleHandler)

	r, _ := newRequest("GET", "/user/dimfeld", nil)

	benchRequest(b, router, r)
}
