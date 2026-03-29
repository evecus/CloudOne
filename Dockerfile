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
    zip \
    ca-certificates \
    tzdata

# 创建数据存储目录
RUN mkdir -p /app/data

ARG TARGETARCH

COPY cloudone-linux-${TARGETARCH} /app/cloudone

# 赋予执行权限
RUN chmod +x /app/cloudone


# 暴露端口（根据你的 gin 配置修改，假设是 8080）
EXPOSE 6677

# 启动命令
CMD ["/app/cloudone"]
