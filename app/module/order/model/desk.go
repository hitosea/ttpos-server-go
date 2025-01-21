package model

import (
	"time"

	"gorm.io/gorm"
)

// Desk 桌台信息表
type Desk struct {
	ID             int    `gorm:"primaryKey;column:id;comment:桌台唯一标识符"`
	Name           string `gorm:"column:name;comment:桌台名称"`
	DeskRegionID   int    `gorm:"column:desk_region_id;comment:桌台区域ID"`
	DeskTypeID     int    `gorm:"column:desk_type_id;comment:桌台类型ID"`
	OrderBy        int    `gorm:"column:order_by;comment:排序序号"`
	Status         string `gorm:"column:status;comment:状态"`
	IsDisable      bool   `gorm:"column:is_disable;comment:是否禁用"`
	QrcodeImageURL string `gorm:"column:qrcode_image_url;comment:二维码图片URL"`
	CreateTime     int64  `gorm:"column:create_time;comment:创建时间"`
	UpdateTime     int64  `gorm:"column:update_time;comment:更新时间"`
	DeleteTime     int64  `gorm:"column:delete_time;comment:删除时间"`
}

// BeforeCreate 创建前钩子
func (m *Desk) BeforeCreate(tx *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	m.UpdateTime = time.Now().Unix()
	return nil
}

// BeforeUpdate 更新前钩子
func (m *Desk) BeforeUpdate(tx *gorm.DB) error {
	m.UpdateTime = time.Now().Unix()
	return nil
}
