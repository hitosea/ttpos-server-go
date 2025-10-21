package erp

import "time"

// StockEntry 库存变动单据
type StockEntry struct {
	Name                  string             `json:"name,omitempty"`                    // 单据编号
	Owner                 string             `json:"owner,omitempty"`                   // 所有者
	Creation              time.Time          `json:"creation,omitempty"`                // 创建时间
	Modified              time.Time          `json:"modified,omitempty"`                // 修改时间
	ModifiedBy            string             `json:"modified_by,omitempty"`             // 修改人
	DocStatus             int                `json:"docstatus,omitempty"`               // 单据状态
	Idx                   int                `json:"idx,omitempty"`                     // 索引
	NamingSeries          string             `json:"naming_series,omitempty"`           // 编号系列
	StockEntryType        string             `json:"stock_entry_type,omitempty"`        // 库存变动类型
	Purpose               string             `json:"purpose,omitempty"`                 // 用途
	AddToTransit          int                `json:"add_to_transit,omitempty"`          // 添加到中转
	Company               string             `json:"company,omitempty"`                 // 公司
	PostingDate           string             `json:"posting_date,omitempty"`            // 过账日期
	PostingTime           string             `json:"posting_time,omitempty"`            // 过账时间
	SetPostingTime        int                `json:"set_posting_time,omitempty"`        // 设置过账时间
	InspectionRequired    int                `json:"inspection_required,omitempty"`     // 需要检验
	ApplyPutawayRule      int                `json:"apply_putaway_rule,omitempty"`      // 应用上架规则
	FromBom               int                `json:"from_bom,omitempty"`                // 来自BOM
	UseMultiLevelBom      int                `json:"use_multi_level_bom,omitempty"`     // 使用多级BOM
	FgCompletedQty        float64            `json:"fg_completed_qty,omitempty"`        // 完成数量
	ProcessLossPercentage float64            `json:"process_loss_percentage,omitempty"` // 工艺损耗百分比
	ProcessLossQty        float64            `json:"process_loss_qty,omitempty"`        // 工艺损耗数量
	TotalOutgoingValue    float64            `json:"total_outgoing_value,omitempty"`    // 总出库价值
	TotalIncomingValue    float64            `json:"total_incoming_value,omitempty"`    // 总入库价值
	ValueDifference       float64            `json:"value_difference,omitempty"`        // 价值差额
	TotalAdditionalCosts  float64            `json:"total_additional_costs,omitempty"`  // 总额外成本
	IsOpening             string             `json:"is_opening,omitempty"`              // 是否期初
	PerTransferred        float64            `json:"per_transferred,omitempty"`         // 转移百分比
	TotalAmount           float64            `json:"total_amount,omitempty"`            // 总金额
	IsReturn              int                `json:"is_return,omitempty"`               // 是否退货
	DocType               string             `json:"doctype,omitempty"`                 // 文档类型
	Items                 []StockEntryDetail `json:"items,omitempty"`                   // 明细项目
	AdditionalCosts       []interface{}      `json:"additional_costs,omitempty"`        // 额外成本
}

