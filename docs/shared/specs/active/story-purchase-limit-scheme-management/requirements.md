# 品牌采购限购方案管理 需求文档

> 本文档定义品牌采购限购方案管理功能的详细需求和验收标准（Phase 1: 限购方案 CRUD + 数据迁移）。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/v2.15.0-purchase-limit-scheme-adjustment.md](../../../../team/proposals/2026-01/v2.15.0-purchase-limit-scheme-adjustment.md) |
| **创建日期**      | 2026-01-20                                                                                                 |
| **负责人**        | 待分配                                                                                                       |
| **目标 Sprint**   | Sprint 15 (v2.15.0)                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |
| **关联任务**      | DooTask #38970                                                                                               |
| **拆分说明**      | 从 story-purchase-limit-scheme 拆分，Phase 1: 方案管理                                                         |
| **后续 Story**    | [story-purchase-limit-scheme-validation](../story-purchase-limit-scheme-validation/) - Phase 2: 限购校验      |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 开发中 |
| **审核人**   | weifashi             |
| **审核日期** | 2026-01-20             |
| **审核意见** | 仅后端实现，Phase 1 先完成方案管理和数据迁移         |
| **设计状态** | ✅ 已完成 |
| **设计人**   | weifashi |
| **设计日期** | 2026-01-20 |

---

## 📋 概述

在商家管理端的"参数设置"模块中新增"限购业务"功能，支持总部创建和管理多个限购方案。本 Story 专注于限购方案的 CRUD 操作和数据迁移，为后续的限购校验功能提供数据基础。

### Phase 1 范围

1. **限购方案管理**：创建、读取、更新、删除限购方案
2. **周期配置**：选择限购方案的生效周期（星期几）
3. **物品配置**：选择物品并为每个物品配置限额（数量+单位）
4. **门店配置**：选择限购方案适用的门店范围
5. **数据迁移**：将旧表 `ttpos_purchase_quota_config` 和 `ttpos_purchase_quota_config_shop` 的数据迁移到新方案表，然后删除旧表

## 🎯 产品对齐

本功能支持以下产品目标：

1. **提升采购管理精细化**：通过方案化管理，支持按时间段（星期）、门店、物品进行精细化限购控制
2. **优化库存管理**：引导门店合理规划采购，减少库存积压和浪费
3. **规范采购行为**：通过限额和次数控制，规范门店采购行为，降低总部审核工作量

## 📝 用户故事

### 主要用户故事：总部采购管理员

**作为** 总部采购管理员  
**我想** 创建和管理多个限购方案，为不同物品配置精细化的申请限额规则（按工作日、门店、单位、数量）  
**以便于** 引导门店合理规划采购，优化库存管理，规范采购行为

---

## 功能需求

### Requirement 1: 限购方案管理（总部视角）

**用户故事**: 作为总部采购管理员，我想在"参数设置"中管理多个限购方案，以便于灵活控制不同场景的采购限额

#### 验收标准

1. **WHEN** 用户进入商家管理端的"参数设置"页面 **THEN** 系统 **SHALL** 显示"限购业务"卡片入口
2. **WHEN** 用户查看"限购业务"卡片 **THEN** 系统 **SHALL** 显示描述文字："限制采购的数量、物品、门店等"
3. **WHEN** 用户点击"限购业务"卡片 **THEN** 系统 **SHALL** 跳转到限购方案列表页面
4. **WHEN** 用户进入限购方案列表页面 **THEN** 系统 **SHALL** 以卡片形式展示所有限购方案
5. **WHEN** 限购方案卡片展示 **THEN** 系统 **SHALL** 显示：方案名称、周期（如"周一、周三、周五"）、每日申请次数（如"5次"）、门店数量（如"10门店"或"全部门店"）、物品数量（如"45物品"）
6. **WHEN** 用户点击方案卡片 **THEN** 系统 **SHALL** 进入设置限购方案页面（编辑模式）
7. **WHEN** 用户点击页面底部的"新增"按钮 **THEN** 系统 **SHALL** 进入设置限购方案页面（新增模式）

#### 具体要求

- [x] 1.1 API 路由：`GET /shop/purchase/limit_scheme/list` - 获取限购方案列表
- [x] 1.2 API 路由：`GET /shop/purchase/limit_scheme/:id` - 获取限购方案详情
- [x] 1.3 限购方案列表返回核心信息摘要（名称、周期、次数、门店、物品数量）
- [x] 1.4 支持分页查询（可选）

