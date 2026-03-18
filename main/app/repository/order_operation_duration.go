package repository

import (
	"fmt"
	"time"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

type IOrderOperationDurationRepo interface {
	Delete7DaysAgo() error
}

func NewOrderOperationDurationRepo(db *gorm.DB) IOrderOperationDurationRepo {
	return &orderOperationDurationRepo{db: db}
}

type orderOperationDurationRepo struct {
	db *gorm.DB
}

// Delete7DaysAgo 分批删除7天前的数据，避免大批量删除导致锁表
// 注意：GORM 的 Delete 方法会忽略 Limit，因此使用原生 SQL 执行分批删除
func (r *orderOperationDurationRepo) Delete7DaysAgo() error {
	deadline := time.Now().Add(-7 * 24 * time.Hour).Unix()
	table := (&model.OrderOperationDuration{}).TableName()
	batchSize := 10000
	for {
		result := r.db.Exec(
			fmt.Sprintf("DELETE FROM `%s` WHERE `create_time` > 0 AND `create_time` < ? LIMIT ?", table),
			deadline, batchSize,
		)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			break
		}
	}
	return nil
}
