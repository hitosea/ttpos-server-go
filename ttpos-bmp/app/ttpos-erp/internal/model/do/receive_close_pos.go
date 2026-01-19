// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// ReceiveClosePos is the golang structure of table erp_receive_close_pos for DAO operations like Where/Data.
type ReceiveClosePos struct {
	g.Meta           `orm:"table:erp_receive_close_pos, do:true"`
	Id               any // ID
	PosOpenEntryName any // 开帐名称
	PeriodEndDate    any // 结账时间
	Docstatus        any // 文档状态，参考erpnext
	CreatedAt        any // 创建时间
	UpdatedAt        any // 更新时间
	ReqMessage       any // 请求数据,base64编码
	RespMessage      any // 响应数据,base64编码
	SiteCode         any // erp_site_code, 用来区分调那个租户
	ReqBody          any // 请求文本，如果能转换
	RespBody         any // 响应文本，如果能转换
}
