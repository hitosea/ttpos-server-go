# 收银机接单权限子项管理 需求文档

> 本文档定义收银机接单权限子项管理功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | 任务 37492: 权限-Grab外卖对接                                                                                |
| **关联前端 Spec** | [pos-grab-order-process](../../../../ttpos-flutter/docs/shared/specs/active/pos-grab-order-process/requirements.md) |
| **创建日期**      | 2025-12-05                                                                                                 |
| **负责人**        | 待分配                                                                                                       |
| **目标版本**      | v2.11.0                                                                                                     |
| **涉及技术栈**    | [x] Go (main/) [x] PHP (admin/)                                                                             |

## 📋 审核状态

| 项目         | 内容     |
| ------------ | -------- |
| **审核状态** | 已通过   |
| **审核人**   | 开发团队 |
| **审核日期** | 2025-12-09 |
| **审核意见** | 需求已通过审核，可进入技术设计阶段 |

---

## 📋 概述

在收银机权限下新增"接单"权限（与沽清权限平级），并将原有的接单权限和外送权限调整为该新权限的子项，同时新增外卖接单权限作为第三项子项，实现细粒度的接单权限控制。根据云平台外卖开关状态动态显示/隐藏"外卖"权限子项，并在云平台开启任意一个外卖平台时，自动为默认角色（Cashier、Store Manager）勾选"外卖"权限。

**核心功能**：
1. 在收银机权限（1704880670）下新增"接单"权限，与沽清权限（1724220513）平级，排序在沽清后（sort=11）
2. 将原有的接单权限（1724320522）调整为新接单权限的子项，名称改为"扫码接单"
3. 将原有的外送权限（1752716650）调整为新接单权限的子项
4. 新增外卖接单权限，作为新接单权限的第三项子项
5. 根据云平台外卖开关状态（BaseInfo API 返回）动态显示/隐藏"外卖"权限子项
6. 当云平台开启任意一个外卖平台时，系统自动为默认角色（Cashier、Store Manager）勾选"外卖"权限
7. POS 端接单菜单根据用户权限和云平台开关状态，动态显示/隐藏外卖订单

**实现范围**：
- 数据库：在 `access` 表中新增"接单"权限，调整原有接单权限和外送权限的 parent_uuid，新增"外卖"权限子项
- 权限筛选：在权限筛选逻辑中根据云平台外卖开关状态动态过滤"外卖"权限
- 默认角色：在云平台开启外卖时，自动为默认角色分配"外卖"权限和新"接单"权限
- API：BaseInfo API 返回外卖开关状态，供前端判断

## 🎯 产品对齐

- 实现收银机接单权限的细粒度控制
- 根据云平台外卖开关状态动态控制权限可见性
- 提升权限管理的灵活性和用户体验
- 支持多外卖平台的权限统一管理

## 📝 用户故事

**作为** 系统管理员  
**我想** 在角色权限配置中为收银机接单权限设置子项（扫码接单、外送、外卖）  
**以便于** 实现细粒度的接单权限控制

**作为** 系统  
**我想** 根据云平台外卖开关状态动态显示/隐藏"外卖"权限子项  
**以便于** 确保权限配置与实际业务功能保持一致

**作为** 系统  
**我想** 在云平台开启外卖时自动为默认角色分配"外卖"权限  
**以便于** 减少管理员的手动配置工作

---

## 功能需求

### Requirement 1: 权限数据结构 - 接单权限重构

**用户故事**: 作为系统，我想在收银机权限下新增"接单"权限，并将原有接单权限、外送权限调整为该权限的子项，同时新增外卖接单权限，以便于实现细粒度的权限控制

#### 验收标准

1. **WHEN** 查看收银机权限结构 **THEN** 系统 **SHALL** 显示"接单"权限（与沽清权限平级）及其子项（扫码接单、外送、外卖）
2. **WHEN** 权限数据迁移完成时 **THEN** 系统 **SHALL** 正确创建新的"接单"权限，并调整原有权限的层级关系
3. **WHEN** 默认角色查询权限时 **THEN** 系统 **SHALL** 显示新"接单"权限和"外卖"权限已分配给 Cashier 和 Store Manager 角色

