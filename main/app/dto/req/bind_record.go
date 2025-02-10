package req

type AddBindRecordReq struct {
	DeviceId         string `json:"device_id"`          // 设备标识
	Brand            string `json:"brand"`              // 品牌
	Source           string `json:"source"`             // 来源：cashier、tablet、kitchen、assistant
	FinallyLoginUuid uint64 `json:"finally_login_uuid"` // 最后登录的员工ID
	FinallyLoginTime int    `json:"finally_login_time"` // 最后登录的时间
	CompanyUuid      uint64 `json:"company_uuid"`       // 公司ID
	PrintPortUuid    uint64 `json:"print_port_uuid"`    // 打印端口ID
	UserAgent        string `json:"user_agent"`         // 用户代理
	Remark           string `json:"remark"`             // 备注
	Address          string `json:"address"`            // 地址
	Port             uint   `json:"port"`               // 端口
	DeviceIP         string `json:"device_ip"`          // 设备IP
}
