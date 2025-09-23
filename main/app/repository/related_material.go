package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IRelatedMaterialRepo interface {
	CreateRelatedMaterials(relatedMaterials []model.RelatedMaterial) error
	GetProductBomCardUuidsByMaterialUuid(materialUuid uint64) ([]uint64, error) // 通过物品uuid获取该材料相关的成本卡uuid列表
}

type relatedMaterialRepoImpl struct {
	db *gorm.DB
}

func NewRelatedMaterialRepo(db *gorm.DB) IRelatedMaterialRepo {
	return &relatedMaterialRepoImpl{db: db}
}

func (r *relatedMaterialRepoImpl) CreateRelatedMaterials(relatedMaterials []model.RelatedMaterial) error {
	// 如果relatedMaterials为空，则不创建
	if len(relatedMaterials) == 0 {
		return nil
	}
	// 清空关联对象
	list := make([]model.RelatedMaterial, 0)
	for _, relatedMaterial := range relatedMaterials {
		relatedMaterial.SetNil()
		list = append(list, relatedMaterial)
	}

	// 创建related_material表数据
	if err := r.db.Model(&model.RelatedMaterial{}).Create(list).Error; err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// 获取成本卡uuid列表
func (r *relatedMaterialRepoImpl) GetProductBomCardUuidsByMaterialUuid(materialUuid uint64) ([]uint64, error) {
	var productBomCardUuids []uint64
	if err := r.db.Model(&model.RelatedMaterial{}).Where("material_uuid = ?", materialUuid).Where("delete_time = 0").Pluck("related_uuid", &productBomCardUuids).Error; err != nil {
		return nil, errors.WithMessage(err)
	}
	return productBomCardUuids, nil
}
