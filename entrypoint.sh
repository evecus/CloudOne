#!/bin/sh

# 如果开启了 CF 且 Token 不为空，则以后台模式启动 cloudflared 隧道
if [ "$CF" = "true" ] && [ -n "$TOKEN" ]; then
    echo "Starting cloudflared tunnel..."
    /app/cloudflared tunnel --no-autoupdate run --token "$TOKEN" &
fi

# 启动主程序
# 使用 exec 确保主程序能接收到系统信号（如 SIGTERM）
exec /app/cloudone "$@"
