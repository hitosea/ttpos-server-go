package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMaterialRepo 物品仓库接口
type IMaterialRepo interface {
	GetMaterialListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.Material, int64, error)
	GetMaterialByUuid(uuid uint64, opts ...DBOption) (model.Material, error)
	CreateMaterial(material model.Material) (uint64, error)
	UpdateMaterial(material model.Material) error
	DeleteMaterial(uuid uint64) error
	UpdateMaterialStatus(uuid uint64, status bool) error
	SearchMaterials(keyword string, opts ...DBOption) ([]model.Material, error)
	GetMaterialByBarcode(barcode string, opts ...DBOption) (model.Material, error)
	GetMaterialByName(name string, opts ...DBOption) (model.Material, error)
	GetMaterialCategoryByName(name string) (*model.MaterialCategory, error)
	CreateMaterialCategory(materialCategory model.MaterialCategory) (uint64, error)
	GetMaterialCategoryList() ([]model.MaterialCategory, error)
}

// NewMaterialRepo 创建新的物品仓库
func NewMaterialRepo(db *gorm.DB) IMaterialRepo {
	return NewMaterialRepoImpl(db)
}

// NewMaterialRepoImpl 创建新的物品仓库实现
func NewMaterialRepoImpl(db *gorm.DB) *MaterialRepoImpl {
	return &MaterialRepoImpl{db: db}
}

type MaterialRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// GetMaterialListWithPagination 获取物品列表（分页）
func (r *MaterialRepoImpl) GetMaterialListWithPagination(pageNo, pageSize int, opts ...DBOption) ([]model.Material, int64, error) {
	var materials []model.Material
	var total int64

	// 构建查询
	query := r.db.Model(&model.Material{}).Where("delete_time = ?", 0)

	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.WithMessage(err, "计算物品总数失败")
	}

	// 分页查询
	offset := (pageNo - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("create_time DESC").Find(&materials).Error; err != nil {
		return nil, 0, errors.WithMessage(err, "查询物品列表失败")
	}

	return materials, total, nil
}

// GetMaterialByUuid 根据UUID获取物品详情
func (r *MaterialRepoImpl) GetMaterialByUuid(uuid uint64, opts ...DBOption) (model.Material, error) {
	var material model.Material

	query := r.db.Model(&model.Material{}).Where("uuid = ? AND delete_time = ?", uuid, 0)

	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}

	if err := query.First(&material).Error; err != nil {
		return model.Material{}, errors.WithMessage(err, "查询物品详情失败")
	}

	return material, nil
}

// CreateMaterial 创建物品
func (r *MaterialRepoImpl) CreateMaterial(material model.Material) (uint64, error) {
	if err := r.db.Create(&material).Error; err != nil {
		return 0, errors.WithMessage(err, "创建物品失败")
	}
	return material.Uuid, nil
}

// UpdateMaterial 更新物品
func (r *MaterialRepoImpl) UpdateMaterial(material model.Material) error {
	if err := r.db.Model(&model.Material{}).Where("uuid = ?", material.Uuid).Updates(material).Error; err != nil {
		return errors.WithMessage(err, "更新物品失败")
	}
	return nil
}

// DeleteMaterial 删除物品（软删除）
func (r *MaterialRepoImpl) DeleteMaterial(uuid uint64) error {
	if err := r.db.Model(&model.Material{}).Where("uuid = ?", uuid).Update("delete_time", time.Now().Unix()).Error; err != nil {
		return errors.WithMessage(err, "删除物品失败")
	}
	return nil
}

// UpdateMaterialStatus 更新物品状态
func (r *MaterialRepoImpl) UpdateMaterialStatus(uuid uint64, status bool) error {
	if err := r.db.Model(&model.Material{}).Where("uuid = ?", uuid).Update("status", status).Error; err != nil {
		return errors.WithMessage(err, "更新物品状态失败")
	}
	return nil
}

// SearchMaterials 搜索物品
func (r *MaterialRepoImpl) SearchMaterials(keyword string, opts ...DBOption) ([]model.Material, error) {
	var materials []model.Material

	query := r.db.Model(&model.Material{}).Where("delete_time = ? AND (name LIKE ? OR barcode_value LIKE ?)", 0, "%"+keyword+"%", "%"+keyword+"%")

	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}

	if err := query.Order("create_time DESC").Find(&materials).Error; err != nil {
		return nil, errors.WithMessage(err, "搜索物品失败")
	}

	return materials, nil
}

// GetMaterialByBarcode 根据条形码获取物品
func (r *MaterialRepoImpl) GetMaterialByBarcode(barcode string, opts ...DBOption) (model.Material, error) {
	var material model.Material

	query := r.db.Model(&model.Material{}).Where("barcode_value = ? AND delete_time = ?", barcode, 0)

	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}

	if err := query.First(&material).Error; err != nil {
		return model.Material{}, errors.WithMessage(err, "根据条形码查询物品失败")
	}

	return material, nil
}

// GetMaterialByName 根据名称获取物品
func (r *MaterialRepoImpl) GetMaterialByName(name string, opts ...DBOption) (model.Material, error) {
	var material model.Material

	query := r.db.Model(&model.Material{}).Where("name = ? AND delete_time = ?", name, 0)

	// 应用查询选项
	for _, opt := range opts {
		query = opt(query)
	}

	if err := query.First(&material).Error; err != nil {
		return model.Material{}, errors.WithMessage(err, "根据名称查询物品失败")
	}

	return material, nil
}

func (r *MaterialRepoImpl) GetMaterialCategoryByName(name string) (*model.MaterialCategory, error) {
	var materialCategory model.MaterialCategory

	query := r.db.Model(&model.MaterialCategory{}).Where("name = ? AND delete_time = ?", name, 0)

	if err := query.First(&materialCategory).Error; err != nil {
		return nil, errors.WithMessage(err, "根据名称查询物品类别失败")
	}

	return &materialCategory, nil
}

func (r *MaterialRepoImpl) CreateMaterialCategory(materialCategory model.MaterialCategory) (uint64, error) {
	if err := r.db.Create(&materialCategory).Error; err != nil {
		return 0, errors.WithMessage(err, "创建物品类别失败")
	}
	return materialCategory.Uuid, nil
}

func (r *MaterialRepoImpl) GetMaterialCategoryList() ([]model.MaterialCategory, error) {
	var materialCategories []model.MaterialCategory

	if err := r.db.Model(&model.MaterialCategory{}).Where("delete_time = ?", 0).Order("create_time DESC").Find(&materialCategories).Error; err != nil {
		return nil, errors.WithMessage(err, "获取物品类别列表失败")
	}

	return materialCategories, nil
}
