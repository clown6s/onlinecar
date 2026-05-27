package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	districtFieldNames = []string{
		"id", "ad_code", "name", "level", "center",
		"parent_ad_code", "gmt_create", "gmt_modified",
	}
)

type (
	DistrictModel interface {
		Insert(ctx context.Context, data *District) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*District, error)
		FindByAdCode(ctx context.Context, adCode uint32) (*District, error)
		FindByParentAdCode(ctx context.Context, parentAdCode uint32) ([]*District, error)
		Update(ctx context.Context, data *District) error
		Delete(ctx context.Context, id int64) error
	}

	defaultDistrictModel struct {
		conn  sqlx.SqlConn
		table string
	}

	District struct {
		Id           int64      `db:"id"`
		AdCode       uint32     `db:"ad_code"`
		Name         string     `db:"name"`
		Level        string     `db:"level"`
		Center       string     `db:"center"`
		ParentAdCode uint32     `db:"parent_ad_code"`
		GmtCreate    *time.Time `db:"gmt_create"`
		GmtModified  *time.Time `db:"gmt_modified"`
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
	data.GmtCreate = &now
	data.GmtModified = &now

	query := fmt.Sprintf(
		"INSERT INTO %s (`ad_code`,`name`,`level`,`center`,`parent_ad_code`) VALUES (?,?,?,?,?)",
		m.table,
	)
	ret, err := m.conn.ExecCtx(ctx, query,
		data.AdCode, data.Name, data.Level, data.Center, data.ParentAdCode,
	)
	return ret, err
}

func (m *defaultDistrictModel) FindOne(ctx context.Context, id int64) (*District, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `id` = ? LIMIT 1", strings.Join(districtFieldNames, ","), m.table)
	var resp District
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
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
			return nil, nil
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultDistrictModel) FindByParentAdCode(ctx context.Context, parentAdCode uint32) ([]*District, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `parent_ad_code` = ? ORDER BY `ad_code` ASC", strings.Join(districtFieldNames, ","), m.table)
	var resp []*District
	err := m.conn.QueryRowsCtx(ctx, &resp, query, parentAdCode)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *defaultDistrictModel) Update(ctx context.Context, data *District) error {
	now := time.Now()
	data.GmtModified = &now

	query := fmt.Sprintf(
		"UPDATE %s SET `ad_code`=?,`name`=?,`level`=?,`center`=?,`parent_ad_code`=?,`gmt_modified`=? WHERE `id` = ?",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query,
		data.AdCode, data.Name, data.Level, data.Center, data.ParentAdCode, data.GmtModified,
		data.Id,
	)
	return err
}

func (m *defaultDistrictModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}
