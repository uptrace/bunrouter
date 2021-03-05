module github.com/vmihailenco/treemux/extra/treemuxgzip

go 1.15

replace github.com/vmihailenco/treemux => ../..

require (
	github.com/klauspost/compress v1.11.11 // indirect
	github.com/vmihailenco/httpgzip v1.2.3
	github.com/vmihailenco/treemux v0.5.3
)