// StockEntryDetail 库存变动明细
type StockEntryDetail struct {
	Name                   string    `json:"name,omitempty"`                      // 明细编号
	Owner                  string    `json:"owner,omitempty"`                     // 所有者
	Creation               time.Time `json:"creation,omitempty"`                  // 创建时间
	Modified               time.Time `json:"modified,omitempty"`                  // 修改时间
	ModifiedBy             string    `json:"modified_by,omitempty"`               // 修改人
	DocStatus              int       `json:"docstatus,omitempty"`                 // 单据状态
	Idx                    int       `json:"idx,omitempty"`                       // 索引
	Barcode                string    `json:"barcode,omitempty"`                   // 条形码
	HasItemScanned         int       `json:"has_item_scanned,omitempty"`          // 已扫描物品
	TWarehouse             string    `json:"t_warehouse,omitempty"`               // 目标仓库
	ItemCode               string    `json:"item_code,omitempty"`                 // 物品编码
	ItemName               string    `json:"item_name,omitempty"`                 // 物品名称
	IsFinishedItem         int       `json:"is_finished_item,omitempty"`          // 是否成品
	IsScrapItem            int       `json:"is_scrap_item,omitempty"`             // 是否废料
	Description            string    `json:"description,omitempty"`               // 描述
	ItemGroup              string    `json:"item_group,omitempty"`                // 物品分组
	Qty                    float64   `json:"qty,omitempty"`                       // 数量
	TransferQty            float64   `json:"transfer_qty,omitempty"`              // 转移数量
	RetainSample           int       `json:"retain_sample,omitempty"`             // 保留样品
	Uom                    string    `json:"uom,omitempty"`                       // 单位
	StockUom               string    `json:"stock_uom,omitempty"`                 // 库存单位
	ConversionFactor       float64   `json:"conversion_factor,omitempty"`         // 转换系数
	SampleQuantity         float64   `json:"sample_quantity,omitempty"`           // 样品数量
	BasicRate              float64   `json:"basic_rate,omitempty"`                // 基础价格
	AdditionalCost         float64   `json:"additional_cost,omitempty"`           // 额外成本
	ValuationRate          float64   `json:"valuation_rate,omitempty"`            // 估值价格
	AllowZeroValuationRate int       `json:"allow_zero_valuation_rate,omitempty"` // 允许零估值价格
	SetBasicRateManually   int       `json:"set_basic_rate_manually,omitempty"`   // 手动设置基础价格
	BasicAmount            float64   `json:"basic_amount,omitempty"`              // 基础金额
	Amount                 float64   `json:"amount,omitempty"`                    // 金额
	UseSerialBatchFields   int       `json:"use_serial_batch_fields,omitempty"`   // 使用序列批次字段
	ExpenseAccount         string    `json:"expense_account,omitempty"`           // 费用科目
	CostCenter             string    `json:"cost_center,omitempty"`               // 成本中心
	ActualQty              float64   `json:"actual_qty,omitempty"`                // 实际数量
	TransferredQty         float64   `json:"transferred_qty,omitempty"`           // 已转移数量
	AllowAlternativeItem   int       `json:"allow_alternative_item,omitempty"`    // 允许替代物品
	Parent                 string    `json:"parent,omitempty"`                    // 父单据
	ParentField            string    `json:"parentfield,omitempty"`               // 父字段
	ParentType             string    `json:"parenttype,omitempty"`                // 父类型
	DocType                string    `json:"doctype,omitempty"`                   // 文档类型
}

// StockEntryConstants 库存变动常量
const (
	// StockEntryTypePurchase 采购
	StockEntryTypePurchase = "Purchase"

	// StockEntryTypeMaterialReceipt 物料入库
	StockEntryTypeMaterialReceipt = "Material Receipt"
	// StockEntryTypeMaterialIssue 物料出库
	StockEntryTypeMaterialIssue = "Material Issue"
	// StockEntryTypeMaterialTransfer 物料转移
	StockEntryTypeMaterialTransfer = "Material Transfer"
	// StockEntryTypeManufacture 生产
	StockEntryTypeManufacture = "Manufacture"
	// StockEntryTypeRepack 重新包装
	StockEntryTypeRepack = "Repack"

	// DocTypeStockEntry 库存变动单据类型
	DocTypeStockEntry = "Stock Entry"
	// DocTypeStockEntryDetail 库存变动明细类型
	DocTypeStockEntryDetail = "Stock Entry Detail"

	// DefaultMaterialRequestSeries 默认申请单命名序列
	DefaultMaterialRequestSeries = "MAT-MR-.YYYY.-"
)

