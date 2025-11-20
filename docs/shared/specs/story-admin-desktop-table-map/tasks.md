# 新管理端-桌面端-桌台地图 任务分解

> 本文档定义「新管理端-桌面端-桌台地图」功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求（如 R1.1, R2.3）

## 📊 进度总览

**总任务数**: 30（Phase 1: 5 + Phase 2: 10 + Phase 3: 11 + Phase 4: 4）  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

> 实际开发过程中，请在完成任务时同步维护以上统计，并更新对应任务的勾选状态。

---

## Phase 1: 云平台商家开关（Requirement 3）

> 根据 `structs.mdc`，本阶段涉及 **Admin 模块 - 管理后台**（PHP + ThinkPHP + Vue 3）

- [ ] 1.1 商家表新增能力开关字段

  - **Module**: Admin - 数据库迁移
  - **File**: `admin/database/migrations/{YYYYMMDDHHMMSS}_alter_ttpos_merchant_add_table_map_fields.php`
  - **Purpose**: 为 `ttpos_merchant` 表增加 `enable_table_map`、`enable_data_management` 字段
  - **Requirements**: R3.1, R3.3
  - **Leverage**: 现有商家相关迁移文件；模板: `docs/agent/templates/database-migration-template.md`
  - **Prompt**: Role: Database Engineer | Task: 为 ttpos_merchant 表新增桌台地图/数据管理能力字段，遵循数据库规范 | Context: 必须包含默认值，兼容历史数据，迁移可重复执行 | Restrictions: 遵循 `.cursor/rules/database.mdc`，谨慎处理生产环境影响 | Success: 字段新增成功，现有数据不受影响

- [ ] 1.2 商家 Model 同步字段

  - **Module**: Admin - 管理后台 Model
  - **File**: `admin/app/admin/model/Merchant.php`
  - **Purpose**: 在 Model 中声明新增字段，支持读写
  - **Requirements**: R3.1
  - **Leverage**: 现有字段定义模式
  - **Prompt**: Role: PHP Developer | Task: 在 Merchant Model 中添加 enable_table_map 和 enable_data_management 字段定义 | Context: 遵循 ThinkPHP Model 规范 | Restrictions: 遵循 `.cursor/rules/php.mdc` | Success: 字段可正常读写

- [ ] 1.3 商家 Service 增加字段处理逻辑

  - **Module**: Admin - 管理后台 Service
  - **File**: `admin/app/admin/service/MerchantService.php`
  - **Purpose**: 在商家创建/编辑方法中处理新增字段的读写和校验
  - **Requirements**: R3.1-R3.4
  - **Leverage**: 现有商家 Service 方法
  - **Prompt**: Role: PHP Developer | Task: 在 MerchantService 的 create/update 方法中增加桌台地图开关字段处理 | Context: 业务逻辑放在 Service 层 | Restrictions: 遵循 `.cursor/rules/php.mdc`，Controller 不包含业务逻辑 | Success: 字段可通过 Service 正确保存和读取

- [ ] 1.4 商家 Validate 增加字段验证

  - **Module**: Admin - 管理后台 Validate
  - **File**: `admin/app/admin/validate/MerchantValidate.php`
  - **Purpose**: 为新增字段添加验证规则（布尔类型，可选）
  - **Requirements**: R3.1
  - **Leverage**: 现有验证器规则
  - **Prompt**: Role: PHP Developer | Task: 在 MerchantValidate 中为 enable_table_map 和 enable_data_management 添加验证规则 | Context: 使用 ThinkPHP 验证器 | Restrictions: 遵循 `.cursor/rules/php.mdc` | Success: 参数验证通过

- [ ] 1.5 云平台商家编辑页增加开关控件

  - **Module**: Admin - 管理后台前端（Vue 3）
  - **File**: `admin/views/admin/pages/merchant/edit.vue`
  - **Purpose**: 在前端表单中增加「桌台地图」「数据管理」两个开关
  - **Requirements**: R3.1
  - **Leverage**: 现有 Boolean 开关控件（如 el-switch）和表单提交流程
  - **Prompt**: Role: Frontend Developer | Task: 在商家编辑表单中增加两个开关控件 | Context: 使用 Element Plus 组件库 | Restrictions: 遵循 `.cursor/rules/vue.mdc` | Success: 开关可正常切换并提交保存

---

## Phase 2: 新管理端-桌台地图配置（Requirement 1）

