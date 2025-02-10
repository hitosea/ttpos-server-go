package old_model

import (
	"fmt"
	"gorm.io/gorm"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/database"
)

type FreeTag struct {
	Id             uint   `gorm:"primaryKey;autoIncrement;comment:'标签唯一标识符'"`
	FreeTag        string `gorm:"type:varchar(2000);not null;default:'';comment:'标签名'"`
	ShopSupplierId int    `gorm:"default:0;comment:'门店id'"`
	AppId          int    `gorm:"default:0;comment:'应用id'"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:'创建时间'"`
	UpdateTime     int64  `gorm:"autoUpdateTime;comment:'更新时间'"`
}

type FreeTagRepository interface {
	GetFreeTagList() ([]*FreeTag, error)
	ConvertFreeTag() error
}

type FreeTagService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *FreeTagService) GetFreeTagList() ([]*FreeTag, error) {
	var freeTags []*FreeTag
	if err := s.db.Find(&freeTags).Error; err != nil {
		return nil, err
	}
	return freeTags, nil
}

func (s *FreeTagService) ConvertFreeTag() error {
	freeTags, err := s.GetFreeTagList()
	if err != nil {
		return err
	}
	for _, freeTag := range freeTags {
		names := Names{}
		err := names.GetNames(freeTag.FreeTag)
		if err != nil {
			return err
		}
		fmt.Println(fmt.Sprintf("free_tag_id: %d, free_tag_name: %+v", freeTag.Id, names))

		id, err := database.GetID()
		fmt.Println(fmt.Sprintf("id: %d", id))

		languageName := names.GenMultiLanguageName(id)

		reason := model.GiftOrFreeOrderReason{
			Uuid:                  freeTag.Id,
			Name:                  names.Zh,
			MultiLanguageNameUuid: uint(id),
			MultiLanguageName:     languageName,
		}
		_, err = base.NewGiftOrFreeOrderReasonRepo(s.targetDB).CreateGiftOrFreeOrderReason(reason)
		if err != nil {
			return err
		}
	}
	return nil
}
