module github.com/uptrace/treemux/example/rate-limiting

go 1.16

replace github.com/uptrace/treemux => ../..

replace github.com/uptrace/treemux/extra/reqlog => ../../extra/reqlog

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/fatih/color v1.12.0 // indirect
	github.com/go-redis/redis/v8 v8.11.3
	github.com/go-redis/redis_rate/v9 v9.1.1
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	github.com/uptrace/treemux v0.7.3
	github.com/uptrace/treemux/extra/reqlog v0.7.3
	golang.org/x/sys v0.0.0-20210910150752-751e447fb3d0 // indirect
)
