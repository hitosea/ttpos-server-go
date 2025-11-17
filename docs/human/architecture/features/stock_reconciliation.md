# 盘点单服务 (Stock Reconciliation Service)

## 概述

`stock_reconciliation.go` 实现了库存盘点单管理服务，负责处理餐饮系统中的库存盘点业务。该服务涵盖了盘点单的完整生命周期管理，包括创建、编辑、提交、审核、驳回、删除等操作，并与 ERP 系统（ERPNext）深度集成，实现库存数据的实时同步和管理。

**文件路径**: `ttpos-server-go/main/app/service/stock_reconciliation.go`

## 核心功能

### 1. 盘点单生命周期管理
- **创建**: 创建新的盘点单，生成唯一单据编号
- **保存**: 保存盘点数据（草稿状态）
- **提交**: 提交到 ERP 系统
- **审核**: 审核通过后更新实际库存
- **驳回**: 驳回已提交的盘点单
- **删除**: 删除草稿状态的盘点单

### 2. 盘点数据管理
- 支持多单位盘点（同一物品可用不同单位盘点）
- 自动计算盘盈盘亏
- 账面库存与实盘数量对比
- 盘点异常检测（差值超过20%）

### 3. ERP 集成
- 与 ERPNext 系统双向同步
- 盘点单提交到 ERP
- ERP 审核结果回写
- 错误信息本地化处理

### 4. 库存更新
- 审核后自动更新仓库物品库存
- 生成盘盈/盘亏出入库日志
- 触发商品库存重新计算

## 接口定义

### IStockReconciliationSrv 接口

```go
type IStockReconciliationSrv interface {
    GetStockReconciliationList(ctx context.Context, req req.StockReconciliationListReq) (resp.StockReconciliationListResp, error)
    GetStockReconciliationDetail(ctx context.Context, req req.StockReconciliationDetailReq) (resp.StockReconciliationDetailResp, error)
    SaveStockReconciliation(ctx context.Context, req req.StockReconciliationSaveReq) (uint64, error)
    DeleteStockReconciliation(ctx context.Context, req req.StockReconciliationDeleteReq) error
    ApproveStockReconciliation(ctx context.Context, req req.StockReconciliationApproveReq) ([]dto.LocaleResponse, error)
    RejectStockReconciliation(ctx context.Context, req req.StockReconciliationRejectReq) error
    CheckMaterials(ctx context.Context, req req.StockReconciliationCheckMaterialsReq) (resp.StockReconciliationCheckMaterialsListResp, error)
}
```

### stockReconciliationSrv 结构体

```go
type stockReconciliationSrv struct {
    productSrv IProductSrv        // 商品服务
    dbm        *database.DBManager // 数据库管理器
    lock       lock.Lock           // 分布式锁
}
```

## 依赖项

### 内部依赖
- **repository.StockReconciliationRepo**: 盘点单数据仓库
- **repository.WarehouseRepo**: 仓库数据仓库
- **repository.WarehouseItemRepo**: 仓库物品数据仓库
- **repository.MaterialRepo**: 物品数据仓库
- **IProductSrv**: 商品服务，用于同步商品库存
- **erp.IErpSrv**: ERP 服务，用于与 ERPNext 系统交互

### 外部依赖
- **database.DBManager**: 数据库管理器
- **lock.Lock**: 并发锁，保证单据编号唯一性和操作原子性
- **decimal**: 精确小数计算库
- **copier**: 结构体拷贝库

## 盘点单生命周期

```
         创建
          ↓
    [已保存 Status=0]
          ↓ 提交
    [已提交 Status=1] ←→ [已驳回 Status=3]
          ↓ 审核
    [已审核 Status=2]
```

### 状态说明

| 状态 | 常量 | 说明 | 允许操作 |
|------|------|------|----------|
| 已保存 | StockReconciliationStatusSaved (0) | 草稿状态 | 编辑、删除、提交 |
| 已提交 | StockReconciliationStatusSubmitted (1) | 已提交到 ERP | 审核、驳回 |
| 已审核 | StockReconciliationStatusApproved (2) | 审核通过，库存已更新 | 无（终态） |
| 已驳回 | StockReconciliationStatusRejected (3) | 审核驳回 | 删除 |

## 核心方法详解

### 1. GetStockReconciliationList - 获取盘点单列表

**方法签名**:
```go
func (s *stockReconciliationSrv) GetStockReconciliationList(ctx context.Context, req req.StockReconciliationListReq) (resp.StockReconciliationListResp, error)
```

**功能**: 分页获取盘点单列表，支持多条件筛选。

**请求参数**:
```go
type StockReconciliationListReq struct {
    PageNo          int      // 页码
    PageSize        int      // 每页大小
    WarehouseUuids  []uint64 // 仓库UUID列表
    Keyword         string   // 关键字（单据编号、ERP单号）
    StartCreateTime int      // 开始创建时间
    EndCreateTime   int      // 结束创建时间
    StatusIn        []int    // 状态列表
}
```

**返回数据**:
```go
type StockReconciliationListResp struct {
    List []StockReconciliationInfo // 盘点单列表
    Meta PageResponse               // 分页信息
}

type StockReconciliationInfo struct {
    Uuid                uint64           // 盘点单UUID
    OrderNo             string           // 单据编号
    Type                int              // 类型
    WarehouseUuid       uint64           // 仓库UUID
    WarehouseLocaleName dto.LocaleResponse // 仓库多语言名称
    Purpose             int              // 盘点目的
    Status              int              // 状态
    ItemsCount          int64            // 物品数量
    SubmitTime          int              // 提交时间
    CreateTime          int              // 创建时间
}
```

**实现流程**:

```65:132:ttpos-server-go/main/app/service/stock_reconciliation.go
func (s *stockReconciliationSrv) GetStockReconciliationList(ctx context.Context, req req.StockReconciliationListReq) (resp.StockReconciliationListResp, error) {
	db := ctx.GetDB()
	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)

	// 构建查询选项
	var opts []repository.DBOption

	// 多仓库筛选
	if len(req.WarehouseUuids) > 0 {
		opts = append(opts, stockReconciliationRepo.WhereWarehouseUuids(req.WarehouseUuids))
	}

	// 关键字搜索（单据编号和ERP盘点单号）
	if req.Keyword != "" {
		opts = append(opts, stockReconciliationRepo.WhereKeyword(req.Keyword))
	}

	// 创建时间范围筛选
	if req.StartCreateTime > 0 || req.EndCreateTime > 0 {
		opts = append(opts, stockReconciliationRepo.WhereCreateTimeRange(req.StartCreateTime, req.EndCreateTime))
	}

	// 状态列表筛选
	if len(req.StatusIn) > 0 {
		opts = append(opts, stockReconciliationRepo.WhereStatusIn(req.StatusIn))
	}

	opts = append(opts, stockReconciliationRepo.WithWarehouseMultiLanguageName())

	// 查询数据
	list, total, err := stockReconciliationRepo.GetStockReconciliationListWithPagination(req.PageNo, req.PageSize, opts...)
	if err != nil {
		return resp.StockReconciliationListResp{}, errors.WithMessage(err, "查询盘点单列表失败")
	}

	// 转换响应数据
	listResp := make([]*resp.StockReconciliationInfo, 0, len(list))

	stockReconciliationUuidList := make([]uint64, 0, len(list))
	for _, item := range list {
		stockReconciliationUuidList = append(stockReconciliationUuidList, item.Uuid)
	}

	// 根据盘点单获取每个盘点单的物品数量，返回map[盘点单UUID]物品数量
	itemsCountMap, err := stockReconciliationRepo.GetStockReconciliationItemCountListByReconciliationUuidList(stockReconciliationUuidList)
	if err != nil {
		return resp.StockReconciliationListResp{}, errors.WithMessage(err, "查询盘点单物品明细失败")
	}

	for _, item := range list {
		info := &resp.StockReconciliationInfo{}
		if err := copier.Copy(info, item); err != nil {
			continue
		}
		info.WarehouseLocaleName = item.Warehouse.MultiLanguageName.GetNames()
		info.ItemsCount = itemsCountMap[item.Uuid]
		listResp = append(listResp, info)
	}

	return resp.StockReconciliationListResp{
		List: listResp,
		Meta: dto.PageResponse{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
			Total:    total,
		},
	}, nil
}
```

