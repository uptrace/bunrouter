module github.com/uptrace/bunrouter/example/rate-limiting

go 1.17

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/go-redis/redis/v8 v8.11.4
	github.com/go-redis/redis_rate/v9 v9.1.2
	github.com/uptrace/bunrouter v1.0.11
	github.com/uptrace/bunrouter/extra/reqlog v1.0.11
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	golang.org/x/sys v0.0.0-20220114195835-da31bd327af9 // indirect
)
