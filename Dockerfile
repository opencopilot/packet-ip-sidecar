FROM golang:alpine

WORKDIR /go/src/github.com/opencopilot/packet-ip-sidecar
COPY . .

# RUN apk update; apk add curl; apk add git;
# RUN apk update;
# RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
# RUN dep ensure -vendor-only -v

RUN go build -o cmd/packet-ip-sidecar

ENTRYPOINT [ "cmd/packet-ip-sidecar" ]