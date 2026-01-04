# ERP 支付方式 PaymentID 字段 任务分解

> 本文档定义 ERP 支付方式 PaymentID 字段 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 10  
**已完成**: 7  
**进行中**: -  
**完成率**: 70%

---

## Phase 1: ERP 字段创建和迁移

- [x] 1.1 创建 ERP 迁移脚本

  - File: `ttpos-bmp/app/ttpos-erp/manifest/erp-migrate/v2.12/01_custom_field/01_custom_payment_id.json`
  - Purpose: 创建 ERP 自定义字段迁移脚本
  - Requirements: 1.1, 1.4
  - Leverage: 现有迁移脚本: `ttpos-bmp/app/ttpos-erp/manifest/erp-migrate/v2.7/02_custom_field/`
  - Success: 迁移脚本已创建，字段定义完整

- [ ] 1.2 执行 ERP 字段迁移

  - File: -
  - Purpose: 在 ERPNext 系统中创建 custom_payment_id 字段
  - Requirements: 1.1
  - Leverage: Task 1.1 的迁移脚本
  - Command: 通过 ERPNext 迁移工具或 API 执行迁移
  - Success: custom_payment_id 字段成功添加到 Mode of Payment DocType

- [ ] 1.3 验证字段创建

  - File: -
  - Purpose: 确认字段在 ERPNext 中可用
  - Requirements: 1.1
  - Leverage: ERPNext UI 或 API 查询
  - Success: 可以在 ERPNext 中看到并使用 custom_payment_id 字段

---

## Phase 2: Protobuf 定义更新

- [x] 2.1 更新 ModeOfPayment 消息

  - File: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
  - Purpose: 在 ModeOfPayment 消息中添加 payment_id 字段
  - Requirements: 1.1
  - Leverage: 现有 protobuf 定义
  - Success: payment_id 字段已添加到 ModeOfPayment 消息

- [x] 2.2 更新 SaveModeOfPaymentReq 消息

  - File: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
  - Purpose: 在 SaveModeOfPaymentReq 消息中添加 payment_id 字段
  - Requirements: 1.2
  - Leverage: 现有 protobuf 定义
  - Success: payment_id 字段已添加到 SaveModeOfPaymentReq 消息

- [x] 2.3 重新生成 API 代码

  - File: `ttpos-bmp/app/ttpos-erp/api/selling/selling.pb.go`
  - Purpose: 根据更新的 protobuf 定义重新生成 Go 代码
  - Requirements: 1.1, 1.2
  - Command: `cd ttpos-bmp/app/ttpos-erp && gf gen pb`
  - Success: API 代码重新生成成功，包含 payment_id 字段

---

## Phase 3: Logic 层实现

- [x] 3.1 修改 GetModeOfPaymentList 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 查询时包含 custom_payment_id 字段，返回时映射到 payment_id
  - Requirements: 1.1, 2.1
  - Leverage: 现有 GetModeOfPaymentList 实现
  - Success: 查询支付方式列表时正确返回 payment_id 字段

- [x] 3.2 修改 createModeOfPayment 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 创建时支持 PaymentID 自动生成（PID + 16位数字）
  - Requirements: 1.2, 1.3
  - Leverage: `ttpos-bmp/utility/uuid` 包，现有 createModeOfPayment 实现
  - Success: 创建支付方式时自动生成或使用提供的 PaymentID

- [x] 3.3 修改 updateModeOfPayment 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 更新时支持 payment_id 字段修改
  - Requirements: 2.2, 3.2
  - Leverage: 现有 updateModeOfPayment 实现
  - Success: 更新支付方式时可以修改 payment_id 字段

---

## Phase 4: 测试

- [x] 4.1 单元测试：PaymentID 生成逻辑

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/paymentid_test.go`
  - Purpose: 测试 PaymentID 生成格式和唯一性
  - Requirements: 1.2, 1.3
  - Leverage: Task 3.2 的实现
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 编写 PaymentID 生成测试 | Context: 测试自动生成格式（PID+16位数字），测试提供 PaymentID 的情况 | Restrictions: 测试覆盖率 ≥ 80% | Success: ✅ 所有测试通过（3个测试，1000个唯一ID生成）

- [x] 4.2 单元测试：权限校验

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/permission_test.go`
  - Purpose: 测试跨公司修改权限校验
  - Requirements: 3.2
  - Leverage: Task 3.3 的实现
  - Prompt: Role: QA Engineer | Task: 编写权限校验测试 | Context: 测试跨公司修改被拒绝，测试同公司修改成功 | Restrictions: 使用 Mock 数据 | Success: ✅ 所有测试通过（8个测试用例）

- [ ] 4.3 gRPC 接口测试：创建支付方式

  - File: `ttpos-bmp/app/ttpos-erp/test/api/selling_test.go` (或测试目录)
  - Purpose: 测试通过 gRPC 创建支付方式
  - Requirements: 1.1, 1.2, 2.1
  - Leverage: Task 3.2 的实现
  - Prompt: Role: QA Automation Engineer | Task: 编写 SaveModeOfPayment gRPC 测试 | Context: 测试不提供 payment_id 时自动生成，测试提供 payment_id 时使用该值 | Restrictions: 使用真实 gRPC 调用 | Success: 创建成功，PaymentID 正确

