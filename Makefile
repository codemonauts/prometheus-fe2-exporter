test:
	golangci-lint run

build: test
	go mod download
	CC=musl-gcc GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -linkmode external -extldflags -static"