#### 具体要求

- [ ] 1.1 在 `access` 表中新增"接单"权限记录（作为收银机权限的子级）
  - UUID: 待生成（建议使用时间戳，例如：1734000000）
  - name: "接单"
  - path: `cashier_accept_order`
  - parent_uuid: 1704880670（收银机权限UUID）
  - sort: 11（沽清权限sort为10，接单权限sort为11，排序在沽清后）
  - api_path: ''（父级权限，无具体API路径）
  - is_route: 0
  - is_menu: 0
  - is_show: 1
  - is_supplier: 0

- [ ] 1.2 更新原有"接单"权限（UUID: 1724320522），调整为新接单权限的子项
  - name: "扫码接单"（名称修改）
  - path: `cashier_accept_scan_order`（路径修改）
  - parent_uuid: 新接单权限UUID（从1704880670改为新接单权限UUID）
  - sort: 10（作为第一项子级）
  - api_path: `/store/TakeOrder/list`（保持不变）
  - 其他字段保持不变

- [ ] 1.3 更新"外送"权限（UUID: 1752716650），调整为新接单权限的子项
  - name: "外送"（保持不变）
  - path: `cashier_accept_delivery`（保持不变）
  - parent_uuid: 新接单权限UUID（从1704880670改为新接单权限UUID）
  - sort: 20（作为第二项子级）
  - api_path: `/cashier/member_order/list`（保持不变）
  - 其他字段保持不变

- [ ] 1.4 在 `access` 表中新增"外卖"权限记录（作为新接单权限的子项）
  - UUID: 待生成（建议使用时间戳，例如：1734000001）
  - name: "外卖"
  - path: `cashier_accept_delivery_platform`
  - parent_uuid: 新接单权限UUID
  - sort: 30（作为第三项子级）
  - api_path: `/cashier/grab_order/list`（外卖订单列表，待确认）
  - is_route: 0
  - is_menu: 0
  - is_show: 1
  - is_supplier: 0

- [ ] 1.5 为新"接单"权限分配默认角色权限
  - 为 Cashier 角色（role_uuid: 2）分配新"接单"权限
  - 为 Store Manager 角色（role_uuid: 1）分配新"接单"权限
  - 在 `role_access` 表中插入记录

- [ ] 1.6 为"外卖"权限分配默认角色权限
  - 为 Cashier 角色（role_uuid: 2）分配"外卖"权限
  - 为 Store Manager 角色（role_uuid: 1）分配"外卖"权限
  - 在 `role_access` 表中插入记录

- [ ] 1.7 创建数据库迁移脚本，执行权限结构调整
  - 文件路径: `admin/database/migrations/{timestamp}_add_cashier_order_permission_sub_items.php`
  - 参考: `admin/database/migrations/20250717013925_add_cashier_permission_delivery_order_to_access.php`
  - 执行顺序：
    1. 新增"接单"权限
    2. 更新原有"接单"权限（1724320522）的 parent_uuid 和名称
    3. 更新"外送"权限（1752716650）的 parent_uuid
    4. 新增"外卖"权限
    5. 为新"接单"权限和"外卖"权限分配默认角色权限

**参考实现**：
- 外送权限迁移: `admin/database/migrations/20250717013925_add_cashier_permission_delivery_order_to_access.php`
- 权限初始化: `admin/database/seeds/shop_init_data.sql`
- 角色信息: Store Manager (role_uuid=1), Cashier (role_uuid=2)

---

### Requirement 2: 权限筛选 - 根据云平台外卖开关动态显示/隐藏

**用户故事**: 作为系统，我想根据云平台外卖开关状态动态显示/隐藏"外卖"权限子项，以便于确保权限配置与实际业务功能保持一致

#### 验收标准

