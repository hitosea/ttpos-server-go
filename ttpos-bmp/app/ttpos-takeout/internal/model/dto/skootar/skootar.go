package skootar

import "ttpos-bmp/app/ttpos-takeout/internal/consts"

type ReqBase struct {
	ApiKey   string `json:"apiKey"`
	UserName string `json:"userName"`
	Channel  string `json:"channel"`
	// 客户参数 (可选)
	CustomerEmail string `json:"customerEmail,omitempty"`
	CustomerMoble string `json:"customerMoble,omitempty"`
}

type RespBase struct {
	ResponseCode string `json:"responseCode"`
	ResponseDesc string `json:"responseDesc"`
}
type Location struct {
	AddressName  string         `json:"addressName,omitempty"`
	Address      string         `json:"address,omitempty"`
	Lat          string         `json:"lat,omitempty"`
	Lng          string         `json:"lng,omitempty"`
	ContactName  string         `json:"contactName,omitempty"`
	ContactPhone string         `json:"contactPhone,omitempty"`
	CashFee      consts.CashFee `json:"cashFee,omitempty"` //This indicate to collect cash point from customer (available only payment type is ""cash"") The value is ""Y"" (yes) or ""N"" (no)
	Seq          int            `json:"seq"`               //Sequence of point in the trip. First point start at integer value 1, 1. pickup point, 2. delivery point

	Remark string `json:"remark,omitempty"`
}
