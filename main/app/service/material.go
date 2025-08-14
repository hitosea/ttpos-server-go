package service

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp/material_resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
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
		}
		materialList = append(materialList, respMaterial)
	}

	return material_resp.MaterialListWithPaginationResp{
		List: material_resp.MaterialListResp{
			List: materialList,
		},
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
		nameId, err := repository.NewMultiLanguageNameRepoImpl(tx).CreateMultiLanguageName(multiLanguageName)
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

	materialUuid, _ := utils.GetID()

	if err := repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
		materialRepo := repository.NewMaterialRepo(tx)

		// 创建多语言名称
		multiLanguageName := model.MultiLanguageName{}
		multiLanguageName.InitByLocaleResponse(req.LocaleName)
		nameId, err := repository.NewMultiLanguageNameRepoImpl(tx).CreateMultiLanguageName(multiLanguageName)
		if err != nil {
			return errors.WithMessage(err, "创建多语言名称失败")
		}

		unitMap := make(map[uint64]uint64) // 单位ID -> 原料单位ID

		// 创建material_unit
		unitUuid, err := repository.NewMaterialUnitRepo(tx).CreateMaterialUnit(model.MaterialUnit{
			Name:           req.LocaleName.ZH,
			UnitUuid:       req.UnitUuid,
			ConversionRate: 1,
			IsDefault:      1,
			FromUnitUuid:   0,
			MaterialUuid:   materialUuid,
		})
		if err != nil {
			return errors.WithMessage(err, "创建原料单位失败")
		}

		unitMap[req.UnitUuid] = unitUuid

		// 创建非基准单位 material_unit
		for _, unit := range req.UnitList {
			productUnit, err := repository.NewProductRepo(tx).GetProductUnitByUnitUuid(unit.UnitUuid)
			if err != nil {
				return errors.WithMessage(err, "获取产品单位失败")
			}

			materialUnitUuid, err := repository.NewMaterialUnitRepo(tx).CreateMaterialUnit(model.MaterialUnit{
				Name:           productUnit.MultiLanguageName.GetNameByLang("zh"),
				UnitUuid:       productUnit.Uuid,
				ConversionRate: unit.ConversionRate,
				FromUnitUuid:   unitUuid,
				MaterialUuid:   materialUuid,
			})
			if err != nil {
				return errors.WithMessage(err, "创建原料单位失败")
			}
			unitMap[unit.UnitUuid] = materialUnitUuid
		}

		// 创建物品
		material := model.Material{
			BaseModel: model.BaseModel{
				Uuid: materialUuid,
			},
			Name:                  req.LocaleName.ToJson(),
			Code:                  "", // TODO: 从ERP获取编码
			Valuation:             req.Valuation,
			InitStock:             req.InitStock,
			MultiLanguageNameUuid: nameId,
			CategoryUuid:          req.CategoryUuid,
			UnitUuid:              unitMap[req.UnitUuid],
			BarcodeValue:          req.BarcodeValue,
			PurchaseUnitUuid:      unitMap[req.PurchaseUnitUuid],
			CostUnitUuid:          unitMap[req.CostUnitUuid],
			Status: func() bool {
				if req.Status == 1 {
					return true
				}
				return false
			}(),
		}

		_, err = materialRepo.CreateMaterial(material)
		if err != nil {
			return errors.WithMessage(err, "创建物品失败")
		}

		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
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
			multiLanguageNameRepo := repository.NewMultiLanguageNameRepoImpl(tx)
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
			if _, ok := exitUnitMap[unit.UnitUuid]; !ok {
				productUnit, err := repository.NewProductRepo(tx).GetProductUnitByUnitUuid(unit.UnitUuid)
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
				unitMap[unit.UnitUuid] = materialUnitUuid
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

	if len(req.Materials.List) == 0 {
		return errors.New("物品列表不能为空")
	}

	nameUuid, _ := utils.GetID()
	cardUuid, _ := utils.GetID()
	multiLanguageName := model.MultiLanguageName{}
	multiLanguageName.InitByLocaleResponse(req.ProductBomCardName)
	multiLanguageName.Uuid = nameUuid
	_, err := repository.NewMultiLanguageNameRepoImpl(db).CreateMultiLanguageName(multiLanguageName)
	if err != nil {
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
		if err := repository.NewProductBomRepo(tx).UpdateProductBomCard(req.ProductBomUuid, cardUuid); err != nil {
			return errors.WithMessage(err, "更新成本卡失败")
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}
