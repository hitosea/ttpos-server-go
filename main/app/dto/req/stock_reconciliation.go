package req

import "github.com/shopspring/decimal"

// StockReconciliationListReq 盘点单列表请求
type StockReconciliationListReq struct {
	PageNo          int      `json:"page_no" form:"page_no" binding:"required,min=1"`     // 页码
	PageSize        int      `json:"page_size" form:"page_size" binding:"required,min=1"` // 每页数量
	WarehouseUuids  []uint64 `json:"warehouse_uuids" form:"warehouse_uuids"`              // 仓库UUID列表
	Keyword         string   `json:"keyword" form:"keyword"`                              // 关键字（搜索单据编号和ERP盘点单号）
	StartCreateTime int      `json:"start_create_time" form:"start_create_time"`          // 创建开始时间（时间戳）
	EndCreateTime   int      `json:"end_create_time" form:"end_create_time"`              // 创建结束时间（时间戳）
	StatusIn        []int    `json:"status_in" form:"status_in"`                          // 状态列表
}

// StockReconciliationDetailReq 盘点单详情请求
type StockReconciliationDetailReq struct {
	Uuid uint64 `json:"uuid" form:"uuid" binding:"required"` // 盘点单UUID
}

// StockReconciliationItemReq 盘点单物品明细请求
type StockReconciliationItemReq struct {
	MaterialUuid uint64                            `json:"material_uuid" binding:"required"`    // 物品UUID
	Units        []*StockReconciliationItemUnitReq `json:"units" binding:"required,min=1,dive"` // 盘点单物品单位明细
}

// StockReconciliationItemUnitReq 盘点单物品单位明细请求
type StockReconciliationItemUnitReq struct {
	MaterialUnitUuid uint64           `json:"material_unit_uuid" binding:"required"` // 单位UUID
	Quantity         *decimal.Decimal `json:"quantity"`                              // 单位数量
}

// StockReconciliationSaveReq 更新盘点单请求
type StockReconciliationSaveReq struct {
	Uuid            uint64                        `json:"uuid"`              // 盘点单UUID，如果为0，表示新建
	IsSubmit        bool                          `json:"is_submit"`         // 是否提交
	SubmitAfterSave bool                          `json:"submit_after_save"` // 是否在保存后提交
	WarehouseUuid   uint64                        `json:"warehouse_uuid"`    // 仓库UUID
	Purpose         int                           `json:"purpose"`           // 盘点目的 1-库存盘点 2-期初盘点
	Type            int                           `json:"type"`              // 盘点类型 1-指定物品盘点 2-全部物品盘点
	Items           []*StockReconciliationItemReq `json:"items"`             // 盘点单物品明细
}

// StockReconciliationDeleteReq 删除盘点单请求
type StockReconciliationDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 盘点单UUID
}

// StockReconciliationApproveReq 审核盘点单请求
type StockReconciliationApproveReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 盘点单UUID
}

// StockReconciliationRejectReq 驳回盘点单请求
type StockReconciliationRejectReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 盘点单UUID
}

// StockReconciliationCopyReq 新增盘点单请求
type StockReconciliationCopyReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 盘点单UUID
}

// StockReconciliationReCheckDiffReq 复盘差异请求
type StockReconciliationReCheckDiffReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 盘点单UUID
}

type StockReconciliationCheckMaterialsReq struct {
	Uuid          uint64   `json:"uuid"`            // 盘点单UUID
	MaterialUuids []uint64 `json:"material_uuids" ` // 待盘点物品UUID列表
}
