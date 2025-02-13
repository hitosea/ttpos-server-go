package old_model

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type ReturnReason struct {
	Id             uint   `gorm:"primaryKey;autoIncrement;comment:'退菜原因唯一标识符'"`
	Reason         string `gorm:"default:'';comment:'原因'"`
	ShopSupplierId int    `gorm:"default:0;comment:'门店id'"`
	AppId          int    `gorm:"default:0;comment:'应用id'"`
	CreateTime     int64  `gorm:"autoCreateTime;comment:'创建时间'"`
	UpdateTime     int64  `gorm:"autoUpdateTime;comment:'更新时间'"`
}

type ReturnReasonRepository interface {
	GetReturnReasonList() ([]*ReturnReason, error)
	ConvertReturnReason() error
}

type ReturnReasonService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *ReturnReasonService) GetReturnReasonList() ([]*ReturnReason, error) {
	var returnReasons []*ReturnReason
	if err := s.db.Find(&returnReasons).Error; err != nil {
		return nil, err
	}
	return returnReasons, nil
}

func (s *ReturnReasonService) ConvertReturnReason() error {
	returnReasons, err := s.GetReturnReasonList()
	if err != nil {
		return err
	}
	for _, returnReason := range returnReasons {
		fmt.Println(fmt.Sprintf("-------迁移return_reason: %+v", returnReason))
		names := Names{}
		err := names.GetNames(returnReason.Reason)
		if err != nil {
			return err
		}
		fmt.Println(fmt.Sprintf("return_reason_id: %d, return_reason_name: %+v", returnReason.Id, names))

		id, err := utils.GetID()
		if err != nil {
			return err
		}
		fmt.Println(fmt.Sprintf("id: %d", id))

		languageName := names.GenMultiLanguageName(id)

		reason := model.ReturnFoodReason{
			BaseModel: model.BaseModel{
				Uuid:       uint64(returnReason.Id),
				CreateTime: returnReason.CreateTime,
				UpdateTime: returnReason.UpdateTime,
			},
			Name:                  names.Zh,
			MultiLanguageNameUuid: uint(id),
			MultiLanguageName:     languageName,
		}
		_, err = base.NewReturnFoodReasonRepo(s.targetDB).CreateReturnFoodReason(reason)
		if err != nil {
			return err
		}
	}
	return nil
}
