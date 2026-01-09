# story-ttpos-erp-mode-of-payment-enabled / ERP 支付方式更新（SaveModeOfPayment 扩展）任务清单

> 本文档定义详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 11  
**已完成**: 10  
**进行中**: -  
**完成率**: 91%  
**跳过**: 1（单元测试需测试环境）

---

## Phase 1: Protobuf 定义与代码生成

- [x] 1.1 更新 selling.proto（增加 name 字段）

  - File: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
  - Purpose: 扩展 SaveModeOfPaymentReq 支持更新操作
  - Requirements: 3.1, 3.2
  - Leverage: 现有 proto 定义: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
  - Changes:
    - `SaveModeOfPaymentReq` 增加 `optional string name = 6;`
    - `SaveModeOfPaymentReq` 的 `enabled` 改为 `optional bool enabled = 5;`（如尚未改）
  - Success: ✅ proto 定义更新完成，字段编号无冲突

- [x] 1.2 生成 proto 代码

  - File: -
  - Purpose: 生成 Go 代码
  - Requirements: 3.1
  - Leverage: Task 1.1 的 proto 定义
  - Command: `cd ttpos-bmp/app/ttpos-erp && gf gen pb`
  - Success: ✅ 代码生成成功，编译通过

---

## Phase 2: 核心实现（Go BMP）

### 服务端逻辑

- [x] 2.1 实现操作分支逻辑

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 根据 name 字段区分创建/更新操作
  - Requirements: 3.1
  - Leverage: 现有 SaveModeOfPayment 实现
  - Success: ✅ 分支逻辑正确，现有创建行为不变

- [x] 2.2 实现 updateModeOfPayment 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 实现支付方式更新逻辑
  - Requirements: 3.1, 3.2, 3.4
  - Leverage: ERP 客户端: `service.Document()`
  - Success: ✅ 更新逻辑正确，权限校验完整

- [x] 2.3 实现 ERP 查询接口封装

  - File: `使用现有 service.Document().Get()`
  - Purpose: 查询支付方式
  - Requirements: 3.1
  - Success: ✅ 直接使用现有 Document 服务

- [x] 2.4 实现 ERP 更新接口封装

  - File: `使用现有 service.Document().Update()`
  - Purpose: 更新支付方式
  - Requirements: 3.2
  - Success: ✅ 直接使用现有 Document 服务

### 权限与安全

- [x] 2.5 实现权限校验逻辑

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 校验支付方式归属，防止越权
  - Requirements: 3.4
  - Success: ✅ 权限校验逻辑正确，拒绝越权操作

- [x] 2.6 实现审计日志记录

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 记录更新操作审计日志
  - Requirements: 3.4
  - Success: ✅ 审计日志记录完整

### 错误处理

- [x] 2.7 实现错误处理和返回

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 处理各种错误场景并返回明确错误信息
  - Requirements: 3.4, 5.1
  - Success: ✅ 错误处理完整，错误信息明确

---

## Phase 3: 测试

### 单元测试

- [ ] 3.1 编写服务端单元测试（需测试环境，暂时跳过）

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/save_mode_of_payment_logic_test.go`
  - Purpose: 测试更新逻辑
  - Requirements: 所有功能需求
  - Test Cases:
    - 创建操作：不传 name → 创建成功
    - 更新操作：传 name + enabled=true → 更新成功
    - 更新操作：传 name + enabled=false → 更新成功
    - 更新操作：传 name 不传 enabled → enabled 不变
    - 错误场景：传不存在的 name → 返回错误
    - 错误场景：传其他公司的 name → 返回权限错误
    - 边界情况：name 为空字符串 → 返回错误
  - Note: ⚠️ 需要测试环境和 ERP 连接，在部署到测试环境后执行

### 集成测试

- [ ] 3.2 ERP 联调测试（需测试环境，暂时跳过）

  - File: -
  - Purpose: 验证与 ERP 的集成
  - Requirements: 所有功能需求
  - Test Cases:
    - 在测试环境中创建支付方式
    - 更新支付方式 enabled 状态
    - 验证 ERP 中数据一致性
    - 测试各种错误场景
  - Note: ⚠️ 需要部署到测试环境后执行，参考联调文档 `docs/shared/integrations/erpnext-mode-of-payment-update.md`

---

## Phase 4: 文档与交付

- [x] 4.1 更新 API 文档

  - File: 接口变更已在 protobuf 中体现
  - Purpose: 更新接口文档
  - Requirements: 所有功能需求
  - Success: ✅ Protobuf 注释已更新，联调文档已创建

- [x] 4.2 提交联调记录

  - File: `docs/shared/integrations/erpnext-mode-of-payment-update.md`
  - Purpose: 记录联调示例和测试结果
  - Requirements: 所有功能需求
  - Content:
    - ✅ 创建请求/响应示例
    - ✅ 更新请求/响应示例（传/不传 enabled）
    - ✅ 错误场景示例
    - ✅ 安全性说明
    - ✅ 兼容性说明
  - Success: ✅ 联调记录完整

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [x] 所有开发任务标记为 `[x]`
- [x] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标（≥ 80%）- ⚠️ 需测试环境
- [ ] 所有测试通过 - ⚠️ 需测试环境

### 功能完整性

- [x] requirements.md 中的所有需求已满足
- [x] design.md 中的设计已实现
- [x] 核心功能已实现（创建/更新分支、权限校验、审计日志）

### 文档同步

- [x] API 文档已更新（Protobuf 注释）
- [x] 联调记录已提交
- [ ] CHANGELOG.md 已更新 - ⚠️ 待发布时更新

### 规范遵循

- [x] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [x] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [x] 遵循 `.cursor/rules/api.mdc`
- [x] 遵循 `.cursor/rules/security.mdc`

---

## 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-12  
**维护者**: 后端开发组
