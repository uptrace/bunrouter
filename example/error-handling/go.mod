module github.com/vmihailenco/treemux/example/error-handling

go 1.16

replace github.com/vmihailenco/treemux => ../..

replace github.com/vmihailenco/treemux/extra/reqlog => ../../extra/reqlog

require (
	github.com/vmihailenco/treemux v0.6.0
	github.com/vmihailenco/treemux/extra/reqlog v0.6.0
)
