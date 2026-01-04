# ERP 支付方式 PaymentID 查询与自动解析 任务分解

> 本文档定义 ERP 支付方式 PaymentID 查询与自动解析功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 10  
**已完成**: 8  
**进行中**: -  
**完成率**: 80%

**任务清单**：
- Phase 1: Protobuf 定义更新（3 任务，3 已完成）✅
- Phase 2: Logic 层实现（3 任务，3 已完成）✅
- Phase 3: Service 接口生成（2 任务，2 已完成）✅
- Phase 4: 测试（2 任务，待执行）
- Phase 5: 文档和部署（1 任务，待执行）

---

## Phase 1: Protobuf 定义更新（1h）

### 任务说明

更新 `selling.proto` 文件，新增 `GetModeOfPayment` 查询接口，修改 `PosInvoicePayment` 消息支持 `payment_id` 字段。

---

- [x] 1.1 新增 GetModeOfPaymentReq 和 GetModeOfPaymentResp 消息

  - File: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
  - Purpose: 定义查询单个支付方式的请求和响应消息
  - Requirements: Requirement 1 (1.1, 1.2)
  - Leverage: 现有 Protobuf 定义: `selling.proto`，现有 `ModeOfPayment` 消息
  - Completed: ✅ 已完成（在 Proposal 中已定义）
  - Details:
    ```protobuf
    message GetModeOfPaymentReq {
      optional string name = 1;       // 支付方式名称（精确匹配）
      optional string payment_id = 2; // 支付方式唯一标识（PaymentID）
    }
    
    message GetModeOfPaymentResp {
      ModeOfPayment mode_of_payment = 1; // 支付方式信息
    }
    ```

- [x] 1.2 修改 PosInvoicePayment 消息，新增 payment_id 字段

  - File: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
  - Purpose: 支持在 POS 发票支付中使用 PaymentID
  - Requirements: Requirement 2 (2.1, 2.2)
  - Leverage: 现有 `PosInvoicePayment` 消息
  - Completed: ✅ 已完成（在 Proposal 中已定义）
  - Details:
    ```protobuf
    message PosInvoicePayment {
      string mode_of_payment = 1;     // 支付方式，与 payment_id 二选一
      double amount = 2;              // 金额，必填
      optional string payment_id = 3; // 支付方式唯一标识（PaymentID）
      // 注意：当 payment_id 不为空时，系统自动调用 GetModeOfPayment 查询 mode_of_payment 值
    }
    ```

- [x] 1.3 在 SellingService 中添加 GetModeOfPayment RPC 方法

  - File: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
  - Purpose: 定义 gRPC 服务接口
  - Requirements: Requirement 1 (1.3)
  - Leverage: 现有 `SellingService` 定义
  - Completed: ✅ 已完成（2025-12-23 10:27）
  - Command: `cd ttpos-bmp/app/ttpos-erp && gf gen pb`
  - Success: Protobuf 生成成功，Go 代码已更新

---

## Phase 2: Logic 层实现（2.5h）

### 任务说明

实现 `GetModeOfPayment` 查询逻辑和 `SavePosInvoice` 支付流程集成，包括 payment_id 自动解析。

---

- [x] 2.1 实现 GetModeOfPayment 查询逻辑

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 实现通过 name 或 payment_id 查询单个支付方式
  - Requirements: Requirement 1 (1.4, 1.5, 1.6)
  - Completed: ✅ 已完成（2025-12-23 10:30）
  - Success: GetModeOfPayment 方法实现完整，支持两种查询方式，包含完整的日志和错误处理

- [x] 2.2 实现 resolvePaymentIDs 辅助方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 批量解析支付列表中的 payment_id，实现缓存优化
  - Requirements: Requirement 3 (3.1, 3.2, 3.3, 3.6)
  - Completed: ✅ 已完成（2025-12-23 10:31）
  - Success: resolvePaymentIDs 方法实现完整，缓存优化生效，错误信息详细

- [x] 2.3 修改 SavePosInvoice，集成 payment_id 自动解析

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
  - Purpose: 在保存 POS 发票前自动解析 payment_id
  - Requirements: Requirement 3 (3.4, 3.5, 3.7)
  - Completed: ✅ 已完成（2025-12-23 10:31）
  - Success: SavePosInvoice 集成完成，payment_id 自动解析生效，向后兼容

---

## Phase 3: Service 接口生成（1h）

### 任务说明

使用 GoFrame 工具生成 Service 接口和 Controller 层代码。

---

- [x] 3.1 生成 Service 接口

  - File: `ttpos-bmp/app/ttpos-erp/internal/service/selling.go`（自动生成接口）
  - Purpose: 自动生成 Service 接口定义
  - Requirements: Requirement 1 (1.3)
  - Completed: ✅ 已完成（2025-12-23 10:32）
  - Command: `cd ttpos-bmp/app/ttpos-erp && gf gen service`
  - Success: Service 接口生成成功，包含 GetModeOfPayment 方法

- [x] 3.2 添加 Controller 层代码

  - File: `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/selling/selling.go`
  - Purpose: 添加 GetModeOfPayment Controller 方法
  - Requirements: Requirement 1
  - Completed: ✅ 已完成（2025-12-23 10:33）
  - Success: Controller 层 GetModeOfPayment 方法添加成功，调用 Logic 层正确

---

## Phase 4: 测试（1.5h）

### 任务说明

编写单元测试和集成测试，确保功能正确性和性能。

---

