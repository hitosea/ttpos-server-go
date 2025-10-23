package model

// MaterialSupplier 原料供应商关联表 `ttpos_material_supplier`
type MaterialSupplier struct {
	BaseModel
	MaterialUuid    uint64 `gorm:"default:0;column:material_uuid;comment:'原料UUID'"`
	MaterialCode    string `gorm:"default:'';column:material_code;comment:'原料编码'"`
	SupplierUuid    uint64 `gorm:"default:0;column:supplier_uuid;comment:'供应商UUID'"`
	SupplierErpCode string `gorm:"default:'';column:supplier_erp_code;comment:'供应商ERP编码'"`
	HeadquarterUuid uint64 `gorm:"default:0;column:headquarter_uuid;comment:'总部UUID'"`

	// 关联模型
	Material *Material `gorm:"foreignKey:material_uuid;references:uuid"` // 原料信息
	Supplier *Supplier `gorm:"foreignKey:supplier_uuid;references:uuid"` // 供应商信息
}

// SetNil 清空关联模型
func (model *MaterialSupplier) SetNil() {
	model.Material = nil
	model.Supplier = nil
}

// TableName 获取表名
func (MaterialSupplier) TableName() string {
	return "ttpos_material_supplier"
}
