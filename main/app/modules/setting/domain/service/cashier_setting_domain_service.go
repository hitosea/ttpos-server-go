package service

import (
	"errors"
	"ttpos-server-go/app/modules/setting/domain/entity"
)

// CashierSettingDomainServiceImpl 收银机设置域服务实现
type CashierSettingDomainServiceImpl struct{}

// NewCashierSettingDomainService 创建收银机设置域服务
func NewCashierSettingDomainService() ICashierSettingDomainService {
	return &CashierSettingDomainServiceImpl{}
}

// ValidateAutoOrderSettings 验证自动接单设置
func (c *CashierSettingDomainServiceImpl) ValidateAutoOrderSettings(isAutoOrder, autoOrderLimit string) error {
	if isAutoOrder == "1" && autoOrderLimit == "" {
		return errors.New("开启自动接单时必须设置金额上限")
	}
	return nil
}

// UpdateAutoOrderSettings 更新自动接单设置
func (c *CashierSettingDomainServiceImpl) UpdateAutoOrderSettings(cashier *entity.CashierSetting, isAutoOrder, autoOrderLimit, isAutoVoice string) {
	cashier.UpdateAutoOrderSettings(isAutoOrder, autoOrderLimit, isAutoVoice)
}

// UpdateAutoMemberOrderSettings 更新自动接单会员订单设置
func (c *CashierSettingDomainServiceImpl) UpdateAutoMemberOrderSettings(cashier *entity.CashierSetting, isAutoMemberOrder, autoMemberOrderLimit, isAutoVoice string) {
	cashier.UpdateAutoMemberOrderSettings(isAutoMemberOrder, autoMemberOrderLimit, isAutoVoice)
}
