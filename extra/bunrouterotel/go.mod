module github.com/uptrace/bunrouter/extra/bunrouterotel

go 1.22.0

toolchain go1.24.1

replace github.com/uptrace/bunrouter => ../..

require (
	github.com/uptrace/bunrouter v1.0.22
	go.opentelemetry.io/otel v1.35.0
	go.opentelemetry.io/otel/trace v1.35.0
)
