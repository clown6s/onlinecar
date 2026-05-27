# OnlineCar 网约车平台

基于 go-zero 的微服务架构网约车平台，包含司机服务、乘客服务和 API 网关。

## 技术栈

| 组件 | 版本 | 说明 |
|------|------|------|
| Go | 1.26.1 | 开发语言 |
| go-zero | v1.10.1 | 微服务框架 |
| goctl | v1.10.1 | 代码生成工具 |
| gRPC | v1.81.1 | 微服务间 RPC 通信 |
| etcd | v3.6.9 | 服务注册与发现（3 节点集群） |
| MySQL | 8.4 | 持久化存储（分库） |
| Redis | 8.6.2 | 缓存 / Geo 位置 / 验证码 / 限流 |
| Jaeger v2 | - | 分布式链路追踪 |
| Loki + Grafana | - | 日志聚合与可视化 |
| Docker Compose | - | 容器编排 |

## 项目结构

```
onlinecar/
├── deploy/                        # 部署配置
│   ├── docker-compose.yaml        # Docker 编排（11 个容器）
│   ├── config/                    # 生产配置模板
│   ├── driver/Dockerfile
│   ├── passenger/Dockerfile
│   └── gateway/Dockerfile
├── driver/                        # 司机 gRPC 微服务（:6000）
├── passenger/                     # 乘客 gRPC 微服务（:6001）
└── gateway/                       # HTTP API 网关（:7000）
```

## 架构图

```
                     Client (App / Web)
                           │
                    HTTP :7000
                           │
               ┌───────────┴───────────┐
               │      Gateway (REST)    │
               │  JWT + 限流 + CORS      │
               │  安全头 + Panic 恢复    │
               └───────────┬───────────┘
                           │ gRPC (via etcd)
          ┌────────────────┴────────────────┐
          │                                 │
   ┌──────┴──────┐                  ┌──────┴──────┐
   │   Driver     │                  │  Passenger  │
   │   RPC :6000  │                  │  RPC :6001  │
   └──────┬──────┘                  └──────┬──────┘
          │                                │
    ┌─────┼────────────────────────────────┼─────┐
    │     │                                │     │
  etcd   MySQL                        MySQL   Redis
 :2379 onlinecar_driver          onlinecar_passenger :6379
  (3节点集群)
```

## 服务详情

| 服务 | 端口 | etcd Key | 数据库 | 职责 |
|------|------|----------|--------|------|
| **Gateway** | :7000 | — | — | HTTP 入口，JWT 双角色认证，限流，安全头，gRPC 转发 |
| **Driver** | :6000 | `driver.rpc` | `onlinecar_driver` | 司机注册/登录，位置上报(Geo)，工作状态管理 |
| **Passenger** | :6001 | `passenger.rpc` | `onlinecar_passenger` | 乘客注册/登录，行程创建 |

### gRPC 接口

**Driver 服务：**

| 方法 | 说明 |
|------|------|
| `Register` | 司机注册（姓名/手机号/车牌/车型/密码） |
| `Login` | 司机登录（手机号+密码） |
| `CheckPhone` | 检查手机号是否已注册 |
| `UpdateLocation` | 实时位置上报（Redis Geo） |
| `GetDriver` | 查询司机详情（含车辆、工作状态） |
| `GetDriverWorkStatus` | 查询工作状态 |
| `UpdateDriverWorkStatus` | 切换工作状态（收车/出车/暂停） |

**Passenger 服务：**

| 方法 | 说明 |
|------|------|
| `Register` | 乘客注册（姓名/手机号/密码） |
| `Login` | 乘客登录（手机号+密码） |
| `CheckPhone` | 检查手机号是否已注册 |
| `GetPassenger` | 查询乘客详情 |
| `CreateTrip` | 创建行程（叫车） |

### HTTP 接口

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/v1/driver/register` | 无 | 司机注册 |
| POST | `/api/v1/driver/login` | 无 | 司机登录（返回 JWT） |
| POST | `/api/v1/driver/location` | Driver JWT | 上报实时位置 |
| POST | `/api/v1/passenger/register` | 无 | 乘客注册 |
| POST | `/api/v1/passenger/login` | 无 | 乘客登录（返回 JWT） |
| POST | `/api/v1/sms/send` | 无 | 发送短信验证码 |
| POST | `/api/v1/trip` | Passenger JWT | 发起行程（叫车） |
| GET | `/healthz` | 无 | 健康检查 |

## 数据库设计

| 数据库 | 服务 | 表 |
|--------|------|----|
| `onlinecar_driver` | Driver | driver, car, driver_car_binding_relationship, driver_work_status |
| `onlinecar_passenger` | Passenger | passenger_user, trip |

```sql
-- 行程状态：1待接单 2已接单 3行程中 4已完成 5已取消

