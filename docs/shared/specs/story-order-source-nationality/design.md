# 点餐方式增加订单来源-国籍选择 设计文档

> 本文档定义「点餐方式增加订单来源-国籍选择」功能的技术设计和实现方案。

## 📋 概述

本功能围绕五个子场景展开：

1. 新管理端业务设置中提供外卖来源配置（开关 + 增删改查）；
2. 新管理端业务设置中提供国籍配置（开关 + 增删改查）；
3. 收银端点餐方式增加「外卖」选项，可选择外卖来源；
4. 收银端/点餐助手点餐和桌台管理中增加国籍选择；
5. 订单详情（收银端/商家后台）显示订单来源和国籍信息。

整体方案采用「后台配置管理 → 终端选择展示 → 订单数据记录 → 订单详情显示」的链路，确保在未开启功能时，终端保持原有体验不受影响。

---

## 🎯 终端与模块影响范围总览

> 根据 `structs.mdc` 规范，明确本功能涉及的终端、用户角色和技术模块。

### 涉及终端（按 structs.mdc 终端定义）

| 终端 | 服务对象 | 影响场景 | 风险等级 |
| ---- | -------- | -------- | -------- |
| **shop** (新管理端) | 店长、运营人员 | 业务设置-外卖来源和国籍配置 | 🟢 低 |
| **pos** (收银端) | 前台收银员、店长 | 点餐方式增加外卖选项、点餐/桌台增加国籍选择、订单详情显示 | 🟡 中 |
| **assistant** (点餐助手) | 店员、收银辅助人员 | 点餐/桌台增加国籍选择 | 🟡 中 |
| **admin** (云平台管理后台) | 平台运营（如需） | 商家后台订单详情显示 | 🟢 低 |

### 涉及技术模块（按 structs.mdc 目录结构）

| 模块 | 技术栈 | 主要路径 | 变更类型 |
| ---- | ------ | -------- | -------- |
| **Admin - 店铺后台** | PHP + ThinkPHP | `admin/app/shop/` | 配置管理 API + Service + Model |
| **Main - 核心业务** | Go + Gin | `main/app/api/v1/{cashier,assistant}/` | 订单创建/查询接口扩展 |
| **前端 - 新管理端** | Flutter | `ttpos-flutter/apps/shop/` | 业务设置页增加配置界面 |
| **前端 - 收银端** | Flutter | `ttpos-flutter/apps/pos/` | 点餐方式增加外卖选项、国籍选择、订单详情 |
| **前端 - 点餐助手** | Flutter | `ttpos-flutter/apps/assistant/` | 点餐/桌台增加国籍选择 |
| **数据库** | MySQL 8.0+ | `admin/database/migrations/` | 新增配置表 + 订单表字段扩展 |

### 服务通信链路

```
新管理端配置 (admin/app/shop/) 
  ↓ 数据库写入
配置数据 (ttpos_order_source, ttpos_nationality)
  ↓ HTTP API 读取
终端展示 (main/app/api/v1/{cashier,assistant}/)
  ↓ Flutter 渲染
用户选择 (ttpos-flutter/apps/{pos,assistant}/)
  ↓ HTTP API 写入
订单数据 (ttpos_order 扩展字段)
  ↓ HTTP API 读取
订单详情展示 (ttpos-flutter/apps/{pos,shop}/)
```

### 关键约束与风险

1. **不影响现有订单流程**：未开启功能时，订单创建和详情展示保持原有逻辑。
2. **数据兼容性**：历史订单没有订单来源和国籍字段，需处理空值展示。
3. **多终端一致性**：收银端和点餐助手的国籍选择交互需保持一致。
4. **软删除保护**：删除配置后，历史订单仍需显示原名称。

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

本功能主要扩展订单模块，需遵循：

- Service 只依赖其他 Service 接口；
- Repository 只持有 db 实例；
- URL 使用 snake_case；
- data 字段必须是对象；
- 不使用 panic，统一返回 error。

### PHP 规范 (php.mdc)

配置管理模块遵循：

- 遵循 MVC 分层：Controller 只做参数获取/校验与结果返回，业务逻辑放在 Service；
- 使用验证器验证参数；
- 使用软删除（`delete_time` 字段）。

### API 设计规范 (api.mdc)

- URL 使用 snake_case，如 `/api/v1/admin/order_source`；
- 响应格式统一 `{code, message, data{}}`；
- data 字段为对象，不为 null 或数组；
- 分页信息统一放在 `meta` 中。

### 数据库规范 (database.mdc)

