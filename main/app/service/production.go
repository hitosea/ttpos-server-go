package service

import (
	"fmt"
	"slices"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/eventbus/event"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IProductionSrv interface {
	GetProductListByOrder(ctx context.Context, req req.ProductionListReq) (resp.ProductionListWithPagination, error)              // 根据订单获取送厨商品
	GetProductListByCategory(ctx context.Context, req req.ProductionListByCategoryReq) (resp.ProductionListWithPagination, error) // 根据分类获取送厨商品
	GetHistory(ctx context.Context, req req.HistoryReq) (resp.ProductionHistory, error)                                           // 获取制作完成、传菜完成历史
	Finish(ctx context.Context, req req.FinishReq) error                                                                          // 完成制作、传菜
	Recovery(ctx context.Context, req req.RecoveryReq) error                                                                      // 恢复制作
	ConfirmReturn(ctx context.Context, productUuid uint64) error                                                                  // 厨显端确认退菜
	ConfirmReturnAll(ctx context.Context, saleBillUuid uint64) error                                                              // 厨显端确认退菜整单
}

// productionSrv 收银服务结构体
type productionSrv struct {
	dbm        *database.DBManager // 数据库管理器
	settingSrv setting.ISrv
}

// NewProductionSrv 创建新的收银产品类别服务
func NewProductionSrv(dbm *database.DBManager, settingSrv setting.ISrv) IProductionSrv {
	return NewProductionSrvImpl(dbm, settingSrv)
}

// NewProductionSrvImpl 创建新的收银服务实现
func NewProductionSrvImpl(dbm *database.DBManager, settingSrv setting.ISrv) IProductionSrv {
	return &productionSrv{
		dbm:        dbm,
		settingSrv: settingSrv,
	}
}

func (s *productionSrv) getMode(ctx context.Context, reqMode uint) (*uint, error) {
	var mode *uint = nil
	kitchenSetting, err := s.settingSrv.GetKitchenSetting(ctx, ctx.GetCompanySetting(), nil)
	if err != nil {
		return mode, errors.WithMessage(errors.New("获取厨显设置失败"), err.Error())
	}
	if kitchenSetting.IsSmartKitchen == "1" {
		deviceRepo := repository.NewDeviceRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
		device, err := deviceRepo.GetDevice(deviceRepo.WhereSn(ctx.GetDeviceSn()), deviceRepo.WhereSource(ctx.GetSource()))
		if err != nil {
			return nil, errors.WithMessage(errors.New("获取设备信息失败"), err.Error())
		}
		switch device.KdsMode {
		case constant.KdsModeDefault: // 单传菜模式，只接受 req.Mode = 0
			if reqMode != constant.ReqModeSend {
				return nil, errors.WithMessage(errors.New("本机不支持查看待制作的菜品"))
			}
		case constant.KdsModeMake: // 仅制作模式，只接受 req.Mode = 1
			if reqMode != constant.ReqModeMake {
				return nil, errors.WithMessage(errors.New("本机不支持查看待传菜的菜品"))
			}
		case constant.KdsModeMakeAndSend: // 制作+传菜模式，只接受 req.Mode = 1 or req.Mode = 0
			if reqMode != constant.ReqModeMake && reqMode != constant.ReqModeSend {
				return nil, errors.WithMessage(errors.New("本机不支持查看当前状态的菜品"))
			}
		}
		mode = &reqMode
	}
	return mode, nil
}

