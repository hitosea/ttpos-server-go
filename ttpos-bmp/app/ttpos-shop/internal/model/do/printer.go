// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Printer is the golang structure of table ttpos_printer for DAO operations like Where/Data.
type Printer struct {
	g.Meta            `orm:"table:ttpos_printer, do:true"`
	Id                interface{} // 自增ID
	Uuid              interface{} // 打印机ID
	Name              interface{} // 打印机名称
	PrinterTypeUuid   interface{} // 打印机类型ID
	ConfigJson        interface{} // 打印机json配置
	IsUsb             interface{} // 是否usb 0-否 1-是
	IsEnableUsb       interface{} // 是否启用usb 0-否 1-是
	Status            interface{} // 状态 0-离线 1-在线
	LastHeartbeatTime interface{} // 最后心跳时间
	SourceDeviceSn    interface{} // 来源设备SN
	Copies            interface{} // 打印份数
	Sort              interface{} // 排序
	CreateTime        interface{} // 创建时间(时间戳)
	UpdateTime        interface{} // 更新时间(时间戳)
	DeleteTime        interface{} // 删除时间(时间戳)
}
