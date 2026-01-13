package service

import (
	"slices"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// IProductCheckSrv 定义产品检查服务接口
type IProductCheckSrv interface {
	CheckProductType(typ int) error                                                                                             // 检查商品类型
	CheckProductName(ctx context.Context, uuid uint64, localeName dto.LocaleResponse) error                                     // 检查商品名称
	CheckProductSellingPoint(ctx context.Context, localeName dto.LocaleResponse) error                                          // 检查商品卖点
	CheckMaterialName(ctx context.Context, uuid uint64, localeName dto.LocaleResponse) error                                    // 检查物品名称
	CheckProductCategory(db *gorm.DB, categoryUuid uint64) error                                                                // 检查商品分类
	CheckProductUnique(db *gorm.DB, uuid uint64) error                                                                          // 检查商品唯一性
	CheckProductFlavor(db *gorm.DB, flavors []CheckProductFlavorParam) (*CheckProductFlavorResult, error)                       // 检查商品规格
	CheckProductAttribute(db *gorm.DB, attributes []CheckProductAttributeGroupParam) ([]CheckProductAttributeGroupParam, error) // 检查商品属性
	CheckProductSauce(db *gorm.DB, param CheckProductSauceParam) (*CheckProductSauceResult, error)                              // 检查商品加料
	CheckProductTax(ctx context.Context, db *gorm.DB, taxReq CheckProductTaxParam) error                                        // 检查商品税类
	CheckProductStatus(status int) error                                                                                        // 检查商品状态
	CheckProductImage(ctx context.Context, db *gorm.DB, imageFileUuid uint64) error                                             // 检查商品图片
	CheckProductNumType(numType int) error                                                                                      // 检查商品计价方式
	CheckProductDeductStockType(deductStockType int) error                                                                      // 检查商品库存计算方式
	CheckProductShow(show CheckProductShowParam) error                                                                          // 检查商品显示设置
	CheckProductMemberDiscount(isEnableMemberDiscount int) error                                                                // 检查商品会员折扣
	CheckProductOverallDiscount(isEnableOverallDiscount int) error                                                              // 检查商品整单折扣
	CheckProductPackage(ctx context.Context, db *gorm.DB, param CheckProductPackageParam) (*CheckProductPackageResult, error)   // 检查商品套餐
	CheckProductPrinter(ctx context.Context, db *gorm.DB, productPackageUuid uint64, productPrinterUuids []uint64) error        // 检查商品打印机
}

// productCheckSrv 产品检查服务结构
type productCheckSrv struct {
	dbm        *database.DBManager // 数据库管理器
	localeSrv  ILocaleSrv          // 多语言名称服务
	settingSrv setting.ISrv        // 设置服务
}

// NewProductCheckSrv 创建产品检查服务实例
func NewProductCheckSrv(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv) IProductCheckSrv {
	return NewProductCheckSrvImpl(dbm, localeSrv, settingSrv)
}

// NewProductCheckSrvImpl 创建产品检查服务实例实现
func NewProductCheckSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv) IProductCheckSrv {
	return &productCheckSrv{
		dbm:        dbm,
		localeSrv:  localeSrv,
		settingSrv: settingSrv,
	}
}

func (s *productCheckSrv) CheckProductType(typ int) error {
	// 商品类型
	if !slices.Contains([]int{constant.ProductTypeProduct, constant.ProductTypePackage}, typ) {
		return errors.New("类型只能为商品或套餐")
	}
	return nil
}

func (s *productCheckSrv) CheckProductName(ctx context.Context, uuid uint64, localeName dto.LocaleResponse) error {
	// 商品名称
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !localeName.CheckRequiredLocale(storeLanguages) {
		return errors.New("名称不能为空")
	}
	if !localeName.CheckLenLocal(storeLanguages, 150) {
		return errors.New("名称长度不能超过150")
	}
	return nil
}

func (s *productCheckSrv) CheckProductSellingPoint(ctx context.Context, localeName dto.LocaleResponse) error {
	if localeName.IsNull() {
		return nil
	}
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !localeName.CheckLenLocal(storeLanguages, 500) {
		return errors.New("卖点长度不能超过500")
	}
	return nil
}

// 检查商品名称
func (s *productCheckSrv) CheckMaterialName(ctx context.Context, uuid uint64, localeName dto.LocaleResponse) error {
	// 物品名称
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !localeName.CheckRequiredLocale(storeLanguages) {
		return errors.New("名称不能为空")
	}
	if !localeName.CheckLenLocal(storeLanguages, 150) {
		return errors.New("名称长度不能超过150")
	}
	return nil
}

func (s *productCheckSrv) CheckProductCategory(db *gorm.DB, categoryUuid uint64) error {
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	// 商品分类
	if categoryUuid == 0 {
		return errors.New("分类不能为空")
	}
	category, err := productRepo.GetProductCategory(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereUuid(categoryUuid),
	)
	if err != nil || category.ID == 0 {
		return errors.New("分类不存在")
	}
	childCount, _ := productRepo.GetProductCategoryCount(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereParentUuid(category.Uuid),
	)
	if childCount > 0 {
		return errors.New("分类下有子分类，不能添加商品")
	}
	return nil
}

