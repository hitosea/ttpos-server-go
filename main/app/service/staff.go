package service

import (
	"time"
	"ttpos-server-go/app/constant"
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
	PaginateGetStaffs(ctx context.Context, getStaffListReq req.GetStaffListReq) (resp.StaffListPaginationResp, error) // 员工列表
	AddStaff(ctx context.Context, addReq req.AddStaffReq) (error, []string)                                           // 添加员工
	UpdateStaff(ctx context.Context, updateReq req.UpdateStaffReq) (error, []string)                                  // 修改员工
	GetRoleList(ctx context.Context, pageReq dto.PageReq) (resp.RoleListResp, error)                                  // 角色列表

	SaasPaginateGetStaffs(ctx context.Context, getStaffListReq req.GetStaffListReq) (resp.StaffListPaginationResp, error) // 统一账号员工列表
	SaasAddStaff(ctx context.Context, addReq req.AddStaffReq) (error, []string)                                           // 统一账号添加员工
	SaasUpdateStaff(ctx context.Context, updateReq req.UpdateStaffReq) (error, []string)                                  // 统一账号修改员工
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
func (s *staffSrv) PaginateGetStaffs(ctx context.Context, getStaffListReq req.GetStaffListReq) (resp.StaffListPaginationResp, error) {
	staffRepo := repository.NewStaffRepo(s.dbm.GetDB(ctx.GetDbId()))

	opts := []repository.DBOption{staffRepo.WithRoles()}
	if getStaffListReq.IsFilterSuper == 1 {
		opts = append(opts, staffRepo.WhereIsSuper(0))
	}
	if getStaffListReq.Keyword != "" {
		opts = append(opts, staffRepo.WhereRealNameOrUsernameOrPhone(getStaffListReq.Keyword))
	}

	staffs, total, err := staffRepo.PaginateGetStaffs(getStaffListReq.PageNo, getStaffListReq.PageSize, opts...)
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
			Uuid:              staff.Uuid,
			Username:          staff.Username,
			Phone:             staff.Phone,
			RealName:          staff.RealName,
			Roles:             roles,
			IsDisable:         staff.IsDisable,
			IsSuper:           staff.IsSuper,
			HasDataPermission: staff.HasDataPermission == 1,
			CreateTime:        staff.CreateTime,

			CompanyList: make([]resp.CompanyRoleInfo, 0),
		})
	}
	return resp.StaffListPaginationResp{
		List: staffList,
		Meta: dto.PageResponse{
			PageNo:   getStaffListReq.PageNo,
			PageSize: getStaffListReq.PageSize,
			Total:    total,
		},
	}, nil
}

