# 外卖平台集成模块（Takeout）

> 基于 DDD 架构的多平台外卖订单和菜单管理模块

---

## 📋 概述

本模块提供 TTPOS 与外卖平台（Grab、LINE MAN、FoodPanda 等）之间的完整集成能力，包括菜单同步和订单管理，采用领域驱动设计（DDD）架构，支持多平台扩展。

### 核心功能

#### ✅ 订单管理
- 多平台订单接收和解析（Grab、Lineman）
- 订单商品和修饰符映射（flavor/sauce/attr/commodity）
- 订单状态管理（待接单、配餐中、配送中、已完成）
- 库存检查和原料扣减
- 送厨单生成（ttpos_production_order）
- 商家联/客户联打印（支持 TTPOS 名称和平台名称）

#### ✅ 菜单同步
- Grab 菜单导出（TTPOS → Grab 格式）
- Grab 菜单导入（Grab 格式 → TTPOS）
- 菜单预览（实时预览转换结果）
- 商品和规格价格映射

#### ⏳ 扩展支持
- LINE MAN 平台（开发中）
- FoodPanda 平台（规划中）

---

## 🏗️ 架构设计

### DDD 分层结构

```
app/modules/takeout/
├── domain/                          # 领域层（核心业务逻辑）
│   ├── model/                       # 领域模型（聚合根）
│   │   ├── takeout.go               # 平台状态管理
│   │   ├── takeout_order.go         # 订单聚合根
│   │   ├── takeout_order_item.go    # 订单商品
│   │   ├── takeout_order_item_modifier.go  # 商品修饰符
│   │   ├── takeout_order_receiver.go       # 收货人信息
│   │   ├── takeout_order_campaign.go       # 订单活动
│   │   ├── takeout_order_promo.go          # 订单促销
│   │   ├── takeout_order_material.go       # 订单原料
│   │   └── takeout_settings.go      # 平台配置
│   ├── types/                       # 值对象和类型定义
│   │   ├── modifier_info.go         # 修饰符信息（含名称和数量）
│   │   └── product_info.go          # 商品信息（含显示名和标识名）
│   ├── value_object/                # 值对象
│   │   ├── item_id_parser.go        # 商品ID解析器
│   │   ├── modifier.go              # 修饰符值对象
│   │   └── takeout_platform.go      # 平台枚举
│   ├── event/                       # 领域事件
│   │   ├── event.go                 # 事件基类
│   │   ├── order_created.go         # 订单创建事件
│   │   ├── order_accepted.go        # 订单接单事件
│   │   ├── order_rejected.go        # 订单拒单事件
│   │   ├── order_ready.go           # 订单准备完成事件
│   │   ├── order_rider_processing.go # 骑手配送中事件
│   │   ├── order_completed.go       # 订单完成事件
│   │   └── order_cancelled.go       # 订单取消事件
│   ├── repository/                  # 仓储接口
│   │   ├── menu_data_repository.go  # 菜单数据仓储接口
│   │   └── takeout_import_log_repository.go  # 导入日志仓储接口
│   ├── service/                     # 领域服务
│   │   ├── takeout_order_service.go # 订单领域服务
│   │   ├── takeout_order_material_service.go  # 订单原料服务
│   │   ├── takeout_service.go       # 平台管理服务
│   │   ├── takeout_settings_service.go  # 平台设置服务
│   │   ├── order_converter.go       # 订单转换服务
│   │   ├── platform_converter.go    # 平台转换器接口（策略模式）
│   │   └── import_progress_service.go  # 导入进度服务
│   └── helper/                      # 辅助工具
│       └── comparison.go            # 数据对比工具
├── application/                     # 应用层（编排与协调）
│   ├── takeout_order_service.go     # 订单应用服务
│   └── takeout_app_service.go       # 外卖应用服务（平台管理、菜单导入等）
└── infrastructure/                  # 基础设施层
    ├── persistence/                 # 持久化实现
    │   ├── base.go                  # 仓储基类
    │   ├── takeout_repository.go    # 平台管理仓储
    │   ├── takeout_settings_repo.go # 平台设置仓储
    │   ├── takeout_order_repo.go    # 订单仓储实现
    │   ├── takeout_order_item_repo.go  # 订单商品仓储
    │   ├── takeout_order_item_modifier_repo.go  # 订单修饰符仓储
    │   ├── takeout_order_receiver_repo.go       # 订单收货人仓储
    │   ├── takeout_order_campaign_repo.go       # 订单活动仓储
    │   ├── takeout_order_material_repo.go       # 订单原料仓储
    │   ├── takeout_bom_mapping_repo.go          # BOM映射仓储
    │   ├── takeout_import_log_repository_impl.go # 导入日志仓储
    │   └── menu_data_repository_impl.go         # 菜单数据仓储实现
    └── adapter/                     # 平台适配器
        ├── grab/                    # Grab 平台适配器
        │   ├── grab_menu_converter.go   # Grab 菜单转换器
        │   └── grab_order_converter.go  # Grab 订单转换器
        └── rpc/                     # RPC 客户端
            ├── bmp_client.go        # BMP 微服务客户端
            └── takeout_rpc_service.go   # 外卖 RPC 服务
```

