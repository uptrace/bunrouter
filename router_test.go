package bunrouter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func simpleHandler(w http.ResponseWriter, req Request) error {
	return nil
}

type TestScenario struct {
	description string
}

var scenarios = []TestScenario{
	{"Test with URL.Path and normal ServeHTTP"},
}

func TestRequestWithContext(t *testing.T) {
	router := New()
	router.GET("/user/:param", func(w http.ResponseWriter, req Request) error {
		value1 := req.Param("param")
		require.Equal(t, "hello", value1)

		value2 := req.WithContext(context.TODO()).Param("param")
		require.Equal(t, value1, value2)

		return nil
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/user/hello", nil)
	router.ServeHTTP(w, req)
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
		result = "" // reset

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
	var calledNotFound int

	notFoundHandler := func(w http.ResponseWriter, req Request) error {
		calledNotFound++
		return nil
	}

	router := New()
	router.GET("/user/abc", simpleHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/abc/", nil)

	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
	require.Equal(t, 0, calledNotFound)

	// Now try with a custome handler.
	router = New(WithNotFoundHandler(notFoundHandler))
	router.GET("/user/abc", simpleHandler)

	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusNotFound, w.Code)
	require.Equal(t, 1, calledNotFound)
}

func TestMethodNotAllowed(t *testing.T) {
	var calledMethodNotAllowed int

	methodNotAllowedHandler := func(w http.ResponseWriter, req Request) error {
		calledMethodNotAllowed++
		return nil
	}

	router := New()
	router.POST("/abc", simpleHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/abc", nil)

	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.Equal(t, 0, calledMethodNotAllowed)

	// Now try with a custome handler.
	router = New(WithMethodNotAllowedHandler(methodNotAllowedHandler))
	router.POST("/abc", simpleHandler)

	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusMethodNotAllowed, w.Code)
	require.Equal(t, 1, calledMethodNotAllowed)
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

func TestRouteWithNamedAndWildcardParams(t *testing.T) {
	t.Run("without static nodes", func(t *testing.T) {
		router := New()

		var params map[string]string
		router.GET("/:id/*path", func(w http.ResponseWriter, req Request) error {
			params = req.Params().Map()
			return nil
		})

		t.Run("with path", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/123/hello/world", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, map[string]string{
				"id":   "123",
				"path": "hello/world",
			}, params)
		})

		t.Run("without path", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/123/", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, map[string]string{
				"id":   "123",
				"path": "",
			}, params)
		})
	})

	t.Run("with static nodes", func(t *testing.T) {
		router := New()

		var params map[string]string
		router.GET("/:id/foo/*path", func(w http.ResponseWriter, req Request) error {
			params = req.Params().Map()
			return nil
		})

		t.Run("with path", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/123/foo/hello/world", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, map[string]string{
				"id":   "123",
				"path": "hello/world",
			}, params)
		})

		t.Run("without path", func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/123/foo/", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, map[string]string{
				"id":   "123",
				"path": "",
			}, params)
		})
	})
}

