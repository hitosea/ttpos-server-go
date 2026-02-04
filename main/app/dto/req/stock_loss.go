package req

import "github.com/shopspring/decimal"

// StockLossListReq 报损单列表请求
type StockLossListReq struct {
	PageNo          int      `json:"page_no" form:"page_no" binding:"required,min=1"`     // 页码
	PageSize        int      `json:"page_size" form:"page_size" binding:"required,min=1"` // 每页数量
	WarehouseUuids  []uint64 `json:"warehouse_uuids" form:"warehouse_uuids"`              // 仓库UUID列表
	Keyword         string   `json:"keyword" form:"keyword"`                              // 关键字（搜索单据编号和ERP单据编号）
	StartCreateTime int      `json:"start_create_time" form:"start_create_time"`          // 创建开始时间（时间戳）
	EndCreateTime   int      `json:"end_create_time" form:"end_create_time"`              // 创建结束时间（时间戳）
	StatusIn        []int    `json:"status_in" form:"status_in"`                          // 状态列表
	LossType        int      `json:"loss_type" form:"loss_type"`                          // 报损类型 1:物品损坏 2:物品报废 3:物品过期
}

// StockLossDetailReq 报损单详情请求
type StockLossDetailReq struct {
	Uuid uint64 `json:"uuid" form:"uuid" binding:"required"` // 报损单UUID
}

// StockLossItemReq 报损单物品明细请求
type StockLossItemReq struct {
	MaterialUuid uint64                  `json:"material_uuid"` // 物料UUID
	Units        []*StockLossItemUnitReq `json:"units"`         // 报损单物品单位明细
}

// StockLossItemUnitReq 报损单物品单位明细请求
type StockLossItemUnitReq struct {
	MaterialUnitUuid uint64           `json:"material_unit_uuid" binding:"required"` // 单位UUID
	Quantity         *decimal.Decimal `json:"quantity"`                              // 单位数量
}

// StockLossSaveReq 保存/提交报损单请求
// 当 Uuid > 0 时，可以只传 Uuid 直接提交已保存的报损单，不需要传递 Items 等字段
// 当 Uuid == 0 时，WarehouseUuid、LossType、Items 为必填
type StockLossSaveReq struct {
	Uuid          uint64              `json:"uuid"`           // 报损单UUID，如果为0，表示新建
	IsSubmit      bool                `json:"is_submit"`      // 是否提交
	WarehouseUuid uint64              `json:"warehouse_uuid"` // 仓库UUID（新建时必填）
	LossType      int                 `json:"loss_type"`      // 报损类型 1:物品损坏 2:物品报废 3:物品过期（新建时必填）
	Reason        string              `json:"reason"`         // 报损原因
	Items         []*StockLossItemReq `json:"items"`          // 报损单物品明细（新建时必填）
	FileUuids     []uint64            `json:"file_uuids"`     // 附件UUID列表

	isResubmit bool // 是否重新提交（内部使用，不暴露到接口文档）
}

// SetIsResubmit 设置是否重新提交
func (r *StockLossSaveReq) SetIsResubmit(isResubmit bool) {
	r.isResubmit = isResubmit
}

// GetIsResubmit 获取是否重新提交
func (r *StockLossSaveReq) GetIsResubmit() bool {
	return r.isResubmit
}

// StockLossDeleteReq 删除报损单请求
type StockLossDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 报损单UUID
}

// StockLossApproveReq 审核通过报损单请求
type StockLossApproveReq struct {
	Uuid       uint64 `json:"uuid" binding:"required"` // 报损单UUID
	Annotation string `json:"annotation"`              // 批注内容（非必填）
}

// StockLossRejectReq 驳回报损单请求
type StockLossRejectReq struct {
	Uuid       uint64 `json:"uuid" binding:"required"` // 报损单UUID
	Annotation string `json:"annotation"`              // 批注内容（非必填）
}

// StockLossAnnotationListReq 报损单批注列表请求
type StockLossAnnotationListReq struct {
	StockLossUuid uint64 `json:"stock_loss_uuid" form:"stock_loss_uuid" binding:"required"` // 报损单UUID
}
