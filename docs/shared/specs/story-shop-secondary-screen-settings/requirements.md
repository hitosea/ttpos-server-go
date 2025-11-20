# 副屏设置接口开发 需求文档

> 本文档定义副屏设置接口开发的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-11-20-secondary-screen-settings.md](../../../team/proposals/2025-11-20-secondary-screen-settings.md) |
| **创建日期**      | 2025-11-20                                                                                                 |
| **负责人**        | 曾振华                                                                                                       |
| **目标 Sprint**   | Sprint {N}                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

---

## 📋 概述

**当前状态**：收银端设置中已有未点餐时的副屏轮播图设置（`cashier.carousel`），但缺少轮播间隔设置，以及点餐时的设置（展示模式和轮播间隔）。

**本次开发**：在现有收银机设置基础上扩展，添加：
1. 未点餐时的轮播间隔设置
2. 点餐时的展示模式设置（轮播内容、点餐内容、点餐内容+轮播内容）
3. 点餐时的轮播间隔设置

收银端通过接口获取副屏配置，前端根据当前点餐状态动态展示相应内容。

**本次开发仅限接口处理，前端不需要处理。**

## 🎯 产品对齐

该功能支持门店管理需求：
- **提升营销效果**：通过副屏轮播广告，增加品牌曝光和促销信息展示
- **改善用户体验**：顾客可以在副屏查看订单详情，减少询问
- **增强品牌形象**：专业的双屏收银体验，提升门店形象
- **提高运营效率**：统一管理副屏内容，支持多门店配置

## 📝 用户故事

**作为** 商户管理员  
**我想** 在管理端配置收银机副屏的显示内容和模式  
**以便于** 提升营销效果和顾客体验

**作为** 收银员  
**我想** 收银端能够获取副屏配置并展示内容  
**以便于** 为顾客提供更好的服务体验

---

## 功能需求

### Requirement 1: 未点餐时轮播间隔设置

**用户故事**: 作为商户管理员，我想设置未点餐时副屏轮播图的切换间隔，以便于控制轮播速度

**说明**: 当前系统已有未点餐时的副屏轮播图设置（`cashier.carousel`），本次新增轮播间隔设置。轮播内容最多支持15个图片或视频。

#### 验收标准

1. **WHEN** 调用收银机设置保存接口，传入未点餐时轮播间隔参数（10-120秒） **THEN** 系统 **SHALL** 保存到 `cashier.no_order_carousel_interval` 字段
2. **WHEN** 未传入轮播间隔参数 **THEN** 系统 **SHALL** 使用默认值10秒
3. **WHEN** 传入间隔时间不在有效范围（10-120秒） **THEN** 系统 **SHALL** 返回参数验证错误
4. **WHEN** 调用收银机设置查询接口 **THEN** 系统 **SHALL** 返回未点餐时轮播间隔设置

#### 具体要求

- [ ] 1.1 在现有收银机设置（`cashier`）中扩展字段 `no_order_carousel_interval`（未点餐时轮播间隔）
- [ ] 1.2 轮播间隔范围：10-120秒，默认10秒
- [ ] 1.3 参数验证（范围检查）
- [ ] 1.4 保存到 `setting` 表的 `cashier` 配置 JSON 中
- [ ] 1.5 轮播内容数量限制：最多15个图片或视频

---

### Requirement 2: 点餐时副屏展示模式设置

**用户故事**: 作为商户管理员，我想设置点餐时副屏的展示模式，以便于控制点餐时的显示内容

#### 验收标准

1. **WHEN** 调用收银机设置保存接口，传入点餐时展示模式参数 **THEN** 系统 **SHALL** 保存到 `cashier.order_display_mode` 字段
2. **WHEN** 调用收银机设置查询接口 **THEN** 系统 **SHALL** 返回当前配置的点餐时展示模式
3. **WHEN** 传入无效的展示模式值 **THEN** 系统 **SHALL** 返回参数验证错误
4. **WHEN** 显示模式为"点餐内容+轮播内容" **THEN** 系统 **SHALL** 保存该模式配置

#### 具体要求

