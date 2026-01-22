# 旧商家后台-统计报表-导出-加上Grab数据/LINEMAN数据 需求文档

> 本文档定义 旧商家后台统计报表导出增加 Grab 和 LINE MAN 数据 的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2026-01/legacy-shop-statistics-export-add-grab-lineman-data.md](../../../../team/proposals/2026-01/legacy-shop-statistics-export-add-grab-lineman-data.md) |
| **创建日期**      | 2026-01-19                                                                                                 |
| **负责人**        | 待分配                                                                                                       |
| **目标 Sprint**   | 待定                                                                                                   |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                   |

## 📋 审核状态

| 项目         | 内容                     |
| ------------ | ------------------------ |
| **审核状态** | 待审核 |
| **审核人**   | -             |
| **审核日期** | -             |
| **审核意见** | -         |

---

## 📋 概述

在旧商家后台的统计报表导出功能中，为销售统计（按天）和支付数据（按天）两个报表增加 Grab 和 LINE MAN 平台的数据支持。确保这两个平台的数据能够正确统计并显示在导出的报表中，且 Grab 和 LINE MAN 数据排列在最后。

## 🎯 产品对齐

该功能支持商家全面分析各平台（包括 Grab 和 LINE MAN）的销售表现，提供完整的多渠道销售数据视图，帮助商家更好地进行经营决策和渠道优化。

## 📝 用户故事

**作为** 商户管理员  
**我想** 在统计报表导出中查看包含 Grab 和 LINE MAN 平台的完整销售和支付数据  
**以便于** 全面分析各渠道的经营表现，做出更好的经营决策

---

## 功能需求

### Requirement 1: 销售统计（按天）导出增加 Grab 和 LINE MAN 数据

**用户故事**: 作为商户管理员，我想在销售统计（按天）导出报表中看到 Grab 和 LINE MAN 平台的销售数据，以便于全面了解各渠道的销售情况

#### 验收标准

1. **WHEN** 商户管理员导出销售统计（按天）报表 **THEN** 系统 **SHALL** 包含 Grab 和 LINE MAN 平台的销售数据
2. **IF** Grab 或 LINE MAN 平台有订单数据 **THEN** 系统 **SHALL** 正确统计并显示在导出的报表中
3. **WHEN** 导出报表时 **THEN** Grab 和 LINE MAN 数据 **SHALL** 排列在最后

#### 具体要求

- [ ] 1.1 在 `CountSaleResp` 结构体中添加 Grab 和 LINE MAN 的统计字段（订单数、最小/最大/平均订单金额）
- [ ] 1.2 在 `CountSaleDays` 方法中增加对 Grab 和 LINE MAN 渠道的统计支持
- [ ] 1.3 计算 Grab 和 LINE MAN 的统计指标：订单数、最小订单金额、最大订单金额、平均订单金额
- [ ] 1.4 确保 Grab 和 LINE MAN 数据在导出报表中正确显示
- [ ] 1.5 确保 Grab 和 LINE MAN 数据排列在所有其他渠道数据之后
- [ ] 1.6 验证数据统计的准确性和完整性（包括统计指标的计算准确性）

---

### Requirement 2: 支付数据（按天）导出增加 Grab 和 LINE MAN 数据

**用户故事**: 作为商户管理员，我想在支付数据（按天）导出报表中看到 Grab 和 LINE MAN 平台的支付数据，以便于全面了解各渠道的支付情况

#### 验收标准

1. **WHEN** 商户管理员导出支付数据（按天）报表 **THEN** 系统 **SHALL** 包含 Grab 和 LINE MAN 平台的支付数据
2. **IF** Grab 或 LINE MAN 平台有支付数据 **THEN** 系统 **SHALL** 正确统计并显示在导出的报表中
3. **WHEN** 导出报表时 **THEN** Grab 和 LINE MAN 数据 **SHALL** 排列在最后

#### 具体要求

- [ ] 2.1 在 `CountPaymentDays` 方法中增加对 Grab 和 LINE MAN 渠道的统计支持
- [ ] 2.2 确保 Grab 和 LINE MAN 支付数据在导出报表中正确显示
- [ ] 2.3 确保 Grab 和 LINE MAN 数据排列在所有其他支付方式数据之后
- [ ] 2.4 验证支付数据统计的准确性和完整性

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

