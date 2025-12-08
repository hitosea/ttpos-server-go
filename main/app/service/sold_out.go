package service

import (
	"time"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"
)

// ISoldOutSrv 定义沽清服务接口
type ISoldOutSrv interface {
	GetSoldOutList(companyUuid uint64, soldOutReq req.SoldOutListReq) (resp.SoldOutPaginationResp, error) // 获取沽清商品列表
	CancelSoldOut(companyUuid uint64, productBomUuid uint64) error                                        // 取消单个沽清商品
	CancelAllSoldOut(companyUuid uint64) error                                                            // 取消全部沽清商品
	AddSoldOut(companyUuid uint64, items []req.SoldOutItem) error                                         // 添加商品沽清
	GetSettings(companyUuid uint64, req *req.GetSoldOutSettingsReq) (*resp.SoldOutSettingsResp, error)    // 获取商品沽清设置
}

// soldOutSrv 沽清服务结构体
type soldOutSrv struct {
	dbm       *database.DBManager // 数据库管理
	localeSrv ILocaleSrv          // 多语言名称服务
}

// NewSoldOutSrv 创建新的收银产品类别服务
func NewSoldOutSrv(dbm *database.DBManager, localSrv ILocaleSrv) ISoldOutSrv {
	return NewSoldOutSrvImpl(dbm, localSrv)
}

// NewSoldOutSrvImpl 创建新的收银服务实现
func NewSoldOutSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv) ISoldOutSrv {
	return &soldOutSrv{
		dbm:       dbm,
		localeSrv: localeSrv,
	}
}

// GetSoldOutList 获取沽清商品列表
func (s *soldOutSrv) GetSoldOutList(companyUuid uint64, soldOutReq req.SoldOutListReq) (resp.SoldOutPaginationResp, error) {

	productRepo := repository.NewProductRepo(s.dbm.GetDB(companyUuid))

	boms, total, err := productRepo.GetSoldOutWithPagination(soldOutReq.PageNo, soldOutReq.PageSize,
		productRepo.WithProductPackage(),
		productRepo.WithProductPackageMultiLanguageName(),
		productRepo.WithProductFlavor(),
		productRepo.WithProductFlavorMultiLanguageName())

	if err != nil {
		return resp.SoldOutPaginationResp{}, errors.WithMessage(err, "获取沽清商品列表失败")
	}

	soldOuts := make([]resp.SoldOut, 0, len(boms))

	for _, bom := range boms {
		soldOuts = append(soldOuts, resp.SoldOut{
			LocaleProductName:    bom.ProductPackage.MultiLanguageName.GetNames(),
			ProductBomUuid:       bom.Uuid,
			LocaleProductBomName: bom.ProductFlavor.MultiLanguageName.GetNames(),
		})
	}

	return resp.SoldOutPaginationResp{
		List: soldOuts,
		Meta: dto.PageResponse{
			PageNo:   soldOutReq.PageNo,
			PageSize: soldOutReq.PageSize,
			Total:    total,
		},
	}, nil
}

// CancelSoldOut 取消单个沽清商品
func (s *soldOutSrv) CancelSoldOut(companyUuid uint64, productBomUuid uint64) error {
	productRepo := repository.NewProductRepo(s.dbm.GetDB(companyUuid))
	if err := productRepo.UpdateProductBomSoldOut([]repository.DBOption{productRepo.WhereBomUuid(productBomUuid)}, map[string]any{
		"is_sold_out": 0,
	}); err != nil {
		return errors.WithMessage(err, "取消沽清商品失败")
	}
	// 推送沽清商品
	utils.Go(func() {
		websocket.PushClient(companyUuid, websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_PRODUCT, map[string]interface{}{
			"type":         "update",
			"product_uuid": productBomUuid,
			"update_time":  time.Now().Unix(),
		})
	})
	return nil
}

// CancelAllSoldOut 取消全部沽清商品
func (s *soldOutSrv) CancelAllSoldOut(companyUuid uint64) error {
	productRepo := repository.NewProductRepo(s.dbm.GetDB(companyUuid))

	if err := productRepo.UpdateProductBomSoldOut([]repository.DBOption{productRepo.WhereBomIsSoldOut()}, map[string]any{
		"is_sold_out": 0,
	}); err != nil {
		return errors.WithMessage(err, "全部取消沽清商品失败")
	}

	// 推送沽清商品
	utils.Go(func() {
		websocket.PushClient(companyUuid, websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_PRODUCT, map[string]interface{}{
			"type":         "update",
			"product_uuid": 0,
			"update_time":  time.Now().Unix(),
		})
	})
	return nil
}

// AddSoldOut 添加商品沽清
func (s *soldOutSrv) AddSoldOut(companyUuid uint64, items []req.SoldOutItem) error {
	productRepo := repository.NewProductRepo(s.dbm.GetDB(companyUuid))
	for _, item := range items {
		soldOutMap := map[bool]uint{
			true:  1,
			false: 0,
		}
		if err := productRepo.UpdateProductBomSoldOut([]repository.DBOption{productRepo.WhereBomUuid(item.ProductBomUuid)}, map[string]any{
			"is_sold_out": soldOutMap[*item.IsSoldOut],
		}); err != nil {
			return errors.WithMessage(err, "沽清商品失败")
		}
		// 推送沽清商品
		utils.Go(func() {
			websocket.PushClient(companyUuid, websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_PRODUCT, map[string]interface{}{
				"type":         "update",
				"product_uuid": item.ProductBomUuid,
				"update_time":  time.Now().Unix(),
			})
		})
	}
	return nil
}

// GetSettings 获取商品沽清设置
func (s *soldOutSrv) GetSettings(companyUuid uint64, req *req.GetSoldOutSettingsReq) (*resp.SoldOutSettingsResp, error) {
	productBomRepo := repository.NewProductBomRepo(s.dbm.GetDB(companyUuid))

	// 查询该商品包的所有规格
	boms, err := productBomRepo.GetProductBoms(
		repository.CommonRepo.WhereByProductPackageUuid(req.ProductPackageUuid),
		repository.CommonRepo.WhereBySoftDelete(),
	)
	if err != nil {
		return nil, errors.WithMessage(err, "获取商品规格失败")
	}

	settings := make([]resp.SoldOutSetting, 0, len(boms))
	for _, bom := range boms {
		bomCardStockNum := 0.0
		if bom.UseBomCardStock == 1 {
			// TODO: 调用库存服务计算成本卡库存
			// 暂时返回 0，后续实现成本卡库存计算
		}

		settings = append(settings, resp.SoldOutSetting{
			ProductBomUuid:   bom.Uuid,
			UseBomCardStock:  bom.UseBomCardStock == 1,
			BomCardStockNum:  bomCardStockNum,
			IsSoldOut:        bom.IsSoldOut == 1,
			SellableQuantity: bom.SellableQuantity,
		})
	}

	return &resp.SoldOutSettingsResp{
		Settings: settings,
	}, nil
}