- [ ] 2.1 在现有收银机设置（`cashier`）中扩展字段 `order_display_mode`（点餐时展示模式）
- [ ] 2.2 支持三种显示模式：carousel(轮播内容)、order(点餐内容)、order_carousel(点餐内容+轮播内容)
- [ ] 2.3 显示模式参数验证（仅允许上述三种值）
- [ ] 2.4 默认值：carousel（轮播内容）
- [ ] 2.5 保存到 `setting` 表的 `cashier` 配置 JSON 中

---

### Requirement 3: 点餐时轮播间隔设置

**用户故事**: 作为商户管理员，我想设置点餐时副屏轮播内容的切换间隔，以便于控制轮播速度

#### 验收标准

1. **WHEN** 调用收银机设置保存接口，传入点餐时轮播间隔参数（10-120秒） **THEN** 系统 **SHALL** 保存到 `cashier.order_carousel_interval` 字段
2. **WHEN** 未传入轮播间隔参数 **THEN** 系统 **SHALL** 使用默认值10秒
3. **WHEN** 传入间隔时间不在有效范围（10-120秒） **THEN** 系统 **SHALL** 返回参数验证错误
4. **WHEN** 调用收银机设置查询接口 **THEN** 系统 **SHALL** 返回点餐时轮播间隔设置

#### 具体要求

- [ ] 3.1 在现有收银机设置（`cashier`）中扩展字段 `order_carousel_interval`（点餐时轮播间隔）
- [ ] 3.2 轮播间隔范围：10-120秒，默认10秒
- [ ] 3.3 参数验证（范围检查）
- [ ] 3.4 保存到 `setting` 表的 `cashier` 配置 JSON 中

---

### Requirement 4: 收银端获取副屏配置接口

**说明**: 收银端通过现有的收银机设置接口获取配置，本次扩展返回新增的字段。

**用户故事**: 作为收银端应用，我想获取门店的副屏配置信息，以便于根据配置展示副屏内容

#### 验收标准

1. **WHEN** 调用现有收银机设置接口（`GetCashierSetting`） **THEN** 系统 **SHALL** 返回扩展后的配置，包含新增字段：
   - `no_order_carousel_interval`（未点餐时轮播间隔）
   - `order_display_mode`（点餐时展示模式）
   - `order_carousel_interval`（点餐时轮播间隔）
2. **WHEN** 门店未配置新增字段 **THEN** 系统 **SHALL** 返回默认值（轮播间隔10秒，展示模式carousel）
3. **WHEN** 返回配置信息 **THEN** 系统 **SHALL** 包含：
   - 未点餐时：`carousel`（轮播内容列表，已存在，最多15个）、`no_order_carousel_interval`（轮播间隔）
   - 点餐时：`order_display_mode`（展示模式）、`order_carousel_interval`（轮播间隔，当模式包含轮播时）
4. **WHEN** 前端根据配置和当前点餐状态处理显示逻辑 **THEN** 后端不做点餐状态判断和内容切换处理
5. **WHEN** 保存轮播内容时，传入超过15个内容 **THEN** 系统 **SHALL** 返回参数验证错误

#### 具体要求

- [ ] 4.1 扩展现有 `CashierResp` 结构体，添加新字段：
  - `NoOrderCarouselInterval int` - 未点餐时轮播间隔
  - `OrderDisplayMode string` - 点餐时展示模式
  - `OrderCarouselInterval int` - 点餐时轮播间隔
- [ ] 4.2 在 `GetCashierSetting` 方法中解析并返回新增字段
- [ ] 4.3 未配置时返回默认值（轮播间隔10秒，展示模式carousel）
- [ ] 4.4 后端不做点餐状态判断，仅返回配置信息
- [ ] 4.5 轮播内容数量限制：最多15个图片或视频（在保存接口中验证）

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Service → Repository 分层
- **单一职责原则**: 每个文件应有单一、明确的目的
- **模块化设计**: Service 和 Repository 应独立且可复用
- **依赖管理**: Service 只能依赖其他 Service 接口，不能直接依赖 Repository
- **遵循规范**:
  - `.cursor/rules/go-main.mdc` - Go Main 开发规范
  - `.cursor/rules/api.mdc` - API 设计规范
  - `.cursor/rules/database.mdc` - 数据库开发规范

