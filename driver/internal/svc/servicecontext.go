package svc

import (
	"onlinecar-driver/internal/config"
	"onlinecar-driver/internal/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config                config.Config
	Conn                  sqlx.SqlConn
	DriverUserModel       model.DriverUserModel
	CarModel              model.CarModel
	DriverCarBindingModel model.DriverCarBindingModel
	DriverWorkStatusModel model.DriverWorkStatusModel
	Redis                 *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:                c,
		Conn:                  conn,
		DriverUserModel:       model.NewDriverUserModel(conn),
		CarModel:              model.NewCarModel(conn),
		DriverCarBindingModel: model.NewDriverCarBindingModel(conn),
		DriverWorkStatusModel: model.NewDriverWorkStatusModel(conn),
		Redis:                 redis.MustNewRedis(c.RedisConf),
	}
}
