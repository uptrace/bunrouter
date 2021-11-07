module github.com/uptrace/bunrouter/example/panic-recover

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/uptrace/bunrouter v1.0.3
	github.com/uptrace/bunrouter/extra/reqlog v1.0.3
	golang.org/x/sys v0.0.0-20211106132015-ebca88c72f68 // indirect
)
