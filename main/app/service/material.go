package service

import (
	"fmt"
	"ttpos-bmp/app/ttpos-erp/api/manufacturing"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp/material_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

// IMaterialSrv 物品服务接口
type IMaterialSrv interface {
	GetMaterialList(ctx context.Context, req req.MaterialListReq) (material_resp.MaterialListWithPaginationResp, error)
	GetMaterialDetail(ctx context.Context, req req.MaterialDetailReq) (material_resp.MaterialDetailResp, error)
	AddMaterial(ctx context.Context, req req.MaterialAddReq) error
	EditMaterial(ctx context.Context, req req.MaterialEditReq) error
	DeleteMaterial(ctx context.Context, req req.MaterialDeleteReq) error
	UpdateMaterialStatusBatch(ctx context.Context, req req.MaterialStatusReq) error
	AddMaterialCategory(ctx context.Context, req req.MaterialCategoryAddReq) error
	GetMaterialCategoryList(ctx context.Context, req req.MaterialCategoryListReq) (material_resp.MaterialCategoryListResp, error)
	GetMaterialUnitList(ctx context.Context, req req.MaterialUnitListReq) (material_resp.MaterialUnitListResp, error)
	AddProductBomCard(ctx context.Context, req req.ProductBomCardAddReq) error
	GetProductBomCardDetail(ctx context.Context, req req.ProductBomCardDetailReq) (*material_resp.ProductBomCardDetailResp, error)
	UnlinkProductBomCard(ctx context.Context, req req.ProductBomCardUnlinkReq) error
	CopyProductBomCard(ctx context.Context, req req.ProductBomCardCopyReq) error
	ImportProductBomCard(ctx context.Context, req req.ProductBomCardImportReq) error
}

type materialSrv struct {
	dbm        *database.DBManager // 数据库管理器
	localeSrv  ILocaleSrv          // 多语言名称服务
	settingSrv setting.ISrv
}

func NewMaterialSrv(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv) IMaterialSrv {
	return NewMaterialSrvImpl(dbm, localeSrv, settingSrv)
}

func NewMaterialSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv, settingSrv setting.ISrv) IMaterialSrv {
	return &materialSrv{
		dbm:        dbm,
		localeSrv:  localeSrv,
		settingSrv: settingSrv,
	}
}

// GetMaterialList 获取物品列表
func (s *materialSrv) GetMaterialList(ctx context.Context, req req.MaterialListReq) (material_resp.MaterialListWithPaginationResp, error) {
	dbId := ctx.GetDbId()
	materialRepo := repository.NewMaterialRepo(s.dbm.GetDB(dbId))

	// 构建查询选项
	var dbOptions []repository.DBOption
	commonRepo := repository.NewCommonRepo()

	if req.Keyword != "" {
		dbOptions = append(dbOptions, commonRepo.WhereLike("name", "%"+req.Keyword+"%"))
	}
	if req.CategoryUuid != 0 {
		dbOptions = append(dbOptions, commonRepo.WhereByCategoryUuid(req.CategoryUuid))
	}
	if req.Status != 0 {
		dbOptions = append(dbOptions, commonRepo.WhereByStatus(uint(req.Status)))
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
		if material.InitStock == 0 {
			continue // 过滤掉非移动管理端添加的物品
		}
		unitList := []material_resp.MaterialUnit{}
		for _, unit := range material.NotBaseUnitList {
			unitList = append(unitList, material_resp.MaterialUnit{
				Uuid:           unit.Uuid,
				Name:           unit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
				ConversionRate: unit.ConversionRate,
			})
		}
		respMaterial := material_resp.Material{
			Uuid:         material.Uuid,
			Name:         material.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
			CategoryUuid: material.CategoryUuid,
			Image:        material.GetImage(utils.GetBaseURL(ctx.GetGin().Request)),
			Status: func() int {
				if material.Status {
					return 1
				}
				return 2
			}(),
			UnitName:         material.Unit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
			PurchaseUnitName: material.PurchaseUnit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
			CostUnitName:     material.CostUnit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
			UnitList:         unitList,
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
			Name:           materialUnit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
			ConversionRate: materialUnit.ConversionRate,
		})
	}
	return material_resp.MaterialDetailResp{
		Uuid:             material.Uuid,
		LocaleName:       material.MultiLanguageName.GetNames(),
		Code:             material.Code,
		CategoryUuid:     material.CategoryUuid,
		CategoryName:     material.Category.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
		Status:           int(utils.BoolToUint(material.Status)),
		Valuation:        material.Valuation,
		BarcodeValue:     material.BarcodeValue,
		UnitName:         material.Unit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
		UnitList:         material_resp.MaterialUnitListResp{List: unitList},
		PurchaseUnitName: material.PurchaseUnit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
		PurchaseUnitUuid: material.PurchaseUnitUuid,
		CostUnitName:     material.CostUnit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
		CostUnitUuid:     material.CostUnitUuid,
	}, nil
}

