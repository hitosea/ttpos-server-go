# ERP 班次支付方式锁定与验证 任务分解

> 本文档定义 ERP 班次支付方式锁定与验证功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 8  
**已完成**: 5  
**进行中**: -  
**待手动测试**: 3  
**完成率**: 62.5%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_opening_payment_methods_to_staff_shift_log.php`
  - Purpose: 在 `ttpos_staff_shift_log` 表中新增 `opening_payment_methods` 字段
  - Requirements: 1.1
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，在 ttpos_staff_shift_log 表中新增 opening_payment_methods 字段（varchar(2000) 类型，可空） | Context: 字段用于存储逗号分隔的支付方式UUID列表，必须兼容历史数据（NULL） | Restrictions: 遵循 .cursor/rules/database.mdc，字段名使用 snake_case | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中新增字段
  - Requirements: 1.1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [x] 1.3 更新 Go Model

  - File: `main/app/model/staff.go`
  - Purpose: 在 `StaffShiftLog` 结构体中新增 `OpeningPaymentMethods` 字段，并添加 `OpeningPaymentMethod` 结构体定义
  - Requirements: 1.1, 1.3
  - Leverage: 现有 Model: `main/app/model/staff.go`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 在 StaffShiftLog 结构体中新增 OpeningPaymentMethods 字段（string 类型，gorm 标签指向 opening_payment_methods） | Context: 使用 gorm 标签，字段类型为 string（对应数据库 varchar(2000) 类型），用于存储逗号分隔的支付方式UUID列表 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段映射正确

---

## Phase 2: 核心实现（Go Main）

### Service 层

- [x] 2.1 修改 `CreateWorkingLog` 方法，保存支付方式列表

  - File: `main/app/service/staff_shift.go`
  - Purpose: 在班次开账时，当公司开启 ERP 时，保存当前已启用的支付方式列表到班次记录
  - Requirements: 1.1, 1.2, 1.3, 1.4
  - Leverage: 现有代码: `main/app/service/staff_shift.go` 的 `CreateWorkingLog` 方法（87-138 行），`main/app/repository/payment_method.go` 的 `GetPaymentMethodList` 方法
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 CreateWorkingLog 方法中，当公司开启 ERP 时，在获取支付方式列表后，将支付方式UUID用逗号连接并保存到 shiftLog.OpeningPaymentMethods 字段 | Context: 使用 strings.Join 将 UUID 数组连接成逗号分隔的字符串，格式如 "123456,123457,123458" | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，需要导入 fmt 和 strings 包 | Success: 支付方式UUID列表正确保存到班次记录，格式正确（逗号分隔）

- [x] 2.2 在 `IStaffShiftSrv` 接口中新增 `ValidatePaymentMethod` 方法

  - File: `main/app/service/staff_shift.go`
  - Purpose: 定义支付方式验证接口
  - Requirements: 2.1
  - Leverage: 现有接口: `main/app/service/staff_shift.go` 的 `IStaffShiftSrv` 接口（36-47 行）
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 在 IStaffShiftSrv 接口中新增 ValidatePaymentMethod 方法签名：`ValidatePaymentMethod(ctx context.Context, shiftNo string, paymentMethodUuid uint64) (bool, error)` | Context: 方法用于验证支付方式是否在开账时保存的列表中，shiftNo 是交班编号（string 类型） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

- [x] 2.3 实现 `ValidatePaymentMethod` 方法

  - File: `main/app/service/staff_shift.go`
  - Purpose: 实现支付方式验证逻辑
  - Requirements: 2.2, 2.3, 2.4
  - Leverage: 现有代码: `main/app/service/staff_shift.go`，`main/app/repository/staff_shift_log.go` 的 `GetShiftLog` 方法
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 ValidatePaymentMethod 方法，根据 shiftNo（交班编号）查询班次记录，如果未保存支付方式列表（历史数据）返回 false，如果保存了则用 strings.Split 分割逗号分隔的字符串并检查支付方式 UUID 是否在列表中 | Context: 使用 repository.CommonRepo.WhereByShiftNo(shiftNo) 查询班次记录，使用 strings.Split 分割字符串，使用 strings.TrimSpace 去除空格，遍历匹配 uuid，如果班次记录不存在返回错误，需要导入 strings 和 fmt 包 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error，使用 errors.WithMessage 包装错误 | Success: 方法实现完整，逻辑正确，错误处理完善，历史数据处理正确（返回 false），使用 shiftNo 查询正确

---

## Phase 3: 手动测试

- [ ] 3.1 手动测试 `CreateWorkingLog` 支付方式列表保存逻辑

  - File: -
  - Purpose: 手动验证支付方式UUID列表保存逻辑
  - Requirements: 1.1, 1.2, 1.3, 1.4
  - 测试步骤:
    1. 开启 ERP，创建班次开账
    2. 查询班次记录，检查 `opening_payment_methods` 字段
    3. 验证格式为逗号分隔的 UUID 列表（如 "1001,1002,1003"）
    4. 关闭 ERP，创建班次开账
    5. 验证 `opening_payment_methods` 字段为空
  - Success: 手动测试通过，支付方式列表保存正确

- [ ] 3.2 手动测试 `ValidatePaymentMethod` 验证逻辑

  - File: -
  - Purpose: 手动验证支付方式验证逻辑
  - Requirements: 2.2, 2.3, 2.4
  - 测试步骤:
    1. 创建班次记录，设置 `opening_payment_methods` 为 "1001,1002,1003"
    2. 调用 `ValidatePaymentMethod`，传入 shiftNo 和 paymentMethodUuid=1001
    3. 验证返回 true
    4. 调用 `ValidatePaymentMethod`，传入 shiftNo 和 paymentMethodUuid=9999
    5. 验证返回 false
    6. 创建历史班次记录（opening_payment_methods 为空）
    7. 调用 `ValidatePaymentMethod`，验证返回 false
    8. 调用 `ValidatePaymentMethod`，传入不存在的 shiftNo，验证返回错误
    9. 测试支付方式列表包含空格的情况（" 1001 , 1002 "），验证能正确处理
  - Success: 手动测试通过，所有场景验证正确

- [ ] 3.3 手动集成测试

  - File: -
  - Purpose: 手动测试端到端功能
  - Requirements: 所有功能需求
  - 测试步骤:
    1. 开启 ERP，创建班次开账
    2. 查询班次记录，确认支付方式列表已保存
    3. 调用 `ValidatePaymentMethod` 验证支付方式
    4. 验证结果符合预期
  - Success: 集成测试通过

---

## Phase 4: 文档更新

- [ ] 4.1 更新 API 文档（如需要）

  - File: `docs/shared/api/`（如有相关文档）
  - Purpose: 记录新增的 Service 方法
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Success: 文档已更新（本次不涉及 HTTP API，可跳过）

- [ ] 4.2 更新 CHANGELOG

  - File: `CHANGELOG.md`
  - Purpose: 记录功能变更
  - Requirements: 文档要求
  - Leverage: 现有 CHANGELOG
  - Success: CHANGELOG 已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%（使用现有方法，无需新增）
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] CHANGELOG.md 已更新
- [ ] 代码注释完整

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-erp-shift-payment-method-validation/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-erp-shift-payment-method-validation/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-erp-shift-payment-method-validation/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-erp-shift-payment-method-validation/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-erp-shift-payment-method-validation/tasks.md)" | bc
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
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-30  
**维护者**: 后端开发组

