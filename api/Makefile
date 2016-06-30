SWAGGER_SPEC=swagger.yml
TERESA_API_NAME=Teresa

help:
	@echo "Targets: api-server, api-client\n"
	@echo "validate: validate swagger file $(SWAGGER_SPEC) aginst swagger specification 2.0"
	@echo "api-server: generate the API server code as described by $(SWAGGER_SPEC)"
	@echo "api-client: generate the API client code as described by $(SWAGGER_SPEC)"
	@echo

all: help

gen-api-server:
	swagger generate server -A $(TERESA_API_NAME) -f $(SWAGGER_SPEC)

run-api-server:
	PORT=8080 go run ./cmd/teresa-server/main.go

gen-api-client:
	swagger generate client -A $(TERESA_API_NAME) -f $(SWAGGER_SPEC)

validate:
	swagger validate $(SWAGGER_SPEC)

swagger-docs:
	go run docs/webserver.go swagger.yml
