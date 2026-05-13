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
	carFieldNames = []string{
		"id", "address", "vehicle_no", "plate_color", "seats",
		"brand", "model", "vehicle_type", "owner_name", "vehicle_color",
		"engine_id", "vin", "certify_date_a", "fue_type", "engine_displace",
		"trans_agency", "trans_area", "trans_date_start", "trans_date_end",
		"certify_date_b", "fix_state", "next_fix_date", "check_state",
		"fee_print_id", "gps_brand", "gps_model", "gps_install_date",
		"register_date", "commercial_type", "fare_type", "state",
		"tid", "trid", "trname", "gmt_create", "gmt_modified",
	}
)

type (
	CarModel interface {
		Insert(ctx context.Context, data *Car) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*Car, error)
		FindByVehicleNo(ctx context.Context, vehicleNo string) (*Car, error)
		Update(ctx context.Context, data *Car) error
		Delete(ctx context.Context, id int64) error
	}

	defaultCarModel struct {
		conn  sqlx.SqlConn
		table string
	}

	Car struct {
		Id             int64      `db:"id"`
		Address        string     `db:"address"`
		VehicleNo      string     `db:"vehicle_no"`
		PlateColor     *string    `db:"plate_color"`
		Seats          *int       `db:"seats"`
		Brand          *string    `db:"brand"`
		Model          *string    `db:"model"`
		VehicleType    *string    `db:"vehicle_type"`
		OwnerName      *string    `db:"owner_name"`
		VehicleColor   *string    `db:"vehicle_color"`
		EngineId       *string    `db:"engine_id"`
		Vin            *string    `db:"vin"`
		CertifyDateA   *time.Time `db:"certify_date_a"`
		FueType        *string    `db:"fue_type"`
		EngineDisplace *string    `db:"engine_displace"`
		TransAgency    *string    `db:"trans_agency"`
		TransArea      *string    `db:"trans_area"`
		TransDateStart *time.Time `db:"trans_date_start"`
		TransDateEnd   *time.Time `db:"trans_date_end"`
		CertifyDateB   *time.Time `db:"certify_date_b"`
		FixState       *string    `db:"fix_state"`
		NextFixDate    *time.Time `db:"next_fix_date"`
		CheckState     *string    `db:"check_state"`
		FeePrintId     *string    `db:"fee_print_id"`
		GpsBrand       *string    `db:"gps_brand"`
		GpsModel       *string    `db:"gps_model"`
		GpsInstallDate *time.Time `db:"gps_install_date"`
		RegisterDate   *time.Time `db:"register_date"`
		CommercialType *int       `db:"commercial_type"`
		FareType       *string    `db:"fare_type"`
		State          *int8      `db:"state"`
		Tid            *string    `db:"tid"`
		Trid           *string    `db:"trid"`
		Trname         *string    `db:"trname"`
		GmtCreate      *time.Time `db:"gmt_create"`
		GmtModified    *time.Time `db:"gmt_modified"`
	}
)

func NewCarModel(conn sqlx.SqlConn) CarModel {
	return &defaultCarModel{
		conn:  conn,
		table: "`car`",
	}
}

func (m *defaultCarModel) Insert(ctx context.Context, data *Car) (sql.Result, error) {
	now := time.Now()
	data.GmtCreate = &now
	data.GmtModified = &now

	query := fmt.Sprintf(
		"INSERT INTO %s (`address`,`vehicle_no`,`plate_color`,`seats`,`brand`,`model`,`vehicle_type`,`owner_name`,`vehicle_color`,`engine_id`,`vin`,`certify_date_a`,`fue_type`,`engine_displace`,`trans_agency`,`trans_area`,`trans_date_start`,`trans_date_end`,`certify_date_b`,`fix_state`,`next_fix_date`,`check_state`,`fee_print_id`,`gps_brand`,`gps_model`,`gps_install_date`,`register_date`,`commercial_type`,`fare_type`,`state`,`tid`,`trid`,`trname`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		m.table,
	)
	ret, err := m.conn.ExecCtx(ctx, query,
		data.Address, data.VehicleNo, data.PlateColor, data.Seats, data.Brand,
		data.Model, data.VehicleType, data.OwnerName, data.VehicleColor, data.EngineId,
		data.Vin, data.CertifyDateA, data.FueType, data.EngineDisplace, data.TransAgency,
		data.TransArea, data.TransDateStart, data.TransDateEnd, data.CertifyDateB, data.FixState,
		data.NextFixDate, data.CheckState, data.FeePrintId, data.GpsBrand, data.GpsModel,
		data.GpsInstallDate, data.RegisterDate, data.CommercialType, data.FareType, data.State,
		data.Tid, data.Trid, data.Trname,
	)
	return ret, err
}

func (m *defaultCarModel) FindOne(ctx context.Context, id int64) (*Car, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `id` = ? LIMIT 1", strings.Join(carFieldNames, ","), m.table)
	var resp Car
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultCarModel) FindByVehicleNo(ctx context.Context, vehicleNo string) (*Car, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `vehicle_no` = ? LIMIT 1", strings.Join(carFieldNames, ","), m.table)
	var resp Car
	err := m.conn.QueryRowCtx(ctx, &resp, query, vehicleNo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultCarModel) Update(ctx context.Context, data *Car) error {
	now := time.Now()
	data.GmtModified = &now

	query := fmt.Sprintf(
		"UPDATE %s SET `address`=?,`vehicle_no`=?,`plate_color`=?,`seats`=?,`brand`=?,`model`=?,`vehicle_type`=?,`owner_name`=?,`vehicle_color`=?,`engine_id`=?,`vin`=?,`certify_date_a`=?,`fue_type`=?,`engine_displace`=?,`trans_agency`=?,`trans_area`=?,`trans_date_start`=?,`trans_date_end`=?,`certify_date_b`=?,`fix_state`=?,`next_fix_date`=?,`check_state`=?,`fee_print_id`=?,`gps_brand`=?,`gps_model`=?,`gps_install_date`=?,`register_date`=?,`commercial_type`=?,`fare_type`=?,`state`=?,`tid`=?,`trid`=?,`trname`=?,`gmt_modified`=? WHERE `id` = ?",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query,
		data.Address, data.VehicleNo, data.PlateColor, data.Seats, data.Brand,
		data.Model, data.VehicleType, data.OwnerName, data.VehicleColor, data.EngineId,
		data.Vin, data.CertifyDateA, data.FueType, data.EngineDisplace, data.TransAgency,
		data.TransArea, data.TransDateStart, data.TransDateEnd, data.CertifyDateB, data.FixState,
		data.NextFixDate, data.CheckState, data.FeePrintId, data.GpsBrand, data.GpsModel,
		data.GpsInstallDate, data.RegisterDate, data.CommercialType, data.FareType, data.State,
		data.Tid, data.Trid, data.Trname, data.GmtModified, data.Id,
	)
	return err
}

func (m *defaultCarModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}