func (s *productCheckSrv) CheckProductUnique(db *gorm.DB, unitUuid uint64) error {
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	// 商品单位
	if unitUuid == 0 {
		return errors.New("单位不能为空")
	}
	unit, err := productRepo.GetProductUnit(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereUuid(unitUuid),
	)
	if err != nil || unit.ID == 0 {
		return errors.New("单位不存在")
	}
	return nil
}

type CheckProductFlavorParam struct {
	Uuid         uint64  `json:"uuid"`          // 商品规格UUID
	Price        float64 `json:"price"`         // 商品规格价格
	BarcodeValue string  `json:"barcode_value"` // 商品加料条码值
	InternalCode string  `json:"internal_code"` // 商品规格内部编码
	BomUuid      uint64  `json:"bom_uuid"`      // 商品bomUUID, 如果是新增，则传0，编辑或删除时传商品BOM UUID
	IsDelete     bool    `json:"is_delete"`     // 是否删除, 如果是新增/编辑，则传false，删除时传true
}

type CheckProductFlavorResult struct {
	MinPrice  float64                        `json:"min_price"`  // 最小价格
	MaxPrice  float64                        `json:"max_price"`  // 最大价格
	StockNum  float64                        `json:"stock_num"`  // 库存数量
	Status    int                            `json:"status"`     // 商品状态 0-下架 1-上架
	Flavors   []CheckProductFlavorItemResult `json:"flavors"`    // 商品规格列表
	IsPackage bool                           `json:"is_package"` // 是否套餐
}

type CheckProductFlavorItemResult struct {
	Uuid             uint64  `json:"uuid"`               // 商品规格UUID
	Name             string  `json:"name"`               // 商品规格名称
	BarcodeValue     string  `json:"barcode_value"`      // 商品规格条码值
	InternalCode     string  `json:"internal_code"`      // 商品规格内部编码
	Price            float64 `json:"price"`              // 商品规格价格
	BomUuid          uint64  `json:"bom_uuid"`           // 商品bomUUID, 如果是新增，则传0，编辑或删除时传商品BOM UUID
	IsDelete         bool    `json:"is_delete"`          // 是否删除, 如果是新增/编辑，则传false，删除时传true
	ErpnextGroupName string  `json:"erpnext_group_name"` // ERPNext规格组名称
	ErpnextValueName string  `json:"erpnext_value_name"` // ERPNext规格值名称
}

func (s *productCheckSrv) CheckProductFlavor(db *gorm.DB, flavors []CheckProductFlavorParam) (*CheckProductFlavorResult, error) {
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	// 商品规格
	if len(flavors) == 0 {
		return nil, errors.New("规格不能为空")
	}
	flavorCount := 0
	minPrice := 0.0
	maxPrice := 0.0
	flavorResults := make([]CheckProductFlavorItemResult, 0)

	// 已删除的条码值
	excludeBarcode := make([]string, 0)
	for _, flavor := range flavors {
		if flavor.IsDelete && flavor.BarcodeValue != "" {
			excludeBarcode = append(excludeBarcode, flavor.BarcodeValue)
		}
	}

	for _, flavorReq := range flavors {
		isDelete := flavorReq.IsDelete
		// 商品规格UUID
		if flavorReq.Uuid == 0 {
			return nil, errors.New("规格不能为空")
		}
		flavor, err := productRepo.GetProductFlavor(
			commonRepo.WhereBySoftDelete(),
			productRepo.WhereUuid(flavorReq.Uuid),
		)
		if err != nil || flavor.ID == 0 {
			return nil, errors.New("规格不存在")
		}
		if !isDelete && !productRepo.CheckPrice(flavorReq.Price, 0, 100000000, 2) {
			return nil, errors.New("规格价格范围错误")
		}
		// 商品规格条码值
		if !isDelete && flavorReq.BarcodeValue != "" {
			// 纯数字，最大13位
			if !productRepo.CheckBarcodeFormat(flavorReq.BarcodeValue) {
				return nil, errors.New("规格条码值格式不正确")
			}
			// 判断条码是否存在
			exists := productRepo.CheckBarcodeExist(flavorReq.BarcodeValue, flavorReq.BomUuid)
			// 不在已删除的条码值中，且数据库已存在
			if !slices.Contains(excludeBarcode, flavorReq.BarcodeValue) && exists {
				return nil, errors.New("规格条码值已存在")
			}
		}
		if minPrice == 0 {
			if !isDelete {
				minPrice = flavorReq.Price
			}
		} else {
			if !isDelete {
				minPrice = utils.IfFloat64(flavorReq.Price <= minPrice, flavorReq.Price, minPrice)
			}
		}
		if maxPrice == 0 {
			if !isDelete {
				maxPrice = flavorReq.Price
			}
		} else {
			if !isDelete {
				maxPrice = utils.IfFloat64(flavorReq.Price >= maxPrice, flavorReq.Price, maxPrice)
			}
		}
		flavorResults = append(flavorResults, CheckProductFlavorItemResult{
			Uuid:             flavorReq.Uuid,
			Name:             flavor.Name,
			BarcodeValue:     flavorReq.BarcodeValue,
			InternalCode:     flavorReq.InternalCode,
			Price:            flavorReq.Price,
			BomUuid:          flavorReq.BomUuid,
			IsDelete:         flavorReq.IsDelete,
			ErpnextGroupName: flavor.ErpnextGroupName,
			ErpnextValueName: flavor.ErpnextValueName,
		})
		if !isDelete {
			flavorCount++
		}
	}
	if flavorCount > 20 {
		return nil, errors.New("规格不能超过20个")
	}
	return &CheckProductFlavorResult{
		MinPrice: minPrice,
		MaxPrice: maxPrice,
		Flavors:  flavorResults,
	}, nil
}

