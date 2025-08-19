// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Device is the golang structure of table ttpos_device for DAO operations like Where/Data.
type Device struct {
	g.Meta             `orm:"table:ttpos_device, do:true"`
	Id                 interface{} // 自增ID
	Uuid               interface{} // 绑定记录ID
	FinallyLoginUuid   interface{} // 最后一个登录id, 退出会清为0
	FinallyLoginTime   interface{} // 最后登录时间
	Source             interface{} // 来源 cashier-收银机 tablet-平板端 kitchen-厨显端
	DeviceId           interface{} // 唯一设备标识id
	IsMain             interface{} // 是否主设备 0-常规 1-主
	ProductPrinterUuid interface{} // 打印档口Uuid
	Address            interface{} // 绑定地址
	Port               interface{} // 绑定端口
	DeviceIp           interface{} // 设备ip
	Remark             interface{} // 备注
	Brand              interface{} // 品牌名称
	Platform           interface{} // 平台,0-Web-网页, 1-Android-安卓, 2-iPhone-苹果, 3-Mobile-移动端
	UserAgent          interface{} // 请求头信息
	CashSign           interface{} // 收银终端标识
	CashBoxId          interface{} // 现金箱ID
	AccessToken        interface{} // 访问令牌
	QueueUrl           interface{} // 关联队列url
	CreateTime         interface{} // 创建时间(时间戳)
	UpdateTime         interface{} // 更新时间(时间戳)
	DeleteTime         interface{} // 删除时间(时间戳)
}
