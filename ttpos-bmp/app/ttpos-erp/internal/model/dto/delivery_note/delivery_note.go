package delivery_note

// CreateDeliveryNoteReq 创建送货单请求
type CreateDeliveryNoteReq struct {
	CompanyAbbr      string                  `json:"company_abbr" v:"required#公司缩写不能为空"` // 公司缩写，必填
	Branch           string                  `json:"branch"`                             // 分支名称，可选
	Customer         string                  `json:"customer" v:"required#客户不能为空"`       // 客户，必填
	CustomerName     string                  `json:"customer_name"`                      // 客户名称，可选
	PostingDate      string                  `json:"posting_date" v:"required#过账日期不能为空"` // 过账日期，必填
	PostingTime      string                  `json:"posting_time"`                       // 过账时间，可选
	SetWarehouse     string                  `json:"set_warehouse"`                      // 设置仓库，可选
	SellingPriceList string                  `json:"selling_price_list"`                 // 销售价格表，可选
	Items            []DeliveryNoteItemInput `json:"items" v:"required#送货项不能为空"`         // 送货项，必填
}

// DeliveryNoteItemInput 送货单明细项输入
type DeliveryNoteItemInput struct {
	ItemCode          string  `json:"item_code" v:"required#物品编码不能为空"`      // 物品编码，必填
	ItemName          string  `json:"item_name"`                            // 物品名称，可选
	Qty               float64 `json:"qty" v:"required|gt:0#数量不能为空|数量必须大于0"` // 数量，必填
	Uom               string  `json:"uom"`                                  // 单位，可选
	Rate              float64 `json:"rate"`                                 // 单价，可选
	Warehouse         string  `json:"warehouse"`                            // 仓库，可选
	AgainstSalesOrder string  `json:"against_sales_order"`                  // 关联销售订单，可选
	SoDetail          string  `json:"so_detail"`                            // 销售订单明细，可选
}

// CreateDeliveryNoteResp 创建送货单响应
type CreateDeliveryNoteResp struct {
	DeliveryNoteName string `json:"delivery_note_name"` // 送货单号
}

// GetDeliveryNoteReq 获取送货单详情请求
type GetDeliveryNoteReq struct {
	DeliveryNoteName string `json:"delivery_note_name" v:"required#送货单号不能为空"` // 送货单号，必填
}

// DeliveryNoteData 送货单数据
type DeliveryNoteData struct {
	Name             string                  `json:"name"`               // 送货单号
	Company          string                  `json:"company"`            // 公司
	Customer         string                  `json:"customer"`           // 客户
	CustomerName     string                  `json:"customer_name"`      // 客户名称
	PostingDate      string                  `json:"posting_date"`       // 过账日期
	PostingTime      string                  `json:"posting_time"`       // 过账时间
	SetWarehouse     string                  `json:"set_warehouse"`      // 设置仓库
	SellingPriceList string                  `json:"selling_price_list"` // 销售价格表
	TotalQty         float64                 `json:"total_qty"`          // 总数量
	GrandTotal       float64                 `json:"grand_total"`        // 总金额
	Status           string                  `json:"status"`             // 状态
	Docstatus        int                     `json:"docstatus"`          // 单据状态
	Items            []*DeliveryNoteItemData `json:"items,omitempty"`    // 送货项
}

// DeliveryNoteItemData 送货单明细项数据
type DeliveryNoteItemData struct {
	ItemCode          string  `json:"item_code"`           // 物品编码
	ItemName          string  `json:"item_name"`           // 物品名称
	Qty               float64 `json:"qty"`                 // 数量
	Uom               string  `json:"uom"`                 // 单位
	Rate              float64 `json:"rate"`                // 单价
	Amount            float64 `json:"amount"`              // 金额
	Warehouse         string  `json:"warehouse"`           // 仓库
	AgainstSalesOrder string  `json:"against_sales_order"` // 关联销售订单
}

// GetDeliveryNoteResp 获取送货单详情响应
type GetDeliveryNoteResp struct {
	DeliveryNote *DeliveryNoteData `json:"delivery_note"` // 送货单信息
}

// GetDeliveryNoteListReq 获取送货单列表请求
type GetDeliveryNoteListReq struct {
	CompanyAbbr  string `json:"company_abbr" v:"required#公司缩写不能为空"` // 公司缩写，必填
	Branch       string `json:"branch"`                             // 分支名称，可选
	Customer     string `json:"customer"`                           // 客户，可选
	Warehouse    string `json:"warehouse"`                          // 仓库，可选
	Status       string `json:"status"`                             // 状态，可选
	FromDate     string `json:"from_date"`                          // 开始日期，可选
	ToDate       string `json:"to_date"`                            // 结束日期，可选
	Limit        int32  `json:"limit"`                              // 查询限制数量，可选
	IncludeItems bool   `json:"include_items"`                      // 是否包含明细项，可选
	PoNo         string `json:"po_no"`                              // 采购订单号，可选（通过关联销售订单反查）
	SoNo         string `json:"so_no"`                              // 销售订单号，可选（通过 Delivery Note Item.against_sales_order 过滤）
}

// GetDeliveryNoteListResp 获取送货单列表响应
type GetDeliveryNoteListResp struct {
	DeliveryNoteList []*DeliveryNoteData `json:"delivery_note_list"` // 送货单列表
}

// UpdateDeliveryNoteReq 更新送货单请求
type UpdateDeliveryNoteReq struct {
	DeliveryNoteName string                  `json:"delivery_note_name" v:"required#送货单号不能为空"` // 送货单号，必填
	Customer         string                  `json:"customer"`                                 // 客户，可选
	CustomerName     string                  `json:"customer_name"`                            // 客户名称，可选
	PostingDate      string                  `json:"posting_date"`                             // 过账日期，可选
	SetWarehouse     string                  `json:"set_warehouse"`                            // 设置仓库，可选
	Items            []DeliveryNoteItemInput `json:"items"`                                    // 送货项，可选
}

// UpdateDeliveryNoteResp 更新送货单响应
type UpdateDeliveryNoteResp struct {
	Message string `json:"message"` // 操作结果消息
}

// CreateDeliveryNoteFromSaleOrderReq 从销售订单创建送货单请求
type CreateDeliveryNoteFromSaleOrderReq struct {
	SourceName       string `json:"source_name" v:"required#销售订单名称不能为空"` // 销售订单名称，必填
	SourceWarehouse  string `json:"source_warehouse,omitempty"`          // 源仓库，可选
	TargetWarehouse  string `json:"target_warehouse,omitempty"`          // 目标仓库，可选
	SellingPriceList string `json:"selling_price_list,omitempty"`        // 销售价格表，可选
	PostingDate      string `json:"posting_date,omitempty"`              // 过账日期，可选
}

// CreateDeliveryNoteFromSaleOrderResp 从销售订单创建送货单响应
type CreateDeliveryNoteFromSaleOrderResp struct {
	Name       string  `json:"name"`        // 送货单号
	Status     string  `json:"status"`      // 送货单状态
	GrandTotal float64 `json:"grand_total"` // 总金额
}
