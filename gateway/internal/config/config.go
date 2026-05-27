package config

import (
	"os"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type JwtConf struct {
	AccessSecret string        `json:",default="`
	AccessExpire time.Duration `json:",default=86400s"`
}

// GetAccessSecret 优先读取环境变量，无则使用配置文件值
func (c *JwtConf) GetAccessSecret(envKey string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return c.AccessSecret
}

type Config struct {
	rest.RestConf
	AuthDriver    JwtConf
	AuthPassenger JwtConf
	DriverRpc     zrpc.RpcClientConf
	PassengerRpc  zrpc.RpcClientConf
	DistrictRpc   zrpc.RpcClientConf

	RedisConf redis.RedisConf
	CORS      CORSConf
}

type CORSConf struct {
	AllowOrigins []string `json:",default=[]"`
}
