# ========== 构建阶段 ==========
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /passenger passenger.go

# ========== 运行阶段 ==========
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S -g 1000 appgroup && adduser -S -u 1000 -G appgroup appuser && \
    mkdir -p /var/log/onlinecar && \
    chown appuser:appgroup /var/log/onlinecar

ENV TZ=Asia/Shanghai

WORKDIR /app

USER appuser

COPY --from=builder --chown=appuser:appgroup /passenger .
COPY --from=builder --chown=appuser:appgroup /app/etc/passenger.yaml ./etc/passenger.yaml

EXPOSE 6001
CMD ["/passenger", "-f", "etc/passenger.yaml"]
