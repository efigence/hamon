# generate version number
version=$(shell git describe --tags --long --always --dirty|sed 's/^v//')
binfile=hamon
.PHONY: static
all:
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(version)" $(binfile).go
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(version)" -o hamon-ipset-loader cmd/ipset-loader.go
	-@go fmt

static:
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o $(binfile).static $(binfile).go
	CGO_ENABLED=0 go build -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o hamon-ipset-loader.static cmd/ipset-loader.go
	-@go fmt

arm:
	CGO_ENABLED=0 GOARCH=arm CGO_ENABLED=0 go build  -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o $(binfile).arm $(binfile).go
	CGO_ENABLED=0 GOARCH=arm64 CGO_ENABLED=0 go build  -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o $(binfile).arm64 $(binfile).go
	CGO_ENABLED=0 GOARCH=arm CGO_ENABLED=0 go build  -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o hamon-ipset-loader.arm cmd/ipset-loader.go
	CGO_ENABLED=0 GOARCH=arm64 CGO_ENABLED=0 go build  -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o hamon-ipset-loader.arm64 cmd/ipset-loader.go
version:
	@echo $(version)