// GetProductListByOrder 根据订单获取送厨商品
func (s *productionSrv) GetProductListByOrder(ctx context.Context, req req.ProductionListReq) (resp.ProductionListWithPagination, error) {
	mode, err := s.getMode(ctx, req.Mode)
	if err != nil {
		return resp.ProductionListWithPagination{}, err
	}
	productPackageUuids, saleBillUuids, emptyRes, err := s.getProductPackageUuidsAndSaleBillUuids(ctx)
	if err != nil || len(productPackageUuids) == 0 {
		return emptyRes, err
	}
	productionRepo := repository.NewProductionRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	statusOpt := productionRepo.WhereProductStatus(constant.ProductionOrderProductStatusCooking)
	productPackageUuidOpt := productionRepo.WhereProductPackageUuidIn(productPackageUuids)
	saleBillUuidOpt := productionRepo.WhereSaleBillUuidIn(saleBillUuids)

	limitedProductOpts := []repository.DBOption{
		statusOpt,
		productPackageUuidOpt,
		saleBillUuidOpt,
	}
	if mode != nil && *mode == 1 { // 制作模式，只显示未制作完成和已恢复到制作中的商品
		limitedProductOpts = append(limitedProductOpts, productionRepo.WhereProductMakeStatus([]uint{
			constant.ProductionOrderProductMakeStatusDefault,
			constant.ProductionOrderProductMakeStatusRecovery,
		}))
	}
	// 2.4.0 版本之前，只显示大于0的商品
	if !ctx.Version(context.GTE, "2.4.0") {
		limitedProductOpts = append(limitedProductOpts, productionRepo.WhereProductNumGT0())
	}

	// 非分批商品、或者分批已送厨商品
	limitedProductOpts = append(limitedProductOpts, productionRepo.WhereIsNotBatchOrBatchTimeGT0())
	limitedProducts, total, err := productionRepo.GetLimitedProducts(constant.ProductionOrderProductColumnSaleBill, req.PageNo, req.PageSize, limitedProductOpts...)
	if err != nil {
		return resp.ProductionListWithPagination{}, errors.ErrInternal
	}
	var uuids []uint64
	for _, limitedProduct := range limitedProducts {
		uuids = append(uuids, limitedProduct.SaleBillUuid)
	}
	productOpts := []repository.DBOption{
		productPackageUuidOpt,
		saleBillUuidOpt,
		productionRepo.WhereSaleBillUuidIn(uuids),
	}
	if mode != nil && *mode == 1 { // 制作模式，只显示未制作完成和已恢复到制作中的商品
		productOpts = append(productOpts, productionRepo.WhereProductMakeStatus([]uint{
			constant.ProductionOrderProductMakeStatusDefault,
			constant.ProductionOrderProductMakeStatusRecovery,
		}))
	}
	// 2.4.0 版本之前，只显示大于0的商品
	if !ctx.Version(context.GTE, "2.4.0") {
		productOpts = append(productOpts, productionRepo.WhereProductNumGT0())
	}
	sendKitchenNum, products, err := productionRepo.GetProducts(0, repository.CreateTimeAsc, statusOpt, productOpts...)
	if err != nil {
		return resp.ProductionListWithPagination{}, errors.ErrInternal
	}
	finishedList, err := s.getFinishedList(productionRepo, mode, productPackageUuidOpt, saleBillUuidOpt)
	if err != nil {
		return resp.ProductionListWithPagination{}, err
	}
	return resp.ProductionListWithPagination{
		SendKitchenNum: sendKitchenNum,
		List:           s.groupByOrder(ctx, limitedProducts, products, nil),
		FinishedList:   finishedList,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

func (s *productionSrv) getProductPackageUuidsAndSaleBillUuids(ctx context.Context) ([]uint64, []uint64, resp.ProductionListWithPagination, error) {
	var productPackageUuids, saleBillUuids []uint64
	emptyResp := resp.ProductionListWithPagination{
		List: make([]resp.ProductionGroup, 0),
		FinishedList: resp.ProductionList{
			List: make([]resp.ProductionItem, 0),
		},
	}
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	// 获取厨显设备信息
	deviceRepo := repository.NewDeviceRepo(db)
	device, err := deviceRepo.GetDevice(deviceRepo.WhereSn(ctx.GetDeviceSn()), deviceRepo.WhereSource(ctx.GetSource()))
	if err != nil {
		return productPackageUuids, saleBillUuids, emptyResp, errors.ErrInternal
	}
	if device.ProductPrinterUuid == 0 {
		return productPackageUuids, saleBillUuids, emptyResp, nil
	}
	// 厨显绑定了商品打印机
	// 从商品打印设置中获取打印机关联的商品Uuid
	productPrinterRepo := repository.NewProductPrinterRepo(db)
	productPackageUuids, err = productPrinterRepo.GetProductPackageUuids(productPrinterRepo.WhereProductPrinterUuid(device.ProductPrinterUuid))
	if err != nil {
		return productPackageUuids, saleBillUuids, emptyResp, errors.ErrInternal
	}

	var opt repository.DBOption
	// 如果版本号大于等于2.4.0，则获取厨显端未确认退菜整单的账单Uuid；否则获取未被删除的账单Uuid
	if ctx.Version(context.GTE, "2.4.0") {
		opt = productPrinterRepo.WhereSaleBillIsKitchenConfirm(0)
	} else {
		opt = productPrinterRepo.WhereSaleBillNotDeletedOrIsNotCanceled() // 未被删除的，未整单取消的
	}

	// 从商品打印设置中获取区域ID，根据区域ID获取销售账单Uuid
	saleBillUuids, err = productPrinterRepo.GetProductionSaleBillUuid(device.ProductPrinterUuid, opt)
	if err != nil {
		return productPackageUuids, saleBillUuids, emptyResp, errors.ErrInternal
	}
	return productPackageUuids, saleBillUuids, emptyResp, nil
}

// GetProductListByCategory 根据订单获取送厨商品
func (s *productionSrv) GetProductListByCategory(ctx context.Context, req req.ProductionListByCategoryReq) (resp.ProductionListWithPagination, error) {
	mode, err := s.getMode(ctx, req.Mode)
	if err != nil {
		return resp.ProductionListWithPagination{}, err
	}
	productPackageUuids, saleBillUuids, emptyRes, err := s.getProductPackageUuidsAndSaleBillUuids(ctx)
	if err != nil || len(productPackageUuids) == 0 {
		return emptyRes, err
	}
	language := ctx.GetLanguage()
	productionRepo := repository.NewProductionRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	statusOpt := productionRepo.WhereProductStatus(constant.ProductionOrderProductStatusCooking)
	productPackageUuidOpt := productionRepo.WhereProductPackageUuidIn(productPackageUuids)
	saleBillUuidOpt := productionRepo.WhereSaleBillUuidIn(saleBillUuids)
	limitedProductOpts := []repository.DBOption{
		statusOpt,
		productPackageUuidOpt,
		saleBillUuidOpt,
	}
	if mode != nil && *mode == 1 { // 制作模式，只显示未制作完成和已恢复到制作中的商品
		limitedProductOpts = append(limitedProductOpts, productionRepo.WhereProductMakeStatus([]uint{
			constant.ProductionOrderProductMakeStatusDefault,
			constant.ProductionOrderProductMakeStatusRecovery,
		}))
	}
	// 2.4.0 版本之前，只显示大于0的商品
	if !ctx.Version(context.GTE, "2.4.0") {
		limitedProductOpts = append(limitedProductOpts, productionRepo.WhereProductNumGT0())
	}

	if req.CategoryUuid != 0 { // 指定分类Uuid
		limitedProductOpts = append(limitedProductOpts, productionRepo.WhereProductFirstCategoryUuidIn([]uint64{req.CategoryUuid}))
	}
	// 非分批商品、或者分批已送厨商品
	limitedProductOpts = append(limitedProductOpts, productionRepo.WhereIsNotBatchOrBatchTimeGT0())
	limitedProducts, total, err := productionRepo.GetLimitedProducts(constant.ProductionOrderProductColumnCategory, req.PageNo, req.PageSize, limitedProductOpts...)
	if err != nil {
		return resp.ProductionListWithPagination{}, errors.WithMessage(errors.ErrInternal)
	}
	var uuids []uint64
	for _, product := range limitedProducts {
		uuids = append(uuids, product.FirstCategoryUuid)
	}
	productOpts := []repository.DBOption{
		productPackageUuidOpt,
		saleBillUuidOpt,
		productionRepo.WhereProductFirstCategoryUuidIn(uuids),
		productionRepo.WithProductCategory(),
		productionRepo.WithProductCategoryMultiLanguageName(),
	}
	if mode != nil && *mode == 1 { // 制作模式，只显示未制作完成和已恢复到制作中的商品
		productOpts = append(productOpts, productionRepo.WhereProductMakeStatus([]uint{
			constant.ProductionOrderProductMakeStatusDefault,
			constant.ProductionOrderProductMakeStatusRecovery,
		}))
	}
	// 2.4.0 版本之前，只显示大于0的商品
	if !ctx.Version(context.GTE, "2.4.0") {
		productOpts = append(productOpts, productionRepo.WhereProductNumGT0())
	}
	sendKitchenNum, products, err := productionRepo.GetProducts(0, repository.CreateTimeAsc, statusOpt, productOpts...)
	if err != nil {
		return resp.ProductionListWithPagination{}, errors.WithMessage(errors.ErrInternal)
	}
	groups := make([]resp.ProductionGroup, 0)
	for _, paginatedProduct := range limitedProducts {
		var group resp.ProductionGroup
		items := make([]resp.ProductionItem, 0)
		for _, product := range products {
			// 获取门店业务设置
			businessSetting, errGet := s.settingSrv.GetBusinessSetting(ctx)
			if errGet != nil {
				return resp.ProductionListWithPagination{}, errors.WithMessage(errors.ErrInternal)
			}
			if businessSetting.OpenIsBatch() {
				// 如果开启了分批送厨， 如果商品是分批商品，且处于预送厨阶段，则不显示
				if product.IsBatchBool() && product.IsPreCooking() {
					continue
				}
			}
			if paginatedProduct.FirstCategoryUuid != product.FirstCategoryUuid {
				continue
			}
			if product.ProductCategory.MultiLanguageName.Uuid != 0 && group.LocaleName == nil {
				localName := product.ProductCategory.MultiLanguageName.GetNames()
				group.LocaleName = &localName
			}
			var item resp.ProductionItem
			copier.Copy(&item, product)

			// 如果商品是分批商品，则以分批送厨的时间为正式的送厨时间
			if product.IsBatchBool() {
				item.CreateTime = product.BatchTime
			}

			if product.SaleOrderProduct.IsPackageSubProduct() && item.Remark != "" {
				item.Remark = i18n.Translate(language, "套餐备注：") + item.Remark
			}
			item.LocaleName = product.SaleOrderProduct.MultiLanguageName.GetNames()
			item.ProductAttributeNames = product.SaleOrderProduct.GetAttributeName()
			item.SerialNo = product.SaleBill.SerialNo
			item.DiningMethod = product.GetWrapStatus()                                                                            // 订单商品的打包状态
			item.IsSaleBillDeleted = product.SaleBill.DeleteTime > 0 || product.SaleBill.Status == constant.SaleBillStatusCanceled // 是否已经整单取消
			item.BatchTag = func() resp.BatchTagInfo {
				// 获取门店业务设置
				businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
				if err != nil {
					return resp.BatchTagInfo{}
				}
				if businessSetting.OpenIsBatch() {
					if product.IsBatchBool() && product.BatchTag != nil {
						return resp.BatchTagInfo{
							Uuid:       product.BatchTagUuid,
							LocaleName: product.BatchTag.MultiLanguageName.GetNames(),
							Color:      product.BatchTag.Color,
						}
					}
				}
				return resp.BatchTagInfo{}
			}()
			item.OrderRemark = func() resp.OrderRemarkRes {
				remark, err := product.SaleBill.GetOrderRemarkRes()
				if err != nil {
					return resp.OrderRemarkRes{}
				}
				if remark == nil {
					return resp.OrderRemarkRes{}
				}
				return *remark
			}()
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
	finishedList, err := s.getFinishedList(productionRepo, mode, productPackageUuidOpt, saleBillUuidOpt)
	if err != nil {
		return resp.ProductionListWithPagination{}, err
	}
	return resp.ProductionListWithPagination{
		SendKitchenNum: sendKitchenNum,
		List:           groups,
		FinishedList:   finishedList,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// getFinishedList 获取最近上菜历史
func (s *productionSrv) getFinishedList(productionRepo repository.IProductionOrderRepo, mode *uint, opts ...repository.DBOption) (resp.ProductionList, error) {
	var finishStatusOpt repository.DBOption = nil
	var orderBy string = repository.FinishedTimeDesc
	if mode != nil {
		switch *mode {
		case constant.KdsModeMake: // 制作模式，只显示已制作完成的商品
			opts = append(opts, productionRepo.WhereProductMakeStatus([]uint{constant.ProductionOrderProductMakeStatusFinished}))
			orderBy = repository.MadeTimeDesc
		case constant.KdsModeDefault: // 传菜模式
			finishStatusOpt = productionRepo.WhereProductStatus(constant.ProductionOrderProductStatusFinished)
		}
	} else {
		finishStatusOpt = productionRepo.WhereProductStatus(constant.ProductionOrderProductStatusFinished)
	}
	finishedList, err := s.getLatestFinishedList(productionRepo, orderBy, finishStatusOpt, opts...)
	if err != nil {
		errMsg := "获取最近上菜历史失败"
		if mode != nil {
			switch *mode {
			case constant.KdsModeMake:
				errMsg = "获取最近制作历史失败"
			case constant.KdsModeDefault:
				errMsg = "获取最近上菜历史失败"
			}
		}
		return resp.ProductionList{}, errors.WithMessage(errors.New(errMsg), err.Error())
	}
	return finishedList, nil
}

// 最近上菜历史
func (s *productionSrv) getLatestFinishedList(productionRepo repository.IProductionOrderRepo, orderBy string, statusOpt repository.DBOption, opts ...repository.DBOption) (resp.ProductionList, error) {
	_, products, err := productionRepo.GetProducts(3, orderBy, statusOpt, opts...)
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

// GetHistory 获取制作完成、传菜完成历史
func (s *productionSrv) GetHistory(ctx context.Context, req req.HistoryReq) (resp.ProductionHistory, error) {
	// 获取过去24小时内的制作、传菜历史，按照制作、传菜时间倒序
	productionRepo := repository.NewProductionRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))

	mode, err := s.getMode(ctx, req.Mode)
	if err != nil {
		return resp.ProductionHistory{}, err
	}

	var statusOpt, finishedTimeOpt repository.DBOption
	var limitedOrderField, productsOrderBy string

	if mode != nil && *mode == constant.KdsModeMake {
		statusOpt = productionRepo.WhereProductMakeStatus([]uint{constant.ProductionOrderProductMakeStatusFinished})
		finishedTimeOpt = productionRepo.WhereProductMadeTime(time.Now().Add(-1 * time.Hour * 24).Unix())
		limitedOrderField = "made_time"
		productsOrderBy = repository.MadeTimeDesc
	} else {
		statusOpt = productionRepo.WhereProductStatus(constant.ProductionOrderProductStatusFinished)
		finishedTimeOpt = productionRepo.WhereProductFinishedTime(time.Now().Add(-1 * time.Hour * 24).Unix())
		limitedOrderField = "finished_time"
		productsOrderBy = repository.FinishedTimeDesc
	}

	limitProducts, err := productionRepo.GetLimitedHistoryProducts(limitedOrderField, statusOpt, finishedTimeOpt)
	if err != nil {
		return resp.ProductionHistory{}, errors.WithMessage(errors.ErrInternal)
	}
	_, products, err := productionRepo.GetProducts(0, productsOrderBy, statusOpt, finishedTimeOpt, productionRepo.SaleBillUuidOpt())
	if err != nil {
		return resp.ProductionHistory{}, errors.WithMessage(errors.ErrInternal)
	}
	return resp.ProductionHistory{
		List: s.groupByOrder(ctx, limitProducts, products, mode),
	}, nil
}

// 根据销售账单分组
func (s *productionSrv) groupByOrder(ctx context.Context, limitProducts []model.ProductionOrderProduct, products []model.ProductionOrderProduct, mode *uint) []resp.ProductionGroup {
	groups := make([]resp.ProductionGroup, 0, len(limitProducts))
	language := ctx.GetLanguage()
	modeMake := mode != nil && *mode == constant.ReqModeMake
	modeSend := mode != nil && *mode == constant.ReqModeSend
	for _, paginatedProduct := range limitProducts {
		var group resp.ProductionGroup
		items := make([]resp.ProductionItem, 0) // 生产单商品列表
		for _, product := range products {
			// 获取门店业务设置
			businessSetting, errGet := s.settingSrv.GetBusinessSetting(ctx)
			if errGet != nil {
				return nil
			}
			if businessSetting.OpenIsBatch() {
				// 如果开启了分批送厨， 如果商品是分批商品，且处于预送厨阶段，则不显示
				if product.IsBatchBool() && product.IsPreCooking() {
					continue
				}
			}
			if paginatedProduct.SaleBillUuid != product.SaleBillUuid {
				continue
			}
			group.DiningMethod = product.SaleBill.DiningMethod                                                                      // 订单商品的打包状态
			group.SaleBillUuid = product.SaleBillUuid                                                                               // 销售账单Uuid
			group.IsSaleBillDeleted = product.SaleBill.DeleteTime > 0 || product.SaleBill.Status == constant.SaleBillStatusCanceled // 是否已经整单取消
			group.IsTakeoutBill = product.SaleBill.IsTakeoutBill()
			group.OrderRemark = func() resp.OrderRemarkRes {
				remark, err := product.SaleBill.GetOrderRemarkRes()
				if err != nil {
					return resp.OrderRemarkRes{}
				}
				if remark == nil {
					return resp.OrderRemarkRes{}
				}
				return *remark
			}()
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
					SV:   product.SaleBill.SerialNo,
				}
			}
			var item resp.ProductionItem
			err := copier.Copy(&item, product)
			if err != nil {
				logger.Logger.Error("copier error", zap.Error(err))
			}

			// 如果商品是分批商品，则以分批送厨的时间为正式的送厨时间
			if product.IsBatchBool() {
				item.CreateTime = product.BatchTime
			}

			if product.SaleOrderProduct.IsPackageSubProduct() && item.Remark != "" {
				item.Remark = i18n.Translate(language, "套餐备注：") + item.Remark
			}
			if modeMake {
				item.FinishedTime = product.MadeTime
				item.MakeDuration = product.MakeDuration
			}
			if modeSend {
				item.SendDuration = func() int64 {
					if product.SendDuration == 0 {
						return product.MakeDuration // 如果传菜时长为0，则以制作完成时间作为传菜时间. 这个场景下是表示这个菜是在未开启智能后厨时制作完成的,所以没有传菜时长记录.
					}
					return product.SendDuration
				}()
			}
			item.LocaleName = product.SaleOrderProduct.MultiLanguageName.GetNames()
			item.ProductAttributeNames = product.SaleOrderProduct.GetAttributeName()
			item.SerialNo = product.SaleBill.SerialNo
			item.DiningMethod = product.GetWrapStatus()                                                                            // 订单商品的打包状态
			item.IsSaleBillDeleted = product.SaleBill.DeleteTime > 0 || product.SaleBill.Status == constant.SaleBillStatusCanceled // 是否已经整单取消
			item.BatchTag = func() resp.BatchTagInfo {
				// 获取门店业务设置
				businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
				if err != nil {
					return resp.BatchTagInfo{}
				}
				if businessSetting.OpenIsBatch() {
					if product.IsBatchBool() && product.BatchTag != nil {
						return resp.BatchTagInfo{
							Uuid:       product.BatchTagUuid,
							LocaleName: product.BatchTag.MultiLanguageName.GetNames(),
							Color:      product.BatchTag.Color,
						}
					}
				}
				return resp.BatchTagInfo{}
			}()
			items = append(items, item)
		}
		if group.LocaleName == nil {
			group.LocaleName = &dto.LocaleResponse{}
		}
		group.ProductionList = resp.ProductionList{
			List: items,
		}
		// 如果没有子商品，则不显示
		if len(group.ProductionList.List) == 0 {
			continue
		}
		groups = append(groups, group)
	}

	return groups
}

// 更新生产单套餐商品状态
func (s *productionSrv) updatePackageProduct(tx *gorm.DB, saleOrderProductUuids []uint64, finishedTime int64) error {
	saleOrderProductRepo := repository.NewSaleOrderProductRepo(tx)
	saleOrderProducts, err := saleOrderProductRepo.GetSaleOrderProductsByUuids(saleOrderProductUuids)
	if err != nil {
		return err
	}

	var packageUuids []uint64
	for _, saleOrderProduct := range saleOrderProducts {
		if !saleOrderProduct.IsPackageSubProduct() || slices.Contains(packageUuids, saleOrderProduct.PackageUuid) {
			continue
		}
		packageUuids = append(packageUuids, saleOrderProduct.PackageUuid)
	}
	if len(packageUuids) == 0 {
		return nil
	}

	for _, packageUuid := range packageUuids {
		productionRepo := repository.NewProductionRepo(tx)
		// 根据销售单套餐uuid获取送厨单子商品商品，不存在子商品一半状态是退菜
		// packageUuid := saleOrderProduct.PackageUuid // 套餐商品在saleOrderProduct表的uuid
		productionProducts, err := productionRepo.GetProductsByPackageUuid(packageUuid)
		if err != nil {
			return err
		}
		for _, productionProduct := range productionProducts {
			if productionProduct.Status == constant.ProductionOrderProductStatusCooking {
				// 修改生产单套餐商品为制作中
				if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereSaleOrderProductUuid(packageUuid)}, map[string]any{
					"status":        constant.ProductionOrderProductStatusCooking,
					"finished_time": 0,
					"make_duration": 0, // 清空时长记录
					"all_duration":  0,
				}); err != nil {
					return errors.WithMessage(errors.New("更新送厨单套餐商品状态失败"), err.Error())
				}
				return nil // 套餐子商品状态是制作中，直接返回. 只要有一个子商品未制作完成，则套餐商品状态就为制作中
			}
		}
		// 如果套餐下所有子商品都已制作完成(都不是制作中)，则修改生产单套餐商品为制作完成.
		// 并记录套餐的制作耗时,现在时间-送厨时间
		var makeDuration int64 = 0
		if len(productionProducts) > 0 {
			createTime := productionProducts[0].GetCreateTime()
			if finishedTime > createTime {
				makeDuration = finishedTime - createTime
			}
		}
		if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereSaleOrderProductUuid(packageUuid)}, map[string]any{
			"status":        constant.ProductionOrderProductStatusFinished,
			"finished_time": time.Now().Unix(),
			"make_duration": makeDuration,
			"all_duration":  makeDuration,
		}); err != nil {
			return errors.WithMessage(errors.New("更新送厨单套餐商品状态失败"), err.Error())
		}
	}

	return nil
}

