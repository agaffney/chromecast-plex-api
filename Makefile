BINARY=chromecast-plex-api
DOCKER_IMAGE=agaffney1548/chromecast-plex-api

.PHONY: all clean

all: $(BINARY)

clean:
	rm -f $(BINARY)

$(BINARY): $(shell find -name '*.go')
	GOOS=linux CGO_ENABLED=0 go build -o $(BINARY)

.PHONY: run test image

run:
	DEBUG=1 go run main.go

test:
	go test -v ./...

image:
	docker build -t $(DOCKER_IMAGE) .
