package repository

import (
	goCtx "context"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	"ttpos-server-go/app/modules/objectstorage/infrastructure/controller"
	"ttpos-server-go/pkg/context"

	"gorm.io/gorm"
)

type ICompanySettingRepo interface {
	WhereErpnextCompanyAbbr(erpnextCompanyAbbr string) DBOption
	WhereErpnextCompanyAbbrNotEmpty() DBOption
	WhereSiteCode(siteCode string) DBOption

	GetOne(opts ...DBOption) (model.CompanySetting, error)
	Get() model.CompanySetting
	GetCompanySetting() model.CompanySetting
	GetCompanySettingsByCompanyUuids(companyUuids []uint64) ([]*model.CompanySetting, error)
	GetAllByHeadquarterUuid(headquarterUuid uint64) ([]model.CompanySetting, error)         // 获取总部下所有公司的设置
	UpdateSmsQuota(companyUuid uint64, quota int) error                                     // 扣减公司的短信余额
	UpdateByCompanyUuid(companyUuid uint64, data map[string]any) error                      // 根据公司UUID更新设置
	UpdateByCompanyUuids(companyUuids []uint64, data map[string]any) error                  // 根据公司UUID列表批量更新设置
	UpdateBySiteCodeAndCompanyAbbr(siteCode, companyAbbr string, data map[string]any) error // 根据站点编码和公司缩写更新设置

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
		return r.GetCompanySetting()
	}

	// 检查是否启用对象存储缓存
	if !adapter.IsObjectStorageCacheEnabled(companyUuid) {
		// 未启用缓存，直接查询数据库
		return r.GetCompanySetting()
	}

	// 使用对象存储模块缓存查询
	companySetting, err := r.getCompanySettingWithCache(companyUuid)
	if err != nil {
		// 缓存查询失败，降级到直接查询数据库
		return r.GetCompanySetting()
	}
	return *companySetting
}

// queryCompanySetting 查询商户设置（数据库查询）
func (r *companySettingRepo) GetCompanySetting() model.CompanySetting {
	var companySetting model.CompanySetting
	r.db.Model(&model.CompanySetting{}).First(&companySetting)
	return companySetting
}

// getCompanySettingWithCache 使用对象存储模块缓存查询商户设置
func (r *companySettingRepo) getCompanySettingWithCache(companyUuid uint64) (*model.CompanySetting, error) {
	if companyUuid == 0 {
		return nil, errors.New("getCompanySettingWithCache companyUuid cannot be 0")
	}

	// 创建包含 companyUuid 的 context
	ctx := context.NewContext(
		context.WithCompanyUuid(companyUuid),
		context.WithContext(goCtx.Background()),
	)

	// 使用控制器查询
	controller := controller.GetCompanySettingController()
	result, err := controller.GetByUuid(ctx, r.db, companyUuid)
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

// GetCompanySettingsByCompanyUuids 批量通过companyUuids列表查询商户设置
func (r *companySettingRepo) GetCompanySettingsByCompanyUuids(companyUuids []uint64) ([]*model.CompanySetting, error) {
	if len(companyUuids) == 0 {
		return []*model.CompanySetting{}, nil
	}
	var companySettings []*model.CompanySetting
	err := r.db.Model(&model.CompanySetting{}).
		Scopes(NotDeleted).
		Where("company_uuid IN (?)", companyUuids).
		Find(&companySettings).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return companySettings, nil
}

// UpdateByCompanyUuid 根据公司UUID更新设置
func (r *companySettingRepo) UpdateByCompanyUuid(companyUuid uint64, data map[string]any) error {
	return r.db.Model(&model.CompanySetting{}).Where("company_uuid = ?", companyUuid).Updates(data).Error
}

// UpdateByCompanyUuids 根据公司UUID列表批量更新设置
func (r *companySettingRepo) UpdateByCompanyUuids(companyUuids []uint64, data map[string]any) error {
	if len(companyUuids) == 0 {
		return nil
	}
	return r.db.Model(&model.CompanySetting{}).Where("company_uuid IN (?)", companyUuids).Updates(data).Error
}

// UpdateBySiteCodeAndCompanyAbbr 根据站点编码和公司缩写更新设置
func (r *companySettingRepo) UpdateBySiteCodeAndCompanyAbbr(siteCode, companyAbbr string, data map[string]any) error {
	return r.db.Model(&model.CompanySetting{}).Where("erpnext_site_code = ? AND erpnext_company_abbr = ?", siteCode, companyAbbr).Updates(data).Error
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
