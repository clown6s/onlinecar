FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY driver/go.mod driver/go.sum ./
RUN go mod download

COPY driver/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /driver driver.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /driver .
COPY --from=builder /app/etc/driver.yaml ./etc/driver.yaml

EXPOSE 9001
CMD ["/driver", "-f", "etc/driver.yaml"]
