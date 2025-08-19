// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ProductMustPlan is the golang structure of table ttpos_product_must_plan for DAO operations like Where/Data.
type ProductMustPlan struct {
	g.Meta       `orm:"table:ttpos_product_must_plan, do:true"`
	Id           interface{} // 自增ID
	Uuid         interface{} // 商品必选商品计划ID
	Name         interface{} // 方案名称
	UseChannel   interface{} // 使用渠道 10-点餐方式 20-桌台方式
	MustType     interface{} // 必点类型 0-每笔订单必点1份 1-每人必点1份
	MustRule     interface{} // 必点规则 0-固定商品 1-可选商品
	Status       interface{} // 状态,1-开启 0-关闭
	AutoCart     interface{} // 自动加入购物车 1-是 0-否
	AutoChange   interface{} // 顾客可修改必点数量 1-是 0-否
	AutoCheck    interface{} // 下单时检查必点商品 1-是 0-否
	AutoCheckout interface{} // 结账时检查必点商品 1-是 0-否
	CreateTime   interface{} // 创建时间(时间戳)
	UpdateTime   interface{} // 更新时间(时间戳)
	DeleteTime   interface{} // 删除时间(时间戳)
}
