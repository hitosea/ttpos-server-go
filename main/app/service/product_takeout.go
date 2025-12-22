package service

import (
	"fmt"
	"sync"
	"time"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	product_resp "ttpos-server-go/app/dto/resp/product_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IProductTakeoutSrv interface {
	// 外卖商品管理
	AddProductTakeoutShop(ctx context.Context, req req.ProductTakeoutShopAddReq) (*model.ProductPackageTakeout, error)
	EditProductTakeoutShop(ctx context.Context, req req.ProductTakeoutShopEditReq) error
	GetProductTakeoutShopDetail(ctx context.Context, req req.ProductTakeoutShopDetailReq) (*product_resp.ProductTakeoutShopDetailResp, error)
	DeleteProductTakeoutShop(ctx context.Context, req req.ProductTakeoutShopDeleteReq) error
	UpdateProductTakeoutShopStatus(ctx context.Context, req req.ProductTakeoutShopStatusReq) error

	// 批量操作
	BatchCreateProducts(ctx context.Context, req req.TakeoutBatchCreateReq) (*product_resp.TakeoutBatchResp, error)
	BatchOnlineProducts(ctx context.Context, req req.TakeoutBatchOnlineReq) (*product_resp.TakeoutBatchResp, error)
	BatchOfflineProducts(ctx context.Context, req req.TakeoutBatchOfflineReq) (*product_resp.TakeoutBatchResp, error)
	BatchDeleteProducts(ctx context.Context, req req.TakeoutBatchDeleteReq) (*product_resp.TakeoutBatchResp, error)

	// GetProductCount 获取外卖商品统计
	GetProductCount(ctx context.Context, companyUuid uint64, platform string, forceRefresh bool) (int64, error)
}

type productTakeoutSrv struct {
	dbm        *database.DBManager
	cache      cache.Cache
	localeSrv  ILocaleSrv
	productSrv IProductSrv
}

func NewProductTakeoutSrv(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv, cache cache.Cache, translateSrv ITranslateSrv) IProductTakeoutSrv {
	return &productTakeoutSrv{
		dbm:        dbm,
		cache:      cache,
		localeSrv:  localeSrv,
		productSrv: NewProductSrv(dbm, localeSrv, settingSrv, cache, translateSrv),
	}
}

