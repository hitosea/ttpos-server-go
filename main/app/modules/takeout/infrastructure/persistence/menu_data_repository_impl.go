package persistence

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	inventoryApp "ttpos-server-go/app/modules/inventory/application"
	takeoutModel "ttpos-server-go/app/modules/takeout/domain/model"
	menuRepo "ttpos-server-go/app/modules/takeout/domain/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"

	"github.com/grab/grabfood-api-sdk-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// menuDataRepositoryImpl 菜单数据仓储实现
type menuDataRepositoryImpl struct {
	dbm *database.DBManager
}

// NewMenuDataRepository 创建菜单数据仓储
func NewMenuDataRepository(dbm *database.DBManager) menuRepo.IMenuDataRepository {
	return &menuDataRepositoryImpl{
		dbm: dbm,
	}
}

// parseMenuJSON 统一解析菜单 JSON 的辅助函数
func parseMenuJSON(menuInterface interface{}) ([]byte, error) {
	// 处理 nil
	if menuInterface == nil {
		return nil, errors.New("菜单数据为空")
	}

	// 处理指针类型
	if ptr, ok := menuInterface.(*interface{}); ok && ptr != nil {
		menuInterface = *ptr
	}

	// 根据类型处理
	switch v := menuInterface.(type) {
	case []byte:
		return tryDecodeBase64(v), nil
	case string:
		return tryDecodeBase64([]byte(v)), nil
	default:
		// 其他类型，序列化为 JSON
		return json.Marshal(v)
	}
}

// tryDecodeBase64 尝试解码 Base64
func tryDecodeBase64(data []byte) []byte {
	// 移除可能的引号
	str := strings.Trim(string(data), "\"")

	// 检查是否看起来像 Base64（只包含 Base64 字符）
	if len(str) > 0 && isLikelyBase64(str) {
		// 尝试解码
		decoded, err := base64.StdEncoding.DecodeString(str)
		if err == nil && len(decoded) > 0 {
			// 检查解码后的数据是否像 JSON
			trimmed := strings.TrimSpace(string(decoded))
			if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "{") {
				return decoded
			}
		}
	}

	return data
}

// isLikelyBase64 检查字符串是否可能是 Base64 编码
func isLikelyBase64(s string) bool {
	// Base64 字符集: A-Z, a-z, 0-9, +, /, =
	// 且长度应该是 4 的倍数（允许有少量误差）
	if len(s) < 20 { // 太短不太可能是Base64编码的JSON
		return false
	}

	validChars := 0
	for _, c := range s {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=' {
			validChars++
		}
	}

	// 如果 95% 以上的字符都是 Base64 字符，认为是 Base64
	return float64(validChars)/float64(len(s)) > 0.95
}

// GetTakeoutCategories 获取外卖分类列表
func (r *menuDataRepositoryImpl) GetTakeoutCategories(ctx context.Context, companyUuid uint64, categoryIDs []uint64) ([]*model.ProductCategory, error) {
	db := ctx.GetDB()
	var categories []*model.ProductCategory

	query := db.Model(&model.ProductCategory{}).
		Where("is_display_in_takeout = ?", 1).
		Where("delete_time = ?", 0).
		Preload("MultiLanguageName", "delete_time = ?", 0).
		Order("sort ASC, id ASC")

	// 如果指定了分类ID，则只查询指定分类
	if len(categoryIDs) > 0 {
		query = query.Where("uuid IN ?", categoryIDs)
	}

	if err := query.Find(&categories).Error; err != nil {
		return nil, err
	}

	return categories, nil
}

