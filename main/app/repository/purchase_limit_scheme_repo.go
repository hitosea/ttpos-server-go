package repository

import (
	"time"

	"ttpos-server-go/app/model"

	"gorm.io/gorm"
)

// IPurchaseLimitSchemeRepo 限购方案数据访问接口
type IPurchaseLimitSchemeRepo interface {
	// Create 创建限购方案
	Create(scheme *model.PurchaseLimitScheme) error

	// Update 更新限购方案
	Update(scheme *model.PurchaseLimitScheme) error

	// GetByUuid 根据UUID查询限购方案
	GetByUuid(uuid uint64, options ...DBOption) (*model.PurchaseLimitScheme, error)

	// GetList 查询限购方案列表
	GetList(options ...DBOption) ([]*model.PurchaseLimitScheme, int64, error)

	// GetActiveSchemes 查询所有启用的限购方案（按门店过滤）
	GetActiveSchemes(companyUuid uint64) ([]*model.PurchaseLimitScheme, error)

	// GetMinQuotaLimitBatchByMaterialCodes 批量获取物品的最小限购数量
	GetMinQuotaLimitBatchByMaterialCodes(companyUuid uint64, materialCodes []string, currentWeekday int8) (map[string]float64, error)

	// GetDisallowedPurchaseMaterialCodes 批量获取禁止采购的物品编码
	GetDisallowedPurchaseMaterialCodes(companyUuid uint64, materialCodes []string, currentWeekday int8) ([]string, error)

	// GetMinDailyLimit 获取当天最小的每日申请次数限制
	GetMinDailyLimit(companyUuid uint64, currentWeekday int8) (int, error)

	// Delete 软删除限购方案
	Delete(uuid uint64) error

	// 选项方法
	WhereUuid(uuid uint64) DBOption
	WhereStatus(status int8) DBOption
	WhereName(name string) DBOption
	WhereNameLike(name string) DBOption
	Paginate(pageNo, pageSize int) DBOption
}

type purchaseLimitSchemeRepoImpl struct {
	db *gorm.DB // ✅ 只持有 db 实例
}

// NewPurchaseLimitSchemeRepo 创建限购方案仓储实例
func NewPurchaseLimitSchemeRepo(db *gorm.DB) IPurchaseLimitSchemeRepo {
	return &purchaseLimitSchemeRepoImpl{db: db}
}

// Create 创建限购方案
func (r *purchaseLimitSchemeRepoImpl) Create(scheme *model.PurchaseLimitScheme) error {
	return r.db.Create(scheme).Error
}

// Update 更新限购方案
func (r *purchaseLimitSchemeRepoImpl) Update(scheme *model.PurchaseLimitScheme) error {
	return r.db.Save(scheme).Error
}

// GetByUuid 根据UUID查询限购方案
func (r *purchaseLimitSchemeRepoImpl) GetByUuid(uuid uint64, options ...DBOption) (*model.PurchaseLimitScheme, error) {
	var scheme model.PurchaseLimitScheme
	db := r.db.Where("delete_time = ?", 0)

	for _, option := range options {
		db = option(db)
	}

	if err := db.Where("uuid = ?", uuid).First(&scheme).Error; err != nil {
		return nil, err
	}
	return &scheme, nil
}

