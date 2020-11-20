all:
	go test ./...
	go test ./... -short -race
	go test ./... -run=NONE -bench=. -benchmem
	env GOOS=linux GOARCH=386 go test ./...
	go vet
	golangci-lint run

tag:
	git tag $(VERSION)
	git tag extra/reqlog/$(VERSION)
	git tag extra/treemuxgzip/$(VERSION)
	git tag extra/treemuxotel/$(VERSION)
