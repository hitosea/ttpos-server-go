package model

import (
	"time"

	"gorm.io/gorm"
)

// SaleOrderProduct 销售订单商品表
type SaleOrderProduct struct {
	ID                  int     `gorm:"primaryKey;column:id;comment:销售订单商品唯一标识符"`
	Name                string  `gorm:"column:name;comment:产品名称"`
	FlavorName          string  `gorm:"column:flavor_name;comment:口味名称"`
	MultiLanguageNameID int     `gorm:"column:multi_language_name_id;comment:多语言名称ID"`
	Num                 int     `gorm:"column:num;comment:数量"`
	CustomPrice         float64 `gorm:"column:custom_price;comment:自定义价格"`
	UnitPrice           float64 `gorm:"column:unit_price;comment:单价"`
	Price               float64 `gorm:"column:price;comment:最终单价"`
	Status              string  `gorm:"column:status;comment:状态"`
	Remark              string  `gorm:"column:remark;comment:备注"`
	IsGift              bool    `gorm:"column:is_gift;comment:是否赠品"`
	GiftReason          string  `gorm:"column:gift_reason;comment:赠品原因"`
	OrderProductID      int     `gorm:"column:order_product_id;comment:订单产品ID"`
	ProductionOrderID   int     `gorm:"column:production_order_id;comment:生产订单ID"`
	Sign                string  `gorm:"column:sign;comment:商品签名"`
	ProductPackageID    int     `gorm:"column:product_package_id;comment:产品包ID"`
	SaleBillID          int     `gorm:"column:sale_bill_id;comment:账单ID"`
	CreateTime          int64   `gorm:"column:create_time;comment:创建时间"`
	UpdateTime          int64   `gorm:"column:update_time;comment:更新时间"`
	DeleteTime          int64   `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *SaleOrderProduct) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *SaleOrderProduct) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
