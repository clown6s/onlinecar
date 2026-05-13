# OnlineCar Gateway

> 网约车平台 — HTTP API 网关

## 概述

Gateway 是 OnlineCar 网约车平台的统一入口（BFF），负责将外部 HTTP 请求转发到后端 gRPC 微服务。基于 go-zero REST 框架构建，提供 RESTful API 接口、JWT 身份认证、请求路由转发等功能。

Gateway 作为唯一的对外暴露层，所有客户端（App、小程序、Web）都通过 Gateway 访问平台功能。

## 技术栈

| 组件 | 版本 | 说明 |
|------|------|------|
| Go | 1.25.0 | 开发语言 |
| go-zero | v1.10.1 | REST 框架 |
| gRPC | v1.81.0 | 后端 RPC 通信 |
| etcd | 3.5 | 服务发现 |
| golang-jwt | v4.5.2 | JWT Token 签发与验证 |

## 项目结构

```
gateway/
├── gateway.go                         # 程序入口
├── gateway.api                        # go-zero API 定义文件（路由 + 类型）
├── go.mod / go.sum                    # Go 模块依赖
├── etc/
│   └── gateway.yaml                   # 运行时配置
├── common/
│   ├── ctxdata/
│   │   └── ctxdata.go                 # Context UID 传递工具
│   ├── errorx/
│   │   └── errors.go                  # 统一错误码定义
│   └── response/
│       └── response.go                # 统一 JSON 响应格式
├── pb/
│   ├── driver/
│   │   ├── driver.pb.go               # Driver Protobuf 消息
│   │   └── driver_grpc.pb.go          # Driver gRPC Client Stub
│   └── passenger/
│       ├── passenger.pb.go            # Passenger Protobuf 消息
│       └── passenger_grpc.pb.go       # Passenger gRPC Client Stub
└── internal/
    ├── config/
    │   └── config.go                  # 配置结构体
    ├── handler/
    │   ├── routes.go                  # 路由注册（公开 + JWT 鉴权）
    │   ├── driver/
    │   │   └── driver.go              # 司机 HTTP Handler
    │   ├── passenger/
    │   │   └── passenger.go           # 乘客 HTTP Handler
    │   └── trip/
    │       └── trip.go                # 行程 HTTP Handler
    ├── logic/
    │   ├── driver/
    │   │   └── driver.go              # 司机业务逻辑 + JWT 签发
    │   ├── passenger/
    │   │   └── passenger.go           # 乘客业务逻辑 + JWT 签发
    │   └── trip/
    │       └── trip.go                # 行程创建逻辑
    ├── middleware/
    │   └── auth.go                    # 自定义 JWT 鉴权中间件
    ├── svc/
    │   └── servicecontext.go          # 服务上下文（gRPC 客户端注入）
    └── types/
        └── types.go                   # 请求/响应类型定义
```

## 架构设计

```
┌──────────────────────────────────────────────────────────┐
│                     Client (App / Web)                    │
└──────────────────────────┬───────────────────────────────┘
                           │ HTTP :8888
                           ▼
┌──────────────────────────────────────────────────────────┐
│                      Gateway (REST)                       │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────┐  │
│  │  JWT Auth    │  │  Route Match │  │  Response      │  │
│  │  Middleware  │  │              │  │  Wrapper       │  │
│  └─────────────┘  └──────┬───────┘  └───────────────┘  │
│                          │ gRPC                          │
│              ┌───────────┴───────────┐                   │
│              │     etcd Discovery    │                   │
│              └───────────┬───────────┘                   │
└──────────────────────────┼──────────────────────────────┘
                           │
          ┌────────────────┴────────────────┐
          ▼                                 ▼
┌──────────────────┐              ┌──────────────────┐
│  Driver Service   │              │ Passenger Service │
│  gRPC :9001       │              │ gRPC :9002        │
│  (driver.rpc)     │              │ (passenger.rpc)   │
└──────────────────┘              └──────────────────┘
```

## API 接口列表

### 公开接口（无需认证）

| 方法 | 路径 | 说明 | Handler |
|------|------|------|---------|
| POST | `/api/v1/driver/register` | 司机注册 | `DriverRegister` |
| POST | `/api/v1/driver/login` | 司机登录（返回 JWT） | `DriverLogin` |
| POST | `/api/v1/passenger/register` | 乘客注册 | `PassengerRegister` |
| POST | `/api/v1/passenger/login` | 乘客登录（返回 JWT） | `PassengerLogin` |

### 司机认证接口（需要 Driver JWT）

| 方法 | 路径 | 说明 | Handler |
|------|------|------|---------|
| POST | `/api/v1/driver/location` | 上报实时位置 | `UpdateDriverLocation` |

