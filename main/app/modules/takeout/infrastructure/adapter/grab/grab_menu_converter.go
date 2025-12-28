package grab

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	menuRepo "ttpos-server-go/app/modules/takeout/domain/repository"
	"ttpos-server-go/app/modules/takeout/domain/value_object"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	grabfood "github.com/grab/grabfood-api-sdk-go"
)

// GrabConverter Grab 平台转换器实现
type GrabConverter struct {
	dbm      *database.DBManager
	menuRepo menuRepo.IMenuDataRepository
	cache    cache.Cache
}

// NewGrabConverter 创建 Grab 转换器
func NewGrabConverter(dbm *database.DBManager, cache cache.Cache) *GrabConverter {
	return &GrabConverter{
		dbm:      dbm,
		menuRepo: persistence.NewMenuDataRepository(dbm),
		cache:    cache,
	}
}

// GetPlatformName 获取平台名称
func (c *GrabConverter) GetPlatformName() string {
	return "Grab"
}

// ConvertToTTPOS 从 Grab 格式转换为 TTPOS 数据
// 现在直接返回 grabfood.GetMenuNewResponse，不再转换为 entity
func (c *GrabConverter) ConvertToTTPOS(ctx context.Context, platformData interface{}) (interface{}, error) {
	// 将 interface{} 转换为 GrabMenu
	var grabMenu grabfood.GetMenuNewResponse

	// 尝试类型断言
	if gm, ok := platformData.(*grabfood.GetMenuNewResponse); ok {
		return gm, nil
	} else if gm, ok := platformData.(grabfood.GetMenuNewResponse); ok {
		return &gm, nil
	}

	// 尝试 JSON 序列化/反序列化
	jsonData, err := json.Marshal(platformData)
	if err != nil {
		return nil, errors.WithMessage(err, "平台数据格式错误")
	}
	if err := json.Unmarshal(jsonData, &grabMenu); err != nil {
		return nil, errors.WithMessage(err, "解析 Grab 菜单数据失败")
	}

	// 验证数据
	if err := c.ValidateData(&grabMenu); err != nil {
		return nil, errors.WithMessage(err, "Grab 菜单数据验证失败")
	}

	return &grabMenu, nil
}

// ValidateData 验证 Grab 菜单数据格式
func (c *GrabConverter) ValidateData(platformData interface{}) error {
	// 尝试多种类型断言
	if grabMenu, ok := platformData.(*grabfood.GetMenuNewResponse); ok {
		if grabMenu.GetCurrency().Code == "" {
			return errors.New("货币代码不能为空")
		}
		if len(grabMenu.GetCategories()) == 0 {
			return errors.New("至少需要一个分类")
		}
		return nil
	}

	if grabMenu, ok := platformData.(grabfood.GetMenuNewResponse); ok {
		if grabMenu.GetCurrency().Code == "" {
			return errors.New("货币代码不能为空")
		}
		if len(grabMenu.GetCategories()) == 0 {
			return errors.New("至少需要一个分类")
		}
		return nil
	}

	// 尝试从 map 转换
	if dataMap, ok := platformData.(map[string]interface{}); ok {
		if currency, ok := dataMap["currency"].(map[string]interface{}); ok {
			if code, ok := currency["code"].(string); ok && code != "" {
				return nil
			}
		}
		return errors.New("货币代码不能为空")
	}

	return errors.New("数据格式不是 Grab 菜单格式")
}

// LoadMenuFromDatabase 从数据库加载菜单数据（辅助方法）
func (c *GrabConverter) LoadMenuFromDatabase(ctx context.Context, companyUuid uint64, currencyUnit string, categoryIDs []uint64) (*grabfood.GetMenuNewResponse, error) {
	menu := grabfood.NewGetMenuNewResponseWithDefaults()
	// 设置 PartnerMerchantID
	menu.SetPartnerMerchantID(strconv.FormatUint(companyUuid, 10))

	// 根据货币符号推断货币代码和 exponent
	currencySymbol := utils.IfString(currencyUnit == "", "฿", currencyUnit)
	currencyCode, exponent := c.getCurrencyInfoBySymbol(currencySymbol)
	menu.SetCurrency(grabfood.Currency{
		Code:     currencyCode,
		Symbol:   currencySymbol,
		Exponent: int32(exponent),
	})

	// 设置售卖时段
	grabSellingTime := grabfood.NewSellingTimeWithDefaults()
	grabSellingTime.SetId("SELLINGTIME-01")
	grabSellingTime.SetName("全天")
	grabSellingTime.SetStartTime("1000-01-01 00:00:00")
	grabSellingTime.SetEndTime("9999-12-31 23:59:59")
	grabSellingTime.SetServiceHours(grabfood.ServiceHours{
		Mon: grabfood.ServiceHour{
			OpenPeriodType: "OpenAllDay",
		},
		Tue: grabfood.ServiceHour{
			OpenPeriodType: "OpenAllDay",
		},
		Wed: grabfood.ServiceHour{
			OpenPeriodType: "OpenAllDay",
		},
		Thu: grabfood.ServiceHour{
			OpenPeriodType: "OpenAllDay",
		},
		Fri: grabfood.ServiceHour{
			OpenPeriodType: "OpenAllDay",
		},
		Sat: grabfood.ServiceHour{
			OpenPeriodType: "OpenAllDay",
		},
		Sun: grabfood.ServiceHour{
			OpenPeriodType: "OpenAllDay",
		},
	})
	menu.SetSellingTimes([]grabfood.SellingTime{
		*grabSellingTime,
	})

	// 从数据库加载实际的分类和商品数据
	categories, err := c.menuRepo.GetTakeoutCategories(ctx, companyUuid, categoryIDs)
	if err != nil {
		return nil, errors.WithMessage(err, "查询外卖分类失败")
	}

	// 转换分类和加载商品
	for idx, cat := range categories {
		category, err := c.convertTTPOSCategory(ctx, cat, idx+1)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("转换分类失败: %s", cat.Name))
		}

		// 为该分类加载商品
		if err := c.loadCategoryProducts(ctx, companyUuid, category, cat.Uuid); err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("加载分类商品失败: %s", cat.Name))
		}

		// 分类下无商品则不添加
		if len(category.Items) == 0 {
			continue
		}

		// 添加分类到菜单
		menu.SetCategories(append(menu.GetCategories(), *category))
	}

	return menu, nil
}

