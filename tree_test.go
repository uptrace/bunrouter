package httptreemux

import (
	"net/http"
	"strings"
	"testing"
)

func dummyHandler(w http.ResponseWriter, r *http.Request, urlParams map[string]string) {

}

func addPath(t *testing.T, tree *node, path string) {
	t.Logf("Adding path %s", path)
	n := tree.addPath(path[1:])
	handler := func(w http.ResponseWriter, r *http.Request, urlParams map[string]string) {
		urlParams["path"] = path
	}
	n.setHandler("GET", handler)
}

var test *testing.T

func testPath(t *testing.T, tree *node, path string, expectPath string, expectedParams map[string]string) {
	if t.Failed() {
		t.Log(tree.dumpTree("", " "))
		t.FailNow()
	}

	expectCatchAll := strings.Contains(expectPath, "/*")

	t.Log("Testing", path)
	var params map[string]string
	n := tree.search(path[1:], &params)
	if expectPath != "" && n == nil {
		t.Errorf("No match for %s, expected %s", path, expectPath)
		return
	} else if expectPath == "" && n != nil {
		t.Errorf("Expected no match for %s but got %v with params %v", path, n, expectedParams)
		t.Error("Node and subtree was\n" + n.dumpTree("", " "))
		return
	}

	if n == nil {
		return
	}

	if expectCatchAll != n.isCatchAll {
		t.Errorf("For path %s expectCatchAll %v but saw %v", path, expectCatchAll, n.isCatchAll)
	}

	handler, ok := n.leafHandler["GET"]
	if !ok {
		t.Errorf("Path %s returned node without handler", path)
		t.Error("Node and subtree was\n" + n.dumpTree("", " "))
		return
	}

	pathMap := make(map[string]string)
	handler(nil, nil, pathMap)
	matchedPath := pathMap["path"]

	if matchedPath != expectPath {
		t.Errorf("Path %s matched %s, expected %s", path, matchedPath, expectPath)
		t.Error("Node and subtree was\n" + n.dumpTree("", " "))
	}

	if expectedParams == nil {
		if len(params) != 0 {
			t.Errorf("Path %p expected no parameters, saw %v", path, params)
		}
	} else {
		for key, val := range expectedParams {
			sawVal, ok := params[key]
			if !ok {
				t.Errorf("Path %s matched without key %s", path, key)
			} else if sawVal != val {
				t.Errorf("Path %s expected param %s to be %s, saw %s", path, key, val, sawVal)
			}

			delete(params, key)
		}

		for key, val := range params {
			t.Errorf("Path %s returned unexpected param %s=%s", path, key, val)
		}
	}

}

