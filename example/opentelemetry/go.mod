module github.com/uptrace/bunrouter/example/opentelemetry

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

replace github.com/uptrace/bunrouter/extra/bunrouterotel => ../../extra/bunrouterotel

require (
	github.com/klauspost/compress v1.13.6
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/uptrace/bunrouter v1.0.10
	github.com/uptrace/bunrouter/extra/bunrouterotel v1.0.10
	github.com/uptrace/bunrouter/extra/reqlog v1.0.10
	github.com/uptrace/opentelemetry-go-extra/otelplay v0.1.7
	github.com/uptrace/uptrace-go v1.3.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.28.0
	go.opentelemetry.io/otel/trace v1.3.0
	golang.org/x/net v0.0.0-20220107192237-5cfca573fb4d // indirect
	google.golang.org/genproto v0.0.0-20220107163113-42d7afdf6368 // indirect
)
