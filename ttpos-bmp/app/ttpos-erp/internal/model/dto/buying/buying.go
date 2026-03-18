package buying

type CreatePurchaseFromMqReq struct {
	SourceName      string `json:"source_name,omitempty"`
	Supplier        string `json:"supplier,omitempty"`
	RequiredBy      string `json:"required_by,omitempty"`
	TargetWarehouse string `json:"target_warehouse,omitempty"`
	BuyingPriceList string `json:"buying_price_list,omitempty" description:"采购价格表"`
	TaxesAndCharges string `json:"taxes_and_charges,omitempty" description:"采购税费模板名称"`
	TaxCategory     string `json:"tax_category,omitempty" description:"税类别"`

	// 自定义字段数据（ERPNext make_purchase_order 不会从 MR 携带自定义字段，需在 PO 创建时补写）
	CompanyAbbr         string             `json:"company_abbr,omitempty" description:"门店公司缩写，用于查询门店库存"`
	SourceCompanyAbbr   string             `json:"source_company_abbr,omitempty" description:"总部公司缩写，用于查询总部默认仓库库存"`
	ItemLastPurchaseQty map[string]float64 `json:"item_last_purchase_qty,omitempty" description:"物品上次采购数量(默认销售单位) key:item_code"`
}

type CreateInnerSaleOrderFromPurchaseOrderReq struct {
	SourceName       string `json:"source_name,omitempty"`
	DeliveryDate     string `json:"delivery_date,omitempty"`
	SourceWarehouse  string `json:"source_warehouse,omitempty"`
	SellingPriceList string `json:"selling_price_list,omitempty" description:"销售价格表"`
	TaxesAndCharges  string `json:"taxes_and_charges,omitempty" description:"销售税费模板名称"`
	TaxCategory      string `json:"tax_category,omitempty" description:"税类别"`
	AutoApprove      bool   `json:"auto_approve,omitempty" description:"品牌采购自动审批标记"`
}

type AddSupplerTransactCompanyReq struct {
	Supplier        string `json:"supplier"`
	WithCompanyAbbr string `json:"companyAbbr"`
}

type CreateDeliveryNoteFromInnerSaleOrderReq struct {
	SourceName      string `json:"source_name,omitempty"`
	SourceWarehouse string `json:"source_warehouse,omitempty"`
	TargetWarehouse string `json:"target_warehouse,omitempty"`
}

type CreatePurchaseOrderFromSalesOrderReq struct {
	SourceName      string `json:"source_name,omitempty"`       // 销售订单名称
	Supplier        string `json:"supplier,omitempty"`          // 供应商
	ScheduleDate    string `json:"schedule_date,omitempty"`     // 预计交付日期
	TargetWarehouse string `json:"target_warehouse,omitempty"`  // 目标仓库
	BuyingPriceList string `json:"buying_price_list,omitempty"` // 采购价格表
}
