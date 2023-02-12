# Fast and flexible HTTP router for Go

[![build workflow](https://github.com/uptrace/bunrouter/actions/workflows/build.yml/badge.svg)](https://github.com/uptrace/bunrouter/actions)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/uptrace/bunrouter)](https://pkg.go.dev/github.com/uptrace/bunrouter)
[![Documentation](https://img.shields.io/badge/bunrouter-documentation-informational)](https://bunrouter.uptrace.dev/)
[![Chat](https://discordapp.com/api/guilds/752070105847955518/widget.png)](https://discord.gg/rWtp5Aj)

> BunRouter is brought to you by :star: [**uptrace/uptrace**](https://github.com/uptrace/uptrace).
> Uptrace is an open-source APM tool that supports distributed tracing, metrics, and logs. You can
> use it to monitor applications and set up automatic alerts to receive notifications via email,
> Slack, Telegram, and others. Star it as well!

**TLDR** BunRouter is as fast as httprouter, but supports middlewares, routing rules priority, and
error handling.

BunRouter is an extremely fast HTTP router for Go with unique combination of features:

- [Middlewares](https://bunrouter.uptrace.dev/guide/golang-http-middlewares.html) allow to extract
  common operations from HTTP handlers into reusable functions.
- [Error handling](https://bunrouter.uptrace.dev/guide/golang-http-error-handling.html) allows to
  further reduce the size of HTTP handlers by handling errors in middlewares.
- [Routes priority](https://bunrouter.uptrace.dev/guide/golang-router.html#routes-priority) enables
  meaningful matching priority for routing rules: first static nodes, then named nodes, lastly
  wildcard nodes.
- net/http compatible API which means using minimal API without constructing huge wrappers that try
  to do everything: from serving static files to XML generation (for example, `gin.Context` or
  `echo.Context`).

| Router          | Middlewares        | Error handling     | Routes priority    | net/http API       |
| --------------- | ------------------ | ------------------ | ------------------ | ------------------ |
| BunRouter       | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: |
| [httprouter][1] | :x:                | :x:                | :x:                | :heavy_check_mark: |
| [Chi][2]        | :heavy_check_mark: | :x:                | :heavy_check_mark: | :heavy_check_mark: |
| [Echo][3]       | :heavy_check_mark: | :heavy_check_mark: | :x:                | :x:                |
| [Gin][4]        | :heavy_check_mark: | :heavy_check_mark: | :x:                | :x:                |

[1]: https://github.com/julienschmidt/httprouter
[2]: https://github.com/go-chi/chi
[3]: https://github.com/labstack/echo
[4]: https://github.com/go-gin/gin

Learn:

- [Documentation](https://bunrouter.uptrace.dev/)
- [Reference](https://pkg.go.dev/github.com/uptrace/bunrouter)

Examples:

- [Basic example](/example/basic/)
- [http.HandlerFunc example](/example/basic-compat/)
- [httprouter.Handle example](/example/basic-verbose/)
- [CORS example](/example/cors/)
- [Basic auth example](/example/basicauth/)

Projects using BunRouter:

- [Distributed tracing tool](https://github.com/uptrace/uptrace)
- [input-output-hk/cicero](https://github.com/input-output-hk/cicero)
- [RealWorld example application](https://github.com/go-bun/bun-realworld-app)

Benchmarks:

- [web-frameworks-benchmark](https://web-frameworks-benchmark.netlify.app/result?l=go)
- [go-http-routing-benchmark](https://github.com/go-bun/go-http-routing-benchmark)

<details>
  <summary>Benchmark results</summary>

```
BenchmarkGin_Param               	16019718	        74.16 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_Param        	12560001	        95.04 ns/op	      32 B/op	       1 allocs/op
BenchmarkBunrouter_Param         	50015306	        23.81 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_Param5              	 8997234	       131.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_Param5       	 4809441	       261.3 ns/op	     160 B/op	       1 allocs/op
BenchmarkBunrouter_Param5        	10789635	       114.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_Param20             	 3953041	       302.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_Param20      	 1661373	       743.3 ns/op	     640 B/op	       1 allocs/op
BenchmarkBunrouter_Param20       	 2462354	       482.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_ParamWrite          	 9258986	       128.0 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_ParamWrite   	 9908178	       123.0 ns/op	      32 B/op	       1 allocs/op
BenchmarkBunrouter_ParamWrite    	15511226	        70.62 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_GithubStatic        	12781513	        94.17 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_GithubStatic 	30077443	        37.36 ns/op	       0 B/op	       0 allocs/op
BenchmarkBunrouter_GithubStatic  	37160334	        32.41 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_GithubParam         	 6971791	       169.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_GithubParam  	 5464755	       217.4 ns/op	      96 B/op	       1 allocs/op
BenchmarkBunrouter_GithubParam   	12047902	       101.2 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_GithubAll           	   32758	     37382 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_GithubAll    	   27324	     43932 ns/op	   13792 B/op	     167 allocs/op
BenchmarkBunrouter_GithubAll     	   57910	     20914 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_GPlusStatic         	17788194	        69.13 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_GPlusStatic  	60191341	        19.84 ns/op	       0 B/op	       0 allocs/op
BenchmarkBunrouter_GPlusStatic   	87114368	        14.06 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_GPlusParam          	10075399	       119.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_GPlusParam   	 8272046	       149.2 ns/op	      64 B/op	       1 allocs/op
BenchmarkBunrouter_GPlusParam    	37359979	        32.43 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_GPlus2Params        	 7375279	       162.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_GPlus2Params 	 6538942	       186.7 ns/op	      64 B/op	       1 allocs/op
BenchmarkBunrouter_GPlus2Params  	19681939	        61.51 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_GPlusAll            	  647716	      1752 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_GPlusAll     	  590356	      2085 ns/op	     640 B/op	      11 allocs/op
BenchmarkBunrouter_GPlusAll      	 1685287	       712.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_ParseStatic         	14566458	        76.58 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_ParseStatic  	52994076	        21.02 ns/op	       0 B/op	       0 allocs/op
BenchmarkBunrouter_ParseStatic   	50583933	        23.83 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_ParseParam          	13443874	        90.66 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_ParseParam   	 8825664	       135.6 ns/op	      64 B/op	       1 allocs/op
BenchmarkBunrouter_ParseParam    	38058278	        31.33 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_Parse2Params        	10179813	       118.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_Parse2Params 	 7801735	       152.9 ns/op	      64 B/op	       1 allocs/op
BenchmarkBunrouter_Parse2Params  	23704574	        50.78 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_ParseAll            	  394884	      3073 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_ParseAll     	  410238	      3011 ns/op	     640 B/op	      16 allocs/op
BenchmarkBunrouter_ParseAll      	  810908	      1487 ns/op	       0 B/op	       0 allocs/op
BenchmarkGin_StaticAll           	   50658	     23699 ns/op	       0 B/op	       0 allocs/op
BenchmarkHttpRouter_StaticAll    	  105313	     11518 ns/op	       0 B/op	       0 allocs/op
BenchmarkBunrouter_StaticAll     	   99674	     12188 ns/op	       0 B/op	       0 allocs/op
```

</details>

## Quickstart

Install:

```shell
go get github.com/uptrace/bunrouter
```

Run the [example](/example/basic/):

```go
package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/uptrace/bunrouter"
	"github.com/uptrace/bunrouter/extra/reqlog"
)

func main() {
	router := bunrouter.New(
		bunrouter.Use(reqlog.NewMiddleware()),
	)

	router.GET("/", indexHandler)

	router.WithGroup("/api", func(g *bunrouter.Group) {
		g.GET("/users/:id", debugHandler)
		g.GET("/users/current", debugHandler)
		g.GET("/users/*path", debugHandler)
	})

	log.Println("listening on http://localhost:9999")
	log.Println(http.ListenAndServe(":9999", router))
}

func indexHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return indexTemplate().Execute(w, nil)
}

func debugHandler(w http.ResponseWriter, req bunrouter.Request) error {
	return bunrouter.JSON(w, bunrouter.H{
		"route":  req.Route(),
		"params": req.Params().Map(),
	})
}

var indexTmpl = `
<html>
  <h1>Welcome</h1>
  <ul>
    <li><a href="/api/users/123">/api/users/123</a></li>
    <li><a href="/api/users/current">/api/users/current</a></li>
    <li><a href="/api/users/foo/bar">/api/users/foo/bar</a></li>
  </ul>
</html>
`

func indexTemplate() *template.Template {
	return template.Must(template.New("index").Parse(indexTmpl))
}
```

See the [Golang Router documentation](https://bunrouter.uptrace.dev/) for details.

## See also

- [Golang ORM](https://bun.uptrace.dev/)
- [Golang msgpack](https://msgpack.uptrace.dev/)
- [Golang message task queue](https://github.com/vmihailenco/taskq)
- [Distributed tracing tools ](https://get.uptrace.dev/compare/distributed-tracing-tools.html)
