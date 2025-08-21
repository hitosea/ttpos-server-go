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
	PaginateGetStaffs(ctx context.Context, pageReq dto.PageReq) (resp.StaffListPaginationResp, error) // 员工列表
	AddStaff(ctx context.Context, addReq req.AddStaffReq) (error, []string)                           // 添加员工
	UpdateStaff(ctx context.Context, updateReq req.UpdateStaffReq) (error, []string)                  // 修改员工
	GetRoleList(ctx context.Context, pageReq dto.PageReq) (resp.RoleListResp, error)                  // 角色列表
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

// PaginateGetStaffs 获取员工列表
func (s *staffSrv) PaginateGetStaffs(ctx context.Context, pageReq dto.PageReq) (resp.StaffListPaginationResp, error) {
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(ctx.GetDbId()))

	staffs, total, err := staffRepo.PaginateGetStaffs(pageReq.PageNo, pageReq.PageSize, staffRepo.WithRoles())
	if err != nil {
		return resp.StaffListPaginationResp{}, errors.WithMessage(errors.New("获取员工列表失败"), err.Error())
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

// UpdateStaff 编辑员工
func (s *staffSrv) UpdateStaff(ctx context.Context, updateReq req.UpdateStaffReq) (error, []string) {
	var exists []string
	db := s.dbm.GetDB(ctx.GetDbId())
	// 获取员工
	staffRepo := repository.NewStaffRepo(db)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(updateReq.Uuid))
	if err != nil {
		return errors.WithMessage(errors.New("获取员工失败"), err.Error()), exists
	}
	if staff.IsSuper == 1 {
		return errors.New("超级员工不能修改"), exists
	}
	saasDB := s.dbm.GetDB(0)
	companyStaffRepo := repository.NewCompanyStaffRepo(saasDB)
	// 手机号、邮箱必填，可以这样判断
	existsCompanyStaff := companyStaffRepo.GetCompanyStaff(companyStaffRepo.WhereUsernameOrPhone(updateReq.Username, updateReq.Phone), companyStaffRepo.WhereNotUuid(updateReq.Uuid))
	if existsCompanyStaff.Username == updateReq.Username {
		exists = append(exists, "username")
	}
	if existsCompanyStaff.Phone == updateReq.Phone {
		exists = append(exists, "phone")
	}
	existsStaff, _ := staffRepo.GetStaff(staffRepo.WhereRealName(updateReq.RealName), staffRepo.WhereNotUuid(updateReq.Uuid))
	if existsStaff.Uuid != 0 {
		exists = append(exists, "real_name")
	}
	if len(exists) > 0 {
		return errors.New("此内容已被占用"), exists
	}
	// 判断角色参数是否正确
	roleRepo := repository.NewRoleRepo(db)
	roles, err := roleRepo.GetRoleList([]repository.DBOption{roleRepo.WhereUuids(updateReq.Roles)}...)
	if err != nil {
		return errors.WithMessage(errors.New("获取角色失败"), err.Error()), exists
	}
	if len(roles) != len(updateReq.Roles) {
		return errors.New("角色参数错误"), exists
	}
	err = saasDB.Model(&model.CompanyStaff{}).Where("uuid = ?", updateReq.Uuid).Updates(map[string]any{
		"username": updateReq.Username,
		"phone":    updateReq.Phone,
	}).Error
	if err != nil {
		return errors.WithMessage(errors.New("编辑员工失败"), err.Error()), exists
	}
	update := map[string]any{
		"username":  updateReq.Username,
		"real_name": updateReq.RealName,
		"phone":     updateReq.Phone,
	}
	if updateReq.Password != "" {
		update["password"] = utils.EncryptPassword(updateReq.Password)
		update["password_change_time"] = time.Now().Unix()
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		staffRepo := repository.NewStaffRepo(tx)
		err = staffRepo.Update(updateReq.Uuid, update)
		if err != nil {
			return err
		}
		// 更新员工角色
		err = staffRepo.UpdateStaffRoles(updateReq.Uuid, updateReq.Roles)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.WithMessage(errors.New("编辑员工失败"), err.Error()), exists
	}

	// 删除收银机缓存
	tc := cache.NewTaggedCache(s.cache)
	tc.TagClear("cashier")

	// 推送配置更新
	go websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_PERMISSION, map[string]any{
		"staff_uuid":  updateReq.Uuid,
		"update_time": time.Now().Unix(),
	})

	return nil, exists
}

// AddStaff 添加员工
func (s *staffSrv) AddStaff(ctx context.Context, addReq req.AddStaffReq) (error, []string) {
	saasDB := s.dbm.GetDB(0)
	companyStaffRepo := repository.NewCompanyStaffRepo(saasDB)
	existsCompanyStaff := companyStaffRepo.GetCompanyStaff(companyStaffRepo.WhereUsernameOrPhone(addReq.Username, addReq.Phone))
	var exists []string
	if existsCompanyStaff.Username == addReq.Username {
		exists = append(exists, "username")
	}
	if existsCompanyStaff.Phone == addReq.Phone {
		exists = append(exists, "phone")
	}

	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	staffRepo := repository.NewStaffRepo(db)
	staff, _ := staffRepo.GetStaff(staffRepo.WhereRealName(addReq.RealName))
	if staff.Uuid != 0 {
		exists = append(exists, "real_name")
	}
	if len(exists) > 0 {
		return errors.New("此内容已被占用"), exists
	}
	// 判断角色是否存在
	roleRepo := repository.NewRoleRepo(db)
	roles, err := roleRepo.GetRoleList([]repository.DBOption{roleRepo.WhereUuids(addReq.Roles)}...)
	if err != nil {
		return errors.WithMessage(errors.New("角色不存在"), err.Error()), exists
	}
	if len(roles) != len(addReq.Roles) {
		return errors.New("角色参数错误"), exists
	}

	companyStaff := model.CompanyStaff{
		Username:    addReq.Username,
		Phone:       addReq.Phone,
		IsSuper:     0,
		CompanyUuid: ctx.GetCompanyUuid(),
	}
	// 保存saas库
	saasDB.Model(&model.CompanyStaff{}).Create(&companyStaff)

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
		// 保存员工角色
		err = repository.NewStaffRepo(tx).UpdateStaffRoles(staff.Uuid, addReq.Roles)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return errors.New("添加员工失败"), exists
	}
	return nil, exists
}

// GetRoleList 获取角色列表
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
