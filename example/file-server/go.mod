module github.com/uptrace/bunrouter/example/file-server

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/uptrace/bunrouter v1.0.5
	github.com/uptrace/bunrouter/extra/reqlog v1.0.5
)
