package resp

import (
	"ttpos-server-go/app/dto"
)

// StockLossListResp 报损单列表响应
type StockLossListResp struct {
	List []*StockLossInfo `json:"list"` // 报损单列表
	Meta dto.PageResponse `json:"meta"` // 分页信息
}

// StockLossInfo 报损单列表信息
type StockLossInfo struct {
	Uuid                uint64             `json:"uuid"`                  // 报损单UUID
	Code                string             `json:"code"`                  // 单据编号
	ErpCode             string             `json:"erp_code"`              // ERP单据编号
	LossType            int                `json:"loss_type"`             // 报损类型 1:物品损坏 2:物品报废 3:物品过期
	WarehouseUuid       uint64             `json:"warehouse_uuid"`        // 仓库UUID
	WarehouseLocaleName dto.LocaleResponse `json:"warehouse_locale_name"` // 仓库多语言名称
	Status              int                `json:"status"`                // 状态 0:已保存 1:已提交 2:已审核通过 3:已驳回
	ItemsCount          int                `json:"items_count"`           // 物品明细数量
	SubmitTime          int                `json:"submit_time"`           // 提交时间
	CreateTime          int                `json:"create_time"`           // 创建时间
}

// StockLossDetailResp 报损单详情响应
type StockLossDetailResp struct {
	Uuid          uint64                    `json:"uuid"`           // 报损单UUID
	Code          string                    `json:"code"`           // 单据编号
	ErpCode       string                    `json:"erp_code"`       // ERP单据编号
	LossType      int                       `json:"loss_type"`      // 报损类型 1:物品损坏 2:物品报废 3:物品过期
	WarehouseUuid uint64                    `json:"warehouse_uuid"` // 仓库UUID
	WarehouseName dto.LocaleResponse        `json:"warehouse_name"` // 仓库名称（多语言）
	Reason        string                    `json:"reason"`         // 报损原因
	Status        int                       `json:"status"`         // 状态 0:已保存 1:已提交 2:已审核通过 3:已驳回
	Items         []*StockLossItemInfo      `json:"items"`          // 报损单物品明细
	Files         []*StockLossFileInfo      `json:"files"`          // 附件列表
	Annotations   []*StockLossAnnotationInfo `json:"annotations"`    // 批注列表（按创建时间倒序）
	SubmitTime    int                       `json:"submit_time"`    // 提交时间
	ApproveTime   int                       `json:"approve_time"`   // 审核通过时间
	RejectTime    int                       `json:"reject_time"`    // 驳回时间
	CreateTime    int                       `json:"create_time"`    // 创建时间
	UpdateTime    int                       `json:"update_time"`    // 更新时间
	IsCanResubmit bool                      `json:"is_can_resubmit"` // 是否可重新提交（已驳回状态且为发起人）
}

// StockLossItemInfo 报损单物品明细信息
type StockLossItemInfo struct {
	MaterialUuid    uint64                      `json:"material_uuid"`    // 物料UUID
	MaterialBarcode string                      `json:"material_barcode"` // 物料条码
	InternalCode    string                      `json:"internal_code"`    // 内部编码
	MaterialCode    string                      `json:"material_code"`    // 物料编码
	LocaleName      dto.LocaleResponse          `json:"locale_name"`      // 物料名称（多语言）
	BaseQuantity    float64                     `json:"base_quantity"`    // 基准单位数量（所有单位转换后的总量）
	ItemUnits       []*StockLossItemUnitInfo    `json:"item_units"`       // 报损单物品单位明细
	Units           []MaterialUnitInfo          `json:"units"`            // 所有单位（包含基准单位）
	CreateTime      int                         `json:"create_time"`      // 创建时间
}

// StockLossItemUnitInfo 报损单物品单位明细信息
type StockLossItemUnitInfo struct {
	MaterialUnitUuid uint64             `json:"material_unit_uuid"` // 单位UUID
	LocaleName       dto.LocaleResponse `json:"locale_name"`        // 单位名称
	Quantity         *float64           `json:"quantity"`           // 单位数量
	ConversionRate   float64            `json:"conversion_rate"`    // 转换率（相对于基准单位）
}

// StockLossFileInfo 报损单附件信息
type StockLossFileInfo struct {
	Uuid      uint64 `json:"uuid"`       // 文件UUID
	Name      string `json:"name"`       // 文件名称
	Url       string `json:"url"`        // 文件URL
	FileType  string `json:"file_type"`  // 文件类型
	SortOrder int    `json:"sort_order"` // 排序顺序
}

// StockLossAnnotationInfo 报损单批注信息
type StockLossAnnotationInfo struct {
	Uuid         uint64             `json:"uuid"`          // 批注UUID
	Action       string             `json:"action"`        // 操作类型 submit/resubmit/approve/reject
	LocaleName   dto.LocaleResponse `json:"locale_name"`   // 操作类型名称（多语言）
	Content      string             `json:"content"`       // 批注内容
	OperatorUuid uint64             `json:"operator_uuid"` // 操作人UUID
	OperatorName string             `json:"operator_name"` // 操作人姓名
	CreateTime   int                `json:"create_time"`   // 创建时间
}

// StockLossUuidResp 保存/提交报损单响应
type StockLossUuidResp struct {
	Uuid uint64 `json:"uuid"` // 报损单UUID
}

// StockLossAnnotationListResp 报损单批注列表响应
type StockLossAnnotationListResp struct {
	List []*StockLossAnnotationInfo `json:"list"` // 批注列表
}
