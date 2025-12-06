BINARY_NAME=udns-api
MAIN_FILE=main.go
GOPRIVATE=git.code.oa.com,git.woa.com

all: test build
build: build-linux

test:
	@go test -v ./...
clean:
	@go clean && rm -f $(BINARY_NAME)
run:
	@go build -o $(BINARY_NAME) -v && ./$(BINARY_NAME) start
deps:
	@go mod tidy

# 交叉编译
build-linux: ## 默认编译linux
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME) -v