// AddProductTakeoutShop 添加外卖商品
func (s *productTakeoutSrv) AddProductTakeoutShop(ctx context.Context, addReq req.ProductTakeoutShopAddReq) (*model.ProductPackageTakeout, error) {
	companySetting := ctx.GetCompanySetting()
	db := ctx.GetDB()

	// 设置默认外卖类型
	if addReq.TakeoutType == 0 {
		addReq.TakeoutType = constant.TakeoutTypeGrab
	}

	// 检查商品包是否存在
	productPackage, err := repository.NewProductPackageRepo(db).GetProductPackage(
		repository.CommonRepo.WhereByUuid(addReq.ProductPackageUuid),
		repository.CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return nil, errors.WithMessage(errors.New("商品不存在"))
	}
	if productPackage.IsDelete() {
		return nil, errors.WithMessage(errors.New("商品已删除"))
	}

	// 检查是否是总部商品，并且当前不是总店，则总部商品不能添加为外卖商品
	if productPackage.HeadquarterUuid != 0 && !companySetting.IsHeadquarter() {
		return nil, errors.WithMessage(errors.New("总部商品不能添加为外卖商品"))
	}

	// 检查是否已存在同类型外卖商品（包括软删除的记录）
	takeoutRepo := repository.NewProductPackageTakeoutRepo(db)
	existingTakeout, err := takeoutRepo.GetProductPackageTakeoutIncludeSoftDelete(addReq.ProductPackageUuid, uint(addReq.TakeoutType))

	// 如果存在未删除的记录，返回错误
	if err == nil && existingTakeout.DeleteTime == 0 {
		return existingTakeout, nil
	}

	// 如果存在已软删除的记录，还原它
	if err == nil && existingTakeout.DeleteTime != 0 {
		return s.restoreProductTakeoutShop(ctx, existingTakeout, addReq)
	}

	// 生成UUID
	uuid, err := utils.GetID()
	if err != nil {
		return nil, errors.WithMessage(err, "生成UUID失败")
	}

	// 处理多语言名称
	var multiLanguageNameUuid uint64
	var productName string

	if !addReq.LocaleName.IsNull() {
		productName = addReq.LocaleName.ToJson()
		multiLanguageName := model.NewMultiLanguageName(productName)
		multiLanguageNameUuid, err = repository.NewMultiLanguageNameRepo(db).CreateMultiLanguageName(*multiLanguageName)
		if err != nil {
			return nil, errors.WithMessage(err, "创建多语言名称失败")
		}
	} else {
		// 未提供自定义名称，使用店内商品的名称
		multiLanguageNameUuid = productPackage.MultiLanguageNameUuid
		productName = productPackage.Name
	}

	// 处理卖点多语言
	var describeMultiLanguageNameUuid uint64
	var describe string

	if !addReq.Describe.IsNull() {
		describe = addReq.Describe.ToJson()
		describeMultiLanguageName := model.NewMultiLanguageName(describe)
		describeMultiLanguageNameUuid, err = repository.NewMultiLanguageNameRepo(db).CreateMultiLanguageName(*describeMultiLanguageName)
		if err != nil {
			return nil, errors.WithMessage(err, "创建卖点多语言失败")
		}
	} else {
		// 未提供自定义卖点，使用店内商品的卖点
		describeMultiLanguageNameUuid = productPackage.DescribeMultiLanguageNameUuid
		describe = productPackage.Describe
	}

	// 创建外卖商品
	productPackageTakeout := &model.ProductPackageTakeout{
		BaseModel: model.BaseModel{
			Uuid: uuid,
		},
		ProductPackageUuid:            addReq.ProductPackageUuid,
		MultiLanguageNameUuid:         multiLanguageNameUuid,
		DescribeMultiLanguageNameUuid: describeMultiLanguageNameUuid,
		HeadquarterUuid:               productPackage.HeadquarterUuid,
		Name:                          productName,
		Describe:                      describe,
		ProductType:                   uint(productPackage.ProductType),
		TakeoutType:                   uint(addReq.TakeoutType),
		Status:                        uint(addReq.Status),
		CategoryUuid:                  addReq.CategoryUuid,
		SpecialCategoryUuid:           addReq.SpecialCategoryUuid,
		ImageFileUuid:                 addReq.ImageFileUuid,
		Source:                        addReq.Source,
		SourceProductId:               addReq.SourceProductId,
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// 创建外卖商品记录
		if err := repository.NewProductPackageTakeoutRepo(tx).CreateProductPackageTakeout(productPackageTakeout); err != nil {
			return errors.WithMessage(err, "创建外卖商品失败")
		}

		// 处理外卖规格价格
		if len(addReq.Flavors) > 0 {
			productBomTakeoutRepo := repository.NewProductBomTakeoutRepo(tx)

			for _, flavorReq := range addReq.Flavors {
				// 创建外卖规格价格记录
				productBomTakeout := &model.ProductBomTakeout{
					ProductPackageTakeoutUuid: productPackageTakeout.Uuid,
					ProductBomUuid:            flavorReq.BomUuid,
					HeadquarterUuid:           productPackage.HeadquarterUuid,
					Price:                     flavorReq.Price,
					GrabModifierId:            flavorReq.GrabModifierId,
				}
				if err := productBomTakeoutRepo.CreateProductBomTakeout(productBomTakeout); err != nil {
					return errors.WithMessage(err, "创建外卖规格价格失败")
				}
			}
		}

		// 处理外卖属性价格
		if len(addReq.Attributes) > 0 {
			productPackageAttributeTakeoutRepo := repository.NewProductPackageAttributeTakeoutRepo(tx)

			for _, attributeReq := range addReq.Attributes {
				// 创建外卖属性价格记录
				productPackageAttributeTakeout := &model.ProductPackageAttributeTakeout{
					ProductPackageTakeoutUuid:   productPackageTakeout.Uuid,
					ProductPackageAttributeUuid: attributeReq.ProductPackageAttributeUuid,
					HeadquarterUuid:             productPackage.HeadquarterUuid,
					Price:                       attributeReq.Price,
				}
				if err := productPackageAttributeTakeoutRepo.CreateProductPackageAttributeTakeout(productPackageAttributeTakeout); err != nil {
					return errors.WithMessage(err, "创建外卖属性价格失败")
				}
			}
		}

		// 处理外卖套餐子商品价格（仅当商品类型为套餐时）
		if productPackage.ProductType == constant.ProductTypePackage && len(addReq.PackageGroupItems) > 0 {
			packageGroupItemTakeoutRepo := repository.NewProductPackageGroupItemTakeoutRepo(tx)
			productPackageGroupRepo := repository.NewProductPackageGroupRepo(tx)

			for _, groupItemReq := range addReq.PackageGroupItems {
				// 验证套餐子商品是否存在
				groupItem, err := productPackageGroupRepo.GetProductPackageGroupItem(
					repository.CommonRepo.WhereByUuid(groupItemReq.ProductPackageGroupItemUuid),
					repository.CommonRepo.WhereBySoftDelete(),
				)
				if err != nil {
					return errors.WithMessage(err, "套餐子商品不存在")
				}
				if groupItem.IsDelete() {
					return errors.WithMessage(errors.New("套餐子商品已删除"))
				}

				// 创建外卖套餐子商品价格记录
				packageGroupItemTakeout := &model.ProductPackageGroupItemTakeout{
					ProductPackageTakeoutUuid:   productPackageTakeout.Uuid,
					ProductPackageGroupItemUuid: groupItemReq.ProductPackageGroupItemUuid,
					ProductPackageGroupUuid:     groupItem.ProductPackageGroupUuid,
					HeadquarterUuid:             productPackage.HeadquarterUuid,
					AddPrice:                    groupItemReq.AddPrice,
				}
				if err := packageGroupItemTakeoutRepo.CreateProductPackageGroupItemTakeout(packageGroupItemTakeout); err != nil {
					return errors.WithMessage(err, "创建外卖套餐子商品价格失败")
				}
			}
		}

		return nil
	})

	if err != nil {
		logger.Logger.Error("添加外卖商品失败", zap.Any("func", "AddProductTakeoutShop"), zap.Any("params", addReq), zap.Error(err))
		return nil, errors.WithMessage(err, "添加外卖商品失败")
	}

	// 自动设置分类在外卖平台显示
	if addReq.CategoryUuid != 0 {
		_ = s.productSrv.SetCategoryDisplayInTakeout(ctx, addReq.CategoryUuid)
	}
	if addReq.SpecialCategoryUuid != 0 {
		_ = s.productSrv.SetCategoryDisplayInTakeout(ctx, addReq.SpecialCategoryUuid)
	}

	return productPackageTakeout, nil
}

