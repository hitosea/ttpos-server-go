package resp

import "ttpos-server-go/app/dto"

// CompanyInfoResp 门店信息响应
type CompanyInfoResp struct {
	Uuid          uint64 `json:"uuid"`            // 门店UUID
	Name          string `json:"name"`            // 门店名称
	SuperRealName string `json:"super_real_name"` // 超管姓名
	SuperPhone    string `json:"super_phone"`     // 超管手机号
	Status        int    `json:"status"`          // 状态 1-启用 0-禁用

	Roles []RoleItem `json:"roles"` // 角色列表
}

type SaasCompanyListResp struct {
	List []CompanyInfoResp `json:"list"`
}

type CompanyInfoRespWithStoreCode struct {
	Uuid      uint64 `json:"uuid"`       // 门店UUID
	Name      string `json:"name"`       // 门店名称
	StoreCode string `json:"store_code"` // 店铺编码
	Status    int    `json:"status"`     // 状态 1-启用 0-禁用
}

func (c CompanyInfoRespWithStoreCode) GetStoreCode() string { return c.StoreCode }
func (c CompanyInfoRespWithStoreCode) GetUuid() uint64      { return c.Uuid }
type SaasCompanyListWithStoreCodeResp struct {
	List []CompanyInfoRespWithStoreCode `json:"list"`
}

// Store 商城设置
type CompanyStoreResp struct {
	Uuid           uint64             `json:"uuid"`            // 门店UUID
	Status         int                `json:"status"`          // 状态 1-启用 0-禁用
	BusinessStatus int                `json:"business_status"` // 营业状态: 1-测试营业 2-正常营业
	Name           string             `json:"name"`            // 店铺名称
	LogoUrl        string             `json:"logo_url"`        // 店铺logo，上传后保存url，必填
	Address        string             `json:"address"`         // 地址，必填，最大500个字符
	Coordinates    string             `json:"coordinates"`     // 经纬度
	CompanyName    string             `json:"company_name"`    // 公司名称，区别于店铺名称，最大500个字符
	Phone          string             `json:"phone"`           // 联系电话，必填，最大20个字符
	TaxNumber      string             `json:"tax_number"`      // 税号
	LanguageList   []dto.LanguageItem `json:"language_list"`   // 系统语言，必填，至少一个
	TimeZone       string             `json:"time_zone"`       // 时区，必填
	StoreCode      string             `json:"store_code"`      // 店铺编码，用于发票打印，最大100个字符
	Language       []string           `json:"language"`        // 云平台限制商家的可用语言列表
	IPWhiteList    string             `json:"ip_white_list"`   // ip白名单
}
