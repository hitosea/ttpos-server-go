package model

import (
	"strconv"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
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
	isUpdate   bool   // 用于判断该model是否需要更新
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

// SetUpdate 设置需要更新
func (model *BaseModel) SetUpdate() {
	model.isUpdate = true
}

// GetUpdate 获取是否需要更新
func (model *BaseModel) GetUpdate() bool {
	return model.isUpdate
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
	return
}

// IsDelete 判断记录是否已删除
func (model *BaseModel) IsDelete() bool {
	return model.DeleteTime != constant.NotDeleted
}

// getGormDbUuid 从数据库名称中提取公司UUID
func (model *BaseModel) getGormDbUuid(tx *gorm.DB) uint64 {

	// 获取数据库名称
	var dbName string
	// 根据不同的数据库类型获取数据库名称
	switch tx.Dialector.Name() {
	case "mysql":
		tx.Raw("SELECT DATABASE()").Scan(&dbName)
	case "sqlite":
		// todo 离线版本要重新考虑
		// 对于 SQLite，数据库名称就是文件路径
		// dbName = tx.Dialector.(gorm.Dialector).DSN()
	}
	// 从数据库名称中提取公司UUID
	// 格式为: shop8609817471094784
	if len(dbName) > 4 && strings.HasPrefix(dbName, "shop") {
		uuidStr := strings.TrimPrefix(dbName, "shop")
		uuid, err := strconv.ParseUint(uuidStr, 10, 64)
		if err == nil {
			return uuid
		}
	}
	return 0
}

// getDeviceIdFromTx 从事务上下文中获取设备ID
func (model *BaseModel) getDeviceIdFromTx(tx *gorm.DB) string {
	// 尝试从事务上下文中获取设备ID
	if value, ok := tx.Get("device_id"); ok {
		if deviceId, ok := value.(string); ok {
			return deviceId
		}
	}
	return "" // 如果没有找到设备ID，返回空字符串
}
