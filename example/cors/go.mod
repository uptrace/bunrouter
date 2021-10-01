module github.com/uptrace/bunrouter/example/basic

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/fatih/color v1.13.0 // indirect
	github.com/mattn/go-colorable v0.1.11 // indirect
	github.com/uptrace/bunrouter v0.8.0
	github.com/uptrace/bunrouter/extra/reqlog v0.8.0
	golang.org/x/sys v0.0.0-20210930212924-f542c8878de8 // indirect
)
