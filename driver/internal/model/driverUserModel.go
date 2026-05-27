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
		FindDriverDetail(ctx context.Context, driverId int64) (*DriverUser, *DriverCarBinding, *Car, error)
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
			return nil, ErrNotFound
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
			return nil, ErrNotFound
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
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &resp, nil
}

// driverDetailRow 用于 JOIN 查询的结果扫描
type driverDetailRow struct {
	// driver_user 字段
	DuId                          int64      `db:"du_id"`
	DuAddress                     string     `db:"du_address"`
	DuDriverName                  string     `db:"du_driver_name"`
	DuDriverPhone                 string     `db:"du_driver_phone"`
	DuDriverGender                *int64     `db:"du_driver_gender"`
	DuDriverBirthday              *time.Time `db:"du_driver_birthday"`
	DuDriverNation                string     `db:"du_driver_nation"`
	DuDriverContactAddress        string     `db:"du_driver_contact_address"`
	DuLicenseId                   string     `db:"du_license_id"`
	DuGetDriverLicenseDate        *time.Time `db:"du_get_driver_license_date"`
	DuDriverLicenseOn             *time.Time `db:"du_driver_license_on"`
	DuDriverLicenseOff            *time.Time `db:"du_driver_license_off"`
	DuTaxiDriver                  *int64     `db:"du_taxi_driver"`
	DuCertificateNo               string     `db:"du_certificate_no"`
	DuNetworkCarIssueOrganization string     `db:"du_network_car_issue_organization"`
	DuNetworkCarIssueDate         *time.Time `db:"du_network_car_issue_date"`
	DuGetNetworkCarProofDate      *time.Time `db:"du_get_network_car_proof_date"`
	DuNetworkCarProofOn           *time.Time `db:"du_network_car_proof_on"`
	DuNetworkCarProofOff          *time.Time `db:"du_network_car_proof_off"`
	DuRegisterDate                *time.Time `db:"du_register_date"`
	DuCommercialType              *int64     `db:"du_commercial_type"`
	DuContractCompany             string     `db:"du_contract_company"`
	DuContractOn                  *time.Time `db:"du_contract_on"`
	DuContractOff                 *time.Time `db:"du_contract_off"`
	DuState                       *int64     `db:"du_state"`
	DuGmtCreate                   *time.Time `db:"du_gmt_create"`
	DuGmtModified                 *time.Time `db:"du_gmt_modified"`
	// driver_car_binding 字段
	DcbId            int64      `db:"dcb_id"`
	DcbDriverId      *int64     `db:"dcb_driver_id"`
	DcbCarId         *int64     `db:"dcb_car_id"`
	DcbBindState     *int       `db:"dcb_bind_state"`
	DcbBindingTime   *time.Time `db:"dcb_binding_time"`
	DcbUnBindingTime *time.Time `db:"dcb_un_binding_time"`
	// car 字段
	CId             int64      `db:"c_id"`
	CAddress        string     `db:"c_address"`
	CVehicleNo      string     `db:"c_vehicle_no"`
	CPlateColor     *string    `db:"c_plate_color"`
	CSeats          *int       `db:"c_seats"`
	CBrand          *string    `db:"c_brand"`
	CModel          *string    `db:"c_model"`
	CVehicleType    *string    `db:"c_vehicle_type"`
	COwnerName      *string    `db:"c_owner_name"`
	CVehicleColor   *string    `db:"c_vehicle_color"`
	CEngineId       *string    `db:"c_engine_id"`
	CVin            *string    `db:"c_vin"`
	CCertifyDateA   *time.Time `db:"c_certify_date_a"`
	CFuelType       *string    `db:"c_fue_type"`
	CEngineDisplace *string    `db:"c_engine_displace"`
	CTransAgency    *string    `db:"c_trans_agency"`
	CTransArea      *string    `db:"c_trans_area"`
	CTransDateStart *time.Time `db:"c_trans_date_start"`
	CTransDateEnd   *time.Time `db:"c_trans_date_end"`
	CCertifyDateB   *time.Time `db:"c_certify_date_b"`
	CFixState       *string    `db:"c_fix_state"`
	CNextFixDate    *time.Time `db:"c_next_fix_date"`
	CCheckState     *string    `db:"c_check_state"`
	CFeePrintId     *string    `db:"c_fee_print_id"`
	CGpsBrand       *string    `db:"c_gps_brand"`
	CGpsModel       *string    `db:"c_gps_model"`
	CGpsInstallDate *time.Time `db:"c_gps_install_date"`
	CRegisterDate   *time.Time `db:"c_register_date"`
	CCommercialType *int       `db:"c_commercial_type"`
	CFareType       *string    `db:"c_fare_type"`
	CState          *int8      `db:"c_state"`
	CTid            *string    `db:"c_tid"`
	CTrid           *string    `db:"c_trid"`
	CTrname         *string    `db:"c_trname"`
	CGmtCreate      *time.Time `db:"c_gmt_create"`
	CGmtModified    *time.Time `db:"c_gmt_modified"`
}

