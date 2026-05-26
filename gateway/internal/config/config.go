package config

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type JwtConf struct {
	AccessSecret string        `json:",default="`
	AccessExpire time.Duration `json:",default=86400s"`
}

type Config struct {
	rest.RestConf
	AuthDriver    JwtConf
	AuthPassenger JwtConf
	DriverRpc     zrpc.RpcClientConf
	PassengerRpc  zrpc.RpcClientConf
	RedisConf     redis.RedisConf
}
