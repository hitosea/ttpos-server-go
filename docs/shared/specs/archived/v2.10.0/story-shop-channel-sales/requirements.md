> ⚠️ **已归档** - 此 Spec 已随 v2.10.0 发布。
>
> - 归档时间: 2025-12-05
> - 归档人: weifashi

# 渠道营业统计 API 需求文档

> 本文档定义渠道营业统计查询与导出的后端 API 需求与验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-11/new-admin-home-report-channel-sales.md](../../../../team/proposals/2025-11/new-admin-home-report-channel-sales.md) |
| **创建日期**      | 2025-11-26                                                                                                   |
| **负责人**        | 待指派                                                                                                       |
| **目标 Sprint**   | Sprint 待定                                                                                                  |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 已通过                   |
| **审核人**   | 产品组                   |
| **审核日期** | 2025-11-26               |
| **审核意见** | API 仅含 start/end，满足需求 |

---

## 📋 概述

在新管理端提供面向店长/运营的渠道营业统计接口，只实现后端 API：一个用于查询当前筛选条件下的渠道营业表现，一个用于导出相同数据的报表。统计逻辑复用 `main/app/repository/statistics.go` 中的 `CountSale` 能力，通过新增 Repository 方法抽取渠道相关的聚合结果，并沿用既有时间/时区处理实现的“今日”默认范围。

## 🎯 产品对齐

该能力支撑“新管理端首页渠道营业统计”与“报表中心渠道营业统计”需求，帮助店长无需多系统比对即可查看各渠道（堂食、点餐-店内、点餐-外卖、外送等）订单规模与客单，辅助渠道投放与排班决策，并提供导出用于经营复盘。

## 📝 用户故事

**作为** 商家管理端店长/运营  
**我想** 通过后端接口获取并导出按渠道拆分的营业数据  
**以便于** 在首页小部件与报表中心中快速展示/下载渠道表现

---

## 功能需求

### Requirement 1: 渠道营业统计查询 API

**用户故事**: 作为店长，我想在管理端请求渠道营业统计数据，以便在前端展示各渠道表现。

#### 验收标准

1. **WHEN** 未传时间范围 **THEN** 系统 **SHALL** 使用现有时间工具确定当前日期 00:00:00-23:59:59 作为默认查询区间。
2. **IF** 传入 `start_time` > `end_time`（Unix 秒） **THEN** 系统 **SHALL** 返回参数错误。
3. **WHEN** 商户已开启外卖/外送渠道 **AND** 有订单数据 **THEN** 结果中 **SHALL** 包含对应渠道的统计分组。
4. **WHEN** 任一渠道无订单 **THEN** 结果 **SHALL** 仍返回该渠道，金额与数量为 0。
5. **WHEN** 请求成功 **THEN** 响应结构 **SHALL** 符合 `{code, message, data{summary, dine_in, pickup, takeout_shop, takeout_delivery}}` 格式且各字段取自 Repository 聚合结果。

#### 具体要求

- [ ] 1.1 新增 `GET /api/v1/shop/statistics/channel_sales`（最终路径以 `main/app/api/v1/shop` 命名规范为准，使用 snake_case）。
- [ ] 1.2 请求参数：
  - `start_time`、`end_time`（int64，Unix 秒，可选；默认今日范围，依赖既有时间/时区实现）。
- [ ] 1.3 Controller 调用 Service；Service 调用 Repository 新增方法（例如 `CountChannelSale`）以复用 `CountSale` 逻辑。
- [ ] 1.4 返回字段（整数或 decimal，金额统一保留 2 位）：
  - `order_count`, `min_amount`, `max_amount`, `avg_amount`。
  - 对桌台渠道额外返回 `table_count`, `guest_count`。
- [ ] 1.5 对请求做权限校验（必须为当前店铺管理员会话）。
- [ ] 1.6 响应 `meta` 中附带查询时间区间与渠道列表，便于前端显示。

---

### Requirement 2: 渠道营业统计导出 API

**用户故事**: 作为店长，我想导出渠道营业统计报表，以便向老板或财务分享。

#### 验收标准

1. **WHEN** 调用导出接口 **THEN** 系统 **SHALL** 生成与查询接口相同口径的数据并以表格形式输出（见需求附件示例）。
2. **IF** 时间范围为空 **THEN** 默认与 Requirement 1 相同的今日范围。
3. **WHEN** 用户选择超出允许范围的时间（例如未来日期） **THEN** 接口 **SHALL** 返回校验错误。
4. **WHEN** 导出成功 **THEN** 返回带签名 URL 或直接流下载，文件名格式 `channel_sales_{shop_id}_{YYYYMMDDHHMM}.xlsx`。

