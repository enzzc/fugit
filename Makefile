all: test build
test:
	go test -race
build: fmt
	CGO_ENABLED=0 go build -v fugit
fmt:
	go fmt *go
