FROM golang:1.16.5-alpine3.14 as builder
WORKDIR $GOPATH/src/go.k6.io/k6
RUN go install github.com/k6io/xk6/cmd/xk6@latest
RUN /go/bin/xk6 build v0.33.0 --output /go/bin/k6 --with github.com/BratSinot/xk6-graphql

FROM alpine:3.14
RUN apk add --no-cache ca-certificates && \
    adduser -D -u 12345 -g 12345 k6
COPY --from=builder /go/bin/k6 /usr/bin/k6

USER 12345
WORKDIR /home/k6
ENTRYPOINT ["k6"]
