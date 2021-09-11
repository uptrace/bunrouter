module github.com/uptrace/treemux/example/rate-limiting

go 1.16

replace github.com/uptrace/treemux => ../..

replace github.com/uptrace/treemux/extra/reqlog => ../../extra/reqlog

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/go-redis/redis/v8 v8.11.3
	github.com/go-redis/redis_rate/v9 v9.1.1
	github.com/stretchr/testify v1.7.0 // indirect
	github.com/uptrace/treemux v0.8.0
	github.com/uptrace/treemux/extra/reqlog v0.8.0
)
