package constant

// 导出类型常量
const (
	ExportTypeTimeBusiness           uint8 = 1 // 时段营业统计
	ExportTypeComprehensiveOperation uint8 = 2 // 综合运营统计
	ExportTypeBusinessReceivable     uint8 = 3 // 营业应收统计
	ExportTypeDishDetail             uint8 = 4 // 菜品出品明细
	ExportTypeDishEfficiency         uint8 = 5 // 菜品出品详情
)

// 导出记录状态常量
const (
	ExportRecordStatusExporting uint8 = 0 // 导出中
	ExportRecordStatusSuccess   uint8 = 1 // 导出成功
	ExportRecordStatusFailed    uint8 = 2 // 导出失败
)
