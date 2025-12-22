package model

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID         uint   `gorm:"column:id;type:int(11) unsigned;primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid       uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:'绑定记录ID';"`
	CreateTime int64  `gorm:"autoCreateTime;column:create_time;type:int(10);comment:'创建时间(时间戳)'"`
	UpdateTime int64  `gorm:"autoUpdateTime;column:update_time;type:int(10);comment:'更新时间(时间戳)'"`
	DeleteTime int64  `gorm:"column:delete_time;type:int(10);default:0;comment:'删除时间(时间戳)'"`
}

// DBTableName 获取表名
func GetTableName(table string) string {
	prefix := config.Database.TablePrefix
	return prefix + table
}

// NoPrimaryKey 判断是否无主键
func (model *BaseModel) NoPrimaryKey() bool {
	if model.ID == 0 {
		return true
	}
	if model.Uuid == 0 {
		return true
	}
	return false
}

// SetDelete 设置需要删除
func (model *BaseModel) SetDelete() {
	model.DeleteTime = time.Now().Unix()
}

// BeforeCreate 创建记录前生成UUID
func (model *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if model.Uuid == 0 {
		uuid, err := utils.GetID()
		if err != nil {
			logger.Logger.Error("生成雪花ID失败", zap.Error(err))
			return errors.WithMessage(err)
		}
		model.Uuid = uuid
	}
	// 如果就是要uuid为0
	// 18446744073709551615 为uint64的最大值
	if model.Uuid == 18446744073709551615 {
		model.Uuid = 0
	}
	return
}
