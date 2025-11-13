# 沽清服务 (SoldOut Service)

## 概述

`sold_out.go` 实现了沽清（Sold Out）管理服务，负责管理餐饮系统中商品的沽清状态。沽清是指某个商品或规格暂时缺货、售罄的状态，在餐饮行业中是常见的库存管理功能。

**文件路径**: `ttpos-server-go/main/app/service/sold_out.go`

## 核心功能

### 1. 商品沽清管理
- 获取沽清商品列表（支持分页）
- 标记商品为沽清状态
- 取消单个商品的沽清
- 批量取消所有沽清商品

### 2. 实时通知
- 通过 WebSocket 实时推送沽清状态变更
- 支持向所有客户端广播更新

### 3. 多语言支持
- 商品名称和规格名称支持多语言显示

## 接口定义

### ISoldOutSrv 接口

```go
type ISoldOutSrv interface {
    GetSoldOutList(companyUuid uint64, soldOutReq req.SoldOutListReq) (resp.SoldOutPaginationResp, error)
    CancelSoldOut(companyUuid uint64, productBomUuid uint64) error
    CancelAllSoldOut(companyUuid uint64) error
    AddSoldOut(companyUuid uint64, items []req.SoldOutItem) error
}
```

### soldOutSrv 结构体

```go
type soldOutSrv struct {
    dbm       *database.DBManager // 数据库管理器
    localeSrv ILocaleSrv          // 多语言服务
}
```

## 依赖项

### 内部依赖
- **repository.ProductRepo**: 产品数据仓库，用于查询和更新商品沽清状态
- **ILocaleSrv**: 多语言服务，处理商品名称的多语言显示

### 外部依赖
- **database.DBManager**: 数据库管理器，提供数据库连接
- **websocket**: WebSocket 服务，用于推送实时更新
- **utils**: 工具包，提供协程管理等功能

## 核心方法详解

### 1. GetSoldOutList - 获取沽清商品列表

**方法签名**:
```go
func (s *soldOutSrv) GetSoldOutList(companyUuid uint64, soldOutReq req.SoldOutListReq) (resp.SoldOutPaginationResp, error)
```

**功能**: 分页获取当前标记为沽清状态的商品列表。

**参数说明**:
- `companyUuid`: 公司唯一标识
- `soldOutReq`: 请求参数，包含分页信息
  - `PageNo`: 页码
  - `PageSize`: 每页大小

**返回值**:
```go
type SoldOutPaginationResp struct {
    List []SoldOut      // 沽清商品列表
    Meta PageResponse   // 分页元数据
}

type SoldOut struct {
    LocaleProductName    map[string]string // 商品多语言名称
    ProductBomUuid       uint64            // 商品BOM UUID
    LocaleProductBomName map[string]string // 商品规格多语言名称
}
```

**实现流程**:

```45:64:ttpos-server-go/main/app/service/sold_out.go
	productRepo := repository.NewProductRepo(s.dbm.GetDB(companyUuid))

	boms, total, err := productRepo.GetSoldOutWithPagination(soldOutReq.PageNo, soldOutReq.PageSize,
		productRepo.WithProductPackage(),
		productRepo.WithProductPackageMultiLanguageName(),
		productRepo.WithProductFlavor(),
		productRepo.WithProductFlavorMultiLanguageName())

	if err != nil {
		return resp.SoldOutPaginationResp{}, errors.WithMessage(err, "获取沽清商品列表失败")
	}

	soldOuts := make([]resp.SoldOut, 0, len(boms))

	for _, bom := range boms {
		soldOuts = append(soldOuts, resp.SoldOut{
			LocaleProductName:    bom.ProductPackage.MultiLanguageName.GetNames(),
			ProductBomUuid:       bom.Uuid,
			LocaleProductBomName: bom.ProductFlavor.MultiLanguageName.GetNames(),
		})
	}
```

**关键点**:
1. 使用 `GetSoldOutWithPagination` 从数据库获取沽清商品
2. 预加载关联数据（产品包、口味、多语言名称）
3. 转换为响应格式，提供多语言名称支持

---

### 2. CancelSoldOut - 取消单个沽清商品

