package erp

type CreateWarehouseInp struct {
	Branch      string `json:"branch"`
	WhType      string `json:"whType"`
	AliasName   string `json:"aliasname"`
	Company     string `json:"company"`
	CompanyAbbr string `json:"company_abbr"`
}

type CreateWarehouseOut struct {
	WarehouseName string `json:"warehouse_name"`
}

type CreatePosProfileInp struct {
	PosProfileName string `json:"pos_profile_name"`
	CompanyAbbr    string `json:"company_abbr"`
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

type CreateUserInp struct {
	UserEmail string `json:"user_email"`
	FirstName string `json:"first_name"`
}
