// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// PrinterLog is the golang structure for table printer_log.
type PrinterLog struct {
	Id                 uint   `json:"id"                 orm:"id"                   description:"自增ID"`                                                    // 自增ID
	Uuid               uint64 `json:"uuid"               orm:"uuid"                 description:"打印日志ID"`                                                  // 打印日志ID
	PrinterUuid        uint64 `json:"printerUuid"        orm:"printer_uuid"         description:"打印机id"`                                                   // 打印机id
	ProductPrinterUuid uint64 `json:"productPrinterUuid" orm:"product_printer_uuid" description:"商品打印机id"`                                                 // 商品打印机id
	CashierDeviceId    string `json:"cashierDeviceId"    orm:"cashier_device_id"    description:"收银机绑定的id"`                                                // 收银机绑定的id
	ReadDeviceId       string `json:"readDeviceId"       orm:"read_device_id"       description:"读取设备id"`                                                  // 读取设备id
	RelatedType        int    `json:"relatedType"        orm:"related_type"         description:"关联订单类型：0-销售订单；1-充值订单"`                                    // 关联订单类型：0-销售订单；1-充值订单
	RelatedUuid        uint64 `json:"relatedUuid"        orm:"related_uuid"         description:"销售账单、充值订单id"`                                             // 销售账单、充值订单id
	Data               string `json:"data"               orm:"data"                 description:"打印数据"`                                                    // 打印数据
	Type               int    `json:"type"               orm:"type"                 description:"类型:0系统默认队列,1云上服务下放"`                                      // 类型:0系统默认队列,1云上服务下放
	DataType           int    `json:"dataType"           orm:"data_type"            description:"数据类型 1-预结账单 2-结账单 3-一菜一单 4-整单打印 5-打印发票 6-打印营业数据 7-打印交班单"` // 数据类型 1-预结账单 2-结账单 3-一菜一单 4-整单打印 5-打印发票 6-打印营业数据 7-打印交班单
	PrintMethod        int    `json:"printMethod"        orm:"print_method"         description:"打印方式 1文本打印, 2图片打印"`                                       // 打印方式 1文本打印, 2图片打印
	PrinterType        string `json:"printerType"        orm:"printer_type"         description:"打印机类型"`                                                   // 打印机类型
	Num                int    `json:"num"                orm:"num"                  description:"打印次数"`                                                    // 打印次数
	Status             int    `json:"status"             orm:"status"               description:"状态(0结束,1进行中,2成功)"`                                        // 状态(0结束,1进行中,2成功)
	Reason             string `json:"reason"             orm:"reason"               description:"原因"`                                                      // 原因
	PrinterTime        int    `json:"printerTime"        orm:"printer_time"         description:"打印时间"`                                                    // 打印时间
	FirstExecution     int    `json:"firstExecution"     orm:"first_execution"      description:"是否首次执行打印 1-是 0-否"`                                        // 是否首次执行打印 1-是 0-否
	CreateTime         uint   `json:"createTime"         orm:"create_time"          description:"创建时间(时间戳)"`                                               // 创建时间(时间戳)
	UpdateTime         uint   `json:"updateTime"         orm:"update_time"          description:"更新时间(时间戳)"`                                               // 更新时间(时间戳)
	DeleteTime         uint   `json:"deleteTime"         orm:"delete_time"          description:"删除时间(时间戳)"`                                               // 删除时间(时间戳)
}
