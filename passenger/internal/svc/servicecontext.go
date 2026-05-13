package svc

import (
	"onlinecar-passenger/internal/config"
	"onlinecar-passenger/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config             config.Config
	PassengerUserModel model.PassengerUserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:             c,
		PassengerUserModel: model.NewPassengerUserModel(conn),
	}
}
