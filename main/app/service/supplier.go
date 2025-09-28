package service

import (
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
	"ttpos-server-go/pkg/utils"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// ISupplierSrv 供应商服务接口
type ISupplierSrv interface {
	GetSupplierList(ctx context.Context, req req.SupplierListReq) (resp.SupplierListResp, error)           // 供应商列表
	CreateSupplier(ctx context.Context, req req.SupplierCreateReq) error                                   // 创建供应商
	UpdateSupplier(ctx context.Context, req req.SupplierUpdateReq) error                                   // 更新供应商
	DeleteSupplier(ctx context.Context, req req.SupplierDeleteReq) error                                   // 删除供应商
	GetSupplierSelect(ctx context.Context, req req.SupplierSelectReq) (resp.SupplierSelectResp, error)     // 获取供应商选择器列表
	GetSupplier(ctx context.Context, req req.SupplierReq) (resp.SupplierResp, error)                       // 获取供应商
	CheckCodeExists(ctx context.Context, req req.CheckCodeExistsReq) (resp.CheckNameCodeExistsResp, error) // 检查编码是否存在

	SyncSupplier(ctx context.Context) error // 同步供应商
}

// NewSupplierSrv 创建供应商服务
func NewSupplierSrv(dbm *database.DBManager) ISupplierSrv {
	return NewSupplierSrvImpl(dbm)
}

// supplierSrv 供应商服务实现
type supplierSrv struct {
	dbm *database.DBManager
}

// NewSupplierSrvImpl 创建供应商服务实现
func NewSupplierSrvImpl(dbm *database.DBManager) ISupplierSrv {
	return &supplierSrv{
		dbm: dbm,
	}
}

// GetSupplierList 获取供应商列表
func (s *supplierSrv) GetSupplierList(ctx context.Context, req req.SupplierListReq) (resp.SupplierListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)
	// 构建查询选项
	var opts []repository.DBOption
	// 名称编码筛选
	if req.Keyword != "" {
		opts = append(opts, supplierRepo.WhereNameOrCodeLike(req.Keyword))
	}
	// 排序
	opts = append(opts, supplierRepo.OrderByCreateTime(true))
	// 分页查询
	suppliers, total, err := supplierRepo.GetListWithPagination(
		req.PageReq.PageNo,
		req.PageReq.PageSize,
		opts...,
	)
	if err != nil {
		return resp.SupplierListResp{}, errors.WithMessage(err, "获取供应商列表失败")
	}
	// 转换响应格式
	supplierList := make([]*resp.SupplierInfo, 0, len(suppliers))
	for _, supplier := range suppliers {
		supplierInfo := &resp.SupplierInfo{}
		err := copier.Copy(supplierInfo, &supplier)
		if err != nil {
			continue
		}
		var localeName dto.LocaleResponse
		if supplier.MultiLanguageName != nil {
			localeName = supplier.MultiLanguageName.GetNames()
		}
		supplierInfo.LocaleName = localeName
		supplierInfo.IsEditable = isEditable(ctx, supplier.HeadquarterUuid)
		supplierInfo.IsHeadquarter = supplier.ErpCode == constant.ErpHeadquartersSupplierCode
		supplierList = append(supplierList, supplierInfo)
	}
	return resp.SupplierListResp{
		List: supplierList,
		Meta: dto.PageResponse{
			PageNo:   req.PageReq.PageNo,
			PageSize: req.PageReq.PageSize,
			Total:    total,
		},
	}, nil
}

// GetSupplier 获取供应商
func (s *supplierSrv) GetSupplier(ctx context.Context, req req.SupplierReq) (resp.SupplierResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)
	supplier, err := supplierRepo.GetByUuid(req.Uuid)
	if err != nil {
		return resp.SupplierResp{}, errors.WithMessage(err, "获取供应商失败")
	}

	var localeName dto.LocaleResponse
	if supplier.MultiLanguageName != nil {
		localeName = supplier.MultiLanguageName.GetNames()
	}
	return resp.SupplierResp{
		SupplierInfo: &resp.SupplierInfo{
			Uuid:          supplier.Uuid,
			LocaleName:    localeName,
			Code:          supplier.Code,
			IsHeadquarter: supplier.ErpCode == constant.ErpHeadquartersSupplierCode,
			IsEditable:    isEditable(ctx, supplier.HeadquarterUuid),
		},
		Address:                 supplier.Address,
		ContactName:             supplier.ContactName,
		ContactPhone:            supplier.ContactPhone,
		Status:                  supplier.Status,
		HasRelatedPurchaseOrder: s.hasRelatedPurchaseOrder(ctx, supplier),
	}, nil
}

