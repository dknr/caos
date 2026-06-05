BINARY_NAME ?= caos
BUILD_DIR   ?= dist

.PHONY: build clean codegen test test-verbose

build: codegen
	cd cmd && go build -o ../$(BUILD_DIR)/$(BINARY_NAME) ./caos

codegen:
	go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen \
		-config openapi-codegen.yaml \
		openapi.yaml

test:
	go test -v -count=1 ./...

test-verbose:
	go test -v -race -count=1 ./...

clean:
	rm -rf $(BUILD_DIR)
