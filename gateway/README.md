# OnlineCar Gateway

> 网约车平台 — HTTP API 网关

## 概述

Gateway 是平台的统一 HTTP 入口（BFF），负责 JWT 认证、限流、安全头、CORS，并将请求转发到后端 gRPC 微服务。基于 go-zero REST 框架构建。

## 技术栈

| 组件 | 版本 | 说明 |
|------|------|------|
| Go | 1.26.1 | 开发语言 |
| go-zero | v1.10.1 | REST 框架 |
| gRPC | v1.81.1 | 后端 RPC 通信 |
| etcd | v3.6.9 | 服务发现 |
| golang-jwt | v4 | JWT Token 签发与验证 |
| Redis | 8.6.2 | 验证码存储 / 频率限制 |

## 项目结构

```
gateway/
├── gateway.go                         # 程序入口
├── gateway.api                        # go-zero API 定义文件
├── go.mod / go.sum
├── etc/
│   └── gateway.yaml                   # 运行时配置
├── common/
│   ├── ctxdata/ctxdata.go             # Context UID 传递工具
│   ├── errorx/errors.go               # 统一错误码 + MapRpcError
│   ├── response/response.go           # 统一 JSON 响应格式
│   ├── validate/validate.go           # 参数校验（手机号等）
│   └── sms/sms.go                     # 验证码脱敏日志
├── pb/                                # 后端 proto 生成文件（拷贝+改 package）
│   ├── driver/
│   │   ├── driver.pb.go
│   │   └── driver_grpc.pb.go
│   └── passenger/
│       ├── passenger.pb.go
│       └── passenger_grpc.pb.go
└── internal/
    ├── config/config.go               # 配置结构体（含 CORS）
    ├── svc/servicecontext.go          # 服务上下文（gRPC 客户端 + Redis）
    ├── handler/
    │   ├── routes.go                  # 路由注册（中间件链 + 公开/认证路由）
    │   ├── driver/driver.go           # 司机 HTTP Handler
    │   ├── passenger/passenger.go     # 乘客 HTTP Handler
    │   ├── trip/trip.go              # 行程 HTTP Handler
    │   └── sms/sms.go               # 短信 HTTP Handler
    ├── logic/
    │   ├── driver/driver.go           # 司机逻辑 + JWT 签发
    │   ├── passenger/passenger.go     # 乘客逻辑 + JWT 签发
    │   ├── trip/trip.go              # 行程创建逻辑
    │   └── sms/sms.go               # 验证码发送+校验（Lua 原子操作）
    ├── middleware/
    │   ├── auth.go                    # JWT 双角色认证中间件
    │   ├── cors.go                    # CORS 跨域中间件
    │   ├── ratelimit.go              # IP 限流中间件（可信代理）
    │   └── security.go              # 安全响应头中间件
    └── types/types.go                 # 请求/响应类型
```

## 架构

```
HTTP Request
    │
    ▼
┌─────────────────────────────────────────────┐
│  Global Middleware                           │
│  CORS → Security Headers → Log → Panic Recovery  │
├─────────────────────────────────────────────┤
│  Route Matching                             │
│  ├── 公开路由 → 限流(20/s) → Handler       │
│  ├── 司机路由 → 限流(100/s) → JWT(Driver) → Handler  │
│  └── 乘客路由 → 限流(50/s) → JWT(Passenger) → Handler │
├─────────────────────────────────────────────┤
│  Handler → Logic → gRPC Client → etcd → RPC │
│  错误处理：MapRpcError 统一映射              │
└─────────────────────────────────────────────┘
```

## 中间件链

| 中间件 | 作用 | 备注 |
|--------|------|------|
| `CORSMiddleware` | 跨域控制 | 从配置读取 AllowOrigins |
| `SecurityHeadersMiddleware` | 安全响应头 | HSTS/X-Frame-Options/X-Content-Type-Options/Referrer-Policy |
| `LogMiddleware` | 访问日志 | 记录 method/path/status/duration |
| `PanicMiddleware` | Panic 恢复 | 防止单个请求 panic 导致服务崩溃 |
| `RateLimitMiddleware` | IP 限流 | 仅信任内网代理的 X-Real-IP |
| `JWTAuthMiddleware` | JWT 认证 | 司机/乘客独立密钥，支持环境变量覆盖 |

## API 接口

| 方法 | 路径 | 认证 | 限流 | 说明 |
|------|------|------|------|------|
| POST | `/api/v1/driver/register` | 无 | 20/s | 司机注册 |
| POST | `/api/v1/driver/login` | 无 | 20/s | 司机登录 |
| POST | `/api/v1/driver/location` | Driver JWT | 100/s | 上报位置 |
| POST | `/api/v1/passenger/register` | 无 | 20/s | 乘客注册 |
| POST | `/api/v1/passenger/login` | 无 | 20/s | 乘客登录 |
| POST | `/api/v1/sms/send` | 无 | 20/s | 发送验证码 |
| POST | `/api/v1/trip` | Passenger JWT | 50/s | 发起行程 |
| GET | `/healthz` | 无 | — | 健康检查 |

