package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IProductUnitRepo 商品单位仓库接口
type IProductUnitRepo interface {
	// 基础CRUD操作
	GetProductUnit(opts ...DBOption) (*model.ProductUnit, error)
	GetProductUnitByErpnextUom(erpnextUom string) (*model.ProductUnit, error)
	GetProductUnitList(opts ...DBOption) ([]*model.ProductUnit, error)
	WhereByUuids(uuids []uint64) DBOption
	WhereByErpnextUom(erpnextUom string) DBOption
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
