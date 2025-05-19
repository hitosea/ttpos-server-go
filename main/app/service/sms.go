package service

import (
	"fmt"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/lock"

	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/sms"
)

// ISmsSrv 短信服务接口
type ISmsSrv interface {
	// SendMemberConsumptionSMS 发送会员消费短信
	SendMemberConsumptionSMS(ctx context.Context, phone string, params *sms.MemberConsumptionRequest) error
	// SendMemberRechargeSMS 发送会员充值短信
	SendMemberRechargeSMS(ctx context.Context, phone string, params *sms.MemberRechargeRequest) error
	// SendMemberRechargeRefundSMS 发送会员充值退款短信
	SendMemberRechargeRefundSMS(ctx context.Context, phone string, params *sms.MemberRechargeRefundRequest) error
	// SendMemberOrderRefundSMS 发送会员用餐订单退款短信
	SendMemberOrderRefundSMS(ctx context.Context, phone string, params *sms.MemberOrderRefundRequest) error
}

// smsSrv 短信服务实现
type smsSrv struct {
	bus    *event.SystemEventBus
	dbm    *database.DBManager // 数据库管理器
	client sms.SMSClient
}

// NewSMSSrv 创建短信服务实例
func NewSMSSrv(dbm *database.DBManager) ISmsSrv {
	return NewSMSSrvImpl(dbm)
}

// NewSMSSrvImpl 创建短信服务实例（实现）
func NewSMSSrvImpl(dbm *database.DBManager) ISmsSrv {
	return &smsSrv{
		bus:    event.NewSystemBus(),
		dbm:    dbm,
		client: sms.GetSMSClient(),
	}
}

// 格式化手机号
// 10位数字的为泰国，前缀为+66，11位数字的为中国，前缀为+86
// 如果泰国手机号以0开头，则去掉0
func (s *smsSrv) formatPhone(phone string) (string, error) {
	// 如果手机号已经有前缀，则去掉+66或+86
	if strings.HasPrefix(phone, constant.ThailandPrefix) {
		phone = strings.TrimPrefix(phone, constant.ThailandPrefix)
	}
	if strings.HasPrefix(phone, constant.ChinaPrefix) {
		phone = strings.TrimPrefix(phone, constant.ChinaPrefix)
	}

	if len(phone) == 10 {
		// 如果手机号以0开头，则去掉0
		if phone[0] == '0' {
			phone = phone[1:]
		}
		return constant.ThailandPrefix + phone, nil
	}
	if len(phone) == 11 {
		return constant.ChinaPrefix + phone, nil
	}
	return "", fmt.Errorf("invalid phone number")
}

// 选择语言。如果语言为中文，则返回zh，否则返回en
func (s *smsSrv) selectLanguage(defaultLanguage string) string {
	if defaultLanguage == "zh" {
		return "zh"
	}
	return "en"
}

func (s *smsSrv) checkQuotaAndFormatPhone(ctx context.Context, phone string) (string, string, string, error) {

	// 检查短信额度
	setting := ctx.GetCompanySetting()
	if !setting.SmsEnabled() {
		err := fmt.Errorf("SMS service is not enabled, EnableSms: %d, SmsQuota: %d", setting.EnableSms, setting.SmsQuota)
		return "", "", "", errors.WithMessage(err, "没有开启短信或没有额度")
	}

	// 格式化手机号
	formattedPhone, err := s.formatPhone(phone)
	if err != nil {
		err := fmt.Errorf("invalid phone number: %s, err:%v", phone, err)
		return "", "", "", errors.WithMessage(err, "手机号格式错误")
	}

	// 选择语言
	defaultLanguage := setting.GetDefaultLanguage()
	language := s.selectLanguage(defaultLanguage)

	return formattedPhone, language, ctx.GetCompany().Name, nil
}

// SendMemberConsumptionSMS 发送会员消费短信
func (s *smsSrv) SendMemberConsumptionSMS(ctx context.Context, phone string, params *sms.MemberConsumptionRequest) error {
	company := ctx.GetCompany()
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(company.Uuid)
		defer lock.NewSystemLock().UnlockUuid(company.Uuid)
		ctx.AddLock()
	}

	formattedPhone, language, companyName, err := s.checkQuotaAndFormatPhone(ctx, phone)
	if err != nil {
		return err
	}

	// 获取公司名称
	if params.Company == "" {
		params.Company = companyName
	}

	// 如果增加的积分和会员支付的金额都为0，则不发送短信
	if params.IncreasePoints == 0 && params.MemberPay == 0 {
		return errors.WithMessage(errors.New("增加的积分和会员支付的金额都为0，不发送短信"))
	}

	// 发送短信
	resp, err := s.client.SendMemberConsumptionSMS(formattedPhone, language, params)
	if err != nil {
		err := fmt.Errorf("failed to send SMS: %v", err)
		return errors.WithMessage(err, "发送短信失败")
	}

	// 如果发送成功，扣减额度
	if resp.Code == sms.ResponseCodeSuccess {
		if err := repository.NewCompanySettingRepo(ctx.GetDB()).UpdateSmsQuota(company.Uuid, 1); err != nil {
			err := fmt.Errorf("failed to update SMS quota: %v", err)
			return errors.WithMessage(err, "扣减短信额度失败")
		}
	} else {
		err := fmt.Errorf("failed to send SMS code: %v, msg: %v", resp.Code, resp.Msg)
		return errors.WithMessage(err, "发送短信失败")
	}

	return nil
}