type CheckProductAttributeGroupParam struct {
	Uuid               uint64                       `json:"uuid"`                 // 属性组UUID
	IsMust             int                          `json:"is_must"`              // 属性组是否必选 0-否 1-是（废弃，使用MinSelection替代）
	MinSelection       int                          `json:"min_selection"`        // 属性组最小选择数量（v2.12新增）
	IsOpenInput        bool                         `json:"is_open_input"`        // 属性组最大选择数量是否开启 false-否 true-是
	MaxSelection       int                          `json:"max_selection"`        // 属性组最大选择数量
	Attributes         []CheckProductAttributeParam `json:"attributes"`           // 属性列表
	IsDelete           bool                         `json:"is_delete"`            // 是否删除, 如果是新增/编辑，则传false，删除时传true
	ProductPackageUuid uint64                       `json:"product_package_uuid"` // 商品包UUID，用于编辑时查询原有值
	ClientVersion      string                       `json:"client_version"`       // 客户端版本号，用于兼容性判断
}

type CheckProductAttributeParam struct {
	Uuid              uint64 `json:"uuid"`                // 属性UUID
	IsDefaultSelected int    `json:"is_default_selected"` // 属性是否默认选中 0-否 1-是
	IsDelete          bool   `json:"is_delete"`           // 是否删除, 如果是新增/编辑，则传false，删除时传true
}

func (s *productCheckSrv) CheckProductAttribute(db *gorm.DB, attributes []CheckProductAttributeGroupParam) ([]CheckProductAttributeGroupParam, error) {
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	productPackageAttributeGroupRepo := repository.NewProductPackageAttributeGroupRepo(db)

	attributeGroupCount := 0
	for idx, attributeGroupReq := range attributes {
		isDelete := attributeGroupReq.IsDelete
		if attributeGroupReq.Uuid == 0 {
			return nil, errors.New("属性组不能为空")
		}
		attributeGroup, _ := productRepo.GetProductAttributeGroup(
			commonRepo.WhereBySoftDelete(),
			productRepo.WhereUuid(attributeGroupReq.Uuid),
		)
		if attributeGroup.ID == 0 {
			return nil, errors.New("属性组不存在")
		}

		// 统计属性数量（用于默认最大可选值）
		attributeCount := 0
		for _, attributeReq := range attributeGroupReq.Attributes {
			if !attributeReq.IsDelete {
				attributeCount++
			}
		}

		// ========== 版本兼容性处理开始 ==========
		// 判断客户端版本：v2.11.x 使用旧逻辑，v2.12+ 使用新逻辑
		clientVersion := attributeGroupReq.ClientVersion
		isV211Client := utils.CompareVersion(clientVersion, utils.VersionLT, "2.12.0")

		if isV211Client {
			// v2.11客户端 → v2.12后端：旧字段转新字段
			if attributeGroupReq.ProductPackageUuid > 0 {
				// 编辑商品：查询原有的 min_selection 值
				existingGroup, _ := productPackageAttributeGroupRepo.GetProductPackageAttributeGroup(
					commonRepo.WhereByProductPackageUuid(attributeGroupReq.ProductPackageUuid),
					commonRepo.WhereByProductAttributeGroupUuid(attributeGroupReq.Uuid),
					commonRepo.WhereBySoftDelete(),
				)

				if attributeGroupReq.IsMust == 0 {
					// 取消必选：重置为0
					attributes[idx].MinSelection = 0
				} else if attributeGroupReq.IsMust == 1 {
					// 勾选必选：保留原有的最小选择数量（如果用户之前设置过）
					if attributeGroupReq.MinSelection == 0 {
						// 前端未传 min_selection，需要判断
						if existingGroup != nil && existingGroup.MinSelection > 0 {
							// 保留数据库中的原值
							attributes[idx].MinSelection = int(existingGroup.MinSelection)
						} else {
							// 原值为0或首次设置，设置为1
							attributes[idx].MinSelection = 1
						}
					}
					// else: 前端已传 min_selection > 0，保留前端值
				}
			} else {
				// 新增商品
				if attributeGroupReq.MinSelection == 0 && attributeGroupReq.IsMust == 1 {
					attributes[idx].MinSelection = 1
				}
			}

			// 处理 max_selection 默认值
			// 考虑 is_open_input 字段：关闭最大可选功能 或 未设置最大可选值
			if !attributeGroupReq.IsOpenInput || attributeGroupReq.MaxSelection == 0 {
				// 未开启最大可选限制 或 未设置最大可选值，默认为属性值数量（表示不限制）
				attributes[idx].MaxSelection = attributeCount
			}
		}

		// v2.12客户端 → v2.12后端：新字段自动计算旧字段（所有版本都执行）
		if attributes[idx].MinSelection > 0 {
			attributes[idx].IsMust = 1
		} else {
			attributes[idx].IsMust = 0
		}
		// ========== 版本兼容性处理结束 ==========

		// 验证可选范围（v2.12新增）
		if attributes[idx].MaxSelection < attributes[idx].MinSelection {
			return nil, errors.New("属性组最大可选不可小于最小可选")
		}

		if !slices.Contains([]int{0, 1}, attributes[idx].IsMust) {
			return nil, errors.New("属性组是否必选不正确")
		}
		if len(attributeGroupReq.Attributes) == 0 {
			return nil, errors.New("属性值不能为空")
		}
		attributeDefaultCount := 0
		for _, attributeReq := range attributeGroupReq.Attributes {
			if attributeReq.Uuid == 0 {
				return nil, errors.New("属性值不能为空")
			}
			attribute, _ := productRepo.GetProductAttribute(
				commonRepo.WhereBySoftDelete(),
				productRepo.WhereUuid(attributeReq.Uuid),
			)
			if attribute.ID == 0 {
				return nil, errors.New("属性值不存在")
			}
			if !slices.Contains([]int{0, 1}, attributeReq.IsDefaultSelected) {
				return nil, errors.New("属性是否默认选中不正确")
			}
			if !attributeReq.IsDelete {
				if attributeReq.IsDefaultSelected == 1 {
					attributeDefaultCount++
				}
				attributeCount++
			}
		}
		if !isDelete {
			// 验证最大可选不超过属性值数量
			if attributes[idx].MaxSelection > attributeCount {
				return nil, errors.New("属性最大可选数量不能超过属性值数量")
			}
			if attributeGroupReq.IsOpenInput && attributeDefaultCount > attributes[idx].MaxSelection {
				return nil, errors.New("默认勾选数量不能大于最大选择数量")
			}
			attributeGroupCount++
		}
	}
	if attributeGroupCount > 100 {
		return nil, errors.New("属性组不能超过100个")
	}
	return attributes, nil
}

