# 新管理端-桌面端-桌台地图 设计文档

> 本文档定义「新管理端-桌面端-桌台地图」功能的技术设计和实现方案。

## 📋 概述

本功能围绕三个子场景展开：

1. 新管理端（桌面端）提供桌台地图配置画布（区域维度）；
2. 收银机/点餐助手终端按配置展示桌台地图模式；
3. 云平台商家管理提供「桌台地图」「数据管理」能力开关。

整体方案采用「云平台能力开关 → 配置端布局建模 → 终端渲染与交互」的链路，确保在不开启或未配置地图时，终端保持原有桌台列表体验不受影响。

---

## 🎯 终端与模块影响范围总览

> 根据 `structs.mdc` 规范，明确本功能涉及的终端、用户角色和技术模块。

### 涉及终端（按 structs.mdc 终端定义）

| 终端 | 服务对象 | 影响场景 | 风险等级 |
| ---- | -------- | -------- | -------- |
| **shop** (新管理端-桌面端) | 店长、运营人员、区域经理 | 桌台地图配置画布（核心功能） | 🟡 中 |
| **pos** (收银端) | 前台收银员、店长 | 桌台地图模式展示与切换 | 🟡 中 |
| **assistant** (点餐助手) | 店员、收银辅助人员 | 桌台地图模式展示与切换 | 🟡 中 |
| **admin** (云平台管理后台) | 平台运营、实施人员 | 商家能力开关配置 | 🟢 低 |

### 涉及技术模块（按 structs.mdc 目录结构）

| 模块 | 技术栈 | 主要路径 | 变更类型 |
| ---- | ------ | -------- | -------- |
| **Admin - 管理后台** | PHP + ThinkPHP | `admin/app/admin/` | 商家表字段扩展 + 开关控件 |
| **Admin - 店铺后台** | PHP + ThinkPHP | `admin/app/shop/` | 桌台地图配置 API + Service |
| **Main - 核心业务** | Go + Gin | `main/app/api/v1/{cashier,assistant}/` | 终端桌台地图布局读取接口 |
| **前端 - 管理后台** | Vue 3 | `admin/views/admin/` | 商家编辑页开关控件 |
| **前端 - 新管理端** | Flutter | `ttpos-flutter/apps/shop/` | 桌台地图配置页面与画布编辑 |
| **前端 - 收银端** | Flutter | `ttpos-flutter/apps/pos/` | 地图模式组件 |
| **前端 - 点餐助手** | Flutter | `ttpos-flutter/apps/assistant/` | 地图模式组件 |
| **数据库** | MySQL 8.0+ | `admin/database/migrations/` | 新增布局表 + 商家表字段 |

### 服务通信链路

```
云平台 (admin/app/admin/) 
  ↓ HTTP API
新管理端配置 (admin/app/shop/) 
  ↓ 数据库写入
布局数据 (ttpos_desk_map_layout)
  ↓ HTTP API 读取
终端展示 (main/app/api/v1/{cashier,assistant}/)
  ↓ Flutter 渲染
用户交互 (ttpos-flutter/apps/{pos,assistant}/)
```

### 关键约束与风险

1. **不影响现有桌台列表功能**：未开启或未配置地图时，终端保持原有体验。
2. **多终端一致性**：配置端保存的布局需在多个终端（pos/assistant）保持一致渲染。
3. **性能考量**：大桌量场景（200+ 桌）需要前端虚拟化渲染和后端缓存优化。
4. **权限控制**：云平台开关需严格控制，避免误开启导致未完成配置的商家看到空白地图。

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

本功能主要落在 Admin + 前端，若后续需要 Main 参与（如桌台状态透传），需遵循：

- Service 只依赖其他 Service 接口；
- Repository 只持有 db 实例；
- URL 使用 snake_case；
- data 字段必须是对象；
- 不使用 panic，统一返回 error。

### PHP 规范 (php.mdc)

- 遵循 MVC 分层：Controller 只做参数获取/校验与结果返回，业务逻辑放在 Service；
- 使用验证器验证参数；
- 使用软删除（如涉及表的删除操作）。

