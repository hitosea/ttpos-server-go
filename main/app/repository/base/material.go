package base

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMaterialRepo 原料仓库接口
type IMaterialRepo interface {
	GetMaterialList() ([]model.Material, error)
	UpdateMaterial(id uint, material model.Material) error
	CreateMaterial(material model.Material) (uint64, error)
	DeleteMaterial(id uint) error
	UpdateMaterials(materials []*model.Material) error
	GetMaterialByUuids(uuid []uint64) ([]*model.Material, error)
}

// NewMaterialRepo 创建新的原料仓库
func NewMaterialRepo(db *gorm.DB) IMaterialRepo {
	return NewMaterialRepoImpl(db)
}

// NewMaterialRepoImpl 创建新的原料仓库实现
func NewMaterialRepoImpl(db *gorm.DB) *MaterialRepoImpl {
	return &MaterialRepoImpl{db: db}
}

type MaterialRepoImpl struct {
	db *gorm.DB // 数据库连接
}

// GetMaterialList 获取原料列表
func (r *MaterialRepoImpl) GetMaterialList() ([]model.Material, error) {
	var materials []model.Material
	err := r.db.Model(&model.Material{}).Preload("MultiLanguageName").Where("delete_time = ?", 0).Find(&materials).Error
	return materials, errors.WithMessage(err)
}

// UpdateMaterial 更新原料
func (r *MaterialRepoImpl) UpdateMaterial(id uint, material model.Material) error {
	tx := r.db.Begin() // 开始事务
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 回滚事务
		}
	}()

	if err := tx.Model(&model.Material{}).Where("id = ?", id).Updates(material).Error; err != nil {
		tx.Rollback() // 更新失败，回滚事务
		return errors.WithMessage(err)
	}

	if err := tx.Model(&material.MultiLanguageName).Where("id = ?", material.MultiLanguageNameUuid).Updates(material.MultiLanguageName).Error; err != nil {
		tx.Rollback() // 更新多语言名称失败，回滚事务
		return errors.WithMessage(err)
	}

	return tx.Commit().Error // 提交事务
}

// CreateMaterial 创建原料
func (r *MaterialRepoImpl) CreateMaterial(material model.Material) (uint64, error) {
	err := r.db.Model(&model.Material{}).Create(&material).Error // 将多语言名称插入数据库
	return material.Uuid, errors.WithMessage(err)
}

// DeleteMaterial 删除原料
func (r *MaterialRepoImpl) DeleteMaterial(id uint) error {
	return r.db.Model(&model.Material{}).Where("id = ?", id).Update("delete_time", uint(time.Now().Unix())).Error
}

// UpdateMaterials 更新原料
func (r *MaterialRepoImpl) UpdateMaterials(materials []*model.Material) error {
	if len(materials) == 0 {
		return nil
	}
	list := make([]model.Material, 0)
	for _, material := range materials {
		material := *material
		material.SetNil()
		list = append(list, material)
	}
	if err := r.db.Model(&model.Material{}).Save(list).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// GetMaterialByUuid 根据uuid获取原料
func (r *MaterialRepoImpl) GetMaterialByUuids(uuids []uint64) ([]*model.Material, error) {
	var materials []*model.Material
	err := r.db.Model(&model.Material{}).Where("uuid in (?)", uuids).Find(&materials).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return materials, nil
}

// GetMaterialByUuidsWithUnit 根据uuids获取原料，并预加载单位信息
// .Preload("Unit").Preload("PurchaseUnit")
