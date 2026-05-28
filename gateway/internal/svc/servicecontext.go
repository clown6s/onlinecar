package svc

import (
	"os"
	"time"

	"onlinecar-gateway/common/cache"
	"onlinecar-gateway/common/sms"
	"onlinecar-gateway/internal/config"
	districtpb "onlinecar-gateway/pb/district"
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
	DistrictRpc   districtpb.DistrictServiceClient
	DriverConn    *grpc.ClientConn
	PassengerConn *grpc.ClientConn
	DistrictConn  *grpc.ClientConn
	Redis         *redis.Redis
	SmsSender     sms.SmsSender
	ProvinceCache *cache.ProvinceCache
}

// Close 优雅关闭所有连接资源
func (s *ServiceContext) Close() error {
	var errs []error
	if s.DriverConn != nil {
		if err := s.DriverConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.PassengerConn != nil {
		if err := s.PassengerConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.DistrictConn != nil {
		if err := s.DistrictConn.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if s.ProvinceCache != nil {
		s.ProvinceCache.Stop()
	}
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func NewServiceContext(c config.Config) *ServiceContext {
	driverConn := zrpc.MustNewClient(c.DriverRpc)
	passengerConn := zrpc.MustNewClient(c.PassengerRpc)
	districtConn := zrpc.MustNewClient(c.DistrictRpc)

	districtClient := districtpb.NewDistrictServiceClient(districtConn.Conn())

	// 短信发送器：通过环境变量 ONLINECAR_SMS_PROVIDER 选择实现
	// 默认 mock（开发环境），设为 "aliyun" 对接阿里云短信
	var smsSender sms.SmsSender = sms.NewMockSmsSender()
	if provider := os.Getenv("ONLINECAR_SMS_PROVIDER"); provider == "aliyun" {
		smsSender = sms.NewAliyunSmsSender()
	}

	return &ServiceContext{
		Config:        c,
		DriverRpc:     driverpb.NewDriverClient(driverConn.Conn()),
		PassengerRpc:  passengerpb.NewPassengerClient(passengerConn.Conn()),
		DistrictRpc:   districtClient,
		DriverConn:    driverConn.Conn(),
		PassengerConn: passengerConn.Conn(),
		DistrictConn:  districtConn.Conn(),
		Redis:         redis.MustNewRedis(c.RedisConf),
		SmsSender:     smsSender,
		ProvinceCache: cache.NewProvinceCache(districtClient, 10*time.Minute),
	}
}