**关键点**:
1. 支持多维度筛选（仓库、关键字、时间范围、状态）
2. 使用选项模式构建灵活的查询条件
3. 批量查询物品数量，优化性能
4. 返回多语言仓库名称

---

### 2. GetStockReconciliationDetail - 获取盘点单详情

**方法签名**:
```go
func (s *stockReconciliationSrv) GetStockReconciliationDetail(ctx context.Context, req req.StockReconciliationDetailReq) (resp.StockReconciliationDetailResp, error)
```

**功能**: 获取盘点单的完整详情，包括所有物品明细和单位明细。

**返回数据**:
```go
type StockReconciliationDetailResp struct {
    Uuid          uint64                          // 盘点单UUID
    OrderNo       string                          // 单据编号
    WarehouseName dto.LocaleResponse              // 仓库名称
    Purpose       int                             // 盘点目的
    Status        int                             // 状态
    Items         []*StockReconciliationItemInfo  // 物品明细列表
    // ... 其他字段
}

type StockReconciliationItemInfo struct {
    MaterialUuid               uint64               // 物品UUID
    LocaleName                 dto.LocaleResponse   // 物品名称
    MaterialCode               string               // 物品编码
    InternalCode               string               // 内部编码
    MaterialBarcode            string               // 物品条码
    BookedQuantity             float64              // 账面数量
    CountedQuantity            float64              // 实盘数量
    DiffQuantity               float64              // 差异数量
    InventoryStatus            int                  // 盘点状态（盘盈/盘亏/正常）
    IsInventoryStatusException bool                 // 是否异常
    Units                      []MaterialUnitInfo   // 可用单位列表
    ItemUnits                  []*StockReconciliationItemUnitInfo // 盘点单位明细
}
```

**实现流程**:

```148:261:ttpos-server-go/main/app/service/stock_reconciliation.go
func (s *stockReconciliationSrv) GetStockReconciliationDetail(ctx context.Context, req req.StockReconciliationDetailReq) (resp.StockReconciliationDetailResp, error) {
	db := ctx.GetDB()
	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)
	var detailResp resp.StockReconciliationDetailResp

	opts := []repository.DBOption{
		stockReconciliationRepo.WhereUuid(req.Uuid),
		stockReconciliationRepo.WithWarehouseMultiLanguageName(),
		stockReconciliationRepo.WithStockReconciliationItemMaterialUnits(),
	}
	// 查询盘点单
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliation(opts...)
	if err != nil {
		return detailResp, errors.WithMessage(err, "查询盘点单失败")
	}
	if stockReconciliation == nil {
		return detailResp, errors.New("盘点单不存在")
	}

	// 转换响应数据
	if err := copier.Copy(&detailResp, stockReconciliation); err != nil {
		return detailResp, errors.WithMessage(err, "转换盘点单数据失败")
	}
	detailResp.WarehouseName = stockReconciliation.Warehouse.MultiLanguageName.GetNames()

	bookedQuantityMap, err := s.getBookedQuantityMap(db, stockReconciliation.WarehouseUuid)
	if err != nil {
		return detailResp, errors.WithMessage(errors.New("查询仓库物品失败"), err.Error())
	}

	// 物品单位明细
	itemsResp := make([]*resp.StockReconciliationItemInfo, 0, len(stockReconciliation.StockReconciliationItems))
	for _, item := range stockReconciliation.StockReconciliationItems {
		// 明细中的物品已删除，跳过
		if item.DeleteTime > 0 {
			continue
		}
		itemInfo := &resp.StockReconciliationItemInfo{}
		if err := copier.Copy(itemInfo, item); err != nil {
			continue
		}

		itemInfo.BookedQuantity = item.BookedQuantity.InexactFloat64()
		itemInfo.CountedQuantity = item.CountedQuantity.InexactFloat64()

		// 查询物品信息
		if item.Material != nil {
			itemInfo.LocaleName = *language.JsonToLocaleResponse(item.Material.Name)
			itemInfo.MaterialCode = item.Material.Code
			itemInfo.InternalCode = item.Material.InternalCode
			itemInfo.MaterialBarcode = item.Material.BarcodeValue
		}

		itemInfo.Units = make([]resp.MaterialUnitInfo, 0)
		for _, unit := range item.Material.NotBaseUnitList {
			if unit.Unit != nil {
				itemInfo.Units = append(itemInfo.Units, resp.MaterialUnitInfo{
					MaterialUnitUuid: unit.Uuid,
					UnitUuid:         unit.UnitUuid,
					UnitName:         unit.Unit.MultiLanguageName.GetNames(),
					ConversionRate:   unit.ConversionRate,
					IsDefault:        unit.IsDefault,
				})
			}
		}

		// 查询单位明细
		itemUnits, err := stockReconciliationRepo.GetStockReconciliationItemUnitListByItemUuid(item.Uuid)
		if err == nil && len(itemUnits) > 0 {
			unitsResp := make([]*resp.StockReconciliationItemUnitInfo, 0, len(itemUnits))
			for _, itemUnit := range itemUnits {
				if itemUnit.MaterialUnit == nil || itemUnit.MaterialUnit.Unit == nil || itemUnit.MaterialUnit.Unit.MultiLanguageName.Uuid == 0 ||
					(stockReconciliation.Status == constant.StockReconciliationStatusSubmitted && itemUnit.Quantity == nil) {
					continue
				}
				unitInfo := &resp.StockReconciliationItemUnitInfo{}
				if err := copier.Copy(unitInfo, itemUnit); err != nil {
					continue
				}
				unitInfo.LocaleName = itemUnit.MaterialUnit.Unit.MultiLanguageName.GetNames()
				for _, unit := range itemInfo.Units {
					if unit.MaterialUnitUuid == itemUnit.MaterialUnitUuid {
						unitInfo.ConversionRate = unit.ConversionRate
						break
					}
				}
				unitsResp = append(unitsResp, unitInfo)
			}
			itemInfo.ItemUnits = unitsResp
		}

		bookedQuantity := item.BookedQuantity
		// 已保存状态，账面库存数量要实时读取；其他状态，账面库存数量为盘点单中的数量
		if stockReconciliation.Status == constant.StockReconciliationStatusSaved {
			bookedQuantity = bookedQuantityMap[item.MaterialUuid]
		}
		// 盘盈盘亏状态
		if item.CountedQuantity.GreaterThan(bookedQuantity) {
			itemInfo.InventoryStatus = constant.StockReconciliationInventoryStatusProfit
		} else if item.CountedQuantity.LessThan(bookedQuantity) {
			itemInfo.InventoryStatus = constant.StockReconciliationInventoryStatusLoss
		} else {
			itemInfo.InventoryStatus = constant.StockReconciliationInventoryStatusNormal
		}
		// 是否盘盈盘亏异常（账面和实盘数量差值的绝对值大于20%）
		itemInfo.IsInventoryStatusException = s.getIsInventoryStatusException(bookedQuantity, item.CountedQuantity)
		itemInfo.DiffQuantity = item.CountedQuantity.Sub(bookedQuantity).Truncate(3).InexactFloat64()
		itemsResp = append(itemsResp, itemInfo)
	}
	detailResp.Items = itemsResp

	return detailResp, nil
}
```

**关键点**:
1. **动态账面库存**: 已保存状态时实时读取账面库存，其他状态使用快照值
2. **多单位支持**: 展示物品的所有可用单位和盘点单位明细
3. **盘盈盘亏计算**: 自动判断盘盈、盘亏或正常
4. **异常检测**: 差值超过20%标记为异常
5. **软删除过滤**: 跳过已删除的物品明细

---

### 3. SaveStockReconciliation - 保存盘点单

**方法签名**:
```go
func (s *stockReconciliationSrv) SaveStockReconciliation(ctx context.Context, saveReq req.StockReconciliationSaveReq) (uint64, error)
```

**功能**: 创建新盘点单或更新现有盘点单，支持保存后提交到 ERP。

