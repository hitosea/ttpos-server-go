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
	// 初始化 Desk 对象的缓存控制器
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

	// 初始化 SaleBillSetting 对象的缓存控制器
	objectStorageController.InitSaleBillSettingController(
		underlyingCache,
		ttl,
		func(db *gorm.DB, saleBillUuid uint64) (*model.SaleBillSetting, error) {
			orderRepo := NewOrderRepo(db)
			return orderRepo.GetSaleBillSetting(saleBillUuid)
		},
		func(db *gorm.DB, saleBillUuids []uint64) ([]*model.SaleBillSetting, error) {
			// 批量查询 SaleBillSetting
			orderRepo := NewOrderRepo(db)
			settings := make([]*model.SaleBillSetting, 0, len(saleBillUuids))
			for _, saleBillUuid := range saleBillUuids {
				if saleBillUuid > 0 {
					setting, err := orderRepo.GetSaleBillSetting(saleBillUuid)
					if err != nil {
						// 单个查询失败不影响其他，继续查询
						continue
					}
					if setting != nil {
						settings = append(settings, setting)
					}
				}
			}
			return settings, nil
		},
	)

	// 初始化 ProductBomCard 对象的缓存控制器
	objectStorageController.InitProductBomCardController(
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.ProductBomCard, error) {
			productBomCardRepo := NewProductBomCardRepo(db)
			return productBomCardRepo.GetProductBomCardWithMaterials(uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.ProductBomCard, error) {
			productBomCardRepo := NewProductBomCardRepo(db)
			return productBomCardRepo.GetProductBomCardsWithMaterials(uuids)
		},
	)

	// 初始化 Member 对象的缓存控制器
	objectStorageController.InitMemberController(
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.Member, error) {
			memberRepo := NewMemberRepo(db)
			return memberRepo.GetMemberWithAssociations(uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.Member, error) {
			memberRepo := NewMemberRepo(db)
			return memberRepo.GetMembersWithAssociations(uuids)
		},
	)
}
