FROM golang:1.8

RUN mkdir -p /go/src/github.com/luizalabs/teresa
WORKDIR /go/src/github.com/luizalabs/teresa
COPY . /go/src/github.com/luizalabs/teresa

RUN make build-server

ENTRYPOINT ["./teresa-server"]
CMD ["run"]
EXPOSE 50051
