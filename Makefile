tools:
	@go get golang.org/x/tools/cmd/stringer

build: tools
	@go build ./...

install: build
	@go install ./...

target:
	mkdir -p target

release: build target
	@CGO_ENABLED=0 GOOS=linux go build -a -o target/xgeo ./cmd/xgeo

release-all: build target
	@CGO_ENABLED=0 GOOS=darwin GOARCH=386 go build -a -o target/xgeo.darwin-386 ./cmd/xgeo
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -o target/xgeo.darwin-amd64 ./cmd/xgeo
	@CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -a -o target/xgeo.linux-386 ./cmd/xgeo
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o target/xgeo.linux-amd64 ./cmd/xgeo
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -a -o target/xgeo.linux-arm ./cmd/xgeo
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -o target/xgeo.linux-arm64 ./cmd/xgeo
	@CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -a -o target/xgeo.windows-386.exe ./cmd/xgeo
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -o target/xgeo.windows-amd64.exe ./cmd/xgeo

clean:
	@rm -rf target

.PHONY: tools build install release release-all clean
