FROM golang:1.25.0-alpine3.22 AS builder
WORKDIR /go/src/pub
COPY . .
RUN \
    apk add -U --no-cache \
        protoc \
        protobuf-dev \
        make \
        git \
        curl \
        ca-certificates && \
    make build && \
    curl -L -o emoji_utf8_lexicon.txt https://raw.githubusercontent.com/drankou/go-vader/master/data/emoji_utf8_lexicon.txt && \
    curl -L -o vader_lexicon.txt https://raw.githubusercontent.com/drankou/go-vader/master/data/vader_lexicon.txt

FROM scratch
COPY --from=builder /go/src/pub/pub /bin/pub
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/pub/emoji_utf8_lexicon.txt /data/emoji_utf8_lexicon.txt
COPY --from=builder /go/src/pub/vader_lexicon.txt /data/vader_lexicon.txt
ENTRYPOINT ["/bin/pub"]