// SaasPaginateGetStaffs 获取统一账号员工列表（跨门店查询）
func (s *staffSrv) SaasPaginateGetStaffs(ctx context.Context, getStaffListReq req.GetStaffListReq) (resp.StaffListPaginationResp, error) {
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	saasStaffRepo := repository.NewSaasStaffRepo(saasDB)
	companyStaffRepo := repository.NewCompanyStaffRepo(saasDB)
	companyRepo := repository.NewCompanyRepo(saasDB)

	saasStaffListPaginationResp := resp.StaffListPaginationResp{
		List: make([]resp.Staff, 0),
	}

	// 获取当前商家可看到的所有商家列表
	currentCompanyUuid := ctx.GetCompanyUuid()
	visibleCompanies, err := companyRepo.GetVisibleCompanyList(currentCompanyUuid)
	if err != nil {
		return saasStaffListPaginationResp, errors.WithMessage(errors.New("获取可见门店列表失败"), err.Error())
	}

	// 构建可见门店UUID集合
	visibleCompanyUuids := make(map[uint64]bool)
	companyUuidToName := make(map[uint64]string)
	for _, company := range visibleCompanies {
		visibleCompanyUuids[company.Uuid] = true
		companyUuidToName[company.Uuid] = company.Name
	}

	companyUuids := make([]uint64, 0)
	// 如果传递了 company_uuid，验证是否在可见门店列表中
	if getStaffListReq.CompanyUuid > 0 {
		if !visibleCompanyUuids[getStaffListReq.CompanyUuid] {
			return saasStaffListPaginationResp, errors.New("无权限查看该门店的员工")
		}
		companyUuids = append(companyUuids, getStaffListReq.CompanyUuid)
	} else {
		for companyUuid := range visibleCompanyUuids {
			companyUuids = append(companyUuids, companyUuid)
		}
	}

	// 从 saas.ttpos_company_staff 获取员工UUID列表（根据可见门店筛选）
	var allStaffUuids []uint64
	staffUuidSet := make(map[uint64]bool)
	companyStaffList, err := companyStaffRepo.GetByCompanyUuids(companyUuids)
	if err != nil {
		return saasStaffListPaginationResp, errors.WithMessage(errors.New("获取门店员工列表失败"), err.Error())
	}
	for _, cs := range companyStaffList {
		if !staffUuidSet[cs.Uuid] {
			allStaffUuids = append(allStaffUuids, cs.Uuid)
			staffUuidSet[cs.Uuid] = true
		}
	}
	if len(allStaffUuids) == 0 {
		return saasStaffListPaginationResp, nil
	}

	saasStaffList, total, err := saasStaffRepo.PaginateGetStaffs(getStaffListReq.PageNo, getStaffListReq.PageSize, saasStaffRepo.WhereUuids(allStaffUuids))
	if err != nil {
		return saasStaffListPaginationResp, errors.WithMessage(errors.New("获取员工列表失败"), err.Error())
	}

	// 构建响应数据（包含员工在各门店的角色信息）
	staffList := make([]resp.Staff, 0, len(saasStaffList))
	for _, saasStaff := range saasStaffList {
		// 获取员工关联的所有门店（在可见范围内）
		companyStaffList, _ := companyStaffRepo.GetByStaffUuid(saasStaff.Uuid)

		var isDisable int
		var isSuper int
		// 构建门店列表（包含角色信息）
		companyList := make([]resp.CompanyRoleInfo, 0)
		currentCompanyRoles := make([]resp.StaffRole, 0)
		for _, cs := range companyStaffList {

			// 如果当前门店是员工所在的门店，则获取禁用状态
			if cs.CompanyUuid == currentCompanyUuid {
				isDisable = cs.IsDisable
				isSuper = cs.IsSuper
			}

			companyUuid := cs.CompanyUuid
			// 只处理可见门店
			if !visibleCompanyUuids[companyUuid] {
				continue
			}

			// 如果传递了 company_uuid，只返回该门店的信息
			if getStaffListReq.CompanyUuid > 0 && companyUuid != getStaffListReq.CompanyUuid {
				continue
			}

			// 从门店数据库查询员工角色信息
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

			roleList := make([]resp.StaffRole, 0, len(roles))
			for _, role := range roles {
				roleList = append(roleList, resp.StaffRole{
					Uuid: role.Uuid,
					Name: role.Name,
				})
			}

			if companyUuid == currentCompanyUuid {
				currentCompanyRoles = roleList
			}

			companyName := companyUuidToName[companyUuid]
			companyList = append(companyList, resp.CompanyRoleInfo{
				CompanyUuid: companyUuid,
				CompanyName: companyName,
				Roles:       roleList,
				IsSuper:     cs.IsSuper,
				IsDisable:   cs.IsDisable,
			})
		}

		staffList = append(staffList, resp.Staff{
			Uuid:        saasStaff.Uuid,
			Username:    saasStaff.Email,
			Phone:       saasStaff.Phone,
			RealName:    saasStaff.RealName,
			Roles:       currentCompanyRoles,
			IsDisable:   isDisable,
			IsSuper:     isSuper,
			CreateTime:  saasStaff.CreateTime,
			CompanyList: companyList,
		})
	}

	return resp.StaffListPaginationResp{
		List: staffList,
		Meta: dto.PageResponse{
			PageNo:   getStaffListReq.PageNo,
			PageSize: getStaffListReq.PageSize,
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
	staff, _ := staffRepo.GetStaff(staffRepo.WhereUuid(updateReq.Uuid))
	if staff.Uuid == 0 {
		return errors.New("获取员工失败"), exists
	}
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	// saas库查询是否存在相同的邮箱或手机号
	saasStaffRepo := repository.NewSaasStaffRepo(saasDB)
	existsSaasStaff := saasStaffRepo.GetSaasStaff(saasStaffRepo.WhereEmailOrPhone(updateReq.Username, updateReq.Phone), saasStaffRepo.WhereNotUuid(updateReq.Uuid))
	// 手机号、邮箱必填，可以这样判断
	if existsSaasStaff.Email == updateReq.Username {
		exists = append(exists, "username")
	}
	if existsSaasStaff.Phone == updateReq.Phone {
		exists = append(exists, "phone")
	}
	if len(exists) > 0 {
		return errors.New("此内容已被占用"), exists
	}
	// 判断角色参数是否正确
	if len(updateReq.Roles) > 0 {
		roleRepo := repository.NewRoleRepo(db)
		roles, err := roleRepo.GetRoleList([]repository.DBOption{roleRepo.WhereUuids(updateReq.Roles)}...)
		if err != nil {
			return errors.WithMessage(errors.New("获取角色失败"), err.Error()), exists
		}
		if len(roles) != len(updateReq.Roles) {
			return errors.New("角色参数错误"), exists
		}
	}
	err := saasDB.Model(&model.CompanyStaff{}).Where("uuid = ?", updateReq.Uuid).Updates(map[string]any{
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
	saasStaffUpdate := map[string]any{
		"email":     updateReq.Username,
		"real_name": updateReq.RealName,
		"phone":     updateReq.Phone,
	}
	if updateReq.Password != "" {
		update["password"] = utils.EncryptPassword(updateReq.Password)
		update["password_change_time"] = time.Now().Unix()

		saasStaffUpdate["password"] = utils.EncryptPassword(updateReq.Password)
		saasStaffUpdate["password_change_time"] = time.Now().Unix()
	}
	if updateReq.PermissionPassword != "" {
		update["permission_password"] = utils.EncryptPassword(updateReq.PermissionPassword)
	}

	// 更新统一账号表
	if err := repository.NewSaasStaffRepo(saasDB).Update(updateReq.Uuid, saasStaffUpdate); err != nil {
		return errors.WithMessage(errors.New("编辑员工失败"), err.Error()), exists
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
	utils.Go(func() {
		websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_PERMISSION, map[string]any{
			"staff_uuid":  updateReq.Uuid,
			"update_time": time.Now().Unix(),
		})
	})

	return nil, exists
}

// AddStaff 添加员工
func (s *staffSrv) AddStaff(ctx context.Context, addReq req.AddStaffReq) (error, []string) {
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	// 判断邮箱或手机号是否被占用
	saasStaffRepo := repository.NewSaasStaffRepo(saasDB)
	existsSaasStaff := saasStaffRepo.GetSaasStaff(saasStaffRepo.WhereEmailOrPhone(addReq.Username, addReq.Phone))
	var exists []string
	if existsSaasStaff.Email == addReq.Username {
		exists = append(exists, "username")
	}
	if existsSaasStaff.Phone == addReq.Phone {
		exists = append(exists, "phone")
	}
	if len(exists) > 0 {
		return errors.New("此内容已被占用"), exists
	}
	db := s.dbm.GetDB(ctx.GetCompanyUuid())
	// 判断角色是否存在
	roleRepo := repository.NewRoleRepo(db)
	roles, err := roleRepo.GetRoleList([]repository.DBOption{roleRepo.WhereUuids(addReq.Roles)}...)
	if err != nil {
		return errors.WithMessage(errors.New("角色不存在"), err.Error()), exists
	}
	if len(roles) != len(addReq.Roles) {
		return errors.New("角色参数错误"), exists
	}

	saasStaff := model.SaasStaff{
		Email:     addReq.Username,
		Phone:     addReq.Phone,
		RealName:  addReq.RealName,
		Password:  utils.EncryptPassword(addReq.Password),
		IsDisable: 0,
	}
	saasDB.Transaction(func(tx *gorm.DB) error {
		err = tx.Model(&model.SaasStaff{}).Create(&saasStaff).Error
		if err != nil {
			return err
		}
		companyStaff := model.CompanyStaff{
			Username:    addReq.Username,
			Phone:       addReq.Phone,
			IsSuper:     0,
			IsDisable:   0,
			CompanyUuid: ctx.GetCompanyUuid(),
		}
		companyStaff.Uuid = saasStaff.Uuid
		// 保存saas库
		return tx.Model(&model.CompanyStaff{}).Create(&companyStaff).Error
	})
	if err != nil {
		return errors.WithMessage(errors.New("添加员工失败"), err.Error()), exists
	}

	staff := model.Staff{
		CompanyUuid:        ctx.GetCompanyUuid(),
		Username:           addReq.Username,
		RealName:           addReq.RealName,
		Phone:              addReq.Phone,
		Password:           utils.EncryptPassword(addReq.Password),
		PermissionPassword: utils.EncryptPassword(addReq.PermissionPassword),
		IsDisable:          0,
		IsSuper:            0,
	}
	// 确保关联得上
	staff.Uuid = saasStaff.Uuid

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

// SaasAddStaff 统一账号添加员工
func (s *staffSrv) SaasAddStaff(ctx context.Context, addReq req.AddStaffReq) (error, []string) {
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	// 判断邮箱或手机号是否被占用
	saasStaffRepo := repository.NewSaasStaffRepo(saasDB)
	existsSaasStaff := saasStaffRepo.GetSaasStaff(saasStaffRepo.WhereEmailOrPhone(addReq.Username, addReq.Phone))
	var exists []string
	if existsSaasStaff.Email == addReq.Username {
		exists = append(exists, "username")
	}
	if existsSaasStaff.Phone == addReq.Phone {
		exists = append(exists, "phone")
	}
	if len(exists) > 0 {
		return errors.New("此内容已被占用"), exists
	}

	// 在 saas 上创建 ttpos_staff
	saasStaff := model.SaasStaff{
		Email:     addReq.Username,
		Phone:     addReq.Phone,
		RealName:  addReq.RealName,
		Password:  utils.EncryptPassword(addReq.Password),
		IsDisable: 0,
	}
	if addReq.IsDisable != nil {
		saasStaff.IsDisable = *addReq.IsDisable
	}

	err := saasDB.Model(&model.SaasStaff{}).Create(&saasStaff).Error
	if err != nil {
		return errors.WithMessage(errors.New("创建统一账号失败"), err.Error()), exists
	}

	currentCompanyUuid := ctx.GetCompanyUuid()
	companySetting := ctx.GetCompanySetting()

	// 判断当前 company 是否是总部或者有子公司
	isHeadquarter := companySetting.IsHeadquarter()
	hasChildren := companySetting.HasChildren == 1

	if isHeadquarter || hasChildren {
		// 如果是总部或者有子公司，走多门店逻辑
		if len(addReq.CompanyRoleList) == 0 {
			return errors.New("多门店配置不能为空"), exists
		}

		// 获取当前门店可见的所有门店列表
		companyRepo := repository.NewCompanyRepo(saasDB)
		visibleCompanies, err := companyRepo.GetVisibleCompanyList(currentCompanyUuid)
		if err != nil {
			return errors.WithMessage(errors.New("获取可见门店列表失败"), err.Error()), exists
		}

		// 构建可见门店UUID映射
		visibleCompanyMap := make(map[uint64]bool)
		for _, company := range visibleCompanies {
			visibleCompanyMap[company.Uuid] = true
		}

		// 遍历 CompanyRoleList，验证并创建员工
		for _, companyRoleItem := range addReq.CompanyRoleList {
			// 判断 company_uuid 是否为当前公司可见
			if !visibleCompanyMap[companyRoleItem.CompanyUuid] {
				return errors.New("门店不存在或无权限访问"), exists
			}

			// 查询对应商家数据库的 RoleUuids 是否存在
			shopDB := s.dbm.GetDB(companyRoleItem.CompanyUuid)
			if shopDB == nil {
				return errors.New("获取门店数据库连接失败"), exists
			}

			roleRepo := repository.NewRoleRepo(shopDB)
			roles, err := roleRepo.GetRoleList([]repository.DBOption{roleRepo.WhereUuids(companyRoleItem.RoleUuids)}...)
			if err != nil {
				return errors.WithMessage(errors.New("获取角色失败"), err.Error()), exists
			}
			if len(roles) != len(companyRoleItem.RoleUuids) {
				return errors.New("角色参数错误"), exists
			}

			// 在对应商家数据库 ttpos_staff 添加数据
			staff := model.Staff{
				CompanyUuid:        companyRoleItem.CompanyUuid,
				Username:           addReq.Username,
				RealName:           addReq.RealName,
				Phone:              addReq.Phone,
				Password:           utils.EncryptPassword(addReq.Password),
				PermissionPassword: utils.EncryptPassword(addReq.PermissionPassword),
				IsDisable:          0,
				IsSuper:            0,
			}
			if addReq.IsDisable != nil {
				staff.IsDisable = *addReq.IsDisable
			}
			staff.Uuid = saasStaff.Uuid

			err = shopDB.Transaction(func(tx *gorm.DB) error {
				// 创建员工
				err := tx.Model(&model.Staff{}).Create(&staff).Error
				if err != nil {
					return err
				}
				// 添加员工和角色的关联关系
				err = repository.NewStaffRepo(tx).UpdateStaffRoles(staff.Uuid, companyRoleItem.RoleUuids)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return errors.WithMessage(errors.New("添加员工失败"), err.Error()), exists
			}

			// 在 saas 创建 company_staff 关联关系
			companyStaff := model.CompanyStaff{
				CompanyUuid: companyRoleItem.CompanyUuid,
				Username:    addReq.Username,
				Phone:       addReq.Phone,
				IsSuper:     0,
				IsDisable:   0,
			}
			companyStaff.Uuid = saasStaff.Uuid
			if addReq.IsDisable != nil {
				companyStaff.IsDisable = *addReq.IsDisable
			}
			err = saasDB.Model(&model.CompanyStaff{}).Create(&companyStaff).Error
			if err != nil {
				return errors.WithMessage(errors.New("创建门店关联失败"), err.Error()), exists
			}
		}
	} else {
		// 既不是总部，也没有子公司，走单门店逻辑
		if len(addReq.Roles) == 0 {
			return errors.New("角色不能为空"), exists
		}

		// 检查 Roles 参数，是否是当前商家数据库的角色
		db := s.dbm.GetDB(currentCompanyUuid)
		roleRepo := repository.NewRoleRepo(db)
		roles, err := roleRepo.GetRoleList([]repository.DBOption{roleRepo.WhereUuids(addReq.Roles)}...)
		if err != nil {
			return errors.WithMessage(errors.New("获取角色失败"), err.Error()), exists
		}
		if len(roles) != len(addReq.Roles) {
			return errors.New("角色参数错误"), exists
		}

		// 在当前商家数据库添加 ttpos_staff
		staff := model.Staff{
			CompanyUuid:        currentCompanyUuid,
			Username:           addReq.Username,
			RealName:           addReq.RealName,
			Phone:              addReq.Phone,
			Password:           utils.EncryptPassword(addReq.Password),
			PermissionPassword: utils.EncryptPassword(addReq.PermissionPassword),
			IsDisable:          0,
			IsSuper:            0,
		}
		if addReq.IsDisable != nil {
			staff.IsDisable = *addReq.IsDisable
		}
		staff.Uuid = saasStaff.Uuid

		err = db.Transaction(func(tx *gorm.DB) error {
			// 创建员工
			err := tx.Model(&model.Staff{}).Create(&staff).Error
			if err != nil {
				return err
			}
			// 添加 staff 和角色的关联
			err = repository.NewStaffRepo(tx).UpdateStaffRoles(staff.Uuid, addReq.Roles)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return errors.WithMessage(errors.New("添加员工失败"), err.Error()), exists
		}

		// 在 saas 创建 company_staff 关联关系
		companyStaff := model.CompanyStaff{
			CompanyUuid: currentCompanyUuid,
			Username:    addReq.Username,
			Phone:       addReq.Phone,
			IsSuper:     0,
			IsDisable:   0,
		}
		companyStaff.Uuid = saasStaff.Uuid
		if addReq.IsDisable != nil {
			companyStaff.IsDisable = *addReq.IsDisable
		}
		err = saasDB.Model(&model.CompanyStaff{}).Create(&companyStaff).Error
		if err != nil {
			return errors.WithMessage(errors.New("创建门店关联失败"), err.Error()), exists
		}
	}

	return nil, exists
}

// SaasUpdateStaff 统一账号修改员工
func (s *staffSrv) SaasUpdateStaff(ctx context.Context, updateReq req.UpdateStaffReq) (error, []string) {
	var exists []string
	saasDB := s.dbm.GetDB(constant.DefaultDB)
	saasStaffRepo := repository.NewSaasStaffRepo(saasDB)

	// 查询 saas.ttpos_staff 是否存在该 saas 员工，不存在则报错
	saasStaff, err := saasStaffRepo.GetByUuid(updateReq.Uuid)
	if err != nil || saasStaff == nil || saasStaff.Uuid == 0 {
		return errors.New("员工不存在"), exists
	}

	// 修改 saas.ttpos_staff 的 Email、Phone、RealName
	saasStaffUpdate := map[string]any{
		"email":     updateReq.Username,
		"real_name": updateReq.RealName,
		"phone":     updateReq.Phone,
	}
	if updateReq.Password != "" {
		saasStaffUpdate["password"] = utils.EncryptPassword(updateReq.Password)
		saasStaffUpdate["password_change_time"] = time.Now().Unix()
	}
	if err := saasStaffRepo.Update(updateReq.Uuid, saasStaffUpdate); err != nil {
		return errors.WithMessage(errors.New("编辑员工失败"), err.Error()), exists
	}

	currentCompanyUuid := ctx.GetCompanyUuid()
	companySetting := ctx.GetCompanySetting()

	// 判断当前商家是否总部或者有子级商家
	isHeadquarter := companySetting.IsHeadquarter()
	hasChildren := companySetting.HasChildren == 1

	if isHeadquarter || hasChildren {
		// 如果是总部或者有子公司，走多门店逻辑
		if len(updateReq.CompanyRoleList) == 0 {
			return errors.New("多门店配置不能为空"), exists
		}

		// 获取当前门店可见的所有门店列表
		companyRepo := repository.NewCompanyRepo(saasDB)
		visibleCompanies, err := companyRepo.GetVisibleCompanyList(currentCompanyUuid)
		if err != nil {
			return errors.WithMessage(errors.New("获取可见门店列表失败"), err.Error()), exists
		}

		// 构建可见门店UUID映射
		visibleCompanyMap := make(map[uint64]bool)
		for _, company := range visibleCompanies {
			visibleCompanyMap[company.Uuid] = true
		}

		// NOTE: 如果updateReq.CompanyRoleList中包含当前门店，查询员工在当前门店数据库的信息
		// for _, companyRoleItem := range updateReq.CompanyRoleList {
		// 	if companyRoleItem.CompanyUuid == currentCompanyUuid {
		// 		shopDB := s.dbm.GetDB(companyRoleItem.CompanyUuid)
		// 		staffRepo := repository.NewStaffRepo(shopDB)
		// 		staff, _ := staffRepo.GetStaff(staffRepo.WhereUuid(updateReq.Uuid))
		// 		if staff.Uuid != 0 && staff.CashierOnline != 0 && updateReq.IsDisable != nil && *updateReq.IsDisable == 1 {
		// 			return errors.New("当前人员未交班，请先交班"), exists
		// 		}
		// 	}
		// }

		// 遍历 CompanyRoleList，验证并更新员工
		for _, companyRoleItem := range updateReq.CompanyRoleList {
			// 判断 company_uuid 是否为当前公司可见
			if !visibleCompanyMap[companyRoleItem.CompanyUuid] {
				return errors.New("门店不存在或无权限访问"), exists
			}

			// 查询对应商家数据库的 RoleUuids 是否存在
			shopDB := s.dbm.GetDB(companyRoleItem.CompanyUuid)
			if shopDB == nil {
				return errors.New("获取门店数据库连接失败"), exists
			}

			roleRepo := repository.NewRoleRepo(shopDB)
			roles, err := roleRepo.GetRoleList([]repository.DBOption{roleRepo.WhereUuids(companyRoleItem.RoleUuids)}...)
			if err != nil {
				return errors.WithMessage(errors.New("获取角色失败"), err.Error()), exists
			}
			if len(roles) != len(companyRoleItem.RoleUuids) {
				return errors.New("角色参数错误"), exists
			}

			// 检查员工是否存在于该门店数据库中
			staffRepo := repository.NewStaffRepo(shopDB)
			staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(updateReq.Uuid))
			staffExists := err == nil && staff.Uuid != 0

			// 准备员工更新/创建数据
			staffUpdate := map[string]any{
				"username":  updateReq.Username,
				"real_name": updateReq.RealName,
				"phone":     updateReq.Phone,
			}
			if updateReq.Password != "" {
				staffUpdate["password"] = utils.EncryptPassword(updateReq.Password)
				staffUpdate["password_change_time"] = time.Now().Unix()
			}
			if updateReq.PermissionPassword != "" {
				staffUpdate["permission_password"] = utils.EncryptPassword(updateReq.PermissionPassword)
			}
			// 如果 CompanyRoleList 中存在当前商家uuid，更新 is_disable
			if updateReq.IsDisable != nil && companyRoleItem.CompanyUuid == currentCompanyUuid {
				staffUpdate["is_disable"] = *updateReq.IsDisable
			}

			err = shopDB.Transaction(func(tx *gorm.DB) error {
				txStaffRepo := repository.NewStaffRepo(tx)
				if staffExists {
					// 更新员工信息
					err := txStaffRepo.Update(updateReq.Uuid, staffUpdate)
					if err != nil {
						return err
					}
				} else {
					// 创建员工信息
					newStaff := model.Staff{
						CompanyUuid: companyRoleItem.CompanyUuid,
						Username:    updateReq.Username,
						RealName:    updateReq.RealName,
						Phone:       updateReq.Phone,
						IsDisable:   0,
						IsSuper:     0,
					}
					newStaff.Uuid = updateReq.Uuid
					if updateReq.Password != "" {
						newStaff.Password = utils.EncryptPassword(updateReq.Password)
						newStaff.PasswordChangeTime = time.Now().Unix()
					} else {
						// 如果密码为空，使用 saasStaff 的密码
						newStaff.Password = saasStaff.Password
					}
					if updateReq.PermissionPassword != "" {
						newStaff.PermissionPassword = utils.EncryptPassword(updateReq.PermissionPassword)
					}
					if updateReq.IsDisable != nil && companyRoleItem.CompanyUuid == currentCompanyUuid {
						newStaff.IsDisable = *updateReq.IsDisable
					}
					err := tx.Model(&model.Staff{}).Create(&newStaff).Error
					if err != nil {
						return err
					}
				}
				// 更新员工和角色的关联关系
				err = txStaffRepo.UpdateStaffRoles(updateReq.Uuid, companyRoleItem.RoleUuids)
				if err != nil {
					return err
				}
				return nil
			})
			if err != nil {
				return errors.WithMessage(errors.New("更新员工失败"), err.Error()), exists
			}

			// 检查并更新/创建 saas.ttpos_company_staff 表
			var companyStaff model.CompanyStaff
			err = saasDB.Model(&model.CompanyStaff{}).
				Where("uuid = ? AND company_uuid = ?", updateReq.Uuid, companyRoleItem.CompanyUuid).
				First(&companyStaff).Error

			companyStaffUpdate := map[string]any{
				"username": updateReq.Username,
				"phone":    updateReq.Phone,
			}
			// 如果 CompanyRoleList 中存在当前商家uuid，更新 is_disable
			if updateReq.IsDisable != nil && companyRoleItem.CompanyUuid == currentCompanyUuid {
				companyStaffUpdate["is_disable"] = *updateReq.IsDisable
			}

			if err != nil {
				// 记录不存在，创建新记录
				newCompanyStaff := model.CompanyStaff{
					CompanyUuid: companyRoleItem.CompanyUuid,
					Username:    updateReq.Username,
					Phone:       updateReq.Phone,
					IsSuper:     0,
					IsDisable:   0,
				}
				newCompanyStaff.Uuid = updateReq.Uuid
				if updateReq.IsDisable != nil && companyRoleItem.CompanyUuid == currentCompanyUuid {
					newCompanyStaff.IsDisable = *updateReq.IsDisable
				}
				err = saasDB.Model(&model.CompanyStaff{}).Create(&newCompanyStaff).Error
				if err != nil {
					return errors.WithMessage(errors.New("创建门店关联失败"), err.Error()), exists
				}
			} else {
				// 记录存在，更新记录
				err = saasDB.Model(&model.CompanyStaff{}).
					Where("uuid = ? AND company_uuid = ?", updateReq.Uuid, companyRoleItem.CompanyUuid).
					Updates(companyStaffUpdate).Error
				if err != nil {
					return errors.WithMessage(errors.New("更新门店关联失败"), err.Error()), exists
				}
			}
		}
	} else {
		// 如果是子店，判断参数中的 roles，修改当前商家数据库中的 ttpos_staff
		if len(updateReq.Roles) == 0 {
			return errors.New("角色不能为空"), exists
		}

		// 检查 Roles 参数，是否是当前商家数据库的角色
		db := s.dbm.GetDB(currentCompanyUuid)
		roleRepo := repository.NewRoleRepo(db)
		roles, err := roleRepo.GetRoleList([]repository.DBOption{roleRepo.WhereUuids(updateReq.Roles)}...)
		if err != nil {
			return errors.WithMessage(errors.New("获取角色失败"), err.Error()), exists
		}
		if len(roles) != len(updateReq.Roles) {
			return errors.New("角色参数错误"), exists
		}

		// 检查员工是否存在于当前门店数据库中
		staffRepo := repository.NewStaffRepo(db)
		staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(updateReq.Uuid))
		staffExists := err == nil && staff.Uuid != 0

		// 准备员工更新/创建数据
		staffUpdate := map[string]any{
			"username":  updateReq.Username,
			"real_name": updateReq.RealName,
			"phone":     updateReq.Phone,
		}
		if updateReq.Password != "" {
			staffUpdate["password"] = utils.EncryptPassword(updateReq.Password)
			staffUpdate["password_change_time"] = time.Now().Unix()
		}
		if updateReq.PermissionPassword != "" {
			staffUpdate["permission_password"] = utils.EncryptPassword(updateReq.PermissionPassword)
		}
		// 如果是子店，更新 is_disable
		if updateReq.IsDisable != nil {
			staffUpdate["is_disable"] = *updateReq.IsDisable
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			txStaffRepo := repository.NewStaffRepo(tx)
			if staffExists {
				// 更新员工信息
				err := txStaffRepo.Update(updateReq.Uuid, staffUpdate)
				if err != nil {
					return err
				}
			} else {
				// 创建员工信息
				newStaff := model.Staff{
					CompanyUuid: currentCompanyUuid,
					Username:    updateReq.Username,
					RealName:    updateReq.RealName,
					Phone:       updateReq.Phone,
					IsDisable:   0,
					IsSuper:     0,
				}
				newStaff.Uuid = updateReq.Uuid
				if updateReq.Password != "" {
					newStaff.Password = utils.EncryptPassword(updateReq.Password)
					newStaff.PasswordChangeTime = time.Now().Unix()
				} else {
					// 如果密码为空，使用 saasStaff 的密码
					newStaff.Password = saasStaff.Password
				}
				if updateReq.PermissionPassword != "" {
					newStaff.PermissionPassword = utils.EncryptPassword(updateReq.PermissionPassword)
				}
				if updateReq.IsDisable != nil {
					newStaff.IsDisable = *updateReq.IsDisable
				}
				err := tx.Model(&model.Staff{}).Create(&newStaff).Error
				if err != nil {
					return err
				}
			}
			// 更新员工和角色的关联关系
			err = txStaffRepo.UpdateStaffRoles(updateReq.Uuid, updateReq.Roles)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return errors.WithMessage(errors.New("更新员工失败"), err.Error()), exists
		}

		// 检查并更新/创建 saas.ttpos_company_staff 表
		var companyStaff model.CompanyStaff
		err = saasDB.Model(&model.CompanyStaff{}).
			Where("uuid = ? AND company_uuid = ?", updateReq.Uuid, currentCompanyUuid).
			First(&companyStaff).Error

		companyStaffUpdate := map[string]any{
			"username": updateReq.Username,
			"phone":    updateReq.Phone,
		}
		// 如果是子店，更新 is_disable
		if updateReq.IsDisable != nil {
			companyStaffUpdate["is_disable"] = *updateReq.IsDisable
		}

		if err != nil {
			// 记录不存在，创建新记录
			newCompanyStaff := model.CompanyStaff{
				CompanyUuid: currentCompanyUuid,
				Username:    updateReq.Username,
				Phone:       updateReq.Phone,
				IsSuper:     0,
				IsDisable:   0,
			}
			newCompanyStaff.Uuid = updateReq.Uuid
			if updateReq.IsDisable != nil {
				newCompanyStaff.IsDisable = *updateReq.IsDisable
			}
			err = saasDB.Model(&model.CompanyStaff{}).Create(&newCompanyStaff).Error
			if err != nil {
				return errors.WithMessage(errors.New("创建门店关联失败"), err.Error()), exists
			}
		} else {
			// 记录存在，更新记录
			err = saasDB.Model(&model.CompanyStaff{}).
				Where("uuid = ? AND company_uuid = ?", updateReq.Uuid, currentCompanyUuid).
				Updates(companyStaffUpdate).Error
			if err != nil {
				return errors.WithMessage(errors.New("更新门店关联失败"), err.Error()), exists
			}
		}
	}

	// 处理 RemoveCompanyList：移除多门店角色配置
	if len(updateReq.RemoveCompanyList) > 0 {
		// 获取当前门店可见的所有门店列表（用于验证权限）
		companyRepo := repository.NewCompanyRepo(saasDB)
		visibleCompanies, err := companyRepo.GetVisibleCompanyList(currentCompanyUuid)
		if err != nil {
			return errors.WithMessage(errors.New("获取可见门店列表失败"), err.Error()), exists
		}

		// 构建可见门店UUID映射
		visibleCompanyMap := make(map[uint64]bool)
		for _, company := range visibleCompanies {
			visibleCompanyMap[company.Uuid] = true
		}

		// 遍历 RemoveCompanyList，软删除关联关系
		for _, companyUuid := range updateReq.RemoveCompanyList {
			// 验证门店是否为当前公司可见
			if !visibleCompanyMap[companyUuid] {
				return errors.New("门店不存在或无权限访问"), exists
			}

			// 软删除 saas.ttpos_company_staff 中的关联关系
			err = saasDB.Model(&model.CompanyStaff{}).
				Where("uuid = ? AND company_uuid = ?", updateReq.Uuid, companyUuid).
				Update("delete_time", time.Now().Unix()).Error
			if err != nil {
				return errors.WithMessage(errors.New("删除门店关联失败"), err.Error()), exists
			}

			// 软删除对应商家数据库中的 ttpos_staff
			shopDB := s.dbm.GetDB(companyUuid)
			if shopDB != nil {
				err = shopDB.Model(&model.Staff{}).
					Where("uuid = ?", updateReq.Uuid).
					Update("delete_time", time.Now().Unix()).Error
				if err != nil {
					return errors.WithMessage(errors.New("删除员工失败"), err.Error()), exists
				}
			}
		}
	}

	// 删除收银机缓存
	tc := cache.NewTaggedCache(s.cache)
	tc.TagClear("cashier")

	// 推送配置更新
	utils.Go(func() {
		websocket.PushClient(ctx.GetCompanyUuid(), websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_PERMISSION, map[string]any{
			"staff_uuid":  updateReq.Uuid,
			"update_time": time.Now().Unix(),
		})
	})

	return nil, exists
}
