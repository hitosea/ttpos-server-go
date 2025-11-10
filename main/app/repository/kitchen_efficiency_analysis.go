package repository

import (
	"strings"
	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IKitchenEfficiencyAnalysisRepo 后厨效率分析仓库接口
type IKitchenEfficiencyAnalysisRepo interface {
	CreateKitchenEfficiencyAnalysis(analysis model.KitchenEfficiencyAnalysis) error                                                                      // 创建后厨效率分析记录
	GetKitchenEfficiencyAnalysis(opts ...DBOption) (*model.KitchenEfficiencyAnalysis, error)                                                             // 获取后厨效率分析记录
	GetKitchenEfficiencyAnalysisList(pageNo, pageSize int, opts ...DBOption) ([]*model.KitchenEfficiencyAnalysis, int64, error)                          // 获取后厨效率分析记录列表
	UpdateKitchenEfficiencyAnalysis(opts []DBOption, vars map[string]any) error                                                                          // 更新后厨效率分析记录
	GetKitchenEfficiencyAnalysisByProductPackageUuid(productPackageUuids []uint64, startTime, endTime int64) ([]*KitchenEfficiencyAnalysisResult, error) // 根据product_package_uuid查询后厨效率分析记录
	GetKitchenEfficiencyAnalysisAvg(startTime, endTime int64) (float64, error)                                                                           // 获取后厨效率分析平均值
	CalculateKitchenEfficiencyAnalysis(startTime, endTime int64) ([]*model.KitchenEfficiencyAnalysis, error)                                             // 计算后厨效率分析记录
	KitchenEfficiencyAnalysisDay(startTime, endTime int64) ([]*model.KitchenEfficiencyAnalysis, error)                                                   // 查询效率表中当天的数据
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
func (r *kitchenEfficiencyAnalysisRepo) CreateKitchenEfficiencyAnalysis(analysis model.KitchenEfficiencyAnalysis) error {
	if err := r.db.Model(&model.KitchenEfficiencyAnalysis{}).Create(&analysis).Error; err != nil {
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

type KitchenEfficiencyAnalysisResult struct {
	ProductPackageUuid uint64  `json:"product_package_uuid"`
	Min                float64 `json:"min"`
	Max                float64 `json:"max"`
	Avg                float64 `json:"avg"`
}

// 根据product_package_uuid查询后厨效率分析记录
func (r *kitchenEfficiencyAnalysisRepo) GetKitchenEfficiencyAnalysisByProductPackageUuid(productPackageUuids []uint64, startTime, endTime int64) ([]*KitchenEfficiencyAnalysisResult, error) {
	var results []*KitchenEfficiencyAnalysisResult
	db := r.db.Model(&model.KitchenEfficiencyAnalysis{}).Where("product_package_uuid IN (?)", productPackageUuids).Where("date >= ?", startTime).Where("date <= ?", endTime)
	err := db.Select("product_package_uuid, MIN(min) as min, MAX(max) as max, AVG(avg) as avg").Group("product_package_uuid").Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

// 获取后厨效率分析平均值
func (r *kitchenEfficiencyAnalysisRepo) GetKitchenEfficiencyAnalysisAvg(startTime, endTime int64) (float64, error) {
	type KitchenEfficiencyAnalysisAvg struct {
		Avg float64 `json:"avg"`
	}
	var avgResult KitchenEfficiencyAnalysisAvg
	db := r.db.Model(&model.KitchenEfficiencyAnalysis{}).Where("date >= ?", startTime).Where("date <= ?", endTime)
	err := db.Select("AVG(avg) as avg").First(&avgResult).Error
	if err != nil {
		return 0, err
	}
	return avgResult.Avg, nil
}

// 计算某个商品某天的后厨效率
func (r *kitchenEfficiencyAnalysisRepo) CalculateKitchenEfficiencyAnalysis(startTime, endTime int64) ([]*model.KitchenEfficiencyAnalysis, error) {
	var kitchenEfficiencyAnalysisList []*model.KitchenEfficiencyAnalysis
	db := r.db.Model(&model.ProductionOrderProduct{}).Where("finished_time >= ?", startTime).Where("finished_time <= ?", endTime).
		Where("delete_time = 0").
		Where("status = 2").      // 状态: 2-已完成
		Where("all_duration > 0") // 该功能之前的数据都是0,不查询之前的数据
	err := db.Select("product_package_uuid,MIN(all_duration) AS min , MAX(all_duration) AS max ,AVG(all_duration) AS avg, sum(all_duration) AS total, sum(num) AS 'count'").Group("product_package_uuid").Find(&kitchenEfficiencyAnalysisList).Error
	if err != nil {
		return nil, err
	}
	return kitchenEfficiencyAnalysisList, nil
}

// 查询效率表中当天的数据
func (r *kitchenEfficiencyAnalysisRepo) KitchenEfficiencyAnalysisDay(startTime, endTime int64) ([]*model.KitchenEfficiencyAnalysis, error) {
	var kitchenEfficiencyAnalysisList []*model.KitchenEfficiencyAnalysis
	err := r.db.Model(&model.KitchenEfficiencyAnalysis{}).Where("date >= ?", startTime).Where("date <= ?", endTime).Where("delete_time = 0").Find(&kitchenEfficiencyAnalysisList).Error
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return nil, nil
		}
		return nil, err
	}
	return kitchenEfficiencyAnalysisList, nil
}