// StockLedger 库存分类账结构体，记录库存变动明细
type StockLedger struct {
	ItemCode             string  `json:"item_code,omitempty"`               // 物品编码
	Date                 string  `json:"date,omitempty"`                    // 日期
	Warehouse            string  `json:"warehouse,omitempty"`               // 仓库
	PostingDate          string  `json:"posting_date,omitempty"`            // 过账日期
	PostingTime          string  `json:"posting_time,omitempty"`            // 过账时间
	ActualQty            float64 `json:"actual_qty,omitempty"`              // 实际数量
	IncomingRate         float64 `json:"incoming_rate,omitempty"`           // 入库价格
	ValuationRate        float64 `json:"valuation_rate,omitempty"`          // 估值价格
	Company              string  `json:"company,omitempty"`                 // 公司
	VoucherType          string  `json:"voucher_type,omitempty"`            // 凭证类型
	QtyAfterTransaction  float64 `json:"qty_after_transaction,omitempty"`   // 交易后数量
	StockValueDifference float64 `json:"stock_value_difference,omitempty"`  // 库存价值差额
	SerialAndBatchBundle *string `json:"serial_and_batch_bundle,omitempty"` // 序列号和批次包
	VoucherNo            string  `json:"voucher_no,omitempty"`              // 凭证编号
	StockValue           float64 `json:"stock_value,omitempty"`             // 库存价值
	BatchNo              *string `json:"batch_no,omitempty"`                // 批次号
	SerialNo             *string `json:"serial_no,omitempty"`               // 序列号
	Project              *string `json:"project,omitempty"`                 // 项目
	Name                 string  `json:"name,omitempty"`                    // 名称
	ItemName             string  `json:"item_name,omitempty"`               // 物品名称
	Description          string  `json:"description,omitempty"`             // 描述
	ItemGroup            string  `json:"item_group,omitempty"`              // 物品分组
	Brand                *string `json:"brand,omitempty"`                   // 品牌
	StockUom             string  `json:"stock_uom,omitempty"`               // 库存单位
	InQty                float64 `json:"in_qty,omitempty"`                  // 入库数量
	OutQty               float64 `json:"out_qty,omitempty"`                 // 出库数量
	InOutRate            float64 `json:"in_out_rate,omitempty"`             // 进出库价格
}

// UOM 结构体，表示计量单位信息
type UOM struct {
	Name              string `json:"name,omitempty"`                 // 单位名称
	Owner             string `json:"owner,omitempty"`                // 拥有者
	Creation          string `json:"creation,omitempty"`             // 创建时间
	Modified          string `json:"modified,omitempty"`             // 修改时间
	ModifiedBy        string `json:"modified_by,omitempty"`          // 修改人
	Docstatus         int    `json:"docstatus,omitempty"`            // 单据状态
	Idx               int    `json:"idx,omitempty"`                  // 索引
	UomName           string `json:"uom_name,omitempty"`             // 单位名称
	MustBeWholeNumber int    `json:"must_be_whole_number,omitempty"` // 是否必须为整数
	Enabled           int    `json:"enabled,omitempty"`              // 是否启用
	Doctype           string `json:"doctype,omitempty"`              // 单据类型
	// 自定义字段
	CustomAlias          string           `json:"custom_alias,omitempty"`           // 自定义别名
	CustomCompany        string           `json:"custom_company,omitempty"`         // 自定义公司
	CustomBranch         string           `json:"custom_branch,omitempty"`          // 自定义分支
	CustomPermissionRule []PermissionRule `json:"custom_permission_rule,omitempty"` // 自定义权限规则
}

