package model

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
)

// DeskRegion 餐桌区域表,定义餐桌的区域信息 ttpos_desk_region
type DeskRegion struct {
	BaseModel
	Name string `gorm:"default:'';column:name;comment:'餐桌区域名称'"`
	Sort uint   `gorm:"default:0;column:sort;comment:'排序序号'"`

	Desks         []Desk         `gorm:"foreignKey:RegionUuid;references:Uuid"`
	DeskMapLayout *DeskMapLayout `gorm:"foreignKey:RegionUuid;references:Uuid"`
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
	DeskNo                 string `gorm:"column:desk_no;type:varchar(255);comment:桌位编号;NOT NULL" json:"desk_no"`
	RegionUuid             uint64 `gorm:"column:region_uuid;type:bigint(20) unsigned;default:0;comment:桌台区域ID;NOT NULL" json:"region_uuid"`
	TypeUuid               uint64 `gorm:"column:type_uuid;type:bigint(20) unsigned;default:0;comment:桌台类型ID;NOT NULL" json:"type_uuid"`
	Sort                   uint   `gorm:"column:sort;type:int(11);default:0;comment:排序序号;NOT NULL" json:"sort"`
	Status                 uint   `gorm:"column:status;type:tinyint(1);default:0;comment:状态, 0-未开台 1-已开台;NOT NULL" json:"status"`
	IsDisable              uint   `gorm:"column:is_disable;type:tinyint(1);default:0;comment:是否禁用, 0-否 1-是;NOT NULL" json:"is_disable"`
	QrcodeToken            string `gorm:"column:qrcode_token;type:varchar(255);comment:二维码图片URL的token,判断二维码链接是否有效,token相同则二维码链接有效;NOT NULL" json:"qrcode_token"`
	SaleBillUuid           uint64 `gorm:"column:sale_bill_uuid;type:bigint(20) unsigned;default:0;comment:销售账单UUID,销售账单ID,一个桌台只能绑定一个销售账单，一个单结束后才能绑定下一个单;NOT NULL" json:"sale_bill_uuid"`
	DeviceUuid             uint64 `gorm:"column:device_uuid;type:bigint(20) unsigned;default:0;comment:平板设备uuid, 0-未绑定;NOT NULL" json:"device_uuid"`
	DefaultPeopleNum       uint   `gorm:"column:default_people_num;type:int(11);default:0;comment:默认人数;NOT NULL" json:"default_people_num"`
	IsOpenDefaultPeopleNum uint   `gorm:"column:is_open_default_people_num;type:int(1);default:0;comment:是否开启默认人数, 0-否 1-是;NOT NULL" json:"is_open_default_people_num"`

	SaleBill *SaleBill   `gorm:"foreignKey:SaleBillUuid;references:uuid"` // 销售账单
	Device   *Device     `gorm:"foreignKey:DeviceUuid;references:uuid"`   // 关联绑定设备
	Region   *DeskRegion `gorm:"foreignKey:RegionUuid;references:uuid"`   // 关联区域
}

func (model *Desk) SetNil() {
	model.SaleBill = nil
	model.Device = nil
	model.Region = nil
}

// CheckProductChangeDesk 检查商品转桌台是否符合条件。这个是目标桌台
func (model *Desk) CheckProductChangeDesk() error {
	// 判断目标桌台是否已经开台。只能转菜到已开台的桌台上
	if model.IsAvailableDesk() {
		return errors.New("目标桌台未开台")
	}
	// 判断目标桌台是否禁用
	if model.IsDisableDesk() {
		return errors.New("目标桌台已禁用")
	}
	// 判断目标桌台是否存在销售账单
	if model.SaleBillUuid == 0 {
		return errors.New("目标桌台不存在销售账单")
	}
	if model.SaleBill != nil {
		// 判断销售账单是否已结账或已经取消
		if model.SaleBill.IsEndStatus() {
			return errors.New("目标桌台已结账")
		}
		// 判断销售账单是否锁定
		if model.SaleBill.IsLockStatus() {
			return errors.New("目标桌台已锁定")
		}
	}
	return nil
}
func (model *Desk) IsDisableDesk() bool {
	return model.IsDisable == constant.DeskDisable
}

// IsAvailableDesk 判断桌台是否是空闲桌台。
// 如果桌台是关闭状态，肯定是空闲桌台
func (model *Desk) IsAvailableDesk() bool {
	if model.Status == constant.DeskStatusClose {
		return true
	}
	return false
}

