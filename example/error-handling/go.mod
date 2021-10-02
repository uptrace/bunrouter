module github.com/uptrace/bunrouter/example/error-handling

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/uptrace/bunrouter v1.0.0-rc.1
	github.com/uptrace/bunrouter/extra/reqlog v1.0.0-rc.1
)
