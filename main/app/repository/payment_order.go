package repository

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPaymentOrderRepo 定义仓库接口
type IPaymentOrderRepo interface {
	IPaymentOrderQueryRepo
	WhereRelatedUuid(uuid uint64) DBOption
	WhereRelatedUuids(uuids []uint64) DBOption
	WhereStatus(status uint) DBOption
	WhereUuid(uuid uint64) DBOption
	WhereRelatedType(relatedType uint64) DBOption
	WherePaymentMethodUuid(paymentMethodUuid uint64) DBOption

	WithPaymentMethod() DBOption       // 关联支付方式
	WithMemberRechargeOrder() DBOption // 关联会员充值订单

	UpdatePaymentOrderRecord(paymentOrder model.PaymentOrder) error
	Create(order model.PaymentOrder) (model.PaymentOrder, error)   // 创建支付订单
	UpdateOrCreatePaymentOrderRecord(obj model.PaymentOrder) error // 没有主键时创建，有主键时更新
	Update(uuid uint64, vars map[string]any) error                 // 更新支付订单金额
	CreateRefundOrderRecord(refundOrder model.RefundOrder) error   // 创建退款单
	DeletePaymentOrderRecord(uuid uint64) error                    // 删除支付订单
}

// IPaymentOrderQueryRepo 定义仓库查询接口
type IPaymentOrderQueryRepo interface {
	GetPaymentOrder(opts ...DBOption) (model.PaymentOrder, error)      // 获取详细信息
	GetPaymentOrderInfo(opts ...DBOption) (*model.PaymentOrder, error) // 获取详细信息
	GetPaymentOrderList(opts ...DBOption) ([]model.PaymentOrder, error)
	GetPaymentOrderRecordList(opts ...DBOption) ([]*model.PaymentOrder, error)

	GetPaymentOrderRecord(paymentOrderUuid uint64) (*model.PaymentOrder, error)
	GetPaymentOrderListBySaleOrderUuid(saleOrderUuid uint64) ([]*model.PaymentOrder, error)
}

// paymentOrderRepo 仓库
type paymentOrderRepo struct {
	db *gorm.DB
}

// NewPaymentOrderRepo 创建新仓库
func NewPaymentOrderRepo(db *gorm.DB) IPaymentOrderRepo {
	return NewPaymentOrderRepoImpl(db)
}

// NewPaymentOrderRepoImpl 创建新仓库实现
func NewPaymentOrderRepoImpl(db *gorm.DB) IPaymentOrderRepo {
	return &paymentOrderRepo{db: db}
}

// GetPaymentOrder 获取详细信息
func (r *paymentOrderRepo) GetPaymentOrder(opts ...DBOption) (model.PaymentOrder, error) {
	var paymentOrder model.PaymentOrder
	db := r.db
	for _, w := range opts {
		db = w(db)
	}
	db = CommonRepo.WhereBySoftDelete()(db)
	err := db.Order("id asc").First(&paymentOrder).Error
	return paymentOrder, errors.WithMessage(err)
}

// GetPaymentOrderInfo 获取详细信息
func (r *paymentOrderRepo) GetPaymentOrderInfo(opts ...DBOption) (*model.PaymentOrder, error) {
	var paymentOrder model.PaymentOrder
	db := r.db
	for _, w := range opts {
		db = w(db)
	}
	db = CommonRepo.WhereBySoftDelete()(db)
	err := db.Order("id asc").First(&paymentOrder).Error
	return &paymentOrder, errors.WithMessage(err)
}

// GetPaymentOrderList 获取列表
func (r *paymentOrderRepo) GetPaymentOrderList(opts ...DBOption) ([]model.PaymentOrder, error) {
	var paymentOrders []model.PaymentOrder
	db := r.db
	for _, w := range opts {
		db = w(db)
	}
	db = CommonRepo.WhereBySoftDelete()(db)
	err := db.Order("id asc").Find(&paymentOrders).Error
	return paymentOrders, errors.WithMessage(err)
}