> 根据 `structs.mdc`，本阶段涉及：
> - **Admin 模块 - 店铺后台**（PHP + ThinkPHP）：`admin/app/shop/`
> - **新管理端前端**（Flutter）：`ttpos-flutter/apps/shop/`（注意：不是 `admin/views/shop`）

- [ ] 2.1 创建桌台布局表迁移

  - **Module**: Admin - 数据库迁移
  - **File**: `admin/database/migrations/{YYYYMMDDHHMMSS}_create_ttpos_table_map_layout_table.php`
  - **Purpose**: 定义 `ttpos_table_map_layout` 表结构
  - **Requirements**: R1.1-R1.6
  - **Leverage**: design.md 中数据库设计章节；模板: `docs/agent/templates/database-migration-template.md`
  - **Prompt**: Role: Database Engineer | Task: 创建桌台布局表，包含 company_uuid, area_uuid, layout_json 等字段 | Context: 遵循数据库规范，包含标准字段 | Restrictions: 遵循 `.cursor/rules/database.mdc` | Success: 表创建成功，索引合理

- [ ] 2.2 新增布局 Model

  - **Module**: Admin - 店铺后台 Model
  - **File**: `admin/app/shop/model/TableMapLayout.php`
  - **Purpose**: 映射 `ttpos_table_map_layout` 表
  - **Requirements**: R1.1-R1.6
  - **Leverage**: 现有类似配置类 Model（如区域、桌台 Model）
  - **Prompt**: Role: PHP Developer | Task: 创建 TableMapLayout Model 映射布局表 | Context: 使用 ThinkPHP Model 规范 | Restrictions: 遵循 `.cursor/rules/php.mdc` | Success: Model 可正常读写布局数据

- [ ] 2.3 新增布局 Validate

  - **Module**: Admin - 店铺后台 Validate
  - **File**: `admin/app/shop/validate/TableMapValidate.php`
  - **Purpose**: 验证布局保存参数（area_uuid, layout JSON 结构等）
  - **Requirements**: R1.6
  - **Leverage**: 现有验证器规则
  - **Prompt**: Role: PHP Developer | Task: 创建 TableMapValidate 验证器，验证布局数据合法性 | Context: 使用 ThinkPHP 验证器 | Restrictions: 遵循 `.cursor/rules/php.mdc` | Success: 参数验证通过

- [ ] 2.4 新增 TableMap Service

  - **Module**: Admin - 店铺后台 Service
  - **File**: `admin/app/shop/service/TableMapService.php`
  - **Purpose**: 封装区域列表+状态查询、布局读取/保存逻辑
  - **Requirements**: R1.1-R1.6, R3.2-R3.3
  - **Leverage**: 现有餐厅/桌台管理 Service（如 `AreaService.php`, `TableService.php`）
  - **Prompt**: Role: PHP Developer | Task: 创建 TableMapService，实现 getAreaListWithStatus, getAreaLayoutDetail, saveAreaLayout 方法 | Context: 业务逻辑放在 Service 层，可依赖其他 Service | Restrictions: 遵循 `.cursor/rules/php.mdc`，Controller 不包含业务逻辑 | Success: Service 方法可正常调用

- [ ] 2.5 新增 TableMap Controller

  - **Module**: Admin - 店铺后台 Controller
  - **File**: `admin/app/shop/controller/TableMapController.php`
  - **Purpose**: 提供区域列表、布局详情、保存布局三个 API（index, edit, saveLayout）
  - **Requirements**: R1.1-R1.6
  - **Leverage**: 现有 Controller 示例
  - **Prompt**: Role: PHP Developer | Task: 创建 TableMapController，实现 index/edit/saveLayout 方法 | Context: Controller 只做参数获取/校验与结果返回 | Restrictions: 遵循 `.cursor/rules/php.mdc` 和 `.cursor/rules/api.mdc` | Success: API 接口可正常调用

- [ ] 2.6 新增路由

  - **Module**: Admin - 路由配置
  - **File**: `admin/config/route.php` 或对应路由配置
  - **Purpose**: 注册桌台地图相关接口路由
  - **Requirements**: R1.1-R1.6
  - **Prompt**: Role: PHP Developer | Task: 在路由配置中注册 TableMapController 的路由 | Context: 遵循现有路由命名规范 | Restrictions: URL 使用 snake_case | Success: 路由可正常访问

