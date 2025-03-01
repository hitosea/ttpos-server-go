package resp

import "ttpos-server-go/app/dto"

// GiftOrFreeOrderReasonResps 免单原因响应
type GiftOrFreeOrderReasonResps struct {
	List []GiftOrFreeOrderReasonResp `json:"list"`
}

// GiftOrFreeOrderReasonResp 免单原因响应
type GiftOrFreeOrderReasonResp struct {
	Uuid       uint64             `json:"uuid"`
	LocaleName dto.LocaleResponse `json:"locale_name"`
}