---

### Requirement 2: 限购方案配置（总部视角）

**用户故事**: 作为总部采购管理员，我想配置限购方案的各项参数（名称、状态、周期、物品、门店、次数），以便于实现精细化的限购管理

#### 验收标准

8. **WHEN** 用户进入设置限购方案页面（新增模式） **THEN** 系统 **SHALL** 显示表单字段：名称（必填）、状态开关、周期、物品、门店、每日申请次数
9. **WHEN** 用户进入设置限购方案页面（编辑模式） **THEN** 系统 **SHALL** 显示表单字段和已保存的数据，底部显示"删除"和"保存"按钮
10. **WHEN** 用户未填写名称就点击保存 **THEN** 系统 **SHALL** 显示红色错误提示"名称不能为空"并阻止保存
11. **WHEN** 用户切换状态开关 **THEN** 系统 **SHALL** 更新方案状态（开启/关闭）
12. **WHEN** 用户点击周期字段 **THEN** 系统 **SHALL** 进入周期选择内页
13. **WHEN** 用户点击物品字段 **THEN** 系统 **SHALL** 跳转到物品选择页面
14. **WHEN** 用户点击门店字段 **THEN** 系统 **SHALL** 弹出门店选择器
15. **WHEN** 用户在每日申请次数输入框中输入数值 **THEN** 系统 **SHALL** 保存该方案的每日申请次数限制
16. **WHEN** 用户点击保存按钮 **THEN** 系统 **SHALL** 验证必填项（名称），验证通过后保存方案并返回列表页面
17. **WHEN** 用户在编辑模式点击删除按钮 **THEN** 系统 **SHALL** 弹出确认对话框，显示提示语"确定要删除限购方案"{方案名称}"吗？删除后无法恢复。"，确认后删除方案并返回列表页面

#### 具体要求

- [x] 2.1 API 路由：`POST /shop/purchase/limit_scheme/create` - 创建限购方案
- [x] 2.2 API 路由：`PUT /shop/purchase/limit_scheme/:id` - 更新限购方案
- [x] 2.3 API 路由：`DELETE /shop/purchase/limit_scheme/:id` - 删除限购方案
- [x] 2.4 方案名称为必填项，最多 50 个字符
- [x] 2.5 状态开关默认为"开启"（status=1）
- [x] 2.6 周期配置默认为空，保存时至少选择一个星期
- [x] 2.7 物品配置默认为空，保存时至少选择一个物品
- [x] 2.8 门店配置默认为"全部门店"（apply_to_all_shops=1）
- [x] 2.9 每日申请次数默认为空（不限制），如果输入则必须 > 0
- [x] 2.10 保存时校验必填项和数据有效性
- [x] 2.11 删除操作使用软删除（delete_time）

---

### Requirement 3: 周期选择（总部视角）

**用户故事**: 作为总部采购管理员，我想选择限购方案的生效周期（星期几），以便于实现按工作日的精细化限购

#### 验收标准

13. **WHEN** 用户进入周期选择内页 **THEN** 系统 **SHALL** 显示"活动周期"标题和星期选择网格（周一到周日），支持多选
14. **WHEN** 用户在周期选择内页未选择任何星期就点击确定 **THEN** 系统 **SHALL** 显示错误提示"至少需选择一个星期"并阻止保存
15. **WHEN** 用户在周期选择内页选择星期并点击"确定"按钮 **THEN** 系统 **SHALL** 保存周期配置并返回设置限购方案页面，更新周期字段显示（如"周一、周三、周五"）
16. **WHEN** 用户在周期选择内页点击"取消"按钮 **THEN** 系统 **SHALL** 放弃修改，返回设置限购方案页面

#### 具体要求

- [x] 3.1 周期配置存储在 `ttpos_purchase_limit_scheme_weekday` 表
- [x] 3.2 weekday 字段：1-7 分别代表周一到周日
- [x] 3.3 至少选择一个星期才能保存
- [x] 3.4 支持多选，一个方案可以配置多个星期

---

### Requirement 4: 物品选择和限额配置（总部视角）

**用户故事**: 作为总部采购管理员，我想选择物品并为每个物品配置限额（数量+单位），以便于精细化控制不同物品的采购量

#### 验收标准

