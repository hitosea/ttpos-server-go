package service

import (
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	settingSrv "ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"

	"gorm.io/gorm"
)

type ICompanySrv interface {
	GetCompanyList(ctx context.Context) (resp.SaasCompanyListResp, error)                   // 获取本店可看到的所有店列表（包含店铺名称、超管real_name以及超管手机号）
	GetCompanyInfo(ctx context.Context, companyUuid uint64) (*resp.CompanyStoreResp, error) // 获取门店信息
	UpdateCompanyInfo(ctx context.Context, updateReq req.UpdateCompanySettingReq) error     // 修改门店信息
}

type companySrv struct {
	dbm        *database.DBManager
	settingSrv settingSrv.ISrv
}

func NewCompanySrv(dbm *database.DBManager, settingSrv settingSrv.ISrv) ICompanySrv {
	return &companySrv{
		dbm:        dbm,
		settingSrv: settingSrv,
	}
}

// GetCompanyList 获取本店可看到的所有店列表
func (s *companySrv) GetCompanyList(ctx context.Context) (resp.SaasCompanyListResp, error) {
	// 获取当前门店UUID
	currentCompanyUuid := ctx.GetCompanyUuid()
	var saasCompanyListResp resp.SaasCompanyListResp

	// 获取本店可看到的所有店列表（包括自己和parent_company_uuids包含自己的店）
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	companyRepo := repository.NewCompanyRepo(saasDB)
	companies, err := companyRepo.GetVisibleCompanyList(currentCompanyUuid)
	if err != nil {
		return saasCompanyListResp, errors.WithMessage(errors.New("获取门店列表失败"), err.Error())
	}

	// 获取所有门店的UUID列表
	companyUuids := make([]uint64, 0, len(companies))
	for _, company := range companies {
		companyUuids = append(companyUuids, company.Uuid)
	}

	// 查询所有门店的超管信息（is_super=1）
	companyStaffRepo := repository.NewCompanyStaffRepo(saasDB)
	companyStaffList, err := companyStaffRepo.GetByCompanyUuids(companyUuids)
	if err != nil {
		return saasCompanyListResp, errors.WithMessage(errors.New("获取门店员工列表失败"), err.Error())
	}

	// 构建门店UUID到超管信息的映射
	superAdminMap := make(map[uint64]*model.CompanyStaff)
	for _, cs := range companyStaffList {
		if cs.IsSuper == 1 {
			// 如果该门店已有超管，保留第一个（或可以根据需要选择）
			if _, exists := superAdminMap[cs.CompanyUuid]; !exists {
				superAdminMap[cs.CompanyUuid] = cs
			}
		}
	}

	// 获取超管的详细信息（real_name和phone）
	superAdminUuids := make([]uint64, 0)
	for _, cs := range superAdminMap {
		superAdminUuids = append(superAdminUuids, cs.Uuid)
	}

	saasStaffRepo := repository.NewSaasStaffRepo(saasDB)
	superAdminInfoMap := make(map[uint64]*model.SaasStaff)
	if len(superAdminUuids) > 0 {
		saasStaffList, _, err := saasStaffRepo.PaginateGetStaffs(1, 1000, saasStaffRepo.WhereUuids(superAdminUuids))
		if err == nil {
			for i := range saasStaffList {
				superAdminInfoMap[saasStaffList[i].Uuid] = &saasStaffList[i]
			}
		}
	}

	// 构建响应
	companyList := make([]resp.CompanyInfoResp, 0, len(companies))
	for _, company := range companies {

		// 获取该门店的角色列表
		roleRepo := repository.NewRoleRepo(s.dbm.GetDB(company.Uuid))
		roles, _ := roleRepo.GetRoleList()

		roleItems := make([]resp.RoleItem, 0, len(roles))
		for _, role := range roles {
			roleItems = append(roleItems, resp.RoleItem{
				Uuid: role.Uuid,
				Name: role.Name,
			})
		}

		companyInfo := resp.CompanyInfoResp{
			Uuid:  company.Uuid,
			Name:  company.Name,
			Roles: roleItems,
		}

		// 获取该门店的超管信息
		if cs, exists := superAdminMap[company.Uuid]; exists {
			if saasStaff, ok := superAdminInfoMap[cs.Uuid]; ok {
				companyInfo.SuperRealName = saasStaff.RealName
				companyInfo.SuperPhone = saasStaff.Phone
			}
		}

		companyList = append(companyList, companyInfo)
	}

	return resp.SaasCompanyListResp{
		List: companyList,
	}, nil
}

