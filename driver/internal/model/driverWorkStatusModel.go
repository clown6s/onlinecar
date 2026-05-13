package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	driverWorkStatusFieldNames = []string{
		"id", "driver_id", "work_status", "gmt_create", "gmt_modified",
	}
)

type (
	DriverWorkStatusModel interface {
		Upsert(ctx context.Context, driverId, workStatus int64) error
		FindByDriverId(ctx context.Context, driverId int64) (*DriverWorkStatus, error)
	}

	defaultDriverWorkStatusModel struct {
		conn  sqlx.SqlConn
		table string
	}

	DriverWorkStatus struct {
		Id          int64      `db:"id"`
		DriverId    *int64     `db:"driver_id"`
		WorkStatus  *int       `db:"work_status"`
		GmtCreate   *time.Time `db:"gmt_create"`
		GmtModified *time.Time `db:"gmt_modified"`
	}
)

func NewDriverWorkStatusModel(conn sqlx.SqlConn) DriverWorkStatusModel {
	return &defaultDriverWorkStatusModel{
		conn:  conn,
		table: "`driver_user_work_status`",
	}
}

func (m *defaultDriverWorkStatusModel) Upsert(ctx context.Context, driverId, workStatus int64) error {
	now := time.Now()
	query := fmt.Sprintf(
		"INSERT INTO %s (`id`,`driver_id`,`work_status`,`gmt_create`,`gmt_modified`) VALUES (?,?,?,?,?) ON DUPLICATE KEY UPDATE `work_status`=?,`gmt_modified`=?",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query, driverId, driverId, workStatus, now, now, workStatus, now)
	return err
}

func (m *defaultDriverWorkStatusModel) FindByDriverId(ctx context.Context, driverId int64) (*DriverWorkStatus, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `driver_id` = ? LIMIT 1", strings.Join(driverWorkStatusFieldNames, ","), m.table)
	var resp DriverWorkStatus
	err := m.conn.QueryRowCtx(ctx, &resp, query, driverId)
	if err != nil {
		return nil, nil
	}
	return &resp, nil
}
