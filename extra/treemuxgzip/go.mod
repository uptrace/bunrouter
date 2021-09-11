module github.com/uptrace/treemux/extra/treemuxgzip

go 1.15

replace github.com/uptrace/treemux => ../..

require (
	github.com/klauspost/compress v1.13.5 // indirect
	github.com/uptrace/treemux v0.8.0
	github.com/vmihailenco/httpgzip v1.2.3
)
