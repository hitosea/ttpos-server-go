package erp

type CreateWarehouseInp struct {
	Branch      string `json:"branch"`
	WhType      string `json:"whType"`
	AliasName   string `json:"aliasName"`
	Company     string `json:"company"`
	CompanyAbbr string `json:"companyAbbr"`
}

type CreateWarehouseOut struct {
	WarehouseName string `json:"warehouseName"`
}

type CreatePosProfileInp struct {
	PosProfileName string `json:"posProfileName"`
	Company        string `json:"company"`
	Warehouse      string `json:"warehouse"`
	Branch         string `json:"branch"`

	//Payment Methods
	Payments []string `json:"payments"`

	//Accounting
	Currency string `json:"currency"`
	//销帐科目
	WriteOffAccount string `json:"write_off_account"`
	//冲销限额
	WriteOffLimit string `json:"write_off_limit"`
	//销帐成本中心
	WriteOffCostCenter string `json:"write_off_cost_center"`
}