// convertTTPOSCategory 从 TTPOS model 转换分类
func (c *GrabConverter) convertTTPOSCategory(ctx context.Context, cat any, sequence int) (*grabfood.MenuCategory, error) {
	// 类型断言
	category, ok := cat.(*model.ProductCategory)
	if !ok {
		return nil, errors.New("类型断言失败：cat 不是 *model.ProductCategory")
	}

	// 获取分类名称（英文）
	categoryName := category.Name
	if category.MultiLanguageName.Uuid != 0 {
		categoryName = category.MultiLanguageName.GetNameByLangWithFallback("en")
	}

	// 判断分类状态：
	// Status = 1（开启）-> AVAILABLE
	// Status != 1（关闭）-> HIDE
	categoryStatus := value_object.AvailableStatusAvailable
	if category.Status != 1 {
		categoryStatus = value_object.AvailableStatusHide
	}

	categoryVO := grabfood.NewMenuCategoryWithDefaults()
	categoryVO.SetId(func() string {
		if category.SourceId != "" {
			return category.SourceId
		}
		return fmt.Sprintf("TTPOS-CAT-%d", category.Uuid)
	}())
	categoryVO.SetName(categoryName)
	categoryVO.SetSequence(int32(sequence))
	categoryVO.SetAvailableStatus(string(categoryStatus))
	categoryVO.SetSellingTimeID("SELLINGTIME-01")
	categoryVO.SetNameTranslation(c.filterSupportedLanguages(category.MultiLanguageName.ToMap()))
	return categoryVO, nil
}

