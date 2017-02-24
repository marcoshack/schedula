all: test build build-client

test:
	go test github.com/marcoshack/schedula/...

build:
	go build -o bin/schedula github.com/marcoshack/schedula

build-client:
	go build -o bin/client github.com/marcoshack/schedula/examples/client

install:
	go install github.com/marcoshack/schedula
