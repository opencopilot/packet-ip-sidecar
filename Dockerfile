FROM golang:alpine

WORKDIR /go/src/github.com/opencopilot/packet-ip-sidecar
COPY . .

RUN go build -o cmd/packet-ip-sidecar

ENTRYPOINT [ "cmd/packet-ip-sidecar" ]