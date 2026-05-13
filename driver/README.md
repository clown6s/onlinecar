# OnlineCar Driver Service

> 网约车平台 — 司机 gRPC 微服务

## 概述

司机服务是 OnlineCar 网约车平台的核心微服务之一，负责司机用户的全生命周期管理。基于 go-zero 框架的 zRPC 构建，使用 gRPC 协议对外提供服务，通过 etcd 进行服务注册与发现。

## 技术栈

| 组件 | 版本 | 说明 |
|------|------|------|
| Go | 1.25.0 | 开发语言 |
| go-zero | v1.10.1 | 微服务框架 |
| gRPC | v1.81.0 | RPC 通信协议 |
| etcd | 3.5 | 服务注册与发现 |
| MySQL | 8.0 | 持久化存储 |
| Redis | 7.x | 缓存（已配置，待接入） |

## 项目结构

```
driver/
├── driver.go                          # 程序入口
├── driver.proto                       # gRPC 服务定义（Protobuf）
├── go.mod / go.sum                    # Go 模块依赖
├── etc/
│   └── driver.yaml                    # 运行时配置
├── pb/
│   ├── driver.pb.go                   # Protobuf 生成消息
│   └── driver_grpc.pb.go              # gRPC 生成代码
├── common/
│   ├── ctxdata/
│   │   └── ctxdata.go                 # Context UID 传递工具
│   └── errorx/
│       └── errors.go                  # 统一错误码定义
└── internal/
    ├── config/
    │   └── config.go                  # 配置结构体
    ├── svc/
    │   └── servicecontext.go          # 服务上下文（依赖注入）
    ├── server/
    │   └── driver/
    │       └── driverServer.go        # gRPC Server（goctl 生成）
    └── logic/
        └── driver/
            ├── registerLogic.go       # 司机注册
            ├── loginLogic.go          # 司机登录
            ├── getDriverLogic.go      # 获取司机信息
            └── updateLocationLogic.go # 更新司机位置
```

## gRPC 接口定义

### 服务：`Driver`

| 方法 | 请求 | 响应 | 说明 |
|------|------|------|------|
| `Register` | `RegisterReq` | `RegisterResp` | 司机注册 |
| `Login` | `LoginReq` | `LoginResp` | 司机登录 |
| `UpdateLocation` | `UpdateLocationReq` | `Empty` | 实时位置上报 |
| `GetDriver` | `GetDriverReq` | `DriverInfo` | 查询司机详情 |

### Protobuf 消息定义

```protobuf
// 注册请求
message RegisterReq {
  string name     = 1;  // 司机姓名
  string phone    = 2;  // 手机号（唯一）
  string carNo    = 3;  // 车牌号
  string carType  = 4;  // 车型（经济型/舒适型/豪华型）
  string password = 5;  // 密码
}

// 注册响应
message RegisterResp {
  int64  id      = 1;  // 司机ID
  string name    = 2;  // 司机姓名
  string phone   = 3;  // 手机号
  string carNo   = 4;  // 车牌号
  string carType = 5;  // 车型
  int64  status  = 6;  // 状态
}

// 登录请求
message LoginReq {
  string phone    = 1;  // 手机号
  string password = 2;  // 密码
}

// 登录响应
message LoginResp {
  int64  id     = 1;  // 司机ID
  string name   = 2;  // 司机姓名
  int64  status = 3;  // 状态
}

// 位置更新请求
message UpdateLocationReq {
  int64  id  = 1;  // 司机ID
  double lng = 2;  // 经度
  double lat = 3;  // 纬度
}

// 查询司机请求
message GetDriverReq {
  int64 id = 1;  // 司机ID
}

// 司机详细信息
message DriverInfo {
  int64  id      = 1;  // 司机ID
  string name    = 2;  // 司机姓名
  string phone   = 3;  // 手机号
  string carNo   = 4;  // 车牌号
  string carType = 5;  // 车型
  double lng     = 6;  // 经度
  double lat     = 7;  // 纬度
  int64  status  = 8;  // 状态: 0=离线 1=空闲 2=接单中 3=行程中
}
```

## 数据库设计

### 数据库：`onlinecar_driver`

