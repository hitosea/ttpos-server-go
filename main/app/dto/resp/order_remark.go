package resp

import "ttpos-server-go/app/dto"

type OrderRemarkResp struct {
	List []OrderRemark `json:"list"`
}

// OrderRemark 整单备注响应
type OrderRemark struct {
	Uuid       uint64             `json:"uuid"`
	LocaleName dto.LocaleResponse `json:"locale_name"`
}
