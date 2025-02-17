package service

import (
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// IPaymentMethodSrv 定义支付方式服务接口
type IPaymentMethodSrv interface {
	IsEnabled(ctx context.Context, paymentMethod model.PaymentMethod, companySetting model.CompanySetting) bool // 支付方式是否已启用
	CalculatePaymentCommissionFee(paymentMethod model.PaymentMethod, paymentAmount float64) float64             // 计算税费
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

// CalculatePaymentCommissionFee 计算支付手续费
func (s *paymentMethodSrv) CalculatePaymentCommissionFee(paymentMethod model.PaymentMethod, paymentAmount float64) float64 {
	// 将 paymentAmount 和 fee/100 转换为 decimal
	decimalPrice := decimal.NewFromFloat(paymentAmount)
	feeRate := decimal.NewFromFloat(paymentMethod.FeePercent).Div(decimal.NewFromFloat(100))

	// 计算费用，先保留3位小数，然后四舍五入到2位
	feeMoney := decimalPrice.Mul(feeRate).Round(3).Round(2)

	// 转换回 float64
	result, _ := feeMoney.Float64()
	return result
}
