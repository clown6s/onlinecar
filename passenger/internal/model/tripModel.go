package model

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	tripFieldNames = []string{
		"id", "passenger_id", "driver_id", "origin", "destination",
		"origin_lng", "origin_lat", "dest_lng", "dest_lat",
		"status", "fare", "gmt_create", "gmt_modified",
	}
)

type (
	TripModel interface {
		Insert(ctx context.Context, data *Trip) (sql.Result, error)
		InsertWithConn(ctx context.Context, session sqlx.Session, data *Trip) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*Trip, error)
		FindByPassengerId(ctx context.Context, passengerId int64, page, pageSize int) ([]*Trip, error)
		FindByPassengerIdAndStatus(ctx context.Context, passengerId int64, status int8, page, pageSize int) ([]*Trip, error)
		FindActiveTrips(ctx context.Context, passengerId int64, statuses []int8) ([]*Trip, error)
		Update(ctx context.Context, data *Trip) error
	}

	defaultTripModel struct {
		conn  sqlx.SqlConn
		table string
	}

	Trip struct {
		Id          int64      `db:"id"`
		PassengerId int64      `db:"passenger_id"`
		DriverId    int64      `db:"driver_id"`
		Origin      string     `db:"origin"`
		Destination string     `db:"destination"`
		OriginLng   float64    `db:"origin_lng"`
		OriginLat   float64    `db:"origin_lat"`
		DestLng     float64    `db:"dest_lng"`
		DestLat     float64    `db:"dest_lat"`
		Status      int8       `db:"status"` // 0待接单 1进行中 2已完成 3已取消
		Fare        float64    `db:"fare"`
		GmtCreate   *time.Time `db:"gmt_create"`
		GmtModified *time.Time `db:"gmt_modified"`
	}
)

func NewTripModel(conn sqlx.SqlConn) TripModel {
	return &defaultTripModel{
		conn:  conn,
		table: "`trip`",
	}
}

func (m *defaultTripModel) Insert(ctx context.Context, data *Trip) (sql.Result, error) {
	now := time.Now()
	data.GmtCreate = &now
	data.GmtModified = &now

	query := fmt.Sprintf(
		"INSERT INTO %s (`passenger_id`,`origin`,`destination`,`origin_lng`,`origin_lat`,`dest_lng`,`dest_lat`,`status`,`fare`) VALUES (?,?,?,?,?,?,?,?,?)",
		m.table,
	)
	ret, err := m.conn.ExecCtx(ctx, query,
		data.PassengerId, data.Origin, data.Destination,
		data.OriginLng, data.OriginLat, data.DestLng, data.DestLat,
		data.Status, data.Fare,
	)
	return ret, err
}

func (m *defaultTripModel) InsertWithConn(ctx context.Context, session sqlx.Session, data *Trip) (sql.Result, error) {
	now := time.Now()
	data.GmtCreate = &now
	data.GmtModified = &now

	query := fmt.Sprintf(
		"INSERT INTO %s (`passenger_id`,`origin`,`destination`,`origin_lng`,`origin_lat`,`dest_lng`,`dest_lat`,`status`,`fare`) VALUES (?,?,?,?,?,?,?,?,?)",
		m.table,
	)
	ret, err := session.ExecCtx(ctx, query,
		data.PassengerId, data.Origin, data.Destination,
		data.OriginLng, data.OriginLat, data.DestLng, data.DestLat,
		data.Status, data.Fare,
	)
	return ret, err
}

func (m *defaultTripModel) FindOne(ctx context.Context, id int64) (*Trip, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `id` = ? LIMIT 1", strings.Join(tripFieldNames, ","), m.table)
	var resp Trip
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultTripModel) FindByPassengerId(ctx context.Context, passengerId int64, page, pageSize int) ([]*Trip, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf("SELECT %s FROM %s WHERE `passenger_id` = ? ORDER BY `id` DESC LIMIT ? OFFSET ?",
		strings.Join(tripFieldNames, ","), m.table)
	var resp []*Trip
	err := m.conn.QueryRowsCtx(ctx, &resp, query, passengerId, pageSize, offset)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultTripModel) FindByPassengerIdAndStatus(ctx context.Context, passengerId int64, status int8, page, pageSize int) ([]*Trip, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf("SELECT %s FROM %s WHERE `passenger_id` = ? AND `status` = ? ORDER BY `id` DESC LIMIT ? OFFSET ?",
		strings.Join(tripFieldNames, ","), m.table)
	var resp []*Trip
	err := m.conn.QueryRowsCtx(ctx, &resp, query, passengerId, status, pageSize, offset)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// FindActiveTrips 查询乘客指定状态列表中的未完成行程（用于创建行程前的冲突检查）
func (m *defaultTripModel) FindActiveTrips(ctx context.Context, passengerId int64, statuses []int8) ([]*Trip, error) {
	if len(statuses) == 0 {
		return nil, nil
	}

	// 构建 IN (?, ?, ...) 占位符
	placeholders := make([]string, len(statuses))
	args := make([]interface{}, 0, len(statuses)+1)
	args = append(args, passengerId)
	for i, s := range statuses {
		placeholders[i] = "?"
		args = append(args, s)
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE `passenger_id` = ? AND `status` IN (%s) ORDER BY `id` DESC LIMIT 1",
		strings.Join(tripFieldNames, ","), m.table, strings.Join(placeholders, ","))

	logx.Debugf("FindActiveTrips query: %s, args: %v", query, args)

	var trips []*Trip
	err := m.conn.QueryRowsCtx(ctx, &trips, query, args...)
	if err != nil {
		return nil, err
	}
	return trips, nil
}

func (m *defaultTripModel) Update(ctx context.Context, data *Trip) error {
	now := time.Now()
	data.GmtModified = &now

	query := fmt.Sprintf(
		"UPDATE %s SET `driver_id`=?,`status`=?,`fare`=?,`gmt_modified`=? WHERE `id` = ?",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query,
		data.DriverId, data.Status, data.Fare, data.GmtModified, data.Id,
	)
	return err
}
