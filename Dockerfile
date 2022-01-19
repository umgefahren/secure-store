FROM golang:alpine AS builder
WORKDIR /go/src/app
COPY . .

RUN go get -d -cpu qemu,vx=off ./...
RUN go build

FROM alpine:latest

COPY --from=builder /go/src/app/secure-store /usr/local/bin/secure-store
ENV GIN_MODE=release
EXPOSE 8080
CMD ["secure-store"]