# OnlineCar 网约车平台

基于 go-zero 的微服务架构网约车平台，包含司机服务、乘客服务和 API 网关。

## 技术栈

| 组件 | 版本 | 说明 |
|------|------|------|
| Go | 1.25.0 | 开发语言 |
| go-zero | v1.10.1 | 微服务框架 |
| goctl | v1.10.1 | 代码生成工具 |
| gRPC | v1.81.0 | 微服务间 RPC 通信 |
| etcd | 3.5 | 服务注册与发现 |
| MySQL | 8.0 | 持久化存储 |
| Redis | 7.x | 缓存 |
| Docker Compose | 3.8 | 容器编排 |

## 项目结构

```
onlinecar/
├── deploy/                    # 部署配置
│   ├── docker-compose.yaml    # Docker 编排
│   ├── init.sql               # 数据库初始化脚本
│   ├── driver/Dockerfile
│   ├── passenger/Dockerfile
│   └── gateway/Dockerfile
├── driver/                    # 司机 gRPC 微服务（端口 9001）
├── passenger/                 # 乘客 gRPC 微服务（端口 9002）
└── gateway/                   # HTTP API 网关（端口 8888）
```

## 架构图

```
                         Client (App / Web)
                                │
                         HTTP :8888
                                │
                    ┌───────────┴───────────┐
                    │      Gateway (REST)    │
                    │  JWT Auth + Route       │
                    └───────────┬───────────┘
                                │ gRPC (via etcd)
                   ┌────────────┴────────────┐
                   │                         │
            ┌──────┴──────┐          ┌──────┴──────┐
            │   Driver     │          │  Passenger  │
            │   RPC :9001  │          │  RPC :9002  │
            └──────┬──────┘          └──────┬──────┘
                   │                        │
         ┌────────┼────────────────────────┼────────┐
         │        │                        │        │
       etcd     MySQL                  MySQL     Redis
      :2379  onlinecar_driver     onlinecar_passenger :6379
```

---

## 服务详情

### Driver 服务（司机）

- **模块**: `onlinecar-driver`
- **端口**: 9001
- **etcd Key**: `driver.rpc`
- **数据库**: `onlinecar_driver`
- **职责**: 司机注册/登录、位置上报、工作状态管理

#### gRPC 接口

| 方法 | 请求 | 响应 | 说明 |
|------|------|------|------|
| `Register` | `RegisterReq` | `RegisterResp` | 司机注册 |
| `Login` | `LoginReq` | `LoginResp` | 司机登录 |
| `UpdateLocation` | `UpdateLocationReq` | `Empty` | 实时位置上报 |
| `GetDriver` | `GetDriverReq` | `DriverInfo` | 查询司机详情（含车辆信息） |
| `GetDriverWorkStatus` | `GetDriverWorkStatusReq` | `DriverWorkStatusInfo` | 查询工作状态 |
| `UpdateDriverWorkStatus` | `UpdateDriverWorkStatusReq` | `Empty` | 切换工作状态 |

#### Protobuf 消息

```protobuf
message RegisterReq {
  string driver_name    = 1; // 司机姓名
  string driver_phone   = 2; // 手机号（唯一）
  string password       = 3; // 密码
  string vehicle_no     = 4; // 车牌号
  string vehicle_type   = 5; // 车辆类型
  int64  commercial_type = 6; // 服务类型：1网约车 2巡游 3合乘
}

message DriverInfo {
  int64  id              = 1;
  string driver_name     = 2;
  string driver_phone    = 3;
  int64  driver_gender   = 4;  // 1男 2女
  string license_id      = 5;  // 驾驶证号
  string certificate_no  = 6;  // 从业资格证号
  int64  commercial_type = 7;  // 服务类型
  string vehicle_no      = 8;  // 车牌号
  string vehicle_type    = 9;  // 车辆类型
  string brand           = 10; // 车辆品牌
  string model           = 11; // 车辆型号
  int64  state           = 12; // 0有效 1失效
  int64  work_status     = 13; // 0收车 1出车 2暂停
}
```

### Passenger 服务（乘客）

