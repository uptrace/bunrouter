module github.com/uptrace/bunrouter/extra/reqlog

go 1.17

replace github.com/uptrace/bunrouter => ../..

require (
	github.com/fatih/color v1.14.1
	github.com/felixge/httpsnoop v1.0.3
	github.com/uptrace/bunrouter v1.0.20
	go.opentelemetry.io/otel/trace v1.13.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	go.opentelemetry.io/otel v1.13.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
)
