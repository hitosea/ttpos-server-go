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
