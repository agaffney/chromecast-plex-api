FROM golang:1.15 AS build

WORKDIR /code

COPY . .

RUN make

FROM alpine:3.13

COPY --from=build /code/chromecast-plex-api /usr/local/bin

ENTRYPOINT /usr/local/bin/chromecast-plex-api
