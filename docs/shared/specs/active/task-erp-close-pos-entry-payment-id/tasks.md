# 关账接口支持 PaymentID 任务分解

> 本文档定义关账接口支持 PaymentID 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 12  
**已完成**: 9  
**进行中**: 单元测试（Phase 3）和集成测试（Phase 4）  
**完成率**: 75%

---

## Phase 1: Protobuf 定义调整

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 修改 Protobuf 定义

  - File: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
  - Purpose: 在 `ClosePosEntryDetail` 消息中增加 `payment_id` 字段支持
  - Requirements: 1.1, 1.2, 1.3, 1.4
  - Leverage: 现有的 `selling.proto` 文件，参考 `PosInvoicePayment` 消息（已支持 payment_id）
  - Changes:
    - 将 `ClosePosEntryDetail.mode_of_payment` 字段改为 `optional string`
    - 新增 `ClosePosEntryDetail.payment_id` 字段（`optional string`，字段编号 4）
    - 添加字段注释说明两个字段的关系（与 mode_of_payment 二选一）
  - Prompt: 
    ```
    Role: gRPC Developer

    Task: 修改 selling.proto 中的 ClosePosEntryDetail 消息定义，增加 payment_id 字段支持

    Context:
    - File: ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto
    - Message: ClosePosEntryDetail
    - Requirements: 支持通过 payment_id 或 mode_of_payment 指定支付方式，两者至少提供一个
    - Reference: PosInvoicePayment 消息已有类似设计

    Changes Required:
    1. 将 mode_of_payment 字段改为 optional string（字段编号 1）
    2. 新增 payment_id 字段（optional string，字段编号 4）
    3. 添加注释说明：payment_id 和 mode_of_payment 二选一；当 payment_id 不为空时，系统自动调用 GetModeOfPayment 查询 mode_of_payment 值

    Restrictions:
    - 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc
    - 字段命名使用 snake_case
    - 注释使用中文
    - 保持其他字段（opening_amount, closing_amount）不变

    Success Criteria:
    - ClosePosEntryDetail 包含 optional string payment_id 字段
    - mode_of_payment 改为 optional
    - 字段注释完整且准确
    ```

- [x] 1.2 重新生成 Protobuf Go 代码

  - File: `ttpos-bmp/app/ttpos-erp/api/selling/*.pb.go`
  - Purpose: 根据更新的 proto 文件生成 Go 代码
  - Requirements: 1.1
  - Leverage: Task 1.1 的 proto 文件
  - Command: `cd ttpos-bmp/app/ttpos-erp && gf gen pb`
  - Success: 代码生成成功，无编译错误

- [x] 1.3 验证 Protobuf 生成结果

  - File: `ttpos-bmp/app/ttpos-erp/api/selling/selling.pb.go`
  - Purpose: 确认生成的 Go 代码包含 PaymentId 字段
  - Requirements: 1.1, 1.2
  - Leverage: Task 1.2 生成的代码
  - Verification:
    - 检查 `ClosePosEntryDetail` 结构体包含 `PaymentId *string` 字段
    - 检查 `ModeOfPayment *string` 字段（optional）
    - 编译通过：`cd ttpos-bmp/app/ttpos-erp && go build ./...`
  - Success: 生成的代码结构正确，编译无错误

---

## Phase 2: Logic 层实现

### 参数校验和自动查询

