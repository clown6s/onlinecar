package svc

import (
	"onlinecar-gateway/internal/config"
	driverpb "onlinecar-gateway/pb/driver"
	passengerpb "onlinecar-gateway/pb/passenger"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config       config.Config
	DriverRpc    driverpb.DriverClient
	PassengerRpc passengerpb.PassengerClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	driverConn := zrpc.MustNewClient(c.DriverRpc)
	passengerConn := zrpc.MustNewClient(c.PassengerRpc)

	return &ServiceContext{
		Config:       c,
		DriverRpc:    driverpb.NewDriverClient(driverConn.Conn()),
		PassengerRpc: passengerpb.NewPassengerClient(passengerConn.Conn()),
	}
}
