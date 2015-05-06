httptreemux  [![Build Status](https://travis-ci.org/dimfeld/httptreemux.png?branch=master)](https://travis-ci.org/dimfeld/httptreemux) [![GoDoc](http://godoc.org/github.com/dimfeld/httptreemux?status.png)](http://godoc.org/github.com/dimfeld/httptreemux)
===========

High-speed, flexible, tree-based HTTP router for Go.

This is inspired by [Julien Schmidt's httprouter](https://www.github.com/julienschmidt/httprouter), in that it uses a patricia tree, but the implementation is rather different. Specifically, the routing rules are relaxed so that a single path segment may be a wildcard in one route and a static token in another. This gives a nice combination of high performance with a lot of convenience in designing the routing patterns. In [benchmarks](https://github.com/dimfeld/go-http-routing-benchmark), httptreemux is close to, but slightly slower than, httprouter.

## Why?
There are a lot of good routers out there. But looking at the ones that were really lightweight, I couldn't quite get something that fit with the route patterns I wanted. The code itself is simple enough, so I spent an evening writing this.

## Handler
The handler is a simple function with the prototype `func(w http.ResponseWriter, r *http.Request, params map[string]string)`. The params argument contains the parameters parsed from wildcards and catch-alls in the URL, as described below. This type is aliased as httptreemux.HandlerFunc.

## Routing Rules
The syntax here is also modeled after httprouter. Each variable in a path may match on one segment only, except for an optional catch-all variable at the end of the URL.

Some examples of valid URL patterns are:
* /post/all
* /post/:postid
* /post/:postid/page/:page
* /post/:postid/:page
* /images/*path
* /favicon.ico
* /:year/:month/
* /:year/:month/:post
* /:page

Note that all of the above URL patterns may exist concurrently in the router.

Path elements starting with : indicate a wildcard in the path. A wildcard will only match on a single path segment. That is, the pattern `/post/:postid` will match on `/post/1` or `/post/1/`, but not `/post/1/2`.

A path element starting with * is a catch-all, whose value will be a string containing all text in the URL matched by the wildcards. For example, with a pattern of `/images/*path` and a requested URL `images/abc/def`, path would contain `abc/def`.

### Routing Priority
The priority rules in the router are simple.

1. Static path segments take the highest priority. If a segment and its subtree are able to match the URL, that match is returned.
2. Wildcards take second priority. For a particular wildcard to match, that wildcard and its subtree must match the URL.
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

### Special Method Behavior
If TreeMux.HeadCanUseGet is set to true, the router will call the GET handler for a pattern when a HEAD request is processed, if no HEAD handler has been added for that pattern. This behavior is enabled by default.

Go's http.ServeContent and related functions already handle the HEAD method correctly by sending only the header, so in most cases your handlers will not need any special cases for it.

### Trailing Slashes
The router has special handling for paths with trailing slashes. If a pattern is added to the router with a trailing slash, any matches on that pattern without a trailing slash will be redirected to the version with the slash. If a pattern does not have a trailing slash, matches on that pattern with a trailing slash will be redirected to the version without.

The trailing slash flag is only stored once for a pattern. That is, if a pattern is added for a method with a trailing slash, all other methods for that pattern will also be considered to have a trailing slash, regardless of whether or not it is specified for those methods too.
However this behavior can be turned off by setting TreeMux.RedirectTrailingSlash to false. By default it is set to true.

One exception to this rule is catch-all patterns. By default, trailing slash redirection is disabled on catch-all patterns, since the structure of the entire URL and the desired patterns can not be predicted. If trailing slash removal is desired on catch-all patterns, set TreeMux.RemoveCatchAllTrailingSlash to true.

```go
router = httptreemux.New()
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

RedirectBehavior sets the behavior when the router redirects the request to the canonical version of the requested URL using RedirectTrailingSlash or RedirectClean. The default behavior is to return a 301 status, redirecting the browser to the version of the URL that matches the given pattern. 

These are the values accepted for RedirectBehavior. You may also add these values to the RedirectMethodBehavior map to define custom per-method redirect behavior.

* Redirect301 - HTTP 301 Moved Permanently; this is the default.
* Redirect307 - HTTP/1.1 Temporary Redirect
* Redirect308 - RFC7538 Permanent Redirect
* UseHandler - Don't redirect to the canonical path. Just call the handler instead.

#### Rationale/Usage
On a POST request, most browsers that receive a 301 will submit a GET request to the redirected URL, meaning that any data will likely be lost. If you want to handle and avoid this behavior, you may use Redirect307, which causes most browsers to resubmit the request using the original method and request body.

Since 307 is supposed to be a temporary redirect, the new 308 status code has been proposed, which is treated the same, except it indicates correctly that the redirection is permanent. The big caveat here is that the RFC is relatively recent, and older or non-compliant browsers will not handle it. Therefore its use is not recommended unless you really know what you're doing.

Finally, the UseHandler value will simply call the handler function for the pattern, without redirecting to the canonical version of the URL.

### Escaped Slashes
Go automatically processes escaped characters in a URL, converting + to a space and %XX to the corresponding character. This can present issues when the URL contains a %2f, which is unescaped to '/'. This isn't an issue for most applications, but it will prevent the router from correctly matching paths and wildcards.

For example, the pattern `/post/:post` would not match on `/post/abc%2fdef`, which is unescaped to `/post/abc/def`. The desired behavior is that it matches, and the `post` wildcard is set to `abc/def`.

Therefore, this router works with the raw URL, stored in the Request.RequestURI variable. Matching wildcards and catch-alls are then unescaped, to give the desired behavior.

TL;DR: If a requested URL contains a %2f, this router will still do the right thing. Some Go HTTP routers may not due to [Go issue 3659](https://code.google.com/p/go/issues/detail?id=3659).

## Error Handlers

### NotFoundHandler
TreeMux.NotFoundHandler can be set to provide custom 404-error handling. The default implementation is Go's `http.NotFound` function.

### MethodNotAllowedHandler
If a pattern matches, but the pattern does not have an associated handler for the requested method, the router calls the MethodNotAllowedHandler. The default
version of this handler just writes the status code `http.StatusMethodNotAllowed` and sets the response header's `Allowed` field appropriately.

### Panic Handling
TreeMux.PanicHandler can be set to provide custom panic handling. The `SimplePanicHandler` just writes the status code `http.StatusInternalServerError`. The function `ShowErrorsPanicHandler`, adapted from [gocraft/web](https://github.com/gocraft/web), will print panic errors to the browser in an easily-readable format.

## Middleware
This package provides no middleware. But there are a lot of great options out there and it's pretty easy to write your own.

# Acknowledgements

* Inspiration from Julien Schmidt's [httprouter](https://github.com/julienschmidt/httprouter)
* Show Errors panic handler from [gocraft/web](https://github.com/gocraft/web)
