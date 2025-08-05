// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderProductDao is the data access object for the table ttpos_sale_order_product.
type SaleOrderProductDao struct {
	table    string                  // table is the underlying table name of the DAO.
	group    string                  // group is the database configuration group name of the current DAO.
	columns  SaleOrderProductColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler      // handlers for customized model modification.
}

// SaleOrderProductColumns defines and stores column names for the table ttpos_sale_order_product.
type SaleOrderProductColumns struct {
	OpenMemberDiscount     string // 是否开启会员折扣, 0-否 1-是。添加商品时记录下状态不受后台改变，结账时检查是否改变
	Id                     string // 自增ID
	Uuid                   string // 销售订单商品ID
	Name                   string // 商品名称
	FlavorName             string // 规格名称
	Sign                   string // 商品签名,规格、属性、加料、是否改价、是否赠菜、送厨批次、销售价相同的商品签名相同,用于取消拆单时合并商品
	MultiLanguageNameUuid  string // 多语言名称ID
	Num                    string // 商品数量。不能减为0，当数量为1再减时，标记删除
	ImageFileUuid          string // 商品图片ID
	FlavorPrice            string // 规格原价（单商品）,仅某规格商品的原价
	SaucePrice             string // 小料价（单商品）,所有小料的价格之和
	ProductPrice           string // 原始单价（单商品）,规格原价+小料价
	ChangePriceTime        string // 改价时间(时间戳),用于判断是否改价和不同时间改价的商品不合并
	IsBuffet               string // 是否为自助餐商品,0-否 1-是. 如果是自助餐商品，则sale_price为0
	SalePrice              string // 销售价（单商品，折前价）,当自定义价格时，销售价=自定义价格,否则销售价=原始单价
	SalePriceNoTax         string // 销售价,未含税价格（折前）
	TaxRate                string // 税率,单位%.加购时记录税率,结账时再重新核算
	MemberDiscountRate     string // 会员折扣率(0-100%)
	MemberCardDiscountRate string // 会员卡折扣率(0-100%)
	CustomDiscountRate     string // 自定义折扣率(0-100%)
	MemberDiscountPrice    string // 会员折扣后的价格（单商品）=销售价*会员折扣率*会员卡折扣率
	Price                  string // 最终单价(单商品，会员、会员卡和优惠折扣后，折后价)。销售价*折扣率
	ServiceTaxFee          string // 服务费税费（单商品）,0-不收取税费；收取时，服务费税费=服务费*税率
	TaxFee                 string // 商品税费（单商品）。商品已含税时，税费=规格原价*(1-1/(1+税率))；商品未含税时，税费=原始单价*税率
	ServiceFee             string // 服务费（单商品）,0-固定服务费 大于0-按比例收服务费；商品已含税时，服务费=(最终单价-商品税费)*服务费比例；商品未含税时，服务费=最终单价*服务费比例
	TotalPrice             string // 应收金额(单商品)。商品已含税时，应收金额(单商品)=(最终单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=最终单价+服务费+总税费
	OriginTotalPrice       string // 应收金额(单商品)。商品已含税时，应收金额(单商品)=(销售价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=销售价+服务费+总税费
	DiscountFee            string // 打折金额（单商品）=销售价-最终单价。校验：打折金额=会员折扣金额+自定义折扣金额
	MemberDiscountFee      string // 会员折扣金额（单商品）=销售价*（1-会员折扣率*会员卡折扣率）
	CustomDiscountFee      string // 自定义折扣金额（单商品）。自定义折扣金额（单商品）=会员折扣后的价格（单商品）*(1-自定义折扣率) 。校验：自定义折扣金额（单商品）=销售价 - 最终单价（单商品）-会员折扣金额（单商品）；注意，不能这样算，自定义折扣金额（单商品）=销售价*(1-自定义折扣率)
	Status                 string // 状态, 0-未送厨 1-已送厨
	IsRequire              string // 是否必点商品 0-否 1-是。用于在前端显示必点图标
	DeductStockType        string // 库存计算方式,0-下单减库存 1-付款减库存。加购商品时记录，不受后台影响，用于减少查询次数
	DeductStockTime        string // 减库存的时间(时间戳)，0-未减库存。标记是否已减库存，用于取消订单时恢复库存、避免重复减库存、避免漏减库存
	Remark                 string // 备注，顾客对商品的备注信息
	GiftTime               string // 赠菜时间(时间戳),用于判断是否赠菜和不同时间赠送的商品不合并
	CancelTime             string // 退菜时间(时间戳)
	GiftReason             string // 赠菜原因
	CancelReason           string // 退菜原因
	ProductionOrderUuid    string // 生产订单ID
	ProductPackageUuid     string // 商品包ID
	SaleBillUuid           string // 销售账单ID
	SaleOrderUuid          string // 销售订单ID
	MustPlanUuid           string // 必点方案ID,产品要求用这种方式标注各个必点
	DeskUuid               string // 桌台ID, 默认为0是本台，大于0为合并过来的桌台
	H5OrderUuid            string // 扫码订单ID，用于关联扫码订单，用于判断是否为扫码订单商品
	H5OrderProductUuid     string // h5订单商品ID，用于关联h5订单商品，用于判断是否为h5订单商品
	IsAcceptOrder          string // 是否已接单, 0-否 1-是。订单商品默认已接单，h5订单商品只有下单并接单后才改为已接单
	SendKitchenTime        string // 送厨时间(时间戳)
	CreateTime             string // 创建时间(时间戳)
	UpdateTime             string // 更新时间(时间戳)
	DeleteTime             string // 删除时间(时间戳)
}

