# 支付方式系统标识字段 任务分解

> 本文档定义支付方式系统标识字段功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 0.5-1 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 9  
**已完成**: 3  
**进行中**: -  
**完成率**: 33%

---

## Phase 1: Protobuf 定义和代码生成

- [x] 1.1 修改 Protobuf 定义，新增 added_by 字段

  - File: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
  - Purpose: 在 SaveModeOfPaymentReq 中新增 added_by 字段，标识支付方式创建来源
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 现有 Protobuf 定义: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto` (Line 189-204)
  - Prompt:
    ```
    Role: gRPC Developer specializing in Protobuf

    Task: 在 SaveModeOfPaymentReq 消息中新增 optional string added_by 字段

    Context:
    - File: ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto
    - Location: Line 189-204 (SaveModeOfPaymentReq 定义)
    - Requirements: 1.1 - 在 SaveModeOfPaymentReq 中新增 optional string added_by = 8 字段

    Restrictions:
    - 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc
    - 字段编号为 8（当前最后一个字段是 7）
    - 必须使用 optional 修饰符
    - 字段注释使用中文，说明用途和取值规范
    - 注释格式：// 创建来源标识，"sys" 表示系统创建

    Success Criteria:
    - added_by 字段已添加到 SaveModeOfPaymentReq
    - 字段编号为 8
    - 使用 optional 修饰符
    - 注释清晰完整
    ```

- [x] 1.2 生成 Protobuf Go 代码

  - File: -
  - Purpose: 执行 gf gen pb 命令，生成对应的 Go 代码
  - Requirements: 1.4
  - Leverage: Task 1.1 的 Protobuf 定义
  - Command:
    ```bash
    cd ttpos-bmp/app/ttpos-erp
    gf gen pb
    ```
  - Success: 代码生成成功，`api/selling/selling.pb.go` 已更新

---

## Phase 2: 常量定义

- [x] 2.1 创建常量定义文件

  - File: `ttpos-bmp/app/ttpos-erp/internal/consts/payment.go`
  - Purpose: 定义支付方式创建来源和序号相关常量
  - Requirements: 2.2
  - Leverage: 现有常量定义: `ttpos-bmp/app/ttpos-erp/internal/consts/`
  - Prompt:
    ```
    Role: Go Developer specializing in Constants Definition

    Task: 创建 payment.go 常量定义文件，定义支付方式相关常量

    Context:
    - File: ttpos-bmp/app/ttpos-erp/internal/consts/payment.go (新建)
    - Requirements: 2.2 - 使用常量定义 "sys" 标识

    Constants to Define:
    1. PaymentAddedBySystem = "sys"  // 系统创建的支付方式标识
    2. PaymentSeqSystem = 0          // 系统支付方式固定序号

    Restrictions:
    - 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    - 常量名使用大驼峰命名法（PascalCase）
    - 添加完整的中文注释
    - 包名为 consts

    Success Criteria:
    - payment.go 文件已创建
    - 两个常量定义正确
    - 注释清晰完整
    ```

---

## Phase 3: 业务逻辑实现

- [x] 3.1 修改 createModeOfPayment 方法，增加条件判断逻辑

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 根据 added_by 字段值决定使用固定序号还是自动递增序号
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5
  - Leverage: 
    - 现有方法: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go` (Line 1363-1439)
    - Task 2.1 的常量定义
  - Prompt:
    ```
    Role: Go Developer specializing in Business Logic (GoFrame)

    Task: 修改 createModeOfPayment 方法，增加 added_by 字段的条件判断逻辑

    Context:
    - File: ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go
    - Method: createModeOfPayment (Line 1363-1439)
    - Requirements: 2.1-2.5 - 序号生成逻辑调整
    - Constants: consts.PaymentAddedBySystem, consts.PaymentSeqSystem

    Modification Logic:
    在调用 nextModeOfPaymentSeq 之前（约 Line 1380），增加条件判断：
    ```go
    var nextSeq int
    if req.AddedBy != nil && strings.TrimSpace(*req.AddedBy) == consts.PaymentAddedBySystem {
        // 系统创建：使用固定序号
        nextSeq = consts.PaymentSeqSystem
        g.Log().Infof(ctx, "[createModeOfPayment] 系统创建支付方式，使用固定序号 %04d, company=%s, prefix=%s",
            nextSeq, req.CompanyAbbr, prefix)
    } else {
        // 用户创建：自动递增序号
        nextSeq, err = s.nextModeOfPaymentSeq(ctx, prefix, companyName)
        if err != nil {
            return nil, err
        }
        g.Log().Debugf(ctx, "[createModeOfPayment] 用户创建支付方式，使用递增序号 %04d, company=%s, prefix=%s",
            nextSeq, req.CompanyAbbr, prefix)
    }
    ```

    Restrictions:
    - 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    - 使用常量 consts.PaymentAddedBySystem 而不是硬编码 "sys"
    - 日志记录使用中文
    - 系统创建使用 Infof 级别，用户创建使用 Debugf 级别
    - 不破坏现有逻辑

    Success Criteria:
    - 条件判断逻辑已添加
    - 使用常量而不是硬编码字符串
    - 日志记录正确
    - 现有逻辑不受影响
    ```

