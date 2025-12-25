package application

import (
	"context"
	"time"

	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/domain/entity"
	"ttpos-server-go/app/modules/objectstorage/domain/service"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

// IObjectStorageAppService 对象存储应用服务接口
type IObjectStorageAppService interface {
	// PreloadSaleBillAssociations 自动注入 SaleBill 的关联对象
	PreloadSaleBillAssociations(ctx context.Context, saleBill *model.SaleBill, db *gorm.DB) error

	// GetSaleBillAssociations 获取 SaleBill 的关联配置
	GetSaleBillAssociations(ctx context.Context, db *gorm.DB) []entity.Association
}

// objectStorageAppService 对象存储应用服务实现
type objectStorageAppService struct {
	objectStorage service.IObjectStorage[*model.SaleBill]
	config        *service.Config
	dbm           *database.DBManager
}

// NewObjectStorageAppService 创建对象存储应用服务
func NewObjectStorageAppService(
	cacheInstance cache.Cache,
	dbm *database.DBManager,
) IObjectStorageAppService {
	// 创建缓存适配器
	cacheAdapter := adapter.NewCacheAdapter(cacheInstance)

	// 创建配置
	config := &service.Config{
		TTL:          5 * time.Minute, // 默认 5 分钟
		DisableCache: false,
		CacheLayer:   cacheAdapter,
	}

	// 为不同对象类型设置不同的 TTL
	config.SetTTL("sale_bill_setting", 10*time.Minute)
	config.SetTTL("desk", 10*time.Minute)
	config.SetTTL("product_package", 5*time.Minute)
	config.SetTTL("multi_language_name", 30*time.Minute)
	config.SetTTL("product_category", 30*time.Minute)
	config.SetTTL("product_bom", 5*time.Minute)
	config.SetTTL("product_package_attribute_group", 5*time.Minute)
	config.SetTTL("product_attribute", 5*time.Minute)
	config.SetTTL("product_flavor", 5*time.Minute)
	config.SetTTL("product_sauce", 5*time.Minute)
	config.SetTTL("batch_tag", 10*time.Minute)

	return &objectStorageAppService{
		objectStorage: persistence.NewObjectStorage[*model.SaleBill](config),
		config:        config,
		dbm:           dbm,
	}
}

// GetSaleBillAssociations 获取 SaleBill 的关联配置
func (s *objectStorageAppService) GetSaleBillAssociations(ctx context.Context, db *gorm.DB) []entity.Association {
	deskRepo := repository.NewDeskRepo(db)
	productPackageRepo := repository.NewProductPackageRepo(db)
	multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(db)
	productCategoryRepo := repository.NewProductCategoryRepo(db)

	return []entity.Association{
		// 一对一关系：SaleBillSetting
		{
			Path:       "SaleBillSetting",
			ObjectType: "sale_bill_setting",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(*model.SaleBill).Uuid
			},
			QueryFunc: func(ctx context.Context, uuid uint64) (interface{}, error) {
				var setting model.SaleBillSetting
				err := db.Where("sale_bill_uuid = ?", uuid).First(&setting).Error
				if err != nil {
					return nil, err
				}
				return &setting, nil
			},
		},
		// 一对一关系：Desk
		{
			Path:       "Desk",
			ObjectType: "desk",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(*model.SaleBill).DeskUuid
			},
			QueryFunc: func(ctx context.Context, uuid uint64) (interface{}, error) {
				desk, err := deskRepo.GetDesk(
					repository.CommonRepo.WhereByUuid(uuid),
					repository.CommonRepo.WhereBySoftDelete(),
				)
				if err != nil {
					return nil, err
				}
				return &desk, nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.ProductPackage（支持批量优化）
		{
			Path:       "SaleOrders.SaleOrderProducts.ProductPackage",
			ObjectType: "product_package",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(*model.SaleOrderProduct).ProductPackageUuid
			},
			QueryFunc: func(ctx context.Context, uuid uint64) (interface{}, error) {
				return productPackageRepo.GetProductPackage(
					repository.CommonRepo.WhereByUuid(uuid),
					repository.CommonRepo.WhereBySoftDelete(),
				)
			},
			BatchQueryFunc: func(ctx context.Context, uuids []uint64) (map[uint64]interface{}, error) {
				packages, err := productPackageRepo.GetProductPackageListByUuids(uuids)
				if err != nil {
					return nil, err
				}
				result := make(map[uint64]interface{})
				for i := range packages {
					result[packages[i].Uuid] = &packages[i]
				}
				return result, nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.MultiLanguageName（支持批量优化）
		{
			Path:       "SaleOrders.SaleOrderProducts.MultiLanguageName",
			ObjectType: "multi_language_name",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(*model.SaleOrderProduct).MultiLanguageNameUuid
			},
			QueryFunc: func(ctx context.Context, uuid uint64) (interface{}, error) {
				return multiLanguageNameRepo.GetMultiLanguageName(
					repository.CommonRepo.WhereByUuid(uuid),
				)
			},
			BatchQueryFunc: func(ctx context.Context, uuids []uint64) (map[uint64]interface{}, error) {
				result := make(map[uint64]interface{})
				for _, uuid := range uuids {
					name, err := multiLanguageNameRepo.GetMultiLanguageName(
						repository.CommonRepo.WhereByUuid(uuid),
					)
					if err == nil {
						result[uuid] = name
					}
				}
				return result, nil
			},
		},
		// 嵌套关联：SaleOrders.SaleOrderProducts.ProductPackage.ProductCategory
		{
			Path:       "SaleOrders.SaleOrderProducts.ProductPackage.ProductCategory",
			ObjectType: "product_category",
			GetUUID: func(obj interface{}) uint64 {
				return obj.(*model.ProductPackage).CategoryUuid
			},
			QueryFunc: func(ctx context.Context, uuid uint64) (interface{}, error) {
				return productCategoryRepo.GetProductCategory(
					repository.CommonRepo.WhereByUuid(uuid),
				)
			},
		},
	}
}

// PreloadSaleBillAssociations 自动注入 SaleBill 的关联对象
func (s *objectStorageAppService) PreloadSaleBillAssociations(ctx context.Context, saleBill *model.SaleBill, db *gorm.DB) error {
	associations := s.GetSaleBillAssociations(ctx, db)
	return s.objectStorage.PreloadWithConfig(ctx, saleBill, associations)
}