// GetTakeoutProducts 获取指定分类下的外卖商品
func (r *menuDataRepositoryImpl) GetTakeoutProducts(ctx context.Context, companyUuid uint64, categoryUuid uint64) ([]*model.ProductPackageTakeout, error) {
	db := ctx.GetDB()
	var products []*model.ProductPackageTakeout

	err := db.Model(&model.ProductPackageTakeout{}).
		Where("category_uuid = ?", categoryUuid).
		Preload("ProductPackage", func(db *gorm.DB) *gorm.DB {
			return db.Where("delete_time = ?", 0).
				Preload("MultiLanguageName", "delete_time = ?", 0).
				Preload("DescribeMultiLanguageName", "delete_time = ?", 0).
				Preload("ProductBoms", func(db *gorm.DB) *gorm.DB {
					return db.Where("delete_time = ?", 0).
						Preload("ProductFlavor.MultiLanguageName", "delete_time = ?", 0).
						Preload("ProductSauce.MultiLanguageName", "delete_time = ?", 0)
				}).
				Preload("ProductPackageAttributeGroups", func(db *gorm.DB) *gorm.DB {
					return db.Where("delete_time = ?", 0).
						Preload("ProductAttributeGroup.MultiLanguageName", "delete_time = ?", 0).
						Preload("ProductPackageAttributes", func(db *gorm.DB) *gorm.DB {
							return db.Where("delete_time = ?", 0).
								Preload("Attribute.MultiLanguageName", "delete_time = ?", 0)
						})
				}).
				Preload("ProductPackageGroups", func(db *gorm.DB) *gorm.DB {
					return db.Where("delete_time = ?", 0).
						Preload("MultiLanguageName", "delete_time = ?", 0).
						Preload("ProductPackageGroupItems", func(db *gorm.DB) *gorm.DB {
							return db.Where("delete_time = ?", 0).
								Preload("ProductPackage", func(db *gorm.DB) *gorm.DB {
									return db.Where("delete_time = ?", 0).
										Preload("MultiLanguageName", "delete_time = ?", 0)
								}).
								Preload("ProductBom", func(db *gorm.DB) *gorm.DB {
									return db.Where("delete_time = ?", 0).
										Preload("ProductFlavor.MultiLanguageName", "delete_time = ?", 0)
								})
						})
				})
		}).
		Preload("MultiLanguageName", "delete_time = ?", 0).
		Preload("ImageFile").
		Preload("ProductBomTakeouts", func(db *gorm.DB) *gorm.DB {
			return db.Where("delete_time = ?", 0).
				Preload("ProductBom", func(db *gorm.DB) *gorm.DB {
					return db.Where("delete_time = ?", 0)
				}).
				Preload("ProductBom.ProductFlavor.MultiLanguageName", "delete_time = ?", 0).
				Preload("ProductBom.ProductSauce.MultiLanguageName", "delete_time = ?", 0)
		}).
		Preload("ProductPackageAttributeTakeouts", func(db *gorm.DB) *gorm.DB {
			return db.Where("delete_time = ?", 0).
				Preload("ProductPackageAttribute", func(db *gorm.DB) *gorm.DB {
					return db.Where("delete_time = ?", 0)
				}).
				Preload("ProductPackageAttribute.Attribute.MultiLanguageName", "delete_time = ?", 0)
		}).
		Preload("ProductPackageGroupItemTakeouts", func(db *gorm.DB) *gorm.DB {
			return db.Where("delete_time = ?", 0)
		}).
		Order("id ASC").
		Find(&products).Error

	if err != nil {
		return nil, err
	}

	// 注入库存
	r.InjectStockNum(ctx, products)

	return products, nil
}

// InjectStockNum 使用 ProductInventoryAppService 批量查询库存
// 返回库存映射表，key 为 BomUuid，value 为库存数量
func (r *menuDataRepositoryImpl) InjectStockNum(ctx context.Context, takeoutProducts []*model.ProductPackageTakeout) {
	if len(takeoutProducts) == 0 {
		return
	}

	// 使用工厂方法创建库存应用服务实例
	appService := inventoryApp.NewProductInventoryAppServiceWithDependencies(r.dbm, cache.Global)

	// Step 1: 收集所有需要查询库存的 BOM UUID 及其对应的对象引用
	bomUuids := make([]uint64, 0)
	bomMap := make(map[uint64]*model.ProductBom) // 用于快速定位 BOM 对象

	for _, takeoutProduct := range takeoutProducts {
		// 使用索引遍历以获取可修改的引用
		for i := range takeoutProduct.ProductPackage.ProductBoms {
			bom := &takeoutProduct.ProductPackage.ProductBoms[i]
			bomUuids = append(bomUuids, bom.Uuid)
			bomMap[bom.Uuid] = bom
		}
	}

	// Step 2: 如果没有需要查询的 BOM，直接返回
	if len(bomUuids) == 0 {
		return
	}

	// Step 3: 批量查询所有 BOM 的库存
	inventoryMap, err := appService.GetProductInventoriesBatch(ctx, bomUuids)
	if err != nil {
		logger.Logger.Error("批量查询商品规格/小料库存失败", zap.Error(err), zap.Int("bom_count", len(bomUuids)))
		// 查询失败，设置所有 BOM 为无限库存
		for _, bom := range bomMap {
			bom.StockNum = 99999999
		}
		return
	}

	// Step 4: 将库存值注入到对应的 BOM 对象中
	for bomUuid, bom := range bomMap {
		if inventory, ok := inventoryMap[bomUuid]; ok {
			bom.StockNum = inventory
		} else {
			// 如果某个 BOM 没有返回库存数据，设置为无限库存
			logger.Logger.Warn("未查询到商品规格/小料库存，设置为无限库存", zap.Uint64("bom_uuid", bomUuid))
			bom.StockNum = 99999999
		}
	}
}