// SendMemberRechargeSMS 发送会员充值短信
func (s *smsSrv) SendMemberRechargeSMS(ctx context.Context, phone string, params *sms.MemberRechargeRequest) error {
	company := ctx.GetCompany()
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(company.Uuid)
		defer lock.NewSystemLock().UnlockUuid(company.Uuid)
		ctx.AddLock()
	}

	formattedPhone, language, companyName, err := s.checkQuotaAndFormatPhone(ctx, phone)
	if err != nil {
		return err
	}

	// 获取公司名称
	if params.Company == "" {
		params.Company = companyName
	}

	// 发送短信
	resp, err := s.client.SendMemberRechargeSMS(formattedPhone, language, params)
	if err != nil {
		err := fmt.Errorf("failed to send SMS: %v", err)
		return errors.WithMessage(err, "发送短信失败")
	}

	// 如果发送成功，扣减额度
	if resp.Code == sms.ResponseCodeSuccess {
		if err := repository.NewCompanySettingRepo(ctx.GetDB()).UpdateSmsQuota(company.Uuid, 1); err != nil {
			err := fmt.Errorf("failed to update SMS quota: %v", err)
			return errors.WithMessage(err, "扣减短信额度失败")
		}
	} else {
		err := fmt.Errorf("failed to send SMS code: %v, msg: %v", resp.Code, resp.Msg)
		return errors.WithMessage(err, "发送短信失败")
	}

	return nil
}

// SendMemberRechargeRefundSMS 发送会员充值退款短信
func (s *smsSrv) SendMemberRechargeRefundSMS(ctx context.Context, phone string, params *sms.MemberRechargeRefundRequest) error {
	company := ctx.GetCompany()
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(company.Uuid)
		defer lock.NewSystemLock().UnlockUuid(company.Uuid)
		ctx.AddLock()
	}

	formattedPhone, language, companyName, err := s.checkQuotaAndFormatPhone(ctx, phone)
	if err != nil {
		return err
	}

	// 获取公司名称
	if params.Company == "" {
		params.Company = companyName
	}

	// 发送短信
	resp, err := s.client.SendMemberRechargeRefundSMS(formattedPhone, language, params)
	if err != nil {
		err := fmt.Errorf("failed to send SMS: %v", err)
		return errors.WithMessage(err, "发送短信失败")
	}

	// 如果发送成功，扣减额度
	if resp.Code == sms.ResponseCodeSuccess {
		if err := repository.NewCompanySettingRepo(ctx.GetDB()).UpdateSmsQuota(company.Uuid, 1); err != nil {
			err := fmt.Errorf("failed to update SMS quota: %v", err)
			return errors.WithMessage(err, "扣减短信额度失败")
		}
	} else {
		err := fmt.Errorf("failed to send SMS code: %v, msg: %v", resp.Code, resp.Msg)
		return errors.WithMessage(err, "发送短信失败")
	}

	return nil
}

// SendMemberOrderRefundSMS 发送会员用餐订单退款短信
func (s *smsSrv) SendMemberOrderRefundSMS(ctx context.Context, phone string, params *sms.MemberOrderRefundRequest) error {
	company := ctx.GetCompany()
	// 禁止并发操作
	if ctx.NoLock() {
		lock.NewSystemLock().LockUuid(company.Uuid)
		defer lock.NewSystemLock().UnlockUuid(company.Uuid)
		ctx.AddLock()
	}

	formattedPhone, language, companyName, err := s.checkQuotaAndFormatPhone(ctx, phone)
	if err != nil {
		return err
	}

	// 获取公司名称
	if params.Company == "" {
		params.Company = companyName
	}

	// 发送短信
	resp, err := s.client.SendMemberOrderRefundSMS(formattedPhone, language, params)
	if err != nil {
		err := fmt.Errorf("failed to send SMS: %v", err)
		return errors.WithMessage(err, "发送短信失败")
	}

	// 如果发送成功，扣减额度
	if resp.Code == sms.ResponseCodeSuccess {
		if err := repository.NewCompanySettingRepo(ctx.GetDB()).UpdateSmsQuota(company.Uuid, 1); err != nil {
			err := fmt.Errorf("failed to update SMS quota: %v", err)
			return errors.WithMessage(err, "扣减短信额度失败")
		}
	} else {
		err := fmt.Errorf("failed to send SMS code: %v, msg: %v", resp.Code, resp.Msg)
		return errors.WithMessage(err, "发送短信失败")
	}

	return nil
}
