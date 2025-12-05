package inventory

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/modules/inventory/domain/entity"
	"ttpos-server-go/app/modules/inventory/domain/repository"
	domainService "ttpos-server-go/app/modules/inventory/domain/service"
	"ttpos-server-go/app/modules/inventory/domain/specification"
	"ttpos-server-go/app/modules/inventory/domain/valueobject"
	appService "ttpos-server-go/app/service"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

// WarehouseAppService 仓库应用服务
type WarehouseAppService struct {
	warehouseRepo         repository.IWarehouseRepository
	multiLanguageNameRepo repository.IMultiLanguageNameRepository
	domainService         domainService.IWarehouseDomainService
	erpService            domainService.IErpIntegrationService
	dbm                   *database.DBManager
	settingSrv            setting.ISrv
	translateSrv          appService.ITranslateSrv
	checkNameSrv          appService.ICheckNameSrv
}

// NewWarehouseAppService 创建仓库应用服务
func NewWarehouseAppService(
	warehouseRepo repository.IWarehouseRepository,
	multiLanguageNameRepo repository.IMultiLanguageNameRepository,
	domainService domainService.IWarehouseDomainService,
	erpService domainService.IErpIntegrationService,
	dbm *database.DBManager,
	settingSrv setting.ISrv,
	translateSrv appService.ITranslateSrv,
	checkNameSrv appService.ICheckNameSrv,
) *WarehouseAppService {
	return &WarehouseAppService{
		warehouseRepo:         warehouseRepo,
		multiLanguageNameRepo: multiLanguageNameRepo,
		domainService:         domainService,
		erpService:            erpService,
		dbm:                   dbm,
		settingSrv:            settingSrv,
		translateSrv:          translateSrv,
		checkNameSrv:          checkNameSrv,
	}
}

// CreateWarehouse 创建仓库
func (s *WarehouseAppService) CreateWarehouse(ctx context.Context, addReq req.CreateWarehouseReq) error {
	// 准备领域对象
	code, err := valueobject.NewWarehouseCode(addReq.Code)
	if err != nil {
		return err
	}

	// 检查名称长度和是否存在
	if err := s.validateName(ctx, addReq.LocaleName, 0); err != nil {
		return err
	}

	// 创建领域服务请求
	domainReq := domainService.CreateWarehouseRequest{
		Name:        valueobject.NewMultiLanguageName(addReq.LocaleName),
		Type:        valueobject.NewWarehouseType(addReq.Type),
		Code:        code,
		Status:      valueobject.NewWarehouseStatus(addReq.Status),
		ContactInfo: valueobject.NewContactInfo(addReq.Contact, addReq.Phone, addReq.Address),
	}

	// 开启事务
	return ctx.GetDB().Transaction(func(tx *gorm.DB) error {
		// 调用领域服务创建实体
		warehouse, err := s.domainService.CreateWarehouse(ctx, domainReq)
		if err != nil {
			return err
		}

		// 翻译并保存多语言名称
		multiLanguageName, err := s.saveMultiLanguageName(ctx, tx, addReq.LocaleName)
		if err != nil {
			return err
		}
		warehouse.UpdateName(valueobject.NewMultiLanguageNameWithUuid(multiLanguageName.Uuid, addReq.LocaleName))

		// 如果开启了ERP，调用ERP创建仓库
		if ctx.GetCompany().IsOpenErp() {
			companySetting := s.getErpCompanySetting(ctx)
			erpCode, err := s.erpService.CreateWarehouse(ctx, warehouse, companySetting)
			if err != nil {
				return errors.WithMessage(err, "创建仓库失败")
			}
			warehouse.SetErpCode(erpCode)
		}

		// 持久化
		if err := s.warehouseRepo.Save(ctx, warehouse); err != nil {
			return err
		}

		return nil
	})
}

