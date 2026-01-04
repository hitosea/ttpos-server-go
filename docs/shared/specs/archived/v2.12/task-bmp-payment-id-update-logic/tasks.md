# 支付方式更新逻辑优化 任务分解

> 本文档定义支付方式更新逻辑优化的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 6  
**已完成**: 3  
**进行中**: 任务 1.2 编写单元测试  
**完成率**: 50%

---

## Phase 1: Controller 层调整

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 修改 Controller 层验证逻辑

  - **File**: `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go`
  - **Purpose**: 支持通过 `PaymentId` 识别更新操作
  - **Requirements**: 1.1, 2.1
  - **Leverage**: 现有 `validateSaveModeOfPaymentReq` 方法
  - **Prompt**:
    ```
    Role: Go Developer specializing in GoFrame 2.x
    
    Task: 修改 validateSaveModeOfPaymentReq 方法，支持通过 PaymentId 识别更新操作
    
    Context:
    - Current file: ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go
    - Method: validateSaveModeOfPaymentReq
    - Requirements: 1.1, 2.1 - 支持通过 payment_id 识别和更新
    - Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    
    Changes Required:
    1. 修改 isUpdate 判断条件，新增 PaymentId 不为空的判断：
       isUpdate := (req.Name != nil && strings.TrimSpace(*req.Name) != "") || 
                   (req.PaymentId != "" && strings.TrimSpace(req.PaymentId) != "")
    2. 保持其他验证逻辑不变
    3. 更新注释说明新增的判断条件
    
    Restrictions:
    - 遵循 GoFrame 2.x 规范
    - 使用 gerror 包进行错误处理
    - 保持向后兼容，不破坏现有调用方
    
    Success Criteria:
    - isUpdate 判断支持 PaymentId
    - 现有 Name 判断逻辑不受影响
    - 代码通过 go fmt
    ```
  - **Story Points**: 0.5

- [ ] 1.2 编写 Controller 层单元测试

  - **File**: `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling_test.go`
  - **Purpose**: 确保 Controller 层验证逻辑正确
  - **Requirements**: 1.1, 2.1
  - **Leverage**: 现有测试文件 `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`
  - **Prompt**:
    ```
    Role: QA Engineer with Go testing expertise
    
    Task: 为 validateSaveModeOfPaymentReq 编写单元测试
    
    Context:
    - Target file: ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go
    - Test file: ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling_test.go
    - Coverage target: ≥ 80%
    
    Test Cases Required:
    1. 测试 Name 不为空时识别为更新操作
    2. 测试 PaymentId 不为空时识别为更新操作
    3. 测试 Name 和 PaymentId 都不为空时识别为更新操作
    4. 测试创建操作时 pay_type 必填
    5. 测试更新操作时 pay_type 非必填
    
    Restrictions:
    - 遵循 GoFrame 2.x 测试规范
    - 必须包含边界情况测试
    
    Success Criteria:
    - 测试覆盖率 ≥ 80%
    - 所有测试通过
    - 边界情况已覆盖
    ```
  - **Story Points**: 1

---

## Phase 2: Logic 层调整

- [x] 2.1 修改 Logic 层路由逻辑

  - **File**: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - **Purpose**: 支持通过 `PaymentId` 路由到更新方法
  - **Requirements**: 1.1, 2.1
  - **Leverage**: 现有 `SaveModeOfPayment` 方法
  - **Prompt**:
    ```
    Role: Go Developer specializing in GoFrame 2.x business logic
    
    Task: 修改 SaveModeOfPayment 方法，支持通过 PaymentId 路由到更新方法
    
    Context:
    - Current file: ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go
    - Method: SaveModeOfPayment
    - Requirements: 1.1, 2.1 - 支持通过 payment_id 识别和更新
    - Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    
    Changes Required:
    1. 修改路由判断条件，新增 PaymentId 不为空的判断：
       if (req.Name != nil && strings.TrimSpace(*req.Name) != "") || 
          (req.PaymentId != "" && strings.TrimSpace(req.PaymentId) != "") {
           return s.updateModeOfPayment(ctx, req)
       }
    2. 保持创建逻辑不变
    3. 更新注释说明新增的判断条件
    
    Restrictions:
    - 遵循 GoFrame 2.x 规范
    - 保持向后兼容
    - 不破坏现有调用方
    
    Success Criteria:
    - 路由逻辑支持 PaymentId
    - 现有 Name 路由逻辑不受影响
    - 代码通过 go fmt
    ```
  - **Story Points**: 0.5

