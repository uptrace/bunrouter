module github.com/uptrace/bunrouter/example/all-in-one

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/klauspost/compress v1.13.6
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/rs/cors v1.8.2
	github.com/uptrace/bunrouter v1.0.10
	github.com/uptrace/bunrouter/extra/reqlog v1.0.10
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
)