**请求参数**:
```go
type StockReconciliationSaveReq struct {
    Uuid             uint64                            // 盘点单UUID（0表示新建）
    Type             int                               // 类型
    WarehouseUuid    uint64                            // 仓库UUID
    Purpose          int                               // 盘点目的（1-月度盘点，2-年度盘点）
    IsSubmit         bool                              // 是否提交
    SubmitAfterSave  bool                              // 是否保存后提交
    Items            []StockReconciliationSaveItem     // 物品明细
}

type StockReconciliationSaveItem struct {
    MaterialUuid uint64                                 // 物品UUID
    Units        []StockReconciliationSaveItemUnit     // 单位明细
}

type StockReconciliationSaveItemUnit struct {
    MaterialUnitUuid uint64           // 物品单位UUID
    Quantity         *decimal.Decimal // 数量
}
```

**实现流程**:

```264:439:ttpos-server-go/main/app/service/stock_reconciliation.go
func (s *stockReconciliationSrv) SaveStockReconciliation(ctx context.Context, saveReq req.StockReconciliationSaveReq) (uint64, error) {

	db := ctx.GetDB()
	companySetting := ctx.GetCompanySetting()
	timezone := companySetting.GetTimezone()
	stockReconciliationUuid := saveReq.Uuid
	var stockReconciliation *model.StockReconciliation
	var err error

	if saveReq.Uuid == 0 { // 新建
		// 加锁保证单号唯一性（基于公司UUID和日期）
		dateStr := utils.SetTimezone(timezone).Now().Format("20060102")
		lockKey := fmt.Sprintf("stock_reconciliation_%d_%s", ctx.GetCompanyUuid(), dateStr)
		s.lock.LockUuidString(lockKey)
		defer s.lock.UnlockUuidString(lockKey)
	} else { // 修改
		// 加锁
		s.lock.LockUuid(saveReq.Uuid)
		defer s.lock.UnlockUuid(saveReq.Uuid)

		stockReconciliationRepo := repository.NewStockReconciliationRepo(db)
		// 查询盘点单
		stockReconciliation, err = stockReconciliationRepo.GetStockReconciliationByUuid(saveReq.Uuid)
		if err != nil {
			return stockReconciliationUuid, errors.WithMessage(err, "查询盘点单失败")
		}
		if stockReconciliation == nil {
			return stockReconciliationUuid, errors.New("盘点单不存在")
		}

		// 只有已保存状态的盘点单才能修改
		if stockReconciliation.Status != constant.StockReconciliationStatusSaved {
			if saveReq.IsSubmit {
				return stockReconciliationUuid, errors.New("当前状态不允许提交")
			} else {
				return stockReconciliationUuid, errors.New("当前状态不允许修改")
			}
		}
	}

	// A、在列表上直接提交
	if saveReq.IsSubmit && !saveReq.SubmitAfterSave && saveReq.Uuid > 0 {
		// 直接提交
		return stockReconciliationUuid, s.submitStockReconciliation(ctx, saveReq.Uuid, true)
	}

	// B、在详情中保存或者提交

	// 验证仓库和物品明细
	warehouseItems, materials, err := s.validateWarehouseAndItems(db, saveReq)
	if err != nil {
		return stockReconciliationUuid, err
	}

	bookedQuantityMap := map[uint64]float64{}
	for _, warehouseItem := range warehouseItems {
		bookedQuantityMap[warehouseItem.MaterialUuid] = warehouseItem.Stock
	}

	materialUnitMap := make(map[uint64]map[uint64]float64)
	for _, material := range materials {
		materialUnitMap[material.Uuid] = make(map[uint64]float64)
		for _, materialUnit := range material.NotBaseUnitList {
			materialUnitMap[material.Uuid][materialUnit.Uuid] = materialUnit.ConversionRate
		}
	}

	// 开启事务
	err = db.Transaction(func(tx *gorm.DB) error {
		stockReconciliationRepo := repository.NewStockReconciliationRepo(tx)

		if saveReq.Uuid == 0 { // 新建
			// 在事务内部生成单据编号
			orderNo := s.generateOrderNo(tx, timezone)
			// 创建盘点单
			stockReconciliation = &model.StockReconciliation{
				OrderNo:       orderNo,
				Type:          saveReq.Type,
				WarehouseUuid: saveReq.WarehouseUuid,
				Purpose:       saveReq.Purpose,
				Status:        constant.StockReconciliationStatusSaved, // 0-已保存
			}
			if err := stockReconciliationRepo.CreateStockReconciliation(stockReconciliation); err != nil {
				return errors.WithMessage(errors.New("创建盘点单失败"), err.Error())
			}

			stockReconciliationUuid = stockReconciliation.Uuid
		} else { // 更新
			// 更新盘点单
			stockReconciliation.WarehouseUuid = saveReq.WarehouseUuid
			stockReconciliation.Purpose = saveReq.Purpose
			stockReconciliation.Type = saveReq.Type

			if err := stockReconciliationRepo.UpdateStockReconciliation(stockReconciliation); err != nil {
				return errors.WithMessage(err, "更新盘点单失败")
			}
			// 删除原有的物品明细
			if err := stockReconciliationRepo.DeleteStockReconciliationItemByReconciliationUuid(saveReq.Uuid); err != nil {
				return errors.WithMessage(err, "删除盘点单物品明细失败")
			}
			// 删除原有物品单位明细
			if err := stockReconciliationRepo.DeleteStockReconciliationItemUnitByReconciliationUuid(saveReq.Uuid); err != nil {
				return errors.WithMessage(err, "删除盘点单物品单位明细失败")
			}
		}

		var stockReconciliationItemUnits []*model.StockReconciliationItemUnit
		for _, reqItem := range saveReq.Items {
			// 计算实盘数量（基准单位）
			countedQuantity := decimal.Zero
			if len(reqItem.Units) > 0 {
				for _, unitItem := range reqItem.Units {
					if unitItem.Quantity == nil {
						continue
					}
					conversionRate := materialUnitMap[reqItem.MaterialUuid][unitItem.MaterialUnitUuid]
					unitQuantity := unitItem.Quantity.Mul(decimal.NewFromFloat(conversionRate))
					countedQuantity = countedQuantity.Add(unitQuantity)
				}
			}
			countedQuantity = countedQuantity.Truncate(3)

			item := &model.StockReconciliationItem{
				StockReconciliationUuid: stockReconciliation.Uuid,
				MaterialUuid:            reqItem.MaterialUuid,
				BookedQuantity:          decimal.NewFromFloat(bookedQuantityMap[reqItem.MaterialUuid]), // 每次保存都实时读取账面库存数量
				CountedQuantity:         countedQuantity,
			}

			// 先创建item以获取自动生成的uuid
			if err := stockReconciliationRepo.CreateStockReconciliationItem(item); err != nil {
				return errors.WithMessage(errors.New("创建盘点单物品明细失败"), err.Error())
			}

			// 创建单位明细
			for _, unitItem := range reqItem.Units {
				var quantity *float64
				if unitItem.Quantity != nil {
					quantityDecimal := unitItem.Quantity.InexactFloat64()
					quantity = &quantityDecimal
				}
				stockReconciliationItemUnits = append(stockReconciliationItemUnits, &model.StockReconciliationItemUnit{
					StockReconciliationItemUuid: item.Uuid,
					MaterialUnitUuid:            unitItem.MaterialUnitUuid,
					Quantity:                    quantity,
				})
			}
		}

		if len(stockReconciliationItemUnits) > 0 {
			if err := stockReconciliationRepo.CreateStockReconciliationItemUnitBatch(stockReconciliationItemUnits); err != nil {
				return errors.WithMessage(errors.New("创建盘点单物品单位明细失败"), err.Error())
			}
		}

		return nil
	})

	if err != nil {
		errMsg := "保存失败"
		if saveReq.IsSubmit {
			errMsg = "提交失败"
		}
		return stockReconciliationUuid, errors.WithMessage(errors.New(errMsg), err.Error())
	}

	// 提交盘点单
	if saveReq.IsSubmit && ctx.GetCompany().IsOpenErp() {
		err = s.submitStockReconciliation(ctx, stockReconciliation.Uuid, false)
		if err != nil {
			return stockReconciliationUuid, errors.WithMessage(err, "提交盘点单失败")
		}
	}

	return stockReconciliationUuid, nil
}
```

**关键点**:
1. **分布式锁**: 新建时使用日期锁保证单号唯一性，修改时使用 UUID 锁
2. **状态校验**: 只有已保存状态才能修改
3. **两种提交方式**:
   - 列表直接提交：不修改数据，直接调用 submitStockReconciliation
   - 保存后提交：先保存数据，再提交到 ERP
