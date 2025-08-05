package service

import (
	"ttpos-server-go/app/dto"
	"ttpos-server-go/app/dto/req"
	"ttpos-server-go/app/dto/resp"
	"ttpos-server-go/app/errors"
	"ttpos-server-go/app/model"
	"ttpos-server-go/app/repository"
	"ttpos-server-go/pkg/context"
	"ttpos-server-go/pkg/database"
	"ttpos-server-go/pkg/logger"
	"ttpos-server-go/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IStaffSrv interface {
	// 获取管理员列表
	GetStaffs(ctx context.Context, pageReq dto.PageReq) (resp.StaffListPaginationResp, error)
	// 修改管理员
	UpdateStaff(ctx context.Context, updateReq req.UpdateStaffReq) error
	// 设置启用禁用员工
	UpdateStaffStatus(ctx context.Context, updateReq req.UpdateStaffStatusReq) error
	// 删除员工
	DeleteStaff(ctx context.Context, deleteReq req.DeleteStaffReq) error
	// 添加管理员
	AddStaff(ctx context.Context, addReq req.AddStaffReq) error
	// 获取角色列表
	GetRoleList(ctx context.Context, pageReq dto.PageReq) (resp.RoleListResp, error)
	// 添加角色
	AddRole(ctx context.Context, addReq req.AddRoleReq) error
	// 修改角色
	UpdateRole(ctx context.Context, updateReq req.UpdateRoleReq) error
	// 删除角色
	DeleteRole(ctx context.Context, deleteReq req.DeleteRoleReq) error
	// 获取角色详细
	GetRoleAccess(ctx context.Context, getReq req.GetRoleReq) (resp.RoleDetailResp, error)
}

type staffSrv struct {
	dbm *database.DBManager
}

func NewStaffSrvImpl(dbm *database.DBManager) IStaffSrv {
	return &staffSrv{
		dbm: dbm,
	}
}

func NewStaffSrv(dbm *database.DBManager) IStaffSrv {
	return NewStaffSrvImpl(dbm)
}

func (s *staffSrv) GetStaffs(ctx context.Context, pageReq dto.PageReq) (resp.StaffListPaginationResp, error) {
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(ctx.GetDbId()))

	staffs, total, err := staffRepo.PaginateGetStaffs(pageReq.PageNo, pageReq.PageSize, staffRepo.WithRoles())
	if err != nil {
		return resp.StaffListPaginationResp{}, err
	}

	staffList := make([]resp.Staff, 0, len(staffs))

	for _, staff := range staffs {

		logger.Logger.Info("staff", zap.Any("staff", staff))

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
	return nil
}

func (s *staffSrv) UpdateStaffStatus(ctx context.Context, updateReq req.UpdateStaffStatusReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	staffRepo := repository.NewStaffRepo(db)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(updateReq.Uuid))
	if err != nil {
		return err
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
	db := s.dbm.GetDB(ctx.GetDbId())
	staffRepo := repository.NewStaffRepo(db)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(deleteReq.Uuid))
	if err != nil {
		return err
	}
	if staff.IsSuper == 1 {
		return errors.New("超级管理员不能删除")
	}

	err = db.Model(&model.Staff{}).Where("uuid = ?", deleteReq.Uuid).Delete(&model.Staff{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *staffSrv) AddStaff(ctx context.Context, addReq req.AddStaffReq) error {
	db := s.dbm.GetDB(ctx.GetDbId())
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
	staff = model.Staff{
		Username:  addReq.Username,
		RealName:  addReq.RealName,
		Phone:     addReq.Phone,
		Password:  utils.EncryptPassword(addReq.Password),
		IsDisable: 0,
		IsSuper:   0,
	}
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
