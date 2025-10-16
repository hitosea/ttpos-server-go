package service

import (
	"fmt"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
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

	// 同步规格
	err := s.productSrv.UpdateProductFlavorErp(ctx, tx)
	if err != nil {
		return errors.WithMessage(err, "更新规格到erp失败")
	}

	// 同步商品
	commonRepo := repository.NewCommonRepo()
	productPackageRepo := repository.NewProductPackageRepo(tx)
	productPackageList, err := productPackageRepo.GetProductPackageList(
		commonRepo.WhereBySoftDelete(),
		commonRepo.WhereByHeadquarterUuid(0),
		// productPackageRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
		productPackageRepo.WithProductBoms(commonRepo.WhereBySoftDelete()),
		// productPackageRepo.WithProductPackageAttributeGroups(commonRepo.WhereBySoftDelete()),
		// productPackageRepo.WithProductPackageAttributeGroupAttributes(commonRepo.WhereBySoftDelete()),
		productPackageRepo.WithProductPackageGroups(commonRepo.WhereBySoftDelete()),
		productPackageRepo.WithProductPackageGroupItems(commonRepo.WhereBySoftDelete()),
		// productPackageRepo.WithProductPackageGroupMultiLanguageName(commonRepo.WhereBySoftDelete()),
	)
	if err != nil {
		return errors.WithMessage(err, "获取总部商品包列表失败")
	}

	fmt.Println(productPackageList)

	// 商品加料
	// 商品规格
	// 商品包
	// for _, productPackage := range productPackageList {
	// 	multiLanguageName := model.NewMultiLanguageName(request.LocaleName.ToJson())
	// 	enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
	// 	if err != nil {
	// 		return nil, errors.WithMessage(err, "翻译失败")
	// 	}
	// 	productUnit, errGetUnit := repository.NewProductUnitRepo(tx).GetProductUnit(commonRepo.WhereByUuid(request.UnitUuid))
	// 	if errGetUnit != nil {
	// 		return nil, errors.WithMessage(errGetUnit, "获取商品单位失败")
	// 	}

	// 	flavorUuids := make([]uint64, 0, len(request.Flavors))
	// 	for _, v := range request.Flavors {
	// 		flavorUuids = append(flavorUuids, v.Uuid)
	// 	}
	// 	flavorList, err := repository.NewProductFlavorRepo(tx).GetProductFlavorList(flavorUuids...)
	// 	if err != nil {
	// 		return nil, errors.WithMessage(err, "获取商品规格失败")
	// 	}
	// 	var flavors []req.Flavor
	// 	for _, v := range flavorList {
	// 		flavors = append(flavors, req.Flavor{
	// 			Name:  v.ErpnextGroupName,
	// 			Value: v.ErpnextValueName,
	// 		})
	// 	}
	// 	erpSrv := erp.NewIErpSrv(s.dbm)
	// 	itemInfo, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
	// 		ItemName: enName,
	// 		StockUom: productUnit.ErpnextUom,
	// 		Flavors:  flavors,

	// 		// ItemName           string   `json:"item_name" binding:"required"`           // 商品名称, 英文
	// 		// StockUom           string   `json:"stock_uom" binding:"required"`           // 商品单位, 英文
	// 		// ItemCode           string   `json:"item_code" binding:"required"`           // 商品编码，如果为空，则为新增；如果非空，则为编辑
	// 		// TemplateItemCode   string   `json:"template_item_code" binding:"required"`  // 模版商品编码
	// 		// ItemSpecification  string   `json:"item_specification" binding:"required"`  // 商品规格
	// 		// BarcodeValue       string   `json:"barcode_value" binding:"required"`       // 条形码值
	// 		// Classification     string   `json:"classification" binding:"required"`      // 分类
	// 		// ClassificationCode string   `json:"classification_code" binding:"required"` // 分类编码
	// 		// InternalCode       string   `json:"internal_code" `                         // 内部编码
	// 		// Flavors            []Flavor `json:"flavors" `                               // 规格列表
	// 	})
	// 	if errErp != nil {
	// 		return nil, errors.WithMessage(errErp, "同步商品到erp失败")
	// 	}
	// 	erpCode = itemInfo.ItemCode
	// }

	return nil
}

func (s *SyncProductToErpSrv) getEnName(ctx context.Context, locale dto.LocaleResponse) (string, error) {
	return GetEnName(ctx, s.settingSrv, locale)
}
