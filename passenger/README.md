# OnlineCar Passenger Service

> 网约车平台 — 乘客 gRPC 微服务

## 概述

乘客服务负责乘客用户管理和行程创建。基于 go-zero zRPC 构建，gRPC 协议对外服务，etcd 注册发现。

## 技术栈

| 组件 | 版本 | 说明 |
|------|------|------|
| Go | 1.26.1 | 开发语言 |
| go-zero | v1.10.1 | 微服务框架 |
| gRPC | v1.81.1 | RPC 通信协议 |
| etcd | v3.6.9 | 服务注册与发现 |
| MySQL | 8.4 | 持久化存储（`onlinecar_passenger` 库） |
| Redis | 8.6.2 | 缓存 |

## 项目结构

```
passenger/
├── passenger.go                      # 程序入口（含 gRPC 参数配置）
├── passenger.proto                   # gRPC 服务定义
├── go.mod / go.sum
├── etc/
│   └── passenger.yaml                # 运行时配置
├── pb/
│   ├── passenger.pb.go               # Protobuf 消息
│   └── passenger_grpc.pb.go          # gRPC Stub
├── common/
│   ├── ctxdata/ctxdata.go            # Context UID 传递
│   └── errorx/errors.go              # 统一错误码
└── internal/
    ├── config/config.go              # 配置结构体
    ├── svc/servicecontext.go         # 服务上下文（DB + Redis + Model）
    ├── server/passenger/passengerServer.go  # gRPC Server
    ├── model/                        # 数据模型
    │   ├── passengerUserModel.go
    │   └── tripModel.go
    └── logic/passenger/
        ├── registerLogic.go          # 乘客注册（事务）
        ├── loginLogic.go             # 乘客登录
        ├── checkPhoneLogic.go        # 检查手机号
        ├── getPassengerLogic.go      # 查询乘客详情
        └── createTripLogic.go        # 创建行程（叫车）
```

## gRPC 接口

| 方法 | 请求 | 响应 | 说明 |
|------|------|------|------|
| `Register` | `RegisterReq` | `RegisterResp` | 乘客注册（姓名/手机号/密码） |
| `Login` | `LoginReq` | `LoginResp` | 乘客登录（手机号+密码） |
| `CheckPhone` | `CheckPhoneReq` | `CheckPhoneResp` | 手机号是否已注册 |
| `GetPassenger` | `GetPassengerReq` | `PassengerInfo` | 查询乘客详情 |
| `CreateTrip` | `CreateTripReq` | `TripResp` | 创建行程（叫车） |

## 数据库

库：`onlinecar_passenger`，2 张表：

| 表 | 说明 |
|----|------|
| `passenger_user` | 乘客用户（phone 字段） |
| `trip` | 行程（状态 1待接单 2已接单 3行程中 4已完成 5已取消，DEFAULT 1） |

## 配置

```yaml
# etc/passenger.yaml
Name: passenger.rpc
ListenOn: 0.0.0.0:6001
Timeout: 30000

Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: passenger.rpc

DataSource: "root:password@tcp(127.0.0.1:3306)/onlinecar_passenger?charset=utf8mb4&parseTime=true&loc=Local"

RedisConf:
  Host: 127.0.0.1:6379
  Type: node
```

## 启动流程

```
main()
  ├── 加载配置
  ├── 创建 ServiceContext（MySQL + Redis + Models）
  ├── 注册 gRPC Server
  ├── 配置 MaxRecvMsgSize(4MB) + MaxSendMsgSize(4MB)
  ├── 配置 Keepalive 参数
  ├── 注册 UnaryInterceptor（日志+耗时+错误码）
  ├── 注册 gRPC Health Server
  ├── 注册到 etcd (key: passenger.rpc)
  └── 启动监听 :6001
```

## 快速开始

```bash
# 代码生成
goctl rpc protoc passenger.proto --go_out=. --go-grpc_out=. --zrpc_out=. --multiple --style=goZero

# 本地运行
go run passenger.go -f etc/passenger.yaml

# Docker
docker-compose up -d passenger
```

## 关键设计

- **事务注册**：`Register` 使用 `TransactCtx` 确保用户创建的原子性
- **错误脱敏**：所有 DB 查询失败返回固定文案，原始 err 仅记录日志
- **CheckPhone 不降级**：DB 异常直接返回 error，不会静默返回 `exists=false`
- **Trip 状态码**：`1待接单 2已接单 3行程中 4已完成 5已取消`，DDL/Model/Logic 三处统一

## 服务发现

- 启动后注册到 etcd，key: `passenger.rpc`
- Gateway 通过 etcd 发现并负载均衡（P2C 策略）
- 支持多实例水平扩展
