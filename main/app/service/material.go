package service

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"ttpos-bmp/app/ttpos-erp/api/manufacturing"
	v1 "ttpos-bmp/app/ttpos-message/api/message"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp/material_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/repository/base"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/app/service/rpc/message"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/language"
	"ttpos-server-go/pkg/lock"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/jinzhu/copier"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 物品导入状态常量
const (
	MaterialImportStatusStart      = "start"      // 开始导入
	MaterialImportStatusProcessing = "processing" // 正在处理
	MaterialImportStatusFinish     = "finish"     // 导入完成
	MaterialImportStatusError      = "error"      // 导入失败
)

// MaterialImportProgressData 物品导入进度数据结构
type MaterialImportProgressData struct {
	Time     int64                       `json:"time"`     // 时间戳
	Status   string                      `json:"status"`   // 状态：start|processing|finish|error
	Progress int                         `json:"progress"` // 进度百分比 (0-100)
	Total    int                         `json:"total"`    // 总数量
	Current  int                         `json:"current"`  // 当前处理数量
	Success  int                         `json:"success"`  // 成功数量
	Failed   int                         `json:"failed"`   // 失败数量
	Error    string                      `json:"error"`    // 总体错误信息
	Errors   []MaterialImportErrorDetail `json:"errors"`   // 详细错误列表
}

// MaterialImportErrorDetail 物品导入错误详情
type MaterialImportErrorDetail struct {
	Row     int    `json:"row"`     // 行号
	Message string `json:"message"` // 错误信息
}

// IMaterialSrv 物品服务接口
type IMaterialSrv interface {
	GetMaterialList(ctx context.Context, req req.MaterialListReq) (material_resp.MaterialListWithPaginationResp, error)
	GetMaterialDetail(ctx context.Context, req req.MaterialDetailReq) (material_resp.MaterialDetailResp, error)
	GetMaterialStockDetail(ctx context.Context, req req.MaterialStockDetailReq) (material_resp.MaterialStockDetailResp, error)
	AddMaterial(ctx context.Context, req req.MaterialAddReq) error
	AddMaterialByEprItem(ctx context.Context, request req.MaterialAddErpReq) (*model.Material, error)
	EditMaterial(ctx context.Context, req req.MaterialEditReq) error
	UpdateMaterialByEprItem(ctx context.Context, request req.MaterialEditErpReq) error
	DeleteMaterial(ctx context.Context, req req.MaterialDeleteReq) error
	UpdateMaterialStatusBatch(ctx context.Context, req req.MaterialStatusReq) error
	UpdateMaterialVisibleBatch(ctx context.Context, req req.MaterialBatchUpdateVisibleReq) (int, error) // 批量更新物品可见性，返回更新数量
	AddMaterialCategory(ctx context.Context, req req.MaterialCategoryAddReq) error
	GetMaterialCategoryList(ctx context.Context, req req.MaterialCategoryListReq) (material_resp.MaterialCategoryListResp, error)
	GetMaterialCategoryDetail(ctx context.Context, req req.MaterialCategoryDetailReq) (*material_resp.MaterialCategory, error)
	SortMaterialCategory(ctx context.Context, req req.MaterialCategorySortReq) error
	EditMaterialCategory(ctx context.Context, req req.MaterialCategoryEditReq) error
	GetMaterialUnitList(ctx context.Context, req req.MaterialUnitListReq) (material_resp.MaterialUnitListResp, error)
	DeleteMaterialCategory(ctx context.Context, req req.MaterialCategoryDeleteReq) error
	AddProductBomCard(ctx context.Context, req req.ProductBomCardAddReq) error
	GetProductBomCardDetail(ctx context.Context, req req.ProductBomCardDetailReq) (*material_resp.ProductBomCardDetailResp, error)
	UnlinkProductBomCard(ctx context.Context, req req.ProductBomCardUnlinkReq) error
	CopyProductBomCard(ctx context.Context, req req.ProductBomCardCopyReq) error
	ImportProductBomCard(ctx context.Context, req req.ProductBomCardImportReq) error
	ImportMaterialList(ctx context.Context, req req.MaterialImportListReq) (material_resp.MaterialImportResp, error)
	ImportMaterial(ctx context.Context, req req.MaterialImportReq) error
	GetWarehouseItemsByErpCode(ctx context.Context, warehouseErpCode string, pageNo, pageSize int) ([]model.WarehouseItem, int64, error)
	SyncMaterialCategory(ctx context.Context) error // 同步物品分类
	SyncMaterial(ctx context.Context) error         // 同步物品
	SyncProductBomCard(ctx context.Context) error   // 同步成本卡

	GetWarehouseItemConsumption(ctx context.Context, warehouseUuid uint64) (material_resp.MaterialConsumptionListResp, error) // 获取仓库物品消耗量

	CheckMaterialSafetyStock(ctx context.Context, companyUuid uint64) error // 检查物品安全库存
}

type materialSrv struct {
	dbm          *database.DBManager // 数据库管理器
	translateSrv ITranslateSrv       // 翻译服务
	localeSrv    ILocaleSrv          // 多语言名称服务
	settingSrv   setting.ISrv        // 设置服务
	messageSrv   message.IMessageSrv // 消息服务
	systemLock   lock.Lock           // 系统锁
}

func NewMaterialSrv(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv, translateSrv ITranslateSrv, messageSrv message.IMessageSrv) IMaterialSrv {
	return NewMaterialSrvImpl(dbm, localeSrv, settingSrv, translateSrv, messageSrv)
}

func NewMaterialSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv, translateSrv ITranslateSrv, messageSrv message.IMessageSrv) IMaterialSrv {
	return &materialSrv{
		dbm:          dbm,
		translateSrv: translateSrv,
		localeSrv:    localeSrv,
		settingSrv:   settingSrv,
		messageSrv:   messageSrv,
		systemLock:   lock.NewSystemLock(),
	}
}

// pushMaterialImportProgress 推送物品导入进度到前端
func (s *materialSrv) pushMaterialImportProgress(companyUuid uint64, deviceSn string, data MaterialImportProgressData) {
	data.Time = time.Now().Unix()
	utils.Go(func() {
		websocket.PushClient(companyUuid, websocket.SourceShop, deviceSn, websocket.IMPORT_MATERIAL, map[string]any{
			"time":     data.Time,
			"status":   data.Status,
			"progress": data.Progress,
			"total":    data.Total,
			"current":  data.Current,
			"success":  data.Success,
			"failed":   data.Failed,
			"error":    data.Error,
			"errors":   data.Errors,
		})
	})
}