// convertTTPOSProduct 从 TTPOS model 转换商品
func (c *GrabConverter) convertTTPOSProduct(ctx context.Context, pkg any, sequence int) (*grabfood.MenuItem, error) {
	// 类型断言
	takeoutProduct, ok := pkg.(*model.ProductPackageTakeout)
	if !ok {
		return nil, errors.New("类型断言失败：pkg 不是 *model.ProductPackageTakeout")
	}

	// 判断商品状态：
	// 1. 商品下架（Status=0）或删除 -> HIDE
	// 2. 商品所有规格都售罄 -> UNAVAILABLE
	// 3. 正常（Status=1）-> AVAILABLE
	status := value_object.AvailableStatusAvailable
	if takeoutProduct.IsDown() || takeoutProduct.IsDelete() {
		// 商品下架或删除，设置为 HIDE
		status = value_object.AvailableStatusHide
	}

	// 获取商品名称（优先使用外卖商品的多语言名称，否则使用店内商品名称）
	itemName := takeoutProduct.Name
	if takeoutProduct.MultiLanguageName.Uuid != 0 {
		itemName = takeoutProduct.MultiLanguageName.GetNameByLangWithFallback("en")
	} else if takeoutProduct.ProductPackage.MultiLanguageName.Uuid != 0 {
		itemName = takeoutProduct.ProductPackage.MultiLanguageName.GetNameByLangWithFallback("en")
	}

	// 先处理规格，以便计算商品价格（如果有规格，使用最小规格金额）
	// 收集所有规格用于价格计算
	flavorsForPrice := make([]*model.ProductBom, 0)
	for i := range takeoutProduct.ProductPackage.ProductBoms {
		bom := &takeoutProduct.ProductPackage.ProductBoms[i]
		if bom.IsFlavor() && !bom.IsDelete() && bom.Status == 1 {
			flavorsForPrice = append(flavorsForPrice, bom)
		}
	}

	// 创建外卖规格价格映射（bom_uuid -> 外卖价格）
	takeoutPriceMap := make(map[uint64]float64)
	for i := range takeoutProduct.ProductBomTakeouts {
		bomTakeout := &takeoutProduct.ProductBomTakeouts[i]
		if !bomTakeout.IsDelete() {
			takeoutPriceMap[bomTakeout.ProductBomUuid] = bomTakeout.Price
		}
	}

	// 计算商品价格：如果有规格，使用最小规格金额；否则使用商品原价
	var price int64
	if len(flavorsForPrice) > 0 && takeoutProduct.ProductType == 0 {
		// 找到最小规格价格（优先使用外卖价格，否则使用店内价格）
		minPrice := float64(0)
		for idx, bom := range flavorsForPrice {
			// 优先从外卖价格表获取价格
			flavorPrice := bom.Price
			if takeoutPrice, ok := takeoutPriceMap[bom.Uuid]; ok {
				flavorPrice = takeoutPrice
			}

			if idx == 0 || flavorPrice < minPrice {
				minPrice = flavorPrice
			}
		}
		price = int64(minPrice * 100) // 转换为分
	} else {
		// 没有规格，优先使用外卖商品价格，否则使用店内商品原价
		price = int64(takeoutProduct.Price * 100)
	}

	// 创建商品值对象
	menuItem := grabfood.NewMenuItemWithDefaults()
	menuItem.SetSellingTimeID("SELLINGTIME-01")
	menuItem.SetName(itemName)
	menuItem.SetSequence(int32(sequence))
	menuItem.SetAvailableStatus(string(status))
	menuItem.SetPrice(price)
	// 设置多语言名称（只包含 Grab 支持的语言）
	if takeoutProduct.MultiLanguageName.Uuid != 0 {
		menuItem.SetNameTranslation(c.filterSupportedLanguages(takeoutProduct.MultiLanguageName.ToMap()))
	} else if takeoutProduct.ProductPackage.MultiLanguageName.Uuid != 0 {
		menuItem.SetNameTranslation(c.filterSupportedLanguages(takeoutProduct.ProductPackage.MultiLanguageName.ToMap()))
	}
	// 设置商品描述（优先使用多语言字段）
	if takeoutProduct.ProductPackage.DescribeMultiLanguageName.Uuid != 0 {
		// 使用多语言描述
		menuItem.SetDescriptionTranslation(c.filterSupportedLanguages(takeoutProduct.ProductPackage.DescribeMultiLanguageName.ToMap()))
		// 设置默认描述（使用英文，如果没有则回退）
		menuItem.SetDescription(takeoutProduct.ProductPackage.DescribeMultiLanguageName.GetNameByLangWithFallback("en"))
	} else if takeoutProduct.ProductPackage.Describe != "" {
		// 回退到单语言描述字段
		menuItem.SetDescription(takeoutProduct.ProductPackage.Describe)
	}

	// 设置商品图片（使用 GetUrl 方法，如果没有本地图片则使用外部URL）
	if takeoutProduct.ImageFile.Uuid != 0 && takeoutProduct.ImageFile.FileUrl != "" {
		imageUrl := takeoutProduct.ImageFile.GetUrl(utils.GetBaseURL(nil))
		menuItem.SetPhotos([]string{imageUrl})
	} else if takeoutProduct.ProductPackage.ImageUrl != "" {
		// 如果没有本地图片文件，使用外部图片URL
		menuItem.SetPhotos([]string{takeoutProduct.ProductPackage.ImageUrl})
	}

	// 处理修饰符（Modifier）
	if takeoutProduct.ProductType != constant.ProductTypePackage {
		menuItem.SetId(func() string {
			if takeoutProduct.SourceProductId != "" {
				return takeoutProduct.SourceProductId
			}
			return fmt.Sprintf("TTPOS-ITEM-%d", takeoutProduct.ProductPackageUuid)
		}())
		// 1. 处理 ProductFlavor（规格）
		if err := c.convertProductFlavors(ctx, menuItem, takeoutProduct); err != nil {
			return nil, errors.WithMessage(err, "转换商品规格失败")
		}

		// 2. 处理 ProductSauce（小料）
		if err := c.convertProductSauces(ctx, menuItem, &takeoutProduct.ProductPackage); err != nil {
			return nil, errors.WithMessage(err, "转换商品小料失败")
		}

		// 3. 处理 ProductAttributeGroup（属性组）
		if err := c.convertProductAttributeGroups(ctx, menuItem, &takeoutProduct.ProductPackage); err != nil {
			return nil, errors.WithMessage(err, "转换商品属性组失败")
		}
	} else {
		menuItem.SetId(func() string {
			if takeoutProduct.SourceProductId != "" {
				return takeoutProduct.SourceProductId
			}
			return fmt.Sprintf("TTPOS-PACKAGE-%d", takeoutProduct.ProductPackageUuid)
		}())
		// 4. 处理套餐分组（如果是套餐）
		if err := c.convertPackageGroups(ctx, menuItem, takeoutProduct); err != nil {
			return nil, errors.WithMessage(err, "转换套餐分组失败")
		}
	}

	// 检查商品是否所有规格都售罄（仅当商品正常上架时才检查）
	if status == value_object.AvailableStatusAvailable {
		if c.isAllFlavorsSoldOut(takeoutProduct) {
			// 所有规格都售罄，商品状态设为 UNAVAILABLE
			menuItem.SetAvailableStatus(string(value_object.AvailableStatusUnavailable))
		}
	}

	return menuItem, nil
}

// filterSupportedLanguages 过滤出 Grab 支持的语言代码
// 根据 Grab 官方文档，支持的语言：en, zh, th, ms, vi, id, km, my
// 不支持：zhtw, ja, ko, tr, sv
func (c *GrabConverter) filterSupportedLanguages(translations map[string]string) map[string]string {
	supportedLangs := map[string]bool{
		"en": true, // 英文 - 所有国家
		"zh": true, // 中文 - Thailand, Singapore, Indonesia
		"th": true, // 泰语 - Thailand
		"my": true, // 缅甸语 - Myanmar
		"ms": true, // 马来语 - Malaysia
		"vi": true, // 越南语 - Vietnam
		"id": true, // 印尼语 - Indonesia
		"km": true, // 高棉语 - Cambodia
	}

	filtered := make(map[string]string)
	for lang, value := range translations {
		if supportedLangs[lang] && value != "" {
			filtered[lang] = value
		}
	}

	return filtered
}

// loadCategoryProducts 加载分类下的商品
func (c *GrabConverter) loadCategoryProducts(ctx context.Context, companyUuid uint64, category *grabfood.MenuCategory, categoryUuid uint64) error {
	// 查询外卖商品
	takeoutProducts, err := c.menuRepo.GetTakeoutProducts(ctx, companyUuid, categoryUuid)
	if err != nil {
		return errors.WithMessage(err, "查询外卖商品失败")
	}

	// 转换商品并添加到分类
	for idx, pkg := range takeoutProducts {
		menuItem, err := c.convertTTPOSProduct(ctx, pkg, idx+1)
		if err != nil {
			return errors.WithMessage(err, fmt.Sprintf("转换商品失败: %s", pkg.Name))
		}

		category.SetItems(append(category.GetItems(), *menuItem))
	}

	return nil
}

