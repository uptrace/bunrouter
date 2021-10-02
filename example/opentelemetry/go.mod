module github.com/uptrace/bunrouter/example/opentelemetry

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

replace github.com/uptrace/bunrouter/extra/bunrouterotel => ../../extra/bunrouterotel

require (
	github.com/uptrace/bunrouter v0.8.0
	github.com/uptrace/bunrouter/extra/bunrouterotel v0.0.0-20211001140734-672077266ee1
	github.com/uptrace/bunrouter/extra/reqlog v0.8.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.24.0
	go.opentelemetry.io/otel v1.0.1
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.0.1
	go.opentelemetry.io/otel/internal/metric v0.24.0 // indirect
	go.opentelemetry.io/otel/sdk v1.0.1
	golang.org/x/sys v0.0.0-20211001092434-39dca1131b70 // indirect
)
