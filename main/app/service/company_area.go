package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

// ICompanyAreaSrv 门店区域服务接口
type ICompanyAreaSrv interface {
	GetList(ctx context.Context) (resp.CompanyAreaListResp, error)
	GetInfo(ctx context.Context, areaUuid uint64) (resp.CompanyAreaInfoResp, error)
	Create(ctx context.Context, createReq req.CompanyAreaCreateReq) error
	Update(ctx context.Context, updateReq req.CompanyAreaUpdateReq) error
	Delete(ctx context.Context, deleteReq req.CompanyAreaDeleteReq) error
	GetUnassigned(ctx context.Context) (resp.CompanyAreaUnassignedResp, error)
	GetShopList(ctx context.Context) (resp.CompanyAreaShopListResp, error)
}

type companyAreaSrv struct {
	dbm *database.DBManager
}

func NewCompanyAreaSrv(dbm *database.DBManager) ICompanyAreaSrv {
	return &companyAreaSrv{dbm: dbm}
}

// getHeadquarterUuid 获取总部UUID（兼容总部和子店调用）
func (s *companyAreaSrv) getHeadquarterUuid(ctx context.Context) uint64 {
	companySetting := ctx.GetCompanySetting()
	if companySetting.IsHeadquarter() {
		return ctx.GetCompanyUuid()
	}
	return companySetting.HeadquarterUuid
}

// requireHeadquarter 校验当前调用方必须是总部
func (s *companyAreaSrv) requireHeadquarter(ctx context.Context) error {
	companySetting := ctx.GetCompanySetting()
	if !companySetting.IsHeadquarter() {
		return errors.NewWithCode(constant.CodeAccessDenied, "仅总部可操作区域管理")
	}
	return nil
}

// getAllSettingsByHeadquarter 获取总部下所有门店设置
func (s *companyAreaSrv) getAllSettingsByHeadquarter(ctx context.Context) ([]model.CompanySetting, error) {
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	return repository.NewCompanySettingRepo(saasDB).GetAllByHeadquarterUuid(s.getHeadquarterUuid(ctx))
}

// fetchShopsByUuids 批量获取门店信息并转换为 CompanyAreaShop
func (s *companyAreaSrv) fetchShopsByUuids(uuids []uint64) ([]resp.CompanyAreaShop, error) {
	list := make([]resp.CompanyAreaShop, 0)
	if len(uuids) == 0 {
		return list, nil
	}
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	companies, err := repository.NewCompanyRepo(saasDB).GetByUuids(uuids)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	for _, company := range companies {
		list = append(list, resp.CompanyAreaShop{
			Uuid:   company.Uuid,
			Name:   company.Name,
			Status: company.Status,
		})
	}
	return list, nil
}

// GetList 获取区域列表（含门店数量）
func (s *companyAreaSrv) GetList(ctx context.Context) (resp.CompanyAreaListResp, error) {
	if err := s.requireHeadquarter(ctx); err != nil {
		return resp.CompanyAreaListResp{}, err
	}
	db := ctx.GetDB()
	areaRepo := repository.NewCompanyAreaRepo(db)

	areas, err := areaRepo.GetList(
		repository.CommonRepo.Preload(repository.WithPreload{Query: "MultiLanguageName"}),
	)
	if err != nil {
		return resp.CompanyAreaListResp{}, errors.WithMessage(err)
	}

	allSettings, err := s.getAllSettingsByHeadquarter(ctx)
	if err != nil {
		return resp.CompanyAreaListResp{}, errors.WithMessage(err)
	}

	// 统计每个区域的门店数量 + 摘要
	totalShopCount := len(allSettings)
	unassignedShopCount := 0
	areaShopCount := make(map[uint64]int)
	for _, setting := range allSettings {
		if setting.CompanyAreaUuid > 0 {
			areaShopCount[setting.CompanyAreaUuid]++
		} else {
			unassignedShopCount++
		}
	}

	list := make([]resp.CompanyAreaItem, 0, len(areas))
	for _, area := range areas {
		item := resp.CompanyAreaItem{
			Uuid:      area.Uuid,
			Sort:      area.Sort,
			ShopCount: areaShopCount[area.Uuid],
		}
		if area.MultiLanguageName != nil {
			item.LocaleName = area.MultiLanguageName.GetNames()
		}
		list = append(list, item)
	}

	return resp.CompanyAreaListResp{
		TotalShopCount:      totalShopCount,
		UnassignedShopCount: unassignedShopCount,
		List:                list,
	}, nil
}