// getCurrencyInfoBySymbol 根据货币符号返回货币代码和 exponent
func (c *GrabConverter) getCurrencyInfoBySymbol(symbol string) (code string, exponent int) {
	// 货币符号到代码和 exponent 的映射
	symbolToInfo := map[string]struct {
		code     string
		exponent int
	}{
		"฿":  {"THB", 2}, // 泰铢
		"S$": {"SGD", 2}, // 新加坡元
		"RM": {"MYR", 2}, // 马来西亚林吉特
		"Rp": {"IDR", 2}, // 印尼盾
		"₫":  {"VND", 0}, // 越南盾（exponent 为 0）
		"₱":  {"PHP", 2}, // 菲律宾比索
		"៛":  {"KHR", 2}, // 柬埔寨瑞尔
		"K":  {"MMK", 2}, // 缅甸元
		"$":  {"USD", 2}, // 美元
		"￥":  {"CNY", 2}, // 人民币
		"¥":  {"CNY", 2}, // 人民币（另一种表示）
		"€":  {"EUR", 2}, // 欧元
		"£":  {"GBP", 2}, // 英镑
	}

	if info, ok := symbolToInfo[symbol]; ok {
		return info.code, info.exponent
	}

	// 默认返回泰铢
	return "THB", 2
}

// convertProductFlavors 转换商品规格为修饰符组
func (c *GrabConverter) convertProductFlavors(
	_ context.Context,
	menuItem *grabfood.MenuItem,
	takeoutProduct *model.ProductPackageTakeout,
) error {
	// 收集所有规格
	flavors := make([]*model.ProductBomTakeout, 0)
	for i := range takeoutProduct.ProductBomTakeouts {
		bomTakeout := &takeoutProduct.ProductBomTakeouts[i]
		if bomTakeout.IsDelete() || bomTakeout.GrabModifierId != "" {
			continue
		}
		flavors = append(flavors, bomTakeout)
	}

	if len(flavors) == 0 {
		return nil
	}

	// 按价格排序，确保最小价格的规格排在第一位
	sort.Slice(flavors, func(i, j int) bool {
		return flavors[i].Price < flavors[j].Price
	})

	// 找到最小规格价格，用于计算差价（优先使用外卖价格，否则使用店内价格）
	minFlavorPrice := float64(0)
	for idx, bomTakeout := range flavors {
		// 优先从外卖价格表获取价格
		price := bomTakeout.Price
		if idx == 0 || price < minFlavorPrice {
			minFlavorPrice = price
		}
	}

	// 判断规格是否必选：通常如果有多个规格，可能是可选的；如果只有一个规格，可能是必选的
	// 根据图片规则：必选时固定组(必选) min:1 max:1，可选时可选组(非必选) min:0 max:1
	minSelection := 1
	maxSelection := 1 // 规格只能选一个

	// 判断规格组状态：
	// 如果所有规格都售罄（is_sold_out = 1）-> UNAVAILABLE
	// 如果所有规格都下架 -> HIDE
	// 否则 -> AVAILABLE
	modifierGroupStatus := value_object.AvailableStatusAvailable
	allSoldOut := true
	allDown := true
	for i := range takeoutProduct.ProductBomTakeouts {
		bomTakeout := &takeoutProduct.ProductBomTakeouts[i]
		if bomTakeout.IsDelete() || bomTakeout.GrabModifierId != "" {
			continue
		}
		if bomTakeout.ProductBom.StockNum > 0 {
			allSoldOut = false
		}
		if bomTakeout.ProductBom.Status != 0 && !bomTakeout.ProductBom.IsDelete() {
			allDown = false
		}
	}
	if allDown {
		modifierGroupStatus = value_object.AvailableStatusHide
	} else if allSoldOut {
		modifierGroupStatus = value_object.AvailableStatusUnavailable
	}

	modifierGroup := grabfood.NewModifierGroupWithDefaults()
	modifierGroup.SetId(value_object.PrefixFlavorGroup + strconv.FormatUint(takeoutProduct.ProductPackageUuid, 10))
	modifierGroup.SetName("Specifications")
	modifierGroup.SetNameTranslation(map[string]string{
		"en": "Specifications",
		"zh": "规格",
		"th": "ข้อมูลจำเพาะ",
		"ms": "Saiz",
		"vi": "Kích thước",
		"id": "Ukuran",
		"km": "ទិចនាករ",
		"my": "သတ်မှတ်ချက်များ",
	})
	modifierGroup.SetSequence(1)
	modifierGroup.SetAvailableStatus(string(modifierGroupStatus))
	modifierGroup.SetSelectionRangeMin(int32(minSelection))
	modifierGroup.SetSelectionRangeMax(int32(maxSelection))

	// 转换每个规格为修饰符
	for idx, bomTakeout := range flavors {
		// 获取规格名称
		flavorName := bomTakeout.ProductBom.ProductFlavor.Name
		if bomTakeout.ProductBom.ProductFlavor.MultiLanguageName.Uuid != 0 {
			flavorName = bomTakeout.ProductBom.ProductFlavor.MultiLanguageName.GetNameByLangWithFallback("en")
		}

		// 获取规格价格（优先使用外卖价格，否则使用店内价格）
		flavorPrice := bomTakeout.Price

		// 计算规格价格：与最小规格的差价（当前规格价格 - 最小规格价格）
		priceDiff := flavorPrice - minFlavorPrice
		priceInCents := int64(priceDiff * 100) // 转换为分

		// 判断规格状态：
		// 1. 售罄（is_sold_out = 1）-> UNAVAILABLE
		// 2. 下架（Status = 0）或删除 -> HIDE
		// 3. 正常 -> AVAILABLE
		modifierStatus := value_object.AvailableStatusAvailable
		if bomTakeout.ProductBom.StockNum <= 0 {
			// 规格售罄
			modifierStatus = value_object.AvailableStatusUnavailable
		}
		if bomTakeout.ProductBom.Status == 0 || bomTakeout.ProductBom.IsDelete() {
			modifierStatus = value_object.AvailableStatusHide
		}

		// 创建修饰符
		modifier := grabfood.NewMenuModifierWithDefaults()
		modifier.SetId(func() string {
			if bomTakeout.GrabModifierId != "" {
				return bomTakeout.GrabModifierId
			}
			return value_object.PrefixFlavor + strconv.FormatUint(bomTakeout.ProductBomUuid, 10)
		}())
		modifier.SetName(flavorName)
		modifier.SetSequence(int32(idx + 1))
		modifier.SetAvailableStatus(string(modifierStatus))
		modifier.SetPrice(priceInCents)
		if bomTakeout.ProductBom.ProductFlavor.MultiLanguageName.Uuid != 0 {
			modifier.SetNameTranslation(c.filterSupportedLanguages(bomTakeout.ProductBom.ProductFlavor.MultiLanguageName.ToMap()))
		}
		modifierGroup.SetModifiers(append(modifierGroup.GetModifiers(), *modifier))

	}

	menuItem.SetModifierGroups(append(menuItem.GetModifierGroups(), *modifierGroup))
	return nil
}

