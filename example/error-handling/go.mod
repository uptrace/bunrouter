module github.com/vmihailenco/treemux/example/error-handling

go 1.16

replace github.com/vmihailenco/treemux => ../..

replace github.com/vmihailenco/treemux/extra/reqlog => ../../extra/reqlog

require (
	github.com/vmihailenco/treemux v0.5.5
	github.com/vmihailenco/treemux/extra/reqlog v0.5.5
	golang.org/x/sys v0.0.0-20210309074719-68d13333faf2 // indirect
)