// GetMaterialList 获取物品列表
func (s *materialSrv) GetMaterialList(ctx context.Context, req req.MaterialListReq) (material_resp.MaterialListWithPaginationResp, error) {
	dbId := ctx.GetDbId()
	companySetting := ctx.GetCompanySetting()
	materialRepo := repository.NewMaterialRepo(s.dbm.GetDB(dbId))

	// 构建查询选项
	var dbOptions []repository.DBOption
	commonRepo := repository.NewCommonRepo()

	// 根据名称、编码、条码模糊查询
	if req.Keyword != "" {
		dbOptions = append(dbOptions, commonRepo.DBOption(func(db *gorm.DB) *gorm.DB {
			return db.Where("name LIKE ? OR code LIKE ? OR barcode_value LIKE ? OR internal_code LIKE ?",
				"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
		}))
	}

	// WarehouseErpCode 根据仓库ERP编码过滤
	if req.PurchaseType == 2 {
		dbOptions = append(dbOptions, commonRepo.DBOption(func(db *gorm.DB) *gorm.DB {
			return db.Where("headquarter_uuid > ?", 0)
		}))
	} else if req.SupplierErpCode != "" {
		dbOptions = append(dbOptions, commonRepo.WhereByMaterialSupplierErpCode(req.SupplierErpCode))
	}

	if len(req.CategoryUuids) > 0 {
		dbOptions = append(dbOptions, commonRepo.WhereByCategoryUuids(req.CategoryUuids))
	}
	if req.Status != 0 {
		dbOptions = append(dbOptions, commonRepo.WhereByStatus(uint(req.Status)))
	}
	if len(req.MaterialUuids) > 0 {
		dbOptions = append(dbOptions, commonRepo.WhereInUuids(req.MaterialUuids))
	}

	// 子店查询时自动过滤不可见物品
	if companySetting.IsSubShop() {
		dbOptions = append(dbOptions, materialRepo.WhereAllowSubstoreVisible(1))
	}

	// 预加载关联数据
	dbOptions = append(dbOptions, commonRepo.Preload(
		repository.WithPreload{
			Query: "MultiLanguageName",
		},
		repository.WithPreload{
			Query: "Unit.Unit.MultiLanguageName",
		},
		repository.WithPreload{
			Query: "PurchaseUnit.Unit.MultiLanguageName",
		},
		repository.WithPreload{
			Query: "CostUnit.Unit.MultiLanguageName",
		},
		repository.WithPreload{
			Query: "NotBaseUnitList.Unit.MultiLanguageName",
		},
		repository.WithPreload{
			Query: "WarehouseItems.Warehouse",
		},
	))

	// 获取物品列表
	materials, total, err := materialRepo.GetMaterialListWithPagination(
		req.PageNo,
		req.PageSize,
		dbOptions...,
	)

	if err != nil {
		return material_resp.MaterialListWithPaginationResp{}, errors.WithMessage(err, "获取物品列表失败")
	}

	// 转换为响应格式
	var materialList []material_resp.Material
	for _, material := range materials {
		unitList := []material_resp.MaterialUnit{}
		for _, unit := range material.NotBaseUnitList {
			unitList = append(unitList, material_resp.MaterialUnit{
				Uuid: unit.Uuid,
				Name: func() string {
					if unit.Unit == nil {
						return ""
					}
					if unit.Unit.MultiLanguageName == (model.MultiLanguageName{}) {
						return ""
					}
					return unit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
				}(),
				LocaleName: func() dto.LocaleResponse {
					if unit.Unit == nil {
						return dto.LocaleResponse{}
					}
					if unit.Unit.MultiLanguageName == (model.MultiLanguageName{}) {
						return dto.LocaleResponse{}
					}
					return unit.Unit.MultiLanguageName.GetNames()
				}(),
				ConversionRate: unit.ConversionRate,
			})
		}
		purchaseUnit := material.GetUnit(material.PurchaseUnitUuid)
		costUnit := material.GetUnit(material.CostUnitUuid)
		baseUnit := material.GetBaseUnit()
		var purchaseUnitLocaleName, costUnitLocaleName, baseUnitLocaleName dto.LocaleResponse
		if costUnit != nil {
			costUnitLocaleName = *language.JsonToLocaleResponse(costUnit.Name)
		}
		if purchaseUnit != nil {
			purchaseUnitLocaleName = *language.JsonToLocaleResponse(purchaseUnit.Name)
		}
		if baseUnit != nil {
			baseUnitLocaleName = *language.JsonToLocaleResponse(baseUnit.Name)
		}

		// 库存数量、可用库存数量、在途库存数量
		num := decimal.NewFromFloat(0)
		availableNum := decimal.NewFromFloat(0)
		transitNum := decimal.NewFromFloat(0)
		for _, warehouseItem := range material.WarehouseItems {
			if warehouseItem.Warehouse != nil {
				if req.OutWarehouseErpCode != "" && warehouseItem.Warehouse.ErpCode != req.OutWarehouseErpCode {
					continue
				}
				if warehouseItem.Warehouse.IsTransit() {
					transitNum = transitNum.Add(decimal.NewFromFloat(warehouseItem.Stock))
				} else {
					availableNum = availableNum.Add(decimal.NewFromFloat(warehouseItem.Stock))
				}
			}
		}
		stockNum := num.Add(availableNum).Add(transitNum).InexactFloat64()

		// 非基准单位的库存数
		notBasicUnitStocks := make([]material_resp.NotBasicUnitStock, 0)
		for _, unit := range material.NotBaseUnitList {
			if unit.IsDelete() || unit.IsDefault == 1 {
				continue
			}
			notBasicUnitStocks = append(notBasicUnitStocks, material_resp.NotBasicUnitStock{
				LocaleName: func() dto.LocaleResponse {
					if unit.Unit == nil { // 没有关联单位，则使用名称
						name := model.NewMultiLanguageName(unit.Name)
						return name.GetNames()
					}
					return unit.Unit.MultiLanguageName.GetNames()
				}(),
				Num: func() float64 {
					if unit.ConversionRate == 0 {
						return 0
					}
					return decimal.NewFromFloat(stockNum).Div(decimal.NewFromFloat(unit.ConversionRate)).Round(3).InexactFloat64()
				}(),
				ConversionRate: unit.ConversionRate,
			})
		}

		// 响应格式
		respMaterial := material_resp.Material{
			Uuid:               material.Uuid,
			MaterialUuid:       material.Uuid,
			Name:               material.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
			LocaleName:         material.MultiLanguageName.GetNames(),
			ErpCode:            material.Code,
			InternalCode:       material.InternalCode,
			BarcodeValue:       material.BarcodeValue,
			Num:                stockNum,
			SafetyStock:        material.SafetyStock,
			AvailableNum:       availableNum.InexactFloat64(),
			TransitNum:         transitNum.InexactFloat64(),
			NotBasicUnitStocks: material_resp.NotBasicUnitStockList{List: notBasicUnitStocks},
			CategoryUuid:       material.CategoryUuid,
			Image:              material.GetImage(utils.GetBaseURL(ctx.GetGin().Request)),
			Status: func() int {
				if material.Status {
					return 1
				}
				return 0
			}(),
			UnitName: func() string {
				if material.Unit != nil && material.Unit.Unit != nil && material.Unit.Unit.MultiLanguageName != (model.MultiLanguageName{}) {
					return material.Unit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
				}
				return ""
			}(),
			UnitUuid: material.UnitUuid,
			PurchaseUnitName: func() string {
				if material.PurchaseUnit == nil {
					return ""
				}
				if material.PurchaseUnit.Unit == nil {
					return ""
				}
				if material.PurchaseUnit.Unit.MultiLanguageName == (model.MultiLanguageName{}) {
					return ""
				}
				return material.PurchaseUnit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
			}(),
			PurchaseUnitUuid: material.PurchaseUnitUuid,
			CostUnitName: func() string {
				if material.CostUnit == nil {
					return ""
				}
				if material.CostUnit.Unit == nil {
					return ""
				}
				if material.CostUnit.Unit.MultiLanguageName == (model.MultiLanguageName{}) {
					return ""
				}
				return material.CostUnit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
			}(),
			CostUnitUuid:           material.CostUnitUuid,
			PurchaseUnitLocaleName: purchaseUnitLocaleName,
			CostUnitLocaleName:     costUnitLocaleName,
			UnitLocaleName:         baseUnitLocaleName,
			UnitList:               unitList,
		}
		materialList = append(materialList, respMaterial)
	}

	return material_resp.MaterialListWithPaginationResp{
		List: materialList,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// GetMaterialDetail 获取物品详情
func (s *materialSrv) GetMaterialDetail(ctx context.Context, req req.MaterialDetailReq) (material_resp.MaterialDetailResp, error) {
	dbId := ctx.GetDbId()
	materialRepo := repository.NewMaterialRepo(s.dbm.GetDB(dbId))

	// 获取物品详情
	material, err := materialRepo.GetMaterialDetailByUuid(req.Uuid)
	if err != nil {
		return material_resp.MaterialDetailResp{}, errors.WithMessage(err, "获取物品详情失败")
	}
	unitList := []material_resp.MaterialUnit{}
	materialUnitRepo := repository.NewMaterialUnitRepo(s.dbm.GetDB(dbId))
	materialUnitList, err := materialUnitRepo.GetMaterialUnitListByBaseUnitUuid(material.Unit.Uuid)
	if err != nil {
		return material_resp.MaterialDetailResp{}, errors.WithMessage(err, "获取物品详情失败")
	}
	for _, materialUnit := range materialUnitList {
		unitList = append(unitList, material_resp.MaterialUnit{
			Uuid:           materialUnit.Uuid,
			FromUnitUuid:   materialUnit.UnitUuid,
			Name:           materialUnit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
			LocaleName:     materialUnit.Unit.MultiLanguageName.GetNames(),
			ConversionRate: materialUnit.ConversionRate,
		})
	}

	purchaseUnit := material.GetUnit(material.PurchaseUnitUuid)
	costUnit := material.GetUnit(material.CostUnitUuid)
	baseUnit := material.GetBaseUnit()
	var purchaseUnitLocaleName, costUnitLocaleName, baseUnitLocaleName dto.LocaleResponse
	if purchaseUnit != nil {
		purchaseUnitLocaleName = *language.JsonToLocaleResponse(purchaseUnit.Name)
	}
	if costUnit != nil {
		costUnitLocaleName = *language.JsonToLocaleResponse(costUnit.Name)
	}
	if baseUnit != nil {
		baseUnitLocaleName = *language.JsonToLocaleResponse(baseUnit.Name)
	}

	fromUnitUuid := uint64(0)
	if material.Unit != nil {
		fromUnitUuid = material.GetUnit(material.UnitUuid).UnitUuid
	}
	fromPurchaseUnitUuid := uint64(0)
	if material.PurchaseUnit != nil {
		fromPurchaseUnitUuid = material.GetUnit(material.PurchaseUnitUuid).UnitUuid
	}
	fromCostUnitUuid := uint64(0)
	if material.CostUnit != nil {
		fromCostUnitUuid = material.GetUnit(material.CostUnitUuid).UnitUuid
	}
	return material_resp.MaterialDetailResp{
		Uuid:         material.Uuid,
		LocaleName:   material.MultiLanguageName.GetNames(),
		Code:         material.Code,
		CategoryUuid: material.CategoryUuid,
		CategoryName: material.Category.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
		Status:       int(utils.BoolToUint(material.Status)),
		Valuation:    material.Valuation,
		BarcodeValue: material.BarcodeValue,
		InternalCode: material.InternalCode,
		SafetyStock:  material.SafetyStock,
		UnitName:     material.Unit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
		UnitUuid:     material.UnitUuid,
		FromUnitUuid: fromUnitUuid,
		UnitList:     material_resp.MaterialUnitListResp{List: unitList},
		PurchaseUnitName: func() string {
			if material.PurchaseUnit == nil {
				return ""
			}
			return material.PurchaseUnit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
		}(),
		PurchaseUnitUuid:       material.PurchaseUnitUuid,
		FromPurchaseUnitUuid:   fromPurchaseUnitUuid,
		CostUnitName:           material.CostUnit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
		CostUnitUuid:           material.CostUnitUuid,
		FromCostUnitUuid:       fromCostUnitUuid,
		PurchaseUnitLocaleName: purchaseUnitLocaleName,
		CostUnitLocaleName:     costUnitLocaleName,
		UnitLocaleName:         baseUnitLocaleName,
		IsEditable:             !material.IsHeadquarter(), // 总部物品不可编辑
	}, nil
}

// GetMaterialStockDetail 获取物品库存详情
func (s *materialSrv) GetMaterialStockDetail(ctx context.Context, req req.MaterialStockDetailReq) (material_resp.MaterialStockDetailResp, error) {
	dbId := ctx.GetDbId()
	materialRepo := repository.NewMaterialRepo(s.dbm.GetDB(dbId))
	material, err := materialRepo.GetMaterialDetailByUuid(req.Uuid)
	if err != nil {
		return material_resp.MaterialStockDetailResp{}, errors.WithMessage(err, "获取物品库存详情失败")
	}
	// 获取仓库列表
	warehouseRepo := repository.NewWarehouseItemRepo(s.dbm.GetDB(dbId))
	warehouseItems, err := warehouseRepo.GetWarehouseItemsByMaterialUuid(material.Uuid)
	if err != nil {
		return material_resp.MaterialStockDetailResp{}, errors.WithMessage(err, "获取仓库库存列表失败")
	}

	warehouseList := []material_resp.Warehouse{}
	amount := decimal.NewFromFloat(0)
	for _, warehouseItem := range warehouseItems {
		localeName := dto.LocaleResponse{}
		if warehouseItem.Warehouse != nil {
			localeName = warehouseItem.Warehouse.MultiLanguageName.GetNames()
		}
		notBasicUnitStocks := make([]material_resp.NotBasicUnitStock, 0)
		for _, unit := range warehouseItem.Material.NotBaseUnitList {
			notBasicUnitStock := material_resp.NotBasicUnitStock{
				LocaleName: func() dto.LocaleResponse {
					if unit.Unit == nil { // 没有关联单位，则使用名称
						name := model.NewMultiLanguageName(unit.Name)
						return name.GetNames()
					}
					return unit.Unit.MultiLanguageName.GetNames()
				}(),
				Num: func() float64 {
					if unit.ConversionRate == 0 {
						return 0
					}
					return decimal.NewFromFloat(warehouseItem.Stock).Div(decimal.NewFromFloat(unit.ConversionRate)).Round(3).InexactFloat64()
				}(),
				ConversionRate: unit.ConversionRate,
			}
			notBasicUnitStocks = append(notBasicUnitStocks, notBasicUnitStock)
		}
		warehouseList = append(warehouseList, material_resp.Warehouse{
			Uuid:       warehouseItem.WarehouseUuid,
			LocaleName: localeName,
			Num:        warehouseItem.Stock,
			NotBasicUnitStockList: material_resp.NotBasicUnitStockList{
				List: notBasicUnitStocks,
			},
		})
		amount = amount.Add(decimal.NewFromFloat(warehouseItem.Stock))
	}

	return material_resp.MaterialStockDetailResp{
		Uuid:         req.Uuid,
		LocaleName:   material.MultiLanguageName.GetNames(),
		Code:         material.Code,
		InternalCode: material.InternalCode,
		Warehouses: material_resp.WarehouseList{
			Amount: amount.InexactFloat64(),
			List:   warehouseList,
		},
	}, nil
}

// AddMaterialCategory 创建物品类别
func (s *materialSrv) AddMaterialCategory(ctx context.Context, request req.MaterialCategoryAddReq) error {
	// 大写编码
	request.Code = strings.ToUpper(request.Code)
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		materialCategoryRepo := repository.NewMaterialRepo(tx)

		checkService := NewCheckNameSrv(s.dbm)
		names := checkService.MakeCheckNameList(ctx, request.LocaleName)
		for _, name := range names {
			if !checkService.CheckNameLength(ctx, name.Text, 50) {
				return errors.New("名称长度不能超过50")
			}
		}

		// 检查物品类别编码是否已存在
		if request.Code != "" {
			if exist := materialCategoryRepo.CheckMaterialCategoryCodeExist(request.Code, 0); exist {
				return errors.New("物品类别编码已存在")
			}
		}

		// 创建多语言名称
		multiLanguageName := model.MultiLanguageName{}
		multiLanguageName.InitByLocaleResponse(request.LocaleName)
		nameId, err := repository.NewMultiLanguageNameRepo(tx).CreateMultiLanguageName(multiLanguageName)
		if err != nil {
			return errors.WithMessage(err, "创建多语言名称失败")
		}

		maxSort, err := materialCategoryRepo.GetMaterialCategoryMaxSort(
			repository.NewCommonRepo().WhereBySoftDelete(),
		)
		if err != nil {
			return errors.WithMessage(err, "获取一级分类最大排序失败")
		}
		sort := uint(maxSort + 1)

		// 创建物品类别
		materialCategory := model.MaterialCategory{
			MultiLanguageNameUuid: nameId,
			Name:                  request.LocaleName.ToJson(),
			Code:                  request.Code,
			Sort:                  int(sort),
		}

		if request.GetUuid() != 0 {
			materialCategory.Uuid = request.GetUuid()
		}

		_, err = materialCategoryRepo.CreateMaterialCategory(materialCategory)
		if err != nil {
			return errors.WithMessage(err, "创建物品类别失败")
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

func (s *materialSrv) AddMaterialByEprItem(ctx context.Context, request req.MaterialAddErpReq) (*model.Material, error) {
	// 切换为当前数据库
	db := ctx.GetDB()
	if db == nil {
		dbId := ctx.GetDbId()
		db = s.dbm.GetDB(dbId)
	}

	// 创建多语言名称, 异步翻译
	localeName := model.MultiLanguageName{
		EnName:   request.ItemName,
		ZhName:   request.ItemName,
		ZhTwName: request.ItemName,
		ThName:   request.ItemName,
		MyName:   request.ItemName,
		JaName:   request.ItemName,
		KoName:   request.ItemName,
		TrName:   request.ItemName,
		SvName:   request.ItemName,
	}

	// 获取单位信息
	productUnitRepo := repository.NewProductRepo(db)
	productUnit, err := productUnitRepo.GetProductUnitByErpnextUom(request.StockUom)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("基准单位不存在: %s", request.StockUom))
	}

	// 获取采购单位
	var purchaseUnitUuid uint64
	if request.PurchaseUom != "" {
		purchaseUnit, err := productUnitRepo.GetProductUnitByErpnextUom(request.PurchaseUom)
		if err != nil {
			return nil, errors.WithMessage(err, fmt.Sprintf("采购单位不存在: %s %s", request.ItemCode, request.PurchaseUom))
		}
		purchaseUnitUuid = purchaseUnit.Uuid
	}

	var material *model.Material

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		productUnitRepo := repository.NewProductRepo(tx)

		// 获取物品分类信息
		commonRepo := repository.NewCommonRepo()
		materialCategoryRepo := repository.NewMaterialRepo(tx)
		materialCategory, exists, err := materialCategoryRepo.GetMaterialCategoryByCode(request.ClassificationCode)
		if err != nil {
			return errors.WithMessage(err, fmt.Sprintf("获取物品分类失败: %s", request.ClassificationCode))
		}
		var materialCategoryUuid uint64 = 1
		if exists {
			materialCategoryUuid = materialCategory.Uuid
		}

		unitList := []req.MaterialUnitReq{}
		for _, unit := range request.Uoms {
			if unit.Uom == request.StockUom {
				continue
			}
			// 查询单位信息
			productUnit, err := productUnitRepo.GetProductUnitByErpnextUom(unit.Uom)
			if err != nil {
				return errors.WithMessage(err, fmt.Sprintf("单位不存在: %s", unit.Uom))
			}
			unitList = append(unitList, req.MaterialUnitReq{
				Uuid:           productUnit.Uuid,
				ConversionRate: unit.ConversionRate,
			})
		}

		// 获取单位信息
		params := req.MaterialAddReq{
			LocaleName:   localeName.GetNames(),
			CategoryUuid: materialCategoryUuid,
			Status: func() int {
				if request.Disabled {
					return 0
				}
				return 1
			}(),
			Valuation:        request.ValuationRate,
			InitStock:        request.OpeningStock,
			BarcodeValue:     request.BarcodeValue,
			UnitUuid:         productUnit.Uuid,
			UnitList:         unitList,
			PurchaseUnitUuid: purchaseUnitUuid,
			CostUnitUuid:     productUnit.Uuid,
			InternalCode:     request.InternalCode,
		}
		// 获取总部ID
		companySetting := ctx.GetCompanySetting()
		var headquarterUuid uint64
		if companySetting.IsSubShop() {
			headquarterUuid = companySetting.HeadquarterUuid
		}
		params.SetHeadquarterUuid(headquarterUuid)
		// 获取默认仓库ID
		warehouseUuid, err := repository.NewWarehouseRepo(tx).GetDefaultWarehouse()
		if err != nil {
			return errors.WithMessage(err, "默认仓库不存在")
		}
		params.SetWarehouseUuid(warehouseUuid.Uuid)
		params.SetIsSync(true)    // 从总店同步物品到本地，忽略检查内部编码唯一性
		params.SetIsErpSync(true) // 从ERP同步物品到本地，不创建仓库物品
		material, _, err = addMaterial(ctx, tx, s.settingSrv, params)
		if err != nil {
			return errors.WithMessage(err)
		}
		// 更新物品数据, 根据NotForSale判断是否删除
		materialRepo := repository.NewMaterialRepo(tx)
		updateData := map[string]any{"code": request.ItemCode}
		if request.NotForSale {
			updateData["delete_time"] = time.Now().Unix()
		} else {
			updateData["delete_time"] = 0
		}
		err = materialRepo.UpdateMaterialData(updateData, commonRepo.WhereByUuid(material.Uuid))
		if err != nil {
			return errors.WithMessage(err, "更新物品数据失败")
		}
		return nil
	}); err != nil {
		return nil, errors.WithMessage(err)
	}

	return material, nil
}

// AddMaterial 添加物品
func (s *materialSrv) AddMaterial(ctx context.Context, req req.MaterialAddReq) error {
	// 去除两端空格，大写编码
	req.InternalCode = strings.ToUpper(strings.TrimSpace(req.InternalCode))
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)

	// 检查物品名称
	productCheckSrv := NewProductCheckSrv(s.dbm, s.localeSrv, s.settingSrv)
	if err := productCheckSrv.CheckMaterialName(ctx, 0, req.LocaleName); err != nil {
		return errors.WithMessage(err, "检查物品名称失败")
	}

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		material, materialAddErpReq, err := addMaterial(ctx, tx, s.settingSrv, req)
		if err != nil {
			return errors.WithMessage(err)
		}

		code := ""
		if ctx.GetCompany().IsOpenErp() {
			erpSrv := erp.NewIErpSrv(s.dbm)
			itemInfo, errErp := erpSrv.AddMaterial(ctx, *materialAddErpReq)
			if errErp != nil {
				return errors.WithMessage(errErp)
			}
			code = itemInfo.ItemCode
		} else {
			uuid, _ := utils.GetID()
			code = fmt.Sprintf("WPR%d", uuid)
		}
		materialRepo := repository.NewMaterialRepo(tx)
		err = materialRepo.UpdateMaterialCode(material.Uuid, code)
		if err != nil {
			return errors.WithMessage(err, "更新物品编码失败")
		}

		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

func addMaterial(ctx context.Context, tx *gorm.DB, settingSrv setting.ISrv, request req.MaterialAddReq) (*model.Material, *req.MaterialAddErpReq, error) {
	// 非基准单位不能重复
	requestUnitList := make(map[uint64]bool)
	for _, unit := range request.UnitList {
		if requestUnitList[unit.Uuid] {
			return nil, nil, errors.WithMessage(errors.New("非基准单位不能重复"))
		}
		requestUnitList[unit.Uuid] = true
	}

	// 检查条形码唯一性
	if request.BarcodeValue != "" {
		materialRepo := repository.NewMaterialRepo(tx)
		if materialRepo.CheckBarcodeExist(request.BarcodeValue, 0) {
			if request.GetIsSync() {
				logger.Logger.Error("同步物品时，条形码已存在，但已同步物品，忽略检查唯一性", zap.String("barcode", request.BarcodeValue), zap.Any("request", request))
			} else {
				return nil, nil, errors.WithMessage(errors.New("条形码已存在，请使用其他条形码"))
			}
		}
	}
	// 检查内部编码唯一性
	if request.InternalCode != "" {
		materialRepo := repository.NewMaterialRepo(tx)
		if materialRepo.CheckMaterialInternalCodeExist(request.InternalCode, 0) {
			if request.GetIsSync() {
				logger.Logger.Error("同步物品时，内部编码已存在，但已同步物品，忽略检查唯一性", zap.String("internal_code", request.InternalCode), zap.Any("request", request))
			} else {
				return nil, nil, errors.WithMessage(errors.New("内部编码已存在，请使用其他内部编码"), request.InternalCode)
			}
		}
	}

	materialUuid, _ := utils.GetID()
	materialRepo := repository.NewMaterialRepo(tx)
	// 创建多语言名称
	multiLanguageName := model.MultiLanguageName{}
	multiLanguageName.InitByLocaleResponse(request.LocaleName)
	nameId, err := repository.NewMultiLanguageNameRepo(tx).CreateMultiLanguageName(multiLanguageName)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "创建多语言名称失败")
	}

	unitMap := make(map[uint64]uint64) // 单位ID -> 原料单位ID

	// 创建material_unit
	productUnit, err := repository.NewProductRepo(tx).GetProductUnitByUnitUuid(request.UnitUuid)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "获取产品单位失败")
	}
	unit := model.MaterialUnit{
		Name:           productUnit.MultiLanguageName.ToJson(),
		UnitUuid:       request.UnitUuid,
		ConversionRate: 1,
		IsDefault:      1,
		FromUnitUuid:   0,
		MaterialUuid:   materialUuid,
		Unit:           productUnit,
	}
	unitUuid, err := repository.NewMaterialUnitRepo(tx).CreateMaterialUnit(unit)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "创建原料单位失败")
	}

	unitMap[request.UnitUuid] = unitUuid

	notBaseUnitList := []*model.MaterialUnit{}
	unitList := []req.MaterialUomReq{}
	// 创建非基准单位 material_unit
	for _, unit := range request.UnitList {
		productUnit, err := repository.NewProductRepo(tx).GetProductUnitByUnitUuid(unit.Uuid)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "获取产品单位失败")
		}

		materialUnitUuid, _ := utils.GetID()
		notBaseUnitList = append(notBaseUnitList, &model.MaterialUnit{
			BaseModel: model.BaseModel{
				Uuid: materialUnitUuid,
			},
			Name:           productUnit.MultiLanguageName.ToJson(),
			UnitUuid:       productUnit.Uuid,
			ConversionRate: unit.ConversionRate,
			FromUnitUuid:   unitUuid,
			MaterialUuid:   materialUuid,
			Unit:           productUnit,
		})
		unitMap[unit.Uuid] = materialUnitUuid

		unitList = append(unitList, req.MaterialUomReq{
			Uom:            productUnit.ErpnextUom,
			ConversionRate: unit.ConversionRate,
		})
	}

	warehouseUuid := request.GetWarehouseUuid()
	if warehouseUuid == 0 {
		// 获取默认仓库ID
		warehouse, err := repository.NewWarehouseRepo(tx).GetDefaultWarehouse()
		if err != nil {
			return nil, nil, errors.WithMessage(err, "默认仓库不存在")
		}
		warehouseUuid = warehouse.Uuid
	}

	// 创建物品
	material := model.Material{
		BaseModel: model.BaseModel{
			Uuid: materialUuid,
		},
		Name:                  request.LocaleName.ToJson(),
		Code:                  "", // 先添加物品，之后调用erp接口后再更新编码
		Valuation:             request.Valuation,
		InitStock:             request.InitStock,
		StockNum:              request.InitStock,
		MultiLanguageNameUuid: nameId,
		CategoryUuid:          request.CategoryUuid,
		UnitUuid:              unitMap[request.UnitUuid],
		BarcodeValue:          request.BarcodeValue,
		PurchaseUnitUuid:      unitMap[request.PurchaseUnitUuid],
		CostUnitUuid:          unitMap[request.CostUnitUuid],
		SafetyStock:           request.SafetyStock,
		Status: func() bool {
			if request.Status == 1 {
				return true
			}
			return false
		}(),
		NotBaseUnitList: notBaseUnitList,
		Unit:            &unit,
		InternalCode:    request.InternalCode,
		AllowSubstoreVisible: func() int {
			// 如果前端版本号小于 2.10，则都等于 1
			if ctx.Version(context.LT, "2.10.0") {
				return 1
			}
			return request.AllowSubstoreVisible
		}(),
	}

	// 设置总部ID
	headquarterUuid := request.GetHeadquarterUuid()
	if headquarterUuid != 0 {
		material.HeadquarterUuid = headquarterUuid
	}
	// 设置仓库ID
	if warehouseUuid != 0 {
		material.WarehouseUuid = warehouseUuid
	}

	_, err = materialRepo.CreateMaterial(material)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "创建物品失败")
	}

	// 如果不是从ERP同步物品到本地，则创建仓库物品库存记录
	if !request.GetIsErpSync() {
		warehouseItem := model.WarehouseItem{
			WarehouseUuid: warehouseUuid,
			MaterialUuid:  materialUuid,
			MaterialCode:  "",
			Stock:         request.InitStock,
			Valuation:     request.Valuation,
		}
		err = repository.NewWarehouseItemRepo(tx).Create(&warehouseItem)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "创建仓库商品库存记录失败")
		}
	}

	for _, unit := range notBaseUnitList {
		_, err = repository.NewMaterialUnitRepo(tx).CreateMaterialUnit(*unit)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "创建原料单位失败")
		}
	}

	getEnName, err := GetEnName(ctx, settingSrv, request.LocaleName)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "翻译失败")
	}
	// 获取物品分类信息
	materialCategoryRepo := repository.NewMaterialRepo(tx)
	materialCategory, err := materialCategoryRepo.GetMaterialCategoryByUuid(request.CategoryUuid)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "获取物品分类失败")
	}

	getMaterialCategoryName, err := GetEnName(ctx, settingSrv, materialCategory.MultiLanguageName.GetNames())
	if err != nil {
		return nil, nil, errors.WithMessage(err, "翻译失败")
	}

	// 获取采购单位
	var purchaseUom string
	if material.PurchaseUnitUuid != 0 {
		purchaseUnit, err := repository.NewMaterialUnitRepo(tx).GetMaterialUnitsByUuid(material.PurchaseUnitUuid)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "获取采购单位失败")
		}
		if purchaseUnit.Unit != nil {
			purchaseUom = purchaseUnit.Unit.ErpnextUom
		}
	}

	materialAddErpReq := &req.MaterialAddErpReq{
		ItemName:           getEnName,
		StockUom:           productUnit.ErpnextUom,
		BarcodeValue:       request.BarcodeValue,
		Disabled:           request.Status == 0,
		ValuationRate:      request.Valuation,
		OpeningStock:       request.InitStock,
		Uoms:               unitList,
		InternalCode:       request.InternalCode,
		Classification:     getMaterialCategoryName,
		ClassificationCode: materialCategory.Code,
		PurchaseUom:        purchaseUom,
	}

	return &material, materialAddErpReq, nil
}