- [ ] 2.7 新管理端前端页面搭建（Flutter）

  - **Module**: 新管理端 - 桌面端（Flutter）
  - **File**: `ttpos-flutter/apps/shop/lib/pages/table_map/` （具体路径需在 ttpos-flutter 仓库确认）
  - **Purpose**: 实现桌台地图主页面（区域列表 + 编辑入口）
  - **Requirements**: R1.1, R1.2
  - **Leverage**: 现有页面列表/表格组件
  - **Note**: ⚠️ 注意这是 Flutter 前端，不是 `admin/views/shop`（Vue 3 旧店铺后台）

- [ ] 2.8 编辑页左侧桌台列表组件（Flutter）

  - **Module**: 新管理端 - 桌面端（Flutter）
  - **File**: `ttpos-flutter/apps/shop/lib/pages/table_map/components/table_list_panel.dart`
  - **Purpose**: 展示区域桌台列表及勾选状态
  - **Requirements**: R1.3, R1.5

- [ ] 2.9 画布编辑组件（Flutter）

  - **Module**: 新管理端 - 桌面端（Flutter）
  - **File**: `ttpos-flutter/apps/shop/lib/pages/table_map/components/canvas_editor.dart`
  - **Purpose**: 支持桌台拖拽、缩放、样式切换、尺寸调整
  - **Requirements**: R1.4, R1.5, R1.6
  - **Leverage**: 若已有拖拽/画布类组件可复用

- [ ] 2.10 保存交互与校验（Flutter）

  - **Module**: 新管理端 - 桌面端（Flutter）
  - **File**: 同 2.9
  - **Purpose**: 在保存前校验布局非空及数据有效性
  - **Requirements**: R1.6

---

## Phase 3: 终端桌台地图模式（Requirement 2）

> 根据 `structs.mdc`，本阶段涉及：
> - **Main 模块**（Go + Gin）：`main/app/api/v1/{cashier,assistant}/`
> - **终端前端**（Flutter）：`ttpos-flutter/apps/{pos,assistant}/`

### 3.A 收银端（pos）桌台地图

- [ ] 3.1 收银端地图布局读取接口（Go Main）

  - **Module**: Main - 收银端 API
  - **File**: `main/app/api/v1/cashier/table_map_api.go`
  - **Purpose**: 提供收银端获取桌台地图布局的接口（`/api/v1/cashier/table_map/layout`）
  - **Requirements**: R2.1-R2.5
  - **Leverage**: 现有 `main/app/api/v1/cashier/table_api.go` 桌台列表/状态接口
  - **Prompt**: Role: Go Developer | Task: 创建收银端桌台地图布局读取接口 | Context: 遵循 Go Main 三层架构，API → Service → Repository | Restrictions: 遵循 `.cursor/rules/go-main.mdc` 和 `.cursor/rules/api.mdc`，URL 使用 snake_case，data 字段必须是对象 | Success: 接口可正常返回布局数据

- [ ] 3.2 收银端 TableMap Service（Go Main）

  - **Module**: Main - Service 层
  - **File**: `main/app/service/table_map_service.go`
  - **Purpose**: 桌台地图业务逻辑（可复用现有 `table_service.go`）
  - **Requirements**: R2.1-R2.5
  - **Leverage**: 现有 `main/app/service/table_service.go`
  - **Prompt**: Role: Go Developer | Task: 创建或扩展 TableMapService，实现布局数据查询逻辑 | Context: Service 可依赖其他 Service，不直接依赖 Repository | Restrictions: 遵循 `.cursor/rules/go-main.mdc`，不使用 panic，返回 error | Success: Service 方法可正常调用

- [ ] 3.3 收银端 TableMap Repository（Go Main）

  - **Module**: Main - Repository 层
  - **File**: `main/app/repository/table_map_repository.go`
  - **Purpose**: 桌台布局数据访问
  - **Requirements**: R2.1-R2.5
  - **Leverage**: 现有 Repository 示例
  - **Prompt**: Role: Go Developer | Task: 创建 TableMapRepository，实现布局数据查询 | Context: Repository 只持有 db 实例，不持有 DBManager | Restrictions: 遵循 `.cursor/rules/go-main.mdc` | Success: Repository 可正常查询数据

- [ ] 3.4 收银端 DTO 定义（Go Main）

  - **Module**: Main - DTO 层
  - **File**: `main/app/dto/req/table_map_req.go`, `main/app/dto/resp/table_map_resp.go`
  - **Purpose**: 定义请求参数和响应数据结构
  - **Requirements**: R2.1-R2.5
  - **Leverage**: 现有 DTO 示例
  - **Prompt**: Role: Go Developer | Task: 定义桌台地图相关 DTO | Context: 遵循 Go Main DTO 规范 | Restrictions: 遵循 `.cursor/rules/go-main.mdc` | Success: DTO 结构清晰合理