// EditProductTakeoutShop 编辑外卖商品
func (s *productTakeoutSrv) EditProductTakeoutShop(ctx context.Context, editReq req.ProductTakeoutShopEditReq) error {
	companySetting := ctx.GetCompanySetting()
	db := s.dbm.GetDB(ctx.GetDbId())

	// 检查外卖商品是否存在
	takeoutRepo := repository.NewProductPackageTakeoutRepo(db)
	existTakeout, err := takeoutRepo.GetProductPackageTakeout(
		repository.CommonRepo.WhereByUuid(editReq.Uuid),
		repository.CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return errors.WithMessage(errors.New("外卖商品不存在"))
	}
	if existTakeout.DeleteTime != 0 {
		return errors.WithMessage(errors.New("外卖商品已删除"))
	}

	// 判断是否是总部外卖商品
	isHeadquarterProduct := existTakeout.HeadquarterUuid != 0 && !companySetting.IsHeadquarter()

	// 准备更新数据
	updateData := map[string]any{
		"status": editReq.Status,
	}

	// 如果不是总部商品，可以编辑完整信息
	if !isHeadquarterProduct {
		updateData["category_uuid"] = editReq.CategoryUuid
		updateData["special_category_uuid"] = editReq.SpecialCategoryUuid
		updateData["image_file_uuid"] = editReq.ImageFileUuid

		// 处理多语言名称更新
		if !editReq.LocaleName.IsNull() {
			multiLanguageName := model.NewMultiLanguageName(editReq.LocaleName.ToJson())
			err = repository.NewMultiLanguageNameRepo(db).UpdateMultiLanguageName(existTakeout.MultiLanguageNameUuid, *multiLanguageName)
			if err != nil {
				return errors.WithMessage(err, "更新多语言名称失败")
			}
			updateData["name"] = editReq.LocaleName.ToJson()
		}
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		// 更新外卖商品信息
		if err := repository.NewProductPackageTakeoutRepo(tx).UpdateProductPackageTakeout(
			updateData,
			repository.CommonRepo.WhereByUuid(editReq.Uuid),
		); err != nil {
			return errors.WithMessage(err, "更新外卖商品失败")
		}

		// 处理外卖规格价格更新
		if len(editReq.Flavors) > 0 {
			productBomTakeoutRepo := repository.NewProductBomTakeoutRepo(tx)
			commonRepo := repository.NewCommonRepo()

			// 获取当前外卖商品的所有规格价格
			existingBomTakeouts, err := productBomTakeoutRepo.GetProductBomTakeoutList(
				commonRepo.WhereByProductPackageTakeoutUuid(editReq.Uuid),
				commonRepo.WhereBySoftDelete(),
			)
			if err != nil {
				return errors.WithMessage(err, "获取外卖规格价格失败")
			}

			// 构建现有规格的映射（key: product_bom_uuid）
			existingBomMap := make(map[uint64]*model.ProductBomTakeout)
			for _, bomTakeout := range existingBomTakeouts {
				existingBomMap[bomTakeout.ProductBomUuid] = bomTakeout
			}

			// 处理请求中的规格
			requestedBomUuids := make(map[uint64]bool)
			for _, flavorReq := range editReq.Flavors {
				requestedBomUuids[flavorReq.BomUuid] = true

				// 检查是否已存在
				if existingBom, exists := existingBomMap[flavorReq.BomUuid]; exists {
					// 更新价格
					if err := productBomTakeoutRepo.UpdateProductBomTakeout(
						map[string]any{"price": flavorReq.Price},
						commonRepo.WhereByUuid(existingBom.Uuid),
					); err != nil {
						return errors.WithMessage(err, "更新外卖规格价格失败")
					}
				} else {
					// 创建新的外卖规格价格
					productBomTakeout := &model.ProductBomTakeout{
						ProductPackageTakeoutUuid: editReq.Uuid,
						ProductBomUuid:            flavorReq.BomUuid,
						HeadquarterUuid:           existTakeout.HeadquarterUuid,
						Price:                     flavorReq.Price,
					}
					if err := productBomTakeoutRepo.CreateProductBomTakeout(productBomTakeout); err != nil {
						return errors.WithMessage(err, "创建外卖规格价格失败")
					}
				}
			}

			// 删除不再需要的外卖规格价格（软删除）
			for bomUuid, existingBom := range existingBomMap {
				if !requestedBomUuids[bomUuid] {
					if err := productBomTakeoutRepo.DestroyProductBomTakeout(
						commonRepo.WhereByUuid(existingBom.Uuid),
					); err != nil {
						return errors.WithMessage(err, "删除外卖规格价格失败")
					}
				}
			}
		}

		// 处理外卖属性价格更新
		if len(editReq.Attributes) > 0 {
			productPackageAttributeTakeoutRepo := repository.NewProductPackageAttributeTakeoutRepo(tx)
			commonRepo := repository.NewCommonRepo()

			// 获取当前外卖商品的所有属性价格
			existingAttributeTakeouts, err := productPackageAttributeTakeoutRepo.GetProductPackageAttributeTakeoutList(
				func(db *gorm.DB) *gorm.DB {
					return db.Where("product_package_takeout_uuid = ?", editReq.Uuid)
				},
				commonRepo.WhereBySoftDelete(),
			)
			if err != nil {
				return errors.WithMessage(err, "获取外卖属性价格失败")
			}

			// 构建现有属性的映射（key: product_package_attribute_uuid）
			existingAttributeMap := make(map[uint64]*model.ProductPackageAttributeTakeout)
			for _, attributeTakeout := range existingAttributeTakeouts {
				existingAttributeMap[attributeTakeout.ProductPackageAttributeUuid] = attributeTakeout
			}

			// 处理请求中的属性
			requestedAttributeUuids := make(map[uint64]bool)
			for _, attributeReq := range editReq.Attributes {
				requestedAttributeUuids[attributeReq.ProductPackageAttributeUuid] = true

				// 检查是否已存在
				if existingAttribute, exists := existingAttributeMap[attributeReq.ProductPackageAttributeUuid]; exists {
					// 更新价格
					if err := productPackageAttributeTakeoutRepo.UpdateProductPackageAttributeTakeout(
						map[string]any{"price": attributeReq.Price},
						commonRepo.WhereByUuid(existingAttribute.Uuid),
					); err != nil {
						return errors.WithMessage(err, "更新外卖属性价格失败")
					}
				} else {
					// 创建新的外卖属性价格
					productPackageAttributeTakeout := &model.ProductPackageAttributeTakeout{
						ProductPackageTakeoutUuid:   editReq.Uuid,
						ProductPackageAttributeUuid: attributeReq.ProductPackageAttributeUuid,
						HeadquarterUuid:             existTakeout.HeadquarterUuid,
						Price:                       attributeReq.Price,
					}
					if err := productPackageAttributeTakeoutRepo.CreateProductPackageAttributeTakeout(productPackageAttributeTakeout); err != nil {
						return errors.WithMessage(err, "创建外卖属性价格失败")
					}
				}
			}

			// 删除不再需要的外卖属性价格（软删除）
			for attributeUuid, existingAttribute := range existingAttributeMap {
				if !requestedAttributeUuids[attributeUuid] {
					if err := productPackageAttributeTakeoutRepo.DestroyProductPackageAttributeTakeout(
						commonRepo.WhereByUuid(existingAttribute.Uuid),
					); err != nil {
						return errors.WithMessage(err, "删除外卖属性价格失败")
					}
				}
			}
		}

		// 处理外卖套餐子商品价格更新（仅当商品类型为套餐时）
		if existTakeout.ProductType == constant.ProductTypePackage && len(editReq.PackageGroupItems) > 0 {
			packageGroupItemTakeoutRepo := repository.NewProductPackageGroupItemTakeoutRepo(tx)
			productPackageGroupRepo := repository.NewProductPackageGroupRepo(tx)
			commonRepo := repository.NewCommonRepo()

			// 获取当前外卖商品的所有套餐子商品价格
			existingGroupItemTakeouts, err := packageGroupItemTakeoutRepo.GetProductPackageGroupItemTakeoutList(editReq.Uuid)
			if err != nil {
				return errors.WithMessage(err, "获取外卖套餐子商品价格失败")
			}

			// 构建现有套餐子商品的映射（key: product_package_group_item_uuid）
			existingGroupItemMap := make(map[uint64]*model.ProductPackageGroupItemTakeout)
			for _, groupItemTakeout := range existingGroupItemTakeouts {
				existingGroupItemMap[groupItemTakeout.ProductPackageGroupItemUuid] = groupItemTakeout
			}

			// 处理请求中的套餐子商品
			requestedGroupItemUuids := make(map[uint64]bool)
			for _, groupItemReq := range editReq.PackageGroupItems {
				requestedGroupItemUuids[groupItemReq.ProductPackageGroupItemUuid] = true

				// 检查是否已存在
				if existingGroupItem, exists := existingGroupItemMap[groupItemReq.ProductPackageGroupItemUuid]; exists {
					// 更新加价
					if err := packageGroupItemTakeoutRepo.UpdateAddPrice(existingGroupItem.Uuid, groupItemReq.AddPrice); err != nil {
						return errors.WithMessage(err, "更新外卖套餐子商品价格失败")
					}
				} else {
					// 验证套餐子商品是否存在
					groupItem, err := productPackageGroupRepo.GetProductPackageGroupItem(
						commonRepo.WhereByUuid(groupItemReq.ProductPackageGroupItemUuid),
						commonRepo.WhereBySoftDelete(),
					)
					if err != nil {
						return errors.WithMessage(err, "套餐子商品不存在")
					}
					if groupItem.IsDelete() {
						return errors.WithMessage(errors.New("套餐子商品已删除"))
					}

					// 创建新的外卖套餐子商品价格
					packageGroupItemTakeout := &model.ProductPackageGroupItemTakeout{
						ProductPackageTakeoutUuid:   editReq.Uuid,
						ProductPackageGroupItemUuid: groupItemReq.ProductPackageGroupItemUuid,
						ProductPackageGroupUuid:     groupItem.ProductPackageGroupUuid,
						HeadquarterUuid:             existTakeout.HeadquarterUuid,
						AddPrice:                    groupItemReq.AddPrice,
					}
					if err := packageGroupItemTakeoutRepo.CreateProductPackageGroupItemTakeout(packageGroupItemTakeout); err != nil {
						return errors.WithMessage(err, "创建外卖套餐子商品价格失败")
					}
				}
			}

			// 删除不再需要的外卖套餐子商品价格（软删除）
			for groupItemUuid, existingGroupItem := range existingGroupItemMap {
				if !requestedGroupItemUuids[groupItemUuid] {
					if err := packageGroupItemTakeoutRepo.SoftDelete(existingGroupItem.Uuid); err != nil {
						return errors.WithMessage(err, "删除外卖套餐子商品价格失败")
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		logger.Logger.Error("编辑外卖商品失败", zap.Any("func", "EditProductTakeoutShop"), zap.Any("params", editReq), zap.Error(err))
		return errors.WithMessage(err, "编辑外卖商品失败")
	}

	// 自动设置分类在外卖平台显示（仅在非总部商品时才处理分类更新）
	if !isHeadquarterProduct {
		// 如果修改了分类UUID，自动设置新分类的外卖显示
		if editReq.CategoryUuid != 0 {
			_ = s.productSrv.SetCategoryDisplayInTakeout(ctx, editReq.CategoryUuid)
		}
		// 如果修改了特色分类UUID，自动设置新特色分类的外卖显示
		if editReq.SpecialCategoryUuid != 0 && editReq.SpecialCategoryUuid != editReq.CategoryUuid {
			_ = s.productSrv.SetCategoryDisplayInTakeout(ctx, editReq.SpecialCategoryUuid)
		}
	}

	return nil
}

// GetProductTakeoutShopDetail 获取外卖商品详情
func (s *productTakeoutSrv) GetProductTakeoutShopDetail(ctx context.Context, detailReq req.ProductTakeoutShopDetailReq) (*product_resp.ProductTakeoutShopDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	// 获取原商品的完整信息，包括套餐、加料、属性等
	productRepo := repository.NewProductRepo(db)
	productPackage, err := productRepo.GetProductDetail(detailReq.Uuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取原商品信息失败")
	}

	// 获取外卖商品信息
	takeoutRepo := repository.NewProductPackageTakeoutRepo(db)
	takeout, err := takeoutRepo.GetProductPackageTakeout(
		func(db *gorm.DB) *gorm.DB {
			return db.Where("product_package_uuid = ?", detailReq.Uuid)
		},
		repository.CommonRepo.WhereBySoftDelete(),
		takeoutRepo.WithProductPackage(),
		takeoutRepo.WithProductPackageMultiLanguageName(),
		takeoutRepo.WithProductCategory(),
		takeoutRepo.WithProductSpecialCategory(),
		takeoutRepo.WithImageFile(),
		takeoutRepo.WhereByTakeoutType(uint(s.platformToTakeoutType(detailReq.Platform))),
	)
	if err != nil {
		return &product_resp.ProductTakeoutShopDetailResp{}, nil
	}

	// 获取外卖商品的规格价格列表
	productBomTakeoutRepo := repository.NewProductBomTakeoutRepo(db)
	commonRepo := repository.NewCommonRepo()
	takeoutBoms, err := productBomTakeoutRepo.GetProductBomTakeoutList(
		commonRepo.WhereByProductPackageTakeoutUuid(takeout.Uuid),
		commonRepo.WhereBySoftDelete(),
		productBomTakeoutRepo.WithProductBom(),
		productBomTakeoutRepo.WithProductBomProductFlavor(),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "获取外卖规格价格失败")
	}

	// 构建外卖规格价格映射表
	takeoutPriceMap := make(map[uint64]float64)
	for _, bomTakeout := range takeoutBoms {
		if !bomTakeout.IsDelete() {
			takeoutPriceMap[bomTakeout.ProductBomUuid] = bomTakeout.Price
		}
	}

	// 构建规格响应（包含所有规格，标记是否有外卖价格）
	flavors := make([]product_resp.ProductTakeoutShopFlavorResp, 0)
	for _, productBom := range productPackage.ProductBoms {
		if !productBom.IsDelete() && productBom.IsFlavor() {
			takeoutPrice, hasTakeoutPrice := takeoutPriceMap[productBom.Uuid]
			if !hasTakeoutPrice {
				// 如果没有设置外卖价格，使用原商品价格
				takeoutPrice = productBom.Price
			}

			flavors = append(flavors, product_resp.ProductTakeoutShopFlavorResp{
				BomUuid:    productBom.Uuid,
				LocaleName: productBom.ProductFlavor.MultiLanguageName.GetNames(),
				Price:      takeoutPrice,
			})
		}
	}

	// 获取图片URL
	imageUrl := ""
	if takeout.ImageFileUuid != 0 {
		imageUrl = takeout.ImageFile.GetUrl(utils.GetBaseURL(ctx.GetGin().Request))
	}

	result := &product_resp.ProductTakeoutShopDetailResp{
		Uuid:                takeout.Uuid,
		ProductPackageUuid:  takeout.ProductPackageUuid,
		ProductType:         uint(productPackage.ProductType),
		TakeoutType:         int(takeout.TakeoutType),
		LocaleName:          takeout.ProductPackage.MultiLanguageName.GetNames(),
		CategoryUuid:        takeout.CategoryUuid,
		CategoryName:        takeout.ProductCategory.MultiLanguageName.GetNames(),
		SpecialCategoryUuid: takeout.SpecialCategoryUuid,
		SpecialCategoryName: takeout.ProductSpecialCategory.MultiLanguageName.GetNames(),
		Status:              int(takeout.Status),
		ImageFileUuid:       takeout.ImageFileUuid,
		ImageUrl:            imageUrl,
		HeadquarterUuid:     takeout.HeadquarterUuid,
		Flavors:             flavors,
		// 关联原商品的套餐信息
		PackageSubProductGroups: product_resp.ProductPackageSubProductGroupList{
			List: productPackage.GetRespPackageSubProductGroupList(),
		},
	}

	return result, nil
}

// DeleteProductTakeoutShop 删除外卖商品
func (s *productTakeoutSrv) DeleteProductTakeoutShop(ctx context.Context, deleteReq req.ProductTakeoutShopDeleteReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())

	takeoutRepo := repository.NewProductPackageTakeoutRepo(db)
	takeout, err := takeoutRepo.GetProductPackageTakeout(
		takeoutRepo.WhereByProductPackageUuid(deleteReq.Uuid),
		takeoutRepo.WhereByTakeoutType(uint(s.platformToTakeoutType(deleteReq.Platform))),
	)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return errors.WithMessage(errors.New("外卖商品不存在"))
	}

	// 检查是否是总部外卖商品，总部商品不能删除
	if takeout.HeadquarterUuid != 0 {
		return errors.WithMessage(errors.New("总部外卖商品不能删除"))
	}

	// TODO: 检查是否有未完成的外卖订单
	// 注意：目前外卖订单功能未开发，暂不检查
	// 后续需要检查 ttpos_sale_order 表中是否有该外卖商品的未完成订单
	// 如果有，应该阻止删除或给出提示

	err = db.Transaction(func(tx *gorm.DB) error {
		// 删除外卖商品
		if err := repository.NewProductPackageTakeoutRepo(tx).DestroyProductPackageTakeout(
			repository.CommonRepo.WhereByUuid(takeout.Uuid),
		); err != nil {
			return errors.WithMessage(err, "删除外卖商品失败")
		}

		// 删除关联的外卖规格价格
		productBomTakeoutRepo := repository.NewProductBomTakeoutRepo(tx)
		if err := productBomTakeoutRepo.DestroyProductBomTakeout(
			repository.CommonRepo.WhereByProductPackageTakeoutUuid(takeout.Uuid),
		); err != nil {
			return errors.WithMessage(err, "删除外卖规格价格失败")
		}

		return nil
	})

	if err != nil {
		logger.Logger.Error("删除外卖商品失败", zap.Any("func", "DeleteProductTakeoutShop"), zap.Any("params", deleteReq), zap.Error(err))
		return err
	}

	return nil
}

// UpdateProductTakeoutShopStatus 更新外卖商品状态
func (s *productTakeoutSrv) UpdateProductTakeoutShopStatus(ctx context.Context, statusReq req.ProductTakeoutShopStatusReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())

	takeoutRepo := repository.NewProductPackageTakeoutRepo(db)
	takeout, err := takeoutRepo.GetProductPackageTakeout(
		takeoutRepo.WhereByProductPackageUuid(statusReq.Uuid),
		takeoutRepo.WhereByTakeoutType(uint(s.platformToTakeoutType(statusReq.Platform))),
	)
	if err != nil {
		return errors.WithMessage(errors.New("外卖商品不存在"))
	}

	if err := takeoutRepo.UpdateProductPackageTakeout(
		map[string]any{"status": *statusReq.Status},
		repository.CommonRepo.WhereByUuid(takeout.Uuid),
	); err != nil {
		logger.Logger.Error("更新外卖商品状态失败", zap.Any("func", "UpdateProductTakeoutShopStatus"), zap.Any("params", statusReq), zap.Error(err))
		return errors.WithMessage(err, "更新外卖商品状态失败")
	}

	return nil
}

