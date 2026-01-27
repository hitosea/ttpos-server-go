package valueobject

import "ttpos-server-go/app/dto"

// MultiLanguageName 多语言名称值对象
type MultiLanguageName struct {
	uuid uint64
	zh   string
	th   string
	en   string
	zhtw string
	ja   string
	ko   string
	my   string
	tr   string
	sv   string
}

// NewMultiLanguageName 创建多语言名称
func NewMultiLanguageName(locale dto.LocaleResponse) MultiLanguageName {
	return MultiLanguageName{
		zh:   locale.ZH,
		th:   locale.TH,
		en:   locale.EN,
		zhtw: locale.ZHTW,
		ja:   locale.JA,
		ko:   locale.KO,
		my:   locale.MY,
		tr:   locale.TR,
		sv:   locale.SV,
	}
}

// NewMultiLanguageNameWithUuid 创建带UUID的多语言名称
func NewMultiLanguageNameWithUuid(uuid uint64, locale dto.LocaleResponse) MultiLanguageName {
	return MultiLanguageName{
		uuid: uuid,
		zh:   locale.ZH,
		th:   locale.TH,
		en:   locale.EN,
		zhtw: locale.ZHTW,
		ja:   locale.JA,
		ko:   locale.KO,
		my:   locale.MY,
		tr:   locale.TR,
		sv:   locale.SV,
	}
}

// Uuid 获取UUID
func (m MultiLanguageName) Uuid() uint64 {
	return m.uuid
}

// ToLocaleResponse 转换为LocaleResponse
func (m MultiLanguageName) ToLocaleResponse() dto.LocaleResponse {
	return dto.LocaleResponse{
		ZH:   m.zh,
		TH:   m.th,
		EN:   m.en,
		ZHTW: m.zhtw,
		JA:   m.ja,
		KO:   m.ko,
		MY:   m.my,
		TR:   m.tr,
		SV:   m.sv,
	}
}

// ZH 获取中文名称
func (m MultiLanguageName) ZH() string {
	return m.zh
}

// EN 获取英文名称
func (m MultiLanguageName) EN() string {
	return m.en
}