- [ ] 3.5 收银端前端地图模式切换按钮（Flutter）

  - **Module**: 收银端前端（Flutter）
  - **File**: `ttpos-flutter/apps/pos/lib/pages/table/` （具体路径需在 ttpos-flutter 仓库确认）
  - **Purpose**: 在桌台页增加列表/地图模式切换按钮
  - **Requirements**: R2.1-R2.3
  - **Leverage**: 现有桌台列表页面

- [ ] 3.6 收银端地图视图组件（Flutter）

  - **Module**: 收银端前端（Flutter）
  - **File**: `ttpos-flutter/apps/pos/lib/pages/table/components/table_map_view.dart`
  - **Purpose**: 渲染布局及状态，支持筛选后隐藏无关桌台
  - **Requirements**: R2.2-R2.5

- [ ] 3.7 收银端模式切换与状态保持（Flutter）

  - **Module**: 收银端前端（Flutter）
  - **File**: 同 3.6
  - **Purpose**: 确保列表/地图模式间筛选条件和选中桌台一致
  - **Requirements**: R2.2, R2.5

### 3.B 点餐助手端（assistant）桌台地图

- [ ] 3.8 点餐助手端地图布局读取接口（Go Main）

  - **Module**: Main - 点餐助手端 API
  - **File**: `main/app/api/v1/assistant/table_map_api.go`
  - **Purpose**: 提供点餐助手端获取桌台地图布局的接口（`/api/v1/assistant/table_map/layout`）
  - **Requirements**: R2.1-R2.5
  - **Leverage**: 复用 3.2-3.4 的 Service/Repository/DTO，只需新增 API 层
  - **Prompt**: Role: Go Developer | Task: 创建点餐助手端桌台地图布局读取接口 | Context: 复用 TableMapService | Restrictions: 遵循 `.cursor/rules/go-main.mdc` 和 `.cursor/rules/api.mdc` | Success: 接口可正常返回布局数据

- [ ] 3.9 点餐助手端前端地图模式切换按钮（Flutter）

  - **Module**: 点餐助手端前端（Flutter）
  - **File**: `ttpos-flutter/apps/assistant/lib/pages/table/`
  - **Purpose**: 在桌台页增加列表/地图模式切换按钮
  - **Requirements**: R2.1-R2.3
  - **Leverage**: 复用收银端地图组件（3.6-3.7）

- [ ] 3.10 点餐助手端地图视图组件（Flutter）

  - **Module**: 点餐助手端前端（Flutter）
  - **File**: `ttpos-flutter/apps/assistant/lib/pages/table/components/table_map_view.dart`
  - **Purpose**: 渲染布局及状态，支持筛选后隐藏无关桌台
  - **Requirements**: R2.2-R2.5
  - **Leverage**: 复用收银端地图组件

- [ ] 3.11 点餐助手端模式切换与状态保持（Flutter）

  - **Module**: 点餐助手端前端（Flutter）
  - **File**: 同 3.10
  - **Purpose**: 确保列表/地图模式间筛选条件和选中桌台一致
  - **Requirements**: R2.2, R2.5

---

## Phase 4: 测试与优化

- [ ] 4.1 后端单元测试

  - File: 对应 Service/Model 测试文件
  - Purpose: 覆盖核心业务逻辑
  - Requirements: 所有

- [ ] 4.2 API 集成测试

  - File: 对应 API 测试文件
  - Purpose: 覆盖 Admin 配置接口 + 终端布局接口

- [ ] 4.3 前端交互测试

  - File: E2E 或手动测试用例
  - Purpose: 覆盖画布交互、模式切换、筛选联动

- [ ] 4.4 性能与大场景验证

  - File: 测试脚本/记录
  - Purpose: 验证大桌量场景下地图模式的可用性与性能

---

## 提交清单

- [ ] requirements.md 中的需求条目全部覆盖到具体任务；
- [ ] design.md 中的重要设计点均有对应实现任务；
- [ ] 所有新增/修改 API 已在文档中更新；
- [ ] 数据库迁移脚本已通过测试环境验证，可安全重复执行；
- [ ] 前后端联调通过，终端体验满足验收标准。

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出桌台地图建模/大场景性能优化等经验，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-19  
**维护者**: 后端/前端联合小组


