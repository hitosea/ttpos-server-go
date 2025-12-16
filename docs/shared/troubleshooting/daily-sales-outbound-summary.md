# 销售出库汇总定时任务逻辑文档

## 概述

销售出库汇总定时任务（`DailySalesOutboundSummaryTask`）是一个每小时执行一次的定时任务，用于在门店营业结束后自动统计当天的销售出库记录，并将汇总结果写入 `ttpos_warehouse_in_out_log` 表。

## 执行时机

- **定时执行**: 每小时整点执行（Cron: `0 0 * * * *`）
- **触发条件**: 门店营业结束时间到达后
- **执行频率**: 每个门店每个营业时段只统计一次

## 核心流程

### 1. 任务入口 (`Execute`)

```42:78:main/app/tasks/daily_sales_outbound_summary.go
// Execute 执行定时任务
func (t *DailySalesOutboundSummaryTask) Execute() {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Error("每日销售出库汇总定时任务,发生panic: %v", zap.Any("panic", r))
		}
	}()
	logger.Logger.Info("开始执行每日销售出库汇总定时任务")

	start := time.Now()
	// 分布式锁,避免多个节点同时执行
	lock.NewSystemLock().LockUuid(lock.DailySalesOutboundSummaryLock)
	defer lock.NewSystemLock().UnlockUuid(lock.DailySalesOutboundSummaryLock)
	spend := time.Since(start)
	logger.Logger.Info("分布式锁耗时: %v", zap.Duration("spend", spend))
	if spend > 1*time.Second {
		logger.Logger.Warn("其他节点已经处理该任务，本节点跳过")
		return
	}

	// 获取所有门店
	companies, err := t.getAllCompanies()
	if err != nil {
		logger.Logger.Error("获取门店列表失败: %v", zap.Error(err))
		return
	}

	logger.Logger.Info("找到 %d 个门店，开始检查营业结束时间", zap.Int("company_count", len(companies)))

	for _, company := range companies {
		if err := t.ProcessCompany(company); err != nil {
			logger.Logger.Error("处理门店 %s (%d) 失败: %v", zap.String("company_name", company.Name), zap.Uint64("company_uuid", company.Uuid), zap.Error(err))
			continue
		}
	}

	logger.Logger.Info("每日销售出库汇总定时任务执行完成")
}
```

**关键点**:
- 使用分布式锁防止多节点重复执行
- 如果获取锁耗时超过 1 秒，说明其他节点正在处理，当前节点跳过
- 遍历所有门店，逐个处理

### 2. 获取门店列表 (`getAllCompanies`)

```80:93:main/app/tasks/daily_sales_outbound_summary.go
// getAllCompanies 获取所有门店
func (t *DailySalesOutboundSummaryTask) getAllCompanies() ([]*model.Company, error) {
	var companies []*model.Company
	db := t.dbm.GetDB(0)

	// 查询所有未删除且启用的门店，包含设置信息
	err := db.Model(&model.Company{}).
		Joins("LEFT JOIN ttpos_company_setting ON ttpos_company.uuid = ttpos_company_setting.company_uuid").
		Where("ttpos_company.delete_time = 0 AND ttpos_company.status = 1").
		Preload("CompanySetting").
		Find(&companies).Error

	return companies, err
}
```

**查询条件**:
- `delete_time = 0`: 未删除的门店
- `status = 1`: 启用状态的门店
- 预加载门店设置信息（用于获取时区）

### 3. 处理单个门店 (`ProcessCompany`)

