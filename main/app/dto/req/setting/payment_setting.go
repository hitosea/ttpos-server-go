package setting

type Payment struct {
	IsCash    string `json:"is_cash"`
	IsBalance string `json:"is_balance"`
	IsOther   string `json:"is_other"`
}