// restoreProductTakeoutShop 还原已软删除的外卖商品
func (s *productTakeoutSrv) restoreProductTakeoutShop(ctx context.Context, existingTakeout *model.ProductPackageTakeout, addReq req.ProductTakeoutShopAddReq) (*model.ProductPackageTakeout, error) {
	db := ctx.GetDB()

	// 准备更新数据
	updateData := map[string]any{
		"delete_time":           0, // 还原软删除
		"status":                addReq.Status,
		"category_uuid":         addReq.CategoryUuid,
		"special_category_uuid": addReq.SpecialCategoryUuid,
		"image_file_uuid":       addReq.ImageFileUuid,
		"source":                addReq.Source,
		"source_product_id":     addReq.SourceProductId,
	}

	// 处理多语言名称
	if !addReq.LocaleName.IsNull() {
		productName := addReq.LocaleName.ToJson()
		multiLanguageName := model.NewMultiLanguageName(productName)

		// 如果已有多语言记录，更新它；否则创建新的
		if existingTakeout.MultiLanguageNameUuid != 0 {
			err := repository.NewMultiLanguageNameRepo(db).UpdateMultiLanguageName(existingTakeout.MultiLanguageNameUuid, *multiLanguageName)
			if err != nil {
				return nil, errors.WithMessage(err, "更新多语言名称失败")
			}
		} else {
			multiLanguageNameUuid, err := repository.NewMultiLanguageNameRepo(db).CreateMultiLanguageName(*multiLanguageName)
			if err != nil {
				return nil, errors.WithMessage(err, "创建多语言名称失败")
			}
			updateData["multi_language_name_uuid"] = multiLanguageNameUuid
		}
		updateData["name"] = productName
	}

	// 处理卖点多语言
	if !addReq.Describe.IsNull() {
		describe := addReq.Describe.ToJson()
		describeMultiLanguageName := model.NewMultiLanguageName(describe)

		// 如果已有卖点多语言记录，更新它；否则创建新的
		if existingTakeout.DescribeMultiLanguageNameUuid != 0 {
			err := repository.NewMultiLanguageNameRepo(db).UpdateMultiLanguageName(existingTakeout.DescribeMultiLanguageNameUuid, *describeMultiLanguageName)
			if err != nil {
				return nil, errors.WithMessage(err, "更新卖点多语言失败")
			}
		} else {
			describeMultiLanguageNameUuid, err := repository.NewMultiLanguageNameRepo(db).CreateMultiLanguageName(*describeMultiLanguageName)
			if err != nil {
				return nil, errors.WithMessage(err, "创建卖点多语言失败")
			}
			updateData["describe_multi_language_name_uuid"] = describeMultiLanguageNameUuid
		}
		updateData["describe"] = describe
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		// 还原外卖商品记录
		takeoutRepo := repository.NewProductPackageTakeoutRepo(tx)
		if err := takeoutRepo.UpdateProductPackageTakeout(
			updateData,
			repository.CommonRepo.WhereByUuid(existingTakeout.Uuid),
		); err != nil {
			return errors.WithMessage(err, "还原外卖商品失败")
		}

		// 删除旧的外卖规格价格记录（软删除的可能已过期）
		productBomTakeoutRepo := repository.NewProductBomTakeoutRepo(tx)
		commonRepo := repository.NewCommonRepo()

		// 软删除所有旧的规格价格
		if err := productBomTakeoutRepo.DestroyProductBomTakeout(
			commonRepo.WhereByProductPackageTakeoutUuid(existingTakeout.Uuid),
		); err != nil {
			return errors.WithMessage(err, "清理旧规格价格失败")
		}

		// 添加新的外卖规格价格
		if len(addReq.Flavors) > 0 {
			for _, flavorReq := range addReq.Flavors {
				productBomTakeout := &model.ProductBomTakeout{
					ProductPackageTakeoutUuid: existingTakeout.Uuid,
					ProductBomUuid:            flavorReq.BomUuid,
					HeadquarterUuid:           existingTakeout.HeadquarterUuid,
					Price:                     flavorReq.Price,
					GrabModifierId:            flavorReq.GrabModifierId,
				}
				if err := productBomTakeoutRepo.CreateProductBomTakeout(productBomTakeout); err != nil {
					return errors.WithMessage(err, "创建外卖规格价格失败")
				}
			}
		}

		// 删除旧的外卖属性价格记录
		productPackageAttributeTakeoutRepo := repository.NewProductPackageAttributeTakeoutRepo(tx)

		// 软删除所有旧的属性价格
		if err := productPackageAttributeTakeoutRepo.DestroyProductPackageAttributeTakeout(
			commonRepo.WhereByProductPackageTakeoutUuid(existingTakeout.Uuid),
		); err != nil {
			return errors.WithMessage(err, "清理旧属性价格失败")
		}

		// 添加新的外卖属性价格
		if len(addReq.Attributes) > 0 {
			for _, attributeReq := range addReq.Attributes {
				productPackageAttributeTakeout := &model.ProductPackageAttributeTakeout{
					ProductPackageTakeoutUuid:   existingTakeout.Uuid,
					ProductPackageAttributeUuid: attributeReq.ProductPackageAttributeUuid,
					HeadquarterUuid:             existingTakeout.HeadquarterUuid,
					Price:                       attributeReq.Price,
				}
				if err := productPackageAttributeTakeoutRepo.CreateProductPackageAttributeTakeout(productPackageAttributeTakeout); err != nil {
					return errors.WithMessage(err, "创建外卖属性价格失败")
				}
			}
		}

		return nil
	})

	if err != nil {
		logger.Logger.Error("还原外卖商品失败", zap.Any("func", "restoreProductTakeoutShop"), zap.Any("params", addReq), zap.Error(err))
		return nil, errors.WithMessage(err, "还原外卖商品失败")
	}

	// 自动设置分类在外卖平台显示
	if addReq.CategoryUuid != 0 {
		_ = s.productSrv.SetCategoryDisplayInTakeout(ctx, addReq.CategoryUuid)
	}
	if addReq.SpecialCategoryUuid != 0 {
		_ = s.productSrv.SetCategoryDisplayInTakeout(ctx, addReq.SpecialCategoryUuid)
	}

	// 重新加载并返回还原后的记录
	restoredTakeout, err := repository.NewProductPackageTakeoutRepo(db).GetProductPackageTakeout(
		repository.CommonRepo.WhereByUuid(existingTakeout.Uuid),
		repository.CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "获取还原后的外卖商品失败")
	}

	return restoredTakeout, nil
}