4. **多单位计算**: 根据换算率将各单位数量转换为基准单位
5. **事务保证**: 整个保存过程在事务中执行
6. **实时账面库存**: 保存时重新读取账面库存数量

---

### 4. submitStockReconciliation - 提交盘点单（私有方法）

**方法签名**:
```go
func (s *stockReconciliationSrv) submitStockReconciliation(ctx context.Context, stockReconciliationUuid uint64, isDirectSubmit bool) error
```

**功能**: 将盘点单提交到 ERP 系统，更新状态为已提交。

**参数说明**:
- `stockReconciliationUuid`: 盘点单 UUID
- `isDirectSubmit`: 是否列表直接提交（true: 列表提交，需更新账面库存；false: 保存后提交）

**实现流程**:

```441:563:ttpos-server-go/main/app/service/stock_reconciliation.go
func (s *stockReconciliationSrv) submitStockReconciliation(ctx context.Context, stockReconciliationUuid uint64, isDirectSubmit bool) error {
	db := ctx.GetDB()
	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)

	opts := []repository.DBOption{
		stockReconciliationRepo.WhereUuid(stockReconciliationUuid),
		stockReconciliationRepo.WithStockReconciliationItemsMultiLanguageName(),
		stockReconciliationRepo.WithStockReconciliationItemsUnits(),
		stockReconciliationRepo.WithWarehouse(),
	}
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliation(opts...)

	if err != nil {
		return errors.WithMessage(errors.New("查询盘点单失败"), err.Error())
	}
	if stockReconciliation == nil {
		return errors.New("盘点单不存在")
	}

	bookedQuantityMap := make(map[uint64]decimal.Decimal)
	if isDirectSubmit {
		var err error
		bookedQuantityMap, err = s.getBookedQuantityMap(db, stockReconciliation.WarehouseUuid)
		if err != nil {
			return errors.WithMessage(errors.New("查询仓库物品失败"), err.Error())
		}
	}

	companySetting := ctx.GetCompanySetting()

	// 根据时区获取过账日期和时间
	now := utils.SetTimezone(companySetting.GetTimezone()).Now()

	var erpItems []*stock.StockReconciliationItem
	err = db.Transaction(func(tx *gorm.DB) error {
		stockReconciliationRepo := repository.NewStockReconciliationRepo(tx)
		for _, item := range stockReconciliation.StockReconciliationItems {
			// 物品已禁用，标记item的delete_time(删除)
			if !item.Material.Status {
				if err := stockReconciliationRepo.DeleteStockReconciliationItem(item.Uuid); err != nil {
					return errors.WithMessage(errors.New("提交盘点单时移除已关闭物品失败"), err.Error())
				}
				continue
			}
			// 不往erp传递已删除的物品明细
			if item.DeleteTime > 0 {
				continue
			}

			var unitExists bool
			for _, unit := range item.StockReconciliationItemUnits {
				if unit.DeleteTime == 0 && unit.Quantity != nil {
					unitExists = true
					break
				}
			}
			if !unitExists {
				if err := stockReconciliationRepo.DeleteStockReconciliationItem(item.Uuid); err != nil {
					return errors.WithMessage(errors.New("提交盘点单时移除待盘点物品失败"), err.Error())
				}
				continue
			}

			if isDirectSubmit {
				stockReconciliationItem := *item
				stockReconciliationItem.BookedQuantity = bookedQuantityMap[item.MaterialUuid]
				if err := stockReconciliationRepo.UpdateStockReconciliationItem(&stockReconciliationItem); err != nil {
					return errors.WithMessage(errors.New("更新盘点单物品明细失败"), err.Error())
				}
			}

			erpItems = append(erpItems, &stock.StockReconciliationItem{
				ItemCode: item.Material.Code,
				Qty:      item.CountedQuantity.InexactFloat64(),
			})
		}

		if len(erpItems) == 0 {
			return errors.New("物品列表为空，请先添加物品后再操作")
		}
		erpSrv := erp.NewIErpSrv(s.dbm)
		erpReq, err := erpSrv.SubmitStockReconciliation(ctx, companySetting, &stock.SaveStockReconciliationReq{
			CompanyAbbr: companySetting.ErpnextCompanyAbbr,
			Branch:      companySetting.ErpnextBranchName,
			PostingDate: now.Format("2006-01-02"),
			PostingTime: now.Format("15:04:05"),
			Warehouse:   stockReconciliation.Warehouse.ErpCode,
			Items:       erpItems,
		})
		if err != nil {
			// 提取物品名称
			itemName := s.extractName("Item", "is disabled", err.Error())
			for _, item := range stockReconciliation.StockReconciliationItems {
				if item.Material.Code == itemName {
					materialName := item.Material.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
					message := i18n.Translate(ctx.GetLanguage(), "物品%s状态已关闭，请修改物品状态", materialName)
					return errors.New(message)
				}
			}
			if itemName != "" {
				return errors.New(i18n.Translate(ctx.GetLanguage(), "物品%s状态已关闭，请修改物品状态"), itemName)
			}
			return errors.WithMessage(errors.New("提交盘点单失败"), err.Error())
		}
		// 更新盘点单erp_code和提交时间
		stockReconciliation.ErpCode = erpReq.StockReconciliationName
		stockReconciliation.SubmitTime = int(time.Now().Unix())
		stockReconciliation.Status = constant.StockReconciliationStatusSubmitted
		if err := stockReconciliationRepo.UpdateStockReconciliation(stockReconciliation); err != nil {
			return errors.WithMessage(errors.New("更新盘点单状态失败"), err.Error())
		}

		return nil
	})
	if err != nil {
		return errors.WithMessage(err)
	}

	return nil
}
```

**关键点**:
1. **数据清理**: 提交前移除禁用物品和未盘点物品
2. **列表提交**: 直接提交时需更新账面库存快照
3. **ERP 集成**: 调用 ERP 接口提交盘点单
4. **错误本地化**: 提取 ERP 错误信息并转换为多语言
5. **状态更新**: 提交成功后更新状态为已提交，记录 ERP 单号

---

### 5. ApproveStockReconciliation - 审核盘点单

**方法签名**:
```go
func (s *stockReconciliationSrv) ApproveStockReconciliation(ctx context.Context, req req.StockReconciliationApproveReq) ([]dto.LocaleResponse, error)
```

**功能**: 审核通过盘点单，更新仓库实际库存，生成出入库日志，同步到 ERP。

**返回值**:
- `[]dto.LocaleResponse`: 禁用物品列表（如果有）
- `error`: 错误信息

**实现流程**:

```600:757:ttpos-server-go/main/app/service/stock_reconciliation.go
func (s *stockReconciliationSrv) ApproveStockReconciliation(ctx context.Context, req req.StockReconciliationApproveReq) ([]dto.LocaleResponse, error) {
	db := ctx.GetDB()

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)

	opts := []repository.DBOption{
		stockReconciliationRepo.WhereUuid(req.Uuid),
		stockReconciliationRepo.WithStockReconciliationItemsMultiLanguageName(),
		stockReconciliationRepo.WithStockReconciliationItemsMaterialBaseUnit(),
	}
	// 查询盘点单
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliation(opts...)
	if err != nil {
		return nil, errors.WithMessage(err, "查询盘点单失败")
	}
	if stockReconciliation == nil {
		return nil, errors.New("盘点单不存在")
	}

	// 只有已提交状态的盘点单才能审核
	if stockReconciliation.Status != constant.StockReconciliationStatusSubmitted {
		return nil, errors.New("当前状态不允许审核")
	}

	disabledMaterials := make([]dto.LocaleResponse, 0)
	for _, item := range stockReconciliation.StockReconciliationItems {
		// 跳过已删除物品明细
		if item.DeleteTime > 0 {
			continue
		}
		// 如果从提交到审核期间，物品已禁用，则提示请求该物品为开启
		if !item.Material.Status {
			disabledMaterials = append(disabledMaterials, item.Material.MultiLanguageName.GetNames())
		}
	}

	if len(disabledMaterials) > 0 {
		return disabledMaterials, errors.New("请修改物品状态")
	}

	// 获取仓库Uuid获取仓库物品信息列表
	warehouseItems, err := repository.NewWarehouseItemRepo(db).GetByWarehouseUuid(stockReconciliation.WarehouseUuid)
	if err != nil {
		return disabledMaterials, errors.WithMessage(errors.New("查询仓库物品失败"), err.Error())
	}
	warehouseMaterialUuids := make(map[uint64]struct{})
	for _, item := range warehouseItems {
		warehouseMaterialUuids[item.MaterialUuid] = struct{}{}
	}

	// 开启事务
	err = db.Transaction(func(tx *gorm.DB) error {
		stockReconciliationRepo := repository.NewStockReconciliationRepo(tx)

		// 更新盘点单状态为已审核
		updateData := map[string]any{
			"status":      constant.StockReconciliationStatusApproved, // 2-已审核
			"update_time": int(time.Now().Unix()),
		}
		if err := stockReconciliationRepo.UpdateStockReconciliationData(updateData, stockReconciliationRepo.WhereUuid(req.Uuid)); err != nil {
			return errors.WithMessage(errors.New("审核盘点单失败"), err.Error())
		}
		// 遍历所有物品
		var warehouseLogs []*model.WarehouseInOutLog
		for _, item := range stockReconciliation.StockReconciliationItems {
			// 跳过已删除物品明细
			if item.DeleteTime > 0 {
				continue
			}
			material := item.Material
			// 审核时，物品被删除，则删除物品明细
			if item.Material.DeleteTime > 0 {
				if err := stockReconciliationRepo.DeleteStockReconciliationItem(item.Uuid); err != nil {
					return errors.WithMessage(errors.New("审核盘点单时移除已删除物品失败"), err.Error())
				}
				continue
			}
			stockQuantity := item.CountedQuantity.Truncate(3).InexactFloat64()
			_, exists := warehouseMaterialUuids[item.MaterialUuid]
			if !exists { // 仓库不存在该物品，则新增仓库物品，库存数量为实盘数量
				if err := tx.Create(&model.WarehouseItem{
					WarehouseUuid: stockReconciliation.WarehouseUuid,
					MaterialUuid:  item.MaterialUuid,
					MaterialCode:  item.Material.Code,
					Stock:         stockQuantity,
					Valuation:     1.0, // 和同步仓库物品库存一致
				}).Error; err != nil {
					return errors.WithMessage(errors.New("创建仓库物品失败"), err.Error())
				}
			} else { // 仓库存在该物品，则更新仓库物品库存为实盘数量
				if err := tx.Model(&model.WarehouseItem{}).
					Where("warehouse_uuid = ?", stockReconciliation.WarehouseUuid).
					Where("material_uuid = ?", item.MaterialUuid).Update("stock", stockQuantity).Error; err != nil {
					return errors.WithMessage(errors.New("更新仓库物品库存失败"), err.Error())
				}
			}

			// 有盈亏才记录日志
			if !item.CountedQuantity.Equal(item.BookedQuantity) {
				scene := constant.WarehouseInOutLogSceneProfitIn        // 盘盈
				logType := constant.WarehouseInOutLogLogTypeIn          // 入库
				if item.CountedQuantity.LessThan(item.BookedQuantity) { // 盘亏出库
					logType = constant.WarehouseInOutLogLogTypeOut
					scene = constant.WarehouseInOutLogSceneLossOut
				}
				diff := item.CountedQuantity.Sub(item.BookedQuantity).Abs()
				warehouseLogs = append(warehouseLogs, &model.WarehouseInOutLog{
					LogType:              logType,
					Scene:                scene,
					WarehouseUuid:        stockReconciliation.WarehouseUuid,
					MaterialUuid:         item.MaterialUuid,
					MaterialName:         item.MaterialName,
					MaterialBaseUnitUuid: material.Unit.Uuid, // 基准单位
					MaterialBaseUnitName: material.Unit.Name, // 基准单位名称
					Num:                  diff.Truncate(3).InexactFloat64(),
					Price:                material.Valuation,
					Amount:               decimal.NewFromFloat(material.Valuation).Mul(diff).Truncate(3).InexactFloat64(),
					OrderNo:              stockReconciliation.OrderNo,
				})
			}
		}
		if len(warehouseLogs) > 0 {
			if err := tx.Create(&warehouseLogs).Error; err != nil {
				return errors.WithMessage(errors.New("创建盘盈盘亏出入库记录失败"), err.Error())
			}
		}

		// 调用erp接口审核盘点单
		if ctx.GetCompany().IsOpenErp() && stockReconciliation.ErpCode != "" {
			companySetting := ctx.GetCompanySetting()
			erpSrv := erp.NewIErpSrv(s.dbm)
			_, err := erpSrv.ApproveStockReconciliation(ctx, companySetting, &stock.SubmitStockReconciliationReq{
				StockReconciliationName: stockReconciliation.ErpCode,
			})
			if err != nil {
				return errors.WithMessage(errors.New("审核盘点单失败"), err.Error())
			}
		}

		return nil
	})
	if err != nil {
		return disabledMaterials, err
	}
	utils.Go(func() {
		// 计算所有关联成本卡的商品的库存
		err = s.productSrv.SyncProductStockByBomCard(ctx)
		if err != nil {
			logger.Logger.Error("审核通过盘点单-计算商品库存失败", zap.Error(err))
		}
	})
	return disabledMaterials, nil
}
```

**关键点**:
1. **状态校验**: 只有已提交状态才能审核
2. **物品状态检查**: 禁用物品返回列表，不阻断审核
3. **库存更新**:
   - 仓库不存在物品：新增仓库物品记录
   - 仓库存在物品：更新库存为实盘数量
4. **出入库日志**: 只记录有差异的物品
   - 盘盈：入库日志
   - 盘亏：出库日志
5. **ERP 同步**: 调用 ERP 接口审核
6. **商品库存同步**: 异步计算商品库存（基于成本卡）

---

### 6. DeleteStockReconciliation - 删除盘点单

**方法签名**:
```go
func (s *stockReconciliationSrv) DeleteStockReconciliation(ctx context.Context, req req.StockReconciliationDeleteReq) error
```

**功能**: 删除已保存状态的盘点单。

**实现流程**:

```565:598:ttpos-server-go/main/app/service/stock_reconciliation.go
func (s *stockReconciliationSrv) DeleteStockReconciliation(ctx context.Context, req req.StockReconciliationDeleteReq) error {
	db := ctx.GetDB()

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)

	// 查询盘点单
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliationByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "查询盘点单失败")
	}
	if stockReconciliation == nil {
		return errors.New("盘点单不存在")
	}

	// 只有已保存和已驳回状态的盘点单才能删除
	if stockReconciliation.Status != constant.StockReconciliationStatusSaved {
		return errors.New("当前状态不允许删除")
	}

	// 删除盘点单以及物品明细和单位明细
	if err := stockReconciliationRepo.DeleteStockReconciliation(req.Uuid); err != nil {
		return errors.WithMessage(err, "删除盘点单失败")
	}

	// 清理UUID锁资源
	s.lock.ClearUuidLock(req.Uuid)

	return nil
}
```

**关键点**:
1. 只允许删除已保存状态的盘点单
2. 级联删除物品明细和单位明细
3. 清理分布式锁资源

---

### 7. RejectStockReconciliation - 驳回盘点单

**方法签名**:
```go
func (s *stockReconciliationSrv) RejectStockReconciliation(ctx context.Context, req req.StockReconciliationRejectReq) error
```

**功能**: 驳回已提交的盘点单。

**实现流程**:

```759:793:ttpos-server-go/main/app/service/stock_reconciliation.go
func (s *stockReconciliationSrv) RejectStockReconciliation(ctx context.Context, req req.StockReconciliationRejectReq) error {
	db := ctx.GetDB()

	// 加锁
	s.lock.LockUuid(req.Uuid)
	defer s.lock.UnlockUuid(req.Uuid)

	stockReconciliationRepo := repository.NewStockReconciliationRepo(db)

	// 查询盘点单
	stockReconciliation, err := stockReconciliationRepo.GetStockReconciliationByUuid(req.Uuid)
	if err != nil {
		return errors.WithMessage(err, "查询盘点单失败")
	}
	if stockReconciliation == nil {
		return errors.New("盘点单不存在")
	}

	// 只有已提交状态的盘点单才能驳回
	if stockReconciliation.Status != constant.StockReconciliationStatusSubmitted {
		return errors.New("盘点单状态不允许驳回")
	}

	// 更新盘点单状态为已驳回
	updateData := map[string]any{
		"status":      constant.StockReconciliationStatusRejected, // 3-已驳回
		"update_time": int(time.Now().Unix()),
	}
	if err := stockReconciliationRepo.UpdateStockReconciliationData(updateData, stockReconciliationRepo.WhereUuid(req.Uuid)); err != nil {
		return errors.WithMessage(err, "驳回盘点单失败")
	}

	return nil
}
```

