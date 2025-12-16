package service

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
)

// ISaasStaffSrv saas数据库统一账号服务接口（仅用于门店切换等功能）
type ISaasStaffSrv interface {
	// UpdateLastCompany 更新上次登录商家UUID（门店切换时调用）
	UpdateLastCompany(ctx context.Context, staffUuid, companyUuid uint64) error
	// GetDefaultCompanyUuid 获取默认门店UUID（登录时使用，优先使用 last_company_uuid）
	GetDefaultCompanyUuid(ctx context.Context, staffUuid uint64) (uint64, error)
	// SearchStaff 根据关键字（email或phone）搜索员工，返回在当前门店可见范围内的门店和角色信息
	SearchStaff(ctx context.Context, keyword string) *resp.SearchStaffResp
}

func NewSaasStaffSrv(dbm *database.DBManager) ISaasStaffSrv {
	return NewSaasStaffSrvImpl(dbm)
}

type saasStaffSrv struct {
	dbm *database.DBManager
}

func NewSaasStaffSrvImpl(dbm *database.DBManager) ISaasStaffSrv {
	return &saasStaffSrv{
		dbm: dbm,
	}
}

// UpdateLastCompany 更新上次登录商家UUID（门店切换时调用）
func (s *saasStaffSrv) UpdateLastCompany(ctx context.Context, staffUuid, companyUuid uint64) error {
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	staffRepo := repository.NewSaasStaffRepo(saasDB)
	companyStaffRepo := repository.NewCompanyStaffRepo(saasDB)

	// 验证员工是否有该门店权限
	companyStaff, err := companyStaffRepo.GetByStaffAndCompany(staffUuid, companyUuid)
	if err != nil || companyStaff == nil {
		return errors.New("无权限访问该门店")
	}

	// 更新 last_company_uuid
	return staffRepo.Update(staffUuid, map[string]any{
		"last_company_uuid": companyUuid,
		"update_time":       time.Now().Unix(),
	})
}

// GetDefaultCompanyUuid 获取默认门店UUID（登录时使用，优先使用 last_company_uuid）
func (s *saasStaffSrv) GetDefaultCompanyUuid(ctx context.Context, staffUuid uint64) (uint64, error) {
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	staffRepo := repository.NewSaasStaffRepo(saasDB)
	companyStaffRepo := repository.NewCompanyStaffRepo(saasDB)

	// 获取员工信息
	staff, err := staffRepo.GetByUuid(staffUuid)
	if err != nil {
		return 0, errors.WithMessage(err, "获取员工信息失败")
	}

	// 获取员工关联的门店列表
	companyList, err := companyStaffRepo.GetByStaffUuid(staffUuid)
	if err != nil {
		return 0, errors.WithMessage(err, "获取门店列表失败")
	}

	if len(companyList) == 0 {
		return 0, errors.New("员工未关联任何门店")
	}

	// 如果只有一个门店，直接返回
	if len(companyList) == 1 {
		return companyList[0].CompanyUuid, nil
	}

	// 多个门店时，优先使用 last_company_uuid
	if staff.LastCompanyUuid > 0 {
		// 验证员工是否有该门店权限
		for _, cs := range companyList {
			if cs.CompanyUuid == staff.LastCompanyUuid {
				return staff.LastCompanyUuid, nil
			}
		}
	}

	// 如果 last_company_uuid 无效，返回第一个门店
	return companyList[0].CompanyUuid, nil
}

// SearchStaff 根据关键字（email或phone）搜索员工，返回在当前门店可见范围内的门店和角色信息
func (s *saasStaffSrv) SearchStaff(ctx context.Context, keyword string) *resp.SearchStaffResp {
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	saasStaffRepo := repository.NewSaasStaffRepo(saasDB)
	companyStaffRepo := repository.NewCompanyStaffRepo(saasDB)

	// 1. 根据关键字从 saas.ttpos_staff 精确搜索（email或phone）
	var saasStaff *model.SaasStaff
	var err error

	// 通过邮箱或手机号查询
	saasStaff, err = saasStaffRepo.GetByEmailOrPhone(keyword)
	if err != nil || saasStaff == nil {
		return &resp.SearchStaffResp{
			CompanyList: make([]*resp.SearchStaffCompanyInfo, 0),
		}
	}

	// 2. 查询关联的 saas.ttpos_company_staff
	companyStaffList, _ := companyStaffRepo.GetByStaffUuid(saasStaff.Uuid, companyStaffRepo.WithCompany())

	// 3. 获取当前门店能看到的所有门店列表（包括自己和在 parent_company_uuids 包含自己uuid的记录）
	currentCompanyUuid := ctx.GetCompanyUuid()
	visibleCompanies, err := repository.NewCompanyRepo(s.dbm.GetDB(constant.DefaultDB)).GetVisibleCompanyList(currentCompanyUuid)
	if err != nil {
		return &resp.SearchStaffResp{
			CompanyList: make([]*resp.SearchStaffCompanyInfo, 0),
		}
	}

	// 构建可见门店UUID集合
	visibleCompanyUuids := make(map[uint64]bool)
	companyUuidToName := make(map[uint64]string)
	for _, company := range visibleCompanies {
		visibleCompanyUuids[company.Uuid] = true
		companyUuidToName[company.Uuid] = company.Name
	}

	// 4. 匹配 company_uuid，过滤出在当前门店可见范围内的门店
	matchedCompanyList := make([]*resp.SearchStaffCompanyInfo, 0)
	for _, cs := range companyStaffList {

		companyUuid := cs.CompanyUuid
		if !visibleCompanyUuids[companyUuid] {
			continue
		}

		// 5. 根据员工uuid从匹配的company_uuid的商家查找员工在该门店下的角色信息
		shopDB := s.dbm.GetDB(companyUuid)
		if shopDB == nil {
			continue
		}

		staffRoleRepo := repository.NewStaffRoleRepo(shopDB)
		roleUuids, err := staffRoleRepo.GetRoleUuidsByStaffUuid(saasStaff.Uuid)
		if err != nil {
			continue
		}

		roleRepo := repository.NewRoleRepo(shopDB)
		roles, err := roleRepo.GetRoleList(roleRepo.WhereUuids(roleUuids))
		if err != nil {
			continue
		}

		roleItems := make([]resp.RoleItem, 0, len(roles))
		for _, role := range roles {
			roleItems = append(roleItems, resp.RoleItem{
				Uuid: role.Uuid,
				Name: role.Name,
			})
		}

		matchedCompanyList = append(matchedCompanyList, &resp.SearchStaffCompanyInfo{
			CompanyUuid: companyUuid,
			CompanyName: companyUuidToName[companyUuid],
			Roles:       roleItems,
			IsSuper:     cs.IsSuper,
		})
	}

	return &resp.SearchStaffResp{
		Uuid:        saasStaff.Uuid,
		Email:       saasStaff.Email,
		Phone:       saasStaff.Phone,
		RealName:    saasStaff.RealName,
		CompanyList: matchedCompanyList,
	}
}
