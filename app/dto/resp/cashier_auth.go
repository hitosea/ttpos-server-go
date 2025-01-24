package resp

type CashierLoginResponse struct {
	Token       string        `json:"token"`
	Permissions []*Permission `json:"permissions"`
}

type Permission struct {
	ID             int           `json:"id"`
	Name           string        `json:"name"`
	Path           string        `json:"path"`
	APIPath        string        `json:"api_path"`
	ParentId       int           `json:"parent_id"`
	Sort           int           `json:"-"`
	Icon           string        `json:"-"`
	RedirectName   string        `json:"redirect_name"`
	IsRoute        int           `json:"is_route"`
	IsMenu         int           `json:"is_menu"`
	Alias          string        `json:"alias"`
	IsShow         int           `json:"is_show"`
	PlusCategoryId int           `json:"-"`
	Remark         string        `json:"-"`
	IsSupplier     int           `json:"-"`
	AppId          int           `json:"-"`
	CreateTime     string        `json:"-"`
	UpdateTime     string        `json:"-"`
	Children       []*Permission `json:"children"`
}
