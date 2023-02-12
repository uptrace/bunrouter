module github.com/uptrace/bunrouter/extra/bunrouterotel

go 1.17

replace github.com/uptrace/bunrouter => ../..

require (
	github.com/uptrace/bunrouter v1.0.19
	go.opentelemetry.io/otel v1.13.0
	go.opentelemetry.io/otel/trace v1.13.0
)
