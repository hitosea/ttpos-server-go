// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// ReceiveReturnPosInvoice is the golang structure of table erp_receive_return_pos_invoice for DAO operations like Where/Data.
type ReceiveReturnPosInvoice struct {
	g.Meta           `orm:"table:erp_receive_return_pos_invoice, do:true"`
	Id               any         // ID
	OrderNo          any         // 退款订单号，来自ttpos
	OpenPosEntryName any         // POS开帐名称
	PostingDatetime  *gtime.Time // 过账日期时间
	CompanyAbbr      any         // 公司缩写
	Docstatus        any         // 文档状态,参考 erpnext
	CreatedAt        any         // 创建时间
	UpdatedAt        any         // 更新时间
	ReqMessage       any         // 请求数据,base64编码
	RespMessage      any         // 响应数据,base64编码
	SiteCode         any         // erp_site_code, 用来区分调那个租户
	ReqBody          any         // 请求文本，如果能转换
	RespBody         any         // 响应文本，如果能转换
}