- 表名使用 `ttpos_` 前缀，snake_case；
- 必须包含 `id`, `uuid`, `create_time`, `update_time`, `delete_time`；
- 时间字段使用 int，默认 0；
- 按需设计索引。

---

## 🔄 代码复用分析

### 可复用的现有组件（待开发阶段具体检索）

- **订单 Model & Service**: `admin/app/shop/model/Order.php` 和 `main/app/service/order_service.go` 可扩展字段和方法。
- **配置管理模块**: 参考现有桌台类型、区域管理等配置模块的增删改查逻辑。
- **多语言系统**: 复用现有多语言翻译机制，国籍名称走系统翻译。

### 集成点

- 订单创建接口：扩展参数，接收订单来源和国籍信息；
- 订单详情接口：扩展返回字段，包含订单来源和国籍名称；
- 配置读取接口：新增外卖来源和国籍列表查询接口，供终端调用。

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

#### 新管理端-外卖来源管理（Admin 模块 - 店铺后台）

根据 `structs.mdc` Admin 模块结构：

**后端实现路径**（PHP + ThinkPHP）：
- **Model**: `admin/app/shop/model/OrderSource.php` - 外卖来源数据模型
  - 表：`ttpos_order_source`
  - 字段：`id`, `uuid`, `multi_language_name_uuid`, `sort`, `status`, `create_time`, `update_time`, `delete_time`
- **Service**: `admin/app/shop/service/OrderSourceService.php` - 外卖来源业务逻辑
  - `getList()` - 获取外卖来源列表（JOIN ttpos_multi_language_name）
  - `create($data)` - 创建外卖来源（先创建多语言名称，再创建外卖来源）
  - `update($uuid, $data)` - 更新外卖来源（更新多语言名称）
  - `delete($uuid)` - 软删除外卖来源（需校验是否有订单使用）
- **Controller**: `admin/app/shop/controller/OrderSourceController.php` - 外卖来源管理接口
  - `index()` - 获取列表
  - `create()` - 创建
  - `update()` - 更新
  - `delete()` - 删除
- **Validate**: `admin/app/shop/validate/OrderSourceValidate.php` - 参数验证器

**前端实现路径**（Flutter）：
- `ttpos-flutter/apps/shop/lib/pages/business_settings/` - 业务设置页
  - 外卖来源配置界面（开关 + 列表 + 增删改查）

**数据库迁移**：
- `admin/database/migrations/{YYYYMMDDHHMMSS}_create_ttpos_order_source_table.php`

**PHP 规范遵循**：
- ✅ Controller 只做参数获取/校验与结果返回
- ✅ 业务逻辑放在 Service 层
- ✅ 使用验证器验证参数
- ✅ 使用软删除（`delete_time` 字段）

#### 新管理端-国籍管理（Admin 模块 - 店铺后台）

**后端实现路径**（PHP + ThinkPHP）：
- **Model**: `admin/app/shop/model/Nationality.php` - 国籍数据模型
  - 表：`ttpos_nationality`
  - 字段：`id`, `uuid`, `multi_language_name_uuid`, `sort`, `status`, `create_time`, `update_time`, `delete_time`
- **Service**: `admin/app/shop/service/NationalityService.php` - 国籍业务逻辑
  - `getList()` - 获取国籍列表（JOIN ttpos_multi_language_name）
  - `create($data)` - 创建国籍（先创建多语言名称，再创建国籍）
  - `update($uuid, $data)` - 更新国籍（更新多语言名称）
  - `delete($uuid)` - 软删除国籍（需校验是否有订单使用）
- **Controller**: `admin/app/shop/controller/NationalityController.php` - 国籍管理接口
- **Validate**: `admin/app/shop/validate/NationalityValidate.php` - 参数验证器

**前端实现路径**（Flutter）：
- `ttpos-flutter/apps/shop/lib/pages/business_settings/` - 业务设置页
  - 国籍配置界面（开关 + 列表 + 增删改查）

**数据库迁移**：
- `admin/database/migrations/{YYYYMMDDHHMMSS}_create_ttpos_nationality_table.php`

#### 收银端-订单创建扩展（Main 模块）

根据 `structs.mdc` 终端与模块映射规则：

**涉及终端**：
- `pos` (收银端)：主要 API 前缀 `/api/v1/cashier`
- `assistant` (点餐助手端)：主要 API 前缀 `/api/v1/assistant`

**后端实现路径**（Go Main 模块）：
- **API 层**: 
  - `main/app/api/v1/cashier/order_api.go` - 扩展订单创建接口（增加参数）
  - `main/app/api/v1/assistant/order_api.go` - 扩展订单创建接口（增加参数）
