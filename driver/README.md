# OnlineCar Driver Service

> 网约车平台 — 司机 gRPC 微服务

## 概述

司机服务负责司机用户的全生命周期管理：注册、登录、位置上报、工作状态管理。基于 go-zero zRPC 构建，gRPC 协议对外服务，etcd 注册发现。

## 技术栈

| 组件 | 版本 | 说明 |
|------|------|------|
| Go | 1.26.1 | 开发语言 |
| go-zero | v1.10.1 | 微服务框架 |
| gRPC | v1.81.1 | RPC 通信协议 |
| etcd | v3.6.9 | 服务注册与发现 |
| MySQL | 8.4 | 持久化存储（`onlinecar_driver` 库） |
| Redis | 8.6.2 | Geo 位置存储 / 缓存 |

## 项目结构

```
driver/
├── driver.go                          # 程序入口（含 gRPC 参数配置）
├── driver.proto                       # gRPC 服务定义
├── go.mod / go.sum
├── etc/
│   └── driver.yaml                    # 运行时配置
├── pb/
│   ├── driver.pb.go                   # Protobuf 消息
│   └── driver_grpc.pb.go              # gRPC Stub
├── common/
│   ├── ctxdata/ctxdata.go             # Context UID 传递
│   └── errorx/errors.go               # 统一错误码
└── internal/
    ├── config/config.go               # 配置结构体
    ├── svc/servicecontext.go          # 服务上下文（DB + Redis + Model）
    ├── server/driver/driverServer.go  # gRPC Server
    ├── model/                         # 数据模型（GORM 风格）
    │   ├── driverModel.go
    │   ├── carModel.go
    │   ├── driverCarBindingModel.go
    │   └── driverWorkStatusModel.go
    └── logic/driver/
        ├── registerLogic.go           # 司机注册（事务）
        ├── loginLogic.go              # 司机登录
        ├── checkPhoneLogic.go         # 检查手机号
        ├── getDriverLogic.go          # 查询司机详情（含车辆+状态）
        ├── getDriverWorkStatusLogic.go
        ├── updateDriverWorkStatusLogic.go
        └── updateLocationLogic.go     # 位置上报（Redis Geo + 24h TTL）
```

## gRPC 接口

| 方法 | 请求 | 响应 | 说明 |
|------|------|------|------|
| `Register` | `RegisterReq` | `RegisterResp` | 司机注册（事务：创建用户+车辆+绑定+工作状态） |
| `Login` | `LoginReq` | `LoginResp` | 司机登录（手机号+密码） |
| `CheckPhone` | `CheckPhoneReq` | `CheckPhoneResp` | 手机号是否已注册 |
| `GetDriver` | `GetDriverReq` | `DriverInfo` | 查询详情（含车辆、工作状态） |
| `GetDriverWorkStatus` | `GetDriverWorkStatusReq` | `DriverWorkStatusInfo` | 查询工作状态 |
| `UpdateDriverWorkStatus` | `UpdateDriverWorkStatusReq` | `Empty` | 切换状态（0收车/1出车/2暂停） |
| `UpdateLocation` | `UpdateLocationReq` | `Empty` | 实时位置上报 |

## 数据库

库：`onlinecar_driver`，4 张表：

| 表 | 说明 |
|----|------|
| `driver` | 司机基本信息（phone 唯一索引） |
| `car` | 车辆信息（车牌、品牌、型号） |
| `driver_car_binding_relationship` | 司机-车辆绑定关系 |
| `driver_work_status` | 司机工作状态（收车/出车/暂停） |

## 配置

```yaml
# etc/driver.yaml
Name: driver.rpc
ListenOn: 0.0.0.0:6000
Timeout: 30000

Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: driver.rpc

DataSource: "root:password@tcp(127.0.0.1:3306)/onlinecar_driver?charset=utf8mb4&parseTime=true&loc=Local"

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
  ├── 注册到 etcd (key: driver.rpc)
  └── 启动监听 :6000
```

## 快速开始

```bash
# 代码生成
goctl rpc protoc driver.proto --go_out=. --go-grpc_out=. --zrpc_out=. --multiple --style=goZero

# 本地运行
go run driver.go -f etc/driver.yaml

# Docker
docker-compose up -d driver
```

## 关键设计

- **事务注册**：`Register` 使用 `TransactCtx` 将司机/车辆/绑定/状态 4 张表的插入放在一个事务中，防并发竞态
- **位置上报**：Redis `GEOADD` + `EXPIRE` 原子操作，key `driver:location`，24h TTL 自动清理不活跃司机
- **错误脱敏**：所有 DB 查询失败返回固定文案 `"查询失败，请稍后重试"`，原始 err 仅记录日志
- **CheckPhone 不降级**：DB 异常直接返回 error，不会静默返回 `exists=false`

## 服务发现

- 启动后注册到 etcd，key: `driver.rpc`
- Gateway 通过 etcd 发现并负载均衡（P2C 策略）
- 支持多实例水平扩展
