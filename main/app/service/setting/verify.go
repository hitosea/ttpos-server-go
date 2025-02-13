package setting

import (
	"errors"
)

// VerifyAdvancedPassword 验证高级密码
func (s *Srv) VerifyAdvancedPassword(companyUuid uint64, password string) error {
	businessSetting, err := s.GetBusinessSetting(companyUuid, "")
	if err != nil {
		return err
	}
	if businessSetting.IsNeedPassword == "1" {
		if password == "" {
			return errors.New("请输入确认密码")
		}
		cashier, err := s.GetCashierSetting(companyUuid, "", nil, nil)
		if err != nil {
			return err
		}
		if password != cashier.AdvancedPassword {
			return errors.New("确认密码错误")
		}
	}
	return nil
}