```95:156:main/app/tasks/daily_sales_outbound_summary.go
// processCompany 处理单个门店
func (t *DailySalesOutboundSummaryTask) ProcessCompany(company *model.Company) error {
	// 判断是否是erp商品,不是的不处理
	if !company.IsOpenErp() {
		return nil
	}

	// 获取门店营业时间
	openingHours, err := t.getOpeningHours(company.Uuid)
	if err != nil {
		return fmt.Errorf("获取门店营业时间失败: %w", err)
	}

	// 检查是否到达营业结束时间
	isBusinessEndTime, startTime, endTime := t.isBusinessEndTime(company, openingHours)
	if !isBusinessEndTime {
		logger.Logger.Info("门店 %s 未到达营业结束时间，跳过处理", zap.String("company_name", company.Name))
		return nil
	}

	// 判断该营业时段是否已经统计过
	year := time.Now().Format("20060102") // 20251023
	openingYearHours := fmt.Sprintf("%s %s", year, openingHours)
	db := t.dbm.GetDB(company.Uuid)
	warehouseLogRepo := repository.NewWarehouseInOutLogRepo(db)
	opts := []repository.DBOption{
		warehouseLogRepo.WhereLogType(1), // 出库
		warehouseLogRepo.WhereScene(1),   // 销售出库
		func(db *gorm.DB) *gorm.DB {
			return db.Where("opening_hours = ?", openingYearHours)
		},
	}
	existingLogs, err := warehouseLogRepo.GetWarehouseInOutLogs(opts...)
	if err != nil {
		return fmt.Errorf("获取营业时段记录失败: %w", err)
	}
	if len(existingLogs) > 0 {
		logger.Logger.Info(fmt.Sprintf("门店 %s 该营业时段%s已统计过，跳过处理", company.Name, openingYearHours))
		return nil
	}

	logger.Logger.Info("门店 %s 已到达营业结束时间，开始统计销售出库记录", zap.String("company_name", company.Name))

	// 统计当天销售出库记录
	outboundRecords, err := t.getDailySalesOutboundRecords(company.Uuid, startTime, endTime)
	if err != nil {
		return fmt.Errorf("获取销售出库记录失败: %w", err)
	}

	if len(outboundRecords) == 0 {
		logger.Logger.Info("门店 %s 今日无销售出库记录", zap.String("company_name", company.Name))
		return nil
	}

	// 生成汇总记录并写入数据库
	if err := t.saveOutboundSummaryRecords(company.Uuid, outboundRecords, openingYearHours); err != nil {
		return fmt.Errorf("保存出库汇总记录失败: %w", err)
	}

	logger.Logger.Info("门店 %s 销售出库汇总完成，共处理 %d 条记录", zap.String("company_name", company.Name), zap.Int("outbound_record_count", len(outboundRecords)))
	return nil
}
```

**处理步骤**:
1. **ERP 检查**: 只处理开启了 ERP 功能的门店
2. **获取营业时间**: 从门店设置中获取营业时间（格式：`HH:mm-HH:mm`）
3. **判断营业结束**: 检查当前时间是否已超过营业结束时间
4. **去重检查**: 检查该营业时段是否已经统计过（通过 `opening_hours` 字段）
5. **获取出库记录**: 查询营业时段内的销售订单原料明细
6. **保存汇总**: 将汇总结果写入 `ttpos_warehouse_in_out_log` 表

### 4. 获取营业时间 (`getOpeningHours`)

```158:173:main/app/tasks/daily_sales_outbound_summary.go
// getOpeningHours 获取门店营业时间
func (t *DailySalesOutboundSummaryTask) getOpeningHours(companyUuid uint64) (string, error) {
	setting := setting.NewSrvImpl(database.GetDBManager(config.Database), cache.Global)
	ctx := context.NewContext()
	ctx.SetCompanyUuid(companyUuid)
	businessSetting, err := setting.GetBusinessSetting(ctx)
	if err != nil {
		return "", err
	}
	if businessSetting.OpeningHours != "" {
		return businessSetting.OpeningHours, nil
	}

	// 返回默认值
	return "00:00-23:59", nil
}
```

**返回值**:
- 格式: `HH:mm-HH:mm`（例如：`09:00-22:00`）
- 默认值: `00:00-23:59`（如果门店未配置）

### 5. 判断营业结束时间 (`isBusinessEndTime`)

```175:192:main/app/tasks/daily_sales_outbound_summary.go
// isBusinessEndTime 判断是否到达营业结束时间
func (t *DailySalesOutboundSummaryTask) isBusinessEndTime(company *model.Company, openingHours string) (bool, int64, int64) {
	timezone := "Asia/Shanghai" // 默认时区
	if company.CompanySetting != nil {
		timezone = company.CompanySetting.GetTimezone()
	}
	logger.Logger.Info(fmt.Sprintf("时区: %s", timezone))
	timeUtil := utils.Timezone(timezone)

	// 获取营业时间的开始和结束时间戳
	startTime, endTime := timeUtil.OpeningHoursStartEndUnix(openingHours, utils.WithOpeningHoursType(1))
	logger.Logger.Info(fmt.Sprintf("openingHours: %s,营业开始时间: %s(%d),营业结束时间: %s(%d)", openingHours, timeUtil.FormatUnixTime(startTime, "2006-01-02 15:04:05"), startTime, timeUtil.FormatUnixTime(endTime, "2006-01-02 15:04:05"), endTime))
	now := timeUtil.Now().Unix()

	// 如果当前时间在营业结束时间之后，则认为已到达营业结束时间
	logger.Logger.Info(fmt.Sprintf("当前时间: %s(%d), 营业结束时间: %s(%d)", timeUtil.FormatUnixTime(now, "2006-01-02 15:04:05"), now, timeUtil.FormatUnixTime(endTime, "2006-01-02 15:04:05"), endTime))
	return now >= endTime, startTime, endTime
}
```

