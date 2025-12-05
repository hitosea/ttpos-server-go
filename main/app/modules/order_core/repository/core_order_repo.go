package repository

import (
	"ttpos-server-go/app/modules/order_core/model"

	"gorm.io/gorm"
)

type CoreOrderRepo struct {
	db *gorm.DB
}

func NewCoreOrderRepo(db *gorm.DB) ICoreOrderRepo {
	return &CoreOrderRepo{db: db}
}

// Bill 相关
func (r *CoreOrderRepo) CreateBill(bill *model.CoreSaleBill) error {
	return r.db.Create(bill).Error
}

func (r *CoreOrderRepo) UpdateBillStatus(uuid uint64, status uint) error {
	return r.db.Model(&model.CoreSaleBill{}).Where("uuid = ?", uuid).Update("status", status).Error
}

func (r *CoreOrderRepo) GetBillByUuid(uuid uint64) (*model.CoreSaleBill, error) {
	var bill model.CoreSaleBill
	if err := r.db.Where("uuid = ?", uuid).First(&bill).Error; err != nil {
		return nil, err
	}
	return &bill, nil
}

// Order 相关
func (r *CoreOrderRepo) CreateOrder(order *model.CoreSaleOrder) error {
	return r.db.Create(order).Error
}

func (r *CoreOrderRepo) UpdateOrderStatus(uuid uint64, status uint) error {
	return r.db.Model(&model.CoreSaleOrder{}).Where("uuid = ?", uuid).Update("status", status).Error
}

func (r *CoreOrderRepo) GetOrderByUuid(uuid uint64) (*model.CoreSaleOrder, error) {
	var order model.CoreSaleOrder
	if err := r.db.Where("uuid = ?", uuid).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *CoreOrderRepo) GetOrdersByBillUuid(billUuid uint64) ([]*model.CoreSaleOrder, error) {
	var orders []*model.CoreSaleOrder
	if err := r.db.Where("sale_bill_uuid = ?", billUuid).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// OrderProduct 相关
func (r *CoreOrderRepo) CreateOrderProduct(product *model.CoreSaleOrderProduct) error {
	return r.db.Create(product).Error
}

func (r *CoreOrderRepo) GetOrderProductsByOrderUuid(orderUuid uint64) ([]*model.CoreSaleOrderProduct, error) {
	var products []*model.CoreSaleOrderProduct
	if err := r.db.Where("sale_order_uuid = ?", orderUuid).Find(&products).Error; err != nil {
		return nil, err
	}
	return products, nil
}
