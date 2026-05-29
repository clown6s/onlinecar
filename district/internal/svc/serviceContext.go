package svc

import (
	"time"

	"onlinecar-district/internal/config"
	"onlinecar-district/internal/model"

	"github.com/zeromicro/go-zero/core/logx"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config        config.Config
	Conn          sqlx.SqlConn
	DistrictModel model.DistrictModel
	Redis         *redis.Redis // 可为 nil（Redis 不可用时降级为直查 DB）
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

	// Redis 非关键依赖：连接失败时降级启动，缓存失效但不影响核心功能
	r, err := redis.NewRedis(c.RedisConf)
	if err != nil {
		logx.Severef("[WARN] Redis 连接失败，District 服务将以无缓存模式启动: %v", err)
	}

	return &ServiceContext{
		Config:        c,
		Conn:          conn,
		DistrictModel: model.NewDistrictModel(conn),
		Redis:         r,
	}
}
