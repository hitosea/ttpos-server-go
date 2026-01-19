package resp

import "ttpos-server-go/app/dto"

type OrderItemRemarkResp struct {
	List []OrderItemRemark `json:"list"`
}

// OrderItemRemark 单品备注原因响应
type OrderItemRemark struct {
	Uuid       uint64             `json:"uuid"`
	LocaleName dto.LocaleResponse `json:"locale_name"`
}
