package service

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// ISyncProductToErpSrv同步服务接口
type ISyncProductToErpSrv interface {
	Sync(ctx context.Context) error // 同步商品数据到ERP
}

// SyncProductToErpSrv同步服务结构体
type SyncProductToErpSrv struct {
	dbm        *database.DBManager
	productSrv IProductSrv
	settingSrv setting.ISrv
}

// NewSyncProductToErpSrv 创建新同步服务
func NewSyncProductToErpSrv(dbm *database.DBManager, productSrv IProductSrv, settingSrv setting.ISrv) ISyncProductToErpSrv {
	return &SyncProductToErpSrv{
		dbm:        dbm,
		productSrv: productSrv,
		settingSrv: settingSrv,
	}
}

// NewSyncProductToErpSrvImpl 创建新同步服务实现
func NewSyncProductToErpSrvImpl(dbm *database.DBManager, productSrv IProductSrv, settingSrv setting.ISrv) ISyncProductToErpSrv {
	return &SyncProductToErpSrv{
		dbm:        dbm,
		productSrv: productSrv,
		settingSrv: settingSrv,
	}
}

// Sync 同步商品数据到ERP
func (s *SyncProductToErpSrv) Sync(ctx context.Context) error {
	tx := ctx.GetDB()

	commonRepo := repository.NewCommonRepo()

	// 商品加料
	sauceRepo := repository.NewProductSauceRepo(tx)
	sauceList := sauceRepo.GetProductSauceList(
		commonRepo.WhereBySoftDelete(),
		commonRepo.WhereByHeadquarterUuid(0),
		sauceRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
	)
	for _, sauce := range sauceList {
		multiLanguageName := model.NewMultiLanguageName(sauce.MultiLanguageName.ToJson())
		enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
		if err != nil {
			return errors.WithMessage(err, "翻译失败")
		}
		erpSrv := erp.NewIErpSrv(s.dbm)
		itemInfo, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
			ItemName: enName,
			StockUom: "Nos",
		})
		if errErp != nil {
			return errors.WithMessage(errErp, "同步新增加料到erp失败")
		}
		err = sauceRepo.UpdateProductSauce(map[string]any{"erp_code": itemInfo.ItemCode}, commonRepo.WhereByUuid(sauce.Uuid))
		if err != nil {
			return errors.WithMessage(err, "更新加料erp编码失败")
		}
	}

	// 同步规格
	err := s.productSrv.UpdateProductFlavorErp(ctx, tx)
	if err != nil {
		return errors.WithMessage(err, "更新规格到erp失败")
	}

	// 同步商品
	productPackageRepo := repository.NewProductPackageRepo(tx)
	productBomRepo := repository.NewProductBomRepo(tx)
	productPackageList, err := productPackageRepo.GetProductPackageList(
		commonRepo.WhereBySoftDelete(),
		commonRepo.WhereByHeadquarterUuid(0),
		productPackageRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
		productPackageRepo.WithProductBoms(commonRepo.WhereBySoftDelete()),
		productPackageRepo.WithProductBomsProductFlavor(commonRepo.WhereBySoftDelete()),
		productPackageRepo.WithProductCategory(commonRepo.WhereBySoftDelete()),
		productPackageRepo.WithProductCategoryMultiLanguageName(commonRepo.WhereBySoftDelete()),
	)
	if err != nil {
		return errors.WithMessage(err, "获取总部商品包列表失败")
	}

	// 商品包
	for _, productPackage := range productPackageList {
		erpSrv := erp.NewIErpSrv(s.dbm)
		var productErpCode string
		classification := productPackage.ProductCategory.MultiLanguageName.GetNames()
		getEnClassification, err := s.getEnName(ctx, classification)
		if err != nil {
			return errors.WithMessage(err, "翻译失败")
		}
		if !productPackage.IsPackage() {
			multiLanguageName := model.NewMultiLanguageName(productPackage.MultiLanguageName.ToJson())
			enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
			if err != nil {
				return errors.WithMessage(err, "翻译失败")
			}
			productCategory, errGetCategory := repository.NewProductCategoryRepo(tx).GetProductCategory(commonRepo.WhereByUuid(productPackage.CategoryUuid))
			if errGetCategory != nil {
				return errors.WithMessage(errGetCategory, "获取商品分类失败")
			}
			productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(productPackage.UnitUuid))
			if errGetUnit != nil {
				return errors.WithMessage(errGetUnit, "获取商品单位失败")
			}
			var flavorUuids []uint64
			for _, v := range productPackage.ProductBoms {
				flavorUuids = append(flavorUuids, v.ProductFlavorUuid)
			}
			flavorList, err := repository.NewProductFlavorRepo(tx).GetProductFlavorList(flavorUuids...)
			if err != nil {
				return errors.WithMessage(err, "获取商品规格失败")
			}
			var flavors []req.Flavor
			for _, v := range flavorList {
				flavors = append(flavors, req.Flavor{
					Name:  v.ErpnextGroupName,
					Value: v.ErpnextValueName,
				})
			}
			itemInfo, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
				ItemName:           enName,
				StockUom:           productUnit.ErpnextUom,
				Classification:     getEnClassification,
				ClassificationCode: productCategory.Code,
				Flavors:            flavors,
			})
			if errErp != nil {
				return errors.WithMessage(errErp, "同步商品到erp失败")
			}
			productErpCode = itemInfo.ItemCode
			err = productPackageRepo.UpdateProductPackage(map[string]any{"erp_code": productErpCode}, commonRepo.WhereByUuid(productPackage.Uuid))
			if err != nil {
				return errors.WithMessage(err, "更新商品包erp编码失败")
			}
		}
		for _, bom := range productPackage.ProductBoms {
			var bomErpCode string
			if productPackage.IsPackage() {
				// 套餐
				productPackage, err := productPackageRepo.GetProductPackage(
					commonRepo.WhereByUuid(productPackage.Uuid),
					commonRepo.Preload(
						repository.WithPreload{
							Query: "ProductUnit",
						},
						repository.WithPreload{
							Query: "ProductCategory.MultiLanguageName",
						},
					),
				)
				if err != nil {
					return errors.WithMessage(err, "获取商品包失败")
				}

				multiLanguageName := model.NewMultiLanguageName(productPackage.Name)
				enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
				if err != nil {
					return errors.WithMessage(err, "翻译失败")
				}
				productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(productPackage.UnitUuid))
				if errGetUnit != nil {
					return errors.WithMessage(errGetUnit, "获取商品单位失败")
				}
				stockUom := productUnit.ErpnextUom

				params := req.PackageAddErpReq{
					ItemName:           enName,
					StockUom:           stockUom,
					InternalCode:       bom.InternalCode,
					Classification:     getEnClassification,
					ClassificationCode: productPackage.ProductCategory.Code,
				}
				itemInfo, errErp := erpSrv.AddPackage(ctx, params)
				if errErp != nil {
					return errors.WithMessage(errErp)
				}
				bomErpCode = itemInfo.ItemCode
			} else {
				// 商品-规格
				itemInfo, errErp := erpSrv.AddProductBom(ctx, req.ProductBomAddErpReq{
					VariantsOf:   productErpCode,
					InternalCode: bom.InternalCode,
					Flavors: []req.Flavor{
						{
							Name:  bom.ProductFlavor.ErpnextGroupName,
							Value: bom.ProductFlavor.ErpnextValueName,
						},
					},
				})
				if errErp != nil {
					return errors.WithMessage(errErp, "同步商品bom到erp失败")
				}
				bomErpCode = itemInfo.ItemCode
			}
			err = productBomRepo.UpdateProductBom(map[string]any{"erp_code": bomErpCode}, commonRepo.WhereByUuid(bom.Uuid))
			if err != nil {
				return errors.WithMessage(err, "更新商品bomerp编码失败")
			}
		}
	}

	return nil
}

func (s *SyncProductToErpSrv) getEnName(ctx context.Context, locale dto.LocaleResponse) (string, error) {
	return GetEnName(ctx, s.settingSrv, locale)
}
