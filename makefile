
default: help

.PHONY: help
help:
	@echo 'chariot'
	@echo 'usage: make [target] ...'

.PHONY: install-tool
install-tool:
	go get -u github.com/golang/mock/gomock
	go get -u github.com/golang/mock/mockgen

.PHONY: install-dependency
install-dependency:
	go mod tidy
	go mod verify
	go mod vendor

.PHONY: clean-dependency
clean-dependency:
	rm -f go.sum
	rm -rf vendor
	go clean -modcache

.PHONY: install
install:
	go install -v ./...

.PHONY: test
test:
	go test ./.../ -p 1 -race -coverprofile coverage.out
	go tool cover -func coverage.out | grep ^total:

.PHONY: test-coverage
test-coverage:
	ginkgo -r -v -p -race --progress --randomize-all --randomize-suites -cover -coverprofile="coverage.out"

.PHONY: test-unit
test-unit:
	ginkgo -r -v -p -race --label-filter="unit" -cover -coverprofile="coverage.out"

.PHONY: test-integration
test-integration:
	ginkgo -r -v -p -race --label-filter="integration" -cover -coverprofile="coverage.out"

.PHONY: test-watch-unit
test-watch-unit:
	ginkgo watch -r -v -p -race --trace --label-filter="unit"

.PHONY: test-watch-integration
test-watch-integration:
	ginkgo watch -r -v -p -race --trace --label-filter="integration"

.PHONY: generate-mock
generate-mock:
	mockgen -package=mock_auth -source internal/auth/client.go -destination=internal/auth/mock/client_mock.go
	mockgen -package=mock_auth -source internal/auth/basic.go -destination=internal/auth/mock/basic_mock.go
	mockgen -package=mock_barrel -source internal/barrel/barrel.go -destination=internal/barrel/mock/barrel_mock.go
	mockgen -package=mock_file -source internal/file/file.go -destination=internal/file/mock/file_mock.go
	mockgen -package=mock_healthcheck -source internal/healthcheck/health.go -destination=internal/healthcheck/mock/health_mock.go
	mockgen -package=mock_queue -source internal/queue/queue.go -destination=internal/queue/mock/queue_mock.go
	mockgen -package=mock_repository -source internal/repository/auth.go -destination=internal/repository/mock/auth_mock.go
	mockgen -package=mock_repository -source internal/repository/barrel.go -destination=internal/repository/mock/barrel_mock.go
	mockgen -package=mock_repository -source internal/repository/file.go -destination=internal/repository/mock/file_mock.go
	mockgen -package=mock_repository -source internal/repository/provider.go -destination=internal/repository/mock/provider_mock.go
	mockgen -package=mock_restapp -source internal/restapp/server.go -destination=internal/restapp/mock/server_mock.go
	mockgen -package=mock_session -source internal/session/session.go -destination=internal/session/mock/session_mock.go
	mockgen -package=mock_signature -source internal/signature/signature.go -destination=internal/signature/mock/signature_mock.go
	mockgen -package=mock_storage -source internal/storage/storage.go -destination=internal/storage/mock/storage_mock.go
	mockgen -package=mock_storage -source internal/storage/router/router.go -destination=internal/storage/mock/router_mock.go

.PHONY: generate-proto
generate-proto:
	buf generate api/

.PHONY: verify-swagger
verify-swagger:
	swagger-cli bundle api/restapp/main.yml --type json > api/restapp/main.all.json
	swagger-cli validate api/restapp/main.all.json 

.PHONY: generate-swagger
generate-swagger:
	swagger-cli bundle api/restapp/main.yml --type yaml > api/restapp/main.all.yml

.PHONY: generate-oapi
generate-oapi:
	make generate-swagger
	make generate-oapi-type
	make generate-oapi-server

.PHONY: generate-oapi-type
generate-oapi-type:
	oapi-codegen -old-config-style -config api/restapp/type.gen.yaml api/restapp/main.all.yml

.PHONY: generate-oapi-server
generate-oapi-server:
	oapi-codegen -old-config-style -config api/restapp/server.gen.yaml api/restapp/main.all.yml

.PHONY: run-restapp
run-restapp:
	go run cmd/restapp/main.go

.PHONY: build-restapp
build-restapp:
	go build -o ./build/restapp/ ./cmd/restapp/main.go

ifeq (migrate-mysql,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "migrate-mysql"
  MIGRATE_MYSQL_RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(MIGRATE_MYSQL_RUN_ARGS):dummy;@:)
endif

ifeq (migrate-mysql-create,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "migrate-mysql-create"
  MIGRATE_MYSQL_RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(MIGRATE_MYSQL_RUN_ARGS):dummy;@:)
endif

dummy: ## used by migrate script as do-nothing targets
	@:


MYSQL_DB_URI=mysql://admin:123456@tcp(localhost:3411)/chariot?x-tls-insecure-skip-verify=true

.PHONY: migrate-mysql
migrate-mysql:
	migrate -database "$(MYSQL_DB_URI)" -path ./migration/mysql $(MIGRATE_MYSQL_RUN_ARGS)

.PHONY: migrate-mysql-create
migrate-mysql-create:
	migrate create -dir migration/mysql -ext .sql $(MIGRATE_MYSQL_RUN_ARGS)
