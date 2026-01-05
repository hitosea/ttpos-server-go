package repository

import (
	"time"

	"ttpos-server-go/app/model"
	objectStorageController "ttpos-server-go/app/modules/objectstorage/infrastructure/controller"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

// InitCacheObjectController 初始化 Desk 对象的缓存控制器
// 在 repository 包中初始化，可以访问 DeskRepo 和 CommonRepo
// 参数：
//   - underlyingCache: 底层缓存实例
//   - ttl: 缓存 TTL（默认 10 分钟）
func InitCacheObjectController(underlyingCache cache.Cache, ttl time.Duration) {
	objectStorageController.InitDeskController(
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.Desk, error) {
			deskRepo := NewDeskRepo(db)
			desk, err := deskRepo.GetDesk(
				CommonRepo.WhereByUuid(uuid),
				CommonRepo.WhereBySoftDelete(),
			)
			if err != nil {
				return nil, err
			}
			return &desk, nil
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.Desk, error) {
			deskRepo := NewDeskRepo(db)
			desks, err := deskRepo.GetDesks(
				CommonRepo.WhereInUuids(uuids),
				CommonRepo.WhereBySoftDelete(),
			)
			if err != nil {
				return nil, err
			}
			return desks, nil
		},
	)
}