**方法签名**:
```go
func (s *soldOutSrv) CancelSoldOut(companyUuid uint64, productBomUuid uint64) error
```

**功能**: 将指定商品从沽清状态恢复为可售状态。

**参数说明**:
- `companyUuid`: 公司唯一标识
- `productBomUuid`: 商品BOM的唯一标识

**实现流程**:

```78:94:ttpos-server-go/main/app/service/sold_out.go
func (s *soldOutSrv) CancelSoldOut(companyUuid uint64, productBomUuid uint64) error {
	productRepo := repository.NewProductRepo(s.dbm.GetDB(companyUuid))
	if err := productRepo.UpdateProductBomSoldOut([]repository.DBOption{productRepo.WhereBomUuid(productBomUuid)}, map[string]any{
		"is_sold_out": 0,
	}); err != nil {
		return errors.WithMessage(err, "取消沽清商品失败")
	}
	// 推送沽清商品
	utils.Go(func() {
		websocket.PushClient(companyUuid, websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_PRODUCT, map[string]interface{}{
			"type":         "update",
			"product_uuid": productBomUuid,
			"update_time":  time.Now().Unix(),
		})
	})
	return nil
}
```

**关键点**:
1. 更新数据库，将 `is_sold_out` 字段设置为 0
2. 异步推送 WebSocket 消息通知所有客户端
3. 推送消息包含商品 UUID 和更新时间

**WebSocket 推送详情**:
- **接收端**: 所有客户端（`SourceAll`）
- **消息类型**: `UPDATE_PRODUCT`
- **消息内容**:
  - `type`: "update" - 更新类型
  - `product_uuid`: 具体商品的 UUID
  - `update_time`: 更新时间戳

---

### 3. CancelAllSoldOut - 取消全部沽清商品

**方法签名**:
```go
func (s *soldOutSrv) CancelAllSoldOut(companyUuid uint64) error
```

**功能**: 批量取消所有商品的沽清状态，一键恢复所有商品为可售状态。

**参数说明**:
- `companyUuid`: 公司唯一标识

**实现流程**:

```97:115:ttpos-server-go/main/app/service/sold_out.go
func (s *soldOutSrv) CancelAllSoldOut(companyUuid uint64) error {
	productRepo := repository.NewProductRepo(s.dbm.GetDB(companyUuid))

	if err := productRepo.UpdateProductBomSoldOut([]repository.DBOption{productRepo.WhereBomIsSoldOut()}, map[string]any{
		"is_sold_out": 0,
	}); err != nil {
		return errors.WithMessage(err, "全部取消沽清商品失败")
	}

	// 推送沽清商品
	utils.Go(func() {
		websocket.PushClient(companyUuid, websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_PRODUCT, map[string]interface{}{
			"type":         "update",
			"product_uuid": 0,
			"update_time":  time.Now().Unix(),
		})
	})
	return nil
}
```

**关键点**:
1. 使用 `WhereBomIsSoldOut` 条件批量更新所有沽清商品
2. 异步推送全局更新通知
3. 推送消息中 `product_uuid` 为 0，表示全局更新

**与单个取消的区别**:
- 单个取消：`product_uuid` 为具体商品 UUID
- 批量取消：`product_uuid` 为 0，表示所有商品都可能受影响

---

### 4. AddSoldOut - 添加商品沽清

**方法签名**:
```go
func (s *soldOutSrv) AddSoldOut(companyUuid uint64, items []req.SoldOutItem) error
```

**功能**: 批量设置商品的沽清状态，支持将商品标记为沽清或取消沽清。

**参数说明**:
- `companyUuid`: 公司唯一标识
- `items`: 沽清项目列表
  ```go
  type SoldOutItem struct {
      ProductBomUuid uint64 // 商品BOM UUID
      IsSoldOut      *bool  // 是否沽清
  }
  ```

**实现流程**:

```118:140:ttpos-server-go/main/app/service/sold_out.go
func (s *soldOutSrv) AddSoldOut(companyUuid uint64, items []req.SoldOutItem) error {
	productRepo := repository.NewProductRepo(s.dbm.GetDB(companyUuid))
	for _, item := range items {
		soldOutMap := map[bool]uint{
			true:  1,
			false: 0,
		}
		if err := productRepo.UpdateProductBomSoldOut([]repository.DBOption{productRepo.WhereBomUuid(item.ProductBomUuid)}, map[string]any{
			"is_sold_out": soldOutMap[*item.IsSoldOut],
		}); err != nil {
			return errors.WithMessage(err, "沽清商品失败")
		}
		// 推送沽清商品
		utils.Go(func() {
			websocket.PushClient(companyUuid, websocket.SourceAll, websocket.SourceAll, websocket.UPDATE_PRODUCT, map[string]interface{}{
				"type":         "update",
				"product_uuid": item.ProductBomUuid,
				"update_time":  time.Now().Unix(),
			})
		})
	}
	return nil
}
```

**关键点**:
1. 支持批量操作多个商品
2. 使用 `soldOutMap` 将布尔值转换为数值（true -> 1, false -> 0）
3. 每个商品更新后都会发送独立的 WebSocket 通知
4. 循环处理，遇到错误立即返回

**注意事项**:
- 此方法既可以设置沽清（`IsSoldOut = true`），也可以取消沽清（`IsSoldOut = false`）
- 每次数据库更新后都触发 WebSocket 推送
- 如果某个商品更新失败，会中断整个批量操作

---

## 服务创建

### 构造函数

```go
func NewSoldOutSrv(dbm *database.DBManager, localSrv ILocaleSrv) ISoldOutSrv {
    return NewSoldOutSrvImpl(dbm, localSrv)
}

func NewSoldOutSrvImpl(dbm *database.DBManager, localeSrv ILocaleSrv) ISoldOutSrv {
    return &soldOutSrv{
        dbm:       dbm,
        localeSrv: localeSrv,
    }
}
```

**依赖注入**:
- `dbm`: 数据库管理器
- `localeSrv`: 多语言服务（虽然在当前实现中未直接使用，但保留用于未来扩展）

---

## 数据模型

### 数据库字段
商品 BOM 表中的沽清相关字段：
- `is_sold_out`: TINYINT 类型
  - `0`: 正常可售
  - `1`: 已沽清

### ProductBom 概念
- **ProductBom** (Bill of Materials): 商品物料清单，代表商品的特定规格或口味
- 一个商品（Product）可以有多个规格（Bom）
- 沽清是在 Bom 级别进行管理的，而不是商品级别

**示例**:
- 商品: "奶茶"
- 规格1: "大杯-珍珠奶茶" (BomUuid: 101) - 可能沽清
- 规格2: "中杯-珍珠奶茶" (BomUuid: 102) - 正常供应

---

## WebSocket 实时通知机制

### 推送策略

所有沽清状态的变更都会通过 WebSocket 实时推送到客户端：

```go
websocket.PushClient(
    companyUuid,           // 公司标识
    websocket.SourceAll,   // 发送源：所有
    websocket.SourceAll,   // 接收源：所有
    websocket.UPDATE_PRODUCT, // 消息类型：商品更新
    payload                // 消息载荷
)
```

### 消息格式

**单个商品更新**:
```json
{
    "type": "update",
    "product_uuid": 12345,
    "update_time": 1699876543
}
```

**全局更新**（取消所有沽清时）:
```json
{
    "type": "update",
    "product_uuid": 0,
    "update_time": 1699876543
}
```

### 客户端处理建议
- 接收到 `product_uuid > 0`: 刷新特定商品的状态
- 接收到 `product_uuid = 0`: 刷新所有商品列表
- 所有客户端类型（收银、助手、平板等）都会收到通知

---

## 使用场景

### 1. 餐厅日常运营
```
场景: 厨房发现某道菜的主要食材用完了
流程:
1. 厨房或前台工作人员标记该菜品为沽清
2. 调用 AddSoldOut 方法
3. WebSocket 实时推送到所有终端
4. 收银员看到该商品显示"沽清"，无法继续销售
```

### 2. 早餐时段结束
```
场景: 早餐时段结束，需要将所有早餐商品标记为沽清
流程:
1. 批量调用 AddSoldOut，传入所有早餐商品
2. 系统逐个标记并推送更新
3. 所有终端实时同步，早餐商品不再可选
```

