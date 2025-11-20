# 点餐方式增加订单来源-国籍选择 任务分解

> 本文档定义「点餐方式增加订单来源-国籍选择」功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求（如 R1.1, R2.3）

## 📊 进度总览

**总任务数**: 25（Phase 1: 6 + Phase 2: 6 + Phase 3: 5 + Phase 4: 4 + Phase 5: 4）  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

> 实际开发过程中，请在完成任务时同步维护以上统计，并更新对应任务的勾选状态。

---

## Phase 1: 数据库设计与迁移

> 根据 `database.mdc` 规范，采用多租户架构（每个商户独立数据库），多语言通过 `ttpos_multi_language_name` 表实现

- [ ] 1.1 创建外卖来源配置表迁移

  - **Module**: Admin - 数据库迁移
  - **File**: `admin/database/migrations/{YYYYMMDDHHMMSS}_create_ttpos_order_source_table.php`
  - **Purpose**: 定义 `ttpos_order_source` 表结构
  - **Requirements**: R1.1-R1.6
  - **Leverage**: 模板: `docs/agent/templates/database-migration-template.md`；参考现有配置表结构
  - **Key Points**: 
    - 必须包含标准字段：`id`, `uuid`, `create_time`, `update_time`, `delete_time`
    - 包含 `multi_language_name_uuid` 关联多语言表
    - 不需要 `company_uuid`（多租户架构）
    - 不需要 `name_lang`（使用多语言表）
    - 时间戳使用 shell 命令 `date +%Y%m%d%H%M%S` 生成
  - **Prompt**: Role: Database Engineer | Task: 创建外卖来源配置表，包含 uuid, multi_language_name_uuid, sort, status 等字段 | Context: 遵循数据库规范，包含标准字段，关联多语言表 | Restrictions: 遵循 `.cursor/rules/database.mdc`，不包含 company_uuid 和 name_lang | Success: 表创建成功，索引合理，可重复执行

- [ ] 1.2 创建国籍配置表迁移

  - **Module**: Admin - 数据库迁移
  - **File**: `admin/database/migrations/{YYYYMMDDHHMMSS}_create_ttpos_nationality_table.php`
  - **Purpose**: 定义 `ttpos_nationality` 表结构
  - **Requirements**: R2.1-R2.7
  - **Leverage**: 1.1 的表结构模式，参数几乎一致
  - **Key Points**: 同 1.1
  - **Prompt**: Role: Database Engineer | Task: 创建国籍配置表，结构与外卖来源表类似 | Context: 遵循数据库规范，关联多语言表 | Restrictions: 遵循 `.cursor/rules/database.mdc` | Success: 表创建成功

- [ ] 1.3 扩展订单表字段迁移

  - **Module**: Admin - 数据库迁移
  - **File**: `admin/database/migrations/{YYYYMMDDHHMMSS}_alter_ttpos_order_add_source_nationality_fields.php`
  - **Purpose**: 为 `ttpos_order` 表新增 `order_source_uuid` 和 `nationality_uuid` 字段
  - **Requirements**: R3, R4, R5
  - **Leverage**: 现有订单表结构
  - **Key Points**: 
    - 字段类型 bigint unsigned，默认值 0
    - 0 表示未记录（店内/未选择）
    - 添加索引用于统计查询
  - **Prompt**: Role: Database Engineer | Task: 为订单表新增订单来源和国籍字段 | Context: 扩展现有表，默认值 0，兼容历史数据 | Restrictions: 遵循 `.cursor/rules/database.mdc`，谨慎处理生产环境影响 | Success: 字段新增成功，现有数据不受影响

- [ ] 1.4 预置默认数据（外卖来源）

  - **Module**: Admin - 数据库 Seed
  - **File**: `admin/database/seeds/order_source_default.sql` 或在迁移中处理
  - **Purpose**: 预置 4 个外卖来源：Grab、Line Man、悟空外卖、Foodpanda
  - **Requirements**: R1.4
  - **Key Points**: 
    - 先创建多语言名称记录
    - 然后创建外卖来源记录并关联 multi_language_name_uuid
    - 检查是否已存在，避免重复插入

- [ ] 1.5 预置默认数据（国籍）

  - **Module**: Admin - 数据库 Seed
  - **File**: `admin/database/seeds/nationality_default.sql` 或在迁移中处理
  - **Purpose**: 预置 8 个国家：泰国、中国、美国、日本、韩国、英国、法国、俄罗斯
  - **Requirements**: R2.4
  - **Key Points**: 同 1.4

