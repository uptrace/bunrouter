module github.com/uptrace/bunrouter/extra/bunrouterotel

go 1.17

replace github.com/uptrace/bunrouter => ../..

require (
	github.com/uptrace/bunrouter v1.0.17
	go.opentelemetry.io/otel v1.7.0
	go.opentelemetry.io/otel/trace v1.7.0
)
