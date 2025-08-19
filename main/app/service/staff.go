package service

import (
	"time"
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/cache"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/utils"
	"ttpos-server-go/pkg/websocket"

	"gorm.io/gorm"
)

type IStaffSrv interface {
	PaginateGetStaffs(ctx context.Context, pageReq dto.PageReq) (resp.StaffListPaginationResp, error) // 获取管理员列表
	UpdateStaff(ctx context.Context, updateReq req.UpdateStaffReq) error                              // 修改管理员
	UpdateStaffStatus(ctx context.Context, updateReq req.UpdateStaffStatusReq) error                  // 设置启用禁用管理员
	DeleteStaff(ctx context.Context, deleteReq req.DeleteStaffReq) error                              // 删除管理员
	AddStaff(ctx context.Context, addReq req.AddStaffReq) error                                       // 添加管理员
	GetRoleList(ctx context.Context, pageReq dto.PageReq) (resp.RoleListResp, error)                  // 获取角色列表
	AddRole(ctx context.Context, addReq req.AddRoleReq) error                                         // 添加角色
	UpdateRole(ctx context.Context, updateReq req.UpdateRoleReq) error                                // 修改角色
	DeleteRole(ctx context.Context, deleteReq req.DeleteRoleReq) error                                // 删除角色
	GetRoleAccess(ctx context.Context, getReq req.GetRoleReq) (resp.RoleDetailResp, error)            // 获取角色详细
	GetPermissionGroup(ctx context.Context) (resp.PermissionGroup, error)                             // 获取所有角色权限
}

type staffSrv struct {
	cache         cache.Cache
	dbm           *database.DBManager
	roleAccessSrv IRoleAccessSrv
}

func NewStaffSrvImpl(dbm *database.DBManager, cache cache.Cache, roleAccessSrv IRoleAccessSrv) IStaffSrv {
	return &staffSrv{
		cache:         cache,
		dbm:           dbm,
		roleAccessSrv: roleAccessSrv,
	}
}

func NewStaffSrv(dbm *database.DBManager, cache cache.Cache, roleAccessSrv IRoleAccessSrv) IStaffSrv {
	return NewStaffSrvImpl(dbm, cache, roleAccessSrv)
}

func (s *staffSrv) PaginateGetStaffs(ctx context.Context, pageReq dto.PageReq) (resp.StaffListPaginationResp, error) {
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(ctx.GetDbId()))

	staffs, total, err := staffRepo.PaginateGetStaffs(pageReq.PageNo, pageReq.PageSize, staffRepo.WithRoles())
	if err != nil {
		return resp.StaffListPaginationResp{}, err
	}

	staffList := make([]resp.Staff, 0, len(staffs))

	for _, staff := range staffs {
		roles := make([]resp.StaffRole, 0, len(staff.Roles))
		for _, role := range staff.Roles {
			roles = append(roles, resp.StaffRole{
				Uuid: role.Uuid,
				Name: role.Name,
			})
		}
		staffList = append(staffList, resp.Staff{
			Uuid:       staff.Uuid,
			Username:   staff.Username,
			Phone:      staff.Phone,
			RealName:   staff.RealName,
			Roles:      roles,
			IsDisable:  staff.IsDisable,
			IsSuper:    staff.IsSuper,
			CreateTime: staff.CreateTime,
		})
	}
	return resp.StaffListPaginationResp{
		List: staffList,
		Meta: dto.PageResponse{
			PageNo:   pageReq.PageNo,
			PageSize: pageReq.PageSize,
			Total:    total,
		},
	}, nil
}

