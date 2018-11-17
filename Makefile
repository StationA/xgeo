tools:
	@go get golang.org/x/tools/cmd/stringer

build: tools
	@go build ./...

install: build
	@go install ./...

release: build
	@CGO_ENABLED=0 GOOS=linux go install -a ./...

.PHONY: tools build install release
