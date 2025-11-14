package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IMaterialStockAlertLogRepo 物料库存预警邮件记录仓储接口
type IMaterialStockAlertLogRepo interface {
	// GetAlertLog 获取物料的预警记录
	GetAlertLog(companyUuid, materialUuid, warehouseUuid uint64) (*model.MaterialStockAlertLog, error)
	// CreateAlertLog 创建预警记录
	CreateAlertLog(log *model.MaterialStockAlertLog) error
	// UpdateAlertLog 更新预警记录
	UpdateAlertLog(log *model.MaterialStockAlertLog) error
	// ShouldSendAlert 判断是否需要发送预警（24小时内未发送或首次发送）
	ShouldSendAlert(companyUuid, materialUuid, warehouseUuid uint64) (bool, *model.MaterialStockAlertLog, error)
	// ClearAlertLog 清除预警记录（软删除）
	ClearAlertLog(companyUuid, materialUuid, warehouseUuid uint64) error
}

type materialStockAlertLogRepo struct {
	db *gorm.DB
}

// NewMaterialStockAlertLogRepo 创建物料库存预警邮件记录仓储实例
func NewMaterialStockAlertLogRepo(db *gorm.DB) IMaterialStockAlertLogRepo {
	return &materialStockAlertLogRepo{
		db: db,
	}
}

// GetAlertLog 获取物料的预警记录
func (r *materialStockAlertLogRepo) GetAlertLog(companyUuid, materialUuid, warehouseUuid uint64) (*model.MaterialStockAlertLog, error) {
	var log model.MaterialStockAlertLog
	err := r.db.Where("company_uuid = ? AND material_uuid = ? AND warehouse_uuid = ? AND delete_time = 0",
		companyUuid, materialUuid, warehouseUuid).
		First(&log).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, errors.WithMessage(err, "查询预警记录失败")
	}
	return &log, nil
}

// CreateAlertLog 创建预警记录
func (r *materialStockAlertLogRepo) CreateAlertLog(log *model.MaterialStockAlertLog) error {
	now := time.Now().Unix()
	log.CreateTime = now
	log.UpdateTime = now

	err := r.db.Create(log).Error
	if err != nil {
		return errors.WithMessage(err, "创建预警记录失败")
	}
	return nil
}

// UpdateAlertLog 更新预警记录
func (r *materialStockAlertLogRepo) UpdateAlertLog(log *model.MaterialStockAlertLog) error {
	log.UpdateTime = time.Now().Unix()

	err := r.db.Model(log).Where("uuid = ?", log.Uuid).Updates(map[string]any{
		"current_stock":   log.CurrentStock,
		"safety_stock":    log.SafetyStock,
		"last_alert_time": log.LastAlertTime,
		"alert_count":     log.AlertCount,
		"send_status":     log.SendStatus,
		"recipient":       log.Recipient,
		"error_message":   log.ErrorMessage,
		"update_time":     log.UpdateTime,
	}).Error

	if err != nil {
		return errors.WithMessage(err, "更新预警记录失败")
	}
	return nil
}

// ShouldSendAlert 判断是否需要发送预警
// 规则：
// 1. 如果没有预警记录，返回true（首次发送，第1次）
// 2. 如果已发送2次或以上，返回false（最多只发送2次）
// 3. 如果已发送1次：
//   - 上次发送成功且距离上次发送不足24小时，返回false（等待24小时后发送第2次）
//   - 上次发送成功且距离上次发送超过24小时，返回true（发送第2次）
//   - 上次发送失败，返回true（重试发送）
func (r *materialStockAlertLogRepo) ShouldSendAlert(companyUuid, materialUuid, warehouseUuid uint64) (bool, *model.MaterialStockAlertLog, error) {
	log, err := r.GetAlertLog(companyUuid, materialUuid, warehouseUuid)
	if err != nil {
		return false, nil, err
	}

	// 没有记录，首次发送（第1次）
	if log == nil {
		return true, nil, nil
	}

	// 已经发送过2次或以上，不再发送
	if log.AlertCount >= 2 {
		return false, log, nil
	}

	// 已发送1次的情况
	if log.AlertCount == 1 {
		// 如果上次发送成功
		if log.SendStatus == model.SendStatusSuccess {
			// 检查是否超过24小时
			now := time.Now().Unix()
			if now-int64(log.LastAlertTime) < 10*60 {
				// 24小时内不发送第2次
				return false, log, nil
			}
			// 超过24小时，可以发送第2次
			return true, log, nil
		}
		// 上次发送失败，需要重试
		return true, log, nil
	}

	// 其他情况（alert_count == 0，但有记录），需要发送
	return true, log, nil
}

// ClearAlertLog 清除预警记录（软删除）
// 当库存恢复正常时调用，以便下次库存不足时重新计时
func (r *materialStockAlertLogRepo) ClearAlertLog(companyUuid, materialUuid, warehouseUuid uint64) error {
	now := time.Now().Unix()
	err := r.db.Model(&model.MaterialStockAlertLog{}).
		Where("company_uuid = ? AND material_uuid = ? AND warehouse_uuid = ? AND delete_time = 0",
			companyUuid, materialUuid, warehouseUuid).
		Update("delete_time", now).Error

	if err != nil {
		return errors.WithMessage(err, "清除预警记录失败")
	}
	return nil
}
