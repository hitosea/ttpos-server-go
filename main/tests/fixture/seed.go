package fixture

import (
	"database/sql"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
)

// StaffOptions contains options for seeding a staff member.
type StaffOptions struct {
	UUID        int64
	CompanyUUID int64
	RealName    string
	Username    string
	Password    string
	Phone       string
	IsSuper     int
	DutyNo      string
	BindKey     string // device serial number bound to this cashier (bind_key column)
	Status      int
}

// SeedStaff creates a staff member in the database.
func SeedStaff(tb testing.TB, db *sql.DB, opts ...func(*StaffOptions)) StaffOptions {
	tb.Helper()

	// Default values
	opt := StaffOptions{
		UUID:        generateSnowflakeID(),
		CompanyUUID: 1234567890,
		RealName:    "Test Staff",
		Username:    "teststaff",
		Password:    "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZRGdjGj/n3.iW8jCPTu3QPTu3QPTu", // "password"
		Phone:       "0812345678",
		IsSuper:     0,
		DutyNo:      "",
		Status:      1,
	}

	// Apply options
	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_staff (uuid, company_uuid, real_name, username, password, phone, is_super, duty_no, bind_key, create_time, update_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, opt.UUID, opt.CompanyUUID, opt.RealName, opt.Username, opt.Password, opt.Phone, opt.IsSuper, opt.DutyNo, opt.BindKey, now, now)

	if err != nil {
		tb.Fatalf("failed to seed staff: %v", err)
	}

	return opt
}

// WithStaffUUID sets a custom UUID for the staff member.
func WithStaffUUID(uuid int64) func(*StaffOptions) {
	return func(o *StaffOptions) {
		o.UUID = uuid
	}
}

// WithStaffCompanyUUID sets the company UUID for the staff member.
func WithStaffCompanyUUID(companyUUID int64) func(*StaffOptions) {
	return func(o *StaffOptions) {
		o.CompanyUUID = companyUUID
	}
}

// WithStaffName sets the real_name for the staff member.
func WithStaffName(name string) func(*StaffOptions) {
	return func(o *StaffOptions) {
		o.RealName = name
	}
}

// WithStaffDutyNo sets the duty_no for the staff member (non-empty means on-duty).
func WithStaffDutyNo(dutyNo string) func(*StaffOptions) {
	return func(o *StaffOptions) {
		o.DutyNo = dutyNo
	}
}

// WithStaffIsSuper sets is_super for the staff member (1 = super admin, skips role/permission tables).
func WithStaffIsSuper(isSuper int) func(*StaffOptions) {
	return func(o *StaffOptions) {
		o.IsSuper = isSuper
	}
}

// WithStaffUsername sets the username for the staff member.
func WithStaffUsername(username string) func(*StaffOptions) {
	return func(o *StaffOptions) {
		o.Username = username
	}
}

// WithStaffBindKey sets the bind_key (device serial number) for the staff member.
// Required for SubmitShift: the service looks up the device by staff.BindKey.
func WithStaffBindKey(key string) func(*StaffOptions) {
	return func(o *StaffOptions) {
		o.BindKey = key
	}
}

// DeskOptions contains options for seeding a desk.
type DeskOptions struct {
	UUID        int64
	CompanyUUID int64
	Name        string
	AreaUUID    int64
	Status      int
}

// SeedDesk creates a desk in the database.
func SeedDesk(tb testing.TB, db *sql.DB, opts ...func(*DeskOptions)) DeskOptions {
	tb.Helper()

	opt := DeskOptions{
		UUID:        generateSnowflakeID(),
		CompanyUUID: 1234567890,
		Name:        "Table 1",
		AreaUUID:    0,
		Status:      1,
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_desk (uuid, company_uuid, name, area_uuid, status, create_time, update_time)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, opt.UUID, opt.CompanyUUID, opt.Name, opt.AreaUUID, opt.Status, now, now)

	if err != nil {
		tb.Fatalf("failed to seed desk: %v", err)
	}

	return opt
}