// UpdateWarehouse 更新仓库
func (s *WarehouseAppService) UpdateWarehouse(ctx context.Context, updateReq req.UpdateWarehouseReq) error {
	// 查找仓库
	warehouse, err := s.warehouseRepo.FindByUuid(ctx, updateReq.Uuid)
	if err != nil {
		return errors.WithMessage(err, "获取仓库信息失败")
	}
	if warehouse == nil {
		return errors.New("仓库不存在")
	}

	// 检查是否可编辑
	if err := warehouse.CanBeEdited(ctx.GetCompanyUuid()); err != nil {
		return err
	}

	// 准备更新请求
	code, err := valueobject.NewWarehouseCode(updateReq.Code)
	if err != nil {
		return err
	}

	// 检查名称
	if err := s.validateName(ctx, updateReq.LocaleName, warehouse.Uuid()); err != nil {
		return err
	}

	domainReq := domainService.UpdateWarehouseRequest{
		Name:        valueobject.NewMultiLanguageName(updateReq.LocaleName),
		Type:        valueobject.NewWarehouseType(updateReq.Type),
		Code:        code,
		Status:      valueobject.NewWarehouseStatus(updateReq.Status),
		ContactInfo: valueobject.NewContactInfo(updateReq.Contact, updateReq.Phone, updateReq.Address),
	}

	// 验证更新操作
	if err := s.domainService.ValidateForUpdate(ctx, warehouse, domainReq); err != nil {
		return err
	}

	// 开启事务
	return ctx.GetDB().Transaction(func(tx *gorm.DB) error {
		// 更新多语言名称
		if err := s.updateMultiLanguageName(ctx, tx, warehouse.Name().Uuid(), updateReq.LocaleName); err != nil {
			return err
		}

		// 更新仓库属性
		warehouse.UpdateName(valueobject.NewMultiLanguageNameWithUuid(warehouse.Name().Uuid(), updateReq.LocaleName))
		warehouse.UpdateType(domainReq.Type)
		warehouse.UpdateStatus(domainReq.Status)
		warehouse.UpdateContactInfo(domainReq.ContactInfo)

		// 如果开启了ERP，调用ERP更新仓库
		if ctx.GetCompany().IsOpenErp() && warehouse.ErpCode() != "" {
			companySetting := s.getErpCompanySetting(ctx)
			if err := s.erpService.UpdateWarehouse(ctx, warehouse, companySetting); err != nil {
				return errors.WithMessage(err, "更新仓库失败")
			}
		}

		// 持久化
		if err := s.warehouseRepo.Save(ctx, warehouse); err != nil {
			return err
		}

		// 从待翻译集合中删除
		s.translateSrv.RemoveMultiLanguageNameUuidFromSet(ctx.GetCompanyUuid(), warehouse.Name().Uuid())

		return nil
	})
}

// DeleteWarehouse 删除仓库
func (s *WarehouseAppService) DeleteWarehouse(ctx context.Context, deleteReq req.DeleteWarehouseReq) error {
	// 查找仓库
	warehouse, err := s.warehouseRepo.FindByUuid(ctx, deleteReq.Uuid)
	if err != nil {
		return errors.WithMessage(err, "获取仓库信息失败")
	}
	if warehouse == nil {
		return errors.New("仓库不存在")
	}

	// 验证删除操作
	if err := s.domainService.ValidateForDelete(ctx, warehouse, ctx.GetCompanyUuid()); err != nil {
		return err
	}

	// 如果开启了ERP，调用ERP删除仓库
	if ctx.GetCompany().IsOpenErp() && warehouse.ErpCode() != "" {
		companySetting := s.getErpCompanySetting(ctx)
		if err := s.erpService.DeleteWarehouse(ctx, warehouse.ErpCode(), companySetting); err != nil {
			return errors.WithMessage(err, "删除仓库失败")
		}
	}

	// 删除仓库
	if err := s.warehouseRepo.Remove(ctx, warehouse.Uuid()); err != nil {
		return errors.WithMessage(err, "删除仓库失败")
	}

	return nil
}

