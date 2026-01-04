# POS Invoice 外卖订单字段扩展 任务分解

> 本文档定义 POS Invoice 外卖订单字段扩展 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 6（BMP 模块）  
**待 Main 模块实现**: 3  
**进行中**: -  
**完成率**: 40%（BMP 模块部分）

---

## Phase 1: ERPNext 自定义字段创建

- [x] 1.1 创建 `01_pos_invoice_takeout_order_no.json` 迁移文件

  - File: `ttpos-bmp/app/ttpos-erp/manifest/erp-migrate/v2.12/01_custom_field/01_pos_invoice_takeout_order_no.json`
  - Purpose: 定义 POS Invoice 的外卖订单号自定义字段
  - Requirements: 1.1
  - Leverage: 现有迁移文件: `ttpos-bmp/app/ttpos-erp/manifest/erp-migrate/v2.12/01_custom_field/01_custom_payment_id.json`
  - Prompt: Role: ERPNext Developer | Task: 创建 POS Invoice 的 custom_takeout_order_no 字段迁移文件 | Context: 字段类型 Data，字段标签 "Takeout Order No"，插入位置 custom_pos_opening_entry 之后 | Restrictions: 遵循 ERPNext 自定义字段规范，参考 01_custom_payment_id.json 格式 | Success: 迁移文件创建成功，字段定义正确

- [x] 1.2 创建 `02_pos_invoice_takeout_provider.json` 迁移文件

  - File: `ttpos-bmp/app/ttpos-erp/manifest/erp-migrate/v2.12/01_custom_field/02_pos_invoice_takeout_provider.json`
  - Purpose: 定义 POS Invoice 的外卖平台提供商自定义字段
  - Requirements: 1.2
  - Leverage: Task 1.1 的迁移文件
  - Prompt: Role: ERPNext Developer | Task: 创建 POS Invoice 的 custom_takeout_provider 字段迁移文件 | Context: 字段类型 Data，字段标签 "Takeout Provider"，插入位置 custom_takeout_order_no 之后 | Restrictions: 遵循 ERPNext 自定义字段规范 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.3 验证 ERPNext 字段创建

  - File: -
  - Purpose: 在测试环境验证字段创建成功
  - Requirements: 1.1, 1.2
  - Leverage: Task 1.1, 1.2 的迁移文件
  - Command: 在 ERPNext 中执行迁移脚本或手动创建字段
  - Success: 字段在 ERPNext POS Invoice 中可见且可编辑

---

## Phase 2: Protobuf 和 DTO 更新

- [x] 2.1 更新 Protobuf 定义

  - File: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
  - Purpose: 在 SavePosInvoiceReq 中增加外卖订单字段
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: 现有 Protobuf 定义: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
  - Prompt: Role: gRPC Developer | Task: 在 SavePosInvoiceReq 中增加 takeout_order_no 和 takeout_provider 可选字段 | Context: 字段编号从 17 开始，字段类型 optional string，添加中文注释 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc，字段必须为可选 | Success: Protobuf 定义更新成功，字段编号正确

- [x] 2.2 重新生成 protobuf Go 代码

  - File: -
  - Purpose: 生成更新后的 Go 代码
  - Requirements: 2.4
  - Leverage: Task 2.1 的 Protobuf 定义
  - Command: `cd ttpos-bmp/app/ttpos-erp && gf gen pb`
  - Success: Go 代码生成成功，新字段可用

- [x] 2.3 更新 POS Invoice DTO 结构

  - File: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/selling.go`
  - Purpose: 在 POSInvoice 结构体中增加外卖订单字段
  - Requirements: 3.1, 3.2, 3.3
  - Leverage: 现有 DTO: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/selling.go`
  - Prompt: Role: Go Developer | Task: 在 POSInvoice 结构体中增加 CustomTakeoutOrderNo 和 CustomTakeoutProvider 字段 | Context: 字段位置放在 CustomPosOpeningEntry 之后，使用 omitempty 标签 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: DTO 结构更新成功，字段定义正确

- [ ] 2.4 更新 Main 模块 Request DTO（需在 Main 模块中单独实现）

  - File: `main/app/dto/req/erpnext.go`
  - Purpose: 在 SavePosInvoiceReq 中增加外卖订单字段
  - Requirements: 5.1, 5.2, 5.3, 5.4
  - Leverage: 现有 DTO: `main/app/dto/req/erpnext.go`
  - Prompt: Role: Go Developer | Task: 在 SavePosInvoiceReq 中增加 TakeoutOrderNo 和 TakeoutProvider 字段 | Context: 字段为可选，使用 form 和 json 标签 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，字段定义正确

---

## Phase 3: 业务逻辑实现

