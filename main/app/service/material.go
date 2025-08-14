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
	UpdateMaterialStatus(ctx context.Context, req req.MaterialStatusReq) error
	AddMaterialCategory(ctx context.Context, req req.MaterialCategoryAddReq) error
	GetMaterialCategoryList(ctx context.Context, req req.MaterialCategoryListReq) (material_resp.MaterialCategoryListResp, error)
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
			UnitName: material.Unit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
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

	// 构建查询选项
	commonRepo := repository.NewCommonRepo()
	dbOptions := []repository.DBOption{
		commonRepo.Preload(
			repository.WithPreload{
				Query: "MultiLanguageName",
			},
			repository.WithPreload{
				Query: "Category.MultiLanguageName",
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
		),
	}

	// 获取物品详情
	material, err := materialRepo.GetMaterialByUuid(req.Uuid, dbOptions...)
	if err != nil {
		return material_resp.MaterialDetailResp{}, errors.WithMessage(err, "获取物品详情失败")
	}

	return material_resp.MaterialDetailResp{
		Uuid:         material.Uuid,
		LocaleName:   material.MultiLanguageName.GetNames(),
		Code:         material.Code,
		CategoryUuid: material.CategoryUuid,
		CategoryName: material.Category.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
		Status: func() int {
			if material.Status {
				return 1
			}
			return 2
		}(),
		Valuation:    material.Valuation,
		BarcodeValue: material.BarcodeValue,
		//UnitName:         material.Unit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
		UnitList: material_resp.MaterialUnitListResp{List: []material_resp.MaterialUnit{}},
		//PurchaseUnitName: material.PurchaseUnit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
		PurchaseUnitUuid: material.PurchaseUnitUuid,
		//CostUnitName:     material.CostUnit.Unit.MultiLanguageName.GetNameByLang(ctx.GetLanguage()),
		CostUnitUuid: material.CostUnitUuid,
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
			})
			if err != nil {
				return errors.WithMessage(err, "创建原料单位失败")
			}
			unitMap[unit.UnitUuid] = materialUnitUuid
		}

		// 创建物品
		material := model.Material{
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
	materialRepo := repository.NewMaterialRepo(s.dbm.GetDB(dbId))

	// 检查物品是否存在
	existingMaterial, err := materialRepo.GetMaterialByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "物品不存在")
	}

	// 更新物品
	material := model.Material{
		MultiLanguageNameUuid: existingMaterial.MultiLanguageNameUuid,
		CategoryUuid:          req.CategoryUuid,
		BarcodeValue:          req.BarcodeValue,
		PurchaseUnitUuid:      req.PurchaseUnitUuid,
		CostUnitUuid:          req.CostUnitUuid,
		Status: func() bool {
			if req.Status == 1 {
				return true
			}
			return false
		}(),
	}

	err = materialRepo.UpdateMaterial(material)
	if err != nil {
		return errors.WithMessage(err, "更新物品失败")
	}

	return nil
}

// DeleteMaterial 删除物品
func (s *materialSrv) DeleteMaterial(ctx context.Context, req req.MaterialDeleteReq) error {
	dbId := ctx.GetDbId()
	materialRepo := repository.NewMaterialRepo(s.dbm.GetDB(dbId))

	// 检查物品是否存在
	_, err := materialRepo.GetMaterialByUuid(req.Uuid)
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

// UpdateMaterialStatus 更新物品状态
func (s *materialSrv) UpdateMaterialStatus(ctx context.Context, req req.MaterialStatusReq) error {
	dbId := ctx.GetDbId()
	materialRepo := repository.NewMaterialRepo(s.dbm.GetDB(dbId))

	// 检查物品是否存在
	_, err := materialRepo.GetMaterialByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "物品不存在")
	}

	// 更新物品状态
	err = materialRepo.UpdateMaterialStatus(req.Uuid, req.Status)
	if err != nil {
		return errors.WithMessage(err, "更新物品状态失败")
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