**关键逻辑**:
- 使用门店配置的时区（默认：`Asia/Shanghai`）
- 将营业时间字符串转换为当天的时间戳
- 判断当前时间是否 >= 营业结束时间
- 返回：是否到达结束时间、开始时间戳、结束时间戳

### 6. 获取销售出库记录 (`getDailySalesOutboundRecords`)

```194:253:main/app/tasks/daily_sales_outbound_summary.go
// getDailySalesOutboundRecords 获取当天销售出库记录
func (t *DailySalesOutboundSummaryTask) getDailySalesOutboundRecords(companyUuid uint64, startTime int64, endTime int64) ([]*OutboundRecord, error) {
	db := t.dbm.GetDB(companyUuid)

	// 使用 repository 方法查询出库单明细
	saleOrderMaterialRepo := repository.NewSaleOrderMaterialRepo(db)
	saleOrderMaterials, err := saleOrderMaterialRepo.GetSaleOrderMaterialByCreateTimeBetween(
		startTime,
		endTime,
	)
	if err != nil {
		return nil, err
	}

	// 按仓库和物料分组汇总数量
	recordMap := make(map[string]*OutboundRecord)
	for _, item := range saleOrderMaterials {
		key := fmt.Sprintf("%d_%d", item.WarehouseUuid, item.MaterialUuid)
		if record, exists := recordMap[key]; exists {
			record.TotalNum += item.Num
		} else {
			materialName := ""
			materialBaseUnitUuid := uint64(0)
			materialBaseUnitName := ""

			// 从关联的物料信息中获取数据
			if item.Material != nil {
				materialName = item.Material.Name
				materialBaseUnitUuid = item.Material.UnitUuid
			}

			// 这里需要额外查询单位名称，因为关联查询比较复杂
			if materialBaseUnitUuid > 0 {
				var unitName string
				materialBaseUnitName = item.Material.Unit.Name
				materialBaseUnitName = unitName
			}

			recordMap[key] = &OutboundRecord{
				Uuid:                 item.Uuid,
				WarehouseUuid:        item.WarehouseUuid,
				MaterialUuid:         item.MaterialUuid,
				TotalNum:             item.Num,
				Valuation:            item.Material.Valuation,
				SupplierUuid:         item.Material.SupplierUuid,
				MaterialName:         materialName,
				MaterialBaseUnitUuid: materialBaseUnitUuid,
				MaterialBaseUnitName: materialBaseUnitName,
			}
		}
	}

	// 转换为切片返回
	var records []*OutboundRecord
	for _, record := range recordMap {
		records = append(records, record)
	}

	return records, nil
}
```

**查询条件**（`GetSaleOrderMaterialByCreateTimeBetween`）:
```46:51:main/app/repository/sale_order_material.go
// GetSaleOrderMaterialByCreateTimeBetween 获取某时间范围内的销售订单原料
func (r *SaleOrderMaterialRepoImpl) GetSaleOrderMaterialByCreateTimeBetween(startTime, endTime int64) ([]*model.SaleOrderMaterial, error) {
	var saleOrderMaterials []*model.SaleOrderMaterial
	err := r.db.Model(&model.SaleOrderMaterial{}).Preload("Material.Unit").Where("create_time BETWEEN ? AND ? AND delete_time = 0 AND is_summarized = 0", startTime, endTime).Find(&saleOrderMaterials).Error
	return saleOrderMaterials, errors.WithMessage(err)
}
```

**关键点**:
- 查询条件：`create_time BETWEEN startTime AND endTime`
- 只查询未删除的记录：`delete_time = 0`
- 只查询未统计的记录：`is_summarized = 0`
- 预加载物料和单位信息
- **按仓库和物料分组汇总**：相同仓库+物料的记录合并，数量累加

