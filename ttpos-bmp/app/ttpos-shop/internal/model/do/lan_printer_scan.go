// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// LanPrinterScan is the golang structure of table ttpos_lan_printer_scan for DAO operations like Where/Data.
type LanPrinterScan struct {
	g.Meta         `orm:"table:ttpos_lan_printer_scan, do:true"`
	Id             interface{} //
	Uuid           interface{} // uuid
	Ip             interface{} // ip
	Port           interface{} // 端口
	Status         interface{} // 状态 0: 离线 1: 在线
	Remark         interface{} // 备注
	SourceDeviceSn interface{} // 来源设备SN
	CreateTime     interface{} // 创建时间
	UpdateTime     interface{} // 更新时间
	DeleteTime     interface{} // 删除时间
}
