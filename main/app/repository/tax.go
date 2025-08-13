package repository

import (
	"errors"

	apperrors "ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ITaxRepo 税种
type ITaxRepo interface {
	CreateTax(tax model.Tax) error
	GetTaxCategoryUuidByNameOptimized(name string) (uint64, error)
}

func NewTaxRepo(db *gorm.DB) ITaxRepo {
	return NewTaxRepoImpl(db)
}

// NewTaxRepoImpl 创建新的角色仓库实现
func NewTaxRepoImpl(db *gorm.DB) *TaxRepoImpl {
	return &TaxRepoImpl{db: db}
}

type TaxRepoImpl struct {
	db *gorm.DB
}

// CreateTax 创建税种
func (r *TaxRepoImpl) CreateTax(tax model.Tax) error {
	if err := r.db.Create(&tax).Error; err != nil {
		return apperrors.WithMessage(err, "创建税种失败")
	}
	return nil
}

// GetTaxCategoryUuidByNameOptimized 获取税种UUID（找不到时返回0，不报错）
func (r *TaxRepoImpl) GetTaxCategoryUuidByNameOptimized(name string) (uint64, error) {
	var taxCategories model.Tax
	err := r.db.Model(&model.Tax{}).Where("delete_time = ?", 0).Where("name = ?", name).First(&taxCategories).Error
	if err != nil {
		// 如果是记录不存在的错误，返回0而不是错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}
		// 其他错误仍然返回
		return 0, apperrors.WithMessage(err)
	}
	return taxCategories.Uuid, nil
}
