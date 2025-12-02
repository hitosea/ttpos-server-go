# 订单来源追踪 任务分解

> 本文档定义 订单来源追踪 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 18  
**已完成**: 12  
**进行中**: -  
**完成率**: 67%

---

## Phase 0: 数据库迁移

- [x] 0.1 创建数据库迁移文件，新增 client_version 字段

  - File: `admin/database/migrations/{timestamp}_add_client_version_to_sale_bill.php`
  - Purpose: 在 `ttpos_sale_bill` 表中新增 `client_version` 字段
  - Requirements: 0.1, 0.2, 0.3, 0.4
  - Leverage: 参考: `admin/database/migrations/20251126201148_add_source_to_sale_bill.php`
  - Prompt: Role: PHP Developer specializing in Database Migrations | Task: 创建数据库迁移文件，在 ttpos_sale_bill 表中新增 client_version 字段 | Context: 字段类型 varchar(20)，默认值空字符串，注释"客户端版本号（如 2.10.0、2.9.0）"，位置在 source 字段之后，添加索引 idx_client_version | Restrictions: 遵循 .cursor/rules/database.mdc，参考现有迁移文件格式 | Success: 迁移文件创建成功，可以正常执行和回滚

- [x] 0.2 更新 SaleBill 模型，添加 ClientVersion 字段定义

  - File: `main/app/model/sale_bill.go`
  - Purpose: 在 SaleBill 模型中添加 ClientVersion 字段
  - Requirements: 0.5
  - Leverage: 现有模型: `main/app/model/sale_bill.go`，参考 Source 字段定义
  - Prompt: Role: Go Developer specializing in Model | Task: 在 SaleBill 模型中添加 ClientVersion 字段定义 | Context: 字段类型 string，gorm tag: column:client_version;type:varchar(20);default:'';comment:客户端版本号（如 2.10.0、2.9.0），json tag: client_version | Restrictions: 遵循 .cursor/rules/go-main.mdc，字段位置在 Source 字段之后 | Success: ClientVersion 字段定义正确，代码通过 go fmt 和 go vet

- [ ] 0.3 验证迁移脚本可以正常执行和回滚

  - File: `admin/database/migrations/{timestamp}_add_client_version_to_sale_bill.php`
  - Purpose: 确保数据库迁移脚本可以正常执行和回滚
  - Requirements: 0.6
  - Leverage: 测试环境
  - Success: 迁移脚本可以正常执行和回滚，字段已成功添加

---

## Phase 1: 常量定义和映射函数

- [x] 1.1 创建 Source 映射常量文件

  - File: `main/app/constant/sale_bill_source.go`
  - Purpose: 定义 SaleBill.source 字段的常量值和映射函数
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 现有常量: `main/app/constant/jwt/jwt.go`，参考: `main/app/constant/order.go`
  - Prompt: Role: Go Developer specializing in Constants | Task: 创建 sale_bill_source.go 文件，定义 Source 映射函数 MapJwtSourceToSaleBillSource | Context: 使用 switch 语句实现映射，支持 cashier→1, assistant→2, tablet→3, h5→4, member→5, 其他→0 | Restrictions: 遵循 .cursor/rules/go-main.mdc，函数名使用 MapJwtSourceToSaleBillSource，注意常量已在 device.go 中定义 | Success: 映射函数正确，代码通过 go fmt 和 go vet

