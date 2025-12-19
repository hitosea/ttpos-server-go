package grab

import (
	"encoding/json"
	"fmt"
	"sort"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/takeout/domain/menu/entity"
	menuRepo "ttpos-server-go/app/modules/takeout/domain/menu/repository"
	"ttpos-server-go/app/modules/takeout/domain/menu/valueobject"
	"ttpos-server-go/app/modules/takeout/domain/service"
	"ttpos-server-go/app/modules/takeout/infrastructure/persistence"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"
)

// GrabConverter Grab 平台转换器实现
type GrabConverter struct {
	dbm      *database.DBManager
	menuRepo menuRepo.IMenuDataRepository
	cache    cache.Cache
}

// NewGrabConverter 创建 Grab 转换器
func NewGrabConverter(dbm *database.DBManager, cache cache.Cache) service.IPlatformConverter {
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

// ConvertFromTTPOS 从 TTPOS 数据转换为 Grab 格式
func (c *GrabConverter) ConvertFromTTPOS(ctx context.Context, menu *entity.TakeoutMenu) (interface{}, error) {
	if menu == nil {
		return nil, errors.New("菜单数据不能为空")
	}

	// 验证菜单
	if err := menu.Validate(); err != nil {
		return nil, errors.WithMessage(err, "菜单验证失败")
	}

	// 转换货币
	grabCurrency := GrabCurrency{
		Code:     menu.Currency.Code,
		Symbol:   menu.Currency.Symbol,
		Exponent: menu.Currency.Exponent,
	}

	// 转换售卖时段
	grabSellingTimes := make([]GrabSellingTime, 0, len(menu.SellingTimes))
	for _, st := range menu.SellingTimes {
		grabST := c.convertSellingTime(st)
		grabSellingTimes = append(grabSellingTimes, grabST)
	}

	// 转换分类
	grabCategories := make([]GrabCategory, 0, len(menu.Categories))
	for _, cat := range menu.Categories {
		grabCat := c.convertCategory(cat)
		grabCategories = append(grabCategories, grabCat)
	}

	grabMenu := &GrabMenu{
		Currency:     grabCurrency,
		SellingTimes: grabSellingTimes,
		Categories:   grabCategories,
	}

	return grabMenu, nil
}

// convertSellingTime 转换售卖时段
func (c *GrabConverter) convertSellingTime(st *valueobject.SellingTime) GrabSellingTime {
	// 默认全天营业配置
	defaultDayHours := GrabDayServiceHours{
		OpenPeriodType: "OpenPeriod",
		Periods: []GrabServicePeriod{
			{
				StartTime: "00:00",
				EndTime:   "23:59",
			},
		},
	}

	return GrabSellingTime{
		ID:       st.ID,
		Name:     st.Name,
		Sequence: st.Sequence,
		ServiceHours: GrabServiceHours{
			Mon: defaultDayHours,
			Tue: defaultDayHours,
			Wed: defaultDayHours,
			Thu: defaultDayHours,
			Fri: defaultDayHours,
			Sat: defaultDayHours,
			Sun: defaultDayHours,
		},
		StartTime: st.StartTime,
		EndTime:   st.EndTime,
	}
}

// convertCategory 转换分类
func (c *GrabConverter) convertCategory(cat *valueobject.Category) GrabCategory {
	grabItems := make([]GrabItem, 0, len(cat.Items))
	for _, item := range cat.Items {
		grabItem := c.convertItem(item)
		grabItems = append(grabItems, grabItem)
	}

	return GrabCategory{
		ID:              cat.ID,
		Name:            cat.Name,
		NameTranslation: cat.NameTranslation,
		Sequence:        cat.Sequence,
		AvailableStatus: string(cat.AvailableStatus),
		Items:           grabItems,
		SellingTimeID:   cat.SellingTimeID,
	}
}

// convertItem 转换商品
func (c *GrabConverter) convertItem(item *valueobject.MenuItem) GrabItem {
	grabModifierGroups := make([]GrabModifierGroup, 0, len(item.ModifierGroups))
	for _, mg := range item.ModifierGroups {
		grabMG := c.convertModifierGroup(mg)
		grabModifierGroups = append(grabModifierGroups, grabMG)
	}

	grabItem := GrabItem{
		ID:                     item.ID,
		Name:                   item.Name,
		NameTranslation:        item.NameTranslation,
		Sequence:               item.Sequence,
		AvailableStatus:        string(item.AvailableStatus),
		Price:                  item.Price,
		Description:            item.Description,
		DescriptionTranslation: item.DescriptionTranslation,
		Photos:                 item.Photos,
		ModifierGroups:         grabModifierGroups,
		SellingTimeID:          item.SellingTimeID,
	}

	// 转换营销活动信息
	if item.CampaignInfo != nil {
		grabItem.CampaignInfo = &GrabCampaignInfo{
			OriginalPrice: item.CampaignInfo.OriginalPrice,
			DiscountType:  item.CampaignInfo.DiscountType,
			DiscountValue: item.CampaignInfo.DiscountValue,
		}
	}

	return grabItem
}

// convertModifierGroup 转换修饰符组
func (c *GrabConverter) convertModifierGroup(mg *valueobject.ModifierGroup) GrabModifierGroup {
	grabModifiers := make([]GrabModifier, 0, len(mg.Modifiers))
	for _, m := range mg.Modifiers {
		grabM := GrabModifier{
			ID:              m.ID,
			Name:            m.Name,
			NameTranslation: m.NameTranslation,
			Sequence:        m.Sequence,
			AvailableStatus: string(m.AvailableStatus),
			Price:           m.Price,
		}
		grabModifiers = append(grabModifiers, grabM)
	}

	return GrabModifierGroup{
		ID:                mg.ID,
		Name:              mg.Name,
		NameTranslation:   mg.NameTranslation,
		Sequence:          mg.Sequence,
		AvailableStatus:   string(mg.AvailableStatus),
		SelectionRangeMin: mg.SelectionRangeMin,
		SelectionRangeMax: mg.SelectionRangeMax,
		Modifiers:         grabModifiers,
	}
}

// ConvertToTTPOS 从 Grab 格式转换为 TTPOS 数据
func (c *GrabConverter) ConvertToTTPOS(ctx context.Context, platformData interface{}) (*entity.TakeoutMenu, error) {
	// 将 interface{} 转换为 GrabMenu
	var grabMenu GrabMenu

	// 尝试类型断言
	if gm, ok := platformData.(*GrabMenu); ok {
		grabMenu = *gm
	} else if gm, ok := platformData.(GrabMenu); ok {
		grabMenu = gm
	} else {
		// 尝试 JSON 序列化/反序列化
		jsonData, err := json.Marshal(platformData)
		if err != nil {
			return nil, errors.WithMessage(err, "平台数据格式错误")
		}
		if err := json.Unmarshal(jsonData, &grabMenu); err != nil {
			return nil, errors.WithMessage(err, "解析 Grab 菜单数据失败")
		}
	}

	// 验证数据
	if err := c.ValidateData(&grabMenu); err != nil {
		return nil, errors.WithMessage(err, "Grab 菜单数据验证失败")
	}

	// 获取门店和公司信息
	company := ctx.GetCompany()

	// 转换货币
	currency, err := valueobject.NewCurrency(
		grabMenu.Currency.Code,
		grabMenu.Currency.Symbol,
		grabMenu.Currency.Exponent,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "创建货币对象失败")
	}

	// 创建菜单聚合根
	menu, err := entity.NewTakeoutMenu(company.Uuid)
	if err != nil {
		return nil, errors.WithMessage(err, "创建菜单聚合根失败")
	}

	// 设置货币
	if err := menu.SetCurrency(currency.Code, currency.Symbol, currency.Exponent); err != nil {
		return nil, errors.WithMessage(err, "设置货币失败")
	}

	// 转换售卖时段
	for _, gst := range grabMenu.SellingTimes {
		st, err := c.convertFromGrabSellingTime(&gst)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("转换售卖时段失败: %s", gst.Name))
		}
		if err := menu.AddSellingTime(st); err != nil {
			return nil, errors.WithMessage(err, "添加售卖时段失败")
		}
	}

	// 转换分类
	for _, gcat := range grabMenu.Categories {
		cat, err := c.convertFromGrabCategory(&gcat)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("转换分类失败: %s", gcat.Name))
		}
		if err := menu.AddCategory(cat); err != nil {
			return nil, errors.WithMessage(err, "添加分类失败")
		}
	}

	return menu, nil
}