- [x] 3.1 在 buildPosInvoice 中增加字段赋值逻辑

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 支持将请求参数中的外卖订单字段传递到 DTO
  - Requirements: 4.1, 4.2, 4.3
  - Leverage: 现有方法: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go` 的 buildPosInvoice 方法
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 在 buildPosInvoice 方法中增加外卖订单字段赋值逻辑 | Context: 检查 req.TakeoutOrderNo 和 req.TakeoutProvider 是否为空，非空时赋值给 posInvoice 对应字段 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，确保空值检查 | Success: 字段赋值逻辑实现正确，空值处理正确

- [ ] 3.2 在 Main 模块 SavePosInvoice 中获取外卖订单信息（需在 Main 模块中单独实现）

  - File: `main/app/service/order.go`
  - Purpose: 从 MemberSaleOrder 获取外卖订单信息并传递给 SavePosInvoice
  - Requirements: 5.1, 5.2, 5.3, 5.4
  - Leverage: 现有方法: `main/app/service/order.go` 的 SavePosInvoice 方法
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 SavePosInvoice 方法中检查订单是否为外卖订单，并获取 RelatedOrderNo 和 RelatedOrderType | Context: 检查 saleBill.OrderSourceUuid > 0，查询 MemberSaleOrder，获取 RelatedOrderNo 和 RelatedOrderType | Restrictions: 遵循 .cursor/rules/go-main.mdc，错误处理要记录日志但不中断流程 | Success: 外卖订单信息获取正确，非外卖订单不设置字段

- [ ] 3.3 在 RPC Service 中传递字段到 BMP 模块（需在 Main 模块中单独实现）

  - File: `main/app/service/rpc/erp/selling.go`
  - Purpose: 将 Main 模块的字段值传递到 BMP 模块的 Protobuf 请求
  - Requirements: 5.4
  - Leverage: 现有方法: `main/app/service/rpc/erp/selling.go` 的 SavePosInvoice 方法
  - Prompt: Role: Go Developer with gRPC expertise | Task: 在 SavePosInvoice RPC 方法中将 TakeoutOrderNo 和 TakeoutProvider 传递到 BMP 模块 | Context: 检查 savePosInvoiceReq 中的字段值，非空时设置到 params 中 | Restrictions: 遵循 .cursor/rules/go-main.mdc，注意 optional 字段的处理 | Success: 字段传递正确，空值不设置

---

## Phase 4: 测试和验证

- [ ] 4.1 编写 Logic 层单元测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`
  - Purpose: 测试 buildPosInvoice 方法中的字段赋值逻辑
  - Requirements: 4.1, 4.2, 4.3
  - Leverage: 现有测试: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 buildPosInvoice 方法编写单元测试，覆盖外卖订单字段赋值场景 | Context: 测试有字段值、无字段值、空字符串等场景 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 4.2 编写集成测试

  - File: `test/integration/erp_pos_invoice_takeout_test.go` (如需要)
  - Purpose: 测试端到端流程：从订单支付到 ERPNext 保存
  - Requirements: 5.1, 5.2, 5.3, 5.4
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试，验证外卖订单字段正确传递和保存 | Context: 测试外卖订单支付完成流程，验证字段值在 ERPNext 中正确保存 | Restrictions: 测试真实用户场景 | Success: 集成测试通过，字段值正确

- [ ] 4.3 向后兼容性测试

  - File: -
  - Purpose: 验证不传新字段时现有功能正常
  - Requirements: 2.2, 4.3
  - Leverage: 现有测试用例
  - Command: 运行现有测试套件，确保不传新字段时功能正常
  - Success: 所有现有测试通过，向后兼容性验证通过

- [ ] 4.4 手动测试验证

  - File: -
  - Purpose: 在 ERPNext 中手动验证字段值和显示
  - Requirements: 1.1, 1.2, 5.1, 5.2
  - Leverage: ERPNext 测试环境
  - Steps:
    1. 创建外卖订单并完成支付
    2. 在 ERPNext 中查看对应的 POS Invoice
    3. 验证 custom_takeout_order_no 和 custom_takeout_provider 字段值正确
  - Success: 字段在 ERPNext 中可见且值正确

---

## Phase 5: 文档和代码审查

- [ ] 5.1 代码审查

  - File: 所有修改的文件
  - Purpose: 确保代码质量和规范遵循
  - Requirements: 所有功能需求
  - Leverage: `.cursor/rules/go-rules.mdc`, `.cursor/rules/proto-rules.mdc`
  - Checklist:
    - [ ] 代码格式正确（go fmt）
    - [ ] 代码检查通过（go vet）
    - [ ] 遵循项目规范
    - [ ] 注释完整
  - Success: 代码审查通过

- [ ] 5.2 更新相关文档

  - File: `docs/shared/api/erp_api.md` (如需要)
  - Purpose: 更新 API 文档（如有变更）
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Success: 文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] Protobuf 代码已重新生成
- [ ] 测试覆盖率达标
  - Logic 层: ≥ 70%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
- [ ] ERPNext 字段创建成功
- [ ] 外卖订单字段正确传递和保存

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] 迁移文档已更新
- [ ] CHANGELOG.md 已更新（如需要）

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-erp-pos-invoice-takeout-fields/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-erp-pos-invoice-takeout-fields/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-erp-pos-invoice-takeout-fields/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-erp-pos-invoice-takeout-fields/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-erp-pos-invoice-takeout-fields/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Go BMP 模块开发

```
Role: Go Developer specializing in GoFrame and gRPC

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc, ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 禁止修改 dao/entity/do/ 目录（自动生成）
- Protobuf 文件修改后需重新生成 Go 代码
- Logic 层实现业务逻辑
- 使用 erp.ApiResponse 包装 gRPC 响应
- 字段必须为可选（optional），确保向后兼容

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70% (Logic 层)
```

### Go Main 模块开发

```
Role: Go Developer specializing in Gin framework

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-main.mdc

Restrictions:
- Service 只依赖其他 Service 接口
- Repository 只持有 db 实例
- 不使用 panic，返回 error
- 使用 errors.WithMessage 包装错误

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70% (Service)
```

### Protobuf 开发

```
Role: gRPC Developer with Protobuf expertise

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 字段命名使用 snake_case
- 字段编号按顺序递增
- 字段必须为可选（optional）
- 请求消息以 Req 结尾，响应消息以 Resp 结尾

Success Criteria:
- {成功标准1}
- Protobuf 定义正确
- 重新生成 Go 代码成功
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-26  
**维护者**: rikugun

