package setting

// ServiceCharge 服务费
type ServiceCharge struct {
	IsOpen              string  `json:"is_open"`                // 是否开启 0关闭 1开启
	ChargeType          string  `json:"charge_type"`            // 服务费类型 1-固定金额 2-百分比
	ServiceCharge       string  `json:"service_charge"`         // 服务费金额
	ServiceChargeRate   string  `json:"service_charge_rate"`    // 服务费率
	IsOpenTax           string  `json:"is_open_tax"`            // 税费 1-收取税费 0-不收取税费
	ApplyScope          string  `json:"apply_scope"`            // 适用范围 1-全部 2-部分
	ApplyScopeOrdering  string  `json:"apply_scope_ordering"`   // 适用范围-点餐 0-关闭 1-开启
	ApplyScopeTable     string  `json:"apply_scope_table"`      // 适用范围-桌台 0-关闭 1-开启
	ApplyScopeTableList []int64 `json:"apply_scope_table_list"` // 适用范围-桌台id列表
}
