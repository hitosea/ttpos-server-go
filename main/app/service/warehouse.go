package service

import (
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/rpc/erp"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// IWarehouseSrv 仓库服务接口
type IWarehouseSrv interface {
	GetWarehouseList(ctx context.Context, req req.WarehouseListReq) (resp.WarehouseListResp, error)                   // 仓库列表
	GetHeadquarterWarehouseList(ctx context.Context) (resp.WarehouseListResp, error)                                  // 总部仓库列表
	CreateWarehouse(ctx context.Context, addReq req.CreateWarehouseReq) error                                         // 创建仓库
	UpdateWarehouse(ctx context.Context, req req.UpdateWarehouseReq) error                                            // 更新仓库
	DeleteWarehouse(ctx context.Context, req req.DeleteWarehouseReq) error                                            // 删除仓库
	SetDefaultWarehouse(ctx context.Context, req req.SetDefaultWarehouseReq) error                                    // 设置默认仓库
	GetWarehouse(ctx context.Context, req req.WarehouseReq) (resp.WarehouseResp, error)                               // 获取仓库
	GetWarehouseInOutList(ctx context.Context, req req.GetWarehouseInOutListReq) (resp.WarehouseInOutListResp, error) // 出入库明细列表
	CheckCodeExists(ctx context.Context, req req.CheckCodeExistsReq) (resp.CheckNameCodeExistsResp, error)            // 检查仓库编码是否存在

	SyncWarehouse(ctx context.Context) error // 同步仓库列表
}

// NewWarehouseSrv 创建仓库服务
func NewWarehouseSrv(dbm *database.DBManager) IWarehouseSrv {
	return NewWarehouseSrvImpl(dbm)
}

// warehouseSrv 仓库服务实现
type warehouseSrv struct {
	dbm *database.DBManager
}

// NewWarehouseSrvImpl 创建仓库服务实现
func NewWarehouseSrvImpl(dbm *database.DBManager) IWarehouseSrv {
	return &warehouseSrv{
		dbm: dbm,
	}
}

// GetWarehouseList 获取仓库列表
func (s *warehouseSrv) GetWarehouseList(ctx context.Context, req req.WarehouseListReq) (resp.WarehouseListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	warehouseRepo := repository.NewWarehouseRepo(db)
	// 构建查询选项
	var opts []repository.DBOption
	// 名称筛选
	if req.Keyword != "" {
		opts = append(opts, warehouseRepo.WhereNameOrCodeLike(req.Keyword))
	}
	// 类型筛选
	if req.Type != "" {
		opts = append(opts, warehouseRepo.WhereType(req.Type))
	}
	// 状态筛选
	if req.Status != nil && slice.Contain([]int{0, 1}, *req.Status) {
		opts = append(opts, warehouseRepo.WhereStatus(*req.Status))
	}
	// 排序
	opts = append(opts, warehouseRepo.OrderByCreateTime(true))
	// 分页查询
	warehouses, total, err := warehouseRepo.GetListWithPagination(
		req.PageReq.PageNo,
		req.PageReq.PageSize,
		opts...,
	)
	if err != nil {
		return resp.WarehouseListResp{}, errors.WithMessage(err, "获取仓库列表失败")
	}
	// 构建响应数据
	warehouseList := make([]resp.WarehouseResp, 0, len(warehouses))
	for _, warehouse := range warehouses {
		warehouseList = append(warehouseList, s.buildWarehouseResp(warehouse))
	}

	return resp.WarehouseListResp{
		List: warehouseList,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}

// GetHeadquarterWarehouseList 获取总部仓库列表
func (s *warehouseSrv) GetHeadquarterWarehouseList(ctx context.Context) (resp.WarehouseListResp, error) {
	// 获取总部公司
	headquarterUuid := ctx.GetCompanySetting().HeadquarterUuid
	db := s.dbm.GetDB(headquarterUuid)
	company, err := repository.NewCompanyRepo(db).GetCompanyInfoByUuid(headquarterUuid)
	if err != nil {
		logger.Logger.Error("获取总部公司数据失败", zap.Error(err))
		return resp.WarehouseListResp{}, errors.WithMessage(errors.New("获取总部公司数据失败"), err.Error())
	}

	// 获取总部仓库列表
	companySetting := company.CompanySetting
	warehouseList, err := erp.NewIErpSrv(s.dbm).GetWarehouseList(ctx.GetContext(), req.GetErpnextWarehouseListReq{
		SiteCode:    companySetting.ErpnextSiteCode,
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
		Branch:      companySetting.ErpnextBranchName,
	})
	if err != nil {
		return resp.WarehouseListResp{}, errors.WithMessage(errors.New("同步仓库失败"), err.Error())
	}

	// 转换 warehouseList 为 resp.WarehouseResp 格式
	warehouseRespList := make([]resp.WarehouseResp, 0, len(warehouseList))
	for _, warehouse := range warehouseList {
		warehouseResp := resp.WarehouseResp{
			LocalName: dto.LocaleResponse{
				ZH:   warehouse.WarehouseName, // 使用仓库显示名称作为中文名称
				EN:   warehouse.WarehouseName, // 使用仓库全称作为英文名称
				TH:   warehouse.WarehouseName, // 使用仓库全称作为泰文名称
				ZHTW: warehouse.WarehouseName, // 使用仓库全称作为粤文名称
				JA:   warehouse.WarehouseName, // 使用仓库全称作为日文名称
				KO:   warehouse.WarehouseName, // 使用仓库全称作为韩文名称
				MY:   warehouse.WarehouseName, // 使用仓库全称作为缅文名称
				TR:   warehouse.WarehouseName, // 使用仓库全称作为土耳其文名称
				SV:   warehouse.WarehouseName, // 使用仓库全称作为瑞典文名称
			},
			Code: warehouse.Name, // 使用仓库全称作为编码
		}
		warehouseRespList = append(warehouseRespList, warehouseResp)
	}

	return resp.WarehouseListResp{
		List: warehouseRespList,
		Meta: dto.PageResponse{
			PageNo:   1,
			PageSize: len(warehouseRespList),
			Total:    int64(len(warehouseRespList)),
		},
	}, nil
}

// GetWarehouse 获取仓库
func (s *warehouseSrv) GetWarehouse(ctx context.Context, req req.WarehouseReq) (resp.WarehouseResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	warehouseRepo := repository.NewWarehouseRepo(db)
	warehouse, err := warehouseRepo.GetByUuid(req.Uuid)
	if err != nil {
		return resp.WarehouseResp{}, errors.WithMessage(err, "获取仓库失败")
	}
	return s.buildWarehouseResp(*warehouse), nil
}

// CreateWarehouse 创建仓库
func (s *warehouseSrv) CreateWarehouse(ctx context.Context, addReq req.CreateWarehouseReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	var syncEver int64
	db.Model(&model.Warehouse{}).Count(&syncEver)
	if syncEver == 0 {
		return errors.New("仓库未同步")
	}
	warehouseRepo := repository.NewWarehouseRepo(db)
	// 检查仓库编码是否已存在
	exists, err := warehouseRepo.IsCodeExists(addReq.Code, 0)
	if err != nil {
		return errors.WithMessage(err, "检查仓库编码失败")
	}
	if exists {
		return errors.New("仓库编码已存在")
	}
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, addReq.LocaleName)
	for _, name := range names {
		if !checkService.CheckNameLength(ctx, name.Text, 150) {
			return errors.New("仓库名称长度不能超过150")
		}
	}
	exists = checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Source: constant.CheckNameSourceCategory,
		Names:  names,
	})
	if exists {
		return errors.New("仓库名称已存在")
	}
	// 生成UUID
	uuid, err := utils.GetID()
	if err != nil {
		return errors.WithMessage(err, "生成UUID失败")
	}
	// 创建仓库模型
	warehouse := model.Warehouse{
		BaseModel: model.BaseModel{
			Uuid:       uuid,
			CreateTime: time.Now().Unix(),
			UpdateTime: time.Now().Unix(),
		},
	}
	// 复制请求数据到模型
	err = copier.Copy(&warehouse, &addReq)
	if err != nil {
		return errors.WithMessage(err, "数据复制失败")
	}

	companySetting := ctx.GetCompanySetting()
	typeMap := map[string]string{
		"normal":  "Normal",
		"transit": "Transit",
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		warehouseName, err := GetEnName(ctx, addReq.LocaleName)
		if err != nil {
			return errors.WithMessage(errors.New("翻译失败"), err.Error())
		}
		// 保存多语言
		multiLanguageName := model.MultiLanguageName{
			ZhName:   addReq.LocaleName.ZH,
			ThName:   addReq.LocaleName.TH,
			EnName:   addReq.LocaleName.EN,
			ZhTwName: addReq.LocaleName.ZHTW,
			JaName:   addReq.LocaleName.JA,
			KoName:   addReq.LocaleName.KO,
			MyName:   addReq.LocaleName.MY,
			TrName:   addReq.LocaleName.TR,
			SvName:   addReq.LocaleName.SV,
		}
		err = tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
		if err != nil {
			return errors.WithMessage(err, "创建多语言名称失败")
		}
		warehouse.Name = addReq.LocaleName.ToJson()
		warehouse.MultiLanguageNameUuid = multiLanguageName.Uuid
		var erpCode string
		// 调用erp接口
		if ctx.GetCompany().IsOpenErp() {
			erpCode, err = erp.NewIErpSrv(s.dbm).CreateWarehouse(ctx.GetContext(), req.CreateErpnextWarehouseReq{
				SiteCode:      companySetting.ErpnextSiteCode,
				WarehouseName: warehouseName,
				AliasName:     warehouseName,
				CompanyAbbr:   companySetting.ErpnextCompanyAbbr,
				Branch:        companySetting.ErpnextBranchName,
				Disabled:      addReq.Status == 0,
				WarehouseType: typeMap[addReq.Type], // 修改为大驼峰
			})
			if err != nil {
				return errors.WithMessage(errors.New("创建仓库失败"), err.Error())
			}
		}
		warehouse.ErpCode = erpCode
		// 保存到数据库
		err = tx.Model(&model.Warehouse{}).Create(&warehouse).Error
		if err != nil {
			return errors.WithMessage(err, "创建仓库失败")
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(errors.New("创建仓库失败"), err.Error())
	}

	// 返回创建的仓库信息
	return nil
}

