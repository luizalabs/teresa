FROM golang:1.6

RUN mkdir -p /go/src/github.com/luizalabs/teresa-api
WORKDIR /go/src/github.com/luizalabs/teresa-api
COPY . /go/src/github.com/luizalabs/teresa-api

RUN make build-server

CMD ["./teresa-server", "run", "--port", "8080"]
EXPOSE 8080