// EditMaterial 编辑物品
func (s *materialSrv) EditMaterial(ctx context.Context, request req.MaterialEditReq) error {
	// 去除两端空格，大写编码
	request.InternalCode = strings.ToUpper(strings.TrimSpace(request.InternalCode))
	// 验证请求参数
	if err := request.Validate(); err != nil {
		return errors.WithMessage(err)
	}

	// 检查物品名称
	productCheckSrv := NewProductCheckSrv(s.dbm, s.localeSrv, s.settingSrv)
	if err := productCheckSrv.CheckMaterialName(ctx, request.Uuid, request.LocaleName); err != nil {
		return errors.WithMessage(err, "检查物品名称失败")
	}

	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		materialRepo := repository.NewMaterialRepo(tx)

		// 检查物品是否存在
		existingMaterial, err := materialRepo.GetMaterialDetailByUuid(request.Uuid)
		if err != nil {
			return errors.WithMessage(err, "物品不存在")
		}

		// 如果是总部物品，则不能修改
		if existingMaterial.IsHeadquarter() {
			return errors.WithMessage(errors.New("总部物品不能修改"))
		}

		// 检查条形码唯一性
		if request.BarcodeValue != "" && request.BarcodeValue != existingMaterial.BarcodeValue {
			if materialRepo.CheckBarcodeExist(request.BarcodeValue, request.Uuid) {
				return errors.WithMessage(errors.New("条形码已存在，请使用其他条形码"))
			}
		}
		// 检查内部编码唯一性
		if request.InternalCode != "" && request.InternalCode != existingMaterial.InternalCode {
			if materialRepo.CheckMaterialInternalCodeExist(request.InternalCode, request.Uuid) {
				return errors.WithMessage(errors.New("内部编码已存在，请使用其他内部编码"))
			}
		}

		// 判断名称是否修改
		if existingMaterial.MultiLanguageName.IsNameChanged(request.LocaleName) {
			// 更新多语言名称
			multiLanguageName := model.MultiLanguageName{}
			multiLanguageName.InitByLocaleResponse(request.LocaleName)
			multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
			err = multiLanguageNameRepo.UpdateMultiLanguageName(existingMaterial.MultiLanguageNameUuid, multiLanguageName)
			if err != nil {
				return errors.WithMessage(err, "创建多语言名称失败")
			}
			existingMaterial.Name = multiLanguageName.ToJson()
		}

		// 判断非基准单位是否有新增
		unitMap := make(map[uint64]uint64)       // 单位ID -> 原料单位ID
		exitUnitMap := make(map[uint64]bool)     //  单位ID 是否存在
		notBaseUnitList := make(map[uint64]bool) // 已经添加的物品非基准单位列表
		for _, unit := range existingMaterial.NotBaseUnitList {
			unitMap[unit.Unit.Uuid] = unit.Uuid
			exitUnitMap[unit.UnitUuid] = true // 单位ID 是否存在
			notBaseUnitList[unit.Uuid] = true
		}

		// 过滤掉旧的基准单位
		unitList := make([]req.MaterialUnitReq, 0)
		for _, unit := range request.UnitList {
			if _, ok := notBaseUnitList[unit.Uuid]; !ok {
				unitList = append(unitList, req.MaterialUnitReq{
					Uuid:           unit.Uuid,
					ConversionRate: unit.ConversionRate,
				})
			}
		}
		request.UnitList = unitList

		for _, unit := range request.UnitList {
			if _, ok := exitUnitMap[unit.Uuid]; !ok {
				productUnit, err := repository.NewProductRepo(tx).GetProductUnitByUnitUuid(unit.Uuid)
				if err != nil {
					return errors.WithMessage(err, "获取产品单位失败")
				}
				// 新增非基准单位
				materialUnitUuid, err := repository.NewMaterialUnitRepo(tx).CreateMaterialUnit(model.MaterialUnit{
					Name:           productUnit.MultiLanguageName.ToJson(),
					UnitUuid:       productUnit.Uuid,
					ConversionRate: unit.ConversionRate,
					FromUnitUuid:   existingMaterial.UnitUuid,
					MaterialUuid:   request.Uuid,
				})
				if err != nil {
					return errors.WithMessage(err, "创建原料单位失败")
				}
				unitMap[unit.Uuid] = materialUnitUuid
			} else {
				return errors.New("非基准单位已存在")
			}
		}

		// 更新物品
		material := model.Material{
			BaseModel: model.BaseModel{
				Uuid: request.Uuid,
			},
			Name:         existingMaterial.Name,
			CategoryUuid: request.CategoryUuid,
			Status: func() bool {
				if request.Status == 1 {
					return true
				}
				return false
			}(),
			Valuation:    request.Valuation,
			BarcodeValue: request.BarcodeValue,
			InternalCode: request.InternalCode,
			SafetyStock:  request.SafetyStock,
			PurchaseUnitUuid: func() uint64 {
				// 如果选择了已经存在的单位，则使用已存在的单位
				if _, ok := notBaseUnitList[request.PurchaseUnitUuid]; ok {
					return request.PurchaseUnitUuid
				}
				// 如果选择了新的单位，则创建新的单位
				return unitMap[request.PurchaseUnitUuid]
			}(),
			CostUnitUuid: func() uint64 {
				if _, ok := notBaseUnitList[request.CostUnitUuid]; ok {
					return request.CostUnitUuid
				}
				return unitMap[request.CostUnitUuid]
			}(),
			AllowSubstoreVisible: func() int {
				// 如果前端版本号小于 2.10，则都等于 1
				if ctx.Version(context.LT, "2.10.0") {
					return 1
				}
				return request.AllowSubstoreVisible
			}(),
		}

		err = materialRepo.UpdateMaterial(material)
		if err != nil {
			return errors.WithMessage(err, "更新物品失败")
		}

		// 物品 - 将多语言名称uuid从待翻译集合中删除
		s.translateSrv.RemoveMultiLanguageNameUuidFromSet(ctx.GetCompanyUuid(), existingMaterial.MultiLanguageNameUuid)

		if request.Status == 0 {
			err = materialRepo.UpdateMaterialStatus(request.Uuid, false)
			if err != nil {
				return errors.WithMessage(err, "更新物品状态失败")
			}
		}

		if request.BarcodeValue == "" {
			err = materialRepo.ClearMaterialBarcodeValue(request.Uuid)
			if err != nil {
				return errors.WithMessage(err, "清空物品条形码值失败")
			}
		}

		if request.Valuation == 0 {
			err = materialRepo.ClearMaterialValuation(request.Uuid)
			if err != nil {
				return errors.WithMessage(err, "清空物品估值率失败")
			}
		}

		if request.InternalCode == "" {
			err = materialRepo.ClearMaterialInternalCode(request.Uuid)
			if err != nil {
				return errors.WithMessage(err, "清空物品内部编码失败")
			}
		}

		// 如果安全库存为nil，则清空安全库存字段
		if request.SafetyStock == nil {
			err = materialRepo.ClearMaterialSafetyStock(request.Uuid)
			if err != nil {
				return errors.WithMessage(err, "清空物品安全库存失败")
			}
		}

		if ctx.GetCompany().IsOpenErp() {
			erpSrv := erp.NewIErpSrv(s.dbm)
			enName, err := GetEnName(ctx, s.settingSrv, request.LocaleName)
			if err != nil {
				return errors.WithMessage(err, "翻译失败")
			}

			unitList := []req.MaterialUomReq{}
			// 新增的非基准单位
			for _, unit := range request.UnitList {
				productUnit, err := repository.NewProductRepo(tx).GetProductUnitByUnitUuid(unit.Uuid)
				if err != nil {
					return errors.WithMessage(err, "获取产品单位失败")
				}

				unitList = append(unitList, req.MaterialUomReq{
					Uom:            productUnit.ErpnextUom,
					ConversionRate: unit.ConversionRate,
				})
			}
			// 旧的非基准单位
			for _, unit := range existingMaterial.NotBaseUnitList {
				unitList = append(unitList, req.MaterialUomReq{
					Uom:            unit.Unit.ErpnextUom,
					ConversionRate: unit.ConversionRate,
				})
			}

			// 获取物品分类信息
			materialCategoryRepo := repository.NewMaterialRepo(tx)
			materialCategory, err := materialCategoryRepo.GetMaterialCategoryByUuid(request.CategoryUuid)
			if err != nil {
				return errors.WithMessage(err, "获取物品分类失败")
			}
			materialCategoryName := language.JsonToLocaleResponse(materialCategory.MultiLanguageName.ToJson())
			getMaterialCategoryName, err := GetEnName(ctx, s.settingSrv, *materialCategoryName)
			if err != nil {
				return errors.WithMessage(err, "翻译失败")
			}

			// 获取采购单位
			var purchaseUom string
			purchaseUnit, err := repository.NewMaterialUnitRepo(tx).GetMaterialUnitsByUuid(material.PurchaseUnitUuid)
			if err != nil {
				return errors.WithMessage(err, "获取采购单位失败")
			}
			if purchaseUnit.Unit != nil {
				purchaseUom = purchaseUnit.Unit.ErpnextUom
			}

			_, errErp := erpSrv.AddMaterial(ctx, req.MaterialAddErpReq{
				ItemCode:      existingMaterial.Code,
				ItemName:      enName,
				StockUom:      existingMaterial.Unit.Unit.ErpnextUom,
				Disabled:      request.Status == 0,
				ValuationRate: material.Valuation,
				BarcodeValue:  material.BarcodeValue,
				Uoms:          unitList,
				InternalCode: func() string {
					if request.InternalCode != "" {
						return request.InternalCode
					}
					return " " // 内部编码为空时，传空格给ErpNext
				}(),
				Classification:     getMaterialCategoryName,
				ClassificationCode: materialCategory.Code,
				PurchaseUom:        purchaseUom,
			})
			if errErp != nil {
				return errors.WithMessage(errErp)
			}
		}

		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

func (s *materialSrv) UpdateMaterialByEprItem(ctx context.Context, request req.MaterialEditErpReq) error {
	db := ctx.GetDB()
	if db == nil {
		dbId := ctx.GetDbId()
		db = s.dbm.GetDB(dbId)
	}

	commonRepo := repository.NewCommonRepo()

	err := commonRepo.Transaction(db, func(tx *gorm.DB) error {
		materialRepo := repository.NewMaterialRepo(tx)
		productUnitRepo := repository.NewProductUnitRepo(tx)
		materialUnitRepo := repository.NewMaterialUnitRepo(tx)

		updateData := map[string]any{
			"valuation":     request.ValuationRate, // 估值率
			"barcode_value": request.BarcodeValue,  // 条形码值
			"internal_code": request.InternalCode,  // 内部编码
			"status": func() int {
				if request.Disabled {
					return 0
				}
				return 1
			}(),
		}

		material, err := materialRepo.GetMaterialDetailContainsDeletedByUuid(request.Uuid)
		if err != nil {
			return errors.WithMessage(err, "物品不存在:"+strconv.FormatUint(request.Uuid, 10))
		}

		// 同步物品分类
		var materialCategoryUuid uint64 = 1
		if request.ClassificationCode != "" {
			materialCategory, exists, err := materialRepo.GetMaterialCategoryByCode(request.ClassificationCode)
			if err != nil {
				return errors.WithMessage(err, "获取物品分类失败："+request.ClassificationCode)
			}
			if exists {
				materialCategoryUuid = materialCategory.Uuid
			}
		}
		updateData["category_uuid"] = materialCategoryUuid

		// 基准单位
		stockUnit, err := productUnitRepo.GetProductUnitByErpnextUom(request.StockUom)
		if err != nil {
			return errors.WithMessage(err, fmt.Sprintf("基准单位不存在：%s %s", request.ItemCode, request.StockUom))
		}

		// 采购单位
		var purchaseUnitUuid uint64
		if request.PurchaseUom != "" {
			purchaseUnit, err := productUnitRepo.GetProductUnitByErpnextUom(request.PurchaseUom)
			if err != nil {
				return errors.WithMessage(err, fmt.Sprintf("采购单位不存在: %s %s", request.ItemCode, request.PurchaseUom))
			}
			purchaseUnitUuid = purchaseUnit.Uuid
		}

		// 同步单位
		saveUnitUuids := []uint64{}
		for _, uom := range request.Uoms {
			productUnit, _ := productUnitRepo.GetProductUnitByErpnextUom(uom.Uom)
			if productUnit == nil {
				return errors.WithMessage(errors.New(fmt.Sprintf("采购单位不存在: %s %s", request.ItemCode, uom.Uom)))
			}

			if productUnit.Uuid == stockUnit.Uuid {
				// 基准单位
				err := materialUnitRepo.UpdateMaterialUnit(map[string]any{
					"name":        productUnit.Name,
					"unit_uuid":   productUnit.Uuid,
					"is_default":  1,
					"delete_time": 0,
				}, commonRepo.WhereByUuid(material.Unit.Uuid))
				if err != nil {
					return errors.WithMessage(err, "更新基准单位失败")
				}
				saveUnitUuids = append(saveUnitUuids, material.Unit.Uuid)
				if productUnit.Uuid == purchaseUnitUuid {
					updateData["purchase_unit_uuid"] = material.Unit.Uuid
				}
			} else {
				// 非基准单位
				existUnit, exist := slice.FindBy(material.NotBaseUnitList, func(index int, item *model.MaterialUnit) bool {
					return item.UnitUuid == productUnit.Uuid && item.DeleteTime == 0
				})
				if exist {
					err := materialUnitRepo.UpdateMaterialUnit(map[string]any{
						"name":            productUnit.Name,
						"conversion_rate": uom.ConversionRate,
					}, commonRepo.WhereByUuid(existUnit.Uuid))
					if err != nil {
						return errors.WithMessage(err, "更新非基准单位失败")
					}
					saveUnitUuids = append(saveUnitUuids, existUnit.Uuid)
					if productUnit.Uuid == purchaseUnitUuid {
						updateData["purchase_unit_uuid"] = existUnit.Uuid
					}
				} else {
					uuid, err := materialUnitRepo.CreateMaterialUnit(model.MaterialUnit{
						Name:           productUnit.Name,
						UnitUuid:       productUnit.Uuid,
						ConversionRate: uom.ConversionRate,
						FromUnitUuid:   material.Unit.Uuid,
						MaterialUuid:   material.Uuid,
					})
					if err != nil {
						return errors.WithMessage(err, "创建非基准单位失败")
					}
					saveUnitUuids = append(saveUnitUuids, uuid)
					if productUnit.Uuid == purchaseUnitUuid {
						updateData["purchase_unit_uuid"] = uuid
					}
				}
			}
		}

		if len(saveUnitUuids) > 0 {
			err := materialUnitRepo.DestroyMaterialUnit(
				commonRepo.WhereByUuidNotIn(saveUnitUuids),
				commonRepo.WhereBySoftDelete(),
				commonRepo.WhereByMaterialUuid(material.Uuid),
			)
			if err != nil {
				return errors.WithMessage(err, "删除非基准单位失败")
			}
		}

		if !slices.Contains(saveUnitUuids, material.CostUnitUuid) {
			updateData["cost_unit_uuid"] = 0
		}

		// 根据NotForSale设置删除时间，物品下架时设置为当前时间，上架时重置为0
		if request.NotForSale && material.DeleteTime == 0 {
			updateData["delete_time"] = time.Now().Unix()
		} else if !request.NotForSale {
			updateData["delete_time"] = 0
		}

		return materialRepo.UpdateMaterialData(updateData, commonRepo.WhereByUuid(material.Uuid))

	})

	if err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

// DeleteMaterial 删除物品
func (s *materialSrv) DeleteMaterial(ctx context.Context, req req.MaterialDeleteReq) error {
	dbId := ctx.GetDbId()
	materialRepo := repository.NewMaterialRepo(s.dbm.GetDB(dbId))

	// 检查物品是否存在
	material, err := materialRepo.GetMaterialDetailByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "物品不存在")
	}

	// 如果是总部物品，则不能删除
	if material.IsHeadquarter() {
		return errors.WithMessage(errors.New("总部物品不能删除"))
	}

	// 删除物品
	err = materialRepo.DeleteMaterial(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "删除物品失败")
	}

	return nil
}

// UpdateMaterialStatusBatch 批量修改物品状态
func (s *materialSrv) UpdateMaterialStatusBatch(ctx context.Context, request req.MaterialStatusReq) error {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		materialRepo := repository.NewMaterialRepo(tx)
		if err := materialRepo.UpdateMaterialStatusBatch(request.Uuids, request.Status); err != nil {
			return errors.WithMessage(err)
		}

		// 调用erp接口,更新状态
		if ctx.GetCompany().IsOpenErp() {
			existingMaterials, err := materialRepo.GetMaterialDetailByUuids(request.Uuids)
			if err != nil {
				return errors.WithMessage(err, "获取物品详情失败")
			}
			erpSrv := erp.NewIErpSrv(s.dbm)
			for _, existingMaterial := range existingMaterials {
				// 如果是总部物品，则不能更新状态
				if existingMaterial.IsHeadquarter() {
					continue
				}

				enName, err := GetEnName(ctx, s.settingSrv, existingMaterial.MultiLanguageName.GetNames())
				if err != nil {
					return errors.WithMessage(err, "翻译失败")
				}

				// 旧的非基准单位
				unitList := []req.MaterialUomReq{}
				for _, unit := range existingMaterial.NotBaseUnitList {
					unitList = append(unitList, req.MaterialUomReq{
						Uom:            unit.Unit.ErpnextUom,
						ConversionRate: unit.ConversionRate,
					})
				}

				// 获取采购单位
				var purchaseUom string
				purchaseUnit, err := repository.NewMaterialUnitRepo(db).GetMaterialUnitsByUuid(existingMaterial.PurchaseUnitUuid)
				if err != nil {
					return errors.WithMessage(err, "获取采购单位失败")
				}
				if purchaseUnit.Unit != nil {
					purchaseUom = purchaseUnit.Unit.ErpnextUom
				}

				_, errErp := erpSrv.AddMaterial(ctx, req.MaterialAddErpReq{
					ItemCode:      existingMaterial.Code,
					ItemName:      enName,
					StockUom:      existingMaterial.Unit.Unit.ErpnextUom,
					Disabled:      request.Status == 0,
					ValuationRate: existingMaterial.Valuation,
					BarcodeValue:  existingMaterial.BarcodeValue,
					Uoms:          unitList,
					InternalCode:  existingMaterial.InternalCode,
					PurchaseUom:   purchaseUom,
				})
				if errErp != nil {
					return errors.WithMessage(errErp)
				}
			}
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

// UpdateMaterialVisibleBatch 批量更新物品可见性
func (s *materialSrv) UpdateMaterialVisibleBatch(ctx context.Context, request req.MaterialBatchUpdateVisibleReq) (int, error) {
	// 验证请求参数
	if err := request.Validate(); err != nil {
		return 0, errors.WithMessage(err)
	}

	// 检查权限：仅总店可以更新可见性
	companySetting := ctx.GetCompanySetting()
	if companySetting.IsSubShop() {
		return 0, errors.WithMessage(errors.New("子店无权修改物品可见性设置"))
	}

	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)

	var updatedCount int64
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		materialRepo := repository.NewMaterialRepo(tx)

		// 检查物品是否存在且不是总部物品（总部物品不能修改）
		materials, err := materialRepo.GetMaterialDetailByUuids(request.Uuids)
		if err != nil {
			return errors.WithMessage(err, "获取物品详情失败")
		}

		// 过滤掉总部物品
		validUuids := make([]uint64, 0)
		for _, material := range materials {
			if !material.IsHeadquarter() {
				validUuids = append(validUuids, material.Uuid)
			}
		}

		if len(validUuids) == 0 {
			return errors.WithMessage(errors.New("没有可更新的物品（总部物品不能修改）"))
		}

		// 批量更新可见性
		if err := materialRepo.UpdateMaterialVisibleBatch(validUuids, request.AllowSubstoreVisible); err != nil {
			return errors.WithMessage(err)
		}

		updatedCount = int64(len(validUuids))
		return nil
	}); err != nil {
		return 0, errors.WithMessage(err)
	}

	return int(updatedCount), nil
}

// GetMaterialCategoryList 获取物品类别列表
func (s *materialSrv) GetMaterialCategoryList(ctx context.Context, req req.MaterialCategoryListReq) (material_resp.MaterialCategoryListResp, error) {
	dbId := ctx.GetDbId()
	materialCategoryRepo := repository.NewMaterialRepo(s.dbm.GetDB(dbId))

	// 获取物品类别列表
	materialCategories, err := materialCategoryRepo.GetMaterialCategoryList()
	if err != nil {
		return material_resp.MaterialCategoryListResp{}, errors.WithMessage(err, "获取物品类别列表失败")
	}

	// 转换为响应格式
	var materialCategoryList []material_resp.MaterialCategory
	for _, materialCategory := range materialCategories {
		materialCategoryList = append(materialCategoryList, material_resp.MaterialCategory{
			Uuid:       materialCategory.Uuid,
			Name:       materialCategory.Name,
			LocaleName: materialCategory.MultiLanguageName.GetNames(),
			Code:       materialCategory.Code,
			Sort:       materialCategory.Sort,
		})
	}

	return material_resp.MaterialCategoryListResp{
		List: materialCategoryList,
	}, nil
}

func (s *materialSrv) GetMaterialCategoryDetail(ctx context.Context, req req.MaterialCategoryDetailReq) (*material_resp.MaterialCategory, error) {
	dbId := ctx.GetDbId()
	materialCategoryRepo := repository.NewMaterialRepo(s.dbm.GetDB(dbId))
	materialCategory, err := materialCategoryRepo.GetMaterialCategoryByUuid(req.Uuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取物品类别失败")
	}
	// 是否关联了物品
	materialRepo := repository.NewMaterialRepo(s.dbm.GetDB(dbId))
	materials, err := materialRepo.GetMaterialByCategoryUuid(materialCategory.Uuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取物品失败")
	}
	return &material_resp.MaterialCategory{
		Uuid:       materialCategory.Uuid,
		Name:       materialCategory.Name,
		LocaleName: materialCategory.MultiLanguageName.GetNames(),
		Code:       materialCategory.Code,
		Sort:       materialCategory.Sort,
		IsRelated:  len(materials) > 0,
		IsEditable: !materialCategory.IsHeadquarter(),
	}, nil
}

func (s *materialSrv) SortMaterialCategory(ctx context.Context, req req.MaterialCategorySortReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	commonRepo := repository.NewCommonRepo()
	productRepo := repository.NewProductRepo(db)

	materialCategoryUuids := make([]uint64, 0, len(req.List))
	for _, item := range req.List {
		materialCategoryUuids = append(materialCategoryUuids, item.Uuid)
	}
	productCategories, _ := productRepo.GetMaterialCategoryCount(
		commonRepo.WhereBySoftDelete(),
		productRepo.WhereUuidIn(materialCategoryUuids),
	)
	if productCategories != int64(len(materialCategoryUuids)) {
		return errors.New("分类不存在")
	}

	sorts := make(map[uint64]int)
	for _, item := range req.List {
		if item.Sort == 0 {
			return errors.New("排序不能为0")
		}
		sorts[item.Uuid] = item.Sort
	}
	err := productRepo.BatchUpdateSort(&model.MaterialCategory{}, sorts)
	if err != nil {
		return errors.WithMessage(errors.New("排序分类失败"), err.Error())
	}
	return nil
}

