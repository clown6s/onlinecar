package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource      string
	MaxOpenConns    int `json:",default=100"`
	MaxIdleConns    int `json:",default=20"`
	ConnMaxLifetime int `json:",default=300"` // 秒
	RedisConf       redis.RedisConf
	MetricsPort     int `json:",default=9100"`
}
