BINARY=chromecast-plex-api
DOCKER_IMAGE=agaffney/chromecast-plex-api

.PHONY: all clean

all: $(BINARY)

clean:
	rm -f $(BINARY)

$(BINARY): $(shell find -name '*.go')
	go build

.PHONY: run test

run:
	DEBUG=1 go run main.go

test:
	go test -v ./...
