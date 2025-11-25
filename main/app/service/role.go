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
)

// IRoleSrv 角色服务接口
type IRoleSrv interface {
	GetRoleList(ctx context.Context, pageReq dto.PageReq) (resp.RoleListResp, error)                                     // 获取角色列表
	GetRoleDetail(ctx context.Context, uuid uint64) (resp.RoleDetailResp, error)                                         // 获取角色详情
	CreateRole(ctx context.Context, createReq req.AddRoleReq) (*resp.Role, error)                                        // 创建角色
	UpdateRole(ctx context.Context, updateReq req.UpdateRoleReq) error                                                   // 更新角色
	DeleteRole(ctx context.Context, deleteReq req.DeleteRoleReq) error                                                   // 删除角色
	GetCompanyPermissionGroup(ctx context.Context, includeRouteNames []constant.RouteName) (resp.PermissionGroup, error) // 获取店铺权限树（用于角色权限配置）
}

type roleSrv struct {
	dbm           *database.DBManager
	roleAccessSrv IRoleAccessSrv
}

func NewRoleSrv(dbm *database.DBManager, roleAccessSrv IRoleAccessSrv) IRoleSrv {
	return NewRoleSrvImpl(dbm, roleAccessSrv)
}

func NewRoleSrvImpl(dbm *database.DBManager, roleAccessSrv IRoleAccessSrv) IRoleSrv {
	return &roleSrv{
		dbm:           dbm,
		roleAccessSrv: roleAccessSrv,
	}
}

// GetRoleList 获取角色列表
func (s *roleSrv) GetRoleList(ctx context.Context, pageReq dto.PageReq) (resp.RoleListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	roleRepo := repository.NewRoleRepo(db)

	roles, total, err := roleRepo.PaginateGetRoleList(pageReq.PageNo, pageReq.PageSize)
	if err != nil {
		return resp.RoleListResp{}, errors.WithMessage(err, "获取角色列表失败")
	}

	roleList := make([]resp.Role, 0, len(roles))
	for _, role := range roles {
		roleList = append(roleList, resp.Role{
			Uuid:       role.Uuid,
			Name:       role.Name,
			CreateTime: role.CreateTime,
		})
	}

	return resp.RoleListResp{
		List: roleList,
		Meta: dto.PageResponse{
			PageNo:   pageReq.PageNo,
			PageSize: pageReq.PageSize,
			Total:    total,
		},
	}, nil
}

// GetRoleDetail 获取角色详情
func (s *roleSrv) GetRoleDetail(ctx context.Context, uuid uint64) (resp.RoleDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	roleRepo := repository.NewRoleRepo(db)
	staffRoleRepo := repository.NewStaffRoleRepo(db)

	role, err := roleRepo.GetRole(roleRepo.WhereUuid(uuid), roleRepo.WithAccesses())
	if err != nil {
		return resp.RoleDetailResp{}, errors.WithMessage(err, "获取角色详情失败")
	}

	// 获取角色权限UUID列表
	accessUuids := make([]uint64, 0, len(role.Accesses))
	accessUuidSet := make(map[uint64]bool)
	for _, access := range role.Accesses {
		accessUuids = append(accessUuids, access.AccessUuid)
		accessUuidSet[access.AccessUuid] = true
	}

	// 获取角色关联的员工UUID列表
	staffUuids, _ := staffRoleRepo.GetStaffUuidsByRoleUuid(role.Uuid)

	// 获取权限树（排除管理后台）
	permissionGroup, err := s.GetCompanyPermissionGroup(ctx, []constant.RouteName{constant.ShopAppRouteName, constant.CashierRouteName, constant.AssistantRouteName})
	if err != nil {
		return resp.RoleDetailResp{}, errors.WithMessage(err, "获取权限树失败")
	}

	// 统计各个权限组的叶子节点数量之和（管理APP、收银机、点餐助手）
	totalLeafCount := 0
	selectedLeafCount := 0

	for _, root := range permissionGroup.List {
		// 统计该权限组的叶子节点数量
		leafCount := s.countLeafNodes(root)
		totalLeafCount += leafCount

		// 统计角色已选择权限中的叶子节点数量
		selectedLeafCount += s.countSelectedLeafNodes(root, accessUuidSet)
	}

	return resp.RoleDetailResp{
		Uuid:              role.Uuid,
		Name:              role.Name,
		AccessUuids:       accessUuids,
		StaffCount:        len(staffUuids),
		StaffUuids:        staffUuids,
		SelectedLeafCount: selectedLeafCount,
		TotalLeafCount:    totalLeafCount,
	}, nil
}

// countLeafNodes 统计权限树的叶子节点数量
func (s *roleSrv) countLeafNodes(permission *resp.Permission) int {
	if len(permission.Children) == 0 {
		return 1
	}
	count := 0
	for _, child := range permission.Children {
		count += s.countLeafNodes(child)
	}
	return count
}

