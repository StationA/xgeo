tools:
	go get golang.org/x/tools/cmd/stringer
	go get github.com/pointlander/peg

build: tools
	peg lang/grammar.peg
	stringer -type=Op -trimprefix=Op vm
	go build

install: build
	go install
