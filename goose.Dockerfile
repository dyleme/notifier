ARG GOLANG_VERSION=1.21
FROM golang:${GOLANG_VERSION} AS builder

ARG GOOSE_VERSION=latest
RUN go install github.com/pressly/goose/v3/cmd/goose@${GOOSE_VERSION}

FROM alpine:3.18.3
RUN apk add libc6-compat
COPY --from=builder /go/bin/goose /usr/local/bin
WORKDIR /data
ENTRYPOINT [ "/usr/local/bin/goose"]
CMD [ "status" ]

