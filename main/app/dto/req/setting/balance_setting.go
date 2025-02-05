package setting

type Balance struct {
	IsOpen   string `json:"is_open"`
	IsPlan   string `json:"is_plan"`
	MinMoney int    `json:"min_money"`
	Describe string `json:"describe"`
}
