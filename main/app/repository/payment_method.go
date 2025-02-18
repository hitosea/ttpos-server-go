package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPaymentMethodRepo 定义仓库接口
type IPaymentMethodRepo interface {
	WhereUuid(uuid uint64) DBOption
	GetPaymentMethod(opts ...DBOption) model.PaymentMethod
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

// GetPaymentMethod 根据Uuid获取支付方式
func (r *paymentMethodRepo) GetPaymentMethod(opts ...DBOption) model.PaymentMethod {
	var paymentMethod model.PaymentMethod
	db := r.db
	for _, opt := range opts {
		db = opt(db)
	}
	db.Model(&model.PaymentMethod{}).First(&paymentMethod)
	return paymentMethod
}

func (r *paymentMethodRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}
