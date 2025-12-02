package model

const (
	ExportTypeBusinessData              = 1 // 时段营业统计
	ExportTypeBusinessDataSummary       = 2 // 综合运营统计
	ExportTypeBusinessDataPaymentMethod = 3 // 营业收款统计
	ExportTypeKitchenProductionDetail   = 4 // 菜品出品明细
	ExportTypeKitchenEfficiencyAnalysis = 5 // 菜品出品详情
	ExportTypeProductSales              = 6 // 商品销售统计
	ExportTypeChannelSales              = 7 // 渠道营业统计
	ExportTypeUserAnalysis              = 8 // 用户分析统计
)

const (
	ExportStatusPending = 0 // 导出中
	ExportStatusSuccess = 1 // 导出成功
	ExportStatusFailed  = 2 // 导出失败
)

// ExportRecord 导出记录表 ttpos_export_record
type ExportRecord struct {
	BaseModel
	ExportType   uint8  `gorm:"column:export_type;type:tinyint;default:0;comment:导出类型: 1-时段营业统计, 2-综合运营统计, 3-营业应收统计, 4-菜品出品明细, 5-菜品出品详情;NOT NULL" json:"export_type"`
	ExportName   string `gorm:"column:export_name;type:varchar(200);default:'';comment:导出文件名称;NOT NULL" json:"export_name"`
	FileUuid     uint64 `gorm:"column:file_uuid;type:bigint;default:0;comment:文件UUID，关联ttpos_file表;NOT NULL" json:"file_uuid"`
	Status       uint8  `gorm:"column:status;type:tinyint;default:0;comment:状态: 0-导出中, 1-导出成功, 2-导出失败;NOT NULL" json:"status"`
	ErrorMsg     string `gorm:"column:error_msg;type:text;comment:错误信息" json:"error_msg"`
	ExportParams string `gorm:"column:export_params;type:text;comment:导出参数JSON" json:"export_params"`
	StaffUuid    uint64 `gorm:"column:staff_uuid;type:bigint;default:0;comment:操作员工UUID;NOT NULL" json:"staff_uuid"`

	// 关联
	File *File `gorm:"foreignKey:FileUuid;references:Uuid" json:"file,omitempty"`
}
