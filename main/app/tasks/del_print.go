package tasks

import (
	"fmt"
	"log"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/config"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"
)

// 秒级任务
type delPrintTask struct {
	dbm   *database.DBManager
	cache cache.Cache
}

// NewUsbPrintTask 创建新的秒级任务
func NewDelPrintTask(dbm *database.DBManager, cache cache.Cache) *delPrintTask {
	return &delPrintTask{
		dbm:   dbm,
		cache: cache,
	}
}

// Execute 执行任务
func (t *delPrintTask) Execute() {
	// 根据 APP 表实例化数据库连接，并左关联 CompanySetting
	prefix := config.Database.TablePrefix
	var companies []model.Company
	if err := t.dbm.GetDB(0).
		Joins(fmt.Sprintf("LEFT JOIN %scompany_setting ON %scompany.uuid = %scompany_setting.company_uuid", prefix, prefix, prefix)).
		Where(fmt.Sprintf("%scompany.delete_time = 0", prefix)).
		Select(fmt.Sprintf("%scompany.uuid, %scompany.name", prefix, prefix)). // 选择需要的字段
		Find(&companies).Error; err != nil {
		log.Fatalf("Error querying companies with settings: %s", err)
		return
	}
	//
	for _, company := range companies {
		repository.NewPrinterLogRepo(t.dbm.GetDB(company.Uuid)).Delete7DaysAgo()
	}
}
