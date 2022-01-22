FROM golang:latest AS builder
WORKDIR /go/src/app
COPY . .

# RUN apt-get update
# RUN apt-get install -y build-essential libsqlite3-dev
RUN go get -d ./...
RUN go build

FROM ubuntu:latest

COPY --from=builder /go/src/app/secure-store /usr/local/bin/secure-store
ENV GIN_MODE=release
EXPOSE 8080
CMD ["secure-store"]