# Shop 端分批类型管理功能增强 任务分解

> 本文档定义 Shop 端分批类型管理功能增强的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 12  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

**注意**：多语言名称功能已在 v2.9.0 版本中实现，本次无需实现。

---

## Phase 1: 数据库设计和迁移

- [ ] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_abbreviation_to_batch_tag_table.php`
  - Purpose: 在 `ttpos_batch_tag` 表中增加 `abbreviation` 字段
  - Requirements: 2.1
  - Leverage: 现有迁移文件: `admin/database/migrations/20251011134625_add_product_batch_tag_table.php`，参考模板: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建数据库迁移文件，在 ttpos_batch_tag 表中增加 abbreviation 字段（VARCHAR(255), NOT NULL, DEFAULT ''）| Context: 字段位置在 multi_language_name_uuid 之后，参考现有迁移文件格式 | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查字段是否存在 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中执行迁移，增加 abbreviation 字段
  - Requirements: 2.1
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [ ] 1.3 更新 Go Model

  - File: `main/app/model/product.go`
  - Purpose: 在 BatchTag 结构体中增加 Abbreviation 字段
  - Requirements: 2.1
  - Leverage: 现有 Model: `main/app/model/product.go` - BatchTag 结构体（第 957-970 行）
  - Prompt: Role: Go Developer | Task: 在 BatchTag 结构体中增加 Abbreviation 字段 | Context: 字段类型为 string，gorm 标签为 `gorm:"default:'';column:abbreviation;comment:'名称缩写'"`，位置在 MultiLanguageNameUuid 之后 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 更新成功，字段映射正确

- [ ] 1.4 编写数据迁移脚本（为现有数据设置默认缩写）

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_migrate_batch_tag_abbreviation_data.php`
  - Purpose: 为现有的分批类型设置默认缩写
  - Requirements: 2.2
  - Leverage: 现有数据迁移脚本，参考 `main/app/model/product.go` 中的 BatchTag 结构，多语言名称已在 v2.9.0 实现
  - Prompt: Role: Database Engineer | Task: 编写数据迁移脚本，为现有分批类型设置默认缩写 | Context: 从 multi_language_name 表中提取中文名称（en_name）作为默认缩写，如果中文名称为空，则使用名称的前几个字符 | Restrictions: 遵循 .cursor/rules/database.mdc，确保数据不丢失 | Success: 迁移脚本创建成功，现有数据已处理

---

## Phase 2: 核心实现（Go Main）

### DTO 层

- [ ] 2.1 更新 Request DTO（增加 Abbreviation 字段）

  - File: `main/app/dto/req/product.go`
  - Purpose: 在 BatchTagAddReq 和 BatchTagEditReq 中增加 Abbreviation 字段
  - Requirements: 1.2, 1.3
  - Leverage: 现有 DTO: `main/app/dto/req/product.go` - BatchTagAddReq（第 609-613 行），BatchTagEditReq（第 615-620 行），多语言字段已存在
  - Prompt: Role: Go Developer | Task: 在 BatchTagAddReq 和 BatchTagEditReq 结构体中增加 Abbreviation 字段 | Context: 字段类型为 string，binding 标签为 `binding:"required"`，JSON 标签为 `json:"abbreviation"`，位置在 LocaleName 之后 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，validation 正确

- [ ] 2.2 更新 Response DTO（增加 Abbreviation 字段）

  - File: `main/app/dto/resp/product_resp/product.go`
  - Purpose: 在 BatchTag 和 BatchTagDetail 中增加 Abbreviation 字段
  - Requirements: 1.4
  - Leverage: 现有 DTO: `main/app/dto/resp/product_resp/product.go` - BatchTag（第 94-100 行），BatchTagDetail（第 107-113 行），多语言字段已存在
  - Prompt: Role: Go Developer | Task: 在 BatchTag 和 BatchTagDetail 结构体中增加 Abbreviation 字段 | Context: 字段类型为 string，JSON 标签为 `json:"abbreviation"`，位置在 LocaleName 之后 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 更新成功，响应格式正确

### Service 层

- [ ] 2.3 更新 AddBatchTag 方法（增加缩写字段处理）

  - File: `main/app/service/product.go`
  - Purpose: 在 AddBatchTag 方法中增加缩写字段的验证和处理
  - Requirements: 1.2
  - Leverage: 现有 Service: `main/app/service/product.go` - AddBatchTag 方法（第 7924-7967 行），多语言处理已存在
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 AddBatchTag 方法中增加缩写字段的验证和处理 | Context: 验证缩写字段必填和长度限制（1-10个字符），在创建 BatchTag 时设置 Abbreviation 字段，多语言处理保持不变 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 更新成功，业务逻辑正确，验证正确

