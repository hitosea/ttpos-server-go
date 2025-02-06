package setting

// Currency 货币单位
type Currency struct {
	Unit             string `json:"unit"`               // 货币单位，默认泰铢
	PrintUnit        string `json:"print_unit"`         // 货币单位 - 打印专用，默认泰铢
	UnitPosition     string `json:"unit_position"`      // 主货币显示位置 0-金额前 1-金额后
	IsOpen           string `json:"is_open"`            // 副货币单位开关
	ViceUnit         string `json:"vice_unit"`          // 副货币单位
	ViceUnitPosition string `json:"vice_unit_position"` // 副货币显示位置 0-金额前 1-金额后
	UnitRate         string `json:"unit_rate"`          // 单位汇率
}