## JWT 双角色认证

司机和乘客使用独立的 JWT 密钥：

```yaml
AuthDriver:
  AccessSecret: "onlinecar-driver-secret-2026"
  AccessExpire: 86400s

AuthPassenger:
  AccessSecret: "onlinecar-passenger-secret-2026"
  AccessExpire: 86400s
```

密钥支持环境变量覆盖：`ONLINECAR_DRIVER_JWT_SECRET` / `ONLINECAR_PASSENGER_JWT_SECRET`

JWT Claims：`uid`（用户ID）、`subject`（driver/passenger）

## 验证码机制

- Redis key: `sms:{role}:{phone}`，5 分钟 TTL
- 频率限制: `sms:limit:{phone}`，60 秒 TTL
- 校验使用 **Lua 脚本**原子化 GET + DEL，防并发重放
- 频率限制 Redis 写入失败时回滚验证码，阻断发送

## 统一响应格式

```json
{
  "code": 0,
  "msg": "ok",
  "data": { ... }
}
```

## 错误处理

### RPC 错误映射（MapRpcError）

跨 gRPC 传输后 CodeError 类型丢失为普通 error，通过 `errorx.MapRpcError` 关键词匹配恢复：

```go
// 所有 Gateway Logic 中 RPC 调用必须使用 MapRpcError
resp, err := l.svcCtx.DriverRpc.Login(l.ctx, req)
if err != nil {
    return nil, errorx.MapRpcError(err)  // 不直接返回 err
}
```

### 内部错误脱敏

- RPC 层：`fmt.Errorf("查询失败，请稍后重试")`（不带原始 err）
- Gateway：非 CodeError 的 err 返回 `"服务器内部错误，请稍后重试"`（500）

## 配置

```yaml
# etc/gateway.yaml
Name: gateway-api
Host: 0.0.0.0
Port: 7000
Timeout: 30000

AuthDriver:
  AccessSecret: "onlinecar-driver-secret-2026"
  AccessExpire: 86400s
AuthPassenger:
  AccessSecret: "onlinecar-passenger-secret-2026"
  AccessExpire: 86400s

DriverRpc:
  Etcd:
    Hosts: [127.0.0.1:2379]
    Key: driver.rpc
PassengerRpc:
  Etcd:
    Hosts: [127.0.0.1:2379]
    Key: passenger.rpc

RedisConf:
  Host: 127.0.0.1:6379
  Type: node

CORS:
  AllowOrigins: ["*"]
```

## 快速开始

```bash
# 本地运行（确保 etcd + 后端 RPC 已启动）
go run gateway.go -f etc/gateway.yaml

# Docker
docker-compose up -d gateway
```

## 如何接入新的 RPC 服务

添加新微服务时，Gateway 需要 6 步改造：

### 1. 拷贝 proto 生成文件

```bash
mkdir -p gateway/pb/{service}/
cp {service}/pb/{service}.pb.go gateway/pb/{service}/
cp {service}/pb/{service}_grpc.pb.go gateway/pb/{service}/
# 修改 package 声明为 package {service}
```

### 2. 配置加 RPC 客户端

```yaml
# gateway/etc/gateway.yaml
{Service}Rpc:
  Etcd:
    Hosts: [127.0.0.1:2379]
    Key: {service}.rpc
```

### 3. Config 结构体加字段

```go
// gateway/internal/config/config.go
type Config struct {
    // ...
    {Service}Rpc zrpc.RpcClientConf
}
```

### 4. ServiceContext 初始化客户端

```go
// gateway/internal/svc/servicecontext.go
type ServiceContext struct {
    // ...
    {Service}Rpc {service}pb.{Service}Client
}

func NewServiceContext(c config.Config) *ServiceContext {
    conn := zrpc.MustNewClient(c.{Service}Rpc)
    return &ServiceContext{
        // ...
        {Service}Rpc: {service}pb.New{Service}Client(conn.Conn()),
    }
}
```

### 5. 注册路由

```go
// gateway/internal/handler/routes.go
server.AddRoutes([]rest.Route{
    {Method: http.MethodPost, Path: "/api/v1/...",
        Handler: {rateLimit}({auth}(handler.{Handler}(svcCtx)))},
})
```

### 6. 实现 Logic（用 MapRpcError）

```go
resp, err := l.svcCtx.{Service}Rpc.{Method}(l.ctx, req)
if err != nil {
    l.Errorf("{Service}Rpc.{Method} 失败, err=%v", err)
    return nil, errorx.MapRpcError(err)
}
```

## 集群与扩展

- Gateway 无状态，支持多副本 + Nginx 负载均衡
- 后端 RPC 通过 etcd 自动发现，go-zero P2C 负载均衡
- 详见根目录 README