### API 设计规范 (api.mdc)

- URL 使用 snake_case，如 `/api/v1/admin/desk_map/layout`；
- 响应格式统一 `{code, message, data{}}`；
- data 字段为对象，不为 null 或数组；
- 分页信息统一放在 `meta` 中。

### 数据库规范 (database.mdc)

- 表名使用 `ttpos_` 前缀，snake_case；
- 必须包含 `id`, `uuid`, `create_time`, `update_time`, `delete_time`；
- 时间字段使用 int，默认 0；
- 按需设计索引。

---

### 集成点

- 云平台商家表：在现有商家信息表基础上新增能力开关字段；
- Mian 新管理端：在餐厅/桌台管理模块下挂载「桌台地图」配置功能；
- 终端（收银机/点餐助手）：在现有桌台列表页增加地图模式入口，并新增布局数据加载接口。

---

## 🏗️ 架构设计

### 分层设计原则

后端仍采用标准三层架构：

```
API 层 (Controller/API)
  ↓ 依赖
业务层 (Service)
  ↓ 依赖
数据层 (Repository)
```

依赖规则：

- 上层可依赖下层；
- 禁止下层依赖上层；
- Service 不直接依赖 DBManager（Go 场景）；PHP 中 Service 不直接操作请求上下文。

### 模块划分（初步）

#### 云平台（Admin 模块 - 管理后台）

根据 `structs.mdc` Admin 模块结构：

**后端实现路径**（PHP + ThinkPHP）：
- **Model**: `admin/app/admin/model/Merchant.php` - 扩展商家表字段
  - 新增：`enable_desk_map`, `enable_data_management`
- **Service**: `admin/app/admin/service/MerchantService.php` - 商家能力开关业务逻辑
- **Controller**: `admin/app/admin/controller/MerchantController.php` - 商家管理接口
- **Validate**: `admin/app/admin/validate/MerchantValidate.php` - 参数验证器

**前端实现路径**（Vue 3）：
- `admin/views/admin/pages/merchant/edit.vue` - 商家编辑页增加开关控件
- `admin/views/admin/components/` - 可复用的开关组件

**数据库迁移**：
- `admin/database/migrations/{YYYYMMDDHHMMSS}_alter_ttpos_merchant_add_desk_map_fields.php`

#### 新管理端-桌面端（餐厅管理）

根据 `structs.mdc` 新管理端映射规则：

**涉及终端**：
- `shop` (商家/门店管理端)：**新管理端-桌面端**（在 `ttpos-flutter/apps/shop`），不是旧的 `admin/views/shop`

**后端实现路径**（Go Main 模块）：
- **API 前缀**: `/api/v1/shop/desk_map`
- **API 层**: `main/app/api/v1/shop/desk_map_api.go`
  - `GetAreaList()`：获取区域+状态列表
  - `GetLayoutDetail()`：获取某区域桌台基础信息 + 布局信息
  - `SaveLayout()`：保存布局
- **Service 层**: `main/app/service/desk_map_service.go` - 桌台地图配置业务逻辑
- **Repository 层**: `main/app/repository/desk_map_repository.go` - 桌台布局数据访问
- **Model**: `main/app/model/desk_map_layout.go` - 桌台布局数据模型
- **DTO**: 
  - `main/app/dto/req/desk_map_req.go` - 请求参数
  - `main/app/dto/resp/desk_map_resp.go` - 响应数据

**前端实现路径**（ttpos-flutter 仓库）：
- `ttpos-flutter/apps/shop/` - 新管理端-桌面端
  - 桌台地图配置页面与画布编辑组件

**数据库迁移**：
- `admin/database/migrations/{YYYYMMDDHHMMSS}_create_desk_map_layout_table.php`

**Go Main 规范遵循**：
- ✅ API 层只做参数绑定/校验与结果返回
- ✅ 业务逻辑放在 Service 层
- ✅ 数据访问通过 Repository 层
- ✅ 使用软删除（`delete_time` 字段）
- ✅ 遵循三层依赖架构：API → Service → Repository

