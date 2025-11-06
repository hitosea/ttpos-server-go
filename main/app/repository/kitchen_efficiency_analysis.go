package repository

import (
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IKitchenEfficiencyAnalysisRepo 后厨效率分析仓库接口
type IKitchenEfficiencyAnalysisRepo interface {
	CreateKitchenEfficiencyAnalysis(analysis *model.KitchenEfficiencyAnalysis) error                                            // 创建后厨效率分析记录
	GetKitchenEfficiencyAnalysis(opts ...DBOption) (*model.KitchenEfficiencyAnalysis, error)                                    // 获取后厨效率分析记录
	GetKitchenEfficiencyAnalysisList(pageNo, pageSize int, opts ...DBOption) ([]*model.KitchenEfficiencyAnalysis, int64, error) // 获取后厨效率分析记录列表
	UpdateKitchenEfficiencyAnalysis(opts []DBOption, vars map[string]any) error                                                 // 更新后厨效率分析记录
	WhereCompanyUUID(uuid uint64) DBOption                                                                                      // 公司UUID条件
	WhereShopUUID(uuid uint64) DBOption                                                                                         // 门店UUID条件
	WhereProductCategoryUUID(uuid uint64) DBOption                                                                              // 商品分类UUID条件
	WhereProductBomUUID(uuid uint64) DBOption                                                                                   // 商品BOM UUID条件
	WhereStatisticsDate(date uint) DBOption                                                                                     // 统计日期条件
	WhereStatisticsDateBetween(startTime, endTime uint) DBOption                                                                // 统计日期范围条件
}

type kitchenEfficiencyAnalysisRepo struct {
	db *gorm.DB
}

// NewKitchenEfficiencyAnalysisRepo 创建新的后厨效率分析仓库
func NewKitchenEfficiencyAnalysisRepo(db *gorm.DB) IKitchenEfficiencyAnalysisRepo {
	return NewKitchenEfficiencyAnalysisRepoImpl(db)
}

// NewKitchenEfficiencyAnalysisRepoImpl 创建新的后厨效率分析仓库实现
func NewKitchenEfficiencyAnalysisRepoImpl(db *gorm.DB) IKitchenEfficiencyAnalysisRepo {
	return &kitchenEfficiencyAnalysisRepo{db: db}
}

// CreateKitchenEfficiencyAnalysis 创建后厨效率分析记录
func (r *kitchenEfficiencyAnalysisRepo) CreateKitchenEfficiencyAnalysis(analysis *model.KitchenEfficiencyAnalysis) error {
	if err := r.db.Model(analysis).Create(analysis).Error; err != nil {
		return err
	}
	return nil
}

// GetKitchenEfficiencyAnalysis 获取后厨效率分析记录
func (r *kitchenEfficiencyAnalysisRepo) GetKitchenEfficiencyAnalysis(opts ...DBOption) (*model.KitchenEfficiencyAnalysis, error) {
	var analysis model.KitchenEfficiencyAnalysis
	db := r.db.Model(&model.KitchenEfficiencyAnalysis{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	result := db.First(&analysis)
	if result.Error != nil {
		return nil, result.Error
	}
	return &analysis, nil
}

// GetKitchenEfficiencyAnalysisList 获取后厨效率分析记录列表
func (r *kitchenEfficiencyAnalysisRepo) GetKitchenEfficiencyAnalysisList(pageNo, pageSize int, opts ...DBOption) ([]*model.KitchenEfficiencyAnalysis, int64, error) {
	var analysisList []*model.KitchenEfficiencyAnalysis
	var total int64
	db := r.db.Model(&model.KitchenEfficiencyAnalysis{}).Scopes(NotDeleted)
	for _, opt := range opts {
		db = opt(db)
	}
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = db.Offset((pageNo - 1) * pageSize).Limit(pageSize).Find(&analysisList).Error
	if err != nil {
		return nil, 0, err
	}
	return analysisList, total, nil
}

// UpdateKitchenEfficiencyAnalysis 更新后厨效率分析记录
func (r *kitchenEfficiencyAnalysisRepo) UpdateKitchenEfficiencyAnalysis(opts []DBOption, vars map[string]any) error {
	db := r.db.Model(&model.KitchenEfficiencyAnalysis{})
	for _, opt := range opts {
		db = opt(db)
	}
	return db.Updates(vars).Error
}

// WhereCompanyUUID 公司UUID条件
func (r *kitchenEfficiencyAnalysisRepo) WhereCompanyUUID(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("company_uuid = ?", uuid)
	}
}

// WhereShopUUID 门店UUID条件
func (r *kitchenEfficiencyAnalysisRepo) WhereShopUUID(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("shop_uuid = ?", uuid)
	}
}

// WhereProductCategoryUUID 商品分类UUID条件
func (r *kitchenEfficiencyAnalysisRepo) WhereProductCategoryUUID(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("product_category_uuid = ?", uuid)
	}
}

// WhereProductBomUUID 商品BOM UUID条件
func (r *kitchenEfficiencyAnalysisRepo) WhereProductBomUUID(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("product_bom_uuid = ?", uuid)
	}
}

// WhereStatisticsDate 统计日期条件
func (r *kitchenEfficiencyAnalysisRepo) WhereStatisticsDate(date uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("statistics_date = ?", date)
	}
}

// WhereStatisticsDateBetween 统计日期范围条件
func (r *kitchenEfficiencyAnalysisRepo) WhereStatisticsDateBetween(startTime, endTime uint) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("statistics_date BETWEEN ? AND ?", startTime, endTime)
	}
}
