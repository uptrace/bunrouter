module github.com/uptrace/bunrouter/example/basic-verbose

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/uptrace/bunrouter v1.0.0-rc.1
	github.com/uptrace/bunrouter/extra/reqlog v1.0.0-rc.1
	golang.org/x/sys v0.0.0-20211002104244-808efd93c36d // indirect
)
