// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// Site is the golang structure for table site.
type Site struct {
	Id        int64  `json:"id"        orm:"id"         description:"主键"`             // 主键
	Uuid      string `json:"uuid"      orm:"uuid"       description:"UUID"`           // UUID
	SiteName  string `json:"siteName"  orm:"site_name"  description:"站点名称"`           // 站点名称
	SiteUrl   string `json:"siteUrl"   orm:"site_url"   description:"站点地址"`           // 站点地址
	Remark    string `json:"remark"    orm:"remark"     description:"备注"`             // 备注
	SiteCode  string `json:"siteCode"  orm:"site_code"  description:"站点编码， 与ttpos映射"` // 站点编码， 与ttpos映射
	ApiKey    string `json:"apiKey"    orm:"api_key"    description:""`               //
	ApiSecret string `json:"apiSecret" orm:"api_secret" description:""`               //
}