// GetInfo 获取区域详情（含门店列表）
func (s *companyAreaSrv) GetInfo(ctx context.Context, areaUuid uint64) (resp.CompanyAreaInfoResp, error) {
	if err := s.requireHeadquarter(ctx); err != nil {
		return resp.CompanyAreaInfoResp{}, err
	}
	db := ctx.GetDB()
	areaRepo := repository.NewCompanyAreaRepo(db)

	area, err := areaRepo.GetByUuid(areaUuid)
	if err != nil {
		return resp.CompanyAreaInfoResp{}, errors.WithMessage(err)
	}

	result := resp.CompanyAreaInfoResp{
		Uuid: area.Uuid,
		Sort: area.Sort,
	}
	if area.MultiLanguageName != nil {
		result.LocaleName = area.MultiLanguageName.GetNames()
	}

	allSettings, err := s.getAllSettingsByHeadquarter(ctx)
	if err != nil {
		return resp.CompanyAreaInfoResp{}, errors.WithMessage(err)
	}

	var companyUuids []uint64
	for _, setting := range allSettings {
		if setting.CompanyAreaUuid == areaUuid {
			companyUuids = append(companyUuids, setting.CompanyUuid)
		}
	}

	shopList, err := s.fetchShopsByUuids(companyUuids)
	if err != nil {
		return resp.CompanyAreaInfoResp{}, err
	}
	result.ShopList = shopList

	return result, nil
}

// Create 创建区域
func (s *companyAreaSrv) Create(ctx context.Context, createReq req.CompanyAreaCreateReq) error {
	if err := s.requireHeadquarter(ctx); err != nil {
		return err
	}
	db := ctx.GetDB()

	// 校验名称长度
	if !createReq.LocaleName.CheckLen(100) {
		return errors.NewWithCode(constant.CodeParamError, "区域名称超过100字符")
	}

	// 校验名称不重复
	if err := s.checkNameDuplicate(db, createReq.LocaleName, 0); err != nil {
		return err
	}

	// 校验门店唯一归属
	if len(createReq.CompanyUuids) > 0 {
		if err := s.checkShopsUnassigned(createReq.CompanyUuids, 0); err != nil {
			return err
		}
	}

	// 创建多语言名称
	multiLangRepo := repository.NewMultiLanguageNameRepo(db)
	multiLangName := model.MultiLanguageName{}
	multiLangName.InitByLocaleResponse(createReq.LocaleName)
	multiLangUuid, err := multiLangRepo.CreateMultiLanguageName(multiLangName)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 创建区域
	areaRepo := repository.NewCompanyAreaRepo(db)
	area := &model.CompanyArea{
		Name:                  multiLangName.ToJson(),
		MultiLanguageNameUuid: multiLangUuid,
		Sort:                  createReq.Sort,
	}
	if err := areaRepo.Create(area); err != nil {
		return errors.WithMessage(err)
	}

	// 分配门店到区域（双写saas+商户库）
	if len(createReq.CompanyUuids) > 0 {
		if err := s.assignShops(area.Uuid, createReq.CompanyUuids); err != nil {
			return err
		}
	}

	return nil
}

// Update 编辑区域
func (s *companyAreaSrv) Update(ctx context.Context, updateReq req.CompanyAreaUpdateReq) error {
	if err := s.requireHeadquarter(ctx); err != nil {
		return err
	}
	db := ctx.GetDB()
	areaRepo := repository.NewCompanyAreaRepo(db)

	// 获取现有区域
	area, err := areaRepo.GetByUuid(updateReq.Uuid)
	if err != nil {
		return errors.WithMessage(err)
	}

	// 校验名称长度
	if !updateReq.LocaleName.CheckLen(100) {
		return errors.NewWithCode(constant.CodeParamError, "区域名称超过100字符")
	}

	// 校验名称不重复（排除自己）
	if err := s.checkNameDuplicate(db, updateReq.LocaleName, updateReq.Uuid); err != nil {
		return err
	}

	// 更新多语言名称
	multiLangRepo := repository.NewMultiLanguageNameRepo(db)
	multiLangName := model.MultiLanguageName{}
	multiLangName.InitByLocaleResponse(updateReq.LocaleName)
	if err := multiLangRepo.UpdateMultiLanguageName(area.MultiLanguageNameUuid, multiLangName); err != nil {
		return errors.WithMessage(err)
	}

	// 更新区域信息
	if err := areaRepo.Update(updateReq.Uuid, map[string]any{
		"name": multiLangName.ToJson(),
		"sort": updateReq.Sort,
	}); err != nil {
		return errors.WithMessage(err)
	}

	// 差量更新门店归属
	if err := s.updateShopAssignment(ctx, updateReq.Uuid, updateReq.CompanyUuids); err != nil {
		return err
	}

	return nil
}

