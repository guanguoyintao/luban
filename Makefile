# 推荐系统框架 Makefile

# 变量定义
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=recommendation-system
BINARY_UNIX=$(BINARY_NAME)_unix
MAIN_PATH=cmd/main.go

# 版本信息
VERSION?=1.0.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# 默认目标
.PHONY: all
all: test build

# 构建项目
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build completed: bin/$(BINARY_NAME)"

# 构建Linux版本
.PHONY: build-linux
build-linux:
	@echo "Building Linux version..."
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_UNIX) $(MAIN_PATH)
	@echo "Linux build completed: bin/$(BINARY_UNIX)"

# 运行测试
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# 运行单元测试
.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) -v -short ./internal/...

# 运行集成测试
.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./test/integration/...

# 运行端到端测试
.PHONY: test-e2e
test-e2e:
	@echo "Running end-to-end tests..."
	$(GOTEST) -v -tags=e2e ./test/e2e/...

# 运行基准测试
.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./pkg/...

# 代码覆盖率
.PHONY: coverage
coverage:
	@echo "Running coverage tests..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# 下载依赖
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# 更新依赖
.PHONY: update-deps
update-deps:
	@echo "Updating dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	$(GOGET) -u ./...

# 清理构建文件
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html
	@echo "Clean completed"

# 运行项目
.PHONY: run
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

# 开发模式运行
.PHONY: dev
dev:
	@echo "Running in development mode..."
	CONFIG_PATH=configs/development/config.yaml $(GOCMD) run $(MAIN_PATH)

# 生产模式运行
.PHONY: prod
prod:
	@echo "Running in production mode..."
	CONFIG_PATH=configs/production/config.yaml $(GOCMD) run $(MAIN_PATH)

# 格式化代码
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# 静态代码分析
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# 安全检查
.PHONY: security
security:
	@echo "Running security checks..."
	gosec ./...

# 生成文档
.PHONY: docs
docs:
	@echo "Generating documentation..."
	godoc -http=:6060

# Docker构建
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):$(VERSION) .
	@echo "Docker image built: $(BINARY_NAME):$(VERSION)"

# Docker运行
.PHONY: docker-run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(BINARY_NAME):$(VERSION)

# Docker Compose启动
.PHONY: compose-up
compose-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

# Docker Compose停止
.PHONY: compose-down
compose-down:
	@echo "Stopping services with Docker Compose..."
	docker-compose down

# 数据库迁移
.PHONY: migrate
migrate:
	@echo "Running database migrations..."
	$(GOCMD) run scripts/migrate/main.go

# 数据库回滚
.PHONY: rollback
rollback:
	@echo "Rolling back database migrations..."
	$(GOCMD) run scripts/migrate/main.go rollback

# 生成测试数据
.PHONY: seed
seed:
	@echo "Seeding test data..."
	$(GOCMD) run scripts/seed/main.go

# 性能测试
.PHONY: load-test
load-test:
	@echo "Running load tests..."
	$(GOCMD) run scripts/load-test/main.go

# 健康检查
.PHONY: health
health:
	@echo "Checking health status..."
	curl -f http://localhost:8080/health || exit 1

# 帮助信息
.PHONY: help
help:
	@echo "推荐系统框架 Makefile"
	@echo ""
	@echo "使用方法:"
	@echo "  make [目标]"
	@echo ""
	@echo "常用目标:"
	@echo "  build         - 构建项目"
	@echo "  test          - 运行所有测试"
	@echo "  run           - 运行项目"
	@echo "  dev           - 开发模式运行"
	@echo "  prod          - 生产模式运行"
	@echo "  clean         - 清理构建文件"
	@echo "  deps          - 下载依赖"
	@echo "  fmt           - 格式化代码"
	@echo "  lint          - 静态代码分析"
	@echo "  docker-build  - 构建Docker镜像"
	@echo "  compose-up    - 启动Docker Compose服务"
	@echo "  help          - 显示帮助信息"
	@echo ""
	@echo "高级目标:"
	@echo "  build-linux   - 构建Linux版本"
	@echo "  test-unit     - 运行单元测试"
	@echo "  test-integration - 运行集成测试"
	@echo "  test-e2e      - 运行端到端测试"
	@echo "  benchmark     - 运行基准测试"
	@echo "  coverage      - 生成代码覆盖率报告"
	@echo "  security      - 运行安全检查"
	@echo "  migrate       - 运行数据库迁移"
	@echo "  load-test     - 运行性能测试"
	@echo "  health        - 健康检查"