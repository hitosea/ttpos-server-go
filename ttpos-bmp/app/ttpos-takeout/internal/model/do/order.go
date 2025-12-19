// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// Order is the golang structure of table takeout_order for DAO operations like Where/Data.
type Order struct {
	g.Meta             `orm:"table:takeout_order, do:true"`
	Id                 any         // 主键
	Uuid               any         // 系统唯一ID
	ShopUuid           any         // TTPOS店铺UUID
	ProviderMerchantId any         // 渠道商户ID (Provider Merchant ID)
	ProviderOrderId    any         // 平台订单号 (Grab Order ID)
	ShopRefNo          any         // TTPOS内部订单号
	ShortOrderNumber   any         // 短单号
	ProviderName       any         // 渠道: grab, foodpanda
	OrderType          any         // 订单类型: DeliveryByGrab, DeliveryByRestaurant, TakeAway, DineIn
	OrderTime          *gtime.Time // 下单时间
	OrderStatus        any         // 订单状态: PENDING, ACCEPTED, PREPARING, READY, COMPLETED, CANCELLED
	ScheduledTime      *gtime.Time // 预约送达时间
	Currency           any         // 货币代码
	Subtotal           any         // 商品小计
	TotalAmount        any         // 总金额
	MerchantCharge     any         // 商户收取金额
	TaxAmount          any         // 税费
	DiscountAmount     any         // 折扣金额
	MerchantFundPromo  any         // 商户承担促销金额
	PaymentType        any         // 支付方式: CASH, CASHLESS
	IsMexEditOrder     any         // 是否为商户编辑订单
	Cutlery            any         // 是否需要餐具
	EaterCount         any         // 用餐人数
	CustomerName       any         // 顾客姓名
	CustomerPhone      any         // 顾客电话 (虚拟号)
	DeliveryAddress    any         // 配送地址 (JSON)
	DriverEta          any         // 骑手预计到达时间(分钟)
	ReadyTime          *gtime.Time // 商家预估准备完成时间
	Note               any         // 订单备注
	RawData            any         // 原始JSON数据
	CreatedAt          *gtime.Time // 创建时间
	UpdatedAt          *gtime.Time // 更新时间
	DeletedAt          *gtime.Time // 软删除
}