### API 设计要求

- [ ] URL 使用 snake_case 命名（如：`/api/v1/secondary_screen/setting`）
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

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（Redis，收银端配置缓存）
- [ ] 并发处理（使用 UUID 锁）

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] 集成测试覆盖核心流程
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 安全要求

- [ ] 所有 API 需要身份验证
- [ ] 文件上传校验（格式、大小）
- [ ] SQL 注入防护（使用参数化查询）
- [ ] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **未点餐时轮播间隔**: 支持设置未点餐时轮播间隔（10-120秒，默认10秒），保存到 `cashier.no_order_carousel_interval`
2. **点餐时展示模式**: 支持设置点餐时展示模式（carousel/order/order_carousel），保存到 `cashier.order_display_mode`
3. **点餐时轮播间隔**: 支持设置点餐时轮播间隔（10-120秒，默认10秒），保存到 `cashier.order_carousel_interval`
4. **轮播内容数量限制**: 轮播内容（carousel）最多15个图片或视频，保存时进行数量验证
5. **配置获取**: 收银端通过现有接口能够正确获取扩展后的副屏配置信息（包含新增字段）

### 测试验收

1. **单元测试**: 覆盖率达标（Service ≥ 70%, Repository ≥ 80%）
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整
3. **数据库文档**: 迁移脚本和表结构文档完整
4. **测试文档**: tasks.md 中的测试任务完成

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架
- 接口以 `I` 开头，实现以 `Impl` 结尾
- Service 只能依赖其他 Service 接口
- Repository 只能持有 db 实例，不能持有 DBManager
- 不使用 panic，返回 error

### 业务约束

- 仅限接口开发，前端不需要处理
- 设备尺寸信息（商米、COMPAX）仅作为前端页面提示，后端不做设备类型区分
- 预览效果由前端根据配置信息自行处理，不需要专门的预览接口
- 收银端仅提供配置获取接口，不处理点餐状态判断和内容切换逻辑

### 资源约束

- 开发时间: 5-7 天
- Story Point: 8 SP（待技术评审确认，需拆分至 ≤ 5）

---

## 依赖关系

### 技术依赖

- 现有收银机设置服务：`main/app/service/setting/setting.go` - `GetCashierSetting` 方法
- 现有设置存储：`setting` 表，key 为 `cashier`，values 为 JSON 格式
- 现有轮播内容：`cashier.carousel` 字段（已存在，无需修改）

### 服务依赖

- **Frontend → Main**: HTTP API 调用（获取配置）

### 业务依赖

- 门店ID：从context上下文中获取（`ctx.GetCompanyUuid()`），不需要额外接口

---

## 风险和缓解

### 风险 1: 现有配置兼容性

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 新增字段使用默认值，不影响现有功能
- 向后兼容：未配置新字段时返回默认值
- 测试现有收银机设置接口不受影响

### 风险 2: JSON 字段扩展

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 使用 JSON 合并策略，保留现有字段
- 字段验证确保数据格式正确

---

## 时间表

- **Phase 1 - 数据结构扩展**: 0.5 天（扩展 CashierResp 和设置 JSON）
- **Phase 2 - 管理端接口扩展**: 1-2 天（扩展现有收银机设置保存接口）
- **Phase 3 - 收银端接口扩展**: 1 天（扩展 GetCashierSetting 返回新字段）
- **Phase 4 - 测试联调**: 1 天
- **总计**: 3.5-4.5 天（SP = 5，符合要求）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范
- `.cursor/rules/database.mdc` - 数据库开发规范
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-main-architecture.md` - Go Main 架构

### 开发指南

- `docs/human/guides/go-main-development.md` - Go Main 开发指南
- `docs/human/guides/api-design-guide.md` - API 设计指南
- `docs/human/guides/database-guide.md` - 数据库开发指南

### 外部参考

- 任务 #36935：新管理端-副屏设置（详细需求）
- 任务 #36982：收银机-副屏（应用实现）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-20  
**作者**: 曾振华  
**审核者**: {待审核}