type CheckProductSauceParam struct {
	IsMust             int                          `json:"is_must"`              // 商品加料是否必选 0-否 1-是（废弃，使用MinSelection替代）
	MinSelection       int                          `json:"min_selection"`        // 商品加料最小选择数量（v2.12新增）
	IsOpenInput        bool                         `json:"is_open_input"`        // 商品加料最大选择数量是否开启 0-否 1-是
	MaxSelection       int                          `json:"max_selection"`        // 商品加料最大选择数量
	Sauces             []CheckProductSauceItemParam `json:"items"`                // 商品加料列表
	ProductPackageUuid uint64                       `json:"product_package_uuid"` // 商品包UUID，用于编辑时查询原有值
	ClientVersion      string                       `json:"client_version"`       // 客户端版本号，用于兼容性判断
}

type CheckProductSauceItemParam struct {
	Uuid              uint64   `json:"uuid"`                // 商品加料项UUID
	IsDefaultSelected int      `json:"is_default_selected"` // 商品加料项是否默认选中 0-否 1-是
	BomUuid           uint64   `json:"bom_uuid"`            // 商品加料项bomUUID, 如果是新增，则传0，编辑或删除时传商品BOM UUID
	IsDelete          bool     `json:"is_delete"`           // 是否删除, 如果是新增/编辑，则传false，删除时传true
	Price             *float64 `json:"price"`               // 商品加料价格，可不传递，兼容旧版客户端，v2.14后必传
}

type CheckProductSauceResult struct {
	IsMust       int                           `json:"is_must"`       // 商品加料是否必选 0-否 1-是（v2.11废弃）
	MinSelection int                           `json:"min_selection"` // 商品加料最小选择数量（v2.12新增）
	MaxSelection int                           `json:"max_selection"` // 商品加料最大选择数量
	Status       int                           `json:"status"`        // 商品状态 0-下架 1-上架
	Sauces       []CheckProductSauceItemResult `json:"sauces"`        // 商品加料列表
}

type CheckProductSauceItemResult struct {
	Uuid              uint64  `json:"uuid"`                // 商品小料UUID
	Name              string  `json:"name"`                // 商品小料名称
	Price             float64 `json:"price"`               // 商品小料价格
	IsDefaultSelected int     `json:"is_default_selected"` // 商品小料是否默认选中 0-否 1-是
	BomUuid           uint64  `json:"bom_uuid"`            // 商品小料bomUUID, 如果是新增，则传0，编辑或删除时传商品BOM UUID
	IsDelete          bool    `json:"is_delete"`           // 是否删除, 如果是新增/编辑，则传false，删除时传true
}

