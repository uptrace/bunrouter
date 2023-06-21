package bunrouter

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmptyGroupAndMapping(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			// everything is good, it paniced
		} else {
			t.Error(`Expected NewGroup("")`)
		}
	}()
	New().GET("", func(w http.ResponseWriter, _ Request) error {
		return nil
	})
}

func TestSubGroupSlashMapping(t *testing.T) {
	r := New()
	r.NewGroup("/foo").GET("/", func(w http.ResponseWriter, _ Request) error {
		w.WriteHeader(200)
		return nil
	})

	var req *http.Request
	var recorder *httptest.ResponseRecorder

	req, _ = http.NewRequest("GET", "/foo", nil)
	recorder = httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	if recorder.Code != 301 { // should get redirected
		t.Error(`/foo on NewGroup("/foo").GET("/") should result in 301 response, got:`, recorder.Code)
	}

	req, _ = http.NewRequest("GET", "/foo/", nil)
	recorder = httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	if recorder.Code != 200 {
		t.Error(`/foo/ on NewGroup("/foo").GET("/") should result in 200 response, got:`, recorder.Code)
	}
}

func TestSubGroupEmptyMapping(t *testing.T) {
	r := New()
	r.NewGroup("/foo").GET("", func(w http.ResponseWriter, _ Request) error {
		w.WriteHeader(200)
		return nil
	})
	req, _ := http.NewRequest("GET", "/foo", nil)
	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, req)
	if recorder.Code != 200 {
		t.Error(`/foo on NewGroup("/foo").GET("") should result in 200 response, got:`, recorder.Code)
	}
}

func TestGroupMethods(t *testing.T) {
	for _, scenario := range scenarios {
		t.Log(scenario.description)
		testGroupMethods(t)
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

// Liberally borrowed from router_test
func testGroupMethods(t *testing.T) {
	var result string
	makeHandler := func(method string) HandlerFunc {
		return func(w http.ResponseWriter, r Request) error {
			result = method
			return nil
		}
	}
	router := New()

	// Testing with a sub-group of a group as that will test everything at once
	const (
		basePath     = "/base"
		userPath     = "/user"
		fullUserPath = basePath + userPath
	)
	g := router.NewGroup(basePath).NewGroup(userPath)
	g.GET("/:param", makeHandler("GET"))
	g.POST("/:param", makeHandler("POST"))
	g.PATCH("/:param", makeHandler("PATCH"))
	g.PUT("/:param", makeHandler("PUT"))
	g.DELETE("/:param", makeHandler("DELETE"))

	require.Equal(t, g.Path(), fullUserPath)

	testMethod := func(method, expect string) {
		result = ""

		url := fmt.Sprintf("%s/%s", fullUserPath, method)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(method, url, nil)
		router.ServeHTTP(w, r)

		if expect == "" {
			require.Equal(t, http.StatusMethodNotAllowed, w.Code)
		}

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

	router.HEAD(fullUserPath+"/:param", makeHandler("HEAD"))
	testMethod("HEAD", "HEAD")
}
