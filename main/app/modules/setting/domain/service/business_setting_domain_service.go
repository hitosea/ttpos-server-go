package service

import (
	"errors"
	"ttpos-server-go/app/modules/setting/domain/entity"
)

// BusinessSettingDomainServiceImpl 业务设置域服务实现
type BusinessSettingDomainServiceImpl struct{}

// NewBusinessSettingDomainService 创建业务设置域服务
func NewBusinessSettingDomainService() IBusinessSettingDomainService {
	return &BusinessSettingDomainServiceImpl{}
}

// ValidateBatchSettings 验证分批设置
func (b *BusinessSettingDomainServiceImpl) ValidateBatchSettings(isBatch string, batchTagNum int) error {
	if isBatch == "1" && batchTagNum == 0 {
		return errors.New("开启分批送厨时必须至少有一个分批类型")
	}
	return nil
}

// UpdateBatchCookingMode 更新分批烹饪模式
func (b *BusinessSettingDomainServiceImpl) UpdateBatchCookingMode(business *entity.BusinessSetting, cookingMode string) {
	business.UpdateBatchCookingMode(cookingMode)
}
