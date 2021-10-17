module github.com/uptrace/bunrouter/example/file-server

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/uptrace/bunrouter v1.0.1
	github.com/uptrace/bunrouter/extra/reqlog v0.0.0-00010101000000-000000000000
)
