#!/bin/sh

# 2. 如果开启了 CF 且 Token 不为空，则启动 cloudflared 隧道
if [ "$CF" = "true" ] && [ -n "$TOKEN" ]; then
    echo "Starting cloudflared tunnel..."
    /app/cloudflared tunnel --no-autoupdate run --token "$TOKEN" &
fi

# 3. 启动主程序
# 使用 exec 确保主程序成为 PID 1，能够接收系统信号
echo "Starting main application..."
exec /app/cloudone "$@"
