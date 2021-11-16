module github.com/uptrace/bunrouter/example/opentelemetry

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

replace github.com/uptrace/bunrouter/extra/bunrouterotel => ../../extra/bunrouterotel

require (
	github.com/klauspost/compress v1.13.6
	github.com/uptrace/bunrouter v1.0.8
	github.com/uptrace/bunrouter/extra/bunrouterotel v1.0.8
	github.com/uptrace/bunrouter/extra/reqlog v1.0.8
	github.com/uptrace/opentelemetry-go-extra/otelplay v0.1.4
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.26.1
	go.opentelemetry.io/otel/trace v1.1.0
	golang.org/x/net v0.0.0-20211105192438-b53810dc28af // indirect
)
