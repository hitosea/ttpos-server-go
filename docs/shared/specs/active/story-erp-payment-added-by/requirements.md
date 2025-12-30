# 支付方式系统标识字段 需求文档

> 本文档定义支付方式系统标识字段功能的详细需求和验收标准。

## 📋 基本信息

| 项目              | 内容                                                                                                                     |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------ |
| **来源 Proposal** | [docs/team/proposals/2025-12/v2.12-payment-system-added-by.md](../../../../team/proposals/2025-12/v2.12-payment-system-added-by.md) |
| **创建日期**      | 2025-12-30                                                                                                               |
| **负责人**        | rikugun                                                                                                                  |
| **目标 Sprint**   | Sprint TBD                                                                                                               |
| **涉及技术栈**    | [x] Go (ttpos-bmp/) [ ] Go (main/) [ ] PHP (admin/) [ ] Vue (admin/views/)                                               |

## 📋 审核状态

| 项目         | 内容                             |
| ------------ | -------------------------------- |
| **审核状态** | 已通过                           |
| **审核人**   | rikugun                          |
| **审核日期** | 2025-12-30                       |
| **审核意见** | 需求清晰，技术设计合理，进入开发 |

---

## 📋 概述

在 ERP 支付方式管理中，需要区分系统自动创建的默认支付方式（如 Cash、Balance、Free Meal 等）和用户手动创建的自定义支付方式。通过在创建请求中新增 `added_by` 字段，当值为 "sys" 时，支付方式序号固定为 0000，便于系统识别和管理。

该功能主要用于：
1. 系统初始化时创建标准支付方式
2. 数据迁移场景下快速识别系统支付方式
3. 保持多公司间系统支付方式命名一致性
4. 简化运维和故障排查

## 🎯 产品对齐

该功能支持以下产品目标：
- **提升系统可维护性**：通过固定序号标识系统支付方式，便于运维识别和管理
- **规范数据管理**：明确区分系统数据和用户数据，避免误操作
- **支持多公司部署**：确保不同公司的系统支付方式命名一致
- **简化数据迁移**：在系统升级或迁移时快速处理系统级支付方式

## 📝 用户故事

**作为** 系统管理员  
**我想** 在创建支付方式时标识其来源（系统自动创建 vs 用户手动创建）  
**以便于** 快速识别系统默认支付方式，避免误操作，简化数据迁移和故障排查

---

## 功能需求

### Requirement 1: 新增 added_by 字段

**用户故事**: 作为开发者，我想在 SaveModeOfPaymentReq 中新增 added_by 字段，以便于标识支付方式的创建来源。

#### 验收标准

1. **WHEN** 调用 SaveModeOfPayment 接口 **AND** added_by 字段存在 **THEN** 系统 **SHALL** 根据 added_by 值决定序号生成策略
2. **WHEN** added_by = "sys" **THEN** 系统 **SHALL** 使用固定序号 0000
3. **WHEN** added_by 为空或其他值 **THEN** 系统 **SHALL** 使用自动递增序号（0001, 0002, 0003...）
4. **WHEN** 未传入 added_by 字段 **THEN** 系统 **SHALL** 保持现有行为（自动递增序号）

#### 具体要求

- [ ] 1.1 在 `SaveModeOfPaymentReq` protobuf 定义中新增 `optional string added_by = 8` 字段
- [ ] 1.2 字段为可选类型（optional），确保向后兼容
- [ ] 1.3 字段注释清晰，说明用途和取值规范
- [ ] 1.4 生成对应的 Go 代码（使用 `gf gen pb` 命令）

---

### Requirement 2: 序号生成逻辑调整

**用户故事**: 作为系统管理员，我想系统创建的支付方式使用固定序号 0000，以便于快速识别系统支付方式。

#### 验收标准

1. **WHEN** createModeOfPayment 处理请求 **AND** added_by = "sys" **THEN** 系统 **SHALL** 跳过 nextModeOfPaymentSeq 调用
2. **WHEN** added_by = "sys" **THEN** 系统 **SHALL** 将 nextSeq 设置为 0
3. **WHEN** added_by 为空或其他值 **THEN** 系统 **SHALL** 调用 nextModeOfPaymentSeq 获取序号
4. **WHEN** 生成支付方式名称 **THEN** 系统 **SHALL** 使用 %04d 格式化序号（如 0000, 0001...）

