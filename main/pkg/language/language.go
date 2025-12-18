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