// convertProductSauces 转换商品小料为修饰符组
func (c *GrabConverter) convertProductSauces(ctx context.Context, menuItem *grabfood.MenuItem, productPackage *model.ProductPackage) error {
	// 收集所有小料
	sauces := make([]*model.ProductBom, 0)
	for i := range productPackage.ProductBoms {
		bom := &productPackage.ProductBoms[i]
		if bom.IsSauce() && !bom.IsDelete() && bom.Status == 1 {
			sauces = append(sauces, bom)
		}
	}

	if len(sauces) == 0 {
		return nil
	}

	// 根据 ProductPackage 的配置判断小料是否必选
	maxSelection := int(productPackage.SauceMaxSelection)
	if maxSelection == 0 {
		maxSelection = len(sauces) // 如果未配置，使用小料总数
	}
	// 确保最大选择数量至少为 1
	if maxSelection < 1 {
		maxSelection = 1
	}

	// 判断小料组状态：
	// 如果所有小料都售罄（is_sold_out = 1 或 StockNum <= 0）-> UNAVAILABLE
	// 如果所有小料都下架 -> HIDE
	// 否则 -> AVAILABLE
	modifierGroupStatus := value_object.AvailableStatusAvailable
	allSoldOut := true
	allDown := true
	for i := range productPackage.ProductBoms {
		bom := &productPackage.ProductBoms[i]
		if !bom.IsSauce() {
			continue
		}
		if bom.StockNum > 0 {
			allSoldOut = false
		}
		if !bom.IsDown() {
			allDown = false
		}
	}
	if allDown {
		modifierGroupStatus = value_object.AvailableStatusHide
	} else if allSoldOut {
		modifierGroupStatus = value_object.AvailableStatusUnavailable
	}

	modifierGroup := grabfood.NewModifierGroupWithDefaults()
	modifierGroup.SetId(fmt.Sprintf("TTPOS-SAUCE-GROUP-%d", productPackage.Uuid))
	modifierGroup.SetName("Add Toppings")
	modifierGroup.SetNameTranslation(map[string]string{
		"en": "Add Toppings",
		"zh": "加料",
		"th": "เพิ่มเครื่อง",
		"ms": "Tambahkan Topping",
		"vi": "Thêm gia vị",
		"id": "Tambahkan Topping",
		"km": "ការបង្ហាញ",
		"my": "အပိုထည့်ခြင်း",
	})
	modifierGroup.SetSequence(2)
	modifierGroup.SetAvailableStatus(string(modifierGroupStatus))
	modifierGroup.SetSelectionRangeMin(int32(productPackage.SauceMinSelection))
	modifierGroup.SetSelectionRangeMax(int32(maxSelection))

	// 转换每个小料为修饰符
	for idx, bom := range sauces {
		// 获取小料名称
		sauceName := bom.ProductSauce.Name
		if bom.ProductSauce.MultiLanguageName.Uuid != 0 {
			sauceName = bom.ProductSauce.MultiLanguageName.GetNameByLangWithFallback("en")
		}

		// 判断小料状态：
		// 1. 售罄（is_sold_out = 1 或 StockNum <= 0）-> UNAVAILABLE
		// 2. 下架（Status = 0）或删除 -> HIDE
		// 3. 正常 -> AVAILABLE
		modifierStatus := value_object.AvailableStatusAvailable
		if bom.StockNum <= 0 {
			// 小料售罄
			modifierStatus = value_object.AvailableStatusUnavailable
		} else if bom.IsDown() {
			// 小料下架或删除
			modifierStatus = value_object.AvailableStatusHide
		}

		modifier := grabfood.NewMenuModifierWithDefaults()
		modifier.SetId(value_object.PrefixSauce + strconv.FormatUint(bom.Uuid, 10))
		modifier.SetName(sauceName)
		modifier.SetSequence(int32(idx + 1))
		modifier.SetAvailableStatus(string(modifierStatus))
		modifier.SetPrice(int64(bom.Price * 100)) // 转换为分
		if bom.ProductSauce.MultiLanguageName.Uuid != 0 {
			modifier.SetNameTranslation(c.filterSupportedLanguages(bom.ProductSauce.MultiLanguageName.ToMap()))
		}
		modifierGroup.SetModifiers(append(modifierGroup.GetModifiers(), *modifier))
	}

	menuItem.SetModifierGroups(append(menuItem.GetModifierGroups(), *modifierGroup))
	return nil
}

