package erp

// Item 结构体，表示商品信息
type Item struct {
	Name                             string                `json:"name,omitempty"`                                // 名称
	Owner                            string                `json:"owner,omitempty"`                               // 拥有者
	Creation                         string                `json:"creation,omitempty"`                            // 创建时间
	Modified                         string                `json:"modified,omitempty"`                            // 修改时间
	ModifiedBy                       string                `json:"modified_by,omitempty"`                         // 修改人
	Docstatus                        int                   `json:"docstatus,omitempty"`                           // 单据状态
	Idx                              int                   `json:"idx,omitempty"`                                 // 索引
	NamingSeries                     string                `json:"naming_series,omitempty"`                       // 编码规则
	ItemCode                         string                `json:"item_code,omitempty"`                           // 商品编码
	ItemName                         string                `json:"item_name,omitempty"`                           // 商品名称
	ItemGroup                        string                `json:"item_group,omitempty"`                          // 商品分组
	StockUom                         string                `json:"stock_uom,omitempty"`                           // 库存计量单位
	Disabled                         int                   `json:"disabled,omitempty"`                            // 是否禁用
	AllowAlternativeItem             int                   `json:"allow_alternative_item,omitempty"`              // 是否允许替代商品
	IsStockItem                      int                   `json:"is_stock_item,omitempty"`                       // 是否库存商品
	HasVariants                      int                   `json:"has_variants,omitempty"`                        // 是否有变体
	OpeningStock                     float64               `json:"opening_stock,omitempty"`                       // 期初库存
	ValuationRate                    float64               `json:"valuation_rate,omitempty"`                      // 计价单价
	StandardRate                     float64               `json:"standard_rate,omitempty"`                       // 标准单价
	IsFixedAsset                     int                   `json:"is_fixed_asset,omitempty"`                      // 是否固定资产
	AutoCreateAssets                 int                   `json:"auto_create_assets,omitempty"`                  // 是否自动创建资产
	IsGroupedAsset                   int                   `json:"is_grouped_asset,omitempty"`                    // 是否分组资产
	OverDeliveryReceiptAllowance     float64               `json:"over_delivery_receipt_allowance,omitempty"`     // 超收容差
	OverBillingAllowance             float64               `json:"over_billing_allowance,omitempty"`              // 超开票容差
	Description                      string                `json:"description,omitempty"`                         // 描述
	ShelfLifeInDays                  int                   `json:"shelf_life_in_days,omitempty"`                  // 保质期天数
	EndOfLife                        string                `json:"end_of_life,omitempty"`                         // 失效日期
	DefaultMaterialRequestType       string                `json:"default_material_request_type,omitempty"`       // 默认请购类型
	ValuationMethod                  string                `json:"valuation_method,omitempty"`                    // 计价方法
	WeightPerUnit                    float64               `json:"weight_per_unit,omitempty"`                     // 单位重量
	AllowNegativeStock               int                   `json:"allow_negative_stock,omitempty"`                // 是否允许负库存
	HasBatchNo                       int                   `json:"has_batch_no,omitempty"`                        // 是否有批次号
	CreateNewBatch                   int                   `json:"create_new_batch,omitempty"`                    // 是否创建新批次
	HasExpiryDate                    int                   `json:"has_expiry_date,omitempty"`                     // 是否有过期日期
	RetainSample                     int                   `json:"retain_sample,omitempty"`                       // 是否保留样品
	SampleQuantity                   int                   `json:"sample_quantity,omitempty"`                     // 样品数量
	HasSerialNo                      int                   `json:"has_serial_no,omitempty"`                       // 是否有序列号
	VariantBasedOn                   string                `json:"variant_based_on,omitempty"`                    // 变体依据
	EnableDeferredExpense            int                   `json:"enable_deferred_expense,omitempty"`             // 是否启用递延费用
	NoOfMonthsExp                    int                   `json:"no_of_months_exp,omitempty"`                    // 递延费用月数
	EnableDeferredRevenue            int                   `json:"enable_deferred_revenue,omitempty"`             // 是否启用递延收入
	NoOfMonths                       int                   `json:"no_of_months,omitempty"`                        // 递延收入月数
	MinOrderQty                      float64               `json:"min_order_qty,omitempty"`                       // 最小订购量
	SafetyStock                      float64               `json:"safety_stock,omitempty"`                        // 安全库存
	IsPurchaseItem                   int                   `json:"is_purchase_item,omitempty"`                    // 是否采购商品
	LeadTimeDays                     int                   `json:"lead_time_days,omitempty"`                      // 采购提前期
	LastPurchaseRate                 float64               `json:"last_purchase_rate,omitempty"`                  // 最近采购价
	IsCustomerProvidedItem           int                   `json:"is_customer_provided_item,omitempty"`           // 是否客户提供商品
	DeliveredBySupplier              int                   `json:"delivered_by_supplier,omitempty"`               // 是否供应商交付
	CountryOfOrigin                  string                `json:"country_of_origin,omitempty"`                   // 原产国
	GrantCommission                  int                   `json:"grant_commission,omitempty"`                    // 是否允许佣金
	IsSalesItem                      int                   `json:"is_sales_item,omitempty"`                       // 是否销售商品
	MaxDiscount                      float64               `json:"max_discount,omitempty"`                        // 最大折扣
	InspectionRequiredBeforePurchase int                   `json:"inspection_required_before_purchase,omitempty"` // 采购前是否需要检验
	InspectionRequiredBeforeDelivery int                   `json:"inspection_required_before_delivery,omitempty"` // 发货前是否需要检验
	IncludeItemInManufacturing       int                   `json:"include_item_in_manufacturing,omitempty"`       // 是否参与生产
	IsSubContractedItem              int                   `json:"is_sub_contracted_item,omitempty"`              // 是否分包商品
	CustomerCode                     string                `json:"customer_code,omitempty"`                       // 客户编码
	TotalProjectedQty                float64               `json:"total_projected_qty,omitempty"`                 // 预计总库存
	Doctype                          string                `json:"doctype,omitempty"`                             // 单据类型
	Barcodes                         []BarCode             `json:"barcodes,omitempty"`                            // 条码
	ReorderLevels                    []interface{}         `json:"reorder_levels,omitempty"`                      // 补货水平
	Taxes                            []interface{}         `json:"taxes,omitempty"`                               // 税费
	Uoms                             []UomConversionDetail `json:"uoms,omitempty"`                                // 计量单位换算明细
	ItemDefaults                     []ItemDefault         `json:"item_defaults,omitempty"`                       // 商品默认配置
	Attributes                       []interface{}         `json:"attributes,omitempty"`                          // 属性
	SupplierItems                    []interface{}         `json:"supplier_items,omitempty"`                      // 供应商商品
	VariantOf                        string                `json:"variant_of,omitempty"`                          // 变体依据商品编码
	PurchaseUom                      string                `json:"purchase_uom,omitempty"`                        // 采购单位

	CustomerItems []interface{} `json:"customer_items,omitempty"` // 客户商品
	//自定义字段
	CustomCompany        string           `json:"custom_company,omitempty"`       // 自定义公司
	CustomBranch         string           `json:"custom_branch,omitempty"`        // 自定义分公司
	CustomSpecification  string           `json:"custom_specification,omitempty"` // 自定义规格
	Classification       string           `json:"custom_classification,omitempty"`
	ClassificationCode   string           `json:"custom_classification_code,omitempty"`
	CustomInternalCode   string           `json:"custom_internal_code,omitempty"`
	CustomPermissionRule []PermissionRule `json:"custom_permission_rule,omitempty"` //自定权限清单 多选表格
	CustomNotForSale     bool             `json:"custom_not_for_sale,omitempty"`    //是否禁售
}

