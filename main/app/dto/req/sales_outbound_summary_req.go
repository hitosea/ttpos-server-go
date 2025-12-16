package req

// RegenerateSalesOutboundSummaryReq 重新生成销售出库汇总记录请求
type RegenerateSalesOutboundSummaryReq struct {
	CompanyUuid uint64 `json:"company_uuid" binding:"required"` // 门店UUID
	Date        string `json:"date" binding:"required"`          // 日期，格式：YYYY-MM-DD
}

