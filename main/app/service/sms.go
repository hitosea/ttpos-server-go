package service

import (
	"fmt"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	srvSetting "ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/lock"

	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/sms"
)

const (
	ISmsSrvLockSuffix = 1000
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
	// SendMemberCodeSMS 发送会员验证码短信
	SendMemberCodeSMS(ctx context.Context, phone string, params *sms.MemberSendCodeRequest) error
	// SendMemberRegisterCodeSMS 发送会员注册验证码短信
	SendMemberRegisterCodeSMS(ctx context.Context, phone string, params *sms.MemberSendCodeRequest) error
	// SendMemberAuthOrderCodeSMS 发送会员认证验证码短信
	SendMemberAuthOrderCodeSMS(ctx context.Context, phone string, params *sms.MemberSendCodeRequest) error
	// SendMemberPointsSMS 发送会员积分短信
	SendMemberPointsSMS(ctx context.Context, phone string, params *sms.MemberPointsRequest) error
	// SendMemberCouponSMS 发送会员优惠券短信
	SendMemberCouponSMS(ctx context.Context, phone string, params *sms.MemberCouponRequest) error
	// SendDeliveryOrderCancelSMS 发送外送订单取消短信
	SendDeliveryOrderCancelSMS(ctx context.Context, phone string, params *sms.DeliveryOrderCancel) error
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

	if len(phone) == 10 || len(phone) == 9 {
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
	setting := repository.NewCompanySettingRepo(ctx.GetDB()).Get()
	if !setting.SmsEnabled() {
		err := fmt.Errorf("SMS service is not enabled, EnableSms: %d, SmsQuota: %d", setting.EnableSms, setting.SmsQuota)
		return "", "", "", errors.WithMessage(err, "没有开启短信或没有额度")
	}
	// 检查手机号格式
	formattedPhone, language, companyName, err := s.checkFormatPhone(ctx, phone)
	if err != nil {
		return "", "", "", err
	}
	return formattedPhone, language, companyName, nil
}

func (s *smsSrv) checkFormatPhone(ctx context.Context, phone string) (string, string, string, error) {
	// 格式化手机号
	formattedPhone, err := s.formatPhone(phone)
	if err != nil {
		return "", "", "", errors.New("手机号格式错误")
	}

	// 获取门店设置
	settingSvr := srvSetting.NewSrvImpl(s.dbm, cache.Global)
	storeSetting, err := settingSvr.GetStoreSetting(ctx)
	if err != nil {
		return "", "", "", errors.WithMessage(err, "获取门店设置失败")
	}
	var defaultLanguage string
	if len(storeSetting.Language) > 0 {
		defaultLanguage = storeSetting.Language[0].Name
	}

	// 选择语言
	language := s.selectLanguage(defaultLanguage)

	return formattedPhone, language, ctx.GetCompany().Name, nil
}

// 扣减短信额度. 同时扣减saas库和商家库的短信额度
func (s *smsSrv) deductQuota(ctx context.Context, companyUuid uint64) error {
	if err := repository.NewCompanySettingRepo(s.dbm.GetDB(ctx.GetCompanyUuid())).UpdateSmsQuota(companyUuid, 1); err != nil {
		err := fmt.Errorf("failed to update SMS quota: %v", err)
		return errors.WithMessage(err, "扣减短信额度失败")
	}
	if err := repository.NewCompanySettingRepo(s.dbm.GetDB(constant.DefaultDB)).UpdateSmsQuota(companyUuid, 1); err != nil {
		err := fmt.Errorf("failed to update saas SMS quota: %v", err)
		return errors.WithMessage(err, "saas扣减短信额度失败")
	}
	return nil
}

// SendMemberConsumptionSMS 发送会员消费短信
func (s *smsSrv) SendMemberConsumptionSMS(ctx context.Context, phone string, params *sms.MemberConsumptionRequest) error {
	company := ctx.GetCompany()

	// 禁止并发操作
	lock.NewSystemLock().LockUuid(company.Uuid + ISmsSrvLockSuffix)
	defer lock.NewSystemLock().UnlockUuid(company.Uuid + ISmsSrvLockSuffix)

	// 检查短信额度
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
		if err := s.deductQuota(ctx, company.Uuid); err != nil {
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
	lock.NewSystemLock().LockUuid(company.Uuid + ISmsSrvLockSuffix)
	defer lock.NewSystemLock().UnlockUuid(company.Uuid + ISmsSrvLockSuffix)

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
		if err := s.deductQuota(ctx, company.Uuid); err != nil {
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
	lock.NewSystemLock().LockUuid(company.Uuid + ISmsSrvLockSuffix)
	defer lock.NewSystemLock().UnlockUuid(company.Uuid + ISmsSrvLockSuffix)

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
		if err := s.deductQuota(ctx, company.Uuid); err != nil {
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
	lock.NewSystemLock().LockUuid(company.Uuid + ISmsSrvLockSuffix)
	defer lock.NewSystemLock().UnlockUuid(company.Uuid + ISmsSrvLockSuffix)

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
		if err := s.deductQuota(ctx, company.Uuid); err != nil {
			return errors.WithMessage(err, "扣减短信额度失败")
		}
	} else {
		err := fmt.Errorf("failed to send SMS code: %v, msg: %v", resp.Code, resp.Msg)
		return errors.WithMessage(err, "发送短信失败")
	}

	return nil
}

// SendMemberCodeSMS 发送会员发送验证码短信
func (s *smsSrv) SendMemberCodeSMS(ctx context.Context, phone string, params *sms.MemberSendCodeRequest) error {
	company := ctx.GetCompany()

	// 禁止并发操作
	lock.NewSystemLock().LockUuid(company.Uuid + ISmsSrvLockSuffix)
	defer lock.NewSystemLock().UnlockUuid(company.Uuid + ISmsSrvLockSuffix)

	formattedPhone, language, companyName, err := s.checkFormatPhone(ctx, phone)
	if err != nil {
		return err
	}

	// 获取公司名称
	if params.Company == "" {
		params.Company = companyName
	}

	// 发送短信
	resp, err := s.client.SendMemberCodeSMS(formattedPhone, language, params)
	if err != nil {
		err := fmt.Errorf("failed to send SMS: %v", err)
		return errors.WithMessage(err, "发送短信失败")
	}

	// 如果发送失败，返回错误
	if resp.Code != sms.ResponseCodeSuccess {
		err := fmt.Errorf("failed to send SMS code: %v, msg: %v", resp.Code, resp.Msg)
		return errors.WithMessage(err, "发送短信失败")
	}

	return nil
}

// SendMemberRegisterCodeSMS 发送会员注册验证码短信
func (s *smsSrv) SendMemberRegisterCodeSMS(ctx context.Context, phone string, params *sms.MemberSendCodeRequest) error {
	company := ctx.GetCompany()

	// 禁止并发操作
	lock.NewSystemLock().LockUuid(company.Uuid + ISmsSrvLockSuffix)
	defer lock.NewSystemLock().UnlockUuid(company.Uuid + ISmsSrvLockSuffix)

	formattedPhone, language, companyName, err := s.checkFormatPhone(ctx, phone)
	if err != nil {
		return err
	}

	// 获取公司名称
	if params.Company == "" {
		params.Company = companyName
	}

	// 发送短信
	resp, err := s.client.SendMemberRegisterCodeSMS(formattedPhone, language, params)
	if err != nil {
		err := fmt.Errorf("failed to send SMS: %v", err)
		return errors.WithMessage(err, "发送短信失败")
	}

	// 如果发送失败，返回错误
	if resp.Code != sms.ResponseCodeSuccess {
		err := fmt.Errorf("failed to send SMS code: %v, msg: %v", resp.Code, resp.Msg)
		return errors.WithMessage(err, "发送短信失败")
	}

	return nil

}

// SendMemberAuthOrderCodeSMS 发送会员认证验证码短信
func (s *smsSrv) SendMemberAuthOrderCodeSMS(ctx context.Context, phone string, params *sms.MemberSendCodeRequest) error {
	company := ctx.GetCompany()

	// 禁止并发操作
	lock.NewSystemLock().LockUuid(company.Uuid + ISmsSrvLockSuffix)
	defer lock.NewSystemLock().UnlockUuid(company.Uuid + ISmsSrvLockSuffix)

	formattedPhone, language, companyName, err := s.checkFormatPhone(ctx, phone)
	if err != nil {
		return err
	}

	// 获取公司名称
	if params.Company == "" {
		params.Company = companyName
	}

	// 发送短信
	resp, err := s.client.SendMemberAuthOrderCodeSMS(formattedPhone, language, params)
	if err != nil {
		err := fmt.Errorf("failed to send SMS: %v", err)
		return errors.WithMessage(err, "发送短信失败")
	}

	// 如果发送失败，返回错误
	if resp.Code != sms.ResponseCodeSuccess {
		err := fmt.Errorf("failed to send SMS code: %v, msg: %v", resp.Code, resp.Msg)
		return errors.WithMessage(err, "发送短信失败")
	}

	return nil

}

// SendMemberPointsSMS 发送会员积分短信
func (s *smsSrv) SendMemberPointsSMS(ctx context.Context, phone string, params *sms.MemberPointsRequest) error {
	company := ctx.GetCompany()

	// 禁止并发操作
	lock.NewSystemLock().LockUuid(company.Uuid + ISmsSrvLockSuffix)
	defer lock.NewSystemLock().UnlockUuid(company.Uuid + ISmsSrvLockSuffix)

	formattedPhone, language, companyName, err := s.checkQuotaAndFormatPhone(ctx, phone)
	if err != nil {
		return err
	}

	// 获取公司名称
	if params.Company == "" {
		params.Company = companyName
	}

	// 发送短信
	resp, err := s.client.SendMemberPointsSMS(formattedPhone, language, params)
	if err != nil {
		err := fmt.Errorf("failed to send SMS: %v", err)
		return errors.WithMessage(err, "发送短信失败")
	}

	// 如果发送成功，扣减额度
	if resp.Code == sms.ResponseCodeSuccess {
		if err := s.deductQuota(ctx, company.Uuid); err != nil {
			return errors.WithMessage(err, "扣减短信额度失败")
		}
	} else {
		err := fmt.Errorf("failed to send SMS code: %v, msg: %v", resp.Code, resp.Msg)
		return errors.WithMessage(err, "发送短信失败")
	}

	return nil
}

// SendMemberCouponSMS 发送会员优惠券短信
func (s *smsSrv) SendMemberCouponSMS(ctx context.Context, phone string, params *sms.MemberCouponRequest) error {
	company := ctx.GetCompany()

	// 禁止并发操作
	lock.NewSystemLock().LockUuid(company.Uuid + ISmsSrvLockSuffix)
	defer lock.NewSystemLock().UnlockUuid(company.Uuid + ISmsSrvLockSuffix)

	formattedPhone, language, companyName, err := s.checkQuotaAndFormatPhone(ctx, phone)
	if err != nil {
		return err
	}

	// 获取公司名称
	if params.Company == "" {
		params.Company = companyName
	}

	// 发送短信
	resp, err := s.client.SendMemberCouponSMS(formattedPhone, language, params)
	if err != nil {
		err := fmt.Errorf("failed to send SMS: %v", err)
		return errors.WithMessage(err, "发送短信失败")
	}

	// 如果发送成功，扣减额度
	if resp.Code == sms.ResponseCodeSuccess {
		if err := s.deductQuota(ctx, company.Uuid); err != nil {
			return errors.WithMessage(err, "扣减短信额度失败")
		}
	} else {
		err := fmt.Errorf("failed to send SMS code: %v, msg: %v", resp.Code, resp.Msg)
		return errors.WithMessage(err, "发送短信失败")
	}

	return nil
}

// SendDeliveryOrderBySelfCancelSMS 发送外送订单取消短信
func (s *smsSrv) SendDeliveryOrderCancelSMS(ctx context.Context, phone string, params *sms.DeliveryOrderCancel) error {
	company := ctx.GetCompany()

	// 禁止并发操作
	lock.NewSystemLock().LockUuid(company.Uuid + ISmsSrvLockSuffix)
	defer lock.NewSystemLock().UnlockUuid(company.Uuid + ISmsSrvLockSuffix)

	formattedPhone, language, companyName, err := s.checkQuotaAndFormatPhone(ctx, phone)
	if err != nil {
		return err
	}

	// 获取公司名称
	if params.Company == "" {
		params.Company = companyName
	}

	// 发送短信
	resp, err := s.client.SendDeliveryOrderCancelSMS(formattedPhone, language, params)
	if err != nil {
		err := fmt.Errorf("failed to send SMS: %v", err)
		return errors.WithMessage(err, "发送短信失败")
	}

	// 如果发送成功，扣减额度
	if resp.Code == sms.ResponseCodeSuccess {
		if err := s.deductQuota(ctx, company.Uuid); err != nil {
			return errors.WithMessage(err, "扣减短信额度失败")
		}
	} else {
		err := fmt.Errorf("failed to send SMS code: %v, msg: %v", resp.Code, resp.Msg)
		return errors.WithMessage(err, "发送短信失败")
	}

	return nil
}
