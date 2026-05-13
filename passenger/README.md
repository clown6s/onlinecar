# OnlineCar Passenger Service

> 网约车平台 — 乘客 gRPC 微服务

## 概述

乘客服务是 OnlineCar 网约车平台的核心微服务之一，负责乘客用户管理和行程创建。基于 go-zero 框架的 zRPC 构建，使用 gRPC 协议对外提供服务，通过 etcd 进行服务注册与发现。

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
passenger/
├── passenger.go                      # 程序入口
├── passenger.proto                   # gRPC 服务定义（Protobuf）
├── go.mod / go.sum                   # Go 模块依赖
├── etc/
│   └── passenger.yaml                # 运行时配置
├── pb/
│   ├── passenger.pb.go               # Protobuf 生成消息
│   └── passenger_grpc.pb.go          # gRPC 生成代码
└── internal/
    ├── config/
    │   └── config.go                 # 配置结构体
    ├── svc/
    │   └── servicecontext.go         # 服务上下文（依赖注入）
    ├── server/
    │   └── passenger/
    │       └── passengerServer.go    # gRPC Server（goctl 生成）
    └── logic/
        └── passenger/
            ├── registerLogic.go      # 乘客注册
            ├── loginLogic.go         # 乘客登录
            ├── getPassengerLogic.go  # 获取乘客信息
            └── createTripLogic.go    # 创建行程
```

## gRPC 接口定义

### 服务：`Passenger`

| 方法 | 请求 | 响应 | 说明 |
|------|------|------|------|
| `Register` | `RegisterReq` | `RegisterResp` | 乘客注册 |
| `Login` | `LoginReq` | `LoginResp` | 乘客登录 |
| `GetPassenger` | `GetPassengerReq` | `PassengerInfo` | 查询乘客详情 |
| `CreateTrip` | `CreateTripReq` | `TripResp` | 创建行程（叫车） |

### Protobuf 消息定义

```protobuf
// 注册请求
message RegisterReq {
  string passenger_name   = 1;  // 乘客姓名
  string passenger_phone  = 2;  // 手机号（唯一）
  int64  passenger_gender = 3;  // 性别: 0=未知 1=男 2=女
}

// 注册响应
message RegisterResp {
  int64  id               = 1;  // 乘客ID
  string passenger_name   = 2;  // 乘客姓名
  string passenger_phone  = 3;  // 手机号
  int64  passenger_gender = 4;  // 性别
  int64  state            = 5;  // 账户状态
}

// 登录请求
message LoginReq {
  string passenger_phone = 1;  // 手机号
}

// 登录响应
message LoginResp {
  int64  id              = 1;  // 乘客ID
  string passenger_name  = 2;  // 乘客姓名
  string passenger_phone = 3;  // 手机号
  int64  state           = 4;  // 账户状态
  string profile_photo   = 5;  // 头像 URL
}

// 查询乘客请求
message GetPassengerReq {
  int64 id = 1;  // 乘客ID
}

// 乘客详细信息
message PassengerInfo {
  int64  id               = 1;  // 乘客ID
  string passenger_name   = 2;  // 乘客姓名
  string passenger_phone  = 3;  // 手机号
  int64  passenger_gender = 4;  // 性别
  int64  state            = 5;  // 账户状态
  string profile_photo    = 6;  // 头像 URL
}

// 创建行程请求
message CreateTripReq {
  int64   passenger_id = 1;  // 乘客ID（由 Gateway 从 JWT 注入）
  string  origin       = 2;  // 出发地描述
  string  destination  = 3;  // 目的地描述
  double  origin_lng   = 4;  // 出发地经度
  double  origin_lat   = 5;  // 出发地纬度
  double  dest_lng     = 6;  // 目的地经度
  double  dest_lat     = 7;  // 目的地纬度
}