**关键点**:
1. 只允许驳回已提交状态的盘点单
2. 驳回后状态变为已驳回，可以删除
3. 不影响 ERP 中的数据

---

### 8. CheckMaterials - 检查物品

**方法签名**:
```go
func (s *stockReconciliationSrv) CheckMaterials(ctx context.Context, checkReq req.StockReconciliationCheckMaterialsReq) (resp.StockReconciliationCheckMaterialsListResp, error)
```

**功能**: 检查物品的盘点状态、异常情况、禁用状态等，用于前端提示。

**请求参数**:
```go
type StockReconciliationCheckMaterialsReq struct {
    Uuid          uint64                                    // 盘点单UUID
    WarehouseUuid uint64                                    // 仓库UUID
    Items         []StockReconciliationCheckMaterialsItem   // 物品列表
}

type StockReconciliationCheckMaterialsItem struct {
    MaterialUuid    uint64          // 物品UUID
    CountedQuantity decimal.Decimal // 实盘数量
}
```

**返回数据**:
```go
type StockReconciliationCheckMaterialsResp struct {
    LocaleName                 dto.LocaleResponse // 物品名称
    Status                     bool               // 物品状态（启用/禁用）
    IsDeleted                  bool               // 是否已删除
    IsInventoryStatusException bool               // 是否盘点异常
    UnitCount                  uint               // 已盘点单位数量
}
```

**关键点**:
1. 支持根据盘点单 UUID 或仓库 UUID 查询
2. 检测盘点异常（差值超过20%）
3. 检测物品状态（禁用、删除）
4. 统计已盘点单位数量

---

### 9. generateOrderNo - 生成单据编号

**方法签名**:
```go
func (s *stockReconciliationSrv) generateOrderNo(db *gorm.DB, timezone string) string
```

**功能**: 在事务内部生成唯一的盘点单编号。

**编号格式**: `ST + 年月日 + 4位序列号`
- 例如: `ST202510160001`
- 序列号从 0001 开始，每天重置

**实现流程**:

```795:834:ttpos-server-go/main/app/service/stock_reconciliation.go
func (s *stockReconciliationSrv) generateOrderNo(db *gorm.DB, timezone string) string {
	// 生成格式：ST + 年月日 + 0000（4位序列号）
	// 例如：ST202510160001
	// 序列号从0001开始递增，每天重置
	repo := repository.NewStockReconciliationRepo(db)

	// 使用商家时区格式化日期
	dateStr := utils.SetTimezone(timezone).Now().Format("20060102")
	prefix := fmt.Sprintf("ST%s", dateStr)

	// 查询当天最大的订单号
	maxOrderNo, err := repo.GetMaxOrderNoByPrefix(prefix)
	if err != nil {
		logger.Logger.Error("查询最大单据编号失败", zap.Error(err))
		// 如果查询失败，返回第一个序列号
		return fmt.Sprintf("%s0001", prefix)
	}

	// 如果没有找到当天的订单号，从0001开始
	if maxOrderNo == "" {
		return fmt.Sprintf("%s0001", prefix)
	}

	// 从订单号中提取序列号（最后4位）
	if len(maxOrderNo) < 4 {
		return fmt.Sprintf("%s0001", prefix)
	}

	// 获取序列号部分
	seqStr := maxOrderNo[len(maxOrderNo)-4:]
	seq := 0
	fmt.Sscanf(seqStr, "%d", &seq)

	// 序列号+1
	seq++

	// 生成新的订单号
	return fmt.Sprintf("%s%04d", prefix, seq)
}
```

**关键点**:
1. 必须在事务内调用，保证并发安全
2. 使用商家时区，而非服务器时区
3. 序列号每天重置，从 0001 开始
4. 查询失败时返回第一个序列号，保证业务连续性

---

### 10. validateWarehouseAndItems - 验证仓库和物品

**方法签名**:
```go
func (s *stockReconciliationSrv) validateWarehouseAndItems(db *gorm.DB, req req.StockReconciliationSaveReq) ([]model.WarehouseItem, []*model.Material, error)
```

**功能**: 验证仓库和物品明细的有效性。

**验证内容**:
1. 仓库是否存在且类型为 normal（普通仓库）
2. 盘点目的参数是否正确（1-月度盘点，2-年度盘点）
3. 物品是否存在
4. 物品单位是否正确

**返回值**:
- `[]model.WarehouseItem`: 仓库物品列表
- `[]*model.Material`: 物品详情列表（含单位信息）
- `error`: 错误信息

---

### 11. getIsInventoryStatusException - 判断盘点异常

**方法签名**:
```go
func (s *stockReconciliationSrv) getIsInventoryStatusException(bookedQuantity decimal.Decimal, countedQuantity decimal.Decimal) bool
```

**功能**: 判断盘点是否异常（差值超过账面库存的20%）。

**异常判断逻辑**:

```913:922:ttpos-server-go/main/app/service/stock_reconciliation.go
func (s *stockReconciliationSrv) getIsInventoryStatusException(bookedQuantity decimal.Decimal, countedQuantity decimal.Decimal) bool {
	if bookedQuantity.IsZero() {
		if countedQuantity.IsZero() {
			return false
		}
		return true
	}
	return countedQuantity.Sub(bookedQuantity).Abs().Div(bookedQuantity).GreaterThan(decimal.NewFromFloat(0.2))
}
```

**特殊情况**:
- 账面为0，实盘为0：正常
- 账面为0，实盘不为0：异常
- 账面不为0：差值绝对值 / 账面库存 > 20% 为异常

---

### 12. getBookedQuantityMap - 获取账面库存

**方法签名**:
```go
func (s *stockReconciliationSrv) getBookedQuantityMap(db *gorm.DB, warehouseUuid uint64) (map[uint64]decimal.Decimal, error)
```

**功能**: 获取仓库中所有物品的账面库存数量。

**返回值**: `map[物品UUID]账面库存数量`

---

### 13. extractName - 从错误信息中提取名称

**方法签名**:
```go
func (s *stockReconciliationSrv) extractName(name, after, errorMsg string) string
```

**功能**: 从 ERP 返回的错误信息中提取物品名称或其他信息，用于本地化错误提示。

**使用示例**:
```go
// ERP 错误: "Item ABC123 is disabled"
itemName := s.extractName("Item", "is disabled", errMsg)
// 返回: "ABC123"
```

**实现流程**:

```1027:1054:ttpos-server-go/main/app/service/stock_reconciliation.go
func (s *stockReconciliationSrv) extractName(name, after, errorMsg string) string {
	// 转义正则表达式中的特殊字符
	escapedName := regexp.QuoteMeta(name)
	escapedAfter := regexp.QuoteMeta(after)
	// 使用正则表达式匹配，支持HTML标签和普通文本
	var re *regexp.Regexp
	if strings.Contains(name, "<") || strings.Contains(after, "<") {
		// HTML标签模式：不要求空格分隔
		re = regexp.MustCompile(escapedName + `(.+?)` + escapedAfter)
	} else {
		// 普通文本模式：要求空格分隔
		re = regexp.MustCompile(escapedName + `\s+(.+?)\s+` + escapedAfter)
	}
	matches := re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		supplierInfo := strings.TrimSpace(matches[1])
		// 如果包含编码#名称格式，提取名称部分
		if strings.Contains(supplierInfo, "#") {
			parts := strings.SplitN(supplierInfo, "#", 2)
			if len(parts) == 2 {
				return parts[1] // 返回物品erp_code
			}
		}
		return supplierInfo
	}
	return ""
}
```

---

## ERP 集成详解

### 集成架构

```
TTPOS Backend ←→ TTPOS-BMP (中台) ←→ ERPNext
```

### 集成流程

#### 1. 提交盘点单到 ERP