- [ ] 3.2 可选：增加序号冲突检测逻辑

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 检查序号 0000 是否已被占用，避免冲突
  - Requirements: 4.1, 4.2, 4.3, 4.4
  - Leverage: 
    - `service.Doctype().Count` 方法
    - Task 3.1 的修改
  - Prompt:
    ```
    Role: Go Developer specializing in Error Handling

    Task: 在 createModeOfPayment 方法中增加序号冲突检测

    Context:
    - File: ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go
    - Method: createModeOfPayment
    - Requirements: 4.1-4.4 - 序号冲突检测（可选增强）
    - Location: 在 nextSeq 赋值后，name 生成前

    Detection Logic:
    ```go
    if nextSeq == consts.PaymentSeqSystem {
        // 检查序号 0000 是否已存在
        existingName := fmt.Sprintf("%s%04d - %s", prefix, nextSeq, req.CompanyAbbr)
        count, err := service.Doctype().Count(ctx, &erp.ErpReq{
            DocType: erp.DocTypeModeOfPayment,
        }, &erp.RequestParams{
            Filters: [][]string{{"name", "=", existingName}},
        })
        if err != nil {
            return nil, gerror.Wrapf(err, "检查支付方式序号冲突失败")
        }
        if count > 0 {
            g.Log().Errorf(ctx, "系统支付方式序号 0000 已被占用: %s", existingName)
            return nil, gerror.Newf("系统支付方式序号 0000 已被占用: %s", existingName)
        }
    }
    ```

    Restrictions:
    - 仅当 nextSeq == consts.PaymentSeqSystem 时执行检查
    - 使用 gerror.Wrapf 和 gerror.Newf 包装错误
    - 错误信息使用中文
    - 记录 Error 级别日志

    Success Criteria:
    - 冲突检测逻辑已添加
    - 错误处理正确
    - 日志记录完整
    ```

---

## Phase 4: 测试

- [ ] 4.1 编写单元测试 - 系统创建场景

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`
  - Purpose: 测试 added_by = "sys" 时使用固定序号 0000
  - Requirements: 1.2, 2.2, 3.3
  - Leverage: 现有测试文件（如存在）
  - Prompt:
    ```
    Role: QA Engineer with Go testing expertise (GoFrame)

    Task: 编写单元测试，验证 added_by = "sys" 时使用固定序号 0000

    Context:
    - Test File: ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go
    - Target Method: createModeOfPayment
    - Requirements: 测试系统创建支付方式场景

    Test Case:
    - Name: Test_createModeOfPayment_SystemCreated
    - Input: added_by = "sys"
    - Expected: nextSeq = 0, name contains "0000"

    Restrictions:
    - 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    - 使用 GoFrame 测试框架
    - 测试必须可独立运行

    Success Criteria:
    - 测试用例已添加
    - 测试通过
    - 覆盖系统创建场景
    ```

- [ ] 4.2 编写单元测试 - 用户创建场景（added_by 为空）

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`
  - Purpose: 测试 added_by 未传入时使用自动递增序号
  - Requirements: 1.4, 3.1
  - Leverage: Task 4.1 的测试框架
  - Test Case:
    - Name: Test_createModeOfPayment_UserCreated_Empty
    - Input: added_by = nil
    - Expected: 调用 nextModeOfPaymentSeq，使用自动递增序号