### 设计模式

1. **策略模式**：不同平台使用不同的转换策略
2. **适配器模式**：将平台特定格式适配为通用领域模型
3. **仓储模式**：封装数据访问逻辑
4. **值对象模式**：不可变的业务概念
5. **聚合根模式**：订单作为聚合根管理订单商品和修饰符

### 核心设计理念

#### 1. 名称冗余策略（Name Redundancy）

为了减少查询和保持历史数据一致性，在订单创建时保存名称快照：

```go
// 商品名称：双名称设计
item.ItemName = productInfo.Name          // 显示用（优先外卖表）
item.TtposItemName = productInfo.TtposName // 标识用（核心表）

// 修饰符名称：双名称设计
modifier.ModifierName = modifierInfo.Name        // 显示用
modifier.TtposModifierName = modifierInfo.TtposName // 标识用

// commodity 类型额外保存规格
if modifier.TtposModifierType == "commodity" {
    modifier.TtposFlavorUuid = modifierInfo.TtposFlavorUuid
    modifier.TtposFlavorName = modifierInfo.TtposFlavorName
}
```

**优势**：
- ✅ 打印时无需额外查询数据库
- ✅ 历史订单不受商品改名影响
- ✅ 商家联/客户联可显示不同名称
- ✅ 送厨单直接使用 TTPOS 标准名称

#### 2. 修饰符类型系统

支持 4 种修饰符类型，精确映射平台数据：

| Type | 说明 | 关联表 | 示例 |
|------|------|--------|------|
| `flavor` | 规格 | `ttpos_product_bom` | 大/中/小杯 |
| `sauce` | 加料 | `ttpos_product_bom` | 珍珠/椰果 |
| `attr` | 属性 | `ttpos_product_package_attribute` | 冰度/糖度 |
| `commodity` | 套餐商品 | `ttpos_product_package_group_item` | 珍珠奶茶(大杯) x2 |

#### 3. 数据一致性保证

- **事务处理**：订单创建使用数据库事务
- **幂等性**：平台订单ID唯一索引防止重复
- **原始数据保留**：`raw_data` 字段保存平台完整数据
- **软删除**：使用 `delete_time` 实现软删除

---

## 🔌 API 接口

### 订单管理 API

#### 1. 接收订单（Webhook）

```http
POST /api/v1/takeout/order/webhook
Content-Type: application/json
X-Platform: grab

{
  "orderID": "GF-123456",
  "orderState": "PENDING",
  "items": [...],
  "receiver": {...},
  ...
}
```

#### 2. 接单

```http
POST /api/v1/takeout/order/accept
Authorization: Bearer {token}
Content-Type: application/json

{
  "orderUuid": 123456789,
  "estimatedReadyTime": 1703836800
}
```

#### 3. 拒单

```http
POST /api/v1/takeout/order/reject
Authorization: Bearer {token}
Content-Type: application/json

{
  "orderUuid": 123456789,
  "rejectReasonCode": "SOLD_OUT",
  "rejectReason": "商品已售罄"
}
```

#### 4. 订单列表

```http
GET /api/v1/takeout/order/list?platform=grab&state=1&page=1&pageSize=20
Authorization: Bearer {token}
```

### 菜单管理 API

#### 1. 导出菜单

```http
POST /api/v1/takeout/menu/export
Authorization: Bearer {token}
Content-Type: application/json

{
  "platform": "grab",
  "shopUuid": 0,
  "categoryIds": [],
  "sellingTimeIds": []
}
```

#### 2. 导入菜单