18. **WHEN** 用户进入物品选择页面 **THEN** 系统 **SHALL** 显示搜索框和物品列表（按分类展示）
19. **WHEN** 用户未选择任何物品就点击确定 **THEN** 系统 **SHALL** 在顶部显示红色提示"最少需选1个物品"并阻止确定
20. **WHEN** 用户在搜索框输入关键词 **THEN** 系统 **SHALL** 按"物品名称/内部编码/条形码"搜索并过滤物品列表
21. **WHEN** 物品按分类展示 **THEN** 系统 **SHALL** 显示分类名称和分类下的物品数量（红色徽章）
22. **WHEN** 用户查看物品列表 **THEN** 系统 **SHALL** 为每个物品显示：物品名称、内部编码、限额配置区域
23. **WHEN** 用户为物品配置限额项 **THEN** 系统 **SHALL** 显示数量输入框（带"+"和"-"按钮）、单位显示（固定显示该物品的销售单位，如果物品没有销售单位则显示基准单位，不可选择）、删除按钮（红色圆圈减号图标）
24. **WHEN** 用户点击物品限额配置区域的"+"按钮 **THEN** 系统 **SHALL** 添加新的限额项（数量+单位组合，单位固定为该物品的销售单位，如果物品没有销售单位则使用基准单位）
25. **WHEN** 用户点击限额项的删除按钮 **THEN** 系统 **SHALL** 从列表中移除该限额项
26. **WHEN** 用户配置物品的限额项数量 **THEN** 系统 **SHALL** 保存配置，数量可以为空（为空则不限制数量），如果输入数量则最小值为 0.001，单位固定为该物品的销售单位（如果物品没有销售单位则使用基准单位，不可选择）
27. **WHEN** 用户勾选底部"全选"复选框 **THEN** 系统 **SHALL** 选中所有物品
28. **WHEN** 用户选择物品后 **THEN** 系统 **SHALL** 在底部显示"已选X个"
29. **WHEN** 用户点击确定按钮 **THEN** 系统 **SHALL** 验证至少选择1个物品，验证通过后保存选择的物品和限额配置，返回设置限购方案页面并更新物品字段显示

#### 具体要求

- [x] 4.1 物品配置存储在 `ttpos_purchase_limit_scheme_item` 表
- [x] 4.2 限额配置的单位固定为物品的销售单位（如果没有销售单位则使用基准单位）
- [x] 4.3 数量配置默认为空（表示不限制数量）
- [x] 4.4 数量输入框支持小数，最小值 0.001，最大值 999999.999
- [x] 4.5 至少选择 1 个物品才能保存
- [x] 4.6 物品选择使用现有的物品查询接口

---

### Requirement 5: 门店选择（总部视角）

**用户故事**: 作为总部采购管理员，我想选择限购方案适用的门店范围（全部或部分），以便于针对不同门店实施不同的限购策略

#### 验收标准

30. **WHEN** 用户点击门店字段 **THEN** 系统 **SHALL** 弹出门店选择器
31. **WHEN** 用户选择"全部门店" **THEN** 系统 **SHALL** 保存为全部门店，在设置页面显示"已选择 (全部)"
32. **WHEN** 用户选择部分门店并点击确定 **THEN** 系统 **SHALL** 更新门店配置，并在设置页面显示"已选择 (X)"（X为门店数量）

#### 具体要求

- [x] 5.1 门店配置存储在 `ttpos_purchase_limit_scheme_shop` 表和 `apply_to_all_shops` 字段
- [x] 5.2 默认选择"全部门店"（apply_to_all_shops=1）
- [x] 5.3 指定门店模式（apply_to_all_shops=0），门店UUID存储在 `ttpos_purchase_limit_scheme_shop` 表

---

### Requirement 6: 数据迁移（系统维护）

**用户故事**: 作为系统维护人员，我需要将旧表数据迁移到新方案表，以便于保持业务连续性

#### 验收标准

33. **WHEN** 执行数据迁移脚本 **THEN** 系统 **SHALL** 将 `ttpos_purchase_quota_config` 表的数据迁移到新方案表
34. **WHEN** 执行数据迁移脚本 **THEN** 系统 **SHALL** 将 `ttpos_purchase_quota_config_shop` 表的数据迁移到新方案门店表
35. **WHEN** 数据迁移完成后 **THEN** 系统 **SHALL** 删除旧表 `ttpos_purchase_quota_config` 和 `ttpos_purchase_quota_config_shop`
36. **WHEN** 迁移过程中发生错误 **THEN** 系统 **SHALL** 回滚所有操作，保持数据一致性

#### 具体要求