// saleOrderProductColumns holds the columns for the table ttpos_sale_order_product.
var saleOrderProductColumns = SaleOrderProductColumns{
	OpenMemberDiscount:     "open_member_discount",
	Id:                     "id",
	Uuid:                   "uuid",
	Name:                   "name",
	FlavorName:             "flavor_name",
	Sign:                   "sign",
	MultiLanguageNameUuid:  "multi_language_name_uuid",
	Num:                    "num",
	ImageFileUuid:          "image_file_uuid",
	FlavorPrice:            "flavor_price",
	SaucePrice:             "sauce_price",
	ProductPrice:           "product_price",
	ChangePriceTime:        "change_price_time",
	IsBuffet:               "is_buffet",
	SalePrice:              "sale_price",
	SalePriceNoTax:         "sale_price_no_tax",
	TaxRate:                "tax_rate",
	MemberDiscountRate:     "member_discount_rate",
	MemberCardDiscountRate: "member_card_discount_rate",
	CustomDiscountRate:     "custom_discount_rate",
	MemberDiscountPrice:    "member_discount_price",
	Price:                  "price",
	ServiceTaxFee:          "service_tax_fee",
	TaxFee:                 "tax_fee",
	ServiceFee:             "service_fee",
	TotalPrice:             "total_price",
	OriginTotalPrice:       "origin_total_price",
	DiscountFee:            "discount_fee",
	MemberDiscountFee:      "member_discount_fee",
	CustomDiscountFee:      "custom_discount_fee",
	Status:                 "status",
	IsRequire:              "is_require",
	DeductStockType:        "deduct_stock_type",
	DeductStockTime:        "deduct_stock_time",
	Remark:                 "remark",
	GiftTime:               "gift_time",
	CancelTime:             "cancel_time",
	GiftReason:             "gift_reason",
	CancelReason:           "cancel_reason",
	ProductionOrderUuid:    "production_order_uuid",
	ProductPackageUuid:     "product_package_uuid",
	SaleBillUuid:           "sale_bill_uuid",
	SaleOrderUuid:          "sale_order_uuid",
	MustPlanUuid:           "must_plan_uuid",
	DeskUuid:               "desk_uuid",
	H5OrderUuid:            "h5_order_uuid",
	H5OrderProductUuid:     "h5_order_product_uuid",
	IsAcceptOrder:          "is_accept_order",
	SendKitchenTime:        "send_kitchen_time",
	CreateTime:             "create_time",
	UpdateTime:             "update_time",
	DeleteTime:             "delete_time",
}

// NewSaleOrderProductDao creates and returns a new DAO object for table data access.
func NewSaleOrderProductDao(handlers ...gdb.ModelHandler) *SaleOrderProductDao {
	return &SaleOrderProductDao{
		group:    "default",
		table:    "ttpos_sale_order_product",
		columns:  saleOrderProductColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SaleOrderProductDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SaleOrderProductDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SaleOrderProductDao) Columns() SaleOrderProductColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SaleOrderProductDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SaleOrderProductDao) Ctx(ctx context.Context) *gdb.Model {
	model := dao.DB().Model(dao.table)
	for _, handler := range dao.handlers {
		model = handler(model)
	}
	return model.Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rolls back the transaction and returns the error if function f returns a non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note: Do not commit or roll back the transaction in function f,
// as it is automatically handled by this function.
func (dao *SaleOrderProductDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
