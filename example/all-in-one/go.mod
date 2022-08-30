module github.com/uptrace/bunrouter/example/all-in-one

go 1.17

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/klauspost/compress v1.15.9
	github.com/rs/cors v1.8.2
	github.com/uptrace/bunrouter v1.0.19
	github.com/uptrace/bunrouter/extra/reqlog v1.0.19
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	go.opentelemetry.io/otel v1.9.0 // indirect
	go.opentelemetry.io/otel/trace v1.9.0 // indirect
	golang.org/x/sys v0.0.0-20220829200755-d48e67d00261 // indirect
)