type BarCode struct {
	Barcode string `json:"barcode,omitempty"` // 条码
}

// ItemDefault 结构体，表示商品默认配置
type ItemDefault struct {
	Name             string `json:"name,omitempty"`              // 名称
	Owner            string `json:"owner,omitempty"`             // 拥有者
	Creation         string `json:"creation,omitempty"`          // 创建时间
	Modified         string `json:"modified,omitempty"`          // 修改时间
	ModifiedBy       string `json:"modified_by,omitempty"`       // 修改人
	Docstatus        int    `json:"docstatus,omitempty"`         // 单据状态
	Idx              int    `json:"idx,omitempty"`               // 索引
	Company          string `json:"company,omitempty"`           // 公司
	DefaultWarehouse string `json:"default_warehouse,omitempty"` // 默认仓库
	IncomeAccount    string `json:"income_account,omitempty"`    // 收入账户
	Parent           string `json:"parent,omitempty"`            // 父级
	Parentfield      string `json:"parentfield,omitempty"`       // 父级字段
	Parenttype       string `json:"parenttype,omitempty"`        // 父级类型
	Doctype          string `json:"doctype,omitempty"`           // 单据类型
}

// UomConversionDetail 结构体，表示计量单位换算明细
type UomConversionDetail struct {
	Name             string  `json:"name,omitempty"`              // 名称
	Owner            string  `json:"owner,omitempty"`             // 拥有者
	Creation         string  `json:"creation,omitempty"`          // 创建时间
	Modified         string  `json:"modified,omitempty"`          // 修改时间
	ModifiedBy       string  `json:"modified_by,omitempty"`       // 修改人
	Docstatus        int     `json:"docstatus,omitempty"`         // 单据状态
	Idx              int     `json:"idx,omitempty"`               // 索引
	Uom              string  `json:"uom,omitempty"`               // 计量单位
	ConversionFactor float64 `json:"conversion_factor,omitempty"` // 换算系数
	Parent           string  `json:"parent,omitempty"`            // 父级
	Parentfield      string  `json:"parentfield,omitempty"`       // 父级字段
	Parenttype       string  `json:"parenttype,omitempty"`        // 父级类型
	Doctype          string  `json:"doctype,omitempty"`           // 单据类型
}