// SetDefaultWarehouse 设置默认仓库
func (s *WarehouseAppService) SetDefaultWarehouse(ctx context.Context, setReq req.SetDefaultWarehouseReq) error {
	// 查找仓库
	warehouse, err := s.warehouseRepo.FindByUuid(ctx, setReq.Uuid)
	if err != nil {
		return errors.WithMessage(err, "获取仓库信息失败")
	}
	if warehouse == nil {
		return errors.New("仓库不存在")
	}

	// 检查是否可编辑
	if err := warehouse.CanBeEdited(ctx.GetCompanyUuid()); err != nil {
		return errors.New("仓库不可设置为默认仓库")
	}

	// 调用领域服务设置默认仓库
	return s.domainService.SetDefaultWarehouse(ctx, setReq.Uuid)
}

// GetWarehouseList 获取仓库列表
func (s *WarehouseAppService) GetWarehouseList(ctx context.Context, listReq req.WarehouseListReq, isHeadquarters ...bool) (resp.WarehouseListResp, error) {
	isHeadquarter := false
	if len(isHeadquarters) > 0 {
		isHeadquarter = isHeadquarters[0]
	}

	// 构建查询规格
	spec := specification.NewWarehouseQuerySpec()

	if listReq.Keyword != "" {
		spec = spec.WithKeyword(listReq.Keyword)
	}

	if listReq.Type != "" {
		spec = spec.WithType(valueobject.NewWarehouseType(listReq.Type))
	}

	if listReq.Status != nil {
		spec = spec.WithStatus(valueobject.NewWarehouseStatus(*listReq.Status))
	}

	if isHeadquarter {
		spec = spec.WithIsHeadquarter(true)
	} else {
		spec = spec.WithHeadquarterUuid(0)
	}

	// 查询仓库列表
	warehouses, total, err := s.warehouseRepo.FindWithPagination(ctx, spec, listReq.PageNo, listReq.PageSize)
	if err != nil {
		return resp.WarehouseListResp{}, errors.WithMessage(err, "获取仓库列表失败")
	}

	// 转换为响应
	list := make([]resp.WarehouseResp, 0, len(warehouses))
	for _, warehouse := range warehouses {
		list = append(list, s.buildWarehouseResp(ctx, warehouse, isHeadquarter))
	}

	return resp.WarehouseListResp{
		List: list,
		Meta: dto.PageResponse{
			PageNo:   listReq.PageNo,
			PageSize: listReq.PageSize,
			Total:    total,
		},
	}, nil
}

// GetWarehouse 获取仓库详情
func (s *WarehouseAppService) GetWarehouse(ctx context.Context, getReq req.WarehouseReq) (resp.WarehouseResp, error) {
	warehouse, err := s.warehouseRepo.FindByUuid(ctx, getReq.Uuid)
	if err != nil {
		return resp.WarehouseResp{}, errors.WithMessage(err, "获取仓库失败")
	}
	if warehouse == nil {
		return resp.WarehouseResp{}, errors.New("仓库不存在")
	}

	return s.buildWarehouseResp(ctx, warehouse, false), nil
}

// 辅助方法

// validateName 验证名称
func (s *WarehouseAppService) validateName(ctx context.Context, localeName dto.LocaleResponse, excludeUuid uint64) error {
	names := s.checkNameSrv.MakeCheckNameList(ctx, localeName)
	for _, name := range names {
		if !s.checkNameSrv.CheckNameLength(ctx, name.Text, 140) {
			return errors.New("仓库名称长度不能超过140")
		}
	}

	exists := s.checkNameSrv.InnerCheckNameExists(ctx, req.CheckNameRequest{
		Uuid:   excludeUuid,
		Source: "warehouse",
		Names:  names,
	})
	if exists {
		return errors.New("仓库名称已存在")
	}

	return nil
}

