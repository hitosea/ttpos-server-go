package repository

import (
	"fmt"
	"time"

	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// NumberSequenceRepo 编号序列仓储
type NumberSequenceRepo struct {
	db *gorm.DB
}

// NewNumberSequenceRepo 创建编号序列仓储实例
func NewNumberSequenceRepo(db *gorm.DB) *NumberSequenceRepo {
	return &NumberSequenceRepo{db: db}
}

// GetNextSequence 获取指定类型的下一个序列号
// companyUuid: 商家UUID
// numberType: 编号类型（如 constant.NumberTypeInvoice）
// date: 日期字符串（格式：2006-01-02）
func (r *NumberSequenceRepo) GetNextSequence(companyUuid uint64, numberType string, date string) (uint64, error) {
	currentTime := time.Now().Unix()
	var seq uint64

	err := r.db.Transaction(func(tx *gorm.DB) error {
		var record model.NumberSequence
		err := tx.Where("company_uuid = ? AND type = ? AND date = ?", companyUuid, numberType, date).
			First(&record).Error

		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("查询编号序列失败: %v", err)
		}

		if err == gorm.ErrRecordNotFound {
			// 创建新记录
			record = model.NumberSequence{
				CompanyUuid: companyUuid,
				Type:        numberType,
				Date:        date,
				Sequence:    1,
				CreateTime:  currentTime,
				UpdateTime:  currentTime,
			}
			if err := tx.Create(&record).Error; err != nil {
				return fmt.Errorf("创建编号序列失败: %v", err)
			}
			seq = 1
		} else {
			// 更新序列号
			result := tx.Model(&model.NumberSequence{}).
				Where("company_uuid = ? AND type = ? AND date = ?", companyUuid, numberType, date).
				UpdateColumn("sequence", gorm.Expr("sequence + 1")).
				UpdateColumn("update_time", currentTime)
			if result.Error != nil {
				return fmt.Errorf("更新编号序列失败: %v", result.Error)
			}
			// 重新查询获取最新序列号
			if err := tx.Where("company_uuid = ? AND type = ? AND date = ?", companyUuid, numberType, date).
				First(&record).Error; err != nil {
				return fmt.Errorf("查询更新后的序列失败: %v", err)
			}
			seq = record.Sequence
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return seq, nil
}
