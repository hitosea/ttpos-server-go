package resp

type CashierLoginResponse struct {
	Token       string        `json:"token"`
	Permissions []*Permission `json:"permissions"`
}

type Permission struct {
	AccessId       int           `json:"access_id"`
	Name           string        `json:"name"`
	Path           string        `json:"path"`
	APIPath        string        `json:"api_path"`
	ParentId       int           `json:"parent_id"`
	Sort           int           `json:"sort"`
	Icon           string        `json:"icon"`
	RedirectName   string        `json:"redirect_name"`
	IsRoute        int           `json:"is_route"`
	IsMenu         int           `json:"is_menu"`
	Alias          string        `json:"alias"`
	IsShow         int           `json:"is_show"`
	PlusCategoryId int           `json:"plus_category_id"`
	Remark         string        `json:"remark"`
	IsSupplier     int           `json:"is_supplier"`
	AppId          int           `json:"app_id"`
	CreateTime     string        `json:"create_time"`
	UpdateTime     string        `json:"update_time"`
	Children       []*Permission `json:"children"`
}