### 乘客认证接口（需要 Passenger JWT）

| 方法 | 路径 | 说明 | Handler |
|------|------|------|---------|
| POST | `/api/v1/trip` | 发起行程（叫车） | `CreateTrip` |

## JWT 认证机制

Gateway 实现了**双域名隔离的 JWT 认证**，司机和乘客使用不同的密钥：

### 配置

```yaml
AuthDriver:
  AccessSecret: "onlinecar-driver-secret-2026"
  AccessExpire: 86400s                # 24 小时

AuthPassenger:
  AccessSecret: "onlinecar-passenger-secret-2026"
  AccessExpire: 86400s                # 24 小时
```

### JWT Claims

```json
{
  "uid": 12345,           // 用户 ID
  "subject": "driver",    // 身份标识: "driver" 或 "passenger"
  "exp": 1700000000,      // 过期时间
  "iat": 1699913600       // 签发时间
}
```

### 认证流程

```
1. 登录 → 后端 RPC 验证手机号 → 签发 JWT Token
2. 请求受保护接口 → Header 携带 Authorization: Bearer <token>
3. JWT Middleware → 验证签名 + 过期时间 → 提取 uid 注入 Context
4. Handler → 从 Context 获取 uid → 传递给后端 RPC
```

### 自定义中间件

```go
// internal/middleware/auth.go
type JWTAuthMiddleware struct {
    Secret string
}

func (m *JWTAuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
    // 1. 提取 Authorization: Bearer <token>
    // 2. 解析验证 JWT
    // 3. 提取 uid claim
    // 4. 通过 ctxdata.SetUID 注入 Context
    // 5. 调用 next handler
}
```

## 请求/响应格式

### 统一响应结构

所有接口返回统一的 JSON 格式：

```json
{
  "code": 0,
  "msg": "ok",
  "data": { ... }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | int | 状态码，0 表示成功，-1 表示业务错误 |
| `msg` | string | 状态描述 |
| `data` | any | 响应数据（可选） |

### 请求类型

#### 司机注册

```json
POST /api/v1/driver/register

{
  "name": "张师傅",
  "phone": "13800138000",
  "carNo": "京A12345",
  "carType": "舒适型",
  "password": "123456"
}
```

#### 司机登录

```json
POST /api/v1/driver/login

{
  "phone": "13800138000",
  "password": "123456"
}
```

**响应：**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expire": 86400,
    "driverId": 1,
    "driverName": "张师傅"
  }
}
```

#### 乘客注册

```json
POST /api/v1/passenger/register

{
  "name": "李先生",
  "phone": "13900139000",
  "password": "123456"
}
```

#### 乘客登录

```json
POST /api/v1/passenger/login

{
  "phone": "13900139000",
  "password": "123456"
}
```

**响应：**
```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "expire": 86400,
    "passengerId": 1,
    "passengerName": "李先生"
  }
}
```

#### 上报位置（需 Driver JWT）

```json
POST /api/v1/driver/location
Authorization: Bearer <driver_token>

{
  "lng": 116.397128,
  "lat": 39.916527
}
```

#### 创建行程（需 Passenger JWT）

```json
POST /api/v1/trip
Authorization: Bearer <passenger_token>

{
  "origin": "中关村软件园",
  "destination": "望京SOHO",
  "originLng": 116.310,
  "originLat": 40.055,
  "destLng": 116.480,
  "destLat": 39.998
}
```

## 配置说明

### `etc/gateway.yaml`

```yaml
Name: gateway-api                    # 网关服务名
Host: 0.0.0.0                        # 监听地址
Port: 8888                           # 监听端口

# JWT 配置（司机）
AuthDriver:
  AccessSecret: "onlinecar-driver-secret-2026"
  AccessExpire: 86400s

# JWT 配置（乘客）
AuthPassenger:
  AccessSecret: "onlinecar-passenger-secret-2026"
  AccessExpire: 86400s

# Driver RPC 客户端（通过 etcd 发现）
DriverRpc:
  Etcd:
    Hosts:
      - 10.211.55.5:2379
    Key: driver.rpc

# Passenger RPC 客户端（通过 etcd 发现）
PassengerRpc:
  Etcd:
    Hosts:
      - 10.211.55.5:2379
    Key: passenger.rpc
```

### 配置结构体

