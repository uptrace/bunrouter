module github.com/uptrace/bunrouter/example/basic

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/uptrace/bunrouter v1.0.10
	github.com/uptrace/bunrouter/extra/reqlog v1.0.10
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
)