func (s *supplierSrv) hasRelatedPurchaseOrder(_ context.Context, supplier *model.Supplier) bool {
	return supplier.PurchaseOrders != nil && len(supplier.PurchaseOrders) > 0
}

// CreateSupplier 创建供应商
func (s *supplierSrv) CreateSupplier(ctx context.Context, createSupplierReq req.SupplierCreateReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)
	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, createSupplierReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Source: constant.CheckNameSourceUnit,
		Names:  names,
	})
	if exists {
		return errors.New("供应商名称已存在")
	}
	for _, name := range names {
		if !checkService.CheckNameLength(ctx, name.Text, 150) {
			return errors.New("供应商名称长度不能超过150")
		}
	}
	// 检查供应商编码是否重复
	codeExists, err := supplierRepo.IsCodeExists(createSupplierReq.Code, 0)
	if err != nil {
		return errors.WithMessage(err, "检查供应商编码失败")
	}
	if codeExists {
		return errors.New("供应商编码已存在")
	}
	var erpCode string
	// 调用erp接口
	if ctx.GetCompany().IsOpenErp() {
		enName, err := GetEnName(ctx, createSupplierReq.LocaleName)
		if err != nil {
			return errors.WithMessage(errors.New("翻译失败"), err.Error())
		}
		companySetting := ctx.GetCompanySetting()
		erpCode, err = erp.NewIErpSrv(s.dbm).CreateSupplier(ctx.GetContext(), req.CreateSupplierReq{
			SiteCode:     companySetting.ErpnextSiteCode,
			SupplierName: enName,
			CompanyAbbr:  companySetting.ErpnextCompanyAbbr,
			Branch:       companySetting.ErpnextBranchName,
			Disabled:     createSupplierReq.Status == 0,
		})
		if err != nil {
			return errors.WithMessage(errors.New("创建供应商失败"), err.Error())
		}
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		multiLanguageName := model.MultiLanguageName{
			EnName:   createSupplierReq.LocaleName.EN,
			ZhName:   createSupplierReq.LocaleName.ZH,
			ThName:   createSupplierReq.LocaleName.TH,
			MyName:   createSupplierReq.LocaleName.MY,
			JaName:   createSupplierReq.LocaleName.JA,
			KoName:   createSupplierReq.LocaleName.KO,
			TrName:   createSupplierReq.LocaleName.TR,
			SvName:   createSupplierReq.LocaleName.SV,
			ZhTwName: createSupplierReq.LocaleName.ZHTW,
		}
		err := tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName).Error
		if err != nil {
			return errors.WithMessage(errors.New("创建多语言名称失败"), err.Error())
		}
		// 创建供应商
		err = tx.Model(&model.Supplier{}).Create(&model.Supplier{
			Name:                  createSupplierReq.LocaleName.ToJson(),
			MultiLanguageNameUuid: multiLanguageName.Uuid,
			Code:                  createSupplierReq.Code,
			Address:               createSupplierReq.Address,
			ContactName:           createSupplierReq.ContactName,
			ContactPhone:          createSupplierReq.ContactPhone,
			Status:                createSupplierReq.Status,
			ErpCode:               erpCode,
		}).Error
		if err != nil {
			return errors.WithMessage(errors.New("创建供应商失败"), err.Error())
		}
		return nil
	})

	return err
}

