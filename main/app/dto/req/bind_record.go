package req

type AddBindRecordReq struct {
	DeviceId         string `json:"device_id"`          // 设备标识
	Brand            string `json:"brand"`              // 品牌
	Source           string `json:"source"`             // 来源：cashier、tablet、kitchen、assistant
	FinallyLoginId   uint64 `json:"finally_login_id"`   // 最后登录的员工ID
	FinallyLoginTime int    `json:"finally_login_time"` // 最后登录的时间
	CompanyId        uint64 `json:"company_id"`         // 公司ID
	PrintPortId      int    `json:"print_port_id"`      // 打印端口ID
	UserAgent        string `json:"user_agent"`         // 用户代理
	Remark           string `json:"remark"`             // 备注
	Address          string `json:"address"`            // 地址
	Port             uint   `json:"port"`               // 端口
	DeviceIP         string `json:"device_ip"`          // 设备IP
}
