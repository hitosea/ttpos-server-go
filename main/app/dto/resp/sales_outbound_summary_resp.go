package resp

// RegenerateSalesOutboundSummaryResp 重新生成销售出库汇总记录响应
type RegenerateSalesOutboundSummaryResp struct {
	DeletedCount   int   `json:"deleted_count"`   // 删除的记录数
	GeneratedCount int   `json:"generated_count"` // 生成的记录数
	DurationMs     int64 `json:"duration_ms"`      // 操作耗时（毫秒）
}

