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
	Id               interface{} // ID
	OrderNo          interface{} // 退款订单号，来自ttpos
	OpenPosEntryName interface{} // POS开帐名称
	PostingDatetime  *gtime.Time // 过账日期时间
	CompanyAbbr      interface{} // 公司缩写
	Docstatus        interface{} // 文档状态,参考 erpnext
	CreatedAt        interface{} // 创建时间
	UpdatedAt        interface{} // 更新时间
	ReqMessage       interface{} // 请求数据,base64编码
	RespMessage      interface{} // 响应数据,base64编码
}
