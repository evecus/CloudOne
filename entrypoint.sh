#!/bin/sh

# 1. 启动 SSH 服务
# -D 选项让 sshd 不脱离控制台（通常用于调试），但在脚本中我们通常直接启动它
# 如果你的镜像里没有配置好 host keys，可能需要先执行 ssh-keygen
echo "Starting SSH server..."
/usr/sbin/sshd

# 2. 如果开启了 CF 且 Token 不为空，则启动 cloudflared 隧道
if [ "$CF" = "true" ] && [ -n "$TOKEN" ]; then
    echo "Starting cloudflared tunnel..."
    /app/cloudflared tunnel --no-autoupdate run --token "$TOKEN" &
fi

# 3. 启动主程序
# 使用 exec 确保主程序成为 PID 1，能够接收系统信号
echo "Starting main application..."
exec /app/cloudone "$@"