- [x] URL 使用 snake_case 命名（如：`/api/v1/order_info`）
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

### 性能要求

- [ ] 本地响应时间 < 200ms
- [ ] 数据库查询优化（使用索引）
- [ ] 缓存策略（Redis）
- [ ] 并发处理（使用 UUID 锁）

### 浏览器兼容性（管理后台）

- [ ] Chrome 90+
- [ ] Safari 14+
- [ ] Firefox 88+
- [ ] Edge 90+

### 测试要求

- [ ] Service 层测试覆盖率 ≥ 70%
- [ ] Repository 层测试覆盖率 ≥ 80%
- [ ] 集成测试覆盖核心流程
- [ ] API 测试覆盖所有接口
- [ ] 参考: `.cursor/rules/go-main.mdc` - 测试规范

### 国际化要求

- [ ] 支持 10 种语言（中文、英文、日语、韩语等）
- [ ] 所有文案使用多语言实现
- [ ] 参考: `main/i18n/` - 国际化配置

### 安全要求

- [x] 所有 API 需要身份验证
- [x] 敏感数据加密存储
- [x] SQL 注入防护（使用参数化查询）
- [x] XSS 防护（前端输入校验）
- [x] CSRF 防护（Token 验证）
- [x] 参考: `.cursor/rules/security.mdc` - 安全开发规范

### 可靠性要求

- [ ] 网络异常时优雅降级
- [ ] 事务管理（保证数据一致性）
- [ ] 错误日志记录（使用 Logger）
- [ ] 故障恢复机制

---

## 验收标准

### 功能验收

1. **销售统计导出**: 导出的销售统计（按天）报表包含 Grab 和 LINE MAN 数据，且排列在最后
2. **支付数据导出**: 导出的支付数据（按天）报表包含 Grab 和 LINE MAN 数据，且排列在最后
3. **数据准确性**: Grab 和 LINE MAN 的数据统计准确，与订单数据一致

### 测试验收

1. **单元测试**: 覆盖率达标
2. **API 测试**: 所有接口测试通过
3. **集成测试**: 端到端流程测试通过
4. **手动测试**: 导出功能测试通过，验证数据完整性和排序

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: API 接口文档完整（如有）
3. **测试文档**: tasks.md 中的测试任务完成

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

- 必须保持与现有导出格式的兼容性
- Grab 和 LINE MAN 数据必须排列在最后
- 数据统计逻辑必须与现有平台数据保持一致

### 资源约束

- 开发时间: 3-5 天
- Story Point: 待技术评审确认（必须 ≤ 5）

---

## 依赖关系

### 技术依赖

- `main/app/service/statistics.go` - 统计服务
- `main/app/repository/statistics.go` - 统计数据仓库
- `main/app/api/v1/shop/shop_statistics.go` - 统计 API 接口

### 服务依赖

- **Main → Main**: 内部服务调用

### 业务依赖

- 依赖现有的统计报表导出功能
- 依赖 Grab 和 LINE MAN 订单数据源

---

## 风险和缓解

### 风险 1: Grab 和 LINE MAN 数据源可能不完整或格式不一致

**影响**: 中  
**概率**: 中  
**缓解措施**:

- 提前验证 Grab 和 LINE MAN 数据源的完整性和格式
- 在开发前进行数据源调研和测试

### 风险 2: 需要确保数据统计逻辑与现有平台数据保持一致

**影响**: 高  
**概率**: 中  
**缓解措施**:

- 复用现有的统计逻辑，确保数据一致性
- 进行充分的数据对比测试

### 风险 3: 导出格式需要兼容现有报表结构

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 在现有导出格式基础上扩展，保持向后兼容
- 进行回归测试确保不影响现有功能

---

## 时间表

- **Phase 1 - 需求分析和技术调研**: 1 天
- **Phase 2 - 开发实现**: 2-3 天
- **Phase 3 - 测试和优化**: 1 天
- **总计**: 3-5 天（SP = 待技术评审确认）

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

### 相关代码

- `main/app/service/statistics.go` - 统计服务实现
- `main/app/service/business.go` - 营业数据服务（包含渠道统计）
- `main/app/api/v1/shop/shop_statistics.go` - 统计 API 接口
- `main/app/repository/statistics.go` - 统计数据仓库

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-19  
**作者**: 王昱  
**审核者**: 待审核
