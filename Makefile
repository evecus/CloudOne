.PHONY: all frontend backend build build-arm64-musl clean run dev-frontend

# 默认目标：完整构建（前端 + 后端，dist 编译期内嵌）
all: build

# 第一步：构建前端（必须先于后端）
frontend:
	@echo ">>> 构建前端..."
	cd frontend && npm install && npm run build
	@echo ">>> 前端构建完成"

# 第二步：构建后端（依赖前端 dist，dist 缺失时编译会直接失败）
backend: frontend
	@echo ">>> 构建后端..."
	cargo build --release
	@echo ">>> 后端构建完成: target/release/cloudone"

build: backend

# 原生编译 linux/arm64 (musl 静态链接，需在 ARM64 主机上执行)
# 前置：rustup target add aarch64-unknown-linux-musl && apt install musl-tools
build-arm64-musl: frontend
	@echo ">>> 原生编译 aarch64-unknown-linux-musl..."
	CARGO_TARGET_AARCH64_UNKNOWN_LINUX_MUSL_LINKER=musl-gcc \
		cargo build --release --target aarch64-unknown-linux-musl
	@echo ">>> 完成: target/aarch64-unknown-linux-musl/release/cloudone"

clean:
	rm -rf target
	rm -rf frontend/dist
	@echo ">>> 清理完成"

run: build
	./target/release/cloudone

# 仅运行前端开发服务（热重载，API 代理到 6677）
dev-frontend:
	cd frontend && npm run dev