// GetProductNameByUuid 根据商品UUID获取多语言名称
// 优先返回中文名称，如果没有则返回英文名称
func (r *menuDataRepositoryImpl) GetProductNameByUuid(ctx context.Context, productUuid uint64, productType int) (string, error) {
	db := ctx.GetDB()

	// 第一步：查询店内商品
	var productPackage model.ProductPackage
	err := db.
		Model(&model.ProductPackage{}).
		Where("uuid = ? AND delete_time = ? and product_type = ?", productUuid, 0, productType).
		Preload("MultiLanguageName", "delete_time = ?", 0).
		First(&productPackage).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.New("店内商品不存在")
		}
		return "", errors.WithMessage(err, "查询店内商品失败")
	}

	// 第二步：关联查询外卖商品
	var takeoutProduct model.ProductPackageTakeout
	err = db.
		Model(&model.ProductPackageTakeout{}).
		Where("product_package_uuid = ? AND delete_time = ?", productUuid, 0).
		Preload("MultiLanguageName", "delete_time = ?", 0).
		First(&takeoutProduct).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		return "", errors.WithMessage(err, "查询外卖商品失败")
	}

	// 最后回退到外卖商品的单语言名称
	if err == nil && takeoutProduct.Name != "" {
		return takeoutProduct.Name, nil
	}

	// 如果都没有，返回店内商品名称
	return productPackage.Name, nil
}

// GetProductNamesByUuids 批量根据商品UUID获取多语言名称
// 返回 map[productUuid]name
func (r *menuDataRepositoryImpl) GetProductNamesByUuids(ctx context.Context, productUuids []uint64, productTypes map[uint64]int) map[uint64]string {
	db := ctx.GetDB()

	if len(productUuids) == 0 {
		return make(map[uint64]string)
	}

	// 第一步：批量查询店内商品
	var productPackages []model.ProductPackage
	err := db.
		Model(&model.ProductPackage{}).
		Where("uuid IN ? AND delete_time = ?", productUuids, 0).
		Preload("MultiLanguageName", "delete_time = ?", 0).
		Find(&productPackages).Error

	if err != nil {
		return make(map[uint64]string)
	}

	// 建立 productPackage 映射
	packageMap := make(map[uint64]model.ProductPackage)
	for _, pkg := range productPackages {
		packageMap[pkg.Uuid] = pkg
	}

	// 第二步：批量查询外卖商品
	var takeoutProducts []model.ProductPackageTakeout
	err = db.
		Model(&model.ProductPackageTakeout{}).
		Where("product_package_uuid IN ? AND delete_time = ?", productUuids, 0).
		Preload("MultiLanguageName", "delete_time = ?", 0).
		Find(&takeoutProducts).Error

	if err != nil {
		return make(map[uint64]string)
	}

	// 建立 takeoutProduct 映射
	takeoutMap := make(map[uint64]model.ProductPackageTakeout)
	for _, takeout := range takeoutProducts {
		takeoutMap[takeout.ProductPackageUuid] = takeout
	}

	// 第三步：构建返回结果
	result := make(map[uint64]string)
	for _, productUuid := range productUuids {
		var name string

		// 优先使用外卖商品的多语言名称
		if takeoutProduct, ok := takeoutMap[productUuid]; ok {
			if takeoutProduct.MultiLanguageName.Uuid != 0 {
				name = takeoutProduct.MultiLanguageName.ToJson()
			}
			// 如果多语言为空，尝试外卖商品的单语言名称
			if name == "" && takeoutProduct.Name != "" {
				name = takeoutProduct.Name
			}
		}

		// 回退到店内商品的多语言名称
		if name == "" {
			if productPackage, ok := packageMap[productUuid]; ok {
				if productPackage.MultiLanguageName.Uuid != 0 {
					name = productPackage.MultiLanguageName.ToJson()
				}
				// 最后回退到店内商品名称
				if name == "" {
					name = productPackage.Name
				}
			}
		}

		result[productUuid] = name
	}

	return result
}