func (s *materialSrv) EditMaterialCategory(ctx context.Context, request req.MaterialCategoryEditReq) error {
	// 大写编码
	request.Code = strings.ToUpper(request.Code)
	db := s.dbm.GetDB(ctx.GetDbId())
	materialCategoryRepo := repository.NewMaterialRepo(db)
	materialCategory, err := materialCategoryRepo.GetMaterialCategoryByUuid(request.Uuid)
	if err != nil {
		return errors.WithMessage(err, "获取物品类别失败")
	}
	changeCode := false
	if materialCategory.Code != request.Code {
		changeCode = true
	}
	materialCategory.Name = request.LocaleName.ToJson()
	materialCategory.Code = request.Code
	materialCategory.MultiLanguageName.InitByLocaleResponse(request.LocaleName)

	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, request.LocaleName)
	for _, name := range names {
		if !checkService.CheckNameLength(ctx, name.Text, 50) {
			return errors.New("名称长度不能超过50")
		}
	}

	// 检查物品类别编码是否已存在
	if request.Code != "" {
		if exist := materialCategoryRepo.CheckMaterialCategoryCodeExist(request.Code, request.Uuid); exist {
			return errors.New("物品类别编码已存在")
		}
	}

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		materialCategoryRepo := repository.NewMaterialRepo(tx)
		if err := materialCategoryRepo.UpdateMaterialCategory(*materialCategory); err != nil {
			return errors.WithMessage(err, "更新物品类别失败")
		}
		// 更新多语言名称
		if err = repository.NewMultiLanguageNameRepo(tx).UpdateMultiLanguageName(materialCategory.MultiLanguageNameUuid, materialCategory.MultiLanguageName); err != nil {
			return errors.WithMessage(err, "更新多语言名称失败")
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	// 异步执行同步erp物品，避免阻塞主线程造成前端请求超时
	utils.Go(func() {
		// 如果修改了物品编码，同步更新所有关联了这个分类的erp物品
		if ctx.GetCompany().IsOpenErp() {
			if changeCode {
				materialRepo := repository.NewMaterialRepo(db)
				materials, err := materialRepo.GetMaterialByCategoryUuid(materialCategory.Uuid)
				if err != nil {
					ctx.Log().Error("获取物品失败", zap.Error(err))
					return
				}
				erpSrv := erp.NewIErpSrv(s.dbm)
				for _, material := range materials {
					// 如果是总部物品，则不能更新
					if material.IsHeadquarter() {
						continue
					}

					enName, err := GetEnName(ctx, s.settingSrv, material.MultiLanguageName.GetNames())
					if err != nil {
						ctx.Log().Error("翻译失败", zap.Error(err))
						return
					}

					// 旧的非基准单位
					unitList := []req.MaterialUomReq{}
					for _, unit := range material.NotBaseUnitList {
						unitList = append(unitList, req.MaterialUomReq{
							Uom:            unit.Unit.ErpnextUom,
							ConversionRate: unit.ConversionRate,
						})
					}

					getMaterialCategoryName, err := GetEnName(ctx, s.settingSrv, materialCategory.MultiLanguageName.GetNames())
					if err != nil {
						ctx.Log().Error("翻译失败", zap.Error(err))
						return
					}

					// 获取采购单位
					var purchaseUom string
					purchaseUnit, err := repository.NewMaterialUnitRepo(db).GetMaterialUnitsByUuid(material.PurchaseUnitUuid)
					if err != nil {
						ctx.Log().Error("获取采购单位失败", zap.Error(err))
						return
					}
					if purchaseUnit.Unit != nil {
						purchaseUom = purchaseUnit.Unit.ErpnextUom
					}

					erpSrv.AddMaterial(ctx, req.MaterialAddErpReq{
						ItemCode:       material.Code,
						ItemName:       enName,
						StockUom:       material.Unit.Unit.ErpnextUom,
						Disabled:       material.Status == false,
						ValuationRate:  material.Valuation,
						BarcodeValue:   material.BarcodeValue,
						Uoms:           unitList,
						InternalCode:   material.InternalCode,
						Classification: getMaterialCategoryName,
						ClassificationCode: func() string {
							if materialCategory.Code != "" {
								return materialCategory.Code
							}
							return " " // 分类编码为空时，传空格给ErpNext
						}(),
						PurchaseUom: purchaseUom,
					})
				}
			}
		}
	})
	return nil
}

func (s *materialSrv) GetMaterialUnitList(ctx context.Context, req req.MaterialUnitListReq) (material_resp.MaterialUnitListResp, error) {
	dbId := ctx.GetDbId()
	materialRepo := repository.NewMaterialRepo(s.dbm.GetDB(dbId))
	material, err := materialRepo.GetMaterialDetailByUuid(req.Uuid)
	if err != nil {
		return material_resp.MaterialUnitListResp{}, errors.WithMessage(err, "获取物品失败")
	}

	var materialUnitListResp []material_resp.MaterialUnit
	for _, materialUnit := range material.NotBaseUnitList {
		materialUnitListResp = append(materialUnitListResp, material_resp.MaterialUnit{
			Uuid:           materialUnit.Uuid,
			Name:           materialUnit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
			LocaleName:     materialUnit.Unit.MultiLanguageName.GetNames(),
			ConversionRate: materialUnit.ConversionRate,
		})
	}

	return material_resp.MaterialUnitListResp{
		List: materialUnitListResp,
	}, nil
}

func (s *materialSrv) DeleteMaterialCategory(ctx context.Context, req req.MaterialCategoryDeleteReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	materialCategoryRepo := repository.NewMaterialRepo(db)
	materialCategory, err := materialCategoryRepo.GetMaterialCategoryByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "获取物品类别失败")
	}
	// 检查物品类别是否关联了物品
	materialRepo := repository.NewMaterialRepo(db)
	materials, err := materialRepo.GetMaterialByCategoryUuid(materialCategory.Uuid)
	if err != nil {
		return errors.WithMessage(err, "获取物品失败")
	}
	if len(materials) > 0 {
		return errors.New("该类别已经关联了物品，不可删除")
	}
	if err := materialCategoryRepo.DeleteMaterialCategory(materialCategory.Uuid, materialCategory.MultiLanguageNameUuid); err != nil {
		return errors.WithMessage(err, "删除物品类别失败")
	}
	return nil
}

func (s *materialSrv) AddProductBomCard(ctx context.Context, req req.ProductBomCardAddReq) error {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	ctx.SetDB(db)
	if len(req.Materials.List) == 0 {
		return errors.New("物品列表不能为空")
	}

	if req.RelatedType == constant.ProductBomCardRelatedTypeFlavor {
		return s.addProductBomCard(ctx, req)
	} else if req.RelatedType == constant.ProductBomCardRelatedTypeSauce {
		return s.addSauceBomCard(ctx, req)
	}
	return errors.New("关联类型错误")

}

// 创建成本卡(同步erp数据)
func (s *materialSrv) AddProductBomCardForSync(ctx context.Context, productBomCard model.ProductBomCard) error {
	// db := ctx.GetDB()
	// if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
	// 	repository.NewProductBomCardRepo(tx).CreateProductBomCard(productBomCard)
	// }); err != nil {
	// 	return errors.WithMessage(err)
	// }
	return nil
}

