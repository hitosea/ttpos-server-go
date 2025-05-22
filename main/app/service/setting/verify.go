package setting

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/context"
)

type VerifyAdvancedPasswordOption struct {
	IsAssistantCheckOrder bool // 是否是助手端下单校验高级密码
}

func WithIsAssistantCheckOrder() func(option *VerifyAdvancedPasswordOption) {
	return func(option *VerifyAdvancedPasswordOption) {
		option.IsAssistantCheckOrder = true
	}
}

// VerifyAdvancedPassword 验证高级密码
func (s *Srv) VerifyAdvancedPassword(ctx context.Context, password string, options ...func(option *VerifyAdvancedPasswordOption)) error {
	option := &VerifyAdvancedPasswordOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}

	if option.IsAssistantCheckOrder {
		// 获取点餐助手设置
		assistant, err := s.GetAssistantSetting(ctx, nil)
		if err != nil {
			return errors.WithMessage(err)
		}
		// 如果开启下单校验高级密码，则校验高级密码
		if assistant.IsCheckOrderPassword() {
			if password == "" {
				return errors.New("请输入确认密码")
			}
			// 如果需要校验高级密码，则校验高级密码
			if password != assistant.AdvancedPassword {
				return errors.New("确认密码错误")
			}
		}
		return nil
	}

	businessSetting, err := s.GetBusinessSetting(ctx)
	if err != nil {
		return errors.WithMessage(err)
	}
	if businessSetting.IsNeedPassword == "1" {
		if password == "" {
			return errors.New("请输入确认密码")
		}
		if ctx.GetSource() == constant.SourceCashier {
			// 获取收银机设置
			cashier, err := s.GetCashierSetting(ctx, nil)
			if err != nil {
				return errors.WithMessage(err)
			}
			if password != cashier.AdvancedPassword {
				return errors.New("确认密码错误")
			}
		}
		if ctx.GetSource() == constant.SourceAssistant {
			// 获取点餐助手设置
			assistant, err := s.GetAssistantSetting(ctx, nil)
			if err != nil {
				return errors.WithMessage(err)
			}
			if password != assistant.AdvancedPassword {
				return errors.New("确认密码错误")
			}
		}
		if ctx.GetSource() == constant.SourceKitchen {
			// 获取厨显设置
			kitchen, err := s.GetKitchenSetting(ctx, model.CompanySetting{}, nil)
			if err != nil {
				return errors.WithMessage(err)
			}
			if password != kitchen.AdvancedPassword {
				return errors.New("确认密码错误")
			}
		}
		if ctx.GetSource() == constant.SourceTablet {
			// 获取平板端设置
			tablet, err := s.GetTabletSetting(ctx, nil)
			if err != nil {
				return errors.WithMessage(err)
			}
			if password != tablet.AdvancedPassword {
				return errors.New("确认密码错误")
			}
		}
	}
	return nil
}
