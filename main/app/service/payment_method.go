package service

import (
	"slices"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
)

// IPaymentMethodSrv 定义支付方式服务接口
type IPaymentMethodSrv interface {
	IsEnabled(ctx context.Context, paymentMethod model.PaymentMethod, companySetting model.CompanySetting) bool // 支付方式是否已启用
	GetList(ctx context.Context, typ string) resp.PaymentMethodList                                             // 获取支付方式列表
}

// paymentMethodSrv  支付方式服务结构体
type paymentMethodSrv struct {
	dbm        *database.DBManager // 数据库管理器
	settingSrv setting.ISrv
}

// NewPaymentMethodSrv 创建新的收银产品类别服务
func NewPaymentMethodSrv(dbm *database.DBManager, settingSrv setting.ISrv) IPaymentMethodSrv {
	return NewPaymentMethodSrvImpl(dbm, settingSrv)
}

// NewPaymentMethodSrvImpl 创建新的收银服务实现
func NewPaymentMethodSrvImpl(dbm *database.DBManager, settingSrv setting.ISrv) IPaymentMethodSrv {
	return &paymentMethodSrv{
		dbm:        dbm,
		settingSrv: settingSrv,
	}
}

// IsEnabled 支付方式是否已启用
func (s *paymentMethodSrv) IsEnabled(ctx context.Context, paymentMethod model.PaymentMethod, companySetting model.CompanySetting) bool {
	// 获取支付设置
	paymentSetting, err := s.settingSrv.GetPaymentSetting(ctx, companySetting)
	if err != nil {
		ctx.Log().Error("获取支付设置失败", zap.Error(err))
		return false
	}
	var availableCodes []int
	if paymentSetting.IsBalance == "1" {
		availableCodes = append(availableCodes, constant.PaymentMethodCodeBalance)
	}
	if paymentSetting.IsCash == "1" {
		availableCodes = append(availableCodes, constant.PaymentMethodCodeCash)
	}
	if paymentSetting.IsOther == "1" && paymentMethod.Status == 1 {
		availableCodes = append(availableCodes, paymentMethod.Code)
	}
	return slices.Contains(availableCodes, paymentMethod.Code)
}

// GetList 获取支付方式列表
func (s *paymentMethodSrv) GetList(ctx context.Context, typ string) resp.PaymentMethodList {
	if !slices.Contains([]string{constant.PaymentMethodShowAll, constant.PaymentMethodShowRecharge, constant.PaymentMethodShowCheckout}, typ) {
		return resp.PaymentMethodList{}
	}
	paymentMethodRepo := repository.NewPaymentMethodRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	companySetting := ctx.GetCompanySetting()
	opts := []repository.DBOption{
		paymentMethodRepo.WhereStatus(constant.PaymentMethodStatusEnable),
	}
	if ctx.GetSource() == constant.SourceCashier {
		if typ != constant.PaymentMethodShowAll {
			switch typ {
			case constant.PaymentMethodShowRecharge:
				opts = append(opts, paymentMethodRepo.WhereCashierMemberRecharge())
			case constant.PaymentMethodShowCheckout:
				opts = append(opts, paymentMethodRepo.WhereCashier())
			}
		}
	} else if ctx.GetSource() == constant.SourceAssistant {
		if typ != constant.PaymentMethodShowAll {
			switch typ {
			case constant.PaymentMethodShowRecharge:
				return resp.PaymentMethodList{}
			case constant.PaymentMethodShowCheckout:
				opts = append(opts, paymentMethodRepo.WhereAssistant())
			}
		}
	}
	opts = append(opts, paymentMethodRepo.WithLogoFile(), paymentMethodRepo.WithQrcodeFile())
	paymentMethods := paymentMethodRepo.GetPaymentMethodList(opts...)

	paymentMethodItems := make([]resp.PaymentMethodItem, 0, len(paymentMethods))
	for _, method := range paymentMethods {
		// 不显示免单
		if method.Code == constant.PaymentMethodCodeFreePay {
			continue
		}
		// 充值不显示余额
		if method.Code == constant.PaymentMethodCodeBalance &&
			(companySetting.IsOpenMember != 1 || typ == constant.PaymentMethodShowRecharge) {
			continue
		}
		var logo, qrcode string
		baseUrl := utils.GetBaseURL(ctx.GetGin().Request)
		if method.LogoFile != nil {
			logo = method.LogoFile.GetUrl(baseUrl)
		}
		if logo == "" && method.DefaultImg != "" {
			logo = strings.TrimRight(baseUrl, "/") + method.DefaultImg
		}
		if method.QrcodeFile != nil {
			qrcode = method.QrcodeFile.GetUrl(baseUrl)
		}
		paymentMethodItems = append(paymentMethodItems, resp.PaymentMethodItem{
			SourceText:    i18n.Translate(i18n.GetAcceptLanguage(ctx.GetGin()), constant.PaymentMethodSourceTextMap[method.Source]),
			Uuid:          method.Uuid,
			PaymentName:   method.PaymentName,
			PaymentMethod: method.Name,
			FeePercent:    method.FeePercent,
			Logo:          logo,
			Qrcode:        qrcode,
			Code:          method.Code,
			Source:        method.Source,
		})
	}
	return resp.PaymentMethodList{List: paymentMethodItems}
}