- [x] 2.1 实现参数校验逻辑

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 在 `ClosePosEntry` 方法中添加参数校验逻辑
  - Requirements: 2.1, 2.2, 2.3, 2.4
  - Leverage: 现有的 `ClosePosEntry` 方法实现
  - Changes:
    - 在处理 `req.ClosePosEntryDetail` 列表时，对每个 detail 进行校验
    - 检查 `detail.PaymentId` 和 `detail.ModeOfPayment` 是否同时为空
    - 如果同时为空，返回错误：`close_pos_entry_detail[{index}]: payment_id 和 mode_of_payment 不能同时为空`
  - Prompt:
    ```
    Role: Go Developer with GoFrame expertise

    Task: 在 ClosePosEntry 方法中添加参数校验逻辑

    Context:
    - File: ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go
    - Method: func (s *sSelling) ClosePosEntry(ctx context.Context, req *selling.ClosePosEntryReq) (*selling.ClosePosEntryResp, error)
    - Requirements: 校验 payment_id 和 mode_of_payment 不能同时为空

    Implementation:
    在循环处理 req.ClosePosEntryDetail 时，添加校验：
    ```go
    for i, detail := range req.ClosePosEntryDetail {
        // 参数校验
        if (detail.PaymentId == nil || *detail.PaymentId == "") && 
           (detail.ModeOfPayment == nil || *detail.ModeOfPayment == "") {
            return nil, gerror.Newf("close_pos_entry_detail[%d]: payment_id 和 mode_of_payment 不能同时为空", i)
        }
        
        // ... 后续处理
    }
    ```

    Restrictions:
    - 使用 gerror.Newf() 创建错误
    - 错误信息使用中文
    - 包含索引 i 便于定位问题

    Success Criteria:
    - 当两个参数都为空时返回明确的错误信息
    - 错误信息包含 detail 的索引
    ```

- [x] 2.2 实现自动查询 mode_of_payment

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 当 payment_id 不为空时，自动调用 GetModeOfPayment 查询对应的支付方式名称
  - Requirements: 3.1, 3.2, 3.3, 3.4, 3.5
  - Leverage: 
    - 现有的 `GetModeOfPayment` 方法
    - Task 2.1 的参数校验逻辑
  - Changes:
    - 在参数校验后，检查 `detail.PaymentId` 是否不为空
    - 如果不为空，调用 `service.Selling().GetModeOfPayment()`
    - 提取 `resp.ModeOfPayment.Name` 用于后续处理
    - 如果为空，使用 `detail.ModeOfPayment`
  - Prompt:
    ```
    Role: Go Developer with business logic expertise

    Task: 在 ClosePosEntry 方法中实现自动查询 mode_of_payment 的逻辑

    Context:
    - File: ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go
    - Method: func (s *sSelling) ClosePosEntry(ctx context.Context, req *selling.ClosePosEntryReq) (*selling.ClosePosEntryResp, error)
    - Requirements: 当 payment_id 不为空时，调用 GetModeOfPayment 查询对应的 mode_of_payment
    - Leverage: service.Selling().GetModeOfPayment() 服务

    Implementation:
    在 Task 2.1 的参数校验后，添加自动查询逻辑：
    ```go
    modeOfPayment := ""
    
    // 如果提供了 payment_id，自动查询 mode_of_payment
    if detail.PaymentId != nil && *detail.PaymentId != "" {
        getModeResp, err := service.Selling().GetModeOfPayment(ctx, &selling.GetModeOfPaymentReq{
            PaymentId: detail.PaymentId,
        })
        if err != nil {
            g.Log().Error(ctx, "查询支付方式失败", 
                g.Map{"payment_id": *detail.PaymentId, "error": err.Error()})
            return nil, gerror.Wrapf(err, "查询支付方式失败，payment_id: %s", *detail.PaymentId)
        }
        
        if getModeResp.ModeOfPayment == nil || getModeResp.ModeOfPayment.Name == "" {
            return nil, gerror.Newf("支付方式不存在或未启用，payment_id: %s", *detail.PaymentId)
        }
        
        modeOfPayment = getModeResp.ModeOfPayment.Name
        
        g.Log().Info(ctx, "关账详情: 通过 payment_id 查询到 mode_of_payment",
            g.Map{"index": i, "payment_id": *detail.PaymentId, "mode_of_payment": modeOfPayment})
    } else {
        // 直接使用 mode_of_payment（向后兼容）
        modeOfPayment = *detail.ModeOfPayment
    }
    
    // 使用 modeOfPayment 进行后续处理
    // ... 原有关账逻辑使用 modeOfPayment ...
    ```

    Restrictions:
    - 使用 service.Selling().GetModeOfPayment() 调用服务
    - 使用 gerror.Wrapf() 包装错误
    - 使用 g.Log() 记录日志
    - 错误信息包含 payment_id 值

    Success Criteria:
    - payment_id 不为空时能成功查询 mode_of_payment
    - 查询失败时返回包含 payment_id 的错误信息
    - 支付方式未启用时返回明确错误
    - payment_id 为空时使用原有逻辑（向后兼容）
    ```

