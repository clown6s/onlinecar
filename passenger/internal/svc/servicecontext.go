package svc

import (
	"onlinecar-passenger/internal/config"
	"onlinecar-passenger/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config             config.Config
	Conn               sqlx.SqlConn
	PassengerUserModel model.PassengerUserModel
	TripModel          model.TripModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:             c,
		Conn:               conn,
		PassengerUserModel: model.NewPassengerUserModel(conn),
		TripModel:          model.NewTripModel(conn),
	}
}