- [ ] 4.4 gRPC 接口测试：查询支付方式

  - File: `ttpos-bmp/app/ttpos-erp/test/api/selling_test.go`
  - Purpose: 测试通过 gRPC 查询支付方式
  - Requirements: 1.1, 2.1
  - Leverage: Task 3.1 的实现
  - Prompt: Role: QA Automation Engineer | Task: 编写 GetModeOfPaymentList gRPC 测试 | Context: 测试查询返回 payment_id 字段 | Restrictions: 使用真实 gRPC 调用 | Success: 查询成功，返回数据包含 payment_id

- [ ] 4.5 gRPC 接口测试：更新支付方式

  - File: `ttpos-bmp/app/ttpos-erp/test/api/selling_test.go`
  - Purpose: 测试通过 gRPC 更新支付方式
  - Requirements: 2.2, 3.2
  - Leverage: Task 3.3 的实现
  - Prompt: Role: QA Automation Engineer | Task: 编写 SaveModeOfPayment 更新模式测试 | Context: 测试传入 name 参数时执行更新，测试更新 payment_id 和 enabled 字段 | Restrictions: 使用真实 gRPC 调用 | Success: 更新成功，数据正确

- [ ] 4.6 集成测试：端到端流程

  - File: -
  - Purpose: 测试完整的创建、查询、更新流程
  - Requirements: 所有功能需求
  - Leverage: 所有已实现的功能
  - Success: TTPOS 通过 gRPC 创建 → 自动生成 PaymentID → 查询返回 PaymentID → 更新成功

---

## Phase 5: 文档和部署

- [ ] 5.1 更新 API 文档

  - File: `docs/shared/api/erp_api.md`
  - Purpose: 记录 PaymentID 字段的使用说明
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Success: API 文档完整准确

- [ ] 5.2 更新 CHANGELOG

  - File: `CHANGELOG.md`
  - Purpose: 记录功能变更
  - Requirements: 文档要求
  - Leverage: 现有 CHANGELOG 格式
  - Success: CHANGELOG 更新完成

- [ ] 5.3 创建部署说明

  - File: `docs/shared/specs/active/story-erp-mode-of-payments-paymentid/deployment.md`
  - Purpose: 记录部署步骤和注意事项
  - Requirements: 运维要求
  - Leverage: 现有部署文档
  - Success: 部署说明清晰完整，包含 ERP 迁移执行步骤

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [x] Phase 1-3 任务标记为 `[x]`
- [ ] Phase 4-5 任务标记为 `[x]`
- [x] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Logic: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [x] Protobuf 定义已更新
- [x] Logic 层已实现
- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成：
  - ✅ PaymentID 自动生成（PID + 16位数字）
  - ✅ gRPC 接口支持 PaymentID 读写
  - ⏳ ERP 字段迁移完成
  - ⏳ 测试通过

### 文档同步

- [ ] API 文档已更新
- [ ] CHANGELOG.md 已更新
- [ ] 部署文档已创建
- [ ] ERP 迁移脚本已验证

### 规范遵循

- [x] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [x] 遵循 `ttpos-bmp/.cursor/rules/erpnext.mdc`
- [x] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [x] 日志使用中文描述

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-erp-mode-of-payments-paymentid/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-erp-mode-of-payments-paymentid/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-erp-mode-of-payments-paymentid/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-erp-mode-of-payments-paymentid/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-erp-mode-of-payments-paymentid/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务（按 Phase 顺序）
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 查看 design.md 中的设计细节
4. **查看复用**: 检查 Leverage 中的可复用代码
5. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
6. **实现代码**: 按照规范实现功能
7. **运行检查**: `go fmt`, `go vet`, `go test`
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Go BMP 开发

```
Role: Go Developer specializing in GoFrame and ERPNext integration

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc, ttpos-bmp/.cursor/rules/erpnext.mdc

Restrictions:
- 使用 GoFrame 2.x 框架
- 禁止修改 dao/entity/do/ 目录（自动生成）
- 与 ERPNext 交互使用通用服务
- 使用 g.Log() 记录日志（中文描述）
- 不使用 panic，返回 error
- DTO 定义在 internal/model/dto/erp/

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 80%
```

### Protobuf 开发

```
Role: Go Developer with Protocol Buffers expertise

Task: {具体任务描述}

Context:
- Current file: {protobuf 文件路径}
- Requirements: {需求编号}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 请求消息以 Req 结尾
- 响应消息以 Resp 结尾
- 字段名使用 snake_case
- 添加中文注释说明字段用途

Success Criteria:
- Protobuf 定义格式正确
- gf gen pb 命令执行成功
- 生成的 Go 代码可编译
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2025-12/2025-12-23.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v2.0.0  
**创建日期**: 2025-12-23  
**更新日期**: 2025-12-23  
**维护者**: ttpos-erp 开发组
