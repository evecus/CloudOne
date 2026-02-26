# 直接使用 build job 产出的二进制，按平台选择对应文件
FROM alpine:3.19

ARG TARGETARCH
ARG VERSION=latest

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

RUN mkdir -p /app/data
RUN mkdir -p /app/data/storage

# 根据目标架构复制对应二进制
COPY cloudone-linux-${TARGETARCH} ./cloudone
RUN chmod +x ./cloudone

VOLUME ["/app/data"]

EXPOSE 6677

CMD ["./cloudone"]