func (s *productCheckSrv) CheckProductSauce(db *gorm.DB, param CheckProductSauceParam) (*CheckProductSauceResult, error) {
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	productPackageRepo := repository.NewProductPackageRepo(db)

	// 统计小料数量（用于默认最大可选值）
	sauceCount := 0
	for _, sauceReq := range param.Sauces {
		if !sauceReq.IsDelete {
			sauceCount++
		}
	}

	// ========== 版本兼容性处理开始 ==========
	// 判断客户端版本：v2.11.x 使用旧逻辑，v2.12+ 使用新逻辑
	clientVersion := param.ClientVersion
	isV211Client := utils.CompareVersion(clientVersion, utils.VersionLT, "2.12.0")

	if isV211Client {
		// v2.11客户端 → v2.12后端：旧字段转新字段
		if param.ProductPackageUuid > 0 {
			// 编辑商品：查询原有的 sauce_min_selection 值
			existingPackage, _ := productPackageRepo.GetProductPackage(
				commonRepo.WhereByUuid(param.ProductPackageUuid),
				commonRepo.WhereBySoftDelete(),
			)

			if param.IsMust == 0 {
				// 取消必选：重置为0
				param.MinSelection = 0
			} else if param.IsMust == 1 {
				// 勾选必选：保留原有的最小选择数量（如果用户之前设置过）
				if param.MinSelection == 0 {
					// 前端未传 min_selection，需要判断
					if existingPackage != nil && existingPackage.SauceMinSelection > 0 {
						// 保留数据库中的原值
						param.MinSelection = int(existingPackage.SauceMinSelection)
					} else {
						// 原值为0或首次设置，设置为1
						param.MinSelection = 1
					}
				}
				// else: 前端已传 min_selection > 0，保留前端值
			}
		} else {
			// 新增商品
			if param.MinSelection == 0 && param.IsMust == 1 {
				param.MinSelection = 1
			}
		}

		// 处理 max_selection 默认值
		// 考虑 is_open_input 字段：关闭最大可选功能 或 未设置最大可选值
		if !param.IsOpenInput || param.MaxSelection == 0 {
			// 未开启最大可选限制 或 未设置最大可选值，默认为加料值数量（表示不限制）
			param.MaxSelection = sauceCount
		}

	}

	// v2.12客户端 → v2.12后端：新字段自动计算旧字段（所有版本都执行）
	if param.MinSelection > 0 {
		param.IsMust = 1
	} else {
		param.IsMust = 0
	}
	// ========== 版本兼容性处理结束 ==========

	// 验证可选范围（v2.12新增）
	if param.MaxSelection < param.MinSelection {
		return nil, errors.New("加料最大可选不可小于最小可选")
	}

	if !slices.Contains([]int{0, 1}, param.IsMust) {
		return nil, errors.New("是否必选不正确")
	}
	sauceDefaultCount := 0
	sauceResults := make([]CheckProductSauceItemResult, 0)
	for _, sauceReq := range param.Sauces {
		if sauceReq.Uuid == 0 {
			return nil, errors.New("加料不能为空")
		}
		sauce, _ := productRepo.GetProductSauce(
			commonRepo.WhereBySoftDelete(),
			productRepo.WhereUuid(sauceReq.Uuid),
		)
		if sauce.ID == 0 {
			return nil, errors.New("加料不存在")
		}
		price := sauce.Price
		if sauceReq.Price != nil {
			price = *sauceReq.Price
		}
		sauceResults = append(sauceResults, CheckProductSauceItemResult{
			Uuid:              sauce.Uuid,
			Name:              sauce.Name,
			Price:             price,
			IsDefaultSelected: sauceReq.IsDefaultSelected,
			BomUuid:           sauceReq.BomUuid,
			IsDelete:          sauceReq.IsDelete,
		})
		if !sauceReq.IsDelete {
			if sauceReq.IsDefaultSelected == 1 {
				sauceDefaultCount++
			}
		}
	}

	// 重新统计实际的 sauceCount（用于验证）
	actualSauceCount := len(sauceResults)
	for _, sr := range sauceResults {
		if sr.IsDelete {
			actualSauceCount--
		}
	}

	if param.IsOpenInput && param.MaxSelection > actualSauceCount {
		return nil, errors.New("加料项数量不能大于最大选择数量")
	}
	if param.IsOpenInput && sauceDefaultCount > param.MaxSelection {
		return nil, errors.New("默认勾选数量不能大于最大选择数量")
	}
	if actualSauceCount > 100 {
		return nil, errors.New("加料不能超过100个")
	}
	return &CheckProductSauceResult{
		IsMust:       param.IsMust,
		MinSelection: param.MinSelection,
		MaxSelection: param.MaxSelection,
		Sauces:       sauceResults,
	}, nil
}

type CheckProductTaxParam struct {
	DineUuid    uint64 `json:"dine_uuid"`    // 堂食税类UUID
	TakeoutUuid uint64 `json:"takeout_uuid"` // 外带税类UUID
}