- **Service 层**: 
  - `main/app/service/order_service.go` - 扩展订单创建方法，处理订单来源和国籍字段
  - `main/app/service/order_source_service.go` - 新增外卖来源查询服务
  - `main/app/service/nationality_service.go` - 新增国籍查询服务
- **Repository 层**: 
  - `main/app/repository/order_repository.go` - 扩展订单创建/查询方法
  - `main/app/repository/order_source_repository.go` - 新增外卖来源数据访问
  - `main/app/repository/nationality_repository.go` - 新增国籍数据访问
- **Model 层**: 
  - `main/app/model/order.go` - 扩展订单模型（增加字段）
  - `main/app/model/order_source.go` - 新增外卖来源模型
  - `main/app/model/nationality.go` - 新增国籍模型
- **DTO 层**:
  - `main/app/dto/req/order_req.go` - 扩展订单创建请求参数
  - `main/app/dto/resp/order_resp.go` - 扩展订单详情响应数据
  - `main/app/dto/resp/order_source_resp.go` - 新增外卖来源响应数据
  - `main/app/dto/resp/nationality_resp.go` - 新增国籍响应数据

**前端实现路径**（Flutter）：
- `ttpos-flutter/apps/pos/lib/pages/order/` - 收银端点餐和订单详情页
- `ttpos-flutter/apps/assistant/lib/pages/order/` - 点餐助手点餐页

**依赖规则遵循**：
- ✅ API → Service → Repository 严格分层
- ✅ Service 可依赖其他 Service（如 OrderSourceService、NationalityService）
- ❌ Service 不直接依赖 Repository
- ❌ Repository 不持有 DBManager，只持有 db 实例

#### 配置开关管理（扩展商户配置）

**可能的实现方式**：
1. 在商户配置表中新增字段：`enable_order_source`, `enable_nationality`
2. 或在独立的功能开关表中管理

**具体方案在 tasks 中细化**。

---

## 🗄️ 数据库设计（草案）

> 具体迁移在 tasks 中细化，这里给出核心表和字段草稿。

### 表 1: 外卖来源配置表

- 表：`ttpos_order_source`

核心字段：

- `id` bigint unsigned AUTO_INCREMENT
- `uuid` bigint unsigned NOT NULL DEFAULT 0
- `multi_language_name_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '多语言名称 UUID'
- `sort` int NOT NULL DEFAULT 0 COMMENT '排序'
- `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态（1=启用，0=禁用）'
- `create_time` int NOT NULL DEFAULT 0
- `update_time` int NOT NULL DEFAULT 0
- `delete_time` int NOT NULL DEFAULT 0

索引设计：

- PK: `PRIMARY KEY (id)`
- 唯一索引：`UNIQUE KEY uk_uuid (uuid)`
- 索引：`KEY idx_multi_language_name_uuid (multi_language_name_uuid)`
- 索引：`KEY idx_delete_time (delete_time)`

说明：

- 多语言名称通过关联 `ttpos_multi_language_name` 表实现
- 每个商户使用独立数据库（shop{商户ID}），无需 company_uuid 字段
- 默认预置 4 个外卖来源：Grab、Line Man、悟空外卖、Foodpanda

### 表 2: 国籍配置表

- 表：`ttpos_nationality`

核心字段：

- `id` bigint unsigned AUTO_INCREMENT
- `uuid` bigint unsigned NOT NULL DEFAULT 0
- `multi_language_name_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '多语言名称 UUID'
- `sort` int NOT NULL DEFAULT 0 COMMENT '排序'
- `status` tinyint(1) NOT NULL DEFAULT 1 COMMENT '状态（1=启用，0=禁用）'
- `create_time` int NOT NULL DEFAULT 0
- `update_time` int NOT NULL DEFAULT 0
- `delete_time` int NOT NULL DEFAULT 0

索引设计：

- PK: `PRIMARY KEY (id)`
- 唯一索引：`UNIQUE KEY uk_uuid (uuid)`
- 索引：`KEY idx_multi_language_name_uuid (multi_language_name_uuid)`
- 索引：`KEY idx_delete_time (delete_time)`

说明：

- 多语言名称通过关联 `ttpos_multi_language_name` 表实现
- 每个商户使用独立数据库（shop{商户ID}），无需 company_uuid 字段
- 默认预置 8 个国家：泰国、中国、美国、日本、韩国、英国、法国、俄罗斯

### 表 3: 订单表扩展（修改现有表）

- 表：`ttpos_order`（假设已存在）
- 新增字段：
  - `order_source_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '订单来源 UUID（0=店内，>0=外卖）';
  - `nationality_uuid` bigint unsigned NOT NULL DEFAULT 0 COMMENT '国籍 UUID（0=未记录）';

