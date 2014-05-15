httptreemux
===========

Simple tree-based HTTP router for Go.

This is inspired by [Julien Schmidt's httprouter](https://www.github.com/julienschmidt/httprouter), in that it uses a patricia tree, but the implementation is rather different. Specifically, the routing rules are relaxed to allow wildcards and static tokens in a path segment. This gives a nice combination of high performance with a lot of convenience in designing the routing patterns. In [benchmarks](https://github.com/dimfeld/go-http-routing-benchmark), httptreemux is close to, but slightly slower than, httprouter.

## Why?
There are a lot of good routers out there. But looking at the ones that were really lightweight, I couldn't quite get something that fit with the route patterns I wanted. The code itself is simple enough, so I spent an evening writing this.

## Handler
The handler is a simple function with the prototype `func(w http.ResponseWriter, r *http.Request, params map[string]string`. The params argument contains the parameters parsed from wildcards and catch-alls in the URL, as described below. This type is aliased as httptreemux.HandlerFunc.

## Routing Rules
The syntax here is also modeled after httprouter. Each variable in a path may match on ones segment only, except for an optional catch-all variable at the end of the URL.

Some examples of valid URL patterns are:
* /post/all
* /post/:postid
* /post/:postid/page/:page
* /images/*path
* /favicon.ico
* /:year/:month/
* /:year/:month/:post
* /:page

Note that all of the above URL patterns may exist concurrently in the router.

Path elements starting with : indicate a wildcard in the path. A wildcard will only match on a single path segment. A path element starting with * is a catch-all, whose value will be a string containing all text in the URL matched by the wildcards. For example, with a pattern of `/images/*path` and a requested URL `images/abc/def`, path would contain `abc/def`.

### Routing Priority
The priority rules in the router are simple.
1. Static path segments take the highest priority. If a segment and its subtree are able to match the URL, that match is returned.
2. Wildcards take seconds priority. For a particular wildcard to match, that wildcard and its subtree must match the URL.
3. Finally, a catch-all rule will match when the earlier path segments have matched, and none of the static or wildcard conditions have matched. Catch-all rules must be at the end of a pattern.

So with the following patterns adapted from [simpleblog](https://www.github.com/dimfeld/simpleblog), we'll see certain matches:
```go
router = httptreemux.New()
router.GET("/:page", pageHandler)
router.GET("/:year/:month/:post", postHandler)
router.GET("/:year/:month", archiveHandler)
router.GET("/images/*path", staticHandler)
router.GET("/favicon.ico", staticHandler)

/abc will match /:page
/2014/05 will match /:year/:month
/2014/05/really-great-blog-post will match /:year/:month/:post
/images/CoolImage.gif will match /images/*path
/images/2014/05/MayImage.jpg will also match /images/*path, with all the text after /images stored in the variable path.
/favicon.ico will match /favicon.ico
```

## NotFoundHandler
TreeMux.NotFoundHandler can be set to provide custom 404-error handling. The default implementation is Go's built-in http.NotFound function.

## Panic handling
TreeMux.PanicHandler can be set to provide custom panic handling. The default implementation just returns error 500. The function ShowErrorsPanicHandler, adapted from [gocraft/web](https://github.com/gocraft/web), will print panic errors to the browser in an easily-readable format.

## Middleware
This package provides no middleware. But there are a lot of great options out there and it's pretty easy to write your own.

# Acknowledgements

* Inspiration and CleanPath function from Julien Schmidt's [httprouter](https://github.com/julienschmidt/httprouter)
* Show Errors panic handler from [gocraft/web](https://github.com/gocraft/web)