// GetModifierNamesByUuids 批量根据修饰符UUID和类型获取多语言名称
// modifierTypes: map[uuid]type, type 可能是 "flavor"/"sauce"/"attr"/"commodity"
// 返回 map[modifierUuid]name
func (r *menuDataRepositoryImpl) GetModifierNamesByUuids(ctx context.Context, modifierUuids []uint64, modifierTypes map[uint64]string) map[uint64]string {
	db := ctx.GetDB()

	if len(modifierUuids) == 0 {
		return make(map[uint64]string)
	}

	result := make(map[uint64]string)

	// 按类型分组
	flavorUuids := make([]uint64, 0)
	sauceUuids := make([]uint64, 0)
	attrUuids := make([]uint64, 0)
	commodityUuids := make([]uint64, 0)

	for uuid, modType := range modifierTypes {
		switch modType {
		case "flavor":
			flavorUuids = append(flavorUuids, uuid)
		case "sauce":
			sauceUuids = append(sauceUuids, uuid)
		case "attr":
			attrUuids = append(attrUuids, uuid)
		case "commodity":
			commodityUuids = append(commodityUuids, uuid)
		}
	}

	// 批量查询规格（Flavor）
	if len(flavorUuids) > 0 {
		var flavors []model.ProductBom
		err := db.
			Model(&model.ProductBom{}).
			Where("uuid IN ? AND delete_time = ?", flavorUuids, 0).
			Preload("ProductFlavor.MultiLanguageName", "delete_time = ?", 0).
			Find(&flavors).Error

		if err != nil {
			logger.Logger.Error("批量查询规格失败", zap.Error(err))
		} else {
			for _, flavor := range flavors {
				name := flavor.ProductFlavor.MultiLanguageName.ToJson()
				result[flavor.Uuid] = name
			}
		}
	}

	// 批量查询小料（Sauce）
	if len(sauceUuids) > 0 {
		var sauces []model.ProductSauce
		err := db.
			Model(&model.ProductSauce{}).
			Where("uuid IN ? AND delete_time = ?", sauceUuids, 0).
			Preload("MultiLanguageName", "delete_time = ?", 0).
			Find(&sauces).Error

		if err != nil {
			logger.Logger.Error("批量查询小料失败", zap.Error(err))
		} else {
			for _, sauce := range sauces {
				name := sauce.MultiLanguageName.ToJson()
				result[sauce.Uuid] = name
			}
		}
	}

	// 批量查询属性（Attribute）
	if len(attrUuids) > 0 {
		var attrs []model.ProductPackageAttribute
		err := db.
			Model(&model.ProductPackageAttribute{}).
			Where("uuid IN ? AND delete_time = ?", attrUuids, 0).
			Preload("Attribute.MultiLanguageName", "delete_time = ?", 0).
			Find(&attrs).Error

		if err != nil {
			logger.Logger.Error("批量查询属性失败", zap.Error(err))
		} else {
			for _, attr := range attrs {
				name := attr.Attribute.MultiLanguageName.ToJson()
				result[attr.Uuid] = name
			}
		}
	}

	// 批量查询套餐商品（Commodity = ProductPackage）
	if len(commodityUuids) > 0 {
		var packages []model.ProductPackageGroupItem
		err := db.
			Model(&model.ProductPackageGroupItem{}).
			Where("uuid IN ? AND delete_time = ?", commodityUuids, 0).
			Preload("ProductPackage.MultiLanguageName").
			Find(&packages).Error

		if err != nil {
			logger.Logger.Error("批量查询套餐商品失败", zap.Error(err))
		} else {
			for _, pkg := range packages {
				name := pkg.ProductPackage.MultiLanguageName.ToJson()
				result[pkg.Uuid] = name
			}
		}
	}

	return result
}