- [x] 1.2 编写 Source 映射函数单元测试

  - File: `main/app/constant/sale_bill_source_test.go`
  - Purpose: 确保 Source 映射函数正确性
  - Requirements: 1.4
  - Leverage: 现有测试: `main/app/constant/*_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 MapJwtSourceToSaleBillSource 编写单元测试，覆盖率 100% | Context: 测试所有 JWT Source 值映射到正确的 source 值（包括 member→5），测试未知来源返回默认值 0 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 100%，所有测试通过

- [x] 1.3 实现 NormalizeClientVersion 函数（版本号格式标准化）

  - File: `main/app/constant/sale_bill_source.go`
  - Purpose: 标准化客户端版本号格式，去除前缀 "v" 或 "V"，统一为 x.y.z 格式
  - Requirements: 2.4, 3.4, 4.4
  - Leverage: 现有文件: `main/app/constant/sale_bill_source.go`
  - Prompt: Role: Go Developer specializing in Constants | Task: 在 sale_bill_source.go 中添加 NormalizeClientVersion 函数，用于标准化客户端版本号格式 | Context: 函数功能：去除前缀 "v" 或 "V"，去除前后空格，统一为 x.y.z 格式。如果为空则返回空字符串。示例: "v2.10.0" -> "2.10.0", "V2.10.0" -> "2.10.0", "" -> "" | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 strings.TrimPrefix 和 strings.TrimSpace | Success: 函数实现正确，代码通过 go fmt 和 go vet

- [x] 1.4 编写 NormalizeClientVersion 函数单元测试

  - File: `main/app/constant/sale_bill_source_test.go`
  - Purpose: 确保版本号标准化函数正确性
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/constant/sale_bill_source_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 NormalizeClientVersion 编写单元测试，覆盖率 100% | Context: 测试正常版本号、带 v/V 前缀、带空格、空字符串等场景 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 100%，所有测试通过

---

## Phase 2: 修改订单创建逻辑

### 即时订单

- [x] 2.1 修改 CreateInstantOrder 方法设置 source

  - File: `main/app/service/order_base.go`
  - Purpose: 在创建即时订单时设置 source 字段
  - Requirements: 2.1, 2.2, 2.3
  - Leverage: 现有代码: `main/app/service/order_base.go:58`，Task 1.1 的映射函数
  - Prompt: Role: Go Developer specializing in Order Service | Task: 在 CreateInstantOrder 方法中，创建 SaleBill 时设置 Source 字段 | Context: 使用 constant.MapJwtSourceToSaleBillSource(ctx.GetSource()) 获取 source 值，在 CreateSaleBill 调用时设置 Source 字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不影响现有业务逻辑 | Success: source 字段正确设置，代码通过 go fmt 和 go vet

- [x] 2.1.1 修改 CreateInstantOrder 方法同时设置 client_version

  - File: `main/app/service/order_base.go`
  - Purpose: 在创建即时订单时**同时**设置 client_version 字段（与 source 一起设置）
  - Requirements: 2.2, 2.6
  - Leverage: 现有代码: `main/app/service/order_base.go:64`，Task 1.3 的版本号标准化函数
  - Prompt: Role: Go Developer specializing in Order Service | Task: 在 CreateInstantOrder 方法中，创建 SaleBill 时**同时**设置 Source 和 ClientVersion 字段 | Context: ⚠️ 重要：source 和 client_version 必须在同一位置、同一时机一起设置。使用 constant.MapJwtSourceToSaleBillSource(ctx.GetSource()) 设置 Source，使用 constant.NormalizeClientVersion(ctx.GetVersion()) 设置 ClientVersion | Restrictions: 遵循 .cursor/rules/go-main.mdc，确保 source 和 client_version 一起设置，不影响现有业务逻辑 | Success: source 和 client_version 字段同时正确设置，代码通过 go fmt 和 go vet

- [ ] 2.2 编写 CreateInstantOrder 单元测试

  - File: `main/app/service/order_base_test.go`
  - Purpose: 验证即时订单 source 和 client_version 字段设置正确
  - Requirements: 2.7
  - Leverage: 现有测试: `main/app/service/order_base_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 CreateInstantOrder 添加 source 和 client_version 字段设置的测试用例 | Context: ⚠️ 重要：测试 source 和 client_version 一起设置。测试不同来源（cashier, assistant, tablet, h5, member）和不同版本（2.10.0, v2.10.0, 空字符串）创建即时订单时字段正确设置 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率达标，所有测试通过，验证 source 和 client_version 一起设置

### 桌台订单

- [x] 2.3 修改 CreateDeskOrder 方法设置 source

  - File: `main/app/service/order_base.go`
  - Purpose: 在创建桌台订单时设置 source 字段
  - Requirements: 3.1, 3.2, 3.3
  - Leverage: 现有代码: `main/app/service/order_base.go:179`，Task 1.1 的映射函数
  - Prompt: Role: Go Developer specializing in Order Service | Task: 在 CreateDeskOrder 方法中，创建 SaleBill 时设置 Source 字段 | Context: 在创建 SaleBill 之前，使用 constant.MapJwtSourceToSaleBillSource(ctx.GetSource()) 设置 saleBill.Source | Restrictions: 遵循 .cursor/rules/go-main.mdc，不影响现有业务逻辑 | Success: source 字段正确设置，代码通过 go fmt 和 go vet