- [x] 2.2 优化 Logic 层查询逻辑

  - **File**: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - **Purpose**: 优先使用 `PaymentId` 查询，统一使用 List 接口
  - **Requirements**: 1.1, 2.1, 2.2
  - **Leverage**: 现有 `updateModeOfPayment` 方法，`service.Document().List()` 方法
  - **Prompt**:
    ```
    Role: Go Developer specializing in GoFrame 2.x and ERPNext integration
    
    Task: 重构 updateModeOfPayment 方法，优先使用 PaymentId 查询，统一使用 List 接口
    
    Context:
    - Current file: ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go
    - Method: updateModeOfPayment
    - Requirements: 1.1, 2.1, 2.2 - 优先使用 PaymentId 查询，统一使用 List 接口
    - Leverage: service.Document().List() 方法
    - Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    
    Changes Required:
    1. 查询优先级调整：
       - 优先使用 PaymentId 查询（业务主键）
       - 其次使用 Name 查询（ERP 主键）
       - 至少提供一个
    
    2. 统一查询接口：
       - 使用 service.Document().List() 替代 Get()
       - 使用 Filters 参数过滤
       - 使用 Limit: 1 减少数据传输
    
    3. 增强日志记录：
       - 记录查询键值（payment_id 或 name）
       - 记录查询方式和结果
       - 记录更新操作和数据
    
    4. 保持其他逻辑不变：
       - 权限校验逻辑
       - 更新逻辑
       - 返回数据格式
    
    Code Template:
    ```go
    func (s *sSelling) updateModeOfPayment(ctx context.Context, req *selling.SaveModeOfPaymentReq) (*selling.SaveModeOfPaymentResp, error) {
        var resp *gjson.Json
        var err error
        var name string
        var queryKey string

        // 构建查询过滤器
        var filters [][]string
        
        // 1. 优先使用 PaymentId 查询（业务主键）
        if req.PaymentId != "" && strings.TrimSpace(req.PaymentId) != "" {
            paymentId := strings.TrimSpace(req.PaymentId)
            queryKey = fmt.Sprintf("payment_id=%s", paymentId)
            filters = [][]string{{"custom_payment_id", "=", paymentId}}
            g.Log().Infof(ctx, "[updateModeOfPayment] 通过 payment_id 查询支付方式: %s", queryKey)
        } else if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
            // 2. 使用 Name 查询（ERP 主键）
            name = strings.TrimSpace(*req.Name)
            queryKey = fmt.Sprintf("name=%s", name)
            filters = [][]string{{"name", "=", name}}
            g.Log().Infof(ctx, "[updateModeOfPayment] 通过 name 查询支付方式: %s", queryKey)
        } else {
            return nil, gerror.New("name 或 payment_id 至少提供一个")
        }

        // 3. 统一使用 List 接口查询
        resp, err = service.Document().List(ctx, &erp.ErpReq{
            DocType: erp.DocTypeModeOfPayment,
        }, &erp.RequestParams{
            Fields:  []string{"name", "custom_company", "custom_branch", "enabled", "custom_payment_id"},
            Filters: filters,
            Limit:   1,
        })
        if err != nil {
            g.Log().Errorf(ctx, "[updateModeOfPayment] 查询支付方式失败: %s, err=%v", queryKey, err)
            return nil, gerror.Wrapf(err, "查询支付方式失败")
        }

        // 4. 检查查询结果
        dataArray := resp.GetJsons("data")
        if len(dataArray) == 0 {
            g.Log().Warningf(ctx, "[updateModeOfPayment] 支付方式不存在: %s", queryKey)
            return nil, gerror.Newf("支付方式不存在: %s", queryKey)
        }

        // 5-10. 保持原有逻辑：获取信息、权限校验、构建更新数据、执行更新、记录日志、返回结果
        // (参考 design.md 中的完整代码)
    }
    ```
    
    Restrictions:
    - 遵循 GoFrame 2.x 规范
    - 使用 gerror 包进行错误处理
    - 保持向后兼容
    - 不破坏现有调用方
    
    Success Criteria:
    - 查询优先级正确（PaymentId > Name）
    - 统一使用 List 接口
    - 日志记录完善
    - 代码通过 go fmt
    - 原有逻辑不受影响
    ```
  - **Story Points**: 2

