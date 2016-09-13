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

func TestHandleWithContext(t *testing.T) {
	for _, scenario := range scenarios {
		t.Log(scenario.description)
		testHandleWithContext(t, scenario.RequestCreator, true)
		testHandleWithContext(t, scenario.RequestCreator, false)
	}
}

func testHandleWithContext(t *testing.T, reqGen RequestCreator, headCanUseGet bool) {
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

	router := New().UsingContext()
	router.TreeMux().HeadCanUseGet = headCanUseGet

	cg := router.NewContextGroup("/base").NewContextGroup("/user")
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

	router.Handle("HEAD", "/base/user/:param", makeHandler("HEAD"))
	testMethod("HEAD", "HEAD")
}