// convertNameToMultiLanguageJSON 将名称转换为多语言 JSON 字符串
// nameTranslation: 多语言名称对象（可能为 map[string]string）
// fallbackName: 如果多语言为空，使用此单语言名称作为回退
// 返回: 多语言格式的 JSON 字符串
func convertNameToMultiLanguageJSON(nameTranslation map[string]string, fallbackName string) string {
	var result string

	// 处理多语言名称
	if nameTranslation != nil && len(nameTranslation) > 0 {
		// 辅助函数：安全获取语言值，如果不存在则使用回退值
		getLanguageValue := func(lang string) string {
			if val, ok := nameTranslation[lang]; ok && val != "" {
				return val
			}
			// 如果当前语言不存在，尝试使用其他语言作为回退
			// 优先级：zh > en > 其他任意非空值 > fallbackName
			if zh, ok := nameTranslation["zh"]; ok && zh != "" {
				return zh
			}
			if en, ok := nameTranslation["en"]; ok && en != "" {
				return en
			}
			// 遍历找第一个非空值
			for _, v := range nameTranslation {
				if v != "" {
					return v
				}
			}
			return fallbackName
		}

		// 标准化多语言对象，确保包含所有语言字段
		singleLang := map[string]string{
			"en":   getLanguageValue("en"),
			"zh":   getLanguageValue("zh"),
			"th":   getLanguageValue("th"),
			"zhtw": getLanguageValue("zhtw"),
			"ja":   getLanguageValue("ja"),
			"ko":   getLanguageValue("ko"),
			"my":   getLanguageValue("my"),
			"tr":   getLanguageValue("tr"),
			"sv":   getLanguageValue("sv"),
		}
		if nameBytes, err := json.Marshal(singleLang); err == nil {
			result = string(nameBytes)
		}
	}

	// 如果多语言名称为空，使用单语言名称（也包装成多语言格式的 JSON）
	if result == "" && fallbackName != "" {
		singleLang := map[string]string{
			"en":   fallbackName,
			"zh":   fallbackName,
			"th":   fallbackName,
			"zhtw": fallbackName,
			"ja":   fallbackName,
			"ko":   fallbackName,
			"my":   fallbackName,
			"tr":   fallbackName,
			"sv":   fallbackName,
		}
		if nameBytes, err := json.Marshal(singleLang); err == nil {
			result = string(nameBytes)
		}
	}

	return result
}

// FetchTakeoutMenuByPlatform 根据平台查询外卖菜单记录（通用方法）
// 返回外卖菜单记录和解析后的 grabfood 菜单数据
func (r *menuDataRepositoryImpl) FetchTakeoutMenuByPlatform(ctx context.Context, platform string) (*takeoutModel.Takeout, *grabfood.GetMenuNewResponse, error) {
	db := ctx.GetDB()

	// 从数据库查询菜单 JSON
	var takeoutMenu takeoutModel.Takeout
	err := db.Model(&takeoutModel.Takeout{}).
		Where("platform = ? AND delete_time = ?", platform, 0).
		First(&takeoutMenu).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Logger.Warn("未找到外卖菜单", zap.String("platform", platform))
			return nil, nil, nil
		}
		logger.Logger.Error("查询外卖菜单失败", zap.Error(err), zap.String("platform", platform))
		return nil, nil, errors.WithMessage(err, "查询外卖菜单失败")
	}

	// 解析菜单数据为 grabfood 结构
	var grabMenu grabfood.GetMenuNewResponse
	if takeoutMenu.Menu != nil {
		// 使用统一的解析函数
		menuBytes, err := parseMenuJSON(takeoutMenu.Menu)
		if err != nil {
			logger.Logger.Error("解析菜单数据失败", zap.Error(err))
			logger.Logger.Warn("跳过菜单解析，返回空结果")
			return &takeoutMenu, nil, nil
		}

		if err := json.Unmarshal(menuBytes, &grabMenu); err != nil {
			logger.Logger.Error("解析Grab菜单失败", zap.Error(err), zap.String("type", fmt.Sprintf("%T", takeoutMenu.Menu)), zap.Int("dataLength", len(menuBytes)))
			logger.Logger.Warn("跳过菜单解析，返回空结果")
			return &takeoutMenu, nil, nil
		}
	}

	return &takeoutMenu, &grabMenu, nil
}