**数据结构**:
```255:266:main/app/tasks/daily_sales_outbound_summary.go
// OutboundRecord 出库记录汇总
type OutboundRecord struct {
	Uuid                 uint64  `json:"uuid"` // 出库记录ID
	WarehouseUuid        uint64  `json:"warehouse_uuid"`
	MaterialUuid         uint64  `json:"material_uuid"`
	TotalNum             float64 `json:"total_num"`
	Valuation            float64 `json:"valuation"` // 估值率
	SupplierUuid         uint64  `json:"supplier_uuid"`
	MaterialName         string  `json:"material_name"`
	MaterialBaseUnitUuid uint64  `json:"material_base_unit_uuid"`
	MaterialBaseUnitName string  `json:"material_base_unit_name"`
}
```

### 7. 保存汇总记录 (`saveOutboundSummaryRecords`)

```268:317:main/app/tasks/daily_sales_outbound_summary.go
// saveOutboundSummaryRecords 保存出库汇总记录到ttpos_warehouse_in_out_log表
func (t *DailySalesOutboundSummaryTask) saveOutboundSummaryRecords(companyUuid uint64, records []*OutboundRecord, openingHours string) error {
	db := t.dbm.GetDB(companyUuid)

	// 生成出库单号
	orderNo, err := t.generateOrderNo(companyUuid)
	if err != nil {
		return fmt.Errorf("生成出库单号失败: %w", err)
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		// 使用 repository 方法创建记录
		warehouseLogRepo := repository.NewWarehouseInOutLogRepo(tx)

		uuids := make([]uint64, 0)
		for _, record := range records {
			uuids = append(uuids, record.Uuid)
			logRecord := &model.WarehouseInOutLog{
				LogType:              constant.WarehouseInOutLogLogTypeOut, // 出库
				Scene:                constant.WarehouseInOutLogSceneSale,  // 销售出库
				WarehouseUuid:        record.WarehouseUuid,
				MaterialUuid:         record.MaterialUuid,
				MaterialName:         record.MaterialName,
				MaterialBaseUnitUuid: record.MaterialBaseUnitUuid,
				MaterialBaseUnitName: record.MaterialBaseUnitName,
				Num:                  record.TotalNum,
				Price:                record.Valuation,
				Amount:               decimal.NewFromFloat(record.TotalNum).Mul(decimal.NewFromFloat(record.Valuation)).Round(2).InexactFloat64(),
				SupplierUuid:         record.SupplierUuid,
				OrderNo:              orderNo,
				OpeningHours:         openingHours,
			}

			if err := warehouseLogRepo.Create(logRecord); err != nil {
				logger.Logger.Error("保存出库记录失败: warehouse_uuid=%d, material_uuid=%d, error=%v", zap.Uint64("warehouse_uuid", record.WarehouseUuid), zap.Uint64("material_uuid", record.MaterialUuid), zap.Error(err))
				continue
			}
		}
		// 更新销售订单原料的统计状态
		saleOrderMaterialRepo := repository.NewSaleOrderMaterialRepo(tx)
		if err := saleOrderMaterialRepo.UpdateSaleOrderMaterialIsSummarized(uuids); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
```

**写入逻辑**:
1. **生成出库单号**: 格式 `SSCK + YYYYMMDD + 4位序号`（例如：`SSCK202510230001`）
2. **事务处理**: 使用数据库事务确保数据一致性
3. **批量写入**: 遍历汇总记录，逐条写入 `ttpos_warehouse_in_out_log` 表
4. **更新统计状态**: 将已统计的销售订单原料的 `is_summarized` 字段更新为 `1`

**写入字段映射**:

| 字段 | 值 | 说明 |
|------|-----|------|
| `log_type` | `1` | 出库（`WarehouseInOutLogLogTypeOut`） |
| `scene` | `1` | 销售出库（`WarehouseInOutLogSceneSale`） |
| `warehouse_uuid` | `record.WarehouseUuid` | 仓库ID |
| `material_uuid` | `record.MaterialUuid` | 物料ID |
| `material_name` | `record.MaterialName` | 物料名称 |
| `material_base_unit_uuid` | `record.MaterialBaseUnitUuid` | 物料基准单位ID |
| `material_base_unit_name` | `record.MaterialBaseUnitName` | 物料基准单位名称 |
| `num` | `record.TotalNum` | 出库数量（已汇总） |
| `price` | `record.Valuation` | 单价（估值率） |
| `amount` | `TotalNum × Valuation` | 金额（保留2位小数） |
| `supplier_uuid` | `record.SupplierUuid` | 供应商ID |
| `order_no` | `orderNo` | 出库单号 |
| `opening_hours` | `openingHours` | 营业时段（格式：`YYYYMMDD HH:mm-HH:mm`） |

### 8. 生成出库单号 (`generateOrderNo`)

