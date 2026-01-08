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

	// 初始化 PaymentMethod 对象的缓存控制器
	objectStorageController.InitPaymentMethodController(
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.PaymentMethod, error) {
			paymentMethodRepo := NewPaymentMethodRepo(db)
			return paymentMethodRepo.GetPaymentMethodByUuid(uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.PaymentMethod, error) {
			paymentMethodRepo := NewPaymentMethodRepo(db)
			return paymentMethodRepo.GetPaymentMethodsByUuids(uuids)
		},
	)

	// 初始化 Nationality 对象的缓存控制器
	objectStorageController.InitNationalityController(
		underlyingCache,
		5*time.Minute, // 使用 5 分钟 TTL，与现有代码保持一致
		func(db *gorm.DB, uuid uint64) (*model.Nationality, error) {
			nationalityRepo := NewNationalityRepo(db)
			// 使用 FindByUuidWithDeleted，保证历史订单可显示已删除的配置名称
			return nationalityRepo.FindByUuidWithDeleted(uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.Nationality, error) {
			nationalityRepo := NewNationalityRepo(db)
			// 使用批量查询方法，提高性能
			return nationalityRepo.FindByUuidsWithDeleted(uuids)
		},
	)

	// 初始化 OrderSource 对象的缓存控制器
	objectStorageController.InitOrderSourceController(
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.OrderSource, error) {
			orderSourceRepo := NewOrderSourceRepo(db)
			// 使用 FindByUuidWithDeleted，保证历史订单可显示已删除的配置名称
			return orderSourceRepo.FindByUuidWithDeleted(uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.OrderSource, error) {
			orderSourceRepo := NewOrderSourceRepo(db)
			// 使用批量查询方法，提高性能
			return orderSourceRepo.FindByUuidsWithDeleted(uuids)
		},
	)

	// 初始化 MemberCoupon 对象的缓存控制器
	objectStorageController.InitMemberCouponController(
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.MemberCoupon, error) {
			memberCouponRepo := NewMemberCouponRepo(db)
			return memberCouponRepo.GetMemberCouponByUuid(uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.MemberCoupon, error) {
			memberCouponRepo := NewMemberCouponRepo(db)
			return memberCouponRepo.GetMembersByUuids(uuids)
		},
	)

	// 初始化 MarketingCoupon 对象的缓存控制器
	objectStorageController.InitMarketingCouponController(
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.MarketingCoupon, error) {
			marketingCouponRepo := NewMarketingCouponRepo(db)
			return marketingCouponRepo.GetCouponByUuid(uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.MarketingCoupon, error) {
			marketingCouponRepo := NewMarketingCouponRepo(db)
			return marketingCouponRepo.GetCoupons(
				CommonRepo.WhereInUuids(uuids),
			)
		},
	)

	// 初始化 CompanySetting 对象的缓存控制器
	objectStorageController.InitCompanySettingController(
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.CompanySetting, error) {
			companySettingRepo := NewCompanySettingRepo(db)
			// 每个商户只有一条设置记录，直接查询
			companySetting := companySettingRepo.GetCompanySetting()
			return &companySetting, nil
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.CompanySetting, error) {
			companySettingRepo := NewCompanySettingRepo(db)
			return companySettingRepo.GetCompanySettingsByCompanyUuids(uuids)
		},
	)

	// 初始化 ProductPackage 对象的缓存控制器
	objectStorageController.InitProductPackageController(
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.ProductPackage, error) {
			productPackageRepo := NewProductPackageRepo(db)
			return productPackageRepo.GetProductPackageByUuidWithAssociations(uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.ProductPackage, error) {
			productPackageRepo := NewProductPackageRepo(db)
			return productPackageRepo.GetProductPackagesByUuidsWithAssociations(uuids)
		},
	)

	// 初始化 ProductAttribute 对象的缓存控制器
	objectStorageController.InitProductAttributeController(
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.ProductAttribute, error) {
			productAttributeRepo := NewProductAttributeRepo(db)
			return productAttributeRepo.GetProductAttributeByUuid(uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.ProductAttribute, error) {
			productAttributeRepo := NewProductAttributeRepo(db)
			return productAttributeRepo.GetProductAttributesByUuids(uuids)
		},
	)

	// 初始化 ProductBom 对象的缓存控制器
	objectStorageController.InitProductBomController(
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.ProductBom, error) {
			productBomRepo := NewProductBomRepo(db)
			return productBomRepo.GetProductBomByUuid(uuid)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.ProductBom, error) {
			productBomRepo := NewProductBomRepo(db)
			return productBomRepo.GetProductBomsByUuidsWithAssociations(uuids)
		},
	)

	// 初始化 MultiLanguageName 对象的缓存控制器
	objectStorageController.InitMultiLanguageNameController(
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.MultiLanguageName, error) {
			multiLanguageNameRepo := NewMultiLanguageNameRepo(db)
			// GetMultiLanguageNameByUuid 返回值类型，需要转换为指针
			name, err := multiLanguageNameRepo.GetMultiLanguageNameByUuid(uuid)
			if err != nil {
				return nil, err
			}
			return &name, nil
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.MultiLanguageName, error) {
			multiLanguageNameRepo := NewMultiLanguageNameRepo(db)
			return multiLanguageNameRepo.GetMultiLanguageNameListByUuids(uuids)
		},
	)

	// 初始化 SaleBill 对象的缓存控制器
	objectStorageController.InitSaleBillController(
		underlyingCache,
		ttl,
		func(db *gorm.DB, uuid uint64) (*model.SaleBill, error) {
			return NewOrderRepo(db).QuerySaleBillForObjectStorage(uuid, nil)
		},
		func(db *gorm.DB, uuids []uint64) ([]*model.SaleBill, error) {
			saleBills := make([]*model.SaleBill, 0, len(uuids))
			for _, uuid := range uuids {
				saleBill, err := NewOrderRepo(db).QuerySaleBillForObjectStorage(uuid, nil)
				if err != nil {
					continue
				}
				saleBills = append(saleBills, saleBill)
			}
			return saleBills, nil
		},
	)
}