func TestQueryString(t *testing.T) {
	var param string

	handler := func(w http.ResponseWriter, r Request) error {
		param = r.Params().ByName("param")
		return nil
	}

	router := New()
	router.GET("/static", handler)
	router.GET("/named/:param", handler)
	router.GET("/wildcard/*param", handler)

	w := httptest.NewRecorder()

	req, _ := http.NewRequest("GET", "/static?abc=def&ghi=jkl", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, "", param)

	req, _ = http.NewRequest("GET", "/named/aaa?abc=def", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, "aaa", param)

	req, _ = http.NewRequest("GET", "/wildcard/bbb?abc=def", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, "bbb", param)
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

func TestMiddlewares(t *testing.T) {
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

func TestCORSMiddleware(t *testing.T) {
	corsMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, req Request) error {
			if req.Method == http.MethodOptions {
				return nil
			}
			return next(w, req)
		}
	}

	router := New()

	router.NewGroup("/api",
		// Install CORS only for this group.
		WithMiddleware(corsMiddleware),
		WithGroup(func(g *Group) {
			g.GET("/users", simpleHandler)
		}))

	t.Run("normal request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/users", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("CORS request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodOptions, "/api/users", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("CORS to a non-existant route", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodOptions, "/api", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("not allowed method", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/users", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
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
	checkRoute("OPTIONS", "/apple", "", "", 301, nil)
}

func TestWildcardNode(t *testing.T) {
	var route string
	var params map[string]string

	handler := func(w http.ResponseWriter, req Request) error {
		route = req.Params().Route()
		params = req.Params().Map()
		return nil
	}

	router := New()
	router.GET("/*path", handler)
	router.GET("/static/*path", handler)

	type Test struct {
		path   string
		params map[string]string
	}

	for _, test := range []Test{
		{"/", map[string]string{"path": ""}},
		{"/foo", map[string]string{"path": "foo"}},
		{"/foo/bar", map[string]string{"path": "foo/bar"}},
		{"/static", map[string]string{"path": "static"}},
	} {
		t.Run(fmt.Sprintf("path=%s", test.path), func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, test.path, nil)
			router.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, "/*path", route)
			require.Equal(t, test.params, params)
		})
	}

	for _, test := range []Test{
		{"/static/", map[string]string{"path": ""}},
		{"/static/foo", map[string]string{"path": "foo"}},
		{"/static/foo/bar", map[string]string{"path": "foo/bar"}},
	} {
		t.Run(fmt.Sprintf("path=%s", test.path), func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, test.path, nil)
			router.ServeHTTP(w, req)

			require.Equal(t, http.StatusOK, w.Code)
			require.Equal(t, "/static/*path", route)
			require.Equal(t, test.params, params)
		})
	}
}

func TestFiveColonRoute(t *testing.T) {
	router := New()

	router.GET("/", simpleHandler)
	router.GET("/:a/:b/:c/:d/:e", simpleHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test/test/test/test/test", nil)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)
}

func TestRoutesWithCommonPrefix(t *testing.T) {
	router := New()

	router.GET("/campaigns", simpleHandler)
	router.GET("/causes", simpleHandler)

	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/ca", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code)
	}

	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code)
	}
}

func TestMethodNotAllowedWithMiddlewares(t *testing.T) {
	var stack []string

	middleware := func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, req Request) error {
			stack = append(stack, "middleware")
			return next(w, req)
		}
	}

	handler := func(w http.ResponseWriter, req Request) error {
		stack = append(stack, "handler")
		return nil
	}

	router := New()

	router.NewGroup("/hello",
		WithMiddleware(middleware),
		WithGroup(func(group *Group) {
			group.GET("", handler)
		}),
	)

	t.Run("existing route", func(t *testing.T) {
		stack = nil
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hello", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, []string{"middleware", "handler"}, stack)
	})

	t.Run("not allowed method", func(t *testing.T) {
		stack = nil
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/hello", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusMethodNotAllowed, w.Code)
		require.Equal(t, []string{"middleware"}, stack)
	})

	t.Run("not found route", func(t *testing.T) {
		stack = nil
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hello/world", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusNotFound, w.Code)
		require.Nil(t, stack)
	})
}

func TestNamedAndWildcard(t *testing.T) {
	router := New()

	router.GET("/api/internal", func(w http.ResponseWriter, req Request) error {
		require.Equal(t, "/api/internal", req.Route())
		return nil
	})
	router.GET("/api/internal/*params", func(w http.ResponseWriter, req Request) error {
		require.Equal(t, "/api/internal/*params", req.Route())
		return nil
	})

	t.Run("named route", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/internal", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("empty wildcard route", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/internal/", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("non-empty wildcard route", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/internal/foo/bar", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	})
}

// https://github.com/golang/go/issues/3659
func TestURIEncodedFilepath(t *testing.T) {
	router := New()
	var file string

	router.GET("/files/:file", func(w http.ResponseWriter, req Request) error {
		file = req.Param("file")
		return nil
	})

	t.Run("one", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/files/foo%252fbar", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "foo%2fbar", file)
	})

	t.Run("two", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/files/foo%2fbar", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
		require.Equal(t, "foo%2fbar", file)
	})
}

