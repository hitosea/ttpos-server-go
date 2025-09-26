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
	UpdataTax(data map[string]any, opts ...DBOption) error
	GetTaxCategoryList(opts ...DBOption) ([]model.Tax, error)
	GetTaxCategoryUuidByNameOptimized(name string) (uint64, error)
	GetTaxCategory(opts ...DBOption) (model.Tax, error)
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

// UpdataTax 更新税种
func (r *TaxRepoImpl) UpdataTax(data map[string]any, opts ...DBOption) error {
	db := r.db.Model(&model.Tax{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(data).Error
}

// GetTaxCategoryList 获取税种列表
func (r *TaxRepoImpl) GetTaxCategoryList(opts ...DBOption) ([]model.Tax, error) {
	var taxCategories []model.Tax
	db := r.db.Model(&model.Tax{}).Where("delete_time = ?", 0)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&taxCategories).Error
	return taxCategories, apperrors.WithMessage(err)
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

// GetTaxCategory 获取税种
func (r *TaxRepoImpl) GetTaxCategory(opts ...DBOption) (model.Tax, error) {
	var tax model.Tax
	db := r.db.Model(&model.Tax{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.First(&tax).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Tax{}, nil
		}
	}
	return tax, apperrors.WithMessage(err)
}
