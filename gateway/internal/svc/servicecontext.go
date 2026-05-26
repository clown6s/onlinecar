package svc

import (
	"onlinecar-gateway/common/sms"
	"onlinecar-gateway/internal/config"
	driverpb "onlinecar-gateway/pb/driver"
	passengerpb "onlinecar-gateway/pb/passenger"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type ServiceContext struct {
	Config        config.Config
	DriverRpc     driverpb.DriverClient
	PassengerRpc  passengerpb.PassengerClient
	DriverConn    *grpc.ClientConn
	PassengerConn *grpc.ClientConn
	Redis         *redis.Redis
	SmsSender     sms.SmsSender
}

func NewServiceContext(c config.Config) *ServiceContext {
	driverConn := zrpc.MustNewClient(c.DriverRpc)
	passengerConn := zrpc.MustNewClient(c.PassengerRpc)

	return &ServiceContext{
		Config:        c,
		DriverRpc:     driverpb.NewDriverClient(driverConn.Conn()),
		PassengerRpc:  passengerpb.NewPassengerClient(passengerConn.Conn()),
		DriverConn:    driverConn.Conn(),
		PassengerConn: passengerConn.Conn(),
		Redis:         redis.MustNewRedis(c.RedisConf),
		SmsSender:     sms.NewMockSmsSender(),
	}
}