// GetMenuNamesByPlatformItemIds 批量根据平台商品ID获取菜单名称
// 从 ttpos_takeout 表的 menu JSON 字段中解析
// 返回 map[platformItemId]menuName
func (r *menuDataRepositoryImpl) GetMenuNamesByPlatformItemIds(ctx context.Context, platform string, platformItemIds []string) map[string]string {
	if len(platformItemIds) == 0 {
		return make(map[string]string)
	}

	result := make(map[string]string)
	missedIds := make([]string, 0)

	// 先尝试从缓存获取
	for _, itemId := range platformItemIds {
		cacheKey := "takeout:menu_name:" + platform + ":" + itemId
		cachedValue, found := cache.Global.Get(cacheKey)
		if found {
			if cachedName, ok := cachedValue.(string); ok && cachedName != "" {
				result[itemId] = cachedName
				continue
			}
		}
		missedIds = append(missedIds, itemId)
	}

	// 如果全部命中缓存，直接返回
	if len(missedIds) == 0 {
		return result
	}

	// 使用通用方法查询菜单
	_, grabMenu, err := r.FetchTakeoutMenuByPlatform(ctx, platform)
	if err != nil {
		return result
	}

	// 如果菜单为空，直接返回
	if grabMenu == nil {
		return result
	}

	// 遍历 categories -> items 查找商品名称
	for _, category := range grabMenu.GetCategories() {
		for _, item := range category.GetItems() {
			itemId := item.GetId()
			// 检查是否是需要查询的商品
			for _, missedId := range missedIds {
				if itemId == missedId {
					// 获取商品名称（存储为 JSON 字符串）
					itemName := convertNameToMultiLanguageJSON(item.GetNameTranslation(), item.GetName())
					if itemName != "" {
						result[itemId] = itemName
						// 写入缓存（30分钟过期）
						cacheKey := "takeout:menu_name:" + platform + ":" + itemId
						_ = cache.Global.Set(cacheKey, itemName, 1800)
					}
					break
				}
			}
		}
	}

	return result
}

// GetModifierNamesByPlatformIds 批量根据平台修饰符ID获取修饰符名称
// 从 ttpos_takeout 表的 menu JSON 字段中解析
// 返回 map[platformModifierId]modifierName
func (r *menuDataRepositoryImpl) GetModifierNamesByPlatformIds(ctx context.Context, platform string, platformModifierIds []string) map[string]string {
	if len(platformModifierIds) == 0 {
		return make(map[string]string)
	}

	result := make(map[string]string)
	missedIds := make([]string, 0)

	// 先尝试从缓存获取
	for _, modifierId := range platformModifierIds {
		cacheKey := "takeout:modifier_name:" + platform + ":" + modifierId
		cachedValue, found := cache.Global.Get(cacheKey)
		if found {
			if cachedName, ok := cachedValue.(string); ok && cachedName != "" {
				result[modifierId] = cachedName
				continue
			}
		}
		missedIds = append(missedIds, modifierId)
	}

	// 如果全部命中缓存，直接返回
	if len(missedIds) == 0 {
		return result
	}

	// 使用通用方法查询菜单
	_, grabMenu, err := r.FetchTakeoutMenuByPlatform(ctx, platform)
	if err != nil {
		return result
	}

	// 如果菜单为空，直接返回
	if grabMenu == nil {
		return result
	}

	// 遍历 categories -> items -> modifierGroups -> modifiers 查找修饰符名称
	for _, category := range grabMenu.GetCategories() {
		for _, item := range category.GetItems() {
			if item.HasModifierGroups() {
				for _, modifierGroup := range item.GetModifierGroups() {
					if modifierGroup.HasModifiers() {
						for _, modifier := range modifierGroup.GetModifiers() {
							modifierId := modifier.GetId()
							// 检查是否是需要查询的修饰符
							for _, missedId := range missedIds {
								if modifierId == missedId {
									// 获取修饰符名称（存储为 JSON 字符串）
									modifierName := convertNameToMultiLanguageJSON(modifier.GetNameTranslation(), modifier.GetName())
									if modifierName != "" {
										result[modifierId] = modifierName
										// 写入缓存（30分钟过期）
										cacheKey := "takeout:modifier_name:" + platform + ":" + modifierId
										_ = cache.Global.Set(cacheKey, modifierName, 1800)
									}
									break
								}
							}
						}
					}
				}
			}
		}
	}

	return result
}
