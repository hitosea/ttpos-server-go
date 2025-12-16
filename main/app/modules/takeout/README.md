# 外卖平台集成模块（Takeout）

> 基于 DDD 架构的外卖平台菜单集成模块

---

## 📋 概述

本模块提供 TTPOS 与外卖平台（Grab、LINE MAN 等）之间的数据集成能力，采用领域驱动设计（DDD）架构，支持多平台扩展。

### 当前功能

- ✅ Grab 菜单导出（TTPOS → Grab 格式）
- ✅ Grab 菜单导入（Grab 格式 → TTPOS）
- ✅ 菜单预览（实时预览转换结果）
- ⏳ 其他平台扩展（LINE MAN、Foodpanda 等）

---

## 🏗️ 架构设计

### DDD 分层结构

```
app/modules/takeout/
├── domain/                          # 领域层（核心业务逻辑）
│   ├── menu/                        # 菜单聚合
│   │   ├── entity/                  # 聚合根
│   │   │   └── takeout_menu.go      # 外卖菜单聚合根
│   │   ├── valueobject/             # 值对象（平台无关）
│   │   │   ├── currency.go          # 货币
│   │   │   ├── selling_time.go      # 售卖时段
│   │   │   ├── category.go          # 分类
│   │   │   ├── menu_item.go         # 商品
│   │   │   └── modifier.go          # 规格/加料
│   │   └── repository/              # 仓储接口
│   │       └── takeout_menu_repository.go
│   └── service/                     # 领域服务
│       └── platform_converter.go    # 平台转换器接口（策略模式）
├── application/                     # 应用层（编排与协调）
│   └── takeout_menu_app_service.go  # 应用服务
└── infrastructure/                  # 基础设施层
    ├── persistence/                 # 持久化实现
    │   └── takeout_menu_repository_impl.go
    └── adapter/                     # 平台适配器（适配器模式）
        └── grab/                    # Grab 平台适配器
            ├── grab_converter.go    # Grab 转换器实现
            └── grab_models.go       # Grab 数据模型
```

### 设计模式

1. **策略模式**：不同平台使用不同的转换策略
2. **适配器模式**：将平台特定格式适配为通用领域模型
3. **仓储模式**：封装数据访问逻辑
4. **值对象模式**：不可变的业务概念

---

## 🔌 API 接口

### 1. 导出菜单

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

### 2. 导入菜单

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

### 3. 预览菜单

```http
GET /api/v1/takeout/menu/preview?platform=grab
Authorization: Bearer {token}
```

---

## 🚀 扩展新平台

### Step 1：创建平台适配器

在 `infrastructure/adapter/{platform}/` 下创建：

```go
// {platform}_converter.go
type {Platform}Converter struct {
    dbm *database.DBManager
}

func New{Platform}Converter(dbm *database.DBManager) service.IPlatformConverter {
    return &{Platform}Converter{dbm: dbm}
}

// 实现 IPlatformConverter 接口
func (c *{Platform}Converter) ConvertFromTTPOS(ctx context.Context, menu *entity.TakeoutMenu) (interface{}, error) {
    // 实现转换逻辑
}

func (c *{Platform}Converter) ConvertToTTPOS(ctx context.Context, platformData interface{}) (*entity.TakeoutMenu, error) {
    // 实现转换逻辑
}

func (c *{Platform}Converter) ValidateData(platformData interface{}) error {
    // 实现验证逻辑
}

func (c *{Platform}Converter) GetPlatformName() string {
    return "{platform}"
}
```

### Step 2：注册转换器

在 `application/takeout_menu_app_service.go` 中注册：

```go
func NewTakeoutMenuAppService(dbm *database.DBManager, menuRepo repository.ITakeoutMenuRepository) ITakeoutMenuAppService {
    converters := make(map[string]service.IPlatformConverter)
    converters["grab"] = grab.NewGrabConverter(dbm)
    converters["lineman"] = lineman.NewLinemanConverter(dbm)  // 新增平台
    // ...
}
```

### Step 3：测试

使用相同的 API，只需修改 `platform` 参数即可。

---

## 📚 相关文档

- [Grab 菜单集成文档](../../../../docs/shared/integrations/grab/grab-menu-integration.md)
- [DDD 模块开发指南](../README.md)
- [Go Main 开发规范](../../../../.cursor/rules/go-main.mdc)

---

## ⚠️ 注意事项

1. **价格单位**：所有价格均使用"分"为单位（int64）
2. **多语言支持**：使用 `dto.LocaleResponse` 结构
3. **错误处理**：使用 `errors.WithMessage` 包装错误
4. **协程使用**：使用 `utils.Go` 方法
5. **数据验证**：导入前进行完整的数据验证

---

## 🔄 当前状态

- ✅ 核心架构完成
- ✅ Grab 平台基础支持
- ⏳ 数据库持久化（TODO）
- ⏳ 完整的商品数据加载（TODO）
- ⏳ LINE MAN 平台支持（TODO）

---

**创建时间**：2025-12-09  
**维护者**：TTPOS Team