// convertFromGrabSellingTime 从 Grab 格式转换售卖时段
func (c *GrabConverter) convertFromGrabSellingTime(gst *GrabSellingTime) (*valueobject.SellingTime, error) {
	return valueobject.NewSellingTime(
		gst.ID,
		gst.Name,
		gst.Sequence,
		gst.StartTime,
		gst.EndTime,
	)
}

// convertFromGrabCategory 从 Grab 格式转换分类
func (c *GrabConverter) convertFromGrabCategory(gcat *GrabCategory) (*valueobject.Category, error) {
	status := valueobject.AvailableStatusAvailable
	if gcat.AvailableStatus == "UNAVAILABLE" {
		status = valueobject.AvailableStatusUnavailable
	}

	cat, err := valueobject.NewCategory(gcat.ID, gcat.Name, gcat.Sequence, status)
	if err != nil {
		return nil, err
	}

	cat.SellingTimeID = gcat.SellingTimeID
	cat.NameTranslation = gcat.NameTranslation

	// 转换商品
	for _, gitem := range gcat.Items {
		item, err := c.convertFromGrabItem(&gitem)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("转换商品失败: %s", gitem.Name))
		}
		cat.AddItem(item)
	}

	return cat, nil
}

// convertFromGrabItem 从 Grab 格式转换商品
func (c *GrabConverter) convertFromGrabItem(gitem *GrabItem) (*valueobject.MenuItem, error) {
	status := valueobject.AvailableStatusAvailable
	if gitem.AvailableStatus == "UNAVAILABLE" {
		status = valueobject.AvailableStatusUnavailable
	}

	item, err := valueobject.NewMenuItem(gitem.ID, gitem.Name, gitem.Sequence, status, gitem.Price)
	if err != nil {
		return nil, err
	}

	item.Description = gitem.Description
	item.DescriptionTranslation = gitem.DescriptionTranslation
	item.Photos = gitem.Photos
	item.SellingTimeID = gitem.SellingTimeID
	item.NameTranslation = gitem.NameTranslation

	// 转换营销活动信息
	if gitem.CampaignInfo != nil {
		item.CampaignInfo = &valueobject.CampaignInfo{
			OriginalPrice: gitem.CampaignInfo.OriginalPrice,
			DiscountType:  gitem.CampaignInfo.DiscountType,
			DiscountValue: gitem.CampaignInfo.DiscountValue,
		}
	}

	// 转换修饰符组
	for _, gmg := range gitem.ModifierGroups {
		mg, err := c.convertFromGrabModifierGroup(&gmg)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("转换修饰符组失败: %s", gmg.Name))
		}
		item.AddModifierGroup(mg)
	}

	return item, nil
}