// UpdateSupplier 更新供应商
func (s *supplierSrv) UpdateSupplier(ctx context.Context, updateSupplierReq req.SupplierUpdateReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)

	// 检查供应商是否存在
	supplier, err := supplierRepo.GetByUuid(updateSupplierReq.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("供应商不存在")
		}
		return errors.WithMessage(err, "查询供应商失败")
	}

	if !isEditable(ctx, supplier.HeadquarterUuid) {
		return errors.New("供应商不可编辑")
	}

	checkService := NewCheckNameSrv(s.dbm)
	names := checkService.MakeCheckNameList(ctx, updateSupplierReq.LocaleName)
	exists := checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Uuid:   supplier.Uuid,
		Source: constant.CheckNameSourceUnit,
		Names:  names,
	})
	if exists {
		return errors.New("供应商名称已存在")
	}
	for _, name := range names {
		if !checkService.CheckNameLength(ctx, name.Text, 150) {
			return errors.New("供应商名称长度不能超过150")
		}
	}
	// 检查供应商编码是否重复（排除自己）
	codeExists, err := supplierRepo.IsCodeExists(updateSupplierReq.Code, updateSupplierReq.Uuid)
	if err != nil {
		return errors.WithMessage(err, "检查供应商编码失败")
	}
	if codeExists {
		return errors.New("供应商编码已存在")
	}

	// 调用erp接口，只能修改自己创建的供应商
	if ctx.GetCompany().IsOpenErp() {
		enName, err := GetEnName(ctx, updateSupplierReq.LocaleName)
		if err != nil {
			return errors.WithMessage(errors.New("翻译失败"), err.Error())
		}
		companySetting := ctx.GetCompanySetting()
		err = erp.NewIErpSrv(s.dbm).UpdateSupplier(ctx.GetContext(), req.UpdateSupplierReq{
			CreateSupplierReq: req.CreateSupplierReq{
				SupplierName: enName,
				SiteCode:     companySetting.ErpnextSiteCode,
				CompanyAbbr:  companySetting.ErpnextCompanyAbbr,
				Branch:       companySetting.ErpnextBranchName,
				Disabled:     updateSupplierReq.Status == 0,
			},
			Name: supplier.ErpCode,
		})
		if err != nil {
			return errors.WithMessage(errors.New("更新供应商失败"), err.Error())
		}
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", supplier.MultiLanguageNameUuid).Updates(map[string]any{
			"zh_name":    updateSupplierReq.LocaleName.ZH,
			"th_name":    updateSupplierReq.LocaleName.TH,
			"en_name":    updateSupplierReq.LocaleName.EN,
			"zh_tw_name": updateSupplierReq.LocaleName.ZHTW,
			"ja_name":    updateSupplierReq.LocaleName.JA,
			"ko_name":    updateSupplierReq.LocaleName.KO,
			"my_name":    updateSupplierReq.LocaleName.MY,
			"tr_name":    updateSupplierReq.LocaleName.TR,
			"sv_name":    updateSupplierReq.LocaleName.SV,
		}).Error
		if err != nil {
			return errors.WithMessage(err, "更新多语言名称失败")
		}
		err = tx.Model(&model.Supplier{}).Where("uuid = ?", supplier.Uuid).Updates(map[string]any{
			"name":          updateSupplierReq.LocaleName.ToJson(),
			"address":       updateSupplierReq.Address,
			"contact_name":  updateSupplierReq.ContactName,
			"contact_phone": updateSupplierReq.ContactPhone,
			"status":        updateSupplierReq.Status,
		}).Error
		if err != nil {
			return errors.WithMessage(err, "更新供应商失败")
		}
		return nil
	})
	return err
}

// DeleteSupplier 删除供应商
func (s *supplierSrv) DeleteSupplier(ctx context.Context, deleteSupplierReq req.SupplierDeleteReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)

	// 检查供应商是否存在
	supplier, err := supplierRepo.GetByUuid(deleteSupplierReq.Uuid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("供应商不存在")
		}
		return errors.WithMessage(err, "查询供应商失败")
	}
	if !isEditable(ctx, supplier.HeadquarterUuid) {
		return errors.New("供应商不可删除")
	}
	if s.hasRelatedPurchaseOrder(ctx, supplier) {
		return errors.New("该供应商存在关联的采购订单，无法删除")
	}

	companySetting := ctx.GetCompanySetting()
	// 调用erp接口，只能删除自己创建的供应商
	if ctx.GetCompany().IsOpenErp() {
		err = erp.NewIErpSrv(s.dbm).DeleteSupplier(ctx.GetContext(), req.DeleteSupplierReq{
			SiteCode: companySetting.ErpnextSiteCode,
			Name:     supplier.ErpCode,
		})
		if err != nil {
			return errors.WithMessage(errors.New("删除供应商失败"), err.Error())
		}
	}
	// 软删除供应商
	err = supplierRepo.Delete(deleteSupplierReq.Uuid)
	if err != nil {
		return errors.WithMessage(err, "删除供应商失败")
	}
	return nil
}