func (s *productCheckSrv) CheckProductTax(ctx context.Context, db *gorm.DB, taxReq CheckProductTaxParam) error {
	commonRepo := repository.NewCommonRepo()
	taxRepo := repository.NewTaxRepo(db)
	// 税率
	taxRateSetting, err := s.settingSrv.GetTaxRateSetting(ctx)
	if err != nil {
		return errors.WithMessage(err, "获取税率设置失败")
	}
	if taxRateSetting.IsOpen == "1" {
		if taxReq.DineUuid == 0 {
			return errors.New("堂食税类不能为空")
		}
		dineTax, err := taxRepo.GetTaxCategory(
			commonRepo.WhereBySoftDelete(),
			commonRepo.WhereByUuid(taxReq.DineUuid),
		)
		if err != nil {
			return errors.WithMessage(err, "获取堂食税类失败")
		}
		if dineTax.ID == 0 {
			return errors.New("堂食税类不存在")
		}
		if taxReq.TakeoutUuid == 0 {
			return errors.New("外带税类不能为空")
		}
		takeoutTax, err := taxRepo.GetTaxCategory(
			commonRepo.WhereBySoftDelete(),
			commonRepo.WhereByUuid(taxReq.TakeoutUuid),
		)
		if err != nil {
			return errors.WithMessage(err, "获取外带税类失败")
		}
		if takeoutTax.ID == 0 {
			return errors.New("外带税类不存在")
		}
	}

	return nil
}

func (s *productCheckSrv) CheckProductStatus(status int) error {
	if !slices.Contains([]int{constant.ProductStatusOnSale, constant.ProductStatusOffSale}, status) {
		return errors.New("商品状态不正确")
	}
	return nil
}

func (s *productCheckSrv) CheckProductImage(ctx context.Context, db *gorm.DB, imageFileUuid uint64) error {
	uploadFileSrv := NewUploadFileSrv(s.dbm)
	uploadFile, err := uploadFileSrv.GetUploadFile(ctx, imageFileUuid)
	if err != nil {
		return errors.WithMessage(err, "获取商品图片失败")
	}
	if uploadFile.Uuid == 0 {
		return errors.New("商品图片不存在")
	}

	return nil
}

func (s *productCheckSrv) CheckProductNumType(numType int) error {
	if !slices.Contains([]int{constant.ProductNumTypeInteger, constant.ProductNumTypeDecimal}, numType) {
		return errors.New("计价方式不正确")
	}
	return nil
}

func (s *productCheckSrv) CheckProductDeductStockType(deductStockType int) error {
	if !slices.Contains([]int{constant.ProductPackageDeductStockTypePay, constant.ProductPackageDeductStockTypeCooking}, deductStockType) {
		return errors.New("库存计算方式不正确")
	}
	return nil
}

type CheckProductShowParam struct {
	IsShowCashier   int `json:"is_show_cashier"`   // 是否显示在收银端 0-不显示 1-显示
	IsShowTablet    int `json:"is_show_tablet"`    // 是否显示在平板端 0-不显示 1-显示
	IsShowKitchen   int `json:"is_show_kitchen"`   // 是否显示在厨显端 0-不显示 1-显示
	IsShowAssistant int `json:"is_show_assistant"` // 是否显示在点餐助手 0-不显示 1-显示
	IsShowH5        int `json:"is_show_h5"`        // 是否显示在h5 0-不显示 1-显示
	IsShowDelivery  int `json:"is_show_delivery"`  // 是否显示在外送 0-不显示 1-显示
	IsShowKiosk     int `json:"is_show_kiosk"`     // 是否在自助点餐机显示 0-否 1-是
}

func (s *productCheckSrv) CheckProductShow(show CheckProductShowParam) error {
	// 商品显示设置
	if show.IsShowCashier != 0 && show.IsShowCashier != 1 {
		return errors.New("是否显示在收银端不正确")
	}
	if show.IsShowTablet != 0 && show.IsShowTablet != 1 {
		return errors.New("是否显示在平板端不正确")
	}
	if show.IsShowKitchen != 0 && show.IsShowKitchen != 1 {
		return errors.New("是否显示在厨显端不正确")
	}
	if show.IsShowAssistant != 0 && show.IsShowAssistant != 1 {
		return errors.New("是否显示在点餐助手不正确")
	}
	if show.IsShowH5 != 0 && show.IsShowH5 != 1 {
		return errors.New("是否显示在h5不正确")
	}
	if show.IsShowDelivery != 0 && show.IsShowDelivery != 1 {
		return errors.New("是否显示在外送不正确")
	}
	if show.IsShowKiosk != 0 && show.IsShowKiosk != 1 {
		return errors.New("是否在自助点餐机显示不正确")
	}
	return nil
}

func (s *productCheckSrv) CheckProductMemberDiscount(isEnableMemberDiscount int) error {
	if isEnableMemberDiscount != 0 && isEnableMemberDiscount != 1 {
		return errors.New("是否开启会员折扣不正确")
	}
	return nil
}

func (s *productCheckSrv) CheckProductOverallDiscount(isEnableOverallDiscount int) error {
	if isEnableOverallDiscount != 0 && isEnableOverallDiscount != 1 {
		return errors.New("是否开启整单折扣不正确")
	}
	return nil
}

type CheckProductPackageParam struct {
	Price         float64                         `json:"price"`          // 套餐价格
	Groups        []CheckProductPackageGroupParam `json:"groups"`         // 套餐分组列表
	ClientVersion string                          `json:"client_version"` // 客户端版本号，用于兼容性判断
}

