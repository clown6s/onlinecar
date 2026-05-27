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

var ErrNotFound = errors.New("record not found")

var (
	districtFieldNames = []string{
		"id", "parent_adcode", "ad_code", "name",
		"level", "center", "created_at",
	}
)

type (
	DistrictModel interface {
		Insert(ctx context.Context, data *District) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*District, error)
		FindByAdCode(ctx context.Context, adCode uint32) (*District, error)
		FindByParentAdCode(ctx context.Context, parentAdCode uint32) ([]*District, error)
		FindProvinces(ctx context.Context) ([]*District, error)
		Update(ctx context.Context, data *District) error
		Delete(ctx context.Context, id int64) error
	}

	defaultDistrictModel struct {
		conn  sqlx.SqlConn
		table string
	}

	District struct {
		Id           int64      `db:"id"`
		ParentAdcode *uint32    `db:"parent_adcode"`
		AdCode       *uint32    `db:"ad_code"`
		Name         string     `db:"name"`
		Level        string     `db:"level"`
		Center       string     `db:"center"`
		CreatedAt    *time.Time `db:"created_at"`
	}
)

func NewDistrictModel(conn sqlx.SqlConn) DistrictModel {
	return &defaultDistrictModel{
		conn:  conn,
		table: "`district`",
	}
}

func (m *defaultDistrictModel) Insert(ctx context.Context, data *District) (sql.Result, error) {
	now := time.Now()
	data.CreatedAt = &now

	query := fmt.Sprintf(
		"INSERT INTO %s (`parent_adcode`,`ad_code`,`name`,`level`,`center`,`created_at`) VALUES (?,?,?,?,?,?)",
		m.table,
	)
	ret, err := m.conn.ExecCtx(ctx, query,
		data.ParentAdcode, data.AdCode, data.Name, data.Level, data.Center, data.CreatedAt,
	)
	return ret, err
}

func (m *defaultDistrictModel) FindOne(ctx context.Context, id int64) (*District, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `id` = ? LIMIT 1", strings.Join(districtFieldNames, ","), m.table)
	var resp District
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultDistrictModel) FindByAdCode(ctx context.Context, adCode uint32) (*District, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `ad_code` = ? LIMIT 1", strings.Join(districtFieldNames, ","), m.table)
	var resp District
	err := m.conn.QueryRowCtx(ctx, &resp, query, adCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

// FindByParentAdCode 查询子级行政区
func (m *defaultDistrictModel) FindByParentAdCode(ctx context.Context, parentAdCode uint32) ([]*District, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `parent_adcode` = ? ORDER BY `ad_code` LIMIT 1000", strings.Join(districtFieldNames, ","), m.table)
	var resp []*District
	err := m.conn.QueryRowsCtx(ctx, &resp, query, parentAdCode)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// FindProvinces 查询省级行政区
// 注意：确保所有省级数据的 parent_adcode 统一为 0（非 NULL）
// 如果存在 NULL 值，需先执行: UPDATE district SET parent_adcode = 0 WHERE parent_adcode IS NULL AND level = 0
func (m *defaultDistrictModel) FindProvinces(ctx context.Context) ([]*District, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `parent_adcode` = 0 ORDER BY `ad_code` ASC", strings.Join(districtFieldNames, ","), m.table)
	var resp []*District
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultDistrictModel) Update(ctx context.Context, data *District) error {
	query := fmt.Sprintf(
		"UPDATE %s SET `parent_adcode`=?,`ad_code`=?,`name`=?,`level`=?,`center`=?,`created_at`=? WHERE `id` = ?",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query,
		data.ParentAdcode, data.AdCode, data.Name, data.Level, data.Center, data.CreatedAt,
		data.Id,
	)
	return err
}

func (m *defaultDistrictModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}
