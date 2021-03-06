FROM golang:alpine as build

RUN apk update && apk add git

WORKDIR /go/src/github.com/a-tal/esi-isk
ADD . /go/src/github.com/a-tal/esi-isk
RUN go install -v ./...

FROM alpine:latest
MAINTAINER Adam Talsma <adam@talsma.ca>

COPY --from=build /go/bin/worker /worker
COPY --from=build /etc/ssl/certs /etc/ssl/certs

# nobody
USER 65534:65534
CMD ["/worker"]
