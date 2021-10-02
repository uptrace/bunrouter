module github.com/uptrace/bunrouter/example/basic

go 1.16

replace github.com/uptrace/bunrouter => ../..

replace github.com/uptrace/bunrouter/extra/reqlog => ../../extra/reqlog

require (
	github.com/uptrace/bunrouter v0.8.0
	github.com/uptrace/bunrouter/extra/reqlog v0.8.0
	golang.org/x/sys v0.0.0-20211001092434-39dca1131b70 // indirect
)
