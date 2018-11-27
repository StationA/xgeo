tools:
	go get golang.org/x/tools/cmd/stringer
	go get github.com/pointlander/peg

build: tools
	peg gx/grammar.peg
	stringer -type=Op -trimprefix=Op gx
	go build ./...

install: build
	go install ./...
