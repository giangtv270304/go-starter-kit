SERVICE_NAME = go_starter_kit
VIEW_SPEC_ENV	=
API_SPEC    	= $(shell find . -type file -path '*apispec/*' -name 'swagger.yaml')
ifneq (,$(API_SPEC))
VIEW_SPEC_ENV += -v $(PWD)/apispec/:/usr/share/nginx/html/apispec
endif

.PHONY: build
build:
	env GOOS=$(TARGETOS) GOARCH=$(TARGETARCH) go build -ldflags="-s -w" -trimpath -o $(SERVICE_NAME)

.PHONY: migrate-create
migrate-create:
ifeq (,$(MIGRATION))
	@echo "please specific MIGRATION=<new migration filename>"
else
	migrate create -seq -ext sql -dir db/migrations $(MIGRATION)
endif

MIGRATION_USER ?= postgres
MIGRATION_PASSWORD ?= password
MIGRATION_SCHEMA ?=
.PHONY: migrate-up-local
migrate-up-local:
	migrate -path db/migrations -database "postgres://$(MIGRATION_USER):$(MIGRATION_PASSWORD)@localhost:5432/$(SERVICE_NAME)?sslmode=disable$(MIGRATION_SCHEMA)" up

.PHONY: migrate-down-local
migrate-down-local:
	migrate -path db/migrations -database "postgres://$(MIGRATION_USER):$(MIGRATION_PASSWORD)@localhost:5432/$(SERVICE_NAME)?sslmode=disable$(MIGRATION_SCHEMA)" down 1

.PHONY:pre-lint
pre-lint:
	command -v gofumpt >/dev/null 2>&1 || go install mvdan.cc/gofumpt@v0.9.2
	command -v golangci-lint >/dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

.PHONY:lint
lint: gci pre-lint
	go mod tidy
	gofumpt -l -w .
	go vet ./...
	golangci-lint run

.PHONY:pre-gci
pre-gci:
	command -v gci >/dev/null 2>&1 || go install github.com/daixiang0/gci@v0.13.7

.PHONY:gci
gci: pre-gci
	gci write --skip-generated -s standard -s default .

.PHONY:pre-api-doc
pre-api-doc:
	command -v swag >/dev/null 2>&1 || go install github.com/swaggo/swag/cmd/swag@v1.16.6

.PHONY:api-doc
api-doc: pre-api-doc
	swag init -d ./,./internal/service/ -q -o apispec -ot go,yaml --parseDependency --parseInternal

.PHONY:view-api-doc
view-api-doc:
ifeq (,$(VIEW_SPEC_ENV))
	@echo "no specs found"
else
	docker run --rm -ti \
		-p 8090:8080 \
		$(VIEW_SPEC_ENV) \
		-e URLS='[ $(shell for spec in $(API_SPEC); do \
				n=$${spec##*/}; \
				n=$${n%%.yaml}; \
				echo '{ "url": "'$$spec'", "name": "'$$n' spec" },'; \
			done)]' \
		swaggerapi/swagger-ui
endif
