# 新管理端商品管理增加外卖商品模块 需求文档

> 本文档定义 Shop 端外卖商品管理功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                    |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.11.0-shop-grab-integration.md](../../../../team/proposals/2025-12/v2.11.0-shop-grab-integration.md) |
| **创建日期**      | 2025-12-08                                                                                                              |
| **负责人**        | weifashi                                                                                                                |
| **目标 Sprint**   | Sprint TBD                                                                                                              |
| **目标版本**      | v2.11.0                                                                                                                 |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [x] Vue (admin/views/)                                              |
| **关联任务**      | DooTask #37501                                                                                                          |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | 已通过     |
| **审核人**   | weifashi   |
| **审核日期** | 2025-12-09 |
| **审核意见** | 后端接口已实现，继续创建设计文档 |

---

## 📋 概述

在新管理端（Go Main 模块）商品添加/编辑页面增加外卖商品管理功能，实现商品在"店内"和"外卖"（如 Grab）两个渠道的独立配置。

**核心功能**：
1. **外卖商品添加**：基于店内商品创建外卖商品配置（独立的分类、规格价格、上下架状态）
2. **外卖商品编辑**：编辑外卖商品的专属信息
3. **多平台支持**：通过 `takeout_type` 字段支持多种外卖平台（Grab、FoodPanda 等）

## 🎯 产品对齐

- **差异化定价**：外卖商品可设置与店内不同的价格
- **独立管理**：外卖商品有独立的分类、状态控制
- **多平台扩展**：预留外卖类型字段，支持未来接入更多外卖平台

## 📝 用户故事

**作为** 商户管理员  
**我想** 在商品添加/编辑页面切换到外卖 Tab 配置外卖商品信息  
**以便于** 为外卖平台设置独立的分类、价格和上下架状态

---

## 功能需求

### Requirement 1: 外卖商品添加

**用户故事**: 作为商户管理员，我想为店内商品创建外卖配置，以便于在外卖平台上架销售

#### 验收标准

1. **WHEN** 商户在商品添加页面切换到外卖 Tab **THEN** 系统 **SHALL** 显示外卖商品配置表单
2. **WHEN** 商户填写外卖商品信息并保存 **THEN** 系统 **SHALL** 创建外卖商品记录
3. **WHEN** 同一商品已存在相同类型的外卖配置 **THEN** 系统 **SHALL** 提示错误

#### 具体要求

- [x] 1.1 创建 `ttpos_product_package_takeout` 表存储外卖商品信息
- [x] 1.2 实现 `/shop/product/takeout/add` 接口
- [ ] 1.3 前端商品添加页面增加外卖 Tab 切换
- [ ] 1.4 外卖商品默认状态为"下架"

---

### Requirement 2: 外卖商品编辑

**用户故事**: 作为商户管理员，我想编辑外卖商品的配置，以便于调整外卖价格和上下架状态

#### 验收标准

1. **WHEN** 商户在商品编辑页面切换到外卖 Tab **THEN** 系统 **SHALL** 显示已有的外卖配置
2. **WHEN** 商户修改外卖商品信息并保存 **THEN** 系统 **SHALL** 更新外卖商品记录
3. **WHEN** 外卖商品不存在 **THEN** 系统 **SHALL** 提示用户先创建

#### 具体要求

- [x] 2.1 实现 `/shop/product/takeout/edit` 接口
- [x] 2.2 实现 `/shop/product/takeout/detail` 接口
- [x] 2.3 实现 `/shop/product/takeout/status` 接口
- [x] 2.4 实现 `/shop/product/takeout/delete` 接口
- [ ] 2.5 前端商品编辑页面增加外卖 Tab 切换

---

## 数据库设计

### 新增表：`ttpos_product_package_takeout`

外卖商品表，存储商品的外卖专属信息。

| 字段                     | 类型           | 说明                                        |
| ------------------------ | -------------- | ------------------------------------------- |
| `id`                     | bigint         | 自增ID                                      |
| `uuid`                   | bigint         | UUID                                        |
| `product_package_uuid`   | bigint         | 关联店内商品UUID                            |
| `name`                   | text           | 外卖商品名称                                |
| `multi_language_name_uuid` | bigint       | 多语言名称ID                                |
| `product_type`           | int            | 商品类型 0-商品 1-套餐                      |
| `takeout_type`           | int            | **外卖类型 1-Grab 2-FoodPanda 3-其他**      |
| `status`                 | int            | 外卖状态 0-下架 1-上架                      |
| `category_uuid`          | bigint         | 外卖分类UUID                                |
| `special_category_uuid`  | bigint         | 外卖特色分类UUID                            |
| `image_file_uuid`        | bigint         | 外卖商品图片UUID                            |
| `create_time`            | int            | 创建时间                                    |
| `update_time`            | int            | 更新时间                                    |
| `delete_time`            | int            | 删除时间                                    |