- [ ] 2.4 更新 EditBatchTag 方法（增加缩写字段处理）

  - File: `main/app/service/product.go`
  - Purpose: 在 EditBatchTag 方法中增加缩写字段的验证和处理
  - Requirements: 1.3
  - Leverage: 现有 Service: `main/app/service/product.go` - EditBatchTag 方法（第 7970-8000 行），多语言处理已存在
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 EditBatchTag 方法中增加缩写字段的验证和处理 | Context: 验证缩写字段必填和长度限制（1-10个字符），在更新 BatchTag 时设置 Abbreviation 字段，多语言处理保持不变 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 更新成功，业务逻辑正确，验证正确

- [ ] 2.5 更新 GetBatchTag 方法（返回缩写字段）

  - File: `main/app/service/product.go`
  - Purpose: 在 GetBatchTag 方法中返回缩写字段
  - Requirements: 1.4
  - Leverage: 现有 Service: `main/app/service/product.go` - GetBatchTag 方法（第 7908-7921 行），多语言处理已存在
  - Prompt: Role: Go Developer | Task: 在 GetBatchTag 方法的返回值中增加 Abbreviation 字段 | Context: 从 batchTag.Abbreviation 读取并设置到响应结构体中，多语言处理保持不变 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 更新成功，返回数据正确

- [ ] 2.6 更新 GetBatchTagList 方法（返回缩写字段）

  - File: `main/app/service/product.go`
  - Purpose: 在 GetBatchTagList 方法中返回缩写字段
  - Requirements: 1.4
  - Leverage: 现有 Service: `main/app/service/product.go` - GetBatchTagList 方法（第 7880-7905 行），多语言处理已存在
  - Prompt: Role: Go Developer | Task: 在 GetBatchTagList 方法的返回值中增加 Abbreviation 字段 | Context: 在转换 BatchTag 列表时，为每个项设置 Abbreviation 字段，多语言处理保持不变 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 更新成功，返回数据正确

- [ ] 2.7 编写 Service 单元测试

  - File: `main/app/service/product_test.go`
  - Purpose: 为 AddBatchTag 和 EditBatchTag 方法编写单元测试，测试缩写字段验证
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/service/product_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 AddBatchTag 和 EditBatchTag 方法编写单元测试，重点测试缩写字段验证 | Context: 测试缩写字段必填验证、长度限制验证（1-10个字符）、正常创建和编辑流程，多语言功能已在 v2.9.0 测试 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

### API 层

- [ ] 2.8 更新 AddBatchTag API（参数验证）

  - File: `main/app/api/v1/shop/shop_batch_product.go`
  - Purpose: 在 AddBatchTag API 中增加缩写字段的参数验证
  - Requirements: 1.2
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_batch_product.go` - AddBatchTag 方法（第 85-98 行），多语言验证已存在
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 在 AddBatchTag API 中增加缩写字段的参数验证 | Context: 验证缩写字段必填和长度限制（1-10个字符），使用 helper.ErrorWithDetail 返回错误，多语言验证保持不变 | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON() | Success: API 更新成功，参数验证正确

- [ ] 2.9 更新 EditBatchTag API（参数验证）

  - File: `main/app/api/v1/shop/shop_batch_product.go`
  - Purpose: 在 EditBatchTag API 中增加缩写字段的参数验证
  - Requirements: 1.3
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_batch_product.go` - EditBatchTag 方法（第 111-124 行），多语言验证已存在
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 在 EditBatchTag API 中增加缩写字段的参数验证 | Context: 验证缩写字段必填和长度限制（1-10个字符），使用 helper.ErrorWithDetail 返回错误，多语言验证保持不变 | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON() | Success: API 更新成功，参数验证正确

- [ ] 2.10 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_batch_product_test.go`
  - Purpose: 测试分批类型 API 接口，验证缩写字段功能
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为分批类型 API 编写集成测试 | Context: 测试创建、编辑、详情接口，测试缩写字段验证（必填、长度限制），测试响应格式，多语言功能已在 v2.9.0 测试 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 3: 测试和优化

- [ ] 3.1 集成测试

  - File: `test/integration/batch_tag_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试创建分批类型（包含缩写字段）、编辑分批类型、查看详情、数据迁移后的兼容性，多语言功能已在 v2.9.0 测试 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 3.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 本地响应时间 < 200ms

- [ ] 3.3 文档更新

  - File: `main/docs/swagger.yaml`, `CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: API 文档（Swagger），数据库文档，CHANGELOG | Restrictions: 文档准确完整 | Success: 所有文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（Swagger）
- [ ] 数据库文档已更新（迁移脚本）
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
grep -c "^- \[" docs/shared/specs/story-shop-batch-tag-enhancement/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/story-shop-batch-tag-enhancement/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/story-shop-batch-tag-enhancement/tasks.md
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
**最后更新**: 2025-11-20  
**维护者**: xiezhihuan