// UpdateWarehouse 更新仓库
func (s *warehouseSrv) UpdateWarehouse(ctx context.Context, updateReq req.UpdateWarehouseReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	warehouseRepo := repository.NewWarehouseRepo(db)

	// 检查仓库是否存在
	existingWarehouse, err := warehouseRepo.GetByUuid(updateReq.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("仓库不存在")
		}
		return errors.WithMessage(err, "获取仓库信息失败")
	}

	// 检查仓库编码是否已存在（排除当前仓库）
	exists, err := warehouseRepo.IsCodeExists(updateReq.Code, existingWarehouse.Uuid)
	if err != nil {
		return errors.WithMessage(err, "检查仓库编码失败")
	}
	if exists {
		return errors.New("仓库编码已存在")
	}
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, updateReq.LocaleName)
	for _, name := range names {
		if !checkService.CheckNameLength(ctx, name.Text, 150) {
			return errors.New("仓库名称长度不能超过150")
		}
	}
	exists = checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Uuid:   existingWarehouse.Uuid,
		Source: constant.CheckNameSourceCategory,
		Names:  names,
	})
	if exists {
		return errors.New("仓库名称已存在")
	}

	updateData := map[string]any{
		"name":    updateReq.LocaleName.ToJson(),
		"type":    updateReq.Type,
		"code":    updateReq.Code,
		"status":  updateReq.Status,
		"contact": updateReq.Contact,
		"phone":   updateReq.Phone,
		"address": updateReq.Address,
	}

	companySetting := ctx.GetCompanySetting()
	typeMap := map[string]string{
		"normal":  "Normal",
		"transit": "Transit",
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		warehouseName, err := GetEnName(ctx, updateReq.LocaleName)
		if err != nil {
			return errors.WithMessage(errors.New("翻译失败"), err.Error())
		}
		// 更新多语言名称
		err = tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", existingWarehouse.MultiLanguageNameUuid).Updates(map[string]any{
			"zh_name":    updateReq.LocaleName.ZH,
			"th_name":    updateReq.LocaleName.TH,
			"en_name":    updateReq.LocaleName.EN,
			"zh_tw_name": updateReq.LocaleName.ZHTW,
			"ja_name":    updateReq.LocaleName.JA,
			"ko_name":    updateReq.LocaleName.KO,
			"my_name":    updateReq.LocaleName.MY,
			"tr_name":    updateReq.LocaleName.TR,
			"sv_name":    updateReq.LocaleName.SV,
		}).Error
		if err != nil {
			return errors.WithMessage(err, "更新多语言名称失败")
		}
		// 更新仓库
		err = tx.Model(&model.Warehouse{}).Where("uuid = ?", existingWarehouse.Uuid).Updates(updateData).Error
		if err != nil {
			return errors.WithMessage(err, "更新仓库失败")
		}

		if ctx.GetCompany().IsOpenErp() {
			err = erp.NewIErpSrv(s.dbm).UpdateWarehouse(ctx.GetContext(), req.UpdateErpnextWarehouseReq{
				CreateErpnextWarehouseReq: req.CreateErpnextWarehouseReq{
					SiteCode:      companySetting.ErpnextSiteCode,
					WarehouseName: warehouseName,
					AliasName:     warehouseName,
					CompanyAbbr:   companySetting.ErpnextCompanyAbbr,
					Branch:        companySetting.ErpnextBranchName,
					Disabled:      updateReq.Status == 0,
					WarehouseType: typeMap[updateReq.Type],
				},
				Name: existingWarehouse.ErpCode,
			})
			if err != nil {
				return errors.WithMessage(errors.New("更新仓库失败"), err.Error())
			}
		}

		return nil
	})
	if err != nil {
		return errors.WithMessage(errors.New("更新仓库失败"), err.Error())
	}
	return nil
}

