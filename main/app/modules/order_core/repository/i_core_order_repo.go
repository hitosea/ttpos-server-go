package repository

import "ttpos-server-go/app/modules/order_core/model"

// ICoreOrderRepo 核心订单仓储接口
type ICoreOrderRepo interface {
	// Bill 相关
	CreateBill(bill *model.CoreSaleBill) error
	UpdateBillStatus(uuid uint64, status uint) error
	GetBillByUuid(uuid uint64) (*model.CoreSaleBill, error)

	// Order 相关
	CreateOrder(order *model.CoreSaleOrder) error
	UpdateOrderStatus(uuid uint64, status uint) error
	GetOrderByUuid(uuid uint64) (*model.CoreSaleOrder, error)
	GetOrdersByBillUuid(billUuid uint64) ([]*model.CoreSaleOrder, error)

	// OrderProduct 相关
	CreateOrderProduct(product *model.CoreSaleOrderProduct) error
	GetOrderProductsByOrderUuid(orderUuid uint64) ([]*model.CoreSaleOrderProduct, error)
}
