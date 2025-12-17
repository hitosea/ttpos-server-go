package dto

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"unicode/utf8"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// PageResponse 分页响应
type PageResponse struct {
	PageNo   int   `json:"page_no"`   // 当前页码
	PageSize int   `json:"page_size"` // 每页大小
	Total    int64 `json:"total"`     // 总数
}

// LocaleResponse 语言响应
type LocaleResponse struct {
	ZH   string `json:"zh"`   // 中文
	TH   string `json:"th"`   // 泰语
	EN   string `json:"en"`   // 英语
	ZHTW string `json:"zhtw"` // 粤语
	JA   string `json:"ja"`   // 日语
	KO   string `json:"ko"`   // 韩语
	MY   string `json:"my"`   // 缅甸语
	TR   string `json:"tr"`   // 土耳其语
	SV   string `json:"sv"`   // 瑞典语
}

func (l *LocaleResponse) IsNull() bool {
	return l.ZH == "" && l.TH == "" && l.EN == "" && l.ZHTW == "" && l.JA == "" && l.KO == "" && l.MY == "" && l.TR == "" && l.SV == ""
}

func (l *LocaleResponse) CheckLen(length int) bool {
	return utf8.RuneCountInString(l.ZH) <= length &&
		utf8.RuneCountInString(l.TH) <= length &&
		utf8.RuneCountInString(l.EN) <= length &&
		utf8.RuneCountInString(l.ZHTW) <= length &&
		utf8.RuneCountInString(l.JA) <= length &&
		utf8.RuneCountInString(l.KO) <= length &&
		utf8.RuneCountInString(l.MY) <= length &&
		utf8.RuneCountInString(l.TR) <= length &&
		utf8.RuneCountInString(l.SV) <= length
}

type LocaleType string

const (
	LocaleZH   LocaleType = "zh"
	LocaleTH   LocaleType = "th"
	LocaleEN   LocaleType = "en"
	LocaleZHTW LocaleType = "zhtw"
	LocaleJA   LocaleType = "ja"
	LocaleKO   LocaleType = "ko"
	LocaleMY   LocaleType = "my"
	LocaleTR   LocaleType = "tr"
	LocaleSV   LocaleType = "sv"
)

// GetLocale 获取语言
func (l *LocaleResponse) GetLocale(locale string) string {
	switch locale {
	case "zh":
		return l.ZH
	case "th":
		return l.TH
	case "en":
		return l.EN
	case "zhtw":
		return l.ZHTW
	case "ja":
		return l.JA
	case "ko":
		return l.KO
	case "my":
		return l.MY
	case "tr":
		return l.TR
	case "sv":
		return l.SV
	}
	return l.ZH
}

func (l *LocaleResponse) SetLocale(locale string, value string) {
	switch locale {
	case "zh":
		l.ZH = value
	case "th":
		l.TH = value
	case "en":
		l.EN = value
	case "zhtw":
		l.ZHTW = value
	case "ja":
		l.JA = value
	case "ko":
		l.KO = value
	case "my":
		l.MY = value
	case "tr":
		l.TR = value
	case "sv":
		l.SV = value
	}
}

func (l *LocaleResponse) SetAllEmptyLocale(value string) {
	if l.ZH == "" {
		l.ZH = value
	}
	if l.TH == "" {
		l.TH = value
	}
	if l.EN == "" {
		l.EN = value
	}
	if l.ZHTW == "" {
		l.ZHTW = value
	}
	if l.JA == "" {
		l.JA = value
	}
	if l.KO == "" {
		l.KO = value
	}
	if l.MY == "" {
		l.MY = value
	}
	if l.TR == "" {
		l.TR = value
	}
	if l.SV == "" {
		l.SV = value
	}
}

// ToJson 获取语言json
func (l *LocaleResponse) ToJson() string {
	str, _ := json.Marshal(l)
	return string(str)
}

// CheckRequiredLocale 检查是否包含所有语言
func (l *LocaleResponse) CheckRequiredLocale(locales []string) bool {
	for _, locale := range locales {
		if l.GetLocale(locale) == "" {
			return false
		}
	}
	return true
}

// CheckLenLocal 检查语言长度
func (l *LocaleResponse) CheckLenLocal(locales []string, length int) bool {
	for _, locale := range locales {
		localeValue := l.GetLocale(locale)
		if localeValue != "" && utf8.RuneCountInString(localeValue) > length {
			return false
		}
	}
	return true
}

func (l *LocaleResponse) GetMd5() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(l.ToJson())))
}

func (l *LocaleResponse) GetNames() []string {
	return []string{l.ZH, l.TH, l.EN, l.ZHTW, l.JA, l.KO, l.MY, l.TR, l.SV}
}

// ToMultiLanguageUpdateMap 转换为 MultiLanguageName 更新 map
func (l *LocaleResponse) ToMultiLanguageUpdateMap(updateTime int64) map[string]interface{} {
	return map[string]interface{}{
		"zh_name":     l.ZH,
		"en_name":     l.EN,
		"th_name":     l.TH,
		"zh_tw_name":  l.ZHTW,
		"my_name":     l.MY,
		"ja_name":     l.JA,
		"ko_name":     l.KO,
		"tr_name":     l.TR,
		"sv_name":     l.SV,
		"update_time": updateTime,
	}
}