// DeleteWarehouse 删除仓库
func (s *warehouseSrv) DeleteWarehouse(ctx context.Context, deleteWarehouseReq req.DeleteWarehouseReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	warehouseRepo := repository.NewWarehouseRepo(db)

	// 检查仓库是否存在
	existingWarehouse, err := warehouseRepo.GetByUuid(deleteWarehouseReq.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("仓库不存在")
		}
		return errors.WithMessage(err, "获取仓库信息失败")
	}
	if existingWarehouse.IsDefault == 1 {
		return errors.New("默认仓库不可删除")
	}
	// TODO: 这里可以添加业务逻辑检查，比如检查仓库是否有关联数据
	// 例如：检查是否有库存等

	companySetting := ctx.GetCompanySetting()
	if ctx.GetCompany().IsOpenErp() && existingWarehouse.ErpCode != "" {
		err = erp.NewIErpSrv(s.dbm).DeleteWarehouse(ctx.GetContext(), req.DeleteErpnextWarehouseReq{
			SiteCode: companySetting.ErpnextSiteCode,
			Name:     existingWarehouse.ErpCode,
		})
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return errors.WithMessage(errors.New("删除仓库失败"), err.Error())
		}
	}
	// 执行软删除
	err = warehouseRepo.Delete(existingWarehouse.Uuid)
	if err != nil {
		return errors.WithMessage(errors.New("删除仓库失败"), err.Error())
	}

	return nil
}