```
前端提交 → SaveStockReconciliation 
         → submitStockReconciliation 
         → erp.SubmitStockReconciliation 
         → ERPNext API
```

**提交数据**:
```go
type SaveStockReconciliationReq struct {
    CompanyAbbr string                       // 公司简称
    Branch      string                       // 分店名称
    PostingDate string                       // 过账日期
    PostingTime string                       // 过账时间
    Warehouse   string                       // 仓库 ERP 编码
    Items       []*StockReconciliationItem   // 物品列表
}

type StockReconciliationItem struct {
    ItemCode string  // 物品 ERP 编码
    Qty      float64 // 实盘数量
}
```

**ERP 返回**:
- `StockReconciliationName`: ERP 盘点单编号
- 保存到本地 `erp_code` 字段

#### 2. 审核盘点单到 ERP

```
前端审核 → ApproveStockReconciliation 
         → 更新本地库存 
         → erp.ApproveStockReconciliation 
         → ERPNext API
```

**审核数据**:
```go
type SubmitStockReconciliationReq struct {
    StockReconciliationName string // ERP 盘点单编号
}
```

### 错误处理

#### 1. 物品禁用错误
```
ERP 错误: "Item ABC123 is disabled"
↓
extractName 提取物品编码
↓
查找本地物品名称
↓
返回多语言错误: "物品[珍珠奶茶]状态已关闭，请修改物品状态"
```

#### 2. 其他 ERP 错误
- 直接返回 ERP 原始错误信息
- 前缀添加"提交盘点单失败"或"审核盘点单失败"

---

## 数据模型

### StockReconciliation - 盘点单主表

```go
type StockReconciliation struct {
    Uuid          uint64 `gorm:"primary_key"` // 盘点单UUID
    OrderNo       string                      // 单据编号
    Type          int                         // 类型
    WarehouseUuid uint64                      // 仓库UUID
    Purpose       int                         // 盘点目的（1-月度，2-年度）
    Status        int                         // 状态（0-已保存，1-已提交，2-已审核，3-已驳回）
    ErpCode       string                      // ERP盘点单编号
    SubmitTime    int                         // 提交时间
    CreateTime    int                         // 创建时间
    UpdateTime    int                         // 更新时间
    DeleteTime    int                         // 删除时间（软删除）
    
    // 关联
    Warehouse                 Warehouse
    StockReconciliationItems  []StockReconciliationItem
}
```

### StockReconciliationItem - 盘点单物品明细

```go
type StockReconciliationItem struct {
    Uuid                    uint64          // 明细UUID
    StockReconciliationUuid uint64          // 盘点单UUID
    MaterialUuid            uint64          // 物品UUID
    MaterialName            string          // 物品名称（冗余）
    BookedQuantity          decimal.Decimal // 账面数量（基准单位）
    CountedQuantity         decimal.Decimal // 实盘数量（基准单位）
    DeleteTime              int             // 删除时间
    
    // 关联
    Material                       Material
    StockReconciliationItemUnits   []StockReconciliationItemUnit
}
```

### StockReconciliationItemUnit - 盘点单位明细

```go
type StockReconciliationItemUnit struct {
    Uuid                        uint64   // 单位明细UUID
    StockReconciliationItemUuid uint64   // 盘点单物品明细UUID
    MaterialUnitUuid            uint64   // 物品单位UUID
    Quantity                    *float64 // 数量（可能为空，表示未盘点）
    DeleteTime                  int      // 删除时间
    
    // 关联
    MaterialUnit MaterialUnit
}
```

### WarehouseInOutLog - 仓库出入库日志

```go
type WarehouseInOutLog struct {
    LogType              int     // 日志类型（1-入库，2-出库）
    Scene                int     // 场景（盘盈入库、盘亏出库）
    WarehouseUuid        uint64  // 仓库UUID
    MaterialUuid         uint64  // 物品UUID
    MaterialName         string  // 物品名称
    MaterialBaseUnitUuid uint64  // 基准单位UUID
    MaterialBaseUnitName string  // 基准单位名称
    Num                  float64 // 数量
    Price                float64 // 单价
    Amount               float64 // 金额
    OrderNo              string  // 关联单据编号
}
```

---

## 业务规则总结

### 1. 状态流转规则

| 当前状态 | 允许操作 | 目标状态 |
|----------|----------|----------|
| 已保存 | 编辑 | 已保存 |
| 已保存 | 提交 | 已提交 |
| 已保存 | 删除 | （删除） |
| 已提交 | 审核 | 已审核 |
| 已提交 | 驳回 | 已驳回 |
| 已审核 | （无） | - |
| 已驳回 | 删除 | （删除） |

### 2. 账面库存规则
- **已保存状态**: 每次查询/保存时实时读取账面库存
- **已提交/已审核状态**: 使用快照账面库存（保存时的值）

### 3. 盘点单位规则
- 同一物品可用多个单位盘点
- 系统自动按换算率转换为基准单位
- 实盘数量 = Σ(各单位数量 × 换算率)

### 4. 数据清理规则

**提交时自动清理**:
- 已禁用的物品
- 未盘点的物品（所有单位数量都为空）

**审核时自动清理**:
- 审核期间被删除的物品

### 5. 库存更新规则

**审核通过后**:
1. 仓库不存在物品：新增记录，库存 = 实盘数量
2. 仓库存在物品：更新库存 = 实盘数量
3. 生成出入库日志（仅盘盈盘亏）
4. 异步同步商品库存

### 6. 单据编号规则
- 格式: `ST + YYYYMMDD + 序列号`
- 序列号4位，从 0001 开始
- 每天重置
- 使用商家时区

---

## 使用场景

### 场景1: 月度库存盘点

```
背景: 每月末进行库存盘点
流程:
1. 创建盘点单，选择仓库和盘点目的（月度盘点）
2. 添加需要盘点的物品
3. 仓库人员使用不同单位进行盘点（箱、瓶、个等）
4. 保存盘点单（草稿状态）
5. 检查无误后提交到 ERP
6. 财务或管理员审核
7. 审核通过，系统自动更新库存
8. 生成盘盈盘亏报告
```

### 场景2: 年度全面盘点

```
背景: 年末进行全仓库盘点
流程:
1. 创建盘点单，盘点目的选择年度盘点
2. 批量导入仓库所有物品
3. 多人协作盘点，实时保存
4. 系统提示盘点异常（差值超20%）
5. 核对异常物品后提交
6. 管理层审核
7. 库存更新，生成年度盘点报告
```

### 场景3: 盘点异常处理

```
背景: 盘点发现大量差异
流程:
1. 提交盘点单
2. 审核人员发现多个物品异常（IsInventoryStatusException = true）
3. 驳回盘点单
4. 重新盘点异常物品
5. 修改盘点数据
6. 再次提交
7. 审核通过
```

### 场景4: 列表快速提交

```
背景: 已在 ERP 完成盘点，只需同步
流程:
1. 在列表页直接点击提交
2. 系统实时读取账面库存
3. 提交到 ERP
4. 快速审核
```

---

## 并发控制

### 1. 分布式锁策略

**新建盘点单**:
```go
lockKey := fmt.Sprintf("stock_reconciliation_%d_%s", companyUuid, dateStr)
s.lock.LockUuidString(lockKey)
defer s.lock.UnlockUuidString(lockKey)
```
- 锁粒度: 公司 + 日期
- 目的: 保证单据编号唯一性

**修改/删除/审核/驳回**:
```go
s.lock.LockUuid(stockReconciliationUuid)
defer s.lock.UnlockUuid(stockReconciliationUuid)
```
- 锁粒度: 盘点单 UUID
- 目的: 防止并发修改

### 2. 数据库事务

所有涉及多表操作的方法都使用事务：
- SaveStockReconciliation
- submitStockReconciliation
- ApproveStockReconciliation

### 3. 锁资源清理

删除盘点单时清理 UUID 锁资源：
```go
s.lock.ClearUuidLock(req.Uuid)
```

---

## 错误处理

### 1. 业务错误

| 错误场景 | 错误消息 | 处理方式 |
|----------|----------|----------|
| 盘点单不存在 | "盘点单不存在" | 返回错误 |
| 状态不允许操作 | "当前状态不允许XXX" | 返回错误 |
| 物品禁用 | "物品[XX]状态已关闭" | 返回错误 |
| 仓库参数错误 | "仓库参数错误" | 返回错误 |
| 物品列表为空 | "物品列表为空，请先添加物品" | 返回错误 |

