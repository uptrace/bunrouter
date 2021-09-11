module github.com/uptrace/treemux/example/error-handling

go 1.16

replace github.com/uptrace/treemux => ../..

replace github.com/uptrace/treemux/extra/reqlog => ../../extra/reqlog

require (
	github.com/uptrace/treemux v0.8.0
	github.com/uptrace/treemux/extra/reqlog v0.8.0
)
