# 放弃 Alpine，改用 Debian 精简版
FROM debian:bookworm-slim

ARG TARGETARCH
RUN apt-get update && apt-get install -y ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN mkdir -p /app/data
RUN mkdir -p /app/data/storage

# 根据目标架构复制对应二进制
COPY cloudone-linux-${TARGETARCH} ./cloudone
RUN chmod +x ./cloudone

VOLUME ["/app/data"]

EXPOSE 6677

CMD ["./cloudone"]
