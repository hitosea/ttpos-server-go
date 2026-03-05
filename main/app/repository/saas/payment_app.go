package saas

import (
	"gorm.io/gorm"

	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
)

type paymentAppRepo struct {
	db *gorm.DB
}

type IPaymentAppRepo interface {
	GetPaymentAppCompanyUuid(uuid uint64) (*model.PaymentApp, error)
	UpdatePaymentApp(data map[string]any, companyUuid uint64) error
	Create(paymentApp *model.PaymentApp) error
}

func NewPaymentAppRepo(db *gorm.DB) IPaymentAppRepo {
	return NewPaymentAppRepoImpl(db)
}

func NewPaymentAppRepoImpl(db *gorm.DB) IPaymentAppRepo {
	return &paymentAppRepo{db: db}
}

func (r *paymentAppRepo) GetPaymentAppCompanyUuid(uuid uint64) (*model.PaymentApp, error) {
	var paymentApp model.PaymentApp
	err := r.db.Model(&model.PaymentApp{}).Where("company_uuid = ?", uuid).First(&paymentApp).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &paymentApp, nil
}

// Create 创建支付配置
func (r *paymentAppRepo) Create(paymentApp *model.PaymentApp) error {
	return r.db.Create(paymentApp).Error
}

func (r *paymentAppRepo) UpdatePaymentApp(data map[string]any, companyUuid uint64) error {
	err := r.db.Model(&model.PaymentApp{}).Where("company_uuid = ?", companyUuid).Updates(data).Error
	if err != nil {
		return errors.WithMessage(err)
	}
	return nil
}