// GetList 查询限购方案列表
func (r *purchaseLimitSchemeRepoImpl) GetList(options ...DBOption) ([]*model.PurchaseLimitScheme, int64, error) {
	var list []*model.PurchaseLimitScheme
	var total int64

	db := r.db.Where("delete_time = ?", 0)

	for _, option := range options {
		db = option(db)
	}

	if err := db.Model(&model.PurchaseLimitScheme{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := db.Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

// Delete 软删除限购方案
func (r *purchaseLimitSchemeRepoImpl) Delete(uuid uint64) error {
	return r.db.Model(&model.PurchaseLimitScheme{}).
		Where("uuid = ?", uuid).
		Update("delete_time", time.Now().Unix()).Error
}

// GetActiveSchemes 查询所有启用的限购方案（按门店过滤）
func (r *purchaseLimitSchemeRepoImpl) GetActiveSchemes(companyUuid uint64) ([]*model.PurchaseLimitScheme, error) {
	var schemes []*model.PurchaseLimitScheme

	// 查询启用状态的方案
	query := r.db.Where("delete_time = ?", 0).Where("status = ?", 1)

	// 子查询：查找应用到该门店的方案UUID
	// 1. 应用到全部门店的方案（apply_to_all_shops = 1）
	// 2. 或者在 purchase_limit_scheme_shop 表中关联到该门店的方案
	subQuery := r.db.Table("ttpos_purchase_limit_scheme_shop").
		Select("scheme_uuid").
		Where("company_uuid = ?", companyUuid).
		Where("delete_time = ?", 0)

	query = query.Where("apply_to_all_shops = 1 OR uuid IN (?)", subQuery)

	if err := query.Find(&schemes).Error; err != nil {
		return nil, err
	}

	return schemes, nil
}

// 选项方法
func (r *purchaseLimitSchemeRepoImpl) WhereUuid(uuid uint64) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	}
}

func (r *purchaseLimitSchemeRepoImpl) WhereStatus(status int8) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("status = ?", status)
	}
}

func (r *purchaseLimitSchemeRepoImpl) WhereName(name string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name = ?", name)
	}
}

func (r *purchaseLimitSchemeRepoImpl) WhereNameLike(name string) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("name LIKE ?", "%"+name+"%")
	}
}