// SetOpenDesk 设置开台信息
func (model *Desk) SetOpenDesk(saleBillUuid uint64) {
	model.Status = constant.DeskStatusOpen
	model.SaleBillUuid = saleBillUuid
}

// SetCloseDesk 设置关闭桌台信息，将桌台设置为空闲桌台
func (model *Desk) SetCloseDesk() {
	model.Status = constant.DeskStatusClose
	model.SaleBillUuid = 0
}

// SetCloseDesk 设置待清台信息，将桌台改为待清台状态
func (model *Desk) SetWaitClearDesk() {
	model.Status = constant.DeskStatusOpen
	model.SaleBillUuid = 0
}

// GetRecord 只获取该模型的数据，去掉关联的对象
func (model *Desk) GetRecord() Desk {
	record := *model // 复制一份Desk对象
	record.SetNil()  // 将Desk对象中的关联对象只为空
	return record
}

// getCustomerCount 获取桌台人数
func (model *Desk) getCustomerCount() uint {
	if model.SaleBill != nil {
		// 获取销售账单人数
		return model.SaleBill.MealNum
	}
	return 0
}

func (model *Desk) getLockStatus() bool {
	if model.SaleBill != nil {
		return model.SaleBill.IsLockStatus()
	}
	return false
}

func (model *Desk) getIsBuffet() bool {
	if model.SaleBill != nil {
		return model.SaleBill.IsBuffet == constant.SaleBillIsBuffetYes
	}
	return false
}

// GetIsWaitClearStatus 获取桌台是否待清台
func (model *Desk) GetIsWaitClearStatus() bool {
	// 销售账单为空且桌台状态为开台时，表示待清台
	if model.SaleBillUuid == 0 && model.Status == constant.DeskStatusOpen {
		return true
	}
	return false
}

// 如果是自助餐，计算剩余时间; 非自助餐，显示已用时间
func (model *Desk) getTime() int64 {
	// 没有销售账单时返回0
	if model.SaleBill == nil {
		return 0
	}
	// 如果是自助餐，计算剩余时间; 非自助餐，显示已用时间
	if model.getIsBuffet() {
		return model.SaleBill.GetTotalRemainingSeconds()
	} else {
		return time.Now().Unix() - model.SaleBill.CreateTime
	}
}

// 获取销售账单金额
func (model *Desk) getSaleBillAmount() float64 {
	if model.SaleBill != nil {
		return model.SaleBill.Amount
	}
	// 销售账单不存在
	return 0
}

func (model *Desk) getSaleBillRemark() string {
	if model.SaleBill != nil {
		return model.SaleBill.Remark
	}
	// 销售账单不存在
	return ""
}

// GetDeskResp 获取桌台信息
func (model *Desk) GetDeskResp() resp.Desk {

	var saleOrderUuid uint64
	var isSplitOrder bool
	var lockTime int64
	if model.SaleBill != nil && len(model.SaleBill.SaleOrders) > 0 {
		saleOrderUuid = model.SaleBill.SaleOrders[0].Uuid
		isSplitOrder = len(model.SaleBill.SaleOrders) > 1
		if model.SaleBill.IsLockStatus() {
			lockTime = time.Now().Unix() - model.SaleBill.LockTime
		}
	}

	return resp.Desk{
		Uuid:                   model.Uuid,
		DeskNo:                 model.DeskNo,
		CustomerCount:          model.getCustomerCount(),
		Status:                 model.Status,
		IsLock:                 model.getLockStatus(),
		IsBuffet:               model.getIsBuffet(),
		IsWait:                 model.GetIsWaitClearStatus(),
		Time:                   model.getTime(),
		Price:                  model.getSaleBillAmount(),
		Remark:                 model.getSaleBillRemark(),
		TypeUuid:               model.TypeUuid,
		RegionUuid:             model.RegionUuid,
		SaleBillUuid:           model.SaleBillUuid,
		SaleOrderUuid:          saleOrderUuid,
		IsSplitOrder:           isSplitOrder,
		UpdateTime:             time.Now().Unix(),
		LockTime:               lockTime,
		IsDisabled:             model.IsDisable == constant.DeskDisable,
		DefaultPeopleNum:       model.DefaultPeopleNum,
		IsOpenDefaultPeopleNum: model.IsOpenDefaultPeopleNum == constant.Yes,
		BatchTagColor: func() string {
			if model.SaleBill != nil && model.SaleBill.BatchTag != nil {
				return model.SaleBill.BatchTag.Color
			}
			return ""
		}(),
	}
}
