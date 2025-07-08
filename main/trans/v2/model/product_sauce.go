package model

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	pkgUtils "ttpos-server-go/pkg/utils"
	v1 "ttpos-server-go/trans/v1"
)

// 全局小料Uuid缓存
var ProductSauceUuidMap map[int]uint64 // 旧表小料id -> 新表小料id

func init() {
	ProductSauceUuidMap = make(map[int]uint64)
}

func NewFeedUuid(FeedID int) (uint64, error) {
	// 生成新的雪花uuid
	uuid, err := pkgUtils.GetID()
	if err != nil {
		return 0, errors.WithMessage(err, "获取uuid失败")
	}
	if _, ok := ProductSauceUuidMap[FeedID]; !ok {
		ProductSauceUuidMap[FeedID] = uuid
	} else {
		uuid = ProductSauceUuidMap[FeedID]
	}
	return uuid, nil
}

func NewProductSauce(feed *v1.Feed) (*model.ProductSauce, error) {
	languageName, err := NewMultiLanguageName(feed.FeedName)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMultiLanguageName failed")
	}

	// 生成新的雪花uuid
	uuid, err := NewFeedUuid(feed.FeedID)
	if err != nil {
		return nil, errors.WithMessage(err, "获取uuid失败")
	}

	sauceMaterials := make([]*model.RelatedMaterial, 0)
	for _, material := range feed.ProductFeedMaterials {
		relatedMaterialUuid, err := pkgUtils.GetID()
		if err != nil {
			return nil, errors.WithMessage(err, "获取uuid失败")
		}
		sauceMaterials = append(sauceMaterials, &model.RelatedMaterial{
			BaseModel: model.BaseModel{
				// Uuid: uint64(material.MaterialID),
				Uuid:       relatedMaterialUuid,
				CreateTime: int64(material.CreateTime),
				UpdateTime: int64(material.CreateTime),
			},
			MaterialUuid: uint64(material.MaterialID),
			// RelatedUuid:  uint64(feed.FeedID), // 生成新的雪花uuid
			RelatedUuid: uuid,
			Num:         material.MaterialNum,
		})
	}
	productSauce := &model.ProductSauce{
		BaseModel: model.BaseModel{
			// Uuid: uint64(feed.FeedID),
			Uuid:       uuid,
			CreateTime: int64(feed.CreateTime),
			UpdateTime: int64(feed.UpdateTime),
		},
		Name:                  languageName.ToJson(),
		Price:                 feed.Price,
		MultiLanguageNameUuid: languageName.Uuid,
		MultiLanguageName:     *languageName,
		SauceMaterials:        sauceMaterials,
	}
	return productSauce, nil
}
