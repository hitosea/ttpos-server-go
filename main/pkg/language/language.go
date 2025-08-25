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
