package ai_agent

// ProcurementState holds the entire workflow state for a procurement analysis run.
type ProcurementState struct {
	// Input
	WarehouseUuid uint64 `json:"warehouse_uuid"`
	ForecastDays  int    `json:"forecast_days"`

	// Collected data
	Materials    []MaterialInfo    `json:"materials"`
	Suppliers    []SupplierInfo    `json:"suppliers"`
	SalesData    []ProductSaleInfo `json:"sales_data"`
	LimitSchemes []LimitSchemeInfo `json:"limit_schemes"`

	// Forecast results
	Forecasts     []ForecastItem `json:"forecasts"`
	NeedsPurchase bool           `json:"needs_purchase"`

	// Proposals (grouped by supplier)
	Proposals []PurchaseProposal `json:"proposals"`

	// Review
	ReviewDecision string `json:"review_decision"` // "approved" | "rejected" | ""
	ReviewComment  string `json:"review_comment"`

	// Created orders
	CreatedOrders []CreatedOrder `json:"created_orders"`

	// Anomalies
	Anomalies []AnomalyInfo `json:"anomalies"`

	// Execution tracking
	StepLog []string `json:"step_log"`
	Error   string   `json:"error"`
	Status  string   `json:"status"` // "running" | "awaiting_review" | "completed" | "failed"
}

// MaterialInfo represents enriched material data (warehouse stock + safety_stock merged).
type MaterialInfo struct {
	MaterialUuid     uint64  `json:"material_uuid"`
	MaterialCode     string  `json:"material_code"`
	MaterialNameZH   string  `json:"material_name_zh"`
	BookedQuantity   float64 `json:"booked_quantity"`   // from warehouse
	SafetyStock      float64 `json:"safety_stock"`      // from material list
	SupplierUuid     uint64  `json:"supplier_uuid"`     // from material model
	SupplierErpCode  string  `json:"supplier_erp_code"` // from material model
	PurchaseUnitUuid uint64  `json:"purchase_unit_uuid"`
	AvailableNum     float64 `json:"available_num"`
	TransitNum       float64 `json:"transit_num"`
}

// SupplierInfo represents a supplier.
type SupplierInfo struct {
	Uuid    uint64 `json:"uuid"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	ErpCode string `json:"erp_code"`
}

// ProductSaleInfo represents aggregated product sales.
type ProductSaleInfo struct {
	ProductName  string  `json:"product_name"`
	CategoryName string  `json:"category_name"`
	SalesNum     float64 `json:"sales_num"`
	SalesAmount  float64 `json:"sales_amount"`
}

// LimitSchemeInfo represents a purchase limit scheme.
type LimitSchemeInfo struct {
	Uuid   uint64 `json:"uuid"`
	Name   string `json:"name"`
	Status int    `json:"status"`
}

// ForecastItem represents a single material demand forecast.
type ForecastItem struct {
	MaterialUuid    uint64  `json:"material_uuid"`
	MaterialCode    string  `json:"material_code"`
	MaterialNameZH  string  `json:"material_name_zh"`
	CurrentStock    float64 `json:"current_stock"`
	SafetyStock     float64 `json:"safety_stock"`
	PredictedDemand float64 `json:"predicted_demand"`
	Shortage        float64 `json:"shortage"`
	OrderQuantity   float64 `json:"order_quantity"`
	SupplierName    string  `json:"supplier_name"`
	SupplierErpCode string  `json:"supplier_erp_code"`
}

// PurchaseProposal represents a grouped purchase proposal for one supplier.
type PurchaseProposal struct {
	SupplierName    string         `json:"supplier_name"`
	SupplierErpCode string         `json:"supplier_erp_code"`
	Items           []ProposalItem `json:"items"`
	TotalQuantity   float64        `json:"total_quantity"`
}

// ProposalItem represents a single item in a purchase proposal.
type ProposalItem struct {
	MaterialUuid   uint64  `json:"material_uuid"`
	MaterialCode   string  `json:"material_code"`
	MaterialNameZH string  `json:"material_name_zh"`
	OrderQuantity  float64 `json:"order_quantity"`
	UnitUuid       uint64  `json:"unit_uuid"`
}

// CreatedOrder represents a successfully created purchase order.
type CreatedOrder struct {
	Uuid    uint64 `json:"uuid"`
	OrderNo string `json:"order_no"`
}

// AnomalyInfo represents a detected stock anomaly.
type AnomalyInfo struct {
	MaterialUuid   uint64  `json:"material_uuid"`
	MaterialCode   string  `json:"material_code"`
	MaterialNameZH string  `json:"material_name_zh"`
	AnomalyType    string  `json:"anomaly_type"` // "zero_stock" | "below_safety_stock"
	Severity       string  `json:"severity"`     // "high" | "medium" | "low"
	CurrentStock   float64 `json:"current_stock"`
	SafetyStock    float64 `json:"safety_stock"`
	Message        string  `json:"message"`
}