索引设计：

- 索引：`KEY idx_order_source (order_source_uuid)` - 用于渠道统计
- 索引：`KEY idx_nationality (nationality_uuid)` - 用于国籍统计

说明：

- `order_source_uuid = 0` 表示店内来源
- `nationality_uuid = 0` 表示未记录国籍
- 删除外卖来源或国籍后，订单仍保留原 UUID，通过软删除查询历史名称

### 表 4: 功能开关配置（可选方案）

- 表：可在现有配置表中新增字段，或创建独立的功能开关表
- 新增字段：
  - `enable_order_source` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否启用外卖来源';
  - `enable_nationality` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否启用国籍记录';

说明：

- 功能开关存储在商户独立数据库中
- 具体存储方案根据现有架构确定（可能是系统配置表或独立开关表）

---

## 📊 数据模型

> 具体语言实现（Go/PHP）在开发阶段按规范补充，这里仅描述逻辑模型。

### 外卖来源模型（逻辑）

```json
{
  "uuid": 123,
  "multi_language_name_uuid": 999,
  "name": "Grab",  // 从 ttpos_multi_language_name 表 JOIN 获取
  "sort": 1,
  "status": 1
}
```

### 国籍模型（逻辑）

```json
{
  "uuid": 456,
  "multi_language_name_uuid": 888,
  "name": "泰国",  // 从 ttpos_multi_language_name 表 JOIN 获取，根据当前语言返回对应翻译
  "sort": 1,
  "status": 1
}
```

### 订单扩展模型（逻辑）

```json
{
  "order_uuid": 789,
  "order_source_uuid": 123,  // 0=店内，>0=外卖来源UUID
  "order_source_name": "Grab",  // 冗余字段，方便显示（可选）
  "nationality_uuid": 456,  // 0=未记录
  "nationality_name": "泰国"  // 冗余字段，方便显示（可选）
}
```

说明：

- 是否在订单表中冗余存储名称，根据查询性能和数据一致性权衡决定
- 建议不冗余，通过 JOIN 或二次查询获取名称，保证删除后仍能查到历史名称

---

## 🔌 API 设计（草案）

以下以 REST 风格描述，实际路径按现有项目规范微调。

### Admin 新管理端-外卖来源管理

#### API 1: 获取外卖来源列表

- **URL**: `/api/v1/admin/order_source/list`
- **Method**: `GET`
- **Request Params**: 无（商户信息从登录上下文/数据库连接获取）
- **Response**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {
        "uuid": 123,
        "name": "Grab",
        "sort": 1,
        "status": 1
      }
    ]
  }
}
```

#### API 2: 创建外卖来源

- **URL**: `/api/v1/admin/order_source/create`
- **Method**: `POST`
- **Body**:

```json
{
  "names": {
    "zh": "美团外卖",
    "en": "Meituan",
    "th": "Meituan"
  },
  "sort": 5
}
```

说明：
- 后端先创建 `ttpos_multi_language_name` 记录
- 将返回的 uuid 赋值给 `multi_language_name_uuid`

- **Response**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "uuid": 124
  }
}
```

#### API 3: 更新外卖来源

- **URL**: `/api/v1/admin/order_source/update`
- **Method**: `POST`
- **Body**:

```json
{
  "uuid": 123,
  "names": {
    "zh": "Grab (更新)",
    "en": "Grab (Updated)"
  },
  "sort": 2
}
```

说明：
- 更新时会同步更新关联的 `ttpos_multi_language_name` 记录

#### API 4: 删除外卖来源

- **URL**: `/api/v1/admin/order_source/delete`
- **Method**: `POST`
- **Body**:

```json
{
  "uuid": 123
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

- 删除前需校验是否有订单使用该来源，有则返回错误提示
- 使用软删除，`delete_time` 记录删除时间

### Admin 新管理端-国籍管理

API 设计与外卖来源管理类似，路径为 `/api/v1/admin/nationality/*`

### 终端-外卖来源和国籍列表查询

#### API 5: 获取外卖来源列表（终端）

- **URL**: `/api/v1/terminal/order_source/list`
- **Method**: `GET`
- **Params**: 无（商户信息从登录上下文获取）
- **Response**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {
        "uuid": 123,
        "name": "Grab"
      }
    ]
  }
}
```

说明：

- 终端只需要 uuid 和 name，不需要其他字段
- name 从 `ttpos_multi_language_name` 表 JOIN 获取，根据当前语言返回对应翻译
- 按 sort 排序返回
- 只返回 status=1 且 delete_time=0 的外卖来源

#### API 6: 获取国籍列表（终端）

API 设计与外卖来源列表类似，路径为 `/api/v1/terminal/nationality/list`

### 终端-订单创建扩展

#### API 7: 创建订单（扩展参数）

- **URL**: `/api/v1/cashier/order/create` 或 `/api/v1/assistant/order/create`
- **Method**: `POST`
- **Body**（扩展字段）:

```json
{
  "table_uuid": 111,
  "items": [...],
  "order_source_uuid": 123,  // 新增：0=店内，>0=外卖来源UUID
  "nationality_uuid": 456    // 新增：0=未记录
}
```

- **Response**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "order_uuid": 789
  }
}
```

### 终端-订单详情扩展

#### API 8: 获取订单详情（扩展字段）

- **URL**: `/api/v1/cashier/order/detail` 或 `/api/v1/assistant/order/detail`
- **Method**: `GET`
- **Params**:
  - `order_uuid`
- **Response**（扩展字段）:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "order_uuid": 789,
    "table_name": "A01",
    "total_amount": 100.00,
    "order_source": {
      "uuid": 123,
      "name": "Grab"
    },
    "nationality": {
      "uuid": 456,
      "name": "泰国"
    }
  }
}
```

说明：

- 如果 `order_source_uuid = 0`，则 `order_source` 返回 `{"uuid": 0, "name": "店内"}` 或不返回该字段
- 如果 `nationality_uuid = 0`，则 `nationality` 返回 `{"uuid": 0, "name": "未记录"}` 或不返回该字段
- 如果外卖来源或国籍已被删除，仍需返回历史名称（通过软删除查询）

---

## 🧩 组件和接口（高层）

### 后端 Service（示例：PHP）

- `OrderSourceService`
  - `getList()` - 获取外卖来源列表（JOIN ttpos_multi_language_name）
  - `create($data)` - 创建外卖来源（先创建多语言名称）
  - `update($uuid, $data)` - 更新外卖来源（更新多语言名称）
  - `delete($uuid)` - 软删除外卖来源
  - `checkCanDelete($uuid)` - 校验是否可删除（是否有订单使用）

- `NationalityService`
  - 方法同 OrderSourceService，处理国籍数据

- `OrderService`（Go Main）
  - `CreateOrder(req *dto.CreateOrderReq)` - 扩展订单创建方法
  - `GetOrderDetail(orderUuid uint64)` - 扩展订单详情查询

### 前端组件（新管理端）

- 页面：`BusinessSettingsPage`
  - 子组件：
    - `OrderSourceConfig`：外卖来源配置（开关 + 列表 + 增删改查）
    - `NationalityConfig`：国籍配置（开关 + 列表 + 增删改查）

### 前端组件（收银端）

- 页面：`OrderCreatePage`
  - 子组件：
    - `OrderSourceSelector`：外卖来源选择组件
    - `NationalitySelector`：国籍选择组件

- 页面：`OrderDetailPage`
  - 扩展显示字段：订单来源、国籍

---

## ⚡ 缓存与性能

- 外卖来源和国籍配置读多写少，可在终端侧增加缓存（Redis），Key 形如：
  - `ttpos:order_source:list:{company_uuid}`
  - `ttpos:nationality:list:{company_uuid}`
- 缓存过期时间：5 分钟或配置变更后主动清除
- 订单创建时不需要缓存，直接写入数据库

---

## 🚨 错误处理与安全

- 后端接口严格校验 `company_uuid`，避免跨商家访问；
- 所有接口需通过现有鉴权中间件；
- 删除配置时，需校验是否有订单使用，避免数据丢失；
- 参数校验防止 SQL 注入、XSS 攻击；
- 出错时记录详细日志，返回统一错误码与友好提示。

---

## 🧪 测试策略（概要）

- 单元测试：对 `OrderSourceService`、`NationalityService`、`OrderService` 主要方法进行覆盖；
- API 测试：覆盖配置管理接口、订单创建/详情接口；
- 集成测试：模拟从后台配置 → 终端选择 → 订单创建 → 订单详情展示的完整链路；
- 数据兼容性测试：测试历史订单在各终端的显示效果；
- 多语言测试：测试国籍名称在不同语言环境下的显示。

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