func (s *materialSrv) addSauceBomCard(ctx context.Context, req req.ProductBomCardAddReq) error {
	db := ctx.GetDB()
	// 获取小料名称
	sauceRepo := repository.NewProductSauceRepo(db)
	sauce, err := sauceRepo.GetSauceByUuid(req.RelatedUuid)
	if err != nil {
		return errors.WithMessage(err, "获取小料失败")
	}

	if sauce.HeadquarterUuid != 0 {
		return errors.New("无法为总部小料添加成本卡")
	}

	productBomCardName := sauce.MultiLanguageName.GetNames()

	nameUuid, _ := utils.GetID()
	cardUuid, _ := utils.GetID()
	multiLanguageName := model.MultiLanguageName{}
	multiLanguageName.InitByLocaleResponse(productBomCardName)
	multiLanguageName.Uuid = nameUuid
	_, errName := repository.NewMultiLanguageNameRepo(db).CreateMultiLanguageName(multiLanguageName)
	if errName != nil {
		return errors.WithMessage(err, "创建多语言名称失败")
	}

	materialList := []*model.RelatedMaterial{}
	for _, materialParam := range req.Materials.List {
		// 查询物品信息
		material, err := repository.NewMaterialRepo(db).GetMaterialDetailByUuid(materialParam.MaterialUuid)
		if err != nil {
			return errors.WithMessage(err, "获取物品信息失败")
		}
		if !material.IsUnit(materialParam.UnitUuid) {
			return errors.New("物品没有该单位")
		}
		// 获取成本单位信息
		materialUnit := material.GetUnit(materialParam.UnitUuid)
		baseUnit := material.GetBaseUnit()

		materialList = append(materialList, &model.RelatedMaterial{
			RelatedUuid:            cardUuid,
			MaterialUuid:           materialParam.MaterialUuid,
			Num:                    materialParam.Num,
			UnitUuid:               materialParam.UnitUuid,
			UnitName:               materialUnit.Unit.MultiLanguageName.ToJson(),
			UnitUom:                materialUnit.Unit.ErpnextUom,
			BaseUnitUuid:           baseUnit.Uuid,
			BaseUnitName:           baseUnit.Unit.MultiLanguageName.ToJson(),
			BaseUnitUom:            baseUnit.Unit.ErpnextUom,
			BaseUnitConversionRate: materialUnit.ConversionRate,
			IsUsed:                 1, // 成本卡被使用
			Material:               material,
		})
	}

	productBomCard := model.ProductBomCard{
		BaseModel: model.BaseModel{
			Uuid: cardUuid,
		},
		Name:                  multiLanguageName.ToJson(),
		MultiLanguageNameUuid: nameUuid,
		Num:                   float64(req.Num),
		IsUsed:                1, // 成本卡被使用
		RelatedMaterials:      materialList,
	}

	productBomCardLog, err := newProductBomCardLog(ctx, float64(req.Num), cardUuid, multiLanguageName.ToJson(),
		req.RelatedUuid, sauce.MultiLanguageName.ToJson(), materialList, constant.ProductBomCardLogOperationTypeCreate)
	if err != nil {
		return errors.WithMessage(err, "创建成本卡日志失败")
	}

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 修改旧的成本卡为未使用
		if err := repository.NewProductBomCardRepo(tx).UpdateProductBomCardIsUsed(sauce.ProductBomCardUuid, 0); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
		// 创建新成本卡
		productBomCardRepo := repository.NewProductBomCardRepo(tx)
		if err := productBomCardRepo.CreateProductBomCard(productBomCard); err != nil {
			return errors.WithMessage(err, "创建成本卡失败")
		}
		for _, material := range materialList {
			if err := productBomCardRepo.CreateProductBomCardMaterial(*material); err != nil {
				return errors.WithMessage(err, "创建成本卡材料失败")
			}
		}
		if err := repository.NewProductSauceRepo(tx).UpdateProductBomCard(req.RelatedUuid, cardUuid); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
		if err := repository.NewProductBomCardLogRepo(tx).CreateProductBomCardLog(*productBomCardLog); err != nil {
			return errors.WithMessage(err, "创建成本卡日志失败")
		}
		if ctx.GetCompany().IsOpenErp() {
			// 同步成本卡到erp
			erpBomItemList := []*manufacturing.BomItem{}
			for _, material := range materialList {
				unitName := model.NewMultiLanguageName(material.UnitName)
				erpnextUom := material.GetUnitErpnextUom()
				// 兼容开发阶段产生的脏数据
				if erpnextUom == "" {
					enName, err := GetEnName(ctx, s.settingSrv, unitName.GetNames())
					if err != nil {
						return errors.WithMessage(err, "翻译失败")
					}
					erpnextUom = enName
				}
				erpBomItemList = append(erpBomItemList, &manufacturing.BomItem{
					ItemCode: material.Material.Code,
					Rate:     material.Material.Valuation,
					Qty:      material.Num,
					Uom:      erpnextUom,
				})
			}
			erpSrv := erp.NewIErpSrv(s.dbm)
			erpBomResp, errErp := erpSrv.AddProductBomCard(ctx, erp.ProductBomCardAddErpReq{
				ItemCode: sauce.ErpCode,    // 商品编码
				Quantity: float64(req.Num), // 数量
				Uom:      "Nos",            // 单位,小料的单位固定为Nos
				Items:    erpBomItemList,
			})
			if errErp != nil {
				if strings.Contains(errErp.Error(), "Must be Whole Number") {
					return errors.WithMessage(errors.New("请输入正整数"), errErp.Error())
				}
				return errors.WithMessage(errErp)
			}
			productBomCard.ErpCode = erpBomResp.BomName // 记录erp成本卡编码

			// 更新成本卡ErpCode
			if err := productBomCardRepo.UpdateProductBomCardErpCode(cardUuid, erpBomResp.BomName); err != nil {
				return errors.WithMessage(err, "更新成本卡ErpCode失败")
			}
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

func (s *materialSrv) addProductBomCard(ctx context.Context, req req.ProductBomCardAddReq) error {
	db := ctx.GetDB()
	// 获取商品名称
	productBomRepo := repository.NewProductBomRepo(db)
	productBom, err := productBomRepo.GetProductBom(
		repository.CommonRepo.WhereByUuid(req.RelatedUuid),
		repository.CommonRepo.Preload(
			repository.WithPreload{
				Query: "ProductPackage.MultiLanguageName",
			},
			repository.WithPreload{
				Query: "ProductPackage.ProductUnit.MultiLanguageName",
			},
		),
	)
	if err != nil {
		return errors.WithMessage(err, "获取商品规格失败")
	}
	productBomCardName := productBom.ProductPackage.MultiLanguageName.GetNames()

	if productBom.ProductPackage.HeadquarterUuid != 0 {
		return errors.New("无法为总部商品添加成本卡")
	}

	nameUuid, _ := utils.GetID()
	cardUuid, _ := utils.GetID()
	multiLanguageName := model.MultiLanguageName{}
	multiLanguageName.InitByLocaleResponse(productBomCardName)
	multiLanguageName.Uuid = nameUuid
	_, errName := repository.NewMultiLanguageNameRepo(db).CreateMultiLanguageName(multiLanguageName)
	if errName != nil {
		return errors.WithMessage(err, "创建多语言名称失败")
	}

	materialList := []*model.RelatedMaterial{}
	for _, materialParam := range req.Materials.List {
		// 查询物品信息
		material, err := repository.NewMaterialRepo(db).GetMaterialDetailByUuid(materialParam.MaterialUuid)
		if err != nil {
			return errors.WithMessage(err, "获取物品信息失败")
		}
		if !material.IsUnit(materialParam.UnitUuid) {
			return errors.New("物品没有该单位")
		}
		// 获取成本单位信息
		materialUnit := material.GetUnit(materialParam.UnitUuid)
		baseUnit := material.GetBaseUnit()

		item := &model.RelatedMaterial{
			RelatedUuid:            cardUuid,
			MaterialUuid:           materialParam.MaterialUuid,
			Num:                    materialParam.Num,
			UnitUuid:               materialParam.UnitUuid,
			UnitName:               materialUnit.Unit.MultiLanguageName.ToJson(),
			UnitUom:                materialUnit.Unit.ErpnextUom,
			BaseUnitUuid:           baseUnit.Uuid,
			BaseUnitName:           baseUnit.Unit.MultiLanguageName.ToJson(),
			BaseUnitUom:            baseUnit.Unit.ErpnextUom,
			BaseUnitConversionRate: materialUnit.ConversionRate,
			Material:               material,
			IsUsed:                 1, // 成本卡被使用
		}
		item.SetUnitErpnextUom(materialUnit.Unit.ErpnextUom)
		item.SetExpectedProductionNum(item.CalculateExpectedProductionNum())
		materialList = append(materialList, item)
	}

	productBomCard := model.ProductBomCard{
		BaseModel: model.BaseModel{
			Uuid: cardUuid,
		},
		Name:                  multiLanguageName.ToJson(),
		MultiLanguageNameUuid: nameUuid,
		Num:                   float64(req.Num),
		IsUsed:                1, // 成本卡被使用
		RelatedMaterials:      materialList,
	}

	productBomCardLog, err := newProductBomCardLog(ctx, float64(req.Num), cardUuid, multiLanguageName.ToJson(),
		req.RelatedUuid, productBom.ProductPackage.MultiLanguageName.ToJson(), materialList, constant.ProductBomCardLogOperationTypeCreate)
	if err != nil {
		return errors.WithMessage(err, "创建成本卡日志失败")
	}

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 修改旧的成本卡为未使用
		if err := repository.NewProductBomCardRepo(tx).UpdateProductBomCardIsUsed(productBom.ProductBomCardUuid, 0); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
		// 创建新成本卡
		productBomCardRepo := repository.NewProductBomCardRepo(tx)
		if err := productBomCardRepo.CreateProductBomCard(productBomCard); err != nil {
			return errors.WithMessage(err, "创建成本卡失败")
		}
		for _, material := range materialList {
			if err := productBomCardRepo.CreateProductBomCardMaterial(*material); err != nil {
				return errors.WithMessage(err, "创建成本卡材料失败")
			}
		}
		stockNum := productBomCard.CalculateExpectedProductionNum()
		if err := repository.NewProductBomRepo(tx).UpdateProductBomCard(req.RelatedUuid, cardUuid, stockNum); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
		if err := repository.NewProductBomCardLogRepo(tx).CreateProductBomCardLog(*productBomCardLog); err != nil {
			return errors.WithMessage(err, "创建成本卡日志失败")
		}

		if ctx.GetCompany().IsOpenErp() {
			// 同步成本卡到erp
			erpBomItemList := []*manufacturing.BomItem{}
			for _, material := range materialList {
				unitName := model.NewMultiLanguageName(material.UnitName)
				erpnextUom := material.GetUnitErpnextUom()
				// 兼容开发阶段产生的脏数据
				if erpnextUom == "" {
					enName, err := GetEnName(ctx, s.settingSrv, unitName.GetNames())
					if err != nil {
						return errors.WithMessage(err, "翻译失败")
					}
					erpnextUom = enName
				}
				erpBomItemList = append(erpBomItemList, &manufacturing.BomItem{
					ItemCode: material.Material.Code,
					Rate:     material.Material.Valuation,
					Qty:      material.Num,
					Uom:      erpnextUom,
				})
			}
			erpSrv := erp.NewIErpSrv(s.dbm)
			erpBomResp, errErp := erpSrv.AddProductBomCard(ctx, erp.ProductBomCardAddErpReq{
				ItemCode: productBom.ErpCode,                               // 商品编码
				Quantity: float64(req.Num),                                 // 数量
				Uom:      productBom.ProductPackage.ProductUnit.ErpnextUom, // 单位
				Items:    erpBomItemList,
			})
			if errErp != nil {
				if strings.Contains(errErp.Error(), "Must be Whole Number") {
					return errors.WithMessage(errors.New("请输入正整数"), errErp.Error())
				}
				return errors.WithMessage(errErp)
			}
			productBomCard.ErpCode = erpBomResp.BomName // 记录erp成本卡编码

			// 更新成本卡ErpCode
			if err := productBomCardRepo.UpdateProductBomCardErpCode(cardUuid, erpBomResp.BomName); err != nil {
				return errors.WithMessage(err, "更新成本卡ErpCode失败")
			}
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

func newProductBomCardLog(ctx context.Context, num float64, cardUuid uint64, cardName string, relatedUuid uint64, relatedName string, materialList []*model.RelatedMaterial, operationType uint8) (*model.ProductBomCardLog, error) {
	var data string

	// 新增成本卡
	if cardUuid == 0 {
		productBomCardLogDataItemList := []*model.MaterialItem{}
		for _, material := range materialList {
			productBomCardLogDataItemList = append(productBomCardLogDataItemList, &model.MaterialItem{
				MaterialName:       material.Material.MultiLanguageName.ToJson(),
				MaterialCode:       material.Material.Code,
				Num:                material.Num,
				UnitUuid:           material.UnitUuid,
				UnitName:           material.UnitName,
				BaseUnitUuid:       material.BaseUnitUuid,
				BaseUnitName:       material.BaseUnitName,
				UnitConversionRate: material.BaseUnitConversionRate,
			})
		}
		productBomCardLogData := model.ProductBomCardLogData{
			Num:              num,
			RelatedMaterials: productBomCardLogDataItemList,
		}
		data = utils.ToJson(productBomCardLogData)
	} else {
		// 删除成本卡关联时，data为空
		data = ""
	}

	productBomCardLog := model.ProductBomCardLog{
		ProductBomCardUuid: cardUuid,
		ProductBomCardName: cardName,
		RelatedUuid:        relatedUuid,
		RelatedName:        relatedName,
		Data:               data,
		StaffUuid:          ctx.GetStaffUuid(),
		OperationType:      operationType,
	}

	return &productBomCardLog, nil
}

func (s *materialSrv) GetProductBomCardDetail(ctx context.Context, req req.ProductBomCardDetailReq) (*material_resp.ProductBomCardDetailResp, error) {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)

	productBomCardRepo := repository.NewProductBomCardRepo(db)
	productBomCard, err := productBomCardRepo.GetProductBomCardDetail(req.Uuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取成本卡失败")
	}

	materialList := []material_resp.ProductBomCardMaterial{}
	for _, material := range productBomCard.RelatedMaterials {
		unitList := []material_resp.MaterialUnit{}
		for _, unit := range material.Material.NotBaseUnitList {
			unitList = append(unitList, material_resp.MaterialUnit{
				Uuid:           unit.Uuid,
				Name:           unit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
				LocaleName:     unit.Unit.MultiLanguageName.GetNames(),
				ConversionRate: unit.ConversionRate,
			})
		}
		materialList = append(materialList, material_resp.ProductBomCardMaterial{
			Material: material_resp.MaterialInfo{
				Uuid:         material.MaterialUuid,
				Name:         material.Material.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
				Code:         material.Material.Code,
				InternalCode: material.Material.InternalCode,
			},
			Num: material.Num,
			Unit: material_resp.MaterialUnit{
				Uuid:           material.UnitUuid,
				Name:           material.GetUnitName(ctx.GetLanguage()),
				LocaleName:     material.GetUnitLocaleName(),
				ConversionRate: material.BaseUnitConversionRate,
			},
			UnitList: unitList,
		})
	}

	return &material_resp.ProductBomCardDetailResp{
		Uuid:       productBomCard.Uuid,
		Name:       productBomCard.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
		LocaleName: productBomCard.MultiLanguageName.GetNames(),
		Num:        productBomCard.Num,
		Materials:  materialList,
		IsEditable: !productBomCard.IsHeadquarter(),
	}, nil
}

func (s *materialSrv) UnlinkProductBomCard(ctx context.Context, req req.ProductBomCardUnlinkReq) error {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	ctx.SetDB(db)
	if req.RelatedType == constant.ProductBomCardRelatedTypeFlavor {
		return s.unlinkProductBomCard(ctx, req)
	} else if req.RelatedType == constant.ProductBomCardRelatedTypeSauce {
		return s.unlinkSauceBomCard(ctx, req)
	}
	return errors.New("关联类型错误")

}

func (s *materialSrv) unlinkSauceBomCard(ctx context.Context, req req.ProductBomCardUnlinkReq) error {
	db := ctx.GetDB()
	sauceRepo := repository.NewProductSauceRepo(db)
	sauce, err := sauceRepo.GetSauceByUuid(req.RelatedUuid)
	if err != nil {
		return errors.WithMessage(err, "获取小料失败")
	}

	if sauce.HeadquarterUuid != 0 {
		return errors.New("无法删除总部小料的成本卡")
	}

	productBomCardLog, err := newProductBomCardLog(ctx, 0, 0, "", req.RelatedUuid, sauce.MultiLanguageName.ToJson(), nil, constant.ProductBomCardLogOperationTypeDelete)
	if err != nil {
		return errors.WithMessage(err, "创建成本卡日志失败")
	}

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 修改旧的成本卡为未使用
		if err := repository.NewProductBomCardRepo(tx).UpdateProductBomCardIsUsed(sauce.ProductBomCardUuid, 0); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
		// 解除成本卡关联
		productBomCardRepo := repository.NewProductSauceRepo(tx)
		if err := productBomCardRepo.UpdateProductBomCard(req.RelatedUuid, 0); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
		if err := repository.NewProductBomCardLogRepo(tx).CreateProductBomCardLog(*productBomCardLog); err != nil {
			return errors.WithMessage(err, "创建成本卡日志失败")
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

func (s *materialSrv) unlinkProductBomCard(ctx context.Context, req req.ProductBomCardUnlinkReq) error {
	db := ctx.GetDB()
	productBomRepo := repository.NewProductBomRepo(db)
	productBom, err := productBomRepo.GetFlavorProductBomByUuid(req.RelatedUuid)
	if err != nil {
		return errors.WithMessage(err, "获取商品规格失败")
	}

	if productBom.ProductPackage.HeadquarterUuid != 0 {
		return errors.New("无法删除总部商品的成本卡")
	}

	productBomCardLog, err := newProductBomCardLog(ctx, 0, 0, "", req.RelatedUuid, productBom.ProductPackage.MultiLanguageName.ToJson(), nil, constant.ProductBomCardLogOperationTypeDelete)
	if err != nil {
		return errors.WithMessage(err, "创建成本卡日志失败")
	}

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 修改旧的成本卡为未使用
		if err := repository.NewProductBomCardRepo(tx).UpdateProductBomCardIsUsed(productBom.ProductBomCardUuid, 0); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
		// 解除成本卡关联
		productBomCardRepo := repository.NewProductBomRepo(tx)
		if err := productBomCardRepo.UpdateProductBomCard(req.RelatedUuid, 0, 999999999); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
		if err := repository.NewProductBomCardLogRepo(tx).CreateProductBomCardLog(*productBomCardLog); err != nil {
			return errors.WithMessage(err, "创建成本卡日志失败")
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

func (s *materialSrv) CopyProductBomCard(ctx context.Context, req req.ProductBomCardCopyReq) error {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	ctx.SetDB(db)

	productBomCardRepo := repository.NewProductBomCardRepo(db)
	productBomCard, err := productBomCardRepo.GetProductBomCardDetail(req.ProductBomCardUuid)
	if err != nil {
		return errors.WithMessage(err, "获取成本卡失败")
	}

	copyProductBomCard := productBomCard.Copy()
	stockNum := copyProductBomCard.CalculateExpectedProductionNum()

	relatedName := ""
	if req.RelatedType == constant.ProductBomCardRelatedTypeFlavor {
		productBomRepo := repository.NewProductBomRepo(db)
		productBom, err := productBomRepo.GetFlavorProductBomByUuid(req.RelatedUuid)
		if err != nil {
			return errors.WithMessage(err, "获取商品规格失败")
		}
		relatedName = productBom.ProductPackage.MultiLanguageName.ToJson()
	} else if req.RelatedType == constant.ProductBomCardRelatedTypeSauce {
		productSauceRepo := repository.NewProductSauceRepo(db)
		productSauce, err := productSauceRepo.GetSauceByUuid(req.RelatedUuid)
		if err != nil {
			return errors.WithMessage(err, "获取小料失败")
		}
		relatedName = productSauce.MultiLanguageName.ToJson()
	} else {
		return errors.New("关联类型错误")
	}

	productBomCardLog, err := newProductBomCardLog(ctx, copyProductBomCard.Num, copyProductBomCard.Uuid, copyProductBomCard.MultiLanguageName.ToJson(), req.RelatedUuid, relatedName, copyProductBomCard.RelatedMaterials, constant.ProductBomCardLogOperationTypeCreate)
	if err != nil {
		return errors.WithMessage(err, "创建成本卡日志失败")
	}

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		productBomCardRepo := repository.NewProductBomCardRepo(tx)
		if err := productBomCardRepo.CreateProductBomCard(*copyProductBomCard); err != nil {
			return errors.WithMessage(err, "创建成本卡失败")
		}
		for _, material := range copyProductBomCard.RelatedMaterials {
			if err := productBomCardRepo.CreateProductBomCardMaterial(*material); err != nil {
				return errors.WithMessage(err, "创建成本卡材料失败")
			}
		}
		if _, err := repository.NewMultiLanguageNameRepo(tx).CreateMultiLanguageName(*copyProductBomCard.MultiLanguageName); err != nil {
			return errors.WithMessage(err, "创建多语言名称失败")
		}
		// 更新成本卡关联
		if req.RelatedType == constant.ProductBomCardRelatedTypeFlavor {
			if err := repository.NewProductBomRepo(tx).UpdateProductBomCard(req.RelatedUuid, copyProductBomCard.Uuid, stockNum); err != nil {
				return errors.WithMessage(err, "更新成本卡失败")
			}
		} else if req.RelatedType == constant.ProductBomCardRelatedTypeSauce {
			if err := repository.NewProductSauceRepo(tx).UpdateProductBomCard(req.RelatedUuid, copyProductBomCard.Uuid); err != nil {
				return errors.WithMessage(err, "更新成本卡失败")
			}
		}
		if err := repository.NewProductBomCardLogRepo(tx).CreateProductBomCardLog(*productBomCardLog); err != nil {
			return errors.WithMessage(err, "创建成本卡日志失败")
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

// ImportProductBomCard 从菜品导入成本卡. 先创建物品，再创建成本卡，再关联成本卡到规格商品
func (s *materialSrv) ImportProductBomCard(ctx context.Context, req req.ProductBomCardImportReq) error {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)
	ctx.SetDB(db)

	// 检查物品名称
	productCheckSrv := NewProductCheckSrv(s.dbm, s.localeSrv, s.settingSrv)
	if err := productCheckSrv.CheckMaterialName(ctx, 0, req.LocaleName); err != nil {
		return errors.WithMessage(err, "检查物品名称失败")
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 创建物品
		material, materialAddErpReq, err := addMaterial(ctx, db, s.settingSrv, req.MaterialAddReq)
		if err != nil {
			return errors.WithMessage(err)
		}
		code := ""
		if ctx.GetCompany().IsOpenErp() {
			erpSrv := erp.NewIErpSrv(s.dbm)
			itemInfo, errErp := erpSrv.AddMaterial(ctx, *materialAddErpReq)
			if errErp != nil {
				return errors.WithMessage(err)
			}
			code = itemInfo.ItemCode
		} else {
			uuid, _ := utils.GetID()
			code = fmt.Sprintf("WPR%d", uuid)
		}
		materialRepo := repository.NewMaterialRepo(db)
		err = materialRepo.UpdateMaterialCode(material.Uuid, code)
		if err != nil {
			return errors.WithMessage(err, "更新物品编码失败")
		}

		// 获取商品名称
		productBom, err := repository.NewProductBomRepo(db).GetProductBom(
			repository.CommonRepo.WhereByUuid(req.RelatedUuid),
			repository.CommonRepo.Preload(
				repository.WithPreload{
					Query: "ProductPackage.MultiLanguageName",
				},
			),
		)
		if err != nil {
			return errors.WithMessage(err, "获取商品规格失败")
		}
		productBomCardName := productBom.ProductPackage.MultiLanguageName.GetNames()

		// 创建成本卡
		productBomCardUuid, _ := utils.GetID()
		nameUuid, _ := utils.GetID()
		multiLanguageName := model.MultiLanguageName{}
		multiLanguageName.InitByLocaleResponse(productBomCardName)
		multiLanguageName.Uuid = nameUuid
		_, errName := repository.NewMultiLanguageNameRepo(db).CreateMultiLanguageName(multiLanguageName)
		if errName != nil {
			return errors.WithMessage(err, "创建多语言名称失败")
		}
		materialList := []*model.RelatedMaterial{}
		item := &model.RelatedMaterial{
			RelatedUuid:            productBomCardUuid,
			MaterialUuid:           material.Uuid,
			Num:                    req.Num,
			UnitUuid:               material.UnitUuid,
			UnitName:               material.Unit.Name,
			UnitUom:                material.Unit.Unit.ErpnextUom,
			BaseUnitUuid:           material.UnitUuid,
			BaseUnitName:           material.Unit.Name,
			BaseUnitUom:            material.Unit.Unit.ErpnextUom,
			BaseUnitConversionRate: 1,
			Material:               material,
			IsUsed:                 1, // 成本卡被使用
		}
		item.SetUnitErpnextUom(material.Unit.Unit.ErpnextUom)
		item.SetExpectedProductionNum(item.CalculateExpectedProductionNum()) // 计算预计可生产的产品数量
		materialList = append(materialList, item)

		productBomCard := model.ProductBomCard{
			BaseModel: model.BaseModel{
				Uuid: productBomCardUuid,
			},
			Name:                  multiLanguageName.ToJson(),
			MultiLanguageNameUuid: nameUuid,
			Num:                   1, // 加工份数,目前成本卡的加工份数固定为1
			IsUsed:                1, // 成本卡被使用
			RelatedMaterials:      materialList,
		}

		productBomCardLog, err := newProductBomCardLog(ctx, 1, productBomCardUuid, multiLanguageName.ToJson(),
			req.RelatedUuid, req.MaterialAddReq.LocaleName.ToJson(), materialList, constant.ProductBomCardLogOperationTypeCreate)
		if err != nil {
			return errors.WithMessage(err, "创建成本卡日志失败")
		}

		if ctx.GetCompany().IsOpenErp() {
			// 同步成本卡到erp
			erpBomItemList := []*manufacturing.BomItem{}
			for _, material := range materialList {
				erpnextUom := material.GetUnitErpnextUom()
				if erpnextUom == "" {
					unitName := model.NewMultiLanguageName(material.UnitName)
					enName, err := GetEnName(ctx, s.settingSrv, unitName.GetNames())
					if err != nil {
						return errors.WithMessage(err, "翻译失败")
					}
					erpnextUom = enName
				}
				erpBomItemList = append(erpBomItemList, &manufacturing.BomItem{
					ItemCode: code,
					Rate:     material.Material.Valuation,
					Qty:      material.Num,
					Uom:      erpnextUom,
				})
			}
			productBomRepo := repository.NewProductBomRepo(db)
			productBom, err := productBomRepo.GetFlavorProductBomByUuid(req.RelatedUuid)
			if err != nil {
				return errors.WithMessage(err, "获取商品规格失败")
			}
			uom := productBom.ProductPackage.ProductUnit.ErpnextUom
			erpSrv := erp.NewIErpSrv(s.dbm)
			erpBomResp, errErp := erpSrv.AddProductBomCard(ctx, erp.ProductBomCardAddErpReq{
				ItemCode: productBom.ErpCode,
				Quantity: 1,
				Uom:      uom,
				Items:    erpBomItemList,
			})
			if errErp != nil {
				return errors.WithMessage(errErp)
			}
			productBomCard.ErpCode = erpBomResp.BomName // 记录erp成本卡编码
		}

		productBomCardRepo := repository.NewProductBomCardRepo(db)
		if err := productBomCardRepo.CreateProductBomCard(productBomCard); err != nil {
			return errors.WithMessage(err, "创建成本卡失败")
		}
		for _, material := range materialList {
			if err := productBomCardRepo.CreateProductBomCardMaterial(*material); err != nil {
				return errors.WithMessage(err, "创建成本卡材料失败")
			}
		}
		if err := repository.NewProductBomCardLogRepo(db).CreateProductBomCardLog(*productBomCardLog); err != nil {
			return errors.WithMessage(err, "创建成本卡日志失败")
		}
		stockNum := productBomCard.CalculateExpectedProductionNum()
		if err := repository.NewProductBomRepo(db).UpdateProductBomCard(req.RelatedUuid, productBomCardUuid, stockNum); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}

		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}
	return nil
}

// ImportMaterialList 导入物品列表
func (s *materialSrv) ImportMaterialList(ctx context.Context, reqs req.MaterialImportListReq) (material_resp.MaterialImportResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 初始化返回
	var materialImportResp material_resp.MaterialImportResp
	materialImportResp.List = make([]material_resp.MaterialImportListItem, 0, len(reqs.List))
	for _, item := range reqs.List {
		// 复制物品信息
		material := material_resp.MaterialImportListItem{}
		copier.Copy(&material, item)
		// 获取分类ID
		categoryUuid := uint64(0)
		if categoryName := strings.TrimSpace(item.CategoryName); categoryName != "" {
			categoryUuidTmp, err := repository.NewMaterialRepo(db).GetCategoryUuidByNameOptimized(categoryName)
			if err != nil {
				return material_resp.MaterialImportResp{}, err
			}
			categoryUuid = categoryUuidTmp
		}
		// 获取单位ID
		unitUuid := uint64(0)
		if unitName := strings.TrimSpace(item.UnitName); unitName != "" {
			unitUuidTmp, err := base.NewProductUnitRepo(db).GetProductUnitUuidByNameOptimized(unitName)
			if err != nil {
				return material_resp.MaterialImportResp{}, err
			}
			unitUuid = unitUuidTmp
		}
		// 处理条码：过滤空格、非数字字符，截取13位
		material.BarcodeValue = utils.ProcessBarcode(material.BarcodeValue)
		// 设置分类ID、单位ID
		material.CategoryUuid = categoryUuid
		material.UnitUuid = unitUuid
		// 验证是否已经存在
		material.LocaleNameIsExist = repository.NewMaterialRepo(db).CheckMultiLanguageNameExist(item.LocaleName)
		material.BarcodeIsExist = repository.NewMaterialRepo(db).CheckBarcodeExist(item.BarcodeValue, 0)
		// 添加到列表
		materialImportResp.List = append(materialImportResp.List, material)
	}

	// 获取分类列表
	materialCategories, err := s.GetMaterialCategoryList(ctx, req.MaterialCategoryListReq{})
	if err != nil {
		return material_resp.MaterialImportResp{}, errors.WithMessage(err, "获取分类列表失败")
	}
	materialImportResp.CategoryList = materialCategories.List

	// 获取单位列表
	unitList, err := base.NewProductUnitRepo(db).GetProductUnitList()
	if err != nil {
		return material_resp.MaterialImportResp{}, errors.WithMessage(err, "获取单位列表失败")
	}
	for _, unit := range unitList {
		materialImportResp.UnitList = append(materialImportResp.UnitList, material_resp.MaterialImportUnitListItem{
			Uuid:       unit.Uuid,
			LocaleName: unit.MultiLanguageName.GetNames(),
		})
	}

	return materialImportResp, nil
}

// ImportMaterial 导入物品
func (s *materialSrv) ImportMaterial(ctx context.Context, reqs req.MaterialImportReq) error {
	companyUuid := ctx.GetCompanyUuid()
	deviceSn := ctx.GetDeviceSn()
	db := s.dbm.GetDB(ctx.GetDbId())
	language := ctx.GetLanguage()
	// 生成锁的key
	lockKey := fmt.Sprintf("%d_v1_import_material", companyUuid)

	// 用信道锁 禁止并发导入 - 按公司UUID加锁确保同一公司的物品导入操作不会并发执行
	if !s.systemLock.TryLockUuidString(lockKey) {
		return nil
	}

	// 处理条码：过滤空格、非数字字符，截取13位
	for i := range reqs.List {
		reqs.List[i].BarcodeValue = utils.ProcessBarcode(reqs.List[i].BarcodeValue)
	}

	// 验证条形码是否重复
	barcodeDuplicateMap := make(map[string][]int)
	for _, item := range reqs.List {
		if item.BarcodeValue != "" {
			barcodeDuplicateMap[item.BarcodeValue] = append(barcodeDuplicateMap[item.BarcodeValue], item.Row)
		}
	}
	for barcode, rows := range barcodeDuplicateMap {
		if len(rows) > 1 {
			s.systemLock.UnlockUuidString(lockKey)
			return errors.New(i18n.Translate(language, "条码") + "[" + barcode + "]" + i18n.Translate(language, "在行") + fmt.Sprintf("%v", rows) + i18n.Translate(language, "中重复"))
		}
	}

	// 预验证阶段 - 检查物品名称和条形码是否已存在
	for _, item := range reqs.List {
		// 验证是否已经存在
		materialNameIsExist := repository.NewMaterialRepo(db).CheckMultiLanguageNameExist(item.LocaleName)
		if !materialNameIsExist.IsNull() {
			s.systemLock.UnlockUuidString(lockKey)
			return errors.New(i18n.Translate(language, "行") + "[" + strconv.Itoa(item.Row) + "]: " + i18n.Translate(language, "物品名称已存在"))
		}
		// 验证条形码存在性检查
		if repository.NewMaterialRepo(db).CheckBarcodeExist(item.BarcodeValue, 0) {
			s.systemLock.UnlockUuidString(lockKey)
			return errors.New(i18n.Translate(language, "行") + "[" + strconv.Itoa(item.Row) + "]: " + i18n.Translate(language, "物品条码已存在"))
		}
	}

	// 异步导入
	utils.Go(func() {
		defer s.systemLock.UnlockUuidString(lockKey)

		totalCount := len(reqs.List)
		progressData := MaterialImportProgressData{
			Time:    time.Now().Unix(),
			Status:  MaterialImportStatusStart,
			Total:   totalCount,
			Current: 0,
			Success: 0,
			Failed:  0,
			Errors:  make([]MaterialImportErrorDetail, 0),
		}

		// 推送开始导入进度
		time.Sleep(300 * time.Millisecond)
		s.pushMaterialImportProgress(companyUuid, deviceSn, progressData)

		totalItems := len(reqs.List)
		currentIndex := 0

		for _, item := range reqs.List {
			currentIndex++
			progressData.Current = currentIndex
			progressData.Progress = 30 + int(float64(currentIndex)/float64(totalItems)*70) // 导入占70%进度，从30%开始

			// 添加物品
			tmp := req.MaterialAddReq{
				LocaleName:       item.LocaleName,
				CategoryUuid:     item.CategoryUuid,
				UnitUuid:         item.UnitUuid,
				Status:           item.Status,
				Valuation:        item.Valuation,
				InitStock:        item.InitStock,
				BarcodeValue:     item.BarcodeValue,
				PurchaseUnitUuid: item.UnitUuid,
				CostUnitUuid:     item.UnitUuid,
			}
			err := s.AddMaterial(ctx, tmp)
			if err != nil {
				progressData.Failed++
				progressData.Errors = append(progressData.Errors, MaterialImportErrorDetail{
					Row:     item.Row,
					Message: err.Error(),
				})
				logger.Logger.Error("导入物品失败",
					zap.Int("row", item.Row),
					zap.Error(err),
					zap.Uint64("companyUuid", companyUuid))
			} else {
				progressData.Success++
			}

			// 推送导入进度
			if currentIndex%5 == 0 || currentIndex == totalItems { // 每5条或最后一条推送进度
				s.pushMaterialImportProgress(companyUuid, deviceSn, progressData)
			}
		}

		// 推送最终结果
		progressData.Progress = 100
		if progressData.Failed > 0 {
			progressData.Status = MaterialImportStatusError
			progressData.Error = fmt.Sprintf("导入完成，成功%d条，失败%d条", progressData.Success, progressData.Failed)
		} else {
			progressData.Status = MaterialImportStatusFinish
			progressData.Error = fmt.Sprintf("导入成功，共处理%d条物品", progressData.Success)
		}

		// 延迟500毫秒
		time.Sleep(500 * time.Millisecond)
		progressData.Time = time.Now().Unix()
		s.pushMaterialImportProgress(companyUuid, deviceSn, progressData)
	})

	return nil
}

// GetWarehouseItemsByErpCode 根据仓库ERP编码获取仓库商品库存列表
func (s *materialSrv) GetWarehouseItemsByErpCode(ctx context.Context, warehouseErpCode string, pageNo, pageSize int) ([]model.WarehouseItem, int64, error) {
	dbId := ctx.GetDbId()
	warehouseItemRepo := repository.NewWarehouseItemRepo(s.dbm.GetDB(dbId))

	// 构建查询选项
	var dbOptions []repository.DBOption
	commonRepo := repository.NewCommonRepo()

	// 添加仓库ERP编码过滤条件
	dbOptions = append(dbOptions, warehouseItemRepo.WhereWarehouseErpCode(warehouseErpCode))

	// 预加载仓库信息
	dbOptions = append(dbOptions, commonRepo.Preload(
		repository.WithPreload{
			Query: "Warehouse",
		},
		repository.WithPreload{
			Query: "Warehouse.MultiLanguageName",
		},
	))

	// 获取仓库商品库存列表
	warehouseItems, total, err := warehouseItemRepo.GetListWithWarehouseInfo(pageNo, pageSize, dbOptions...)
	if err != nil {
		return nil, 0, errors.WithMessage(err, "根据仓库ERP编码获取库存列表失败")
	}

	return warehouseItems, total, nil
}

// 同步物品分类
func (s *materialSrv) SyncMaterialCategory(ctx context.Context) error {
	companySetting := ctx.GetCompanySetting()
	if companySetting.IsSubShop() {
		// 获取总部company_uuid
		headquarterUuid := companySetting.HeadquarterUuid
		headquarterDb := s.dbm.GetDB(headquarterUuid)
		// 获取总部的分类
		headquarterMaterialCategoryList, err := repository.NewMaterialRepo(headquarterDb).GetMaterialCategoryList()
		if err != nil {
			return errors.WithMessage(err, "获取总部分类列表失败")
		}
		// 获取子公司的分类
		subShopDb := s.dbm.GetDB(ctx.GetCompanyUuid())
		subShopMaterialCategoryList, err := repository.NewMaterialRepo(subShopDb).GetMaterialCategoryList()
		if err != nil {
			return errors.WithMessage(err, "获取子公司分类列表失败")
		}

		headquarterMaterialCategoryMap := make(map[uint64]model.MaterialCategory)
		for _, category := range headquarterMaterialCategoryList {
			headquarterMaterialCategoryMap[category.Uuid] = category
		}

		// 获取子公司分类中的总部物品分类。 更新这些分类
		headquarterMaterialCategoryInSubShop := s.GetHeadquarterMaterialCategoryInSubShop(ctx, subShopMaterialCategoryList)

		// 获取不在子公司分类中的总部物品分类。新建这些分类
		headquarterMaterialCategoryNotInSubShop := s.GetHeadquarterMaterialCategoryNotInSubShop(ctx, headquarterMaterialCategoryList, subShopMaterialCategoryList)

		if err := repository.CommonRepo.Transaction(subShopDb, func(tx *gorm.DB) error {
			// 更新已存在的物品分类
			for _, category := range headquarterMaterialCategoryInSubShop {
				if headquarterCategory, ok := headquarterMaterialCategoryMap[category.Uuid]; ok {
					category.UpdateFromHeadquarter(headquarterCategory) // 用总部分类信息更新子公司分类
					if err := repository.NewMaterialRepo(tx).UpdateMaterialCategory(category); err != nil {
						return errors.WithMessage(err, "更新总部物品分类失败")
					}
				}
			}
			// 创建子店不存在的物品分类（多语言由 SyncMultiLanguage 任务处理）
			for _, category := range headquarterMaterialCategoryNotInSubShop {
				newCategory := model.MaterialCategory{
					BaseModel: model.BaseModel{
						Uuid:       category.Uuid,
						CreateTime: category.CreateTime,
						UpdateTime: category.UpdateTime,
						DeleteTime: category.DeleteTime,
					},
					Name:                  category.Name,
					MultiLanguageNameUuid: category.MultiLanguageNameUuid,
					Code:                  category.Code,
					Sort:                  category.Sort,
					HeadquarterUuid:       headquarterUuid,
				}
				if _, err := repository.NewMaterialRepo(tx).CreateMaterialCategory(newCategory); err != nil {
					return errors.WithMessage(err, "创建总部物品分类失败")
				}
			}
			return nil
		}); err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}

// 不在子公司分类中的总部物品分类
func (s *materialSrv) GetHeadquarterMaterialCategoryNotInSubShop(ctx context.Context, headquarterMaterialCategoryList []model.MaterialCategory, subShopMaterialCategoryList []model.MaterialCategory) []model.MaterialCategory {
	categoryList := []model.MaterialCategory{}
	subMaterialCategoryMap := make(map[uint64]model.MaterialCategory)
	for i, category := range subShopMaterialCategoryList {
		subMaterialCategoryMap[category.Uuid] = subShopMaterialCategoryList[i]
	}
	for i, category := range headquarterMaterialCategoryList {
		if _, ok := subMaterialCategoryMap[category.Uuid]; !ok {
			categoryList = append(categoryList, headquarterMaterialCategoryList[i])
		}
	}
	return categoryList
}

// 获取子公司分类中的总部物品分类
func (s *materialSrv) GetHeadquarterMaterialCategoryInSubShop(ctx context.Context, subShopMaterialCategoryList []model.MaterialCategory) []model.MaterialCategory {
	categoryList := []model.MaterialCategory{}
	for _, category := range subShopMaterialCategoryList {
		if category.IsHeadquarter() {
			categoryList = append(categoryList, category)
		}
	}
	return categoryList
}

// SyncMaterial 同步物品
func (s *materialSrv) SyncMaterial(ctx context.Context) error {
	companySetting := ctx.GetCompanySetting()

	// 从erp获取物品列表
	erpSrv := erp.NewIErpSrv(s.dbm)
	materialList, err := erpSrv.GetMaterialList(ctx, erp.GetMaterialListReq{
		SiteCode:        companySetting.ErpnextSiteCode,
		Branch:          companySetting.ErpnextBranchName,
		CompanyAbbr:     companySetting.ErpnextCompanyAbbr,
		ContainDisabled: true,
	})
	if err != nil {
		return errors.WithMessage(err, "获取物品列表失败")
	}

	var multiLanguageNameUuids []uint64

	db := ctx.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		commonRepo := repository.NewCommonRepo()
		materialRepo := repository.NewMaterialRepo(tx)
		multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
		copyCtx := ctx.Copy()
		copyCtx.SetDB(tx)
		for _, itemInfo := range materialList.ItemList {
			uoms := []req.MaterialUomReq{}
			for _, uom := range itemInfo.Uoms {
				uoms = append(uoms, req.MaterialUomReq{
					Uom:            uom.Uom,
					ConversionRate: uom.ConversionFactor,
				})
			}
			existingMaterial := materialRepo.GetMaterial(
				commonRepo.WhereByErpCode(itemInfo.ItemCode),
				materialRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
			)
			if existingMaterial.Uuid != 0 { // 如果物品已存在
				if err := s.UpdateMaterialByEprItem(copyCtx, req.MaterialEditErpReq{
					Uuid:               existingMaterial.Uuid,
					ItemCode:           itemInfo.ItemCode,
					ItemName:           itemInfo.ItemName,
					StockUom:           itemInfo.StockUom,
					Disabled:           itemInfo.Disabled,
					ValuationRate:      itemInfo.ValuationRate,
					OpeningStock:       itemInfo.OpeningStock,
					InternalCode:       itemInfo.InternalCode,
					Classification:     itemInfo.Classification,
					ClassificationCode: itemInfo.ClassificationCode,
					Uoms:               uoms,
					PurchaseUom:        itemInfo.PurchaseUom,
					NotForSale:         itemInfo.NotForSale,
				}); err != nil {
					logger.Logger.Error("同步erp物品列表失败-01", zap.Error(err))
				}
				// 添加多语言uuid到待翻译集合中
				if existingMaterial.MultiLanguageNameUuid == 0 || existingMaterial.MultiLanguageName.Uuid == 0 {
					newMultiLanguageName := model.MultiLanguageName{
						EnName:   itemInfo.ItemName,
						ZhName:   itemInfo.ItemName,
						ZhTwName: itemInfo.ItemName,
						ThName:   itemInfo.ItemName,
						MyName:   itemInfo.ItemName,
						JaName:   itemInfo.ItemName,
						KoName:   itemInfo.ItemName,
						TrName:   itemInfo.ItemName,
						SvName:   itemInfo.ItemName,
					}
					nameUuid, err := multiLanguageNameRepo.CreateMultiLanguageName(newMultiLanguageName)
					if err != nil {
						logger.Logger.Error("同步erp物品-更新物品-创建多语言失败", zap.String("itemCode", itemInfo.ItemCode), zap.Error(err))
					} else {
						err = materialRepo.UpdateMaterialData(map[string]any{"multi_language_name_uuid": nameUuid}, commonRepo.WhereByUuid(existingMaterial.Uuid))
						if err != nil {
							logger.Logger.Error("同步erp物品-更新物品-更新多语言UUID", zap.String("itemCode", itemInfo.ItemCode), zap.Error(err))
						} else {
							multiLanguageNameUuids = append(multiLanguageNameUuids, nameUuid)
						}
					}
				} else if existingMaterial.MultiLanguageName.Uuid != 0 {
					if existingMaterial.MultiLanguageName.EnName == "" {
						multiLanguageNameUuids = append(multiLanguageNameUuids, existingMaterial.MultiLanguageName.Uuid)
					}
				}
			} else {
				material, err := s.AddMaterialByEprItem(copyCtx, req.MaterialAddErpReq{
					ItemCode:           itemInfo.ItemCode,
					ItemName:           itemInfo.ItemName,
					Disabled:           itemInfo.Disabled,
					ValuationRate:      itemInfo.ValuationRate,
					OpeningStock:       itemInfo.OpeningStock,
					InternalCode:       itemInfo.InternalCode,
					Classification:     itemInfo.Classification,
					ClassificationCode: itemInfo.ClassificationCode,
					Uoms:               uoms,
					StockUom:           itemInfo.StockUom,
					PurchaseUom:        itemInfo.PurchaseUom,
					NotForSale:         itemInfo.NotForSale,
				})
				if err != nil {
					logger.Logger.Error("同步erp物品列表失败-02", zap.Error(err))
				} else {
					multiLanguageNameUuids = append(multiLanguageNameUuids, material.MultiLanguageNameUuid)
				}
			}
		}

		return nil
	})

	// 从总部同步物品到子店
	if companySetting.IsSubShop() {
		headquarterDb := s.dbm.GetDB(companySetting.HeadquarterUuid)
		commonRepo := repository.NewCommonRepo()
		materialRepo := repository.NewMaterialRepo(headquarterDb)
		headMaterialList := materialRepo.GetMaterialList(
			commonRepo.WhereByHeadquarterUuid(0),
			materialRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
			materialRepo.WithNotBaseUnitList(commonRepo.WhereBySoftDelete()),
		)
		db := ctx.GetDB()
		// 同步总部物品到子店（多语言由 SyncMultiLanguage 任务处理）
		err = db.Transaction(func(tx *gorm.DB) error {
			copyCtx := ctx.Copy()
			copyCtx.SetDB(tx)

			commonRepo := repository.NewCommonRepo()
			materialRepo := repository.NewMaterialRepo(tx)
			materialUnitRepo := repository.NewMaterialUnitRepo(tx)

			// 需要删除的物品、物品单位
			delMaterialUuidList := []uint64{}
			delMaterialUnitUuidList := []uint64{}
			subShopMaterialList := materialRepo.GetMaterialList(
				commonRepo.WhereByHeadquarterUuid(companySetting.HeadquarterUuid),
				materialRepo.WithMultiLanguageName(commonRepo.WhereBySoftDelete()),
				materialRepo.WithNotBaseUnitList(commonRepo.WhereBySoftDelete()),
			)
			for _, subShopMaterial := range subShopMaterialList {
				delMaterialUuidList = append(delMaterialUuidList, subShopMaterial.Uuid)
				for _, unit := range subShopMaterial.NotBaseUnitList {
					delMaterialUnitUuidList = append(delMaterialUnitUuidList, unit.Uuid)
				}
			}

			// 需要保存的物品、物品单位
			addMaterialList := []model.Material{}
			addMaterialUnitList := []model.MaterialUnit{}
			for _, material := range headMaterialList {
				addMaterialList = append(addMaterialList, model.Material{
					BaseModel:             model.BaseModel{Uuid: material.Uuid, CreateTime: material.CreateTime, UpdateTime: material.UpdateTime, DeleteTime: material.DeleteTime},
					Name:                  material.Name,
					Code:                  material.Code,
					Valuation:             material.Valuation,
					InitStock:             material.InitStock,
					MultiLanguageNameUuid: material.MultiLanguageNameUuid,
					CategoryUuid:          material.CategoryUuid,
					SupplierUuid:          material.SupplierUuid,
					ImageUuid:             material.ImageUuid,
					ImageName:             material.ImageName,
					UnitUuid:              material.UnitUuid,
					PurchaseUnitUuid:      material.PurchaseUnitUuid,
					CostUnitUuid:          material.CostUnitUuid,
					Price:                 material.Price,
					StockNum:              material.StockNum,
					SafetyStock:           material.SafetyStock, // 安全库存跟随总部
					ActualSaleNum:         material.ActualSaleNum,
					BarcodeValue:          material.BarcodeValue,
					InternalCode:          material.InternalCode,
					Status:                material.Status,
					HeadquarterUuid:       companySetting.HeadquarterUuid,
					WarehouseUuid:         material.WarehouseUuid,
					AllowSubstoreVisible:  material.AllowSubstoreVisible, // 同步可见性字段
				})
				for _, unit := range material.NotBaseUnitList {
					addMaterialUnitList = append(addMaterialUnitList, model.MaterialUnit{
						BaseModel:      model.BaseModel{Uuid: unit.Uuid, CreateTime: unit.CreateTime, UpdateTime: unit.UpdateTime, DeleteTime: unit.DeleteTime},
						Name:           unit.Name,
						UnitUuid:       unit.UnitUuid,
						ConversionRate: unit.ConversionRate,
						FromUnitUuid:   unit.FromUnitUuid,
						IsDefault:      unit.IsDefault,
						MaterialUuid:   unit.MaterialUuid,
					})
				}
			}
			if len(delMaterialUuidList) > 0 {
				err := materialRepo.DestroyMaterial(commonRepo.WhereInUuids(delMaterialUuidList))
				if err != nil {
					return errors.WithMessage(err, "删除总部物品列表失败")
				}
			}
			if len(delMaterialUnitUuidList) > 0 {
				err := materialRepo.DestroyMaterialUnit(commonRepo.WhereInUuids(delMaterialUnitUuidList))
				if err != nil {
					return errors.WithMessage(err, "删除总部物品单位列表失败")
				}
			}
			if len(addMaterialList) > 0 {
				err := materialRepo.CreateMaterialList(addMaterialList)
				if err != nil {
					return errors.WithMessage(err, "创建总部物品列表失败")
				}
			}
			if len(addMaterialUnitList) > 0 {
				err := materialUnitRepo.CreateMaterialUnitList(addMaterialUnitList)
				if err != nil {
					return errors.WithMessage(err, "创建总部物品单位列表失败")
				}
			}
			return nil
		})
	}

	// 添加多语言uuid到待翻译集合中
	if len(multiLanguageNameUuids) > 0 {
		err := s.translateSrv.AddMultiLanguageNameUuidToSet(ctx.GetCompanyUuid(), multiLanguageNameUuids...)
		if err != nil {
			logger.Logger.Error("同步erp物品列表添加多语言uuid到待翻译集合中失败", zap.Error(err))
		}
	}

	if err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

// 同步成本卡
func (s *materialSrv) SyncProductBomCard(ctx context.Context) error {
	companySetting := ctx.GetCompanySetting()
	erpBoms := []*manufacturing.BomInfo{} // erp成本卡列表
	erpSrv := erp.NewIErpSrv(s.dbm)
	productBomCardList, err := erpSrv.GetProductBomCardList(ctx) // erp成本卡列表,但这个接口没有物品列表信息
	if err != nil {
		return errors.WithMessage(err, "获取erp成本卡列表失败")
	}
	for _, bom := range productBomCardList.BomList {
		bomResp, err := erpSrv.GetProductBomCardDetail(ctx, req.ErpProductBomCardDetailReq{ // 单个成本卡的详情，包括了物品信息
			BomName: bom.BomName,
		})
		if err != nil {
			return errors.WithMessage(err, "获取erp成本卡详情失败")
		}
		erpBoms = append(erpBoms, bomResp.BomInfo)
	}

	erpBomsMap := make(map[string]*manufacturing.BomInfo)
	for _, bom := range erpBoms {
		erpBomsMap[bom.BomName] = bom
	}

	db := ctx.GetDB()
	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		// 获取ttpos的成本卡列表
		headquarterProductBomCardList, err := repository.NewProductBomCardRepo(tx).GetSubShopProductBomCardList(ctx.GetCompanySetting().HeadquarterUuid)
		if err != nil {
			return errors.WithMessage(err, "获取子店中来自总部的成本卡列表失败")
		}
		headquarterProductBomCardMap := make(map[string]*model.ProductBomCard)
		for i, bom := range headquarterProductBomCardList {
			key := bom.ErpCode
			headquarterProductBomCardMap[key] = headquarterProductBomCardList[i]
		}
		needCreateProductBomCardList := s.getNeedCreateProductBomCardList(headquarterProductBomCardMap, erpBoms)

		// 需要新建的成本卡列表。erpnext有，而ttpos没有时
		needCreate := needCreateProductBomCardList.CreateBoms // 这些成本卡来自两种场景：1. erpnext为还没有成本卡的商品添加成本卡；2. erpnext为已经添加成本卡的商品修改成本卡
		// 需要失效的成本卡列表。erpnext没有，而ttpos有时
		needDisable := needCreateProductBomCardList.DisableBoms // 这些成本卡来自1种场景：1. erpnext为已经添加成本卡的商品删除成本卡
		// 已经存在的成本卡列表。erpnext有，而ttpos也有
		existingProductBomCardList := needCreateProductBomCardList.ExistProductBomCardList
		for _, bomCard := range existingProductBomCardList {
			if bom, ok := erpBomsMap[bomCard.ErpCode]; ok {
				itemCode := bom.ItemCode                                              // 商品、小料的erpnext编码
				objectByItemCodeResp, err := s.getObjectByItemCode(ctx, itemCode, tx) // 获取物品或加料
				if err != nil {
					logger.Logger.Error("同步成本卡时，获取物品或加料失败", zap.String("bom_name", bomCard.Name), zap.Error(err), zap.Any("bom_card", bomCard), zap.Any("itemCode", itemCode))
					continue
				}
				if objectByItemCodeResp.RelatedType == constant.ProductBomCardRelatedTypeFlavor { // 规格商品
					// 更新product_bom表的成本卡uuid
					if err := tx.Model(&model.ProductBom{}).Where("uuid = ?", objectByItemCodeResp.ProductBom.Uuid).Update("product_bom_card_uuid", bomCard.Uuid).Error; err != nil {
						logger.Logger.Error("同步成本卡时，更新product_bom表的成本卡uuid失败", zap.String("bom_name", bomCard.Name), zap.Error(err), zap.Any("bom_card", bomCard))
						continue
					}
				} else if objectByItemCodeResp.RelatedType == constant.ProductBomCardRelatedTypeSauce { // 小料
					// 更新product_sauce表的成本卡uuid
					if err := tx.Model(&model.ProductSauce{}).Where("uuid = ?", objectByItemCodeResp.ProductSauce.Uuid).Update("product_bom_card_uuid", bomCard.Uuid).Error; err != nil {
						logger.Logger.Error("同步成本卡时，更新product_sauce表的成本卡uuid失败", zap.String("bom_name", bomCard.Name), zap.Error(err), zap.Any("bom_card", bomCard))
						continue
					}
				}
			}
		}
		for _, bom := range needCreate {
			if err := s.createProductBomCardByErpBom(ctx, tx, bom); err != nil {
				if strings.Contains(err.Error(), "物品或加料不存在") || strings.Contains(err.Error(), "单位不存在") {
					logger.Logger.Error("同步成本卡时，创建成本卡失败，物品或加料或物料单位不存在", zap.String("bom_name", bom.BomName), zap.Any("bom", bom))
					continue
				} else {
					return errors.WithMessage(err, "创建成本卡失败")
				}
			}
		}
		for _, bomCard := range needDisable {
			if err := s.disableProductBomCard(ctx, tx, bomCard); err != nil {
				if strings.Contains(err.Error(), "获取商品或加料失败") {
					logger.Logger.Error("同步成本卡时，失效成本卡失败，商品或加料不存在", zap.String("bom_name", bomCard.Name), zap.Any("bom_card", bomCard))
					continue
				} else {
					return errors.WithMessage(err, "失效成本卡失败")
				}
			}
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	// 从ttpos总店同步
	if companySetting.IsSubShop() {
		// 同步总部成本卡到子店（多语言由 SyncMultiLanguage 任务处理）
		headquarterDb := s.dbm.GetDB(companySetting.HeadquarterUuid)
		commonRepo := repository.NewCommonRepo()
		productBomCardList, err := repository.NewProductBomCardRepo(headquarterDb).GetProductBomCardList(
			commonRepo.WhereBySoftDelete(),
			commonRepo.Preload(
				repository.WithPreload{
					Query: "MultiLanguageName",
				},
				repository.WithPreload{
					Query: "RelatedMaterials",
				},
			),
		)
		if err != nil {
			return errors.WithMessage(err, "获取总部成本卡列表失败")
		}

		if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
			// 删除所有总部的成本卡
			tx.Model(&model.ProductBomCard{}).Where("headquarter_uuid = ?", companySetting.HeadquarterUuid).Delete(&model.ProductBomCard{})
			productBomCardUuids := []uint64{}
			for _, productBomCard := range productBomCardList {
				productBomCardUuids = append(productBomCardUuids, productBomCard.Uuid)
			}
			tx.Model(&model.RelatedMaterial{}).Where("related_uuid IN (?)", productBomCardUuids).Delete(&model.RelatedMaterial{})

			for _, productBomCard := range productBomCardList {
				// 过滤掉数据不完整的成本卡（如无多语言名称）
				if productBomCard.MultiLanguageName == nil {
					continue
				}
				// 创建成本卡
				productBomCard.BaseModel = model.BaseModel{
					Uuid: productBomCard.Uuid,
				}
				productBomCard.HeadquarterUuid = companySetting.HeadquarterUuid
				if err := repository.NewProductBomCardRepo(tx).CreateProductBomCard(*productBomCard); err != nil {
					return errors.WithMessage(err, "创建成本卡失败")
				}
				// 创建成本卡材料
				for _, material := range productBomCard.RelatedMaterials {
					material.BaseModel = model.BaseModel{
						Uuid: material.Uuid,
					}
					if err := repository.NewProductBomCardRepo(tx).CreateProductBomCardMaterial(*material); err != nil {
						return errors.WithMessage(err, "创建成本卡材料失败")
					}
				}
			}
			return nil
		}); err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}

type ObjectByItemCodeResp struct {
	RelatedUuid  uint64
	ProductSauce *model.ProductSauce
	ProductBom   *model.ProductBom
	RelatedType  uint8
}

// 根据item_code查询本地ttpos数据库，判断该item_code是商品还是加料，并返回商品或加料信息
func (s *materialSrv) getObjectByItemCode(_ context.Context, itemCode string, tx *gorm.DB) (*ObjectByItemCodeResp, error) {
	// 查询物品
	productBom, err := repository.NewProductBomRepo(tx).GetProductBomByItemCode(itemCode)
	if err != nil {
		if !strings.Contains(err.Error(), "record not found") {
			return nil, errors.WithMessage(err, "获取物品失败")
		}
	}
	if productBom == nil {
		// 查询加料
		sauce, err := repository.NewProductSauceRepo(tx).GetSauceByErpCode(itemCode)
		if err != nil {
			if !strings.Contains(err.Error(), "record not found") {
				return nil, errors.WithMessage(err, "获取加料失败")
			}
		}
		if sauce == nil {
			return nil, errors.WithMessage(errors.New("物品或加料不存在"), "物品或加料不存在")
		} else {
			return &ObjectByItemCodeResp{
				RelatedUuid:  sauce.Uuid,
				RelatedType:  constant.ProductBomCardRelatedTypeSauce,
				ProductBom:   nil,
				ProductSauce: sauce,
			}, nil
		}
	} else {
		return &ObjectByItemCodeResp{
			RelatedUuid:  productBom.Uuid,
			RelatedType:  constant.ProductBomCardRelatedTypeFlavor,
			ProductBom:   productBom,
			ProductSauce: nil,
		}, nil
	}
}

// 根据成本卡uuid查询本地ttpos数据库，判断该item_code是商品还是加料，并返回商品或加料信息
func (s *materialSrv) getObjectByProductBomCardUuid(_ context.Context, productBomCardUuid uint64, tx *gorm.DB) (*ObjectByItemCodeResp, error) {
	// 查询商品
	productBom, err := repository.NewProductBomRepo(tx).GetProductBomByProductBomCardUuid(productBomCardUuid)
	if err != nil {
		if !strings.Contains(err.Error(), "record not found") {
			return nil, errors.WithMessage(err, "获取商品失败")
		}
	}
	if productBom == nil {
		// 查询加料
		sauce, err := repository.NewProductSauceRepo(tx).GetSauceByProductBomCardUuid(productBomCardUuid)
		if err != nil {
			if !strings.Contains(err.Error(), "record not found") {
				return nil, errors.WithMessage(err, "获取加料失败")
			}
		}
		if sauce == nil {
			return nil, errors.WithMessage(errors.New("商品或加料不存在"), "商品或加料不存在")
		} else {
			return &ObjectByItemCodeResp{
				RelatedUuid:  sauce.Uuid,
				RelatedType:  constant.ProductBomCardRelatedTypeSauce,
				ProductBom:   nil,
				ProductSauce: sauce,
			}, nil
		}
	} else {
		return &ObjectByItemCodeResp{
			RelatedUuid:  productBom.Uuid,
			RelatedType:  constant.ProductBomCardRelatedTypeFlavor,
			ProductBom:   productBom,
			ProductSauce: nil,
		}, nil
	}
}

type ProductBomCardList struct {
	CreateBoms              []*manufacturing.BomInfo
	ExistProductBomCardList []*model.ProductBomCard
	DisableBoms             []*model.ProductBomCard
}

// 获取需要新建、失效的成本卡列表、维持不变的成本卡列表
func (s *materialSrv) getNeedCreateProductBomCardList(headquarterProductBomCardMap map[string]*model.ProductBomCard, erpBoms []*manufacturing.BomInfo) *ProductBomCardList {
	createBoms := []*manufacturing.BomInfo{}
	existProductBomCardList := []*model.ProductBomCard{}
	for _, bom := range erpBoms {
		key := bom.BomName
		if _, ok := headquarterProductBomCardMap[key]; !ok {
			// 需要新建的成本卡。erpnext有，而ttpos没有
			createBoms = append(createBoms, bom)
		} else {
			// erpnext有，ttpos也有。无需处理
			existProductBomCardList = append(existProductBomCardList, headquarterProductBomCardMap[key])
			delete(headquarterProductBomCardMap, key)
		}
	}

	// 需要失效的成本卡。erpnext没有，而ttpos有时
	disableBoms := []*model.ProductBomCard{}
	for _, productBomCard := range headquarterProductBomCardMap {
		disableBoms = append(disableBoms, productBomCard)
	}

	return &ProductBomCardList{
		CreateBoms:              createBoms,
		ExistProductBomCardList: existProductBomCardList,
		DisableBoms:             disableBoms,
	}
}

// 根据erp bom创建ttpos成本卡
func (s *materialSrv) createProductBomCardByErpBom(ctx context.Context, db *gorm.DB, bom *manufacturing.BomInfo) error {
	itemCode := bom.ItemCode
	objectByItemCodeResp, err := s.getObjectByItemCode(ctx, itemCode, db) // 获取物品或加料
	if err != nil {
		return errors.WithMessage(err, "获取物品或加料失败")
	}
	relatedUuid := objectByItemCodeResp.RelatedUuid // 物品或加料的uuid
	relatedType := objectByItemCodeResp.RelatedType // 关联的类型。1:商品，2:加料

	list := []req.ProductBomCardMaterialReq{}
	for _, item := range bom.Items {
		// 获取物品
		materialDetail, err := repository.NewMaterialRepo(db).GetMaterialDetailByErpCode(item.ItemCode)
		if err != nil {
			return errors.WithMessage(err, "获取物品失败")
		}
		// 获取物品成本卡单位
		unitUuid, err := materialDetail.GetUnitUuidByUom(item.Uom)
		if err != nil {
			return errors.WithMessage(err, fmt.Sprintf("获取物品 %s 单位失败：%s", materialDetail.Code, item.Uom))
		}

		list = append(list, req.ProductBomCardMaterialReq{
			MaterialUuid: materialDetail.Uuid,
			Num:          item.Qty,
			UnitUuid:     unitUuid,
		})
	}
	cardUuid, _ := utils.GetID()

	relatedMaterials := []*model.RelatedMaterial{}
	for _, bomItem := range bom.Items {
		// 通过物品的erp_code获取物品详情
		materialDetail, err := repository.NewMaterialRepo(db).GetMaterialDetailByErpCode(bomItem.ItemCode)
		if err != nil {
			return errors.WithMessage(err, "获取物品失败")
		}
		unit, err := materialDetail.GetUnitByUom(bomItem.Uom)
		if err != nil {
			return errors.WithMessage(err, "获取物品单位失败")
		}
		relatedMaterials = append(relatedMaterials, &model.RelatedMaterial{
			RelatedUuid:            cardUuid,
			MaterialUuid:           materialDetail.Uuid,
			Num:                    bomItem.Qty,
			UnitUuid:               unit.Uuid,
			UnitName:               unit.Unit.MultiLanguageName.ToJson(),
			UnitUom:                bomItem.Uom,
			BaseUnitUuid:           materialDetail.UnitUuid,
			BaseUnitName:           materialDetail.Unit.Unit.MultiLanguageName.ToJson(),
			BaseUnitUom:            materialDetail.Unit.Unit.ErpnextUom,
			BaseUnitConversionRate: unit.ConversionRate,
			IsUsed:                 1,
		})
	}

	// 通过商品获取成本卡名称
	cardName := dto.LocaleResponse{}
	if relatedType == constant.ProductBomCardRelatedTypeFlavor {
		cardName = objectByItemCodeResp.ProductBom.ProductPackage.MultiLanguageName.GetNames()
	} else if relatedType == constant.ProductBomCardRelatedTypeSauce {
		cardName = objectByItemCodeResp.ProductSauce.MultiLanguageName.GetNames()
	}

	nameUuid, _ := utils.GetID()
	multiLanguageName := model.MultiLanguageName{BaseModel: model.BaseModel{Uuid: nameUuid}} // 成本卡名称. 与绑定的商品名称相同
	multiLanguageName.InitByLocaleResponse(cardName)

	card := model.ProductBomCard{
		BaseModel: model.BaseModel{
			Uuid: cardUuid,
		},
		Name:                  cardName.ToJson(),
		ErpCode:               bom.ItemCode,
		MultiLanguageNameUuid: nameUuid,
		Num:                   bom.Quantity,
		IsUsed:                1,
		HeadquarterUuid:       ctx.GetCompanySetting().HeadquarterUuid,
		MultiLanguageName:     &multiLanguageName,
		RelatedMaterials:      relatedMaterials,
	}
	// 开始更新DB数据
	// 修改旧的成本卡为未使用
	if relatedType == constant.ProductBomCardRelatedTypeFlavor {
		if err := repository.NewProductBomCardRepo(db).UpdateProductBomCardIsUsed(objectByItemCodeResp.ProductBom.ProductBomCardUuid, 0); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
	} else if relatedType == constant.ProductBomCardRelatedTypeSauce {
		if err := repository.NewProductBomCardRepo(db).UpdateProductBomCardIsUsed(objectByItemCodeResp.ProductSauce.ProductBomCardUuid, 0); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
	}
	// 创建成本卡
	if err := repository.NewProductBomCardRepo(db).CreateProductBomCard(card); err != nil {
		return errors.WithMessage(err, "创建成本卡失败")
	}
	// 创建成本卡材料
	for _, material := range relatedMaterials {
		if err := repository.NewProductBomCardRepo(db).CreateProductBomCardMaterial(*material); err != nil {
			return errors.WithMessage(err, "创建成本卡材料失败")
		}
	}
	// 更新成本卡关联
	if relatedType == constant.ProductBomCardRelatedTypeFlavor {
		if err := repository.NewProductSauceRepo(db).UpdateProductBomCard(relatedUuid, cardUuid); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
	} else if relatedType == constant.ProductBomCardRelatedTypeSauce {
		if err := repository.NewProductSauceRepo(db).UpdateProductBomCard(relatedUuid, cardUuid); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
	}
	// 创建成本卡日志
	productBomCardLog, err := newProductBomCardLog(ctx, 1, cardUuid, cardName.ToJson(),
		relatedUuid, cardName.ToJson(), relatedMaterials, constant.ProductBomCardLogOperationTypeCreate)
	if err != nil {
		return errors.WithMessage(err, "创建成本卡日志失败")
	}
	if err := repository.NewProductBomCardLogRepo(db).CreateProductBomCardLog(*productBomCardLog); err != nil {
		return errors.WithMessage(err, "创建成本卡日志失败")
	}
	return nil
}

func (s *materialSrv) disableProductBomCard(ctx context.Context, db *gorm.DB, productBomCard *model.ProductBomCard) error {
	objectByItemCodeResp, err := s.getObjectByProductBomCardUuid(ctx, productBomCard.Uuid, db) // 获取商品或加料
	if err != nil {
		return errors.WithMessage(err, "获取商品或加料失败")
	}
	relatedUuid := objectByItemCodeResp.RelatedUuid // 商品或加料的uuid
	relatedType := objectByItemCodeResp.RelatedType // 关联的类型。1:商品，2:加料

	// 修改旧的成本卡为未使用
	if relatedType == constant.ProductBomCardRelatedTypeFlavor {
		if err := repository.NewProductBomCardRepo(db).UpdateProductBomCardIsUsed(objectByItemCodeResp.ProductBom.ProductBomCardUuid, 0); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
	} else if relatedType == constant.ProductBomCardRelatedTypeSauce {
		if err := repository.NewProductBomCardRepo(db).UpdateProductBomCardIsUsed(objectByItemCodeResp.ProductSauce.ProductBomCardUuid, 0); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
	}
	// 解除成本卡关联
	if relatedType == constant.ProductBomCardRelatedTypeFlavor {
		productBomCardRepo := repository.NewProductBomRepo(db)
		if err := productBomCardRepo.UpdateProductBomCard(relatedUuid, 0, 999999999); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
	} else if relatedType == constant.ProductBomCardRelatedTypeSauce {
		productBomCardRepo := repository.NewProductSauceRepo(db)
		if err := productBomCardRepo.UpdateProductBomCard(relatedUuid, 0); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
	}

	// 创建成本卡日志
	cardLog := &model.ProductBomCardLog{}
	if relatedType == constant.ProductBomCardRelatedTypeFlavor {
		productBomCardLog, err := newProductBomCardLog(ctx, 0, 0, "", relatedUuid, objectByItemCodeResp.ProductBom.ProductPackage.MultiLanguageName.ToJson(), nil, constant.ProductBomCardLogOperationTypeDelete)
		if err != nil {
			return errors.WithMessage(err, "创建成本卡日志失败")
		}
		cardLog = productBomCardLog
	} else if relatedType == constant.ProductBomCardRelatedTypeSauce {
		productBomCardLog, err := newProductBomCardLog(ctx, 0, 0, "", relatedUuid, objectByItemCodeResp.ProductSauce.MultiLanguageName.ToJson(), nil, constant.ProductBomCardLogOperationTypeDelete)
		if err != nil {
			return errors.WithMessage(err, "创建成本卡日志失败")
		}
		cardLog = productBomCardLog
	}
	if err := repository.NewProductBomCardLogRepo(db).CreateProductBomCardLog(*cardLog); err != nil {
		return errors.WithMessage(err, "创建成本卡日志失败")
	}
	return nil
}

// GetWarehouseItemConsumption 获取仓库物品消耗量
func (s *materialSrv) GetWarehouseItemConsumption(ctx context.Context, warehouseUuid uint64) (material_resp.MaterialConsumptionListResp, error) {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	// 查询当前未交班的班次列表
	staffShiftLogList, err := repository.NewShiftLogRepo(db).GetShiftLogList(
		repository.CommonRepo.WhereByStatus(uint(constant.StaffNotHandedOver)),
	)
	if err != nil {
		return material_resp.MaterialConsumptionListResp{}, errors.WithMessage(err, "获取当前未交班的班次列表失败")
	}
	staffShiftLogUuids := make([]uint64, 0)
	for _, staffShiftLog := range staffShiftLogList {
		staffShiftLogUuids = append(staffShiftLogUuids, staffShiftLog.Uuid)
	}
	// 查询这些班次中指定仓库下的物品消耗量
	itemLogs, err := repository.NewWarehouseFormRepo(db).GetWarehouseOutFormItem(
		func(db *gorm.DB) *gorm.DB {
			return db.Where("warehouse_uuid = ? AND scene = ?", warehouseUuid, 0) // 场景为0，表示销售出库
		},
		func(db *gorm.DB) *gorm.DB {
			return db.Where("staff_shift_log_uuid IN ?", staffShiftLogUuids)
		},
		func(db *gorm.DB) *gorm.DB {
			return db.Where("revoke_time = 0 AND material_uuid != 0") // 未撤销且物品uuid不为0, 获取有效的物品出库记录
		},
	)
	if err != nil {
		return material_resp.MaterialConsumptionListResp{}, errors.WithMessage(err, "获取仓库物品消耗量失败")
	}

	// 将物品消耗量合并
	itemLogsMap := make(map[uint64]material_resp.MaterialConsumption)
	for _, itemLog := range itemLogs {
		if material, ok := itemLogsMap[itemLog.MaterialUuid]; ok {
			material.Consumption = decimal.NewFromFloat(material.Consumption).Add(decimal.NewFromFloat(itemLog.Num)).Round(4).InexactFloat64()
			itemLogsMap[itemLog.MaterialUuid] = material
		} else {
			itemLogsMap[itemLog.MaterialUuid] = material_resp.MaterialConsumption{
				MaterialUuid: itemLog.MaterialUuid,
				Consumption:  itemLog.Num,
			}
		}
	}

	// 获取物品列表
	materialUuids := make([]uint64, 0)
	for _, material := range itemLogsMap {
		materialUuids = append(materialUuids, material.MaterialUuid)
	}
	materialList, err := repository.NewMaterialRepo(db).GetMaterialByUuids(materialUuids)
	if err != nil {
		return material_resp.MaterialConsumptionListResp{}, errors.WithMessage(err, "获取物品列表失败")
	}
	materialMap := make(map[uint64]*model.Material)
	for _, material := range materialList {
		materialMap[material.Uuid] = material
	}

	// 将物品消耗量与物品信息合并
	materialConsumptionList := make([]material_resp.MaterialConsumption, 0)
	for _, itemLog := range itemLogsMap {
		materialConsumptionList = append(materialConsumptionList, material_resp.MaterialConsumption{
			MaterialUuid: itemLog.MaterialUuid,
			MaterialCode: materialMap[itemLog.MaterialUuid].Code,
			Consumption:  itemLog.Consumption,
		})
	}

	return material_resp.MaterialConsumptionListResp{
		List: materialConsumptionList,
	}, nil
}

func (s *materialSrv) CheckMaterialSafetyStock(ctx context.Context, companyUuid uint64) error {
	// 加锁，当前方法执行完才能再次调用
	lockKey := fmt.Sprintf("check_material_safety_stock_%d", companyUuid)
	s.systemLock.LockUuidString(lockKey)
	defer s.systemLock.UnlockUuidString(lockKey)

	var companyUuids []uint64
	if companyUuid == 0 {
		s.dbm.GetDB(constant.DefaultDB).Model(&model.Company{}).Scopes(repository.NotDeleted).Pluck("uuid", &companyUuids)
	} else {
		companyUuids = append(companyUuids, companyUuid)
	}

	// 使用带缓冲的channel控制并发数量，最多10个协程
	maxConcurrent := 10
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	// 用于收集错误
	errChan := make(chan error, len(companyUuids))

	// 分批处理每个公司的安全库存检查
	for _, cid := range companyUuids {
		wg.Add(1)
		utils.Go(func() {
			func(companyId uint64) {
				defer wg.Done()

				// 获取信号量，控制并发数
				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// 获取Context副本用于协程
				ctxCopy := ctx.Copy()

				// 处理单个公司的安全库存检查
				if err := s.checkCompanySafetyStock(ctxCopy, companyId); err != nil {
					errChan <- err
				}
			}(cid)
		})
	}

	// 等待所有协程完成
	wg.Wait()
	close(errChan)

	// 收集所有错误
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		// 返回第一个错误（或者可以返回所有错误的组合）
		return errs[0]
	}

	return nil
}

type SendParams struct {
	CompanyUuid   uint64
	CompanyName   string
	MaterialUuid  uint64
	MaterialName  string
	WarehouseUuid uint64
	WarehouseName string
	CurrentStock  float64
	SafetyStock   float64
	MaterialUnit  string
	TriggerTime   time.Time
}

// sendStockAlertEmail 发送库存预警邮件（带重试机制）
// 参数：
//   - ctx: 上下文
//   - alertType: 预警类型
//   - sendReq: 发送参数
func (s *materialSrv) sendStockAlertEmail(ctx context.Context, alertType uint8, sendReq *SendParams) {
	// 获取预警记录repository
	alertLogRepo := repository.NewMaterialStockAlertLogRepo(s.dbm.GetDB(sendReq.CompanyUuid))

	// 判断是否需要发送预警邮件
	shouldSend, existingLog, err := alertLogRepo.ShouldSendAlert(sendReq.CompanyUuid, sendReq.MaterialUuid, sendReq.WarehouseUuid)
	if err != nil {
		logger.Logger.Error("检查是否需要发送预警邮件失败",
			zap.Uint64("company_uuid", sendReq.CompanyUuid),
			zap.Uint64("material_uuid", sendReq.MaterialUuid),
			zap.Uint64("warehouse_uuid", sendReq.WarehouseUuid),
			zap.Error(err),
		)
		return
	}

	if !shouldSend {
		// 记录跳过原因
		var skipReason string
		var alertCount uint32 = 0
		if existingLog != nil {
			alertCount = existingLog.AlertCount
			if alertCount >= 2 {
				skipReason = "已发送2次预警邮件，不再发送"
			} else {
				skipReason = "24小时内已发送过预警邮件，等待24小时后发送第2次"
			}
		} else {
			skipReason = "跳过发送预警邮件"
		}
		logger.Logger.Info(skipReason,
			zap.Uint64("company_uuid", sendReq.CompanyUuid),
			zap.Uint64("material_uuid", sendReq.MaterialUuid),
			zap.Uint64("warehouse_uuid", sendReq.WarehouseUuid),
			zap.Uint32("alert_count", alertCount),
		)
		return
	}

	// 获取收件人邮箱（从设置中获取，这里暂时使用配置，实际应从数据库获取）
	recipient := s.getStockAlertRecipient(ctx, sendReq.CompanyUuid)
	if recipient == "" {
		logger.Logger.Warn("未配置库存预警邮箱，跳过发送",
			zap.Uint64("company_uuid", sendReq.CompanyUuid),
		)
		return
	}

	triggerTime := sendReq.TriggerTime.Format("15:04, January 2, 2006")
	triggerDate := sendReq.TriggerTime.Format("January 2, 2006")
	// 准备邮件内容
	var subject, messageArgs, templateUuid string
	if sendReq.WarehouseUuid == 0 { // 公司维度
		templateUuid = constant.TplCompanySafetyStockAlert
		messageArgs = fmt.Sprintf(`{"company":"%s","material":"%s","current_stock":"%s %s","safety_stock":"%s %s","time":"%s","date":"%s"}`,
			sendReq.CompanyName, sendReq.MaterialName, utils.FormatFloat(sendReq.CurrentStock), sendReq.MaterialUnit, utils.FormatFloat(sendReq.SafetyStock), sendReq.MaterialUnit, triggerTime, triggerDate)
		subject = fmt.Sprintf("[Alert] %s - Insufficient Safety Stock of %s!", sendReq.CompanyName, sendReq.MaterialName)
	} else {
		templateUuid = constant.TplWarehouseSafetyStockAlert // 仓库维度
		messageArgs = fmt.Sprintf(`{"company":"%s","material":"%s","warehouse":"%s","current_stock":"%s %s","safety_stock":"%s %s","time":"%s","date":"%s"}`,
			sendReq.CompanyName, sendReq.MaterialName, sendReq.WarehouseName, utils.FormatFloat(sendReq.CurrentStock), sendReq.MaterialUnit, utils.FormatFloat(sendReq.SafetyStock), sendReq.MaterialUnit, triggerTime, triggerDate)
		subject = fmt.Sprintf("[Alert] %s - %s - Insufficient Safety Stock of %s!", sendReq.CompanyName, sendReq.WarehouseName, sendReq.MaterialName)
	}

	messageUuid, _ := utils.GetID()
	// 构建发送请求
	sendMessageReq := &v1.SendMessageReq{
		MessageUuid:  strconv.FormatUint(messageUuid, 10),
		TemplateUuid: templateUuid,
		MessageArgs:  messageArgs,
		MessageType:  "email",
		Recipient:    recipient,
		Subject:      subject,
		CompanyUuid:  strconv.FormatUint(sendReq.CompanyUuid, 10),
		OperatorUuid: recipient,
	}

	// 尝试发送邮件，失败时重试2次（共3次）
	maxRetries := 2
	var lastErr error
	sendSuccess := false

	for i := 0; i <= maxRetries; i++ {
		if i > 0 {
			logger.Logger.Info("重试发送库存预警邮件",
				zap.Uint64("company_uuid", sendReq.CompanyUuid),
				zap.Uint64("material_uuid", sendReq.MaterialUuid),
				zap.Uint64("warehouse_uuid", sendReq.WarehouseUuid),
				zap.Int("retry_count", i),
			)
			time.Sleep(time.Second * 2) // 重试前等待2秒
		}

		_, err = s.messageSrv.SendMessage(ctx.GetContext(), sendMessageReq)
		if err == nil {
			sendSuccess = true
			logger.Logger.Info("库存预警邮件发送成功",
				zap.Uint64("company_uuid", sendReq.CompanyUuid),
				zap.Uint64("material_uuid", sendReq.MaterialUuid),
				zap.Uint64("warehouse_uuid", sendReq.WarehouseUuid),
				zap.String("material_name", sendReq.MaterialName),
				zap.String("recipient", recipient),
			)
			break
		}

		lastErr = err
		logger.Logger.Warn("库存预警邮件发送失败",
			zap.Uint64("company_uuid", sendReq.CompanyUuid),
			zap.Uint64("material_uuid", sendReq.MaterialUuid),
			zap.Uint64("warehouse_uuid", sendReq.WarehouseUuid),
			zap.Int("retry_count", i),
			zap.Error(err),
		)
	}

	// 创建或更新预警记录
	now := uint64(time.Now().Unix())
	if existingLog == nil {
		// 创建新记录
		newLog := &model.MaterialStockAlertLog{
			CompanyUuid:   sendReq.CompanyUuid,
			MessageUuid:   messageUuid,
			MaterialUuid:  sendReq.MaterialUuid,
			WarehouseUuid: sendReq.WarehouseUuid,
			AlertType:     alertType,
			CurrentStock:  sendReq.CurrentStock,
			SafetyStock:   sendReq.SafetyStock,
			Recipient:     recipient,
		}

		if sendSuccess {
			// 只有发送成功时才设置计数和时间
			newLog.AlertCount = 1
			newLog.LastAlertTime = now
			newLog.SendStatus = model.SendStatusSuccess
		} else {
			// 发送失败时计数为0，不设置发送时间，以便重试
			newLog.AlertCount = 0
			newLog.LastAlertTime = 0
			newLog.SendStatus = model.SendStatusFailed
			if lastErr != nil {
				newLog.ErrorMessage = lastErr.Error()
			}
		}

		err = alertLogRepo.CreateAlertLog(newLog)
		if err != nil {
			logger.Logger.Error("创建预警记录失败",
				zap.Uint64("company_uuid", sendReq.CompanyUuid),
				zap.Uint64("material_uuid", sendReq.MaterialUuid),
				zap.Uint64("warehouse_uuid", sendReq.WarehouseUuid),
				zap.Error(err),
			)
		}
	} else {
		// 更新现有记录
		existingLog.CurrentStock = sendReq.CurrentStock
		existingLog.SafetyStock = sendReq.SafetyStock
		existingLog.Recipient = recipient
		existingLog.MessageUuid = messageUuid

		if sendSuccess {
			// 只有发送成功时才更新计数和时间
			existingLog.AlertCount++
			existingLog.LastAlertTime = now
			existingLog.SendStatus = model.SendStatusSuccess
			existingLog.ErrorMessage = ""
		} else {
			// 发送失败时不增加计数，不更新发送时间，以便重试
			existingLog.SendStatus = model.SendStatusFailed
			if lastErr != nil {
				existingLog.ErrorMessage = lastErr.Error()
			}
		}

		err = alertLogRepo.UpdateAlertLog(existingLog)
		if err != nil {
			logger.Logger.Error("更新预警记录失败",
				zap.Uint64("company_uuid", sendReq.CompanyUuid),
				zap.Uint64("material_uuid", sendReq.MaterialUuid),
				zap.Uint64("warehouse_uuid", sendReq.WarehouseUuid),
				zap.Error(err),
			)
		}
	}
}

// getStockAlertRecipient 获取库存预警邮件收件人
// 从staff表中查询is_super=1的记录，获取user_name字段（邮箱）
func (s *materialSrv) getStockAlertRecipient(_ context.Context, companyUuid uint64) string {
	// 获取数据库连接
	db := s.dbm.GetDB(companyUuid)
	staffRepo := repository.NewStaffRepo(db)

	// 查询该公司所有is_super=1的员工
	staffs := staffRepo.GetStaffs(
		staffRepo.WhereIsSuper(1),
	)

	if len(staffs) == 0 {
		logger.Logger.Warn("未找到超级管理员，无法获取预警邮箱",
			zap.Uint64("company_uuid", companyUuid),
		)
		return ""
	}

	// 收集所有超级管理员的邮箱（Username字段存储邮箱）
	var emails []string
	for _, staff := range staffs {
		if staff.Username != "" {
			emails = append(emails, staff.Username)
		}
	}

	if len(emails) == 0 {
		logger.Logger.Warn("超级管理员的邮箱为空",
			zap.Uint64("company_uuid", companyUuid),
		)
		return ""
	}

	return emails[0]
}

// checkCompanySafetyStock 检查单个公司的物料安全库存
func (s *materialSrv) checkCompanySafetyStock(ctx context.Context, companyUuid uint64) error {
	// 设置上下文信息
	ctx.SetCompanyUuid(companyUuid)
	ctx.SetLog(logger.Logger)

	company, err := repository.NewCompanyRepo(s.dbm.GetDB(companyUuid)).GetCompanyInfoByUuid(companyUuid)
	if err != nil {
		return errors.WithMessage(err, "查询公司信息失败")
	}
	companySetting := company.CompanySetting
	if companySetting == nil {
		return errors.WithMessage(errors.New("公司设置为空"), "公司设置为空")
	}

	tz := utils.SetTimezone(companySetting.GetTimezone())
	now := tz.Now()

	defaultLanguage := companySetting.GetDefaultLanguage()
	companyName := company.Name

	// 检查是否有safety_stock不为空的物品
	materialRepo := repository.NewMaterialRepo(s.dbm.GetDB(companyUuid))
	materialList := materialRepo.GetMaterialList(repository.NotDeleted, materialRepo.WithMultiLanguageName(), func(db *gorm.DB) *gorm.DB {
		return db.Where("safety_stock IS NOT NULL")
	}, func(db *gorm.DB) *gorm.DB {
		return db.Preload("Unit")
	})
	if len(materialList) == 0 {
		return nil
	}

	// 获取业务设置
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return errors.WithMessage(err, "获取业务设置失败")
	}

	// 提取物料UUID列表
	var materialUuids []uint64
	for _, material := range materialList {
		materialUuids = append(materialUuids, material.Uuid)
	}

	// 创建仓库物品repository和预警记录repository
	warehouseItemRepo := repository.NewWarehouseItemRepo(s.dbm.GetDB(companyUuid))
	alertLogRepo := repository.NewMaterialStockAlertLogRepo(s.dbm.GetDB(companyUuid))

	if businessSetting.SafetyStockType == "1" { // 门店维度
		// 获取所有物料在非在途仓库中的总库存
		stockMap, err := warehouseItemRepo.GetMaterialStockInNormalWarehouses(materialUuids)
		if err != nil {
			return errors.WithMessage(err, "查询物料库存失败")
		}

		// 检查每个物料的库存是否低于安全库存
		for _, material := range materialList {
			unit := material.Unit
			var unitName string
			if unit != nil {
				unitName = model.NewMultiLanguageName(unit.Name).GetNames().EN
				if unitName == "" {
					unitName = model.NewMultiLanguageName(unit.Name).GetNameByLang(defaultLanguage)
				}
			}
			materialName := material.MultiLanguageName.GetNames().EN
			if materialName == "" {
				materialName = material.MultiLanguageName.GetNameByLang(defaultLanguage)
			}

			totalStock := stockMap[material.Uuid] // 如果没有记录，默认为0
			if material.SafetyStock != nil && totalStock < *material.SafetyStock {
				// 库存低于安全库存，记录日志
				logger.Logger.Warn("物料库存低于安全库存",
					zap.Uint64("company_uuid", companyUuid),
					zap.Uint64("material_uuid", material.Uuid),
					zap.String("material_name", material.MultiLanguageName.GetNameByLang(defaultLanguage)),
					zap.Float64("current_stock", totalStock),
					zap.Float64("safety_stock", *material.SafetyStock),
				)

				// 发送库存预警邮件（异步）
				utils.Go(func() {
					s.sendStockAlertEmail(
						ctx.Copy(),
						model.AlertTypeCompany,
						&SendParams{
							CompanyUuid:  companyUuid,
							CompanyName:  companyName,
							MaterialUuid: material.Uuid,
							MaterialName: materialName,
							CurrentStock: totalStock,
							SafetyStock:  *material.SafetyStock,
							MaterialUnit: unitName,
							TriggerTime:  now,
						},
					)
				})
			} else if material.SafetyStock != nil && totalStock >= *material.SafetyStock {
				// 库存已恢复正常，清除预警记录
				err := alertLogRepo.ClearAlertLog(companyUuid, material.Uuid, 0)
				if err != nil {
					logger.Logger.Error("清除预警记录失败",
						zap.Uint64("company_uuid", companyUuid),
						zap.Uint64("material_uuid", material.Uuid),
						zap.Error(err),
					)
				}
			}
		}

	} else if businessSetting.SafetyStockType == "2" { // 仓库维度
		// 获取所有物料在各个仓库中的库存
		stockResults, err := warehouseItemRepo.GetMaterialStockByWarehouse(materialUuids)
		if err != nil {
			return errors.WithMessage(err, "查询物料仓库库存失败")
		}

		// 将结果转换为 map[物料UUID]map[仓库UUID]库存
		stockByWarehouseMap := make(map[uint64]map[uint64]float64)
		for _, result := range stockResults {
			if _, exists := stockByWarehouseMap[result.MaterialUuid]; !exists {
				stockByWarehouseMap[result.MaterialUuid] = make(map[uint64]float64)
			}
			stockByWarehouseMap[result.MaterialUuid][result.WarehouseUuid] = result.Stock
		}

		warehouseNameMap := make(map[uint64]string)
		// 检查每个物料在每个仓库的库存是否低于安全库存
		for _, material := range materialList {
			unit := material.Unit
			var unitName string
			if unit != nil {
				unitName = model.NewMultiLanguageName(unit.Name).GetNames().EN
				if unitName == "" {
					unitName = model.NewMultiLanguageName(unit.Name).GetNameByLang(defaultLanguage)
				}
			}

			materialName := material.MultiLanguageName.GetNames().EN
			if materialName == "" {
				materialName = material.MultiLanguageName.GetNameByLang(defaultLanguage)
			}
			warehouseStocks := stockByWarehouseMap[material.Uuid]
			for warehouseUuid, stock := range warehouseStocks {
				warehouseName, exists := warehouseNameMap[warehouseUuid]
				if !exists {
					var warehouse model.Warehouse
					err := s.dbm.GetDB(companyUuid).Model(&model.Warehouse{}).Preload("MultiLanguageName").Where("uuid = ?", warehouseUuid).Find(&warehouse).Error
					if err != nil {
						logger.Logger.Error("查询仓库失败",
							zap.Uint64("company_uuid", companyUuid),
							zap.Uint64("warehouse_uuid", warehouseUuid),
							zap.Error(err),
						)
						continue
					}
					warehouseName = warehouse.MultiLanguageName.GetNames().EN
					if warehouseName == "" {
						warehouseName = warehouse.MultiLanguageName.GetNameByLang(defaultLanguage)
					}
					warehouseNameMap[warehouseUuid] = warehouseName
				}
				if material.SafetyStock != nil && stock < *material.SafetyStock {
					// 库存低于安全库存，记录日志
					logger.Logger.Warn("物料仓库库存低于安全库存",
						zap.Uint64("company_uuid", companyUuid),
						zap.Uint64("material_uuid", material.Uuid),
						zap.String("material_name", material.MultiLanguageName.GetNameByLang(defaultLanguage)),
						zap.Uint64("warehouse_uuid", warehouseUuid),
						zap.Float64("current_stock", stock),
						zap.Float64("safety_stock", *material.SafetyStock),
					)
					// 发送库存预警邮件（异步）
					utils.Go(func() {
						s.sendStockAlertEmail(
							ctx.Copy(),
							model.AlertTypeWarehouse,
							&SendParams{
								CompanyUuid:   companyUuid,
								CompanyName:   companyName,
								MaterialUuid:  material.Uuid,
								MaterialName:  materialName,
								WarehouseUuid: warehouseUuid,
								WarehouseName: warehouseName,
								CurrentStock:  stock,
								SafetyStock:   *material.SafetyStock,
								MaterialUnit:  unitName,
								TriggerTime:   now,
							},
						)
					})
				} else if material.SafetyStock != nil && stock >= *material.SafetyStock {
					// 库存已恢复正常，清除预警记录
					err := alertLogRepo.ClearAlertLog(companyUuid, material.Uuid, warehouseUuid)
					if err != nil {
						logger.Logger.Error("清除预警记录失败",
							zap.Uint64("company_uuid", companyUuid),
							zap.Uint64("material_uuid", material.Uuid),
							zap.Uint64("warehouse_uuid", warehouseUuid),
							zap.Error(err),
						)
					}
				}
			}
		}
	}

	return nil
}
