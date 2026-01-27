package constant

const (
	// StockReconciliationAnnotationTypeResubmit 批注类型-重新发起
	StockReconciliationAnnotationTypeResubmit = 1
	// StockReconciliationAnnotationTypeReject 批注类型-驳回
	StockReconciliationAnnotationTypeReject = 2
	// StockReconciliationAnnotationTypeApprove 批注类型-通过
	StockReconciliationAnnotationTypeApprove = 3
)

// StockReconciliationAnnotationTypeNameMap 批注类型名称映射
var StockReconciliationAnnotationTypeNameMap = map[int]string{
	StockReconciliationAnnotationTypeResubmit: "重新发起",
	StockReconciliationAnnotationTypeReject:   "驳回",
	StockReconciliationAnnotationTypeApprove:  "通过",
}