// AddMaterialCategory 创建物品类别
func (s *materialSrv) AddMaterialCategory(ctx context.Context, req req.MaterialCategoryAddReq) error {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		materialCategoryRepo := repository.NewMaterialRepo(tx)

		// 检查物品类别名称是否已存在
		_, err := materialCategoryRepo.GetMaterialCategoryByName(req.LocaleName.ZH)
		if err == nil {
			return errors.New("物品类别名称已存在")
		}

		// 创建多语言名称
		multiLanguageName := model.MultiLanguageName{}
		multiLanguageName.InitByLocaleResponse(req.LocaleName)
		nameId, err := repository.NewMultiLanguageNameRepo(tx).CreateMultiLanguageName(multiLanguageName)
		if err != nil {
			return errors.WithMessage(err, "创建多语言名称失败")
		}

		// 创建物品类别
		materialCategory := model.MaterialCategory{
			MultiLanguageNameUuid: nameId,
			Name:                  req.LocaleName.ZH,
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

// AddMaterial 添加物品
func (s *materialSrv) AddMaterial(ctx context.Context, req req.MaterialAddReq) error {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		material, materialAddErpReq, err := addMaterial(ctx, tx, req)
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

func addMaterial(ctx context.Context, tx *gorm.DB, request req.MaterialAddReq) (*model.Material, *req.MaterialAddErpReq, error) {
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

	// 创建物品
	material := model.Material{
		BaseModel: model.BaseModel{
			Uuid: materialUuid,
		},
		Name:                  request.LocaleName.ToJson(),
		Code:                  "", // TODO: 从ERP获取编码
		Valuation:             request.Valuation,
		InitStock:             request.InitStock,
		MultiLanguageNameUuid: nameId,
		CategoryUuid:          request.CategoryUuid,
		UnitUuid:              unitMap[request.UnitUuid],
		BarcodeValue:          request.BarcodeValue,
		PurchaseUnitUuid:      unitMap[request.PurchaseUnitUuid],
		CostUnitUuid:          unitMap[request.CostUnitUuid],
		Status: func() bool {
			if request.Status == 1 {
				return true
			}
			return false
		}(),
		NotBaseUnitList: notBaseUnitList,
		Unit:            &unit,
	}

	_, err = materialRepo.CreateMaterial(material)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "创建物品失败")
	}

	for _, unit := range notBaseUnitList {
		_, err = repository.NewMaterialUnitRepo(tx).CreateMaterialUnit(*unit)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "创建原料单位失败")
		}
	}

	materialAddErpReq := &req.MaterialAddErpReq{
		ItemName:      request.LocaleName.GetLocale(string(dto.LocaleEN)),
		StockUom:      productUnit.ErpnextUom,
		ValuationRate: request.Valuation,
		OpeningStock:  request.InitStock,
		Uoms:          unitList,
	}

	return &material, materialAddErpReq, nil
}

