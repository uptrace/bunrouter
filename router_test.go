package treemux

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func simpleHandler(w http.ResponseWriter, r Request) error {
	return nil
}

func newRequest(method, path string, body io.Reader) (*http.Request, error) {
	r, _ := http.NewRequest(method, path, body)
	u, _ := url.ParseRequestURI(path)
	r.URL = u
	r.RequestURI = path
	return r, nil
}

type RequestCreator func(string, string, io.Reader) (*http.Request, error)

type TestScenario struct {
	RequestCreator RequestCreator
	ServeStyle     bool
	description    string
}

var scenarios = []TestScenario{
	TestScenario{newRequest, false, "Test with RequestURI and normal ServeHTTP"},
	TestScenario{http.NewRequest, false, "Test with URL.Path and normal ServeHTTP"},
	TestScenario{newRequest, true, "Test with RequestURI and LookupResult"},
	TestScenario{http.NewRequest, true, "Test with URL.Path and LookupResult"},
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

func serve(router *TreeMux, w http.ResponseWriter, r *http.Request, useLookup bool) bool {
	if useLookup {
		result, found := router.Lookup(w, r)
		router.ServeLookupResult(w, r, result)
		return found
	} else {
		router.ServeHTTP(w, r)
		return true
	}
}

func TestMethods(t *testing.T) {
	for _, scenario := range scenarios {
		t.Log(scenario.description)
		testMethods(t, scenario.RequestCreator, true, scenario.ServeStyle)
		testMethods(t, scenario.RequestCreator, false, scenario.ServeStyle)
	}
}

func testMethods(t *testing.T, newRequest RequestCreator, headCanUseGet bool, useSeparateLookup bool) {
	var result string

	makeHandler := func(method string) HandlerFunc {
		return func(w http.ResponseWriter, r Request) error {
			result = method
			return nil
		}
	}

	router := New(WithHeadCanUseGet(headCanUseGet))
	router.GET("/user/:param", makeHandler("GET"))
	router.POST("/user/:param", makeHandler("POST"))
	router.PATCH("/user/:param", makeHandler("PATCH"))
	router.PUT("/user/:param", makeHandler("PUT"))
	router.DELETE("/user/:param", makeHandler("DELETE"))

	testMethod := func(method, expect string) {
		result = ""
		w := httptest.NewRecorder()
		r, _ := newRequest(method, "/user/"+method, nil)
		found := serve(router, w, r, useSeparateLookup)

		if useSeparateLookup && expect == "" && found {
			t.Errorf("Lookup unexpectedly succeeded for method %s", method)
		}

		if expect == "" && w.Code != http.StatusMethodNotAllowed {
			t.Errorf("Method %s not expected to match but saw code %d", method, w.Code)
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
	if headCanUseGet {
		t.Log("Test implicit HEAD with HeadCanUseGet = true")
		testMethod("HEAD", "GET")
	} else {
		t.Log("Test implicit HEAD with HeadCanUseGet = false")
		testMethod("HEAD", "")
	}

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
	r, _ := newRequest("GET", "/abc/", nil)
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
		t.Log("Testing with all 301")
		testRedirect(t, Redirect301, Redirect301, Redirect301, false, scenario.RequestCreator, scenario.ServeStyle)
		t.Log("Testing with all UseHandler")
		testRedirect(t, UseHandler, UseHandler, UseHandler, false, scenario.RequestCreator, scenario.ServeStyle)
		t.Log("Testing with default 301, GET 307, POST UseHandler")
		testRedirect(t, Redirect301, Redirect307, UseHandler, true, scenario.RequestCreator, scenario.ServeStyle)
		t.Log("Testing with default UseHandler, GET 301, POST 308")
		testRedirect(t, UseHandler, Redirect301, Redirect308, true, scenario.RequestCreator, scenario.ServeStyle)
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
	newRequest RequestCreator, serveStyle bool) {
	redirHandler := func(w http.ResponseWriter, r Request) error {
		// Returning this instead of 200 makes it easy to verify that the handler is actually getting called.
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	expectedCodeMap := map[string]int{"PUT": behaviorToCode(defaultBehavior)}
	var router *TreeMux

	if customMethods {
		router = New(
			WithRedirectBehavior(defaultBehavior),
			WithRedirectMethodBehavior(map[string]RedirectBehavior{
				"GET":  getBehavior,
				"POST": postBehavior,
			}),
		)

		expectedCodeMap["GET"] = behaviorToCode(getBehavior)
		expectedCodeMap["POST"] = behaviorToCode(postBehavior)
	} else {
		router = New(
			WithRedirectBehavior(defaultBehavior),
		)
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
		found := serve(router, w, r, serveStyle)
		if found == false {
			t.Errorf("/slash: found returned false")
		}
		if w.Code != expectedCode {
			t.Errorf("/slash expected code %d, saw %d", expectedCode, w.Code)
		}
		if expectedCode != http.StatusNoContent && w.Header().Get("Location") != "/slash/" {
			t.Errorf("/slash was not redirected to /slash/")
		}

		r, _ = newRequest(method, "/noslash/", nil)
		w = httptest.NewRecorder()
		found = serve(router, w, r, serveStyle)
		if found == false {
			t.Errorf("/noslash: found returned false")
		}
		if w.Code != expectedCode {
			t.Errorf("/noslash/ expected code %d, saw %d", expectedCode, w.Code)
		}
		if expectedCode != http.StatusNoContent && w.Header().Get("Location") != "/noslash" {
			t.Errorf("/noslash/ was redirected to `%s` instead of /noslash", w.Header().Get("Location"))
		}

		r, _ = newRequest(method, "//noslash/", nil)
		if r.RequestURI == "//noslash/" { // http.NewRequest parses this out differently
			w = httptest.NewRecorder()
			found = serve(router, w, r, serveStyle)
			if found == false {
				t.Errorf("//noslash/: found returned false")
			}
			if w.Code != expectedCode {
				t.Errorf("//noslash/ expected code %d, saw %d", expectedCode, w.Code)
			}
			if expectedCode != http.StatusNoContent && w.Header().Get("Location") != "/noslash" {
				t.Errorf("//noslash/ was redirected to %s, expected /noslash", w.Header().Get("Location"))
			}
		}

		// Test nonredirect cases
		r, _ = newRequest(method, "/noslash", nil)
		w = httptest.NewRecorder()
		found = serve(router, w, r, serveStyle)
		if found == false {
			t.Errorf("/noslash (non-redirect): found returned false")
		}
		if w.Code != http.StatusNoContent {
			t.Errorf("/noslash (non-redirect) expected code %d, saw %d", http.StatusNoContent, w.Code)
		}

		r, _ = newRequest(method, "/noslash?a=1&b=2", nil)
		w = httptest.NewRecorder()
		found = serve(router, w, r, serveStyle)
		if found == false {
			t.Errorf("/noslash (non-redirect): found returned false")
		}
		if w.Code != http.StatusNoContent {
			t.Errorf("/noslash (non-redirect) expected code %d, saw %d", http.StatusNoContent, w.Code)
		}

		r, _ = newRequest(method, "/slash/", nil)
		w = httptest.NewRecorder()
		found = serve(router, w, r, serveStyle)
		if found == false {
			t.Errorf("/slash/ (non-redirect): found returned false")
		}
		if w.Code != http.StatusNoContent {
			t.Errorf("/slash/ (non-redirect) expected code %d, saw %d", http.StatusNoContent, w.Code)
		}

		r, _ = newRequest(method, "/slash/?a=1&b=2", nil)
		w = httptest.NewRecorder()
		found = serve(router, w, r, serveStyle)
		if found == false {
			t.Errorf("/slash/?a=1&b=2: found returned false")
		}
		if w.Code != http.StatusNoContent {
			t.Errorf("/slash/?a=1&b=2 expected code %d, saw %d", http.StatusNoContent, w.Code)
		}

		// Test querystring and fragment cases
		r, _ = newRequest(method, "/slash?a=1&b=2", nil)
		w = httptest.NewRecorder()
		found = serve(router, w, r, serveStyle)
		if found == false {
			t.Errorf("/slash?a=1&b=2 : found returned false")
		}
		if w.Code != expectedCode {
			t.Errorf("/slash?a=1&b=2 expected code %d, saw %d", expectedCode, w.Code)
		}
		if expectedCode != http.StatusNoContent && w.Header().Get("Location") != "/slash/?a=1&b=2" {
			t.Errorf("/slash?a=1&b=2 was redirected to %s", w.Header().Get("Location"))
		}

		r, _ = newRequest(method, "/noslash/?a=1&b=2", nil)
		w = httptest.NewRecorder()
		found = serve(router, w, r, serveStyle)
		if found == false {
			t.Errorf("/noslash/?a=1&b=2: found returned false")
		}
		if w.Code != expectedCode {
			t.Errorf("/noslash/?a=1&b=2 expected code %d, saw %d", expectedCode, w.Code)
		}
		if expectedCode != http.StatusNoContent && w.Header().Get("Location") != "/noslash?a=1&b=2" {
			t.Errorf("/noslash/?a=1&b=2 was redirected to %s", w.Header().Get("Location"))
		}
	}
}

func TestSkipRedirect(t *testing.T) {
	router := New(
		WithRedirectTrailingSlash(false),
		WithRedirectCleanPath(false),
	)

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
	var router *TreeMux

	testPath := func(path string) {
		r, _ := newRequest("GET", "/abc/"+path, nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		endingSlash := strings.HasSuffix(path, "/")

		var expectedCode int
		if endingSlash && router.redirectTrailingSlash && router.removeCatchAllTrailingSlash {
			expectedCode = http.StatusMovedPermanently
		} else {
			expectedCode = http.StatusOK
		}

		if w.Code != expectedCode {
			t.Errorf("Path %s with RedirectTrailingSlash %v, RemoveCatchAllTrailingSlash %v "+
				" expected code %d but saw %d", path,
				router.redirectTrailingSlash, router.removeCatchAllTrailingSlash,
				expectedCode, w.Code)
		}
	}

	redirectSettings := []bool{false, true}
	for _, redirectSetting := range redirectSettings {
		for _, removeCatchAllSlash := range redirectSettings {
			router = New(
				WithRemoveCatchAllTrailingSlash(removeCatchAllSlash),
				WithRedirectTrailingSlash(redirectSetting),
			)
			router.GET("/abc/*path", simpleHandler)

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
		handler := func(w http.ResponseWriter, r Request) error {
			handlerCalled = true
			return nil
		}
		router := New()
		router.GET("/", handler)

		r, _ := scenario.RequestCreator("GET", "/", nil)
		w := new(mockResponseWriter)
		serve(router, w, r, scenario.ServeStyle)

		if !handlerCalled {
			t.Error("Handler not called for root path")
		}
	}
}

func TestWildcardAtSplitNode(t *testing.T) {
	var suppliedParam string
	simpleHandler := func(w http.ResponseWriter, r Request) error {
		t.Log(r.Params.Map())
		suppliedParam, _ = r.Params.Get("slug")
		return nil
	}

	router := New()
	router.GET("/pumpkin", simpleHandler)
	router.GET("/passing", simpleHandler)
	router.GET("/:slug", simpleHandler)
	router.GET("/:slug/abc", simpleHandler)

	t.Log(router.root.dumpTree("", " "))

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
	handler := func(w http.ResponseWriter, r Request) error {
		param = r.Params.Text("param")
		return nil
	}
	ymHandler := func(w http.ResponseWriter, r Request) error {
		param = r.Params.Text("year") + " " + r.Params.Text("month")
		return nil
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
		handler := func(w http.ResponseWriter, r Request) error {
			param = r.Params.Text("param")
			return nil
		}
		router := New()
		router.GET("/static", handler)
		router.GET("/wildcard/:param", handler)
		router.GET("/catchall/*param", handler)

		r, _ := scenario.RequestCreator("GET", "/static?abc=def&ghi=jkl", nil)
		w := new(mockResponseWriter)

		param = "nomatch"
		serve(router, w, r, scenario.ServeStyle)
		if param != "" {
			t.Error("No match on", r.RequestURI)
		}

		r, _ = scenario.RequestCreator("GET", "/wildcard/aaa?abc=def", nil)
		serve(router, w, r, scenario.ServeStyle)
		if param != "aaa" {
			t.Error("Expected wildcard to match aaa, saw", param)
		}

		r, _ = scenario.RequestCreator("GET", "/catchall/bbb?abc=def", nil)
		serve(router, w, r, scenario.ServeStyle)
		if param != "bbb" {
			t.Error("Expected wildcard to match bbb, saw", param)
		}
	}
}

func TestPathSource(t *testing.T) {
	var called string

	appleHandler := func(w http.ResponseWriter, r Request) error {
		called = "apples"
		return nil
	}

	bananaHandler := func(w http.ResponseWriter, r Request) error {
		called = "bananas"
		return nil
	}

	var router *TreeMux

	newRouter := func(opts ...Option) {
		router = New(opts...)
		router.GET("/apples", appleHandler)
		router.GET("/bananas", bananaHandler)
	}

	// Set up a request with different values in URL and RequestURI.
	r, _ := newRequest("GET", "/apples", nil)
	r.RequestURI = "/bananas"
	w := new(mockResponseWriter)

	// Default setting should be RequestURI
	newRouter()
	router.ServeHTTP(w, r)
	if called != "bananas" {
		t.Error("Using default, expected bananas but saw", called)
	}

	newRouter(WithPathSource(URLPath))
	router.ServeHTTP(w, r)
	if called != "apples" {
		t.Error("Using URLPath, expected apples but saw", called)
	}

	newRouter(WithPathSource(RequestURI))
	router.ServeHTTP(w, r)
	if called != "bananas" {
		t.Error("Using RequestURI, expected bananas but saw", called)
	}
}

// Create a bunch of paths for testing.
func createRoutes(numRoutes int) []string {
	letters := "abcdefghijhklmnopqrstuvwxyz"
	wordMap := map[string]bool{}
	for i := 0; i < numRoutes/2; i += 1 {
		length := (i % 4) + 4

		wordBytes := make([]byte, length)
		for charIndex := 0; charIndex < length; charIndex += 1 {
			wordBytes[charIndex] = letters[(i*3+charIndex*4)%len(letters)]
		}
		wordMap[string(wordBytes)] = true
	}

	words := make([]string, 0, len(wordMap))
	for word := range wordMap {
		words = append(words, word)
	}

	routes := make([]string, 0, numRoutes)
	createdRoutes := map[string]bool{}
	rand.Seed(0)
	for len(routes) < numRoutes {
		first := words[rand.Int()%len(words)]
		second := words[rand.Int()%len(words)]
		third := words[rand.Int()%len(words)]
		route := fmt.Sprintf("/%s/%s/%s", first, second, third)

		if createdRoutes[route] {
			continue
		}
		createdRoutes[route] = true
		routes = append(routes, route)
	}

	return routes
}

func TestLookup(t *testing.T) {
	var router *TreeMux

	newRouter := func(opts ...Option) {
		router = New(opts...)
		router.GET("/", simpleHandler)
		router.GET("/user/dimfeld", simpleHandler)
		router.POST("/user/dimfeld", simpleHandler)
		router.GET("/abc/*", simpleHandler)
		router.POST("/abc/*", simpleHandler)
	}

	tryLookup := func(method, path string, expectFound bool, expectCode int) {
		r, _ := newRequest(method, path, nil)
		w := &mockResponseWriter{}
		lr, found := router.Lookup(w, r)
		if found != expectFound {
			t.Errorf("%s %s expected found %v, saw %v", method, path, expectFound, found)
		}

		if lr.StatusCode != expectCode {
			t.Errorf("%s %s expected status code %d, saw %d", method, path, expectCode, lr.StatusCode)
		}

		if w.code != 0 {
			t.Errorf("%s %s unexpectedly wrote status %d", method, path, w.code)
		}

		if w.calledWrite {
			t.Errorf("%s %s unexpectedly wrote data", method, path)
		}
	}

	newRouter()
	tryLookup("GET", "/", true, http.StatusOK)
	tryLookup("GET", "/", true, http.StatusOK)
	tryLookup("POST", "/user/dimfeld", true, http.StatusOK)
	tryLookup("POST", "/user/dimfeld/", true, http.StatusMovedPermanently)
	tryLookup("PATCH", "/user/dimfeld", false, http.StatusMethodNotAllowed)
	tryLookup("GET", "/abc/def/ghi", true, http.StatusOK)

	newRouter(WithRedirectBehavior(Redirect307))
	tryLookup("POST", "/user/dimfeld/", true, http.StatusTemporaryRedirect)
}

func TestRedirectEscapedPath(t *testing.T) {
	router := New()

	testHandler := func(w http.ResponseWriter, r Request) error {
		return nil
	}

	router.GET("/:escaped/", testHandler)

	w := httptest.NewRecorder()
	u, err := url.Parse("/Test P@th")
	if err != nil {
		t.Error(err)
		return
	}

	r, _ := newRequest("GET", u.String(), nil)

	router.ServeHTTP(w, r)

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("Expected status 301 but saw %d", w.Code)
	}

	path := w.Header().Get("Location")
	expected := "/Test%20P@th/"
	if path != expected {
		t.Errorf("Given path wasn't escaped correctly.\n"+
			"Expected: %q\nBut got: %q", expected, path)
	}
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

		req, _ := newRequest("GET", "/h1", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, []string{"h1"}, execLog)
	}

	g := router.NewGroup("", WithMiddleware(newMiddleware("m1")))
	g.GET("/h2", newHandler("h2"))

	// Test route with and without middleware.
	{
		execLog = nil

		req, _ := newRequest("GET", "/h1", nil)
		router.ServeHTTP(w, req)

		req, _ = newRequest("GET", "/h2", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, []string{"h1", "m1", "h2"}, execLog)
	}

	// NewGroup inherits middlewares but has its own stack.
	{
		execLog = nil
		g := g.NewGroup("/g1", WithMiddleware(newMiddleware("m2")))
		g.GET("/h3", newHandler("h3"))

		req, _ := newRequest("GET", "/h2", nil)
		router.ServeHTTP(w, req)

		req, _ = newRequest("GET", "/g1/h3", nil)
		router.ServeHTTP(w, req)

		require.Equal(t, []string{"m1", "h2", "m1", "m2", "h3"}, execLog)
	}

	// Middleware can modify params.
	{
		execLog = nil
		g := g.NewGroup("/g2", WithMiddleware(func(next HandlerFunc) HandlerFunc {
			return func(w http.ResponseWriter, r Request) error {
				record("m4")
				r.Params = append(r.Params, Param{
					Name:  "foo",
					Value: "bar",
				})
				return next(w, r)
			}
		}))
		g.GET("/h6", func(w http.ResponseWriter, r Request) error {
			record("h6")
			if v := r.Params.Text("foo"); v != "bar" {
				t.Fatalf("got %q, wanted %q", v, "bar")
			}
			return nil
		})

		req, _ := newRequest("GET", "/g2/h6", nil)
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

		req, _ := newRequest("GET", "/h5", nil)
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

func BenchmarkRouterSimple(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/dimfeld", simpleHandler)

	r, _ := newRequest("GET", "/user/dimfeld", nil)

	benchRequest(b, router, r)
}

func BenchmarkRouterRoot(b *testing.B) {
	router := New()

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

func BenchmarkRouterLongParams(b *testing.B) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/user/:name/:resource", simpleHandler)

	r, _ := newRequest("GET", "/user/aaaabbbbccccddddeeeeffff/asdfghjkl", nil)

	benchRequest(b, router, r)
}
