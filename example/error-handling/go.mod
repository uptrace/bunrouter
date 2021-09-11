module github.com/uptrace/treemux/example/error-handling

go 1.16

replace github.com/uptrace/treemux => ../..

replace github.com/uptrace/treemux/extra/reqlog => ../../extra/reqlog

require (
	github.com/fatih/color v1.12.0 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/uptrace/treemux v0.7.3
	github.com/uptrace/treemux/extra/reqlog v0.7.3
	golang.org/x/sys v0.0.0-20210910150752-751e447fb3d0 // indirect
)
