package setup

type CreateWarehouseInp struct {
	Branch      string `json:"branch,omitempty"`
	WhType      string `json:"whType,omitempty"`
	AliasName   string `json:"aliasname,omitempty"`
	Company     string `json:"company,omitempty"`
	CompanyAbbr string `json:"company_abbr,omitempty"`
}

type CreateWarehouseOut struct {
	WarehouseName string `json:"warehouse_name,omitempty"`
}

type CreatePosProfileInp struct {
	PosProfileName string `json:"pos_profile_name,omitempty"`
	CompanyAbbr    string `json:"company_abbr,omitempty"`
	Warehouse      string `json:"warehouse,omitempty"`
	Branch         string `json:"branch,omitempty"`

	//Payment Methods
	Payments []string `json:"payments,omitempty"`

	//Accounting
	Currency string `json:"currency,omitempty"`
	//销帐科目
	WriteOffAccount string `json:"write_off_account,omitempty"`
	//冲销限额
	WriteOffLimit float64 `json:"write_off_limit,omitempty"`
	//销帐成本中心
	WriteOffCostCenter string `json:"write_off_cost_center,omitempty"`
}

type CreateUserInp struct {
	UserEmail string `json:"user_email,omitempty"`
	FirstName string `json:"first_name,omitempty"`
}

type CreateModePaymentAccountInp struct {
	CompanyAbbr string `json:"company_abbr,omitempty"`
	PaymentType string `json:"payment_type,omitempty"`
}
