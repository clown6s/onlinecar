package config

import (
	"fmt"
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
func (c *JwtConf) GetAccessSecret(envKey string) (string, error) {
	if v := os.Getenv(envKey); v != "" {
		return v, nil
	}
	if c.AccessSecret == "" {
		return "", fmt.Errorf("JWT secret 未配置：配置文件和环境变量 %s 均为空", envKey)
	}
	return c.AccessSecret, nil
}

// Validate 校验配置必填项
func (c *Config) Validate() error {
	if _, err := c.AuthDriver.GetAccessSecret("ONLINECAR_DRIVER_JWT_SECRET"); err != nil {
		return fmt.Errorf("AuthDriver.JWT Secret: %w", err)
	}
	if _, err := c.AuthPassenger.GetAccessSecret("ONLINECAR_PASSENGER_JWT_SECRET"); err != nil {
		return fmt.Errorf("AuthPassenger.JWT Secret: %w", err)
	}
	return nil
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
