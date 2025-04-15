package dto

import "encoding/json"

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
}

func (l *LocaleResponse) IsNull() bool {
	return l.ZH == "" && l.TH == "" && l.EN == "" && l.ZHTW == "" && l.JA == "" && l.KO == "" && l.MY == "" && l.TR == ""
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
	}
	return l.ZH
}

// ToJson 获取语言json
func (l *LocaleResponse) ToJson() string {
	str, _ := json.Marshal(l)
	return string(str)
}