// Finish 完成制作
func (s *productionSrv) Finish(ctx context.Context, req req.FinishReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	var allProductUuids, productUuids []uint64 // allProductUuids: 所有菜品ID列表，包括确认退菜的菜品; productUuids: 要完成的菜品ID列表
	if req.ProductUuid != 0 {
		productUuids = append(productUuids, req.ProductUuid)
	}
	if len(req.ProductUuids) > 0 {
		productUuids = append(productUuids, req.ProductUuids...)
	}
	allProductUuids = append(allProductUuids, productUuids...)
	if len(req.ConfirmReturnProductUuids) > 0 {
		allProductUuids = append(allProductUuids, req.ConfirmReturnProductUuids...)
	}
	if len(allProductUuids) == 0 {
		return errors.New("送厨商品ID不能为空")
	}
	productionRepo := repository.NewProductionRepo(db)
	allProducts, _ := productionRepo.GetProductsByUuids(allProductUuids, productionRepo.WithSaleOrderProductAll())
	if len(allProducts) != len(allProductUuids) {
		return errors.New("订单商品不存在")
	}

	// 要完成的菜品
	products := make([]model.ProductionOrderProduct, 0)
	// 要完成的菜品关联的销售订单商品ID列表
	var saleOrderProductUuids []uint64
	for _, product := range allProducts {
		if slices.Contains(productUuids, product.Uuid) {
			if product.Status != constant.ProductionOrderProductStatusCooking {
				return errors.New("订单商品未送厨")
			}
			if !(product.Num > 0) {
				return errors.New("订单商品数量为0")
			}
			if !slices.Contains(saleOrderProductUuids, product.SaleOrderProductUuid) {
				saleOrderProductUuids = append(saleOrderProductUuids, product.SaleOrderProductUuid)
			}
			products = append(products, product)
		}
		if slices.Contains(req.ConfirmReturnProductUuids, product.Uuid) {
			if product.Num > 0 {
				return errors.New("订单商品数量大于0")
			}
		}
	}

	confirmReturn := func(tx *gorm.DB) error {
		if len(req.ConfirmReturnProductUuids) > 0 {
			productionRepo := repository.NewProductionRepo(tx)
			if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereProductUuidIn(req.ConfirmReturnProductUuids)}, map[string]any{
				"delete_time": time.Now().Unix(),
			}); err != nil {
				return errors.ErrInternal
			}
		}
		return nil
	}

	// 只有确认退菜的商品，直接确认退菜
	if len(products) == 0 {
		if err := confirmReturn(db); err != nil {
			return errors.WithMessage(errors.New("更新送厨单商品状态失败"), err.Error())
		}
		// 恢复制作后，推送更新厨显
		utils.Go(func() {
			websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]any{
				"update_time": time.Now().Unix(),
			})
		})
		return nil
	}

	mode, err := s.getMode(ctx, req.Mode)
	if err != nil {
		return err
	}

	// 制作模式页面，双击完成制作
	if mode != nil {
		switch *mode {
		case constant.KdsModeMake: // 制作模式
			for _, product := range products {
				if product.MakeStatus == constant.ProductionOrderProductMakeStatusFinished {
					return errors.New("订单商品已制作完成")
				}
			}
			nowTime := time.Now().Unix()
			err := db.Transaction(func(tx *gorm.DB) error {
				if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereProductUuidIn(productUuids)}, map[string]any{
					"make_status":   constant.ProductionOrderProductMakeStatusFinished,
					"made_time":     nowTime,
					"make_duration": gorm.Expr("CASE WHEN is_batch = 1 THEN ? - batch_time ELSE ? - create_time END", nowTime, nowTime),
				}); err != nil {
					return errors.WithMessage(errors.New("更新送厨单商品状态失败"), err.Error())
				}
				if err := confirmReturn(tx); err != nil {
					return errors.WithMessage(errors.New("更新送厨单商品状态失败"), err.Error())
				}
				return nil
			})
			if err != nil {
				return err
			}

			// 完成制作后，推送更新厨显
			utils.Go(func() {
				websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]any{
					"update_time": time.Now().Unix(),
				})
			})

			return nil
		case constant.KdsModeDefault: // 传菜模式
			for _, product := range products {
				if product.MakeStatus != constant.ProductionOrderProductMakeStatusFinished {
					return errors.New("订单商品未制作完成")
				}
			}
		default:
			return errors.New("不支持的工作模式")
		}
	}

	finishedTime := time.Now().Unix()

	err = db.Transaction(func(tx *gorm.DB) error {
		productionRepo := repository.NewProductionRepo(tx)
		if mode != nil {
			// 如果开启智能厨显，则计算传菜时长. 当前时间-制作完成时间
			if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereProductUuidIn(productUuids)}, map[string]any{
				"status":        constant.ProductionOrderProductStatusFinished,
				"finished_time": finishedTime,
				"send_duration": gorm.Expr(`CASE 
					WHEN made_time = 0 THEN 
						CASE 
							WHEN is_batch = 1 THEN ? - batch_time 
							ELSE ? - create_time 
						END
					ELSE ? - made_time 
				END`, finishedTime, finishedTime, finishedTime),
				"all_duration": gorm.Expr(`make_duration + CASE 
					WHEN made_time = 0 THEN 
						CASE 
							WHEN is_batch = 1 THEN ? - batch_time 
							ELSE ? - create_time 
						END
					ELSE ? - made_time 
				END`, finishedTime, finishedTime, finishedTime),
			}); err != nil {
				return errors.WithMessage(errors.New("更新送厨单商品状态失败"), err.Error())
			}
		} else {
			// 如果没有开启智能厨显，则计算制作时长
			if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereProductUuidIn(productUuids)}, map[string]any{
				"status":        constant.ProductionOrderProductStatusFinished,
				"finished_time": finishedTime,
				"make_duration": gorm.Expr("CASE WHEN is_batch = 1 THEN ? - batch_time ELSE ? - create_time END", finishedTime, finishedTime),
				"all_duration":  gorm.Expr("CASE WHEN is_batch = 1 THEN ? - batch_time ELSE ? - create_time END", finishedTime, finishedTime),
			}); err != nil {
				return errors.WithMessage(errors.New("更新送厨单商品状态失败"), err.Error())
			}
		}

		if len(saleOrderProductUuids) > 0 {
			if err := s.updatePackageProduct(tx, saleOrderProductUuids, finishedTime); err != nil {
				return errors.WithMessage(errors.New("更新送厨单套餐商品状态失败"), err.Error())
			}
		}

		if err := confirmReturn(tx); err != nil {
			return errors.WithMessage(errors.New("更新送厨单商品状态失败"), err.Error())
		}
		return nil
	})
	if err != nil {
		return err
	}

	// 完成制作事件
	utils.Go(func() {
		// 把 products  按 SaleBillUuid 分组
		productsBySaleBillUuid := make(map[uint64][]model.ProductionOrderProduct, 0)
		for _, product := range products {
			productsBySaleBillUuid[product.SaleBillUuid] = append(productsBySaleBillUuid[product.SaleBillUuid], product)
		}
		// 获取套餐商品
		if len(req.ProductUuids) > 0 {
			packageSubProductUuidMap := make(map[uint64]bool)
			for _, product := range products {
				if product.SaleOrderProduct.IsPackageSubProduct() {
					packageSubProductUuidMap[product.SaleOrderProduct.PackageUuid] = true
				}
			}
			packageSubProductUuids := make([]uint64, 0, len(packageSubProductUuidMap))
			for uuid := range packageSubProductUuidMap {
				packageSubProductUuids = append(packageSubProductUuids, uuid)
			}
			if len(packageSubProductUuids) > 0 {
				productionRepo := repository.NewSaleOrderProductRepo(db)
				saleOrderProducts, err := productionRepo.GetSaleOrderProductsByUuidIn(packageSubProductUuids, productionRepo.WithSaleOrderProductAll())
				if err != nil {
					logger.Logger.Error("获取套餐子商品失败", zap.Error(err))
				} else {
					for _, saleOrderProduct := range saleOrderProducts {
						productsBySaleBillUuid[saleOrderProduct.SaleBillUuid] = append(productsBySaleBillUuid[saleOrderProduct.SaleBillUuid], model.ProductionOrderProduct{
							BaseModel: model.BaseModel{
								Uuid: saleOrderProduct.Uuid,
							},
							ProductPackageUuid: saleOrderProduct.PackageUuid,
							SaleOrderProduct:   *saleOrderProduct,
							SaleBill:           *saleOrderProduct.SaleBill,
						})
					}
				}
			}
		}
		for saleBillUuid, products := range productsBySaleBillUuid {
			event.NewSystemBus().PublishFinishMenuEvent(event.FinishMenuPayload{
				BasePayload: event.BasePayload{
					Ctx:           ctx,
					CompanyUuid:   ctx.GetCompanyUuid(),
					Source:        ctx.GetSource(),
					SaleBillUuid:  saleBillUuid,
					SaleOrderUuid: saleBillUuid,
				},
				FinishedTime: finishedTime,
				Products: func() event.Products {
					orderProducts := make(event.Products, 0)
					for _, product := range products {
						if product.SaleOrderProduct.ID == 0 {
							continue
						}
						// 套餐子商品不显示送厨记录
						if len(req.ProductUuids) > 0 && product.SaleOrderProduct.IsPackageSubProduct() {
							continue
						}
						orderProducts = append(orderProducts, s.convertToEventOrderProduct(&product, products))
					}
					return orderProducts
				}(),
			})
		}
	})

	// 完成制作后，推送更新厨显
	utils.Go(func() {
		websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]any{
			"update_time": time.Now().Unix(),
		})
	})

	// 完成制作后，推送更新订单
	utils.Go(func() {
		pushedMap := make(map[string]bool)
		for _, product := range products {
			key := fmt.Sprintf("%d-%d", product.SaleBillUuid, product.SaleBill.DeskUuid)
			if _, ok := pushedMap[key]; !ok {
				pushedMap[key] = true
				websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_ORDER, map[string]any{
					"update_time":    time.Now().Unix(),
					"sale_bill_uuid": product.SaleBillUuid,
					"desk_uuid":      product.SaleBill.DeskUuid,
				})
			}
		}
	})
	return nil

}

