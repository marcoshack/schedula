all: test build

test:
	go test ./...

build:
	mkdir -p bin
	go build -o bin/schedula
