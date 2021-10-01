module github.com/uptrace/bunrouter/example/rate-limiting

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/go-redis/redis/v8 v8.11.3
	github.com/go-redis/redis_rate/v9 v9.1.1
	github.com/mattn/go-colorable v0.1.11 // indirect
	github.com/uptrace/bunrouter v0.8.0
	github.com/uptrace/bunrouter/extra/reqlog v0.8.0
	golang.org/x/sys v0.0.0-20210930212924-f542c8878de8 // indirect
)
