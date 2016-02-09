package httptreemux

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGroupMethods(t *testing.T) {
	for _, scenario := range scenarios {
		t.Log(scenario.description)
		testGroupMethods(t, scenario.RequestCreator)
	}
}

func TestInvalidHandle(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("Bad handle path should have caused a panic")
		}
	}()
	New().NewGroup("/foo").GET("bar", nil)
}

func TestInvalidSubPath(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("Bad sub-path should have caused a panic")
		}
	}()
	New().NewGroup("/foo").NewGroup("bar")
}

func TestInvalidPath(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Error("Bad path should have caused a panic")
		}
	}()
	New().NewGroup("foo")
}

//Liberally borrowed from router_test
func testGroupMethods(t *testing.T, reqGen RequestCreator) {
	var result string
	makeHandler := func(method string) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			result = method
		}
	}
	router := New()
	// Testing with a sub-group of a group as that will test everything at once
	g := router.NewGroup("/base").NewGroup("/user")
	g.GET("/:param", makeHandler("GET"))
	g.POST("/:param", makeHandler("POST"))
	g.PATCH("/:param", makeHandler("PATCH"))
	g.PUT("/:param", makeHandler("PUT"))
	g.DELETE("/:param", makeHandler("DELETE"))

	testMethod := func(method, expect string) {
		result = ""
		w := httptest.NewRecorder()
		r, _ := reqGen(method, "/base/user/"+method, nil)
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

	router.HEAD("/base/user/:param", makeHandler("HEAD"))

	t.Log("Test HeadCanUseGet = false with explicit HEAD handler")
	testMethod("HEAD", "HEAD")
	router.HeadCanUseGet = true
	t.Log("Test HeadCanUseGet = true with explicit HEAD handler")
	testMethod("HEAD", "HEAD")
}
