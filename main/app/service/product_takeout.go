package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	product_resp "ttpos-server-go/app/dto/resp/product_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IProductTakeoutSrv interface {
	// 外卖商品管理
	AddProductTakeoutShop(ctx context.Context, req req.ProductTakeoutShopAddReq) (uint64, error)
	EditProductTakeoutShop(ctx context.Context, req req.ProductTakeoutShopEditReq) error
	GetProductTakeoutShopDetail(ctx context.Context, req req.ProductTakeoutShopDetailReq) (*product_resp.ProductTakeoutShopDetailResp, error)
	DeleteProductTakeoutShop(ctx context.Context, req req.ProductTakeoutShopDeleteReq) error
	UpdateProductTakeoutShopStatus(ctx context.Context, req req.ProductTakeoutShopStatusReq) error
}

type productTakeoutSrv struct {
	dbm       *database.DBManager
	localeSrv ILocaleSrv
}

func NewProductTakeoutSrv(dbm *database.DBManager, localeSrv ILocaleSrv) IProductTakeoutSrv {
	return &productTakeoutSrv{
		dbm:       dbm,
		localeSrv: localeSrv,
	}
}

// AddProductTakeoutShop 添加外卖商品
func (s *productTakeoutSrv) AddProductTakeoutShop(ctx context.Context, addReq req.ProductTakeoutShopAddReq) (uint64, error) {
	companySetting := ctx.GetCompanySetting()
	db := s.dbm.GetDB(ctx.GetDbId())

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
		return 0, errors.WithMessage(errors.New("商品不存在"))
	}
	if productPackage.IsDelete() {
		return 0, errors.WithMessage(errors.New("商品已删除"))
	}

	// 检查是否是总部商品，并且当前不是总店，则总部商品不能添加为外卖商品
	if productPackage.HeadquarterUuid != 0 && !companySetting.IsHeadquarter() {
		return 0, errors.WithMessage(errors.New("总部商品不能添加为外卖商品"))
	}

	// 检查是否已存在同类型外卖商品
	takeoutRepo := repository.NewProductPackageTakeoutRepo(db)
	if takeoutRepo.CheckProductPackageTakeoutExist(addReq.ProductPackageUuid, uint(addReq.TakeoutType)) {
		return 0, errors.WithMessage(errors.New("该商品已存在相同类型的外卖配置"))
	}

	// 生成UUID
	uuid, err := utils.GetID()
	if err != nil {
		return 0, errors.WithMessage(err, "生成UUID失败")
	}

	// 处理多语言名称
	var multiLanguageNameUuid uint64
	var productName string

	if !addReq.LocaleName.IsNull() {
		productName = addReq.LocaleName.ToJson()
		multiLanguageName := model.NewMultiLanguageName(productName)
		multiLanguageNameUuid, err = repository.NewMultiLanguageNameRepo(db).CreateMultiLanguageName(*multiLanguageName)
		if err != nil {
			return 0, errors.WithMessage(err, "创建多语言名称失败")
		}
	} else {
		// 未提供自定义名称，使用店内商品的名称
		multiLanguageNameUuid = productPackage.MultiLanguageNameUuid
		productName = productPackage.Name
	}

	// 创建外卖商品
	productPackageTakeout := &model.ProductPackageTakeout{
		BaseModel: model.BaseModel{
			Uuid: uuid,
		},
		ProductPackageUuid:    addReq.ProductPackageUuid,
		MultiLanguageNameUuid: multiLanguageNameUuid,
		HeadquarterUuid:       productPackage.HeadquarterUuid,
		Name:                  productName,
		ProductType:           uint(productPackage.ProductType),
		TakeoutType:           uint(addReq.TakeoutType),
		Status:                uint(addReq.Status),
		CategoryUuid:          addReq.CategoryUuid,
		SpecialCategoryUuid:   addReq.SpecialCategoryUuid,
		ImageFileUuid:         addReq.ImageFileUuid,
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
				}
				if err := productBomTakeoutRepo.CreateProductBomTakeout(productBomTakeout); err != nil {
					return errors.WithMessage(err, "创建外卖规格价格失败")
				}
			}
		}

		return nil
	})

	if err != nil {
		logger.Logger.Error("添加外卖商品失败", zap.Any("func", "AddProductTakeoutShop"), zap.Any("params", addReq), zap.Error(err))
		return 0, errors.WithMessage(err, "添加外卖商品失败")
	}

	return uuid, nil
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

		return nil
	})

	if err != nil {
		logger.Logger.Error("编辑外卖商品失败", zap.Any("func", "EditProductTakeoutShop"), zap.Any("params", editReq), zap.Error(err))
		return errors.WithMessage(err, "编辑外卖商品失败")
	}

	return nil
}

