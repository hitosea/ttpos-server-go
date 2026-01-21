# TTPOS 库存实现详细评估报告

> **评估日期**: 2026-01-18
> **评估范围**: 主服务库存模块 (main/app/modules/inventory/ 及相关模块)
> **评估维度**: 架构设计、代码质量、性能、可维护性、业务逻辑、安全性

---

## 目录

1. [评估总览](#评估总览)
2. [架构设计评估](#架构设计评估)
3. [代码质量评估](#代码质量评估)
4. [性能评估](#性能评估)
5. [业务逻辑评估](#业务逻辑评估)
6. [数据一致性评估](#数据一致性评估)
7. [安全性评估](#安全性评估)
8. [可维护性评估](#可维护性评估)
9. [与 ERPNext 集成评估](#与-erpnext-集成评估)
10. [综合评分](#综合评分)
11. [改进建议](#改进建议)

---

## 评估总览

### 整体评价

TTPOS 库存系统采用了**较为先进的 DDD 架构**和**策略模式**，代码分层清晰，业务逻辑封装良好。但在**数据一致性保障**、**缓存管理**、**错误处理**和**性能优化**方面存在较多可改进空间。

### 评分概览

| 维度 | 评分 | 说明 |
|------|------|------|
| 架构设计 | ⭐⭐⭐⭐ (8/10) | DDD + 策略模式优秀，但缺少 CQRS |
| 代码质量 | ⭐⭐⭐⭐ (7.5/10) | 整体规范，但存在一些设计缺陷 |
| 性能 | ⭐⭐⭐ (6/10) | 缓存机制不完善，N+1 查询风险 |
| 业务逻辑 | ⭐⭐⭐⭐ (8/10) | 策略模式封装良好，但逻辑复杂 |
| 数据一致性 | ⭐⭐⭐ (6.5/10) | 事务使用正确，但缺少补偿机制 |
| 安全性 | ⭐⭐⭐⭐ (7/10) | 锁机制完善，但缺少审计日志 |
| 可维护性 | ⭐⭐⭐⭐ (7.5/10) | 分层清晰，但文档不足 |
| ERPNext 集成 | ⭐⭐⭐ (6/10) | 同步机制简陋，错误处理不足 |

**综合评分**: ⭐⭐⭐⭐ (7.1/10)

---

## 架构设计评估

### ✅ 优势

#### 1. DDD 分层架构清晰

```
应用服务层 (Application Service)
    → 领域服务层 (Domain Service)
        → 领域策略 (Domain Strategy)
            → 仓储层 (Repository)
                → 数据模型 (Model)
```

**优点**:
- 职责明确，应用服务处理缓存和编排，领域服务处理核心业务
- 符合单一职责原则
- 易于单元测试（每层可独立测试）

**示例**:
```go
// 应用服务层 - 处理缓存
func (s *ProductInventoryAppService) GetProductInventory(ctx, productBomUuid) {
    // 1. 查缓存
    if cached, exists := s.cache.Get(cacheKey); exists { ... }

    // 2. 调用领域服务
    inventory := s.domainService.GetProductInventory(ctx, productBomUuid)

    // 3. 写缓存
    s.cache.Set(cacheKey, inventory, TTL)
}

// 领域服务层 - 处理业务逻辑
func (s *productInventoryDomainService) GetProductInventory(ctx, productBomUuid) {
    // 选择策略
    strategy := s.selectStrategy(productBom)

    // 计算库存
    return strategy.CalculateInventory(ctx, productBom)
}
```

#### 2. 策略模式应用优雅

**6 种策略自动适配不同商品类型**:

```go
strategies := map[string]IInventoryStrategy{
    "bom_card":            BomCardProductInventoryStrategy,
    "flavor_materials":    FlavorMaterialsProductInventoryStrategy,
    "sauce_bom_card":      SauceBomCardProductInventoryStrategy,
    "sauce_materials":     SauceMaterialsProductInventoryStrategy,
    "sauce_non_bom_card":  SauceNonBomCardProductInventoryStrategy,
    "non_bom_card":        NonBomCardProductInventoryStrategy,
}
```

**优点**:
- 避免了大量的 if-else 判断
- 符合开闭原则（新增策略无需修改现有代码）
- 每个策略职责单一，易于理解和测试

**代码示例**:
```go
// 策略选择逻辑清晰
if productBom.HasProductBomCard() {
    strategy = s.strategies["bom_card"]
} else if len(productBom.FlavorMaterials) > 0 {
    strategy = s.strategies["flavor_materials"]
} else if productBom.IsSauce() {
    // 小料的三种策略
    ...
} else {
    strategy = s.strategies["non_bom_card"]
}
```

#### 3. 值对象 (Value Object) 封装

```go
type Stock struct {
    value float64  // 不可变
}

// 方法：Add, Subtract, GreaterThan, LessThan...
```

**优点**:
- 不可变对象，线程安全
- 自动处理负数（转为 0）
- 封装库存相关的业务规则（如保留两位小数）

#### 4. 接口定义规范

```go
type IProductInventoryDomainService interface {
    GetProductInventory(ctx, productBomUuid) (float64, error)
    GetProductInventoriesBatch(ctx, productBomUuids) (map[uint64]float64, error)
    GetProductPackageInventory(ctx, productPackageUuid, opts) (float64, error)
    GetProductPackageInventoriesBatch(ctx, productPackageUuids, opts) (map[uint64]float64, error)
}
```

**优点**:
- 便于 Mock 测试
- 支持依赖注入
- 符合依赖倒置原则

### ❌ 劣势

#### 1. 缺少 CQRS 分离

**问题**:
- 查询操作（GetProductInventory）和命令操作（UpdateStock）混在同一个服务中
- 查询路径没有针对性优化（如使用只读副本、物化视图）

**影响**:
- 查询性能受写操作影响
- 无法针对查询场景做专门优化（如使用 Elasticsearch）

**改进建议**:
```go
// 查询服务 - 可以使用只读副本、缓存、搜索引擎
type IInventoryQueryService interface {
    GetProductInventory(productBomUuid) (*InventoryQueryResult, error)
    GetProductInventoriesBatch(productBomUuids) (map[uint64]*InventoryQueryResult, error)
}

// 命令服务 - 使用主库，保证一致性
type IInventoryCommandService interface {
    ReduceStock(saleBillUuid, items) error
    AddStock(saleBillUuid, items) error
}
```

#### 2. 缺少聚合根 (Aggregate Root) 概念

**问题**:
- `ProductBom` 和 `Material` 之间的关系没有明确的聚合根
- 库存变更时，可能直接修改多个实体（违反聚合根原则）

**示例问题**:
```go
// ❌ 直接修改多个实体
productBoms[uuid].StockNum -= quantity
materials[uuid].ReduceStockNum += quantity
```

**改进建议**:
```go
// ✅ 通过聚合根统一管理
type StockAggregateRoot struct {
    productBoms  map[uint64]*ProductBom
    materials    map[uint64]*Material
}

func (a *StockAggregateRoot) ReduceStock(saleBillUuid, items) error {
    // 统一验证和修改
    // 保证聚合内部一致性
}
```

#### 3. 缺少领域事件 (Domain Event)

**问题**:
- 库存变更后，没有发布领域事件
- 其他模块（如通知、报表）无法感知库存变更

**当前实现**:
```go
func ReduceStock(...) {
    // 扣减库存
    UpdateProductBoms(...)
    UpdateMaterialsStockNum(...)

    // ❌ 缺少领域事件
    // ❌ 外卖平台同步是硬编码的副作用
    utils.Go(func() {
        takeoutService.SyncMenuChanges(...)  // 耦合
    })
}
```

**改进建议**:
```go
func ReduceStock(...) {
    // 扣减库存
    UpdateProductBoms(...)
    UpdateMaterialsStockNum(...)

    // ✅ 发布领域事件
    domainEvents.Publish(&StockReducedEvent{
        SaleBillUuid: saleBillUuid,
        Items:        items,
        Timestamp:    time.Now(),
    })
}

// 其他模块订阅事件
domainEvents.Subscribe("StockReduced", func(event *StockReducedEvent) {
    takeoutService.SyncMenuChanges(...)  // 解耦
    notificationService.NotifyLowStock(...)
    analyticsService.RecordStockChange(...)
})
```

#### 4. 缺少防腐层 (Anti-Corruption Layer)

**问题**:
- 直接使用 ERPNext 的数据结构（如 `ItemCode`, `Warehouse`）
- ERPNext 接口变更会影响 TTPOS 核心业务

**示例**:
```go
// ❌ 直接使用 ERPNext 数据结构
erpReq.Items = append(erpReq.Items, &stock.SaveStockReconciliationItem{
    ItemCode:  item.Material.Code,       // ERPNext 概念
    Warehouse: warehouse.ErpCode,        // ERPNext 概念
    Qty:       item.CountedQuantity,
})
```

**改进建议**:
```go
// ✅ 使用防腐层适配器
type ErpAdapter interface {
    SubmitStockReconciliation(domainReconciliation *DomainReconciliation) error
}

type ErpNextAdapter struct {
    client stock.StockServiceClient
}

func (a *ErpNextAdapter) SubmitStockReconciliation(domainReconciliation *DomainReconciliation) error {
    // 将领域模型转换为 ERPNext 模型
    erpReq := a.toDomainToErpModel(domainReconciliation)
    return a.client.SaveStockReconciliation(erpReq)
}
```

---

## 代码质量评估

### ✅ 优势

#### 1. 代码规范性良好

**命名规范**:
- 接口以 `I` 开头：`IProductInventoryDomainService`
- 实现类以具体名称结尾：`ProductInventoryAppService`
- 方法名清晰表达意图：`GetProductInventory`, `ReduceStock`

**代码组织**:
- 按层级分包：`application/`, `domain/service/`, `infrastructure/persistence/`
- 一个文件一个职责

#### 2. 错误处理使用 errors.WithMessage

```go
if err != nil {
    return 0, errors.WithMessage(err, "查询商品BOM失败")
}
```

**优点**:
- 保留错误堆栈
- 提供上下文信息

#### 3. 使用选项模式 (Functional Options)

```go
// 定义选项
type GetProductPackageInventoryOption struct {
    Strategy IProductPackageInventoryStrategy
}

func WithStrategy(strategy IProductPackageInventoryStrategy) func(option *GetProductPackageInventoryOption) {
    return func(option *GetProductPackageInventoryOption) {
        option.Strategy = strategy
    }
}

// 使用选项
inventory := s.GetProductPackageInventory(ctx, packageUuid, WithStrategy(minStrategy))
```

**优点**:
- 参数可选，向后兼容
- 代码可读性高

#### 4. Repository 模式封装良好

```go
type IProductBomRepository interface {
    FindByUuid(ctx, uuid) (*model.ProductBom, error)
    FindByUuids(ctx, uuids) ([]*model.ProductBom, error)
    FindByProductPackageUuid(ctx, packageUuid) ([]*model.ProductBom, error)
}
```

**优点**:
- 数据访问逻辑与业务逻辑分离
- 便于切换数据源（如从 MySQL 切换到 MongoDB）

### ❌ 劣势

#### 1. 缺少单元测试

**问题**:
- 核心业务逻辑（策略计算）缺少单元测试
- 无法保证重构后的行为一致性

**影响**:
- 代码修改风险高
- 难以发现边界条件的 Bug

**改进建议**:
```go
// 应该有的测试
func TestBomCardProductInventoryStrategy_CalculateInventory(t *testing.T) {
    tests := []struct {
        name     string
        bom      *model.ProductBom
        expected float64
        wantErr  bool
    }{
        {
            name: "成本卡控制开启，材料充足",
            bom: &model.ProductBom{
                UseBomCardStock: constant.Yes,
                ProductBomCard: &model.ProductBomCard{
                    RelatedMaterials: []*model.RelatedMaterial{
                        {Material: &model.Material{...}, Usage: 10},
                    },
                },
            },
            expected: 100,  // 材料库存 1000 / 用量 10 = 100
            wantErr:  false,
        },
        // ... 更多测试用例
    }
}
```

#### 2. 魔法数字 (Magic Number) 未定义为常量

**问题**:
```go
// ❌ 魔法数字
return 99999999  // 无限库存

// ❌ 魔法 TTL
ProductInventoryCacheTTL = 30 * time.Second
```

**改进建议**:
```go
const (
    InfiniteStock = 99999999
    ProductInventoryCacheTTL = 30 * time.Second
    ProductPackageInventoryCacheTTL = 30 * time.Second
)
```

#### 3. 类型断言缺少错误处理

**问题**:
```go
// ❌ 直接 panic
bom, ok := productBom.(*model.ProductBom)
if !ok {
    return 0, errors.New("商品BOM类型错误")  // ✅ 这个还好
}

// ❌ 更严重的问题
productBom := bomInterface.(*model.ProductBom)  // 直接断言，可能 panic
```

**影响**:
- 运行时 panic 导致服务崩溃

**改进建议**:
```go
// ✅ 始终使用安全断言
productBom, ok := bomInterface.(*model.ProductBom)
if !ok {
    logger.Logger.Error("类型断言失败", zap.Any("bomInterface", bomInterface))
    continue  // 或 return error
}
```

#### 4. 错误日志缺少上下文信息

**问题**:
```go
// ❌ 缺少关键上下文
logger.Logger.Error("计算商品库存失败", zap.Error(err))

// ❌ 没有记录导致错误的输入
logger.Logger.Error("查询商品BOM失败", zap.Error(err))
```

**改进建议**:
```go
// ✅ 记录完整上下文
logger.Logger.Error("计算商品库存失败",
    zap.Uint64("company_uuid", ctx.GetCompanyUuid()),
    zap.Uint64("product_bom_uuid", productBom.Uuid),
    zap.String("product_type", productBom.Type),
    zap.Bool("has_bom_card", productBom.HasProductBomCard()),
    zap.Error(err),
)
```

#### 5. 硬编码的缓存键格式

**问题**:
```go
// ❌ 硬编码
cacheKey := fmt.Sprintf("product_inventory:%d:%d", companyUuid, productBomUuid)
```

**问题**:
- 缓存键格式散落在代码各处
- 修改格式需要全局搜索替换

**改进建议**:
```go
// ✅ 定义缓存键生成器
type CacheKeyGenerator struct{}

func (g *CacheKeyGenerator) ProductInventoryKey(companyUuid, productBomUuid uint64) string {
    return fmt.Sprintf("inventory:product:%d:%d", companyUuid, productBomUuid)
}

func (g *CacheKeyGenerator) ProductPackageInventoryKey(companyUuid, packageUuid uint64) string {
    return fmt.Sprintf("inventory:package:%d:%d", companyUuid, packageUuid)
}
```

#### 6. Repository 接口过于庞大

**问题**:
```go
type IWarehouseItemRepo interface {
    // 17 个方法！！！
    Create(...)
    CreateBatch(...)
    Update(...)
    Delete(...)
    GetByUuid(...)
    GetByWarehouseUuid(...)
    GetByMaterialUuid(...)
    // ... 还有 10 个
}
```

**影响**:
- 违反接口隔离原则
- Mock 测试时需要实现所有方法

**改进建议**:
```go
// ✅ 拆分为多个小接口
type IWarehouseItemReader interface {
    GetByUuid(uuid) (*model.WarehouseItem, error)
    GetByWarehouseUuid(warehouseUuid) ([]*model.WarehouseItem, error)
}

type IWarehouseItemWriter interface {
    Create(item *model.WarehouseItem) error
    Update(item *model.WarehouseItem) error
}

type IWarehouseItemStockUpdater interface {
    UpdateStock(uuid, stock, reservedStock float64) error
    AddStock(uuid, stock float64) error
    ReduceStock(uuid, stock float64) error
}
```

---

## 性能评估

### ✅ 优势

#### 1. 批量查询优化

```go
// ✅ 一次性查询所有 BOM
productBomInterfaces := s.productBomRepo.FindByUuids(ctx, productBomUuids)

// ✅ 避免循环查询
for _, uuid := range productBomUuids {
    // ❌ 如果这样写会有 N+1 查询问题
    // productBom := s.productBomRepo.FindByUuid(ctx, uuid)
}
```

**优点**:
- 减少数据库往返次数
- 提高查询性能

#### 2. Redis 缓存机制

```go
// 缓存键: product_inventory:{company_uuid}:{product_bom_uuid}
// TTL: 30 秒
```

**优点**:
- 减少数据库查询压力
- 提高响应速度

#### 3. 数据库连接池复用

```go
// 使用 DBManager 管理连接池
db := ctx.GetDB()
```

**优点**:
- 避免频繁创建连接
- 提高并发性能

### ❌ 劣势

#### 1. 缓存 TTL 过短且不可配置

**问题**:
```go
// ❌ 硬编码 30 秒
ProductInventoryCacheTTL = 30 * time.Second
```

**影响**:
- 库存数据变化不频繁时，30 秒过短导致缓存命中率低
- 无法根据业务场景调整

**数据支持**:
- 假设库存查询 QPS = 1000
- 缓存未命中率 = 50%（因为 TTL 过短）
- 实际数据库查询 = 500 QPS
- 如果 TTL = 5 分钟，缓存未命中率 ≈ 10%，数据库查询 = 100 QPS

**改进建议**:
```go
// ✅ 可配置的 TTL
type InventoryCacheConfig struct {
    ProductInventoryTTL        time.Duration  // 默认 5 分钟
    ProductPackageInventoryTTL time.Duration  // 默认 5 分钟
    EnableCache                bool           // 开关
}

// 根据业务场景动态调整
config := &InventoryCacheConfig{
    ProductInventoryTTL: 5 * time.Minute,  // 普通商品
}

// 高频变化的商品可以使用更短的 TTL
if product.IsHotSale() {
    config.ProductInventoryTTL = 30 * time.Second
}
```

#### 2. 缓存失效策略不完善

**问题**:
```go
// ❌ 缓存失效逻辑缺失
func ReduceStock(...) {
    // 扣减库存
    UpdateProductBoms(...)
    UpdateMaterialsStockNum(...)

    // ❌ 没有清理缓存
    // 导致查询到的库存数据是旧的（最多 30 秒）
}
```

**影响**:
- 库存扣减后，查询仍然返回旧数据
- 可能导致超卖（虽然概率不高，因为 TTL 只有 30 秒）

**改进建议**:
```go
// ✅ 库存变更后立即失效缓存
func ReduceStock(ctx, db, saleBillUuid) {
    // 扣减库存
    UpdateProductBoms(ProductBomsList)
    UpdateMaterialsStockNum(...)

    // ✅ 清理相关缓存
    for _, productBom := range ProductBomsList {
        inventoryAppService.InvalidateProductInventoryCache(
            ctx.GetCompanyUuid(),
            productBom.Uuid,
        )

        // ✅ 同时清理商品包缓存
        inventoryAppService.InvalidateProductPackageInventoryCache(
            ctx,
            productBom.ProductPackageUuid,
        )
    }
}
```

#### 3. 存在 N+1 查询风险

**问题**:
```go
// 查询 ProductBom 时，可能存在 N+1 查询
productBom := productBomRepo.FindByUuid(ctx, uuid)

// ❌ 如果没有预加载，会触发多次查询
// 1. SELECT * FROM ttpos_product_bom WHERE uuid = ?
// 2. SELECT * FROM ttpos_product_bom_card WHERE product_bom_uuid = ?
// 3. SELECT * FROM ttpos_related_material WHERE product_bom_uuid = ?
// 4. SELECT * FROM ttpos_material WHERE uuid IN (...)
// 5. SELECT * FROM ttpos_warehouse_item WHERE material_uuid IN (...)
```

**影响**:
- 批量查询时，数据库查询次数 = 1 + N × 4
- 性能严重下降

**改进建议**:
```go
// ✅ 使用 Preload 预加载关联数据
productBoms := productBomRepo.FindByUuids(ctx, uuids,
    WithPreload("ProductBomCard"),
    WithPreload("ProductBomCard.RelatedMaterials"),
    WithPreload("ProductBomCard.RelatedMaterials.Material"),
    WithPreload("ProductBomCard.RelatedMaterials.Material.WarehouseItems"),
    WithPreload("FlavorMaterials"),
    WithPreload("FlavorMaterials.Material"),
    WithPreload("FlavorMaterials.Material.WarehouseItems"),
)

// 或者使用 Join 查询
```

#### 4. 缓存雪崩风险

**问题**:
- 所有缓存使用相同的 TTL（30 秒）
- 大量缓存可能同时失效
- 导致数据库瞬时压力激增

**改进建议**:
```go
// ✅ 添加随机偏移
ttl := ProductInventoryCacheTTL + time.Duration(rand.Intn(10)) * time.Second

// ✅ 或者使用分层 TTL
type CacheTier int

const (
    TierHot     CacheTier = iota  // 热门商品，TTL = 5 分钟
    TierWarm                       // 一般商品，TTL = 10 分钟
    TierCold                       // 冷门商品，TTL = 30 分钟
)

func (s *ProductInventoryAppService) GetCacheTTL(productBom *model.ProductBom) time.Duration {
    if productBom.ActualSaleNum > 1000 {  // 热门商品
        return 5 * time.Minute
    } else if productBom.ActualSaleNum > 100 {
        return 10 * time.Minute
    }
    return 30 * time.Minute
}
```

#### 5. 缓存穿透风险

**问题**:
```go
// ❌ 不存在的商品会穿透缓存
inventory := s.GetProductInventory(ctx, 99999999)  // 不存在的 UUID

// 每次都查询数据库
// 如果恶意请求大量不存在的 UUID，数据库压力激增
```

**改进建议**:
```go
// ✅ 缓存空值
func (s *ProductInventoryAppService) GetProductInventory(ctx, productBomUuid) {
    // 1. 查缓存
    if cached, exists := s.cache.Get(cacheKey); exists {
        if cached == "NULL" {  // ✅ 缓存空值
            return 0, errors.New("商品不存在")
        }
        return parseCached(cached), nil
    }

    // 2. 查数据库
    inventory, err := s.domainService.GetProductInventory(ctx, productBomUuid)
    if err != nil {
        // ✅ 缓存空值，TTL 更短（如 1 分钟）
        s.cache.Set(cacheKey, "NULL", 1*time.Minute)
        return 0, err
    }

    // 3. 缓存正常值
    s.cache.Set(cacheKey, inventory, ProductInventoryCacheTTL)
    return inventory, nil
}
```

#### 6. 数据库连接未使用索引优化

**潜在问题**:
```go
// 查询语句
db.Model(&model.WarehouseItem{}).
    Where("material_uuid = ?", materialUuid).
    Where("warehouse_uuid = ?", warehouseUuid).
    Update("stock", gorm.Expr("stock + ?", addStockNum))
```

**需要确认的索引**:
- `idx_warehouse_item_material_uuid`
- `idx_warehouse_item_warehouse_uuid`
- **最佳**：`idx_warehouse_item_material_warehouse` (复合索引)

**改进建议**:
```sql
-- ✅ 创建复合索引
CREATE INDEX idx_warehouse_item_material_warehouse
ON ttpos_warehouse_item(material_uuid, warehouse_uuid);

-- ✅ 添加覆盖索引（包含 stock 字段）
CREATE INDEX idx_warehouse_item_material_warehouse_stock
ON ttpos_warehouse_item(material_uuid, warehouse_uuid, stock);
```

---

## 业务逻辑评估

### ✅ 优势

#### 1. 策略模式封装业务规则清晰

**6 种策略分别处理**:
- 有成本卡商品
- 规格关联材料
- 小料（3 种子策略）
- 无成本卡商品

**优点**:
- 每个策略职责单一
- 业务规则一目了然
- 易于扩展新策略

#### 2. 木桶原理计算库存准确

```go
func (card *ProductBomCard) CalculateExpectedProductionNum() float64 {
    minInventory := math.MaxFloat64

    for _, relatedMaterial := range card.RelatedMaterials {
        materialStock := relatedMaterial.Material.GetStockNum()
        usage := relatedMaterial.Usage

        if usage > 0 {
            expectedNum := materialStock / usage
            minInventory = math.Min(minInventory, expectedNum)  // 取最小值
        }
    }

    return minInventory
}
```

**优点**:
- 符合实际业务逻辑（最稀缺的材料决定生产数量）
- 避免超卖

#### 3. 负库存处理灵活

```go
if material.AllowNegativeStock == constant.Yes {
    return constant.ProductBomInfiniteStock  // 99999999
}
```

**优点**:
- 支持不同的库存管理策略
- 允许负库存的商品不受库存限制

### ❌ 劣势

#### 1. 库存计算逻辑过于复杂

**问题**:
```go
func (s *bomCardProductInventoryStrategy) calculateNonBomCardInventory(bom) {
    // ❌ 多层嵌套判断
    if bom.IsSoldOut == constant.ProductStatusSaleOut {
        return 0
    }

    if bom.IsOpenStockBool() {
        if bom.HasProductBomCard() && bom.ProductBomCard != nil {
            hasMaterialNotAllowNegativeStock := false
            for _, material := range bom.ProductBomCard.RelatedMaterials {
                if material.Material.AllowNegativeStock == constant.No {
                    hasMaterialNotAllowNegativeStock = true
                    break
                }
            }
            if hasMaterialNotAllowNegativeStock {
                bomCardInventory := bom.ProductBomCard.CalculateExpectedProductionNum(...)
                return math.Min(bom.StockNum, bomCardInventory)
            }
        }
        return bom.StockNum
    }

    if bom.HasProductBomCard() && bom.ProductBomCard != nil {
        for _, material := range bom.ProductBomCard.RelatedMaterials {
            if material.Material.AllowNegativeStock == constant.No {
                return bom.ProductBomCard.CalculateExpectedProductionNum(...)
            }
        }
    }

    return constant.ProductBomInfiniteStock
}
```

**影响**:
- 认知负担高
- 容易出错
- 难以维护

**改进建议**:
```go
// ✅ 拆分为多个小方法
func (s *bomCardProductInventoryStrategy) calculateNonBomCardInventory(bom) {
    if bom.IsSoldOut {
        return 0
    }

    if bom.IsOpenStock {
        return s.calculateOpenStockInventory(bom)
    }

    if s.shouldUseBomCardInventory(bom) {
        return s.calculateBomCardInventory(bom)
    }

    return InfiniteStock
}

func (s *bomCardProductInventoryStrategy) calculateOpenStockInventory(bom) float64 {
    if !s.hasMaterialNotAllowNegativeStock(bom) {
        return bom.StockNum
    }

    bomCardInventory := s.calculateBomCardInventory(bom)
    return math.Min(bom.StockNum, bomCardInventory)
}

func (s *bomCardProductInventoryStrategy) shouldUseBomCardInventory(bom) bool {
    return bom.HasProductBomCard() && s.hasMaterialNotAllowNegativeStock(bom)
}

func (s *bomCardProductInventoryStrategy) hasMaterialNotAllowNegativeStock(bom) bool {
    if !bom.HasProductBomCard() {
        return false
    }

    for _, material := range bom.ProductBomCard.RelatedMaterials {
        if material.Material.AllowNegativeStock == constant.No {
            return true
        }
    }
    return false
}
```

#### 2. Material.StockNum 字段废弃但仍存在

**问题**:
```go
type Material struct {
    StockNum float64  // ❌ 已废弃，实际库存在 WarehouseItem.Stock

    WarehouseItems []*WarehouseItem  // ✅ 真正的库存存储位置
}
```

**影响**:
- 容易误用 `Material.StockNum`
- 数据冗余
- 可能导致数据不一致

**改进建议**:
```go
// ✅ 方案 1: 删除废弃字段（需要数据迁移）
type Material struct {
    // StockNum float64  // 删除

    WarehouseItems []*WarehouseItem
}

// ✅ 方案 2: 添加弃用注释
type Material struct {
    StockNum float64  `gorm:"column:stock_num;comment:已废弃，请使用 WarehouseItem.Stock" json:"-"`  // 不在 JSON 中暴露

    WarehouseItems []*WarehouseItem
}

// ✅ 方案 3: 使用计算属性
func (m *Material) GetTotalStock() float64 {
    total := 0.0
    for _, item := range m.WarehouseItems {
        total += item.Stock
    }
    return total
}
```

#### 3. 库存查询和库存更新使用不同的数据结构

**问题**:
```go
// 查询库存
inventory := productBom.ProductBomCard.CalculateExpectedProductionNum()
// 使用: Material.WarehouseItems[].Stock

// 更新库存
UpdateMaterialsStockNum(materialUuid, warehouseUuid, addStockNum)
// 直接更新: WarehouseItem.Stock

// ❌ 两个操作使用了不同的数据路径
```

**影响**:
- 代码理解困难
- 容易导致逻辑不一致

**改进建议**:
```go
// ✅ 统一使用仓储模式
type IWarehouseItemRepository interface {
    GetStock(materialUuid, warehouseUuid) (float64, error)
    UpdateStock(materialUuid, warehouseUuid, delta float64) error
}

// 查询库存
stock := warehouseItemRepo.GetStock(materialUuid, warehouseUuid)

// 更新库存
warehouseItemRepo.UpdateStock(materialUuid, warehouseUuid, -10)  // 扣减
warehouseItemRepo.UpdateStock(materialUuid, warehouseUuid, +10)  // 增加
```

#### 4. 商品包库存策略默认使用最大值不合理

**问题**:
```go
// ❌ 默认策略: MaxProductPackageInventoryStrategy
// 假设商品包有 3 个规格: A(库存 100), B(库存 50), C(库存 10)
// 返回库存: 100 (最大值)

// 实际业务: 用户只能购买 10 份（受 C 的限制）
// 但系统显示可以购买 100 份 → 可能导致超卖
```

**影响**:
- 库存显示不准确
- 用户体验差（下单后发现库存不足）

**改进建议**:
```go
// ✅ 默认应该使用最小值策略
defaultPackageInventoryStrategy: NewMinProductPackageInventoryStrategy()

// 或者根据业务场景选择
// - 套餐类商品: 使用最小值（所有规格都需要）
// - 可选规格商品: 使用最大值（任选其一）
```

#### 5. 盘点单提交到 ERPNext 后无法回滚

**问题**:
```go
func ApproveStockReconciliation(...) {
    // 1. 提交到 ERPNext (SaveStockReconciliation)
    erpResp := erpSrv.SubmitStockReconciliation(...)

    // 2. 审核 ERPNext (SubmitStockReconciliation)
    erpSrv.ApproveStockReconciliation(...)

    // 3. 更新 TTPOS 状态
    stockReconciliationRepo.UpdateStockReconciliation(...)

    // ❌ 如果第 3 步失败，ERPNext 已经更新库存，但 TTPOS 状态未更新
    // ❌ 数据不一致，且无法回滚
}
```

**影响**:
- 分布式事务问题
- ERPNext 和 TTPOS 库存数据不一致

**改进建议**:
```go
// ✅ 使用 Saga 模式或补偿事务
func ApproveStockReconciliation(...) {
    // 1. 记录操作日志
    saga := NewSaga()
    saga.AddStep("SubmitToErpNext", submitToErpNext, rollbackSubmitToErpNext)
    saga.AddStep("ApproveInErpNext", approveInErpNext, rollbackApproveInErpNext)
    saga.AddStep("UpdateTtposStatus", updateTtposStatus, rollbackUpdateTtposStatus)

    // 2. 执行 Saga
    err := saga.Execute()
    if err != nil {
        // 自动执行补偿事务
        saga.Compensate()
    }
}

// 补偿事务
func rollbackApproveInErpNext(...) error {
    return erpSrv.RejectStockReconciliation(...)  // 取消 ERPNext 盘点单
}
```

---

## 数据一致性评估

### ✅ 优势

#### 1. 使用数据库事务保证原子性

```go
repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    // 更新出库单状态
    warehouseFormRepo.UpdateWarehouseOutFormItemRecordsReduceStock(saleBillUuid)

    // 更新商品库存
    repository.NewProductBomRepo(tx).UpdateProductBoms(ProductBomsList)

    // 更新材料库存
    base.NewMaterialRepo(tx).UpdateMaterialsStockNum(...)

    return nil
})
```

**优点**:
- 所有操作在一个事务中
- 保证全部成功或全部失败
- 避免部分更新导致数据不一致

#### 2. 使用分布式锁防止并发问题

```go
// ✅ 加锁防止并发扣减
lock.NewSystemLock().LockUuid(saleBillUuid)
defer lock.NewSystemLock().UnlockUuid(saleBillUuid)

// 扣减库存
ReduceStock(...)
```

**优点**:
- 防止多个送厨事件并发扣减库存
- 保证库存扣减的顺序性

#### 3. 使用 GORM Expr 进行原子更新

```go
// ✅ 使用数据库表达式
db.Model(&model.WarehouseItem{}).
    Where("material_uuid = ?", materialUuid).
    Where("warehouse_uuid = ?", warehouseUuid).
    Update("stock", gorm.Expr("stock + ?", addStockNum))

// 避免了读-改-写的竞态条件
// ❌ 如果这样写会有问题:
// stock := warehouseItem.Stock
// stock += addStockNum
// warehouseItem.Stock = stock
// db.Save(warehouseItem)
```

**优点**:
- 原子操作，避免并发问题
- 性能更好（单条 SQL）

### ❌ 劣势

#### 1. 缓存与数据库不一致

**问题**:
```go
// 扣减库存
ReduceStock(...) {
    // 更新数据库
    UpdateProductBoms(...)
    UpdateMaterialsStockNum(...)

    // ❌ 没有清理缓存
    // 缓存中仍然是旧数据（最多 30 秒）
}

// 查询库存
GetProductInventory(...) {
    // 从缓存获取
    if cached, exists := s.cache.Get(cacheKey); exists {
        return cached  // ❌ 返回旧数据
    }
}
```

**影响**:
- 库存扣减后，查询仍然返回扣减前的库存
- 虽然 TTL 只有 30 秒，但仍有风险

**真实场景**:
1. 用户 A 查询商品库存 → 100（缓存）
2. 用户 B 下单扣减 50
3. 用户 A 再次查询 → 100（缓存未失效）
4. 用户 A 下单 100 → 超卖

**改进建议**:
```go
// ✅ 库存变更后立即失效缓存
func ReduceStock(ctx, db, saleBillUuid) {
    affectedBomUuids := []uint64{}

    // 扣减库存
    for _, productBom := range ProductBomsList {
        affectedBomUuids = append(affectedBomUuids, productBom.Uuid)
    }

    // 事务提交成功后，立即失效缓存
    for _, bomUuid := range affectedBomUuids {
        inventoryAppService.InvalidateProductInventoryCache(ctx.GetCompanyUuid(), bomUuid)
    }
}
```

#### 2. 跨库事务问题

**问题**:
```go
// TTPOS 更新库存
db.Transaction(func(tx *gorm.DB) error {
    UpdateProductBoms(...)
    UpdateMaterialsStockNum(...)
    return nil
})

// ERPNext 更新库存（独立系统）
erpSrv.ApproveStockReconciliation(...)

// ❌ 两个系统的事务无法保证一致性
```

**影响**:
- TTPOS 更新成功，ERPNext 更新失败 → 数据不一致
- ERPNext 更新成功，TTPOS 更新失败 → 数据不一致

**改进建议**:
```go
// ✅ 使用最终一致性 + 补偿机制
type ReconciliationSyncJob struct {
    ReconciliationUuid string
    Status             string  // pending, syncing, synced, failed
    RetryCount         int
    MaxRetry           int
    LastError          string
}

func SyncToErpNext(reconciliation *StockReconciliation) error {
    job := &ReconciliationSyncJob{
        ReconciliationUuid: reconciliation.Uuid,
        Status:             "syncing",
        MaxRetry:           3,
    }

    // 保存同步任务
    jobRepo.Create(job)

    // 异步同步
    utils.Go(func() {
        for job.RetryCount < job.MaxRetry {
            err := erpSrv.ApproveStockReconciliation(...)
            if err == nil {
                job.Status = "synced"
                jobRepo.Update(job)
                return
            }

            job.RetryCount++
            job.LastError = err.Error()
            jobRepo.Update(job)

            time.Sleep(time.Duration(job.RetryCount) * 10 * time.Second)  // 指数退避
        }

        // 重试失败，标记为失败，人工介入
        job.Status = "failed"
        jobRepo.Update(job)

        // 发送告警
        alertService.SendAlert("ERPNext 同步失败", job)
    })

    return nil
}
```

#### 3. 库存扣减没有幂等性保证

**问题**:
```go
func ReduceStock(...) {
    // ❌ 如果重复调用，会重复扣减库存
    warehouseOutFormItems := warehouseFormRepo.GetWarehouseOutFormItemNotProcessed(saleBillUuid)

    for _, item := range warehouseOutFormItems {
        item.ReduceStock = constant.WarehouseOutFormItemReduceStockSuccess
        ProductBoms[item.ProductBomUuid].StockNum -= item.Num
    }

    // 更新数据库
    UpdateProductBoms(ProductBomsList)
}
```

**潜在问题**:
- 如果 `UpdateWarehouseOutFormItemRecordsReduceStock` 失败
- 下次重试时，会再次扣减库存

**真实场景**:
1. 送厨事件触发，扣减库存 50
2. 更新出库单状态时数据库连接断开
3. 事务回滚，但事件会重新发布
4. 再次扣减库存 50 → 实际扣减了 100

**改进建议**:
```go
// ✅ 使用数据库状态保证幂等性
func ReduceStock(...) {
    // 1. 查询未处理的出库单
    warehouseOutFormItems := warehouseFormRepo.GetWarehouseOutFormItemNotProcessed(saleBillUuid)

    if len(warehouseOutFormItems) == 0 {
        // ✅ 已经处理过，直接返回
        return nil
    }

    // 2. 在事务中：标记状态 + 扣减库存
    repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
        // 先标记状态
        warehouseFormRepo.UpdateWarehouseOutFormItemRecordsReduceStock(saleBillUuid)

        // 再扣减库存
        UpdateProductBoms(ProductBomsList)
        UpdateMaterialsStockNum(...)

        return nil
    })
}
```

#### 4. 缺少库存变更审计日志

**问题**:
```go
// ❌ 库存变更没有记录详细日志
UpdateProductBoms(...)
UpdateMaterialsStockNum(...)
```

**影响**:
- 无法追溯库存变更历史
- 出现库存异常时难以排查
- 缺少审计依据

**改进建议**:
```go
// ✅ 记录库存变更日志
type StockChangeLog struct {
    Uuid            uint64
    CompanyUuid     uint64
    MaterialUuid    uint64
    WarehouseUuid   uint64
    ChangeType      string    // reduce, add, adjust
    ChangeDelta     float64   // 变更量（正/负）
    BeforeStock     float64   // 变更前库存
    AfterStock      float64   // 变更后库存
    RelatedBillUuid uint64    // 关联单据
    RelatedBillType string    // 单据类型: sale_bill, purchase_order, stock_reconciliation
    OperatorUuid    uint64
    CreateTime      int64
}

func UpdateMaterialsStockNum(materialUuid, warehouseUuid, delta) error {
    // 1. 查询当前库存
    warehouseItem := warehouseItemRepo.GetByWarehouseAndMaterial(warehouseUuid, materialUuid)
    beforeStock := warehouseItem.Stock

    // 2. 更新库存
    err := db.Model(&model.WarehouseItem{}).
        Where("material_uuid = ?", materialUuid).
        Where("warehouse_uuid = ?", warehouseUuid).
        Update("stock", gorm.Expr("stock + ?", delta)).Error

    // 3. 记录变更日志
    stockChangeLog := &StockChangeLog{
        MaterialUuid:  materialUuid,
        WarehouseUuid: warehouseUuid,
        ChangeType:    getChangeType(delta),
        ChangeDelta:   delta,
        BeforeStock:   beforeStock,
        AfterStock:    beforeStock + delta,
        CreateTime:    time.Now().Unix(),
    }
    stockChangeLogRepo.Create(stockChangeLog)

    return err
}
```

#### 5. 库存警报机制缺失

**问题**:
- 库存低于安全库存时，没有自动告警
- 库存为负时（允许负库存的商品），没有监控

**改进建议**:
```go
// ✅ 库存变更后检查并发送告警
func UpdateMaterialsStockNum(materialUuid, warehouseUuid, delta) error {
    // 更新库存
    ...

    // 查询更新后的库存
    warehouseItem := warehouseItemRepo.GetByWarehouseAndMaterial(warehouseUuid, materialUuid)
    material := materialRepo.GetByUuid(materialUuid)

    // 检查库存告警
    if material.SafetyStock != nil && warehouseItem.Stock < *material.SafetyStock {
        // 发送低库存告警
        alertService.SendLowStockAlert(&LowStockAlert{
            MaterialUuid:  materialUuid,
            MaterialName:  material.Name,
            CurrentStock:  warehouseItem.Stock,
            SafetyStock:   *material.SafetyStock,
            WarehouseUuid: warehouseUuid,
        })
    }

    if warehouseItem.Stock < 0 && material.AllowNegativeStock == constant.No {
        // 发送负库存告警（不应该出现）
        alertService.SendNegativeStockAlert(...)
    }

    return nil
}
```

---

## 安全性评估

### ✅ 优势

#### 1. 分布式锁防止并发扣减

```go
lock.NewSystemLock().LockUuid(saleBillUuid)
defer lock.NewSystemLock().UnlockUuid(saleBillUuid)
```

**优点**:
- 防止多个事件并发修改同一订单的库存
- 保证库存扣减的顺序性

#### 2. 数据库事务保证一致性

```go
repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
    // 所有操作在一个事务中
})
```

**优点**:
- 防止部分更新导致数据不一致

#### 3. 参数校验

```go
if len(bomQuantityMap) == 0 {
    return nil, nil
}
```

**优点**:
- 防止空参数导致的异常

### ❌ 劣势

#### 1. 缺少权限校验

**问题**:
```go
// ❌ 直接删除盘点单，没有权限校验
func DeleteStockReconciliation(ctx, req) error {
    stockReconciliation := stockReconciliationRepo.GetStockReconciliation(...)

    // ❌ 没有检查当前用户是否有权限删除
    // ❌ 没有检查盘点单是否属于当前公司

    stockReconciliationRepo.DeleteStockReconciliation(req.Uuid)
}
```

**影响**:
- 越权删除其他公司的盘点单
- 低权限用户可以删除高权限创建的盘点单

**改进建议**:
```go
// ✅ 添加权限校验
func DeleteStockReconciliation(ctx, req) error {
    // 1. 检查盘点单是否存在且属于当前公司
    stockReconciliation := stockReconciliationRepo.GetStockReconciliation(...)
    if stockReconciliation.CompanyUuid != ctx.GetCompanyUuid() {
        return errors.New("无权限删除其他公司的盘点单")
    }

    // 2. 检查用户权限
    if !ctx.GetUser().HasPermission("stock_reconciliation:delete") {
        return errors.New("无权限删除盘点单")
    }

    // 3. 检查盘点单状态
    if stockReconciliation.Status != constant.StockReconciliationStatusDraft {
        return errors.New("只能删除草稿状态的盘点单")
    }

    // 4. 删除
    stockReconciliationRepo.DeleteStockReconciliation(req.Uuid)
}
```

#### 2. SQL 注入风险（GORM 已防护，但需注意）

**潜在问题**:
```go
// ✅ GORM 使用参数化查询，安全
db.Where("uuid = ?", uuid).First(&item)

// ❌ 如果手动拼接 SQL，有注入风险
db.Raw("SELECT * FROM ttpos_material WHERE uuid = " + uuid)
```

**改进建议**:
- 统一使用 GORM 的参数化查询
- 禁止手动拼接 SQL

#### 3. 敏感信息日志泄露风险

**问题**:
```go
// ❌ 可能记录敏感信息
logger.Logger.Error("查询商品失败", zap.Any("productBom", productBom))

// 如果 productBom 包含敏感信息（如成本价），会被记录到日志
```

**改进建议**:
```go
// ✅ 只记录必要信息
logger.Logger.Error("查询商品失败",
    zap.Uint64("product_bom_uuid", productBom.Uuid),
    zap.String("product_type", productBom.Type),
    zap.Error(err),
)

// ✅ 定义日志安全结构
type ProductBomLogSafe struct {
    Uuid       uint64
    Type       string
    IsOpenStock bool
    // 不包含敏感字段: Price, Cost, StockNum
}

logger.Logger.Error("查询商品失败", zap.Any("productBom", ProductBomLogSafe{...}))
```

#### 4. 缺少审计日志

**问题**:
- 库存扣减、增加、盘点等关键操作没有审计日志
- 无法追溯谁在什么时候做了什么操作

**改进建议**:
```go
type AuditLog struct {
    Uuid         uint64
    CompanyUuid  uint64
    OperatorUuid uint64
    Action       string  // reduce_stock, add_stock, approve_reconciliation
    ResourceType string  // product_bom, material, stock_reconciliation
    ResourceUuid uint64
    OldValue     string  // JSON 序列化的旧值
    NewValue     string  // JSON 序列化的新值
    IP           string
    UserAgent    string
    CreateTime   int64
}

func ReduceStock(...) {
    // 记录审计日志
    auditLog := &AuditLog{
        CompanyUuid:  ctx.GetCompanyUuid(),
        OperatorUuid: ctx.GetStaff().Uuid,
        Action:       "reduce_stock",
        ResourceType: "product_bom",
        ResourceUuid: productBom.Uuid,
        OldValue:     toJSON(productBom.StockNum),
        NewValue:     toJSON(productBom.StockNum - quantity),
        IP:           ctx.GetClientIP(),
        CreateTime:   time.Now().Unix(),
    }
    auditLogRepo.Create(auditLog)
}
```

#### 5. 缺少限流和防刷机制

**问题**:
- 没有对库存查询接口限流
- 恶意请求可能导致缓存穿透和数据库压力

**改进建议**:
```go
// ✅ 使用限流中间件
func RateLimitMiddleware() gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Limit(100), 100)  // 每秒 100 个请求

    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(429, gin.H{"error": "请求过于频繁，请稍后再试"})
            c.Abort()
            return
        }
        c.Next()
    }
}

// 对库存查询接口限流
router.GET("/inventory/:uuid", RateLimitMiddleware(), GetInventory)
```

---

## 可维护性评估

### ✅ 优势

#### 1. 代码分层清晰

```
application/    # 应用服务层
domain/         # 领域层
    service/    # 领域服务
    valueobject/ # 值对象
infrastructure/ # 基础设施层
    persistence/ # 持久化
```

**优点**:
- 职责明确
- 易于定位代码
- 支持独立测试

#### 2. 接口与实现分离

```go
type IProductInventoryDomainService interface { ... }
type productInventoryDomainService struct { ... }
```

**优点**:
- 便于 Mock 测试
- 支持依赖注入
- 易于替换实现

#### 3. 使用 Repository 模式

**优点**:
- 数据访问逻辑集中管理
- 便于切换数据源

### ❌ 劣势

#### 1. 缺少完整的 API 文档

**问题**:
- 缺少 Swagger 文档
- 接口参数、响应格式、错误码没有统一说明

**改进建议**:
```go
// ✅ 添加 Swagger 注释
// @Summary 获取商品库存
// @Description 查询商品规格或小料的库存数量
// @Tags 库存管理
// @Accept json
// @Produce json
// @Param product_bom_uuid path uint64 true "商品规格UUID"
// @Success 200 {object} resp.ProductInventoryResp
// @Failure 400 {object} resp.ErrorResp
// @Failure 500 {object} resp.ErrorResp
// @Router /api/v1/shop/inventory/{product_bom_uuid} [get]
func GetProductInventory(c *gin.Context) { ... }
```

#### 2. 缺少代码注释和设计文档

**问题**:
```go
// ❌ 复杂逻辑缺少注释
func (s *bomCardProductInventoryStrategy) calculateNonBomCardInventory(bom) {
    if bom.IsSoldOut == constant.ProductStatusSaleOut {
        return 0
    }

    if bom.IsOpenStockBool() {
        if bom.HasProductBomCard() && bom.ProductBomCard != nil {
            hasMaterialNotAllowNegativeStock := false
            // ... 10 行代码
        }
    }
    // ... 没有说明为什么这样设计
}
```

**改进建议**:
```go
// ✅ 添加详细注释
// calculateNonBomCardInventory 计算无成本卡控制商品的库存
//
// 库存计算优先级:
// 1. 如果标记为售罄 (IsSoldOut=1)，返回 0
// 2. 如果开启可售量 (IsOpenStock=1)
//    a. 如果有成本卡且材料不允许负库存，返回 min(可售量, 成本卡库存)
//    b. 否则返回可售量
// 3. 如果有成本卡且材料不允许负库存，返回成本卡计算的库存
// 4. 默认返回无限库存 (99999999)
//
// 注意: 这是一个特殊需求，当成本卡中的材料不允许负库存时，
//      即使没有开启成本卡控制，也要限制库存
func (s *bomCardProductInventoryStrategy) calculateNonBomCardInventory(bom) {
    // ...
}
```

#### 3. 错误码不统一

**问题**:
```go
// ❌ 直接返回错误信息，没有错误码
return errors.New("商品不存在")
return errors.New("盘点单状态不允许修改")
```

**影响**:
- 前端无法根据错误码做不同处理
- 多语言支持困难

**改进建议**:
```go
// ✅ 定义错误码
const (
    ErrCodeProductNotFound              = 10001
    ErrCodeReconciliationStatusInvalid  = 10002
    ErrCodeStockInsufficient            = 10003
    // ...
)

// ✅ 错误结构
type BizError struct {
    Code    int
    Message string
    Details any
}

func (e *BizError) Error() string {
    return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// 使用
return &BizError{
    Code:    ErrCodeProductNotFound,
    Message: "商品不存在",
    Details: map[string]any{"product_bom_uuid": productBomUuid},
}
```

#### 4. 缺少性能监控和链路追踪

**问题**:
- 没有记录库存查询的耗时
- 无法分析性能瓶颈

**改进建议**:
```go
// ✅ 添加性能监控
func (s *ProductInventoryAppService) GetProductInventory(ctx, productBomUuid) {
    start := time.Now()
    defer func() {
        metrics.RecordDuration("inventory.get_product_inventory", time.Since(start))
    }()

    // ...
}

// ✅ 使用 OpenTelemetry 链路追踪
func (s *ProductInventoryAppService) GetProductInventory(ctx, productBomUuid) {
    ctx, span := tracer.Start(ctx, "GetProductInventory")
    defer span.End()

    span.SetAttributes(
        attribute.Int64("product_bom_uuid", int64(productBomUuid)),
    )

    // ...
}
```

#### 5. 缺少配置管理

**问题**:
```go
// ❌ 硬编码配置
ProductInventoryCacheTTL = 30 * time.Second
InfiniteStock = 99999999
```

**改进建议**:
```go
// ✅ 使用配置文件
type InventoryConfig struct {
    CacheTTL        time.Duration `yaml:"cache_ttl"`
    InfiniteStock   float64       `yaml:"infinite_stock"`
    EnableCache     bool          `yaml:"enable_cache"`
}

// 从配置文件加载
config := loadConfig("inventory.yaml")
```

---

## 与 ERPNext 集成评估

### ✅ 优势

#### 1. 使用 gRPC 通信

**优点**:
- 性能高（二进制协议）
- 类型安全（Protobuf）
- 支持流式传输

#### 2. 多租户支持

```go
// 使用元数据传递站点编码
func WithSiteCode(ctx, siteCode) context.Context {
    md := metadata.Pairs("site_code", siteCode)
    return metadata.NewOutgoingContext(ctx, md)
}
```

**优点**:
- 支持多个 ERPNext 站点
- 路由灵活

### ❌ 劣势

#### 1. 同步机制过于简单

**问题**:
```go
// ❌ 直接同步调用，没有重试机制
erpResp := erpSrv.SubmitStockReconciliation(...)
if err != nil {
    return err  // 直接返回错误
}
```

**影响**:
- 网络抖动导致同步失败
- ERPNext 服务不可用时，TTPOS 功能受影响

**改进建议**:
```go
// ✅ 使用异步队列 + 重试机制
type ErpSyncTask struct {
    TaskType   string  // submit_reconciliation, approve_reconciliation
    Payload    string  // JSON 序列化的请求参数
    RetryCount int
    MaxRetry   int
    Status     string  // pending, processing, success, failed
}

func SubmitStockReconciliationAsync(reconciliation *StockReconciliation) error {
    // 1. 保存到数据库
    task := &ErpSyncTask{
        TaskType: "submit_reconciliation",
        Payload:  toJSON(reconciliation),
        MaxRetry: 3,
        Status:   "pending",
    }
    taskRepo.Create(task)

    // 2. 发送到消息队列
    mq.Publish("erp_sync_queue", task)

    return nil
}

// 消费者
func ConsumeErpSyncTask(task *ErpSyncTask) {
    for task.RetryCount < task.MaxRetry {
        err := erpSrv.SubmitStockReconciliation(...)
        if err == nil {
            task.Status = "success"
            taskRepo.Update(task)
            return
        }

        task.RetryCount++
        task.Status = "processing"
        taskRepo.Update(task)

        time.Sleep(time.Duration(task.RetryCount) * 10 * time.Second)
    }

    task.Status = "failed"
    taskRepo.Update(task)
}
```

#### 2. 缺少数据一致性保障

**问题**:
- TTPOS 和 ERPNext 的数据可能不一致
- 没有对账机制

**改进建议**:
```go
// ✅ 定期对账
func ReconcileWithErpNext() error {
    // 1. 从 ERPNext 拉取库存数据
    erpStocks := erpSrv.GetBin(warehouseErpCode)

    // 2. 对比 TTPOS 库存
    for _, erpStock := range erpStocks {
        warehouseItem := warehouseItemRepo.GetByMaterialCode(erpStock.ItemCode)

        if warehouseItem.Stock != erpStock.ActualQty {
            // 记录差异
            diffLog := &StockDiffLog{
                MaterialCode:  erpStock.ItemCode,
                TtposStock:    warehouseItem.Stock,
                ErpnextStock:  erpStock.ActualQty,
                Diff:          warehouseItem.Stock - erpStock.ActualQty,
                CreateTime:    time.Now().Unix(),
            }
            diffLogRepo.Create(diffLog)

            // 发送告警
            alertService.SendStockDiffAlert(diffLog)
        }
    }

    return nil
}
```

#### 3. 错误处理不完善

**问题**:
```go
// ❌ 只检查错误码，不处理具体错误类型
if result.Code != "0" {
    return errors.New(result.Message)
}
```

**影响**:
- 无法区分临时错误（网络抖动）和永久错误（数据错误）
- 无法做针对性重试

**改进建议**:
```go
// ✅ 区分错误类型
type ErpError struct {
    Code      string
    Message   string
    IsRetryable bool  // 是否可重试
}

func HandleErpError(result *stock.CommonResp) error {
    switch result.Code {
    case "0":
        return nil
    case "NETWORK_ERROR", "TIMEOUT":
        return &ErpError{Code: result.Code, Message: result.Message, IsRetryable: true}
    case "INVALID_DATA", "DUPLICATE_ENTRY":
        return &ErpError{Code: result.Code, Message: result.Message, IsRetryable: false}
    default:
        return &ErpError{Code: result.Code, Message: result.Message, IsRetryable: false}
    }
}

// 重试逻辑
func SubmitStockReconciliationWithRetry(...) error {
    maxRetry := 3
    for i := 0; i < maxRetry; i++ {
        err := erpSrv.SubmitStockReconciliation(...)
        if err == nil {
            return nil
        }

        erpErr, ok := err.(*ErpError)
        if !ok || !erpErr.IsRetryable {
            return err  // 不可重试的错误，直接返回
        }

        time.Sleep(time.Duration(i+1) * time.Second)  // 重试延迟
    }
    return errors.New("重试次数已达上限")
}
```

#### 4. 缺少降级策略

**问题**:
- ERPNext 不可用时，TTPOS 盘点功能完全不可用

**改进建议**:
```go
// ✅ 降级策略：允许离线盘点
func ApproveStockReconciliation(...) {
    // 1. 尝试同步到 ERPNext
    err := erpSrv.ApproveStockReconciliation(...)

    if err != nil {
        // 2. ERPNext 不可用，启用降级模式
        if isErpUnavailable(err) {
            // 标记为"待同步"状态
            stockReconciliationRepo.UpdateStockReconciliation(uuid, map[string]any{
                "status":       constant.StockReconciliationStatusPendingSync,
                "sync_retry":   0,
                "sync_error":   err.Error(),
            })

            // 发送告警
            alertService.SendErpSyncFailedAlert(...)

            return nil  // 不影响用户操作
        }

        return err
    }

    // 3. 正常同步成功
    stockReconciliationRepo.UpdateStockReconciliation(uuid, map[string]any{
        "status": constant.StockReconciliationStatusApproved,
    })
}

// 后台任务：重试同步
func RetryPendingSyncReconciliations() {
    pendingReconciliations := stockReconciliationRepo.GetByStatus(constant.StockReconciliationStatusPendingSync)

    for _, reconciliation := range pendingReconciliations {
        err := erpSrv.ApproveStockReconciliation(...)
        if err == nil {
            stockReconciliationRepo.UpdateStockReconciliation(reconciliation.Uuid, map[string]any{
                "status": constant.StockReconciliationStatusApproved,
            })
        } else {
            stockReconciliationRepo.UpdateStockReconciliation(reconciliation.Uuid, map[string]any{
                "sync_retry": reconciliation.SyncRetry + 1,
                "sync_error": err.Error(),
            })
        }
    }
}
```

---

## 综合评分

### 评分矩阵

| 维度 | 分数 | 权重 | 加权分 | 关键问题 |
|------|------|------|--------|----------|
| 架构设计 | 8.0 | 20% | 1.6 | 缺少 CQRS、聚合根、领域事件、防腐层 |
| 代码质量 | 7.5 | 15% | 1.125 | 缺少单元测试、魔法数字、错误日志不完善 |
| 性能 | 6.0 | 15% | 0.9 | 缓存 TTL 过短、失效策略不完善、N+1 查询风险 |
| 业务逻辑 | 8.0 | 15% | 1.2 | 库存计算逻辑过于复杂、数据模型冗余 |
| 数据一致性 | 6.5 | 15% | 0.975 | 缓存与数据库不一致、跨库事务、缺少幂等性 |
| 安全性 | 7.0 | 10% | 0.7 | 缺少权限校验、审计日志、限流机制 |
| 可维护性 | 7.5 | 5% | 0.375 | 缺少 API 文档、代码注释、错误码不统一 |
| ERPNext 集成 | 6.0 | 5% | 0.3 | 同步机制简单、缺少重试、降级策略、对账 |

**综合评分**: **7.1 / 10** (71%)

### 评级说明

- **优秀 (8-10)**: 架构先进，代码质量高，性能优秀
- **良好 (7-8)**: 架构合理，代码规范，性能良好，但有改进空间 ⭐ **当前等级**
- **中等 (5-7)**: 基本可用，但存在较多问题
- **较差 (3-5)**: 问题较多，需要重构
- **极差 (0-3)**: 严重问题，不可用

---

## 改进建议

### 🔥 高优先级（必须修复）

#### 1. 缓存一致性问题

**问题**: 库存变更后缓存未失效，导致查询到旧数据

**解决方案**:
```go
func ReduceStock(ctx, db, saleBillUuid) {
    affectedBomUuids := []uint64{}

    // 扣减库存
    repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
        for _, productBom := range ProductBomsList {
            affectedBomUuids = append(affectedBomUuids, productBom.Uuid)
        }

        UpdateProductBoms(ProductBomsList)
        UpdateMaterialsStockNum(...)

        return nil
    })

    // 事务提交成功后，立即失效缓存
    for _, bomUuid := range affectedBomUuids {
        inventoryAppService.InvalidateProductInventoryCache(ctx.GetCompanyUuid(), bomUuid)
        inventoryAppService.InvalidateProductPackageInventoryCache(ctx, productBom.ProductPackageUuid)
    }
}
```

**预期效果**: 消除缓存不一致风险，库存数据实时准确

#### 2. 幂等性保证

**问题**: 库存扣减没有幂等性，重复调用会重复扣减

**解决方案**:
```go
func ReduceStock(...) {
    // 1. 查询未处理的出库单（带状态检查）
    warehouseOutFormItems := warehouseFormRepo.GetWarehouseOutFormItemNotProcessed(saleBillUuid)

    if len(warehouseOutFormItems) == 0 {
        return nil  // 已处理，直接返回
    }

    // 2. 在同一个事务中：标记状态 + 扣减库存
    repository.CommonRepo.Transaction(db, func(tx *gorm.DB) error {
        // 先标记状态（防止并发）
        warehouseFormRepo.UpdateWarehouseOutFormItemRecordsReduceStock(tx, saleBillUuid)

        // 再扣减库存
        UpdateProductBoms(tx, ProductBomsList)
        UpdateMaterialsStockNum(tx, ...)

        return nil
    })
}
```

**预期效果**: 防止重复扣减库存，保证数据准确性

#### 3. 添加单元测试

**解决方案**:
```go
// 测试每个策略的库存计算逻辑
func TestBomCardProductInventoryStrategy(t *testing.T) { ... }
func TestFlavorMaterialsProductInventoryStrategy(t *testing.T) { ... }

// 测试批量查询
func TestGetProductInventoriesBatch(t *testing.T) { ... }

// 测试缓存逻辑
func TestInventoryCache(t *testing.T) { ... }

// 测试并发扣减
func TestConcurrentReduceStock(t *testing.T) { ... }
```

**预期效果**: 覆盖率 > 80%，保证重构安全

### ⚡ 中优先级（建议改进）

#### 4. 优化缓存策略

**改进方案**:
- TTL 可配置（默认 5 分钟）
- 使用分层 TTL（热门商品 5 分钟，普通商品 10 分钟）
- 添加随机偏移避免缓存雪崩
- 缓存空值防止穿透

#### 5. 添加审计日志

**改进方案**:
```go
type StockChangeLog struct {
    Uuid            uint64
    MaterialUuid    uint64
    ChangeType      string  // reduce, add, adjust
    ChangeDelta     float64
    BeforeStock     float64
    AfterStock      float64
    RelatedBillUuid uint64
    OperatorUuid    uint64
    CreateTime      int64
}
```

#### 6. 改进 ERPNext 集成

**改进方案**:
- 使用异步队列 + 重试机制
- 添加降级策略（离线盘点）
- 定期对账（检查 TTPOS 与 ERPNext 库存差异）
- 区分可重试和不可重试错误

### 💡 低优先级（优化建议）

#### 7. 重构复杂的库存计算逻辑

**改进方案**: 拆分为多个小方法，提高可读性

#### 8. 添加性能监控

**改进方案**: 使用 OpenTelemetry 记录耗时和链路

#### 9. 统一错误码

**改进方案**: 定义错误码常量，便于前端处理

#### 10. 添加 API 文档

**改进方案**: 使用 Swagger 生成 API 文档

---

## 总结

### 核心优势

1. ✅ **架构设计良好**: DDD 分层清晰，策略模式应用优雅
2. ✅ **代码规范性高**: 命名清晰，分包合理
3. ✅ **业务逻辑封装好**: 6 种策略处理不同商品类型
4. ✅ **并发控制完善**: 分布式锁 + 数据库事务

### 核心问题

1. ❌ **缓存一致性**: 库存变更后缓存未失效
2. ❌ **幂等性缺失**: 重复调用会重复扣减库存
3. ❌ **缺少单元测试**: 代码修改风险高
4. ❌ **性能优化不足**: 缓存 TTL 过短，N+1 查询风险
5. ❌ **ERPNext 集成简陋**: 缺少重试、降级、对账机制

### 改进路线图

#### 第一阶段（1-2 周）- 修复关键问题

- [ ] 修复缓存一致性问题
- [ ] 添加幂等性保证
- [ ] 编写核心业务逻辑单元测试

#### 第二阶段（2-3 周）- 性能优化

- [ ] 优化缓存策略（可配置 TTL、分层缓存）
- [ ] 添加数据库索引优化
- [ ] 解决 N+1 查询问题

#### 第三阶段（3-4 周）- 完善功能

- [ ] 添加审计日志
- [ ] 改进 ERPNext 集成（异步队列、重试、降级）
- [ ] 添加权限校验
- [ ] 添加限流机制

#### 第四阶段（持续）- 架构优化

- [ ] 引入 CQRS 分离读写
- [ ] 添加领域事件
- [ ] 重构复杂业务逻辑
- [ ] 完善 API 文档

---

**评估结束**

当前库存实现整体良好，但存在一些关键问题需要修复。建议优先解决缓存一致性和幂等性问题，然后逐步优化性能和完善功能。
