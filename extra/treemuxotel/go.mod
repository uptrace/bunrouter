module github.com/uptrace/treemux/extra/treemuxotel

go 1.15

replace github.com/uptrace/treemux => ../..

require (
	github.com/uptrace/treemux v0.8.0
	go.opentelemetry.io/otel v1.0.0-RC3
	go.opentelemetry.io/otel/trace v1.0.0-RC3
)