// saveMultiLanguageName 保存多语言名称
func (s *WarehouseAppService) saveMultiLanguageName(ctx context.Context, tx *gorm.DB, localeName dto.LocaleResponse) (*repository.MultiLanguageNameData, error) {
	// 翻译英文名称
	enName, err := appService.GetEnName(ctx, s.settingSrv, localeName)
	if err != nil {
		return nil, errors.WithMessage(errors.New("翻译失败"), err.Error())
	}

	// 使用 Repository 保存
	data := &repository.MultiLanguageNameData{
		ZH:   localeName.ZH,
		TH:   localeName.TH,
		EN:   enName,
		ZHTW: localeName.ZHTW,
		JA:   localeName.JA,
		KO:   localeName.KO,
		MY:   localeName.MY,
		TR:   localeName.TR,
		SV:   localeName.SV,
	}

	if err := s.multiLanguageNameRepo.Save(ctx, data); err != nil {
		return nil, errors.WithMessage(err, "创建多语言名称失败")
	}

	return data, nil
}

// updateMultiLanguageName 更新多语言名称
func (s *WarehouseAppService) updateMultiLanguageName(ctx context.Context, tx *gorm.DB, uuid uint64, localeName dto.LocaleResponse) error {
	data := &repository.MultiLanguageNameData{
		Uuid: uuid,
		ZH:   localeName.ZH,
		TH:   localeName.TH,
		EN:   localeName.EN,
		ZHTW: localeName.ZHTW,
		JA:   localeName.JA,
		KO:   localeName.KO,
		MY:   localeName.MY,
		TR:   localeName.TR,
		SV:   localeName.SV,
	}

	return s.multiLanguageNameRepo.Save(ctx, data)
}

// getErpCompanySetting 获取ERP公司设置
func (s *WarehouseAppService) getErpCompanySetting(ctx context.Context) domainService.ErpCompanySetting {
	companySetting := ctx.GetCompanySetting()
	return domainService.ErpCompanySetting{
		SiteCode:    companySetting.ErpnextSiteCode,
		CompanyAbbr: companySetting.ErpnextCompanyAbbr,
		BranchName:  companySetting.ErpnextBranchName,
	}
}

// CheckCodeExists 检查仓库编码是否存在
func (s *WarehouseAppService) CheckCodeExists(ctx context.Context, checkReq req.CheckCodeExistsReq) (resp.CheckNameCodeExistsResp, error) {
	// 验证编码格式
	code, err := valueobject.NewWarehouseCode(checkReq.Code)
	if err != nil {
		return resp.CheckNameCodeExistsResp{Exists: false}, nil
	}

	// 检查编码是否存在
	exists, err := s.warehouseRepo.ExistsCode(ctx, code, checkReq.Uuid)
	if err != nil {
		return resp.CheckNameCodeExistsResp{}, errors.WithMessage(err, "检查仓库编码失败")
	}

	return resp.CheckNameCodeExistsResp{Exists: exists}, nil
}

// buildWarehouseResp 构建仓库响应
func (s *WarehouseAppService) buildWarehouseResp(ctx context.Context, warehouse *entity.Warehouse, isHeadquarter bool) resp.WarehouseResp {
	code := warehouse.Code().Value()
	if isHeadquarter {
		code = warehouse.ErpCode()
	}

	isDefault := 0
	if warehouse.IsDefault() {
		isDefault = 1
	}

	return resp.WarehouseResp{
		Uuid:       warehouse.Uuid(),
		LocalName:  warehouse.Name().ToLocaleResponse(),
		Type:       warehouse.WarehouseType().String(),
		Code:       code,
		Status:     warehouse.Status().ToInt(),
		Contact:    warehouse.ContactInfo().Contact(),
		Phone:      warehouse.ContactInfo().Phone(),
		Address:    warehouse.ContactInfo().Address(),
		IsDefault:  isDefault,
		IsEditable: warehouse.CanBeEdited(ctx.GetCompanyUuid()) == nil,
		HasItem:    warehouse.HasItems(),
	}
}