- [ ] 1.6 功能开关配置

  - **Module**: Admin - 数据库迁移或配置
  - **Purpose**: 确定功能开关存储方案，新增 `enable_order_source` 和 `enable_nationality` 字段
  - **Requirements**: R1.1, R2.1
  - **Key Points**: 根据现有系统配置架构确定具体实现方式

---

## Phase 2: 后台配置管理（Admin 模块 - PHP）

> 根据 `structs.mdc`，本阶段涉及 **Admin 模块 - 店铺后台**（PHP + ThinkPHP）：`admin/app/shop/`

- [ ] 2.1 新增外卖来源 Model

  - **Module**: Admin - 店铺后台 Model
  - **File**: `admin/app/shop/model/OrderSource.php`
  - **Purpose**: 映射 `ttpos_order_source` 表
  - **Requirements**: R1.1-R1.6
  - **Leverage**: 现有类似配置类 Model
  - **Prompt**: Role: PHP Developer | Task: 创建 OrderSource Model 映射配置表 | Context: 使用 ThinkPHP Model 规范，关联多语言表 | Restrictions: 遵循 `.cursor/rules/php.mdc` | Success: Model 可正常读写配置数据

- [ ] 2.2 新增外卖来源 Service

  - **Module**: Admin - 店铺后台 Service
  - **File**: `admin/app/shop/service/OrderSourceService.php`
  - **Purpose**: 封装外卖来源增删改查逻辑
  - **Requirements**: R1.1-R1.6
  - **Leverage**: 现有配置管理 Service 模式
  - **Key Points**: 
    - `getList()` - JOIN ttpos_multi_language_name 获取名称
    - `create($data)` - 先创建多语言名称，再创建外卖来源
    - `update($uuid, $data)` - 更新多语言名称
    - `delete($uuid)` - 软删除前校验是否有订单使用
  - **Prompt**: Role: PHP Developer | Task: 创建 OrderSourceService，实现配置管理业务逻辑 | Context: 业务逻辑放在 Service 层，依赖多语言名称 Service | Restrictions: 遵循 `.cursor/rules/php.mdc`，Controller 不包含业务逻辑 | Success: Service 方法可正常调用

- [ ] 2.3 新增外卖来源 Controller 和路由

  - **Module**: Admin - 店铺后台 Controller
  - **File**: `admin/app/shop/controller/OrderSourceController.php`
  - **Purpose**: 提供外卖来源管理 API（list, create, update, delete）
  - **Requirements**: R1.1-R1.6
  - **Leverage**: 现有 Controller 示例
  - **Prompt**: Role: PHP Developer | Task: 创建 OrderSourceController，实现配置管理接口 | Context: Controller 只做参数获取/校验与结果返回 | Restrictions: 遵循 `.cursor/rules/php.mdc` 和 `.cursor/rules/api.mdc` | Success: API 接口可正常调用

- [ ] 2.4 新增国籍 Model

  - **Module**: Admin - 店铺后台 Model
  - **File**: `admin/app/shop/model/Nationality.php`
  - **Purpose**: 映射 `ttpos_nationality` 表
  - **Requirements**: R2.1-R2.7
  - **Leverage**: 2.1 的 Model 结构，几乎一致

- [ ] 2.5 新增国籍 Service

  - **Module**: Admin - 店铺后台 Service
  - **File**: `admin/app/shop/service/NationalityService.php`
  - **Purpose**: 封装国籍增删改查逻辑
  - **Requirements**: R2.1-R2.7
  - **Leverage**: 2.2 的 Service 模式

- [ ] 2.6 新增国籍 Controller 和路由

  - **Module**: Admin - 店铺后台 Controller
  - **File**: `admin/app/shop/controller/NationalityController.php`
  - **Purpose**: 提供国籍管理 API（list, create, update, delete）
  - **Requirements**: R2.1-R2.7
  - **Leverage**: 2.3 的 Controller 模式

---

## Phase 3: 终端订单创建扩展（Main 模块 - Go）

> 根据 `structs.mdc`，本阶段涉及：
> - **Main 模块**（Go + Gin）：`main/app/api/v1/{cashier,assistant}/`
> - 扩展订单创建接口，增加订单来源和国籍参数

