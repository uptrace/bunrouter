module github.com/uptrace/bunrouter/extra/bunrouterotel

go 1.15

replace github.com/uptrace/bunrouter => ../..

require (
	github.com/uptrace/bunrouter v1.0.10
	go.opentelemetry.io/otel v1.3.0
	go.opentelemetry.io/otel/trace v1.3.0
)
