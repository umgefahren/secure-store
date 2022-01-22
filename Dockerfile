FROM golang:latest AS builder
WORKDIR /go/src/app
COPY . .

RUN go get -d ./...
RUN go build

FROM ubuntu:latest

COPY --from=builder /go/src/app/secure-store /usr/local/bin/secure-store
ENV GIN_MODE=release
EXPOSE 8080
CMD ["secure-store"]