// EditMaterial 编辑物品
func (s *materialSrv) EditMaterial(ctx context.Context, req req.MaterialEditReq) error {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		materialRepo := repository.NewMaterialRepo(tx)

		// 检查物品是否存在
		existingMaterial, err := materialRepo.GetMaterialDetailByUuid(req.Uuid)
		if err != nil {
			return errors.WithMessage(err, "物品不存在")
		}

		// 判断名称是否修改
		if existingMaterial.MultiLanguageName.IsNameChanged(req.LocaleName) {
			// 更新多语言名称
			multiLanguageName := model.MultiLanguageName{}
			multiLanguageName.InitByLocaleResponse(req.LocaleName)
			multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
			err = multiLanguageNameRepo.UpdateMultiLanguageName(existingMaterial.MultiLanguageNameUuid, multiLanguageName)
			if err != nil {
				return errors.WithMessage(err, "创建多语言名称失败")
			}
			existingMaterial.Name = multiLanguageName.ToJson()
		}

		// 判断非基准单位是否有新增
		unitMap := make(map[uint64]uint64)   // 单位ID -> 原料单位ID
		exitUnitMap := make(map[uint64]bool) //  单位ID 是否存在
		for _, unit := range existingMaterial.NotBaseUnitList {
			unitMap[unit.Unit.Uuid] = unit.Uuid
			exitUnitMap[unit.UnitUuid] = true // 单位ID 是否存在
		}
		for _, unit := range req.UnitList {
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
			Name:         existingMaterial.Name,
			CategoryUuid: req.CategoryUuid,
			Status: func() bool {
				if req.Status == 1 {
					return true
				}
				return false
			}(),
			Valuation:    req.Valuation,
			BarcodeValue: req.BarcodeValue,
			PurchaseUnitUuid: func() uint64 {
				// 如果选择了已经存在的单位，则使用已存在的单位
				if _, ok := exitUnitMap[req.PurchaseUnitUuid]; ok {
					return req.PurchaseUnitUuid
				}
				// 如果选择了新的单位，则创建新的单位
				return unitMap[req.PurchaseUnitUuid]
			}(),
			CostUnitUuid: func() uint64 {
				if _, ok := exitUnitMap[req.CostUnitUuid]; ok {
					return req.CostUnitUuid
				}
				return unitMap[req.CostUnitUuid]
			}(),
		}

		err = materialRepo.UpdateMaterial(material)
		if err != nil {
			return errors.WithMessage(err, "更新物品失败")
		}

		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

// DeleteMaterial 删除物品
func (s *materialSrv) DeleteMaterial(ctx context.Context, req req.MaterialDeleteReq) error {
	dbId := ctx.GetDbId()
	materialRepo := repository.NewMaterialRepo(s.dbm.GetDB(dbId))

	// 检查物品是否存在
	_, err := materialRepo.GetMaterialDetailByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "物品不存在")
	}

	// 删除物品
	err = materialRepo.DeleteMaterial(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "删除物品失败")
	}

	return nil
}

// UpdateMaterialStatusBatch 批量修改物品状态
func (s *materialSrv) UpdateMaterialStatusBatch(ctx context.Context, req req.MaterialStatusReq) error {
	dbId := ctx.GetDbId()
	db := s.dbm.GetDB(dbId)

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		materialRepo := repository.NewMaterialRepo(tx)
		return materialRepo.UpdateMaterialStatusBatch(req.Uuids, req.Status)
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
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
			Uuid: materialCategory.Uuid,
			Name: materialCategory.Name,
		})
	}

	return material_resp.MaterialCategoryListResp{
		List: materialCategoryList,
	}, nil
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
			ConversionRate: materialUnit.ConversionRate,
		})
	}

	return material_resp.MaterialUnitListResp{
		List: materialUnitListResp,
	}, nil
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

