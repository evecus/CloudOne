# 直接使用 build job 产出的二进制，按平台选择对应文件
FROM alpine:3.19

ARG TARGETARCH
ARG VERSION=latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

RUN mkdir -p /app/data && chmod 755 /app/data
RUN mkdir -p /app/data/storage && chmod 755 /app/data/storage

# 根据目标架构复制对应二进制
COPY cloudone-linux-${TARGETARCH} ./cloudone
RUN chmod +x ./cloudone

VOLUME ["/app/data"]

EXPOSE 6677

LABEL org.opencontainers.image.title="CloudOne" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.description="CloudOne File Manager"

CMD ["./cloudone"]