- **模块**: `onlinecar-passenger`
- **端口**: 9002
- **etcd Key**: `passenger.rpc`
- **数据库**: `onlinecar_passenger`
- **职责**: 乘客注册/登录、行程创建

#### gRPC 接口

| 方法 | 请求 | 响应 | 说明 |
|------|------|------|------|
| `Register` | `RegisterReq` | `RegisterResp` | 乘客注册 |
| `Login` | `LoginReq` | `LoginResp` | 乘客登录（手机号） |
| `GetPassenger` | `GetPassengerReq` | `PassengerInfo` | 查询乘客详情 |
| `CreateTrip` | `CreateTripReq` | `TripResp` | 创建行程 |

#### Protobuf 消息

```protobuf
message CreateTripReq {
  int64  passenger_id = 1;  // 乘客ID（由 Gateway JWT 注入）
  string origin       = 2;  // 出发地
  string destination  = 3;  // 目的地
  double origin_lng   = 4;  // 出发经度
  double origin_lat   = 5;  // 出发纬度
  double dest_lng     = 6;  // 目的经度
  double dest_lat     = 7;  // 目的纬度
}

message TripResp {
  int64  id           = 1;
  int64  passenger_id = 2;
  int64  driver_id    = 3;  // 待匹配时为 0
  string origin       = 4;
  string destination  = 5;
  int64  status       = 6;  // 1待接单 2已接单 3行程中 4已完成 5已取消
  double fare         = 7;
}
```

### Gateway 服务（API 网关）

- **模块**: `onlinecar-gateway`
- **端口**: 8888
- **职责**: 统一 HTTP 入口、JWT 认证、请求路由转发到后端 gRPC 服务

#### JWT 双域名认证

司机和乘客使用独立的 JWT 密钥，互不干扰：

```yaml
AuthDriver:
  AccessSecret: "onlinecar-driver-secret-2026"
  AccessExpire: 86400s    # 24小时

AuthPassenger:
  AccessSecret: "onlinecar-passenger-secret-2026"
  AccessExpire: 86400s
```

JWT Claims 结构：
```json
{
  "uid": 12345,
  "subject": "driver",    // 或 "passenger"
  "exp": 1700000000,
  "iat": 1699913600
}
```

#### HTTP 接口

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/v1/driver/register` | 无 | 司机注册 |
| POST | `/api/v1/driver/login` | 无 | 司机登录（返回 JWT） |
| POST | `/api/v1/driver/location` | Driver JWT | 上报实时位置 |
| POST | `/api/v1/passenger/register` | 无 | 乘客注册 |
| POST | `/api/v1/passenger/login` | 无 | 乘客登录（返回 JWT） |
| POST | `/api/v1/trip` | Passenger JWT | 发起行程（叫车） |

#### 统一响应格式

```json
{
  "code": 0,
  "msg": "ok",
  "data": { ... }
}
```

- `code: 0` 成功，`code: -1` 业务错误，`code: 401` 未授权

#### 请求示例

```bash
# 司机注册
curl -X POST http://localhost:8888/api/v1/driver/register \
  -H "Content-Type: application/json" \
  -d '{"name":"张师傅","phone":"13800138000","carNo":"京A12345","carType":"舒适型","password":"123456"}'

# 司机登录
curl -X POST http://localhost:8888/api/v1/driver/login \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800138000","password":"123456"}'

# 上报位置（需要 Bearer Token）
curl -X POST http://localhost:8888/api/v1/driver/location \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"lng":116.397,"lat":39.916}'

# 乘客注册
curl -X POST http://localhost:8888/api/v1/passenger/register \
  -H "Content-Type: application/json" \
  -d '{"name":"李先生","phone":"13900139000","password":"123456"}'

# 发起行程（需要 Bearer Token）
curl -X POST http://localhost:8888/api/v1/trip \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"origin":"中关村","destination":"望京","originLng":116.31,"originLat":40.055,"destLng":116.48,"destLat":39.998}'
```

---

## 数据库设计

### 数据库划分

| 数据库 | 服务 | 表 |
|--------|------|----|
| `onlinecar_driver` | Driver | driver, car, driver_car_binding_relationship, driver_work_status |
| `onlinecar_passenger` | Passenger | passenger_user, trip |

### 司机库 `onlinecar_driver`

```sql
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
```

### 乘客库 `onlinecar_passenger`

```sql
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