```http
POST /api/v1/takeout/menu/import
Authorization: Bearer {token}
Content-Type: application/json

{
  "platform": "grab",
  "menuData": "{...}",
  "syncMode": "full"
}
```

#### 3. 预览菜单

```http
GET /api/v1/takeout/menu/preview?platform=grab
Authorization: Bearer {token}
```

### 打印 API

#### 1. 打印商家联

```http
POST /api/v1/takeout/order/print
Authorization: Bearer {token}
Content-Type: application/json

{
  "orderUuid": 123456789,
  "receiptType": "merchant"
}
```

#### 2. 打印客户联

```http
POST /api/v1/takeout/order/print
Authorization: Bearer {token}
Content-Type: application/json

{
  "orderUuid": 123456789,
  "receiptType": "customer"
}
```

---

## 🚀 扩展新平台

### 订单接收扩展

#### Step 1：创建订单解析器

在 `domain/service/` 下创建平台特定的订单解析逻辑：

```go
// {platform}_order_parser.go
func Parse{Platform}Order(rawData []byte) (*model.TakeoutOrder, error) {
    // 1. 解析平台订单数据
    var platformOrder {Platform}Order
    if err := json.Unmarshal(rawData, &platformOrder); err != nil {
        return nil, err
    }
    
    // 2. 转换为 TTPOS 订单模型
    order := &model.TakeoutOrder{
        Platform:        "{platform}",
        PlatformOrderId: platformOrder.OrderID,
        // ... 其他字段映射
    }
    
    // 3. 处理订单商品
    for _, item := range platformOrder.Items {
        orderItem := &model.TakeoutOrderItem{
            PlatformItemId: item.ID,
            ItemName:       item.Name,
            // ... 其他字段映射
        }
        
        // 4. 处理修饰符
        for _, mod := range item.Modifiers {
            modifier := &model.TakeoutOrderItemModifier{
                PlatformModifierId: mod.ID,
                ModifierName:       mod.Name,
                // ... 其他字段映射
            }
            orderItem.TakeoutOrderItemModifiers = append(orderItem.TakeoutOrderItemModifiers, *modifier)
        }
        
        order.TakeoutOrderItems = append(order.TakeoutOrderItems, *orderItem)
    }
    
    return order, nil
}
```

#### Step 2：注册平台解析器

在订单服务中注册：

```go
func (s *TakeoutOrderService) CreateOrder(ctx context.Context, platform string, rawData []byte) error {
    var order *model.TakeoutOrder
    var err error
    
    switch platform {
    case "grab":
        order, err = ParseGrabOrder(rawData)
    case "lineman":
        order, err = ParseLinemanOrder(rawData)  // 新增平台
    default:
        return errors.New("unsupported platform")
    }
    
    // ... 统一的订单处理流程
}
```

### 菜单同步扩展

#### Step 1：创建平台菜单转换器

在 `infrastructure/adapter/{platform}/` 下创建：

```go
// {platform}_menu_converter.go
type {Platform}MenuConverter struct {
    dbm *database.DBManager
}

func New{Platform}MenuConverter(dbm *database.DBManager) *{Platform}MenuConverter {
    return &{Platform}MenuConverter{dbm: dbm}
}

// 导出菜单（TTPOS → 平台格式）
func (c *{Platform}MenuConverter) ExportMenu(ctx context.Context, shopUuid uint64) (interface{}, error) {
    // 1. 从 TTPOS 数据库读取菜单数据
    // 2. 转换为平台格式
    // 3. 返回平台菜单数据
}

// 导入菜单（平台格式 → TTPOS）
func (c *{Platform}MenuConverter) ImportMenu(ctx context.Context, platformMenu interface{}) error {
    // 1. 解析平台菜单数据
    // 2. 转换为 TTPOS 格式
    // 3. 保存到数据库
}
```

#### Step 2：注册转换器

在 `application/takeout_app_service.go` 中注册：

```go
func NewTakeoutAppService(dbm *database.DBManager) ITakeoutAppService {
    service := &TakeoutAppService{
        dbm: dbm,
        menuConverters: make(map[string]MenuConverter),
    }
    
    // 注册菜单转换器
    service.menuConverters["grab"] = grab.NewGrabMenuConverter(dbm)
    service.menuConverters["lineman"] = lineman.NewLinemanMenuConverter(dbm)  // 新增平台
    
    return service
}
```

