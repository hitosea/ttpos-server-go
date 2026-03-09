package constant

const (
	StaffNotHandedOver int = 0 // 未交班
	StaffHandedOver    int = 1 // 已交班

	ShiftVersionLegacy int = 1 // 旧版班次（有ERP开关帐）
	ShiftVersionNew    int = 2 // 新版班次（无ERP开关帐）
)
