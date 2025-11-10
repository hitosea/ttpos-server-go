package model

// MaterialStockAlertLog 物料库存预警邮件记录表 `ttpos_material_stock_alert_log`
type MaterialStockAlertLog struct {
	BaseModel
	CompanyUuid   uint64  `gorm:"default:0;column:company_uuid;comment:'公司UUID'"`
	MessageUuid   uint64  `gorm:"default:0;column:message_uuid;comment:'消息UUID，每次发送时随机生成'"`
	MaterialUuid  uint64  `gorm:"default:0;column:material_uuid;comment:'物料UUID'"`
	WarehouseUuid uint64  `gorm:"default:0;column:warehouse_uuid;comment:'仓库UUID，0表示全部维度'"`
	AlertType     uint8   `gorm:"default:1;column:alert_type;comment:'预警类型：1-公司维度 2-仓库维度'"`
	CurrentStock  float64 `gorm:"type:decimal(14,4);default:0;column:current_stock;comment:'当前库存数量'"`
	SafetyStock   float64 `gorm:"type:decimal(14,4);default:0;column:safety_stock;comment:'安全库存数量'"`
	LastAlertTime uint64  `gorm:"default:0;column:last_alert_time;comment:'上次预警时间（时间戳）'"`
	AlertCount    uint32  `gorm:"default:0;column:alert_count;comment:'预警次数'"`
	SendStatus    uint8   `gorm:"default:0;column:send_status;comment:'发送状态：0-待发送 1-发送成功 2-发送失败'"`
	Recipient     string  `gorm:"default:'';column:recipient;comment:'收件人邮箱（多个用逗号分隔）'"`
	ErrorMessage  string  `gorm:"type:text;column:error_message;comment:'错误信息'"`
}

// 预警类型常量
const (
	AlertTypeCompany   uint8 = 1 // 门店维度
	AlertTypeWarehouse uint8 = 2 // 仓库维度
)

// 发送状态常量
const (
	SendStatusPending uint8 = 0 // 待发送
	SendStatusSuccess uint8 = 1 // 发送成功
	SendStatusFailed  uint8 = 2 // 发送失败
)
