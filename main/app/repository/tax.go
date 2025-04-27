package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ITaxRepo 税种
type ITaxRepo interface {
	CreateTax(tax model.Tax) error
}

func NewTaxRepo(db *gorm.DB) ITaxRepo {
	return NewTaxRepoImpl(db)
}

// NewRoleRepoImpl 创建新的角色仓库实现
func NewTaxRepoImpl(db *gorm.DB) *TaxRepoImpl {
	return &TaxRepoImpl{db: db}
}

type TaxRepoImpl struct {
	db *gorm.DB
}

// CreateTax 创建税种
func (r *TaxRepoImpl) CreateTax(tax model.Tax) error {
	if err := r.db.Create(&tax).Error; err != nil {
		return errors.WithMessage(err, "创建税种失败")
	}
	return nil
}
