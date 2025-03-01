package resp

import "ttpos-server-go/app/dto"

type ReturnFoodReasonResps struct {
	List []ReturnFoodReasonResp `json:"list"`
}

// ReturnFoodReasonResp 退菜原因响应
type ReturnFoodReasonResp struct {
	Uuid       uint64             `json:"uuid"`
	LocaleName dto.LocaleResponse `json:"locale_name"`
}
