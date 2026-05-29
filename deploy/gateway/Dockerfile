# ========== 构建阶段 ==========
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# 先复制 common 模块（go.mod 中 replace onlinecar-common => ../common）
COPY common/go.mod common/go.sum /common/

# 复制服务依赖文件并下载
COPY gateway/go.mod gateway/go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

# 复制源代码
COPY common/ /common/
COPY gateway/ .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /gateway .

# ========== 运行阶段 ==========
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata netcat-openbsd && \
    addgroup -S -g 1000 appgroup && adduser -S -u 1000 -G appgroup appuser && \
    mkdir -p /var/log/onlinecar && \
    chown appuser:appgroup /var/log/onlinecar

ENV TZ=Asia/Shanghai

WORKDIR /app

USER appuser

COPY --from=builder --chown=appuser:appgroup /gateway /gateway
COPY --from=builder --chown=appuser:appgroup /app/etc/gateway.yaml ./etc/gateway.yaml

EXPOSE 7000
HEALTHCHECK --interval=10s --timeout=5s --retries=3 --start-period=15s \
  CMD wget --spider -q http://localhost:7000/healthz || exit 1
CMD ["/gateway", "-f", "etc/gateway.yaml"]
