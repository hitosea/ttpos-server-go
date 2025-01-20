package req

type AddBindRecordReq struct {
	DeviceID         string `json:"device_id"`
	Brand            string `json:"brand"`
	Source           string `json:"source"`
	FinallyLoginId   uint   `json:"finally_login_id"`
	FinallyLoginTime int    `json:"finally_login_time"`
	AppId            uint   `json:"app_id"`
	ShopSupplierId   uint   `json:"shop_supplier_id"`
	PrintPortId      int    `json:"print_port_id"`
	UserAgent        string `json:"user_agent"`
	Remark           string `json:"remark"`
	Address          string `json:"address"`
	Port             uint   `json:"port"`
	DeviceIP         string `json:"device_ip"`
}
