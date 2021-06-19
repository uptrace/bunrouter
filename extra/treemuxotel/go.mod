module github.com/vmihailenco/treemux/extra/treemuxotel

go 1.15

replace github.com/vmihailenco/treemux => ../..

require (
	github.com/vmihailenco/treemux v0.6.1
	go.opentelemetry.io/otel v1.0.0-RC1
	go.opentelemetry.io/otel/trace v1.0.0-RC1
)
