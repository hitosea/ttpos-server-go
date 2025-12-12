package grab

import (
	"encoding/json"
	"fmt"
	"time"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/modules/takeout/domain/menu/entity"
	menuRepo "ttpos-server-go/app/modules/takeout/domain/menu/repository"
	"ttpos-server-go/app/modules/takeout/domain/menu/valueobject"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MenuImporter Grab 菜单导入器
type MenuImporter struct {
	dbm *database.DBManager
}

// NewMenuImporter 创建菜单导入器
func NewMenuImporter(dbm *database.DBManager, _ menuRepo.IProductMapRepository) *MenuImporter {
	return &MenuImporter{
		dbm: dbm,
	}
}

// ImportResult 导入结果
type ImportResult struct {
	SuccessCount int
	FailureCount int
	CreatedItems int
	UpdatedItems int
	Failures     []resp.GrabProductImportFailure
}

// ImportMenu 导入 Grab 菜单到 TTPOS 系统
// 根据图片规则：
// 1. 普通商品的规格映射为必选/固定分组(min=1, max=1)，价格为差价
// 2. 属性/加料映射为可选分组(min=0/1, max=配置值)，价格为加料本身价格
// 3. 套餐的固定分组(必选)：min=分组数量, max=分组数量
// 4. 套餐的可选分组(非必选)：min=0, max=配置值
func (mi *MenuImporter) ImportMenu(ctx context.Context, companyUuid uint64, menu *entity.TakeoutMenu, overwrite bool) (*ImportResult, error) {
	result := &ImportResult{
		Failures: make([]resp.GrabProductImportFailure, 0),
	}

	db := ctx.GetDB()

	// 遍历所有分类和商品
	for _, category := range menu.Categories {
		for _, item := range category.Items {
			err := db.Transaction(func(tx *gorm.DB) error {
				// 导入单个商品
				_, err := mi.importItem(tx, companyUuid, category, item)
				return err
			})

			if err != nil {
				result.FailureCount++
				result.Failures = append(result.Failures, resp.GrabProductImportFailure{
					GrabProductId: item.ID,
					Message:       fmt.Sprintf("%s: %s", item.Name, err.Error()),
				})
				logger.Logger.Error("导入商品失败",
					zap.String("grab_product_id", item.ID),
					zap.String("product_name", item.Name),
					zap.Error(err))
			} else {
				result.SuccessCount++
				result.CreatedItems++ // 新创建的商品计数
			}
		}
	}

	return result, nil
}

// importItem 导入单个商品（全新创建，根据 Grab ID 判断是否已存在）
func (mi *MenuImporter) importItem(
	tx *gorm.DB,
	companyUuid uint64,
	category *valueobject.Category,
	item *valueobject.MenuItem,
) (uint64, error) {
	// 1. 检查是否已经导入过该 Grab 商品（通过外卖商品表的 grab_product_id 字段）
	var existingTakeout model.ProductPackageTakeout
	err := tx.Where("grab_product_id = ? AND delete_time = 0", item.ID).
		First(&existingTakeout).Error

	if err == nil {
		// 已存在，返回已有商品的 UUID
		logger.Logger.Info("Grab商品已存在，跳过导入",
			zap.String("grab_product_id", item.ID),
			zap.String("product_name", item.Name),
			zap.Uint64("product_package_uuid", existingTakeout.ProductPackageUuid))
		return existingTakeout.ProductPackageUuid, nil
	}

	if err != gorm.ErrRecordNotFound {
		return 0, errors.WithMessage(err, "查询现有商品失败")
	}

	// 2. 创建新商品（包括店内商品和外卖商品），带 Grab ID 标记
	productPackageUuid, err := mi.createNewProduct(tx, companyUuid, category, item)
	if err != nil {
		return 0, errors.WithMessage(err, "创建新商品失败")
	}

	// 3. 处理规格/属性/加料等修饰符组
	if err := mi.processModifierGroups(tx, companyUuid, item, productPackageUuid, true); err != nil {
		return 0, errors.WithMessage(err, "处理修饰符组失败")
	}

	return productPackageUuid, nil
}

// createNewProduct 创建新商品（店内商品+外卖商品）
func (mi *MenuImporter) createNewProduct(
	tx *gorm.DB,
	companyUuid uint64,
	category *valueobject.Category,
	item *valueobject.MenuItem,
) (uint64, error) {
	// 1. 查找或创建分类
	categoryUuid, err := mi.findOrCreateCategory(tx, companyUuid, category)
	if err != nil {
		return 0, errors.WithMessage(err, "查找或创建分类失败")
	}

	// 2. 创建多语言名称
	multiLanguageNameUuid, err := mi.createMultiLanguageName(tx, item.NameTranslation, item.Name)
	if err != nil {
		return 0, errors.WithMessage(err, "创建多语言名称失败")
	}

	// 3. 创建多语言描述
	var describeMultiLanguageNameUuid uint64
	if item.Description != "" || len(item.DescriptionTranslation) > 0 {
		describeMultiLanguageNameUuid, err = mi.createMultiLanguageName(tx, item.DescriptionTranslation, item.Description)
		if err != nil {
			return 0, errors.WithMessage(err, "创建多语言描述失败")
		}
	}

	// 4. 生成UUID
	uuid, err := utils.GetID()
	if err != nil {
		return 0, errors.WithMessage(err, "生成UUID失败")
	}

	// 5. 创建店内商品包（ProductPackage）
	// 价格：分转换为元
	price := float64(item.Price) / 100.0

	productPackage := &model.ProductPackage{
		BaseModel: model.BaseModel{
			Uuid:       uuid,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:                          item.Name,
		MultiLanguageNameUuid:         multiLanguageNameUuid,
		Describe:                      item.Description,
		DescribeMultiLanguageNameUuid: describeMultiLanguageNameUuid,
		Price:                         price,
		CategoryUuid:                  categoryUuid,
		ProductType:                   0, // 默认为普通商品
		Status:                        0, // 默认下架（根据图片规则）
		IsShowCashier:                 1,
		IsShowTablet:                  1,
		IsShowKitchen:                 1,
		IsShowAssistant:               1,
		IsShowH5:                      0,
		IsShowDelivery:                0,
		IsShowKiosk:                   0,
		NumType:                       1, // 数量类型：1-整数
		DeductStockType:               0, // 不扣库存
	}

	productPackageRepo := repository.NewProductPackageRepo(tx)
	if err := productPackageRepo.CreateProductPackage(productPackage); err != nil {
		return 0, errors.WithMessage(err, "创建店内商品失败")
	}

	// 6. 创建外卖商品（ProductPackageTakeout）
	takeoutUuid, err := utils.GetID()
	if err != nil {
		return 0, errors.WithMessage(err, "生成外卖商品UUID失败")
	}

	productPackageTakeout := &model.ProductPackageTakeout{
		BaseModel: model.BaseModel{
			Uuid:       takeoutUuid,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		ProductPackageUuid:    uuid,
		MultiLanguageNameUuid: multiLanguageNameUuid,
		Name:                  item.Name,
		ProductType:           0, // 默认为普通商品
		TakeoutType:           1, // 1-Grab
		Status:                0, // 默认下架（根据图片规则：以下商品未与店内关联，不选择具体店内商品内容，需先建立店内商品）
		CategoryUuid:          categoryUuid,
		SpecialCategoryUuid:   0,
		GrabProductId:         item.ID, // 保存 Grab 商品 ID 用于去重
	}

	takeoutRepo := repository.NewProductPackageTakeoutRepo(tx)
	if err := takeoutRepo.CreateProductPackageTakeout(productPackageTakeout); err != nil {
		return 0, errors.WithMessage(err, "创建外卖商品失败")
	}

	return uuid, nil
}

// processModifierGroups 处理修饰符组（规格/属性/加料等）
// 根据图片规则：
// - 规格组：必选(min=1, max=1)，价格存储为与最小规格的差价
// - 属性组：根据配置可选/必选(min=0/1, max=配置值)，价格为0
// - 加料组：根据配置可选/必选(min=0/1, max=配置值)，价格为加料本身价格
func (mi *MenuImporter) processModifierGroups(
	tx *gorm.DB,
	companyUuid uint64,
	item *valueobject.MenuItem,
	productPackageUuid uint64,
	isNewProduct bool,
) error {
	// 如果没有修饰符组，直接返回
	if len(item.ModifierGroups) == 0 {
		return nil
	}

	// 遍历所有修饰符组
	for _, modifierGroup := range item.ModifierGroups {
		// 根据修饰符组名称和规则判断类型
		groupType := mi.detectModifierGroupType(modifierGroup)

		switch groupType {
		case "flavor": // 规格
			if err := mi.processFlavors(tx, productPackageUuid, modifierGroup, isNewProduct); err != nil {
				return errors.WithMessage(err, fmt.Sprintf("处理规格组失败: %s", modifierGroup.Name))
			}
		case "sauce": // 加料
			if err := mi.processSauces(tx, productPackageUuid, modifierGroup, isNewProduct); err != nil {
				return errors.WithMessage(err, fmt.Sprintf("处理加料组失败: %s", modifierGroup.Name))
			}
		default:
			// 默认当作属性处理
			if err := mi.processAttributes(tx, productPackageUuid, modifierGroup, isNewProduct); err != nil {
				return errors.WithMessage(err, fmt.Sprintf("处理修饰符组失败: %s", modifierGroup.Name))
			}
		}
	}

	return nil
}

// detectModifierGroupType 检测修饰符组类型
// 根据名称关键字和规则判断是规格、加料还是属性
func (mi *MenuImporter) detectModifierGroupType(group *valueobject.ModifierGroup) string {
	// 根据选择规则判断
	// 规格：通常 min=1, max=1（必选且只能选一个）
	if group.SelectionRangeMin == 1 && group.SelectionRangeMax == 1 {
		return "flavor"
	}

	// 检查名称关键字
	name := group.Name
	// 规格相关关键字
	if contains(name, []string{"规格", "Specifications", "Size", "Saiz", "Kích thước", "Ukuran", "ទិចនាករ", "သတ်မှတ်ချက်များ"}) {
		return "flavor"
	}

	// 加料相关关键字
	if contains(name, []string{"加料", "Add Toppings", "Topping", "Tambahkan", "Thêm gia vị", "အပိုထည့်ခြင်း", "ការបង្ហាញ", "เพิ่มเครื่อง"}) {
		return "sauce"
	}

	// 属性相关关键字
	if contains(name, []string{"属性", "Attribute", "ຄຸນລັກສະນະ", "Atribut", "คุณสมบัติ"}) {
		return "attribute"
	}

	// 如果有价格，更可能是加料
	hasPrice := false
	for _, modifier := range group.Modifiers {
		if modifier.Price > 0 {
			hasPrice = true
			break
		}
	}
	if hasPrice {
		return "sauce"
	}

	// 默认当作属性
	return "attribute"
}

// processFlavors 处理规格（ProductFlavor + ProductBom）
// 根据图片规则：规格价格存储为与最小规格的差价
func (mi *MenuImporter) processFlavors(
	tx *gorm.DB,
	productPackageUuid uint64,
	modifierGroup *valueobject.ModifierGroup,
	isNewProduct bool,
) error {
	if len(modifierGroup.Modifiers) == 0 {
		return nil
	}

	// 找出最小价格（用于计算差价）
	minPrice := modifierGroup.Modifiers[0].Price
	for _, modifier := range modifierGroup.Modifiers {
		if modifier.Price < minPrice {
			minPrice = modifier.Price
		}
	}

	// 创建或查找每个规格
	for _, modifier := range modifierGroup.Modifiers {
		// 查找是否已存在该规格
		flavorRepo := repository.NewProductFlavorRepo(tx)
		var flavor model.ProductFlavor
		err := tx.Where("name = ? AND delete_time = 0", modifier.Name).First(&flavor).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return errors.WithMessage(err, "查询规格失败")
		}

		// 如果不存在，创建新规格
		if flavor.Uuid == 0 {
			flavorUuid, err := utils.GetID()
			if err != nil {
				return errors.WithMessage(err, "生成规格UUID失败")
			}

			// 创建多语言名称
			multiLanguageNameUuid, err := mi.createMultiLanguageName(tx, modifier.NameTranslation, modifier.Name)
			if err != nil {
				return errors.WithMessage(err, "创建规格多语言名称失败")
			}

			flavor = model.ProductFlavor{
				BaseModel: model.BaseModel{
					Uuid:       flavorUuid,
					CreateTime: time.Now().Unix(),
					UpdateTime: time.Now().Unix(),
				},
				Name:                  modifier.Name,
				MultiLanguageNameUuid: multiLanguageNameUuid,
				Sort:                  modifier.Sequence,
			}

			if err := flavorRepo.CreateProductFlavor(flavor); err != nil {
				return errors.WithMessage(err, "创建规格失败")
			}
		}

		// 创建 ProductBom 关联（规格与商品的关联）
		bomRepo := repository.NewProductBomRepo(tx)

		// 检查是否已存在该关联
		var existingBom model.ProductBom
		err = tx.Where("product_package_uuid = ? AND product_flavor_uuid = ? AND delete_time = 0",
			productPackageUuid, flavor.Uuid).First(&existingBom).Error

		if err == gorm.ErrRecordNotFound {
			// 不存在，创建新关联
			bomUuid, err := utils.GetID()
			if err != nil {
				return errors.WithMessage(err, "生成BOM UUID失败")
			}

			// 计算差价（分转换为元）
			priceDiff := float64(modifier.Price-minPrice) / 100.0

			bom := model.ProductBom{
				BaseModel: model.BaseModel{
					Uuid:       bomUuid,
					CreateTime: time.Now().Unix(),
					UpdateTime: time.Now().Unix(),
				},
				ProductPackageUuid: productPackageUuid,
				ProductFlavorUuid:  flavor.Uuid,
				Price:              priceDiff,
			}

			if _, err := bomRepo.CreateProductBom(bom); err != nil {
				return errors.WithMessage(err, "创建BOM关联失败")
			}
		} else if err != nil {
			return errors.WithMessage(err, "查询BOM关联失败")
		}
		// 如果已存在，不做处理（避免重复创建）
	}

	return nil
}

// processSauces 处理加料（ProductSauce + ProductBom）
// 根据图片规则：加料价格存储为加料本身的价格（非差价）
func (mi *MenuImporter) processSauces(
	tx *gorm.DB,
	productPackageUuid uint64,
	modifierGroup *valueobject.ModifierGroup,
	isNewProduct bool,
) error {
	if len(modifierGroup.Modifiers) == 0 {
		return nil
	}

	// 更新商品包的小料配置
	// 根据修饰符组的选择范围设置
	sauceRequired := 0
	if modifierGroup.SelectionRangeMin > 0 {
		sauceRequired = 1 // 必选
	}
	sauceMaxSelection := modifierGroup.SelectionRangeMax
	if sauceMaxSelection == 0 {
		sauceMaxSelection = len(modifierGroup.Modifiers)
	}

	productPackageRepo := repository.NewProductPackageRepo(tx)
	updates := map[string]interface{}{
		"sauce_required":      sauceRequired,
		"sauce_max_selection": sauceMaxSelection,
		"update_time":         time.Now().Unix(),
	}
	if err := productPackageRepo.UpdateProductPackage(updates, repository.CommonRepo.WhereByUuid(productPackageUuid)); err != nil {
		return errors.WithMessage(err, "更新商品包小料配置失败")
	}

	// 创建或查找每个小料
	for _, modifier := range modifierGroup.Modifiers {
		// 查找是否已存在该小料
		sauceRepo := repository.NewProductSauceRepo(tx)
		var sauce model.ProductSauce
		err := tx.Where("name = ? AND delete_time = 0", modifier.Name).First(&sauce).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return errors.WithMessage(err, "查询小料失败")
		}

		// 如果不存在，创建新小料
		if sauce.Uuid == 0 {
			sauceUuid, err := utils.GetID()
			if err != nil {
				return errors.WithMessage(err, "生成小料UUID失败")
			}

			// 创建多语言名称
			multiLanguageNameUuid, err := mi.createMultiLanguageName(tx, modifier.NameTranslation, modifier.Name)
			if err != nil {
				return errors.WithMessage(err, "创建小料多语言名称失败")
			}

			sauce = model.ProductSauce{
				BaseModel: model.BaseModel{
					Uuid:       sauceUuid,
					CreateTime: time.Now().Unix(),
					UpdateTime: time.Now().Unix(),
				},
				Name:                  modifier.Name,
				MultiLanguageNameUuid: multiLanguageNameUuid,
				Sort:                  modifier.Sequence,
				Price:                 float64(modifier.Price) / 100.0,
			}

			if err := sauceRepo.CreateProductSauce(&sauce); err != nil {
				return errors.WithMessage(err, "创建小料失败")
			}
		}

		// 创建 ProductBom 关联（小料与商品的关联）
		bomRepo := repository.NewProductBomRepo(tx)

		// 检查是否已存在该关联
		var existingBom model.ProductBom
		err = tx.Where("product_package_uuid = ? AND product_sauce_uuid = ? AND delete_time = 0",
			productPackageUuid, sauce.Uuid).First(&existingBom).Error

		if err == gorm.ErrRecordNotFound {
			// 不存在，创建新关联
			bomUuid, err := utils.GetID()
			if err != nil {
				return errors.WithMessage(err, "生成BOM UUID失败")
			}

			// 价格：分转换为元（小料是绝对价格，不是差价）
			price := float64(modifier.Price) / 100.0

			bom := model.ProductBom{
				BaseModel: model.BaseModel{
					Uuid:       bomUuid,
					CreateTime: time.Now().Unix(),
					UpdateTime: time.Now().Unix(),
				},
				ProductPackageUuid: productPackageUuid,
				ProductSauceUuid:   sauce.Uuid,
				Price:              price,
			}

			if _, err := bomRepo.CreateProductBom(bom); err != nil {
				return errors.WithMessage(err, "创建BOM关联失败")
			}
		} else if err != nil {
			return errors.WithMessage(err, "查询BOM关联失败")
		}
		// 如果已存在，不做处理
	}

	return nil
}

// processAttributes 处理属性组（ProductAttributeGroup + ProductAttribute）
// 根据图片规则：属性价格为0
func (mi *MenuImporter) processAttributes(
	tx *gorm.DB,
	productPackageUuid uint64,
	modifierGroup *valueobject.ModifierGroup,
	isNewProduct bool,
) error {
	if len(modifierGroup.Modifiers) == 0 {
		return nil
	}

	// 查找或创建属性组
	var attrGroup model.ProductAttributeGroup
	err := tx.Where("name = ? AND delete_time = 0", modifierGroup.Name).First(&attrGroup).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return errors.WithMessage(err, "查询属性组失败")
	}

	// 如果不存在，创建新属性组
	if attrGroup.Uuid == 0 {
		attrGroupUuid, err := utils.GetID()
		if err != nil {
			return errors.WithMessage(err, "生成属性组UUID失败")
		}

		// 创建多语言名称
		multiLanguageNameUuid, err := mi.createMultiLanguageName(tx, modifierGroup.NameTranslation, modifierGroup.Name)
		if err != nil {
			return errors.WithMessage(err, "创建属性组多语言名称失败")
		}

		attrGroup = model.ProductAttributeGroup{
			BaseModel: model.BaseModel{
				Uuid:       attrGroupUuid,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			Name:                  modifierGroup.Name,
			MultiLanguageNameUuid: multiLanguageNameUuid,
			Sort:                  modifierGroup.Sequence,
		}

		if err := tx.Create(&attrGroup).Error; err != nil {
			return errors.WithMessage(err, "创建属性组失败")
		}
	}

	// 创建属性组与商品的关联
	packageAttrGroupRepo := repository.NewProductPackageAttributeGroupRepo(tx)
	var existingPackageAttrGroup model.ProductPackageAttributeGroup
	err = tx.Where("product_package_uuid = ? AND product_attribute_group_uuid = ? AND delete_time = 0",
		productPackageUuid, attrGroup.Uuid).First(&existingPackageAttrGroup).Error

	var packageAttrGroupUuid uint64
	if err == gorm.ErrRecordNotFound {
		// 不存在，创建新关联
		packageAttrGroupUuid, err = utils.GetID()
		if err != nil {
			return errors.WithMessage(err, "生成商品属性组关联UUID失败")
		}

		// 判断是否必选
		isMust := 0
		if modifierGroup.SelectionRangeMin > 0 {
			isMust = 1
		}
		maxSelection := modifierGroup.SelectionRangeMax
		if maxSelection == 0 {
			maxSelection = len(modifierGroup.Modifiers)
		}

		packageAttrGroup := model.ProductPackageAttributeGroup{
			BaseModel: model.BaseModel{
				Uuid:       packageAttrGroupUuid,
				CreateTime: time.Now().Unix(),
				UpdateTime: time.Now().Unix(),
			},
			ProductPackageUuid:        productPackageUuid,
			ProductAttributeGroupUuid: attrGroup.Uuid,
			IsMust:                    uint(isMust),
			MaxSelection:              uint(maxSelection),
		}

		packageAttrGroups := []model.ProductPackageAttributeGroup{packageAttrGroup}
		if err := packageAttrGroupRepo.CreateProductPackageAttributeGroups(packageAttrGroups); err != nil {
			return errors.WithMessage(err, "创建商品属性组关联失败")
		}
	} else if err != nil {
		return errors.WithMessage(err, "查询商品属性组关联失败")
	} else {
		packageAttrGroupUuid = existingPackageAttrGroup.Uuid
	}

	// 创建或查找每个属性
	for _, modifier := range modifierGroup.Modifiers {
		// 查找是否已存在该属性
		var attr model.ProductAttribute
		err := tx.Where("name = ? AND attribute_group_uuid = ? AND delete_time = 0",
			modifier.Name, attrGroup.Uuid).First(&attr).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return errors.WithMessage(err, "查询属性失败")
		}

		// 如果不存在，创建新属性
		if attr.Uuid == 0 {
			attrUuid, err := utils.GetID()
			if err != nil {
				return errors.WithMessage(err, "生成属性UUID失败")
			}

			// 创建多语言名称
			multiLanguageNameUuid, err := mi.createMultiLanguageName(tx, modifier.NameTranslation, modifier.Name)
			if err != nil {
				return errors.WithMessage(err, "创建属性多语言名称失败")
			}

			attr = model.ProductAttribute{
				BaseModel: model.BaseModel{
					Uuid:       attrUuid,
					CreateTime: time.Now().Unix(),
					UpdateTime: time.Now().Unix(),
				},
				AttributeGroupUuid:    attrGroup.Uuid,
				Name:                  modifier.Name,
				MultiLanguageNameUuid: multiLanguageNameUuid,
				Sort:                  modifier.Sequence,
			}

			if err := tx.Create(&attr).Error; err != nil {
				return errors.WithMessage(err, "创建属性失败")
			}
		}

		// 创建属性与商品的关联
		packageAttrRepo := repository.NewProductPackageAttributeRepo(tx)
		var existingPackageAttr model.ProductPackageAttribute
		err = tx.Where("product_package_attribute_group_uuid = ? AND attribute_uuid = ? AND delete_time = 0",
			packageAttrGroupUuid, attr.Uuid).First(&existingPackageAttr).Error

		if err == gorm.ErrRecordNotFound {
			// 不存在，创建新关联
			packageAttrUuid, err := utils.GetID()
			if err != nil {
				return errors.WithMessage(err, "生成商品属性关联UUID失败")
			}

			packageAttr := model.ProductPackageAttribute{
				BaseModel: model.BaseModel{
					Uuid:       packageAttrUuid,
					CreateTime: time.Now().Unix(),
					UpdateTime: time.Now().Unix(),
				},
				ProductPackageAttributeGroupUuid: packageAttrGroupUuid,
				AttributeUuid:                    attr.Uuid,
				IsDefaultSelected:                0,
			}

			packageAttrs := []model.ProductPackageAttribute{packageAttr}
			if err := packageAttrRepo.CreateProductPackageAttributes(packageAttrs); err != nil {
				return errors.WithMessage(err, "创建商品属性关联失败")
			}
		} else if err != nil {
			return errors.WithMessage(err, "查询商品属性关联失败")
		}
		// 如果已存在，不做处理
	}

	return nil
}

