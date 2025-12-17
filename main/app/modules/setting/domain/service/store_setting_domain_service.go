package service

import (
	"errors"
	"ttpos-server-go/app/modules/setting/domain/entity"
	"ttpos-server-go/app/modules/setting/domain/valueobject"
)

// StoreSettingDomainServiceImpl 门店设置域服务实现
type StoreSettingDomainServiceImpl struct{}

// NewStoreSettingDomainService 创建门店设置域服务
func NewStoreSettingDomainService() IStoreSettingDomainService {
	return &StoreSettingDomainServiceImpl{}
}

// ValidateStoreLanguages 验证门店语言设置
func (s *StoreSettingDomainServiceImpl) ValidateStoreLanguages(languages []valueobject.LanguageItem) error {
	if len(languages) == 0 {
		return errors.New("至少需要设置一种语言")
	}

	// 验证语言格式
	for _, lang := range languages {
		if !lang.IsValid() {
			return errors.New("语言设置格式不正确")
		}
	}

	return nil
}

// UpdateStoreBasicInfo 更新门店基本信息
func (s *StoreSettingDomainServiceImpl) UpdateStoreBasicInfo(store *entity.StoreSetting, name, company, address, phone, storeCode string) {
	store.UpdateBasicInfo(name, company, address, phone, storeCode)
}

// UpdateStoreCoordinates 更新门店坐标
func (s *StoreSettingDomainServiceImpl) UpdateStoreCoordinates(store *entity.StoreSetting, coordinates string) {
	store.UpdateCoordinates(coordinates)
}
