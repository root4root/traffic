BINARY_NAME=traffic
BUILD_DIR=./bin

all: build

build:
	@env GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux_amd64 -tags netgo .
	@env GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux_arm64 -tags netgo .
default:
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .

clean:
	@rm -f -r $(BUILD_DIR)

run: build
	@cd $(BUILD_DIR) && ./$(BINARY_NAME)

test:
	@go test ./... -v

.PHONY: all build default clean run test