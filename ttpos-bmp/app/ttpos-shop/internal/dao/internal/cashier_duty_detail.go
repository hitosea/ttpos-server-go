// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// CashierDutyDetailDao is the data access object for the table ttpos_cashier_duty_detail.
type CashierDutyDetailDao struct {
	table    string                   // table is the underlying table name of the DAO.
	group    string                   // group is the database configuration group name of the current DAO.
	columns  CashierDutyDetailColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler       // handlers for customized model modification.
}

// CashierDutyDetailColumns defines and stores column names for the table ttpos_cashier_duty_detail.
type CashierDutyDetailColumns struct {
	Id                           string // 自增ID
	Uuid                         string // 收银交班详情ID
	StaffUuid                    string // 员工ID
	DutyNo                       string // 当班编号
	DutyStartTime                string // 当班开始时间
	DutyEndTime                  string // 当班结束时间
	TotalSales                   string // 总销售额
	TotalServiceFee              string // 总服务费
	TotalPaymentCommissionFee    string // 总支付手续费
	TotalTaxFee                  string // 总税费
	TotalProductQuantity         string // 商品数量
	TotalDiscountFee             string // 总优惠折扣
	TotalRefundFee               string // 总退款
	TotalRevenue                 string // 总营业收入
	TotalActualAmount            string // 总实收金额
	TotalRechargeAmount          string // 充值金额
	TotalGiftAmount              string // 赠送金额
	TotalGiftPoint               string // 赠送积分
	PreviousBalance              string // 上一班遗留备用金
	TotalOffCashWithdrawal       string // 下班取出现金
	TotalCashBalance             string // 本班遗留备用金
	CashDeposit                  string // 中途存入现金
	CashWithdrawal               string // 中途取出现金
	ExceptionReport              string // 异常报备
	TotalReturnFoodCount         string // 退菜次数
	TotalRefundCount             string // 退款次数
	TotalReconciliationCount     string // 反结账次数
	TotalGiftProductCount        string // 赠菜次数
	TotalFreeOrderCount          string // 免单次数
	TotalTransferProductCount    string // 转菜次数
	TotalSinglePriceChangeCount  string // 单品改价次数
	TotalOrderPriceChangeCount   string // 整单改价次数
	TotalOrderDiscoutCount       string // 整单折扣次数
	TotalZeroCheckoutCount       string // 整单结账抹零次数
	TotalOrderCount              string // 所有订单数
	TotalTableCount              string // 桌数
	TotalCustomerCount           string // 人数
	TotalMinOrderAmount          string // 最小订单金额
	TotalMaxOrderAmount          string // 最大订单金额
	TotalAverageOrderAmount      string // 平均订单金额
	TotalTableCustomerCount      string // 桌台人数
	TotalTableMinOrderAmount     string // 桌台最小订单金额
	TotalTableMaxOrderAmount     string // 桌台最大订单金额
	TotalTableAverageOrderAmount string // 桌台人均消费金额
	TotalScanOrderCount          string // 点餐订单数
	TotalScanMinOrderAmount      string // 点餐最小订单金额
	TotalScanMaxOrderAmount      string // 点餐最大订单金额
	TotalScanAverageOrderAmount  string // 点餐平均订单金额
	TotalGiftProductAmount       string // 赠菜金额
	TotalGiftProductPoint        string // 赠菜积分
	CreateTime                   string // 创建时间(时间戳)
	UpdateTime                   string // 更新时间(时间戳)
	DeleteTime                   string // 删除时间(时间戳)
}

// cashierDutyDetailColumns holds the columns for the table ttpos_cashier_duty_detail.
var cashierDutyDetailColumns = CashierDutyDetailColumns{
	Id:                           "id",
	Uuid:                         "uuid",
	StaffUuid:                    "staff_uuid",
	DutyNo:                       "duty_no",
	DutyStartTime:                "duty_start_time",
	DutyEndTime:                  "duty_end_time",
	TotalSales:                   "total_sales",
	TotalServiceFee:              "total_service_fee",
	TotalPaymentCommissionFee:    "total_payment_commission_fee",
	TotalTaxFee:                  "total_tax_fee",
	TotalProductQuantity:         "total_product_quantity",
	TotalDiscountFee:             "total_discount_fee",
	TotalRefundFee:               "total_refund_fee",
	TotalRevenue:                 "total_revenue",
	TotalActualAmount:            "total_actual_amount",
	TotalRechargeAmount:          "total_recharge_amount",
	TotalGiftAmount:              "total_gift_amount",
	TotalGiftPoint:               "total_gift_point",
	PreviousBalance:              "previous_balance",
	TotalOffCashWithdrawal:       "total_off_cash_withdrawal",
	TotalCashBalance:             "total_cash_balance",
	CashDeposit:                  "cash_deposit",
	CashWithdrawal:               "cash_withdrawal",
	ExceptionReport:              "exception_report",
	TotalReturnFoodCount:         "total_return_food_count",
	TotalRefundCount:             "total_refund_count",
	TotalReconciliationCount:     "total_reconciliation_count",
	TotalGiftProductCount:        "total_gift_product_count",
	TotalFreeOrderCount:          "total_free_order_count",
	TotalTransferProductCount:    "total_transfer_product_count",
	TotalSinglePriceChangeCount:  "total_single_price_change_count",
	TotalOrderPriceChangeCount:   "total_order_price_change_count",
	TotalOrderDiscoutCount:       "total_order_discout_count",
	TotalZeroCheckoutCount:       "total_zero_checkout_count",
	TotalOrderCount:              "total_order_count",
	TotalTableCount:              "total_table_count",
	TotalCustomerCount:           "total_customer_count",
	TotalMinOrderAmount:          "total_min_order_amount",
	TotalMaxOrderAmount:          "total_max_order_amount",
	TotalAverageOrderAmount:      "total_average_order_amount",
	TotalTableCustomerCount:      "total_table_customer_count",
	TotalTableMinOrderAmount:     "total_table_min_order_amount",
	TotalTableMaxOrderAmount:     "total_table_max_order_amount",
	TotalTableAverageOrderAmount: "total_table_average_order_amount",
	TotalScanOrderCount:          "total_scan_order_count",
	TotalScanMinOrderAmount:      "total_scan_min_order_amount",
	TotalScanMaxOrderAmount:      "total_scan_max_order_amount",
	TotalScanAverageOrderAmount:  "total_scan_average_order_amount",
	TotalGiftProductAmount:       "total_gift_product_amount",
	TotalGiftProductPoint:        "total_gift_product_point",
	CreateTime:                   "create_time",
	UpdateTime:                   "update_time",
	DeleteTime:                   "delete_time",
}

// NewCashierDutyDetailDao creates and returns a new DAO object for table data access.
func NewCashierDutyDetailDao(handlers ...gdb.ModelHandler) *CashierDutyDetailDao {
	return &CashierDutyDetailDao{
		group:    "default",
		table:    "ttpos_cashier_duty_detail",
		columns:  cashierDutyDetailColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *CashierDutyDetailDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *CashierDutyDetailDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *CashierDutyDetailDao) Columns() CashierDutyDetailColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *CashierDutyDetailDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *CashierDutyDetailDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *CashierDutyDetailDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
