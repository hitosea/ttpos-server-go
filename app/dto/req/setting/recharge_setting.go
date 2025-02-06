package setting

// Recharge 用户充值设置
type Recharge struct {
	IsEntrance  string `json:"is_entrance"`   // 是否允许用户充值
	IsCustom    string `json:"is_custom"`     // 是否允许自定义金额
	IsMatchPlan string `json:"is_match_plan"` // 自定义金额是否自动匹配合适的套餐
	Describe    string `json:"describe"`      // 充值说明
}
