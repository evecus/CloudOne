# CloudOne

一个以 Go + Vue3 构建的现代云端文件管理工具。

## 功能特性

- 📁 文件浏览、上传、下载、删除、移动、复制
- 🗂️ 创建文件夹和文本文件
- 🔗 分享文件（生成分享链接）
- 🌐 公开文件（直链访问，支持 wget/curl）
- 🔒 基于 JWT 的认证
- 🌏 中英文双语支持
- 💙 蓝白主题，现代灵动界面

## 快速开始

### 下载二进制

从 [Releases](../../releases) 页面下载对应平台的二进制文件：

- `cloudone-linux-amd64` — Linux x86_64
- `cloudone-linux-arm64` — Linux ARM64

### 运行

```bash
chmod +x cloudone-linux-amd64

# 推荐：设置持久化 JWT 密钥（重启后登录状态不会失效）
export CLOUDONE_JWT_SECRET="your-long-random-secret-here"

./cloudone-linux-amd64
```

> **注意：** 若不设置 `CLOUDONE_JWT_SECRET`，程序每次重启会随机生成新密钥，所有已登录的会话将失效。生产环境请务必设置此变量。

程序会在运行目录下自动创建 `data/` 目录存储数据库和文件。

访问 `http://your-ip:6677` 即可使用。

### 以 systemd 服务运行（推荐）

```ini
[Unit]
Description=CloudOne File Manager
After=network.target

[Service]
ExecStart=/opt/cloudone/cloudone
WorkingDirectory=/opt/cloudone
Environment=CLOUDONE_JWT_SECRET=your-long-random-secret-here
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### 公开文件访问

将文件或文件夹设置为公开后，可通过直链访问：

```bash
# 查看公开文件列表
curl http://your-ip:6677/public

# 直接下载公开文件（类似 raw.githubusercontent.com）
wget http://your-ip:6677/raw/path/to/file.yaml
curl http://your-ip:6677/raw/path/to/file.yaml
```

### 分享链接

登录后对任意文件/文件夹点击「分享」即可生成分享链接：

```
http://your-ip:6677/s/AbCdEfGh
http://your-ip:6677/s/AbCdEfGh/raw   # 直接下载
```

## 从源码构建

### 前提条件

- Go 1.21+（需要 CGO，用于 SQLite）
- Node.js 18+
- GCC（Linux 系统通常已预装）

### 使用 Makefile（推荐）

```bash
# 一键完整构建（先构建前端，再编译后端）
make

# 构建并运行
make run

# 清理构建产物
make clean
```

### 手动构建

```bash
# 1. 构建前端（必须先于后端，Go embed 需要 frontend/dist 存在）
cd frontend
npm install
npm run build
cd ..

# 2. 构建后端
CGO_ENABLED=1 go build -ldflags="-s -w" -o cloudone .

# 3. 运行
export CLOUDONE_JWT_SECRET="your-long-random-secret-here"
./cloudone
```

> ⚠️ **常见错误：** 若直接运行 `go build .` 而未提前构建前端，会报错：
> `pattern frontend/dist: directory prefix does not exist`
> 原因是 `main.go` 使用了 `//go:embed frontend/dist`，该目录必须在编译时存在。

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `CLOUDONE_JWT_SECRET` | JWT 签名密钥，生产环境必须设置 | 随机生成（重启失效） |

## 端口

程序监听 `:6677`，暂不支持通过环境变量或配置文件修改端口。