// 行程响应
message TripResp {
  int64   id           = 1;  // 行程ID
  int64   passenger_id = 2;  // 乘客ID
  int64   driver_id    = 3;  // 司机ID（待匹配时为 0）
  string  origin       = 4;  // 出发地
  string  destination  = 5;  // 目的地
  int64   status       = 6;  // 行程状态
  double  fare         = 7;  // 费用
}
```

## 数据库设计

### 数据库：`onlinecar_passenger`

#### 乘客用户表 `passenger_user`

```sql
CREATE TABLE `passenger_user` (
  `id`               BIGINT(32) UNSIGNED NOT NULL AUTO_INCREMENT,
  `gmt_create`       DATETIME          NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  `gmt_modified`     DATETIME          NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  `passenger_phone`  VARCHAR(16)       NULL DEFAULT NULL COMMENT '手机号',
  `passenger_name`   VARCHAR(16)       NULL DEFAULT NULL COMMENT '姓名',
  `passenger_gender` TINYINT(1)        NULL DEFAULT NULL COMMENT '0=未知 1=男 2=女',
  `state`            TINYINT(1)        NULL DEFAULT NULL COMMENT '0=有效 1=失效',
  `profile_photo`    VARCHAR(128)      NULL DEFAULT NULL COMMENT '头像URL',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COMMENT = '乘客用户表';
```

**性别枚举：**

| 值 | 含义 |
|----|------|
| 0 | 未知 |
| 1 | 男 |
| 2 | 女 |

**状态枚举：**

| 值 | 含义 |
|----|------|
| 0 | 有效 |
| 1 | 失效 |

#### 行程表 `trip`

```sql
CREATE TABLE `trip` (
    `id`           BIGINT       NOT NULL AUTO_INCREMENT COMMENT '行程ID',
    `passenger_id` BIGINT       NOT NULL DEFAULT 0 COMMENT '乘客ID',
    `driver_id`    BIGINT       NOT NULL DEFAULT 0 COMMENT '司机ID',
    `origin`       VARCHAR(255) NOT NULL DEFAULT '' COMMENT '出发地',
    `destination`  VARCHAR(255) NOT NULL DEFAULT '' COMMENT '目的地',
    `origin_lng`   DOUBLE       NOT NULL DEFAULT 0 COMMENT '出发经度',
    `origin_lat`   DOUBLE       NOT NULL DEFAULT 0 COMMENT '出发纬度',
    `dest_lng`     DOUBLE       NOT NULL DEFAULT 0 COMMENT '目的经度',
    `dest_lat`     DOUBLE       NOT NULL DEFAULT 0 COMMENT '目的纬度',
    `status`       TINYINT      NOT NULL DEFAULT 1 COMMENT '状态',
    `fare`         DECIMAL(10,2) NOT NULL DEFAULT 0.00 COMMENT '费用',
    `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    KEY `idx_passenger` (`passenger_id`),
    KEY `idx_driver` (`driver_id`),
    KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='行程表';
```

**行程状态枚举：**

| 值 | 含义 |
|----|------|
| 1 | 待接单 |
| 2 | 已接单 |
| 3 | 行程中 |
| 4 | 已完成 |
| 5 | 已取消 |

## 配置说明

### `etc/passenger.yaml`

```yaml
Name: passenger.rpc                  # 服务名称（etcd 注册 key）
ListenOn: 0.0.0.0:9002              # gRPC 监听地址

Etcd:                                # 服务发现
  Hosts:
    - 10.211.55.5:2379               # etcd 地址
  Key: passenger.rpc                 # 注册键名

DataSource: "root:123456@tcp(10.211.55.5:3306)/onlinecar_passenger?charset=utf8mb4&parseTime=true&loc=Local"  # MySQL DSN

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
  ├── 加载配置文件 etc/passenger.yaml
  ├── 禁用 logx 统计 (logx.DisableStat)
  ├── 创建 ServiceContext（初始化 MySQL）
  ├── 注册 gRPC Server (pb.RegisterPassengerServer)
  ├── 注册到 etcd (key: passenger.rpc)
  └── 启动监听 (0.0.0.0:9002)
```

## 快速开始

### 环境要求

- Go 1.25+
- MySQL 8.0+
- etcd 3.5+
- protoc + goctl（代码生成）

### 代码生成

```bash
# 在 passenger/ 目录下执行
goctl rpc protoc passenger.proto --go_out=. --go-grpc_out=. --zrpc_out=. --multiple --style=goZero
```

### 本地运行

```bash
# 确保 MySQL 和 etcd 已启动，修改 etc/passenger.yaml 中的连接地址
go run passenger.go -f etc/passenger.yaml
```

### Docker 部署

```bash
# 在项目根目录 deploy/ 下执行
docker-compose up -d passenger
```

## 服务发现与集群

- 服务启动后自动注册到 etcd，key 为 `passenger.rpc`
- Gateway 通过 etcd 发现本服务并发起 gRPC 调用
- 支持多实例部署：启动多个实例注册同一 key，go-zero 客户端自动负载均衡（P2C 策略）
- docker-compose 可通过 `deploy.replicas: N` 扩容

## 依赖关系

```
passenger 服务
  ├── etcd      (服务注册/发现)
  ├── MySQL     (onlinecar_passenger 库)
  └── Redis     (已配置，待接入)
       ↑
  gateway 服务通过 etcd 发现并调用 passenger.rpc
```

## 待开发

- [ ] Logic 层业务逻辑实现（当前为 goctl 生成的 stub）
- [ ] Model 层数据访问层
- [ ] 密码认证（当前登录仅通过手机号，未校验密码）
- [ ] Redis 缓存接入
- [ ] 行程匹配算法（乘客叫车后匹配附近空闲司机）
- [ ] 行程状态流转（接单 → 行程中 → 完成）
