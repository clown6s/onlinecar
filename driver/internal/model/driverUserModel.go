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
	driverUserFieldNames = []string{
		"id", "address", "driver_name", "driver_phone", "driver_gender",
		"driver_birthday", "driver_nation", "driver_contact_address", "license_id",
		"get_driver_license_date", "driver_license_on", "driver_license_off",
		"taxi_driver", "certificate_no", "network_car_issue_organization",
		"network_car_issue_date", "get_network_car_proof_date", "network_car_proof_on",
		"network_car_proof_off", "register_date", "commercial_type",
		"contract_company", "contract_on", "contract_off", "state",
		"gmt_create", "gmt_modified",
	}
)

type (
	DriverUserModel interface {
		Insert(ctx context.Context, data *DriverUser) (sql.Result, error)
		InsertWithConn(ctx context.Context, session sqlx.Session, data *DriverUser) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*DriverUser, error)
		FindByPhone(ctx context.Context, phone string) (*DriverUser, error)
		FindByPhoneWithConn(ctx context.Context, session sqlx.Session, phone string) (*DriverUser, error)
		Update(ctx context.Context, data *DriverUser) error
		Delete(ctx context.Context, id int64) error
	}

	defaultDriverUserModel struct {
		conn  sqlx.SqlConn
		table string
	}

	DriverUser struct {
		Id                          int64      `db:"id"`
		Address                     string     `db:"address"`
		DriverName                  string     `db:"driver_name"`
		DriverPhone                 string     `db:"driver_phone"`
		DriverGender                *int64     `db:"driver_gender"`
		DriverBirthday              *time.Time `db:"driver_birthday"`
		DriverNation                string     `db:"driver_nation"`
		DriverContactAddress        string     `db:"driver_contact_address"`
		LicenseId                   string     `db:"license_id"`
		GetDriverLicenseDate        *time.Time `db:"get_driver_license_date"`
		DriverLicenseOn             *time.Time `db:"driver_license_on"`
		DriverLicenseOff            *time.Time `db:"driver_license_off"`
		TaxiDriver                  *int64     `db:"taxi_driver"`
		CertificateNo               string     `db:"certificate_no"`
		NetworkCarIssueOrganization string     `db:"network_car_issue_organization"`
		NetworkCarIssueDate         *time.Time `db:"network_car_issue_date"`
		GetNetworkCarProofDate      *time.Time `db:"get_network_car_proof_date"`
		NetworkCarProofOn           *time.Time `db:"network_car_proof_on"`
		NetworkCarProofOff          *time.Time `db:"network_car_proof_off"`
		RegisterDate                *time.Time `db:"register_date"`
		CommercialType              *int64     `db:"commercial_type"`
		ContractCompany             string     `db:"contract_company"`
		ContractOn                  *time.Time `db:"contract_on"`
		ContractOff                 *time.Time `db:"contract_off"`
		State                       *int64     `db:"state"`
		GmtCreate                   *time.Time `db:"gmt_create"`
		GmtModified                 *time.Time `db:"gmt_modified"`
	}
)

func NewDriverUserModel(conn sqlx.SqlConn) DriverUserModel {
	return &defaultDriverUserModel{
		conn:  conn,
		table: "`driver_user`",
	}
}

func (m *defaultDriverUserModel) Insert(ctx context.Context, data *DriverUser) (sql.Result, error) {
	now := time.Now()
	data.GmtCreate = &now
	data.GmtModified = &now

	// insertable fields (exclude id, gmt_create, gmt_modified)
	query := fmt.Sprintf(
		"INSERT INTO %s (`address`,`driver_name`,`driver_phone`,`driver_gender`,`driver_birthday`,`driver_nation`,`driver_contact_address`,`license_id`,`get_driver_license_date`,`driver_license_on`,`driver_license_off`,`taxi_driver`,`certificate_no`,`network_car_issue_organization`,`network_car_issue_date`,`get_network_car_proof_date`,`network_car_proof_on`,`network_car_proof_off`,`register_date`,`commercial_type`,`contract_company`,`contract_on`,`contract_off`,`state`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		m.table,
	)
	ret, err := m.conn.ExecCtx(ctx, query,
		data.Address, data.DriverName, data.DriverPhone, data.DriverGender, data.DriverBirthday,
		data.DriverNation, data.DriverContactAddress, data.LicenseId, data.GetDriverLicenseDate,
		data.DriverLicenseOn, data.DriverLicenseOff, data.TaxiDriver, data.CertificateNo,
		data.NetworkCarIssueOrganization, data.NetworkCarIssueDate, data.GetNetworkCarProofDate,
		data.NetworkCarProofOn, data.NetworkCarProofOff, data.RegisterDate, data.CommercialType,
		data.ContractCompany, data.ContractOn, data.ContractOff, data.State,
	)
	return ret, err
}

