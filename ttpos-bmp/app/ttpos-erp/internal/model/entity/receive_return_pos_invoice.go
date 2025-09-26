// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// ReceiveReturnPosInvoice is the golang structure for table receive_return_pos_invoice.
type ReceiveReturnPosInvoice struct {
	Id               int64       `json:"id"               orm:"id"                  description:"ID"`              // ID
	OrderNo          string      `json:"orderNo"          orm:"order_no"            description:"退款订单号，来自ttpos"`   // 退款订单号，来自ttpos
	OpenPosEntryName string      `json:"openPosEntryName" orm:"open_pos_entry_name" description:"POS开帐名称"`         // POS开帐名称
	PostingDatetime  *gtime.Time `json:"postingDatetime"  orm:"posting_datetime"    description:"过账日期时间"`          // 过账日期时间
	CompanyAbbr      string      `json:"companyAbbr"      orm:"company_abbr"        description:"公司缩写"`            // 公司缩写
	Docstatus        string      `json:"docstatus"        orm:"docstatus"           description:"文档状态,参考 erpnext"` // 文档状态,参考 erpnext
	CreatedAt        int         `json:"createdAt"        orm:"created_at"          description:"创建时间"`            // 创建时间
	UpdatedAt        int         `json:"updatedAt"        orm:"updated_at"          description:"更新时间"`            // 更新时间
	ReqMessage       string      `json:"reqMessage"       orm:"req_message"         description:"请求数据,base64编码"`   // 请求数据,base64编码
	RespMessage      string      `json:"respMessage"      orm:"resp_message"        description:"响应数据,base64编码"`   // 响应数据,base64编码
}
