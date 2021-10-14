module github.com/uptrace/bunrouter/example/rate-limiting

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/go-redis/redis/v8 v8.11.3
	github.com/go-redis/redis_rate/v9 v9.1.1
	github.com/uptrace/bunrouter v1.0.0
	github.com/uptrace/bunrouter/extra/reqlog v1.0.0
)
