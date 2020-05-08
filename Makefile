.PHONY: build test install-dev docker

DOCKER_OPTS ?= --rm
PROTO_BUILD_DIR = ./pkg

help:
	@echo "Service building targets"
	@echo "  build : build service command"
	@echo "  test  : run test suites"
	@echo "  docker: build docker image"
	@echo "  install-dev: install dev dependencies"
	@echo "Env:"
	@echo "  DOCKER_OPTS : default docker build options (default : $(DOCKER_OPTS))"
	@echo "  TEST_ARGS : Arguments to pass to go test call"

generate-api:
	if [ ! -d "$(PROTO_BUILD_DIR)" ]; then mkdir -p "$(PROTO_BUILD_DIR)"; else  find "$(PROTO_BUILD_DIR)/api" -type f -delete &&  mkdir -p "$(PROTO_BUILD_DIR)"; fi
	find ./api/*.proto -maxdepth 1 -type f -exec protoc {} --go_opt=paths=source_relative --go_out=plugins=grpc:$(PROTO_BUILD_DIR) \;


build:
	go build .

test:
	go test $(TEST_ARGS) ./...

install-dev:
	go get github.com/golang/mock/gomock
	go install github.com/golang/mock/mockgen

docker:
	docker build $(DOCKER_OPTS) .