- [ ] 2.3 编写 Logic 层单元测试

  - **File**: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`
  - **Purpose**: 确保 Logic 层业务逻辑正确
  - **Requirements**: 1.1, 2.1, 2.2
  - **Leverage**: 现有测试文件
  - **Prompt**:
    ```
    Role: QA Engineer with Go testing expertise
    
    Task: 为 SaveModeOfPayment 和 updateModeOfPayment 编写单元测试
    
    Context:
    - Target file: ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go
    - Test file: ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go
    - Coverage target: ≥ 80%
    
    Test Cases Required:
    
    1. SaveModeOfPayment 路由逻辑测试：
       - 测试 Name 不为空时路由到更新方法
       - 测试 PaymentId 不为空时路由到更新方法
       - 测试 Name 和 PaymentId 都为空时路由到创建方法
    
    2. updateModeOfPayment 查询逻辑测试：
       - 测试优先使用 PaymentId 查询
       - 测试使用 Name 查询
       - 测试 PaymentId 和 Name 都提供时优先使用 PaymentId
       - 测试查询结果为空时返回错误
       - 测试权限校验失败时返回错误
       - 测试查询成功且权限通过时执行更新
    
    3. 边界情况测试：
       - 测试 PaymentId 包含空格时的 Trim 逻辑
       - 测试 Name 包含空格时的 Trim 逻辑
    
    Restrictions:
    - 遵循 GoFrame 2.x 测试规范
    - 使用 Mock 对象模拟 service.Document() 和 service.Company()
    - 必须包含边界情况测试
    
    Success Criteria:
    - 测试覆盖率 ≥ 80%
    - 所有测试通过
    - 边界情况已覆盖
    ```
  - **Story Points**: 2

---

## Phase 3: 测试和验证

- [ ] 3.1 集成测试

  - **File**: -
  - **Purpose**: 测试端到端功能
  - **Requirements**: 所有功能需求
  - **Leverage**: 现有集成测试脚本
  - **Prompt**:
    ```
    Role: QA Automation Engineer
    
    Task: 实现端到端集成测试
    
    Test Flow:
    1. 创建测试数据：创建测试支付方式
    2. 使用 PaymentId 更新：调用接口，验证更新成功
    3. 使用 Name 更新：调用接口，验证更新成功
    4. 越权测试：尝试更新其他公司的支付方式，验证被拒绝
    5. 不存在测试：使用不存在的 PaymentId，验证返回错误
    
    Restrictions:
    - 测试真实用户场景
    - 清理测试数据
    
    Success Criteria:
    - 集成测试通过
    - 覆盖所有用户场景
    ```
  - **Story Points**: 1.5

- [ ] 3.2 性能测试和文档更新

  - **File**: -
  - **Purpose**: 确保性能达标，更新相关文档
  - **Requirements**: 性能要求
  - **Leverage**: 性能测试工具
  - **Tasks**:
    1. 性能测试：
       - 查询响应时间 < 100ms
       - 更新响应时间 < 200ms
    2. 文档更新：
       - 更新 CHANGELOG.md
       - 更新 API 文档（如需要）
  - **Story Points**: 0.5

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Logic: ≥ 80%
  - Controller: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] CHANGELOG.md 已更新
- [ ] API 文档已更新（如需要）

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `.cursor/rules/version.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/task-bmp-payment-id-update-logic/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/task-bmp-payment-id-update-logic/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/task-bmp-payment-id-update-logic/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/task-bmp-payment-id-update-logic/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/task-bmp-payment-id-update-logic/tasks.md)" | bc
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

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2025-12/2025-12-24.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-24  
**维护者**: rikugun

