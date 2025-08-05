// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// PrinterLog is the golang structure of table ttpos_printer_log for DAO operations like Where/Data.
type PrinterLog struct {
	g.Meta             `orm:"table:ttpos_printer_log, do:true"`
	Id                 interface{} // 自增ID
	Uuid               interface{} // 打印日志ID
	PrinterUuid        interface{} // 打印机id
	ProductPrinterUuid interface{} // 商品打印机id
	CashierDeviceId    interface{} // 收银机绑定的id
	ReadDeviceId       interface{} // 读取设备id
	RelatedType        interface{} // 关联订单类型：0-销售订单；1-充值订单
	RelatedUuid        interface{} // 销售账单、充值订单id
	Data               interface{} // 打印数据
	Type               interface{} // 类型:0系统默认队列,1云上服务下放
	DataType           interface{} // 数据类型 1-预结账单 2-结账单 3-一菜一单 4-整单打印 5-打印发票 6-打印营业数据 7-打印交班单
	PrintMethod        interface{} // 打印方式 1文本打印, 2图片打印
	PrinterType        interface{} // 打印机类型
	Num                interface{} // 打印次数
	Status             interface{} // 状态(0结束,1进行中,2成功)
	Reason             interface{} // 原因
	PrinterTime        interface{} // 打印时间
	FirstExecution     interface{} // 是否首次执行打印 1-是 0-否
	CreateTime         interface{} // 创建时间(时间戳)
	UpdateTime         interface{} // 更新时间(时间戳)
	DeleteTime         interface{} // 删除时间(时间戳)
}