// 转换器
func (s *productionSrv) convertToEventOrderProduct(product *model.ProductionOrderProduct, products []model.ProductionOrderProduct) event.OrderProduct {
	orderProduct := event.OrderProduct{
		OrderProductId:        product.Uuid,
		ProductId:             product.ProductPackageUuid,
		ProductName:           product.SaleOrderProduct.MultiLanguageName.GetNames(),
		ProductType:           product.SaleOrderProduct.ProductType,
		ProductAttr:           product.SaleOrderProduct.GetAttributeName(),
		ProductAttrList:       product.SaleOrderProduct.GetAttributeNameList(),
		ProductSauceNamesList: product.SaleOrderProduct.GetSauceNamesList(),
		TotalNum:              product.SaleOrderProduct.Num,
		NumType:               product.SaleOrderProduct.NumType,
		IsBuffet:              product.SaleOrderProduct.IsBuffet == 1,
		IsWrap: func() bool {
			if product.SaleBill.IsTakeout() && product.SaleBill.MemberSaleOrderUuid == 0 {
				return true
			}
			return product.SaleOrderProduct.IsWrapProduct()
		}(),
		Remark: product.SaleOrderProduct.Remark,
	}
	// 如果是套餐主商品，添加子商品
	if product.SaleOrderProduct.IsPackageProduct() && products != nil && len(products) > 0 {
		subProducts := make([]event.OrderProduct, 0)
		for _, subProduct := range products {
			if subProduct.SaleOrderProduct.IsDelete() {
				continue
			}
			if subProduct.SaleOrderProduct.PackageUuid == product.SaleOrderProduct.Uuid {
				subProducts = append(subProducts, s.convertToEventOrderProduct(&subProduct, nil))
			}
		}
		orderProduct.SubProducts = subProducts
	}
	return orderProduct
}

