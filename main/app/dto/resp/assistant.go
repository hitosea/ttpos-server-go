package resp

type OnlineCashier struct {
	CashierUuid uint64 `json:"cashier_uuid"` // 收银员Uuid
	Username    string `json:"username"`     // 收银员账号
	DeviceId    string `json:"device_id"`    // 收银机设备ID
	Remark      string `json:"remark"`       // 收银机备注
}