// convertProductAttributeGroups 转换商品属性组为修饰符组
func (c *GrabConverter) convertProductAttributeGroups(ctx context.Context, menuItem *grabfood.MenuItem, productPackage *model.ProductPackage) error {
	// 遍历所有属性组
	sequence := 3 // 从3开始，因为规格是1，小料是2
	for _, packageAttrGroup := range productPackage.ProductPackageAttributeGroups {
		if packageAttrGroup.IsDelete() {
			continue
		}

		// 获取属性组名称
		groupName := packageAttrGroup.ProductAttributeGroup.Name
		if packageAttrGroup.ProductAttributeGroup.MultiLanguageName.Uuid != 0 {
			groupName = packageAttrGroup.ProductAttributeGroup.MultiLanguageName.GetNameByLangWithFallback("en")
		}

		// 计算最大选择数量，确保至少为 1
		maxSelection := int(packageAttrGroup.MaxSelection)
		if maxSelection < 1 {
			// 如果未配置或为 0，默认使用属性数量，但至少为 1
			maxSelection = len(packageAttrGroup.ProductPackageAttributes)
			if maxSelection < 1 {
				maxSelection = 1
			}
		}

		// 创建修饰符组
		modifierGroup := grabfood.NewModifierGroupWithDefaults()
		modifierGroup.SetId(func() string {
			if packageAttrGroup.ProductAttributeGroup.SourceId != "" {
				return packageAttrGroup.ProductAttributeGroup.SourceId
			}
			return value_object.PrefixAttrGroup + strconv.FormatUint(packageAttrGroup.ProductAttributeGroup.Uuid, 10)
		}())
		modifierGroup.SetName(groupName)
		modifierGroup.SetNameTranslation(c.filterSupportedLanguages(packageAttrGroup.ProductAttributeGroup.MultiLanguageName.ToMap()))
		modifierGroup.SetSequence(int32(sequence))
		modifierGroup.SetAvailableStatus(string(value_object.AvailableStatusAvailable))
		modifierGroup.SetSelectionRangeMin(int32(packageAttrGroup.MinSelection))
		modifierGroup.SetSelectionRangeMax(int32(maxSelection))

		// 转换每个属性为修饰符
		for idx, packageAttr := range packageAttrGroup.ProductPackageAttributes {
			if packageAttr.IsDelete() {
				continue
			}

			// 获取属性名称
			attrName := packageAttr.Attribute.Name
			if packageAttr.Attribute.MultiLanguageName.Uuid != 0 {
				attrName = packageAttr.Attribute.MultiLanguageName.GetNameByLangWithFallback("en")
			}

			modifier := grabfood.NewMenuModifierWithDefaults()
			modifier.SetId(func() string {
				if packageAttr.Attribute.SourceId != "" {
					return packageAttr.Attribute.SourceId
				}
				return value_object.PrefixAttr + strconv.FormatUint(packageAttr.Uuid, 10)
			}())
			modifier.SetName(attrName)
			modifier.SetSequence(int32(idx + 1))
			modifier.SetAvailableStatus(string(value_object.AvailableStatusAvailable))
			modifier.SetPrice(0)
			if packageAttr.Attribute.MultiLanguageName.Uuid != 0 {
				modifier.SetNameTranslation(c.filterSupportedLanguages(packageAttr.Attribute.MultiLanguageName.ToMap()))
			}
			modifierGroup.SetModifiers(append(modifierGroup.GetModifiers(), *modifier))
		}

		// 只有包含修饰符的组才添加
		if len(modifierGroup.Modifiers) > 0 {
			menuItem.SetModifierGroups(append(menuItem.GetModifierGroups(), *modifierGroup))
			sequence++
		}
	}

	return nil
}