// Recovery 恢复制作
func (s *productionSrv) Recovery(ctx context.Context, req req.RecoveryReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	productionRepo := repository.NewProductionRepo(db)
	product, _ := productionRepo.GetProduct(productionRepo.WhereProductUuid(req.ProductUuid))
	if product.Uuid == 0 {
		return errors.New("订单商品不存在")
	}

	mode, err := s.getMode(ctx, req.Mode)
	if err != nil {
		return err
	}

	if mode != nil && *mode == constant.KdsModeMake {
		// 是否已制作完成
		if product.MakeStatus != constant.ProductionOrderProductMakeStatusFinished {
			return errors.New("订单商品未制作完成")
		}
		// 是否已经传菜
		if product.Status == constant.ProductionOrderProductStatusFinished {
			return errors.New("该菜品已传菜，不可恢复！")
		}
		if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereProductUuid(req.ProductUuid)}, map[string]any{
			"make_status":   constant.ProductionOrderProductMakeStatusDefault,
			"made_time":     0,
			"make_duration": 0, // 清空时长记录
			"send_duration": 0,
			"all_duration":  0,
		}); err != nil {
			return errors.WithMessage(errors.New("恢复送厨单商品制作状态失败"), err.Error())
		}
		// 恢复制作后，推送更新厨显
		utils.Go(func() {
			websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]any{
				"update_time": time.Now().Unix(),
			})
		})
		return nil
	}

	if product.Status != constant.ProductionOrderProductStatusFinished {
		return errors.New("订单商品未完成")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		productionRepo := repository.NewProductionRepo(tx)
		if mode != nil {
			// 如果是智能厨显时,取消传菜则清空sent_time、all_time
			if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereProductUuid(req.ProductUuid)}, map[string]any{
				"status":        constant.ProductionOrderProductStatusCooking,
				"finished_time": 0,
				"send_duration": 0,
				"all_duration":  0,
			}); err != nil {
				return errors.WithMessage(errors.New("更新送厨单商品状态失败"), err.Error())
			}
		} else {
			// 如果不是智能厨显时,取消制作完成则清空make_time、all_time
			if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereProductUuid(req.ProductUuid)}, map[string]any{
				"status":        constant.ProductionOrderProductStatusCooking,
				"finished_time": 0,
				"make_duration": 0,
				"all_duration":  0,
			}); err != nil {
				return errors.WithMessage(errors.New("更新送厨单商品状态失败"), err.Error())
			}
		}
		if err := s.updatePackageProduct(tx, []uint64{product.SaleOrderProductUuid}, 0); err != nil {
			return errors.WithMessage(errors.New("更新送厨单套餐商品状态失败"), err.Error())
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(errors.New("更新送厨单商品状态失败"), err.Error())
	}

	// 恢复制作后，推送更新厨显
	utils.Go(func() {
		websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]any{
			"update_time": time.Now().Unix(),
		})
	})
	// 恢复制作后，推送更新订单
	utils.Go(func() {
		websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_ORDER, map[string]any{
			"update_time":    time.Now().Unix(),
			"sale_bill_uuid": product.SaleBillUuid,
			"desk_uuid":      product.SaleBill.DeskUuid,
		})
	})
	return nil
}