```319:357:main/app/tasks/daily_sales_outbound_summary.go
// generateOrderNo 生成出库单号，格式：SSCK + YYYYMMDD + 4位序号
func (t *DailySalesOutboundSummaryTask) generateOrderNo(companyUuid uint64) (string, error) {
	db := t.dbm.GetDB(companyUuid)

	// 获取今天的日期
	now := time.Now()
	dateStr := now.Format("20060102")

	// 使用 repository 方法查询今天已有的SSCK开头的出库单号
	warehouseLogRepo := repository.NewWarehouseInOutLogRepo(db)
	opts := []repository.DBOption{
		warehouseLogRepo.WhereLogType(1), // 出库
		warehouseLogRepo.WhereScene(1),   // 销售出库
		warehouseLogRepo.WhereCreateTimeBetween(
			int(now.Truncate(24*time.Hour).Unix()),
			int(now.Truncate(24*time.Hour).Add(24*time.Hour).Add(-time.Second).Unix()),
		),
	}
	existingLogs, err := warehouseLogRepo.GetWarehouseInOutLogs(opts...)
	if err != nil {
		return "", err
	}

	// 解析最大序号
	sequence := 1
	for _, log := range existingLogs {
		if len(log.OrderNo) >= 16 {
			seqStr := log.OrderNo[12:16] // 取最后4位作为序号
			if seq, err := strconv.Atoi(seqStr); err == nil && seq >= sequence {
				sequence = seq + 1
			}
		}
	}

	// 生成4位序号，不足补0
	sequenceStr := fmt.Sprintf("%04d", sequence)

	return "SSCK" + dateStr + sequenceStr, nil
}
```

**单号格式**:
- 前缀: `SSCK`（销售出库）
- 日期: `YYYYMMDD`（8位）
- 序号: `0001-9999`（4位，不足补0）
- 示例: `SSCK202510230001`

**生成逻辑**:
- 查询当天所有销售出库记录
- 解析已有单号的序号部分（最后4位）
- 取最大序号 + 1 作为新序号

## 数据表结构

### `ttpos_warehouse_in_out_log` 表

```3:33:main/app/model/warehouse_in_out_log.go
// WarehouseInOutLog 仓库出入库记录表 `ttpos_warehouse_in_out_log`
type WarehouseInOutLog struct {
	BaseModel
	LogType              int     `json:"log_type" gorm:"type:int;default:0;comment:日志类型,0-入库 1-出库"`
	Scene                int     `json:"scene" gorm:"type:int;default:0;comment:场景,0-采购入库 1-销售出库 2-发货出库 3-盘盈入库 4-盘亏出库 20-在途入库 21-在途出库"`
	WarehouseUuid        uint64  `json:"warehouse_uuid" gorm:"type:bigint;default:0;comment:仓库ID"`
	MaterialUuid         uint64  `json:"material_uuid" gorm:"type:bigint;default:0;comment:物品ID"`
	MaterialName         string  `json:"material_name" gorm:"type:text;default:'';comment:物品名称JSON,记录当时物品名称"`
	MaterialBaseUnitUuid uint64  `json:"material_base_unit_uuid" gorm:"type:bigint;default:0;comment:物品基准单位ID"`
	MaterialBaseUnitName string  `json:"material_base_unit_name" gorm:"type:text;default:'';comment:物品基准单位名称"`
	Num                  float64 `json:"num" gorm:"type:decimal(22,4);default:0;comment:数量"`
	Price                float64 `json:"price" gorm:"type:decimal(22,4);default:0;comment:单价，物品基准单位单价"`
	Amount               float64 `json:"amount" gorm:"type:decimal(22,4);default:0;comment:金额,单价*数量"`
	SupplierUuid         uint64  `json:"supplier_uuid" gorm:"type:bigint;default:0;comment:供应商ID"`
	SupplierErpCode      string  `json:"supplier_erp_code" gorm:"type:varchar(500);default:'';comment:供应商ERP编码"`
	SupplierName         string  `json:"supplier_name" gorm:"type:text;comment:供应商名称"`
	OrderNo              string  `json:"order_no" gorm:"type:varchar(255);default:'';comment:单据编号"`
	OtherOrgUuid         uint64  `json:"other_org_uuid" gorm:"type:bigint;default:0;comment:对方机构ID"`
	OtherOrgType         uint64  `json:"other_org_type" gorm:"type:bigint;default:0;comment:对方机构类型 0:供应商 1:客户"`
	OtherOrgName         string  `json:"other_org_name" gorm:"type:text;comment:对方机构名称"`
	OpeningHours         string  `json:"opening_hours" gorm:"type:varchar(255);default:'';comment:营业时段,仅用于Scene销售出库的场景"`

	// 关联模型
	Material  *Material  `gorm:"foreignKey:MaterialUuid;references:Uuid" json:"material,omitempty"`
	Supplier  *Supplier  `gorm:"foreignKey:SupplierUuid;references:Uuid" json:"supplier,omitempty"`
	Warehouse *Warehouse `gorm:"foreignKey:WarehouseUuid;references:Uuid" json:"warehouse,omitempty"`
}
```