```go
type JwtConf struct {
    AccessSecret string        `json:",default="`
    AccessExpire time.Duration `json:",default=86400s"`
}

type Config struct {
    rest.RestConf              // 内嵌：Host, Port
    AuthDriver    JwtConf      // 司机 JWT 配置
    AuthPassenger JwtConf      // 乘客 JWT 配置
    DriverRpc     zrpc.RpcClientConf    // Driver gRPC 客户端配置
    PassengerRpc  zrpc.RpcClientConf    // Passenger gRPC 客户端配置
}
```

## 服务上下文

```go
type ServiceContext struct {
    Config       config.Config
    DriverRpc    driverpb.DriverClient       // Driver gRPC 客户端
    PassengerRpc passengerpb.PassengerClient // Passenger gRPC 客户端
}
```

初始化时通过 `zrpc.MustNewClient()` 连接 etcd，发现后端 RPC 服务并创建 gRPC 客户端存根。

## gRPC 客户端连接

Gateway 不直接 import driver/passenger 服务源码，而是通过编译后的 protobuf 代码连接：

```
gateway/pb/driver/     ← 从 driver.proto 编译后拷贝，package 名改为 driver
gateway/pb/passenger/  ← 从 passenger.proto 编译后拷贝，package 名改为 passenger
```

通过 `zrpc.MustNewClient(c.DriverRpc)` 连接 etcd，查找 `driver.rpc` key 获取实际服务地址，使用 P2C 负载均衡策略。

## 启动流程

```
main()
  ├── 加载配置文件 etc/gateway.yaml
  ├── 禁用 logx 统计 (logx.DisableStat)
  ├── 创建 REST Server
  ├── 创建 ServiceContext
  │   ├── 连接 DriverRpc (etcd → driver.rpc)
  │   └── 连接 PassengerRpc (etcd → passenger.rpc)
  ├── 注册路由
  │   ├── 公开路由（/register, /login）
  │   ├── 司机鉴权路由（/driver/location → Driver JWT Middleware）
  │   └── 乘客鉴权路由（/trip → Passenger JWT Middleware）
  └── 启动监听 (0.0.0.0:8888)
```

## 快速开始

### 环境要求

- Go 1.25+
- etcd 3.5+
- 后端 Driver Service 和 Passenger Service 已启动

### 本地运行

```bash
# 确保 etcd 和后端 RPC 服务已启动
# 修改 etc/gateway.yaml 中的 etcd 地址
go run gateway.go -f etc/gateway.yaml
```

### Docker 部署

```bash
# 在项目根目录 deploy/ 下执行（包含全部服务）
docker-compose up -d gateway
```

### 接口测试示例

```bash
# 司机注册
curl -X POST http://localhost:8888/api/v1/driver/register \
  -H "Content-Type: application/json" \
  -d '{"name":"张师傅","phone":"13800138000","carNo":"京A12345","carType":"舒适型","password":"123456"}'

# 司机登录
curl -X POST http://localhost:8888/api/v1/driver/login \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800138000","password":"123456"}'

# 上报位置（替换 <token> 为登录返回的 token）
curl -X POST http://localhost:8888/api/v1/driver/location \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"lng":116.397,"lat":39.916}'
```

## 集群与扩展

### Gateway 水平扩展

Gateway 是无状态服务，可直接多副本部署：

**Docker Compose：**
```yaml
gateway:
  deploy:
    replicas: 3
  # 注意：多副本时需要去掉 ports 映射，前面套 Nginx
```

**Nginx 反向代理：**
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

### 后端 RPC 自动负载均衡

- Gateway 通过 etcd 发现多个 Driver/Passenger 实例
- go-zero 内置 P2C (Power-of-Two-Choices) 负载均衡策略
- 无需额外配置，多实例自动分发请求

## 错误码定义

| 错误 | Code | 说明 |
|------|------|------|
| 成功 | 0 | 请求成功 |
| 业务错误 | -1 | 后端返回的业务错误 |
| `ErrServer` | 500 | 服务器内部错误 |
| `ErrParam` | 400 | 参数错误 |
| `ErrNotFound` | 404 | 资源不存在 |
| `ErrAuth` | 401 | 未授权（缺少/无效 Token） |
| `ErrForbidden` | 403 | 无权限访问 |

## 依赖关系

```
gateway 服务
  ├── etcd           (发现 Driver RPC + Passenger RPC)
  ├── Driver RPC     (司机注册/登录/位置管理)
  └── Passenger RPC  (乘客注册/登录/行程创建)
```

## 待开发

- [ ] 更多 API 接口（司机信息查询、行程状态查询等）
- [ ] 接口限流（Rate Limiting）
- [ ] 请求日志追踪（OpenTelemetry / Trace ID）
- [ ] API 文档自动生成（Swagger）
- [ ] 健康检查端点（/health）
- [ ] 优雅关闭（Graceful Shutdown）