// convertFromGrabModifierGroup 从 Grab 格式转换修饰符组
func (c *GrabConverter) convertFromGrabModifierGroup(gmg *GrabModifierGroup) (*valueobject.ModifierGroup, error) {
	status := valueobject.AvailableStatusAvailable
	if gmg.AvailableStatus == "UNAVAILABLE" {
		status = valueobject.AvailableStatusUnavailable
	}

	mg, err := valueobject.NewModifierGroup(
		gmg.ID,
		gmg.Name,
		c.filterSupportedLanguages(gmg.NameTranslation),
		gmg.Sequence,
		status,
		gmg.SelectionRangeMin,
		gmg.SelectionRangeMax,
	)
	if err != nil {
		return nil, err
	}

	// 转换修饰符
	for _, gm := range gmg.Modifiers {
		mStatus := valueobject.AvailableStatusAvailable
		if gm.AvailableStatus == "UNAVAILABLE" {
			mStatus = valueobject.AvailableStatusUnavailable
		}

		m, err := valueobject.NewModifier(gm.ID, gm.Name, gm.Sequence, mStatus, gm.Price)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("转换修饰符失败: %s", gm.Name))
		}
		m.NameTranslation = gm.NameTranslation
		mg.AddModifier(m)
	}

	return mg, nil
}