// Delete 删除区域
func (s *companyAreaSrv) Delete(ctx context.Context, deleteReq req.CompanyAreaDeleteReq) error {
	if err := s.requireHeadquarter(ctx); err != nil {
		return err
	}
	db := ctx.GetDB()
	areaRepo := repository.NewCompanyAreaRepo(db)

	// 确认区域存在
	if _, err := areaRepo.GetByUuid(deleteReq.Uuid); err != nil {
		return errors.WithMessage(err)
	}

	// 将该区域下所有门店的company_area_uuid置为0
	if err := s.unassignAllShops(ctx, deleteReq.Uuid); err != nil {
		return err
	}

	// 软删除区域
	if err := areaRepo.Delete(deleteReq.Uuid); err != nil {
		return errors.WithMessage(err)
	}

	return nil
}

// GetUnassigned 获取未分配区域的门店列表
func (s *companyAreaSrv) GetUnassigned(ctx context.Context) (resp.CompanyAreaUnassignedResp, error) {
	if err := s.requireHeadquarter(ctx); err != nil {
		return resp.CompanyAreaUnassignedResp{}, err
	}
	allSettings, err := s.getAllSettingsByHeadquarter(ctx)
	if err != nil {
		return resp.CompanyAreaUnassignedResp{}, errors.WithMessage(err)
	}

	var unassignedUuids []uint64
	for _, setting := range allSettings {
		if setting.CompanyAreaUuid == 0 {
			unassignedUuids = append(unassignedUuids, setting.CompanyUuid)
		}
	}

	list, err := s.fetchShopsByUuids(unassignedUuids)
	if err != nil {
		return resp.CompanyAreaUnassignedResp{}, err
	}

	return resp.CompanyAreaUnassignedResp{List: list}, nil
}

// GetShopList 获取区域管理用门店列表（轻量版，含区域归属）
func (s *companyAreaSrv) GetShopList(ctx context.Context) (resp.CompanyAreaShopListResp, error) {
	if err := s.requireHeadquarter(ctx); err != nil {
		return resp.CompanyAreaShopListResp{}, err
	}

	// 查询总部下所有门店设置（含 area_uuid）
	allSettings, err := s.getAllSettingsByHeadquarter(ctx)
	if err != nil {
		return resp.CompanyAreaShopListResp{}, errors.WithMessage(err)
	}

	// 构建 companyUuid → areaUuid 映射
	companyUuids := make([]uint64, 0, len(allSettings))
	areaUuidMap := make(map[uint64]uint64, len(allSettings))
	for _, setting := range allSettings {
		companyUuids = append(companyUuids, setting.CompanyUuid)
		areaUuidMap[setting.CompanyUuid] = setting.CompanyAreaUuid
	}

	// 批量查门店名称
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	companies, err := repository.NewCompanyRepo(saasDB).GetByUuids(companyUuids)
	if err != nil {
		return resp.CompanyAreaShopListResp{}, errors.WithMessage(err)
	}

	// 查区域名称（从总部商户库）
	areaNameMap := make(map[uint64]string)
	headquarterUuid := s.getHeadquarterUuid(ctx)
	if headquarterUuid > 0 {
		hqDB := s.dbm.GetDB(headquarterUuid)
		areas, err := repository.NewCompanyAreaRepo(hqDB).GetList(
			repository.CommonRepo.Preload(repository.WithPreload{Query: "MultiLanguageName"}),
		)
		if err != nil {
			return resp.CompanyAreaShopListResp{}, errors.WithMessage(err)
		}
		companySetting := ctx.GetCompanySetting()
		lang := companySetting.GetDefaultLanguage()
		for _, area := range areas {
			if area.MultiLanguageName != nil {
				areaNameMap[area.Uuid] = area.MultiLanguageName.GetNameByLangWithFallback(lang)
			}
		}
	}

	// 组装响应
	list := make([]resp.CompanyAreaShopItem, 0, len(companies))
	for _, company := range companies {
		areaUuid := areaUuidMap[company.Uuid]
		list = append(list, resp.CompanyAreaShopItem{
			Uuid:     company.Uuid,
			Name:     company.Name,
			AreaUuid: areaUuid,
			AreaName: areaNameMap[areaUuid],
		})
	}

	return resp.CompanyAreaShopListResp{List: list}, nil
}