#### 具体要求

- [ ] 2.1 在 `createModeOfPayment` 方法中增加条件判断逻辑
- [ ] 2.2 当 `req.AddedBy != nil && strings.TrimSpace(*req.AddedBy) == "sys"` 时，nextSeq = 0
- [ ] 2.3 当 added_by 为空或非 "sys" 时，调用 nextModeOfPaymentSeq 获取序号
- [ ] 2.4 保持支付方式名称格式不变：`{prefix}{序号} - {公司缩写}`
- [ ] 2.5 添加日志记录，标注系统创建行为

---

### Requirement 3: 向后兼容性保证

**用户故事**: 作为开发者，我想确保新功能不影响现有客户端，以便于平滑升级。

#### 验收标准

1. **WHEN** 旧版本客户端不传 added_by 字段 **THEN** 系统 **SHALL** 使用自动递增序号
2. **WHEN** 旧版本客户端调用 SaveModeOfPayment **THEN** 系统 **SHALL** 正常工作，不返回错误
3. **WHEN** 新版本客户端传入 added_by = "sys" **THEN** 系统 **SHALL** 使用固定序号 0000

#### 具体要求

- [ ] 3.1 added_by 字段设计为 optional，未传入时保持现有逻辑
- [ ] 3.2 编写单元测试，覆盖旧客户端场景（不传 added_by）
- [ ] 3.3 编写单元测试，覆盖新客户端场景（传 added_by = "sys"）
- [ ] 3.4 编写单元测试，覆盖传入其他 added_by 值的场景

---

### Requirement 4: 序号冲突检测（可选增强）

**用户故事**: 作为系统管理员，我想系统检测序号 0000 是否已被占用，以便于避免冲突。

#### 验收标准

1. **WHEN** added_by = "sys" **AND** 序号 0000 已存在 **THEN** 系统 **SHALL** 返回错误信息
2. **WHEN** 序号 0000 未被占用 **THEN** 系统 **SHALL** 正常创建支付方式
3. **WHEN** 返回错误 **THEN** 错误信息 **SHALL** 包含冲突的支付方式名称

#### 具体要求

- [ ] 4.1 在创建前检查序号 0000 是否已存在（可选实现）
- [ ] 4.2 使用 `service.Doctype().Count` 查询同名支付方式
- [ ] 4.3 如存在冲突，返回清晰的错误信息
- [ ] 4.4 记录冲突日志，便于故障排查

---

## 非功能需求

### 代码架构和模块化

- **分层设计**: 严格遵循 Controller → Logic → DAO 分层
- **单一职责原则**: createModeOfPayment 方法职责清晰
- **模块化设计**: 序号生成逻辑独立可复用
- **遵循规范**:
  - `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
  - `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
  - `ttpos-bmp/.cursor/rules/go-ttpos-erp.mdc` - ttpos-erp 子模块规则

### API 设计要求

- [ ] gRPC 服务响应格式遵循 `erp.ResponseInfo` 规范
- [ ] logic/service 层返回具体业务数据类型，不返回 `erp.ResponseInfo`
- [ ] 错误信息使用中文，便于运维和调试
- [ ] 日志记录关键业务操作，便于追踪

### 数据库设计要求

本需求不涉及数据库表结构变更。

### 性能要求

- [ ] 序号生成逻辑响应时间 < 50ms
- [ ] 不影响现有支付方式创建性能

### 测试要求

- [ ] Logic 层测试覆盖率 ≥ 80%
- [ ] 覆盖所有 added_by 场景（sys/空/其他值/未传）
- [ ] 集成测试覆盖完整创建流程
- [ ] 回归测试确保不影响现有功能

### 国际化要求

- [ ] 日志记录使用中文
- [ ] 错误信息使用中文

### 安全要求