// GetProductCount 获取外卖商品统计
func (s *productTakeoutSrv) GetProductCount(
	ctx context.Context,
	companyUuid uint64,
	platform string,
	forceRefresh bool,
) (int64, error) {
	// 1. 构造缓存 Key
	cacheKey := s.buildCountCacheKey(companyUuid, platform)

	// 2. 检查缓存(如果不是强制刷新)
	if !forceRefresh {
		cached, ok := s.cache.Get(cacheKey)
		if ok {
			if count, ok := cached.(int64); ok {
				logger.Logger.Debug("缓存命中", zap.String("key", cacheKey), zap.Int64("count", count))
				return count, nil
			}
		}
	}

	// 3. 查询数据库
	db := s.dbm.GetDB(ctx.GetDbId())

	// 4. 构造查询条件
	var count int64
	query := db.Model(&model.ProductPackageTakeout{}).
		Joins("LEFT JOIN ttpos_product_package ON ttpos_product_package_takeout.product_package_uuid = ttpos_product_package.uuid").
		Where("ttpos_product_package_takeout.delete_time = ?", 0).
		Where("ttpos_product_package.delete_time = ?", 0)

	// 6. 如果指定了平台,添加平台过滤 (使用source字段匹配)
	if platform != "" {
		query = query.Where("ttpos_product_package_takeout.source = ?", platform)
	}

	// 7. 执行统计
	if err := query.Count(&count).Error; err != nil {
		logger.Logger.Error("查询商品统计失败",
			zap.Uint64("company_uuid", companyUuid),
			zap.String("platform", platform),
			zap.Error(err))
		return 0, errors.WithMessage(err, "查询商品统计失败")
	}

	// 8. 写入缓存(5分钟)
	if err := s.cache.Set(cacheKey, count, 5*60); err != nil {
		logger.Logger.Warn("写入缓存失败", zap.String("key", cacheKey), zap.Error(err))
		// 缓存失败不影响结果返回
	}

	return count, nil
}

