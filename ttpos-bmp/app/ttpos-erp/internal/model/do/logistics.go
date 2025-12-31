// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Logistics is the golang structure of table erp_logistics for DAO operations like Where/Data.
type Logistics struct {
	g.Meta       `orm:"table:erp_logistics, do:true"`
	Id           any // ID
	Uuid         any // UUID
	Vendor       any // 供应商，如 JT:极兔
	VendorUserId any // 供应商用户id,如极兔的货主编码
	InfConf      any // 接口连接信息。如ak/sk 根据不同供应商有所不同
	Remarks      any // 备注信息
	Reserve1     any // 保留字段1
	Reserve2     any // 保留字段2
}