### 3. 补货后恢复
```
场景: 某个沽清的商品重新补货了
流程:
1. 工作人员调用 CancelSoldOut
2. 商品恢复可售状态
3. 所有终端收到更新通知
4. 收银员可以继续销售该商品
```

### 4. 营业日结束
```
场景: 餐厅打烊，第二天重新营业
流程:
1. 调用 CancelAllSoldOut
2. 批量清除所有沽清标记
3. 全局 WebSocket 推送
4. 所有商品恢复可售状态
```

---

## 错误处理

### 错误类型

1. **数据库操作失败**:
   - 获取沽清列表失败：`"获取沽清商品列表失败"`
   - 取消单个沽清失败：`"取消沽清商品失败"`
   - 取消全部沽清失败：`"全部取消沽清商品失败"`
   - 添加沽清失败：`"沽清商品失败"`

2. **错误处理策略**:
   - 使用 `errors.WithMessage` 包装原始错误
   - 提供中文错误消息
   - AddSoldOut 遇到错误会立即中断并返回

### 错误传播
```go
if err != nil {
    return errors.WithMessage(err, "具体错误消息")
}
```

---

## 性能考虑

### 1. 异步推送
所有 WebSocket 推送操作都使用 `utils.Go` 异步执行，不阻塞主流程：

```go
utils.Go(func() {
    websocket.PushClient(...)
})
```

**优点**:
- 提高响应速度
- 即使 WebSocket 推送失败也不影响数据库操作
- 支持高并发场景

### 2. 数据库查询优化
- 使用预加载（Eager Loading）减少 N+1 查询问题
- 支持分页查询，避免一次性加载大量数据

```go
productRepo.GetSoldOutWithPagination(
    pageNo, pageSize,
    productRepo.WithProductPackage(),
    productRepo.WithProductPackageMultiLanguageName(),
    productRepo.WithProductFlavor(),
    productRepo.WithProductFlavorMultiLanguageName()
)
```

### 3. 批量操作的权衡
- `AddSoldOut` 采用循环更新而非批量更新
- 每次更新后立即推送通知
- 优点：客户端可以逐步收到更新
- 缺点：大量商品时性能较低

**改进建议**:
- 可以考虑批量更新后统一推送
- 或者累积一定数量后批量推送

---

## 多语言支持

### 名称处理
商品名称和规格名称都支持多语言：

```go
resp.SoldOut{
    LocaleProductName:    bom.ProductPackage.MultiLanguageName.GetNames(),
    LocaleProductBomName: bom.ProductFlavor.MultiLanguageName.GetNames(),
}
```

### 数据结构
```go
type MultiLanguageName map[string]string
// 例如:
{
    "zh-CN": "珍珠奶茶",
    "en-US": "Pearl Milk Tea",
    "ja-JP": "タピオカミルクティー"
}
```

### 客户端使用
客户端根据当前语言环境选择对应的名称显示：
```go
productName := soldOut.LocaleProductName[currentLocale]
```

---

## 与其他服务的集成

### 1. Product Service
- 依赖产品仓库获取和更新商品信息
- 操作的是 ProductBom 级别的数据

### 2. WebSocket Service
- 所有状态变更都通过 WebSocket 推送
- 确保多终端实时同步

### 3. Locale Service
- 虽然当前未直接使用
- 预留用于未来的多语言扩展

### 4. 可能的集成场景
- **订单服务**: 下单时检查商品是否沽清
- **库存服务**: 库存不足时自动标记沽清
- **报表服务**: 统计沽清频率，优化备货策略

---

## 最佳实践

### 1. 调用建议

**获取列表**:
```go
soldOutList, err := soldOutSrv.GetSoldOutList(companyUuid, req.SoldOutListReq{
    PageNo:   1,
    PageSize: 20,
})
```

**标记单个商品沽清**:
```go
items := []req.SoldOutItem{
    {
        ProductBomUuid: 12345,
        IsSoldOut:      utils.BoolPtr(true),
    },
}
err := soldOutSrv.AddSoldOut(companyUuid, items)
```

