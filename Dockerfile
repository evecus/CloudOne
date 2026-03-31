# 使用轻量级的 Alpine Linux 作为基础镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 安装必要的工具，添加 openssh
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    openssh

# --- SSH 配置 ---
RUN ssh-keygen -A && \
    echo "root:1234" | chpasswd && \
    sed -i 's/#PermitRootLogin.*/PermitRootLogin yes/' /etc/ssh/sshd_config && \
    sed -i 's/#PasswordAuthentication.*/PasswordAuthentication yes/' /etc/ssh/sshd_config

# 创建数据存储目录
RUN mkdir -p /app/data

ARG TARGETARCH

COPY cloudone-linux-${TARGETARCH} /app/cloudone

# 赋予执行权限
RUN chmod +x /app/cloudone

# 暴露业务端口 6677 和 SSH 端口 22
EXPOSE 6677

# 启动命令：先启动 sshd (后台)，再启动 cloudone (前台)
# 使用 -D 参数让 sshd 在后台运行，或者直接执行然后接主程序
CMD ["/bin/sh", "-c", "/usr/sbin/sshd && /app/cloudone"]