func (m *defaultDriverUserModel) FindOne(ctx context.Context, id int64) (*DriverUser, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `id` = ? LIMIT 1", strings.Join(driverUserFieldNames, ","), m.table)
	var resp DriverUser
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultDriverUserModel) FindByPhone(ctx context.Context, phone string) (*DriverUser, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `driver_phone` = ? LIMIT 1", strings.Join(driverUserFieldNames, ","), m.table)
	var resp DriverUser
	err := m.conn.QueryRowCtx(ctx, &resp, query, phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &resp, nil
}

func (m *defaultDriverUserModel) Update(ctx context.Context, data *DriverUser) error {
	now := time.Now()
	data.GmtModified = &now

	query := fmt.Sprintf(
		"UPDATE %s SET `address`=?,`driver_name`=?,`driver_phone`=?,`driver_gender`=?,`driver_birthday`=?,`driver_nation`=?,`driver_contact_address`=?,`license_id`=?,`get_driver_license_date`=?,`driver_license_on`=?,`driver_license_off`=?,`taxi_driver`=?,`certificate_no`=?,`network_car_issue_organization`=?,`network_car_issue_date`=?,`get_network_car_proof_date`=?,`network_car_proof_on`=?,`network_car_proof_off`=?,`register_date`=?,`commercial_type`=?,`contract_company`=?,`contract_on`=?,`contract_off`=?,`state`=?,`gmt_modified`=? WHERE `id` = ?",
		m.table,
	)
	_, err := m.conn.ExecCtx(ctx, query,
		data.Address, data.DriverName, data.DriverPhone, data.DriverGender, data.DriverBirthday,
		data.DriverNation, data.DriverContactAddress, data.LicenseId, data.GetDriverLicenseDate,
		data.DriverLicenseOn, data.DriverLicenseOff, data.TaxiDriver, data.CertificateNo,
		data.NetworkCarIssueOrganization, data.NetworkCarIssueDate, data.GetNetworkCarProofDate,
		data.NetworkCarProofOn, data.NetworkCarProofOff, data.RegisterDate, data.CommercialType,
		data.ContractCompany, data.ContractOn, data.ContractOff, data.State, data.GmtModified,
		data.Id,
	)
	return err
}

func (m *defaultDriverUserModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("DELETE FROM %s WHERE `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// InsertWithConn 使用指定 session 执行插入（事务内使用）
func (m *defaultDriverUserModel) InsertWithConn(ctx context.Context, session sqlx.Session, data *DriverUser) (sql.Result, error) {
	now := time.Now()
	data.GmtCreate = &now
	data.GmtModified = &now

	query := fmt.Sprintf(
		"INSERT INTO %s (`address`,`driver_name`,`driver_phone`,`driver_gender`,`driver_birthday`,`driver_nation`,`driver_contact_address`,`license_id`,`get_driver_license_date`,`driver_license_on`,`driver_license_off`,`taxi_driver`,`certificate_no`,`network_car_issue_organization`,`network_car_issue_date`,`get_network_car_proof_date`,`network_car_proof_on`,`network_car_proof_off`,`register_date`,`commercial_type`,`contract_company`,`contract_on`,`contract_off`,`state`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
		m.table,
	)
	ret, err := session.ExecCtx(ctx, query,
		data.Address, data.DriverName, data.DriverPhone, data.DriverGender, data.DriverBirthday,
		data.DriverNation, data.DriverContactAddress, data.LicenseId, data.GetDriverLicenseDate,
		data.DriverLicenseOn, data.DriverLicenseOff, data.TaxiDriver, data.CertificateNo,
		data.NetworkCarIssueOrganization, data.NetworkCarIssueDate, data.GetNetworkCarProofDate,
		data.NetworkCarProofOn, data.NetworkCarProofOff, data.RegisterDate, data.CommercialType,
		data.ContractCompany, data.ContractOn, data.ContractOff, data.State,
	)
	return ret, err
}

// FindByPhoneWithConn 使用指定 session 查询（事务内使用）
func (m *defaultDriverUserModel) FindByPhoneWithConn(ctx context.Context, session sqlx.Session, phone string) (*DriverUser, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `driver_phone` = ? LIMIT 1", strings.Join(driverUserFieldNames, ","), m.table)
	var resp DriverUser
	err := session.QueryRowCtx(ctx, &resp, query, phone)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &resp, nil
}
