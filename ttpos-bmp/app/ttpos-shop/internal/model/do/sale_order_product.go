// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderProduct is the golang structure of table ttpos_sale_order_product for DAO operations like Where/Data.
type SaleOrderProduct struct {
	g.Meta                 `orm:"table:ttpos_sale_order_product, do:true"`
	OpenMemberDiscount     interface{} // 是否开启会员折扣, 0-否 1-是。添加商品时记录下状态不受后台改变，结账时检查是否改变
	Id                     interface{} // 自增ID
	Uuid                   interface{} // 销售订单商品ID
	Name                   interface{} // 商品名称
	FlavorName             interface{} // 规格名称
	Sign                   interface{} // 商品签名,规格、属性、加料、是否改价、是否赠菜、送厨批次、销售价相同的商品签名相同,用于取消拆单时合并商品
	MultiLanguageNameUuid  interface{} // 多语言名称ID
	Num                    interface{} // 商品数量。不能减为0，当数量为1再减时，标记删除
	ImageFileUuid          interface{} // 商品图片ID
	FlavorPrice            interface{} // 规格原价（单商品）,仅某规格商品的原价
	SaucePrice             interface{} // 小料价（单商品）,所有小料的价格之和
	ProductPrice           interface{} // 原始单价（单商品）,规格原价+小料价
	ChangePriceTime        interface{} // 改价时间(时间戳),用于判断是否改价和不同时间改价的商品不合并
	IsBuffet               interface{} // 是否为自助餐商品,0-否 1-是. 如果是自助餐商品，则sale_price为0
	SalePrice              interface{} // 销售价（单商品，折前价）,当自定义价格时，销售价=自定义价格,否则销售价=原始单价
	SalePriceNoTax         interface{} // 销售价,未含税价格（折前）
	TaxRate                interface{} // 税率,单位%.加购时记录税率,结账时再重新核算
	MemberDiscountRate     interface{} // 会员折扣率(0-100%)
	MemberCardDiscountRate interface{} // 会员卡折扣率(0-100%)
	CustomDiscountRate     interface{} // 自定义折扣率(0-100%)
	MemberDiscountPrice    interface{} // 会员折扣后的价格（单商品）=销售价*会员折扣率*会员卡折扣率
	Price                  interface{} // 最终单价(单商品，会员、会员卡和优惠折扣后，折后价)。销售价*折扣率
	ServiceTaxFee          interface{} // 服务费税费（单商品）,0-不收取税费；收取时，服务费税费=服务费*税率
	TaxFee                 interface{} // 商品税费（单商品）。商品已含税时，税费=规格原价*(1-1/(1+税率))；商品未含税时，税费=原始单价*税率
	ServiceFee             interface{} // 服务费（单商品）,0-固定服务费 大于0-按比例收服务费；商品已含税时，服务费=(最终单价-商品税费)*服务费比例；商品未含税时，服务费=最终单价*服务费比例
	TotalPrice             interface{} // 应收金额(单商品)。商品已含税时，应收金额(单商品)=(最终单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=最终单价+服务费+总税费
	OriginTotalPrice       interface{} // 应收金额(单商品)。商品已含税时，应收金额(单商品)=(销售价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=销售价+服务费+总税费
	DiscountFee            interface{} // 打折金额（单商品）=销售价-最终单价。校验：打折金额=会员折扣金额+自定义折扣金额
	MemberDiscountFee      interface{} // 会员折扣金额（单商品）=销售价*（1-会员折扣率*会员卡折扣率）
	CustomDiscountFee      interface{} // 自定义折扣金额（单商品）。自定义折扣金额（单商品）=会员折扣后的价格（单商品）*(1-自定义折扣率) 。校验：自定义折扣金额（单商品）=销售价 - 最终单价（单商品）-会员折扣金额（单商品）；注意，不能这样算，自定义折扣金额（单商品）=销售价*(1-自定义折扣率)
	Status                 interface{} // 状态, 0-未送厨 1-已送厨
	IsRequire              interface{} // 是否必点商品 0-否 1-是。用于在前端显示必点图标
	DeductStockType        interface{} // 库存计算方式,0-下单减库存 1-付款减库存。加购商品时记录，不受后台影响，用于减少查询次数
	DeductStockTime        interface{} // 减库存的时间(时间戳)，0-未减库存。标记是否已减库存，用于取消订单时恢复库存、避免重复减库存、避免漏减库存
	Remark                 interface{} // 备注，顾客对商品的备注信息
	GiftTime               interface{} // 赠菜时间(时间戳),用于判断是否赠菜和不同时间赠送的商品不合并
	CancelTime             interface{} // 退菜时间(时间戳)
	GiftReason             interface{} // 赠菜原因
	CancelReason           interface{} // 退菜原因
	ProductionOrderUuid    interface{} // 生产订单ID
	ProductPackageUuid     interface{} // 商品包ID
	SaleBillUuid           interface{} // 销售账单ID
	SaleOrderUuid          interface{} // 销售订单ID
	MustPlanUuid           interface{} // 必点方案ID,产品要求用这种方式标注各个必点
	DeskUuid               interface{} // 桌台ID, 默认为0是本台，大于0为合并过来的桌台
	H5OrderUuid            interface{} // 扫码订单ID，用于关联扫码订单，用于判断是否为扫码订单商品
	H5OrderProductUuid     interface{} // h5订单商品ID，用于关联h5订单商品，用于判断是否为h5订单商品
	IsAcceptOrder          interface{} // 是否已接单, 0-否 1-是。订单商品默认已接单，h5订单商品只有下单并接单后才改为已接单
	SendKitchenTime        interface{} // 送厨时间(时间戳)
	CreateTime             interface{} // 创建时间(时间戳)
	UpdateTime             interface{} // 更新时间(时间戳)
	DeleteTime             interface{} // 删除时间(时间戳)
}
