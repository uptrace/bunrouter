module github.com/vmihailenco/treemux/example/basic

go 1.16

replace github.com/vmihailenco/treemux => ../..

replace github.com/vmihailenco/treemux/extra/reqlog => ../../extra/reqlog

require (
	github.com/vmihailenco/treemux v0.5.3
	github.com/vmihailenco/treemux/extra/reqlog v0.5.3
)