### 2. 数据库错误

使用 `errors.WithMessage` 包装错误：
```go
if err != nil {
    return errors.WithMessage(err, "查询盘点单失败")
}
```

### 3. ERP 错误

**物品禁用错误**:
```go
itemName := s.extractName("Item", "is disabled", err.Error())
message := i18n.Translate(ctx.GetLanguage(), "物品%s状态已关闭，请修改物品状态", materialName)
return errors.New(message)
```

**其他 ERP 错误**:
```go
return errors.WithMessage(errors.New("提交盘点单失败"), err.Error())
```

---

## 性能优化

### 1. 批量查询优化

**物品数量统计**:
```go
// 一次查询获取所有盘点单的物品数量
itemsCountMap, err := stockReconciliationRepo.GetStockReconciliationItemCountListByReconciliationUuidList(stockReconciliationUuidList)
```

**物品详情查询**:
```go
// 批量查询物品详情，而非逐个查询
materials, err := materialRepo.GetMaterialDetailByUuids(materialUuids)
```

### 2. 预加载优化

使用 GORM Preload 减少查询次数：
```go
opts := []repository.DBOption{
    stockReconciliationRepo.WithWarehouseMultiLanguageName(),
    stockReconciliationRepo.WithStockReconciliationItemMaterialUnits(),
}
```

### 3. 异步处理

**商品库存同步**:
```go
utils.Go(func() {
    err = s.productSrv.SyncProductStockByBomCard(ctx)
})
```
- 审核后异步计算商品库存
- 不阻塞审核流程

### 4. 精确小数计算

使用 `decimal` 库进行金额和数量计算：
```go
countedQuantity := unitItem.Quantity.Mul(decimal.NewFromFloat(conversionRate))
countedQuantity = countedQuantity.Truncate(3) // 保留3位小数
```

---

## 最佳实践

### 1. 盘点前准备

```go
// 创建盘点单前的检查
1. 确认仓库类型为 normal
2. 确认物品列表已准备
3. 确认物品单位配置正确
4. 了解盘点目的（月度/年度）
```

### 2. 盘点过程

```go
// 推荐的盘点流程
1. 创建盘点单并保存（草稿）
2. 分批盘点，实时保存
3. 使用 CheckMaterials 接口检查异常
4. 核对异常物品
5. 确认无误后提交
```

### 3. 审核建议

```go
// 审核前检查
1. 检查是否有禁用物品
2. 检查盘点异常物品（差值>20%）
3. 确认 ERP 连接正常
4. 审核通过后立即检查库存更新结果
```

### 4. 错误处理

```go
// 调用示例
uuid, err := stockReconciliationSrv.SaveStockReconciliation(ctx, req)
if err != nil {
    // 记录详细日志
    logger.Logger.Error("保存盘点单失败", 
        zap.Error(err),
        zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
        zap.Any("request", req),
    )
    
    // 返回友好错误
    return errors.New("保存盘点单失败，请稍后重试")
}
```

---

## 潜在改进点

### 1. 批量审核

**当前**: 只支持单个审核
**改进**: 支持批量审核多个盘点单
```go
func (s *stockReconciliationSrv) BatchApproveStockReconciliation(ctx context.Context, uuids []uint64) error
```

### 2. 盘点模板

**当前**: 每次手动选择物品
**改进**: 支持保存盘点模板
```go
type StockReconciliationTemplate struct {
    Name          string
    WarehouseUuid uint64
    MaterialUuids []uint64
}
```

### 3. 盘点进度追踪

**当前**: 无进度展示
**改进**: 显示盘点进度
```go
type StockReconciliationProgress struct {
    TotalItems   int // 总物品数
    CountedItems int // 已盘点物品数
    Progress     float64 // 进度百分比
}
```

### 4. 盘点历史对比

**当前**: 无历史对比
**改进**: 支持与上次盘点对比
```go
func (s *stockReconciliationSrv) CompareWithLastReconciliation(ctx context.Context, warehouseUuid uint64) (CompareResult, error)
```

### 5. 盘点照片附件

**当前**: 无附件支持
**改进**: 支持上传盘点现场照片
```go
type StockReconciliationAttachment struct {
    ReconciliationUuid uint64
    FileUuid           uint64
    FileUrl            string
}
```

### 6. 盘点异常自动提醒

**当前**: 需要手动调用 CheckMaterials
**改进**: 保存时自动检测并提醒
```go
// 保存时自动检测
异常物品 > 3个 → 发送通知给管理员
差值 > 50% → 强制要求填写原因
```

### 7. 导入导出功能

**当前**: 无导入导出
**改进**: 支持 Excel 导入导出
```go
func (s *stockReconciliationSrv) ExportToExcel(ctx context.Context, uuid uint64) ([]byte, error)
func (s *stockReconciliationSrv) ImportFromExcel(ctx context.Context, file []byte) error
```

### 8. 权限细化

**当前**: 基于 Context 的基本权限
**改进**: 细化操作权限
```go
permissions := []string{
    "stock_reconciliation.create",
    "stock_reconciliation.edit",
    "stock_reconciliation.submit",
    "stock_reconciliation.approve",
    "stock_reconciliation.reject",
    "stock_reconciliation.delete",
}
```

---

## 多语言支持

### 1. 仓库名称

```go
info.WarehouseLocaleName = item.Warehouse.MultiLanguageName.GetNames()
// 返回: {"zh-CN": "总仓", "en-US": "Main Warehouse"}
```

### 2. 物品名称

```go
itemInfo.LocaleName = item.Material.MultiLanguageName.GetNames()
// 返回: {"zh-CN": "珍珠", "en-US": "Pearl"}
```

### 3. 单位名称

```go
unitInfo.LocaleName = itemUnit.MaterialUnit.Unit.MultiLanguageName.GetNames()
// 返回: {"zh-CN": "箱", "en-US": "Box"}
```

### 4. 错误消息

```go
message := i18n.Translate(ctx.GetLanguage(), "物品%s状态已关闭，请修改物品状态", materialName)
// 根据用户语言返回对应翻译
```

---

## 相关文件

### DTO 定义
- `ttpos-server-go/app/dto/req/stock_reconciliation.go` - 请求参数
- `ttpos-server-go/app/dto/resp/stock_reconciliation.go` - 响应数据

### 数据仓库
- `ttpos-server-go/app/repository/stock_reconciliation.go` - 盘点单仓库
- `ttpos-server-go/app/repository/warehouse.go` - 仓库仓库
- `ttpos-server-go/app/repository/warehouse_item.go` - 仓库物品仓库
- `ttpos-server-go/app/repository/material.go` - 物品仓库

### 数据模型
- `ttpos-server-go/app/model/stock_reconciliation.go` - 盘点单模型
- `ttpos-server-go/app/model/warehouse.go` - 仓库模型
- `ttpos-server-go/app/model/material.go` - 物品模型

### RPC 服务
- `ttpos-server-go/app/service/rpc/erp/stock_reconciliation.go` - ERP 集成

### 常量定义
- `ttpos-server-go/app/constant/stock_reconciliation.go` - 盘点单常量
- `ttpos-server-go/app/constant/warehouse.go` - 仓库常量

---

## 总结

盘点单服务是餐饮库存管理系统的核心模块，具有以下特点：

1. **完整的生命周期管理**: 从创建到审核的完整流程
2. **深度 ERP 集成**: 与 ERPNext 双向同步，保证数据一致性
3. **灵活的多单位支持**: 支持同一物品使用多个单位盘点
4. **智能异常检测**: 自动识别盘点异常（差值超20%）
5. **精确的库存更新**: 审核后自动更新仓库库存和商品库存
6. **完善的并发控制**: 使用分布式锁保证数据一致性
7. **友好的错误处理**: 本地化错误提示，提取 ERP 错误信息
8. **强大的查询能力**: 支持多维度筛选和分页
9. **详细的操作日志**: 记录盘盈盘亏出入库日志
10. **国际化支持**: 所有名称和错误消息支持多语言

该服务适用于各种规模的餐饮企业进行定期或不定期的库存盘点，确保账实相符，为经营决策提供准确的库存数据。