// 修改管理员
func (s *staffSrv) UpdateStaff(ctx context.Context, updateReq req.UpdateStaffReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取管理员
	staffRepo := repository.NewStaffRepo(db)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(updateReq.Uuid))
	if err != nil {
		return err
	}
	if staff.IsSuper == 1 {
		return errors.New("超级管理员不能修改")
	}
	// 判断角色参数是否正确
	roleRepo := repository.NewRoleRepo(db)
	roles, err := roleRepo.GetRoleList([]repository.DBOption{roleRepo.WhereUuids(updateReq.Roles)}...)
	if err != nil {
		return err
	}
	if len(roles) != len(updateReq.Roles) {
		return errors.New("角色参数错误")
	}

	saasDB := s.dbm.GetDB(0)
	companyStaffRepo := repository.NewCompanyStaffRepo(saasDB)
	// 手机号、邮箱必填，可以这样判断
	existsCompanyStaff := companyStaffRepo.GetCompanyStaff(companyStaffRepo.WhereUsernameOrPhone(updateReq.Username, updateReq.Phone), companyStaffRepo.WhereNotUuid(updateReq.Uuid))
	if existsCompanyStaff.Username == updateReq.Username {
		return errors.New("邮箱已存在")
	}
	if existsCompanyStaff.Phone == updateReq.Phone {
		return errors.New("手机号已存在")
	}

	err = saasDB.Model(&model.CompanyStaff{}).Where("uuid = ?", updateReq.Uuid).Updates(map[string]any{
		"username": updateReq.Username,
		"phone":    updateReq.Phone,
	}).Error
	if err != nil {
		return err
	}

	update := map[string]any{
		"username":  updateReq.Username,
		"real_name": updateReq.RealName,
		"phone":     updateReq.Phone,
	}
	if updateReq.Password != "" {
		update["password"] = utils.EncryptPassword(updateReq.Password)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		staffRepo := repository.NewStaffRepo(tx)
		err = staffRepo.Update(updateReq.Uuid, update)
		if err != nil {
			return err
		}
		// 更新管理员角色
		err = staffRepo.UpdateStaffRoles(updateReq.Uuid, updateReq.Roles)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "更新管理员失败")
	}

	// 删除收银机缓存
	tc := cache.NewTaggedCache(s.cache)
	tc.TagClear("cashier")

	// 推送配置更新
	go websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_PERMISSION, map[string]any{
		"staff_uuid":  updateReq.Uuid,
		"update_time": time.Now().Unix(),
	})

	return nil
}

func (s *staffSrv) UpdateStaffStatus(ctx context.Context, updateReq req.UpdateStaffStatusReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	staffRepo := repository.NewStaffRepo(db)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(updateReq.Uuid))
	if err != nil {
		return err
	}
	if staff.CashierOnline == 1 {
		return errors.New("当前人员未交班，请先交班")
	}
	if staff.IsSuper == 1 {
		return errors.New("超级管理员不能修改")
	}

	statusMap := map[int]int{
		1: 0,
		0: 1,
	}
	err = staffRepo.Update(updateReq.Uuid, map[string]any{
		"is_disable": statusMap[*updateReq.Status],
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *staffSrv) DeleteStaff(ctx context.Context, deleteReq req.DeleteStaffReq) error {
	if ctx.GetStaffUuid() == deleteReq.Uuid {
		return errors.New("不能删除当前登录账号")
	}
	db := s.dbm.GetDB(ctx.GetDbId())
	staffRepo := repository.NewStaffRepo(db)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(deleteReq.Uuid))
	if err != nil {
		return err
	}
	if staff.CashierOnline == 1 {
		return errors.New("当前人员未交班，请先交班")
	}
	if staff.IsSuper == 1 {
		return errors.New("超级管理员不能删除")
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(&model.Staff{}).Where("uuid = ?", deleteReq.Uuid).Delete(&model.Staff{}).Error
		if err != nil {
			return err
		}

		err = tx.Model(&model.StaffRole{}).Where("staff_uuid = ?", deleteReq.Uuid).Delete(&model.StaffRole{}).Error
		if err != nil {
			return err
		}

		return nil
	})
	s.dbm.GetDB(0).Model(&model.CompanyStaff{}).Where("uuid = ?", deleteReq.Uuid).Delete(&model.CompanyStaff{})
	if err != nil {
		return err
	}
	return nil
}

func (s *staffSrv) AddStaff(ctx context.Context, addReq req.AddStaffReq) error {
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	staffRepo := repository.NewStaffRepo(db)
	staff, _ := staffRepo.GetStaff(staffRepo.WhereUsername(addReq.Username))
	if staff.Uuid != 0 {
		return errors.New("管理员已存在")
	}
	// 判断角色是否存在
	roleRepo := repository.NewRoleRepo(db)
	roles, err := roleRepo.GetRoleList([]repository.DBOption{roleRepo.WhereUuids(addReq.Roles)}...)
	if err != nil {
		return err
	}
	if len(roles) != len(addReq.Roles) {
		return errors.New("角色参数错误")
	}

	companyStaff := model.CompanyStaff{
		Username:    addReq.Username,
		Phone:       addReq.Phone,
		IsSuper:     0,
		CompanyUuid: ctx.GetCompanyUuid(),
	}
	// 保存saas库
	s.dbm.GetDB(0).Model(&model.CompanyStaff{}).Create(&companyStaff)

	staff = model.Staff{
		CompanyUuid: ctx.GetCompanyUuid(),
		Username:    addReq.Username,
		RealName:    addReq.RealName,
		Phone:       addReq.Phone,
		Password:    utils.EncryptPassword(addReq.Password),
		IsDisable:   0,
		IsSuper:     0,
	}
	// 确保关联得上
	staff.Uuid = companyStaff.Uuid

	err = db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.Staff{}).Create(&staff).Error
		if err != nil {
			return err
		}
		// 保存管理员角色
		err = repository.NewStaffRepo(tx).UpdateStaffRoles(staff.Uuid, addReq.Roles)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "添加管理员失败")
	}
	return nil
}

