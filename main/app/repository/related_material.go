package repository

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IRelatedMaterialRepo interface {
	CreateRelatedMaterials(relatedMaterials []model.RelatedMaterial) error
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
