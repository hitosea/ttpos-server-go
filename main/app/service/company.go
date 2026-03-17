package service

import (
	"sort"
	"strings"
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	settingSrv "ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"

	"gorm.io/gorm"
)

type ICompanySrv interface {
	GetCompanyList(ctx context.Context) (resp.SaasCompanyListResp, error)                           // 获取本店可看到的所有店列表（包含店铺名称、超管real_name以及超管手机号）
	GetCompanyListWithStoreCode(ctx context.Context) (resp.SaasCompanyListWithStoreCodeResp, error) // 获取本店可看到的所有店列表（包含店铺编码）
	GetCompanyInfo(ctx context.Context, companyUuid uint64) (*resp.CompanyStoreResp, error)         // 获取门店信息
	UpdateCompanyInfo(ctx context.Context, updateReq req.UpdateCompanySettingReq) error             // 修改门店信息
	UpdateBusinessStatus(ctx context.Context, updateReq req.UpdateBusinessStatusReq) error          // 修改营业状态
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
			Uuid:   company.Uuid,
			Name:   company.Name,
			Roles:  roleItems,
			Status: company.Status,
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

// GetCompanyListWithStoreCode 获取本店可看到的所有店列表（包含店铺编码）
func (s *companySrv) GetCompanyListWithStoreCode(ctx context.Context) (resp.SaasCompanyListWithStoreCodeResp, error) {
	db := s.dbm.GetDB(constant.DefaultDB)
	companyRepo := repository.NewCompanyRepo(db)
	currentCompanyUuid := ctx.GetCompanyUuid()
	companySetting := ctx.GetCompanySetting()

	// 获取总部下的所有门店
	headquarterUuid := companySetting.HeadquarterUuid
	if companySetting.IsHeadquarter() {
		headquarterUuid = currentCompanyUuid
	}

	companies, err := companyRepo.GetNoDeleteListByHeadquarterUuid(headquarterUuid)
	if err != nil {
		return resp.SaasCompanyListWithStoreCodeResp{}, errors.WithMessage(errors.New("获取总部下的所有门店失败"), err.Error())
	}

	// 转换响应数据
	list := make([]resp.CompanyInfoRespWithStoreCode, 0)
	for _, company := range companies {
		if company.Uuid == currentCompanyUuid {
			continue
		}
		ctxCopy := ctx.Copy()
		ctxCopy.SetCompanyUuid(company.Uuid)
		storeSetting, err := s.settingSrv.GetStoreSetting(ctxCopy)
		if err != nil {
			return resp.SaasCompanyListWithStoreCodeResp{}, errors.WithMessage(errors.New("获取门店设置失败"), err.Error())
		}
		list = append(list, resp.CompanyInfoRespWithStoreCode{
			Uuid:      company.Uuid,
			Name:      company.Name,
			StoreCode: storeSetting.StoreCode,
			Status:    company.Status,
		})
	}

	// 对 list 按 StoreCode 排序
	// 排序规则：
	// 1. StoreCode 为空的排在最前面
	// 2. 空 StoreCode 之间：按 CompanyUuid 升序排序（保证稳定性）
	// 3. 非空 StoreCode：去掉 "No." 前缀后，纯数字按数字大小排序（去掉前缀0），否则按字符串排序（不区分大小写）
	sort.Slice(list, func(i, j int) bool {
		item1 := list[i]
		item2 := list[j]

		isEmpty1 := item1.StoreCode == ""
		isEmpty2 := item2.StoreCode == ""

		// 1. StoreCode 为空的排在最前面
		if isEmpty1 != isEmpty2 {
			return isEmpty1 // true 排在前面
		}

		// 2. 如果都是空字符串，按 CompanyUuid 排序
		if isEmpty1 && isEmpty2 {
			return item1.Uuid < item2.Uuid
		}

		// 3. 如果都有 StoreCode，按新规则排序
		// 去掉 "No." 前缀（如果有）
		processed1 := utils.ProcessStoreCodeForSort(item1.StoreCode)
		processed2 := utils.ProcessStoreCodeForSort(item2.StoreCode)

		// 判断处理后的内容是否只有数字
		isDigits1 := utils.IsOnlyDigits(processed1)
		isDigits2 := utils.IsOnlyDigits(processed2)

		// 如果都是纯数字，按数字大小排序
		if isDigits1 && isDigits2 {
			num1, _ := utils.ParseNumberWithoutLeadingZeros(processed1)
			num2, _ := utils.ParseNumberWithoutLeadingZeros(processed2)
			return num1 < num2
		}

		// 如果一个是数字一个不是，数字排在前面
		if isDigits1 != isDigits2 {
			return isDigits1
		}

		// 否则按不区分大小写的字符串排序
		return strings.ToLower(processed1) < strings.ToLower(processed2)
	})

	return resp.SaasCompanyListWithStoreCodeResp{
		List: list,
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

	// 从时段表推导营业状态
	periodRepo := repository.NewBusinessStatusPeriodRepo(shopDB)
	businessStatus := periodRepo.GetBusinessStatus()

	// 构建响应
	return &resp.CompanyStoreResp{
		Uuid:           companyUuid,
		IPWhiteList:    storeSetting.IPWhiteList,
		Name:           storeSetting.Name,
		LogoUrl:        storeSetting.LogoURL,
		Address:        storeSetting.Address,
		Coordinates:    storeSetting.Coordinates,
		CompanyName:    storeSetting.Company,
		Phone:          storeSetting.Phone,
		TaxNumber:      storeSetting.TaxNumber,
		LanguageList:   storeSetting.Language,
		TimeZone:       storeSetting.TimeZone,
		Language:       companySetting.GetLanguages(),
		StoreCode:      storeSetting.StoreCode,
		Status:         targetCompany.Status,
		BusinessStatus: businessStatus,
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
		StoreCode:   updateReq.StoreCode,
	})
	if err != nil {
		return errors.WithMessage(errors.New("修改门店信息失败"), err.Error())
	}

	return nil
}

// UpdateBusinessStatus 修改营业状态
// 通过 ttpos_business_status_period 表管理状态，不修改 company_setting
func (s *companySrv) UpdateBusinessStatus(ctx context.Context, updateReq req.UpdateBusinessStatusReq) error {
	if updateReq.BusinessStatus != constant.BusinessStatusTest && updateReq.BusinessStatus != constant.BusinessStatusNormal {
		return errors.New("无效的营业状态值")
	}

	companyUuid := updateReq.Uuid

	shopDB := s.dbm.GetDB(companyUuid)
	if shopDB == nil {
		return errors.New("获取门店数据库连接失败")
	}

	periodRepo := repository.NewBusinessStatusPeriodRepo(shopDB)
	openPeriod, err := periodRepo.GetOpenPeriod()
	if err != nil {
		return errors.WithMessage(errors.New("查询营业状态失败"), err.Error())
	}

	// 推导当前状态
	currentStatus := constant.BusinessStatusNormal
	if openPeriod != nil {
		currentStatus = constant.BusinessStatusTest
	}

	// 状态未变，无需操作
	if currentStatus == updateReq.BusinessStatus {
		return nil
	}

	now := time.Now().Unix()

	if updateReq.BusinessStatus == constant.BusinessStatusTest {
		// 切换到测试营业：创建新的测试时段
		period := &model.BusinessStatusPeriod{
			StartTime: now,
		}
		if err := periodRepo.Create(period); err != nil {
			return errors.WithMessage(errors.New("创建测试营业时段失败"), err.Error())
		}
	} else {
		// 切换到正常营业：关闭当前测试时段
		if err := periodRepo.CloseOpenPeriod(now); err != nil {
			return errors.WithMessage(errors.New("关闭测试营业时段失败"), err.Error())
		}
	}

	return nil
}
