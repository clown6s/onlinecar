package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	driverCarBindingFieldNames = []string{
		"id", "driver_id", "car_id", "bind_state", "binding_time", "un_binding_time",
	}
)

type (
	DriverCarBindingModel interface {
		Insert(ctx context.Context, data *DriverCarBinding) (sql.Result, error)
		InsertWithConn(ctx context.Context, session sqlx.Session, data *DriverCarBinding) (sql.Result, error)
		FindByDriverId(ctx context.Context, driverId int64) (*DriverCarBinding, error)
		Update(ctx context.Context, data *DriverCarBinding) error
	}

	defaultDriverCarBindingModel struct {
		conn  sqlx.SqlConn
		table string
	}

	DriverCarBinding struct {
		Id            int64      `db:"id"`
		DriverId      *int64     `db:"driver_id"`
		CarId         *int64     `db:"car_id"`
		BindState     *int       `db:"bind_state"`
		BindingTime   *time.Time `db:"binding_time"`
		UnBindingTime *time.Time `db:"un_binding_time"`
	}
)

func NewDriverCarBindingModel(conn sqlx.SqlConn) DriverCarBindingModel {
	return &defaultDriverCarBindingModel{
		conn:  conn,
		table: "`driver_car_binding_relationship`",
	}
}

func (m *defaultDriverCarBindingModel) Insert(ctx context.Context, data *DriverCarBinding) (sql.Result, error) {
	query := fmt.Sprintf(
		"INSERT INTO %s (`driver_id`,`car_id`,`bind_state`,`binding_time`) VALUES (?,?,?,?)",
		m.table,
	)
	ret, err := m.conn.ExecCtx(ctx, query, data.DriverId, data.CarId, data.BindState, data.BindingTime)
	return ret, err
}

func (m *defaultDriverCarBindingModel) FindByDriverId(ctx context.Context, driverId int64) (*DriverCarBinding, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `driver_id` = ? AND `bind_state` = 1 LIMIT 1", strings.Join(driverCarBindingFieldNames, ","), m.table)
	var resp DriverCarBinding
	err := m.conn.QueryRowCtx(ctx, &resp, query, driverId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultDriverCarBindingModel) Update(ctx context.Context, data *DriverCarBinding) error {
	query := fmt.Sprintf(
		"UPDATE %s SET `bind_state`=?,`binding_time`=?,`un_binding_time`=? WHERE `id` = ?",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query, data.BindState, data.BindingTime, data.UnBindingTime, data.Id)
	return err
}

// InsertWithConn 使用指定 session 执行插入（事务内使用）
func (m *defaultDriverCarBindingModel) InsertWithConn(ctx context.Context, session sqlx.Session, data *DriverCarBinding) (sql.Result, error) {
	query := fmt.Sprintf(
		"INSERT INTO %s (`driver_id`,`car_id`,`bind_state`,`binding_time`) VALUES (?,?,?,?)",
		m.table,
	)
	ret, err := session.ExecCtx(ctx, query, data.DriverId, data.CarId, data.BindState, data.BindingTime)
	return ret, err
}
