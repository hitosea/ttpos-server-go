# Grab 菜单集成文档

> TTPOS 与 Grab 外卖平台的菜单数据集成指南

---

## 📋 概述

本文档描述 TTPOS 系统与 Grab 外卖平台之间的菜单数据双向转换和同步机制。

### 功能特性

- ✅ **导出菜单**：将 TTPOS 菜单数据转换为 Grab 标准格式
- ✅ **导入菜单**：从 Grab 菜单数据导入到 TTPOS 系统
- ✅ **预览菜单**：实时预览转换后的菜单数据（不保存）
- ✅ **多平台支持**：基于 DDD 架构，易于扩展其他外卖平台

---

## 🏗️ 架构设计

### 模块结构

```
app/modules/takeout/                    # 外卖平台集成模块
├── domain/                             # 领域层
│   ├── menu/                           # 菜单聚合
│   │   ├── entity/                     # 聚合根
│   │   ├── valueobject/                # 值对象（通用）
│   │   └── repository/                 # 仓储接口
│   └── service/                        # 领域服务
│       └── platform_converter.go       # 平台转换器接口
├── application/                        # 应用层
│   └── takeout_menu_app_service.go     # 应用服务
└── infrastructure/                     # 基础设施层
    ├── persistence/                    # 持久化
    └── adapter/                        # 平台适配器
        └── grab/                       # Grab 适配器
            ├── grab_converter.go       # 转换器实现
            └── grab_models.go          # Grab 数据模型
```

### 设计模式

- **DDD（领域驱动设计）**：清晰的领域边界和分层
- **策略模式**：不同平台使用不同的转换策略
- **适配器模式**：将平台特定格式适配为通用领域模型

---

## 📊 Grab 菜单格式

### 标准结构

```json
{
  "currency": {
    "code": "SGD",
    "symbol": "S$",
    "exponent": 2
  },
  "sellingTimes": [
    {
      "id": "SELLINGTIME-01",
      "name": "Breakfast",
      "sequence": 1,
      "serviceHours": { ... },
      "startTime": "1000-01-01 00:00:00",
      "endTime": "9999-12-31 23:59:59"
    }
  ],
  "categories": [
    {
      "id": "CATEGORY-01",
      "name": "Savoury Pancakes",
      "sequence": 1,
      "availableStatus": "AVAILABLE",
      "items": [ ... ],
      "sellingTimeID": "SELLINGTIME-01",
      "nameTranslation": {
        "zh": "咸味煎饼"
      }
    }
  ]
}
```

### 字段说明

#### Currency（货币）

| 字段 | 类型 | 说明 |
|------|------|------|
| code | string | 货币代码（如 SGD, USD, THB） |
| symbol | string | 货币符号（如 S$, $, ฿） |
| exponent | int | 小数位数，一般为 2 |

#### SellingTime（售卖时段）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 售卖时段唯一标识 |
| name | string | 售卖时段名称 |
| sequence | int | 排序序号 |
| serviceHours | object | 营业时间配置（周一到周日） |
| startTime | string | 生效开始时间 |
| endTime | string | 生效结束时间 |

#### Category（分类）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 分类唯一标识 |
| name | string | 分类名称 |
| sequence | int | 排序序号 |
| availableStatus | string | 可用状态：AVAILABLE / UNAVAILABLE |
| items | array | 商品列表 |
| sellingTimeID | string | 关联的售卖时段 ID |
| nameTranslation | object | 多语言名称 |

#### Item（商品）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 商品唯一标识 |
| name | string | 商品名称 |
| sequence | int | 排序序号 |
| availableStatus | string | 可用状态 |
| price | int64 | 价格（单位：分） |
| description | string | 商品描述 |
| photos | array | 商品图片 URL 列表 |
| modifierGroups | array | 修饰符组列表（规格/加料） |
| campaignInfo | object | 营销活动信息（可选） |
| nameTranslation | object | 多语言名称 |

#### ModifierGroup（修饰符组）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 修饰符组唯一标识 |
| name | string | 修饰符组名称 |
| sequence | int | 排序序号 |
| availableStatus | string | 可用状态 |
| selectionRangeMin | int | 最小选择数量 |
| selectionRangeMax | int | 最大选择数量 |
| modifiers | array | 修饰符列表 |

#### Modifier（修饰符）

| 字段 | 类型 | 说明 |
|------|------|------|
| id | string | 修饰符唯一标识 |
| name | string | 修饰符名称 |
| sequence | int | 排序序号 |
| availableStatus | string | 可用状态 |
| price | int64 | 价格（单位：分） |
| nameTranslation | object | 多语言名称 |

---

## 🔌 API 接口

### 1. 导出菜单

**接口地址**：`POST /api/v1/takeout/menu/export`

**请求参数**：

```json
{
  "platform": "grab",
  "shopUuid": 0,
  "categoryIds": [],
  "sellingTimeIds": []
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| platform | string | 是 | 平台名称，当前支持：grab |
| shopUuid | uint64 | 否 | 门店 UUID，默认当前门店 |
| categoryIds | []uint64 | 否 | 分类 ID 列表，为空则导出所有分类 |
| sellingTimeIds | []uint64 | 否 | 售卖时段 ID 列表 |

**响应示例**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "platform": "grab",
    "menuData": {
      "currency": { ... },
      "sellingTimes": [ ... ],
      "categories": [ ... ]
    }
  }
}
```

