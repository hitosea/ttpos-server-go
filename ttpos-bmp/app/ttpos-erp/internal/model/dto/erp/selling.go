package erp

// POSProfile 结构体定义
// 用于表示POS配置文件的完整信息
type POSProfile struct {
	Name                          string             `json:"name,omitempty"`                               // POS配置文件名称
	Owner                         string             `json:"owner,omitempty"`                              // 所有者
	Creation                      string             `json:"creation,omitempty"`                           // 创建时间
	Modified                      string             `json:"modified,omitempty"`                           // 修改时间
	ModifiedBy                    string             `json:"modified_by,omitempty"`                        // 修改者
	Docstatus                     int                `json:"docstatus,omitempty"`                          // 文档状态
	Idx                           int                `json:"idx,omitempty"`                                // 索引
	Company                       string             `json:"company,omitempty"`                            // 公司
	Country                       string             `json:"country,omitempty"`                            // 国家
	Disabled                      int                `json:"disabled,omitempty"`                           // 是否禁用
	Warehouse                     string             `json:"warehouse,omitempty"`                          // 仓库
	HideImages                    int                `json:"hide_images,omitempty"`                        // 是否隐藏图片
	HideUnavailableItems          int                `json:"hide_unavailable_items,omitempty"`             // 是否隐藏不可用商品
	AutoAddItemToCart             int                `json:"auto_add_item_to_cart,omitempty"`              // 是否自动添加商品到购物车
	ValidateStockOnSave           int                `json:"validate_stock_on_save,omitempty"`             // 保存时是否验证库存
	PrintReceiptOnOrderComplete   int                `json:"print_receipt_on_order_complete,omitempty"`    // 订单完成时是否打印收据
	UpdateStock                   int                `json:"update_stock,omitempty"`                       // 是否更新库存
	IgnorePricingRule             int                `json:"ignore_pricing_rule,omitempty"`                // 是否忽略定价规则
	AllowRateChange               int                `json:"allow_rate_change,omitempty"`                  // 是否允许汇率变更
	AllowDiscountChange           int                `json:"allow_discount_change,omitempty"`              // 是否允许折扣变更
	DisableGrandTotalToDefaultMop int                `json:"disable_grand_total_to_default_mop,omitempty"` // 是否禁用总计到默认支付方式
	AllowPartialPayment           int                `json:"allow_partial_payment,omitempty"`              // 是否允许部分支付
	SellingPriceList              string             `json:"selling_price_list,omitempty"`                 // 销售价格表
	Currency                      string             `json:"currency,omitempty"`                           // 货币
	WriteOffAccount               string             `json:"write_off_account,omitempty"`                  // 核销账户
	WriteOffCostCenter            string             `json:"write_off_cost_center,omitempty"`              // 核销成本中心
	WriteOffLimit                 float64            `json:"write_off_limit,omitempty"`                    // 核销限额
	DisableRoundedTotal           int                `json:"disable_rounded_total,omitempty"`              // 是否禁用四舍五入总计
	ApplyDiscountOn               string             `json:"apply_discount_on,omitempty"`                  // 折扣应用位置
	Doctype                       string             `json:"doctype,omitempty"`                            // 文档类型
	CustomerGroups                []interface{}      `json:"customer_groups,omitempty"`                    // 客户组列表
	ApplicableForUsers            []POSProfileUser   `json:"applicable_for_users,omitempty"`               // 适用用户列表
	ItemGroups                    []interface{}      `json:"item_groups,omitempty"`                        // 商品组列表
	Payments                      []POSPaymentMethod `json:"payments,omitempty"`                           // 支付方式列表
	Branch                        string             `json:"branch,omitempty"`                             // 分公司
}

