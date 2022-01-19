FROM golang:alpine AS builder
WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go build

FROM alpine:latest

COPY --from=builder /go/src/app/secure-store /usr/local/bin/secure-store
ENV GIN_MODE=release
CMD ["secure-store"]