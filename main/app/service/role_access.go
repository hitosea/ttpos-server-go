package service

import (
	"fmt"
	"slices"
	"sort"
	"time"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	objectStorageAdapter "ttpos-server-go/app/modules/objectstorage/infrastructure/adapter"
	objectStoragePersistence "ttpos-server-go/app/modules/objectstorage/infrastructure/persistence"
	"ttpos-server-go/i18n"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
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
	GetCompanyPermissionGroup(ctx context.Context, includeRouteNames []constant.RouteName) (resp.PermissionGroup, error)
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
		// 暂时去掉收银交班权限
		if permission.Uuid == 1704881155 {
			continue
		}
		// 暂时去掉外卖管理
		if permission.Uuid == 1626688443 {
			continue
		}
		// 授权无进销存权限
		if companySetting.SaleStock == 0 && slices.Contains([]uint64{1711006072, 1711009130}, permission.Uuid) {
			continue
		}
		// 授权无会员权限
		if companySetting.IsOpenMember == 0 && slices.Contains([]uint64{1636183779, 1704881218}, permission.Uuid) {
			continue
		}
		// 授权无平板点餐权限
		if companySetting.IsOpenTablet == 0 && permission.Uuid == 87 {
			continue
		}
		// 授权无H5点餐权限
		if companySetting.IsOpenH5 == 0 && permission.Uuid == 1724220505 {
			continue
		}
		// 授权无点餐助手权限
		if companySetting.IsOpenAssistant == 0 && permission.Uuid == 1720753338 {
			continue
		}
		// 授权无后厨权限
		if companySetting.IsOpenKitchenKds == 0 && permission.Uuid == 88 {
			continue
		}
		// 授权无自助餐权限
		if companySetting.IsOpenBuffet == 0 && permission.Uuid == 1708671616 {
			continue
		}
		// 授权无扫码点餐接单权限
		if companySetting.IsOpenH5Order == 0 && permission.Uuid == 1724320522 {
			continue
		}
		// 授权无外送权限
		if companySetting.DeliveryStatus != 1 && permission.Uuid == 1752716650 {
			continue
		}
		// 新管理端-管理APP-总部无品采收货权限
		if permission.Uuid == 2858548203520000 && companySetting.IsHeadquarter() {
			continue
		}
		// 新管理端-管理APP-未对接erp无进销存权限
		if permission.Uuid == 2857919057920000 && !company.IsOpenErp() {
			continue
		}
		// 新管理端-管理APP-云平台未开启高级票据打印，权限列表无高级票据样式设置
		if permission.Uuid == 2859181543424000 && companySetting.IsOpenAdvancedTicketPrint == 0 {
			continue
		}
		// 新管理端-管理APP-云平台未开启桌台地图，权限列表无桌台地图
		if permission.Uuid == 2859106045952001 && !companySetting.IsOpenTableMap() {
			continue
		}
		// 新管理端-管理APP-散户无品牌采购、品牌收货、调拨单权限
		if slices.Contains([]uint64{2858468511744000, 2858548203520000, 2858825027584000}, permission.Uuid) && companySetting.IsTtposSite() {
			continue
		}
		// 新管理端-非总部不显示门店管理
		if permission.Uuid == 2856866287616001 && !companySetting.IsHeadquarter() {
			continue
		}
		// 新管理端-管理APP-云平台未开启自助点餐机，权限列表无自助点餐机设置
		if permission.Uuid == 2859353116672000 && !companySetting.IsOpenKiosk() {
			continue
		}

		// 新管理端-管理APP-云平台未开启Grab外卖，权限列表无Grab外卖设置
		if slices.Contains([]uint64{2857076002816000, 2857096974336000, 2857117945856000, 2857138917376000}, permission.Uuid) && !companySetting.IsOpenGrabDelivery() {
			continue
		}
		// 新管理端-管理APP-云平台未开启LINE MAN外卖，权限列表无LINE MAN外卖设置
		if slices.Contains([]uint64{2857159888896001, 2857180860416001, 2857201831936000, 2857222803456001, 2859028879360000}, permission.Uuid) {
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
	// 检查是否启用缓存（需要全局开关开启且门店在白名单内）
	enableCache := objectStorageAdapter.IsObjectStorageCacheEnabled(companyUuid)

	// 查询函数（从数据库获取权限）
	queryFunc := func() ([]string, error) {
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

	var permissions []string
	var err error

	if enableCache {
		// 使用缓存（缓存配置和初始化已在 objectstorage 模块中完成）
		cacheLayer := objectStorageAdapter.GetApiPermissionCache[[]string](cache.Global)
		cacheKey := objectStoragePersistence.BuildApiPermissionKey(companyUuid, staffUuid)
		permissions, err = cacheLayer.GET(cacheKey, queryFunc)
	} else {
		// 不使用缓存，直接查询数据库
		permissions, err = queryFunc()
	}

	if err != nil {
		return nil, errors.WithMessage(err)
	}
	return permissions, nil
}

// GetCompanyPermissionGroup 获取店铺的所有权限组（不依赖员工，用于角色权限配置）
func (s *roleAccessSrv) GetCompanyPermissionGroup(ctx context.Context, includeRouteNames []constant.RouteName) (resp.PermissionGroup, error) {
	company := ctx.GetCompany()
	companyUuid := company.Uuid
	db := s.dbm.GetDB(company.Uuid)
	accessRepo := repository.NewAccessRepo(db)
	companySettingRepo := repository.NewCompanySettingRepo(db)

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
	permissions = s.filterPermission(permissions, companySetting, *company)

	// 构建权限树形结构
	roots := s.buildPermissionTreeWithoutFilter(permissions)
	for _, root := range roots {
		// 只返回"管理APP"、"收银机"、"点餐助手"
		if !slices.Contains(includeRouteNames, constant.RouteName(root.Name)) {
			continue
		}

		groupPermission.List = append(groupPermission.List, root)
	}

	// 遍历groupPermission.List，将每个权限以及其所有子孙权限的名称翻译为对应语言
	for _, root := range groupPermission.List {
		s.translatePermission(root, ctx.GetLanguage())
	}
	return groupPermission, nil
}

// 翻译权限及其所有子孙权限的名称
func (s *roleAccessSrv) translatePermission(permission *resp.Permission, language string) {
	permission.Name = i18n.Translate(language, permission.Name)
	for _, child := range permission.Children {
		s.translatePermission(child, language)
	}
}