1. **WHEN** 云平台未开启任何外卖平台（`enable_grab_delivery = 0` 且其他外卖平台也未开启） **THEN** 系统 **SHALL** 在权限列表中隐藏"外卖"权限子项
2. **WHEN** 云平台开启任意一个外卖平台（`enable_grab_delivery = 1` 或其他外卖平台开启） **THEN** 系统 **SHALL** 在权限列表中显示"外卖"权限子项
3. **WHEN** 用户查询权限列表时 **THEN** 系统 **SHALL** 根据当前商家的云平台外卖开关状态动态过滤权限

#### 具体要求

- [ ] 2.1 在 Go Main 模块的权限筛选逻辑中添加外卖权限过滤
  - File: `main/app/service/role_access.go`
  - Function: `filterPermission`
  - Logic: 检查云平台外卖开关状态（`companySetting.IsOpenGrabDelivery()` 或其他外卖平台开关）
  - 如果未开启，则过滤掉"外卖"权限（UUID: 新生成的外卖权限UUID）
  - 参考: 外送权限的过滤逻辑（第190行）

- [ ] 2.2 在 PHP Admin 模块的权限筛选逻辑中添加外卖权限过滤
  - File: `admin/app/common/model/shop/Access.php`
  - Function: `filterPermission` 或相关筛选方法
  - Logic: 检查云平台外卖开关状态
  - 如果未开启，则过滤掉"外卖"权限
  - 参考: 外送权限的过滤逻辑

- [ ] 2.3 确认云平台外卖开关状态的获取方式
  - 检查 `company_setting` 表中的 `enable_grab_delivery` 字段
  - 检查是否有其他外卖平台的开关字段
  - 确认"任意一个外卖平台开启"的判断逻辑

**参考实现**：
- 外送权限过滤: `main/app/service/role_access.go` 第189-192行
- 扫码点餐接单权限过滤: `main/app/service/role_access.go` 第185-188行

---

### Requirement 3: 默认角色权限分配 - 自动勾选外卖权限

**用户故事**: 作为系统，我想在云平台开启外卖时自动为默认角色（Cashier、Store Manager）分配"外卖"权限，以便于减少管理员的手动配置工作

#### 验收标准

1. **WHEN** 云平台开启任意一个外卖平台时 **THEN** 系统 **SHALL** 自动为默认角色（Cashier、Store Manager）分配"外卖"权限
2. **WHEN** 云平台关闭所有外卖平台时 **THEN** 系统 **SHALL** 不自动移除已分配的"外卖"权限（保留现有权限配置）
3. **WHEN** 新建商家且云平台已开启外卖时 **THEN** 系统 **SHALL** 在初始化默认角色时自动分配"外卖"权限

#### 具体要求

- [ ] 3.1 在商家新建/编辑时检查云平台外卖开关状态
  - File: `admin/app/admin/controller/Shop.php`
  - Logic: 在保存商家配置后，检查 `enable_grab_delivery` 或其他外卖平台开关
  - 如果开启，则检查默认角色是否已有"外卖"权限，如果没有则自动分配

- [ ] 3.2 创建默认角色权限分配逻辑
  - File: `admin/app/admin/service/RoleService.php`（如不存在则创建）
  - Function: `assignDefaultDeliveryPlatformPermission`
  - Logic: 为指定角色分配"外卖"权限（如果角色不存在该权限）
  - 参考: `admin/database/migrations/20250717013925_add_cashier_permission_delivery_order_to_access.php` 第38-45行

- [ ] 3.3 在数据库迁移脚本中为默认角色分配权限
  - 在权限迁移脚本中添加默认角色权限分配逻辑
  - 为新"接单"权限分配默认角色（Cashier role_uuid=2, Store Manager role_uuid=1）
  - 为"外卖"权限分配默认角色（Cashier role_uuid=2, Store Manager role_uuid=1）
  - 参考: `admin/database/migrations/20250717013925_add_cashier_permission_delivery_order_to_access.php`

- [ ] 3.4 处理云平台外卖开关状态变更的场景
  - 当商家从"未开启外卖"变为"开启外卖"时，自动为默认角色分配权限
  - 当商家从"开启外卖"变为"未开启外卖"时，不自动移除权限（保留用户配置）

