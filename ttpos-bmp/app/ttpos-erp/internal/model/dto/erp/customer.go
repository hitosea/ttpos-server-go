package erp

// Customer 客户信息数据结构
type Customer struct {
	Name                  string        `json:"name,omitempty"`
	Owner                 string        `json:"owner,omitempty"`
	Creation              string        `json:"creation,omitempty"`
	Modified              string        `json:"modified,omitempty"`
	ModifiedBy            string        `json:"modified_by,omitempty"`
	Docstatus             int           `json:"docstatus,omitempty"`
	Idx                   int           `json:"idx,omitempty"`
	NamingSeries          string        `json:"naming_series,omitempty"`
	CustomerName          string        `json:"customer_name,omitempty"`
	CustomerType          string        `json:"customer_type,omitempty"`
	IsInternalCustomer    int           `json:"is_internal_customer,omitempty"`
	RepresentsCompany     string        `json:"represents_company,omitempty"`
	Language              string        `json:"language,omitempty"`
	DefaultCommissionRate float64       `json:"default_commission_rate,omitempty"`
	SoRequired            int           `json:"so_required,omitempty"`
	DnRequired            int           `json:"dn_required,omitempty"`
	IsFrozen              int           `json:"is_frozen,omitempty"`
	Disabled              int           `json:"disabled,omitempty"`
	Doctype               string        `json:"doctype,omitempty"`
	PortalUsers           []interface{} `json:"portal_users,omitempty"`
	SalesTeam             []interface{} `json:"sales_team,omitempty"`
	Accounts              []interface{} `json:"accounts,omitempty"`
	CreditLimits          []interface{} `json:"credit_limits,omitempty"`
	Companies             []interface{} `json:"companies,omitempty"`
}
