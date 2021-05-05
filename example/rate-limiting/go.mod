module github.com/vmihailenco/treemux/example/rate-limiting

go 1.16

replace github.com/vmihailenco/treemux => ../..

replace github.com/vmihailenco/treemux/extra/reqlog => ../../extra/reqlog

require (
	github.com/go-redis/redis/v8 v8.8.2
	github.com/go-redis/redis_rate/v9 v9.1.1
	github.com/vmihailenco/treemux v0.6.0
	github.com/vmihailenco/treemux/extra/reqlog v0.6.0
	go.opentelemetry.io/otel v0.20.0 // indirect
)
