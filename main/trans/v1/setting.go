package v1

import (
	"fmt"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"

	"gorm.io/gorm"
)

type Setting struct {
	Key            string `gorm:"type:varchar(30);not null;comment:设置项标示"`
	Describe       string `gorm:"type:varchar(255);not null;default:'';comment:设置项描述"`
	Values         string `gorm:"type:mediumtext;not null;comment:设置内容（json格式）"`
	ShopSupplierID int    `gorm:"type:int(11);not null;default:0;comment:商户id"`
	AppID          uint   `gorm:"type:int(11) unsigned;not null;default:0;comment:小程序id"`
	UpdateTime     uint   `gorm:"type:int(11) unsigned;not null;default:0;comment:更新时间"`
}

type SettingRepository interface {
	GetSettingList() ([]*Setting, error)
	ConvertSetting() error
}

func NewSettingService(db *gorm.DB, targetDB *gorm.DB) SettingRepository {
	return &SettingService{db: db, targetDB: targetDB}
}

type SettingService struct {
	db       *gorm.DB
	targetDB *gorm.DB
}

func (s *SettingService) GetSettingList() ([]*Setting, error) {
	var settings []*Setting
	err := s.db.Find(&settings).Error
	return settings, err
}

func (s *SettingService) ConvertSetting() error {
	settings, err := s.GetSettingList()
	if err != nil {
		return err
	}

	for _, setting := range settings {
		fmt.Println(fmt.Sprintf("setting: %+v", setting))
		model := model.Setting{
			Key:        setting.Key,
			Describe:   setting.Describe,
			Values:     setting.Values,
			UpdateTime: int64(setting.UpdateTime),
		}
		_, err = repository.NewSettingRepo(s.targetDB).Create(model)
		if err != nil {
			return err
		}
	}
	return nil
}
