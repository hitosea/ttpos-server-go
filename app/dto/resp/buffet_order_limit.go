package resp

type BuffetOrderItem struct {
	IsLimitTime string `json:"is_limit_time"`
	LimitTime   string `json:"limit_time"`
	IsLimitNum  string `json:"is_limit_num"`
	LimitNum    string `json:"limit_num"`
}
