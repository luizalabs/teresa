FROM golang:1.15 AS builder

WORKDIR /go/src/github.com/luizalabs/teresa
COPY . /go/src/github.com/luizalabs/teresa

RUN make build-server

FROM debian:10-slim
RUN apt-get update && \
apt-get install ca-certificates -y libc6 &&\
rm -rf /var/lib/apt/lists/* &&\
rm -rf /var/cache/apt/archives/*

WORKDIR /app
COPY --from=builder /go/src/github.com/luizalabs/teresa . 

ENTRYPOINT ["./teresa-server"]
CMD ["run"]
EXPOSE 50051