// ValidateData 验证 Grab 菜单数据格式
func (c *GrabConverter) ValidateData(platformData interface{}) error {
	// 尝试多种类型断言
	if grabMenu, ok := platformData.(*GrabMenu); ok {
		if grabMenu.Currency.Code == "" {
			return errors.New("货币代码不能为空")
		}
		if len(grabMenu.Categories) == 0 {
			return errors.New("至少需要一个分类")
		}
		return nil
	}

	if grabMenu, ok := platformData.(GrabMenu); ok {
		if grabMenu.Currency.Code == "" {
			return errors.New("货币代码不能为空")
		}
		if len(grabMenu.Categories) == 0 {
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
func (c *GrabConverter) LoadMenuFromDatabase(ctx context.Context, companyUuid uint64, currencyUnit string, categoryIDs []uint64) (*entity.TakeoutMenu, error) {
	// 创建菜单聚合根
	menu, err := entity.NewTakeoutMenu(companyUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "创建菜单聚合根失败")
	}

	// 设置货币
	// 根据货币符号推断货币代码和 exponent
	currencySymbol := utils.IfString(currencyUnit == "", "฿", currencyUnit)
	currencyCode, exponent := c.getCurrencyInfoBySymbol(currencySymbol)
	if err := menu.SetCurrency(currencyCode, currencySymbol, exponent); err != nil {
		return nil, errors.WithMessage(err, "设置货币失败")
	}

	// 加载售卖时段（示例：添加一个默认时段）
	defaultSellingTime, err := valueobject.NewSellingTime("SELLINGTIME-01", "全天", 1, "1000-01-01 00:00:00", "9999-12-31 23:59:59")
	if err != nil {
		return nil, errors.WithMessage(err, "创建售卖时段失败")
	}
	if err := menu.AddSellingTime(defaultSellingTime); err != nil {
		return nil, errors.WithMessage(err, "添加售卖时段失败")
	}

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
		if err := menu.AddCategory(category); err != nil {
			return nil, errors.WithMessage(err, "添加分类到菜单失败")
		}
	}

	return menu, nil
}

// convertTTPOSCategory 从 TTPOS model 转换分类
func (c *GrabConverter) convertTTPOSCategory(ctx context.Context, cat any, sequence int) (*valueobject.Category, error) {
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
	categoryStatus := valueobject.AvailableStatusAvailable
	if category.Status != 1 {
		categoryStatus = valueobject.AvailableStatusHide
	}

	// 创建分类值对象
	categoryVO, err := valueobject.NewCategory(
		func() string {
			if category.SourceId != "" {
				return category.SourceId
			}
			return fmt.Sprintf("TTPOS-CAT-%d", category.Uuid)
		}(),
		categoryName,
		sequence,
		categoryStatus,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "创建分类值对象失败")
	}

	// 设置售卖时段 ID（默认使用全天）
	categoryVO.SellingTimeID = "SELLINGTIME-01"

	// 设置多语言名称（只包含 Grab 支持的语言）
	if category.MultiLanguageName.Uuid != 0 {
		categoryVO.NameTranslation = c.filterSupportedLanguages(category.MultiLanguageName.ToMap())
	}

	return categoryVO, nil
}

