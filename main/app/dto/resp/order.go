package resp

import "ttpos-server-go/app/dto"

// CreateOrderResp 创建订单响应
type CreateOrderResp struct {
	Uuid uint64 `json:"uuid"` // 订单UUID
}

// BillListsOrder 账单列表订单响应
type BillListsOrder struct {
	SaleBillUuid  uint64         `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64         `json:"sale_order_uuid"` // 销售订单UUID
	BillType      uint           `json:"bill_type"`       // 订单类型	0:桌台订单 1:点餐订单
	SerialNo      string         `json:"serial_no"`       // 桌位编号 (点餐流水号)
	OrderNo       string         `json:"order_no"`        // 订单编号
	Status        uint           `json:"status"`          // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	ConsumerUuids string         `json:"consumer_uuids"`  // 会员id
	FinishTime    int64          `json:"finish_time"`     // 完成时间（支付时间）（时间戳）
	OrderAmount   float64        `json:"order_amount"`    // 订单总金额
	PaymentAmount float64        `json:"payment_amount"`  // 支付金额
	PayTypeName   string         `json:"pay_type_name"`   // 支付类型名称
	Extra         BillListsExtra `json:"extra,omitempty"`
}

// BillListsExtra 账单列表额外信息响应
type BillListsExtra struct { // 通过当前数据控制按钮是否显示
	IsCellRefund        bool `json:"is_cell_refund"`         // 是否可退款
	IsCellCancel        bool `json:"is_cell_cancel"`         // 是否可取消
	IsCellReverseSettle bool `json:"is_cell_reverse_settle"` // 是否可反结账
	IsCellPrint         bool `json:"is_cell_print"`          // 是否可打印小票
	IsCellDelete        bool `json:"is_cell_delete"`         // 是否可删除
	IsCellInvoice       bool `json:"is_cell_invoice"`        // 是否可打印发票
}

// BillLists 订单列表响应
type BillLists struct {
	SaleBillUuid  uint64           `json:"sale_bill_uuid"`  // 销售账单UUID
	SaleOrderUuid uint64           `json:"sale_order_uuid"` // 销售订单UUID,第一个销售订单的uuid
	BillType      uint             `json:"bill_type"`       // 订单类型	0:桌台订单 1:点餐订单
	IsSplit       bool             `json:"is_split"`        // 是否拆单	false:否 true:是
	SerialNo      string           `json:"serial_no"`       // 桌位编号 (点餐流水号)
	OrderNo       string           `json:"order_no"`        // 订单编号
	Status        uint             `json:"status"`          // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	FinishTime    int64            `json:"finish_time"`     // 完成时间（支付时间）（时间戳）
	OrderAmount   float64          `json:"order_amount"`    // 订单总金额
	PaymentAmount float64          `json:"payment_amount"`  // 支付金额
	PayTypeName   string           `json:"pay_type_name"`   // 支付类型名称
	ConsumerUuids string           `json:"consumer_uuids"`  // 会员id
	SaleOrders    []BillListsOrder `json:"sale_orders"`     // 订单列表
	Extra         BillListsExtra   `json:"extra,omitempty"` // 通过当前数据控制按钮是否显示
}

type OrderListMeta struct {
	dto.PageResponse
	TotalNum    int64 `json:"total_num"`    // 总数量
	UnpaidNum   int64 `json:"unpaid_num"`   // 待付款数量
	CompleteNum int64 `json:"complete_num"` // 已完成数量
	CancelNum   int64 `json:"cancel_num"`   // 已取消数量
}

// OrderListPaginationResp 订单列表分页响应
type OrderListPaginationResp struct {
	List []BillLists   `json:"list"` // 订单列表
	Meta OrderListMeta `json:"meta"` // Meta信息
}

type OrderInfoPayTypes struct {
	Uuid            uint64  `json:"sale_bill_uuid"`    // 支付单UUID
	PaymentTypeName string  `json:"payment_type_name"` // 支付类型名称
	CurrencyUnit    string  `json:"currency_unit"`     // 货币单位
	PaymentAmount   float64 `json:"payment_amount"`    // 支付金额
	Code            string  `json:"code"`              // 支付Code
	Status          uint    `json:"status"`            // 支付状态, 0-未支付 1-已支付 2-已退款 3-支付异常
	StatusReason    string  `json:"status_reason"`     // 支付状态原因
	Source          uint    `json:"source"`            // 支付状态, 来源 0-系统 1-手动 2-LianLianPay
	SourceText      string  `json:"source_text"`       // 支付状态, 来源 0-系统 1-手动 2-LianLianPay
}

type OrderProduct struct {
	Uuid                uint64             `json:"uuid"`                  // 销售订单商品ID
	LocaleName          dto.LocaleResponse `json:"locale_name"`           // 产品名称
	LocaleAttributeName dto.LocaleResponse `json:"locale_attribute_name"` // 口味名称
	Price               float64            `json:"price"`                 // 单价. 折前价
	Num                 float64            `json:"num"`                   // 数量
	SalePrice           float64            `json:"sale_price"`            // 销售价 (折前价) (划线价格) 当 sale_price 不等于 total_price 时才显示
	TotalPrice          float64            `json:"total_price"`           // 最终总价(折后价)
	RefundAmount        float64            `json:"refund_amount"`         // 退款金额
	Status              uint               `json:"status"`                // 状态, 0-正常 1-退菜
	Remark              string             `json:"remark"`                // 备注
	IsGift              bool               `json:"is_gift"`               // 是否赠品, false-否 true-是
	IsWrap              bool               `json:"is_wrap"`               // 是否打包, false-否 true-是
	IsBuffet            bool               `json:"is_buffet"`             // 是否自助餐, false-否 true-是
	IsBuffetCustomer    bool               `json:"is_buffet_customer"`    // 是否自助餐顾客, false-否 true-是
	IsDelay             bool               `json:"is_delay"`              // 是否加钟, false-否 true-是
	IsMust              bool               `json:"is_must"`               // 是否必点, false-否 true-是
	GiftReason          string             `json:"gift_reason"`           // 赠品原因
	ImageUrl            string             `json:"image_url"`             // 图片地址
	CancelReason        string             `json:"refund_reason"`         // 退菜原因
}

// OrderInfo 订单信息响应
type OrderInfo struct {
	SaleOrderUuid uint64             `json:"sale_order_uuid"` // 销售订单UUID
	BillType      uint               `json:"bill_type"`       // 订单类型	0:桌台订单 1:点餐订单
	DiningMethod  uint               `json:"dining_method"`   // 用餐方式,0-堂食 1-打包
	SerialNo      string             `json:"serial_no"`       // 桌位编号 (点餐流水号)
	OrderNo       string             `json:"order_no"`        // 订单编号
	Status        uint               `json:"status"`          // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	FinishTime    int64              `json:"finish_time"`     // 完成时间（支付时间）（时间戳）
	OrderAmount   float64            `json:"order_amount"`    // 订单总金额
	PaymentAmount float64            `json:"payment_amount"`  // 支付金额
	RefundAmount  float64            `json:"refund_amount"`   // 退款金额
	PayTypeName   string             `json:"pay_type_name"`   // 支付类型名称
	MemberName    string             `json:"member_name"`     // 会员名称
	MemberUuid    uint64             `json:"member_uuid"`     // 会员名称
	IsFree        bool               `json:"is_free"`         // 是否免单
	FreeReason    dto.LocaleResponse `json:"free_reason"`     // 免单原因
	Products      []OrderProduct     `json:"products"`        // 产品列表
}

// OrderInfos 订单信息响应
type OrderInfos struct {
	SaleBillUuid    uint64              `json:"sale_bill_uuid"`    // 销售账单UUID
	BillType        uint                `json:"bill_type"`         // 订单类型	0:桌台订单 1:点餐订单 2:外送订单
	DiningMethod    uint                `json:"dining_method"`     // 用餐方式,0-堂食(店内就餐) 1-打包
	IsSplit         bool                `json:"is_split"`          // 是否拆单	false:否 true:是
	IsBuffet        bool                `json:"is_buffet"`         // 是否自助餐	false:否 true:是
	SerialNo        string              `json:"serial_no"`         // 桌位编号 (点餐流水号)
	OrderNo         string              `json:"order_no"`          // 订单编号
	Status          uint                `json:"status"`            // 订单状态 订单状态, 0-待付款、1-已完成、2-已取消
	CreateTime      int64               `json:"create_time"`       // 创建时间
	FinishTime      int64               `json:"finish_time"`       // 完成时间（支付时间）（时间戳）
	OrderAmount     float64             `json:"order_amount"`      // 订单总金额
	PaymentAmount   float64             `json:"payment_amount"`    // 支付金额
	RefundAmount    float64             `json:"refund_amount"`     // 退款金额
	MemberNames     string              `json:"member_names"`      // 会员名称
	MemberUuids     string              `json:"member_uuids"`      // 会员名称
	BuffetNames     string              `json:"buffet_names"`      // 自助餐名称
	CancelReason    string              `json:"cancel_reason"`     // 取消原因
	CashierName     string              `json:"cashier_name"`      // 收银员名称
	Remark          string              `json:"remark"`            // 备注
	OrderSourceUuid uint64              `json:"order_source_uuid"` // 订单来源UUID（0=店内）
	OrderSourceName string              `json:"order_source_name"` // 订单来源名称
	NationalityUuid uint64              `json:"nationality_uuid"`  // 国籍UUID（0=未记录）
	NationalityName string              `json:"nationality_name"`  // 国籍名称
	PayTypes        []OrderInfoPayTypes `json:"pay_types"`         // 支付类型
	SaleOrders      []OrderInfo         `json:"sale_orders"`       // 订单列表
}

type OrderOperationLog struct {
	Uuid        uint64                           `json:"uuid"`        // 账单操作记录ID
	RealName    string                           `json:"user_name"`   // 操作人
	Email       string                           `json:"user_email"`  // 操作人邮箱
	Source      string                           `json:"source"`      // 操作来源 cashier-收银 assistant-助手 shop-商家后台
	CreateTime  int64                            `json:"create_time"` // 创建时间
	Description string                           `json:"description"` // 描述
	PayType     []OrderOperationLogPaymentMethod `json:"pay_type"`    // 支付方式列表
	RefundType  int                              `json:"refund_type"` // 退款方式：1-部分退款；2-整单退款
}

type OrderOperationLogPaymentMethod struct {
	Price            string `json:"price"`              // 金额
	Code             int    `json:"value"`              // 支付方式代号
	Name             string `json:"name"`               // 支付方式名称
	Remark           string `json:"remark"`             // 备注
	PaymentOrderUuid uint64 `json:"payment_order_id"`   // 付款单ID
	RefundMoney      string `json:"refund_money"`       // 退款金额
	RefundStatus     int    `json:"refund_status"`      // 退款状态 0-退款失败 1-退款成功
	ReturnOrderUuid  uint64 `json:"return_order_uuid"`  // 退款单ID
	ReturnAmountUuid uint64 `json:"return_amount_uuid"` // 退款金额ID
	BankCode         string `json:"bank_code"`          // 银行代码
	AccountNo        string `json:"account_no"`         // 账号
	AccountName      string `json:"account_name"`       // 账号姓名
	Unit             string `json:"unit"`               // 现金单位
}

type OrderInfosResp struct {
	Detail       OrderInfos `json:"detail"` // 明细
	OperationLog struct {
		List []OrderOperationLog `json:"list"`
	} `json:"operation_log"` // 操作日志
	Extra BillListsExtra `json:"extra"` // 通过当前数据控制按钮是否显示
}

type OrderReturnInfoResp struct {
	CanReturnAmount    float64                    `json:"can_return_amount"`    // 可退款金额. 可退款金额=订单最终应收金额-已退款金额
	ManualReturnPoints bool                       `json:"manual_return_points"` // 是否可以手动退款积分
	DeductiblePoints   float64                    `json:"deductible_points"`    // 可扣除积分
	PaymentRecords     []OrderReturnPaymentRecord `json:"payment_records"`      // 支付记录
	Products           []OrderReturnProduct       `json:"products"`             // 商品列表
}

type OrderReturnPaymentRecord struct {
	PaymentMethodCode int     `json:"-"`                   // 支付类型Code, 用于排序。退款顺序优先退会员、不够退则到现金、再到记录支付（多个时，哪个先后都行）、再到lianlian（多个时，哪个先后都行）
	PaymentOrderUuid  uint64  `json:"payment_order_uuid"`  // 支付单UUID
	PaymentMethodName string  `json:"payment_method_name"` // 支付类型名称
	PaymentMethodUuid uint64  `json:"-"`                   // 支付类型UUID, 用于构建退款金额
	CurrencyUnit      string  `json:"currency_unit"`       // 货币单位
	PaymentAmount     float64 `json:"payment_amount"`      // 支付金额
	CanReturnAmount   float64 `json:"can_return_amount"`   // 可退款金额。剩余可退xx. 可退款金额=支付金额-已退款金额
}

type OrderReturnProduct struct {
	SaleOrderProductUuid uint64             `json:"sale_order_product_uuid"` // 销售订单商品ID
	LocaleName           dto.LocaleResponse `json:"locale_name"`             // 产品名称
	LocaleAttributeName  dto.LocaleResponse `json:"locale_attribute_name"`   // 属性名称
	Num                  float64            `json:"num"`                     // 数量。可退货数量=订单商品数量-已退货数量
	NumType              uint               `json:"num_type"`                // 数量类型 0-整数 1-小数
	Price                float64            `json:"price"`                   // 最终单价(单商品，会员、会员卡和优惠折扣后，折后价)。
	CanReturnAmount      float64            `json:"can_return_amount"`       // 可退款金额. 可退款金额=订单商品数量*最终单价
	CurrencyUnit         string             `json:"currency_unit"`           // 货币单位
}

type OrderReverseSettleInfoResp struct {
	SaleBillUuid    uint64                      `json:"sale_bill_uuid"`              // 销售账单UUID
	SaleBillNo      string                      `json:"sale_bill_no"`                // 销售账单编号
	SaleBillType    uint                        `json:"sale_bill_type"`              // 销售账单类型.0-桌台订单、1-点餐订单
	OrderAmount     float64                     `json:"order_amount"`                // 订单总金额
	PaymentAmount   float64                     `json:"payment_amount"`              // 支付金额
	PayMethods      []string                    `json:"pay_methods"`                 // 支付方式名称列表
	Desks           *OrderReverseSettleDeskList `json:"desks,omitempty"`             // 可用桌台列表。仅桌台订单才有该字段
	HasInstantOrder *bool                       `json:"has_instant_order,omitempty"` // 是否存在即时订单。仅点餐订单才有该字段
}

type OrderReverseSettleDeskList struct {
	OriginDeskAvailable bool                     `json:"origin_desk_available"` // 原桌台是否空闲
	List                []OrderReverseSettleDesk `json:"list"`                  // 桌台编号
}

type OrderReverseSettleDesk struct {
	Uuid     uint64 `json:"uuid"`      // 桌台UUID
	SerialNo string `json:"serial_no"` // 桌台编号
}

// 发票信息
type SaleOrderInvoiceInfo struct {
	CompanyName      string `json:"company_name"`
	CompanyAddr      string `json:"company_addr"`
	CompanyTaxNumber string `json:"company_tax_number"`
	CompanyPhone     string `json:"company_phone"`
}

// OrderBuffetResp 订单自助餐信息
type OrderBuffetResp struct {
	BuffetUuids         []uint64                 `json:"buffet_uuids"`          // 自助餐uuid列表: 非自助餐时, 传空数组; 自助餐时, 元素数量最小为1, 最大为2
	BuffetCustomerTypes []DeskBuffetCustomerType `json:"buffet_customer_types"` // 自助餐顾客类型列表
}

// CheckAuthorizationResp 检查授权响应
type CheckAuthorizationResp struct {
	HasPermission bool `json:"has_permission"` // 是否有权限: true-有权限, false-无权限
}

// VerifyPasswordResp 密码验证响应
type VerifyPasswordResp struct {
	Verified bool `json:"verified"` // 验证结果: true-成功, false-失败
}

// DeskBuffetCustomerType 自助餐顾客类型
type DeskBuffetCustomerType struct {
	Uuid    uint64 `json:"uuid"`     // 自助餐顾客类型uuid
	MealNum uint   `json:"meal_num"` // 就餐人数
}