- [ ] 3.1 扩展订单 Model（Go Main）

  - **Module**: Main - Model 层
  - **File**: `main/app/model/order.go`
  - **Purpose**: 扩展订单模型，增加 `OrderSourceUuid` 和 `NationalityUuid` 字段
  - **Requirements**: R3, R4
  - **Leverage**: 现有订单模型
  - **Prompt**: Role: Go Developer | Task: 在订单模型中增加订单来源和国籍字段 | Context: 遵循 Go Main Model 规范 | Restrictions: 遵循 `.cursor/rules/go-main.mdc` | Success: 字段可正常读写

- [ ] 3.2 新增外卖来源和国籍 Model（Go Main）

  - **Module**: Main - Model 层
  - **File**: `main/app/model/order_source.go`, `main/app/model/nationality.go`
  - **Purpose**: 定义外卖来源和国籍数据模型（用于查询配置）
  - **Requirements**: R3, R4
  - **Leverage**: 现有配置类 Model

- [ ] 3.3 扩展订单 DTO（Go Main）

  - **Module**: Main - DTO 层
  - **File**: `main/app/dto/req/order_req.go`, `main/app/dto/resp/order_resp.go`
  - **Purpose**: 扩展订单创建请求参数和详情响应数据，增加订单来源和国籍字段
  - **Requirements**: R3, R4, R5
  - **Leverage**: 现有 DTO 示例
  - **Prompt**: Role: Go Developer | Task: 扩展订单相关 DTO，增加订单来源和国籍字段 | Context: 遵循 Go Main DTO 规范 | Restrictions: 遵循 `.cursor/rules/go-main.mdc` | Success: DTO 结构清晰合理

- [ ] 3.4 扩展订单 Service（Go Main）

  - **Module**: Main - Service 层
  - **File**: `main/app/service/order_service.go`
  - **Purpose**: 扩展订单创建方法，处理订单来源和国籍字段
  - **Requirements**: R3, R4, R5
  - **Leverage**: 现有订单创建逻辑
  - **Key Points**: 
    - 参数校验：order_source_uuid 和 nationality_uuid 可选
    - 默认值处理：0 表示未记录
  - **Prompt**: Role: Go Developer | Task: 扩展 OrderService 订单创建方法，增加订单来源和国籍处理 | Context: Service 可依赖其他 Service | Restrictions: 遵循 `.cursor/rules/go-main.mdc`，不使用 panic，返回 error | Success: 订单创建成功记录订单来源和国籍

- [ ] 3.5 扩展订单 API（Go Main）

  - **Module**: Main - API 层
  - **File**: `main/app/api/v1/cashier/order_api.go`, `main/app/api/v1/assistant/order_api.go`
  - **Purpose**: 扩展订单创建和详情接口，支持订单来源和国籍字段
  - **Requirements**: R3, R4, R5
  - **Leverage**: 现有订单 API
  - **Prompt**: Role: Go Developer | Task: 扩展订单创建和详情接口，增加订单来源和国籍参数/响应 | Context: API → Service → Repository 严格分层 | Restrictions: 遵循 `.cursor/rules/go-main.mdc` 和 `.cursor/rules/api.mdc`，URL 使用 snake_case | Success: 接口可正常调用

---

## Phase 4: 前端实现（Flutter）

> 根据 `structs.mdc`，本阶段涉及：
> - **新管理端**（Flutter）：`ttpos-flutter/apps/shop/`
> - **收银端**（Flutter）：`ttpos-flutter/apps/pos/`
> - **点餐助手**（Flutter）：`ttpos-flutter/apps/assistant/`

- [ ] 4.1 新管理端-业务设置页（Flutter）

  - **Module**: 新管理端 - 桌面端（Flutter）
  - **File**: `ttpos-flutter/apps/shop/lib/pages/business_settings/`
  - **Purpose**: 实现外卖来源和国籍配置界面（开关 + 列表 + 增删改查）
  - **Requirements**: R1, R2
  - **Leverage**: 现有配置管理页面
  - **Note**: ⚠️ 这是 Flutter 新管理端，不是 Vue 旧店铺后台

- [ ] 4.2 收银端-点餐方式外卖选项（Flutter）

  - **Module**: 收银端前端（Flutter）
  - **File**: `ttpos-flutter/apps/pos/lib/pages/order/`
  - **Purpose**: 点餐方式增加「外卖」按钮，点击弹出外卖来源选择
  - **Requirements**: R3
  - **Leverage**: 现有点餐方式选择逻辑
  - **Key Points**: 
    - 后台未开启时不显示
    - 外卖来源列表从后台配置读取
    - 支持不选择（默认店内）