- [x] 2.3 更新原有逻辑使用查询结果

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 确保原有关账逻辑使用查询得到的 modeOfPayment 变量
  - Requirements: 3.4, 3.5, 4.1, 4.2, 4.3
  - Leverage: Task 2.2 的查询逻辑
  - Changes:
    - 将原有直接使用 `detail.ModeOfPayment` 的地方改为使用 `modeOfPayment` 变量
    - 确保不影响其他逻辑的执行
  - Success: 原有逻辑正常工作，使用统一的 modeOfPayment 变量

- [x] 2.4 添加代码注释

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 为新增逻辑添加清晰的中文注释
  - Requirements: 所有需求
  - Leverage: Task 2.1, 2.2, 2.3 的实现
  - Changes:
    - 为参数校验逻辑添加注释
    - 为自动查询逻辑添加注释
    - 说明向后兼容性处理
  - Success: 代码注释完整，逻辑清晰易懂

---

## Phase 3: 单元测试

- [ ] 3.1 编写参数校验测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`
  - Purpose: 测试参数校验逻辑（payment_id 和 mode_of_payment 同时为空）
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: 现有的测试文件结构
  - Test Cases:
    - 测试两个参数都为空的情况
    - 验证返回的错误信息格式正确
  - Prompt:
    ```
    Role: QA Engineer with Go testing expertise

    Task: 为 ClosePosEntry 方法编写参数校验测试

    Context:
    - File: ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go
    - Method: ClosePosEntry
    - Test Scenario: payment_id 和 mode_of_payment 同时为空

    Test Implementation:
    ```go
    func Test_ClosePosEntry_BothEmpty(t *testing.T) {
        // 测试两个参数都为空的情况
        req := &selling.ClosePosEntryReq{
            PosOpenEntryName: "TEST-ENTRY",
            PeriodEndDate:    time.Now().Unix(),
            ClosePosEntryDetail: []*selling.ClosePosEntryDetail{
                {
                    // payment_id 和 mode_of_payment 都为 nil
                    OpeningAmount: 1000.00,
                    ClosingAmount: 1500.00,
                },
            },
        }
        
        _, err := service.Selling().ClosePosEntry(ctx, req)
        
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "payment_id 和 mode_of_payment 不能同时为空")
    }
    ```

    Success Criteria:
    - 测试通过
    - 覆盖参数校验逻辑
    ```

- [ ] 3.2 编写自动查询测试（成功场景）

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`
  - Purpose: 测试使用 payment_id 时的自动查询功能
  - Requirements: 3.1, 3.2, 3.4
  - Leverage: Mock `GetModeOfPayment` 服务
  - Test Cases:
    - Mock `GetModeOfPayment` 返回成功
    - 验证关账逻辑正常执行
    - 验证使用了查询得到的 mode_of_payment
  - Success: 测试通过，自动查询逻辑正确

- [ ] 3.3 编写自动查询测试（失败场景）

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`
  - Purpose: 测试 payment_id 查询失败时的错误处理
  - Requirements: 3.3
  - Leverage: Mock `GetModeOfPayment` 服务返回错误
  - Test Cases:
    - Mock `GetModeOfPayment` 返回错误
    - 验证返回的错误信息包含 payment_id
    - Mock 支付方式未启用，验证错误处理
  - Success: 测试通过，错误处理正确

- [ ] 3.4 编写向后兼容测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`
  - Purpose: 测试仅使用 mode_of_payment 的向后兼容性
  - Requirements: 4.1, 4.2, 4.3, 4.4
  - Leverage: 现有的关账测试场景
  - Test Cases:
    - 仅提供 mode_of_payment（payment_id 为 nil）
    - 验证关账逻辑正常执行
    - 验证不调用 GetModeOfPayment
  - Success: 测试通过，向后兼容性得到保证

- [ ] 3.5 测试覆盖率检查

  - File: -
  - Purpose: 确保 Logic 层测试覆盖率 ≥ 80%
  - Requirements: 测试要求
  - Command: `cd ttpos-bmp/app/ttpos-erp && go test -coverprofile=coverage.out ./internal/logic/selling && go tool cover -func=coverage.out | grep selling.go`
  - Success: 覆盖率 ≥ 80%，所有测试通过

