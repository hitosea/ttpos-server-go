package resp

import "ttpos-server-go/app/dto"

// PurchaseOrderListResp 采购订单列表响应
type PurchaseOrderListResp struct {
	List []*PurchaseOrderInfo `json:"list"` // 采购订单列表
	Meta dto.PageResponse     `json:"meta"` // 分页信息
}

// PurchaseOrderInfo 采购订单信息
type PurchaseOrderInfo struct {
	Uuid              uint64             `json:"uuid"`                // 采购订单ID
	OrderNo           string             `json:"order_no"`            // 申请单编号（单据编号）
	ErpOrderNo        string             `json:"erp_order_no"`        // ERP申请单编号（采购单号）
	Status            int                `json:"status"`              // 状态 0-待提交 1-待审核 2-已通过 3-已驳回 4-全部收货(完成) 5-待总部审核
	HeadquarterStatus int                `json:"headquarter_status"`  // V2.6 总部状态 0-待提交 1-待审核 2-已通过 3-已驳回 4-全部收货(完成)
	OrderTime         int64              `json:"order_time"`          // 单据日期
	Num               int                `json:"num"`                 // 物品数量
	OrderType         int                `json:"order_type"`          // 申请类型
	SupplierName      string             `json:"supplier_name"`       // 供应商名称
	SupplierErpCode   string             `json:"supplier_erp_code"`   // 供应商编码
	PurchaseType      int                `json:"purchase_type"`       // V2.6 采购类型 1-外部采购 2-内部采购
	WarehouseErpCode  string             `json:"warehouse_erp_code"`  // V2.6 仓库编码
	WarehouseName     dto.LocaleResponse `json:"warehouse_name"`      // V2.6 仓库名称
	ExpectArrivalTime int64              `json:"expect_arrival_time"` // 期望到货日期
	ReceiptProgress   string             `json:"receipt_progress"`    // 收货进度（百分比0.00%）前端直接显示
	CompanyUuid       uint64             `json:"company_uuid"`        // V2.6 公司UUID
	CompanyName       string             `json:"company_name"`        // V2.6 公司名称
}

// PurchaseOrderDetailResp 采购订单详情响应
type PurchaseOrderDetailResp struct {
	PurchaseOrderInfo
	Items []PurchaseOrderItemInfo `json:"items"` // 采购明细
}

type PurchaseOrderItemMaterialUnit struct {
	Uuid       uint64             `json:"uuid"`        // 单位UUID
	LocaleName dto.LocaleResponse `json:"locale_name"` // 单位名称
}

type PurchaseOrderItemUnit struct {
	Num        float64            `json:"num"`              // 数量
	ArrivalNum float64            `json:"arrival_num"`      // 到货数量
	UnitUuid   uint64             `json:"unit_uuid"`        // 单位UUID
	LocaleName dto.LocaleResponse `json:"locale_unit_name"` // 单位名称
}

// PurchaseOrderItemInfo 采购订单商品明细信息
type PurchaseOrderItemInfo struct {
	Uuid               uint64                          `json:"uuid"`                  // 明细ID
	MaterialUuid       uint64                          `json:"material_uuid"`         // 物品ID
	MaterialCode       string                          `json:"material_code"`         // 物品编码
	LocaleName         dto.LocaleResponse              `json:"locale_name"`           // 物品名称
	Num                float64                         `json:"num"`                   // 申请数量
	ArrivalNum         float64                         `json:"arrival_num"`           // 到货数量
	UnitName           string                          `json:"unit_name"`             // 采购单位名称
	LocaleUnitName     dto.LocaleResponse              `json:"locale_unit_name"`      // 采购单位名称
	UnitConversionRate float64                         `json:"unit_conversion_rate"`  // 基准单位转换率
	BaseUnitName       string                          `json:"base_unit_name"`        // 基准单位名称
	LocaleBaseUnitName dto.LocaleResponse              `json:"locale_base_unit_name"` // 基准单位名称
	InternalCode       string                          `json:"internal_code"`         // 内部编码
	BarcodeValue       string                          `json:"barcode_value"`         // 条形码值
	UnitList           []PurchaseOrderItemMaterialUnit `json:"unit_list"`             // 基准单位列表
	Units              []PurchaseOrderItemUnit         `json:"units"`                 // 单位列表
}

