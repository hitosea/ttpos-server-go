package model

import (
	"strconv"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

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

// getCompanyUuid 从数据库名称中提取公司UUID
func (model *BaseModel) getCompanyUuid(tx *gorm.DB) uint64 {
	var dbName string
	tx.Raw("SELECT DATABASE()").Scan(&dbName)
	if len(dbName) > 4 && strings.HasPrefix(dbName, "shop") {
		uuidStr := strings.TrimPrefix(dbName, "shop")
		uuid, err := strconv.ParseUint(uuidStr, 10, 64)
		if err == nil {
			return uuid
		}
	}
	// todo sqlite - 离线版本要重新考虑
	return 0
}

// --------------------------------
// Hook Methods
// --------------------------------

// SaleBill - AfterUpdate 更新销售订单后的逻辑 - 推送订单更新
func (model *SaleBill) AfterUpdate(tx *gorm.DB) (err error) {
	if companyUuid := model.getCompanyUuid(tx); companyUuid > 0 {
		go websocket.PushClient(companyUuid, "*", "*", websocket.UPDATE_ORDER, map[string]interface{}{
			"sale_bill_uuid": model.Uuid,
			"desk_uuid":      model.DeskUuid,
			"update_time":    model.BaseModel.UpdateTime,
		})
	}
	return nil
}

// CustomerCall - AfterCreate 新增客户呼叫后的逻辑 - 推送订单更新
func (model *CustomerCall) AfterCreate(tx *gorm.DB) (err error) {
	if companyUuid := model.getCompanyUuid(tx); companyUuid > 0 {
		go websocket.PushClient(companyUuid, websocket.SourceCashier, "*", websocket.CUSTOMER_CALL, map[string]interface{}{
			"customer_call_uuid": model.Uuid,
			"desk_uuid":          model.DeskUuid,
			"update_time":        model.BaseModel.UpdateTime,
		})
	}
	return nil
}

// PrinterLog - AfterCreate 打印数据
func (model *PrinterLog) AfterCreate(tx *gorm.DB) (err error) {
	if model.Type != 1 || model.Data == "" {
		return
	}
	if companyUuid := model.getCompanyUuid(tx); companyUuid > 0 {
		go websocket.PushClient(
			companyUuid,
			websocket.SourceCashier,
			utils.IfString(model.CashierDeviceId != "", model.CashierDeviceId, "*"),
			websocket.PRINT_DATA,
			map[string]interface{}{
				"print_log_uuid": model.Uuid,
				"update_time":    model.BaseModel.UpdateTime,
			},
		)
	}
	return nil
}
