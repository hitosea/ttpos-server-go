package model

// Supplier 供应商表 ttpos_supplier
type Supplier struct {
	BaseModel
	Name               string `gorm:"default:'';column:name;comment:'供应商名称'"`
	Code               string `gorm:"default:'';column:code;comment:'编码'"`
	Status             int    `gorm:"default:0;column:status;comment:'状态：0-禁用；1-启用'"`
	Address            string `gorm:"default:'';column:address;comment:'供应商地址'"`
	ContactName        string `gorm:"default:'';column:contact_name;comment:'联系人姓名'"`
	ContactPhone       string `gorm:"default:'';column:contact_phone;comment:'联系人电话'"`
	Position           string `gorm:"default:'';column:position;comment:'职位'"`
	StaffUuid          uint64 `gorm:"default:0;column:staff_uuid;comment:'员工ID, 采购负责人'"`
	ErpCode            string `gorm:"default:'';column:erp_code;comment:'关联erpnext'"`
	RepresentsCompany  string `gorm:"default:'';column:represents_company;comment:'代表公司'"`
	IsInternalSupplier int    `gorm:"default:0;column:is_internal_supplier;comment:'是否内部供应商：0-否；1-是'"`
	HeadquarterUuid    uint64 `gorm:"default:0;column:headquarter_uuid;comment:'总部Uuid'"`

	// purchase order 表的supplier_erp_code 关联supplier的erp_code
	PurchaseOrders []*PurchaseOrder `gorm:"foreignKey:SupplierErpCode;references:ErpCode"`
}