// WithDeskUUID sets a custom UUID for the desk.
func WithDeskUUID(uuid int64) func(*DeskOptions) {
	return func(o *DeskOptions) {
		o.UUID = uuid
	}
}

// WithDeskName sets the name for the desk.
func WithDeskName(name string) func(*DeskOptions) {
	return func(o *DeskOptions) {
		o.Name = name
	}
}

// DeviceOptions contains options for seeding a device.
type DeviceOptions struct {
	UUID        int64
	CompanyUUID int64
	Name        string
	DeviceType  string
	DeviceId    string
	Status      int
}

// DeviceId is the device serial number (device_id column), used for auth.
func (o *DeviceOptions) GetDeviceId() string {
	return o.DeviceId
}

// SeedDevice creates a device in the database.
func SeedDevice(tb testing.TB, db *sql.DB, opts ...func(*DeviceOptions)) DeviceOptions {
	tb.Helper()

	opt := DeviceOptions{
		UUID:        generateSnowflakeID(),
		CompanyUUID: 1234567890,
		Name:        "POS Terminal 1",
		DeviceType:  "cashier", // maps to the 'source' column
		DeviceId:    fmt.Sprintf("test-device-%d", time.Now().UnixNano()),
		Status:      1,
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_device (uuid, source, device_id, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.DeviceType, opt.DeviceId, now, now)

	if err != nil {
		tb.Fatalf("failed to seed device: %v", err)
	}

	return opt
}

// WithDeviceUUID sets a custom UUID for the device.
func WithDeviceUUID(uuid int64) func(*DeviceOptions) {
	return func(o *DeviceOptions) {
		o.UUID = uuid
	}
}

// WithDeviceType sets the device source type (cashier, tablet, kitchen, etc.).
func WithDeviceType(deviceType string) func(*DeviceOptions) {
	return func(o *DeviceOptions) {
		o.DeviceType = deviceType
	}
}

// WithDeviceId sets the device serial number used for JWT device_id.
func WithDeviceId(id string) func(*DeviceOptions) {
	return func(o *DeviceOptions) {
		o.DeviceId = id
	}
}

// CompanyOptions contains options for seeding a company.
type CompanyOptions struct {
	UUID        int64
	Name        string
	Status      int
	ExpireTime  int64
	IsEnableErp int
}

// SeedCompany creates a company record in the tenant database.
// status=1 and expire_time=0 are required for auth to pass.
func SeedCompany(tb testing.TB, db *sql.DB, opts ...func(*CompanyOptions)) CompanyOptions {
	tb.Helper()

	opt := CompanyOptions{
		UUID:        generateSnowflakeID(),
		Name:        "Test Company",
		Status:      1,
		ExpireTime:  0, // 0 = never expires
		IsEnableErp: 0,
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_company (uuid, name, status, expire_time, is_enable_erp, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.Name, opt.Status, opt.ExpireTime, opt.IsEnableErp, now, now)

	if err != nil {
		tb.Fatalf("failed to seed company: %v", err)
	}

	return opt
}

// WithCompanyUUID sets the UUID for the company.
func WithCompanyUUID(uuid int64) func(*CompanyOptions) {
	return func(o *CompanyOptions) {
		o.UUID = uuid
	}
}

// WithCompanyIsEnableErp enables or disables ERP for the company (1=enabled, 0=disabled).
func WithCompanyIsEnableErp(v int) func(*CompanyOptions) {
	return func(o *CompanyOptions) {
		o.IsEnableErp = v
	}
}

// CompanySettingOptions contains options for seeding company settings.
type CompanySettingOptions struct {
	UUID                  int64
	CompanyUUID           int64
	ErpnextSiteCode       string
	ErpnextPosProfileName string
	ErpnextAdminEmail     string
	ErpnextCompanyAbbr    string
	ErpnextBranchName     string
	EnableDataManagement  int
}

// SeedCompanySetting creates a company_setting record in the tenant database.
// The uuid must be non-zero for auth to pass (companySetting.Uuid == 0 check).
func SeedCompanySetting(tb testing.TB, db *sql.DB, opts ...func(*CompanySettingOptions)) CompanySettingOptions {
	tb.Helper()

	opt := CompanySettingOptions{
		UUID:        generateSnowflakeID(),
		CompanyUUID: 1234567890,
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_company_setting (uuid, company_uuid, erpnext_site_code, erpnext_pos_profile_name, erpnext_admin_email, erpnext_company_abbr, erpnext_branch_name, enable_data_management, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.CompanyUUID, opt.ErpnextSiteCode, opt.ErpnextPosProfileName, opt.ErpnextAdminEmail, opt.ErpnextCompanyAbbr, opt.ErpnextBranchName, opt.EnableDataManagement, now, now)

	if err != nil {
		tb.Fatalf("failed to seed company_setting: %v", err)
	}

	return opt
}

// WithCompanySettingCompanyUUID sets the company UUID for the company setting.
func WithCompanySettingCompanyUUID(uuid int64) func(*CompanySettingOptions) {
	return func(o *CompanySettingOptions) {
		o.CompanyUUID = uuid
	}
}

// WithCompanySettingErpConfig sets all ERP-related fields in company_setting.
// Required for ERP-enabled tests: siteCode must be non-empty to trigger ERP flows.
func WithCompanySettingErpConfig(siteCode, posProfileName, adminEmail, companyAbbr, branchName string) func(*CompanySettingOptions) {
	return func(o *CompanySettingOptions) {
		o.ErpnextSiteCode = siteCode
		o.ErpnextPosProfileName = posProfileName
		o.ErpnextAdminEmail = adminEmail
		o.ErpnextCompanyAbbr = companyAbbr
		o.ErpnextBranchName = branchName
	}
}

// WithCompanySettingEnableDataManagement enables the data management feature (1=enabled, 0=disabled).
// Required for SetDataManage and GetDataManage endpoints to work.
func WithCompanySettingEnableDataManagement(v int) func(*CompanySettingOptions) {
	return func(o *CompanySettingOptions) {
		o.EnableDataManagement = v
	}
}

// ProductOptions contains options for seeding a product.
type ProductOptions struct {
	UUID         int64
	CompanyUUID  int64
	Name         string
	Price        string
	CategoryUUID int64
	Status       int
}

// SeedProduct creates a product in the database.
func SeedProduct(tb testing.TB, db *sql.DB, opts ...func(*ProductOptions)) ProductOptions {
	tb.Helper()

	opt := ProductOptions{
		UUID:         generateSnowflakeID(),
		CompanyUUID:  1234567890,
		Name:         "Test Product",
		Price:        "100.00",
		CategoryUUID: 0,
		Status:       1,
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_product (uuid, company_uuid, name, price, category_uuid, status, create_time, update_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, opt.UUID, opt.CompanyUUID, opt.Name, opt.Price, opt.CategoryUUID, opt.Status, now, now)

	if err != nil {
		tb.Fatalf("failed to seed product: %v", err)
	}

	return opt
}

// WithProductUUID sets a custom UUID for the product.
func WithProductUUID(uuid int64) func(*ProductOptions) {
	return func(o *ProductOptions) {
		o.UUID = uuid
	}
}

// WithProductName sets the name for the product.
func WithProductName(name string) func(*ProductOptions) {
	return func(o *ProductOptions) {
		o.Name = name
	}
}

// WithProductPrice sets the price for the product.
func WithProductPrice(price string) func(*ProductOptions) {
	return func(o *ProductOptions) {
		o.Price = price
	}
}

// ShiftLogOptions contains options for seeding a staff shift log.
type ShiftLogOptions struct {
	UUID             int64
	StaffUUID        int64
	ShiftNo          string
	Status           int
	ShiftStartTime   int64
	ErpOpenEntryName string // erpnext_open_pos_entry_name: non-empty triggers ClosePosEntry on shift close
}

// SeedShiftLog creates a staff_shift_log record in the tenant database.
// Status=0 means the shift is active (not handed over), which is required for GetShiftInfo.
func SeedShiftLog(tb testing.TB, db *sql.DB, opts ...func(*ShiftLogOptions)) ShiftLogOptions {
	tb.Helper()

	opt := ShiftLogOptions{
		UUID:           generateSnowflakeID(),
		StaffUUID:      0,
		ShiftNo:        "DUTY-TEST-001",
		Status:         0, // 0 = active (not handed over)
		ShiftStartTime: time.Now().Unix(),
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_staff_shift_log (uuid, staff_uuid, shift_no, status, shift_start_time, erpnext_open_pos_entry_name, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.StaffUUID, opt.ShiftNo, opt.Status, opt.ShiftStartTime, opt.ErpOpenEntryName, now, now)

	if err != nil {
		tb.Fatalf("failed to seed shift log: %v", err)
	}

	return opt
}

// WithShiftLogStaffUUID sets the staff UUID for the shift log.
func WithShiftLogStaffUUID(uuid int64) func(*ShiftLogOptions) {
	return func(o *ShiftLogOptions) {
		o.StaffUUID = uuid
	}
}

// WithShiftLogShiftNo sets the shift_no for the shift log (must match staff.duty_no).
func WithShiftLogShiftNo(shiftNo string) func(*ShiftLogOptions) {
	return func(o *ShiftLogOptions) {
		o.ShiftNo = shiftNo
	}
}

// WithShiftLogStatus sets the status for the shift log (0=active, 1=handed over).
func WithShiftLogStatus(status int) func(*ShiftLogOptions) {
	return func(o *ShiftLogOptions) {
		o.Status = status
	}
}

// WithShiftLogErpOpenEntryName sets the erpnext_open_pos_entry_name for the shift log.
// A non-empty value is required for the ERP ClosePosEntry gRPC call to be triggered on shift close.
func WithShiftLogErpOpenEntryName(name string) func(*ShiftLogOptions) {
	return func(o *ShiftLogOptions) {
		o.ErpOpenEntryName = name
	}
}

// PaymentMethodOptions contains options for seeding a payment method.
type PaymentMethodOptions struct {
	UUID              int64
	Name              string
	Code              int
	Source            int    // 0=system, 1=manual, 2=LianLianPay
	Status            int    // 1=enabled
	IsShowCashier     int    // 1=visible on cashier terminal
	ErpnextPayment    string // ERPNext payment name (fallback mode_of_payment)
	ErpnextPaymentId  string // ERPNext payment ID (preferred over ErpnextPayment)
}

// SeedPaymentMethod creates a payment_method record in the tenant database.
func SeedPaymentMethod(tb testing.TB, db *sql.DB, opts ...func(*PaymentMethodOptions)) PaymentMethodOptions {
	tb.Helper()

	opt := PaymentMethodOptions{
		UUID:          generateSnowflakeID(),
		Name:          "Cash",
		Code:          40, // PaymentMethodCodeCash
		Source:        0,  // PaymentMethodSourceSystem
		Status:        1,  // PaymentMethodStatusEnable
		IsShowCashier: 1,  // visible on cashier by default
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_payment_method (uuid, name, code, source, status, is_show_cashier, erpnext_payment, erpnext_payment_id, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.Name, opt.Code, opt.Source, opt.Status, opt.IsShowCashier, opt.ErpnextPayment, opt.ErpnextPaymentId, now, now)

	if err != nil {
		tb.Fatalf("failed to seed payment method: %v", err)
	}

	return opt
}

// WithPaymentMethodUUID sets the UUID for the payment method.
func WithPaymentMethodUUID(uuid int64) func(*PaymentMethodOptions) {
	return func(o *PaymentMethodOptions) {
		o.UUID = uuid
	}
}

// WithPaymentMethodCode sets the code for the payment method.
func WithPaymentMethodCode(code int) func(*PaymentMethodOptions) {
	return func(o *PaymentMethodOptions) {
		o.Code = code
	}
}

// WithPaymentMethodSource sets the source for the payment method (0=system, 1=manual).
func WithPaymentMethodSource(source int) func(*PaymentMethodOptions) {
	return func(o *PaymentMethodOptions) {
		o.Source = source
	}
}

// WithPaymentMethodErpnextPayment sets the ERPNext payment name (used as mode_of_payment fallback).
func WithPaymentMethodErpnextPayment(name string) func(*PaymentMethodOptions) {
	return func(o *PaymentMethodOptions) {
		o.ErpnextPayment = name
	}
}

// WithPaymentMethodName sets the display name for the payment method.
func WithPaymentMethodName(name string) func(*PaymentMethodOptions) {
	return func(o *PaymentMethodOptions) {
		o.Name = name
	}
}

// WithPaymentMethodStatus sets the status for the payment method (0=disabled, 1=enabled, 2=draft).
func WithPaymentMethodStatus(status int) func(*PaymentMethodOptions) {
	return func(o *PaymentMethodOptions) {
		o.Status = status
	}
}

// WithPaymentMethodIsShowCashier sets whether the payment method is visible on the cashier terminal.
func WithPaymentMethodIsShowCashier(v int) func(*PaymentMethodOptions) {
	return func(o *PaymentMethodOptions) {
		o.IsShowCashier = v
	}
}

// WithPaymentMethodErpnextPaymentId sets the ERPNext payment ID (preferred over payment name).
func WithPaymentMethodErpnextPaymentId(id string) func(*PaymentMethodOptions) {
	return func(o *PaymentMethodOptions) {
		o.ErpnextPaymentId = id
	}
}

// OrderItemRemarkOptions contains options for seeding an order item remark.
type OrderItemRemarkOptions struct {
	UUID int64
	Name string
}

// SeedOrderItemRemark creates an order_item_remark record in the tenant database.
// Used to pre-populate remarks, e.g., for testing the 100-item limit.
func SeedOrderItemRemark(tb testing.TB, db *sql.DB, opts ...func(*OrderItemRemarkOptions)) OrderItemRemarkOptions {
	tb.Helper()

	opt := OrderItemRemarkOptions{
		UUID: generateSnowflakeID(),
		Name: `{"zh_name":"备注","en_name":"Remark"}`,
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_order_item_remark (uuid, name, multi_language_name_uuid, create_time, update_time, delete_time)
		VALUES (?, ?, 0, ?, ?, 0)
	`, opt.UUID, opt.Name, now, now)

	if err != nil {
		tb.Fatalf("failed to seed order item remark: %v", err)
	}

	return opt
}

// WithOrderItemRemarkUUID sets a custom UUID for the order item remark.
func WithOrderItemRemarkUUID(uuid int64) func(*OrderItemRemarkOptions) {
	return func(o *OrderItemRemarkOptions) {
		o.UUID = uuid
	}
}

// WithOrderItemRemarkName sets the name (JSON string) for the order item remark.
func WithOrderItemRemarkName(name string) func(*OrderItemRemarkOptions) {
	return func(o *OrderItemRemarkOptions) {
		o.Name = name
	}
}

// SaleBillOptions contains options for seeding a sale bill.
type SaleBillOptions struct {
	UUID       int64
	Status     int   // 1 = complete (SaleBillStatusComplete)
	FinishTime int64 // Unix timestamp; should fall within query range
}

// SeedSaleBill creates a sale_bill record in the tenant database.
// Only the essential fields are set; all other columns default to 0/empty in MySQL.
func SeedSaleBill(tb testing.TB, db *sql.DB, opts ...func(*SaleBillOptions)) SaleBillOptions {
	tb.Helper()

	opt := SaleBillOptions{
		UUID:       generateSnowflakeID(),
		Status:     1,          // SaleBillStatusComplete
		FinishTime: 1750000000, // 2025-06-15, within the standard test query range
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_sale_bill (uuid, status, finish_time, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.Status, opt.FinishTime, now, now)

	if err != nil {
		tb.Fatalf("failed to seed sale bill: %v", err)
	}

	return opt
}

// WithSaleBillUUID sets a custom UUID for the sale bill.
func WithSaleBillUUID(uuid int64) func(*SaleBillOptions) {
	return func(o *SaleBillOptions) {
		o.UUID = uuid
	}
}

// WithSaleBillStatus sets the status for the sale bill.
func WithSaleBillStatus(status int) func(*SaleBillOptions) {
	return func(o *SaleBillOptions) {
		o.Status = status
	}
}

// WithSaleBillFinishTime sets the finish_time for the sale bill (Unix timestamp).
func WithSaleBillFinishTime(ts int64) func(*SaleBillOptions) {
	return func(o *SaleBillOptions) {
		o.FinishTime = ts
	}
}

// snowflakeCounter ensures unique IDs even when called in tight loops within the same millisecond.
var snowflakeCounter atomic.Int64

// generateSnowflakeID generates a simple unique ID for testing.
// No modulo: the counter is never reset, so IDs are guaranteed unique across all seed calls
// within a process regardless of how many seeds occur per millisecond.
func generateSnowflakeID() int64 {
	return time.Now().UnixMilli()<<16 | (snowflakeCounter.Add(1) & 0xFFFF)
}

// GenerateUUID generates a UUID string for testing.
func GenerateUUID() string {
	return uuid.New().String()
}

// GenerateCompanyUUID generates a company UUID for testing.
func GenerateCompanyUUID(tb testing.TB) string {
	tb.Helper()
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// PaymentOrderOptions contains options for seeding a payment order.
type PaymentOrderOptions struct {
	UUID              int64
	PaymentMethodUUID int64
	PaymentMethodName string
	PaymentAmount     float64
	Status            int // 1=paid, 0=unpaid
	RelatedUUID       int64 // sale_order_uuid
	RelatedType       int   // 0=sale_order
}

// SeedPaymentOrder creates a payment_order record in the tenant database.
func SeedPaymentOrder(tb testing.TB, db *sql.DB, opts ...func(*PaymentOrderOptions)) PaymentOrderOptions {
	tb.Helper()

	opt := PaymentOrderOptions{
		UUID:              generateSnowflakeID(),
		PaymentMethodUUID: 0,          // caller must set via WithPaymentOrderPaymentMethodUUID
		PaymentMethodName: "Cash",
		PaymentAmount:     100.0,
		Status:            1,          // paid
		RelatedType:       0,          // sale_order
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_payment_order
		(uuid, payment_method_uuid, payment_method_name, payment_amount, amount,
		 status, related_uuid, related_type, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.PaymentMethodUUID, opt.PaymentMethodName, opt.PaymentAmount, opt.PaymentAmount,
		opt.Status, opt.RelatedUUID, opt.RelatedType, now, now)

	if err != nil {
		tb.Fatalf("failed to seed payment_order: %v", err)
	}

	return opt
}

// WithPaymentOrderUUID sets the UUID for the payment order.
func WithPaymentOrderUUID(uuid int64) func(*PaymentOrderOptions) {
	return func(o *PaymentOrderOptions) {
		o.UUID = uuid
	}
}

// WithPaymentOrderPaymentMethodUUID sets the payment method UUID for the payment order.
func WithPaymentOrderPaymentMethodUUID(paymentMethodUUID int64) func(*PaymentOrderOptions) {
	return func(o *PaymentOrderOptions) {
		o.PaymentMethodUUID = paymentMethodUUID
	}
}

// WithPaymentOrderPaymentMethodName sets the payment method name for the payment order.
func WithPaymentOrderPaymentMethodName(name string) func(*PaymentOrderOptions) {
	return func(o *PaymentOrderOptions) {
		o.PaymentMethodName = name
	}
}

// WithPaymentOrderAmount sets the payment amount for the payment order.
func WithPaymentOrderAmount(amount float64) func(*PaymentOrderOptions) {
	return func(o *PaymentOrderOptions) {
		o.PaymentAmount = amount
	}
}

// WithPaymentOrderStatus sets the status for the payment order (1=paid, 0=unpaid).
func WithPaymentOrderStatus(status int) func(*PaymentOrderOptions) {
	return func(o *PaymentOrderOptions) {
		o.Status = status
	}
}

// WithPaymentOrderRelatedUUID sets the related UUID (sale_order_uuid) for the payment order.
func WithPaymentOrderRelatedUUID(relatedUUID int64) func(*PaymentOrderOptions) {
	return func(o *PaymentOrderOptions) {
		o.RelatedUUID = relatedUUID
	}
}