func TestTree(t *testing.T) {
	test = t
	tree := &node{path: "/"}

	addPath(t, tree, "/")
	addPath(t, tree, "/images")
	addPath(t, tree, "/images/abc.jpg")
	addPath(t, tree, "/images/:imgname")
	addPath(t, tree, "/images/*path")
	addPath(t, tree, "/ima")
	addPath(t, tree, "/ima/:par")
	addPath(t, tree, "/images1")
	addPath(t, tree, "/images2")
	addPath(t, tree, "/apples")
	addPath(t, tree, "/app/les")
	addPath(t, tree, "/apples1")
	addPath(t, tree, "/appeasement")
	addPath(t, tree, "/appealing")
	addPath(t, tree, "/date/:year/:month")
	addPath(t, tree, "/date/:year/month")
	addPath(t, tree, "/date/:year/:month/abc")
	addPath(t, tree, "/date/:year/:month/:post")
	addPath(t, tree, "/date/:year/:month/*post")
	addPath(t, tree, "/:page")
	addPath(t, tree, "/:page/:index")
	addPath(t, tree, "/post/:post/page/:page")
	addPath(t, tree, "/plaster")

	testPath(t, tree, "/paper", "/:page",
		map[string]string{"page": "paper"})
	testPath(t, tree, "/", "/", nil)
	testPath(t, tree, "/images", "/images", nil)
	testPath(t, tree, "/images/abc.jpg", "/images/abc.jpg", nil)
	testPath(t, tree, "/images/something", "/images/:imgname",
		map[string]string{"imgname": "something"})
	testPath(t, tree, "/images/long/path", "/images/*path",
		map[string]string{"path": "long/path"})
	testPath(t, tree, "/images/even/longer/path", "/images/*path",
		map[string]string{"path": "even/longer/path"})
	testPath(t, tree, "/ima", "/ima", nil)
	testPath(t, tree, "/apples", "/apples", nil)
	testPath(t, tree, "/app/les", "/app/les", nil)
	testPath(t, tree, "/abc", "/:page",
		map[string]string{"page": "abc"})
	testPath(t, tree, "/abc/100", "/:page/:index",
		map[string]string{"page": "abc", "index": "100"})
	testPath(t, tree, "/post/a/page/2", "/post/:post/page/:page",
		map[string]string{"post": "a", "page": "2"})
	testPath(t, tree, "/date/2014/5", "/date/:year/:month",
		map[string]string{"year": "2014", "month": "5"})
	testPath(t, tree, "/date/2014/month", "/date/:year/month",
		map[string]string{"year": "2014"})
	testPath(t, tree, "/date/2014/5/abc", "/date/:year/:month/abc",
		map[string]string{"year": "2014", "month": "5"})
	testPath(t, tree, "/date/2014/5/def", "/date/:year/:month/:post",
		map[string]string{"year": "2014", "month": "5", "post": "def"})
	testPath(t, tree, "/date/2014/5/def/hij", "/date/:year/:month/*post",
		map[string]string{"year": "2014", "month": "5", "post": "def/hij"})
	testPath(t, tree, "/date/2014/5/def/hij/", "/date/:year/:month/*post",
		map[string]string{"year": "2014", "month": "5", "post": "def/hij/"})

	testPath(t, tree, "/date/2014/ab%2f", "/date/:year/:month",
		map[string]string{"year": "2014", "month": "ab/"})
	testPath(t, tree, "/post/ab%2fdef/page/2%2f", "/post/:post/page/:page",
		map[string]string{"post": "ab/def", "page": "2/"})

	testPath(t, tree, "/ima/bcd/fgh", "", nil)
	testPath(t, tree, "/date/2014//month", "", nil)
	testPath(t, tree, "/date/2014/05/", "", nil) // Empty catchall should not match
	testPath(t, tree, "/post//abc/page/2", "", nil)
	testPath(t, tree, "/post/abc//page/2", "", nil)
	testPath(t, tree, "/post/abc/page//2", "", nil)
	testPath(t, tree, "//post/abc/page/2", "", nil)
	testPath(t, tree, "//post//abc//page//2", "", nil)

	t.Log("Test retrieval of duplicate paths")
	params := make(map[string]string)
	p := "date/:year/:month/abc"
	n := tree.addPath(p)
	if n == nil {
		t.Errorf("Duplicate add of %s didn't return a node", p)
	} else {
		handler, ok := n.leafHandler["GET"]
		matchPath := ""
		if ok {
			handler(nil, nil, params)
			matchPath = params["path"]
		}

		if len(matchPath) < 2 || matchPath[1:] != p {
			t.Errorf("Duplicate add of %s returned node for %s\n%s", p, matchPath,
				n.dumpTree("", " "))

		}
	}

	t.Log(tree.dumpTree("", " "))
	test = nil
}

func TestPanics(t *testing.T) {
	sawPanic := false

	panicHandler := func() {
		if err := recover(); err != nil {
			sawPanic = true
		}
	}

	addPathPanic := func(p ...string) {
		sawPanic = false
		defer panicHandler()
		tree := &node{path: "/"}
		for _, path := range p {
			tree.addPath(path)
		}
	}

	addPathPanic("abc/*path/")
	if !sawPanic {
		t.Error("Expected panic with slash after catch-all")
	}

	addPathPanic("abc/*path/def")
	if !sawPanic {
		t.Error("Expected panic with path segment after catch-all")
	}

	addPathPanic("abc/*path", "abc/*paths")
	if !sawPanic {
		t.Error("Expected panic when adding conflicting catch-alls")
	}

	func() {
		sawPanic = false
		defer panicHandler()
		tree := &node{path: "/"}
		tree.setHandler("GET", dummyHandler)
		tree.setHandler("GET", dummyHandler)
	}()
	if !sawPanic {
		t.Error("Expected panic when adding a duplicate handler for a pattern")
	}

	addPathPanic("abc/ab:cd")
	if !sawPanic {
		t.Error("Expected panic with : in middle of path segment")
	}

	addPathPanic("abc/ab", "abc/ab:cd")
	if !sawPanic {
		t.Error("Expected panic with : in middle of path segment with existing path")
	}

	addPathPanic("abc/ab*cd")
	if !sawPanic {
		t.Error("Expected panic with * in middle of path segment")
	}

	addPathPanic("abc/ab", "abc/ab*cd")
	if !sawPanic {
		t.Error("Expected panic with * in middle of path segment with existing path")
	}
}

func BenchmarkTreeNullRequest(b *testing.B) {
	b.ReportAllocs()
	tree := &node{path: "/"}
	var params map[string]string

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.search("", &params)
	}
}

func BenchmarkTreeOneStatic(b *testing.B) {
	b.ReportAllocs()
	tree := &node{path: "/"}
	tree.addPath("abc")
	var params map[string]string

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.search("abc", &params)
	}
}

func BenchmarkTreeOneParam(b *testing.B) {
	b.ReportAllocs()
	tree := &node{path: "/"}
	tree.addPath(":abc")
	var params map[string]string

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params = nil
		tree.search("abc", &params)
	}
}
