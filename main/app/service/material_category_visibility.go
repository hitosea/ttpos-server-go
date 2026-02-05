package service

import (
	"sync"

	"ttpos-server-go/app/constant"
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
)

// IMaterialCategoryVisibilitySrv 物品分类可见性配置服务接口
type IMaterialCategoryVisibilitySrv interface {
	// 配置管理
	GetList(ctx context.Context) (*resp.VisibilityListResp, error)
	GetDetail(ctx context.Context, uuid uint64) (*resp.VisibilityDetailResp, error)
	Create(ctx context.Context, req req.VisibilityCreateReq) error
	Update(ctx context.Context, req req.VisibilityUpdateReq) error
	Delete(ctx context.Context, uuid uint64) error

	// 子店角色查询（跨库）
	GetSubShopRoles(ctx context.Context) (*resp.SubShopRolesResp, error)

	// 可见性过滤（核心方法）
	// 返回值: categoryUuids, needFilter
	//   - needFilter=false: 用户可见所有类别，无需过滤
	//   - needFilter=true:  用户仅可见返回的类别列表（白名单模式）
	GetVisibleCategoryUuids(ctx context.Context, staffRoleUuids []uint64, companyUuid uint64) ([]uint64, bool)

	// GetVisibleCategoryUuidsForStaff 获取当前登录员工可见的物品类别UUID列表（便捷方法）
	GetVisibleCategoryUuidsForStaff(ctx context.Context) ([]uint64, bool)
}

// materialCategoryVisibilitySrv 物品分类可见性配置服务实现
type materialCategoryVisibilitySrv struct {
	dbm *database.DBManager
}

// NewMaterialCategoryVisibilitySrv 创建物品分类可见性配置服务
func NewMaterialCategoryVisibilitySrv(dbm *database.DBManager) IMaterialCategoryVisibilitySrv {
	return &materialCategoryVisibilitySrv{dbm: dbm}
}

// GetList 获取可见性配置列表
func (s *materialCategoryVisibilitySrv) GetList(ctx context.Context) (*resp.VisibilityListResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	repo := repository.NewMaterialCategoryVisibilityRepo(db)

	list, err := repo.GetList()
	if err != nil {
		return nil, errors.WithMessage(err, "获取可见性配置列表失败")
	}

	result := &resp.VisibilityListResp{
		List: make([]resp.VisibilityListItem, 0, len(list)),
	}

	for _, item := range list {
		result.List = append(result.List, resp.VisibilityListItem{
			Uuid:          item.Uuid,
			LocaleName:    model.NewMultiLanguageName(item.Name).GetNames(),
			IsAllRoles:    item.IsAllRoles,
			RoleCount:     item.GetRoleCount(),
			CategoryCount: item.GetCategoryCount(),
		})
	}

	return result, nil
}

// GetDetail 获取可见性配置详情
func (s *materialCategoryVisibilitySrv) GetDetail(ctx context.Context, uuid uint64) (*resp.VisibilityDetailResp, error) {
	db := s.dbm.GetDB(ctx.GetDbId())
	repo := repository.NewMaterialCategoryVisibilityRepo(db)

	visibility, err := repo.GetByUuid(uuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取可见性配置详情失败")
	}

	roleConfigs := visibility.GetRoleConfigList()
	respRoleConfigs := make([]resp.VisibilityRoleConfig, 0, len(roleConfigs))
	for _, rc := range roleConfigs {
		respRoleConfigs = append(respRoleConfigs, resp.VisibilityRoleConfig{
			RoleUuid:    rc.RoleUuid,
			CompanyUuid: rc.CompanyUuid,
		})
	}

	return &resp.VisibilityDetailResp{
		Uuid:          visibility.Uuid,
		LocaleName:    model.NewMultiLanguageName(visibility.Name).GetNames(),
		IsAllRoles:    visibility.IsAllRoles,
		CategoryUuids: visibility.GetCategoryUuidList(),
		RoleConfigs:   respRoleConfigs,
	}, nil
}

