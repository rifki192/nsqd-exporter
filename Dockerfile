# Build Image
FROM golang:1.12 AS builder

ADD ./ /go/src/github.com/rifki192/nsqd-exporter

WORKDIR /go/src/github.com/rifki192/nsqd-exporter

RUN GO111MODULE=on go mod vendor
# Build binary
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-s -w" -mod=vendor -v

EXPOSE 12500

# Image for certs
FROM alpine:latest as certs
# Get certs
RUN apk --update add ca-certificates

# Base Image
FROM alpine:latest

RUN mkdir /nsqd-exporter
WORKDIR /nsqd-exporter
# Setup
COPY --from=builder /go/src/github.com/rifki192/nsqd-exporter/nsqd-exporter .
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 12500

ENTRYPOINT ["./nsqd-exporter"]
