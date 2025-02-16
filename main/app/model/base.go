package model

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"

	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID         uint   `gorm:"column:id;type:int(11) unsigned;AUTO_INCREMENT;primary_key;comment:'自增ID'"`
	Uuid       uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:'绑定记录ID';"`
	CreateTime int64  `gorm:"autoCreateTime;column:create_time;type:int(10);comment:'创建时间(时间戳)'"`
	UpdateTime int64  `gorm:"autoUpdateTime;column:update_time;type:int(10);comment:'更新时间(时间戳)'"`
	DeleteTime int64  `gorm:"column:delete_time;type:int(10);default:0;comment:'删除时间(时间戳)'"`
}

// BeforeCreate 创建记录前生成UUID
func (model *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if model.Uuid == 0 {
		uuid, err := utils.GetID()
		if err != nil {
			logger.Logger.Error("生成雪花ID失败", zap.Error(err))
			return err
		}
		model.Uuid = uuid
	}
	return
}

// IsDelete 判断记录是否已删除
func (model *BaseModel) IsDelete() bool {
	return model.DeleteTime != constant.NotDeleted
}
