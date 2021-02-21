# treemux - fast and flexible HTTP router

[![Build Status](https://travis-ci.com/vmihailenco/treemux.png?branch=master)](https://travis-ci.com/vmihailenco/treemux)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/vmihailenco/treemux)](https://pkg.go.dev/github.com/vmihailenco/treemux)
[![Chat](https://discordapp.com/api/guilds/752070105847955518/widget.png)](https://discord.gg/rWtp5Aj)

> :heart: [**Uptrace.dev** - distributed traces, logs, and errors in one place](https://uptrace.dev)

- [Basic example](/example/basic/)
- [CORS example](/example/cors/)
- [Error handling example](/example/error_handling/)
- [Debug logging](/extra/reqlog/)
- [Gzip compression](/extra/treemuxgzip/)
- [OpenTelemetry integration](/extra/treemuxotel/)
- [Writing REST API with Go and PostgreSQL](https://pg.uptrace.dev/rest-api/)
- [RealWorld example application](https://github.com/uptrace/go-treemux-realworld-example-app)
- [Reference](https://pkg.go.dev/github.com/vmihailenco/treemux)

High-speed, flexible, tree-based HTTP router for Go. It is as fast as httprouter, but with relaxed
routing rules.

<details>
  <summary>Benchmark results</summary>

```
#GithubAPI Routes: 203
   HttpRouter: 37088 Bytes
   HttpTreeMux: 78800 Bytes
   VmihailencoTreemux: 58208 Bytes

#GPlusAPI Routes: 13
   HttpRouter: 2760 Bytes
   HttpTreeMux: 7440 Bytes
   VmihailencoTreemux: 5760 Bytes

#ParseAPI Routes: 26
   HttpRouter: 5024 Bytes
   HttpTreeMux: 7848 Bytes
   VmihailencoTreemux: 5816 Bytes

#Static Routes: 157
   HttpRouter: 21680 Bytes
   HttpTreeMux: 73448 Bytes
   VmihailencoTreemux: 49496 Bytes

goos: linux
goarch: amd64
pkg: github.com/julienschmidt/go-http-routing-benchmark
cpu: AMD Ryzen 5 2600 Six-Core Processor
BenchmarkHttpRouter_Param                	13176325	        90.69 ns/op	      32 B/op	       1 allocs/op
BenchmarkHttpTreeMux_Param               	 2846073	       422.0 ns/op	     352 B/op	       3 allocs/op
BenchmarkVmihailencoTreemux_Param        	 8060584	       149.0 ns/op	      32 B/op	       1 allocs/op
BenchmarkHttpRouter_Param5               	 5219818	       230.9 ns/op	     160 B/op	       1 allocs/op
BenchmarkHttpTreeMux_Param5              	 1000000	      1054 ns/op	     576 B/op	       6 allocs/op
BenchmarkVmihailencoTreemux_Param5       	 3114952	       386.3 ns/op	     160 B/op	       1 allocs/op
BenchmarkHttpRouter_Param20              	 1790470	       670.3 ns/op	     640 B/op	       1 allocs/op
BenchmarkHttpTreeMux_Param20             	  250482	      4890 ns/op	    3196 B/op	      10 allocs/op
BenchmarkVmihailencoTreemux_Param20      	 1000000	      1179 ns/op	     640 B/op	       1 allocs/op
BenchmarkHttpRouter_ParamWrite           	 9207016	       131.3 ns/op	      32 B/op	       1 allocs/op
BenchmarkHttpTreeMux_ParamWrite          	 2607104	       461.4 ns/op	     352 B/op	       3 allocs/op
BenchmarkVmihailencoTreemux_ParamWrite   	 6644600	       181.0 ns/op	      32 B/op	       1 allocs/op
BenchmarkHttpRouter_GithubStatic         	28477857	        42.21 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpTreeMux_GithubStatic        	21255450	        54.35 ns/op	       0 B/op	       0 allocs/op
BenchmarkVmihailencoTreemux_GithubStatic 	23992582	        50.81 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_GithubParam          	 5901350	       202.7 ns/op	      96 B/op	       1 allocs/op
BenchmarkHttpTreeMux_GithubParam         	 1801438	       667.4 ns/op	     384 B/op	       4 allocs/op
BenchmarkVmihailencoTreemux_GithubParam  	 3858866	       310.4 ns/op	      64 B/op	       1 allocs/op
BenchmarkHttpRouter_GithubAll            	   29740	     40436 ns/op	   13792 B/op	     167 allocs/op
BenchmarkHttpTreeMux_GithubAll           	   10000	    121313 ns/op	   65856 B/op	     671 allocs/op
BenchmarkVmihailencoTreemux_GithubAll    	   21978	     54883 ns/op	   10848 B/op	     167 allocs/op
BenchmarkHttpRouter_GPlusStatic          	45898861	        25.99 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpTreeMux_GPlusStatic         	33262026	        35.88 ns/op	       0 B/op	       0 allocs/op
BenchmarkVmihailencoTreemux_GPlusStatic  	41993342	        28.61 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_GPlusParam           	 8516348	       142.3 ns/op	      64 B/op	       1 allocs/op
BenchmarkHttpTreeMux_GPlusParam          	 2606780	       460.0 ns/op	     352 B/op	       3 allocs/op
BenchmarkVmihailencoTreemux_GPlusParam   	 6475717	       186.5 ns/op	      32 B/op	       1 allocs/op
BenchmarkHttpRouter_GPlus2Params         	 6980588	       173.2 ns/op	      64 B/op	       1 allocs/op
BenchmarkHttpTreeMux_GPlus2Params        	 1829535	       656.7 ns/op	     384 B/op	       4 allocs/op
BenchmarkVmihailencoTreemux_GPlus2Params 	 3930752	       305.9 ns/op	      64 B/op	       1 allocs/op
BenchmarkHttpRouter_GPlusAll             	  642843	      1940 ns/op	     640 B/op	      11 allocs/op
BenchmarkHttpTreeMux_GPlusAll            	  188905	      6437 ns/op	    4032 B/op	      38 allocs/op
BenchmarkVmihailencoTreemux_GPlusAll     	  474685	      2808 ns/op	     512 B/op	      11 allocs/op
BenchmarkHttpRouter_ParseStatic          	44829891	        26.75 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpTreeMux_ParseStatic         	22076780	        54.17 ns/op	       0 B/op	       0 allocs/op
BenchmarkVmihailencoTreemux_ParseStatic  	22923108	        52.60 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_ParseParam           	 9637926	       124.6 ns/op	      64 B/op	       1 allocs/op
BenchmarkHttpTreeMux_ParseParam          	 2794585	       429.7 ns/op	     352 B/op	       3 allocs/op
BenchmarkVmihailencoTreemux_ParseParam   	 7717832	       155.2 ns/op	      32 B/op	       1 allocs/op
BenchmarkHttpRouter_Parse2Params         	 8440347	       142.8 ns/op	      64 B/op	       1 allocs/op
BenchmarkHttpTreeMux_Parse2Params        	 2006734	       600.1 ns/op	     384 B/op	       4 allocs/op
BenchmarkVmihailencoTreemux_Parse2Params 	 4927276	       243.5 ns/op	      64 B/op	       1 allocs/op
BenchmarkHttpRouter_ParseAll             	  435433	      2870 ns/op	     640 B/op	      16 allocs/op
BenchmarkHttpTreeMux_ParseAll            	  126555	      9631 ns/op	    5728 B/op	      51 allocs/op
BenchmarkVmihailencoTreemux_ParseAll     	  263841	      4452 ns/op	     608 B/op	      16 allocs/op
BenchmarkHttpRouter_StaticAll            	   98608	     12278 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpTreeMux_StaticAll           	   98220	     12231 ns/op	       0 B/op	       0 allocs/op
BenchmarkVmihailencoTreemux_StaticAll    	   97520	     12241 ns/op	       0 B/op	       0 allocs/op
```

</details>

## Installing with Go Modules

When using Go Modules, import this repository with `import "github.com/vmihailenco/treemux"` to
ensure that you get the right version.

## Handler

The handler is a simple function with the prototype
`func(w http.ResponseWriter, req treemux.Request) error`. A `treemux.Request` contains route name
and parameters parsed from wildcards and catch-alls in the URL. This type is aliased as
`treemux.HandlerFunc`.

```go
import "github.com/vmihailenco/treemux"

router := treemux.New()

group := router.NewGroup("/api/v1")

group.GET("/:id", func(w http.ResponseWriter, req treemux.Request) error {
  id := req.Param("id")
  return treemux.JSON(w, treemux.H{
      "url": fmt.Sprintf("GET /api/v1/%s", id),
      "route": req.Route(),
  })
})

log.Println(http.ListenAndServe(":8080", router))
```

### Why not http.HandlerFunc?

`treemux.HandlerFunc` is a thin wrapper over `http.HandlerFunc`:

- `treemux.Request` replaces `*http.Request`. You can get the original request via `req.Request`.
- Handler returns an error just like any other Go function.

Those 2 tiny changes bring us:

- **Shorter and simpler error handling**. In your handlers you just return the error and deal with
  it in a [middleware](/example/error_handling/) in centralized fashion.
- **Easier debugging**. Since middlewares have access to errors you can [log errors](/extra/reqlog/)
  along with other debugging information. OpenTelemetry [integration](/extra/treemuxotel/) uses that
  to record the error.
- **Route name and params**. `*http.Request` was not designed to carry the route name and params.
  You can store that information in the request `context.Context`, but that clones the request and
  therefore requires an allocation
- **Effeciency**. `treemux.Request` is designed so `req.WithContext(ctx)` does not allocate.

Treemux comes with middlewares that handle [gzip compression](/extra/treemuxgzip/),
[CORS](/example/cors/), [OpenTelemetry integration](/extra/treemuxotel/), and
[request logging](/extra/reqlog/). So with minimal changes you can make treemux work nicely with
existing libraries.

## Converting http.HandlerFunc to treemux.HandlerFunc

treemux provides helpers to convert existing `http.HandlerFunc` and `http.Handler` into
`treemux.HandlerFunc`:

```go
// http.HandlerFunc -> treemux.HandlerFunc
router.GET("/foo", treemux.HTTPHandlerFunc(existingHandlerFunc))

// http.Handler -> treemux.HandlerFunc
router.GET("/bar", treemux.HTTPHandler(existingHandler))
```

## Middlewares

Middleware is a function that wraps a handler with another function:

```go
func corsMiddleware(next treemux.HandlerFunc) treemux.HandlerFunc {
    return func(w http.ResponseWriter, req treemux.Request) error {
        if origin := req.Header.Get("Origin"); origin != "" {
            h := w.Header()
            h.Set("Access-Control-Allow-Origin", origin)
            h.Set("Access-Control-Allow-Credentials", "true")
        }
        return next(w, req)
    }
}

router = treemux.New(treemux.WithMiddleware(corsMiddleware))
```

Middlewares are also used for [error handling](/example/error_handler/).

## Routing Rules

The syntax here is modeled after httprouter. Each variable in a path may match on one segment only,
except for an optional catch-all variable at the end of the URL.

Some examples of valid URL patterns are:

- `/post/all`
- `/post/:postid`
- `/post/:postid/page/:page`
- `/post/:postid/:page`
- `/images/*path`
- `/favicon.ico`
- `/:year/:month/`
- `/:year/:month/:post`
- `/:page`

Note that all of the above URL patterns may exist concurrently in the router.

Path elements starting with `:` indicate a wildcard in the path. A wildcard will only match on a
single path segment. That is, the pattern `/post/:postid` will match on `/post/1` or `/post/1/`, but
not `/post/1/2`.

A path element starting with `*` is a catch-all, whose value will be a string containing all text in
the URL matched by the wildcards. For example, with a pattern of `/images/*path` and a requested URL
`images/abc/def`, path would contain `abc/def`. A catch-all path will not match an empty string, so
in this example a separate route would need to be installed if you also want to match `/images/`.

#### Using : and \* in routing patterns

The characters `:` and `*` can be used at the beginning of a path segment by escaping them with a
backslash. A double backslash at the beginning of a segment is interpreted as a single backslash.
These escapes are only checked at the very beginning of a path segment; they are not necessary or
processed elsewhere in a token.

```go
router.GET("/foo/\\*starToken", handler) // matches /foo/*starToken
router.GET("/foo/star*inTheMiddle", handler) // matches /foo/star*inTheMiddle
router.GET("/foo/starBackslash\\*", handler) // matches /foo/starBackslash\*
router.GET("/foo/\\\\*backslashWithStar") // matches /foo/\*backslashWithStar
```

### Routing Groups

Lets you create a new group of routes with a given path prefix. Makes it easier to create clusters
of paths like:

- `/api/v1/foo`
- `/api/v1/bar`

To use this you do:

```go
router = treemux.New()

api := router.NewGroup("/api/v1")
api.GET("/foo", fooHandler) // becomes /api/v1/foo
api.GET("/bar", barHandler) // becomes /api/v1/bar
```

Or using `WithGroup`:

```go
router.WithGroup("/api/v1", func(g *treemux.Group) {
    g.GET("/foo", fooHandler) // becomes /api/v1/foo
    g.GET("/bar", barHandler) // becomes /api/v1/bar
})
```

More complex example:

```go
router := treemux.New()

g := router.NewGroup("/api/v1", treemux.WithMiddleware(ipRateLimitMiddleware))

g.NewGroup("/users/:user_id",
    treemux.WithMiddleware(authMiddleware),
    treemux.WithGroup(func(g *treemux.Group) {
        g.GET("", userHandler)

        g = g.WithMiddleware(adminMiddleware)

        g.PUT("", updateUserHandler)
        g.DELETE("", deleteUserHandler)
    }))

g.NewGroup("/projects/:project_id/articles/:article_id",
    treemux.WithMiddleware(authMiddleware),
    treemux.WithMiddleware(projectMiddleware),
    treemux.WithGroup(func(g *treemux.Group) {
        g.GET("", articleHandler)

        g.Use(quotaMiddleware)

        g.POST("", createArticleHandler)
        g.PUT("", updateArticleHandler)
        g.DELETE("", deleteArticleHandler)
    }))
```

### Routing Priority

The priority rules in the router are simple.

1. Static path segments take the highest priority. If a segment and its subtree are able to match
   the URL, that match is returned.
2. Wildcards take second priority. For a particular wildcard to match, that wildcard and its subtree
   must match the URL.
3. Finally, a catch-all rule will match when the earlier path segments have matched, and none of the
   static or wildcard conditions have matched. Catch-all rules must be at the end of a pattern.

So with the following patterns adapted from [simpleblog](https://www.github.com/dimfeld/simpleblog),
we'll see certain matches:

```go
router = treemux.New()
router.GET("/:page", pageHandler)
router.GET("/:year/:month/:post", postHandler)
router.GET("/:year/:month", archiveHandler)
router.GET("/images/*path", staticHandler)
router.GET("/favicon.ico", staticHandler)
```

#### Example scenarios

- `/abc` will match `/:page`
- `/2014/05` will match `/:year/:month`
- `/2014/05/really-great-blog-post` will match `/:year/:month/:post`
- `/images/CoolImage.gif` will match `/images/*path`
- `/images/2014/05/MayImage.jpg` will also match `/images/*path`, with all the text after `/images`
  stored in the variable path.
- `/favicon.ico` will match `/favicon.ico`

### Special Method Behavior

If TreeMux.HeadCanUseGet is set to true, the router will call the GET handler for a pattern when a
HEAD request is processed, if no HEAD handler has been added for that pattern. This behavior is
enabled by default.

Go's http.ServeContent and related functions already handle the HEAD method correctly by sending
only the header, so in most cases your handlers will not need any special cases for it.

### Trailing Slashes

The router has special handling for paths with trailing slashes. If a pattern is added to the router
with a trailing slash, any matches on that pattern without a trailing slash will be redirected to
the version with the slash. If a pattern does not have a trailing slash, matches on that pattern
with a trailing slash will be redirected to the version without.

The trailing slash flag is only stored once for a pattern. That is, if a pattern is added for a
method with a trailing slash, all other methods for that pattern will also be considered to have a
trailing slash, regardless of whether or not it is specified for those methods too. However this
behavior can be turned off by setting TreeMux.RedirectTrailingSlash to false. By default it is set
to true.

One exception to this rule is catch-all patterns. By default, trailing slash redirection is disabled
on catch-all patterns, since the structure of the entire URL and the desired patterns can not be
predicted. If trailing slash removal is desired on catch-all patterns, set
TreeMux.RemoveCatchAllTrailingSlash to true.

```go
router = treemux.New()
router.GET("/about", pageHandler)
router.GET("/posts/", postIndexHandler)
router.POST("/posts", postFormHandler)

GET /about will match normally.
GET /about/ will redirect to /about.
GET /posts will redirect to /posts/.
GET /posts/ will match normally.
POST /posts will redirect to /posts/, because the GET method used a trailing slash.
```

### Custom Redirects

RedirectBehavior sets the behavior when the router redirects the request to the canonical version of
the requested URL using RedirectTrailingSlash or RedirectClean. The default behavior is to return a
301 status, redirecting the browser to the version of the URL that matches the given pattern.

These are the values accepted for RedirectBehavior. You may also add these values to the
RedirectMethodBehavior map to define custom per-method redirect behavior.

- Redirect301 - HTTP 301 Moved Permanently; this is the default.
- Redirect307 - HTTP/1.1 Temporary Redirect
- Redirect308 - RFC7538 Permanent Redirect
- UseHandler - Don't redirect to the canonical path. Just call the handler instead.

#### Rationale/Usage

On a POST request, most browsers that receive a 301 will submit a GET request to the redirected URL,
meaning that any data will likely be lost. If you want to handle and avoid this behavior, you may
use Redirect307, which causes most browsers to resubmit the request using the original method and
request body.

Since 307 is supposed to be a temporary redirect, the new 308 status code has been proposed, which
is treated the same, except it indicates correctly that the redirection is permanent. The big caveat
here is that the RFC is relatively recent, and older or non-compliant browsers will not handle it.
Therefore its use is not recommended unless you really know what you're doing.

Finally, the UseHandler value will simply call the handler function for the pattern, without
redirecting to the canonical version of the URL.

### RequestURI vs. URL.Path

#### Escaped Slashes

Go automatically processes escaped characters in a URL, converting + to a space and %XX to the
corresponding character. This can present issues when the URL contains a %2f, which is unescaped to
'/'. This isn't an issue for most applications, but it will prevent the router from correctly
matching paths and wildcards.

For example, the pattern `/post/:post` would not match on `/post/abc%2fdef`, which is unescaped to
`/post/abc/def`. The desired behavior is that it matches, and the `post` wildcard is set to
`abc/def`.

Therefore, this router defaults to using the raw URL, stored in the Request.RequestURI variable.
Matching wildcards and catch-alls are then unescaped, to give the desired behavior.

TL;DR: If a requested URL contains a %2f, this router will still do the right thing. Some Go HTTP
routers may not due to [Go issue 3659](https://github.com/golang/go/issues/3659).

#### http Package Utility Functions

Although using RequestURI avoids the issue described above, certain utility functions such as
`http.StripPrefix` modify URL.Path, and expect that the underlying router is using that field to
make its decision. If you are using some of these functions, set the router's `PathSource` member to
`URLPath`. This will give up the proper handling of escaped slashes described above, while allowing
the router to work properly with these utility functions.

## Error Handlers

### NotFoundHandler

`TreeMux.NotFoundHandler` can be set to provide custom 404-error handling. The default
implementation is Go's `http.NotFound` function.

### MethodNotAllowedHandler

If a pattern matches, but the pattern does not have an associated handler for the requested method,
the router calls the MethodNotAllowedHandler. The default version of this handler just writes the
status code `http.StatusMethodNotAllowed`.

## Unexpected Differences from Other Routers

This router is intentionally light on features in the name of simplicity and performance. When
coming from another router that does heavier processing behind the scenes, you may encounter some
unexpected behavior. This list is by no means exhaustive, but covers some nonobvious cases that
users have encountered.

### httprouter and catch-all parameters

When using `httprouter`, a route with a catch-all parameter (e.g. `/images/*path`) will match on
URLs like `/images/` where the catch-all parameter is empty. This router does not match on empty
catch-all parameters, but the behavior can be duplicated by adding a route without the catch-all
(e.g. `/images/`).

## httptreemux

This is a fork of [httptreemux](https://github.com/dimfeld/httptreemux). The original code was
written by [Daniel Imfeld](https://github.com/dimfeld/).

### Changes from httptreemux

- Thin wrapper `treemux.Request` around `http.Request` to expose route via `Request.Route` and route
  params via `req.Params`.

- Setting a `context.Context` does not require an allocation.

- More efficient params encoding using a slice instead of a map.

- Reworked configuration.

- `Group` is immutable to avoid accidental leaking of middlewares into the group.
