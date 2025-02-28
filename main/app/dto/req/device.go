package req

type AddDeviceReq struct {
	DeviceId           string `json:"device_id"`            // 设备标识
	Brand              string `json:"brand"`                // 品牌
	Source             string `json:"source"`               // 来源：cashier、tablet、kitchen、assistant
	FinallyLoginUuid   uint64 `json:"finally_login_uuid"`   // 最后登录的员工ID
	FinallyLoginTime   int64  `json:"finally_login_time"`   // 最后登录的时间
	CompanyUuid        uint64 `json:"company_uuid"`         // 公司ID
	ProductPrinterUuid uint64 `json:"product_printer_uuid"` // 打印端口ID 厨显端修改设置（bind）的时候会传递
	Remark             string `json:"remark"`               // 备注 平板端绑定时会传递
}