func TestSplitRoute(t *testing.T) {
	type Test struct {
		route  string
		parts  []string
		params map[string]int
	}

	tests := []Test{
		{"/", []string{}, nil},
		{"/static", []string{"static"}, nil},
		{"/static/", []string{"static/"}, nil},
		{"/static/foo", []string{"static/foo"}, nil},
		{"/static/:foo", []string{"static/", ":"}, map[string]int{"foo": 0}},
		{"/static/:foo/bar", []string{"static/", ":", "/bar"}, map[string]int{"foo": 0}},
		{"/static/*path", []string{"static/", "*"}, map[string]int{"path": 0}},
		{"/*path", []string{"*"}, map[string]int{"path": 0}},
		{"/:foo/*path", []string{":", "/", "*"}, map[string]int{"foo": 0, "path": 1}},
		{"/:foo/static/*path", []string{":", "/static/", "*"}, map[string]int{"foo": 0, "path": 1}},
		{
			"/:a/:b/:c/:d/:e",
			[]string{":", "/", ":", "/", ":", "/", ":", "/", ":"},
			map[string]int{"a": 0, "b": 1, "c": 2, "d": 3, "e": 4},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("route=%s", test.route), func(t *testing.T) {
			parts, params := splitRoute(test.route)
			require.Equal(t, test.parts, parts)
			require.Equal(t, test.params, params)
		})
	}
}

func TestMultipleMiddlewaresAndMethodNotAllowed(t *testing.T) {
	firstMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, req Request) error {
			w.Header().Add("First", "xxxx")
			return next(w, req)
		}
	}

	secondMiddleware := func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, req Request) error {
			w.Header().Add("Second", "xxxx")
			return next(w, req)
		}
	}

	indexHandler := func(w http.ResponseWriter, req Request) error {
		return JSON(w, H{
			"message": "Hello, World!",
		})
	}

	t.Run("router", func(t *testing.T) {
		router := New(
			Use(firstMiddleware, secondMiddleware),
		)
		router.GET("/", indexHandler)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusMethodNotAllowed, w.Code)

		h := w.Header()
		require.Equal(t, []string{"xxxx"}, h["First"])
	})

	t.Run("group", func(t *testing.T) {
		router := New()
		router.
			Use(firstMiddleware, secondMiddleware).
			WithGroup("/", func(group *Group) {
				group.GET("", indexHandler)
			})

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusMethodNotAllowed, w.Code)

		h := w.Header()
		require.Equal(t, []string{"xxxx"}, h["First"])
	})
}

func TestSameRouteWithDifferentParams(t *testing.T) {
	router := New()

	router.GET("/:foo", func(w http.ResponseWriter, req Request) error {
		require.Equal(t, map[string]string{"foo": "hello"}, req.Params().Map())
		return nil
	})
	router.HEAD("/:bar", func(w http.ResponseWriter, req Request) error {
		require.Equal(t, map[string]string{"bar": "hello"}, req.Params().Map())
		return nil
	})

	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/hello", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	}

	{
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("HEAD", "/hello", nil)
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code)
	}
}

func TestConflictingPlainAndWilcardRoutes(t *testing.T) {
	router := New()

	router.GET("/", dummyHandler)
	router.POST("/*path", dummyHandler)
	require.PanicsWithError(
		t,
		`routes "/" and "/*path" can't both handle GET`,
		func() {
			router.GET("/*path", dummyHandler)
		},
	)
}

func dummyHandler(w http.ResponseWriter, req Request) error {
	return nil
}
