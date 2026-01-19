package language

import (
	"encoding/json"
	"ttpos-server-go/app/dto"
)

func JsonToLocaleResponse(jsonStr string) *dto.LocaleResponse {
	var locale dto.LocaleResponse
	err := json.Unmarshal([]byte(jsonStr), &locale)
	if err != nil {
		return &dto.LocaleResponse{}
	}
	return &dto.LocaleResponse{
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

// MapToLocaleResponse 将 map[string]string 转换为 LocaleResponse
func MapToLocaleResponse(translations map[string]string, names ...string) dto.LocaleResponse {
	locale := dto.LocaleResponse{}
	if translations != nil && len(translations) > 0 {
		if zh, ok := translations["zh"]; ok {
			locale.ZH = zh
		}
		if th, ok := translations["th"]; ok {
			locale.TH = th
		}
		if en, ok := translations["en"]; ok {
			locale.EN = en
		}
		if zhtw, ok := translations["zhtw"]; ok {
			locale.ZHTW = zhtw
		}
		if zhTW, ok := translations["zh-TW"]; ok {
			locale.ZHTW = zhTW
		}
		if ja, ok := translations["ja"]; ok {
			locale.JA = ja
		}
		if ko, ok := translations["ko"]; ok {
			locale.KO = ko
		}
		if my, ok := translations["my"]; ok {
			locale.MY = my
		}
		if tr, ok := translations["tr"]; ok {
			locale.TR = tr
		}
		if sv, ok := translations["sv"]; ok {
			locale.SV = sv
		}
	}
	if len(names) > 0 && locale.EN == "" && names[0] != "" {
		locale.EN = names[0]
	}
	return locale
}

// MergeLocaleResponses 将多个 LocaleResponse 合并成一个
// 使用指定的分隔符连接每个语言的多个值
// 参考：main/app/model/sale_order_product.go getLocaleResponse()
func MergeLocaleResponses(nameList []dto.LocaleResponse, div string) dto.LocaleResponse {
	if len(nameList) == 0 {
		return dto.LocaleResponse{}
	}
	attributeResultNames := dto.LocaleResponse{}
	for index, name := range nameList {
		attributeResultNames.ZH += name.ZH
		attributeResultNames.TH += name.TH
		attributeResultNames.EN += name.EN
		attributeResultNames.ZHTW += name.ZHTW
		attributeResultNames.JA += name.JA
		attributeResultNames.KO += name.KO
		attributeResultNames.MY += name.MY
		attributeResultNames.TR += name.TR
		attributeResultNames.SV += name.SV
		if attributeResultNames.ZH != "" && index != len(nameList)-1 {
			attributeResultNames.ZH += div
			attributeResultNames.TH += div
			attributeResultNames.EN += div
			attributeResultNames.ZHTW += div
			attributeResultNames.JA += div
			attributeResultNames.KO += div
			attributeResultNames.MY += div
			attributeResultNames.TR += div
			attributeResultNames.SV += div
		}
	}
	return attributeResultNames
}
