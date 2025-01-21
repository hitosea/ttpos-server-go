package model

import (
	"time"

	"gorm.io/gorm"
)

// ProductMustProductPlan 产品必选产品计划表
type ProductMustProductPlan struct {
	ID                                 int    `gorm:"primaryKey;column:id;comment:产品必选产品计划唯一标识符"`
	Name                               string `gorm:"column:name;comment:名称"`
	Scene                              string `gorm:"column:scene;comment:场景,order点餐、desk桌台"`
	RequiredType                       string `gorm:"column:required_type;comment:要求类型,per_person每人必点1份、per_order每笔订单必点1份"`
	RequiredRule                       string `gorm:"column:required_rule;comment:要求规则,fixed固定商品、optional可选商品"`
	Status                             string `gorm:"column:status;comment:状态,open开启、close关闭"`
	AutoAddToShoppingCart              bool   `gorm:"column:auto_add_to_shopping_cart;comment:是否自动加入购物车"`
	CustomersCanModifyRequiredQuantity bool   `gorm:"column:customers_can_modify_required_quantity;comment:是否顾客可修改必点数量"`
	RequiredProductCheckInOrder        bool   `gorm:"column:required_product_check_in_order;comment:下单时检查必点商品"`
	RequiredProductCheckInBill         bool   `gorm:"column:required_product_check_in_bill;comment:结账时检查必坚商品"`
	CreateTime                         int64  `gorm:"column:create_time;comment:创建时间"`
	UpdateTime                         int64  `gorm:"column:update_time;comment:更新时间"`
	DeleteTime                         int64  `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *ProductMustProductPlan) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *ProductMustProductPlan) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
