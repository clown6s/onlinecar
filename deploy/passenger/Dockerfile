# ========== 构建阶段 ==========
FROM golang:1.26.1-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /passenger .

# ========== 运行阶段 ==========
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata netcat-openbsd && \
    addgroup -S -g 1000 appgroup && adduser -S -u 1000 -G appgroup appuser && \
    mkdir -p /var/log/onlinecar && \
    chown appuser:appgroup /var/log/onlinecar

ENV TZ=Asia/Shanghai

WORKDIR /app

USER appuser

COPY --from=builder --chown=appuser:appgroup /passenger /passenger
COPY --from=builder --chown=appuser:appgroup /app/etc/passenger.yaml ./etc/passenger.yaml

EXPOSE 6001
HEALTHCHECK --interval=10s --timeout=5s --retries=3 --start-period=15s \
  CMD nc -z localhost 6001 || exit 1
CMD ["/passenger", "-f", "etc/passenger.yaml"]
