package setting

// Balance 充值设置
type Balance struct {
	IsOpen   string `json:"is_open"`   // 是否开启
	IsPlan   string `json:"is_plan"`   // 是否可以自定义
	MinMoney int    `json:"min_money"` // 最低充值金额
	Describe string `json:"describe"`  // 充值说明
}
