package erp

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GenerateModeOfPaymentID 生成 Mode of Payment ID
// 规则：{渠道}-{支付方式}-{序号}-{商家缩写}
// - 渠道：LianLianPay 或空（自行添加和系统默认）
// - 支付方式：TTPOS 的 payment_name 字段
// - 序号：系统默认=0000，自行添加=0001起，LianLianPay=0000起
// - 商家缩写：CompanyAbbr
func GenerateModeOfPaymentID(db *gorm.DB, paymentMethod *model.PaymentMethod, companyAbbr string) (string, error) {
	// 1. 根据 source 确定 channel
	channel := getChannelBySource(paymentMethod.Source)

	// 2. 构建前缀（用于查询同名支付方式）
	prefix := buildPrefix(channel, paymentMethod.PaymentName)

	// 3. 查询已有同名支付方式，确定下一个序号
	nextSeq, err := getNextSequenceNumber(db, prefix, paymentMethod.Source, companyAbbr)
	if err != nil {
		logger.Logger.Error("GenerateModeOfPaymentID-getNextSequenceNumber",
			zap.String("prefix", prefix),
			zap.Int("source", paymentMethod.Source),
			zap.String("company_abbr", companyAbbr),
			zap.Error(err))
		return "", err
	}

	// 4. 生成完整的 Mode of Payment ID
	modeOfPaymentID := fmt.Sprintf("%s%04d-%s", prefix, nextSeq, companyAbbr)

	return modeOfPaymentID, nil
}

// getChannelBySource 根据 source 确定 channel
// source=3(Kbank) -> channel="Kbank"
// source=2(LianLianPay) -> channel="LianLianPay"
// source=1(自行添加) -> channel=""
// source=0(系统默认) -> channel=""
func getChannelBySource(source int) string {
	if source == constant.PaymentMethodSourceLianLianPay {
		return "LianLianPay"
	}
	if source == constant.PaymentMethodSourceKbank {
		return "Kbank"
	}
	return ""
}

// buildPrefix 构建前缀（用于查询同名支付方式）
// 格式：{channel}-{pay_type}- 或 {pay_type}-
func buildPrefix(channel, payType string) string {
	parts := make([]string, 0, 2)
	if strings.TrimSpace(channel) != "" {
		parts = append(parts, channel)
	}
	if strings.TrimSpace(payType) != "" {
		parts = append(parts, payType)
	}
	if len(parts) > 0 {
		return strings.Join(parts, "-") + "-"
	}
	return ""
}

// getNextSequenceNumber 获取下一个序号
// 规则：
// - 系统默认（source=0）：固定使用 0000
// - 自行添加（source=1）：从 0001 起递增
// - LianLianPay（source=2）：从 0000 起递增
func getNextSequenceNumber(db *gorm.DB, prefix string, source int, companyAbbr string) (int, error) {
	// 系统默认：固定使用 0000
	if source == constant.PaymentMethodSourceSystem {
		return 0, nil
	}

	// 查询已有同名支付方式（通过 erpnext_payment 字段匹配）
	paymentMethodRepo := repository.NewPaymentMethodRepo(db)
	existingMethods := paymentMethodRepo.GetPaymentMethodList(
		repository.CommonRepo.WhereBySoftDelete(),
		repository.DBOption(func(db *gorm.DB) *gorm.DB {
			return db.Where("erpnext_payment LIKE ?", prefix+"%")
		}),
	)

	maxSeq := -1
	seqPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(prefix) + `(\d{4})-` + regexp.QuoteMeta(companyAbbr) + `$`)

	for _, pm := range existingMethods {
		if pm.ErpnextPayment == "" {
			continue
		}
		matches := seqPattern.FindStringSubmatch(pm.ErpnextPayment)
		if len(matches) == 2 {
			seq, err := strconv.Atoi(matches[1])
			if err == nil && seq > maxSeq {
				maxSeq = seq
			}
		}
	}

	// 确定下一个序号
	var nextSeq int
	if maxSeq < 0 {
		// 没有找到同名支付方式
		if source == constant.PaymentMethodSourceLianLianPay {
			nextSeq = 0 // LianLianPay 从 0000 起
		} else {
			nextSeq = 1 // 自行添加从 0001 起
		}
	} else {
		// 找到同名支付方式，递增
		nextSeq = maxSeq + 1
		// 自行添加不能使用 0000（系统默认专用）
		if source == constant.PaymentMethodSourceDefault && nextSeq == 0 {
			nextSeq = 1
		}
	}

	return nextSeq, nil
}

// GetChannelBySource 根据 source 获取 channel（公开方法，供外部调用）
func GetChannelBySource(source int) string {
	return getChannelBySource(source)
}