// PurchaseOrderLogInfo 采购订单操作日志信息
type PurchaseOrderLogInfo struct {
	Uuid              uint64 `json:"uuid"`                // 日志ID
	PurchaseOrderUuid uint64 `json:"purchase_order_uuid"` // 采购订单ID
	OperatorUuid      uint64 `json:"operator_uuid"`       // 操作人ID
	OperatorName      string `json:"operator_name"`       // 操作人姓名
	Action            string `json:"action"`              // 操作动作
	ActionDesc        string `json:"action_desc"`         // 操作描述
	OldStatus         int    `json:"old_status"`          // 操作前状态
	NewStatus         int    `json:"new_status"`          // 操作后状态
	Content           string `json:"content"`             // 操作内容详情
	Remark            string `json:"remark"`              // 备注
	CreateTime        int    `json:"create_time"`         // 创建时间
}

// PurchaseOrderCreateResp 创建采购订单响应
type PurchaseOrderCreateResp struct {
	Uuid    uint64 `json:"uuid"`     // 采购订单ID
	OrderNo string `json:"order_no"` // 采购订单编号
}

// PurchaseReceiptCreateResp 创建收货记录响应
type PurchaseReceiptCreateResp struct {
	Uuid      uint64 `json:"uuid"`       // 收货记录ID
	ReceiptNo string `json:"receipt_no"` // 收货单号
}

// PurchaseOrderStatisticsResp 采购订单统计响应
type PurchaseOrderStatisticsResp struct {
	StatusStats   []StatusStatItem   `json:"status_stats"`   // 状态统计
	PriorityStats []PriorityStatItem `json:"priority_stats"` // 优先级统计
	SupplierStats []SupplierStatItem `json:"supplier_stats"` // 供应商统计
	TimeStats     []TimeStatItem     `json:"time_stats"`     // 时间统计
	SummaryStats  SummaryStatItem    `json:"summary_stats"`  // 汇总统计
}

// StatusStatItem 状态统计项
type StatusStatItem struct {
	Status     int     `json:"status"`      // 状态
	StatusText string  `json:"status_text"` // 状态文本
	Count      int     `json:"count"`       // 数量
	Amount     float64 `json:"amount"`      // 金额
}

// PriorityStatItem 优先级统计项
type PriorityStatItem struct {
	Priority     int     `json:"priority"`      // 优先级
	PriorityText string  `json:"priority_text"` // 优先级文本
	Count        int     `json:"count"`         // 数量
	Amount       float64 `json:"amount"`        // 金额
}

// SupplierStatItem 供应商统计项
type SupplierStatItem struct {
	SupplierUuid uint64  `json:"supplier_uuid"` // 供应商ID
	SupplierName string  `json:"supplier_name"` // 供应商名称
	Count        int     `json:"count"`         // 数量
	Amount       float64 `json:"amount"`        // 金额
}

// TimeStatItem 时间统计项
type TimeStatItem struct {
	Date   string  `json:"date"`   // 日期
	Count  int     `json:"count"`  // 数量
	Amount float64 `json:"amount"` // 金额
}

// SummaryStatItem 汇总统计项
type SummaryStatItem struct {
	TotalCount         int     `json:"total_count"`           // 总订单数
	TotalAmount        float64 `json:"total_amount"`          // 总金额
	PendingCount       int     `json:"pending_count"`         // 待处理数量
	CompletedCount     int     `json:"completed_count"`       // 已完成数量
	CancelledCount     int     `json:"cancelled_count"`       // 已取消数量
	AverageAmount      float64 `json:"average_amount"`        // 平均金额
	CompletionRate     float64 `json:"completion_rate"`       // 完成率
	OnTimeDeliveryRate float64 `json:"on_time_delivery_rate"` // 按时交付率
}

// PurchaseReceiptOrderCreateResp 创建收货单响应
type PurchaseReceiptOrderCreateResp struct {
	Uuid    uint64 `json:"uuid"`     // 收货单ID
	OrderNo string `json:"order_no"` // 收货单号
}

