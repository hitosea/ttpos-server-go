package service

import (
	"fmt"
	"slices"
	"sort"
	"time"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/pkg/database"

	"github.com/jinzhu/copier"
	"gorm.io/gorm"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
)

type IRoleAccessSrv interface {
	GetPermission(routerName constant.RouteName, staffUuid, companyUuid uint64) ([]*resp.Permission, error)
	GetPermissionGroup(staffUuid, companyUuid uint64) (resp.PermissionGroup, error)
	GetApiPermission(staffUuid, companyUuid uint64) ([]string, error)
	GetCompanyPermissionGroup(companyUuid uint64, includeRouteNames []constant.RouteName) (resp.PermissionGroup, error)
}

func NewRoleAccessSrv(dbm *database.DBManager) IRoleAccessSrv {
	return NewRoleAccessSrvImpl(dbm)
}

type roleAccessSrv struct {
	dbm *database.DBManager
}

func NewRoleAccessSrvImpl(dbm *database.DBManager) IRoleAccessSrv {
	return &roleAccessSrv{
		dbm: dbm,
	}
}

// 从数据库获取权限
func (s *roleAccessSrv) getDbPermissions(staffUuid, companyUuid uint64) ([]model.Access, model.CompanySetting, model.Company, error) {
	db := s.dbm.GetDB(companyUuid)
	accessRepo := repository.NewAccessRepo(db)
	var companySetting model.CompanySetting
	var company model.Company
	staffRepo := repository.NewStaffRepo(db)
	staff, _ := staffRepo.GetStaff(staffRepo.WhereUuid(staffUuid), staffRepo.WithCompany(), staffRepo.WithCompanySetting())

	if staff.Company == nil || staff.Company.CompanySetting == nil {
		return nil, companySetting, company, errors.New("获取商家信息错误")
	}

	companySetting = *staff.Company.CompanySetting
	company = *staff.Company
	var options []repository.DBOption

	if staff.IsSuper == 1 { // 超级管理员
		if staff.UserType == 1 {
			options = append(options, accessRepo.WhereIsSupplier())
		}
	} else {
		roleUuids, err := repository.NewStaffRoleRepo(s.dbm.GetDB(staff.CompanyUuid)).GetRoleUuidsByStaffUuid(staff.Uuid)
		if err != nil {
			return nil, companySetting, company, errors.WithMessage(err, "获取用户角色失败")
		}
		accessUuids, err := accessRepo.GetAccessUuids(roleUuids)
		if err != nil {
			return nil, companySetting, company, errors.WithMessage(err, "获取角色权限失败")
		}
		options = append(options, accessRepo.WhereUuids(accessUuids))
	}

	dbPermissions, err := accessRepo.GetPermissions(options...)

	if err != nil {
		return nil, companySetting, company, errors.WithMessage(err, "获取权限失败")
	}

	return dbPermissions, companySetting, company, nil
}