type CheckProductPackageGroupParam struct {
	Uuid             uint64                                 `json:"uuid"`               // 套餐组UUID
	LocaleName       dto.LocaleResponse                     `json:"locale_name"`        // 套餐组名称
	GroupType        int                                    `json:"group_type"`         // 分组类型 0-固定 1-可选
	OptionalMinCount int                                    `json:"optional_min_count"` // 最小可选数量（v2.12新增）
	OptionalCount    int                                    `json:"optional_count"`     // 最大可选数量（字段名不变，v2.12语义改为最大可选）
	IsDelete         bool                                   `json:"is_delete"`          // 是否删除, 如果是新增/编辑，则传false，删除时传true
	Products         []CheckProductPackageGroupProductParam `json:"products"`           // 商品套餐组商品列表
}

type CheckProductPackageGroupProductParam struct {
	Uuid       uint64  `json:"uuid"`        // 套餐商品UUID
	BomUuid    uint64  `json:"bom_uuid"`    // 商品BOM UUID
	Num        float64 `json:"num"`         // 商品数量
	Sort       int     `json:"sort"`        // 商品排序
	AddPrice   float64 `json:"add_price"`   // 加价金额，默认0
	IsRequired int     `json:"is_required"` // 必选 0-不必选 1-必选
	IsDefault  int     `json:"is_default"`  // 默认选中 0-默认不选中 1-默认选中
	IsDelete   bool    `json:"is_delete"`   // 是否删除, 如果是新增/编辑，则传false，删除时传true
}

type CheckProductPackageResult struct {
	Price    float64                          `json:"price"`     // 套餐价格
	StockNum float64                          `json:"stock_num"` // 套餐库存
	Groups   []CheckProductPackageGroupResult `json:"groups"`    // 套餐分组列表
}

type CheckProductPackageGroupResult struct {
	Uuid             uint64                                  `json:"uuid"`               // 套餐组UUID
	LocaleName       dto.LocaleResponse                      `json:"locale_name"`        // 套餐组名称
	GroupType        int                                     `json:"group_type"`         // 分组类型 0-固定 1-可选
	OptionalMinCount int                                     `json:"optional_min_count"` // 最小可选数量（v2.12新增）
	OptionalCount    int                                     `json:"optional_count"`     // 最大可选数量（v2.12语义变更）
	IsDelete         bool                                    `json:"is_delete"`          // 是否删除, 如果是新增/编辑，则传false，删除时传true
	Products         []CheckProductPackageGroupProductResult `json:"products"`           // 商品套餐组商品列表
}

type CheckProductPackageGroupProductResult struct {
	Uuid       uint64  `json:"uuid"`        // 套餐商品UUID
	BomUuid    uint64  `json:"bom_uuid"`    // 商品BOM UUID
	Num        float64 `json:"num"`         // 商品数量
	Sort       int     `json:"sort"`        // 商品排序
	AddPrice   float64 `json:"add_price"`   // 加价金额，默认0
	IsRequired int     `json:"is_required"` // 必选 0-不必选 1-必选
	IsDefault  int     `json:"is_default"`  // 默认选中 0-默认不选中 1-默认选中
	IsDelete   bool    `json:"is_delete"`   // 是否删除, 如果是新增/编辑，则传false，删除时传true
}

