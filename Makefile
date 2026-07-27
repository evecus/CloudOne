.PHONY: all build frontend backend clean run

# 默认目标：完整构建
all: build

# 第一步：构建前端（必须先于后端）
frontend:
	@echo ">>> 构建前端..."
	cd frontend && npm install && npm run build
	@echo ">>> 前端构建完成"

# 第二步：构建后端（依赖前端 dist）
backend: frontend
	@echo ">>> 构建后端..."
	CGO_ENABLED=1 go build -ldflags="-s -w" -o cloudone .
	@echo ">>> 后端构建完成: ./cloudone"

# 完整构建
build: backend

# 交叉编译 linux/amd64
build-linux-amd64: frontend
	@echo ">>> 交叉编译 linux/amd64..."
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o cloudone-linux-amd64 .

# 清理构建产物
clean:
	rm -f cloudone cloudone-linux-amd64 cloudone-linux-arm64
	rm -rf frontend/dist
	@echo ">>> 清理完成"

# 构建并运行（用于本地开发）
run: build
	./cloudone

# 仅运行前端开发服务（热重载）
dev-frontend:
	cd frontend && npm run dev
