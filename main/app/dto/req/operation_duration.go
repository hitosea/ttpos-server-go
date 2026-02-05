package req

// OperationDurationRecord 操作耗时记录数据结构
// 用于在内存队列中传输的数据
type OperationDurationRecord struct {
	CompanyUuid   uint64 // 商户UUID
	SaleBillUuid  uint64 // 账单UUID
	SaleOrderUuid uint64 // 订单UUID
	Action        string // 操作类型 (cancel_order/checkout/discount...)
	Source        string // 来源终端 (cashier/assistant/shop/h5...)
	StaffUuid     uint64 // 操作员工UUID
	DeviceSn      string // 设备序列号
	InstanceId    string // 服务实例标识
	StartTime     int64  // 开始时间戳（毫秒）
	EndTime       int64  // 结束时间戳（毫秒）
	DurationMs    int64  // 耗时（毫秒）
	RequestPath   string // 请求路径
	Status        int    // 操作结果: 1成功 0失败
	ErrorMsg      string // 错误信息
}
