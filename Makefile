.PHONY: test build clean

test:
	go test -v -race ./...

build:
	go build -o app .

clean:
	rm -f app