// GetPermission 获取权限
func (s *roleAccessSrv) GetPermission(routerName constant.RouteName, staffUuid, companyUuid uint64) ([]*resp.Permission, error) {

	var permissions []resp.Permission
	dbPermissions, companySetting, company, err := s.getDbPermissions(staffUuid, companyUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	for _, dbPermission := range dbPermissions {
		var permission resp.Permission
		copier.Copy(&permission, dbPermission)
		permission.CreateTime = time.Unix(dbPermission.CreateTime, 0).Format(time.DateTime)
		permission.UpdateTime = time.Unix(dbPermission.UpdateTime, 0).Format(time.DateTime)
		permission.Children = []*resp.Permission{}
		permissions = append(permissions, permission)
	}

	permissions = s.filterPermission(permissions, companySetting, company)

	return s.buildPermissionTree(permissions, routerName), nil
}

// GetPermissionGroup 获取权限
func (s *roleAccessSrv) GetPermissionGroup(staffUuid, companyUuid uint64) (resp.PermissionGroup, error) {

	var permissions []resp.Permission
	groupPermission := resp.PermissionGroup{
		List: []*resp.Permission{},
	}
	dbPermissions, companySetting, company, err := s.getDbPermissions(staffUuid, companyUuid)
	if err != nil {
		return groupPermission, errors.WithMessage(err)
	}

	for _, dbPermission := range dbPermissions {
		var permission resp.Permission
		copier.Copy(&permission, dbPermission)
		// 手动设置ParentUuid字段，因为json标签不匹配导致copier无法正确复制
		permission.ParentUuid = dbPermission.ParentUuid
		permission.CreateTime = time.Unix(dbPermission.CreateTime, 0).Format(time.DateTime)
		permission.UpdateTime = time.Unix(dbPermission.UpdateTime, 0).Format(time.DateTime)
		permission.Children = []*resp.Permission{}
		permissions = append(permissions, permission)
	}

	// 筛选权限
	permissions = s.filterPermission(permissions, companySetting, company)

	// 构建权限树形结构
	roots := s.buildPermissionTreeWithoutFilter(permissions)
	for _, root := range roots {
		groupPermission.List = append(groupPermission.List, root)
	}

	// 返回权限组
	return groupPermission, nil
}

// 筛选权限
func (s *roleAccessSrv) filterPermission(permissions []resp.Permission, companySetting model.CompanySetting, company model.Company) []resp.Permission {
	var filteredPermissions []resp.Permission
	for _, permission := range permissions {
		// 删除无效数据
		if slices.Contains([]uint64{58, 124, 125, 128, 129, 160, 162, 1724320603, 1724320604, 1724320605}, permission.Uuid) {
			continue
		}

		// 暂时去掉外卖管理
		if permission.ID == 1626688443 {
			continue
		}
		// 授权无进销存权限
		if companySetting.SaleStock == 0 && slices.Contains([]int{1711006072, 1711009130}, permission.ID) {
			continue
		}
		// 授权无会员权限
		if companySetting.IsOpenMember == 0 && slices.Contains([]int{1636183779, 1704881218}, permission.ID) {
			continue
		}
		// 授权无平板点餐权限
		if companySetting.IsOpenTablet == 0 && permission.ID == 87 {
			continue
		}
		// 授权无H5点餐权限
		if companySetting.IsOpenH5 == 0 && permission.ID == 1724220505 {
			continue
		}
		// 授权无点餐助手权限
		if companySetting.IsOpenAssistant == 0 && permission.ID == 1720753338 {
			continue
		}
		// 授权无后厨权限
		if companySetting.IsOpenKitchenKds == 0 && permission.ID == 88 {
			continue
		}
		// 授权无自助餐权限
		if companySetting.IsOpenBuffet == 0 && permission.ID == 1708671616 {
			continue
		}
		// 授权无扫码点餐接单权限
		if companySetting.IsOpenH5Order == 0 && permission.ID == 1724320522 {
			continue
		}
		// 授权无外送权限
		if companySetting.DeliveryStatus != 1 && permission.Uuid == 1752716650 {
			continue
		}
		// 总部无品采收货权限
		if permission.Uuid == 2858560786432000 && companySetting.IsHeadquarter() {
			continue
		}
		// 未对接erp无进销存权限
		if permission.Uuid == 2857919057920000 && !company.IsOpenErp() {
			continue
		}
		filteredPermissions = append(filteredPermissions, permission)
	}
	return filteredPermissions
}

// 构建权限树
func (s *roleAccessSrv) buildPermissionTree(permissions []resp.Permission, routerName constant.RouteName) []*resp.Permission {
	roots := s.buildPermissionTreeWithoutFilter(permissions)
	var filteredRoots []*resp.Permission
	for _, root := range roots {
		if root.Name == string(routerName) {
			filteredRoots = append(filteredRoots, root.Children...)
		}
	}
	return filteredRoots
}

// buildPermissionTreeWithoutFilter 构建权限树（不进行路由过滤）
func (s *roleAccessSrv) buildPermissionTreeWithoutFilter(permissions []resp.Permission) []*resp.Permission {
	permissionMap := make(map[uint64]*resp.Permission)
	var roots []*resp.Permission
	var accessIds []string
	format := "%03d%020d"

	// 第一步：建立ID到节点的映射
	for i := range permissions {
		permission := &permissions[i]
		permissionMap[permission.Uuid] = permission
		accessIds = append(accessIds, fmt.Sprintf(format, permission.Sort, permission.Uuid))
	}

	sort.Strings(accessIds)
	// 第二步：构建树结构

	for _, accessId := range accessIds {
		for _, permission := range permissionMap {
			if fmt.Sprintf(format, permission.Sort, permission.Uuid) != accessId {
				continue
			}
			if parent, exists := permissionMap[permission.ParentUuid]; exists {
				parent.Children = append(parent.Children, permission)
			} else {
				roots = append(roots, permission)
			}
		}
	}

	return roots
}

func (s *roleAccessSrv) GetApiPermission(staffUuid, companyUuid uint64) ([]string, error) {
	accesses, _, _, err := s.getDbPermissions(staffUuid, companyUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	var permissions []string
	for _, access := range accesses {
		if !slices.Contains(permissions, access.Path) {
			permissions = append(permissions, access.Path)
		}
	}
	return permissions, nil
}

// GetCompanyPermissionTree 获取店铺的所有权限树（不依赖员工，用于角色权限配置）
func (s *roleAccessSrv) GetCompanyPermissionGroup(companyUuid uint64, includeRouteNames []constant.RouteName) (resp.PermissionGroup, error) {
	db := s.dbm.GetDB(companyUuid)
	accessRepo := repository.NewAccessRepo(db)
	companySettingRepo := repository.NewCompanySettingRepo(db)
	company, err := repository.NewCompanyRepo(db).GetCompany(func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", companyUuid)
	})
	if err != nil {
		return resp.PermissionGroup{}, errors.WithMessage(err, "获取店铺信息失败")
	}
	// 获取店铺设置
	companySetting, err := companySettingRepo.GetOne(func(db *gorm.DB) *gorm.DB {
		return db.Where("company_uuid = ?", companyUuid)
	})
	if err != nil {
		return resp.PermissionGroup{}, errors.WithMessage(err, "获取店铺设置失败")
	}

	// 获取所有权限（不依赖员工）
	var options []repository.DBOption
	// 如果是供应商，只获取供应商权限
	// 这里可以根据需要添加供应商判断逻辑
	dbPermissions, err := accessRepo.GetPermissions(options...)
	if err != nil {
		return resp.PermissionGroup{}, errors.WithMessage(err, "获取权限失败")
	}

	var permissions []resp.Permission
	groupPermission := resp.PermissionGroup{
		List: []*resp.Permission{},
	}

	for _, dbPermission := range dbPermissions {
		var permission resp.Permission
		copier.Copy(&permission, dbPermission)
		permission.ParentUuid = dbPermission.ParentUuid
		permission.CreateTime = time.Unix(dbPermission.CreateTime, 0).Format(time.DateTime)
		permission.UpdateTime = time.Unix(dbPermission.UpdateTime, 0).Format(time.DateTime)
		permission.Children = []*resp.Permission{}
		permissions = append(permissions, permission)
	}

	// 筛选权限
	permissions = s.filterPermission(permissions, companySetting, company)

	// 构建权限树形结构
	roots := s.buildPermissionTreeWithoutFilter(permissions)
	for _, root := range roots {
		// 只返回"管理APP"、"收银机"、"点餐助手"
		if !slices.Contains(includeRouteNames, constant.RouteName(root.Name)) {
			continue
		}

		// 如果是管理APP分组，标记所有子权限为默认勾选（前端可以根据此标记默认勾选）
		// 注意：这里不修改 Permission 结构，前端可以根据权限组名称判断
		// 如果需要后端标记，可以在 Permission 结构中添加 IsDefaultChecked 字段

		groupPermission.List = append(groupPermission.List, root)
	}

	return groupPermission, nil
}
