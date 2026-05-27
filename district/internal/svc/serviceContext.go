package svc

import (
	"onlinecar-district/internal/config"
	"onlinecar-district/internal/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config        config.Config
	Conn          sqlx.SqlConn
	DistrictModel model.DistrictModel
	Redis         *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:        c,
		Conn:          conn,
		DistrictModel: model.NewDistrictModel(conn),
		Redis:         redis.MustNewRedis(c.RedisConf),
	}
}
