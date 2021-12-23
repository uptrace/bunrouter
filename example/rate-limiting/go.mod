module github.com/uptrace/bunrouter/example/rate-limiting

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/go-redis/redis/v8 v8.11.4
	github.com/go-redis/redis_rate/v9 v9.1.2
	github.com/uptrace/bunrouter v1.0.9
	github.com/uptrace/bunrouter/extra/reqlog v1.0.9
)
