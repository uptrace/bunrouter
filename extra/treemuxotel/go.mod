module github.com/vmihailenco/treemux/extra/treemuxotel

go 1.15

replace github.com/vmihailenco/treemux => ../..

require (
	github.com/vmihailenco/treemux v0.5.5
	go.opentelemetry.io/otel v0.18.0
	go.opentelemetry.io/otel/trace v0.18.0
)
