package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// ICompanyAreaRepo 门店区域仓库接口
type ICompanyAreaRepo interface {
	GetList(opts ...DBOption) ([]model.CompanyArea, error)
	GetByUuid(uuid uint64) (*model.CompanyArea, error)
	Create(area *model.CompanyArea) error
	Update(uuid uint64, data map[string]any) error
	Delete(uuid uint64) error
}

func NewCompanyAreaRepo(db *gorm.DB) ICompanyAreaRepo {
	return &companyAreaRepo{db: db}
}

type companyAreaRepo struct {
	db *gorm.DB
}

// GetList 获取区域列表
func (r *companyAreaRepo) GetList(opts ...DBOption) ([]model.CompanyArea, error) {
	var areas []model.CompanyArea
	db := r.db.Model(&model.CompanyArea{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Order("sort ASC, id ASC").Find(&areas).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return areas, nil
}

// GetByUuid 根据UUID获取区域
func (r *companyAreaRepo) GetByUuid(uuid uint64) (*model.CompanyArea, error) {
	var area model.CompanyArea
	err := r.db.Model(&model.CompanyArea{}).
		Scopes(NotDeleted).
		Preload("MultiLanguageName").
		Where("uuid = ?", uuid).
		First(&area).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return &area, nil
}

// Create 创建区域
func (r *companyAreaRepo) Create(area *model.CompanyArea) error {
	return r.db.Create(area).Error
}

// Update 更新区域
func (r *companyAreaRepo) Update(uuid uint64, data map[string]any) error {
	return r.db.Model(&model.CompanyArea{}).Where("uuid = ?", uuid).Updates(data).Error
}

// Delete 软删除区域
func (r *companyAreaRepo) Delete(uuid uint64) error {
	return r.db.Model(&model.CompanyArea{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error
}