---

## 配置说明

### Driver (`driver/etc/driver.yaml`)

```yaml
Name: driver.rpc
ListenOn: 0.0.0.0:9001
Etcd:
  Hosts: [10.211.55.5:2379]
  Key: driver.rpc
DataSource: "root:123456@tcp(10.211.55.5:3306)/onlinecar_driver?charset=utf8mb4&parseTime=true&loc=Local"
RedisConf:
  Host: 10.211.55.5:6379
  Type: node
```

### Passenger (`passenger/etc/passenger.yaml`)

```yaml
Name: passenger.rpc
ListenOn: 0.0.0.0:9002
Etcd:
  Hosts: [10.211.55.5:2379]
  Key: passenger.rpc
DataSource: "root:123456@tcp(10.211.55.5:3306)/onlinecar_passenger?charset=utf8mb4&parseTime=true&loc=Local"
RedisConf:
  Host: 10.211.55.5:6379
  Type: node
```

### Gateway (`gateway/etc/gateway.yaml`)

```yaml
Name: gateway-api
Host: 0.0.0.0
Port: 8888
AuthDriver:
  AccessSecret: "onlinecar-driver-secret-2026"
  AccessExpire: 86400s
AuthPassenger:
  AccessSecret: "onlinecar-passenger-secret-2026"
  AccessExpire: 86400s
DriverRpc:
  Etcd:
    Hosts: [10.211.55.5:2379]
    Key: driver.rpc
PassengerRpc:
  Etcd:
    Hosts: [10.211.55.5:2379]
    Key: passenger.rpc
```

---

## 快速开始

### 环境要求

- Go 1.25+
- protoc + goctl（代码生成）
- MySQL 8.0+
- etcd 3.5+
- Redis 7.x+

### 本地开发

```bash
# 1. 初始化数据库（执行 deploy/init.sql）

# 2. 启动 etcd、MySQL、Redis

# 3. 修改各服务的 etc/*.yaml 中的连接地址

# 4. 启动各服务
cd driver && go run driver.go -f etc/driver.yaml &
cd passenger && go run passenger.go -f etc/passenger.yaml &
cd gateway && go run gateway.go -f etc/gateway.yaml &
```

### 代码生成

```bash
# Driver 服务
cd driver/
goctl rpc protoc driver.proto --go_out=. --go-grpc_out=. --zrpc_out=. --multiple --style=goZero

# Passenger 服务
cd passenger/
goctl rpc protoc passenger.proto --go_out=. --go-grpc_out=. --zrpc_out=. --multiple --style=goZero

# Gateway（生成 handler + types + routes）
cd gateway/
goctl api go -api gateway.api -dir . --style=goZero
```

### Docker 部署

```bash
cd deploy/
docker-compose up -d
```

启动顺序：etcd → MySQL → Redis → driver → passenger → gateway

---

## 如何添加新服务

以添加 `trip`（行程）服务为例，完整步骤如下：

### 第一步：创建 gRPC 微服务

#### 1. 创建目录和 proto

```
onlinecar/
├── trip/                    ← 新服务
│   ├── trip.proto
│   ├── trip.go
│   ├── go.mod
│   ├── etc/trip.yaml
│   ├── common/
│   └── internal/
│       ├── config/
│       ├── svc/
│       ├── server/trip/
│       └── logic/trip/
```

#### 2. 编写 proto 文件

```protobuf
// trip/trip.proto
syntax = "proto3";

package trip;
option go_package = "./pb";

service Trip {
  rpc CreateTrip(CreateTripReq) returns(TripResp);
  rpc CancelTrip(CancelTripReq) returns(Empty);
  rpc GetTrip(GetTripReq) returns(TripResp);
  rpc UpdateTripStatus(UpdateTripStatusReq) returns(TripResp);
}

message Empty {}

message CreateTripReq {
  int64  passenger_id = 1;
  string origin       = 2;
  string destination  = 3;
  double origin_lng   = 4;
  double origin_lat   = 5;
  double dest_lng     = 6;
  double dest_lat     = 7;
}

message CancelTripReq { int64 trip_id = 1; }
message GetTripReq { int64 trip_id = 1; }

message UpdateTripStatusReq {
  int64  trip_id   = 1;
  int64  driver_id = 2;
  int64  status    = 3;
  double fare      = 4;
}

message TripResp {
  int64  id           = 1;
  int64  passenger_id = 2;
  int64  driver_id    = 3;
  string origin       = 4;
  string destination  = 5;
  int64  status       = 6;
  double fare         = 7;
}
```

