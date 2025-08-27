package model

import (
	"slices"
	"strconv"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BaseModel 基础模型
type BaseModel struct {
	ID            uint   `gorm:"column:id;type:int(11) unsigned;primaryKey;autoIncrement;comment:'自增ID'"`
	Uuid          uint64 `gorm:"column:uuid;type:bigint(20) unsigned;default:0;comment:'绑定记录ID';"`
	CreateTime    int64  `gorm:"autoCreateTime;column:create_time;type:int(10);comment:'创建时间(时间戳)'"`
	UpdateTime    int64  `gorm:"autoUpdateTime;column:update_time;type:int(10);comment:'更新时间(时间戳)'"`
	DeleteTime    int64  `gorm:"column:delete_time;type:int(10);default:0;comment:'删除时间(时间戳)'"`
	isUpdate      bool   // 用于判断该model是否需要更新
	operateSource string // 虚拟字段，用于标记添加来源
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
	// 如果就是要uuid为0
	// 18446744073709551615 为uint64的最大值
	if model.Uuid == 18446744073709551615 {
		model.Uuid = 0
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
	// NOTE sqlite - 离线版本要重新考虑
	return 0
}

// 设置添加来源
func (model *SaleBill) SetOperateSource(addSource string) {
	model.operateSource = addSource
}

// 获取添加来源
func (model *SaleBill) GetOperateSource() string {
	return model.operateSource
}

// --------------------------------
// Hook Methods
// --------------------------------

// SaleBill - AfterUpdate 更新销售订单后的逻辑 - 推送订单更新
func (model *SaleBill) AfterUpdate(tx *gorm.DB) (err error) {
	if model.Uuid == 0 {
		return nil
	}
	if operateSource, ok := tx.Statement.Context.Value(constant.OrderOperateSource).(string); ok {
		if operateSource == constant.SourceH5 {
			return nil
		}
		if operateSource == constant.OrderOpenTable {
			return nil
		}
	}
	if companyUuid := model.getCompanyUuid(tx); companyUuid > 0 {
		if model.DeskUuid > 0 {
			// 推送订单更新
			go websocket.PushClient(companyUuid, websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_ORDER, map[string]interface{}{
				"sale_bill_uuid": model.Uuid,
				"desk_uuid":      model.DeskUuid,
				"update_time":    model.BaseModel.UpdateTime,
			})
			// 推送桌台更新
			go websocket.PushClient(companyUuid, websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_DESK, map[string]interface{}{
				"desk_uuid":   model.DeskUuid,
				"update_time": model.BaseModel.UpdateTime,
			})
		} else if model.MemberSaleOrderUuid == 0 {
			go websocket.PushClient(companyUuid, websocket.SourceCashier, websocket.SourceAll, websocket.UPDATE_ORDER, map[string]interface{}{
				"sale_bill_uuid": model.Uuid,
				"desk_uuid":      model.DeskUuid,
				"update_time":    model.BaseModel.UpdateTime,
			})
		}
	}
	return nil
}

// CustomerCall - AfterCreate 新增客户呼叫后的逻辑 - 推送订单更新
func (model *CustomerCall) AfterCreate(tx *gorm.DB) (err error) {
	if companyUuid := model.getCompanyUuid(tx); companyUuid > 0 {
		data := map[string]interface{}{
			"customer_call_uuid": model.Uuid,
			"desk_uuid":          model.DeskUuid,
			"update_time":        model.BaseModel.UpdateTime,
		}
		go websocket.PushClient(companyUuid, websocket.SourceCashier, websocket.SourceAll, websocket.CUSTOMER_CALL, data)
		go websocket.PushClient(companyUuid, websocket.SourceAssistant, websocket.SourceAll, websocket.CUSTOMER_CALL, data)
		go websocket.PushClient(companyUuid, websocket.SourceKitchen, websocket.SourceAll, websocket.CUSTOMER_CALL, data)
	}
	return nil
}

// PrinterLog - AfterCreate 打印数据 - 推送打印数据
func (model *PrinterLog) AfterCreate(tx *gorm.DB) (err error) {
	if model.Type != 1 || model.Data == "" {
		return
	}
	if companyUuid := model.getCompanyUuid(tx); companyUuid > 0 {
		go websocket.PushClient(
			companyUuid,
			websocket.SourceCashier,
			utils.IfString(model.CashierDeviceId != "", model.CashierDeviceId, websocket.SourceAll),
			websocket.PRINT_DATA,
			map[string]interface{}{
				"print_log_uuid": model.Uuid,
				"update_time":    model.BaseModel.UpdateTime,
			},
		)
	}
	return nil
}

// H5Order - AfterCreate 新增H5订单后的逻辑 - 推送订单更新
func (model *H5Order) AfterCreate(tx *gorm.DB) (err error) {
	if companyUuid := model.getCompanyUuid(tx); companyUuid > 0 {
		data := map[string]interface{}{
			"customer_call_uuid": model.Uuid,
			"h5_order_uuid":      model.Uuid,
			"desk_uuid":          model.DeskUuid,
			"update_time":        model.BaseModel.UpdateTime,
		}
		go websocket.PushClient(companyUuid, websocket.SourceCashier, websocket.SourceAll, websocket.H5_ORDER, data)
		go websocket.PushClient(companyUuid, websocket.SourceCashier, websocket.SourceAll, websocket.CUSTOMER_CALL, data)
	}
	return nil
}

// H5Order - AfterUpdate 更新H5订单后的逻辑 - 推送订单更新
func (model *H5Order) AfterUpdate(tx *gorm.DB) (err error) {
	if companyUuid := model.getCompanyUuid(tx); companyUuid > 0 {
		data := map[string]interface{}{
			"h5_order_uuid": model.Uuid,
			"desk_uuid":     model.DeskUuid,
			"update_time":   model.BaseModel.UpdateTime,
		}
		go websocket.PushClient(companyUuid, websocket.SourceCashier, websocket.SourceAll, websocket.H5_ORDER, data)
		go websocket.PushClient(companyUuid, websocket.SourceCashier, websocket.SourceAll, websocket.CUSTOMER_CALL, data)
	}
	return nil
}

// Desk - AfterUpdate 更新桌台后的逻辑 - 推送桌台更新
func (model *Desk) AfterUpdate(tx *gorm.DB) (err error) {
	if companyUuid := model.getCompanyUuid(tx); companyUuid > 0 {
		data := map[string]interface{}{
			"desk_uuid":   model.BaseModel.Uuid,
			"update_time": model.BaseModel.UpdateTime,
		}
		go websocket.PushClient(companyUuid, websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_DESK, data)
	}
	return nil
}

// MemberSaleOrder - AfterUpdate 更新会员端销售订单后的逻辑 - 推送呼叫消息
func (model *MemberSaleOrder) AfterUpdate(tx *gorm.DB) (err error) {
	companyUuid := model.getCompanyUuid(tx)
	if companyUuid == 0 {
		return nil
	}

	// 如果订单状态大于待商家接单，则推送呼叫消息
	if model.Status > constant.MemberSaleOrderStatusPendingPayment {
		go websocket.PushClient(
			companyUuid,
			websocket.SourceCashier,
			websocket.SourceAll,
			websocket.UPDATE_MEMBER_SALE_ORDER,
			map[string]interface{}{
				"member_sale_order_uuid": model.BaseModel.Uuid,
				"status":                 model.Status,
				"update_time":            model.BaseModel.UpdateTime,
			},
		)
	}

	// 只有以下状态的订单才需要推送websocket消息
	if slices.Contains([]uint{
		constant.MemberSaleOrderStatusPendingMerchantAccept,
		constant.MemberSaleOrderStatusCooking,
		constant.MemberSaleOrderStatusPendingRiderPickup,
		constant.MemberSaleOrderStatusCancelled,
	}, model.Status) {
		go websocket.PushClient(
			companyUuid,
			websocket.SourceCashier,
			websocket.SourceAll,
			websocket.CUSTOMER_CALL,
			map[string]interface{}{
				"status":      model.Status,
				"update_time": model.BaseModel.UpdateTime,
			},
		)
	}

	return nil
}

// ProductPackage - AfterUpdate 更新商品包后的逻辑 - 推送商品包更新
func (model *ProductPackage) AfterUpdate(tx *gorm.DB) (err error) {
	if companyUuid := model.getCompanyUuid(tx); companyUuid > 0 {
		data := map[string]interface{}{
			"product_uuid": model.BaseModel.Uuid,
			"type":         "update",
			"update_time":  model.BaseModel.UpdateTime,
		}
		go websocket.PushClient(companyUuid, websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_PRODUCT, data)
	}
	return nil
}

// ProductCategory - AfterUpdate 更新商品分类后的逻辑 - 推送商品分类更新
func (model *ProductCategory) AfterUpdate(tx *gorm.DB) (err error) {
	if companyUuid := model.getCompanyUuid(tx); companyUuid > 0 {
		data := map[string]interface{}{
			"category_uuid": model.BaseModel.Uuid,
			"type":          "update",
			"update_time":   model.BaseModel.UpdateTime,
		}
		go websocket.PushClient(companyUuid, websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_CATEGORY, data)
	}
	return nil
}
