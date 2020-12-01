module github.com/vmihailenco/treemux/example/basic

go 1.16

replace github.com/vmihailenco/treemux => ../..

replace github.com/vmihailenco/treemux/extra/reqlog => ../../extra/reqlog

require (
	github.com/vmihailenco/treemux v0.1.2
	github.com/vmihailenco/treemux/extra/reqlog v0.0.0-00010101000000-000000000000
)