#### 3. goctl 生成代码

```bash
cd trip/
go mod init onlinecar-trip

goctl rpc protoc trip.proto \
  --go_out=. --go-grpc_out=. --zrpc_out=. \
  --multiple --style=goZero

go mod tidy
```

#### 4. 编写配置

```yaml
# trip/etc/trip.yaml
Name: trip.rpc
ListenOn: 0.0.0.0:9003
Etcd:
  Hosts:
    - 10.211.55.5:2379
  Key: trip.rpc
DataSource: "root:123456@tcp(10.211.55.5:3306)/onlinecar_passenger?charset=utf8mb4&parseTime=true&loc=Local"
RedisConf:
  Host: 10.211.55.5:6379
  Type: node
```

#### 5. 实现 Logic 层

```go
// trip/internal/logic/trip/createTripLogic.go
func (l *CreateTripLogic) CreateTrip(in *pb.CreateTripReq) (*pb.TripResp, error) {
    // 写 trip 表，返回行程信息
}
```

### 第二步：Gateway 接入新服务

#### 1. 拷贝 proto 生成文件

```bash
# 从 trip/pb/ 拷贝到 gateway/pb/trip/，改 package 名为 trip
mkdir -p gateway/pb/trip/
cp trip/pb/trip.pb.go gateway/pb/trip/
cp trip/pb/trip_grpc.pb.go gateway/pb/trip/
# 编辑 gateway/pb/trip/trip_grpc.pb.go，将 package 改为 trip
```

#### 2. Gateway 配置加 RPC 客户端

```yaml
# gateway/etc/gateway.yaml — 新增
TripRpc:
  Etcd:
    Hosts:
      - 10.211.55.5:2379
    Key: trip.rpc
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
    TripRpc       zrpc.RpcClientConf     // ← 新增
}
```

#### 4. ServiceContext 初始化客户端

```go
// gateway/internal/svc/servicecontext.go
import trippb "onlinecar-gateway/pb/trip"

type ServiceContext struct {
    Config       config.Config
    DriverRpc    driverpb.DriverClient
    PassengerRpc passengerpb.PassengerClient
    TripRpc      trippb.TripClient         // ← 新增
}

func NewServiceContext(c config.Config) *ServiceContext {
    driverConn := zrpc.MustNewClient(c.DriverRpc)
    passengerConn := zrpc.MustNewClient(c.PassengerRpc)
    tripConn := zrpc.MustNewClient(c.TripRpc)    // ← 新增

    return &ServiceContext{
        Config:       c,
        DriverRpc:    driverpb.NewDriverClient(driverConn.Conn()),
        PassengerRpc: passengerpb.NewPassengerClient(passengerConn.Conn()),
        TripRpc:      trippb.NewTripClient(tripConn.Conn()),   // ← 新增
    }
}
```

#### 5. 注册 HTTP 路由

```go
// gateway/internal/handler/routes.go
// 新增行程相关路由
server.AddRoutes([]rest.Route{
    {Method: http.MethodPost, Path: "/api/v1/trip", Handler: passengerAuth.Handle(triphandler.CreateTripHandler(svcCtx))},
    {Method: http.MethodPost, Path: "/api/v1/trip/cancel", Handler: passengerAuth.Handle(triphandler.CancelTripHandler(svcCtx))},
    {Method: http.MethodGet,  Path: "/api/v1/trip/:id", Handler: triphandler.GetTripHandler(svcCtx)},
})

server.AddRoutes([]rest.Route{
    {Method: http.MethodPost, Path: "/api/v1/trip/status", Handler: driverAuth.Handle(triphandler.UpdateTripStatusHandler(svcCtx))},
})
```

#### 6. 实现 Gateway Logic

