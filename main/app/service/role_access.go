package service

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"time"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/pkg/database"

	"github.com/jinzhu/copier"

	"ttpos-server-go/app/constant"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
)

type IRoleAccessSrv interface {
	GetPermission(routerName constant.RouteName, staffUuid, companyUuid uint64) ([]*resp.Permission, error)
	GetApiPermission(staffUuid, companyUuid uint64) ([]string, error)
}

func NewRoleAccessSrv(dbm *database.DBManager) IRoleAccessSrv {
	return NewRoleAccessSrvImpl(dbm)
}

type RoleAccessSrv struct {
	dbm *database.DBManager
}

func NewRoleAccessSrvImpl(dbm *database.DBManager) *RoleAccessSrv {
	return &RoleAccessSrv{
		dbm: dbm,
	}
}

// 从数据库获取权限
func (s *RoleAccessSrv) getDbPermissions(staffUuid, companyUuid uint64) ([]model.Access, model.CompanySetting, error) {
	accessRepo := repository.NewAccessRepo(s.dbm.GetDB(companyUuid))
	var companySetting model.CompanySetting
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(companyUuid))
	staff := staffRepo.GetByUuid(staffUuid, staffRepo.WithCompany(), staffRepo.WithCompanySetting())

	if staff.Company == nil || staff.Company.CompanySetting == nil {
		return nil, companySetting, errors.New("获取商家信息错误")
	}

	companySetting = *staff.Company.CompanySetting

	var where []repository.Where

	if staff.IsSuper == 1 { // 超级管理员
		if staff.UserType == 1 {
			where = append(where, accessRepo.WhereIsSupplier())
		}
	} else {
		roleUuids, err := repository.NewStaffRoleRepo(s.dbm.GetDB(staff.CompanyUuid)).GetRoleUuidsByStaffUuid(staff.Uuid)
		if err != nil {
			return nil, companySetting, errors.New("获取用户角色失败")
		}
		accessUuids, err := accessRepo.GetAccessUuids(roleUuids)
		if err != nil {
			return nil, companySetting, errors.New("获取角色权限失败")
		}
		where = append(where, accessRepo.WhereUuids(accessUuids))
	}

	dbPermissions, err := accessRepo.GetPermissions(where...)

	if err != nil {
		return nil, companySetting, errors.New("获取权限失败")
	}

	return dbPermissions, companySetting, nil
}

// GetPermission 获取权限
func (s *RoleAccessSrv) GetPermission(routerName constant.RouteName, staffUuid, companyUuid uint64) ([]*resp.Permission, error) {

	var permissions []resp.Permission
	dbPermissions, companySetting, err := s.getDbPermissions(staffUuid, companyUuid)
	if err != nil {
		return nil, err
	}

	for _, dbPermission := range dbPermissions {
		var permission resp.Permission
		copier.Copy(&permission, dbPermission)
		permission.CreateTime = time.Unix(int64(dbPermission.CreateTime), 0).Format(time.DateTime)
		permission.UpdateTime = time.Unix(int64(dbPermission.UpdateTime), 0).Format(time.DateTime)
		permission.Children = []*resp.Permission{}
		permissions = append(permissions, permission)
	}

	permissions = s.filterPermission(permissions, companySetting)

	return s.buildPermissionTree(permissions, routerName), nil
}

// 筛选权限
func (s *RoleAccessSrv) filterPermission(permissions []resp.Permission, companySetting model.CompanySetting) []resp.Permission {
	var filteredPermissions []resp.Permission
	for _, permission := range permissions {
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
		filteredPermissions = append(filteredPermissions, permission)
	}
	return filteredPermissions
}

// 构建权限树
func (s *RoleAccessSrv) buildPermissionTree(permissions []resp.Permission, routerName constant.RouteName) []*resp.Permission {
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

	var filteredRoots []*resp.Permission
	for _, root := range roots {
		if root.Name == string(routerName) {
			filteredRoots = append(filteredRoots, root.Children...)
		}
	}

	return filteredRoots
}

func (s *RoleAccessSrv) GetApiPermission(staffUuid, companyUuid uint64) ([]string, error) {
	accesses, _, err := s.getDbPermissions(staffUuid, companyUuid)
	if err != nil {
		return nil, err
	}
	var permissions []string
	for _, access := range accesses {
		if !slices.Contains(permissions, access.ApiPath) {
			permissions = append(permissions, access.ApiPath)
		}
	}
	return permissions, nil
}
