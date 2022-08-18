# generate version number
version=$(shell git describe --tags --long --always --dirty|sed 's/^v//')
binfile=hamon

all:
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(version)" $(binfile).go
	-@go fmt

static:
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o $(binfile).static $(binfile).go

arm:
	GOARCH=arm CGO_ENABLED=0 go build  -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o $(binfile).arm $(binfile).go
	GOARCH=arm64 CGO_ENABLED=0 go build  -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o $(binfile).arm64 $(binfile).go
version:
	@echo $(version)