**取消沽清**:
```go
// 方式1: 取消单个
err := soldOutSrv.CancelSoldOut(companyUuid, productBomUuid)

// 方式2: 通过 AddSoldOut 取消
items := []req.SoldOutItem{
    {
        ProductBomUuid: 12345,
        IsSoldOut:      utils.BoolPtr(false),
    },
}
err := soldOutSrv.AddSoldOut(companyUuid, items)
```

**批量取消所有沽清**:
```go
err := soldOutSrv.CancelAllSoldOut(companyUuid)
```

### 2. 客户端监听 WebSocket

```javascript
websocket.on('UPDATE_PRODUCT', (data) => {
    if (data.product_uuid === 0) {
        // 刷新所有商品
        refreshAllProducts();
    } else {
        // 刷新特定商品
        refreshProduct(data.product_uuid);
    }
});
```

### 3. 错误处理

```go
if err := soldOutSrv.AddSoldOut(companyUuid, items); err != nil {
    // 记录日志
    log.Error("沽清商品失败", zap.Error(err))
    
    // 返回友好提示
    return errors.New("设置沽清失败，请稍后重试")
}
```

---

## 潜在改进点

### 1. 批量操作优化
**当前问题**: `AddSoldOut` 逐个更新，性能较低

**改进方案**:
```go
// 批量更新数据库
err := productRepo.BatchUpdateProductBomSoldOut(items)

// 统一推送一次通知
websocket.PushClient(..., map[string]interface{}{
    "type": "batch_update",
    "product_uuids": extractUuids(items),
    "update_time": time.Now().Unix(),
})
```

### 2. 事务支持
**当前问题**: `AddSoldOut` 中途失败可能导致部分更新

**改进方案**:
```go
tx := s.dbm.GetDB(companyUuid).Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

// 批量更新
for _, item := range items {
    // 使用事务
}

tx.Commit()
```

### 3. 缓存支持
**问题**: 频繁查询沽清列表可能产生数据库压力

**改进方案**:
- 使用 Redis 缓存沽清商品列表
- 状态变更时更新缓存
- 设置合理的过期时间

### 4. 权限控制
**当前缺失**: 没有权限验证

**建议添加**:
```go
// 检查用户是否有沽清管理权限
if !ctx.HasPermission("soldout.manage") {
    return errors.New("无权限操作")
}
```

### 5. 操作日志
**建议添加**: 记录沽清操作历史

```go
type SoldOutLog struct {
    ProductBomUuid uint64
    Action         string // "set", "cancel"
    OperatorUuid   uint64
    OperateTime    time.Time
}
```

### 6. 自动恢复机制
**增强功能**: 支持定时自动取消沽清

```go
// 例如：每天营业开始时自动取消所有沽清
scheduler.AddDailyTask("08:00", func() {
    soldOutSrv.CancelAllSoldOut(companyUuid)
})
```

---

## 业务规则总结

1. **操作级别**: 沽清是在 ProductBom（商品规格）级别管理的
2. **实时同步**: 所有状态变更都通过 WebSocket 实时推送
3. **灵活控制**: 支持单个、批量、全部三种操作粒度
4. **多语言**: 商品名称支持多语言显示
5. **异步推送**: WebSocket 推送不阻塞主流程
6. **错误处理**: 批量操作遇错立即中断

---

## 相关文件

- **DTO 定义**: `ttpos-server-go/app/dto/req/sold_out.go`, `ttpos-server-go/app/dto/resp/sold_out.go`
- **数据库仓库**: `ttpos-server-go/app/repository/product.go`
- **WebSocket 服务**: `ttpos-server-go/pkg/websocket/`
- **工具函数**: `ttpos-server-go/pkg/utils/`

---

## 总结

沽清服务是餐饮管理系统中的重要功能，用于动态管理商品的可售状态。该服务具有以下特点：

1. **简洁高效**: 接口清晰，功能专一
2. **实时性强**: 通过 WebSocket 确保多终端实时同步
3. **灵活可控**: 支持多种操作粒度
4. **国际化友好**: 原生支持多语言
5. **扩展性好**: 预留了多语言服务等扩展接口

通过合理使用沽清功能，餐厅可以有效管理商品库存，避免缺货销售，提升客户体验。

