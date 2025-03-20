package repository

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ILlPaymentOrderRepo 定义仓库接口
type ILlPaymentOrderRepo interface {
	GetPaymentOrder(opts ...DBOption) (model.LlPaymentOrder, error) // 获取详细信息
	GetPaymentOrderList(opts ...DBOption) ([]model.LlPaymentOrder, error)
	GetPaymentOrderRecordList(opts ...DBOption) ([]*model.LlPaymentOrder, error)

	GetPaymentOrderRecord(paymentOrderUuid uint64) (*model.LlPaymentOrder, error)
	GetPaymentOrderListBySaleOrderUuid(saleOrderUuid uint64) ([]*model.LlPaymentOrder, error)
	UpdatePaymentOrderRecord(paymentOrder model.LlPaymentOrder) error
	Create(order model.LlPaymentOrder) (model.LlPaymentOrder, error) // 创建支付订单
	UpdateOrCreatePaymentOrderRecord(obj model.LlPaymentOrder) error // 没有主键时创建，有主键时更新
	Update(uuid uint64, vars map[string]any) error                   // 更新支付订单金额
}

// llPaymentOrderRepo 仓库
type llPaymentOrderRepo struct {
	db *gorm.DB
}

// NewLlPaymentOrderRepo 创建新仓库
func NewLlPaymentOrderRepo(db *gorm.DB) ILlPaymentOrderRepo {
	return NewLlPaymentOrderRepoImpl(db)
}

// NewLlPaymentOrderRepoImpl 创建新仓库实现
func NewLlPaymentOrderRepoImpl(db *gorm.DB) ILlPaymentOrderRepo {
	return &llPaymentOrderRepo{db: db}
}

// GetPaymentOrder 获取详细信息
func (r *llPaymentOrderRepo) GetPaymentOrder(opts ...DBOption) (model.LlPaymentOrder, error) {
	var paymentOrder model.LlPaymentOrder
	db := r.db
	for _, w := range opts {
		db = w(db)
	}
	db = CommonRepo.WhereBySoftDelete()(db)
	err := db.Order("id asc").First(&paymentOrder).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return paymentOrder, errors.WithMessage(err)
	}
	return paymentOrder, nil
}

// GetPaymentOrderList 获取列表
func (r *llPaymentOrderRepo) GetPaymentOrderList(opts ...DBOption) ([]model.LlPaymentOrder, error) {
	var paymentOrders []model.LlPaymentOrder
	db := r.db
	for _, w := range opts {
		db = w(db)
	}
	db = CommonRepo.WhereBySoftDelete()(db)
	err := db.Order("id asc").Find(&paymentOrders).Error
	return paymentOrders, errors.WithMessage(err)
}

// GetPaymentOrderRecordList 获取支付订单记录列表
func (r *llPaymentOrderRepo) GetPaymentOrderRecordList(opts ...DBOption) ([]*model.LlPaymentOrder, error) {
	var paymentOrders []*model.LlPaymentOrder
	db := r.db
	for _, w := range opts {
		db = w(db)
	}
	db = CommonRepo.WhereBySoftDelete()(db)
	err := db.Order("id asc").Find(&paymentOrders).Error
	return paymentOrders, errors.WithMessage(err)
}

// GetPaymentOrderRecord 获取支付订单记录
func (r *llPaymentOrderRepo) GetPaymentOrderRecord(paymentOrderUuid uint64) (*model.LlPaymentOrder, error) {
	paymentOrder, err := r.GetPaymentOrder(
		CommonRepo.WhereByUuid(paymentOrderUuid),
		CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &paymentOrder, nil
}

// GetPaymentOrderListBySaleOrderUuid 销售订单支付订单列表
func (r *llPaymentOrderRepo) GetPaymentOrderListBySaleOrderUuid(saleOrderUuid uint64) ([]*model.LlPaymentOrder, error) {
	paymentOrders, err := r.GetPaymentOrderRecordList(
		CommonRepo.WhereByRelatedUuid(saleOrderUuid),
		CommonRepo.WhereByStatus(constant.PaymentOrderStatusPaid),
		CommonRepo.WhereByRelatedType(constant.PaymentOrderRelatedTypeSaleOrder),
		CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return paymentOrders, nil
}

// UpdatePaymentOrderRecord 更新支付订单记录
func (r *llPaymentOrderRepo) UpdatePaymentOrderRecord(paymentOrder model.LlPaymentOrder) error {
	if paymentOrder.NoPrimaryKey() {
		return errors.New("PaymentOrder不能没有ID或UUID")
	}
	return r.db.Model(&model.LlPaymentOrder{}).Select("*").Where("uuid = ?", paymentOrder.Uuid).Updates(&paymentOrder).Error
}

func (r *llPaymentOrderRepo) Create(order model.LlPaymentOrder) (model.LlPaymentOrder, error) {
	err := r.db.Model(&model.LlPaymentOrder{}).Create(&order).Error
	return order, errors.WithMessage(err)
}

func (r *llPaymentOrderRepo) UpdateOrCreatePaymentOrderRecord(obj model.LlPaymentOrder) error {
	if obj.NoPrimaryKey() {
		_, err := r.Create(obj)
		return errors.WithMessage(err)
	}
	// 如果标记支付订单需要更新才更新该支付订单
	if obj.GetUpdate() {
		err := r.UpdatePaymentOrderRecord(obj)
		if err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}

// Update 更新支付订单
func (r *llPaymentOrderRepo) Update(uuid uint64, vars map[string]any) error {
	err := r.db.Model(&model.LlPaymentOrder{}).Where("uuid = ?", uuid).Updates(vars).Error
	return errors.WithMessage(err)
}