// convertTTPOSProduct 从 TTPOS model 转换商品
func (c *GrabConverter) convertTTPOSProduct(ctx context.Context, pkg any, sequence int) (*valueobject.MenuItem, error) {
	// 类型断言
	takeoutProduct, ok := pkg.(*model.ProductPackageTakeout)
	if !ok {
		return nil, errors.New("类型断言失败：pkg 不是 *model.ProductPackageTakeout")
	}

	// 判断商品状态：
	// 1. 商品下架（Status=0）或删除 -> HIDE
	// 2. 商品所有规格都售罄 -> UNAVAILABLE
	// 3. 正常（Status=1）-> AVAILABLE
	status := valueobject.AvailableStatusAvailable
	if takeoutProduct.IsDown() || takeoutProduct.IsDelete() {
		// 商品下架或删除，设置为 HIDE
		status = valueobject.AvailableStatusHide
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
	if len(flavorsForPrice) > 0 {
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
		// 没有规格，使用商品原价
		price = int64(takeoutProduct.ProductPackage.Price * 100)
	}

	// 创建商品值对象
	menuItem, err := valueobject.NewMenuItem(
		func() string {
			if takeoutProduct.SourceProductId != "" {
				return takeoutProduct.SourceProductId
			}
			return fmt.Sprintf("TTPOS-ITEM-%d", takeoutProduct.Uuid)
		}(),
		itemName,
		sequence,
		status,
		price,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "创建商品值对象失败")
	}

	// 设置多语言名称（只包含 Grab 支持的语言）
	if takeoutProduct.MultiLanguageName.Uuid != 0 {
		menuItem.NameTranslation = c.filterSupportedLanguages(takeoutProduct.MultiLanguageName.ToMap())
	} else if takeoutProduct.ProductPackage.MultiLanguageName.Uuid != 0 {
		menuItem.NameTranslation = c.filterSupportedLanguages(takeoutProduct.ProductPackage.MultiLanguageName.ToMap())
	}

	// 设置商品描述（优先使用多语言字段）
	if takeoutProduct.ProductPackage.DescribeMultiLanguageName.Uuid != 0 {
		// 使用多语言描述
		menuItem.DescriptionTranslation = c.filterSupportedLanguages(takeoutProduct.ProductPackage.DescribeMultiLanguageName.ToMap())
		// 设置默认描述（使用英文，如果没有则回退）
		menuItem.Description = takeoutProduct.ProductPackage.DescribeMultiLanguageName.GetNameByLangWithFallback("en")
	} else if takeoutProduct.ProductPackage.Describe != "" {
		// 回退到单语言描述字段
		menuItem.Description = takeoutProduct.ProductPackage.Describe
	}

	// 设置商品图片（使用 GetUrl 方法，如果没有本地图片则使用外部URL）
	if takeoutProduct.ImageFile.Uuid != 0 && takeoutProduct.ImageFile.FileUrl != "" {
		imageUrl := takeoutProduct.ImageFile.GetUrl(utils.GetBaseURL(nil))
		menuItem.Photos = []string{imageUrl}
	} else if takeoutProduct.ProductPackage.ImageUrl != "" {
		// 如果没有本地图片文件，使用外部图片URL
		menuItem.Photos = []string{takeoutProduct.ProductPackage.ImageUrl}
	}

	// 设置售卖时段 ID（默认使用全天）
	menuItem.SellingTimeID = "SELLINGTIME-01"

	// 处理修饰符（Modifier）
	if takeoutProduct.ProductPackage.ProductType != constant.ProductTypePackage {

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
		// 4. 处理套餐分组（如果是套餐）
		if err := c.convertPackageGroups(ctx, menuItem, &takeoutProduct.ProductPackage); err != nil {
			return nil, errors.WithMessage(err, "转换套餐分组失败")
		}
	}

	// 检查商品是否所有规格都售罄（仅当商品正常上架时才检查）
	if status == valueobject.AvailableStatusAvailable {
		if c.isAllFlavorsSoldOut(takeoutProduct) {
			// 所有规格都售罄，商品状态设为 UNAVAILABLE
			menuItem.AvailableStatus = valueobject.AvailableStatusUnavailable
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
func (c *GrabConverter) loadCategoryProducts(ctx context.Context, companyUuid uint64, category *valueobject.Category, categoryUuid uint64) error {
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

		category.AddItem(menuItem)
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
	menuItem *valueobject.MenuItem,
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
	modifierGroupStatus := valueobject.AvailableStatusAvailable
	allSoldOut := true
	allDown := true
	for i := range takeoutProduct.ProductBomTakeouts {
		bomTakeout := &takeoutProduct.ProductBomTakeouts[i]
		if bomTakeout.IsDelete() || bomTakeout.GrabModifierId == "" {
			continue
		}
		if bomTakeout.ProductBom.IsSoldOut != 1 && bomTakeout.ProductBom.StockNum > 0 {
			allSoldOut = false
		}
		if !bomTakeout.ProductBom.IsDown() {
			allDown = false
		}
	}
	if allDown {
		modifierGroupStatus = valueobject.AvailableStatusHide
	} else if allSoldOut {
		modifierGroupStatus = valueobject.AvailableStatusUnavailable
	}

	// 创建一个修饰符组，包含所有规格
	modifierGroup, err := valueobject.NewModifierGroup(
		fmt.Sprintf("TTPOS-FLAVOR-GROUP-%d", takeoutProduct.ProductPackageUuid),
		"Specifications",
		map[string]string{
			"en": "Specifications",
			"zh": "规格",
			"th": "ข้อมูลจำเพาะ",
			"ms": "Saiz",
			"vi": "Kích thước",
			"id": "Ukuran",
			"km": "ទិចនាករ",
			"my": "သတ်မှတ်ချက်များ",
		},
		1,
		modifierGroupStatus,
		minSelection,
		maxSelection,
	)
	if err != nil {
		return err
	}

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
		modifierStatus := valueobject.AvailableStatusAvailable
		if bomTakeout.ProductBom.IsSoldOut == 1 || bomTakeout.ProductBom.StockNum <= 0 {
			// 规格售罄
			modifierStatus = valueobject.AvailableStatusUnavailable
		} else if bomTakeout.ProductBom.IsDown() {
			// 规格下架或删除
			modifierStatus = valueobject.AvailableStatusHide
		}

		// 创建修饰符
		modifier, err := valueobject.NewModifier(
			func() string {
				if bomTakeout.GrabModifierId != "" {
					return bomTakeout.GrabModifierId
				}
				return fmt.Sprintf("TTPOS-FLAVOR-%d", bomTakeout.ProductBomUuid)
			}(),
			flavorName,
			idx+1,
			modifierStatus,
			priceInCents,
		)
		if err != nil {
			return errors.WithMessage(err, fmt.Sprintf("创建规格修饰符失败: %s", flavorName))
		}

		// 设置多语言名称
		if bomTakeout.ProductBom.ProductFlavor.MultiLanguageName.Uuid != 0 {
			modifier.NameTranslation = c.filterSupportedLanguages(bomTakeout.ProductBom.ProductFlavor.MultiLanguageName.ToMap())
		}

		modifierGroup.AddModifier(modifier)
	}

	menuItem.AddModifierGroup(modifierGroup)
	return nil
}

// convertProductSauces 转换商品小料为修饰符组
func (c *GrabConverter) convertProductSauces(ctx context.Context, menuItem *valueobject.MenuItem, productPackage *model.ProductPackage) error {
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
	isRequired := productPackage.SauceRequired == 1
	maxSelection := int(productPackage.SauceMaxSelection)
	if maxSelection == 0 {
		maxSelection = len(sauces) // 如果未配置，使用小料总数
	}
	// 确保最大选择数量至少为 1
	if maxSelection < 1 {
		maxSelection = 1
	}

	// 根据图片规则：必选时可选组(必选) min:1 max:配置值，可选时可选组(非必选) min:0 max:配置值
	minSelection := utils.IfInt(isRequired, 1, 0)

	// 判断小料组状态：
	// 如果所有小料都售罄（is_sold_out = 1 或 StockNum <= 0）-> UNAVAILABLE
	// 如果所有小料都下架 -> HIDE
	// 否则 -> AVAILABLE
	modifierGroupStatus := valueobject.AvailableStatusAvailable
	allSoldOut := true
	allDown := true
	for i := range productPackage.ProductBoms {
		bom := &productPackage.ProductBoms[i]
		if !bom.IsSauce() {
			continue
		}
		if bom.IsSoldOut != 1 && bom.StockNum > 0 {
			allSoldOut = false
		}
		if !bom.IsDown() {
			allDown = false
		}
	}
	if allDown {
		modifierGroupStatus = valueobject.AvailableStatusHide
	} else if allSoldOut {
		modifierGroupStatus = valueobject.AvailableStatusUnavailable
	}

	// 创建一个修饰符组，包含所有小料
	modifierGroup, err := valueobject.NewModifierGroup(
		fmt.Sprintf("TTPOS-SAUCE-GROUP-%d", productPackage.Uuid),
		"Add Toppings",
		map[string]string{
			"en": "Add Toppings",
			"zh": "加料",
			"th": "เพิ่มเครื่อง",
			"ms": "Tambahkan Topping",
			"vi": "Thêm gia vị",
			"id": "Tambahkan Topping",
			"km": "ការបង្ហាញ",
			"my": "အပိုထည့်ခြင်း",
		},
		2,
		modifierGroupStatus,
		minSelection,
		maxSelection,
	)
	if err != nil {
		return err
	}

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
		modifierStatus := valueobject.AvailableStatusAvailable
		if bom.IsSoldOut == 1 || bom.StockNum <= 0 {
			// 小料售罄
			modifierStatus = valueobject.AvailableStatusUnavailable
		} else if bom.IsDown() {
			// 小料下架或删除
			modifierStatus = valueobject.AvailableStatusHide
		}

		// 创建修饰符
		modifier, err := valueobject.NewModifier(
			fmt.Sprintf("TTPOS-SAUCE-%d", bom.ProductSauce.Uuid),
			sauceName,
			idx+1,
			modifierStatus,
			int64(bom.Price*100), // 转换为分
		)
		if err != nil {
			return errors.WithMessage(err, fmt.Sprintf("创建小料修饰符失败: %s", sauceName))
		}

		// 设置多语言名称
		if bom.ProductSauce.MultiLanguageName.Uuid != 0 {
			modifier.NameTranslation = c.filterSupportedLanguages(bom.ProductSauce.MultiLanguageName.ToMap())
		}

		modifierGroup.AddModifier(modifier)
	}

	menuItem.AddModifierGroup(modifierGroup)
	return nil
}

// convertProductAttributeGroups 转换商品属性组为修饰符组
func (c *GrabConverter) convertProductAttributeGroups(ctx context.Context, menuItem *valueobject.MenuItem, productPackage *model.ProductPackage) error {
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
		modifierGroup, err := valueobject.NewModifierGroup(
			func() string {
				if packageAttrGroup.ProductAttributeGroup.SourceId != "" {
					return packageAttrGroup.ProductAttributeGroup.SourceId
				}
				return fmt.Sprintf("TTPOS-ATTR-GROUP-%d", packageAttrGroup.ProductAttributeGroup.Uuid)
			}(),
			groupName,
			c.filterSupportedLanguages(packageAttrGroup.ProductAttributeGroup.MultiLanguageName.ToMap()),
			sequence,
			valueobject.AvailableStatusAvailable,
			int(utils.IfInt(packageAttrGroup.IsMust == 1, 1, 0)), // 如果必选，最小选择1个
			maxSelection, // 最大选择数量（确保至少为 1）
		)
		if err != nil {
			return errors.WithMessage(err, fmt.Sprintf("创建属性组修饰符组失败: %s", groupName))
		}

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

			// 创建修饰符（属性没有价格，价格为0）
			modifier, err := valueobject.NewModifier(
				func() string {
					if packageAttr.Attribute.SourceId != "" {
						return packageAttr.Attribute.SourceId
					}
					return fmt.Sprintf("TTPOS-ATTR-%d", packageAttr.Attribute.Uuid)
				}(),
				attrName,
				idx+1,
				valueobject.AvailableStatusAvailable,
				0, // 属性没有价格
			)
			if err != nil {
				return errors.WithMessage(err, fmt.Sprintf("创建属性修饰符失败: %s", attrName))
			}

			// 设置多语言名称
			if packageAttr.Attribute.MultiLanguageName.Uuid != 0 {
				modifier.NameTranslation = c.filterSupportedLanguages(packageAttr.Attribute.MultiLanguageName.ToMap())
			}

			modifierGroup.AddModifier(modifier)
		}

		// 只有包含修饰符的组才添加
		if len(modifierGroup.Modifiers) > 0 {
			menuItem.AddModifierGroup(modifierGroup)
			sequence++
		}
	}

	return nil
}

