module github.com/uptrace/bunrouter/extra/bunroutergzip

go 1.15

replace github.com/uptrace/bunrouter => ../..

require (
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/uptrace/bunrouter v0.8.0
	github.com/vmihailenco/httpgzip v1.2.3
)
