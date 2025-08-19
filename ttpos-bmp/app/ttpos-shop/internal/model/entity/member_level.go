// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// MemberLevel is the golang structure for table member_level.
type MemberLevel struct {
	Id           uint    `json:"id"           orm:"id"            description:"自增ID"`                                   // 自增ID
	Uuid         uint64  `json:"uuid"         orm:"uuid"          description:"会员等级ID"`                                 // 会员等级ID
	Name         string  `json:"name"         orm:"name"          description:"等级名称"`                                   // 等级名称
	OpenMoney    int     `json:"openMoney"    orm:"open_money"    description:"是否开放累计消费额升级，0-否 1-是"`                    // 是否开放累计消费额升级，0-否 1-是
	UpgradeMoney float64 `json:"upgradeMoney" orm:"upgrade_money" description:"升级条件，累计消费额"`                             // 升级条件，累计消费额
	OpenPoint    int     `json:"openPoint"    orm:"open_point"    description:"是否开放累计积分升级，0-否 1-是"`                     // 是否开放累计积分升级，0-否 1-是
	UpgradePoint float64 `json:"upgradePoint" orm:"upgrade_point" description:"升级条件，累计积分"`                              // 升级条件，累计积分
	Discount     float64 `json:"discount"     orm:"discount"      description:"等级权益,百分比折扣,单位%, 如80%为打8折，discount值为0.8"` // 等级权益,百分比折扣,单位%, 如80%为打8折，discount值为0.8
	Priority     int     `json:"priority"     orm:"priority"      description:"等级权重，越大等级越高"`                            // 等级权重，越大等级越高
	IsDefault    int     `json:"isDefault"    orm:"is_default"    description:"是否默认, 1-是 0-否"`                          // 是否默认, 1-是 0-否
	Remark       string  `json:"remark"       orm:"remark"        description:"备注"`                                     // 备注
	CreateTime   uint    `json:"createTime"   orm:"create_time"   description:"创建时间(时间戳)"`                              // 创建时间(时间戳)
	UpdateTime   uint    `json:"updateTime"   orm:"update_time"   description:"更新时间(时间戳)"`                              // 更新时间(时间戳)
	DeleteTime   uint    `json:"deleteTime"   orm:"delete_time"   description:"删除时间(时间戳)"`                              // 删除时间(时间戳)
}
