FROM alpine:3.19

ARG TARGETARCH
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# 明确创建目录
RUN mkdir -p /app/data/storage

# 建议在 COPY 前打印一下，或者利用通配符防止 arch 字符串微小差异
COPY cloudone-linux-${TARGETARCH}* ./cloudone
RUN chmod +x ./cloudone

# 如果 CGO_ENABLED=1 且用 Alpine，必须加这个软链接(或者换 base image)
#RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

VOLUME ["/app/data"]
EXPOSE 6677
CMD ["./cloudone"]