func (r *purchaseLimitSchemeRepoImpl) Paginate(pageNo, pageSize int) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		offset := (pageNo - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// GetMinQuotaLimitBatchByMaterialCodes 批量获取物品的最小限购数量
//
// 参数：
//   - materialCodes: 物品编码列表
//   - companyUuid: 门店UUID（用于判断方案是否应用到该门店）
//   - currentWeekday: 当前星期几（1-7，1=周一，7=周日）
//
// 返回：
//   - map[string]float64: 物品编码 -> 最小限购数量的映射（如果没有限购配置，则不包含该物品）
//   - error: 错误信息
//
// 逻辑：
//  1. 查询所有启用的、应用到该门店的限购方案
//  2. 过滤出包含当前星期的方案
//  3. 在这些方案中查找匹配的物品配置
//  4. 对每个物品取最小的限购数量（排除 0 值，0 表示不限制）
func (r *purchaseLimitSchemeRepoImpl) GetMinQuotaLimitBatchByMaterialCodes(
	companyUuid uint64,
	materialCodes []string,
	currentWeekday int8,
) (map[string]float64, error) {
	if len(materialCodes) == 0 {
		return make(map[string]float64), nil
	}

	// 1. 子查询：查找应用到该门店的方案UUID
	subQuery := r.db.Table("ttpos_purchase_limit_scheme_shop").
		Select("scheme_uuid").
		Where("company_uuid = ?", companyUuid).
		Where("delete_time = ?", 0)

	// 2. 查询所有启用的、应用到该门店的限购方案
	var schemes []struct {
		SchemeUuid uint64 `gorm:"column:scheme_uuid"`
		Weekdays   string `gorm:"column:weekdays"`
	}

	err := r.db.Table("ttpos_purchase_limit_scheme_item as item").
		Select("DISTINCT scheme.uuid as scheme_uuid, scheme.weekdays").
		Joins("INNER JOIN ttpos_purchase_limit_scheme as scheme ON scheme.uuid = item.scheme_uuid").
		Where("item.delete_time = ?", 0).
		Where("item.material_code IN ?", materialCodes).
		Where("scheme.delete_time = ?", 0).
		Where("scheme.status = ?", 1).
		Where("(scheme.apply_to_all_shops = 1 OR scheme.uuid IN (?))", subQuery).
		Find(&schemes).Error

	if err != nil {
		return nil, err
	}

	if len(schemes) == 0 {
		return make(map[string]float64), nil // 没有限购配置
	}

	// 3. 过滤出包含当前星期的方案
	validSchemeUuids := make([]uint64, 0)
	for _, scheme := range schemes {
		if isWeekdayInScheme(currentWeekday, scheme.Weekdays) {
			validSchemeUuids = append(validSchemeUuids, scheme.SchemeUuid)
		}
	}

	if len(validSchemeUuids) == 0 {
		return make(map[string]float64), nil // 当前星期不在限购周期内
	}

	// 4. 查询这些方案中的物品配置，按物品分组取最小限购数量
	var results []struct {
		MaterialCode string  `gorm:"column:material_code"`
		MinQuota     float64 `gorm:"column:min_quota"`
	}

	err = r.db.Table("ttpos_purchase_limit_scheme_item").
		Select("material_code, MIN(quota_limit) as min_quota").
		Where("delete_time = ?", 0).
		Where("material_code IN ?", materialCodes).
		Where("scheme_uuid IN ?", validSchemeUuids).
		Where("quota_limit > ?", 0). // 排除 0 值（0 表示不限制）
		Group("material_code").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	// 5. 转换为 map
	quotaMap := make(map[string]float64, len(results))
	for _, result := range results {
		quotaMap[result.MaterialCode] = result.MinQuota
	}

	return quotaMap, nil
}

// GetDisallowedPurchaseMaterialCodes 批量获取禁止采购的物品编码
//
// 参数：
//   - companyUuid: 门店UUID（用于判断方案是否应用到该门店）
//   - materialCodes: 物品编码列表
//   - currentWeekday: 当前星期几（1-7，1=周一，7=周日）
//
// 返回：
//   - []string: 禁止采购的物品编码列表
//   - error: 错误信息
//
// 逻辑：
//  1. 查询所有启用的、应用到该门店的限购方案
//  2. 过滤出包含当前星期的方案
//  3. 在这些方案中查找 is_allow_purchase = 'no' 的物品
func (r *purchaseLimitSchemeRepoImpl) GetDisallowedPurchaseMaterialCodes(
	companyUuid uint64,
	materialCodes []string,
	currentWeekday int8,
) ([]string, error) {
	if len(materialCodes) == 0 {
		return []string{}, nil
	}

	// 1. 子查询：查找应用到该门店的方案UUID
	subQuery := r.db.Table("ttpos_purchase_limit_scheme_shop").
		Select("scheme_uuid").
		Where("company_uuid = ?", companyUuid).
		Where("delete_time = ?", 0)

	// 2. 查询所有启用的、应用到该门店的限购方案中禁止采购的物品
	var schemes []struct {
		SchemeUuid uint64 `gorm:"column:scheme_uuid"`
		Weekdays   string `gorm:"column:weekdays"`
	}

	err := r.db.Table("ttpos_purchase_limit_scheme_item as item").
		Select("DISTINCT scheme.uuid as scheme_uuid, scheme.weekdays").
		Joins("INNER JOIN ttpos_purchase_limit_scheme as scheme ON scheme.uuid = item.scheme_uuid").
		Where("item.delete_time = ?", 0).
		Where("item.material_code IN ?", materialCodes).
		Where("item.is_allow_purchase = ?", "no").
		Where("scheme.delete_time = ?", 0).
		Where("scheme.status = ?", 1).
		Where("(scheme.apply_to_all_shops = 1 OR scheme.uuid IN (?))", subQuery).
		Find(&schemes).Error

	if err != nil {
		return nil, err
	}

	if len(schemes) == 0 {
		return []string{}, nil
	}

	// 3. 过滤出包含当前星期的方案
	validSchemeUuids := make([]uint64, 0)
	for _, scheme := range schemes {
		if isWeekdayInScheme(currentWeekday, scheme.Weekdays) {
			validSchemeUuids = append(validSchemeUuids, scheme.SchemeUuid)
		}
	}

	if len(validSchemeUuids) == 0 {
		return []string{}, nil
	}

	// 4. 查询这些方案中禁止采购的物品编码
	var results []string

	err = r.db.Table("ttpos_purchase_limit_scheme_item").
		Select("DISTINCT material_code").
		Where("delete_time = ?", 0).
		Where("material_code IN ?", materialCodes).
		Where("scheme_uuid IN ?", validSchemeUuids).
		Where("is_allow_purchase = ?", "no").
		Pluck("material_code", &results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetMinDailyLimit 获取当天最小的每日申请次数限制
//
// 参数：
//   - companyUuid: 门店UUID（用于判断方案是否应用到该门店）
//   - currentWeekday: 当前星期几（1-7，1=周一，7=周日）
//
// 返回：
//   - int: 最小的每日申请次数限制（如果没有配置或所有配置都为0，则返回0表示不限制）
//   - error: 错误信息
//
// 逻辑：
//  1. 查询所有启用的、应用到该门店的限购方案
//  2. 过滤出包含当前星期的方案
//  3. 取这些方案中最小的 daily_limit（排除 -1 值，-1 表示不限制）
func (r *purchaseLimitSchemeRepoImpl) GetMinDailyLimit(
	companyUuid uint64,
	currentWeekday int8,
) (int, error) {
	// 1. 子查询：查找应用到该门店的方案UUID
	subQuery := r.db.Table("ttpos_purchase_limit_scheme_shop").
		Select("scheme_uuid").
		Where("company_uuid = ?", companyUuid).
		Where("delete_time = ?", 0)

	// 2. 查询所有启用的、应用到该门店的限购方案
	var schemes []struct {
		DailyLimit int    `gorm:"column:daily_limit"`
		Weekdays   string `gorm:"column:weekdays"`
	}

	err := r.db.Table("ttpos_purchase_limit_scheme").
		Select("daily_limit, weekdays").
		Where("daily_limit != ?", -1).
		Where("delete_time = ?", 0).
		Where("status = ?", 1).
		Where("apply_to_all_shops = 1 OR uuid IN (?)", subQuery).
		Find(&schemes).Error

	if err != nil {
		return -1, err
	}

	if len(schemes) == 0 {
		return -1, nil // 没有限购配置，返回0表示不限制
	}

	// 3. 过滤出包含当前星期的方案，并取最小的 daily_limit
	minDailyLimit := -1
	for _, scheme := range schemes {
		if isWeekdayInScheme(currentWeekday, scheme.Weekdays) {
			// 排除 0 值（0 表示不限制）
			if scheme.DailyLimit != -1 {
				if minDailyLimit == -1 || scheme.DailyLimit < minDailyLimit {
					minDailyLimit = scheme.DailyLimit
				}
			}
		}
	}

	return minDailyLimit, nil
}

// isWeekdayInScheme 检查当前星期是否在方案的限购周期内
func isWeekdayInScheme(currentWeekday int8, weekdaysStr string) bool {
	if weekdaysStr == "" {
		return false
	}

	// 解析星期配置（逗号分隔，如 "1,3,5"）
	// 直接用字符串查找，避免复杂的解析
	weekdayStr := string(rune('0' + currentWeekday))

	// 检查是否包含当前星期
	// 需要确保是完整匹配，避免 "1" 匹配到 "11"
	if weekdaysStr == weekdayStr {
		return true // 只有一个星期
	}
	if len(weekdaysStr) > 1 && weekdaysStr[0:2] == weekdayStr+"," {
		return true // 在开头
	}
	if len(weekdaysStr) > 1 && weekdaysStr[len(weekdaysStr)-2:] == ","+weekdayStr {
		return true // 在结尾
	}
	// 在中间
	return len(weekdaysStr) > 2 && containsWeekday(weekdaysStr, weekdayStr)
}

// containsWeekday 检查 weekdaysStr 中是否包含指定的星期（作为独立项）
func containsWeekday(weekdaysStr, weekdayStr string) bool {
	pattern := "," + weekdayStr + ","
	return len(weekdaysStr) > 0 &&
		(weekdaysStr == weekdayStr || // 完全匹配
			weekdaysStr[:len(weekdayStr)+1] == weekdayStr+"," || // 开头
			weekdaysStr[len(weekdaysStr)-len(weekdayStr)-1:] == ","+weekdayStr || // 结尾
			len(weekdaysStr) > len(pattern)-1 && containsPattern(weekdaysStr, pattern)) // 中间
}

// containsPattern 检查字符串中是否包含模式
func containsPattern(s, pattern string) bool {
	for i := 0; i <= len(s)-len(pattern); i++ {
		if s[i:i+len(pattern)] == pattern {
			return true
		}
	}
	return false
}
