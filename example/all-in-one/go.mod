module github.com/uptrace/bunrouter/example/all-in-one

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/klauspost/compress v1.13.6
	github.com/rs/cors v1.8.0
	github.com/uptrace/bunrouter v1.0.9
	github.com/uptrace/bunrouter/extra/reqlog v1.0.9
)
