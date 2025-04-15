package service

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/websocket"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
)

type IProductionSrv interface {
	GetProductListByOrder(ctx context.Context, req req.ProductionListReq) (resp.ProductionListWithPagination, error)              // 根据订单获取送厨商品
	GetProductListByCategory(ctx context.Context, req req.ProductionListByCategoryReq) (resp.ProductionListWithPagination, error) // 根据分类获取送厨商品
	GetHistory(ctx context.Context) (resp.ProductionHistory, error)                                                               // 获取上菜历史
	Finish(ctx context.Context, productionUuid uint64) error                                                                      // 完成制作
	Recovery(ctx context.Context, productionUuid uint64) error                                                                    // 恢复制作
}

// productionSrv 收银服务结构体
type productionSrv struct {
	dbm *database.DBManager // 数据库管理器
}

// NewProductionSrv 创建新的收银产品类别服务
func NewProductionSrv(dbm *database.DBManager) IProductionSrv {
	return NewProductionSrvImpl(dbm)
}

// NewProductionSrvImpl 创建新的收银服务实现
func NewProductionSrvImpl(dbm *database.DBManager) IProductionSrv {
	return &productionSrv{
		dbm: dbm,
	}
}

// GetProductListByOrder 根据订单获取送厨商品
func (s *productionSrv) GetProductListByOrder(ctx context.Context, req req.ProductionListReq) (resp.ProductionListWithPagination, error) {
	productPackageUuids, emptyRes, err := s.getProductPackageUuids(ctx)
	if err != nil || len(productPackageUuids) == 0 {
		return emptyRes, err
	}
	productionRepo := repository.NewProductionRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	cookingStatus := productionRepo.WhereProductStatus(constant.ProductionOrderProductStatusCooking)
	inProductPackageUuids := productionRepo.WhereProductPackageUuidIn(productPackageUuids)
	limitedProducts, total, err := productionRepo.GetLimitedProducts(constant.ProductionOrderProductColumnSaleBill, req.PageNo, req.PageSize, cookingStatus, inProductPackageUuids)
	if err != nil {
		return resp.ProductionListWithPagination{}, errors.ErrInternal
	}
	var uuids []uint64
	for _, limitedProduct := range limitedProducts {
		uuids = append(uuids, limitedProduct.SaleBillUuid)
	}
	products, err := productionRepo.GetProducts(0, repository.CreateTimeAsc,
		cookingStatus, inProductPackageUuids, productionRepo.WhereProductSaleBillUuidIn(uuids))
	if err != nil {
		return resp.ProductionListWithPagination{}, errors.ErrInternal
	}
	cookedStatus := productionRepo.WhereProductStatus(constant.ProductionOrderProductStatusFinished)
	finishedList, err := s.getLatestFinishedList(productionRepo, cookedStatus, inProductPackageUuids)
	if err != nil {
		return resp.ProductionListWithPagination{}, errors.ErrInternal
	}
	return resp.ProductionListWithPagination{
		List:         s.groupByOrder(limitedProducts, products),
		FinishedList: finishedList,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

func (s *productionSrv) getProductPackageUuids(ctx context.Context) ([]uint64, resp.ProductionListWithPagination, error) {
	var productPackageUuids []uint64
	emptyResp := resp.ProductionListWithPagination{
		List: make([]resp.ProductionGroup, 0),
		FinishedList: resp.ProductionList{
			List: make([]resp.ProductionItem, 0),
		},
	}
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	deviceRepo := repository.NewDeviceRepo(db)
	device, err := deviceRepo.GetDevice(deviceRepo.WhereSn(ctx.GetDeviceSn()), deviceRepo.WhereSource(ctx.GetSource()))
	if err != nil {
		return productPackageUuids, emptyResp, errors.ErrInternal
	}
	if device.ProductPrinterUuid == 0 {
		return productPackageUuids, emptyResp, nil
	}
	productPrinterRepo := repository.NewProductPrinterRepo(db)
	productPackageUuids, err = productPrinterRepo.GetProductPackageUuids(productPrinterRepo.WhereProductPrinterUuid(device.ProductPrinterUuid))
	if err != nil {
		return productPackageUuids, emptyResp, errors.ErrInternal
	}
	return productPackageUuids, emptyResp, nil
}

// GetProductListByCategory 根据订单获取送厨商品
func (s *productionSrv) GetProductListByCategory(ctx context.Context, req req.ProductionListByCategoryReq) (resp.ProductionListWithPagination, error) {
	productPackageUuids, emptyRes, err := s.getProductPackageUuids(ctx)
	if err != nil || len(productPackageUuids) == 0 {
		return emptyRes, err
	}
	productionRepo := repository.NewProductionRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	cookingStatus := productionRepo.WhereProductStatus(constant.ProductionOrderProductStatusCooking)
	inProductPackageUuids := productionRepo.WhereProductPackageUuidIn(productPackageUuids)
	dbOptions := []repository.DBOption{
		cookingStatus,
		inProductPackageUuids,
	}
	if req.CategoryUuid != 0 {
		dbOptions = append(dbOptions, productionRepo.WhereProductFirstCategoryUuidIn([]uint64{req.CategoryUuid}))
	}
	limitedProducts, total, err := productionRepo.GetLimitedProducts(constant.ProductionOrderProductColumnCategory, req.PageNo, req.PageSize, dbOptions...)
	if err != nil {
		return resp.ProductionListWithPagination{}, errors.WithMessage(errors.ErrInternal)
	}
	var uuids []uint64
	for _, product := range limitedProducts {
		uuids = append(uuids, product.FirstCategoryUuid)
	}

	products, err := productionRepo.GetProducts(0, repository.CreateTimeAsc,
		cookingStatus, inProductPackageUuids, productionRepo.WhereProductFirstCategoryUuidIn(uuids),
		productionRepo.WithProductCategory(), productionRepo.WithProductCategoryMultiLanguageName())
	if err != nil {
		return resp.ProductionListWithPagination{}, errors.WithMessage(errors.ErrInternal)
	}
	groups := make([]resp.ProductionGroup, 0)
	for _, paginatedProduct := range limitedProducts {
		var group resp.ProductionGroup
		items := make([]resp.ProductionItem, 0)
		for _, product := range products {
			if paginatedProduct.FirstCategoryUuid != product.FirstCategoryUuid {
				continue
			}
			if product.ProductCategory.MultiLanguageName.Uuid != 0 && group.LocaleName == nil {
				localName := product.ProductCategory.MultiLanguageName.GetNames()
				group.LocaleName = &localName
			}
			var item resp.ProductionItem
			copier.Copy(&item, product)
			item.LocaleName = product.SaleOrderProduct.MultiLanguageName.GetNames()
			item.ProductAttributeNames = product.SaleOrderProduct.GetAttributeName()
			item.SerialNo = product.SaleBill.SerialNo
			item.DiningMethod = product.SaleBill.DiningMethod
			items = append(items, item)
		}
		if group.LocaleName == nil {
			group.LocaleName = &dto.LocaleResponse{}
		}
		group.ProductionList = resp.ProductionList{
			List: items,
		}
		groups = append(groups, group)
	}
	cookedStatus := productionRepo.WhereProductStatus(constant.ProductionOrderProductStatusFinished)
	finishedList, err := s.getLatestFinishedList(productionRepo, cookedStatus, inProductPackageUuids)
	if err != nil {
		return resp.ProductionListWithPagination{}, errors.WithMessage(errors.ErrInternal)
	}
	return resp.ProductionListWithPagination{
		List:         groups,
		FinishedList: finishedList,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// 最近上菜历史
func (s *productionSrv) getLatestFinishedList(repo repository.IProductionOrderRepo, opts ...repository.DBOption) (resp.ProductionList, error) {
	products, err := repo.GetProducts(3, repository.FinishedTimeDesc, opts...)
	if err != nil {
		return resp.ProductionList{}, errors.ErrInternal
	}
	items := make([]resp.ProductionItem, 0, len(products))
	for _, product := range products {
		var item resp.ProductionItem
		copier.Copy(&item, product)
		item.LocaleName = product.SaleOrderProduct.MultiLanguageName.GetNames()
		item.SerialNo = product.SaleBill.SerialNo
		item.DiningMethod = product.SaleBill.DiningMethod
		items = append(items, item)
	}
	return resp.ProductionList{
		List: items,
	}, nil
}

// GetHistory 获取上菜历史
func (s *productionSrv) GetHistory(ctx context.Context) (resp.ProductionHistory, error) {
	// 获取过去24小时内的上菜历史，按照上菜时间倒序
	productionRepo := repository.NewProductionRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	statusOption := productionRepo.WhereProductStatus(constant.ProductionOrderProductStatusFinished)
	finishedTimeOption := productionRepo.WhereProductFinishedTime(time.Now().Add(-1 * time.Hour * 24).Unix())
	limitProducts, err := productionRepo.GetLimitedHistoryProducts(statusOption, finishedTimeOption)
	if err != nil {
		return resp.ProductionHistory{}, errors.WithMessage(errors.ErrInternal)
	}

	products, err := productionRepo.GetProducts(0, repository.CreateTimeAsc,
		statusOption, finishedTimeOption, productionRepo.WhereProductHistoryCondition())
	if err != nil {
		return resp.ProductionHistory{}, errors.WithMessage(errors.ErrInternal)
	}

	return resp.ProductionHistory{
		List: s.groupByOrder(limitProducts, products),
	}, nil
}

// 根据销售账单分组
func (s *productionSrv) groupByOrder(limitProducts []model.ProductionOrderProduct, products []model.ProductionOrderProduct) []resp.ProductionGroup {
	groups := make([]resp.ProductionGroup, 0, len(limitProducts))
	for _, paginatedProduct := range limitProducts {
		var group resp.ProductionGroup
		items := make([]resp.ProductionItem, 0)
		for _, product := range products {
			if paginatedProduct.SaleBillUuid != product.SaleBillUuid {
				continue
			}
			if product.SaleBill.SerialNo != "" && group.LocaleName == nil {
				group.LocaleName = &dto.LocaleResponse{
					ZH:   product.SaleBill.SerialNo,
					TH:   product.SaleBill.SerialNo,
					EN:   product.SaleBill.SerialNo,
					ZHTW: product.SaleBill.SerialNo,
					JA:   product.SaleBill.SerialNo,
					KO:   product.SaleBill.SerialNo,
					MY:   product.SaleBill.SerialNo,
					TR:   product.SaleBill.SerialNo,
				}
			}
			var item resp.ProductionItem
			err := copier.Copy(&item, product)
			if err != nil {
				logger.Logger.Error("copier error", zap.Error(err))
			}
			item.LocaleName = product.SaleOrderProduct.MultiLanguageName.GetNames()
			item.ProductAttributeNames = product.SaleOrderProduct.GetAttributeName()
			item.SerialNo = product.SaleBill.SerialNo
			item.DiningMethod = product.SaleBill.DiningMethod
			items = append(items, item)
		}
		if group.LocaleName == nil {
			group.LocaleName = &dto.LocaleResponse{}
		}
		group.ProductionList = resp.ProductionList{
			List: items,
		}
		if len(items) > 0 {
			group.DiningMethod = items[0].DiningMethod
		}
		groups = append(groups, group)
	}

	return groups
}

// Finish 完成制作
func (s *productionSrv) Finish(ctx context.Context, productUuid uint64) error {
	productionRepo := repository.NewProductionRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	product, _ := productionRepo.GetProduct(productionRepo.WhereProductUuid(productUuid))
	if product.Uuid == 0 {
		return errors.New("订单商品不存在")
	}
	if product.Status != constant.ProductionOrderProductStatusCooking {
		return errors.New("订单商品未送厨")
	}
	if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereProductUuid(productUuid)}, map[string]any{
		"status":        constant.ProductionOrderProductStatusFinished,
		"finished_time": time.Now().Unix(),
	}); err != nil {
		return errors.ErrInternal
	}
	// 送厨成功后，推送更新订单
	go websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]interface{}{
		"update_time": time.Now().Unix(),
	})
	return nil

}

// Recovery 恢复制作
func (s *productionSrv) Recovery(ctx context.Context, productUuid uint64) error {
	productionRepo := repository.NewProductionRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	product, _ := productionRepo.GetProduct(productionRepo.WhereProductUuid(productUuid))
	if product.Uuid == 0 {
		return errors.New("订单商品不存在")
	}
	if product.Status != constant.ProductionOrderProductStatusFinished {
		return errors.New("订单商品未完成")
	}
	if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereProductUuid(productUuid)}, map[string]any{
		"status":        constant.ProductionOrderProductStatusCooking,
		"finished_time": 0,
	}); err != nil {
		return errors.ErrInternal
	}
	// 送厨成功后，推送更新订单
	go websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]interface{}{
		"update_time": time.Now().Unix(),
	})
	return nil
}