**索引**：
- `idx_uuid` (UNIQUE)
- `idx_product_package_takeout_type` (UNIQUE) - 同一商品同一外卖类型只能有一条记录

**说明**：
- 规格信息共用 `ttpos_product_bom` 表
- 通过 `product_package_uuid` 关联店内商品

---

## API 接口

### 已实现接口

| 方法   | 路径                          | 说明             |
| ------ | ----------------------------- | ---------------- |
| POST   | `/shop/product/takeout/add`    | 添加外卖商品     |
| POST   | `/shop/product/takeout/edit`   | 编辑外卖商品     |
| GET    | `/shop/product/takeout/detail` | 获取外卖商品详情 |
| DELETE | `/shop/product/takeout/delete` | 删除外卖商品     |
| POST   | `/shop/product/takeout/status` | 修改外卖商品状态 |

### 请求/响应示例

#### 添加外卖商品

```json
// POST /shop/product/takeout/add
{
  "product_package_uuid": 123456789,
  "takeout_type": 1,
  "category_uuid": 111,
  "special_category_uuid": 222,
  "status": 0,
  "image_file_uuid": 333,
  "flavors": [
    { "bom_uuid": 444, "price": 15.00 },
    { "bom_uuid": 555, "price": 20.00 }
  ]
}
```

#### 获取外卖商品详情

```json
// GET /shop/product/takeout/detail?uuid=123456789
{
  "code": 0,
  "message": "success",
  "data": {
    "uuid": 123456789,
    "product_package_uuid": 987654321,
    "takeout_type": 1,
    "locale_name": { "zh": "宫保鸡丁", "en": "Kung Pao Chicken" },
    "category_uuid": 111,
    "category_name": { "zh": "热菜", "en": "Hot Dishes" },
    "status": 1,
    "image_url": "https://...",
    "flavors": [
      { "bom_uuid": 444, "locale_name": { "zh": "小份" }, "price": 15.00 },
      { "bom_uuid": 555, "locale_name": { "zh": "大份" }, "price": 20.00 }
    ]
  }
}
```

---

## 代码结构

### 已实现文件

```
main/app/
├── model/
│   └── product_package_takeout.go      # 外卖商品模型
├── dto/
│   ├── req/product_takeout.go          # 请求 DTO
│   └── resp/product_takeout.go         # 响应 DTO
├── repository/
│   └── product_package_takeout.go      # 数据访问层
├── service/
│   └── product_takeout.go              # 业务逻辑层
├── api/v1/shop/
│   └── shop_product_takeout.go         # API Handler
└── constant/
    └── product.go                      # 外卖类型常量

admin/database/
├── migrations/
│   └── 20251208232558_create_product_package_takeout_table.php
└── seeds/
    └── shop_01.sql                     # 表结构
```

### 外卖类型常量

```go
// main/app/constant/product.go
const (
    TakeoutTypeGrab      = 1 // Grab
    TakeoutTypeFoodPanda = 2 // FoodPanda
    TakeoutTypeOther     = 3 // 其他（预留扩展）
)
```

---

## 验收标准

### 功能验收

1. **添加外卖商品**: 能够为店内商品创建外卖配置
2. **编辑外卖商品**: 能够修改外卖商品的分类、状态、图片
3. **状态控制**: 能够独立控制外卖商品的上下架状态
4. **多平台支持**: 同一商品可以有多种外卖平台的配置

### 测试验收

1. **单元测试**: Service 层测试覆盖
2. **API 测试**: 所有接口测试通过
3. **边界测试**: 重复添加、不存在的商品等边界情况

---

## 约束条件

### 技术约束

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- 不使用 panic，返回 error

### 业务约束

- 外卖商品必须关联一个店内商品
- 同一商品同一外卖类型只能有一条外卖配置
- 外卖商品默认状态为"下架"

---

## 涉及终端

- [x] Shop 商家管理端（配置外卖商品）
- [ ] POS 收银端
- [ ] 其他终端

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范

### 相关文件

- `main/app/service/product.go` - 店内商品服务（参考）
- `main/app/api/v1/shop/shop_product.go` - 店内商品 API（参考）

---

**版本**: v1.1.0  
**创建日期**: 2025-12-08  
**更新日期**: 2025-12-08  
**作者**: weifashi  
**审核者**: 待定