// GetSupplierSelect 获取供应商选择器列表
func (s *supplierSrv) GetSupplierSelect(ctx context.Context, req req.SupplierSelectReq) (resp.SupplierSelectResp, error) {

	if ctx.GetCompany().IsOpenErp() && ctx.Version(context.LT, "2.6.0") {
		erpResp, err := erp.NewIErpSrv(s.dbm).GetSupplierList(ctx)
		if err != nil {
			return resp.SupplierSelectResp{}, errors.WithMessage(err, "获取供应商选择器列表失败")
		}
		// 转换响应格式
		var supplierList []*resp.SupplierSimpleInfo
		for _, supplier := range erpResp.SupplierList {
			supplierList = append(supplierList, &resp.SupplierSimpleInfo{
				Name: supplier.SupplierName,
				Code: supplier.Name,
			})
		}
		return resp.SupplierSelectResp{
			List: supplierList,
		}, nil
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)

	// 构建查询选项
	opts := []repository.DBOption{
		supplierRepo.OrderByName(false), // 按名称升序排序
		supplierRepo.WhereNotDeleted(),
	}

	// 如果公司开启了erp，则查询erp供应商
	if ctx.GetCompany().IsOpenErp() {
		opts = append(opts, supplierRepo.WhereErpCodeExists())
	}

	suppliers, err := supplierRepo.GetList(opts...)
	if err != nil {
		return resp.SupplierSelectResp{}, errors.WithMessage(err, "获取供应商选择器列表失败")
	}

	// 转换响应格式
	var supplierList []*resp.SupplierSimpleInfo
	for _, supplier := range suppliers {
		// 外部采购 去掉总部
		if req.PurchaseType == 1 {
			if supplier.ErpCode == constant.ErpHeadquartersSupplierCode {
				continue
			}
		} else {
			// 内部采购 去掉非总部
			if supplier.ErpCode != constant.ErpHeadquartersSupplierCode {
				continue
			}
		}
		supplierList = append(supplierList, &resp.SupplierSimpleInfo{
			Uuid: supplier.Uuid,
			Name: supplier.Name,
			Code: supplier.ErpCode,
		})
	}

	return resp.SupplierSelectResp{
		List: supplierList,
	}, nil
}

func (s *supplierSrv) CheckCodeExists(ctx context.Context, req req.CheckCodeExistsReq) (resp.CheckNameCodeExistsResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierRepo := repository.NewSupplierRepo(db)
	exists, err := supplierRepo.IsCodeExists(req.Code, req.Uuid)
	if err != nil {
		return resp.CheckNameCodeExistsResp{}, errors.WithMessage(err, "检查编码是否存在失败")
	}
	return resp.CheckNameCodeExistsResp{Exists: exists}, nil
}