// Create 创建可见性配置
func (s *materialCategoryVisibilitySrv) Create(ctx context.Context, r req.VisibilityCreateReq) error {
	// 验证参数
	if err := r.Validate(); err != nil {
		return err
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	companySetting := ctx.GetCompanySetting()

	// 只有总店可以创建
	if !companySetting.IsHeadquarter() {
		return errors.New("只有总店可以创建可见性配置")
	}

	// 转换 RoleConfigs
	roleConfigs := make([]model.RoleConfig, 0, len(r.RoleConfigs))
	for _, rc := range r.RoleConfigs {
		roleConfigs = append(roleConfigs, model.RoleConfig{
			RoleUuid:    rc.RoleUuid,
			CompanyUuid: rc.CompanyUuid,
		})
	}

	// 创建配置
	visibility := &model.MaterialCategoryVisibility{
		Name:       r.LocaleName.ToJson(),
		IsAllRoles: r.IsAllRoles,
	}
	visibility.Uuid, _ = utils.GetID()
	visibility.SetCategoryUuidList(r.CategoryUuids)
	visibility.SetRoleConfigList(roleConfigs)

	repo := repository.NewMaterialCategoryVisibilityRepo(db)
	if err := repo.Create(visibility); err != nil {
		return errors.WithMessage(err, "创建可见性配置失败")
	}

	// 异步同步到子店
	s.asyncSyncToSubShops(ctx, visibility)

	return nil
}

// Update 更新可见性配置
func (s *materialCategoryVisibilitySrv) Update(ctx context.Context, r req.VisibilityUpdateReq) error {
	// 验证参数
	if err := r.Validate(); err != nil {
		return err
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	companySetting := ctx.GetCompanySetting()

	// 只有总店可以更新
	if !companySetting.IsHeadquarter() {
		return errors.New("只有总店可以更新可见性配置")
	}

	repo := repository.NewMaterialCategoryVisibilityRepo(db)

	// 检查配置是否存在
	visibility, err := repo.GetByUuid(r.Uuid)
	if err != nil {
		return errors.WithMessage(err, "可见性配置不存在")
	}

	// 转换 RoleConfigs
	roleConfigs := make([]model.RoleConfig, 0, len(r.RoleConfigs))
	for _, rc := range r.RoleConfigs {
		roleConfigs = append(roleConfigs, model.RoleConfig{
			RoleUuid:    rc.RoleUuid,
			CompanyUuid: rc.CompanyUuid,
		})
	}

	// 更新配置
	visibility.Name = r.LocaleName.ToJson()
	visibility.IsAllRoles = r.IsAllRoles
	visibility.SetCategoryUuidList(r.CategoryUuids)
	visibility.SetRoleConfigList(roleConfigs)

	if err := repo.Update(visibility); err != nil {
		return errors.WithMessage(err, "更新可见性配置失败")
	}

	// 异步同步到子店
	s.asyncSyncToSubShops(ctx, visibility)

	return nil
}

// Delete 删除可见性配置
func (s *materialCategoryVisibilitySrv) Delete(ctx context.Context, uuid uint64) error {
	db := s.dbm.GetDB(ctx.GetDbId())
	companySetting := ctx.GetCompanySetting()

	// 只有总店可以删除
	if !companySetting.IsHeadquarter() {
		return errors.New("只有总店可以删除可见性配置")
	}

	repo := repository.NewMaterialCategoryVisibilityRepo(db)

	// 检查配置是否存在
	_, err := repo.GetByUuid(uuid)
	if err != nil {
		return errors.WithMessage(err, "可见性配置不存在")
	}

	if err := repo.Delete(uuid); err != nil {
		return errors.WithMessage(err, "删除可见性配置失败")
	}

	// 异步同步删除到子店
	s.asyncSyncDeleteToSubShops(ctx, uuid)

	return nil
}

// GetSubShopRoles 获取子店角色列表（跨库查询），不包括总店
func (s *materialCategoryVisibilitySrv) GetSubShopRoles(ctx context.Context) (*resp.SubShopRolesResp, error) {
	companySetting := ctx.GetCompanySetting()

	// 只有总店可以查询
	if !companySetting.IsHeadquarter() {
		return nil, errors.New("只有总店可以查询子店角色")
	}

	headquarterUuid := ctx.GetDbId()

	// 获取所有子店（包含总店，后续过滤）
	saasDb := s.dbm.GetDB(constant.DefaultDB)
	companyRepo := repository.NewCompanyRepo(saasDb)
	subShops, err := companyRepo.GetAllSubShopsAndHeadquarterListByCompanyUuid(headquarterUuid)
	if err != nil {
		return nil, errors.WithMessage(err, "获取子店列表失败")
	}

	result := &resp.SubShopRolesResp{
		Shops: make([]resp.ShopRoles, 0, len(subShops)),
	}

	// 并发查询各子店角色
	var wg sync.WaitGroup
	var mu sync.Mutex
	shopRolesChan := make(chan resp.ShopRoles, len(subShops))

	for _, shop := range subShops {
		// 跳过总店，只查询子店
		if shop.Uuid == headquarterUuid {
			continue
		}
		wg.Add(1)
		utils.Go(func() {
			defer wg.Done()

			shopDb := s.dbm.GetDB(shop.Uuid)
			roleRepo := repository.NewRoleRepo(shopDb)
			roles, err := roleRepo.GetRoleList()
			if err != nil {
				logger.Logger.Error("获取子店角色失败",
					zap.Uint64("company_uuid", shop.Uuid),
					zap.Error(err),
				)
				return
			}

			roleItems := make([]resp.RoleItem, 0, len(roles))
			for _, role := range roles {
				roleItems = append(roleItems, resp.RoleItem{
					Uuid: role.Uuid,
					Name: role.Name,
				})
			}

			shopRolesChan <- resp.ShopRoles{
				CompanyUuid: shop.Uuid,
				CompanyCode: "", // Company model 没有 Code 字段，后续可从 CompanySetting 获取
				CompanyName: shop.Name,
				Roles:       roleItems,
			}
		})
	}

	// 等待所有查询完成
	go func() {
		wg.Wait()
		close(shopRolesChan)
	}()

	// 收集结果
	for shopRoles := range shopRolesChan {
		mu.Lock()
		result.Shops = append(result.Shops, shopRoles)
		mu.Unlock()
	}

	return result, nil
}

// GetVisibleCategoryUuidsForStaff 获取当前登录员工可见的物品类别UUID列表
// 返回值: categoryUuids, needFilter
//   - needFilter=false: 用户可见所有类别，无需过滤
//   - needFilter=true:  用户仅可见返回的类别列表（白名单模式）
func (s *materialCategoryVisibilitySrv) GetVisibleCategoryUuidsForStaff(ctx context.Context) ([]uint64, bool) {
	dbId := ctx.GetDbId()
	staffUuid := ctx.GetStaffUuid()

	// 如果没有员工信息，默认全部可见
	if staffUuid == 0 {
		return nil, false
	}

	// 获取员工角色
	db := s.dbm.GetDB(dbId)
	staffRepo := repository.NewStaffRepo(db)
	staff, err := staffRepo.GetStaff(staffRepo.WhereUuid(staffUuid), staffRepo.WithRoles())
	if err != nil {
		logger.Logger.Error("获取员工角色失败",
			zap.Uint64("company_uuid", dbId),
			zap.Uint64("staff_uuid", staffUuid),
			zap.Error(err),
		)
		return nil, false // 出错时默认全部可见
	}

	// 提取角色UUID列表
	staffRoleUuids := make([]uint64, 0, len(staff.Roles))
	for _, role := range staff.Roles {
		staffRoleUuids = append(staffRoleUuids, role.Uuid)
	}

	return s.GetVisibleCategoryUuids(ctx, staffRoleUuids, dbId)
}

// GetVisibleCategoryUuids 获取用户可见的物品类别UUID列表
func (s *materialCategoryVisibilitySrv) GetVisibleCategoryUuids(ctx context.Context, staffRoleUuids []uint64, companyUuid uint64) ([]uint64, bool) {
	db := s.dbm.GetDB(ctx.GetDbId())
	repo := repository.NewMaterialCategoryVisibilityRepo(db)

	// 获取所有可见性配置
	visibilities, err := repo.GetList()
	if err != nil {
		logger.Logger.Error("获取可见性配置列表失败",
			zap.Uint64("company_uuid", companyUuid),
			zap.Error(err),
		)
		return nil, false // 出错时默认全部可见
	}

	// 如果没有任何配置，全部可见
	if len(visibilities) == 0 {
		return nil, false
	}

	// 检查用户角色是否在任何配置中
	matchedVisibilities := make([]*model.MaterialCategoryVisibility, 0)
	for i := range visibilities {
		visibility := &visibilities[i]
		for _, roleUuid := range staffRoleUuids {
			if visibility.ContainsRole(roleUuid, companyUuid) {
				matchedVisibilities = append(matchedVisibilities, visibility)
				break
			}
		}
	}

	// 如果用户所有角色都不在任何配置中，全部可见
	if len(matchedVisibilities) == 0 {
		return nil, false
	}

	// 收集可见类别并集
	categoryUuidSet := make(map[uint64]struct{})
	for _, visibility := range matchedVisibilities {
		for _, categoryUuid := range visibility.GetCategoryUuidList() {
			categoryUuidSet[categoryUuid] = struct{}{}
		}
	}

	// 转换为切片
	categoryUuids := make([]uint64, 0, len(categoryUuidSet))
	for uuid := range categoryUuidSet {
		categoryUuids = append(categoryUuids, uuid)
	}

	return categoryUuids, true
}

// asyncSyncToSubShops 异步同步配置到子店
func (s *materialCategoryVisibilitySrv) asyncSyncToSubShops(ctx context.Context, visibility *model.MaterialCategoryVisibility) {
	headquarterUuid := ctx.GetDbId()

	utils.Go(func() {
		// 获取所有子店
		saasDb := s.dbm.GetDB(constant.DefaultDB)
		companyRepo := repository.NewCompanyRepo(saasDb)
		subShops, err := companyRepo.GetAllSubShopsAndHeadquarterListByCompanyUuid(headquarterUuid)
		if err != nil {
			logger.Logger.Error("同步可见性配置-获取子店列表失败",
				zap.Uint64("company_uuid", headquarterUuid),
				zap.Uint64("visibility_uuid", visibility.Uuid),
				zap.Error(err),
			)
			return
		}

		// 遍历子店同步
		for _, shop := range subShops {
			subDb := s.dbm.GetDB(shop.Uuid)
			subRepo := repository.NewMaterialCategoryVisibilityRepo(subDb)

			// 创建子店配置（设置 HeadquarterUuid）
			subVisibility := &model.MaterialCategoryVisibility{
				Name:            visibility.Name,
				IsAllRoles:      visibility.IsAllRoles,
				CategoryUuids:   visibility.CategoryUuids,
				RoleConfigs:     visibility.RoleConfigs,
				HeadquarterUuid: headquarterUuid,
			}
			subVisibility.Uuid = visibility.Uuid

			// 尝试更新，如果不存在则创建
			existingVisibility, err := subRepo.GetByUuid(visibility.Uuid)
			if err != nil {
				// 不存在，创建
				if err := subRepo.Create(subVisibility); err != nil {
					logger.Logger.Error("同步可见性配置-创建子店配置失败",
						zap.Uint64("company_uuid", shop.Uuid),
						zap.Uint64("visibility_uuid", visibility.Uuid),
						zap.Error(err),
					)
				}
			} else {
				// 存在，更新
				existingVisibility.Name = visibility.Name
				existingVisibility.IsAllRoles = visibility.IsAllRoles
				existingVisibility.CategoryUuids = visibility.CategoryUuids
				existingVisibility.RoleConfigs = visibility.RoleConfigs
				if err := subRepo.Update(existingVisibility); err != nil {
					logger.Logger.Error("同步可见性配置-更新子店配置失败",
						zap.Uint64("company_uuid", shop.Uuid),
						zap.Uint64("visibility_uuid", visibility.Uuid),
						zap.Error(err),
					)
				}
			}
		}

		logger.Logger.Info("同步可见性配置完成",
			zap.Uint64("company_uuid", headquarterUuid),
			zap.Uint64("visibility_uuid", visibility.Uuid),
			zap.Int("sub_shop_count", len(subShops)),
		)
	})
}

// asyncSyncDeleteToSubShops 异步同步删除到子店
func (s *materialCategoryVisibilitySrv) asyncSyncDeleteToSubShops(ctx context.Context, uuid uint64) {
	headquarterUuid := ctx.GetDbId()

	utils.Go(func() {
		// 获取所有子店
		saasDb := s.dbm.GetDB(constant.DefaultDB)
		companyRepo := repository.NewCompanyRepo(saasDb)
		subShops, err := companyRepo.GetAllSubShopsAndHeadquarterListByCompanyUuid(headquarterUuid)
		if err != nil {
			logger.Logger.Error("同步删除可见性配置-获取子店列表失败",
				zap.Uint64("company_uuid", headquarterUuid),
				zap.Uint64("visibility_uuid", uuid),
				zap.Error(err),
			)
			return
		}

		// 遍历子店同步删除
		for _, shop := range subShops {
			subDb := s.dbm.GetDB(shop.Uuid)
			subRepo := repository.NewMaterialCategoryVisibilityRepo(subDb)

			if err := subRepo.Delete(uuid); err != nil {
				logger.Logger.Error("同步删除可见性配置-删除子店配置失败",
					zap.Uint64("company_uuid", shop.Uuid),
					zap.Uint64("visibility_uuid", uuid),
					zap.Error(err),
				)
			}
		}

		logger.Logger.Info("同步删除可见性配置完成",
			zap.Uint64("company_uuid", headquarterUuid),
			zap.Uint64("visibility_uuid", uuid),
			zap.Int("sub_shop_count", len(subShops)),
		)
	})
}