- [ ] 4.3 收银端/点餐助手-国籍选择（Flutter）

  - **Module**: 收银端/点餐助手前端（Flutter）
  - **File**: `ttpos-flutter/apps/pos/lib/pages/order/`, `ttpos-flutter/apps/assistant/lib/pages/order/`
  - **Purpose**: 点餐和桌台管理中增加国籍选择控件
  - **Requirements**: R4
  - **Leverage**: 现有选择控件
  - **Key Points**: 
    - 后台未开启时不显示
    - 国籍列表从后台配置读取，多语言显示
    - 可选字段，支持不选择

- [ ] 4.4 订单详情-显示订单来源和国籍（Flutter）

  - **Module**: 收银端/新管理端前端（Flutter）
  - **File**: `ttpos-flutter/apps/pos/lib/pages/order/`, `ttpos-flutter/apps/shop/lib/pages/order/`
  - **Purpose**: 订单详情页显示订单来源和国籍信息
  - **Requirements**: R5
  - **Leverage**: 现有订单详情页
  - **Key Points**: 
    - JOIN 多语言表获取名称
    - 历史订单显示「未记录」
    - 已删除配置显示原名称

---

## Phase 5: 测试与优化

- [ ] 5.1 后端单元测试

  - **File**: 对应 Service/Model 测试文件
  - **Purpose**: 覆盖核心业务逻辑（配置管理、订单创建扩展）
  - **Requirements**: 所有
  - **Key Points**: 
    - Service 层测试覆盖率 ≥ 70%
    - 测试多语言关联查询
    - 测试软删除校验

- [ ] 5.2 API 集成测试

  - **File**: 对应 API 测试文件
  - **Purpose**: 覆盖配置管理接口 + 订单创建/详情接口
  - **Requirements**: 所有
  - **Key Points**: 
    - 测试配置 CRUD 流程
    - 测试订单创建带订单来源和国籍
    - 测试订单详情返回正确名称

- [ ] 5.3 数据兼容性测试

  - **Purpose**: 测试历史订单在各终端的显示效果
  - **Requirements**: R5
  - **Key Points**: 
    - 历史订单（无订单来源和国籍）显示「未记录」
    - 已删除配置仍显示原名称
    - 不影响订单其他功能

- [ ] 5.4 多语言测试

  - **Purpose**: 测试国籍名称在不同语言环境下的显示
  - **Requirements**: R2.6, R2.7
  - **Key Points**: 
    - 中文、英文、泰语环境下显示正确
    - 多语言表 JOIN 查询正确
    - 新增配置支持多语言输入

---

## 提交清单

- [ ] requirements.md 中的需求条目全部覆盖到具体任务；
- [ ] design.md 中的重要设计点均有对应实现任务；
- [ ] 所有新增/修改 API 已在文档中更新；
- [ ] 数据库迁移脚本已通过测试环境验证，可安全重复执行；
- [ ] 前后端联调通过，多终端体验一致；
- [ ] 历史数据兼容性测试通过。

---

## 📋 数据库规范检查清单

> 基于 `.cursor/rules/database.mdc` 规范

**设计表时：**

- [ ] 包含必需字段：`id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [ ] 时间字段使用 int 类型，默认值 0
- [ ] UUID 字段使用 bigint unsigned
- [ ] 表名使用 `ttpos_` 前缀
- [ ] 字段名使用 snake_case
- [ ] **多语言字段使用 `multi_language_name_uuid` 关联 `ttpos_multi_language_name` 表**
- [ ] **不包含 `company_uuid` 字段**（多租户架构，每个商户独立数据库）
- [ ] **不包含 `name_lang` 等 JSON 多语言字段**（使用多语言表）

**编写迁移时：**

- [ ] 文件命名使用 `date +%Y%m%d%H%M%S` 生成时间戳
- [ ] 检查表是否存在
- [ ] 检查字段是否存在
- [ ] 包含注释说明
- [ ] 可重复执行

**编写代码时：**

- [ ] 先创建/更新 `ttpos_multi_language_name` 记录
- [ ] 将返回的 uuid 赋值给业务表的 `multi_language_name_uuid`
- [ ] 查询时使用 JOIN 关联 `ttpos_multi_language_name` 表
- [ ] 删除配置时使用软删除，避免历史订单数据丢失

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出多语言关联查询优化、配置软删除策略等经验，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-19  
**维护者**: 后端/前端联合小组