func (s *supplierSrv) SyncSupplier(ctx context.Context) error {
	if !ctx.GetCompany().IsOpenErp() {
		return errors.New("公司未授权erp")
	}
	companySetting := ctx.GetCompanySetting()
	db := s.dbm.GetDB(ctx.GetDbId())
	supplierList, err := erp.NewIErpSrv(s.dbm).ListSuppliers(ctx, req.GetErpnextSupplierListReq{
		SiteCode:    companySetting.ErpnextSiteCode,
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
		Branch:      companySetting.ErpnextBranchName,
	})
	if err != nil {
		return errors.WithMessage(errors.New("同步供应商失败"), err.Error())
	}

	supplierHeadquarterMap := make(map[string]uint64)
	// 如果是子店，获取总部允许子店看到的
	if companySetting.IsSubShop() {
		var headquarter model.CompanySetting
		saasDb := s.dbm.GetDB(0)
		err := saasDb.Model(&model.CompanySetting{}).Where("uuid = ?", companySetting.HeadquarterUuid).Scopes(repository.NotDeleted).Debug().First(&headquarter).Error
		if err != nil || headquarter.Uuid == 0 {
			return errors.WithMessage(errors.New("获取总部公司失败"))
		}
		headquarterSupplierList, err := erp.NewIErpSrv(s.dbm).ListSuppliers(ctx, req.GetErpnextSupplierListReq{
			SiteCode:       headquarter.ErpnextSiteCode,
			CompanyAbbr:    headquarter.ErpnextCompanyAbbr,
			Branch:         headquarter.ErpnextBranchName,
			SubCompanyAbbr: companySetting.ErpnextCompanyAbbr,
		})
		if err != nil {
			return errors.WithMessage(errors.New("获取总部供应商失败"), err.Error())
		}
		for _, supplier := range headquarterSupplierList {
			supplierHeadquarterMap[supplier.Name] = headquarter.Uuid
		}
		supplierList = append(supplierList, headquarterSupplierList...)
	}

	// 翻译供应商名称
	var translateItems []utils.TranslateItem
	for _, erpSupplier := range supplierList {
		translateItems = append(translateItems, utils.TranslateItem{
			Lang:    "en",
			Content: erpSupplier.SupplierName,
		})
	}
	translateClient := utils.NewTranslateClient()
	multiLanguageMap := translateClient.TranslateWithRetry(ctx.GetContext(), translateItems, 10)

	var suppliers []model.Supplier
	db.Model(&model.Supplier{}).Scopes(repository.NotDeleted).Find(&suppliers)

	supplierMap := make(map[string]model.Supplier)
	for _, supplier := range suppliers {
		if supplier.ErpCode == "" {
			continue
		}
		supplierMap[supplier.ErpCode] = supplier
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, erpSupplier := range supplierList {

			// 总部不同步“总部-供应商”
			if erpSupplier.Name == constant.ErpHeadquartersSupplierCode && companySetting.IsHeadquarter() {
				continue
			}

			var headquarterCode string
			if erpSupplier.Name == constant.ErpHeadquartersSupplierCode {
				headquarterCode = "SP001"
			}

			localeName, ok := multiLanguageMap[erpSupplier.SupplierName]
			if !ok {
				localeName = dto.LocaleResponse{
					EN:   erpSupplier.SupplierName,
					ZH:   erpSupplier.SupplierName,
					TH:   erpSupplier.SupplierName,
					MY:   erpSupplier.SupplierName,
					JA:   erpSupplier.SupplierName,
					KO:   erpSupplier.SupplierName,
					TR:   erpSupplier.SupplierName,
					SV:   erpSupplier.SupplierName,
					ZHTW: erpSupplier.SupplierName,
				}
			}
			supplier, _ := supplierMap[erpSupplier.Name]
			headquarterUuid, _ := supplierHeadquarterMap[erpSupplier.Name]
			if supplier.Uuid == 0 { // 添加
				multiLanguageName := model.MultiLanguageName{
					EnName:   localeName.EN,
					ZhName:   localeName.ZH,
					ThName:   localeName.TH,
					MyName:   localeName.MY,
					JaName:   localeName.JA,
					KoName:   localeName.KO,
					TrName:   localeName.TR,
					SvName:   localeName.SV,
					ZhTwName: localeName.ZHTW,
				}
				tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName)
				tx.Model(&model.Supplier{}).Create(&model.Supplier{
					Name:                  multiLanguageName.ToJson(),
					MultiLanguageNameUuid: multiLanguageName.Uuid,
					ErpCode:               erpSupplier.Name,
					Status:                1,
					HeadquarterUuid:       headquarterUuid,
					Code:                  headquarterCode,
				})
			} else { // 编辑
				tx.Model(&model.MultiLanguageName{}).Where("uuid = ?", supplier.MultiLanguageNameUuid).Updates(map[string]any{
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
				tx.Model(&model.Supplier{}).Where("uuid = ?", supplier.Uuid).Updates(map[string]any{
					"name":             erpSupplier.SupplierName,
					"headquarter_uuid": headquarterUuid,
					"code":             headquarterCode,
				})
			}
		}
		return nil
	})

	return err
}