- [x] 2.3.1 修改 CreateDeskOrder 方法同时设置 client_version

  - File: `main/app/service/order_base.go`
  - Purpose: 在创建桌台订单时**同时**设置 client_version 字段（与 source 一起设置）
  - Requirements: 3.2, 3.6
  - Leverage: 现有代码: `main/app/service/order_base.go:180`，Task 1.3 的版本号标准化函数
  - Prompt: Role: Go Developer specializing in Order Service | Task: 在 CreateDeskOrder 方法中，创建 SaleBill 时**同时**设置 Source 和 ClientVersion 字段 | Context: ⚠️ 重要：source 和 client_version 必须在同一位置、同一时机一起设置。在创建 SaleBill 之前，使用 constant.MapJwtSourceToSaleBillSource(ctx.GetSource()) 设置 saleBill.Source，使用 constant.NormalizeClientVersion(ctx.GetVersion()) 设置 saleBill.ClientVersion | Restrictions: 遵循 .cursor/rules/go-main.mdc，确保 source 和 client_version 一起设置，不影响现有业务逻辑 | Success: source 和 client_version 字段同时正确设置，代码通过 go fmt 和 go vet

- [ ] 2.4 编写 CreateDeskOrder 单元测试

  - File: `main/app/service/order_base_test.go`
  - Purpose: 验证桌台订单 source 和 client_version 字段设置正确
  - Requirements: 3.7
  - Leverage: 现有测试: `main/app/service/order_base_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 CreateDeskOrder 添加 source 和 client_version 字段设置的测试用例 | Context: ⚠️ 重要：测试 source 和 client_version 一起设置。测试不同来源（cashier, assistant, tablet, h5, member）和不同版本（2.10.0, v2.10.0, 空字符串）创建桌台订单时字段正确设置 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率达标，所有测试通过，验证 source 和 client_version 一起设置

### 会员端订单

- [x] 2.5 修改 createMemberOrder 方法设置 source

  - File: `main/app/service/order.go`
  - Purpose: 在创建会员端订单时设置 source 字段
  - Requirements: 4.1, 4.2, 4.3
  - Leverage: 现有代码: `main/app/service/order.go:617`，Task 1.1 的映射函数
  - Prompt: Role: Go Developer specializing in Order Service | Task: 在 createMemberOrder 方法中，创建 SaleBill 时设置 Source 字段 | Context: 使用 constant.MapJwtSourceToSaleBillSource(ctx.GetSource()) 获取 source 值，在 CreateSaleBill 调用时设置 Source 字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不影响现有业务逻辑 | Success: source 字段正确设置，代码通过 go fmt 和 go vet

- [x] 2.5.1 修改 createMemberOrder 方法同时设置 client_version

  - File: `main/app/service/order.go`
  - Purpose: 在创建会员端订单时**同时**设置 client_version 字段（与 source 一起设置）
  - Requirements: 4.2, 4.6
  - Leverage: 现有代码: `main/app/service/order.go:624`，Task 1.3 的版本号标准化函数
  - Prompt: Role: Go Developer specializing in Order Service | Task: 在 createMemberOrder 方法中，创建 SaleBill 时**同时**设置 Source 和 ClientVersion 字段 | Context: ⚠️ 重要：source 和 client_version 必须在同一位置、同一时机一起设置。使用 constant.MapJwtSourceToSaleBillSource(ctx.GetSource()) 设置 Source，使用 constant.NormalizeClientVersion(ctx.GetVersion()) 设置 ClientVersion | Restrictions: 遵循 .cursor/rules/go-main.mdc，确保 source 和 client_version 一起设置，不影响现有业务逻辑 | Success: source 和 client_version 字段同时正确设置，代码通过 go fmt 和 go vet

- [ ] 2.6 编写 createMemberOrder 单元测试

  - File: `main/app/service/order_test.go`
  - Purpose: 验证会员端订单 source 和 client_version 字段设置正确
  - Requirements: 4.7
  - Leverage: 现有测试: `main/app/service/order_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 createMemberOrder 添加 source 和 client_version 字段设置的测试用例 | Context: ⚠️ 重要：测试 source 和 client_version 一起设置。测试通过会员端创建订单时 source 字段设置为 5，client_version 字段正确设置 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率达标，所有测试通过，验证 source 和 client_version 一起设置

---

## Phase 3: 全面检查和验证

- [x] 3.1 搜索所有 CreateSaleBill 调用位置

  - File: -
  - Purpose: 确保所有创建 SaleBill 的路径都设置了 source
  - Requirements: 5.1, 5.2
  - Leverage: 代码搜索工具（grep）
  - Command: `grep -r "CreateSaleBill" main/app/service/`
  - Success: 确认所有创建路径都已设置 source

- [x] 3.2 检查订单导入逻辑（可选）

  - File: `main/app/service/order_import_service.go`
  - Purpose: 对于导入订单，如果没有来源信息，使用默认值 0
  - Requirements: 5.4
  - Leverage: 现有代码: `main/app/service/order_import_service.go:526`
  - Prompt: Role: Go Developer | Task: 检查订单导入逻辑，确保导入订单的 source 字段设置为默认值 0 | Context: 导入订单可能没有来源信息，应使用 constant.SaleBillSourceDefault | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 导入订单 source 字段正确设置