// countSelectedLeafNodes 统计已选择权限中的叶子节点数量
func (s *roleSrv) countSelectedLeafNodes(permission *resp.Permission, selectedSet map[uint64]bool) int {
	if len(permission.Children) == 0 {
		// 叶子节点
		if selectedSet[permission.Uuid] {
			return 1
		}
		return 0
	}
	count := 0
	for _, child := range permission.Children {
		count += s.countSelectedLeafNodes(child, selectedSet)
	}
	return count
}

// CreateRole 创建角色
func (s *roleSrv) CreateRole(ctx context.Context, createReq req.AddRoleReq) (*resp.Role, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	roleRepo := repository.NewRoleRepo(db)

	// 检查角色名称是否已存在
	existingRole, _ := roleRepo.GetRole(roleRepo.WhereName(createReq.Name))
	if existingRole.Uuid > 0 {
		return nil, errors.New("角色名称已存在")
	}

	// 创建角色（UUID 和时间字段由 GORM 自动生成）
	role := model.Role{
		Name: createReq.Name,
		Sort: 100,
	}

	roleUuid, err := roleRepo.CreateRole(role)
	if err != nil {
		return nil, errors.WithMessage(err, "创建角色失败")
	}

	// 更新角色权限（createReq.AccessUuids 已通过验证，至少包含一个权限）
	err = roleRepo.UpdateRoleAccess(roleUuid, createReq.AccessUuids)
	if err != nil {
		return nil, errors.WithMessage(err, "更新角色权限失败")
	}

	return &resp.Role{
		Uuid:       roleUuid,
		Name:       role.Name,
		CreateTime: int64(role.CreateTime),
	}, nil
}

// UpdateRole 更新角色
func (s *roleSrv) UpdateRole(ctx context.Context, updateReq req.UpdateRoleReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	roleRepo := repository.NewRoleRepo(db)
	staffRoleRepo := repository.NewStaffRoleRepo(db)

	// 检查角色是否存在
	role, err := roleRepo.GetRole(roleRepo.WhereUuid(updateReq.Uuid))
	if err != nil {
		return errors.WithMessage(err, "角色不存在")
	}

	// 检查角色名称是否已被其他角色使用
	if updateReq.Name != role.Name {
		existingRole, _ := roleRepo.GetRole(roleRepo.WhereName(updateReq.Name))
		if existingRole.Uuid > 0 && existingRole.Uuid != updateReq.Uuid {
			return errors.New("角色名称已存在")
		}
	}

	// 更新角色信息（UpdateTime 由 GORM 自动更新）
	role.Name = updateReq.Name
	err = roleRepo.UpdateRole(uint(updateReq.Uuid), role)
	if err != nil {
		return errors.WithMessage(err, "更新角色失败")
	}

	// 更新角色权限（updateReq.AccessUuids 已通过验证，至少包含一个权限）
	err = roleRepo.UpdateRoleAccess(updateReq.Uuid, updateReq.AccessUuids)
	if err != nil {
		return errors.WithMessage(err, "更新角色权限失败")
	}

	// 更新员工角色关联（如果提供了员工列表）
	if updateReq.StaffUuids != nil {
		// 先删除原有的员工关联
		err = staffRoleRepo.DeleteStaffRolesByRoleUuid(updateReq.Uuid)
		if err != nil {
			return errors.WithMessage(err, "删除员工角色关联失败")
		}

		// 创建新的员工关联
		for _, staffUuid := range updateReq.StaffUuids {
			err = staffRoleRepo.CreateStaffRoles(staffUuid, []uint64{updateReq.Uuid})
			if err != nil {
				return errors.WithMessage(err, "创建员工角色关联失败")
			}
		}
	}

	return nil
}

// DeleteRole 删除角色
func (s *roleSrv) DeleteRole(ctx context.Context, deleteReq req.DeleteRoleReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	roleRepo := repository.NewRoleRepo(db)
	staffRoleRepo := repository.NewStaffRoleRepo(db)

	// 检查角色是否存在
	_, err := roleRepo.GetRole(roleRepo.WhereUuid(deleteReq.Uuid))
	if err != nil {
		return errors.WithMessage(err, "角色不存在")
	}

	// 检查角色是否关联员工
	staffUuids, err := staffRoleRepo.GetStaffUuidsByRoleUuid(deleteReq.Uuid)
	if err != nil {
		return errors.WithMessage(err, "检查角色关联员工失败")
	}
	if len(staffUuids) > 0 {
		return errors.New("该角色已关联员工，无法删除")
	}

	// 删除角色（软删除）
	err = roleRepo.DeleteRole(uint(deleteReq.Uuid))
	if err != nil {
		return errors.WithMessage(err, "删除角色失败")
	}

	return nil
}

// GetCompanyPermissionGroup 获取店铺权限树（用于角色权限配置）
func (s *roleSrv) GetCompanyPermissionGroup(ctx context.Context, includeRouteNames []constant.RouteName) (resp.PermissionGroup, error) {
	company := ctx.GetCompany()
	return s.roleAccessSrv.GetCompanyPermissionGroup(company.Uuid, includeRouteNames)
}
