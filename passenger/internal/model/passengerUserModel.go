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
	passengerUserFieldNames = []string{
		"id", "gmt_create", "gmt_modified", "passenger_phone",
		"passenger_name", "passenger_gender", "state", "profile_photo",
	}
)

type (
	PassengerUserModel interface {
		Insert(ctx context.Context, data *PassengerUser) (sql.Result, error)
		InsertWithConn(ctx context.Context, session sqlx.Session, data *PassengerUser) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*PassengerUser, error)
		FindByPhone(ctx context.Context, phone string) (*PassengerUser, error)
		FindByPhoneWithConn(ctx context.Context, session sqlx.Session, phone string) (*PassengerUser, error)
		Update(ctx context.Context, data *PassengerUser) error
		Delete(ctx context.Context, id int64) error
	}

	defaultPassengerUserModel struct {
		conn  sqlx.SqlConn
		table string
	}

	PassengerUser struct {
		Id              int64      `db:"id"`
		GmtCreate       *time.Time `db:"gmt_create"`
		GmtModified     *time.Time `db:"gmt_modified"`
		PassengerPhone  *string    `db:"passenger_phone"`
		PassengerName   *string    `db:"passenger_name"`
		PassengerGender *int8      `db:"passenger_gender"`
		State           *int8      `db:"state"`
		ProfilePhoto    *string    `db:"profile_photo"`
	}
)

func NewPassengerUserModel(conn sqlx.SqlConn) PassengerUserModel {
	return &defaultPassengerUserModel{
		conn:  conn,
		table: "`passenger_user`",
	}
}

func (m *defaultPassengerUserModel) Insert(ctx context.Context, data *PassengerUser) (sql.Result, error) {
	now := time.Now()
	data.GmtCreate = &now
	data.GmtModified = &now

	query := fmt.Sprintf(
		"INSERT INTO %s (`passenger_phone`,`passenger_name`,`passenger_gender`,`state`,`profile_photo`) VALUES (?,?,?,?,?)",
		m.table,
	)
	ret, err := m.conn.ExecCtx(ctx, query,
		data.PassengerPhone, data.PassengerName, data.PassengerGender, data.State, data.ProfilePhoto,
	)
	return ret, err
}

func (m *defaultPassengerUserModel) FindOne(ctx context.Context, id int64) (*PassengerUser, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `id` = ? LIMIT 1", strings.Join(passengerUserFieldNames, ","), m.table)
	var resp PassengerUser
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultPassengerUserModel) FindByPhone(ctx context.Context, phone string) (*PassengerUser, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `passenger_phone` = ? LIMIT 1", strings.Join(passengerUserFieldNames, ","), m.table)
	var resp PassengerUser
	err := m.conn.QueryRowCtx(ctx, &resp, query, phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultPassengerUserModel) Update(ctx context.Context, data *PassengerUser) error {
	now := time.Now()
	data.GmtModified = &now

	query := fmt.Sprintf(
		"UPDATE %s SET `passenger_phone`=?,`passenger_name`=?,`passenger_gender`=?,`state`=?,`profile_photo`=?,`gmt_modified`=? WHERE `id` = ?",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query,
		data.PassengerPhone, data.PassengerName, data.PassengerGender, data.State, data.ProfilePhoto,
		data.GmtModified, data.Id,
	)
	return err
}

func (m *defaultPassengerUserModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// InsertWithConn 使用指定 session 执行插入（事务内使用）
func (m *defaultPassengerUserModel) InsertWithConn(ctx context.Context, session sqlx.Session, data *PassengerUser) (sql.Result, error) {
	now := time.Now()
	data.GmtCreate = &now
	data.GmtModified = &now

	query := fmt.Sprintf(
		"INSERT INTO %s (`passenger_phone`,`passenger_name`,`passenger_gender`,`state`,`profile_photo`) VALUES (?,?,?,?,?)",
		m.table,
	)
	ret, err := session.ExecCtx(ctx, query,
		data.PassengerPhone, data.PassengerName, data.PassengerGender, data.State, data.ProfilePhoto,
	)
	return ret, err
}

// FindByPhoneWithConn 使用指定 session 查询（事务内使用）
func (m *defaultPassengerUserModel) FindByPhoneWithConn(ctx context.Context, session sqlx.Session, phone string) (*PassengerUser, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `passenger_phone` = ? LIMIT 1", strings.Join(passengerUserFieldNames, ","), m.table)
	var resp PassengerUser
	err := session.QueryRowCtx(ctx, &resp, query, phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &resp, nil
}