```sql
CREATE TABLE `driver` (
    `id`         BIGINT       NOT NULL AUTO_INCREMENT COMMENT '司机ID',
    `name`       VARCHAR(50)  NOT NULL DEFAULT '' COMMENT '姓名',
    `phone`      VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '手机号',
    `password`   VARCHAR(255) NOT NULL DEFAULT '' COMMENT '密码(加密)',
    `car_no`     VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '车牌号',
    `car_type`   VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '车型: 经济型/舒适型/豪华型',
    `lng`        DOUBLE       NOT NULL DEFAULT 0 COMMENT '经度',
    `lat`        DOUBLE       NOT NULL DEFAULT 0 COMMENT '纬度',
    `status`     TINYINT      NOT NULL DEFAULT 1 COMMENT '状态: 0=离线 1=空闲 2=接单中 3=行程中',
    `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_phone` (`phone`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='司机表';
```

**状态枚举：**

| 值 | 含义 |
|----|------|
| 0 | 离线 |
| 1 | 空闲（可接单） |
| 2 | 接单中 |
| 3 | 行程中 |

## 配置说明

### `etc/driver.yaml`

```yaml
Name: driver.rpc                    # 服务名称（etcd 注册 key）
ListenOn: 0.0.0.0:9001              # gRPC 监听地址

Etcd:                                # 服务发现
  Hosts:
    - 10.211.55.5:2379               # etcd 地址
  Key: driver.rpc                    # 注册键名

DataSource: "root:123456@tcp(10.211.55.5:3306)/onlinecar_driver?charset=utf8mb4&parseTime=true&loc=Local"  # MySQL DSN

RedisConf:                           # Redis 配置
  Host: 10.211.55.5:6379
  Type: node                         # 单节点模式
```

### 配置结构体

```go
type Config struct {
    zrpc.RpcServerConf          // 内嵌：Name, ListenOn, Etcd
    DataSource     string       // MySQL 连接串
    RedisConf      redis.RedisConf  // Redis 配置
}
```

## 服务上下文

```go
type ServiceContext struct {
    Config config.Config
    DB     sqlx.SqlConn        // MySQL 连接（go-zero sqlx）
}
```

启动时通过 `sqlx.NewMysql(c.DataSource)` 初始化数据库连接。

## 启动流程

```
main()
  ├── 加载配置文件 etc/driver.yaml
  ├── 禁用 logx 统计 (logx.DisableStat)
  ├── 创建 ServiceContext（初始化 MySQL）
  ├── 注册 gRPC Server (pb.RegisterDriverServer)
  ├── 注册到 etcd (key: driver.rpc)
  └── 启动监听 (0.0.0.0:9001)
```

## 快速开始

### 环境要求

- Go 1.25+
- MySQL 8.0+
- etcd 3.5+
- protoc + goctl（代码生成）

### 代码生成

```bash
# 在 driver/ 目录下执行
goctl rpc protoc driver.proto --go_out=. --go-grpc_out=. --zrpc_out=. --multiple --style=goZero
```

### 本地运行

```bash
# 确保 MySQL 和 etcd 已启动，修改 etc/driver.yaml 中的连接地址
go run driver.go -f etc/driver.yaml
```

### Docker 部署

```bash
# 在项目根目录 deploy/ 下执行
docker-compose up -d driver
```

## 服务发现与集群

- 服务启动后自动注册到 etcd，key 为 `driver.rpc`
- Gateway 通过 etcd 发现本服务并发起 gRPC 调用
- 支持多实例部署：启动多个实例注册同一 key，go-zero 客户端自动负载均衡（P2C 策略）
- docker-compose 可通过 `deploy.replicas: N` 扩容

## 依赖关系

```
driver 服务
  ├── etcd      (服务注册/发现)
  ├── MySQL     (onlinecar_driver 库)
  └── Redis     (已配置，待接入)
       ↑
  gateway 服务通过 etcd 发现并调用 driver.rpc
```

## 错误码定义

| 错误 | Code | 说明 |
|------|------|------|
| `ErrServer` | 500 | 服务器内部错误 |
| `ErrParam` | 400 | 参数错误 |
| `ErrNotFound` | 404 | 资源不存在 |
| `ErrAuth` | 401 | 未授权 |
| `ErrForbidden` | 403 | 无权限访问 |

## 待开发

- [ ] Logic 层业务逻辑实现（当前为 goctl 生成的 stub）
- [ ] Model 层数据访问层
- [ ] 密码加密存储（当前 proto 有 password 字段但未实现加密）
- [ ] Redis 缓存接入（位置数据 GEO 存储）
- [ ] 司机工作状态管理（接单/休息切换）