#### 终端（收银机/点餐助手）

根据 `structs.mdc` 终端与模块映射规则：

**涉及终端**：
- `pos` (收银端)：主要 API 前缀 `/api/v1/cashier`，路由分组 `main/app/api/v1/cashier`
- `assistant` (点餐助手端)：主要 API 前缀 `/api/v1/assistant`，路由分组 `main/app/api/v1/assistant`

**后端实现路径**（Go Main 模块）：
- **API 层**: 
  - `main/app/api/v1/cashier/desk_map_api.go` - 收银端桌台地图接口
  - `main/app/api/v1/assistant/desk_map_api.go` - 点餐助手端桌台地图接口
- **Service 层**: 
  - `main/app/service/desk_map_service.go` - 桌台地图业务逻辑（可复用现有 `table_service.go`）
- **Repository 层**: 
  - `main/app/repository/desk_map_repository.go` - 桌台布局数据访问
- **Model 层**: 
  - `main/app/model/desk_map_layout.go` - 桌台布局数据模型
- **DTO 层**:
  - `main/app/dto/req/desk_map_req.go` - 请求参数
  - `main/app/dto/resp/desk_map_resp.go` - 响应数据

**前端实现路径**（ttpos-flutter 仓库）：
- `ttpos-flutter/apps/pos/` - 收银端地图模式组件
- `ttpos-flutter/apps/assistant/` - 点餐助手端地图模式组件

**依赖规则遵循**：
- ✅ API → Service → Repository 严格分层
- ✅ Service 可依赖其他 Service（如 TableService、AreaService）
- ❌ Service 不直接依赖 Repository
- ❌ Repository 不持有 DBManager，只持有 db 实例

---

## 🗄️ 数据库设计（草案）

> 具体迁移在 tasks 中细化，这里给出核心表和字段草稿。

### 表 1: 商家能力开关（扩展现有商家表）

- 表：`ttpos_merchant`（假设）
- 新增字段：
  - `enable_desk_map` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否启用桌台地图能力';
  - `enable_data_management` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否启用数据管理能力';

索引：  
- 若已有主键/业务索引，此处不新增独立索引，可在查询中复用。

### 表 2: 区域桌台布局表

- 表：`ttpos_desk_map_layout`

核心字段（示例）：

- `id` bigint unsigned AUTO_INCREMENT
- `uuid` bigint unsigned NOT NULL DEFAULT 0
- `company_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '商户 UUID'
- `region_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '区域 UUID'
- `layout_json` text NOT NULL COMMENT '画布布局 JSON（含桌台坐标、尺寸、样式等）'
- `create_time` int NOT NULL DEFAULT 0
- `update_time` int NOT NULL DEFAULT 0
- `delete_time` int NOT NULL DEFAULT 0

索引设计：

- PK: `PRIMARY KEY (id)`
- 唯一索引：`UNIQUE KEY uk_company_area (company_uuid, region_uuid)`

说明：

- 布局以 JSON 形式存储，前端负责解析/渲染，后端只负责版本管理和基本合法性检查。

---

## 📊 数据模型

> 具体语言实现（Go/PHP）在开发阶段按规范补充，这里仅描述逻辑模型。

### 区域布局模型（逻辑）

```json
{
  "region_uuid": 123,
  "tables": [
    {
      "table_uuid": 111,
      "shape": "circle",         // circle / rect
      "capacity": 4,
      "x": 100,
      "y": 200,
      "width": 80,
      "height": 80,
      "rotation": 0
    }
  ]
}
```

说明：

- `capacity` 默认等于该桌台类型最大就餐人数；
- 坐标系由前端约定（如以画布左上角为原点），后端不做几何运算，仅存取。

---

## 🔌 API 设计（草案）

以下以 REST 风格描述，实际路径按现有项目规范微调。

### Admin 新管理端-桌台地图配置

#### API 1: 获取区域列表及地图配置状态

- **URL**: `/api/v1/admin/desk_map/areas`
- **Method**: `GET`
- **Request Params**:
  - `company_uuid`（可从登录上下文获取）
