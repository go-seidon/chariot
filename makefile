
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
	go test ./... -coverprofile coverage.out
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

.PHONY: generate-proto
generate-proto:
	buf generate api/

.PHONY: verify-swagger
verify-swagger:
	swagger-cli bundle api/rest-v1/main.yml --type json > generated/rest-v1/main.all.json
	swagger-cli validate generated/rest-v1/main.all.json 

.PHONY: generate-swagger
generate-swagger:
	swagger-cli bundle api/rest-v1/main.yml --type yaml > generated/rest-v1/main.all.yml

.PHONY: generate-oapi
generate-oapi:
	make generate-swagger
	make generate-oapi-type

.PHONY: generate-oapi-type
generate-oapi-type:
	oapi-codegen -old-config-style -config api/rest-v1/type.gen.yaml generated/rest-v1/main.all.yml

.PHONY: run-grpc-app
run-grpc-app:
	go run cmd/grpc-app/main.go

.PHONY: run-rest-app
run-rest-app:
	go run cmd/rest-app/main.go

.PHONY: run-hybrid-app
run-hybrid-app:
	go run cmd/hybrid-app/main.go

.PHONY: build-grpc-app
build-grpc-app:
	go build -o ./build/grpc-app/ ./cmd/grpc-app/main.go

.PHONY: build-rest-app
build-rest-app:
	go build -o ./build/rest-app/ ./cmd/rest-app/main.go

.PHONY: build-hybrid-app
build-hybrid-app:
	go build -o ./build/hybrid-app/ ./cmd/hybrid-app/main.go

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

ifeq (migrate-mongo,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "migrate-mongo"
  MIGRATE_MONGO_RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(MIGRATE_MONGO_RUN_ARGS):dummy;@:)
endif

ifeq (migrate-mongo-create,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "migrate-mongo-create"
  MIGRATE_MONGO_RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(MIGRATE_MONGO_RUN_ARGS):dummy;@:)
endif

dummy: ## used by migrate script as do-nothing targets
	@:


MYSQL_DB_URI=mysql://admin:123456@tcp(localhost:3308)/chariot?x-tls-insecure-skip-verify=true
MONGO_DB_URI=mongodb://admin:123456@localhost:27030/chariot

.PHONY: migrate-mysql
migrate-mysql:
	migrate -database "$(MYSQL_DB_URI)" -path ./migration/mysql $(MIGRATE_MYSQL_RUN_ARGS)

.PHONY: migrate-mysql-create
migrate-mysql-create:
	migrate create -dir migration/mysql -ext .sql $(MIGRATE_MYSQL_RUN_ARGS)

.PHONY: migrate-mongo
migrate-mongo:
	migrate -database "$(MONGO_DB_URI)" -path ./migration/mongo $(MIGRATE_MONGO_RUN_ARGS)

.PHONY: migrate-mongo-create
migrate-mongo-create:
	migrate create -dir migration/mongo -ext .json $(MIGRATE_MONGO_RUN_ARGS)
