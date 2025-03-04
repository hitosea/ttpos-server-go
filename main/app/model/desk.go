package model

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
)

// DeskRegion 餐桌区域表,定义餐桌的区域信息 ttpos_desk_region
type DeskRegion struct {
	BaseModel
	Name string `gorm:"default:'';column:name;comment:'餐桌区域名称'"`
	Sort uint   `gorm:"default:0;column:sort;comment:'排序序号'"`

	Desks []Desk `gorm:"foreignKey:RegionUuid;references:Uuid"`
}

// DeskType 餐桌类型表,定义餐桌的类型信息 ttpos_desk_type
type DeskType struct {
	BaseModel
	Name     string `gorm:"default:'';column:name;comment:'餐桌类型名称'"`
	Sort     uint   `gorm:"default:0;column:sort;comment:'排序序号'"`
	RangeMin uint   `gorm:"default:0;column:range_min;comment:'最少人数'"`
	RangeMax uint   `gorm:"default:0;column:range_max;comment:'最多人数'"`
}

// Desk 桌台信息表,定义桌台的相关信息 ttpos_desk
type Desk struct {
	BaseModel
	DeskNo       string `gorm:"column:desk_no;type:varchar(255);comment:桌位编号;NOT NULL" json:"desk_no"`
	RegionUuid   uint64 `gorm:"column:region_uuid;type:bigint(20) unsigned;default:0;comment:桌台区域ID;NOT NULL" json:"region_uuid"`
	TypeUuid     uint64 `gorm:"column:type_uuid;type:bigint(20) unsigned;default:0;comment:桌台类型ID;NOT NULL" json:"type_uuid"`
	Sort         uint   `gorm:"column:sort;type:int(11);default:0;comment:排序序号;NOT NULL" json:"sort"`
	Status       uint   `gorm:"column:status;type:tinyint(1);default:0;comment:状态, 0-未开台 1-已开台;NOT NULL" json:"status"`
	IsDisable    uint   `gorm:"column:is_disable;type:tinyint(1);default:1;comment:是否禁用, 0-否 1-是;NOT NULL" json:"is_disable"`
	QrcodeToken  string `gorm:"column:qrcode_token;type:varchar(255);comment:二维码图片URL的token,判断二维码链接是否有效,token相同则二维码链接有效;NOT NULL" json:"qrcode_token"`
	SaleBillUuid uint64 `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;default:0;comment:销售账单UUID,销售账单ID,一个桌台只能绑定一个销售账单，一个单结束后才能绑定下一个单;NOT NULL" json:"sale_bill_uuid"`
	DeviceUuid   uint64 `gorm:"column:device_uuid;type:bigint(20) unsigned;default:0;comment:平板设备uuid, 0-未绑定;NOT NULL" json:"device_uuid"`

	SaleBill *SaleBill   `gorm:"foreignKey:SaleBillUuid;references:uuid"` // 销售账单
	Device   *Device     `gorm:"foreignKey:DeviceUuid;references:uuid"`   // 关联绑定设备
	Region   *DeskRegion `gorm:"foreignKey:RegionUuid;references:uuid"`   // 关联区域
}

func (model *Desk) SetNil() {
	model.SaleBill = nil
	model.Device = nil
	model.Region = nil
}

// SetOpenDesk 设置开台信息
func (model *Desk) SetOpenDesk(saleBillUuid uint64) {
	model.Status = constant.DeskStatusOpen
	model.SaleBillUuid = saleBillUuid
}

// getCustomerCount 获取桌台人数
func (d *Desk) getCustomerCount() uint {
	if d.SaleBill != nil {
		// 获取销售账单人数
		return d.SaleBill.MealNum
	}
	return 0
}

func (d *Desk) getLockStatus() bool {
	if d.SaleBill != nil {
		return d.SaleBill.IsLock == constant.SaleBillIsLockYes
	}
	return false
}

func (d *Desk) getIsBuffet() bool {
	if d.SaleBill != nil {
		return d.SaleBill.IsBuffet == constant.SaleBillIsBuffetYes
	}
	return false
}

func (d *Desk) getIsWaitStatus() bool {
	// 销售账单为空且桌台状态为开台时，表示待清台
	if d.SaleBill == nil && d.Status == constant.DeskStatusOpen {
		return true
	}
	return false
}

// 如果是自助餐，计算剩余时间; 非自助餐，显示已用时间
func (d *Desk) getTime() int64 {
	// 没有销售账单时返回0
	if d.SaleBill == nil {
		return 0
	}
	// 从开台到现在经过的时间，单位秒
	passedTime := time.Now().Unix() - d.SaleBill.CreateTime
	// 如果是自助餐，计算剩余时间; 非自助餐，显示已用时间
	if d.getIsBuffet() {
		if d.SaleBill.BuffetDuration == 0 {
			return -1
		}
		// 计算剩余时间, 剩余时间=自助餐可用时长-经过时间
		seconds := int64(d.SaleBill.BuffetDuration) - passedTime
		if seconds < 0 {
			return 0
		}
		return seconds
	} else {
		return passedTime
	}
}

// 获取销售账单金额
func (d *Desk) getSaleBillAmount() float64 {
	if d.SaleBill != nil {
		return d.SaleBill.Amount
	}
	// 销售账单不存在
	return 0
}

func (d *Desk) getSaleBillRemark() string {
	if d.SaleBill != nil {
		return d.SaleBill.Remark
	}
	// 销售账单不存在
	return ""
}

// GetDeskResp 获取桌台信息
func (d *Desk) GetDeskResp() resp.Desk {
	return resp.Desk{
		Uuid:          d.Uuid,
		DeskNo:        d.DeskNo,
		CustomerCount: d.getCustomerCount(),
		Status:        d.Status,
		IsLock:        d.getLockStatus(),
		IsBuffet:      d.getIsBuffet(),
		IsWait:        d.getIsWaitStatus(),
		Time:          d.getTime(),
		Price:         d.getSaleBillAmount(),
		Remark:        d.getSaleBillRemark(),
		TypeUuid:      d.TypeUuid,
		RegionUuid:    d.RegionUuid,
		SaleBillUuid:  d.SaleBillUuid,
	}
}
