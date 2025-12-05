> ⚠️ **已归档** - 此 Spec 已随 v2.10.0 发布。
>
> - 归档时间: 2025-12-05
> - 归档人: weifashi

# 在 /assistant/desk/ping 接口中返回已选国旗 ID 需求文档

> 本文档定义「在 /assistant/desk/ping 接口中返回已选国旗 ID」功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                                    |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| **来源 Proposal** | [docs/team/proposals/2025-11/assistant-desk-ping-nationality.md](../../../../team/proposals/2025-11/assistant-desk-ping-nationality.md) |
| **创建日期**      | 2025-11-25                                                                                                                              |
| **负责人**        | {姓名}                                                                                                                                  |
| **目标 Sprint**   | Sprint {N}                                                                                                                              |
| **涉及技术栈**    | [x] Go (main/) [ ] Go (ttpos-bmp/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                                              |

---

## 📋 概述

在 `/assistant/desk/ping` 接口的响应中添加 `nationality_uuid` 字段，返回当前桌台订单的已选国籍 ID。该功能旨在减少前端接口调用次数，简化前端逻辑，提升用户体验和系统性能。

## 🎯 产品对齐

- **提升前端开发效率**：减少接口调用次数，简化前端逻辑
- **改善用户体验**：前端可以实时显示当前桌台的国籍信息
- **降低系统负载**：减少不必要的接口请求
- **保持数据一致性**：在轮询接口中统一返回桌台相关信息

## 📝 用户故事

**作为** 助手端开发人员  
**我想** 在 `/assistant/desk/ping` 接口响应中获取当前桌台订单的国籍 ID  
**以便于** 在前端 UI 中显示国籍信息，无需额外接口调用

---

## 功能需求

### Requirement 1: API 响应字段扩展

**用户故事**: 作为 助手端开发人员，我想在 `/assistant/desk/ping` 接口响应中获取当前桌台订单的国籍 ID，以便于在前端 UI 中显示国籍信息。

#### 验收标准

1. **WHEN** 调用 `/assistant/desk/ping` 接口 **THEN** 响应中 **SHALL** 包含 `nationality_uuid` 字段
2. **IF** 桌台已开台且订单已设置国籍 **THEN** `nationality_uuid` **SHALL** 返回对应的国籍 UUID（大于 0）
3. **IF** 桌台未开台或订单未设置国籍 **THEN** `nationality_uuid` **SHALL** 返回 `0`
4. **WHEN** 通过 `/assistant/desk/set_nationality` 设置国籍后 **THEN** 下次轮询 `/assistant/desk/ping` **SHALL** 返回更新后的 `nationality_uuid`

#### 具体要求

- [ ] 1.1 在 `resp.DeskPing` 结构体中添加 `NationalityUuid uint64` 字段，JSON 标签为 `nationality_uuid`
- [ ] 1.2 在 `service.GetDeskPing` 方法中，从 `desk.SaleBill.NationalityUuid` 获取值并赋值给响应
- [ ] 1.3 当 `desk.SaleBill` 为 `nil` 时，`nationality_uuid` 返回 `0`
- [ ] 1.4 当 `desk.SaleBill.NationalityUuid` 为 `0` 时，表示未设置国籍，返回 `0`
- [ ] 1.5 更新 Swagger 文档，说明新增字段含义

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

### API 设计要求

- [x] URL 使用 snake_case 命名（已符合：`/assistant/desk/ping`）
- [x] data 字段必须是对象，不能是 null 或数组（已符合）
- [x] 响应格式：`{code, message, data{}}`（已符合）
- [x] 参考: `.cursor/rules/api.mdc` - API 设计规范

### 数据库设计要求

- [x] 无需数据库变更，使用现有 `SaleBill.NationalityUuid` 字段

### 性能要求

- [x] 本地响应时间 < 200ms（无额外查询，性能影响可忽略）
- [x] 无需额外数据库查询（从已有数据中读取）

### 测试要求

- [ ] API 测试覆盖新增字段返回逻辑
- [ ] 测试未开台场景（返回 `0`）
- [ ] 测试已开台但未设置国籍场景（返回 `0`）
- [ ] 测试已设置国籍场景（返回对应 UUID）
- [ ] 测试设置国籍后轮询场景（返回更新后的 UUID）

### 安全要求

- [x] 所有 API 需要身份验证（已符合）
- [x] 无敏感数据变更

### 可靠性要求

- [x] 向后兼容：未设置国籍时返回 `0`，不影响现有前端逻辑
- [x] 错误处理：确保 `desk.SaleBill` 为 `nil` 时不会 panic

---

## 验收标准

### 功能验收

1. **API 响应字段**: `/assistant/desk/ping` 接口响应中包含 `nationality_uuid` 字段
2. **数据正确性**: 返回的 `nationality_uuid` 值与 `SaleBill.NationalityUuid` 一致
3. **边界处理**: 未开台或未设置国籍时返回 `0`
4. **实时性**: 设置国籍后，下次轮询立即返回更新后的值

### 测试验收

1. **API 测试**: 所有场景测试通过
2. **手动测试**: 验证前端可以正确获取和显示国籍信息

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: Swagger 文档已更新

---

## 约束条件

### 技术约束

#### Go Main 模块

- 必须使用 Gin 框架（已符合）
- 不使用 panic，返回 error（已符合）
- Service 只能依赖其他 Service 接口（已符合）

### 业务约束

- 保持向后兼容：未设置国籍时返回 `0`
- 不影响现有前端逻辑

### 资源约束

- 开发时间: 0.5 天
- Story Point: 1 (必须 ≤ 5)

---

## 依赖关系

### 技术依赖

- 无新增依赖

### 服务依赖

- 无新增服务依赖

### 业务依赖

- 依赖现有功能：`/assistant/desk/set_nationality` 接口（已存在）
- 依赖现有数据：`SaleBill.NationalityUuid` 字段（已存在）

---

## 风险和缓解

### 风险 1: 数据一致性

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 在 Service 层添加空值检查，确保 `desk.SaleBill` 不为 `nil` 时再读取
- 使用现有的 `SaleBill.NationalityUuid` 字段，无需额外查询

### 风险 2: 向后兼容

**影响**: 低  
**概率**: 低  
**缓解措施**:

- 未设置国籍时返回 `0`，前端已有处理 `0` 值的逻辑（参考其他接口）
- 在 API 文档中明确说明：`nationality_uuid` 为 `0` 时表示未设置国籍

---

## 时间表

- **Phase 1 - 响应结构扩展**: 0.1 天
- **Phase 2 - Service 层实现**: 0.1 天
- **Phase 3 - 文档更新**: 0.1 天
- **Phase 4 - 测试验证**: 0.2 天
- **总计**: 0.5 天（SP = 1）

---

## 参考资料

### 核心规范

- `.cursor/rules/go-main.mdc` - Go Main 核心约束
- `.cursor/rules/api.mdc` - API 设计规范

### 相关文档

- API 文档: `main/docs/swagger.yaml`
- 响应结构: `main/app/dto/resp/desk.go`
- Service 实现: `main/app/service/desk.go`
- 数据模型: `main/app/model/sale_bill.go`
- 相关接口: `/assistant/desk/set_nationality` - 设置桌台订单国籍接口
- 相关 Spec: `story-order-source-nationality` - 订单来源和国籍功能

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-25  
**作者**: TTPOS Team  
**审核者**: {审核者}