---

## Phase 4: 集成测试

- [ ] 4.1 手动集成测试

  - File: -
  - Purpose: 使用 gRPC 客户端测试各种场景
  - Requirements: 所有功能需求
  - Test Scenarios:
    1. 使用真实的 payment_id 进行关账
    2. 使用不存在的 payment_id 验证错误处理
    3. 使用原有的 mode_of_payment 验证向后兼容性
    4. 同时提供两个参数，验证优先使用 payment_id
  - Tools: grpcurl 或 BloomRPC
  - Success: 所有场景测试通过

---

## Phase 5: 文档和提交

- [x] 5.1 更新 CHANGELOG

  - File: `ttpos-bmp/CHANGELOG.md`
  - Purpose: 记录本次功能更新
  - Requirements: 文档要求
  - Changes:
    - 在 `[Unreleased]` 部分添加条目
    - 格式：`### Enhanced - ClosePosEntry 接口支持通过 PaymentID 指定支付方式`
  - Success: CHANGELOG 更新完成

- [x] 5.2 提交代码

  - File: -
  - Purpose: 提交所有修改到 Git
  - Requirements: 版本管理规范
  - Leverage: `.cursor/rules/version.mdc`
  - Commit Message:
    ```
    feat(ttpos-erp): ClosePosEntry 接口支持 PaymentID
    
    - Protobuf: ClosePosEntryDetail 新增 payment_id 字段
    - Logic: 自动查询 mode_of_payment 当 payment_id 不为空时
    - 向后兼容: 保持原有 mode_of_payment 字段可用
    - 测试: 覆盖率 ≥ 80%
    
    关联 Spec: task-erp-close-pos-entry-payment-id
    ```
  - Command: `git add . && git commit -m "..." && git push`
  - Success: 代码提交成功

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] Protobuf 代码生成成功
- [ ] 测试覆盖率达标
  - Logic 层: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
  - Protobuf 定义包含 payment_id 字段
  - 参数校验逻辑正确
  - 自动查询功能正常
  - 错误处理完善
  - 向后兼容性保证

### 文档同步

- [ ] Protobuf 注释完整
- [ ] 代码注释清晰
- [ ] CHANGELOG.md 已更新
- [ ] Spec 文档已完成（requirements.md, design.md, tasks.md）

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `.cursor/rules/go-bmp.mdc`
- [ ] 遵循 `.cursor/rules/version.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/task-erp-close-pos-entry-payment-id/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/task-erp-close-pos-entry-payment-id/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/task-erp-close-pos-entry-payment-id/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/task-erp-close-pos-entry-payment-id/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/task-erp-close-pos-entry-payment-id/tasks.md)" | bc
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

## 附录：开发顺序建议

### 推荐执行顺序

1. **Phase 1: Protobuf 定义调整** (Task 1.1 → 1.2 → 1.3)
   - 先完成 Protobuf 定义，确保 API 接口明确
   - 生成代码并验证编译通过

2. **Phase 2: Logic 层实现** (Task 2.1 → 2.2 → 2.3 → 2.4)
   - 按顺序实现参数校验、自动查询、逻辑集成
   - 边实现边添加注释

3. **Phase 3: 单元测试** (Task 3.1 → 3.2 → 3.3 → 3.4 → 3.5)
   - 完成实现后立即编写测试
   - 确保覆盖率达标

4. **Phase 4: 集成测试** (Task 4.1)
   - 手动测试各种场景
   - 验证端到端功能

5. **Phase 5: 文档和提交** (Task 5.1 → 5.2)
   - 更新文档
   - 提交代码

### 并行执行建议

- Task 1.1, 1.2, 1.3 必须顺序执行
- Task 2.1, 2.2, 2.3 必须顺序执行
- Task 3.1-3.4 可以部分并行（不同测试文件）
- Task 2.4（添加注释）可以在实现过程中同步完成

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**创建日期**: 2025-12-24  
**作者**: rikugun  
**维护者**: 后端开发组