```go
// gateway/internal/logic/trip/createTripLogic.go
func (l *CreateTripLogic) CreateTrip(req *types.CreateTripReq) (*types.TripResp, error) {
    uid := ctxdata.GetUID(l.ctx)
    resp, err := l.svcCtx.TripRpc.CreateTrip(l.ctx, &tripbp.CreateTripReq{
        PassengerId: uid,
        Origin:      req.Origin,
        Destination: req.Destination,
        OriginLng:   req.OriginLng,
        OriginLat:   req.OriginLat,
        DestLng:     req.DestLng,
        DestLat:     req.DestLat,
    })
    if err != nil {
        return nil, err
    }
    return &types.TripResp{
        Id: resp.Id, PassengerId: resp.PassengerId,
        DriverId: resp.DriverId, Origin: resp.Origin,
        Destination: resp.Destination, Status: resp.Status,
    }, nil
}
```

### 第三步：Docker 部署

#### 1. 新增 Dockerfile

```dockerfile
# deploy/trip/Dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY trip/go.mod trip/go.sum ./
RUN go mod download
COPY trip/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o /trip trip.go

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai
WORKDIR /app
COPY --from=builder /trip .
COPY --from=builder /app/etc/trip.yaml ./etc/trip.yaml
EXPOSE 9003
CMD ["/trip", "-f", "etc/trip.yaml"]
```

#### 2. docker-compose 加服务

```yaml
# deploy/docker-compose.yaml — 新增
  trip:
    build:
      context: ../trip
      dockerfile: ../deploy/trip/Dockerfile
    ports:
      - "9003:9003"
    depends_on:
      - etcd
      - mysql
      - redis
    networks:
      - onlinecar
    restart: on-failure

  # gateway 的 depends_on 加上 trip
  gateway:
    depends_on:
      - etcd
      - driver
      - passenger
      - trip       # ← 新增
```

### 添加新服务检查清单

| # | 位置 | 操作 |
|---|------|------|
| 1 | `trip/trip.proto` | 定义 gRPC 服务和消息 |
| 2 | `trip/` | goctl 生成代码 + 实现 Logic |
| 3 | `trip/etc/trip.yaml` | 配置 etcd key + MySQL + Redis |
| 4 | `gateway/pb/trip/` | 拷贝 proto 生成文件，改 package 名 |
| 5 | `gateway/etc/gateway.yaml` | 加 RPC 客户端配置 |
| 6 | `gateway/internal/config/config.go` | Config 结构体加字段 |
| 7 | `gateway/internal/svc/servicecontext.go` | 初始化 gRPC 客户端 |
| 8 | `gateway/internal/handler/routes.go` | 注册 HTTP 路由 |
| 9 | `gateway/internal/logic/trip/` | 实现 HTTP → gRPC 转发逻辑 |
| 10 | `deploy/trip/Dockerfile` | 编写 Dockerfile |
| 11 | `deploy/docker-compose.yaml` | 添加服务定义 |

---

## 集群部署

### 当前架构天然支持集群

go-zero 的 etcd + zrpc 架构内置负载均衡能力：
- 多实例注册同一 etcd key，客户端自动发现
- go-zero 内置 **P2C (Power-of-Two-Choices)** 负载均衡
- 无需额外配置即可水平扩展

### gRPC 服务扩容（driver/passenger）

```yaml
# docker-compose.yaml
driver:
  deploy:
    replicas: 3    # 跑 3 个实例
```

多实例注册同一 `driver.rpc` key，Gateway 自动负载分发。

### Gateway 扩容 + Nginx

Gateway 无状态，前面套 Nginx 做负载均衡：

```nginx
upstream gateway {
    server gateway_1:8888;
    server gateway_2:8888;
    server gateway_3:8888;
}

server {
    listen 80;
    location / {
        proxy_pass http://gateway;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### etcd 高可用（生产环境）

```yaml
etcd:
  image: bitnami/etcd:3.5
  environment:
    - ALLOW_NONE_AUTHENTICATION=yes
    - ETCD_ADVERTISE_CLIENT_URLS=http://etcd:2379
  deploy:
    replicas: 3    # 生产建议 3 或 5 节点
```

---

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