- [x] 6.1 创建数据迁移文件：`admin/database/migrations/{timestamp}_migrate_purchase_quota_to_limit_scheme.php`
- [x] 6.2 迁移逻辑：
  - 为每个 `ttpos_purchase_quota_config` 记录创建一个限购方案
  - 方案名称：使用物品名称 + "-限购方案"
  - 状态：根据旧表的状态映射（1=开启，0=关闭）
  - 周期：默认全周（周一到周日）
  - 物品配置：从旧表的 material_code 和 quota_limit 映射
  - 门店配置：从 `ttpos_purchase_quota_config_shop` 表映射
- [x] 6.3 删除旧表前先备份数据到日志文件
- [x] 6.4 使用事务保证数据一致性
- [x] 6.5 提供回滚脚本（如果需要恢复旧表）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/structs.mdc` - 项目结构规范

### API 设计要求

- [x] URL 使用 snake_case 命名（如：`/shop/purchase/limit_scheme/list`）
- [x] data 字段必须是对象，不能是 null 或数组
- [x] 分页信息统一放在 meta 中
- [x] 响应格式：`{code, message, data{}}`
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 必须包含: `id`, `uuid`, `create_time`, `update_time`, `delete_time`
- [x] 时间字段使用 int 类型，\_time 结尾，默认值 0
- [x] 金额字段使用 decimal(20,8)
- [x] UUID 字段使用 bigint unsigned
- [x] 表名使用 ttpos\_ 前缀
- [x] 字段名使用 snake_case
- [x] 参考: `.cursor/rules/database.mdc` - 数据库开发规范

**新增表结构**：
- `ttpos_purchase_limit_scheme`：限购方案表
- `ttpos_purchase_limit_scheme_item`：限购方案物品配置表
- `ttpos_purchase_limit_scheme_shop`：限购方案门店配置表
- `ttpos_purchase_limit_scheme_weekday`：限购方案星期配置表

### 性能要求

- [x] 本地响应时间 < 200ms
- [x] 数据库查询优化（使用索引）
  - `ttpos_purchase_limit_scheme`: uuid, company_uuid, status
  - `ttpos_purchase_limit_scheme_item`: scheme_id, material_code
  - `ttpos_purchase_limit_scheme_shop`: scheme_id, company_uuid
  - `ttpos_purchase_limit_scheme_weekday`: scheme_id, weekday
- [x] 缓存策略（Redis）
  - 限购方案配置缓存（TTL: 5分钟）
- [x] 并发处理（使用 UUID 锁）

### 测试要求

- [x] Service 层测试覆盖率 ≥ 70%
- [x] Repository 层测试覆盖率 ≥ 80%
- [x] 集成测试覆盖核心流程
  - 限购方案创建和配置流程
  - 数据迁移流程
- [x] API 测试覆盖所有接口
- [x] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [x] 支持 10 种语言（中文、英文、日语、韩语等）
- [x] 所有文案使用多语言实现
  - 限购方案相关提示文案
  - 错误提示文案
- [x] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证
- [x] 总部和门店权限隔离（总部可以管理限购方案，门店只能查看）
- [x] 敏感数据加密存储
- [x] SQL 注入防护（使用参数化查询）
- [x] XSS 防护（前端输入校验）
- [x] CSRF 防护（Token 验证）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [x] 网络异常时优雅降级
- [x] 事务管理（保证数据一致性）
  - 限购方案创建、更新、删除使用事务
  - 数据迁移使用事务
- [x] 错误日志记录（使用 Logger）
  - 记录限购方案配置错误
  - 记录数据迁移过程和结果
- [x] 故障恢复机制
  - 数据迁移失败时自动回滚
  - 提供手动回滚脚本

---

## 验收标准

### 功能验收

1. **限购方案管理**: 总部可以创建、编辑、删除限购方案，方案列表正确显示所有方案信息
2. **限购方案配置**: 总部可以配置方案的名称、状态、周期、物品、门店、次数，所有配置项保存成功
3. **周期选择**: 总部可以选择方案的生效周期（星期几），至少选择一个星期才能保存
4. **物品选择和限额配置**: 总部可以选择物品并为每个物品配置限额（数量+单位），数量可以为空
5. **门店选择**: 总部可以选择方案适用的门店范围（全部或部分）
6. **数据迁移**: 旧表数据成功迁移到新方案表，旧表被删除，业务连续性得到保证

### 测试验收

1. **单元测试**: Service 层覆盖率 ≥ 70%，Repository 层覆盖率 ≥ 80%
2. **API 测试**: 所有接口测试通过，包括正常流程和异常流程
3. **集成测试**: 端到端流程测试通过，包括限购方案创建、数据迁移
4. **数据迁移测试**: 迁移脚本在测试环境成功执行，数据完整性得到验证

### 文档验收

1. **技术文档**: design.md 完整且准确，包含数据库设计、API 设计
2. **API 文档**: API 接口文档完整，包括请求参数、响应格式、错误码
3. **数据库文档**: 迁移脚本和表结构文档完整，包含新增的 4 个表
4. **迁移文档**: 数据迁移流程文档完整，包含迁移步骤、回滚步骤

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾（如 `IPurchaseLimitSchemeSrv`）
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error
- 参考现有代码：`main/app/service/purchase_order/purchase_quota.go`

### 业务约束

- 限购方案只在总部配置，门店只能查看
- 方案名称必须唯一（同一总部内）
- 至少选择一个星期和一个物品才能保存方案
- 删除方案使用软删除，不物理删除数据
- 数据迁移必须在生产环境维护窗口期执行

### 资源约束

- 开发时间: 3-4 天
- Story Point: 3

---

## 依赖关系

### 技术依赖

- `github.com/gin-gonic/gin` - HTTP 框架
- `gorm.io/gorm` - ORM 框架
- `ttpos-server-go/app/model` - 数据模型
- `ttpos-server-go/app/repository` - 数据访问层
- `ttpos-server-go/app/service/purchase_order` - 采购申请服务（现有）
- `ttpos-server-go/pkg/utils` - 工具包（时区处理）

### 服务依赖

- **Frontend → Admin → Main**: HTTP API 调用
- **Main → Database**: MySQL 数据库（读写）
- **Main → Redis**: 缓存服务（限购方案配置缓存）

### 业务依赖

- 依赖现有的物品管理功能（`product` 模块）
- 依赖现有的门店管理功能（`company` 模块）
- **不依赖** Story 2（限购校验功能）

---

## 风险和缓解

### 风险 1: 数据迁移风险 - 迁移过程中可能出现数据丢失或不一致

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 在测试环境多次验证迁移脚本
- 生产环境迁移前备份旧表数据
- 使用事务保证原子性
- 提供回滚脚本
- 迁移后验证数据完整性

### 风险 2: 旧表删除风险 - 删除旧表后无法恢复

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 删除前将数据导出到日志文件
- 保留备份文件至少 30 天
- 提供从备份恢复旧表的脚本
- 在删除旧表前确认新方案功能正常

### 风险 3: API 兼容性风险 - 新 API 可能与现有系统不兼容

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 使用新的 API 路由，不修改现有 API
- 提供 API 版本标识
- 详细的 API 文档和示例
- 充分的集成测试

---

## 时间表

- **Phase 1 - 数据库设计和迁移**: 1 天
  - 设计 4 个新表结构
  - 编写迁移脚本（创建表 + 迁移数据 + 删除旧表）
  - 创建测试数据
- **Phase 2 - 后端 API 开发**: 2 天
  - 限购方案 CRUD 接口
  - Model 层和 Repository 层
  - Service 层
- **Phase 3 - 测试和文档**: 0.5-1 天
  - 单元测试和集成测试
  - API 文档
  - 数据迁移验证
- **总计**: 3.5-4 天（SP = 3）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范
- `.cursor/rules/structs.mdc` - 项目结构规范

### 现有代码参考

- `main/app/service/purchase_order/purchase_quota.go` - 现有限购功能
- `main/app/service/purchase_order/purchase_order.go` - 采购申请服务
- `main/app/repository/purchase_quota_config.go` - 限购配置 Repository
- `main/app/api/v1/shop/shop_purchase.go` - 采购申请 API

### 原型和需求

- 原型链接: https://modao.cc/proto/NYlDfREZt0gr57g5xvn9XE/sharing?view_mode=device&screen=rbpV8FJlbQCm1ruKH
- DooTask 任务: #38970
- 提案文档: `docs/team/proposals/2026-01/v2.15.0-purchase-limit-scheme-adjustment.md`
- 原 Spec: `docs/shared/specs/active/story-purchase-limit-scheme/` (已拆分)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/weifashi/2026-01/2026-01-20.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-20  
**作者**: weifashi  
**审核者**: weifashi  
**Story Point**: 3