## 关键常量

```go
// 日志类型
WarehouseInOutLogLogTypeOut = 1 // 出库

// 场景类型
WarehouseInOutLogSceneSale = 1 // 销售出库
```

## 去重机制

### 1. 分布式锁
- 使用 Redis 分布式锁防止多节点重复执行
- 锁标识: `DailySalesOutboundSummaryLock`
- 如果获取锁耗时 > 1 秒，说明其他节点正在处理，当前节点跳过

### 2. 营业时段去重
- 通过 `opening_hours` 字段判断该营业时段是否已统计
- `opening_hours` 格式: `YYYYMMDD HH:mm-HH:mm`（例如：`20251023 09:00-22:00`）
- 查询条件:
  - `log_type = 1`（出库）
  - `scene = 1`（销售出库）
  - `opening_hours = {openingYearHours}`

### 3. 数据源去重
- 只查询 `is_summarized = 0` 的记录（未统计）
- 统计完成后，将相关记录的 `is_summarized` 更新为 `1`

## 数据汇总规则

### 分组维度
- **仓库** (`warehouse_uuid`)
- **物料** (`material_uuid`)

### 汇总字段
- **数量** (`num`): 累加相同仓库+物料的数量
- **单价** (`price`): 使用物料的估值率（`valuation`）
- **金额** (`amount`): `数量 × 单价`（保留2位小数）

### 数据来源
- 表: `ttpos_sale_order_material`（销售订单原料表）
- 条件:
  - `create_time BETWEEN startTime AND endTime`
  - `delete_time = 0`
  - `is_summarized = 0`

## 时区处理

- 使用门店配置的时区（`company_setting.timezone`）
- 默认时区: `Asia/Shanghai`
- 营业时间的开始和结束时间戳基于门店时区计算

## 错误处理

1. **Panic 恢复**: 使用 `defer recover()` 捕获 panic
2. **单门店失败不影响其他门店**: 使用 `continue` 跳过失败的门店
3. **事务回滚**: 写入失败时自动回滚
4. **日志记录**: 记录所有关键步骤和错误信息

## 性能优化

1. **批量查询**: 一次性查询所有门店
2. **预加载关联**: 使用 `Preload` 预加载物料和单位信息
3. **分组汇总**: 在内存中按仓库+物料分组，减少数据库写入次数
4. **事务批量写入**: 使用事务批量写入，提高性能

## 注意事项

1. **ERP 功能**: 只处理开启了 ERP 功能的门店
2. **营业时间配置**: 如果门店未配置营业时间，使用默认值 `00:00-23:59`
3. **时区差异**: 不同门店可能使用不同时区，需要分别处理
4. **数据一致性**: 使用事务确保写入和状态更新的原子性
5. **单号唯一性**: 出库单号在同一天内唯一，通过序号递增保证

## 相关文件

- 任务实现: `main/app/tasks/daily_sales_outbound_summary.go`
- 模型定义: `main/app/model/warehouse_in_out_log.go`
- 仓库实现: `main/app/repository/warehouse_in_out_log.go`
- 销售订单原料仓库: `main/app/repository/sale_order_material.go`
- 定时任务注册: `main/command/root.go`

## 执行日志示例

```
开始执行每日销售出库汇总定时任务
分布式锁耗时: 10ms
找到 5 个门店，开始检查营业结束时间
时区: Asia/Shanghai
openingHours: 09:00-22:00,营业开始时间: 2025-10-23 09:00:00(1729641600),营业结束时间: 2025-10-23 22:00:00(1729681200)
当前时间: 2025-10-23 22:05:00(1729681500), 营业结束时间: 2025-10-23 22:00:00(1729681200)
门店 测试门店 已到达营业结束时间，开始统计销售出库记录
门店 测试门店 销售出库汇总完成，共处理 15 条记录
每日销售出库汇总定时任务执行完成
```

---

**最后更新**: 2025-01-27  
**维护者**: TTPOS Team

