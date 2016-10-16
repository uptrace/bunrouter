// +build go1.7

package httptreemux

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContextParams(t *testing.T) {
	m := map[string]string{"id": "123"}
	ctx := context.WithValue(context.Background(), ParamsContextKey, m)

	params := ContextParams(ctx)
	if params == nil {
		t.Errorf("expected '%#v', but got '%#v'", m, params)
	}

	if v := params["id"]; v != "123" {
		t.Errorf("expected '%s', but got '%#v'", m["id"], params["id"])
	}
}

func TestContextGroupMethods(t *testing.T) {
	for _, scenario := range scenarios {
		t.Log(scenario.description)
		testContextGroupMethods(t, scenario.RequestCreator, true)
		testContextGroupMethods(t, scenario.RequestCreator, false)
	}
}

func testContextGroupMethods(t *testing.T, reqGen RequestCreator, headCanUseGet bool) {
	var result string
	makeHandler := func(method string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			result = method

			v, ok := ContextParams(r.Context())["param"]
			if !ok {
				t.Error("missing key 'param' in context")
			}

			if headCanUseGet && method == "GET" && v == "HEAD" {
				return
			}

			if v != method {
				t.Errorf("invalid key 'param' in context; expected '%s' but got '%s'", method, v)
			}
		}
	}

	router := New()
	router.HeadCanUseGet = headCanUseGet

	cg := router.UsingContext().NewContextGroup("/base").NewContextGroup("/user")
	cg.GET("/:param", makeHandler("GET"))
	cg.POST("/:param", makeHandler("POST"))
	cg.PATCH("/:param", makeHandler("PATCH"))
	cg.PUT("/:param", makeHandler("PUT"))
	cg.DELETE("/:param", makeHandler("DELETE"))

	testMethod := func(method, expect string) {
		result = ""
		w := httptest.NewRecorder()
		r, _ := reqGen(method, "/base/user/"+method, nil)
		router.ServeHTTP(w, r)
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

	cg.HEAD("/:param", makeHandler("HEAD"))
	testMethod("HEAD", "HEAD")
}

func TestNewContextGroup(t *testing.T) {
	router := New()
	group := router.NewGroup("/api")

	group.GET("/v1", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		w.Write([]byte(`200 OK GET /api/v1`))
	})

	group.UsingContext().GET("/v2", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`200 OK GET /api/v2`))
	})

	tests := []struct {
		uri, expected string
	}{
		{"/api/v1", "200 OK GET /api/v1"},
		{"/api/v2", "200 OK GET /api/v2"},
	}

	for _, tc := range tests {
		r, err := http.NewRequest("GET", tc.uri, nil)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("GET %s: expected %d, but got %d", tc.uri, http.StatusOK, w.Code)
		}
		if got := w.Body.String(); got != tc.expected {
			t.Errorf("GET %s : expected %q, but got %q", tc.uri, tc.expected, got)
		}

	}
}

type ContextGroupHandler struct{}

//	adhere to the http.Handler interface
func (f ContextGroupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		w.Write([]byte(`200 OK GET /api/v1`))
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func TestNewContextGroupHandler(t *testing.T) {
	router := New()
	group := router.NewGroup("/api")

	group.UsingContext().Handler("GET", "/v1", ContextGroupHandler{})

	tests := []struct {
		uri, expected string
	}{
		{"/api/v1", "200 OK GET /api/v1"},
	}

	for _, tc := range tests {
		r, err := http.NewRequest("GET", tc.uri, nil)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("GET %s: expected %d, but got %d", tc.uri, http.StatusOK, w.Code)
		}
		if got := w.Body.String(); got != tc.expected {
			t.Errorf("GET %s : expected %q, but got %q", tc.uri, tc.expected, got)
		}
	}
}

func TestDefaultContext(t *testing.T) {
	router := New()
	ctx := context.WithValue(context.Background(), "abc", "def")
	expectContext := false

	router.GET("/abc", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		contextValue := r.Context().Value("abc")
		if expectContext {
			x, ok := contextValue.(string)
			if !ok || x != "def" {
				t.Errorf("Unexpected context key value: %+v", contextValue)
			}
		} else {
			if contextValue != nil {
				t.Errorf("Expected blank context but key had value %+v", contextValue)
			}
		}
	})

	r, err := http.NewRequest("GET", "/abc", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	t.Log("Testing without DefaultContext")
	router.ServeHTTP(w, r)

	router.DefaultContext = ctx
	expectContext = true
	w = httptest.NewRecorder()
	t.Log("Testing with DefaultContext")
	router.ServeHTTP(w, r)
}
