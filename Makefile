build:
	peg lang/grammar.peg
	stringer -type=Op -trimprefix=Op vm
	go build

install: build
	go install
