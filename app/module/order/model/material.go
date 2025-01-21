package model

import (
	"time"

	"gorm.io/gorm"
)

// Material 原料信息表
type Material struct {
	ID                  int     `gorm:"primaryKey;column:id;comment:原料唯一标识符"`
	Name                string  `gorm:"column:name;comment:原料名称"`
	MultiLanguageNameID int     `gorm:"column:multi_language_name_id;comment:多语言名称ID"`
	CategoryKey         string  `gorm:"column:category_key;comment:类别关键字"`
	CategoryID          int     `gorm:"column:category_id;comment:类别ID"`
	SupplierID          int     `gorm:"column:supplier_id;comment:供应商ID"`
	ImageURL            string  `gorm:"column:image_url;comment:图片URL"`
	ImageName           string  `gorm:"column:image_name;comment:图片名称"`
	UnitID              int     `gorm:"column:unit_id;comment:单位ID"`
	Price               float64 `gorm:"column:price;comment:采购单价"`
	Num                 int     `gorm:"column:num;comment:库存数量"`
	BarcodeValue        string  `gorm:"column:barcode_value;comment:条形码值"`
	Status              string  `gorm:"column:status;comment:状态,up上架、down下架"`
	CreateTime          int64   `gorm:"column:create_time;comment:创建时间"`
	UpdateTime          int64   `gorm:"column:update_time;comment:更新时间"`
	DeleteTime          int64   `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *Material) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *Material) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
