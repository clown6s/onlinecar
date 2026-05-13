FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY gateway/go.mod gateway/go.sum ./
RUN go mod download

COPY gateway/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway gateway.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /gateway .
COPY --from=builder /app/etc/gateway.yaml ./etc/gateway.yaml

EXPOSE 8888
CMD ["/gateway", "-f", "etc/gateway.yaml"]
