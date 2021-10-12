all: test build
test: fmt
	go test -race
build: fmt
	CGO_ENABLED=0 go build -v fugit
fmt:
	go fmt *go
