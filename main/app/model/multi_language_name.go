package model

import (
	"encoding/json"
	"ttpos-server-go/app/dto"
)

// MultiLanguageName 结构体表示多语言名称 ttpos_multi_language_name
type MultiLanguageName struct {
	BaseModel
	EnName   string `gorm:"default:'';column:en_name;comment:'英文名称'"`
	ZhName   string `gorm:"default:'';column:zh_name;comment:'中文名称'"`
	ZhTwName string `gorm:"default:'';column:zh_tw_name;comment:'繁体中文名称'"`
	ThName   string `gorm:"default:'';column:th_name;comment:'泰语名称'"`
	MyName   string `gorm:"default:'';column:my_name;comment:'缅甸语名称'"`
	JaName   string `gorm:"default:'';column:ja_name;comment:'日语名称'"`
	KoName   string `gorm:"default:'';column:ko_name;comment:'韩语名称'"`
	TrName   string `gorm:"default:'';column:tr_name;comment:'土耳其语名称'"`
	SvName   string `gorm:"default:'';column:sv_name;comment:'瑞典语名称'"`
}

func (m *MultiLanguageName) IsNullName() bool {
	return m.ZhName == "" && m.ThName == "" && m.EnName == "" && m.ZhTwName == "" && m.JaName == "" && m.KoName == "" && m.MyName == "" && m.TrName == "" && m.SvName == ""
}

func NewMultiLanguageName(jsonStr string) *MultiLanguageName {
	multiLanguageName := MultiLanguageName{}
	multiLanguageName.InitByLocaleResponseJson(jsonStr)
	return &multiLanguageName
}

func (m *MultiLanguageName) InitByLocaleResponse(locale dto.LocaleResponse) {
	m.ZhName = locale.ZH
	m.ThName = locale.TH
	m.EnName = locale.EN
	m.ZhTwName = locale.ZHTW
	m.JaName = locale.JA
	m.KoName = locale.KO
	m.MyName = locale.MY
	m.TrName = locale.TR
	m.SvName = locale.SV
}

func (m *MultiLanguageName) InitByLocaleResponseJson(jsonStr string) {
	names := dto.LocaleResponse{}
	json.Unmarshal([]byte(jsonStr), &names)
	m.InitByLocaleResponse(names)
}

// IsNameChanged 判断多语言名称是否修改
func (m *MultiLanguageName) IsNameChanged(locale dto.LocaleResponse) bool {
	names := m.GetNames()
	namesJson := names.ToJson()
	localeJson := locale.ToJson()
	return namesJson != localeJson
}

// GetNames 获取多语言名称
func (m *MultiLanguageName) GetNames() dto.LocaleResponse {
	return dto.LocaleResponse{
		ZH:   m.ZhName,
		TH:   m.ThName,
		EN:   m.EnName,
		ZHTW: m.ZhTwName,
		JA:   m.JaName,
		KO:   m.KoName,
		MY:   m.MyName,
		TR:   m.TrName,
		SV:   m.SvName,
	}
}

func (m *MultiLanguageName) ToJson() string {
	names := m.GetNames()
	json, _ := json.Marshal(names)
	return string(json)
}

func (m *MultiLanguageName) ToMap() map[string]string {
	return map[string]string{
		"zh":   m.ZhName,
		"en":   m.EnName,
		"th":   m.ThName,
		"zhtw": m.ZhTwName,
		"ja":   m.JaName,
		"ko":   m.KoName,
		"my":   m.MyName,
		"tr":   m.TrName,
		"sv":   m.SvName,
	}
}

// GetNameByLang 获取指定语言名称
func (m *MultiLanguageName) GetNameByLang(lang string) string {
	switch lang {
	case "zh":
		return m.ZhName
	case "th":
		return m.ThName
	case "en":
		return m.EnName
	case "zhtw":
		return m.ZhTwName
	case "zh_tw":
		return m.ZhTwName
	case "ja":
		return m.JaName
	case "ko":
		return m.KoName
	case "my":
		return m.MyName
	case "tr":
		return m.TrName
	case "sv":
		return m.SvName
	default:
		return ""
	}
}

// GetNameByLangWithFallback 获取指定语言名称，如果不存在则按优先级顺序尝试其他语言
// 优先级顺序：1. 指定语言 2. 英语(en) 3. 其他语言（中文、泰语、繁体中文、日语、韩语、缅甸语、土耳其语、瑞典语）
func (m *MultiLanguageName) GetNameByLangWithFallback(lang string) string {
	// 1. 先尝试获取指定语言的名称
	if name := m.GetNameByLang(lang); name != "" {
		return name
	}

	// 2. 如果指定语言不存在，尝试英语
	if lang != "en" {
		if name := m.GetNameByLang("en"); name != "" {
			return name
		}
	}

	// 3. 如果英语也不存在，按顺序尝试其他语言
	fallbackLangs := []string{"zh", "th", "zhtw", "ja", "ko", "my", "tr", "sv"}
	for _, fallbackLang := range fallbackLangs {
		// 跳过已经尝试过的语言
		if fallbackLang == lang || fallbackLang == "en" {
			continue
		}
		if name := m.GetNameByLang(fallbackLang); name != "" {
			return name
		}
	}

	// 如果所有语言都没有名称，返回空字符串
	return ""
}