- [x] 3.2.1 检查订单导入逻辑同时设置 client_version（可选）

  - File: `main/app/service/order_import_service.go`
  - Purpose: 对于导入订单，如果没有版本信息，使用空字符串
  - Requirements: 5.5
  - Leverage: 现有代码: `main/app/service/order_import_service.go:526`
  - Prompt: Role: Go Developer | Task: 检查订单导入逻辑，确保导入订单的 client_version 字段设置为空字符串 | Context: ⚠️ 重要：source 和 client_version 必须一起设置。导入订单可能没有版本信息，应使用空字符串 "" | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 导入订单 client_version 字段正确设置，与 source 一起设置

- [ ] 3.3 API 集成测试

  - File: `test/integration/order_source_test.go`
  - Purpose: 测试端到端功能，验证不同来源和版本创建订单时 source 和 client_version 字段正确
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试: `test/integration/`
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试，验证订单来源和版本追踪功能 | Context: ⚠️ 重要：测试 source 和 client_version 一起设置。测试通过收银机、点餐助手、平板、H5、会员端创建订单，测试不同版本（2.10.0, v2.10.0, 空字符串），验证数据库中的 source 和 client_version 字段值 | Restrictions: 测试真实用户场景 | Success: 集成测试通过，所有来源和版本的订单字段正确设置

- [ ] 3.4 手动测试验证

  - File: -
  - Purpose: 通过不同客户端和版本创建订单，验证数据库中的 source 和 client_version 字段值
  - Requirements: 功能验收标准
  - Leverage: 测试环境
  - Success: 所有客户端和版本的订单 source 和 client_version 字段正确设置，验证两个字段一起设置

---

## Phase 4: 代码审查和文档

- [x] 4.1 代码审查检查清单

  - File: -
  - Purpose: 确保所有创建 SaleBill 的路径都设置了 source
  - Requirements: 5.4
  - Leverage: 代码审查工具
  - Success: 代码审查通过，无遗漏
  
  **审查结果**:
  - ✅ CreateInstantOrder: 已设置 source（需补充 client_version）
  - ✅ CreateDeskOrder: 已设置 source（需补充 client_version）
  - ✅ createMemberOrder: 已设置 source（需补充 client_version）
  - ✅ 订单导入: 已设置 source = 0（默认值，需补充 client_version = ""）
  
  **审查重点**:
  - ⚠️ 确保 source 和 client_version 在同一位置、同一时机一起设置
  - ⚠️ 检查所有创建 SaleBill 的路径都同时设置了两个字段

- [ ] 4.2 更新相关文档（如有需要）

  - File: `docs/shared/api/order_api.md`（如有）
  - Purpose: 确保文档与代码同步
  - Requirements: 文档验收
  - Leverage: 现有文档
  - Success: 文档已更新（本功能不涉及 API 变更，可能无需更新）

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Constant: 100%
  - Service: ≥ 70%
  - Payment/Order: 100%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成
  - Source 映射正确性
  - 版本号格式标准化
  - 即时订单 source 和 client_version 设置（一起设置）
  - 桌台订单 source 和 client_version 设置（一起设置）
  - 会员端订单 source 和 client_version 设置（一起设置）
  - 默认值处理
  - 数据一致性（source 和 client_version 一起设置）

### 文档同步

- [ ] API 文档已更新（本功能无新增 API，可能无需更新）
- [ ] 数据库文档已更新（source 字段已存在，无需更新）
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
grep -c "^- \[" docs/shared/specs/active/story-pos-order-source-tracking/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-pos-order-source-tracking/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-pos-order-source-tracking/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-pos-order-source-tracking/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-pos-order-source-tracking/tasks.md)" | bc
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

### Go 后端开发

```
Role: Go Developer specializing in {具体领域}

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 .cursor/rules/go-main.mdc, .cursor/rules/api.mdc, .cursor/rules/database.mdc

Restrictions:
- 接口以 I 开头，实现以 Impl 结尾
- Service 只依赖其他 Service 接口
- Repository 只持有 db 实例
- 不使用 panic，返回 error
- 使用 errors.WithMessage 包装错误

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70% (Service) 或 100% (Constant)
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: 100% (Constant) 或 ≥ 70% (Service)

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- 所有 JWT Source 值映射测试

Restrictions:
- 遵循 .cursor/rules/go-main.mdc
- 必须包含边界情况测试

Success Criteria:
- 测试覆盖率达标
- 所有测试通过
- 边界情况已覆盖
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-27  
**维护者**: 后端开发组

