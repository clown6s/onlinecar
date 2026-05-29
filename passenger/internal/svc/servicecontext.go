package svc

import (
	"time"

	"onlinecar-passenger/internal/config"
	"onlinecar-passenger/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config             config.Config
	Conn               sqlx.SqlConn // 跨包需要公开，仅供内部事务使用，勿直接写SQL
	PassengerUserModel model.PassengerUserModel
	TripModel          model.TripModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	rawDB, err := conn.RawDB()
	if err == nil && rawDB != nil {
		rawDB.SetMaxOpenConns(c.MaxOpenConns)
		rawDB.SetMaxIdleConns(c.MaxIdleConns)
		if c.ConnMaxLifetime > 0 {
			rawDB.SetConnMaxLifetime(time.Duration(c.ConnMaxLifetime) * time.Second)
		}
	}
	return &ServiceContext{
		Config:             c,
		Conn:               conn,
		PassengerUserModel: model.NewPassengerUserModel(conn),
		TripModel:          model.NewTripModel(conn),
	}
}
