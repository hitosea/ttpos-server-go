// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ReceiveCancelPosInvoice is the golang structure of table erp_receive_cancel_pos_invoice for DAO operations like Where/Data.
type ReceiveCancelPosInvoice struct {
	g.Meta           `orm:"table:erp_receive_cancel_pos_invoice, do:true"`
	Id               interface{} // ID
	OrderNo          interface{} // 退款订单号，来自ttpos
	OpenPosEntryName interface{} // POS开帐名称
	Docstatus        interface{} // 文档状态,参考 erpnext
	CreatedAt        interface{} // 创建时间
	UpdatedAt        interface{} // 更新时间
	ReqMessage       interface{} // 请求数据,base64编码
	RespMessage      interface{} // 响应数据,base64编码
	SiteCode         interface{} // erp_site_code, 用来区分调那个租户
	ReqBody          interface{} // 请求文本，如果能转换
	RespBody         interface{} // 响应文本，如果能转换
}
