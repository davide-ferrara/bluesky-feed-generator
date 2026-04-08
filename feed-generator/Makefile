.PHONY: run build clean test analyze collect help

run:
	go run .

build:
	go build -o ./bin/feedgen .

clean:
	rm -rf ./bin/feedgen *.json

test:
	go test -v ./...
