package service

import (
	"time"
	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/app/service/setting"
	"ttpos-server-go/pkg/cache"
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
	SearchStaff(ctx context.Context, searchReq req.SearchStaffByKeywordReq) *resp.SearchStaffResp
	// QueryStaffByContact 根据邮箱或手机号查询员工（支持模糊搜索），仅返回当前门店可见范围内的授权员工
	QueryStaffByContact(ctx context.Context, queryReq req.QueryStaffByContactReq) (*resp.QueryStaffByContactResp, error)
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
func (s *saasStaffSrv) SearchStaff(ctx context.Context, searchReq req.SearchStaffByKeywordReq) *resp.SearchStaffResp {
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	saasStaffRepo := repository.NewSaasStaffRepo(saasDB)
	companyStaffRepo := repository.NewCompanyStaffRepo(saasDB)

	// 1. 根据关键字从 saas.ttpos_staff 精确搜索（email或phone）
	var saasStaff *model.SaasStaff
	var err error

	// 通过邮箱或手机号查询
	saasStaff, err = saasStaffRepo.GetByEmailOrPhone(searchReq.Field, searchReq.Keyword)
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

// QueryStaffByContact 根据邮箱或手机号查询员工（支持模糊搜索），仅返回当前门店可见范围内的授权员工
func (s *saasStaffSrv) QueryStaffByContact(ctx context.Context, queryReq req.QueryStaffByContactReq) (*resp.QueryStaffByContactResp, error) {
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	saasStaffRepo := repository.NewSaasStaffRepo(saasDB)
	companyStaffRepo := repository.NewCompanyStaffRepo(saasDB)

	// 1. 固定限制返回 20 条结果
	limit := 20

	// 2. 获取当前门店下的授权员工列表
	// 2.1 获取业务设置（授权员工配置）
	settingSrv := setting.NewSrv(s.dbm, cache.Global)
	businessSetting, err := settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return &resp.QueryStaffByContactResp{List: []*resp.QueryStaffByContactItem{}}, nil
	}

	// 2.2 根据操作类型选择对应的授权员工列表
	var authorizedStaffIds []uint64
	if queryReq.OperationType == string(SensitiveOperationTypeRefund) {
		// 退款操作：使用退款授权列表
		authorizedStaffIds = businessSetting.RefundAuthorizedStaffIds
	} else {
		// 折扣操作：使用折扣授权列表（默认）
		authorizedStaffIds = businessSetting.DiscountAuthorizedStaffIds
	}

	// 2.3 获取当前门店下的授权员工（统一账号UUID）
	// 业务设置中的授权员工ID就是统一账号UUID，需要验证这些员工是否在当前门店下
	currentCompanyUuid := ctx.GetCompanyUuid()
	authorizedStaffUuids := make([]uint64, 0)
	authorizedStaffUuidsSet := make(map[uint64]bool) // 用于去重

	// 构建授权员工ID映射，便于快速查找
	authorizedStaffIdsMap := make(map[uint64]bool)
	for _, id := range authorizedStaffIds {
		authorizedStaffIdsMap[id] = true
	}

	// 只查询当前门店下的员工
	companyStaffList, _ := companyStaffRepo.GetByCompanyUuid(currentCompanyUuid)
	for _, cs := range companyStaffList {
		// CompanyStaff.Uuid 就是统一账号UUID（SaasStaff.Uuid）
		// 检查该员工是否在授权名单中
		if authorizedStaffIdsMap[cs.Uuid] {
			if !authorizedStaffUuidsSet[cs.Uuid] {
				authorizedStaffUuids = append(authorizedStaffUuids, cs.Uuid)
				authorizedStaffUuidsSet[cs.Uuid] = true
			}
		}
	}

	// 3. 如果没有授权员工，返回空列表
	if len(authorizedStaffUuids) == 0 {
		return &resp.QueryStaffByContactResp{List: []*resp.QueryStaffByContactItem{}}, nil
	}

	// 4. 根据关键词查询员工（模糊搜索）
	staffList, err := saasStaffRepo.QueryByKeyword(queryReq.Keyword, authorizedStaffUuids, limit)
	if err != nil {
		return &resp.QueryStaffByContactResp{List: []*resp.QueryStaffByContactItem{}}, nil
	}

	// 5. 转换为响应格式
	result := make([]*resp.QueryStaffByContactItem, 0, len(staffList))
	for _, staff := range staffList {
		result = append(result, &resp.QueryStaffByContactItem{
			Uuid:     staff.Uuid,
			RealName: staff.RealName,
			Email:    staff.Email,
			Phone:    staff.Phone,
		})
	}

	return &resp.QueryStaffByContactResp{List: result}, nil
}