// buildCountCacheKey 构造缓存Key
func (s *productTakeoutSrv) buildCountCacheKey(companyUuid uint64, platform string) string {
	if platform == "" {
		return fmt.Sprintf("takeout:products:count:%d:all", companyUuid)
	}
	return fmt.Sprintf("takeout:products:count:%d:%s", companyUuid, platform)
}

// ClearProductCountCache 清除商品统计缓存(商品导入/删除时调用)
func (s *productTakeoutSrv) ClearProductCountCache(ctx context.Context, companyUuid uint64, platform string) {
	// 清除指定平台缓存
	if platform != "" {
		key := s.buildCountCacheKey(companyUuid, platform)
		s.cache.Del(key)
	}

	// 清除所有平台缓存
	allKey := s.buildCountCacheKey(companyUuid, "")
	s.cache.Del(allKey)
}

// BatchCreateProducts 批量创建外卖商品映射
func (s *productTakeoutSrv) BatchCreateProducts(ctx context.Context, batchReq req.TakeoutBatchCreateReq) (*product_resp.TakeoutBatchResp, error) {
	// 验证请求参数
	if err := batchReq.Validate(); err != nil {
		return nil, err
	}

	// 转换平台标识为外卖类型
	takeoutType := s.platformToTakeoutType(batchReq.Platform)

	// 创建响应对象
	result := &product_resp.TakeoutBatchResp{
		Total:          len(batchReq.ProductUuids),
		Success:        0,
		Failed:         0,
		FailedProducts: make([]product_resp.TakeoutBatchFailedProduct, 0),
	}

	// 限流器: 每秒10个请求
	limiter := time.NewTicker(100 * time.Millisecond)
	defer limiter.Stop()

	// 使用 WaitGroup 等待所有 Goroutine 完成
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 获取商品名称映射(用于错误信息)
	db := ctx.GetDB()
	productNames := s.getProductNames(db, batchReq.ProductUuids)

	// 并发处理每个商品
	for _, productUuid := range batchReq.ProductUuids {
		wg.Add(1)
		go func(uuid uint64) {
			defer wg.Done()
			<-limiter.C // 限流

			// 调用 AddProductTakeoutShop 逻辑
			addReq := req.ProductTakeoutShopAddReq{
				ProductPackageUuid: uuid,
				TakeoutType:        takeoutType,
			}
			_, err := s.AddProductTakeoutShop(ctx, addReq)
			mu.Lock()
			if err != nil {
				result.Failed++
				result.FailedProducts = append(result.FailedProducts, product_resp.TakeoutBatchFailedProduct{
					ProductUuid: uuid,
					ProductName: productNames[uuid],
					Error:       err.Error(),
				})
			} else {
				result.Success++
			}
			mu.Unlock()
		}(productUuid)
	}

	wg.Wait()
	return result, nil
}

