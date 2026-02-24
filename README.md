# CloudOne

一个以 Go + Vue3 构建的现代云端文件管理工具。

![License](https://img.shields.io/github/license/yourname/cloudone)
![Build](https://img.shields.io/github/actions/workflow/status/yourname/cloudone/build.yml)

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
./cloudone-linux-amd64
```

程序会在运行目录下自动创建 `data/` 目录存储数据库和文件。

访问 `http://your-ip:6677` 即可使用。

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

```bash
# 构建前端
cd frontend
npm install
npm run build
cd ..

# 构建后端
go build -o cloudone .

# 运行
./cloudone
```

## 环境要求

- Linux amd64 或 arm64
- 端口 6677
