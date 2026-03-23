# 使用轻量级的 Alpine Linux 作为基础镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 安装必要的工具
# --no-cache 可以减小镜像层体积
RUN apk add --no-cache \
    curl \
    wget \
    unzip \
    tar \
    ca-certificates \
    tzdata

# 创建数据存储目录
RUN mkdir -p /app/data

# 将 Github Actions 构建好的二进制文件复制到容器内
# 注意：文件名需与你 workflow 中编译输出的文件名一致
# 如果你是通过构建多架构镜像，buildx 会自动处理 context
COPY cloudone-linux-${TARGETARCH} /app/cloudone

# 赋予执行权限
RUN chmod +x /app/cloudone


# 暴露端口（根据你的 gin 配置修改，假设是 8080）
EXPOSE 6677

# 启动命令
CMD ["/app/cloudone"]
