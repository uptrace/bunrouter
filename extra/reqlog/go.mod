module github.com/uptrace/bunrouter/extra/reqlog

go 1.23.0

toolchain go1.24.1

replace github.com/uptrace/bunrouter => ../..

require (
	github.com/fatih/color v1.18.0
	github.com/felixge/httpsnoop v1.0.4
	github.com/uptrace/bunrouter v1.0.23
	go.opentelemetry.io/otel/trace v1.35.0
)

require (
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	go.opentelemetry.io/otel v1.35.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
)