-- 司机表 driver（onlinecar_driver）
CREATE TABLE `driver` (
    `id`        BIGINT       NOT NULL AUTO_INCREMENT COMMENT '司机ID',
    `name`      VARCHAR(50)  NOT NULL DEFAULT '' COMMENT '姓名',
    `phone`     VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '手机号',
    `password`  VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密码(加密)',
    `car_no`    VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '车牌号',
    `car_type`  VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '车型',
    `lng`       DOUBLE       NOT NULL DEFAULT 0 COMMENT '经度',
    `lat`       DOUBLE       NOT NULL DEFAULT 0 COMMENT '纬度',
    `status`    TINYINT      NOT NULL DEFAULT 1 COMMENT '0=离线 1=空闲 2=接单中 3=行程中',
    `created_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='司机表';

-- 乘客用户表 passenger_user（onlinecar_passenger）
CREATE TABLE `passenger_user` (
    `id`               BIGINT(32) UNSIGNED NOT NULL AUTO_INCREMENT,
    `gmt_create`       DATETIME NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    `gmt_modified`     DATETIME NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
    `passenger_phone`  VARCHAR(16) NULL DEFAULT NULL COMMENT '手机号',
    `passenger_name`   VARCHAR(16) NULL DEFAULT NULL COMMENT '姓名',
    `passenger_gender` TINYINT(1)  NULL DEFAULT NULL COMMENT '0未知 1男 2女',
    `state`            TINYINT(1)  NULL DEFAULT NULL COMMENT '0有效 1失效',
    `profile_photo`    VARCHAR(128) NULL DEFAULT NULL COMMENT '头像URL',
    PRIMARY KEY (`id`)
) ENGINE=InnoDB CHARACTER SET=utf8mb4 COMMENT='乘客用户表';

-- 行程表 trip（onlinecar_passenger）
CREATE TABLE `trip` (
    `id`           BIGINT       NOT NULL AUTO_INCREMENT COMMENT '行程ID',
    `passenger_id` BIGINT       NOT NULL DEFAULT 0,
    `driver_id`    BIGINT       NOT NULL DEFAULT 0,
    `origin`       VARCHAR(255) NOT NULL DEFAULT '',
    `destination`  VARCHAR(255) NOT NULL DEFAULT '',
    `origin_lng`   DOUBLE       NOT NULL DEFAULT 0,
    `origin_lat`   DOUBLE       NOT NULL DEFAULT 0,
    `dest_lng`     DOUBLE       NOT NULL DEFAULT 0,
    `dest_lat`     DOUBLE       NOT NULL DEFAULT 0,
    `status`       TINYINT      NOT NULL DEFAULT 1 COMMENT '1待接单 2已接单 3行程中 4已完成 5已取消',
    `fare`         DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_passenger` (`passenger_id`),
    KEY `idx_driver` (`driver_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='行程表';
```

## 快速开始

### 环境要求

- Go 1.26+
- protoc + goctl v1.10.1（代码生成）
- MySQL 8.4+
- etcd 3.6+
- Redis 8.x+

### 本地开发

```bash
# 1. 初始化数据库（执行上面的 DDL）

# 2. 启动 etcd、MySQL、Redis（或用 docker-compose 启动基础设施）
cd deploy && docker-compose up -d etcd-1 etcd-2 etcd-3 mysql redis

# 3. 修改各服务 etc/*.yaml 中的连接地址

# 4. 启动各服务
cd driver && go run driver.go -f etc/driver.yaml &
cd passenger && go run passenger.go -f etc/passenger.yaml &
cd gateway && go run gateway.go -f etc/gateway.yaml &
```

### Docker 部署

```bash
cd deploy/
docker-compose up -d
```

启动顺序：etcd(3节点) → MySQL → Redis → driver → passenger → gateway

## 如何添加新服务

以添加 `payment`（支付服务）为例，完整步骤如下。

### 第一步：创建 gRPC 微服务

#### 1. 创建目录和 proto

```bash
mkdir -p payment
cd payment
```

#### 2. 编写 proto 文件

```protobuf
// payment/payment.proto
syntax = "proto3";

package payment;
option go_package = "./pb";

service Payment {
  rpc CreatePayment(CreatePaymentReq) returns(PaymentResp);
  rpc GetPayment(GetPaymentReq) returns(PaymentResp);
  rpc Refund(RefundReq) returns(Empty);
}

message Empty {}

message CreatePaymentReq {
  int64  trip_id      = 1;
  int64  passenger_id = 2;
  int64  driver_id    = 3;
  double fare         = 4;
}

message GetPaymentReq { int64 payment_id = 1; }
message RefundReq { int64 payment_id = 1; }

message PaymentResp {
  int64  id           = 1;
  int64  trip_id      = 2;
  int64  passenger_id = 3;
  int64  driver_id    = 4;
  double fare         = 5;
  int64  status       = 6;  // 1待支付 2已支付 3已退款
}
```

#### 3. goctl 生成代码

```bash
cd payment/
go mod init onlinecar-payment

goctl rpc protoc payment.proto \
  --go_out=. --go-grpc_out=. --zrpc_out=. \
  --multiple --style=goZero

go mod tidy
```

#### 4. 编写配置

```yaml
# payment/etc/payment.yaml
Name: payment.rpc
ListenOn: 0.0.0.0:6002
Timeout: 30000

Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: payment.rpc

DataSource: "root:password@tcp(127.0.0.1:3306)/onlinecar_payment?charset=utf8mb4&parseTime=true&loc=Local"

RedisConf:
  Host: 127.0.0.1:6379
  Type: node
```

#### 5. 实现 Logic 层

```go
// payment/internal/logic/payment/createPaymentLogic.go
func (l *CreatePaymentLogic) CreatePayment(in *pb.CreatePaymentReq) (*pb.PaymentResp, error) {
    // 实现业务逻辑
}
```

#### 6. 注册 gRPC 服务参数（生产推荐）

在 `payment/payment.go` 中添加消息大小限制和 UnaryInterceptor：

```go
srv.AddOptions(
    grpc.MaxRecvMsgSize(4*1024*1024),
    grpc.MaxSendMsgSize(4*1024*1024),
    grpc.KeepaliveParams(keepalive.ServerParameters{
        MaxConnectionIdle: 5 * time.Minute,
        MaxConnectionAge:  30 * time.Minute,
        Time:              30 * time.Second,
        Timeout:           10 * time.Second,
    }),
    grpc.UnaryInterceptor(func(ctx context.Context, req interface{},
        info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        resp, err := handler(ctx, req)
        logx.Infof("[gRPC] %s 耗时=%v", info.FullMethod, time.Since(start))
        return resp, err
    }),
)
```

### 第二步：Gateway 接入新服务

#### 1. 拷贝 proto 生成文件

```bash
# 从 payment/pb/ 拷贝到 gateway/pb/payment/，改 package 名
mkdir -p gateway/pb/payment/
cp payment/pb/payment.pb.go gateway/pb/payment/
cp payment/pb/payment_grpc.pb.go gateway/pb/payment/
# 编辑 gateway/pb/payment/ 下的文件，将 package 声明改为 package payment
```

#### 2. Gateway 配置加 RPC 客户端

```yaml
# gateway/etc/gateway.yaml — 新增
PaymentRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: payment.rpc
```

#### 3. Config 结构体加字段

```go
// gateway/internal/config/config.go
type Config struct {
    rest.RestConf
    AuthDriver    JwtConf
    AuthPassenger JwtConf
    DriverRpc     zrpc.RpcClientConf
    PassengerRpc  zrpc.RpcClientConf
    PaymentRpc    zrpc.RpcClientConf     // ← 新增
    RedisConf     redis.RedisConf
    CORS          CORSConf
}
```

#### 4. ServiceContext 初始化客户端

```go
// gateway/internal/svc/servicecontext.go
import paymentpb "onlinecar-gateway/pb/payment"

type ServiceContext struct {
    Config       config.Config
    DriverRpc    driverpb.DriverClient
    PassengerRpc passengerpb.PassengerClient
    PaymentRpc   paymentpb.PaymentClient     // ← 新增
}

func NewServiceContext(c config.Config) *ServiceContext {
    driverConn := zrpc.MustNewClient(c.DriverRpc)
    passengerConn := zrpc.MustNewClient(c.PassengerRpc)
    paymentConn := zrpc.MustNewClient(c.PaymentRpc)     // ← 新增

    return &ServiceContext{
        Config:       c,
        DriverRpc:    driverpb.NewDriverClient(driverConn.Conn()),
        PassengerRpc: passengerpb.NewPassengerClient(passengerConn.Conn()),
        PaymentRpc:   paymentpb.NewPaymentClient(paymentConn.Conn()),   // ← 新增
    }
}
```

#### 5. 注册 HTTP 路由（routes.go）

```go
// gateway/internal/handler/routes.go
server.AddRoutes([]rest.Route{
    {Method: http.MethodPost, Path: "/api/v1/payment",
        Handler: passengerRateLimit(passengerAuth(paymentHandler.CreatePaymentHandler(svcCtx)))},
    {Method: http.MethodGet, Path: "/api/v1/payment/:id",
        Handler: passengerRateLimit(passengerAuth(paymentHandler.GetPaymentHandler(svcCtx)))},
})
```

#### 6. 实现 Gateway Logic（使用 MapRpcError）

```go
// gateway/internal/logic/payment/createPaymentLogic.go
func (l *CreatePaymentLogic) CreatePayment(req *types.CreatePaymentReq) (*types.PaymentResp, error) {
    uid := ctxdata.GetUID(l.ctx)
    resp, err := l.svcCtx.PaymentRpc.CreatePayment(l.ctx, &paymentpb.CreatePaymentReq{
        TripId:      req.TripId,
        PassengerId: uid,
        Fare:        req.Fare,
    })
    if err != nil {
        l.Errorf("PaymentRpc.CreatePayment 失败, err=%v", err)
        return nil, errorx.MapRpcError(err)   // ← 重要：统一错误映射
    }
    return &types.PaymentResp{
        Id: resp.Id, Status: resp.Status,
    }, nil
}
```

### 第三步：Docker 部署

#### 1. 新增 Dockerfile

```dockerfile
# deploy/payment/Dockerfile
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY payment/go.mod payment/go.sum ./
RUN go mod download
COPY payment/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /payment payment.go

FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata netcat-openbsd
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /payment .
COPY --from=builder /app/etc/payment.yaml ./etc/payment.yaml
EXPOSE 6002
CMD ["/payment", "-f", "etc/payment.yaml"]
```

#### 2. docker-compose 加服务

```yaml
# deploy/docker-compose.yaml — 新增
  payment:
    build:
      context: ../payment
      dockerfile: ../deploy/payment/Dockerfile
    depends_on:
      etcd-1:
        condition: service_healthy
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - onlinecar
    restart: on-failure
    healthcheck:
      test: ["CMD", "nc", "-z", "localhost", "6002"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 15s
    deploy:
      resources:
        limits:
          cpus: "1.0"
          memory: 512M
    logging:
      driver: json-file
      options:
        max-size: "50m"
        max-file: "3"

  # gateway 的 depends_on 加上 payment
  gateway:
    depends_on:
      driver:
        condition: service_healthy
      passenger:
        condition: service_healthy
      payment:                             # ← 新增
        condition: service_healthy
```

### 添加新服务检查清单

| # | 位置 | 操作 |
|---|------|------|
| 1 | `payment/payment.proto` | 定义 gRPC 服务和消息 |
| 2 | `payment/` | goctl 生成代码 + 实现 Logic |
| 3 | `payment/payment.go` | 配置 MaxRecvMsgSize + UnaryInterceptor |
| 4 | `payment/etc/payment.yaml` | 配置 etcd key + MySQL + Redis |
| 5 | `gateway/pb/payment/` | 拷贝 proto 生成文件，改 package 名 |
| 6 | `gateway/etc/gateway.yaml` | 加 RPC 客户端配置 |
| 7 | `gateway/internal/config/config.go` | Config 结构体加字段 |
| 8 | `gateway/internal/svc/servicecontext.go` | 初始化 gRPC 客户端 |
| 9 | `gateway/internal/handler/routes.go` | 注册 HTTP 路由 |
| 10 | `gateway/internal/logic/payment/` | 实现 HTTP → gRPC 转发（用 `MapRpcError`） |
| 11 | `deploy/payment/Dockerfile` | 编写 Dockerfile（含 healthcheck） |
| 12 | `deploy/docker-compose.yaml` | 添加服务定义 + gateway depends_on |

## 安全加固

| 措施 | 说明 |
|------|------|
| JWT 双角色隔离 | 司机/乘客独立密钥，支持环境变量覆盖 |
| 错误信息脱敏 | RPC 层不透传内部 err，Gateway 统一返回通用文案 |
| 限流 | 基于 IP 的请求限流，仅信任内网代理的 `X-Real-IP` |
| 安全响应头 | HSTS / X-Frame-Options / X-Content-Type-Options / Referrer-Policy |
| 验证码原子校验 | Redis Lua 脚本防并发重放 |
| gRPC 消息大小限制 | MaxRecvMsgSize 4MB，防大包攻击 |
| gRPC 拦截器 | 统一日志 + 耗时 + 错误码记录 |
| Loki 不暴露端口 | 仅容器内网络可访问 |
| Gateway 健康检查 | Docker healthcheck 确保容器就绪 |

## 错误码

| Code | 说明 |
|------|------|
| 0 | 成功 |
| -1 | 业务错误 |
| 400 | 参数错误 |
| 401 | 未授权（Token 缺失/无效/过期） |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |
