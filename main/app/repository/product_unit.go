package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductUnitRepo 商品单位仓库接口
type IProductUnitRepo interface {
	// 基础CRUD操作
	GetProductUnit(opts ...DBOption) (*model.ProductUnit, error)
	GetProductUnitByErpnextUom(erpnextUom string) (*model.ProductUnit, error)
	GetProductUnitByErpnextUomWithMultiLanguageName(erpnextUom string) (*model.ProductUnit, error)
	GetProductUnitList(opts ...DBOption) ([]*model.ProductUnit, error)
	CreateRecord(unit *model.ProductUnit) error
	CreateBatch(units []model.ProductUnit) error
	UpdateProductUnitData(data map[string]any, opts ...DBOption) error
	SoftDeleteByUuids(uuids []uint64) error
	RecoverByUuids(uuids []uint64) error
	HardDeleteByHeadquarter() error
	GetUnitListWithScopes(scopes ...func(*gorm.DB) *gorm.DB) ([]model.ProductUnit, error)
	GetMaxSort(scopes ...func(*gorm.DB) *gorm.DB) (int, error)
	WhereByUuids(uuids []uint64) DBOption
	WhereByErpnextUom(erpnextUom string) DBOption
	UpdateNameByMultiLanguageNameUuid(multiLanguageNameUuid uint64, name string) error
}

// NewMaterialUnitRepo 创建新的原料单位仓库
func NewProductUnitRepo(db *gorm.DB) IProductUnitRepo {
	return NewProductUnitRepoImpl(db)
}

// NewMaterialUnitRepoImpl 创建新的原料单位仓库实现
func NewProductUnitRepoImpl(db *gorm.DB) *ProductUnitRepoImpl {
	return &ProductUnitRepoImpl{db: db}
}

type ProductUnitRepoImpl struct {
	db *gorm.DB // 数据库连接
}

func (r *ProductUnitRepoImpl) GetProductUnit(opts ...DBOption) (*model.ProductUnit, error) {
	var productUnit model.ProductUnit

	query := r.db.Model(&model.ProductUnit{}).Where("delete_time = ?", 0)

	for _, opt := range opts {
		query = opt(query)
	}

	if err := query.First(&productUnit).Error; err != nil {
		return nil, errors.WithMessage(err)
	}

	return &productUnit, nil
}

func (r *ProductUnitRepoImpl) GetProductUnitByErpnextUom(erpnextUom string) (*model.ProductUnit, error) {
	var productUnit model.ProductUnit

	query := r.db.Model(&model.ProductUnit{}).Where("delete_time = ?", 0).Where("erpnext_uom = ?", erpnextUom)
	if err := query.First(&productUnit).Error; err != nil {
		return nil, errors.WithMessage(err)
	}

	return &productUnit, nil
}

func (r *ProductUnitRepoImpl) GetProductUnitByErpnextUomWithMultiLanguageName(erpnextUom string) (*model.ProductUnit, error) {
	var productUnit model.ProductUnit

	query := r.db.Model(&model.ProductUnit{}).Preload("MultiLanguageName").Where("delete_time = ?", 0).Where("erpnext_uom = ?", erpnextUom)
	if err := query.First(&productUnit).Error; err != nil {
		return nil, errors.WithMessage(err)
	}

	return &productUnit, nil
}

func (r *ProductUnitRepoImpl) GetProductUnitList(opts ...DBOption) ([]*model.ProductUnit, error) {
	var productUnits []*model.ProductUnit

	query := r.db.Model(&model.ProductUnit{}).Where("delete_time = ?", 0)

	for _, opt := range opts {
		query = opt(query)
	}

	if err := query.Find(&productUnits).Error; err != nil {
		return nil, errors.WithMessage(err)
	}

	return productUnits, nil
}

func (r *ProductUnitRepoImpl) WhereByUuids(uuids []uint64) DBOption {
	return func(query *gorm.DB) *gorm.DB {
		return query.Where("uuid IN (?)", uuids)
	}
}

func (r *ProductUnitRepoImpl) WhereByErpnextUom(erpnextUom string) DBOption {
	return func(query *gorm.DB) *gorm.DB {
		return query.Where("erpnext_uom = ?", erpnextUom)
	}
}

// UpdateNameByMultiLanguageNameUuid 根据多语言名称UUID更新name字段
func (r *ProductUnitRepoImpl) UpdateNameByMultiLanguageNameUuid(multiLanguageNameUuid uint64, name string) error {
	return r.db.Model(&model.ProductUnit{}).Where("multi_language_name_uuid = ?", multiLanguageNameUuid).Update("name", name).Error
}

func (r *ProductUnitRepoImpl) CreateRecord(unit *model.ProductUnit) error {
	return r.db.Create(unit).Error
}

func (r *ProductUnitRepoImpl) CreateBatch(units []model.ProductUnit) error {
	if len(units) == 0 {
		return nil
	}
	return r.db.Create(&units).Error
}

func (r *ProductUnitRepoImpl) UpdateProductUnitData(data map[string]any, opts ...DBOption) error {
	db := r.db.Model(&model.ProductUnit{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(data).Error
}

func (r *ProductUnitRepoImpl) SoftDeleteByUuids(uuids []uint64) error {
	if len(uuids) == 0 {
		return nil
	}
	return r.db.Model(&model.ProductUnit{}).Where("uuid IN (?)", uuids).Update("delete_time", time.Now().Unix()).Error
}

func (r *ProductUnitRepoImpl) RecoverByUuids(uuids []uint64) error {
	if len(uuids) == 0 {
		return nil
	}
	return r.db.Model(&model.ProductUnit{}).Where("uuid IN (?)", uuids).Update("delete_time", 0).Error
}

func (r *ProductUnitRepoImpl) HardDeleteByHeadquarter() error {
	return r.db.Where("headquarter_uuid > 0").Delete(&model.ProductUnit{}).Error
}

func (r *ProductUnitRepoImpl) GetUnitListWithScopes(scopes ...func(*gorm.DB) *gorm.DB) ([]model.ProductUnit, error) {
	var units []model.ProductUnit
	err := r.db.Model(&model.ProductUnit{}).Scopes(scopes...).Find(&units).Error
	return units, err
}

func (r *ProductUnitRepoImpl) GetMaxSort(scopes ...func(*gorm.DB) *gorm.DB) (int, error) {
	var maxSort int
	err := r.db.Model(&model.ProductUnit{}).Scopes(scopes...).Select("ifnull(max(sort), 0)").Scan(&maxSort).Error
	return maxSort, err
}