// GetPaymentOrderRecordList 获取支付订单记录列表
func (r *paymentOrderRepo) GetPaymentOrderRecordList(opts ...DBOption) ([]*model.PaymentOrder, error) {
	var paymentOrders []*model.PaymentOrder
	db := r.db
	for _, w := range opts {
		db = w(db)
	}
	db = CommonRepo.WhereBySoftDelete()(db)
	err := db.Order("id asc").Find(&paymentOrders).Error
	return paymentOrders, errors.WithMessage(err)
}

// GetPaymentOrderRecord 获取支付订单记录
func (r *paymentOrderRepo) GetPaymentOrderRecord(paymentOrderUuid uint64) (*model.PaymentOrder, error) {
	paymentOrder, err := r.GetPaymentOrder(
		CommonRepo.WhereByUuid(paymentOrderUuid),
		CommonRepo.WhereBySoftDelete(),
		r.WithPaymentMethod(),
	)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &paymentOrder, nil
}

// GetPaymentOrderListBySaleOrderUuid 销售订单支付订单列表
func (r *paymentOrderRepo) GetPaymentOrderListBySaleOrderUuid(saleOrderUuid uint64) ([]*model.PaymentOrder, error) {
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
func (r *paymentOrderRepo) UpdatePaymentOrderRecord(paymentOrder model.PaymentOrder) error {
	paymentOrder.SetNil() // 将关联对象置空，为了不更新这些关联的对象
	if paymentOrder.NoPrimaryKey() {
		return errors.New("PaymentOrder不能没有ID或UUID")
	}
	return r.db.Model(&model.PaymentOrder{}).Select("*").Where("uuid = ?", paymentOrder.Uuid).Updates(&paymentOrder).Error
}

func (r *paymentOrderRepo) Create(order model.PaymentOrder) (model.PaymentOrder, error) {
	order.SetNil()
	err := r.db.Model(&model.PaymentOrder{}).Create(&order).Error
	return order, errors.WithMessage(err)
}

func (r *paymentOrderRepo) UpdateOrCreatePaymentOrderRecord(obj model.PaymentOrder) error {
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
func (r *paymentOrderRepo) Update(uuid uint64, vars map[string]any) error {
	err := r.db.Model(&model.PaymentOrder{}).Where("uuid = ?", uuid).Updates(vars).Error
	return errors.WithMessage(err)
}

// CreateRefundOrderRecord 创建退款单
func (r *paymentOrderRepo) CreateRefundOrderRecord(refundOrder model.RefundOrder) error {
	err := r.db.Model(&model.RefundOrder{}).Create(&refundOrder).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// WhereRelatedUuid 根据related_uuid(销售订单、充值订单)查询
func (r *paymentOrderRepo) WhereRelatedUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("related_uuid = ?", uuid)
	}
}

// WhereRelatedType 根据related_type(销售订单、充值订单)查询
func (r *paymentOrderRepo) WhereRelatedType(relatedType uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("related_type = ?", relatedType)
	}
}

// WherePaymentMethodUuid 根据支付方式Uuid查询
func (r *paymentOrderRepo) WherePaymentMethodUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("payment_method_uuid = ?", uuid)
	}
}

// WhereUuid 根据支付订单Uuid查询
func (r *paymentOrderRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

// WhereRelatedUuids 根据related_uuids(销售订单、充值订单)查询
func (r *paymentOrderRepo) WhereRelatedUuids(uuids []uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("related_uuid in (?)", uuids)
	}
}

// WhereStatus 根据状态查询
func (r *paymentOrderRepo) WhereStatus(status uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

// WithPaymentMethod 关联支付方式
func (r *paymentOrderRepo) WithPaymentMethod() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("PaymentMethod")
	}
}

// WithMemberRechargeOrder 关联会员充值订单
func (r *paymentOrderRepo) WithMemberRechargeOrder() DBOption { // 关联会员充值订单
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("MemberRechargeOrder")
	}
}

// DeletePaymentOrderRecord 删除支付订单
func (r *paymentOrderRepo) DeletePaymentOrderRecord(uuid uint64) error {
	err := r.db.Model(&model.PaymentOrder{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error
	return errors.WithMessage(err)
}