// GetProductTakeoutShopDetail 获取外卖商品详情
func (s *productTakeoutSrv) GetProductTakeoutShopDetail(ctx context.Context, detailReq req.ProductTakeoutShopDetailReq) (*product_resp.ProductTakeoutShopDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())

	takeoutRepo := repository.NewProductPackageTakeoutRepo(db)
	takeout, err := takeoutRepo.GetProductPackageTakeout(
		repository.CommonRepo.WhereByUuid(detailReq.Uuid),
		repository.CommonRepo.WhereBySoftDelete(),
		takeoutRepo.WithProductPackage(),
		takeoutRepo.WithProductPackageMultiLanguageName(),
		takeoutRepo.WithProductCategory(),
		takeoutRepo.WithProductSpecialCategory(),
		takeoutRepo.WithImageFile(),
	)
	if err != nil {
		return nil, errors.WithMessage(errors.New("外卖商品不存在"))
	}

	// 获取原商品的完整信息，包括套餐、加料、属性等
	productRepo := repository.NewProductRepo(db)
	productPackage, err := productRepo.GetProductDetail(takeout.ProductPackageUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取原商品信息失败")
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
		repository.CommonRepo.WhereByUuid(deleteReq.Uuid),
		repository.CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return errors.WithMessage(errors.New("外卖商品不存在"))
	}

	// 检查是否是总部外卖商品，总部商品不能删除
	if takeout.HeadquarterUuid != 0 {
		return errors.WithMessage(errors.New("总部外卖商品不能删除"))
	}

	// 检查是否有未完成的外卖订单
	// 注意：目前外卖订单功能未开发，暂不检查
	// 后续需要检查 ttpos_sale_order 表中是否有该外卖商品的未完成订单
	// 如果有，应该阻止删除或给出提示

	err = db.Transaction(func(tx *gorm.DB) error {
		// 删除外卖商品
		if err := repository.NewProductPackageTakeoutRepo(tx).DestroyProductPackageTakeout(
			repository.CommonRepo.WhereByUuid(deleteReq.Uuid),
		); err != nil {
			return errors.WithMessage(err, "删除外卖商品失败")
		}

		// 删除关联的外卖规格价格
		productBomTakeoutRepo := repository.NewProductBomTakeoutRepo(tx)
		if err := productBomTakeoutRepo.DestroyProductBomTakeout(
			repository.CommonRepo.WhereByProductPackageTakeoutUuid(deleteReq.Uuid),
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
	_, err := takeoutRepo.GetProductPackageTakeout(
		repository.CommonRepo.WhereByUuid(statusReq.Uuid),
		repository.CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return errors.WithMessage(errors.New("外卖商品不存在"))
	}

	if err := takeoutRepo.UpdateProductPackageTakeout(
		map[string]any{"status": *statusReq.Status},
		repository.CommonRepo.WhereByUuid(statusReq.Uuid),
	); err != nil {
		logger.Logger.Error("更新外卖商品状态失败", zap.Any("func", "UpdateProductTakeoutShopStatus"), zap.Any("params", statusReq), zap.Error(err))
		return errors.WithMessage(err, "更新外卖商品状态失败")
	}

	return nil
}