### 测试

使用相同的 API，只需修改 `platform` 参数即可。

---

## 📚 相关文档

### 数据库设计
- [数据库架构说明](../../../../docs/shared/takeout/database-architecture.md) - 完整的表结构和关系
- [订单收货人文档](../../../../docs/shared/takeout/takeout_order_receiver.md) - 收货人信息管理

### 平台集成
- [Grab 菜单集成文档](../../../../docs/shared/integrations/grab/grab-menu-integration.md)
- [LINE MAN 集成指南](../../../../docs/shared/integrations/lineman/)

### 开发规范
- [Go Main 开发规范](../../../../.cursor/rules/go-main.mdc)
- [数据库开发规范](../../../../.cursor/rules/database.mdc)
- [API 设计规范](../../../../.cursor/rules/api.mdc)

---

## ⚠️ 注意事项

### 数据处理

1. **价格单位**
   - 订单级别金额：使用"分"为单位（`bigint`）
   - 商品级别价格：使用"元"为单位（`decimal(20,4)`）
   ```go
   // 订单金额
   order.Subtotal = 10000  // 100.00 元
   
   // 商品价格
   item.Price = 100.0000   // 100.00 元
   ```

2. **多语言支持**
   - 使用 `dto.LocaleResponse` 结构
   - JSON 格式存储多语言名称
   ```go
   // 获取当前语言的名称
   name := language.JsonToLocaleResponse(item.ItemName).GetLocale(ctx.GetLanguage())
   ```

3. **时间处理**
   - 使用 Unix 时间戳（秒）
   - 统一使用 `int(10)` 或 `bigint` 类型

### 开发规范

1. **错误处理**
   ```go
   // 使用 errors.WithMessage 包装错误
   return errors.WithMessage(err, "订单创建失败")
   ```

2. **协程使用**
   ```go
   // 使用 utils.Go 方法
   utils.Go(func() {
       // 异步任务
   })
   ```

3. **事务处理**
   ```go
   // 订单创建必须使用事务
   tx := db.Begin()
   defer func() {
       if r := recover(); r != nil {
           tx.Rollback()
       }
   }()
   
   // ... 业务逻辑
   
   tx.Commit()
   ```

4. **数据验证**
   - 导入前进行完整的数据验证
   - 使用平台转换器的 `ValidateData` 方法

### 性能优化

1. **批量操作**
   ```go
   // 批量插入
   db.CreateInBatches(&items, 100)
   ```

2. **预加载关联**
   ```go
   // 避免 N+1 问题
   db.Preload("TakeoutOrderItems.TakeoutOrderItemModifiers").
      Preload("TakeoutOrderReceiver").
      Find(&orders)
   ```

3. **名称冗余**
   - 订单创建时保存名称快照，减少查询
   - 打印和送厨单直接使用保存的名称

### 数据一致性

1. **幂等性保证**
   - 平台订单ID唯一索引防止重复
   - Webhook 可能重复推送，需要幂等处理

2. **原始数据保留**
   - `raw_data` 字段保存平台完整数据
   - 用于问题追溯和数据对账

3. **软删除**
   - 使用 `delete_time` 实现软删除
   - 查询时始终过滤 `delete_time = 0`

---

## 🔄 当前状态

### ✅ 已完成
- 核心 DDD 架构设计
- 订单接收和处理（Grab、Lineman）
- 订单商品和修饰符映射（4种类型：flavor/sauce/attr/commodity）
- 名称冗余设计（双名称策略）
- 库存检查和原料扣减
- 送厨单生成
- 商家联/客户联打印（支持 TTPOS 名称和平台名称）
- Grab 平台基础支持（菜单同步）

### ⏳ 进行中
- LINE MAN 平台支持
- 订单状态同步（自动接单、准备完成通知）

### 📋 待开发
- FoodPanda 平台支持
- 订单数据统计和报表
- 配送员管理
- 实时订单推送（WebSocket）

---

## 📊 性能指标

### 当前表现
- 订单创建: < 500ms
- 订单查询: < 100ms
- 打印生成: < 200ms
- 送厨单创建: < 300ms

### 优化目标
- 支持 100+ 并发订单处理
- 数据库查询响应时间 < 50ms
- 批量操作性能提升 50%

---

**创建时间**：2025-12-09  
**最后更新**：2025-12-29  
**维护者**：TTPOS Team