// StockReconciliation 库存盘点单据
type StockReconciliation struct {
	Name             string                    `json:"name,omitempty"`              // 单据编号
	Owner            string                    `json:"owner,omitempty"`             // 所有者
	Creation         string                    `json:"creation,omitempty"`          // 创建时间
	Modified         string                    `json:"modified,omitempty"`          // 修改时间
	ModifiedBy       string                    `json:"modified_by,omitempty"`       // 修改人
	DocStatus        int                       `json:"docstatus,omitempty"`         // 单据状态
	Idx              int                       `json:"idx,omitempty"`               // 索引
	NamingSeries     string                    `json:"naming_series,omitempty"`     // 编号系列
	Company          string                    `json:"company,omitempty"`           // 公司
	Purpose          string                    `json:"purpose,omitempty"`           // 用途
	PostingDate      string                    `json:"posting_date,omitempty"`      // 过账日期
	PostingTime      string                    `json:"posting_time,omitempty"`      // 过账时间
	SetPostingTime   int                       `json:"set_posting_time,omitempty"`  // 设置过账时间
	SetWarehouse     string                    `json:"set_warehouse,omitempty"`     // 设置仓库
	ScanMode         int                       `json:"scan_mode,omitempty"`         // 扫描模式
	ExpenseAccount   string                    `json:"expense_account,omitempty"`   // 费用科目
	DifferenceAmount float64                   `json:"difference_amount,omitempty"` // 差异金额
	CostCenter       string                    `json:"cost_center,omitempty"`       // 成本中心
	DocType          string                    `json:"doctype,omitempty"`           // 文档类型
	Items            []StockReconciliationItem `json:"items,omitempty"`             // 明细项目
}

// StockReconciliationItem 库存盘点明细
type StockReconciliationItem struct {
	Name                    string  `json:"name,omitempty"`                       // 明细编号
	Owner                   string  `json:"owner,omitempty"`                      // 所有者
	Creation                string  `json:"creation,omitempty"`                   // 创建时间
	Modified                string  `json:"modified,omitempty"`                   // 修改时间
	ModifiedBy              string  `json:"modified_by,omitempty"`                // 修改人
	DocStatus               int     `json:"docstatus,omitempty"`                  // 单据状态
	Idx                     int     `json:"idx,omitempty"`                        // 索引
	ItemCode                string  `json:"item_code,omitempty"`                  // 物品编码
	ItemName                string  `json:"item_name,omitempty"`                  // 物品名称
	ItemGroup               string  `json:"item_group,omitempty"`                 // 物品分组
	Warehouse               string  `json:"warehouse,omitempty"`                  // 仓库
	Qty                     float64 `json:"qty,omitempty"`                        // 数量
	ValuationRate           float64 `json:"valuation_rate,omitempty"`             // 估值价格
	Amount                  float64 `json:"amount,omitempty"`                     // 金额
	AllowZeroValuationRate  int     `json:"allow_zero_valuation_rate,omitempty"`  // 允许零估值价格
	UseSerialBatchFields    int     `json:"use_serial_batch_fields,omitempty"`    // 使用序列批次字段
	ReconcileAllSerialBatch int     `json:"reconcile_all_serial_batch,omitempty"` // 核对所有序列批次
	CurrentQty              float64 `json:"current_qty,omitempty"`                // 当前数量
	CurrentAmount           float64 `json:"current_amount,omitempty"`             // 当前金额
	CurrentValuationRate    float64 `json:"current_valuation_rate,omitempty"`     // 当前估值价格
	QuantityDifference      string  `json:"quantity_difference,omitempty"`        // 数量差异
	AmountDifference        float64 `json:"amount_difference,omitempty"`          // 金额差异
	Parent                  string  `json:"parent,omitempty"`                     // 父单据
	ParentField             string  `json:"parentfield,omitempty"`                // 父字段
	ParentType              string  `json:"parenttype,omitempty"`                 // 父类型
	DocType                 string  `json:"doctype,omitempty"`                    // 文档类型
}

// StockReconciliationConstants 库存盘点常量
const (
	// StockReconciliationPurposeOpeningStock 期初库存
	StockReconciliationPurposeOpeningStock = "Opening Stock"
	// StockReconciliationPurposeStockReconciliation 库存盘点
	StockReconciliationPurposeStockReconciliation = "Stock Reconciliation"

	// DefaultStockReconciliationSeries 默认库存盘点命名序列
	DefaultStockReconciliationSeries = "MAT-RECO-.YYYY.-"
)