// PurchaseReceiptOrderListResp 收货单列表响应
type PurchaseReceiptOrderListResp struct {
	PurchaseOrderList []*PurchaseOrderInfo        `json:"purchase_order_list"` // 采购单列表
	List              []*PurchaseReceiptOrderInfo `json:"list"`                // 收货单列表
	Meta              dto.PageResponse            `json:"meta"`                // 分页信息
}

// PurchaseReceiptOrderInfo 收货单信息
type PurchaseReceiptOrderInfo struct {
	Uuid              uint64 `json:"uuid"`                // 收货单ID
	Status            int    `json:"status"`              // 状态 0-待收货 1-已收货 2-已取消
	OrderNo           string `json:"order_no"`            // 订单编号
	ErpOrderNo        string `json:"erp_order_no"`        // ERP收货单号
	PurchaseOrderNo   string `json:"purchase_order_no"`   // 采购单号
	PurchaseOrderUuid uint64 `json:"purchase_order_uuid"` // 采购单ID
	Num               int    `json:"num"`                 // 物品数量
	ExpectArrivalTime int    `json:"expect_arrival_time"` // 期望到货日期
	PurchaseTime      uint64 `json:"purchase_time"`       // 采购时间
	ReceiveTime       int    `json:"receive_time"`        // 收货日期
	CancelTime        int    `json:"cancel_time"`         // 取消时间
	CreateTime        int    `json:"create_time"`         // 创建时间
}

// PurchaseReceiptItemInfo 收货明细信息
type PurchaseReceiptItemInfo struct {
	Uuid                  uint64                          `json:"uuid"`                     // 明细ID
	PurchaseOrderItemUuid uint64                          `json:"purchase_order_item_uuid"` // 采购订单明细ID
	MaterialCode          string                          `json:"material_code"`            // 物品编码
	MaterialUuid          uint64                          `json:"material_uuid"`            // 物品ID
	LocaleName            dto.LocaleResponse              `json:"locale_name"`              // 物品名称
	MaterialName          string                          `json:"material_name"`            // 物品名称
	PurchaseNum           float64                         `json:"purchase_num"`             // 采购数量
	ArrivalNum            float64                         `json:"arrival_num"`              // 到货数量
	UnitName              string                          `json:"unit_name"`                // 采购单位名称
	LocaleUnitName        dto.LocaleResponse              `json:"locale_unit_name"`         // 采购单位名称
	UnitConversionRate    float64                         `json:"unit_conversion_rate"`     // 基准单位转换率
	BaseUnitName          string                          `json:"base_unit_name"`           // 基准单位名称
	LocaleBaseUnitName    dto.LocaleResponse              `json:"locale_base_unit_name"`    // 基准单位名称
	InternalCode          string                          `json:"internal_code"`            // 商品Bom内部编码
	BarcodeValue          string                          `json:"barcode_value"`            // 条形码值
	UnitList              []PurchaseOrderItemMaterialUnit `json:"unit_list"`                // 单位列表
	Units                 []PurchaseOrderItemUnit         `json:"units"`                    // 单位列表
}

// PurchaseReceiptOrderDetailResp 收货单详情响应
type PurchaseReceiptOrderDetailResp struct {
	PurchaseReceiptOrderInfo
	Items []PurchaseReceiptItemInfo `json:"items"` // 收货明细
	Files []ReceiptFileInfo         `json:"files"` // 附件列表
}

// ReceiptFileInfo 收货单附件信息
type ReceiptFileInfo struct {
	FileUuid   uint64 `json:"file_uuid"`   // 文件UUID
	FileName   string `json:"file_name"`   // 文件名
	FileSize   int64  `json:"file_size"`   // 文件大小（字节）
	FileType   string `json:"file_type"`   // 文件类型（image/document）
	Extension  string `json:"extension"`   // 文件扩展名
	FilePath   string `json:"file_path"`   // 文件访问路径
	SortOrder  int    `json:"sort_order"`  // 排序
	CreateTime int    `json:"create_time"` // 创建时间
}