- [ ] 4.1 编写 GetModeOfPayment 单元测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`
  - Purpose: 测试 GetModeOfPayment 查询逻辑
  - Requirements: Requirement 1
  - Leverage: 现有测试文件: `selling_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 GetModeOfPayment 方法编写单元测试，覆盖率 ≥ 80% | Context: 测试用例包括：1) 通过 name 查询成功，2) 通过 payment_id 查询成功，3) 参数缺失失败，4) 支付方式不存在失败 | Restrictions: 使用 testing 框架，Mock ERPNext Service，遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过
  - Success: 单元测试通过，覆盖率达标

- [ ] 4.2 编写 SavePosInvoice 支付流程测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`
  - Purpose: 测试 payment_id 自动解析和向后兼容性
  - Requirements: Requirement 2, Requirement 3
  - Leverage: 现有测试文件: `selling_test.go`
  - Prompt: Role: QA Engineer | Task: 为 SavePosInvoice 支付流程编写测试，覆盖 payment_id 自动解析场景 | Context: 测试用例包括：1) payment_id 自动解析成功，2) mode_of_payment 直接使用成功（向后兼容），3) 同时提供两者（优先使用 payment_id），4) 都未提供失败，5) payment_id 无效失败，6) 支付方式已禁用失败，7) 批量查询缓存效果 | Restrictions: Mock ERPNext Service，验证缓存逻辑生效 | Success: 所有测试通过，向后兼容性验证通过，缓存逻辑生效
  - Success: 集成测试通过，性能测试达标

---

## Phase 5: 文档和部署（0.5h）

### 任务说明

更新文档，准备部署。

---

- [ ] 5.1 更新文档和部署

  - File: `docs/shared/specs/active/story-erp-get-mode-of-payment-by-id/`
  - Purpose: 完善文档，准备部署
  - Requirements: All
  - Tasks:
    1. 更新 design.md（如有变更）
    2. 更新 tasks.md（标记完成状态）
    3. 验证所有需求完成
    4. 准备部署命令
  - Command:
    ```bash
    # 1. 运行所有测试
    cd ttpos-bmp/app/ttpos-erp
    go test ./internal/logic/selling/... -v
    
    # 2. 构建
    go build -o bin/ttpos-erp main.go
    
    # 3. 部署（根据实际部署流程）
    ```
  - Success: 文档更新完成，部署成功

---

## 📋 任务执行检查清单

### 执行前检查

- [ ] 确认 `story-erp-mode-of-payments-paymentid` 已完成（前置依赖）
- [ ] 确认 ERPNext 中 `custom_payment_id` 字段已添加
- [ ] 确认开发环境配置正确（GoFrame, gRPC, ERPNext API）

### 执行中检查

- [ ] Protobuf 定义符合规范（proto-rules.mdc）
- [ ] Logic 层实现符合规范（go-rules.mdc）
- [ ] 错误处理完整（使用 g.Log()，返回 error）
- [ ] 日志记录规范（中文描述）

### 执行后检查

- [ ] 所有单元测试通过
- [ ] 所有集成测试通过
- [ ] 性能测试达标（查询响应 < 200ms）
- [ ] 向后兼容性验证通过
- [ ] 文档更新完整

---

## 🔗 相关资源

### 核心文件

- Protobuf: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/selling/selling.proto`
- Logic: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling.go`
- Service: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/document.go`
- Test: `ttpos-bmp/app/ttpos-erp/internal/logic/selling/selling_test.go`

### 参考规范

- `ttpos-bmp/.cursor/rules/go-rules.mdc` - Go BMP 开发规范
- `ttpos-bmp/.cursor/rules/proto-rules.mdc` - Protobuf 开发规范
- `ttpos-bmp/.cursor/rules/erpnext.mdc` - ERPNext 集成规范

### 关联 Spec

- `story-erp-mode-of-payments-paymentid` - PaymentID 字段新增（前置依赖）

### GoFrame 工具命令

```bash
# 生成 Protobuf
cd ttpos-bmp/app/ttpos-erp
gf gen pb

# 生成 Service 接口
gf gen service

# 运行测试
go test ./internal/logic/selling/... -v

# 构建
go build -o bin/ttpos-erp main.go
```

---

## 📝 任务执行日志

### 2025-12-23
- ✅ 完成 Task 1.1: 新增 GetModeOfPaymentReq 和 GetModeOfPaymentResp 消息（在 Proposal 中已定义）
- ✅ 完成 Task 1.2: 修改 PosInvoicePayment 消息，新增 payment_id 字段（在 Proposal 中已定义）
- ✅ 完成 Task 1.3: 在 SellingService 中添加 GetModeOfPayment RPC 方法 (10:27)
- ✅ 完成 Task 2.1: 实现 GetModeOfPayment 查询逻辑 (10:30)
- ✅ 完成 Task 2.2: 实现 resolvePaymentIDs 辅助方法 (10:31)
- ✅ 完成 Task 2.3: 修改 SavePosInvoice，集成 payment_id 自动解析 (10:31)
- ✅ 完成 Task 3.1: 生成 Service 接口 (10:32)
- ✅ 完成 Task 3.2: 添加 Controller 层代码 (10:33)
- 📝 待执行 Task 4.1-4.2: 测试
- 📝 待执行 Task 5.1: 文档和部署

---

**版本**: v1.0.0  
**创建日期**: 2025-12-23  
**作者**: rikugun  
**最后更新**: 2025-12-23

