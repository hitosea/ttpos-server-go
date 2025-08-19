// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Desk is the golang structure of table ttpos_desk for DAO operations like Where/Data.
type Desk struct {
	g.Meta         `orm:"table:ttpos_desk, do:true"`
	Id             interface{} // 自增ID
	Uuid           interface{} // 桌台ID
	DeskNo         interface{} // 桌位编号
	RegionUuid     interface{} // 桌台区域ID
	TypeUuid       interface{} // 桌台类型ID
	Sort           interface{} // 排序序号
	Status         interface{} // 状态, 0-未开台 1-已开台
	IsDisable      interface{} // 是否禁用, 0-否 1-是
	NeedServiceFee interface{} // 是否需要服务费, 0-否 1-是.标记该桌台收取服务费
	QrcodeToken    interface{} // 二维码图片URL的token,判断二维码链接是否有效,token相同则二维码链接有效
	SaleBillUuid   interface{} // 销售账单UUID,销售账单ID,一个桌台只能绑定一个销售账单，一个单结束后才能绑定下一个单
	DeviceUuid     interface{} // 平板设备uuid, 0-未绑定
	CreateTime     interface{} // 创建时间(时间戳)
	UpdateTime     interface{} // 更新时间(时间戳)
	DeleteTime     interface{} // 删除时间(时间戳)
}
