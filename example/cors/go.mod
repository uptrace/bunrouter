module github.com/uptrace/bunrouter/example/basic

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/rs/cors v1.8.0
	github.com/uptrace/bunrouter v1.0.5
	github.com/uptrace/bunrouter/extra/reqlog v1.0.5
)
