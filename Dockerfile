FROM golang:1.8

RUN mkdir -p /go/src/github.com/luizalabs/teresa-api
WORKDIR /go/src/github.com/luizalabs/teresa-api
COPY . /go/src/github.com/luizalabs/teresa-api

RUN make build-server

ENTRYPOINT ["./teresa-server"]
CMD ["run"]
EXPOSE 50051
