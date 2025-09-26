// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ReceivePosInvoice is the golang structure of table erp_receive_pos_invoice for DAO operations like Where/Data.
type ReceivePosInvoice struct {
	g.Meta              `orm:"table:erp_receive_pos_invoice, do:true"`
	Id                  interface{} // ID
	OrderNo             interface{} // 销售订单号，来自ttpos
	OpenPosEntryName    interface{} // POS开帐名称
	PostingDatetime     interface{} // 过账日期时间
	Branch              interface{} // 门店分支，可选
	Docstatus           interface{} // 文档状态,参考 erpnext
	CreatedAt           interface{} // 创建时间
	UpdatedAt           interface{} // 更新时间
	ReqMessage          interface{} // 请求数据,base64编码
	RespMessage         interface{} // 响应数据,base64编码
	ProductsInvoiceName interface{} // 商品销售发票
	MaterialInvoiceName interface{} // 物品销售发票
	SiteCode            interface{} // erp_site_code, 用来区分调那个租户
}
