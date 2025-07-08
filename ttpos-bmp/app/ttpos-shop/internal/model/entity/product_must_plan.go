// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ProductMustPlan is the golang structure for table product_must_plan.
type ProductMustPlan struct {
	Id           uint   `json:"id"           orm:"id"            description:"自增ID"`                     // 自增ID
	Uuid         uint64 `json:"uuid"         orm:"uuid"          description:"商品必选商品计划ID"`               // 商品必选商品计划ID
	Name         string `json:"name"         orm:"name"          description:"方案名称"`                     // 方案名称
	UseChannel   string `json:"useChannel"   orm:"use_channel"   description:"使用渠道 10-点餐方式 20-桌台方式"`     // 使用渠道 10-点餐方式 20-桌台方式
	MustType     int    `json:"mustType"     orm:"must_type"     description:"必点类型 0-每笔订单必点1份 1-每人必点1份"` // 必点类型 0-每笔订单必点1份 1-每人必点1份
	MustRule     int    `json:"mustRule"     orm:"must_rule"     description:"必点规则 0-固定商品 1-可选商品"`       // 必点规则 0-固定商品 1-可选商品
	Status       int    `json:"status"       orm:"status"        description:"状态,1-开启 0-关闭"`             // 状态,1-开启 0-关闭
	AutoCart     int    `json:"autoCart"     orm:"auto_cart"     description:"自动加入购物车 1-是 0-否"`          // 自动加入购物车 1-是 0-否
	AutoChange   int    `json:"autoChange"   orm:"auto_change"   description:"顾客可修改必点数量 1-是 0-否"`        // 顾客可修改必点数量 1-是 0-否
	AutoCheck    int    `json:"autoCheck"    orm:"auto_check"    description:"下单时检查必点商品 1-是 0-否"`        // 下单时检查必点商品 1-是 0-否
	AutoCheckout int    `json:"autoCheckout" orm:"auto_checkout" description:"结账时检查必点商品 1-是 0-否"`        // 结账时检查必点商品 1-是 0-否
	CreateTime   uint   `json:"createTime"   orm:"create_time"   description:"创建时间(时间戳)"`                // 创建时间(时间戳)
	UpdateTime   uint   `json:"updateTime"   orm:"update_time"   description:"更新时间(时间戳)"`                // 更新时间(时间戳)
	DeleteTime   uint   `json:"deleteTime"   orm:"delete_time"   description:"删除时间(时间戳)"`                // 删除时间(时间戳)
}
