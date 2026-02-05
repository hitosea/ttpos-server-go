package model

// OrderOperationDuration 订单操作耗时记录 `ttpos_order_operation_duration`
// 存储在 SaaS 主库中，用于性能分析和分布式追踪
type OrderOperationDuration struct {
	Id            uint64 `gorm:"column:id;type:bigint(20) unsigned;primaryKey;autoIncrement;comment:自增ID"`
	Uuid          uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;uniqueIndex;comment:记录UUID"`
	CompanyUuid   uint64 `gorm:"column:company_uuid;type:bigint(20) unsigned;default:0;index;comment:商户UUID"`
	SaleBillUuid  uint64 `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;default:0;index;comment:账单UUID"`
	SaleOrderUuid uint64 `gorm:"column:sale_order_uuid;type:bigint(20) unsigned;default:0;comment:订单UUID"`
	Action        string `gorm:"column:action;type:varchar(100);default:'';index;comment:操作类型"`
	Source        string `gorm:"column:source;type:varchar(50);default:'';comment:来源终端"`
	StaffUuid     uint64 `gorm:"column:staff_uuid;type:bigint(20) unsigned;default:0;comment:操作员工UUID"`
	DeviceSn      string `gorm:"column:device_sn;type:varchar(255);default:'';comment:设备序列号"`
	InstanceId    string `gorm:"column:instance_id;type:varchar(128);default:'';index;comment:服务实例标识"`
	StartTime     int64  `gorm:"column:start_time;type:bigint(20) unsigned;default:0;comment:开始时间戳(毫秒)"`
	EndTime       int64  `gorm:"column:end_time;type:bigint(20) unsigned;default:0;comment:结束时间戳(毫秒)"`
	DurationMs    int    `gorm:"column:duration_ms;type:int(11) unsigned;default:0;comment:耗时(毫秒)"`
	RequestPath   string `gorm:"column:request_path;type:varchar(255);default:'';comment:请求路径"`
	Status        int    `gorm:"column:status;type:tinyint(3) unsigned;default:1;comment:操作结果(1成功 0失败)"`
	ErrorMsg      string `gorm:"column:error_msg;type:text;comment:错误信息"`
	TimeNodes     string `gorm:"column:time_nodes;type:text;comment:时间节点JSON"`
	CreateTime    int64  `gorm:"column:create_time;type:int(10) unsigned;default:0;comment:创建时间(时间戳)"`
	UpdateTime    int64  `gorm:"column:update_time;type:int(10) unsigned;default:0;comment:更新时间(时间戳)"`
	DeleteTime    int64  `gorm:"column:delete_time;type:int(10) unsigned;default:0;comment:删除时间(时间戳)"`
}

// TableName 指定表名
func (OrderOperationDuration) TableName() string {
	return "ttpos_order_operation_duration"
}
