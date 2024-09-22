module github.com/uptrace/bunrouter/extra/reqlog

go 1.22

replace github.com/uptrace/bunrouter => ../..

require (
	github.com/fatih/color v1.17.0
	github.com/felixge/httpsnoop v1.0.4
	github.com/uptrace/bunrouter v1.0.22
	go.opentelemetry.io/otel/trace v1.30.0
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	go.opentelemetry.io/otel v1.30.0 // indirect
	golang.org/x/sys v0.25.0 // indirect
)
