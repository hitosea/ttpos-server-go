package model

import "ttpos-server-go/app/dto"

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
	default:
		return ""
	}
}
