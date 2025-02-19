package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPaymentMethodRepo 定义仓库接口
type IPaymentMethodRepo interface {
	WhereUuid(uuid uint64) DBOption // uuid条件
	WhereCashier() DBOption         // 在收银端显示
	WhereAssistant() DBOption       // 在助手端显示
	WhereStatus(status int) DBOption

	WithLogoFile() DBOption   // 关联logo文件
	WithQrcodeFile() DBOption // 关联二维码文件

	GetPaymentMethod(opts ...DBOption) model.PaymentMethod
	GetPaymentMethods(opts ...DBOption) []model.PaymentMethod
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

// GetPaymentMethod 获取支付方式
func (r *paymentMethodRepo) GetPaymentMethod(opts ...DBOption) model.PaymentMethod {
	var paymentMethod model.PaymentMethod
	db := r.db.Model(&model.PaymentMethod{})
	for _, opt := range opts {
		db = opt(db)
	}
	db.First(&paymentMethod)
	return paymentMethod
}

// GetPaymentMethods  获取支付方式
func (r *paymentMethodRepo) GetPaymentMethods(opts ...DBOption) []model.PaymentMethod {
	var paymentMethods []model.PaymentMethod
	db := r.db.Model(&model.PaymentMethod{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	db.Find(&paymentMethods)
	return paymentMethods
}

func (r *paymentMethodRepo) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

func (r *paymentMethodRepo) WhereCashier() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_show_cashier = 1")
	}
}

func (r *paymentMethodRepo) WhereAssistant() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("is_show_assistant = 1")
	}
}

func (r *paymentMethodRepo) WhereStatus(status int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

func (r *paymentMethodRepo) WithLogoFile() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("LogoFile")
	}
}

func (r *paymentMethodRepo) WithQrcodeFile() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("QrcodeFile")
	}
}