// buildWarehouseResp 构建仓库响应
func (s *warehouseSrv) buildWarehouseResp(warehouse model.Warehouse) resp.WarehouseResp {
	var localName dto.LocaleResponse
	if warehouse.MultiLanguageName != nil {
		localName = warehouse.MultiLanguageName.GetNames()
	}
	return resp.WarehouseResp{
		Uuid:      warehouse.Uuid,
		LocalName: localName,
		Type:      warehouse.Type,
		Code:      warehouse.Code,
		Status:    warehouse.Status,
		Contact:   warehouse.Contact,
		Phone:     warehouse.Phone,
		Address:   warehouse.Address,
		IsDefault: warehouse.IsDefault,
	}
}

func (s *warehouseSrv) SetDefaultWarehouse(ctx context.Context, req req.SetDefaultWarehouseReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	warehouseRepo := repository.NewWarehouseRepo(db)

	// 检查仓库是否存在
	exists, err := warehouseRepo.GetByUuid(req.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("仓库不存在")
		}
		return errors.WithMessage(err, "获取仓库信息失败")
	}

	// 设置默认仓库
	err = warehouseRepo.UpdateIsDefault(exists.Uuid)
	if err != nil {
		return errors.WithMessage(err, "设置默认仓库失败")
	}

	return nil
}

func (s *warehouseSrv) GetWarehouseInOutList(ctx context.Context, req req.GetWarehouseInOutListReq) (resp.WarehouseInOutListResp, error) {
	return resp.WarehouseInOutListResp{}, nil
}

