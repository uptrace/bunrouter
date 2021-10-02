module github.com/uptrace/bunrouter/extra/bunrouterotel

go 1.15

replace github.com/uptrace/bunrouter => ../..

require (
	github.com/uptrace/bunrouter v1.0.0-rc.1
	go.opentelemetry.io/otel v1.0.1
	go.opentelemetry.io/otel/trace v1.0.1
)