- [ ] 验证 added_by 字段内容，防止注入
- [ ] 只有授权用户可以使用 added_by = "sys" 创建系统支付方式（未来增强）

### 可靠性要求

- [ ] 序号冲突时优雅返回错误
- [ ] 记录错误日志，便于故障排查
- [ ] 事务管理确保数据一致性

---

## 验收标准

### 功能验收

1. **Protobuf 定义**: SaveModeOfPaymentReq 包含 added_by 字段
2. **序号生成逻辑**: added_by = "sys" 时序号为 0000
3. **向后兼容**: 旧客户端调用不受影响
4. **日志记录**: 创建系统支付方式时有日志记录

### 测试验收

1. **单元测试**: 覆盖率 ≥ 80%
2. **集成测试**: 完整创建流程测试通过
3. **回归测试**: 现有支付方式创建功能不受影响
4. **手动测试**: 使用 gRPC 客户端验证功能

### 文档验收

1. **技术文档**: design.md 完整且准确
2. **API 文档**: Protobuf 注释完整
3. **测试文档**: tasks.md 中的测试任务完成
4. **日志文档**: 记录审计日志示例

---

## 约束条件

### 技术约束

#### Go BMP 模块

- 必须使用 GoFrame 2.x
- 禁止修改 dao/entity/do/ 目录（自动生成）
- Protobuf 修改后必须执行 `gf gen pb` 重新生成
- 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- 遵循 `ttpos-bmp/.cursor/rules/go-ttpos-erp.mdc`

#### 特定约束

- added_by 字段必须设计为 optional，确保向后兼容
- 序号生成逻辑不能破坏现有 nextModeOfPaymentSeq 方法
- 日志记录必须使用中文

### 业务约束

- 系统支付方式序号 0000 应保留给系统使用
- 用户不应创建序号为 0000 的支付方式（通过 UI 限制）

### 资源约束

- 开发时间: 0.5 天
- Story Point: 1 SP

---

## 依赖关系

### 技术依赖

- `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto` - Protobuf 定义
- `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go` - 业务逻辑
- `ttpos-bmp/app/ttpos-erp/api/selling/selling.pb.go` - 自动生成的 API 代码

### 服务依赖

- **无外部服务依赖**

### 业务依赖

- **无前置条件**，独立功能

---

## 风险和缓解

### 风险 1: 序号 0000 已被用户占用

**影响**: 中  
**概率**: 低  
**缓解措施**:

- 在创建前检查序号 0000 是否已存在
- 如存在冲突，返回清晰的错误信息
- 提供手动处理指南（数据迁移脚本）

### 风险 2: 向后兼容性问题

**影响**: 高  
**概率**: 低  
**缓解措施**:

- 将 added_by 设计为 optional 字段
- 编写完整的回归测试
- 在测试环境充分验证旧客户端场景

---

## 时间表

- **Phase 1 - Protobuf 定义和代码生成**: 0.5 小时
- **Phase 2 - 业务逻辑实现**: 1 小时
- **Phase 3 - 测试和文档**: 1 小时
- **总计**: 2.5 小时（SP = 1）

---

## 参考资料

### 核心规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `ttpos-bmp/.cursor/rules/go-ttpos-erp.mdc` - ttpos-erp 子模块规则
- `.cursor/rules/security.mdc` - 安全开发规范

### 架构文档

- `docs/human/architecture/go-bmp-architecture.md` - Go BMP 架构

### 开发指南

- `docs/human/guides/go-bmp-development.md` - Go BMP 开发指南
- `README.MD` - 项目 README
- `MIGRATION_QUICK_START.md` - 迁移快速入门
- `DATABASE_MIGRATION_RULES.md` - 数据库迁移规则

### 外部参考

- [GoFrame 官方文档](https://goframe.org)
- [Protobuf 语言指南](https://developers.google.com/protocol-buffers/docs/proto3)

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2025-12/2025-12-30.md`
- 提醒：需求评审或范围调整若形成经验，应同步更新 Episode 并在 Proposal/Spec 互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-30  
**作者**: rikugun  
**审核者**: 待审核