- [ ] 4.3 编写单元测试 - 用户创建场景（added_by 为其他值）

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`
  - Purpose: 测试 added_by 为非 "sys" 值时使用自动递增序号
  - Requirements: 1.3, 3.4
  - Leverage: Task 4.1 的测试框架
  - Test Case:
    - Name: Test_createModeOfPayment_UserCreated_OtherValue
    - Input: added_by = "user"
    - Expected: 调用 nextModeOfPaymentSeq，使用自动递增序号

- [ ] 4.4 可选：编写单元测试 - 序号冲突场景

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`
  - Purpose: 测试序号 0000 已存在时返回错误
  - Requirements: 4.1, 4.3
  - Leverage: Task 4.1 的测试框架, Task 3.2 的冲突检测逻辑
  - Test Case:
    - Name: Test_createModeOfPayment_SeqConflict
    - Setup: 创建一个序号为 0000 的支付方式
    - Input: added_by = "sys"
    - Expected: 返回错误，包含 "序号 0000 已被占用"

- [ ] 4.5 手动测试验证

  - File: -
  - Purpose: 使用 gRPC 客户端验证完整功能
  - Requirements: 所有功能需求
  - Test Steps:
    1. 启动 ttpos-erp 服务
    2. 调用 SaveModeOfPayment，added_by = "sys"
    3. 验证创建的支付方式名称包含 "0000"
    4. 调用 SaveModeOfPayment，不传 added_by
    5. 验证创建的支付方式名称包含自动递增序号（如 "0001"）
  - Success: 两种场景均正常工作

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
  ```bash
  cd ttpos-bmp/app/ttpos-erp
  go fmt ./...
  go vet ./...
  ```
- [ ] 测试覆盖率 ≥ 80%
  ```bash
  cd ttpos-bmp/app/ttpos-erp
  go test -cover ./internal/logic/selling/
  ```
- [ ] 所有测试通过
  ```bash
  cd ttpos-bmp/app/ttpos-erp
  go test ./internal/logic/selling/ -v
  ```

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
  - Requirement 1: 新增 added_by 字段 ✅
  - Requirement 2: 序号生成逻辑调整 ✅
  - Requirement 3: 向后兼容性保证 ✅
  - Requirement 4: 序号冲突检测（可选）✅
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] Protobuf 注释完整
- [ ] 代码注释清晰（中文）
- [ ] CHANGELOG.md 已更新（如需要）

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-ttpos-erp.mdc`
- [ ] 使用常量而不是硬编码字符串
- [ ] 日志记录使用中文

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-erp-payment-added-by/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-erp-payment-added-by/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-erp-payment-added-by/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-erp-payment-added-by/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-erp-payment-added-by/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 按 Phase 顺序执行，每个 Phase 内按顺序执行
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 查看 design.md 中的实现方案
4. **查看复用**: 检查 Leverage 中的可复用代码
5. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
6. **实现代码**: 按照规范实现功能
7. **运行检查**: 
   ```bash
   cd ttpos-bmp/app/ttpos-erp
   go fmt ./...
   go vet ./...
   go test ./internal/logic/selling/ -v
   ```
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Go 后端开发（GoFrame）

```
Role: Go Developer specializing in GoFrame Business Logic

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc, ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 使用 GoFrame 2.x 框架
- 日志记录使用中文
- 错误信息使用中文
- 使用 gerror 包处理错误
- 不使用 panic
- 禁止修改 dao/entity/do/ 目录
- 使用常量而不是硬编码字符串

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 80%
```

### 测试工程师（GoFrame）

```
Role: QA Engineer with Go testing expertise (GoFrame)

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 80%

Test Cases Required:
- 正常场景测试（added_by = "sys"）
- 异常场景测试（added_by 为空）
- 边界条件测试（added_by 为其他值）
- 冲突场景测试（序号 0000 已存在）

Restrictions:
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
- 使用 GoFrame 测试框架
- 测试必须可独立运行

Success Criteria:
- 测试覆盖率 ≥ 80%
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2025-12/2025-12-30.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-30  
**维护者**: rikugun

