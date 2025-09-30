// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// ReceiveClosePos is the golang structure for table receive_close_pos.
type ReceiveClosePos struct {
	Id               int64  `json:"id"               orm:"id"                  description:"ID"`                       // ID
	PosOpenEntryName string `json:"posOpenEntryName" orm:"pos_open_entry_name" description:"开帐名称"`                     // 开帐名称
	PeriodEndDate    int64  `json:"periodEndDate"    orm:"period_end_date"     description:"结账时间"`                     // 结账时间
	Docstatus        string `json:"docstatus"        orm:"docstatus"           description:"文档状态，参考erpnext"`           // 文档状态，参考erpnext
	CreatedAt        int    `json:"createdAt"        orm:"created_at"          description:"创建时间"`                     // 创建时间
	UpdatedAt        int    `json:"updatedAt"        orm:"updated_at"          description:"更新时间"`                     // 更新时间
	ReqMessage       string `json:"reqMessage"       orm:"req_message"         description:"请求数据,base64编码"`            // 请求数据,base64编码
	RespMessage      string `json:"respMessage"      orm:"resp_message"        description:"响应数据,base64编码"`            // 响应数据,base64编码
	SiteCode         string `json:"siteCode"         orm:"site_code"           description:"erp_site_code, 用来区分调那个租户"` // erp_site_code, 用来区分调那个租户
	ReqBody          string `json:"reqBody"          orm:"req_body"            description:"请求文本，如果能转换"`               // 请求文本，如果能转换
	RespBody         string `json:"respBody"         orm:"resp_body"           description:"响应文本，如果能转换"`               // 响应文本，如果能转换
}