// POSProfileUser 结构体定义
// 用于表示POS配置文件中的用户信息
type POSProfileUser struct {
	Name        string `json:"name,omitempty"`        // 名称
	Owner       string `json:"owner,omitempty"`       // 所有者
	Creation    string `json:"creation,omitempty"`    // 创建时间
	Modified    string `json:"modified,omitempty"`    // 修改时间
	ModifiedBy  string `json:"modified_by,omitempty"` // 修改者
	Docstatus   int    `json:"docstatus,omitempty"`   // 文档状态
	Idx         int    `json:"idx,omitempty"`         // 索引
	Default     int    `json:"default,omitempty"`     // 是否默认
	User        string `json:"user,omitempty"`        // 用户
	Parent      string `json:"parent,omitempty"`      // 父级
	Parentfield string `json:"parentfield,omitempty"` // 父级字段
	Parenttype  string `json:"parenttype,omitempty"`  // 父级类型
	Doctype     string `json:"doctype,omitempty"`     // 文档类型
}

// POSPaymentMethod 结构体定义
// 用于表示POS配置文件中的支付方式信息
type POSPaymentMethod struct {
	Name           string `json:"name,omitempty"`             // 名称
	Owner          string `json:"owner,omitempty"`            // 所有者
	Creation       string `json:"creation,omitempty"`         // 创建时间
	Modified       string `json:"modified,omitempty"`         // 修改时间
	ModifiedBy     string `json:"modified_by,omitempty"`      // 修改者
	Docstatus      int    `json:"docstatus,omitempty"`        // 文档状态
	Idx            int    `json:"idx,omitempty"`              // 索引
	Default        int    `json:"default,omitempty"`          // 是否默认
	AllowInReturns int    `json:"allow_in_returns,omitempty"` // 是否允许退货
	ModeOfPayment  string `json:"mode_of_payment,omitempty"`  // 支付方式
	Parent         string `json:"parent,omitempty"`           // 父级
	Parentfield    string `json:"parentfield,omitempty"`      // 父级字段
	Parenttype     string `json:"parenttype,omitempty"`       // 父级类型
	Doctype        string `json:"doctype,omitempty"`          // 文档类型
}

// ModeOfPayment 结构体定义
// 用于表示支付方式的完整信息
type ModeOfPayment struct {
	Name          string                 `json:"name,omitempty"`            // 支付方式名称
	Owner         string                 `json:"owner,omitempty"`           // 所有者
	Creation      string                 `json:"creation,omitempty"`        // 创建时间
	Modified      string                 `json:"modified,omitempty"`        // 修改时间
	ModifiedBy    string                 `json:"modified_by,omitempty"`     // 修改者
	Docstatus     int                    `json:"docstatus,omitempty"`       // 文档状态
	Idx           int                    `json:"idx,omitempty"`             // 索引
	ModeOfPayment string                 `json:"mode_of_payment,omitempty"` // 支付方式
	Enabled       int                    `json:"enabled,omitempty"`         // 是否启用
	Type          string                 `json:"type,omitempty"`            // 类型
	CustomBranch  string                 `json:"custom_branch,omitempty"`   // 自定义分支
	Doctype       string                 `json:"doctype,omitempty"`         // 文档类型
	Accounts      []ModeOfPaymentAccount `json:"accounts,omitempty"`        // 账户列表
}

// ModeOfPaymentAccount 结构体定义
// 用于表示支付方式中的账户信息
type ModeOfPaymentAccount struct {
	Name           string `json:"name,omitempty"`            // 名称
	Owner          string `json:"owner,omitempty"`           // 所有者
	Creation       string `json:"creation,omitempty"`        // 创建时间
	Modified       string `json:"modified,omitempty"`        // 修改时间
	ModifiedBy     string `json:"modified_by,omitempty"`     // 修改者
	Docstatus      int    `json:"docstatus,omitempty"`       // 文档状态
	Idx            int    `json:"idx,omitempty"`             // 索引
	Company        string `json:"company,omitempty"`         // 公司
	DefaultAccount string `json:"default_account,omitempty"` // 默认账户
	Parent         string `json:"parent,omitempty"`          // 父级
	Parentfield    string `json:"parentfield,omitempty"`     // 父级字段
	Parenttype     string `json:"parenttype,omitempty"`      // 父级类型
	Doctype        string `json:"doctype,omitempty"`         // 文档类型
}
