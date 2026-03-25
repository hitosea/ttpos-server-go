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
	UUID       int64
	Name       string // maps to desk_no column
	AreaUUID   int64  // Deprecated: use RegionUUID for full-schema tests
	RegionUUID int64  // region_uuid column in full schema (shop_01.sql)
	Status     int
}

// SeedDesk creates a desk in the database.
// For full-schema (NewTestTenantFull) tests, use WithDeskRegionUUID to set the region.
func SeedDesk(tb testing.TB, db *sql.DB, opts ...func(*DeskOptions)) DeskOptions {
	tb.Helper()

	opt := DeskOptions{
		UUID:       generateSnowflakeID(),
		Name:       "Table 1",
		AreaUUID:   0,
		RegionUUID: 0,
		Status:     0, // 0=vacant (not opened)
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()

	// Try full-schema INSERT first (region_uuid + desk_no columns); fall back to minimal schema (area_uuid)
	_, err := db.Exec(`
		INSERT INTO ttpos_desk (uuid, desk_no, region_uuid, status, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.Name, opt.RegionUUID, opt.Status, now, now)

	if err != nil {
		// Fallback: minimal schema uses area_uuid and name columns
		_, err = db.Exec(`
			INSERT INTO ttpos_desk (uuid, name, area_uuid, status, create_time, update_time)
			VALUES (?, ?, ?, ?, ?, ?)
		`, opt.UUID, opt.Name, opt.AreaUUID, opt.Status, now, now)
		if err != nil {
			tb.Fatalf("failed to seed desk: %v", err)
		}
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

// WithDeskRegionUUID sets the region_uuid for the desk (full-schema only).
func WithDeskRegionUUID(uuid int64) func(*DeskOptions) {
	return func(o *DeskOptions) {
		o.RegionUUID = uuid
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
	UUID                     int64
	CompanyUUID              int64
	ErpnextSiteCode          string
	ErpnextPosProfileName    string
	ErpnextAdminEmail        string
	ErpnextCompanyAbbr       string
	ErpnextHeadquarterAbbr   string
	ErpnextBranchName        string
	HeadquarterUuid          int64
	EnableDataManagement     int
	DeliveryStatus           int // 外送状态: 0-关闭 1-开启
	EnableKiosk              int // 自助点餐机: 0-关闭 1-开启
	IsOpenMemberInstant      int // 扫码点餐到店自取: 0-关闭 1-开启
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
		INSERT INTO ttpos_company_setting (uuid, company_uuid, erpnext_site_code, erpnext_pos_profile_name, erpnext_admin_email, erpnext_company_abbr, erpnext_headquarter_abbr, headquarter_uuid, erpnext_branch_name, enable_data_management, delivery_status, enable_kiosk, is_open_member_instant, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.CompanyUUID, opt.ErpnextSiteCode, opt.ErpnextPosProfileName, opt.ErpnextAdminEmail, opt.ErpnextCompanyAbbr, opt.ErpnextHeadquarterAbbr, opt.HeadquarterUuid, opt.ErpnextBranchName, opt.EnableDataManagement, opt.DeliveryStatus, opt.EnableKiosk, opt.IsOpenMemberInstant, now, now)

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

// WithCompanySettingHeadquarterConfig configures the company as a chain headquarter.
// Sets ErpnextSiteCode, ErpnextCompanyAbbr and ErpnextHeadquarterAbbr so that
// IsHeadquarter() returns true (site_code != "" && company_abbr == headquarter_abbr).
func WithCompanySettingHeadquarterConfig(siteCode, companyAbbr string) func(*CompanySettingOptions) {
	return func(o *CompanySettingOptions) {
		o.ErpnextSiteCode = siteCode
		o.ErpnextCompanyAbbr = companyAbbr
		o.ErpnextHeadquarterAbbr = companyAbbr // same as company_abbr = headquarter
	}
}

// WithCompanySettingHeadquarterUuid sets the headquarter_uuid for the company setting.
func WithCompanySettingHeadquarterUuid(uuid int64) func(*CompanySettingOptions) {
	return func(o *CompanySettingOptions) {
		o.HeadquarterUuid = uuid
	}
}

// WithCompanySettingEnableDataManagement enables the data management feature (1=enabled, 0=disabled).
// Required for SetDataManage and GetDataManage endpoints to work.
func WithCompanySettingEnableDataManagement(v int) func(*CompanySettingOptions) {
	return func(o *CompanySettingOptions) {
		o.EnableDataManagement = v
	}
}

// WithCompanySettingDeliveryStatus sets the delivery_status (0=off, 1=on).
// Controls IsOpenRider() which gates is_show_delivery on products.
func WithCompanySettingDeliveryStatus(v int) func(*CompanySettingOptions) {
	return func(o *CompanySettingOptions) {
		o.DeliveryStatus = v
	}
}

// WithCompanySettingEnableKiosk sets the enable_kiosk flag (0=off, 1=on).
// Controls IsOpenKiosk() which gates is_show_kiosk on products.
func WithCompanySettingEnableKiosk(v int) func(*CompanySettingOptions) {
	return func(o *CompanySettingOptions) {
		o.EnableKiosk = v
	}
}

// WithCompanySettingIsOpenMemberInstant sets the is_open_member_instant flag (0=off, 1=on).
// Controls IsOpenScanOrder() — scan-to-order self-pickup for delivery display.
func WithCompanySettingIsOpenMemberInstant(v int) func(*CompanySettingOptions) {
	return func(o *CompanySettingOptions) {
		o.IsOpenMemberInstant = v
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

// GenerateSnowflakeID generates a unique snowflake-style ID for testing.
func GenerateSnowflakeID() int64 {
	return generateSnowflakeID()
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

// MemberOptions contains options for seeding a member.
type MemberOptions struct {
	UUID      int64
	Nickname  string
	Phone     string
	IsVisitor int
}

// SeedMember creates a member record in the tenant database.
func SeedMember(tb testing.TB, db *sql.DB, opts ...func(*MemberOptions)) MemberOptions {
	tb.Helper()

	opt := MemberOptions{
		UUID:      generateSnowflakeID(),
		Nickname:  "Test Member",
		Phone:     "0800000000",
		IsVisitor: 0,
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_member (uuid, nickname, phone, is_visitor, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.Nickname, opt.Phone, opt.IsVisitor, now, now)

	if err != nil {
		tb.Fatalf("failed to seed member: %v", err)
	}

	return opt
}

// WithMemberUUID sets the UUID for the member.
func WithMemberUUID(uuid int64) func(*MemberOptions) {
	return func(o *MemberOptions) {
		o.UUID = uuid
	}
}

// SettingOptions contains options for seeding a setting record.
type SettingOptions struct {
	Key    string
	Values string
}

// SeedSetting creates a setting record in the tenant database (ttpos_setting).
func SeedSetting(tb testing.TB, db *sql.DB, key string, values string) {
	tb.Helper()

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_setting (` + "`key`" + `, ` + "`describe`" + `, ` + "`values`" + `, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, 0)
	`, key, key, values, now, now)

	if err != nil {
		tb.Fatalf("failed to seed setting %s: %v", key, err)
	}
}

// ProductFlavorOptions contains options for seeding a product flavor.
type ProductFlavorOptions struct {
	UUID        int64
	Name        string
	ProductUUID int64 // set via product_package.product_uuid relation (not stored in flavor table directly)
}

// SeedProductFlavor creates a product_flavor record in the tenant database.
func SeedProductFlavor(tb testing.TB, db *sql.DB, opts ...func(*ProductFlavorOptions)) ProductFlavorOptions {
	tb.Helper()

	opt := ProductFlavorOptions{
		UUID: generateSnowflakeID(),
		Name: `{"zh_name":"默认规格","en_name":"Default"}`,
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_product_flavor (uuid, name, multi_language_name_uuid, create_time, update_time, delete_time)
		VALUES (?, ?, 0, ?, ?, 0)
	`, opt.UUID, opt.Name, now, now)

	if err != nil {
		tb.Fatalf("failed to seed product_flavor: %v", err)
	}

	return opt
}

// WithProductFlavorUUID sets a custom UUID for the product flavor.
func WithProductFlavorUUID(uuid int64) func(*ProductFlavorOptions) {
	return func(o *ProductFlavorOptions) {
		o.UUID = uuid
	}
}

// ProductBomOptions contains options for seeding a product BOM.
type ProductBomOptions struct {
	UUID               int64
	ProductFlavorUUID  int64
	ProductPackageUUID int64
	Price              float64
	Status             int
}

// SeedProductBom creates a product_bom record in the tenant database.
// Links a product_flavor to a price and status (上架/下架).
func SeedProductBom(tb testing.TB, db *sql.DB, opts ...func(*ProductBomOptions)) ProductBomOptions {
	tb.Helper()

	opt := ProductBomOptions{
		UUID:               generateSnowflakeID(),
		ProductFlavorUUID:  0,
		ProductPackageUUID: 0,
		Price:              10.00,
		Status:             1, // 上架
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_product_bom (uuid, product_flavor_uuid, product_package_uuid, price, status, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.ProductFlavorUUID, opt.ProductPackageUUID, opt.Price, opt.Status, now, now)

	if err != nil {
		tb.Fatalf("failed to seed product_bom: %v", err)
	}

	return opt
}

// WithProductBomFlavorUUID sets the product_flavor_uuid for the product BOM.
func WithProductBomFlavorUUID(uuid int64) func(*ProductBomOptions) {
	return func(o *ProductBomOptions) {
		o.ProductFlavorUUID = uuid
	}
}

// WithProductBomPrice sets the price for the product BOM.
func WithProductBomPrice(price float64) func(*ProductBomOptions) {
	return func(o *ProductBomOptions) {
		o.Price = price
	}
}

// WithProductBomPackageUUID sets the product_package_uuid for the product BOM.
func WithProductBomPackageUUID(uuid int64) func(*ProductBomOptions) {
	return func(o *ProductBomOptions) {
		o.ProductPackageUUID = uuid
	}
}

// ProductPackageOptions contains options for seeding a product package.
type ProductPackageOptions struct {
	UUID             int64
	Name             string
	Status           int
	ProductType      int   // 0=普通商品, 1=套餐
	HeadquarterUuid  int64 // 总部UUID, 0=总部原生商品
}

// SeedProductPackage creates a product_package record.
func SeedProductPackage(tb testing.TB, db *sql.DB, opts ...func(*ProductPackageOptions)) ProductPackageOptions {
	tb.Helper()

	opt := ProductPackageOptions{
		UUID:            generateSnowflakeID(),
		Name:            `{"zh_name":"测试商品包","en_name":"Test Package"}`,
		Status:          1,
		ProductType:     0,
		HeadquarterUuid: 0,
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_product_package (uuid, name, multi_language_name_uuid, status, product_type, headquarter_uuid, create_time, update_time, delete_time)
		VALUES (?, ?, 0, ?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.Name, opt.Status, opt.ProductType, opt.HeadquarterUuid, now, now)

	if err != nil {
		tb.Fatalf("failed to seed product_package: %v", err)
	}

	return opt
}

// WithProductPackageUUID sets the UUID for the product package.
func WithProductPackageUUID(uuid int64) func(*ProductPackageOptions) {
	return func(o *ProductPackageOptions) {
		o.UUID = uuid
	}
}

// WithProductPackageProductType sets the product_type for the product package (0=普通, 1=套餐).
func WithProductPackageProductType(productType int) func(*ProductPackageOptions) {
	return func(o *ProductPackageOptions) {
		o.ProductType = productType
	}
}

// WithProductPackageStatus sets the status for the product package (0=下架, 1=上架).
func WithProductPackageStatus(status int) func(*ProductPackageOptions) {
	return func(o *ProductPackageOptions) {
		o.Status = status
	}
}

// WithProductPackageHeadquarterUuid sets the headquarter_uuid for the product package.
func WithProductPackageHeadquarterUuid(uuid int64) func(*ProductPackageOptions) {
	return func(o *ProductPackageOptions) {
		o.HeadquarterUuid = uuid
	}
}

// WithProductBomStatus sets the status for the product BOM (0=下架, 1=上架).
func WithProductBomStatus(status int) func(*ProductBomOptions) {
	return func(o *ProductBomOptions) {
		o.Status = status
	}
}

// ProductPackageGroupOptions contains options for seeding a product package group.
type ProductPackageGroupOptions struct {
	UUID               int64
	Name               string
	ProductPackageUUID int64 // 套餐 product_package 的 UUID
	GroupType          int   // 0=固定, 1=可选
	OptionalMinCount   int
	OptionalCount      int
	Sort               int
}

// SeedProductPackageGroup creates a product_package_group record in the tenant database.
func SeedProductPackageGroup(tb testing.TB, db *sql.DB, opts ...func(*ProductPackageGroupOptions)) ProductPackageGroupOptions {
	tb.Helper()

	opt := ProductPackageGroupOptions{
		UUID:               generateSnowflakeID(),
		Name:               `{"zh_name":"默认分组","en_name":"Default Group"}`,
		ProductPackageUUID: 0,
		GroupType:          0,
		OptionalMinCount:   0,
		OptionalCount:      0,
		Sort:               0,
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_product_package_group (uuid, name, multi_language_name_uuid, product_package_uuid, group_type, optional_min_count, optional_count, sort, create_time, update_time, delete_time)
		VALUES (?, ?, 0, ?, ?, ?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.Name, opt.ProductPackageUUID, opt.GroupType, opt.OptionalMinCount, opt.OptionalCount, opt.Sort, now, now)

	if err != nil {
		tb.Fatalf("failed to seed product_package_group: %v", err)
	}

	return opt
}

// WithProductPackageGroupProductPackageUUID sets the parent product_package_uuid.
func WithProductPackageGroupProductPackageUUID(uuid int64) func(*ProductPackageGroupOptions) {
	return func(o *ProductPackageGroupOptions) {
		o.ProductPackageUUID = uuid
	}
}

// ProductPackageGroupItemOptions contains options for seeding a product package group item.
type ProductPackageGroupItemOptions struct {
	UUID                    int64
	ProductPackageGroupUUID int64   // 所属分组的 UUID
	RelatedUUID             int64   // 关联的子商品 product_package UUID
	ProductBomUUID          int64   // 子商品的 product_bom UUID
	Num                     float64 // 数量
	Sort                    int
	AddPrice                float64 // 加价金额
	IsRequired              int     // 0=不必选, 1=必选
	IsDefault               int     // 0=不默认, 1=默认选中
}

// SeedProductPackageGroupItem creates a product_package_group_item record.
func SeedProductPackageGroupItem(tb testing.TB, db *sql.DB, opts ...func(*ProductPackageGroupItemOptions)) ProductPackageGroupItemOptions {
	tb.Helper()

	opt := ProductPackageGroupItemOptions{
		UUID:                    generateSnowflakeID(),
		ProductPackageGroupUUID: 0,
		RelatedUUID:             0,
		ProductBomUUID:          0,
		Num:                     1,
		Sort:                    0,
		AddPrice:                0,
		IsRequired:              0,
		IsDefault:               0,
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_product_package_group_item (uuid, product_package_group_uuid, related_uuid, product_bom_uuid, num, sort, add_price, is_required, is_default, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.ProductPackageGroupUUID, opt.RelatedUUID, opt.ProductBomUUID, opt.Num, opt.Sort, opt.AddPrice, opt.IsRequired, opt.IsDefault, now, now)

	if err != nil {
		tb.Fatalf("failed to seed product_package_group_item: %v", err)
	}

	return opt
}

// PackageSubItem represents a sub-product inside a package for the convenience seed function.
type PackageSubItem struct {
	Name  string
	Price float64
}

// SeedPackageProductResult holds the result of SeedPackageProductWithSubItems.
type SeedPackageProductResult struct {
	PackageUUID    int64   // The parent package's product_package UUID (used as product_package_uuid in API)
	PackageBomUUID int64   // The parent package's product_bom UUID (used as flavor_uuid in API)
	GroupUUID      int64   // The package group UUID (used as product_package_group_uuid in API)
	SubBomUUIDs    []int64 // Each sub-product's product_bom UUID (used as flavor_uuid in API)
}

// SeedPackageProductWithSubItems creates a complete package product (套餐) with sub-items.
// It creates:
//   - A parent product_package (product_type=1) + product_flavor + product_bom
//   - A product_package_group linked to the parent
//   - For each sub-item: product_package + product_flavor + product_bom + product_package_group_item
//
// Returns the parent package UUID, parent BOM UUID, group UUID, and sub-item BOM UUIDs.
func SeedPackageProductWithSubItems(tb testing.TB, db *sql.DB, packageName string, packagePrice float64, subItems []PackageSubItem) SeedPackageProductResult {
	tb.Helper()

	// 1. Create parent product_package (product_type=1 = 套餐)
	parentPkg := SeedProductPackage(tb, db, func(o *ProductPackageOptions) {
		o.Name = fmt.Sprintf(`{"zh_name":"%s","en_name":"%s"}`, packageName, packageName)
		o.ProductType = 1
	})

	// 2. Create parent product_flavor + product_bom
	parentFlavor := SeedProductFlavor(tb, db)
	parentBom := SeedProductBom(tb, db,
		WithProductBomFlavorUUID(parentFlavor.UUID),
		WithProductBomPackageUUID(parentPkg.UUID),
		WithProductBomPrice(packagePrice),
	)

	// 3. Create a package group
	group := SeedProductPackageGroup(tb, db, func(o *ProductPackageGroupOptions) {
		o.ProductPackageUUID = parentPkg.UUID
	})

	// 4. Create sub-items
	subBomUUIDs := make([]int64, 0, len(subItems))
	for _, sub := range subItems {
		// Each sub-item is its own product_package + product_flavor + product_bom chain
		subPkg := SeedProductPackage(tb, db, func(o *ProductPackageOptions) {
			o.Name = fmt.Sprintf(`{"zh_name":"%s","en_name":"%s"}`, sub.Name, sub.Name)
			o.ProductType = 0
		})
		subFlavor := SeedProductFlavor(tb, db)
		subBom := SeedProductBom(tb, db,
			WithProductBomFlavorUUID(subFlavor.UUID),
			WithProductBomPackageUUID(subPkg.UUID),
			WithProductBomPrice(sub.Price),
		)

		// Link sub-item to the group
		SeedProductPackageGroupItem(tb, db, func(o *ProductPackageGroupItemOptions) {
			o.ProductPackageGroupUUID = group.UUID
			o.RelatedUUID = subPkg.UUID
			o.ProductBomUUID = subBom.UUID
		})

		subBomUUIDs = append(subBomUUIDs, subBom.UUID)
	}

	return SeedPackageProductResult{
		PackageUUID:    parentPkg.UUID,
		PackageBomUUID: parentBom.UUID,
		GroupUUID:      group.UUID,
		SubBomUUIDs:    subBomUUIDs,
	}
}

// SeedProductWithFlavor is a convenience function that creates a complete product
// with package, flavor, and bom all linked together. Returns the product_bom UUID
// which is what the API expects as flavor_uuid in order requests.
//
// Data chain: product_package ← product_bom → product_flavor
// The product_bom.uuid is used as "flavor_uuid" in API requests.
func SeedProductWithFlavor(tb testing.TB, db *sql.DB, productName string, price float64) int64 {
	tb.Helper()

	// 1. Create product_package (the "product" entity in the system)
	pkg := SeedProductPackage(tb, db, func(o *ProductPackageOptions) {
		o.Name = fmt.Sprintf(`{"zh_name":"%s","en_name":"%s"}`, productName, productName)
	})

	// 2. Create product_flavor (spec/variant)
	flavor := SeedProductFlavor(tb, db)

	// 3. Create product_bom linking flavor and package
	bom := SeedProductBom(tb, db,
		WithProductBomFlavorUUID(flavor.UUID),
		WithProductBomPackageUUID(pkg.UUID),
		WithProductBomPrice(price),
	)

	return bom.UUID
}

// MustPlanOptions contains options for seeding a product must plan.
type MustPlanOptions struct {
	UUID        int64
	Name        string
	UseChannel  string // "10"=dining, "20"=desk, "10,20"=both
	MustType    int    // 0=each order, 1=each person
	MustRule    int    // 0=fixed products, 1=optional
	Status      int    // 1=enabled, 0=disabled
	AutoCart    int    // 1=auto add to cart, 0=manual
	AutoChange  int    // 1=customer can change qty
	AutoCheck   int    // 1=check at order time
	AutoCheckout int   // 1=check at checkout time
}

// SeedMustPlan creates a product_must_plan record in the tenant database.
func SeedMustPlan(tb testing.TB, db *sql.DB, opts ...func(*MustPlanOptions)) MustPlanOptions {
	tb.Helper()

	opt := MustPlanOptions{
		UUID:         generateSnowflakeID(),
		Name:         "Test Must Plan",
		UseChannel:   "20",
		MustType:     0,
		MustRule:     0,
		Status:       1,
		AutoCart:     1,
		AutoChange:   1,
		AutoCheck:    1,
		AutoCheckout: 1,
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_product_must_plan
		(uuid, name, use_channel, must_type, must_rule, status, auto_cart, auto_change, auto_check, auto_checkout, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.Name, opt.UseChannel, opt.MustType, opt.MustRule, opt.Status,
		opt.AutoCart, opt.AutoChange, opt.AutoCheck, opt.AutoCheckout, now, now)

	if err != nil {
		tb.Fatalf("failed to seed must plan: %v", err)
	}

	return opt
}

// WithMustPlanUUID sets the UUID for the must plan.
func WithMustPlanUUID(uuid int64) func(*MustPlanOptions) {
	return func(o *MustPlanOptions) { o.UUID = uuid }
}

// WithMustPlanUseChannel sets the use channel.
func WithMustPlanUseChannel(ch string) func(*MustPlanOptions) {
	return func(o *MustPlanOptions) { o.UseChannel = ch }
}

// WithMustPlanAutoCart sets auto_cart (1=auto, 0=manual).
func WithMustPlanAutoCart(v int) func(*MustPlanOptions) {
	return func(o *MustPlanOptions) { o.AutoCart = v }
}

// MustPlanItemOptions contains options for seeding a must plan item.
type MustPlanItemOptions struct {
	UUID                int64
	ProductMustPlanUUID int64
	ProductPackageUUID  int64
}

// SeedMustPlanItem creates a product_must_plan_item record.
func SeedMustPlanItem(tb testing.TB, db *sql.DB, opts ...func(*MustPlanItemOptions)) MustPlanItemOptions {
	tb.Helper()

	opt := MustPlanItemOptions{
		UUID: generateSnowflakeID(),
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_product_must_plan_item
		(uuid, product_must_plan_uuid, product_package_uuid, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.ProductMustPlanUUID, opt.ProductPackageUUID, now, now)

	if err != nil {
		tb.Fatalf("failed to seed must plan item: %v", err)
	}

	return opt
}

// WithMustPlanItemPlanUUID sets the must plan UUID.
func WithMustPlanItemPlanUUID(uuid int64) func(*MustPlanItemOptions) {
	return func(o *MustPlanItemOptions) { o.ProductMustPlanUUID = uuid }
}

// WithMustPlanItemPackageUUID sets the product package UUID.
func WithMustPlanItemPackageUUID(uuid int64) func(*MustPlanItemOptions) {
	return func(o *MustPlanItemOptions) { o.ProductPackageUUID = uuid }
}

// MustPlanRegionOptions contains options for seeding a must plan region.
type MustPlanRegionOptions struct {
	UUID                int64
	ProductMustPlanUUID int64
	DeskRegionUUID      int64
}

// SeedMustPlanRegion creates a product_must_plan_region record.
func SeedMustPlanRegion(tb testing.TB, db *sql.DB, opts ...func(*MustPlanRegionOptions)) MustPlanRegionOptions {
	tb.Helper()

	opt := MustPlanRegionOptions{
		UUID: generateSnowflakeID(),
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_product_must_plan_region
		(uuid, product_must_plan_uuid, desk_region_uuid, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.ProductMustPlanUUID, opt.DeskRegionUUID, now, now)

	if err != nil {
		tb.Fatalf("failed to seed must plan region: %v", err)
	}

	return opt
}

// WithMustPlanRegionPlanUUID sets the must plan UUID.
func WithMustPlanRegionPlanUUID(uuid int64) func(*MustPlanRegionOptions) {
	return func(o *MustPlanRegionOptions) { o.ProductMustPlanUUID = uuid }
}

// WithMustPlanRegionDeskRegionUUID sets the desk region UUID.
func WithMustPlanRegionDeskRegionUUID(uuid int64) func(*MustPlanRegionOptions) {
	return func(o *MustPlanRegionOptions) { o.DeskRegionUUID = uuid }
}

// DeskRegionOptions contains options for seeding a desk region.
type DeskRegionOptions struct {
	UUID int64
	Name string
	Sort int
}

// SeedDeskRegion creates a desk_region record in the tenant database.
func SeedDeskRegion(tb testing.TB, db *sql.DB, opts ...func(*DeskRegionOptions)) DeskRegionOptions {
	tb.Helper()

	opt := DeskRegionOptions{
		UUID: generateSnowflakeID(),
		Name: `{"zh_name":"大厅","en_name":"Main Hall"}`,
		Sort: 0,
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_desk_region (uuid, name, sort, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.Name, opt.Sort, now, now)

	if err != nil {
		tb.Fatalf("failed to seed desk region: %v", err)
	}

	return opt
}

// WithDeskRegionRecordUUID sets the UUID for the desk region record.
func WithDeskRegionRecordUUID(uuid int64) func(*DeskRegionOptions) {
	return func(o *DeskRegionOptions) { o.UUID = uuid }
}

// StatisticsSaleOptions contains options for seeding a statistics_sale record.
type StatisticsSaleOptions struct {
	UUID                 int64
	SaleBillUUID         int64
	DeskUUID             int64
	OrderSourceUUID      int64
	Source               int // 0=default, 1=cashier, 5=scan
	IsMeger              int
	IsSpecial            int
	IsTakeout            int
	PaymentAmount        float64
	RefundAmount         float64
	RefundPaymentBalance float64
	CompleteTime         int64
	MealNum              int
}

// SeedStatisticsSale creates a statistics_sale record in the tenant database.
func SeedStatisticsSale(tb testing.TB, db *sql.DB, opts ...func(*StatisticsSaleOptions)) StatisticsSaleOptions {
	tb.Helper()

	opt := StatisticsSaleOptions{
		UUID:         generateSnowflakeID(),
		SaleBillUUID: generateSnowflakeID(),
		CompleteTime: 1750000000,
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_statistics_sale (
			uuid, sale_bill_uuid, desk_uuid, order_source_uuid, source,
			is_meger, is_special, is_takeout, payment_amount, refund_amount,
			refund_payment_balance, complete_time, meal_num,
			create_time, update_time, delete_time
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.SaleBillUUID, opt.DeskUUID, opt.OrderSourceUUID, opt.Source,
		opt.IsMeger, opt.IsSpecial, opt.IsTakeout, opt.PaymentAmount, opt.RefundAmount,
		opt.RefundPaymentBalance, opt.CompleteTime, opt.MealNum,
		now, now)

	if err != nil {
		tb.Fatalf("failed to seed statistics sale: %v", err)
	}

	return opt
}

// WithStatisticsSaleBillUUID sets the sale_bill_uuid for the statistics_sale record.
func WithStatisticsSaleBillUUID(uuid int64) func(*StatisticsSaleOptions) {
	return func(o *StatisticsSaleOptions) {
		o.SaleBillUUID = uuid
	}
}

// WithStatisticsSaleSource sets the source for the statistics_sale record.
func WithStatisticsSaleSource(source int) func(*StatisticsSaleOptions) {
	return func(o *StatisticsSaleOptions) {
		o.Source = source
	}
}

// WithStatisticsSalePaymentAmount sets the payment_amount.
func WithStatisticsSalePaymentAmount(amount float64) func(*StatisticsSaleOptions) {
	return func(o *StatisticsSaleOptions) {
		o.PaymentAmount = amount
	}
}

// WithStatisticsSaleRefundAmount sets the refund_amount.
func WithStatisticsSaleRefundAmount(amount float64) func(*StatisticsSaleOptions) {
	return func(o *StatisticsSaleOptions) {
		o.RefundAmount = amount
	}
}

// WithStatisticsSaleCompleteTime sets the complete_time.
func WithStatisticsSaleCompleteTime(ts int64) func(*StatisticsSaleOptions) {
	return func(o *StatisticsSaleOptions) {
		o.CompleteTime = ts
	}
}

// SeedCashierPermissions seeds the ttpos_access table with all cashier permissions
// required by constant.CashierPermissions. This is needed for super-admin staff
// because getDbPermissions returns ALL access records (no filter) for is_super=1,
// and the permission check matches access.Path against the CashierPermissions map.
func SeedCashierPermissions(tb testing.TB, db *sql.DB) {
	tb.Helper()

	permissions := []string{
		"cashier_table_open",
		"cashier_table_delete",
	}

	now := time.Now().Unix()
	for _, perm := range permissions {
		_, err := db.Exec(`
			INSERT INTO ttpos_access (uuid, name, path, create_time, update_time, delete_time)
			VALUES (?, ?, ?, ?, ?, 0)
		`, generateSnowflakeID(), perm, perm, now, now)
		if err != nil {
			tb.Fatalf("failed to seed access permission %s: %v", perm, err)
		}
	}
}
