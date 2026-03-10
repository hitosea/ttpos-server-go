package fixture

import (
	"database/sql"
	"fmt"
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
		INSERT INTO ttpos_staff (uuid, company_uuid, real_name, username, password, phone, is_super, duty_no, create_time, update_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, opt.UUID, opt.CompanyUUID, opt.RealName, opt.Username, opt.Password, opt.Phone, opt.IsSuper, opt.DutyNo, now, now)

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
	UUID       int64
	Name       string
	Status     int
	ExpireTime int64
}

// SeedCompany creates a company record in the tenant database.
// status=1 and expire_time=0 are required for auth to pass.
func SeedCompany(tb testing.TB, db *sql.DB, opts ...func(*CompanyOptions)) CompanyOptions {
	tb.Helper()

	opt := CompanyOptions{
		UUID:       generateSnowflakeID(),
		Name:       "Test Company",
		Status:     1,
		ExpireTime: 0, // 0 = never expires
	}

	for _, o := range opts {
		o(&opt)
	}

	now := time.Now().Unix()
	_, err := db.Exec(`
		INSERT INTO ttpos_company (uuid, name, status, expire_time, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, ?, ?, 0)
	`, opt.UUID, opt.Name, opt.Status, opt.ExpireTime, now, now)

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

// CompanySettingOptions contains options for seeding company settings.
type CompanySettingOptions struct {
	UUID        int64
	CompanyUUID int64
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
		INSERT INTO ttpos_company_setting (uuid, company_uuid, create_time, update_time, delete_time)
		VALUES (?, ?, ?, ?, 0)
	`, opt.UUID, opt.CompanyUUID, now, now)

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

// generateSnowflakeID generates a simple unique ID for testing.
// In production, use the actual snowflake ID generator.
func generateSnowflakeID() int64 {
	return time.Now().UnixNano() / 1000000
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
