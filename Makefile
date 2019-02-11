BINARY=cloudsql-proxy-inject

.PHONY: build-linux build-darwin

${BINARY}:
	@CGO_ENABLED=0 go build -o ${BINARY} -a -ldflags '-s' -installsuffix cgo main.go

build-linux:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${BINARY}-linux-amd64 -a -ldflags '-s' -installsuffix cgo main.go

build-darwin:
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ${BINARY}-darwin-amd64 -a -ldflags '-s' -installsuffix cgo main.go
