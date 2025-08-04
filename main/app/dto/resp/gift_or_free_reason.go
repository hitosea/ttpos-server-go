package resp

import "ttpos-server-go/app/dto"

// GiftOrFreeOrderReasonResp 免单原因响应
type GiftOrFreeOrderReasonResp struct {
	List []GiftOrFreeOrderReason `json:"list"`
}

// GiftOrFreeOrderReasonResp 免单原因响应
type GiftOrFreeOrderReason struct {
	Uuid       uint64             `json:"uuid"`
	LocaleName dto.LocaleResponse `json:"locale_name"`
}