### 2. 导入菜单

**接口地址**：`POST /api/v1/takeout/menu/import`

**请求参数**：

```json
{
  "platform": "grab",
  "shopUuid": 0,
  "menuData": "{ ... }",
  "syncMode": "full",
  "overwriteExisting": false
}
```

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| platform | string | 是 | 平台名称 |
| shopUuid | uint64 | 否 | 门店 UUID，默认当前门店 |
| menuData | string | 是 | Grab 菜单 JSON 字符串 |
| syncMode | string | 否 | 同步模式：full（全量）/ incremental（增量） |
| overwriteExisting | bool | 否 | 是否覆盖已存在的数据 |

**响应示例**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "success": true,
    "totalCategories": 10,
    "totalItems": 50,
    "createdCategories": 8,
    "updatedCategories": 2,
    "createdItems": 45,
    "updatedItems": 5,
    "errors": []
  }
}
```

### 3. 预览菜单

**接口地址**：`GET /api/v1/takeout/menu/preview`

**请求参数**：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| platform | string | 是 | 平台名称 |
| shopUuid | uint64 | 否 | 门店 UUID，默认当前门店 |

**响应示例**：

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "platform": "grab",
    "menuData": { ... }
  }
}
```

---

## 🔄 数据映射规则

### TTPOS → Grab

| TTPOS 实体 | TTPOS 表 | Grab 字段 | 转换说明 |
|-----------|----------|----------|---------|
| 货币配置 | company_setting | currency | 从公司配置获取 |
| 商品分类 | product_category | categories | 分类树扁平化 |
| 商品套餐 | product_package | items | 商品主体信息 |
| 商品规格 | product_flavor | modifiers | 映射为修饰符 |
| 商品加料 | product_sauce | modifiers | 映射为修饰符 |
| 属性组 | product_attribute_group | modifierGroups | 属性组映射 |
| 售卖时段 | selling_time | sellingTimes | 时段配置 |

### 价格转换

- **TTPOS 价格单位**：分（整数，int64）
- **Grab 价格单位**：分（整数，int64）
- **转换规则**：直接传递，无需转换

### 多语言处理

- TTPOS 使用 `multi_language_name` 表存储多语言
- Grab 使用 `nameTranslation` 字段存储多语言
- 支持语言：zh（中文）、en（英文）、th（泰语）等

---

## 💡 使用示例

### 示例 1：导出门店菜单到 Grab 格式

```bash
curl -X POST http://localhost:8080/api/v1/takeout/menu/export \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "grab"
  }'
```

### 示例 2：从 Grab 导入菜单

```bash
curl -X POST http://localhost:8080/api/v1/takeout/menu/import \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "platform": "grab",
    "menuData": "{\"currency\":{...},\"categories\":[...]}",
    "syncMode": "full"
  }'
```

### 示例 3：预览菜单

```bash
curl -X GET "http://localhost:8080/api/v1/takeout/menu/preview?platform=grab" \
  -H "Authorization: Bearer {token}"
```

---

## 🚀 扩展其他平台

基于 DDD 架构，扩展新的外卖平台非常简单：

### 步骤 1：创建平台适配器

在 `infrastructure/adapter/{platform}/` 下创建：

- `{platform}_converter.go`：实现 `IPlatformConverter` 接口
- `{platform}_models.go`：定义平台数据模型

### 步骤 2：注册转换器

在 `application/takeout_menu_app_service.go` 中注册：

```go
converters["lineman"] = lineman.NewLinemanConverter(dbm)
```

### 步骤 3：测试

使用相同的 API 接口，只需修改 `platform` 参数。

---

## ⚠️ 注意事项

1. **价格单位统一**：所有价格均使用"分"为单位，避免浮点数精度问题
2. **ID 生成规则**：建议使用 `{TYPE}-{UUID}` 格式，如 `CATEGORY-123`
3. **图片 URL**：确保图片 URL 可公网访问
4. **数据验证**：导入前会进行完整的数据验证
5. **事务处理**：导入操作使用事务，失败会自动回滚

---

## 🐛 故障排查

### 问题 1：导出菜单为空

**原因**：门店没有可用商品或分类

**解决**：
1. 检查 `product_category` 表是否有数据
2. 检查 `product_package` 表是否有数据
3. 确认商品状态为"可用"

### 问题 2：导入失败

**原因**：菜单数据格式不符合 Grab 规范

**解决**：
1. 检查 JSON 格式是否正确
2. 检查必填字段是否完整
3. 查看错误响应中的 `errors` 字段

### 问题 3：价格显示异常

**原因**：价格单位混淆

**解决**：
- 确保所有价格使用"分"为单位
- 例如：$10.00 应该传递为 1000

---

## 📚 相关文档

- [TTPOS API 设计规范](../../.cursor/rules/api.mdc)
- [DDD 模块开发指南](../modules/README.md)
- [LINE MAN 集成文档](./lineman/lineman_partner_integration_summary.md)

---

**最后更新**：2025-12-09  
**维护者**：TTPOS Team