func (s *staffSrv) GetRoleList(ctx context.Context, pageReq dto.PageReq) (resp.RoleListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	roleRepo := repository.NewRoleRepo(db)
	roles, total, err := roleRepo.PaginateGetRoleList(pageReq.PageNo, pageReq.PageSize)
	if err != nil {
		return resp.RoleListResp{}, err
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

func (s *staffSrv) AddRole(ctx context.Context, addReq req.AddRoleReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	roleRepo := repository.NewRoleRepo(db)
	role, _ := roleRepo.GetRole(roleRepo.WhereName(addReq.Name))
	if role.Uuid != 0 {
		return errors.New("角色已存在")
	}
	role = model.Role{
		Name: addReq.Name,
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.Role{}).Create(&role).Error
		if err != nil {
			return err
		}

		// 保存角色权限
		err = repository.NewRoleRepo(tx).UpdateRoleAccess(role.Uuid, addReq.AccessUuids)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "添加角色失败")
	}
	return nil
}

func (s *staffSrv) UpdateRole(ctx context.Context, updateReq req.UpdateRoleReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	roleRepo := repository.NewRoleRepo(db)
	role, err := roleRepo.GetRole(roleRepo.WhereUuid(updateReq.Uuid))
	if err != nil {
		return err
	}
	if role.Uuid == 0 {
		return errors.New("角色不存在")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.Role{}).Where("uuid = ?", updateReq.Uuid).Updates(map[string]any{
			"name": updateReq.Name,
		}).Error
		if err != nil {
			return err
		}
		// 保存角色权限
		err = repository.NewRoleRepo(tx).UpdateRoleAccess(role.Uuid, updateReq.AccessUuids)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "修改角色失败")
	}
	return nil
}

func (s *staffSrv) DeleteRole(ctx context.Context, deleteReq req.DeleteRoleReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	roleRepo := repository.NewRoleRepo(db)
	role, err := roleRepo.GetRole(roleRepo.WhereUuid(deleteReq.Uuid))
	if err != nil {
		return err
	}
	if role.Uuid == 0 {
		return errors.New("角色不存在")
	}

	// 判断角色是否被管理员使用
	staffRepo := repository.NewStaffRepo(db)
	staffs := staffRepo.GetStaffs(staffRepo.WhereRoleUuid(deleteReq.Uuid))
	if err != nil {
		return err
	}
	if len(staffs) > 0 {
		return errors.New("当前角色下存在用户，不允许删除")
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		err := tx.Model(&model.Role{}).Where("uuid = ?", deleteReq.Uuid).Delete(&model.Role{}).Error
		if err != nil {
			return err
		}
		// 删除角色权限
		err = repository.NewRoleRepo(tx).UpdateRoleAccess(role.Uuid, []uint64{})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(err, "删除角色失败")
	}
	return nil
}

func (s *staffSrv) GetRoleAccess(ctx context.Context, getReq req.GetRoleReq) (resp.RoleDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	roleRepo := repository.NewRoleRepo(db)
	role, err := roleRepo.GetRole(roleRepo.WithAccesses(), roleRepo.WhereUuid(getReq.Uuid))
	if err != nil {
		return resp.RoleDetailResp{}, err
	}
	accessUuids := make([]uint64, 0)
	for _, access := range role.Accesses {
		accessUuids = append(accessUuids, access.Uuid)
	}
	return resp.RoleDetailResp{
		Uuid:        role.Uuid,
		AccessUuids: accessUuids,
		Name:        role.Name,
	}, nil
}

func (s *staffSrv) GetPermissionGroup(ctx context.Context) (resp.PermissionGroup, error) {
	permissionGroups, err := s.roleAccessSrv.GetPermissionGroup(ctx.GetStaffUuid(), ctx.GetCompanyUuid())
	if err != nil {
		return resp.PermissionGroup{}, err
	}
	return permissionGroups, nil
}
