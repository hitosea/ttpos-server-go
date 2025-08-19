package resp

import "ttpos-server-go/app/dto"

type ReturnFoodReasonResp struct {
	List []ReturnFoodReason `json:"list"`
}

// ReturnFoodReasonResp 退菜原因响应
type ReturnFoodReason struct {
	Uuid       uint64             `json:"uuid"`
	LocaleName dto.LocaleResponse `json:"locale_name"`
}
