package repository

import (
	"time"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
)

type ICompanySettingRepo interface {
	WhereErpnextCompanyAbbr(erpnextCompanyAbbr string) DBOption
	WhereErpnextCompanyAbbrNotEmpty() DBOption
	WhereSiteCode(siteCode string) DBOption

	GetOne(opts ...DBOption) (model.CompanySetting, error)
	Get() model.CompanySetting
	GetAllByHeadquarterUuid(headquarterUuid uint64) ([]model.CompanySetting, error) // 获取总部下所有公司的设置
	UpdateSmsQuota(companyUuid uint64, quota int) error                             // 扣减公司的短信余额

	GetErpnextCompanyAbbrUuidMap(opts ...DBOption) (map[string]uint64, error)
}

func NewCompanySettingRepo(db *gorm.DB) ICompanySettingRepo {
	return NewCompanySettingRepoImpl(db)
}

type companySettingRepo struct {
	db *gorm.DB
}

func NewCompanySettingRepoImpl(db *gorm.DB) ICompanySettingRepo {
	return &companySettingRepo{db: db}
}

func (r *companySettingRepo) Get() model.CompanySetting {
	// 获取商户UUID
	companyUuid := GetCompanyUuid(r.db)
	if companyUuid == 0 {
		// 如果无法获取商户UUID，直接查询数据库
		return r.queryCompanySetting()
	}

	// 检查是否启用对象存储缓存
	if !adapter.IsObjectStorageCacheEnabled(companyUuid) {
		// 未启用缓存，直接查询数据库
		return r.queryCompanySetting()
	}

	// 使用对象存储模块缓存查询
	companySetting, err := r.getCompanySettingWithCache(companyUuid)
	if err != nil {
		// 缓存查询失败，降级到直接查询数据库
		return r.queryCompanySetting()
	}
	return *companySetting
}

// queryCompanySetting 查询商户设置（数据库查询）
func (r *companySettingRepo) queryCompanySetting() model.CompanySetting {
	var companySetting model.CompanySetting
	r.db.Model(&model.CompanySetting{}).First(&companySetting)
	return companySetting
}

// getCompanySettingWithCache 使用对象存储模块缓存查询商户设置
func (r *companySettingRepo) getCompanySettingWithCache(companyUuid uint64) (*model.CompanySetting, error) {
	if companyUuid == 0 {
		return nil, errors.New("getCompanySettingWithCache companyUuid cannot be 0")
	}

	// 构建缓存 key（使用固定的 uuid=0 表示商户级别的查询，因为每个商户只有一条设置记录）
	key := persistence.BuildKeyWithCompanyUuid(companyUuid, persistence.ObjectTypeCompanySetting, 0)

	// 获取缓存层（使用订单相关对象缓存配置）
	cacheLayer := adapter.GetOrderObjectCache[*model.CompanySetting](cache.Global, 10*time.Minute)

	// 使用缓存查询
	result, err := cacheLayer.GET(key, func() (*model.CompanySetting, error) {
		// 缓存未命中时，从数据库查询
		companySetting := r.queryCompanySetting()
		return &companySetting, nil
	})

	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return result, nil
}

func (r *companySettingRepo) UpdateSmsQuota(companyUuid uint64, quota int) error {
	if err := r.db.Model(&model.CompanySetting{}).Where("company_uuid = ?", companyUuid).Update("sms_quota", gorm.Expr("sms_quota - ?", quota)).Error; err != nil {
		return errors.WithMessage(err, "failed to update SMS quota")
	}
	return nil
}

func (r *companySettingRepo) GetAllByHeadquarterUuid(headquarterUuid uint64) ([]model.CompanySetting, error) {
	var companySettings []model.CompanySetting
	err := r.db.Model(&model.CompanySetting{}).Scopes(NotDeleted).Where("headquarter_uuid = ? or (company_uuid = ? and headquarter_uuid = 0)", headquarterUuid, headquarterUuid).Find(&companySettings).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return companySettings, nil
}

func (r *companySettingRepo) WhereErpnextCompanyAbbr(erpCompanyAbbr string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("erpnext_company_abbr = ?", erpCompanyAbbr)
	}
}

func (r *companySettingRepo) GetOne(opts ...DBOption) (model.CompanySetting, error) {
	var companySetting model.CompanySetting
	db := r.db.Model(&model.CompanySetting{})
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Take(&companySetting).Error
	return companySetting, err
}

func (r *companySettingRepo) GetErpnextCompanyAbbrUuidMap(opts ...DBOption) (map[string]uint64, error) {
	var companySettings []model.CompanySetting
	db := r.db.Model(&model.CompanySetting{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Find(&companySettings).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	erpCompanyAbbrUuidMap := make(map[string]uint64)
	for _, companySetting := range companySettings {
		erpCompanyAbbrUuidMap[companySetting.ErpnextCompanyAbbr] = companySetting.CompanyUuid
	}
	return erpCompanyAbbrUuidMap, nil
}

func (r *companySettingRepo) WhereErpnextCompanyAbbrNotEmpty() DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("erpnext_company_abbr != ''")
	}
}

func (r *companySettingRepo) WhereSiteCode(siteCode string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("erpnext_site_code = ?", siteCode)
	}
}