// convertPackageGroups 转换套餐分组为修饰符组
func (c *GrabConverter) convertPackageGroups(ctx context.Context, menuItem *grabfood.MenuItem, takeoutProduct *model.ProductPackageTakeout) error {
	// 检查是否有外卖套餐子商品配置
	if len(takeoutProduct.ProductPackageGroupItemTakeouts) == 0 {
		return nil
	}

	// 按分组聚合套餐子商品
	// key: ProductPackageGroup.Uuid, value: 该分组下的所有 ProductPackageGroupItemTakeout
	groupMap := make(map[uint64][]*model.ProductPackageGroupItemTakeout)
	groupInfoMap := make(map[uint64]*model.ProductPackageGroup) // 存储分组信息

	for i := range takeoutProduct.ProductPackageGroupItemTakeouts {
		itemTakeout := &takeoutProduct.ProductPackageGroupItemTakeouts[i]
		if itemTakeout.DeleteTime != 0 {
			continue
		}
		// 获取关联的 ProductPackageGroupItem
		if itemTakeout.ProductPackageGroupItemUuid == 0 {
			continue
		}
		groupItem := itemTakeout.ProductPackageGroupItem

		// 获取关联的 ProductPackageGroup
		if groupItem.ProductPackageGroup == nil || groupItem.ProductPackageGroup.IsDelete() {
			continue
		}
		packageGroup := groupItem.ProductPackageGroup

		// 检查子商品的 ProductPackage 是否有效
		if groupItem.ProductPackage == nil || groupItem.ProductPackage.IsDelete() {
			continue
		}

		// 聚合到分组
		groupUuid := packageGroup.Uuid
		groupMap[groupUuid] = append(groupMap[groupUuid], itemTakeout)
		if _, exists := groupInfoMap[groupUuid]; !exists {
			groupInfoMap[groupUuid] = packageGroup
		}
	}

	// 如果没有有效的分组，直接返回
	if len(groupMap) == 0 {
		return nil
	}

	// 将分组按 Sort 排序
	type groupWithSort struct {
		group *model.ProductPackageGroup
		items []*model.ProductPackageGroupItemTakeout
	}
	sortedGroups := make([]groupWithSort, 0, len(groupMap))
	for groupUuid, items := range groupMap {
		sortedGroups = append(sortedGroups, groupWithSort{
			group: groupInfoMap[groupUuid],
			items: items,
		})
	}
	sort.Slice(sortedGroups, func(i, j int) bool {
		if sortedGroups[i].group.Sort != sortedGroups[j].group.Sort {
			return sortedGroups[i].group.Sort < sortedGroups[j].group.Sort
		}
		return sortedGroups[i].group.Uuid < sortedGroups[j].group.Uuid
	})

	// sequence 从 4 开始（规格=1，小料=2，属性组=3）
	sequence := 4

	// 遍历所有套餐分组
	for _, groupData := range sortedGroups {
		packageGroup := groupData.group
		itemTakeouts := groupData.items

		// 按 ProductPackageGroupItem.Sort 排序子商品
		sort.Slice(itemTakeouts, func(i, j int) bool {
			itemI := itemTakeouts[i].ProductPackageGroupItem
			itemJ := itemTakeouts[j].ProductPackageGroupItem
			if itemI.Sort != itemJ.Sort {
				return itemI.Sort < itemJ.Sort
			}
			return itemI.Uuid < itemJ.Uuid
		})

		// 获取分组名称
		groupName := packageGroup.Name
		if packageGroup.MultiLanguageName.Uuid != 0 {
			groupName = packageGroup.MultiLanguageName.GetNameByLangWithFallback("en")
		}

		// 根据 GroupType 设置选择范围
		// GroupType=0: 固定组（必选），必须选择所有子商品，min=子商品数量，max=子商品数量
		// GroupType=1: 可选组，最小选择=0，最大选择=OptionalCount（本组可选数量）
		var minSelection, maxSelection int
		if packageGroup.GroupType == 0 {
			// 固定组：必须选择所有子商品
			minSelection = len(itemTakeouts)
			maxSelection = len(itemTakeouts)
		} else {
			// 可选组：最小选择=0，最大选择=OptionalCount
			maxSelection = packageGroup.OptionalCount
			// 如果 OptionalCount 为 0 或未配置，使用子商品数量
			if maxSelection == 0 {
				maxSelection = len(itemTakeouts)
			}
			minSelection = packageGroup.OptionalMinCount
		}
		// 确保最大选择数量至少为 1
		if maxSelection < 1 {
			maxSelection = 1
		}
		// 确保最小选择数量不超过最大选择数量
		if minSelection > maxSelection {
			minSelection = maxSelection
		}

		// 获取分组的多语言名称
		nameTranslation := make(map[string]string)
		if packageGroup.MultiLanguageName.Uuid != 0 {
			nameTranslation = c.filterSupportedLanguages(packageGroup.MultiLanguageName.ToMap())
		}

		// 创建修饰符组
		modifierGroup := grabfood.NewModifierGroupWithDefaults()
		modifierGroup.SetId(value_object.PrefixPackageGroup + strconv.FormatUint(packageGroup.Uuid, 10))
		modifierGroup.SetName(groupName)
		modifierGroup.SetNameTranslation(nameTranslation)
		modifierGroup.SetSequence(int32(sequence))
		modifierGroup.SetAvailableStatus(string(value_object.AvailableStatusAvailable))
		modifierGroup.SetSelectionRangeMin(int32(minSelection))
		modifierGroup.SetSelectionRangeMax(int32(maxSelection))

		// 转换每个子商品为修饰符
		for idx, itemTakeout := range itemTakeouts {
			groupItem := itemTakeout.ProductPackageGroupItem

			// 获取子商品名称
			itemName := groupItem.ProductPackage.Name
			if groupItem.ProductPackage.MultiLanguageName.Uuid != 0 {
				itemName = groupItem.ProductPackage.MultiLanguageName.GetNameByLangWithFallback("en")
			}

			// 在商品名称后面加上数量
			if groupItem.Num > 0 {
				// 如果数量是整数，显示为整数；否则显示小数
				if groupItem.Num == float64(int64(groupItem.Num)) {
					itemName = fmt.Sprintf("%s * %d", itemName, int64(groupItem.Num))
				} else {
					itemName = fmt.Sprintf("%s * %.2f", itemName, groupItem.Num)
				}
			}

			// 使用外卖平台的加价（从 ProductPackageGroupItemTakeout.AddPrice）
			priceInCents := int64(itemTakeout.AddPrice * 100)

			// 创建修饰符
			modifier := grabfood.NewMenuModifierWithDefaults()
			modifier.SetId(value_object.PrefixPackageItem + strconv.FormatUint(groupItem.Uuid, 10))
			modifier.SetName(itemName)
			modifier.SetSequence(int32(idx + 1))
			modifier.SetAvailableStatus(string(value_object.AvailableStatusAvailable))
			modifier.SetPrice(priceInCents)
			modifier.AdditionalProperties = map[string]interface{}{
				"num": groupItem.Num,
			}

			// 设置多语言名称（也需要加上数量）
			if groupItem.ProductPackage.MultiLanguageName.Uuid != 0 {
				translations := c.filterSupportedLanguages(groupItem.ProductPackage.MultiLanguageName.ToMap())
				// 为每个语言版本添加数量后缀
				for lang, name := range translations {
					if groupItem.Num > 0 {
						if groupItem.Num == float64(int64(groupItem.Num)) {
							translations[lang] = fmt.Sprintf("%s * %d", name, int64(groupItem.Num))
						} else {
							translations[lang] = fmt.Sprintf("%s * %.2f", name, groupItem.Num)
						}
					}
				}
				modifier.SetNameTranslation(translations)
			}

			modifierGroup.SetModifiers(append(modifierGroup.GetModifiers(), *modifier))
		}

		// 只有包含修饰符的组才添加
		if len(modifierGroup.Modifiers) > 0 {
			menuItem.SetModifierGroups(append(menuItem.GetModifierGroups(), *modifierGroup))
			sequence++
		}
	}

	return nil
}

