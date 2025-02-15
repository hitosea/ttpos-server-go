package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPaymentMethodRepo 定义仓库接口
type IPaymentMethodRepo interface {
	GetByUuid(uuid uint64) model.PaymentMethod
}

// paymentMethodRepo 仓库
type paymentMethodRepo struct {
	db *gorm.DB
}

// NewPaymentMethodRepo 创建新仓库
func NewPaymentMethodRepo(db *gorm.DB) IPaymentMethodRepo {
	return NewPaymentMethodRepoImpl(db)
}

// NewPaymentMethodRepoImpl 创建新仓库实现
func NewPaymentMethodRepoImpl(db *gorm.DB) IPaymentMethodRepo {
	return &paymentMethodRepo{db: db}
}

// GetByUuid 根据Uuid获取支付方式
func (r *paymentMethodRepo) GetByUuid(uuid uint64) model.PaymentMethod {
	var paymentMethod model.PaymentMethod
	r.db.Model(&model.PaymentMethod{}).Where("uuid = ?", uuid).First(&paymentMethod)
	return paymentMethod
}