// FindDriverDetail 使用 JOIN 一次性获取司机、车辆绑定和车辆信息
func (m *defaultDriverUserModel) FindDriverDetail(ctx context.Context, driverId int64) (*DriverUser, *DriverCarBinding, *Car, error) {
	query := `SELECT
		du.id AS du_id, du.address AS du_address, du.driver_name AS du_driver_name, du.driver_phone AS du_driver_phone,
		du.driver_gender AS du_driver_gender, du.driver_birthday AS du_driver_birthday, du.driver_nation AS du_driver_nation,
		du.driver_contact_address AS du_driver_contact_address, du.license_id AS du_license_id,
		du.get_driver_license_date AS du_get_driver_license_date, du.driver_license_on AS du_driver_license_on,
		du.driver_license_off AS du_driver_license_off, du.taxi_driver AS du_taxi_driver,
		du.certificate_no AS du_certificate_no, du.network_car_issue_organization AS du_network_car_issue_organization,
		du.network_car_issue_date AS du_network_car_issue_date, du.get_network_car_proof_date AS du_get_network_car_proof_date,
		du.network_car_proof_on AS du_network_car_proof_on, du.network_car_proof_off AS du_network_car_proof_off,
		du.register_date AS du_register_date, du.commercial_type AS du_commercial_type,
		du.contract_company AS du_contract_company, du.contract_on AS du_contract_on, du.contract_off AS du_contract_off,
		du.state AS du_state, du.gmt_create AS du_gmt_create, du.gmt_modified AS du_gmt_modified,
		dcb.id AS dcb_id, dcb.driver_id AS dcb_driver_id, dcb.car_id AS dcb_car_id,
		dcb.bind_state AS dcb_bind_state, dcb.binding_time AS dcb_binding_time, dcb.un_binding_time AS dcb_un_binding_time,
		c.id AS c_id, c.address AS c_address, c.vehicle_no AS c_vehicle_no, c.plate_color AS c_plate_color,
		c.seats AS c_seats, c.brand AS c_brand, c.model AS c_model, c.vehicle_type AS c_vehicle_type,
		c.owner_name AS c_owner_name, c.vehicle_color AS c_vehicle_color, c.engine_id AS c_engine_id, c.vin AS c_vin,
		c.certify_date_a AS c_certify_date_a, c.fue_type AS c_fue_type, c.engine_displace AS c_engine_displace,
		c.trans_agency AS c_trans_agency, c.trans_area AS c_trans_area, c.trans_date_start AS c_trans_date_start,
		c.trans_date_end AS c_trans_date_end, c.certify_date_b AS c_certify_date_b, c.fix_state AS c_fix_state,
		c.next_fix_date AS c_next_fix_date, c.check_state AS c_check_state, c.fee_print_id AS c_fee_print_id,
		c.gps_brand AS c_gps_brand, c.gps_model AS c_gps_model, c.gps_install_date AS c_gps_install_date,
		c.register_date AS c_register_date, c.commercial_type AS c_commercial_type, c.fare_type AS c_fare_type,
		c.state AS c_state, c.tid AS c_tid, c.trid AS c_trid, c.trname AS c_trname,
		c.gmt_create AS c_gmt_create, c.gmt_modified AS c_gmt_modified
	FROM ` + m.table + ` du
	LEFT JOIN ` + "`driver_car_binding_relationship`" + ` dcb ON dcb.driver_id = du.id AND dcb.bind_state = 1
		AND dcb.id = (SELECT MAX(dcb2.id) FROM driver_car_binding_relationship dcb2 WHERE dcb2.driver_id = du.id AND dcb2.bind_state = 1)
	LEFT JOIN ` + "`car`" + ` c ON c.id = dcb.car_id
	WHERE du.id = ?`

	var row driverDetailRow
	err := m.conn.QueryRowCtx(ctx, &row, query, driverId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil, ErrNotFound
		}
		return nil, nil, nil, err
	}

	driverUser := &DriverUser{
		Id:                          row.DuId,
		Address:                     row.DuAddress,
		DriverName:                  row.DuDriverName,
		DriverPhone:                 row.DuDriverPhone,
		DriverGender:                row.DuDriverGender,
		DriverBirthday:              row.DuDriverBirthday,
		DriverNation:                row.DuDriverNation,
		DriverContactAddress:        row.DuDriverContactAddress,
		LicenseId:                   row.DuLicenseId,
		GetDriverLicenseDate:        row.DuGetDriverLicenseDate,
		DriverLicenseOn:             row.DuDriverLicenseOn,
		DriverLicenseOff:            row.DuDriverLicenseOff,
		TaxiDriver:                  row.DuTaxiDriver,
		CertificateNo:               row.DuCertificateNo,
		NetworkCarIssueOrganization: row.DuNetworkCarIssueOrganization,
		NetworkCarIssueDate:         row.DuNetworkCarIssueDate,
		GetNetworkCarProofDate:      row.DuGetNetworkCarProofDate,
		NetworkCarProofOn:           row.DuNetworkCarProofOn,
		NetworkCarProofOff:          row.DuNetworkCarProofOff,
		RegisterDate:                row.DuRegisterDate,
		CommercialType:              row.DuCommercialType,
		ContractCompany:             row.DuContractCompany,
		ContractOn:                  row.DuContractOn,
		ContractOff:                 row.DuContractOff,
		State:                       row.DuState,
		GmtCreate:                   row.DuGmtCreate,
		GmtModified:                 row.DuGmtModified,
	}

	var binding *DriverCarBinding
	if row.DcbId != 0 {
		binding = &DriverCarBinding{
			Id:            row.DcbId,
			DriverId:      row.DcbDriverId,
			CarId:         row.DcbCarId,
			BindState:     row.DcbBindState,
			BindingTime:   row.DcbBindingTime,
			UnBindingTime: row.DcbUnBindingTime,
		}
	}

	var car *Car
	if row.CId != 0 {
		car = &Car{
			Id:             row.CId,
			Address:        row.CAddress,
			VehicleNo:      row.CVehicleNo,
			PlateColor:     row.CPlateColor,
			Seats:          row.CSeats,
			Brand:          row.CBrand,
			Model:          row.CModel,
			VehicleType:    row.CVehicleType,
			OwnerName:      row.COwnerName,
			VehicleColor:   row.CVehicleColor,
			EngineId:       row.CEngineId,
			Vin:            row.CVin,
			CertifyDateA:   row.CCertifyDateA,
			FuelType:       row.CFuelType,
			EngineDisplace: row.CEngineDisplace,
			TransAgency:    row.CTransAgency,
			TransArea:      row.CTransArea,
			TransDateStart: row.CTransDateStart,
			TransDateEnd:   row.CTransDateEnd,
			CertifyDateB:   row.CCertifyDateB,
			FixState:       row.CFixState,
			NextFixDate:    row.CNextFixDate,
			CheckState:     row.CCheckState,
			FeePrintId:     row.CFeePrintId,
			GpsBrand:       row.CGpsBrand,
			GpsModel:       row.CGpsModel,
			GpsInstallDate: row.CGpsInstallDate,
			RegisterDate:   row.CRegisterDate,
			CommercialType: row.CCommercialType,
			FareType:       row.CFareType,
			State:          row.CState,
			Tid:            row.CTid,
			Trid:           row.CTrid,
			Trname:         row.CTrname,
			GmtCreate:      row.CGmtCreate,
			GmtModified:    row.CGmtModified,
		}
	}

	return driverUser, binding, car, nil
}