**参考实现**：
- 默认角色权限分配: `admin/database/migrations/20250717013925_add_cashier_permission_delivery_order_to_access.php` 第38-45行
- 角色权限关联表: `role_access`

---

### Requirement 4: API 支持 - BaseInfo 返回外卖开关状态

**用户故事**: 作为前端，我想从 BaseInfo API 获取云平台外卖开关状态，以便于根据状态动态显示/隐藏权限和功能

#### 验收标准

1. **WHEN** 前端调用 BaseInfo API **THEN** 系统 **SHALL** 返回云平台外卖开关状态字段
2. **WHEN** 云平台开启外卖时 **THEN** 系统 **SHALL** 返回 `is_open_grab_delivery: true`
3. **WHEN** 云平台未开启外卖时 **THEN** 系统 **SHALL** 返回 `is_open_grab_delivery: false`

#### 具体要求

- [ ] 4.1 确认 BaseInfo API 已返回 `is_open_grab_delivery` 字段
  - File: `main/app/dto/resp/base.go`
  - Struct: `Company`
  - Field: `IsOpenGrabDelivery bool`（已存在，第124行）
  - 确认该字段已正确映射到 `company_setting.enable_grab_delivery`

- [ ] 4.2 确认 `/shop/base` 接口返回外卖开关状态
  - File: `main/app/service/auth.go`
  - Function: `GetShopBase`
  - 确认 `Company.IsOpenGrabDelivery` 已正确设置（第1020行）
  - 确认该字段来自 `companySetting.IsOpenGrabDelivery()`

- [ ] 4.3 确认其他 BaseInfo 相关接口返回外卖开关状态
  - `/menu/base`: `main/app/api/v1/menu/menu_handler.go`
  - `/tablet/base`: `main/app/api/v1/tablet/tablet_base.go`
  - `/h5/base`: `main/app/service/h5_service.go`
  - 确认这些接口的 `Company` 结构中都包含 `IsOpenGrabDelivery` 字段

