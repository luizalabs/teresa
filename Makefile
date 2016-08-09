SWAGGER_SPEC=swagger.yml
TERESA_API_NAME=Teresa

help:
	@echo "Targets: run-api-server, validate, gen-api-server, gen-api-client\n"
	@echo "run-api-server: run the API server on localhost:8080"
	@echo "validate: validate swagger file $(SWAGGER_SPEC) aginst swagger specification 2.0"
	@echo "gen-api-server: generate the API server code as described by $(SWAGGER_SPEC)"
	@echo "gen-api-client: generate the API client code as described by $(SWAGGER_SPEC)"
	@echo

all: help

gen-api-server:
	swagger generate server -A $(TERESA_API_NAME) -f $(SWAGGER_SPEC)

run-api-server:
	go run ./cmd/teresa-server/main.go --http-server 127.0.0.1:8080

gen-api-client:
	swagger generate client -A $(TERESA_API_NAME) -f $(SWAGGER_SPEC)

validate:
	swagger validate $(SWAGGER_SPEC)

swagger-docs:
	go run docs/webserver.go swagger.yml