func (s *materialSrv) addSauceBomCard(ctx context.Context, req req.ProductBomCardAddReq) error {
	db := ctx.GetDB()
	// 获取小料名称
	sauceRepo := repository.NewProductSauceRepo(db)
	sauce, err := sauceRepo.GetSauceByUuid(req.RelatedUuid)
	if err != nil {
		return errors.WithMessage(err, "获取小料失败")
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
			BaseUnitUuid:           baseUnit.Uuid,
			BaseUnitName:           baseUnit.Unit.MultiLanguageName.ToJson(),
			BaseUnitConversionRate: materialUnit.ConversionRate,
		})
	}

	productBomCard := model.ProductBomCard{
		BaseModel: model.BaseModel{
			Uuid: cardUuid,
		},
		Name:                  multiLanguageName.ToJson(),
		MultiLanguageNameUuid: nameUuid,
		Num:                   float64(req.Num),
		RelatedMaterials:      materialList,
	}

	productBomCardLog, err := newProductBomCardLog(ctx, float64(req.Num), cardUuid, multiLanguageName.ToJson(),
		req.RelatedUuid, sauce.MultiLanguageName.ToJson(), materialList, constant.ProductBomCardLogOperationTypeCreate)
	if err != nil {
		return errors.WithMessage(err, "创建成本卡日志失败")
	}

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
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
			BaseUnitUuid:           baseUnit.Uuid,
			BaseUnitName:           baseUnit.Unit.MultiLanguageName.ToJson(),
			BaseUnitConversionRate: materialUnit.ConversionRate,
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
		RelatedMaterials:      materialList,
	}

	productBomCardLog, err := newProductBomCardLog(ctx, float64(req.Num), cardUuid, multiLanguageName.ToJson(),
		req.RelatedUuid, productBom.ProductPackage.MultiLanguageName.ToJson(), materialList, constant.ProductBomCardLogOperationTypeCreate)
	if err != nil {
		return errors.WithMessage(err, "创建成本卡日志失败")
	}

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		if ctx.GetCompany().IsOpenErp() {
			// 同步成本卡到erp
			erpBomItemList := []*manufacturing.BomItem{}
			for _, material := range materialList {
				unitName := model.NewMultiLanguageName(material.UnitName)
				erpBomItemList = append(erpBomItemList, &manufacturing.BomItem{
					ItemCode: material.Material.Code,
					Rate:     material.Material.Valuation,
					Qty:      material.Num,
					Uom:      unitName.EnName,
				})
			}
			erpSrv := erp.NewIErpSrv(s.dbm)
			erpBomResp, errErp := erpSrv.AddProductBomCard(ctx, erp.ProductBomCardAddErpReq{
				ItemCode: productBom.ErpCode,                                             // 商品编码
				Quantity: float64(req.Num),                                               // 数量
				Uom:      productBom.ProductPackage.ProductUnit.MultiLanguageName.EnName, // 单位
				Items:    erpBomItemList,
			})
			if errErp != nil {
				return errors.WithMessage(errErp)
			}
			productBomCard.ErpCode = erpBomResp.BomName // 记录erp成本卡编码
		}

		productBomCardRepo := repository.NewProductBomCardRepo(tx)
		if err := productBomCardRepo.CreateProductBomCard(productBomCard); err != nil {
			return errors.WithMessage(err, "创建成本卡失败")
		}
		for _, material := range materialList {
			if err := productBomCardRepo.CreateProductBomCardMaterial(*material); err != nil {
				return errors.WithMessage(err, "创建成本卡材料失败")
			}
		}
		if err := repository.NewProductBomRepo(tx).UpdateProductBomCard(req.RelatedUuid, cardUuid); err != nil {
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
	productBomCard, err := productBomCardRepo.GetProductBomCardDetail(req.ProductBomCardUuid)
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
				ConversionRate: unit.ConversionRate,
			})
		}
		materialList = append(materialList, material_resp.ProductBomCardMaterial{
			Material: material_resp.MaterialInfo{
				Uuid: material.MaterialUuid,
				Name: material.Material.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
				Code: material.Material.Code,
			},
			Num: material.Num,
			Unit: material_resp.MaterialUnit{
				Uuid:           material.UnitUuid,
				Name:           material.GetUnitName(ctx.GetLanguage()),
				ConversionRate: material.BaseUnitConversionRate,
			},
			UnitList: unitList,
		})
	}

	return &material_resp.ProductBomCardDetailResp{
		Uuid:      productBomCard.Uuid,
		Name:      productBomCard.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
		Num:       productBomCard.Num,
		Materials: materialList,
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

	productBomCardLog, err := newProductBomCardLog(ctx, 0, 0, "", req.RelatedUuid, sauce.MultiLanguageName.ToJson(), nil, constant.ProductBomCardLogOperationTypeDelete)
	if err != nil {
		return errors.WithMessage(err, "创建成本卡日志失败")
	}

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
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

	productBomCardLog, err := newProductBomCardLog(ctx, 0, 0, "", req.RelatedUuid, productBom.ProductPackage.MultiLanguageName.ToJson(), nil, constant.ProductBomCardLogOperationTypeDelete)
	if err != nil {
		return errors.WithMessage(err, "创建成本卡日志失败")
	}

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		productBomCardRepo := repository.NewProductBomRepo(tx)
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
			if err := repository.NewProductBomRepo(tx).UpdateProductBomCard(req.RelatedUuid, copyProductBomCard.Uuid); err != nil {
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

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		// 创建物品
		material, materialAddErpReq, err := addMaterial(ctx, db, req.MaterialAddReq)
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

		// 创建成本卡
		nameUuid, _ := utils.GetID()
		multiLanguageName := model.MultiLanguageName{}
		multiLanguageName.InitByLocaleResponse(req.MaterialAddReq.LocaleName)
		multiLanguageName.Uuid = nameUuid
		_, errName := repository.NewMultiLanguageNameRepo(db).CreateMultiLanguageName(multiLanguageName)
		if errName != nil {
			return errors.WithMessage(err, "创建多语言名称失败")
		}
		materialList := []*model.RelatedMaterial{}
		materialList = append(materialList, &model.RelatedMaterial{
			RelatedUuid:            req.RelatedUuid,
			MaterialUuid:           material.Uuid,
			Num:                    req.Num,
			UnitUuid:               material.UnitUuid,
			UnitName:               material.Unit.Name,
			BaseUnitUuid:           material.UnitUuid,
			BaseUnitName:           material.Unit.Name,
			BaseUnitConversionRate: 1,
			Material:               material,
		})

		productBomCardUuid, _ := utils.GetID()
		productBomCard := model.ProductBomCard{
			BaseModel: model.BaseModel{
				Uuid: productBomCardUuid,
			},
			Name:                  multiLanguageName.ToJson(),
			MultiLanguageNameUuid: nameUuid,
			Num:                   1, // 加工份数,目前成本卡的加工份数固定为1
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
				unitName := model.NewMultiLanguageName(material.UnitName)
				erpBomItemList = append(erpBomItemList, &manufacturing.BomItem{
					ItemCode: code,
					Rate:     material.Material.Valuation,
					Qty:      material.Num,
					Uom:      unitName.EnName,
				})
			}
			productBomRepo := repository.NewProductBomRepo(db)
			productBom, err := productBomRepo.GetFlavorProductBomByUuid(req.RelatedUuid)
			if err != nil {
				return errors.WithMessage(err, "获取商品规格失败")
			}
			erpSrv := erp.NewIErpSrv(s.dbm)
			erpBomResp, errErp := erpSrv.AddProductBomCard(ctx, erp.ProductBomCardAddErpReq{
				ItemCode: productBom.ErpCode,
				Quantity: 1,
				Uom:      productBom.ProductPackage.ProductUnit.ErpnextUom,
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
		if err := repository.NewProductBomRepo(db).UpdateProductBomCard(req.RelatedUuid, productBomCardUuid); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}

		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}
	return nil
}
