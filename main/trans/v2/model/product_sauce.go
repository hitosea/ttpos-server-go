package model

import (
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	v1 "ttpos-server-go/trans/v1"
)

func NewProductSauce(feed *v1.Feed) (*model.ProductSauce, error) {
	languageName, err := NewMultiLanguageName(feed.FeedName)
	if err != nil {
		return nil, errors.WithMessage(err, "NewMultiLanguageName failed")
	}

	sauceMaterials := make([]*model.RelatedMaterial, 0)
	for _, material := range feed.ProductFeedMaterials {
		sauceMaterials = append(sauceMaterials, &model.RelatedMaterial{
			BaseModel: model.BaseModel{
				Uuid: uint64(material.MaterialID),
			},
			MaterialUuid: uint64(material.MaterialID),
			RelatedUuid:  uint64(feed.FeedID),
			Num:          material.MaterialNum,
		})
	}
	productSauce := &model.ProductSauce{
		BaseModel: model.BaseModel{
			Uuid: uint64(feed.FeedID),
		},
		Name:                  languageName.ToJson(),
		Price:                 feed.Price,
		MultiLanguageNameUuid: uint(languageName.Uuid),
		MultiLanguageName:     *languageName,
		SauceMaterials:        sauceMaterials,
	}
	return productSauce, nil
}
