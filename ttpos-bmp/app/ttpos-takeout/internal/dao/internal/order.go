// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// OrderDao is the data access object for the table takeout_order.
type OrderDao struct {
	table    string             // table is the underlying table name of the DAO.
	group    string             // group is the database configuration group name of the current DAO.
	columns  OrderColumns       // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler // handlers for customized model modification.
}

// OrderColumns defines and stores column names for the table takeout_order.
type OrderColumns struct {
	Id                string // 主键
	Uuid              string // 系统唯一ID
	MerchantId        string // 商户ID (Partner Merchant ID)
	PartnerOrderId    string // 平台订单号 (Grab Order ID)
	ShopRefNo         string // TTPOS内部订单号
	ShortOrderNumber  string // 短单号
	ProviderName      string // 渠道: grab, foodpanda
	OrderType         string // 订单类型: DeliveryByGrab, DeliveryByRestaurant, TakeAway, DineIn
	OrderTime         string // 下单时间
	OrderStatus       string // 订单状态: PENDING, ACCEPTED, PREPARING, READY, COMPLETED, CANCELLED
	ScheduledTime     string // 预约送达时间
	Currency          string // 货币代码
	Subtotal          string // 商品小计
	TotalAmount       string // 总金额
	MerchantCharge    string // 商户收取金额
	TaxAmount         string // 税费
	DiscountAmount    string // 折扣金额
	MerchantFundPromo string // 商户承担促销金额
	PaymentType       string // 支付方式: CASH, CASHLESS
	IsMexEditOrder    string // 是否为商户编辑订单
	Cutlery           string // 是否需要餐具
	EaterCount        string // 用餐人数
	CustomerName      string // 顾客姓名
	CustomerPhone     string // 顾客电话 (虚拟号)
	DeliveryAddress   string // 配送地址 (JSON)
	DriverEta         string // 骑手预计到达时间(分钟)
	ReadyTime         string // 商家预估准备完成时间
	Note              string // 订单备注
	RawData           string // 原始JSON数据
	CreatedAt         string // 创建时间
	UpdatedAt         string // 更新时间
	DeletedAt         string // 软删除
}

// orderColumns holds the columns for the table takeout_order.
var orderColumns = OrderColumns{
	Id:                "id",
	Uuid:              "uuid",
	MerchantId:        "merchant_id",
	PartnerOrderId:    "partner_order_id",
	ShopRefNo:         "shop_ref_no",
	ShortOrderNumber:  "short_order_number",
	ProviderName:      "provider_name",
	OrderType:         "order_type",
	OrderTime:         "order_time",
	OrderStatus:       "order_status",
	ScheduledTime:     "scheduled_time",
	Currency:          "currency",
	Subtotal:          "subtotal",
	TotalAmount:       "total_amount",
	MerchantCharge:    "merchant_charge",
	TaxAmount:         "tax_amount",
	DiscountAmount:    "discount_amount",
	MerchantFundPromo: "merchant_fund_promo",
	PaymentType:       "payment_type",
	IsMexEditOrder:    "is_mex_edit_order",
	Cutlery:           "cutlery",
	EaterCount:        "eater_count",
	CustomerName:      "customer_name",
	CustomerPhone:     "customer_phone",
	DeliveryAddress:   "delivery_address",
	DriverEta:         "driver_eta",
	ReadyTime:         "ready_time",
	Note:              "note",
	RawData:           "raw_data",
	CreatedAt:         "created_at",
	UpdatedAt:         "updated_at",
	DeletedAt:         "deleted_at",
}

// NewOrderDao creates and returns a new DAO object for table data access.
func NewOrderDao(handlers ...gdb.ModelHandler) *OrderDao {
	return &OrderDao{
		group:    "default",
		table:    "takeout_order",
		columns:  orderColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *OrderDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *OrderDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *OrderDao) Columns() OrderColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *OrderDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *OrderDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *OrderDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
