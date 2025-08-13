package erp

import "ttpos-bmp/app/ttpos-erp/internal/model/dto"

// Item 结构体，表示商品信息
type Item struct {
	Name                             string                    `json:"name"`                                // 名称
	Owner                            string                    `json:"owner"`                               // 拥有者
	Creation                         string                    `json:"creation"`                            // 创建时间
	Modified                         string                    `json:"modified"`                            // 修改时间
	ModifiedBy                       string                    `json:"modified_by"`                         // 修改人
	Docstatus                        int                       `json:"docstatus"`                           // 单据状态
	Idx                              int                       `json:"idx"`                                 // 索引
	NamingSeries                     string                    `json:"naming_series"`                       // 编码规则
	ItemCode                         string                    `json:"item_code"`                           // 商品编码
	ItemName                         string                    `json:"item_name"`                           // 商品名称
	ItemGroup                        string                    `json:"item_group"`                          // 商品分组
	StockUom                         string                    `json:"stock_uom"`                           // 库存计量单位
	Disabled                         int                       `json:"disabled"`                            // 是否禁用
	AllowAlternativeItem             int                       `json:"allow_alternative_item"`              // 是否允许替代商品
	IsStockItem                      int                       `json:"is_stock_item"`                       // 是否库存商品
	HasVariants                      int                       `json:"has_variants"`                        // 是否有变体
	OpeningStock                     float64                   `json:"opening_stock"`                       // 期初库存
	ValuationRate                    float64                   `json:"valuation_rate"`                      // 计价单价
	StandardRate                     float64                   `json:"standard_rate"`                       // 标准单价
	IsFixedAsset                     int                       `json:"is_fixed_asset"`                      // 是否固定资产
	AutoCreateAssets                 int                       `json:"auto_create_assets"`                  // 是否自动创建资产
	IsGroupedAsset                   int                       `json:"is_grouped_asset"`                    // 是否分组资产
	OverDeliveryReceiptAllowance     float64                   `json:"over_delivery_receipt_allowance"`     // 超收容差
	OverBillingAllowance             float64                   `json:"over_billing_allowance"`              // 超开票容差
	Description                      string                    `json:"description"`                         // 描述
	ShelfLifeInDays                  int                       `json:"shelf_life_in_days"`                  // 保质期天数
	EndOfLife                        string                    `json:"end_of_life"`                         // 失效日期
	DefaultMaterialRequestType       string                    `json:"default_material_request_type"`       // 默认请购类型
	ValuationMethod                  string                    `json:"valuation_method"`                    // 计价方法
	WeightPerUnit                    float64                   `json:"weight_per_unit"`                     // 单位重量
	AllowNegativeStock               int                       `json:"allow_negative_stock"`                // 是否允许负库存
	HasBatchNo                       int                       `json:"has_batch_no"`                        // 是否有批次号
	CreateNewBatch                   int                       `json:"create_new_batch"`                    // 是否创建新批次
	HasExpiryDate                    int                       `json:"has_expiry_date"`                     // 是否有过期日期
	RetainSample                     int                       `json:"retain_sample"`                       // 是否保留样品
	SampleQuantity                   int                       `json:"sample_quantity"`                     // 样品数量
	HasSerialNo                      int                       `json:"has_serial_no"`                       // 是否有序列号
	VariantBasedOn                   string                    `json:"variant_based_on"`                    // 变体依据
	EnableDeferredExpense            int                       `json:"enable_deferred_expense"`             // 是否启用递延费用
	NoOfMonthsExp                    int                       `json:"no_of_months_exp"`                    // 递延费用月数
	EnableDeferredRevenue            int                       `json:"enable_deferred_revenue"`             // 是否启用递延收入
	NoOfMonths                       int                       `json:"no_of_months"`                        // 递延收入月数
	MinOrderQty                      float64                   `json:"min_order_qty"`                       // 最小订购量
	SafetyStock                      float64                   `json:"safety_stock"`                        // 安全库存
	IsPurchaseItem                   int                       `json:"is_purchase_item"`                    // 是否采购商品
	LeadTimeDays                     int                       `json:"lead_time_days"`                      // 采购提前期
	LastPurchaseRate                 float64                   `json:"last_purchase_rate"`                  // 最近采购价
	IsCustomerProvidedItem           int                       `json:"is_customer_provided_item"`           // 是否客户提供商品
	DeliveredBySupplier              int                       `json:"delivered_by_supplier"`               // 是否供应商交付
	CountryOfOrigin                  string                    `json:"country_of_origin"`                   // 原产国
	GrantCommission                  int                       `json:"grant_commission"`                    // 是否允许佣金
	IsSalesItem                      int                       `json:"is_sales_item"`                       // 是否销售商品
	MaxDiscount                      float64                   `json:"max_discount"`                        // 最大折扣
	InspectionRequiredBeforePurchase int                       `json:"inspection_required_before_purchase"` // 采购前是否需要检验
	InspectionRequiredBeforeDelivery int                       `json:"inspection_required_before_delivery"` // 发货前是否需要检验
	IncludeItemInManufacturing       int                       `json:"include_item_in_manufacturing"`       // 是否参与生产
	IsSubContractedItem              int                       `json:"is_sub_contracted_item"`              // 是否分包商品
	CustomerCode                     string                    `json:"customer_code"`                       // 客户编码
	TotalProjectedQty                float64                   `json:"total_projected_qty"`                 // 预计总库存
	Doctype                          string                    `json:"doctype"`                             // 单据类型
	Barcodes                         []BarCode                 `json:"barcodes"`                            // 条码
	ReorderLevels                    []interface{}             `json:"reorder_levels"`                      // 补货水平
	Taxes                            []interface{}             `json:"taxes"`                               // 税费
	Uoms                             []dto.UomConversionDetail `json:"uoms"`                                // 计量单位换算明细
	ItemDefaults                     []ItemDefault             `json:"item_defaults"`                       // 商品默认配置
	Attributes                       []interface{}             `json:"attributes"`                          // 属性
	SupplierItems                    []interface{}             `json:"supplier_items"`                      // 供应商商品
	CustomerItems                    []interface{}             `json:"customer_items"`                      // 客户商品
	//自定义字段
	CustomCompany       string `json:"custom_company"`       // 自定义公司
	CustomBranch        string `json:"custom_branch"`        // 自定义分公司
	CustomSpecification string `json:"custom_specification"` // 自定义规格
}

type BarCode struct {
	Barcode string `json:"barcode"` // 条码
}

// ItemDefault 结构体，表示商品默认配置
type ItemDefault struct {
	Name             string `json:"name"`              // 名称
	Owner            string `json:"owner"`             // 拥有者
	Creation         string `json:"creation"`          // 创建时间
	Modified         string `json:"modified"`          // 修改时间
	ModifiedBy       string `json:"modified_by"`       // 修改人
	Docstatus        int    `json:"docstatus"`         // 单据状态
	Idx              int    `json:"idx"`               // 索引
	Company          string `json:"company"`           // 公司
	DefaultWarehouse string `json:"default_warehouse"` // 默认仓库
	IncomeAccount    string `json:"income_account"`    // 收入账户
	Parent           string `json:"parent"`            // 父级
	Parentfield      string `json:"parentfield"`       // 父级字段
	Parenttype       string `json:"parenttype"`        // 父级类型
	Doctype          string `json:"doctype"`           // 单据类型
}
