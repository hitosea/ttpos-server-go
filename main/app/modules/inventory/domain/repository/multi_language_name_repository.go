package repository

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/pkg/context"
)

// IMultiLanguageNameRepository 多语言名称仓储接口（领域层定义）
type IMultiLanguageNameRepository interface {
	// Save 保存多语言名称（创建或更新）
	Save(ctx context.Context, data *MultiLanguageNameData) error

	// FindByUuid 根据UUID查找多语言名称
	FindByUuid(ctx context.Context, uuid uint64) (*MultiLanguageNameData, error)

	// Remove 删除多语言名称
	Remove(ctx context.Context, uuid uint64) error
}

// MultiLanguageNameData 多语言名称数据
type MultiLanguageNameData struct {
	Uuid uint64
	ZH   string
	TH   string
	EN   string
	ZHTW string
	JA   string
	KO   string
	MY   string
	TR   string
	SV   string
}

// ToLocaleResponse 转换为 LocaleResponse
func (m *MultiLanguageNameData) ToLocaleResponse() dto.LocaleResponse {
	return dto.LocaleResponse{
		ZH:   m.ZH,
		TH:   m.TH,
		EN:   m.EN,
		ZHTW: m.ZHTW,
		JA:   m.JA,
		KO:   m.KO,
		MY:   m.MY,
		TR:   m.TR,
		SV:   m.SV,
	}
}

// NewMultiLanguageNameData 从 LocaleResponse 创建
func NewMultiLanguageNameData(locale dto.LocaleResponse) *MultiLanguageNameData {
	return &MultiLanguageNameData{
		ZH:   locale.ZH,
		TH:   locale.TH,
		EN:   locale.EN,
		ZHTW: locale.ZHTW,
		JA:   locale.JA,
		KO:   locale.KO,
		MY:   locale.MY,
		TR:   locale.TR,
		SV:   locale.SV,
	}
}
