package inventory

// // NewWarehouseServiceAdapter 创建仓库服务（DDD架构）
// // 此函数创建完整的DDD服务栈并返回适配器
// func NewWarehouseServiceAdapter(
// 	dbm *database.DBManager,
// 	settingSrv setting.ISrv,
// 	materialSrv service.IMaterialSrv,
// 	translateSrv service.ITranslateSrv,
// ) service.IWarehouseSrv {
// 	// 创建基础设施层依赖
// 	warehouseRepo := infraPersistence.NewWarehouseRepository(dbm)
// 	multiLanguageNameRepo := infraPersistence.NewMultiLanguageNameRepository(dbm)
// 	erpService := integration.NewErpWarehouseService(dbm)

// 	// 创建领域服务
// 	domainServiceInstance := domainService.NewWarehouseDomainService(warehouseRepo)

// 	// 创建应用服务
// 	checkNameSrv := service.NewCheckNameSrv(dbm)
// 	appService := NewWarehouseAppService(
// 		warehouseRepo,
// 		multiLanguageNameRepo,
// 		domainServiceInstance,
// 		erpService,
// 		dbm,
// 		settingSrv,
// 		translateSrv,
// 		checkNameSrv,
// 	)

// 	// 返回适配器
// 	return &WarehouseServiceAdapter{
// 		appService: appService,
// 	}
// }

// // WarehouseServiceAdapter 仓库服务适配器
// // 实现旧的 IWarehouseSrv 接口，委托给新的应用服务
// // @Deprecated 此适配器用于平滑过渡到DDD架构，后续将逐步废弃
// type WarehouseServiceAdapter struct {
// 	appService *WarehouseAppService
// }

// // 确保实现了 IWarehouseSrv 接口
// var _ service.IWarehouseSrv = (*WarehouseServiceAdapter)(nil)

// // GetWarehouseList 获取仓库列表
// func (a *WarehouseServiceAdapter) GetWarehouseList(ctx context.Context, req req.WarehouseListReq, isHeadquarters ...bool) (resp.WarehouseListResp, error) {
// 	return a.appService.GetWarehouseList(ctx, req, isHeadquarters...)
// }

// // GetHeadquarterWarehouseList 获取总部仓库列表
// func (a *WarehouseServiceAdapter) GetHeadquarterWarehouseList(ctx context.Context) (resp.WarehouseListResp, error) {
// 	statusActive := 1
// 	return a.appService.GetWarehouseList(ctx, req.WarehouseListReq{
// 		PageReq: dto.PageReq{
// 			PageNo:   1,
// 			PageSize: 999,
// 		},
// 		Type:   "normal",
// 		Status: &statusActive,
// 	}, true)
// }

// // CreateWarehouse 创建仓库
// func (a *WarehouseServiceAdapter) CreateWarehouse(ctx context.Context, addReq req.CreateWarehouseReq) error {
// 	return a.appService.CreateWarehouse(ctx, addReq)
// }

// // UpdateWarehouse 更新仓库
// func (a *WarehouseServiceAdapter) UpdateWarehouse(ctx context.Context, updateReq req.UpdateWarehouseReq) error {
// 	return a.appService.UpdateWarehouse(ctx, updateReq)
// }

// // DeleteWarehouse 删除仓库
// func (a *WarehouseServiceAdapter) DeleteWarehouse(ctx context.Context, deleteReq req.DeleteWarehouseReq) error {
// 	return a.appService.DeleteWarehouse(ctx, deleteReq)
// }

// // SetDefaultWarehouse 设置默认仓库
// func (a *WarehouseServiceAdapter) SetDefaultWarehouse(ctx context.Context, setReq req.SetDefaultWarehouseReq) error {
// 	return a.appService.SetDefaultWarehouse(ctx, setReq)
// }

// // GetWarehouse 获取仓库详情
// func (a *WarehouseServiceAdapter) GetWarehouse(ctx context.Context, getReq req.WarehouseReq) (resp.WarehouseResp, error) {
// 	return a.appService.GetWarehouse(ctx, getReq)
// }

// // GetWarehouseInOutList 获取仓库出入库明细列表
// // 注意：此方法暂未迁移到DDD，仍使用旧实现
// func (a *WarehouseServiceAdapter) GetWarehouseInOutList(ctx context.Context, req req.GetWarehouseInOutListReq) (resp.WarehouseInOutListResp, error) {
// 	// TODO: 此功能属于出入库日志领域，暂不实现
// 	// 当前返回空列表，待后续迁移出入库日志聚合时实现
// 	return resp.WarehouseInOutListResp{
// 		List: make([]resp.WarehouseInOutResp, 0),
// 		Meta: dto.PageResponse{
// 			PageNo:   req.PageNo,
// 			PageSize: req.PageSize,
// 			Total:    0,
// 		},
// 	}, nil
// }

// // CheckCodeExists 检查仓库编码是否存在
// func (a *WarehouseServiceAdapter) CheckCodeExists(ctx context.Context, req req.CheckCodeExistsReq) (resp.CheckNameCodeExistsResp, error) {
// 	return a.appService.CheckCodeExists(ctx, req)
// }

// // GetOtherOrgList 获取对方机构列表
// func (a *WarehouseServiceAdapter) GetOtherOrgList(ctx context.Context) (resp.OtherOrgListResp, error) {
// 	// TODO: 此功能属于出入库日志领域，待后续实现
// 	return resp.OtherOrgListResp{List: make([]resp.OtherOrgResp, 0)}, nil
// }

// // GetWarehouseMaterialList 获取仓库物品列表
// func (a *WarehouseServiceAdapter) GetWarehouseMaterialList(ctx context.Context, req req.WarehouseMaterialListReq) (resp.WarehouseMaterialListResp, error) {
// 	// TODO: 此功能属于库存物品聚合，待后续实现
// 	return resp.WarehouseMaterialListResp{
// 		List: make([]resp.WarehouseMaterialInfo, 0),
// 		Meta: dto.PageResponse{
// 			PageNo:   req.PageNo,
// 			PageSize: req.PageSize,
// 			Total:    0,
// 		},
// 	}, nil
// }

// // SyncWarehouse 同步仓库列表
// func (a *WarehouseServiceAdapter) SyncWarehouse(ctx context.Context) error {
// 	// TODO: 此功能待实现
// 	return nil
// }

// // SyncWarehouseItemStock 同步仓库物品库存
// func (a *WarehouseServiceAdapter) SyncWarehouseItemStock(ctx context.Context) error {
// 	// TODO: 此功能属于库存物品聚合，待后续实现
// 	return nil
// }
