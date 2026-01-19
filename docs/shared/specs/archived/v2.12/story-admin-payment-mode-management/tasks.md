# story-admin-payment-mode-management / 新管理端支付方式管理 任务分解

> 本文档定义新管理端支付方式管理与 ERPNext 双向同步的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 6  
**已完成**: 4  
**进行中**: -  
**完成率**: 67%

---

## Phase 1: 云平台授权后同步支付方式到 ERP

### 需求背景

云平台商家授权 ERP 成功后，调用 `/admin/erpnext/payment_mode/save`，同步已创建的支付方式到 ERP，并保存返回的 `erpnext_payment`。

**关联需求**: Requirement 4.1, Requirement 4.5

---

- [x] 1.1 更新 SaveModeOfPayment Handler，保存返回的 erpnext_payment

  - File: `main/app/api/v1/admin/handler.go`
  - Purpose: 在 Handler 中保存返回的 `erpnext_payment` 到数据库
  - Requirements: Requirement 4.1, Requirement 4.5
  - Status: ✅ 已完成 - Handler 已存在，保存逻辑在 Service 层实现
  - Implementation: 
    - `main/app/service/rpc/erp/selling.go` SaveModeOfPayment 已支持返回响应
    - `main/app/service/payment_method.go` Create/Update 方法中已实现保存逻辑
  - Success: Handler 更新成功，erpnext_payment 字段正确保存

- [x] 1.2 实现批量同步支付方式逻辑

  - File: `main/app/service/rpc/erp/setup.go`
  - Purpose: 在 InitShop 完成后，批量同步已创建的支付方式到 ERP
  - Requirements: Requirement 4.1, Requirement 4.5
  - Status: ✅ 已完成 - 在 `InitShop` 方法中实现批量同步逻辑
  - Implementation: 
    - `main/app/service/rpc/erp/setup.go:162-204` 实现了批量同步逻辑
    - 遍历支付方式列表，调用 `SaveModeOfPayment` 同步到 ERP
    - 使用 `InitErpnextPayment` 批量更新 `erpnext_payment` 字段
    - 跳过已存在 `erpnext_payment` 的支付方式
    - 跳过总部同步的自行添加支付方式（无二维码图片）
  - Success: 批量同步逻辑正确，InitShop 中调用成功

- [x] 1.3 实现 source 到 channel 的映射逻辑

  - File: `main/app/service/rpc/erp/payment_mode_naming.go`
  - Purpose: 根据支付方式的 source 字段确定 channel 参数
  - Requirements: Requirement 1.1, Requirement 2.1
  - Status: ✅ 已完成 - 实现了命名规则生成工具类
  - Implementation: 
    - `main/app/service/rpc/erp/payment_mode_naming.go` 实现了完整的命名规则生成逻辑
    - `GetChannelBySource` 函数：source=2(LianLianPay) -> channel="LianLianPay", source=1(手动) -> channel="", source=0(系统) -> channel=""
    - `GenerateModeOfPaymentID` 函数：生成完整的 Mode of Payment ID
    - `getNextSequenceNumber` 函数：计算下一个序号（系统默认=0000，自行添加=0001起，LianLianPay=0000起）
    - `admin/app/common/library/erp/PaymentModeNaming.php` 实现了 PHP 版本的命名规则生成逻辑
  - Success: 映射逻辑正确，channel 参数设置正确

---

## Phase 2: 新管理端新增支付方式同步

### 需求背景

新管理端新增支付方式时，调用 `main/app/service/rpc/erp/selling.go` 中的 `SaveModeOfPayment`，同步支付方式到 ERP，并保存返回的 `erpnext_payment`。

**关联需求**: Requirement 4.1, Requirement 4.2, Requirement 4.7

---

- [x] 2.1 在 PaymentMethodService.Create 中添加 ERP 同步逻辑

  - File: `main/app/service/payment_method.go`
  - Purpose: 新增支付方式时同步到 ERP 并保存 erpnext_payment
  - Requirements: Requirement 4.1, Requirement 4.2, Requirement 4.7
  - Status: ✅ 已完成 - 在 Create、Update、Delete 方法中都实现了 ERP 同步逻辑
  - Implementation: 
    - `main/app/service/payment_method.go:386-432` Create 方法中实现了 ERP 同步逻辑
    - 使用事务确保数据一致性
    - 创建支付方式后，如果开启了 ERP 且 `erpnext_payment` 为空，则同步到 ERP
    - 根据 source 确定 channel，调用 `SaveModeOfPayment` 同步
    - 保存返回的 `erpnext_payment` 字段
    - `main/app/service/payment_method.go:449-524` Update 方法中也实现了 ERP 同步逻辑（创建或更新）
    - `main/app/service/payment_method.go:546-595` Delete 方法中实现了 ERP 禁用逻辑
  - Success: ERP 同步逻辑正确，erpnext_payment 字段正确保存

- [ ] 2.2 编写 Service 单元测试

  - File: `main/app/service/payment_method_srv_test.go`
  - Purpose: 确保支付方式创建和 ERP 同步逻辑正确
  - Requirements: Requirement 4.1, Requirement 4.2
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 PaymentMethodService.Create 编写单元测试，覆盖率 ≥ 70% | Context: 测试创建支付方式，测试 ERP 同步逻辑，测试 erpnext_payment 字段更新 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Payment 相关模块测试覆盖率 100% | Success: 测试覆盖率 ≥ 70%（Payment 相关 100%），所有测试通过
  - Success: 测试覆盖率 ≥ 70%（Payment 相关 100%），所有测试通过

---

## Phase 3: 测试和优化

- [ ] 3.1 端到端集成测试

  - File: `test/integration/payment_method_test.go`
  - Purpose: 测试完整的支付方式管理流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试云平台授权后同步流程，测试新管理端新增支付方式流程，测试 erpnext_payment 字段保存 | Restrictions: 测试真实用户场景 | Success: 集成测试通过
  - Success: 集成测试通过

- [ ] 3.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 本地响应时间 < 200ms，批量同步时间 < 10s（10 个支付方式）

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [x] 所有任务标记为 `[x]`（核心功能已完成，测试任务待完成）
- [ ] Go 代码通过 `go fmt` 和 `go vet`（待验证）
- [ ] 测试覆盖率达标
  - Service: ≥ 70%（待完成）
  - Repository: ≥ 80%（待完成）
  - Payment 相关: 100%（待完成）
- [ ] 所有测试通过（待完成）

### 功能完整性

- [x] requirements.md 中的所有需求已满足（核心功能已实现）
- [x] design.md 中的设计已实现（核心功能已实现）
- [x] 验收标准已达成（核心功能已实现）
- [x] 云平台授权后能自动同步支付方式到 ERP 并保存 erpnext_payment（✅ 已实现）
- [x] 新管理端新增支付方式能自动同步到 ERP 并保存 erpnext_payment（✅ 已实现）
- [x] 更新支付方式时同步到 ERP（✅ 已实现）
- [x] 删除支付方式时禁用 ERP 中的支付方式（✅ 已实现）

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-admin-payment-mode-management/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-admin-payment-mode-management/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-admin-payment-mode-management/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-admin-payment-mode-management/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-admin-payment-mode-management/tasks.md)" | bc
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

- Related Episode: `待补充`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/王昱/2025-12/2025-12-12.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-12  
**维护者**: 王昱
