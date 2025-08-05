// ==========================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// ==========================================================================

package internal

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// SaleOrderBuffetCustomerTypeDao is the data access object for the table ttpos_sale_order_buffet_customer_type.
type SaleOrderBuffetCustomerTypeDao struct {
	table    string                             // table is the underlying table name of the DAO.
	group    string                             // group is the database configuration group name of the current DAO.
	columns  SaleOrderBuffetCustomerTypeColumns // columns contains all the column names of Table for convenient usage.
	handlers []gdb.ModelHandler                 // handlers for customized model modification.
}

// SaleOrderBuffetCustomerTypeColumns defines and stores column names for the table ttpos_sale_order_buffet_customer_type.
type SaleOrderBuffetCustomerTypeColumns struct {
	Id                          string // 自增ID
	Uuid                        string // 销售订单顾客类型ID
	Name                        string // 顾客类型名称
	Num                         string // 人数
	SalePrice                   string // 原始单价（单人，折前价）。自助餐顾客类型原价,下单后价格不受后台改变
	Price                       string // 最终单价（折后价），只进行自定义打折，不进行会员打折
	CustomDiscountRate          string // 自定义折扣率, 值为0-1之间(0-100%)
	CustomDiscountFee           string // 自定义折扣金额（单人）。自定义折扣金额（单人）=自助餐顾客类型原价*自定义折扣率
	TaxRate                     string // 税率,值为0-1之间.加购时记录税率,结账时再重新核算
	ServiceTaxFee               string // 服务费税费（单人）,0-不收取税费；收取时，服务费税费=服务费*税率
	TaxFee                      string // 自助餐顾客类型税费（单人）。自助餐顾客类型已含税时，税费=自助餐顾客类型原价*(1-1/(1+税率))；自助餐顾客类型未含税时，税费=自助餐顾客类型原价*税率
	ServiceFee                  string // 服务费（单人）,0-固定服务费 大于0-按比例收服务费；自助餐顾客类型已含税时，服务费=(自助餐顾客类型原价-自助餐顾客类型税费)*服务费比例；自助餐顾客类型未含税时，服务费=自助餐顾客类型原价*服务费比例
	TotalPrice                  string // 应收金额(单人)。商品已含税时，应收金额(单人)=(最终单价-商品税费)+服务费+总税费；商品未含税时，应收金额(单商品)=最终单价+服务费+总税费
	SaleOrderUuid               string // 销售订单ID
	BuffetPackageUuid           string // 自助餐套餐ID
	BuffetCustomerTypePriceUuid string // 自助餐客户类型价格ID
	CreateTime                  string // 创建时间(时间戳)
	UpdateTime                  string // 更新时间(时间戳)
	DeleteTime                  string // 删除时间(时间戳)
}

// saleOrderBuffetCustomerTypeColumns holds the columns for the table ttpos_sale_order_buffet_customer_type.
var saleOrderBuffetCustomerTypeColumns = SaleOrderBuffetCustomerTypeColumns{
	Id:                          "id",
	Uuid:                        "uuid",
	Name:                        "name",
	Num:                         "num",
	SalePrice:                   "sale_price",
	Price:                       "price",
	CustomDiscountRate:          "custom_discount_rate",
	CustomDiscountFee:           "custom_discount_fee",
	TaxRate:                     "tax_rate",
	ServiceTaxFee:               "service_tax_fee",
	TaxFee:                      "tax_fee",
	ServiceFee:                  "service_fee",
	TotalPrice:                  "total_price",
	SaleOrderUuid:               "sale_order_uuid",
	BuffetPackageUuid:           "buffet_package_uuid",
	BuffetCustomerTypePriceUuid: "buffet_customer_type_price_uuid",
	CreateTime:                  "create_time",
	UpdateTime:                  "update_time",
	DeleteTime:                  "delete_time",
}

// NewSaleOrderBuffetCustomerTypeDao creates and returns a new DAO object for table data access.
func NewSaleOrderBuffetCustomerTypeDao(handlers ...gdb.ModelHandler) *SaleOrderBuffetCustomerTypeDao {
	return &SaleOrderBuffetCustomerTypeDao{
		group:    "default",
		table:    "ttpos_sale_order_buffet_customer_type",
		columns:  saleOrderBuffetCustomerTypeColumns,
		handlers: handlers,
	}
}

// DB retrieves and returns the underlying raw database management object of the current DAO.
func (dao *SaleOrderBuffetCustomerTypeDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of the current DAO.
func (dao *SaleOrderBuffetCustomerTypeDao) Table() string {
	return dao.table
}

// Columns returns all column names of the current DAO.
func (dao *SaleOrderBuffetCustomerTypeDao) Columns() SaleOrderBuffetCustomerTypeColumns {
	return dao.columns
}

// Group returns the database configuration group name of the current DAO.
func (dao *SaleOrderBuffetCustomerTypeDao) Group() string {
	return dao.group
}

// Ctx creates and returns a Model for the current DAO. It automatically sets the context for the current operation.
func (dao *SaleOrderBuffetCustomerTypeDao) Ctx(ctx context.Context) *gdb.Model {
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
func (dao *SaleOrderBuffetCustomerTypeDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}