// platformToTakeoutType 平台标识转外卖类型
func (s *productTakeoutSrv) platformToTakeoutType(platform string) int {
	switch platform {
	case "grab":
		return constant.TakeoutTypeGrab
	case "lineman":
		return constant.TakeoutTypeLINEMAN // 使用 LINE MAN 类型代表 LINE MAN
	default:
		return constant.TakeoutTypeGrab
	}
}

// getProductNames 批量获取商品名称
func (s *productTakeoutSrv) getProductNames(db *gorm.DB, productUuids []uint64) map[uint64]string {
	names := make(map[uint64]string)
	if len(productUuids) == 0 {
		return names
	}

	var products []model.ProductPackage
	err := db.Model(&model.ProductPackage{}).
		Where("uuid IN ?", productUuids).
		Where("delete_time = ?", 0).
		Select("uuid, name").
		Find(&products).Error

	if err != nil {
		logger.Logger.Warn("获取商品名称失败", zap.Error(err))
		return names
	}

	for _, product := range products {
		names[product.Uuid] = product.Name
	}

	return names
}

// BatchOnlineProducts 批量上架外卖商品
func (s *productTakeoutSrv) BatchOnlineProducts(ctx context.Context, batchReq req.TakeoutBatchOnlineReq) (*product_resp.TakeoutBatchResp, error) {
	// 验证请求参数
	if err := batchReq.Validate(); err != nil {
		return nil, err
	}

	// 创建响应对象
	result := &product_resp.TakeoutBatchResp{
		Total:          len(batchReq.ProductUuids),
		Success:        0,
		Failed:         0,
		FailedProducts: make([]product_resp.TakeoutBatchFailedProduct, 0),
	}

	// 限流器: 每秒10个请求
	limiter := time.NewTicker(100 * time.Millisecond)
	defer limiter.Stop()

	// 使用 WaitGroup 等待所有 Goroutine 完成
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 获取商品名称映射
	db := ctx.GetDB()
	productNames := s.getProductNames(db, batchReq.ProductUuids)

	// 并发处理每个商品
	for _, productUuid := range batchReq.ProductUuids {
		wg.Add(1)
		go func(uuid uint64) {
			defer wg.Done()
			<-limiter.C // 限流

			// 调用上架逻辑
			status := 1 // 1=上架
			statusReq := req.ProductTakeoutShopStatusReq{
				Uuid:     uuid,
				Platform: batchReq.Platform,
				Status:   &status,
			}
			err := s.UpdateProductTakeoutShopStatus(ctx, statusReq)
			if err != nil {
				// 重试3次
				err = s.retryUpdateStatus(ctx, uuid, 1, 3)
			}

			mu.Lock()
			if err != nil {
				result.Failed++
				result.FailedProducts = append(result.FailedProducts, product_resp.TakeoutBatchFailedProduct{
					ProductUuid: uuid,
					ProductName: productNames[uuid],
					Error:       err.Error(),
				})
			} else {
				result.Success++
			}
			mu.Unlock()
		}(productUuid)
	}

	wg.Wait()
	return result, nil
}

