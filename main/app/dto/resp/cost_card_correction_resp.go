package resp

// CostCardCorrectionPreviewResp 预览修正影响响应
type CostCardCorrectionPreviewResp struct {
	Orders        []OrderCorrectionInfo `json:"orders"`         // 订单修正信息列表
	TotalOrders   int                   `json:"total_orders"`   // 总订单数
	AffectedDates []string              `json:"affected_dates"` // 受影响的日期列表
}

// OrderCorrectionInfo 订单修正信息
type OrderCorrectionInfo struct {
	OrderUuid  uint64                  `json:"order_uuid"`  // 订单UUID
	OrderNo    string                  `json:"order_no"`    // 订单号
	CreateTime int64                   `json:"create_time"` // 创建时间
	Products   []ProductCorrectionInfo `json:"products"`    // 商品修正信息列表
}

// ProductCorrectionInfo 商品修正信息
type ProductCorrectionInfo struct {
	ProductBomUuid uint64                   `json:"product_bom_uuid"` // 商品BOM UUID
	ProductName    string                   `json:"product_name"`     // 商品名称
	BomCardUuid    uint64                   `json:"bom_card_uuid"`    // 成本卡UUID
	Materials      []MaterialCorrectionInfo `json:"materials"`        // 材料修正信息列表
}

// MaterialCorrectionInfo 材料修正信息
type MaterialCorrectionInfo struct {
	MaterialUuid   uint64  `json:"material_uuid"`   // 材料UUID
	MaterialName   string  `json:"material_name"`   // 材料名称
	OldConsumption float64 `json:"old_consumption"` // 旧消耗量
	NewConsumption float64 `json:"new_consumption"` // 新消耗量
	ReturnQuantity float64 `json:"return_quantity"` // 退回数量
}

// CostCardCorrectionResp 成本卡修正执行响应
type CostCardCorrectionResp struct {
	CorrectionUuid uint64            `json:"correction_uuid"` // 修正操作UUID（用于日志追踪）
	SuccessCount   int               `json:"success_count"`   // 成功修正的订单数
	FailCount      int               `json:"fail_count"`      // 失败的订单数
	FailedOrders   []FailedOrderInfo `json:"failed_orders"`   // 失败的订单信息
}

// FailedOrderInfo 失败订单信息
type FailedOrderInfo struct {
	OrderUuid    uint64 `json:"order_uuid"`    // 订单UUID
	OrderNo      string `json:"order_no"`      // 订单号
	ErrorMessage string `json:"error_message"` // 错误信息
}

// CostCardCorrectionLog 修正日志
type CostCardCorrectionLog struct {
	Uuid           uint64 `json:"uuid"`            // 日志UUID
	CorrectionUuid uint64 `json:"correction_uuid"` // 修正操作UUID
	OrderUuid      uint64 `json:"order_uuid"`      // 订单UUID
	OrderNo        string `json:"order_no"`        // 订单号
	OperatorUuid   uint64 `json:"operator_uuid"`   // 操作人UUID
	OperatorName   string `json:"operator_name"`   // 操作人姓名
	OperationType  string `json:"operation_type"`  // 操作类型 (e.g., "preview", "execute")
	Status         string `json:"status"`          // 状态 (e.g., "success", "failed")
	Message        string `json:"message"`         // 消息
	Details        string `json:"details"`         // 详细信息 (JSON)
	CreateTime     int64  `json:"create_time"`     // 创建时间
}

// CostCardCorrectionLogsResp 修正日志列表响应
type CostCardCorrectionLogsResp struct {
	List     []CostCardCorrectionLog `json:"list"`      // 日志列表
	Total    int64                   `json:"total"`     // 总记录数
	PageNo   int                     `json:"page_no"`   // 当前页码
	PageSize int                     `json:"page_size"` // 每页大小
}
