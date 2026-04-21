# dirsearch-go Makefile

.PHONY: build clean test run install help

# 变量
BINARY_NAME=dirsearch
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"
BUILD_DIR=build

# Go 参数
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

help: ## 显示帮助信息
	@echo 'Usage: make [target]'
	@echo ''
	@echo '可用目标:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## 编译二进制文件
	@echo "编译 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/dirsearch/main.go
	@echo "编译完成: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: ## 编译所有平台的二进制文件
	@echo "编译所有平台..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 cmd/dirsearch/main.go
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 cmd/dirsearch/main.go
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 cmd/dirsearch/main.go
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 cmd/dirsearch/main.go
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe cmd/dirsearch/main.go
	@echo "编译完成"

clean: ## 清理构建文件
	@echo "清理..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	@echo "清理完成"

test: ## 运行测试
	@echo "运行测试..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "测试完成"

test-short: ## 运行快速测试
	@echo "运行快速测试..."
	$(GOTEST) -v -short ./...

bench: ## 运行性能测试
	@echo "运行性能测试..."
	$(GOTEST) -bench=. -benchmem ./...

run: ## 运行程序
	@echo "运行 $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/dirsearch/main.go
	$(BUILD_DIR)/$(BINARY_NAME)

deps: ## 下载依赖
	@echo "下载依赖..."
	$(GOMOD) download
	$(GOMOD) tidy

fmt: ## 格式化代码
	@echo "格式化代码..."
	$(GOCMD) fmt ./...

vet: ## 代码检查
	@echo "运行 go vet..."
	$(GOCMD) vet ./...

lint: ## 代码检查 (需要 golangci-lint)
	@echo "运行 golangci-lint..."
	golangci-lint run ./...

install: ## 安装到 $GOPATH/bin
	@echo "安装 $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o $$GOPATH/bin/$(BINARY_NAME) cmd/dirsearch/main.go
	@echo "安装完成"

docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	docker build -t dirsearch-go:$(VERSION) .

docker-run: ## 运行 Docker 容器
	@echo "运行 Docker 容器..."
	docker run --rm -it dirsearch-go:$(VERSION)

.DEFAULT_GOAL := help