// checkNameDuplicate 检查区域名称是否重复
func (s *companyAreaSrv) checkNameDuplicate(db *gorm.DB, localeName dto.LocaleResponse, excludeUuid uint64) error {
	multiLangRepo := repository.NewMultiLanguageNameRepo(db)
	// 检查每种有值的语言是否重复
	langs := map[string]string{
		"en": localeName.EN, "zh": localeName.ZH, "th": localeName.TH,
		"zhtw": localeName.ZHTW, "ja": localeName.JA, "ko": localeName.KO,
		"my": localeName.MY, "tr": localeName.TR, "sv": localeName.SV,
	}
	for lang, name := range langs {
		if name == "" {
			continue
		}
		count, err := multiLangRepo.CountNameExistsByTable("company_area", lang, name, excludeUuid)
		if err != nil {
			return errors.WithMessage(err)
		}
		if count > 0 {
			return errors.NewWithCode(constant.CodeParamError, "区域名称已存在")
		}
	}
	return nil
}

// checkShopsUnassigned 检查门店是否已被分配到其他区域
func (s *companyAreaSrv) checkShopsUnassigned(companyUuids []uint64, excludeAreaUuid uint64) error {
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	settings, err := repository.NewCompanySettingRepo(saasDB).GetCompanySettingsByCompanyUuids(companyUuids)
	if err != nil {
		return errors.WithMessage(err)
	}

	for _, setting := range settings {
		if setting.CompanyAreaUuid > 0 && setting.CompanyAreaUuid != excludeAreaUuid {
			return errors.NewWithCode(constant.CodeParamError, "部分门店已属于其他区域")
		}
	}
	return nil
}

// assignShops 分配门店到区域（双写saas+商户库）
func (s *companyAreaSrv) assignShops(areaUuid uint64, companyUuids []uint64) error {
	return s.updateShopAreaUuid(companyUuids, map[string]any{"company_area_uuid": areaUuid})
}

// unassignAllShops 将区域下所有门店置为未分配
func (s *companyAreaSrv) unassignAllShops(ctx context.Context, areaUuid uint64) error {
	allSettings, err := s.getAllSettingsByHeadquarter(ctx)
	if err != nil {
		return err
	}

	var companyUuids []uint64
	for _, setting := range allSettings {
		if setting.CompanyAreaUuid == areaUuid {
			companyUuids = append(companyUuids, setting.CompanyUuid)
		}
	}

	if len(companyUuids) > 0 {
		return s.updateShopAreaUuid(companyUuids, map[string]any{"company_area_uuid": 0})
	}
	return nil
}

// updateShopAssignment 差量更新门店归属
func (s *companyAreaSrv) updateShopAssignment(ctx context.Context, areaUuid uint64, newCompanyUuids []uint64) error {
	allSettings, err := s.getAllSettingsByHeadquarter(ctx)
	if err != nil {
		return err
	}

	// 找出当前区域的门店
	currentUuids := make(map[uint64]bool)
	for _, setting := range allSettings {
		if setting.CompanyAreaUuid == areaUuid {
			currentUuids[setting.CompanyUuid] = true
		}
	}

	newUuidSet := make(map[uint64]bool)
	for _, uuid := range newCompanyUuids {
		newUuidSet[uuid] = true
	}

	// 需要移除的门店（在当前区域但不在新列表中）
	var toRemove []uint64
	for uuid := range currentUuids {
		if !newUuidSet[uuid] {
			toRemove = append(toRemove, uuid)
		}
	}

	// 需要添加的门店（在新列表中但不在当前区域）
	var toAdd []uint64
	for uuid := range newUuidSet {
		if !currentUuids[uuid] {
			toAdd = append(toAdd, uuid)
		}
	}

	// 校验新增门店是否已被分配到其他区域
	if len(toAdd) > 0 {
		if err := s.checkShopsUnassigned(toAdd, areaUuid); err != nil {
			return err
		}
	}

	// 移除门店
	if len(toRemove) > 0 {
		if err := s.updateShopAreaUuid(toRemove, map[string]any{"company_area_uuid": 0}); err != nil {
			return err
		}
	}

	// 添加门店
	if len(toAdd) > 0 {
		if err := s.updateShopAreaUuid(toAdd, map[string]any{"company_area_uuid": areaUuid}); err != nil {
			return err
		}
	}

	return nil
}

// updateShopAreaUuid 双写saas+商户库的company_area_uuid
func (s *companyAreaSrv) updateShopAreaUuid(companyUuids []uint64, data map[string]any) error {
	// 批量写saas主库
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	if err := repository.NewCompanySettingRepo(saasDB).UpdateByCompanyUuids(companyUuids, data); err != nil {
		return errors.WithMessage(err)
	}

	// 逐个写商户库（每个门店独立数据库，无法批量）
	for _, companyUuid := range companyUuids {
		shopDB := s.dbm.GetDB(companyUuid)
		if err := repository.NewCompanySettingRepo(shopDB).UpdateByCompanyUuid(companyUuid, data); err != nil {
			return errors.WithMessage(err)
		}
	}
	return nil
}