// ConfirmReturn 厨显端确认退菜
func (s *productionSrv) ConfirmReturn(ctx context.Context, productUuid uint64) error {
	productionRepo := repository.NewProductionRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
	product, _ := productionRepo.GetProduct(productionRepo.WhereProductUuid(productUuid))
	if product.Uuid == 0 {
		return errors.New("订单商品不存在")
	}
	if product.Num > 0 {
		return errors.New("订单商品数量大于0")
	}
	if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereProductUuid(productUuid)}, map[string]any{
		"delete_time": time.Now().Unix(),
	}); err != nil {
		return errors.ErrInternal
	}

	// 恢复制作后，推送更新厨显
	utils.Go(func() {
		websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]any{
			"update_time": time.Now().Unix(),
		})
	})

	return nil
}

// ConfirmReturnAll 厨显端确认退菜整单
func (s *productionSrv) ConfirmReturnAll(ctx context.Context, saleBillUuid uint64) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	saleBillRepo := repository.NewSaleBillRepo(db)
	saleBill, err := saleBillRepo.GetSaleBillByUuid(saleBillUuid)
	if err != nil {
		return errors.ErrInternal
	}

	if saleBill.DeleteTime == 0 && saleBill.Status != constant.SaleBillStatusCanceled || saleBill.IsKitchenConfirm != 0 {
		return errors.New("订单未整单取消")
	}
	saleBill.IsKitchenConfirm = 1

	// 销售订单更新为厨显端已确认退整单，并更新送厨商品为已删除
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := repository.NewSaleBillRepo(tx).UpdateSaleBill(saleBill); err != nil {
			return err
		}
		productionRepo := repository.NewProductionRepo(tx)
		if err := productionRepo.UpdateProduct([]repository.DBOption{productionRepo.WhereSaleBillUuid(saleBillUuid)}, map[string]any{
			"delete_time": time.Now().Unix(),
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.ErrInternal
	}

	// 恢复制作后，推送更新厨显
	utils.Go(func() {
		websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceKitchen, websocket.SourceAll, websocket.UPDATE_KITCHEN, map[string]any{
			"update_time": time.Now().Unix(),
		})
	})
	return nil
}