- **Response**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {
        "region_uuid": 123,
        "area_name": "大厅",
        "table_count": 20,
        "layout_status": "set"   // set / unset
      }
    ]
  }
}
```

#### API 2: 获取某区域布局详情

- **URL**: `/api/v1/admin/desk_map/layout_detail`
- **Method**: `GET`
- **Request Params**:
  - `region_uuid` (required)
- **Response**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "area": {
      "region_uuid": 123,
      "area_name": "大厅"
    },
    "tables": [
      {
        "table_uuid": 111,
        "table_name": "A01",
        "capacity": 4,
        "selected": true
      }
    ],
    "layout": {
      "tables": [ /* 见逻辑模型 */ ]
    }
  }
}
```

#### API 3: 保存区域布局

- **URL**: `/api/v1/admin/desk_map/save_layout`
- **Method**: `POST`
- **Body**:

```json
{
  "region_uuid": 123,
  "layout": {
    "tables": [
      {
        "table_uuid": 111,
        "shape": "circle",
        "capacity": 4,
        "x": 100,
        "y": 200,
        "width": 80,
        "height": 80,
        "rotation": 0
      }
    ]
  }
}
```

- **Response**:

```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

校验要点：

- `layout.tables` 非空；
- 所有 `table_uuid` 属于该区域已有桌台；
- 坐标/尺寸在合理范围内（如非负）。

### 终端-桌台地图模式

#### API 4: 获取桌台地图布局（终端）

- **URL**: `/api/v1/terminal/desk_map/layout`
- **Method**: `GET`
- **Params**:
  - `region_uuid`（可选，默认所有区域）
- **Response**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "areas": [ /* 区域列表 */ ],
    "layout": {
      "tables": [ /* 布局 + 状态，如空闲/占用 */ ]
    }
  }
}
```

说明：

- 若商家未开启「桌台地图」或没有配置布局，则可返回空布局并由前端隐藏地图模式入口。

---

## 🧩 组件和接口（高层）

### 后端 Service（示例：PHP）

- `DeskMapService`
  - `getAreaListWithStatus($companyUuid)`
  - `getAreaLayoutDetail($companyUuid, $areaUuid)`
  - `saveAreaLayout($companyUuid, $areaUuid, $layoutData)`

### 前端组件（新管理端）

- 页面：`AdminDeskMapPage`
  - 子组件：
    - `AreaListPanel`：区域列表及状态；
    - `TableListPanel`：区域桌台列表 + 勾选；
    - `CanvasEditor`：画布编辑（拖拽、缩放、样式切换）。

### 前端组件（终端）

- 页面：`DeskMapView`
  - 提供模式切换（列表/地图）；
  - 使用已有桌台状态颜色体系，避免新增一套。

---

## ⚡ 缓存与性能

- 布局数据读多写少，可在终端侧增加短期缓存（如内存/本地存储），但后端暂不强制 Redis 缓存；
- 若后续性能瓶颈明显，可为布局查询增加 Redis 缓存，Key 形如：`ttpos:desk_map:layout:{company_uuid}:{region_uuid}`。

---

## 🚨 错误处理与安全

- 后端接口严格校验 `company_uuid`、`region_uuid` 权限，避免跨商家/跨门店访问；
- 所有接口需通过现有鉴权中间件；
- 对布局 JSON 进行结构校验，防止异常数据导致终端渲染崩溃；
- 出错时记录详细日志，返回统一错误码与友好提示。

---

## 🧪 测试策略（概要）

- 单元测试：对 `DeskMapService` 主要方法进行覆盖，验证布局读写与状态判断逻辑；
- API 测试：覆盖 Admin 配置接口与终端布局获取接口；
- 集成测试：模拟从云平台开启开关 → Admin 配置布局 → 终端拉取并展示的完整链路；
- 前端：对画布核心交互（拖拽、缩放、保存）进行 E2E 或手动测试。

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-19  
**作者**: {团队/个人}  
**审核者**: {审核者}


