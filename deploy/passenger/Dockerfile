FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY passenger/go.mod passenger/go.sum ./
RUN go mod download

COPY passenger/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /passenger passenger.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /passenger .
COPY --from=builder /app/etc/passenger.yaml ./etc/passenger.yaml

EXPOSE 9002
CMD ["/passenger", "-f", "etc/passenger.yaml"]