func (s *productCheckSrv) CheckProductPackage(ctx context.Context, db *gorm.DB, param CheckProductPackageParam) (*CheckProductPackageResult, error) {
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)
	productPackageGroupRepo := repository.NewProductPackageGroupRepo(db)
	storeLanguages, _ := s.settingSrv.GetStoreLanguage(ctx)
	if !productRepo.CheckPrice(param.Price, 0, 100000000, 2) {
		return nil, errors.New("套餐价格范围错误")
	}
	if len(param.Groups) == 0 {
		return nil, errors.New("分组不能为空")
	}

	isV211Client := utils.CompareVersion(param.ClientVersion, utils.VersionLT, "2.12.0")

	var count int64 = 0
	var stockNum float64 = 0
	groups := make([]CheckProductPackageGroupResult, 0)
	for _, group := range param.Groups {
		var productPackageGroup *model.ProductPackageGroup
		// 判断套餐分组是否存在
		if group.Uuid != 0 {
			productPackageGroup, _ = productPackageGroupRepo.GetProductPackageGroup(
				commonRepo.WhereBySoftDelete(),
				commonRepo.WhereByUuid(group.Uuid),
			)
			if productPackageGroup.ID == 0 {
				return nil, errors.New("套餐分组不存在")
			}
		}
		isDelete := group.IsDelete
		if !isDelete && !group.LocaleName.CheckRequiredLocale(storeLanguages) {
			return nil, errors.New("分组名称不能为空")
		}
		if !isDelete && !group.LocaleName.CheckLenLocal(storeLanguages, 150) {
			return nil, errors.New("分组名称长度不能超过150")
		}
		if !isDelete && len(group.Products) == 0 {
			return nil, errors.New("商品不能为空")
		}
		// 验证分组类型
		if !isDelete {
			if group.GroupType != 0 && group.GroupType != 1 {
				return nil, errors.New("分组类型必须为0（固定）或1（可选）")
			}

			// 可选分组：验证可选范围（v2.12新增）
			if group.GroupType == 1 {
				if isV211Client {
					if productPackageGroup != nil {
						group.OptionalMinCount = productPackageGroup.OptionalCount
					} else {
						group.OptionalMinCount = 0
					}
				}
				// 验证可选范围
				if group.OptionalCount < group.OptionalMinCount {
					return nil, errors.New("分组最大可选不可小于最小可选")
				}
			}
		}
		products := make([]CheckProductPackageGroupProductResult, 0)
		requiredOrDefaultCount := 0 // 统计必选或默认的商品数量（去重）
		for _, product := range group.Products {
			if isDelete {
				product.IsDelete = true
			} else {
				if product.Uuid != 0 {
					productPackageGroupItem, err := productPackageGroupRepo.GetProductPackageGroupItem(
						commonRepo.WhereBySoftDelete(),
						commonRepo.WhereByProductPackageGroupUuid(group.Uuid),
						commonRepo.WhereByUuid(product.Uuid),
					)
					if err != nil || productPackageGroupItem.ID == 0 {
						return nil, errors.New("套餐分组商品不存在")
					}
				}
				if !product.IsDelete {
					// 验证必选和默认选中字段
					if product.IsRequired != 0 && product.IsRequired != 1 {
						return nil, errors.New("必选字段必须为0（不必选）或1（必选）")
					}
					if product.IsDefault != 0 && product.IsDefault != 1 {
						return nil, errors.New("默认选中字段必须为0（不默认选中）或1（默认选中）")
					}
					// 统计必选或默认的商品数量（去重：一个商品既是必选又是默认，只算一个）
					if product.IsRequired == 1 || product.IsDefault == 1 {
						requiredOrDefaultCount++
					}
					// 验证加价金额
					if product.AddPrice < 0 {
						return nil, errors.New("加价金额不能为负数")
					}
					bom, _ := productRepo.GetProductBom(
						commonRepo.WhereBySoftDelete(),
						productRepo.WhereUuid(product.BomUuid),
					)
					if bom.ID == 0 {
						return nil, errors.New("商品规格不存在")
					}
					if bom.StockNum >= float64(product.Num) {
						currentStockNum := decimal.NewFromFloat(bom.StockNum).Div(decimal.NewFromFloat(float64(product.Num))).Floor()
						if stockNum == 0 {
							stockNum = currentStockNum.InexactFloat64()
						} else {
							if currentStockNum.LessThan(decimal.NewFromFloat(stockNum)) {
								stockNum = currentStockNum.InexactFloat64()
							}
						}
					}
				}
			}
			products = append(products, CheckProductPackageGroupProductResult(product))
		}
		// 验证必选或默认商品数量不能大于可选数量
		if !isDelete && group.GroupType == 1 {
			if requiredOrDefaultCount > group.OptionalCount {
				return nil, errors.New("必选或默认不可大于可选数量")
			}
		}
		groups = append(groups, CheckProductPackageGroupResult{
			Uuid:             group.Uuid,
			LocaleName:       group.LocaleName,
			GroupType:        group.GroupType,
			OptionalMinCount: group.OptionalMinCount,
			OptionalCount:    group.OptionalCount,
			IsDelete:         group.IsDelete,
			Products:         products,
		})
		if !isDelete {
			count++
		}
	}

	if count > 100 {
		return nil, errors.New("分组不能超过100个")
	}

	return &CheckProductPackageResult{
		Price:    param.Price,
		StockNum: stockNum,
		Groups:   groups,
	}, nil
}

func (s *productCheckSrv) CheckProductPrinter(ctx context.Context, db *gorm.DB, productPackageUuid uint64, productPrinterUuids []uint64) error {
	productPrinterRepo := repository.NewProductPrinterRepo(db)
	// 如果商品打印机列表为空，则直接返回
	if len(productPrinterUuids) == 0 {
		return nil
	}
	// 获取商品打印机列表
	productPrinterList, err := productPrinterRepo.GetProductPrinters()
	if err != nil {
		return errors.New("获取商品打印机列表失败")
	}

	// 将商品打印机列表转换为UUID列表，方便查找
	printProductSelect := 0
	existingPrinterUuids := make([]uint64, 0, len(productPrinterList))
	for _, productPrinter := range productPrinterList {
		existingPrinterUuids = append(existingPrinterUuids, productPrinter.Uuid)
		// 获取打印商品选择
		if slices.Contains(productPrinterUuids, productPrinter.Uuid) && productPrinter.PrintProductSelect == 1 {
			printProductSelect = productPrinter.PrintProductSelect
		}
	}

	// 判断 productPrinterUuids 中的每个UUID是否都在数据库中存在
	for _, productPrinterUuid := range productPrinterUuids {
		if !slices.Contains(existingPrinterUuids, productPrinterUuid) {
			return errors.New("商品打印档口不存在")
		}
	}

	// 判断 productPackageUuid 是否在商品包列表中
	if printProductSelect != 0 {
		productPackage, err := repository.NewProductPackageRepo(db).GetProductPackage(repository.CommonRepo.WhereByUuid(productPackageUuid))
		if err != nil {
			return errors.New("未设置打印标签")
		}
		if productPackage.ID == 0 {
			return errors.New("商品包不存在")
		}
		if productPackage.PrinterTagUuid == 0 {
			return errors.New("未设置打印标签")
		}
	}

	return nil
}
