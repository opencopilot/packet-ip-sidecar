FROM golang:alpine

WORKDIR /go/src/github.com/opencopilot/packet-ip-sidecar
COPY cmd cmd
RUN apk update
ENTRYPOINT [ "cmd/packet-ip-sidecar" ]