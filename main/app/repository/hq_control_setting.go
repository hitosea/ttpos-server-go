package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/pkg/cache"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const hqControlSettingCacheKeyPrefix = "hq_control_setting:%d"
const hqControlSettingCacheTTL = 1 * time.Hour

type IHqControlSettingRepo interface {
	// GetControlMap 获取所有控制设置（优先缓存）
	GetControlMap(headquarterUuid uint64) map[string]int
	// IsUnifiedControl 判断指定字段是否为统一控制
	IsUnifiedControl(headquarterUuid uint64, fieldType string) bool
	// Upsert 更新或插入控制设置
	Upsert(fieldType string, controlMode int) error
	// InvalidateCache 失效缓存
	InvalidateCache(headquarterUuid uint64)
}

type hqControlSettingRepo struct {
	db *gorm.DB
}

func NewHqControlSettingRepo(db *gorm.DB) IHqControlSettingRepo {
	return &hqControlSettingRepo{db: db}
}

// GetControlMap 获取所有控制设置，返回 map[fieldType]controlMode
// 优先从 Redis 缓存读取，miss 时查 DB 并回填
func (r *hqControlSettingRepo) GetControlMap(headquarterUuid uint64) map[string]int {
	return getControlMapCached(r.db, headquarterUuid)
}

// getControlMapCached 从缓存读取控制设置，miss 时查 DB 并回填
// db 可以为 nil，此时仅从缓存读取，miss 时返回空 map（各字段取默认值）
func getControlMapCached(db *gorm.DB, headquarterUuid uint64) map[string]int {
	cacheKey := fmt.Sprintf(hqControlSettingCacheKeyPrefix, headquarterUuid)

	// 尝试从缓存读取
	if cache.Global != nil {
		if data, ok := cache.Global.Get(cacheKey); ok {
			if str, isStr := data.(string); isStr {
				var result map[string]int
				if err := json.Unmarshal([]byte(str), &result); err == nil {
					return result
				}
			}
		}
	}

	// 缓存 miss，如果有 db 则查 DB 并回填
	if db == nil {
		return make(map[string]int) // 返回空 map，各字段取默认值
	}

	var settings []model.HqControlSetting
	db.Where("delete_time = 0").Find(&settings)

	result := make(map[string]int)
	for _, s := range settings {
		result[s.FieldType] = s.ControlMode
	}

	// 回填缓存
	if cache.Global != nil {
		data, _ := json.Marshal(result)
		_ = cache.Global.Set(cacheKey, string(data), hqControlSettingCacheTTL)
	}

	return result
}

// IsUnifiedControl 判断指定字段是否为统一控制
func (r *hqControlSettingRepo) IsUnifiedControl(headquarterUuid uint64, fieldType string) bool {
	controlMap := r.GetControlMap(headquarterUuid)
	if mode, ok := controlMap[fieldType]; ok {
		return mode == constant.HqControlUnified
	}
	// 无记录时返回默认值
	return defaultControlMode(fieldType) == constant.HqControlUnified
}

// Upsert 更新或插入控制设置
func (r *hqControlSettingRepo) Upsert(fieldType string, controlMode int) error {
	setting := model.HqControlSetting{
		FieldType:   fieldType,
		ControlMode: controlMode,
	}
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "field_type"}},
		DoUpdates: clause.AssignmentColumns([]string{"control_mode", "update_time"}),
	}).Create(&setting).Error
}

// InvalidateCache 失效缓存
func (r *hqControlSettingRepo) InvalidateCache(headquarterUuid uint64) {
	if cache.Global != nil {
		cacheKey := fmt.Sprintf(hqControlSettingCacheKeyPrefix, headquarterUuid)
		cache.Global.Del(cacheKey)
	}
}

// GetHqControlMap 包级函数：从缓存读取控制设置 map，miss 时从 hqDB 读取并回填
// hqDB 可以为 nil（仅读缓存，miss 时各字段取默认值）
func GetHqControlMap(hqDB *gorm.DB, headquarterUuid uint64) map[string]int {
	return getControlMapCached(hqDB, headquarterUuid)
}

// IsHqUnifiedControl 包级函数：判断指定字段是否为统一控制
// hqDB 可以为 nil（仅读缓存，miss 时返回默认值）
func IsHqUnifiedControl(hqDB *gorm.DB, headquarterUuid uint64, fieldType string) bool {
	controlMap := getControlMapCached(hqDB, headquarterUuid)
	if mode, ok := controlMap[fieldType]; ok {
		return mode == constant.HqControlUnified
	}
	return defaultControlMode(fieldType) == constant.HqControlUnified
}

// defaultControlMode 返回字段类型的默认控制模式（无记录时使用）
func defaultControlMode(fieldType string) int {
	switch fieldType {
	case constant.HqFieldNegativeStock:
		return constant.HqControlUnified // 负库存默认统一控制
	default:
		return constant.HqControlSeparate // 其他默认分开控制
	}
}