func (s *warehouseSrv) SyncWarehouse(ctx context.Context) error {
	if !ctx.GetCompany().IsOpenErp() {
		return errors.New("公司未授权erp")
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	translateClient := utils.NewTranslateClient()

	companySetting := ctx.GetCompanySetting()
	warehouseList, err := erp.NewIErpSrv(s.dbm).GetWarehouseList(ctx.GetContext(), req.GetErpnextWarehouseListReq{
		SiteCode:    companySetting.ErpnextSiteCode,
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
		Branch:      companySetting.ErpnextBranchName,
	})
	if err != nil {
		return errors.WithMessage(errors.New("同步仓库失败"), err.Error())
	}

	var translateItems []utils.TranslateItem
	for _, erpWarehouse := range warehouseList {
		translateItems = append(translateItems, utils.TranslateItem{
			Lang:    "en",
			Content: erpWarehouse.WarehouseName,
		})
	}
	multiLanguageMap := translateClient.TranslateWithRetry(ctx.GetContext(), translateItems, 10)

	var syncEver int64
	db.Model(&model.Warehouse{}).Count(&syncEver)

	for _, erpWarehouse := range warehouseList {
		localeName, ok := multiLanguageMap[erpWarehouse.WarehouseName]
		if !ok {
			localeName = dto.LocaleResponse{
				ZH:   erpWarehouse.WarehouseName,
				TH:   erpWarehouse.WarehouseName,
				EN:   erpWarehouse.WarehouseName,
				ZHTW: erpWarehouse.WarehouseName,
				JA:   erpWarehouse.WarehouseName,
				KO:   erpWarehouse.WarehouseName,
				MY:   erpWarehouse.WarehouseName,
				TR:   erpWarehouse.WarehouseName,
				SV:   erpWarehouse.WarehouseName,
			}
		}
		var status int
		if !erpWarehouse.Disabled {
			status = 1
		}
		var code string
		if strings.Contains(erpWarehouse.Name, constant.NormalWarehouseCodeContains) {
			code = constant.NormalWarehouseCode
		} else if strings.Contains(erpWarehouse.Name, constant.TransitWarehouseCodeContains) {
			code = constant.TransitWarehouseCode
		}
		var warehouseType string
		if erpWarehouse.WarehouseType == constant.ErpWarehouseTypeNormal1 || erpWarehouse.WarehouseType == constant.ErpWarehouseTypeNormal2 {
			warehouseType = constant.WarehouseTypeNormal
		} else if erpWarehouse.WarehouseType == constant.ErpWarehouseTypeTransit {
			warehouseType = constant.WarehouseTypeTransit
		}
		var warehouse model.Warehouse
		db.Model(&model.Warehouse{}).Where("erp_code = ?", erpWarehouse.Name).Scopes(repository.NotDeleted).First(&warehouse)
		if warehouse.Uuid > 0 { // 如果存在，则更新
			db.Model(&model.MultiLanguageName{}).Where("uuid = ?", warehouse.MultiLanguageNameUuid).Updates(map[string]any{
				"zh_name":    localeName.ZH,
				"th_name":    localeName.TH,
				"en_name":    localeName.EN,
				"zh_tw_name": localeName.ZHTW,
				"ja_name":    localeName.JA,
				"ko_name":    localeName.KO,
				"my_name":    localeName.MY,
				"tr_name":    localeName.TR,
				"sv_name":    localeName.SV,
			})
			db.Model(&model.Warehouse{}).Where("uuid = ?", warehouse.Uuid).Updates(map[string]any{
				"name":   localeName.ToJson(),
				"type":   warehouseType,
				"status": status,
			})
		} else { // 新增
			var isDefault int
			if code == constant.NormalWarehouseCode && syncEver == 0 {
				isDefault = 1
			}
			// 保存多语言
			multiLanguageName := model.MultiLanguageName{
				ZhName:   localeName.ZH,
				ThName:   localeName.TH,
				EnName:   localeName.EN,
				ZhTwName: localeName.ZHTW,
				JaName:   localeName.JA,
				KoName:   localeName.KO,
				MyName:   localeName.MY,
				TrName:   localeName.TR,
				SvName:   localeName.SV,
			}
			err = db.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
			if err != nil {
				return errors.WithMessage(err, "创建多语言名称失败")
			}
			db.Model(&model.Warehouse{}).Create(&model.Warehouse{
				Name:                  localeName.ToJson(),
				MultiLanguageNameUuid: multiLanguageName.Uuid,
				Type:                  warehouseType,
				Code:                  code,
				Status:                status,
				IsDefault:             isDefault,
			})
		}
	}
	return nil
}

func (s *warehouseSrv) CheckCodeExists(ctx context.Context, req req.CheckCodeExistsReq) (resp.CheckNameCodeExistsResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	warehouseRepo := repository.NewWarehouseRepo(db)
	exists, err := warehouseRepo.IsCodeExists(req.Code, req.Uuid)
	if err != nil {
		return resp.CheckNameCodeExistsResp{}, errors.WithMessage(err, "检查仓库编码是否存在失败")
	}
	return resp.CheckNameCodeExistsResp{Exists: exists}, nil
}