**参考实现**：
- BaseInfo 响应结构: `main/app/dto/resp/base.go` 第109-125行
- `/shop/base` 接口: `main/app/service/auth.go` 第1005-1021行

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/php.mdc` - PHP 开发规范
  - `.cursor/rules/database.mdc` - 数据库开发规范

### API 设计要求

- [ ] URL 使用 snake_case 命名
- [ ] data 字段必须是对象，不能是 null 或数组
- [ ] 响应格式：`{code, message, data{}}`
- [ ] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [ ] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [ ] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [ ] UUID 字段使用 bigint unsigned
- [ ] 表名使用 ttpos\_ 前缀
- [ ] 字段名使用 snake_case
- [ ] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

### 性能要求

- [ ] 权限查询响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（Redis，如适用）

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] API 测试覆盖所有接口
- [ ] 权限筛选逻辑测试覆盖所有场景

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] SQL 注入防护（使用参数化查询）
- [ ] 权限验证（确保用户只能访问有权限的数据）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

---

## 验收标准

### 功能验收

1. **权限数据结构**：
   - 收银机权限下存在"接单"权限（与沽清权限平级，sort=11）
   - "接单"权限下存在三个子项：扫码接单、外送、外卖
   - 原有接单权限（1724320522）已调整为"扫码接单"，parent_uuid 已更新
   - 原有外送权限（1752716650）的 parent_uuid 已更新为新接单权限UUID
   - 新增外卖权限已创建，parent_uuid 为新接单权限UUID
   - 权限数据正确存储在数据库中
   - 权限层级关系正确
   - 新"接单"权限和"外卖"权限已分配给 Cashier 和 Store Manager 角色

2. **权限筛选**：
   - 云平台未开启外卖时，"外卖"权限子项不显示在权限列表中
   - 云平台开启外卖时，"外卖"权限子项显示在权限列表中
   - 权限筛选逻辑正确，不影响其他权限

3. **默认角色权限分配**：
   - 云平台开启外卖时，默认角色（Cashier、Store Manager）自动拥有"外卖"权限
   - 新建商家时，如果云平台已开启外卖，默认角色自动分配"外卖"权限

4. **API 支持**：
   - BaseInfo API 正确返回 `is_open_grab_delivery` 字段
   - 字段值与数据库中的 `enable_grab_delivery` 一致

### 测试验收

1. **单元测试**: 覆盖率达标
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 权限筛选和分配逻辑测试通过
4. **手动测试**: 权限配置和显示功能测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确（待创建）
2. **数据库文档**: 迁移脚本和表结构文档完整
3. **测试文档**: tasks.md 中的测试任务完成（待创建）

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Go 1.23+
- 遵循 `.cursor/rules/go-main.mdc` 规范
- 不使用 panic，返回 error
- 所有代码使用中文注释

#### PHP Admin 模块

- 必须使用 ThinkPHP 6.0
- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数
- 使用软删除

### 业务约束

- 权限子项的添加不影响现有权限配置
- 云平台外卖开关状态变更时，不自动移除已分配的权限（保留用户配置）
- 权限筛选逻辑必须与云平台外卖开关状态保持一致

### 资源约束

- 开发时间: 2-3 天
- Story Point: SP3-SP5 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- `admin/app/common/model/shop/Access.php` - 权限模型
- `main/app/service/role_access.go` - 权限服务
- `admin/app/admin/controller/Shop.php` - 商家管理 Controller
- `main/app/service/auth.go` - 认证服务（BaseInfo API）

### 服务依赖

- **Admin → Main**: HTTP API 调用（如需要）
- **Frontend → Admin/Main**: HTTP API 调用（BaseInfo API）

### 业务依赖

- 依赖云平台外卖开关功能（`story-cloud-platform-merchant-grab-delivery-toggle`）
- 依赖权限系统的基础功能
- 依赖角色管理功能

---

## 风险和缓解

### 风险 1: 权限数据迁移影响现有权限配置

**影响**: 中  
**概率**: 低  
**缓解措施**:
- 迁移脚本使用 `updateOrInsertData` 方法，避免重复插入
- 迁移前备份权限数据
- 在测试环境充分测试迁移脚本

### 风险 2: 权限筛选逻辑不完整

**影响**: 高  
**概率**: 中  
**缓解措施**:
- 全面梳理所有权限查询接口
- 确保 Go Main 和 PHP Admin 模块的权限筛选逻辑一致
- 建立完整的测试用例覆盖所有权限筛选场景

### 风险 3: 默认角色权限分配时机不当

**影响**: 中  
**概率**: 低  
**缓解措施**:
- 明确权限分配的触发时机（商家新建/编辑时）
- 确保权限分配逻辑幂等（重复执行不影响结果）
- 记录权限分配日志，便于排查问题

---

## 时间表

- **Phase 1 - 数据库和权限数据**: 1 天
  - 数据库迁移脚本（新增接单权限、更新原有权限、新增外卖权限）
  - 权限数据初始化
  - 默认角色权限分配
- **Phase 2 - 权限筛选逻辑**: 1 天
  - Go Main 模块权限筛选
  - PHP Admin 模块权限筛选
- **Phase 3 - 默认角色权限分配**: 0.5 天
  - 权限分配逻辑实现
  - 迁移脚本中的默认角色权限分配
- **Phase 4 - API 确认和测试**: 0.5 天
  - BaseInfo API 字段确认
  - 单元测试和集成测试
- **总计**: 3 天（SP = 4-5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/php.mdc` - PHP 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 参考实现

- `admin/database/migrations/20250717013925_add_cashier_permission_delivery_order_to_access.php` - 外送权限迁移脚本
- `main/app/service/role_access.go` - 权限筛选逻辑
- `admin/app/common/model/shop/Access.php` - PHP 权限模型
- `docs/shared/specs/active/story-cloud-platform-merchant-grab-delivery-toggle/` - 云平台外卖开关功能

### 架构文档

- `docs/human/architecture/features/role_access.md` - 权限系统架构
- `docs/human/architecture/features/auth.md` - 认证系统架构

### 开发指南

- `docs/human/guides/php-development.md` - PHP 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{user}/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-05  
**作者**: 开发团队  
**审核者**: 待审核
