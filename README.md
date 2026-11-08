# CloudOne (Rust)

CloudOne 的 Rust 后端版本（对齐 Go 版全部行为），前端 Vue3 应用在编译期内嵌进单个二进制文件，部署只需一个可执行文件。

## 与 Go 版的对应关系

- 路由 / API / 鉴权 / 限流 / WebDAV / 终端 / 分享 / 公开文件行为与 Go 版（gin + bbolt + gorilla/websocket）一致；
- 存储层使用 redb（纯 Rust 嵌入式 KV，单文件 `cloudone.redb`，无外部依赖），数据字段语义对齐；
- 静态资源由 `build.rs` 在编译期从 `frontend/dist` 内嵌（等价于 Go 的 `//go:embed all:frontend/dist`），`dist` 缺失时直接编译失败；
- JWT HS256 + AES-256-GCM 加密存储格式与 Go 版互通（相同 master key 推导）。

## 构建

```bash
# 1. 构建前端（必须先于后端）
cd frontend && npm install && npm run build && cd ..

# 2. 构建后端（前端产物编译期内嵌）
cargo build --release
# 产物: target/release/cloudone
```

或直接 `make build`。

### 编译 linux arm64 (musl 静态)

在 **ARM64 主机**（或 GitHub Actions 的 `ubuntu-24.04-arm` 原生 Runner）上原生编译，无需 Docker / cross：

```bash
rustup target add aarch64-unknown-linux-musl
sudo apt install musl-tools
make build-arm64-musl
# 产物: target/aarch64-unknown-linux-musl/release/cloudone
```

GitHub Actions 见 `.github/workflows/build-arm64-musl.yml`（自动构建前端 + ARM64 Runner 原生 musl 编译 + Release 发布）。

## 运行

```bash
# 与 Go 版相同的命令行参数
./cloudone --config ./data --dir /path/to/storage --host 0.0.0.0 --port 6677

# 环境变量
export CLOUDONE_MASTER_KEY="xxx"      # 可选，缺省自动生成 master.key
export CLOUDONE_JWT_SECRET="xxx"      # 可选，建议设置（否则重启后登录失效）
export CLOUDONE_ORIGIN="https://xx"   # 可选，允许的跨域来源（默认拒绝任意跨域）
```

默认在当前目录创建 `./data`（`cloudone.redb` 数据库 + `master.key`）与 `./data/storage`（文件存储）。

## 功能

- 📁 文件浏览、上传（单文件 / 整个文件夹，保留目录结构）、下载、删除、移动、复制、批量操作
- 🗜️ 压缩（zip / tar / tar.gz）与解压（zip / tar / tar.gz / tgz / gz，防 zip-slip）
- 🔗 分享链接（过期时间、最大访问次数、目录分享浏览 / 子路径下载、raw 直链）
- 🌐 公开文件（`/public`、`/pub/list`、`/pub/dl`、`/raw/...`，支持 wget/curl）
- 🖥️ Web 终端（WebSocket + PTY，登录 shell，跟随文件管理器目录）
- 📤 远程 URL 下载（SSRF 防护：DNS 解析后校验、拒绝内网、重定向逐跳校验；可选 GitHub 加速代理，按延迟择优）
- 🔒 JWT 认证、bcrypt 密码、token 版本吊销、登录 / WebDAV 滑动窗口限流
- 📡 WebDAV（Basic 认证、独立密码、PROPFIND/LOCK 等，兼容挂载客户端）
