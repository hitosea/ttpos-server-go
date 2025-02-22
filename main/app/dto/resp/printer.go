package resp

type PrinterLogData struct {
	PrinterUuid     uint64 `json:"printer_uuid"`      // 打印机uuid
	CashierDeviceId string `json:"cashier_device_id"` // 收银机绑定的设备ID
	DataType        int    `json:"data_type"`         // 数据类型 1-预结账单 2-结账单 3-一菜一单 4-整单打印 5-打印发票 6-打印营业数据 7-打印交班单;
	Data            string `json:"data"`              // 打印数据
	Type            int    `json:"type"`              // 类型:0系统默认队列,1云上服务下放
	FirstExecution  int    `json:"first_execution"`   // 是否首次执行打印 1-是 0-否;
	//OrderId         uint64 `json:"order_id"`          // 关联销售订单ID，废弃，直接传递No
	No string `json:"no"` // 桌台号或者呼叫号(如果有)

	Status int    `json:"status"` // 状态(0结束,1进行中,2成功)
	Reason string `json:"reason"` // 原因

	PrintMethod int   `json:"print_method"` // 打印方式 1文本打印, 2图片打印'
	Num         int   `json:"num"`          // 打印次数
	PrinterTime int64 `json:"printer_time"` // 打印时间

	Uuid          uint64 `json:"uuid"`           // 打印日志Uuid
	Copies        uint   `json:"copies"`         // 打印机.份数 => 对应旧表print_times
	PrinterType   string `json:"printer_type"`   // 打印机.类型
	PrinterConfig string `json:"printer_config"` // 打印机.配置
	CreateTime    int64  `json:"create_time"`    // 日志创建时间戳
}

type PrinterInfo struct {
	PrinterType   string // 打印机类型
	PrinterConfig string // 打印机设置
	PrintCopies   uint   // 打印次数
}