// isAllFlavorsSoldOut 检查商品是否所有规格都售罄
// 返回 true 表示所有规格都售罄，false 表示至少有一个规格可用
// isAllFlavorsSoldOut 检查商品的所有规格是否都不可用（售罄或下架）
func (c *GrabConverter) isAllFlavorsSoldOut(takeoutProduct *model.ProductPackageTakeout) bool {
	// 如果不是套餐，返回 false（只有套餐才需要检查子商品）
	if takeoutProduct.ProductType == constant.ProductTypePackage {
		// 如果是套餐，则判断套餐子商品是否都售罄
		if len(takeoutProduct.ProductPackageGroupItemTakeouts) == 0 {
			// 套餐没有子商品，返回 true
			return true
		}
		// 检查所有子商品是否都售罄或下架
		hasAvailableSubProduct := false
		for i := range takeoutProduct.ProductPackageGroupItemTakeouts {
			subItem := &takeoutProduct.ProductPackageGroupItemTakeouts[i]
			if subItem.ProductPackageGroupItem.ProductBom == nil {
				continue
			}
			bom := subItem.ProductPackageGroupItem.ProductBom
			if !bom.IsFlavor() {
				continue
			}
			// 如果有任何一个子商品售罄，则套餐售罄
			if bom.StockNum <= 0 {
				hasAvailableSubProduct = true
				break
			}
		}
		return hasAvailableSubProduct
	} else {
		// 如果没有规格，商品本身就不能售罄
		if len(takeoutProduct.ProductPackage.ProductBoms) == 0 {
			return false
		}
		// 收集所有规格
		hasFlavor := false
		allUnavailable := true
		for i := range takeoutProduct.ProductPackage.ProductBoms {
			bom := &takeoutProduct.ProductPackage.ProductBoms[i]
			// 只检查规格（Flavor）
			if !bom.IsFlavor() {
				continue
			}
			hasFlavor = true
			// 检查规格是否可用：
			// 1. 已下架（Status == 0）
			// 2. 售罄状态（IsSoldOut == 1）
			// 3. 库存为0（StockNum <= 0）
			isOffline := bom.Status == 0 || bom.IsDelete()
			isSoldOut := bom.StockNum <= 0
			// 如果有任何一个规格可用（上架且未售罄），则商品可用
			if !isOffline && !isSoldOut {
				allUnavailable = false
				break
			}
		}
		// 如果没有规格，返回 false
		if !hasFlavor {
			return false
		}
		return allUnavailable
	}
}

// ParseGrabMenu 解析 Grab 菜单数据
// 支持从字符串、字节数组或对象解析为 GrabMenu 结构
func (c *GrabConverter) ParseGrabMenu(menuData interface{}) (*grabfood.GetMenuNewResponse, error) {
	var menuJSON []byte
	var err error

	// 判断 menuData 的类型
	switch v := menuData.(type) {
	case string:
		// 如果已经是 JSON 字符串，直接使用
		menuJSON = []byte(v)
	case []byte:
		// 如果是字节数组，直接使用
		menuJSON = v
	case *grabfood.GetMenuNewResponse:
		// 如果已经是 GetMenuNewResponse 对象，直接返回
		return v, nil
	case grabfood.GetMenuNewResponse:
		// 如果是值类型，返回指针
		return &v, nil
	default:
		// 如果是其他对象类型，需要序列化
		menuJSON, err = json.Marshal(menuData)
		if err != nil {
			return nil, fmt.Errorf("序列化菜单数据失败: %w", err)
		}
	}

	// 尝试检测并处理 Base64 编码的数据
	menuJSON = c.tryDecodeBase64(menuJSON)

	var menu grabfood.GetMenuNewResponse
	if err := json.Unmarshal(menuJSON, &menu); err != nil {
		// 输出前100个字符用于调试
		preview := string(menuJSON)
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		return nil, fmt.Errorf("反序列化菜单数据失败 (preview: %s, length: %d): %w", preview, len(menuJSON), err)
	}

	return &menu, nil
}

// tryDecodeBase64 尝试检测并解码 Base64 数据
func (c *GrabConverter) tryDecodeBase64(data []byte) []byte {
	str := string(data)

	// 移除可能的引号
	str = strings.Trim(str, "\"")

	// 检查是否看起来像 Base64（只包含 Base64 字符）
	if len(str) > 0 && isLikelyBase64(str) {
		// 尝试解码
		decoded, err := base64.StdEncoding.DecodeString(str)
		if err == nil && len(decoded) > 0 {
			// 检查解码后的数据是否像 JSON
			trimmed := strings.TrimSpace(string(decoded))
			if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
				return decoded
			}
		}
	}

	return data
}

// isLikelyBase64 检查字符串是否可能是 Base64 编码
func isLikelyBase64(s string) bool {
	// Base64 字符集: A-Z, a-z, 0-9, +, /, =
	// 如果字符串很短或者包含大量非 Base64 字符，则不太可能是 Base64
	if len(s) < 20 {
		return false
	}

	// 检查是否包含典型的 JSON 字符（如果有，就不是 Base64）
	if strings.Contains(s, "{") || strings.Contains(s, "}") ||
		strings.Contains(s, "[") || strings.Contains(s, "]") ||
		strings.Contains(s, ":") || strings.Contains(s, ",") {
		return false
	}

	// 简单检查：Base64 字符串通常不包含空格和特殊字符（除了 +/=）
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '+' || c == '/' || c == '=') {
			return false
		}
	}

	return true
}