// ItemAttribute 结构体定义
// 用于表示商品属性的完整信息
type ItemAttribute struct {
	Name                string               `json:"name,omitempty"`                  // 商品属性名称
	Owner               string               `json:"owner,omitempty"`                 // 所有者
	Creation            string               `json:"creation,omitempty"`              // 创建时间
	Modified            string               `json:"modified,omitempty"`              // 修改时间
	ModifiedBy          string               `json:"modified_by,omitempty"`           // 修改者
	Docstatus           int                  `json:"docstatus,omitempty"`             // 文档状态
	Idx                 int                  `json:"idx,omitempty"`                   // 索引
	AttributeName       string               `json:"attribute_name,omitempty"`        // 属性名称
	NumericValues       int                  `json:"numeric_values,omitempty"`        // 是否为数值型属性
	Disabled            int                  `json:"disabled,omitempty"`              // 是否禁用
	FromRange           float64              `json:"from_range,omitempty"`            // 范围起始值
	Increment           float64              `json:"increment,omitempty"`             // 增量值
	ToRange             float64              `json:"to_range,omitempty"`              // 范围结束值
	Doctype             string               `json:"doctype,omitempty"`               // 文档类型
	ItemAttributeValues []ItemAttributeValue `json:"item_attribute_values,omitempty"` // 商品属性值列表
	//自定义字段
	CustomCompany        string           `json:"custom_company,omitempty"`         // 自定义公司
	CustomBranch         string           `json:"custom_branch,omitempty"`          // 自定义分公司
	CustomAlias          string           `json:"custom_alias,omitempty"`           // 自定义别名
	CustomPermissionRule []PermissionRule `json:"custom_permission_rule,omitempty"` //自定权限清单 多选表格
}

// ItemAttributeValue 结构体定义
// 用于表示商品属性中的属性值信息
type ItemAttributeValue struct {
	Name           string `json:"name,omitempty"`            // 名称
	Owner          string `json:"owner,omitempty"`           // 所有者
	Creation       string `json:"creation,omitempty"`        // 创建时间
	Modified       string `json:"modified,omitempty"`        // 修改时间
	ModifiedBy     string `json:"modified_by,omitempty"`     // 修改者
	Docstatus      int    `json:"docstatus,omitempty"`       // 文档状态
	Idx            int    `json:"idx,omitempty"`             // 索引
	AttributeValue string `json:"attribute_value,omitempty"` // 属性值
	Abbr           string `json:"abbr,omitempty"`            // 缩写
	Parent         string `json:"parent,omitempty"`          // 父级
	Parentfield    string `json:"parentfield,omitempty"`     // 父级字段
	Parenttype     string `json:"parenttype,omitempty"`      // 父级类型
	Doctype        string `json:"doctype,omitempty"`         // 文档类型
}

type ItemGroup struct {
	ItemGroupName   string `json:"item_group_name"`             // 商品组名称
	IsGroup         bool   `json:"is_group,omitempty"`          // 是否为商品组
	ParentItemGroup string `json:"parent_item_group,omitempty"` // 父级商品组
	CustomCompany   string `json:"custom_company,omitempty"`    // 自定义公司
	CustomBranch    string `json:"custom_branch,omitempty"`     // 自定义分公司
}