// GetCompanyInfo 获取门店信息
func (s *companySrv) GetCompanyInfo(ctx context.Context, companyUuid uint64) (*resp.CompanyStoreResp, error) {
	// 获取当前门店UUID
	currentCompanyUuid := ctx.GetCompanyUuid()

	// 获取当前门店可见的所有门店列表
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	companyRepo := repository.NewCompanyRepo(saasDB)
	visibleCompanies, err := companyRepo.GetVisibleCompanyList(currentCompanyUuid)
	if err != nil {
		return nil, errors.WithMessage(errors.New("获取可见门店列表失败"), err.Error())
	}

	// 验证目标门店是否在可见列表中
	validCompanyUuid := false
	for _, company := range visibleCompanies {
		if company.Uuid == companyUuid {
			validCompanyUuid = true
			break
		}
	}

	if !validCompanyUuid {
		return nil, errors.New("门店不存在或无权限查看")
	}

	// 从 saas 数据库查询 company
	targetCompany, err := companyRepo.GetCompanyInfoByUuid(companyUuid)
	if err != nil {
		return nil, errors.WithMessage(errors.New("获取门店信息失败"), err.Error())
	}

	// 从商家数据库查询 companySetting
	shopDB := s.dbm.GetDB(companyUuid)
	if shopDB == nil {
		return nil, errors.New("获取门店数据库连接失败")
	}
	companySettingRepo := repository.NewCompanySettingRepo(shopDB)
	companySetting, err := companySettingRepo.GetOne(func(db *gorm.DB) *gorm.DB {
		return db.Where("company_uuid = ?", companyUuid)
	})
	if err != nil {
		return nil, errors.WithMessage(errors.New("获取门店设置失败"), err.Error())
	}

	// 设置到 ctx 中
	ctx.SetCompanyUuid(companyUuid)
	ctx.SetCompany(*targetCompany)
	ctx.SetDB(shopDB)
	ctx.SetCompanySetting(companySetting)

	// 从 settingSrv 获取 storeSetting
	storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(errors.New("获取门店设置失败"), err.Error())
	}

	// 构建响应
	return &resp.CompanyStoreResp{
		Uuid:         companyUuid,
		Name:         storeSetting.Name,
		LogoUrl:      storeSetting.LogoURL,
		Address:      storeSetting.Address,
		Coordinates:  storeSetting.Coordinates,
		CompanyName:  storeSetting.Company,
		Phone:        storeSetting.Phone,
		TaxNumber:    storeSetting.TaxNumber,
		LanguageList: storeSetting.Language,
		TimeZone:     storeSetting.TimeZone,
		Language:     companySetting.GetLanguages(),
	}, nil
}

// UpdateCompanyInfo 修改门店信息
func (s *companySrv) UpdateCompanyInfo(ctx context.Context, updateReq req.UpdateCompanySettingReq) error {
	// 查询门店可以看到的所有门店列表
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	companyRepo := repository.NewCompanyRepo(saasDB)
	companies, err := companyRepo.GetVisibleCompanyList(updateReq.Uuid)
	if err != nil {
		return errors.WithMessage(errors.New("获取门店列表失败"), err.Error())
	}

	var editCompany model.Company
	for _, company := range companies {
		if company.Uuid == updateReq.Uuid {
			editCompany = company
			break
		}
	}

	if editCompany.Uuid == 0 {
		return errors.New("门店不存在")
	}

	// 查询门店设置
	companySettingRepo := repository.NewCompanySettingRepo(s.dbm.GetDB(editCompany.Uuid))
	companySetting, err := companySettingRepo.GetOne(func(db *gorm.DB) *gorm.DB {
		return db.Where("company_uuid = ?", editCompany.Uuid)
	})
	if err != nil {
		return errors.WithMessage(errors.New("获取门店设置失败"), err.Error())
	}

	ctx.SetCompanyUuid(updateReq.Uuid)
	ctx.SetCompany(editCompany)
	ctx.SetDB(s.dbm.GetDB(updateReq.Uuid))
	ctx.SetCompanySetting(companySetting)

	storeSetting, err := s.settingSrv.GetStoreSetting(ctx)
	if err != nil {
		return errors.WithMessage(errors.New("获取门店设置失败"), err.Error())
	}

	err = s.settingSrv.EditStoreSetting(ctx, req.UpdateStoreSetting{
		Name:        updateReq.Name,
		LogoUrl:     updateReq.LogoUrl,
		TimeZone:    updateReq.TimeZone,
		CompanyName: updateReq.CompanyName,
		Address:     updateReq.Address,
		Phone:       updateReq.Phone,
		TaxNumber:   updateReq.TaxNumber,
		Language:    updateReq.Language,
		Coordinates: updateReq.Coordinates,
		StoreCode:   storeSetting.StoreCode,
	})
	if err != nil {
		return errors.WithMessage(errors.New("修改门店信息失败"), err.Error())
	}

	return nil
}