// BatchOfflineProducts 批量下架外卖商品
func (s *productTakeoutSrv) BatchOfflineProducts(ctx context.Context, batchReq req.TakeoutBatchOfflineReq) (*product_resp.TakeoutBatchResp, error) {
	// 验证请求参数
	if err := batchReq.Validate(); err != nil {
		return nil, err
	}

	// 创建响应对象
	result := &product_resp.TakeoutBatchResp{
		Total:          len(batchReq.ProductUuids),
		Success:        0,
		Failed:         0,
		FailedProducts: make([]product_resp.TakeoutBatchFailedProduct, 0),
	}

	// 限流器: 每秒10个请求
	limiter := time.NewTicker(100 * time.Millisecond)
	defer limiter.Stop()

	// 使用 WaitGroup 等待所有 Goroutine 完成
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 获取商品名称映射
	db := ctx.GetDB()
	productNames := s.getProductNames(db, batchReq.ProductUuids)

	// 并发处理每个商品
	for _, productUuid := range batchReq.ProductUuids {
		wg.Add(1)
		go func(uuid uint64) {
			defer wg.Done()
			<-limiter.C // 限流

			// 调用下架逻辑
			status := 0 // 0=下架
			statusReq := req.ProductTakeoutShopStatusReq{
				Uuid:     uuid,
				Platform: batchReq.Platform,
				Status:   &status,
			}
			err := s.UpdateProductTakeoutShopStatus(ctx, statusReq)
			if err != nil {
				// 重试3次
				err = s.retryUpdateStatus(ctx, uuid, 0, 3)
			}

			mu.Lock()
			if err != nil {
				result.Failed++
				result.FailedProducts = append(result.FailedProducts, product_resp.TakeoutBatchFailedProduct{
					ProductUuid: uuid,
					ProductName: productNames[uuid],
					Error:       err.Error(),
				})
			} else {
				result.Success++
			}
			mu.Unlock()
		}(productUuid)
	}

	wg.Wait()
	return result, nil
}

// BatchDeleteProducts 批量删除外卖商品
func (s *productTakeoutSrv) BatchDeleteProducts(ctx context.Context, batchReq req.TakeoutBatchDeleteReq) (*product_resp.TakeoutBatchResp, error) {
	// 验证请求参数
	if err := batchReq.Validate(); err != nil {
		return nil, err
	}

	// 创建响应对象
	result := &product_resp.TakeoutBatchResp{
		Total:          len(batchReq.ProductUuids),
		Success:        0,
		Failed:         0,
		FailedProducts: make([]product_resp.TakeoutBatchFailedProduct, 0),
	}

	// 限流器: 每秒10个请求
	limiter := time.NewTicker(100 * time.Millisecond)
	defer limiter.Stop()

	// 使用 WaitGroup 等待所有 Goroutine 完成
	var wg sync.WaitGroup
	var mu sync.Mutex

	// 获取商品名称映射
	db := ctx.GetDB()
	productNames := s.getProductNames(db, batchReq.ProductUuids)

	// 并发处理每个商品
	for _, productUuid := range batchReq.ProductUuids {
		wg.Add(1)
		go func(uuid uint64) {
			defer wg.Done()
			<-limiter.C // 限流

			// 调用删除逻辑
			deleteReq := req.ProductTakeoutShopDeleteReq{
				Uuid:     productUuid,
				Platform: batchReq.Platform,
			}
			err := s.DeleteProductTakeoutShop(ctx, deleteReq)

			mu.Lock()
			if err != nil {
				result.Failed++
				result.FailedProducts = append(result.FailedProducts, product_resp.TakeoutBatchFailedProduct{
					ProductUuid: uuid,
					ProductName: productNames[uuid],
					Error:       err.Error(),
				})
			} else {
				result.Success++
			}
			mu.Unlock()
		}(productUuid)
	}

	wg.Wait()
	return result, nil
}

// retryUpdateStatus 重试更新状态
func (s *productTakeoutSrv) retryUpdateStatus(ctx context.Context, productUuid uint64, status int, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		statusReq := req.ProductTakeoutShopStatusReq{
			Uuid:   productUuid,
			Status: &status,
		}
		err = s.UpdateProductTakeoutShopStatus(ctx, statusReq)
		if err == nil {
			return nil
		}

		// 指数退避
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	return err
}