#### 具体要求

- [ ] 2.1 新增 `GET /api/v1/shop/statistics/channel_sales/export`（或 POST，需携带签名 Token 防止 CSRF）。
- [ ] 2.2 请求参数仅包含 `start_time`、`end_time`（与查询接口一致），默认今日范围，并额外支持 `format`（默认 `xlsx`，后续可扩展 `csv`）。
- [ ] 2.3 导出模板需匹配附件示例：包含“合计、桌台、点餐-店内、点餐-外卖、外送”分组，列顺序与前端 UI 一致。
- [ ] 2.4 Service 端调用同一统计方法，避免重复计算；导出使用现有通用 Excel 工具（若无则复用报表模块）。
- [ ] 2.5 导出结果应写入操作日志，包含操作者、门店、时间范围。
- [ ] 2.6 文件生成需控制大小（≤5MB），必要时分页生成。

---

## 非功能需求

- **分层设计**：Controller → Service → Repository；复用、抽象 `CountSale` 内现有 SQL。
- **API 设计**：遵循 `.cursor/rules/api.mdc`；URL snake_case；响应 `code/message/data`。
- **性能**：默认今日查询需 <200ms；导出需异步或流式生成以避免阻塞。
- **时间处理**：沿用现有时间/时区处理逻辑，将传入 Unix 时间转换为查询区间，无需新增特殊约定。
- **安全**：接口需校验登录态与店铺权限；导出要做限速和审计。
- **可靠性**：当统计服务异常时返回清晰错误并记录日志。

## 验收标准

### 功能验收

1. **接口可用**：查询 API 在默认与自定义时间范围下返回正确渠道数据。
2. **渠道完整**：开启/关闭渠道场景均验证，缺省渠道返回 0。
3. **导出正确**：导出文件字段顺序、内容与示例一致，金金额度和数量与查询接口一致。

### 测试验收

1. **单元测试**：Repository 新增方法 ≥80% 覆盖。
2. **API 测试**：通用接口测试用例覆盖默认、跨天、无数据、导出等场景。
3. **集成测试**：验证首页与报表中心调用链（仅需模拟接口调用）。

### 文档验收

1. **design.md** 需描述接口定义、数据口径。
2. **API 文档** 更新渠道统计接口（包括导出）。
3. **tasks.md** 包含对应开发与测试任务。

---

## 约束条件

### 技术约束

- Go Main：必须使用 Gin，不允许 panic；Service 不直连 DB。
- Repository 新增方法需与 `CountSale` 共用 SQL/视图，避免重复维护。
- 所有金额字段使用 `decimal(20,8)`；返回前做 rounding。

### 业务约束

- 仅针对商家管理端的渠道数据，暂不支持自营多门店聚合。
- 时间范围不得超过未来；最大跨度若需限制由产品评审决定（默认支持任意历史范围，后续可在 design.md 中细化）。

### 资源约束

- 开发时间：5 天
- Story Point: 5（需在设计阶段确认）

---

## 依赖关系

### 技术依赖

- `main/app/repository/statistics.go`：复用 `CountSale`。
- 通用导出工具包（如 `pkg/excel`，待确认）。

### 服务依赖

- Main 模块内部，无跨服务依赖。

### 业务依赖

- 店铺渠道配置接口需可用，以判定启用渠道。

---

## 风险和缓解

### 风险 1: 渠道口径不一致

**影响**: 高  
**概率**: 中  
**缓解措施**：

- 与产品/数据组确认渠道枚举及映射关系。
- 在单元测试中覆盖多渠道组合。

### 风险 2: 导出性能瓶颈

**影响**: 中  
**概率**: 中  
**缓解措施**：

- 控制单次导出时间范围和行数。
- 如需生成大量数据，采用异步任务并通知用户。

---

## 时间表

- **Phase 1 - 统计抽象**: 2 天
- **Phase 2 - API 实现与测试**: 2 天
- **Phase 3 - 导出与验收**: 1 天
- **总计**: 5 天（SP = 5）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc`
- `.cursor/rules/api.mdc`
- `.cursor/rules/security.mdc`

### 架构文档

- `docs/human/architecture/go-main-architecture.md`

### 开发指南

- `docs/human/guides/go-main-development.md`
- `docs/human/guides/api-design-guide.md`

### 外部参考

- DooTask #36937 附件截图（渠道营业统计 UI & 导出示例）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{用户名}/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：若在实现过程中总结统计口径经验，请同步更新 Episode，并保持 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-26  
**作者**: 产品组（代）  
**审核者**: 待定


