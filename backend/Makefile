.PHONY: run build lint swagger test clean help

# 默认目标
.DEFAULT_GOAL := help

# 变量定义
BINARY_NAME=cryptosignal-hunter
MAIN_PATH=./cmd/server
BUILD_DIR=./build
GOPROXY=https://goproxy.cn,direct

# 运行服务（使用air热重载）
run:
	@echo "启动服务（热重载模式）..."
	@air

# 编译二进制文件
build:
	@echo "编译二进制文件..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "编译完成: $(BUILD_DIR)/$(BINARY_NAME)"

# 运行代码检查
lint:
	@echo "运行代码检查..."
	@golangci-lint run --config .golangci.yml

# 生成Swagger文档
swagger:
	@echo "生成Swagger文档..."
	@swag init -g $(MAIN_PATH)/main.go -o docs
	@echo "Swagger文档生成完成"

# 运行测试
test:
	@echo "运行测试..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "测试完成"

# 查看测试覆盖率
coverage: test
	@go tool cover -html=coverage.out

# 安装依赖
deps:
	@echo "安装依赖..."
	@GOPROXY=$(GOPROXY) go mod download
	@GOPROXY=$(GOPROXY) go mod tidy
	@echo "依赖安装完成"

# 安装开发工具
install-tools:
	@echo "安装开发工具..."
	@GOPROXY=$(GOPROXY) go install github.com/cosmtrek/air@latest
	@GOPROXY=$(GOPROXY) go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@GOPROXY=$(GOPROXY) go install github.com/swaggo/swag/cmd/swag@latest
	@echo "开发工具安装完成"

# 清理构建产物
clean:
	@echo "清理构建产物..."
	@rm -rf $(BUILD_DIR)
	@rm -rf tmp
	@rm -f coverage.out
	@echo "清理完成"

# 格式化代码
fmt:
	@echo "格式化代码..."
	@go fmt ./...
	@goimports -w .

# 帮助信息
help:
	@echo "可用命令:"
	@echo "  make run            - 启动服务（热重载模式）"
	@echo "  make build          - 编译二进制文件"
	@echo "  make lint           - 运行代码检查"
	@echo "  make swagger        - 生成Swagger文档"
	@echo "  make test           - 运行测试"
	@echo "  make coverage       - 查看测试覆盖率"
	@echo "  make deps           - 安装依赖"
	@echo "  make install-tools  - 安装开发工具"
	@echo "  make clean          - 清理构建产物"
	@echo "  make fmt            - 格式化代码"
	@echo "  make help           - 显示帮助信息"

