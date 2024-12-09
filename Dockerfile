FROM golang:1.23.4-alpine3.20 AS builder
WORKDIR /go/src/pub
COPY . .
RUN \
    apk add -U --no-cache \
        protoc \
        protobuf-dev \
        make \
        git \
        ca-certificates && \
    make build

FROM scratch
COPY --from=builder /go/src/pub/pub /bin/pub
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/bin/pub"]