// convertPackageGroups 转换套餐分组为修饰符组
func (c *GrabConverter) convertPackageGroups(ctx context.Context, menuItem *valueobject.MenuItem, productPackage *model.ProductPackage) error {
	// 收集所有未删除的套餐分组
	packageGroups := make([]*model.ProductPackageGroup, 0)
	for i := range productPackage.ProductPackageGroups {
		group := &productPackage.ProductPackageGroups[i]
		if !group.IsDelete() {
			packageGroups = append(packageGroups, group)
		}
	}
	if len(packageGroups) == 0 {
		return nil
	}

	// 按 Sort 排序
	sort.Slice(packageGroups, func(i, j int) bool {
		if packageGroups[i].Sort != packageGroups[j].Sort {
			return packageGroups[i].Sort < packageGroups[j].Sort
		}
		return packageGroups[i].Uuid < packageGroups[j].Uuid
	})

	// sequence 从 4 开始（规格=1，小料=2，属性组=3）
	sequence := 4

	// 遍历所有套餐分组
	for _, packageGroup := range packageGroups {
		// 收集分组中未删除的子商品
		groupItems := make([]*model.ProductPackageGroupItem, 0)
		for i := range packageGroup.ProductPackageGroupItems {
			item := &packageGroup.ProductPackageGroupItems[i]
			if !item.IsDelete() && item.ProductPackage != nil && !item.ProductPackage.IsDelete() {
				groupItems = append(groupItems, item)
			}
		}

		// 如果分组中没有有效的子商品，跳过该分组
		if len(groupItems) == 0 {
			continue
		}

		// 按 Sort 排序子商品
		sort.Slice(groupItems, func(i, j int) bool {
			if groupItems[i].Sort != groupItems[j].Sort {
				return groupItems[i].Sort < groupItems[j].Sort
			}
			return groupItems[i].Uuid < groupItems[j].Uuid
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
			minSelection = len(groupItems)
			maxSelection = len(groupItems)
		} else {
			// 可选组：最小选择=0，最大选择=OptionalCount
			maxSelection = packageGroup.OptionalCount
			// 如果 OptionalCount 为 0 或未配置，使用子商品数量
			if maxSelection == 0 {
				maxSelection = len(groupItems)
			}
			minSelection = packageGroup.OptionalCount
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
		modifierGroup, err := valueobject.NewModifierGroup(
			fmt.Sprintf("TTPOS-PACKAGE-GROUP-%d", packageGroup.Uuid),
			groupName,
			nameTranslation,
			sequence,
			valueobject.AvailableStatusAvailable,
			minSelection,
			maxSelection,
		)
		if err != nil {
			return errors.WithMessage(err, fmt.Sprintf("创建套餐分组修饰符组失败: %s", groupName))
		}

		// 转换每个子商品为修饰符
		for idx, groupItem := range groupItems {
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

			// 计算价格：AddPrice（加价金额，转换为分）
			priceInCents := int64(groupItem.AddPrice * 100)

			// 创建修饰符
			modifier, err := valueobject.NewModifier(
				fmt.Sprintf("TTPOS-PACKAGE-ITEM-%d", groupItem.Uuid),
				itemName,
				idx+1,
				valueobject.AvailableStatusAvailable,
				priceInCents,
			)
			if err != nil {
				return errors.WithMessage(err, fmt.Sprintf("创建套餐子商品修饰符失败: %s", itemName))
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
				modifier.NameTranslation = translations
			}

			modifierGroup.AddModifier(modifier)
		}

		// 只有包含修饰符的组才添加
		if len(modifierGroup.Modifiers) > 0 {
			menuItem.AddModifierGroup(modifierGroup)
			sequence++
		}
	}

	return nil
}

// isAllFlavorsSoldOut 检查商品是否所有规格都售罄
// 返回 true 表示所有规格都售罄，false 表示至少有一个规格可用
func (c *GrabConverter) isAllFlavorsSoldOut(takeoutProduct *model.ProductPackageTakeout) bool {
	// 如果没有规格，商品本身就不能售罄
	if len(takeoutProduct.ProductPackage.ProductBoms) == 0 {
		return false
	}

	// 收集所有规格
	hasFlavor := false
	allSoldOut := true
	for i := range takeoutProduct.ProductPackage.ProductBoms {
		bom := &takeoutProduct.ProductPackage.ProductBoms[i]
		// 只检查规格和小料
		if !bom.IsFlavor() {
			continue
		}
		hasFlavor = true
		// 如果有任何一个规格未售罄，则商品不是售罄状态
		if bom.IsSoldOut != 1 && bom.StockNum > 0 {
			allSoldOut = false
			break
		}
	}

	// 如果没有规格，返回 false
	if !hasFlavor {
		return false
	}

	return allSoldOut
}