// findOrCreateCategory 查找或创建分类（根据 Grab 分类 ID 去重）
func (mi *MenuImporter) findOrCreateCategory(tx *gorm.DB, companyUuid uint64, category *valueobject.Category) (uint64, error) {
	// 1. 先根据 Grab 分类 ID 查找
	categoryRepo := repository.NewProductCategoryRepo(tx)
	var existingCategory model.ProductCategory
	err := tx.Where("source = ? AND source_id = ? AND delete_time = 0", "grab", category.ID).
		First(&existingCategory).Error

	if err == nil {
		// 分类已存在，返回UUID
		logger.Logger.Info("Grab分类已存在，复用",
			zap.String("grab_category_id", category.ID),
			zap.String("category_name", category.Name),
			zap.Uint64("category_uuid", existingCategory.Uuid))
		return existingCategory.Uuid, nil
	}

	if err != gorm.ErrRecordNotFound {
		return 0, errors.WithMessage(err, "查询分类失败")
	}

	// 2. 分类不存在，创建新分类
	categoryUuid, err := utils.GetID()
	if err != nil {
		return 0, errors.WithMessage(err, "生成分类UUID失败")
	}

	// 创建多语言名称
	multiLanguageNameUuid, err := mi.createMultiLanguageName(tx, category.NameTranslation, category.Name)
	if err != nil {
		return 0, errors.WithMessage(err, "创建分类多语言名称失败")
	}

	newCategory := model.ProductCategory{
		BaseModel: model.BaseModel{
			Uuid:       categoryUuid,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
		Name:                  category.Name,
		Source:                "grab",      // 标记来源为 Grab
		SourceId:              category.ID, // 保存 Grab 分类 ID
		MultiLanguageNameUuid: multiLanguageNameUuid,
		Status:                1, // 启用
		IsDisplayInStore:      1, // 在店内显示
		IsDisplayInTakeout:    1, // 在外卖显示
		IsSpecial:             0, // 非特殊分类
		Sort:                  uint(category.Sequence),
	}

	if _, err := categoryRepo.CreateProductCategory(newCategory); err != nil {
		return 0, errors.WithMessage(err, "创建分类失败")
	}

	logger.Logger.Info("创建Grab分类成功",
		zap.String("grab_category_id", category.ID),
		zap.String("category_name", category.Name),
		zap.Uint64("category_uuid", categoryUuid))

	return categoryUuid, nil
}

// createMultiLanguageName 创建多语言名称
func (mi *MenuImporter) createMultiLanguageName(tx *gorm.DB, translations map[string]string, defaultName string) (uint64, error) {
	// 构建多语言JSON字符串
	nameMap := make(map[string]string)

	// 添加所有翻译
	for lang, translation := range translations {
		if translation != "" {
			nameMap[lang] = translation
		}
	}

	// 如果没有英文翻译，使用默认名称
	if _, ok := nameMap["en"]; !ok && defaultName != "" {
		nameMap["en"] = defaultName
	}

	// 转换为JSON字符串
	jsonBytes, err := json.Marshal(nameMap)
	if err != nil {
		return 0, errors.WithMessage(err, "序列化多语言名称失败")
	}

	// 创建多语言名称记录
	multiLanguageName := model.NewMultiLanguageName(string(jsonBytes))
	multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
	uuid, err := multiLanguageNameRepo.CreateMultiLanguageName(*multiLanguageName)
	if err != nil {
		return 0, errors.WithMessage(err, "创建多语言名称记录失败")
	}

	return uuid, nil
}

// contains 检查字符串是否包含任意一个子串
func contains(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			// 简单的子串检查
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}
