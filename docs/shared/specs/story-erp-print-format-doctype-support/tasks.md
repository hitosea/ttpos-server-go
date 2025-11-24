# ERP Print Format Doctype 通用服务支持 任务分解

> 本文档定义 ERP Print Format Doctype 通用服务支持的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 13  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 数据结构定义

- [ ] 1.1 定义 Print Format DTO

  - File: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/print_format.go`
  - Purpose: 定义 Print Format 相关的数据结构
  - Requirements: 所有功能需求
  - Leverage: 现有 DTO: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/`，ERPNext Print Format 标准结构
  - Prompt: Role: Go Developer | Task: 创建 Print Format DTO 结构体 | Context: 包含 PrintFormatListReq, PrintFormatCreateUpdateReq, PrintFormatDetailResp, PrintFormatListResp | Restrictions: 遵循 .cursor/rules/go-bmp.mdc | Success: DTO 定义完整，字段映射正确

- [ ] 1.2 定义常量

  - File: `ttpos-bmp/app/ttpos-erp/internal/model/dto/erp/print_format.go`
  - Purpose: 定义 Print Format 相关常量
  - Requirements: 所有功能需求
  - Code: `const DocTypePrintFormat = "Print Format"`
  - Success: 常量定义正确

- [ ] 1.3 生成 Print Format 迁移文档

  - File: `ttpos-bmp/app/ttpos-erp/manifest/erp-migrate/v2.9/02_print_format/` (新建目录)
  - Purpose: 从 CSV 文件生成 Print Format 迁移 JSON 文档
  - Requirements: 所有功能需求
  - Leverage: CSV 文件: `ttpos-bmp/app/ttpos-erp/manifest/printformat/Wallace Print Format.csv`，参考 ERPNext Print Format JSON 结构
  - Prompt: Role: Go Developer | Task: 编写脚本或工具，读取 CSV 文件并生成 Print Format JSON 迁移文档 | Context: CSV 包含 ID, Standard, Module, DocType, Print Format For, Custom CSS, HTML, Custom Format 字段，需要为每个 Print Format 生成一个 JSON 文件，JSON 结构需符合 ERPNext Print Format 标准 | Restrictions: JSON 文件命名使用 Print Format 名称（如 wallace_delivery_note.json），遵循 erp-migrate 目录结构规范 | Success: 所有 Print Format 的 JSON 文件已生成，格式正确，可用于 ERPNext 迁移
  - Output: 每个 Print Format 生成一个 JSON 文件，例如：
    - `wallace_delivery_note.json`
    - `wallace_purchase_receipt.json`
    - `wallace_material_request.json`
    - `wallace_sales_invoice.json`
    - `wallace_sales_order.json`
    - `wallace_purchase_order.json`
  - JSON 结构示例:
    ```json
    {
      "name": "Wallace Delivery Note",
      "doc_type": "Delivery Note",
      "standard": 0,
      "module": "Stock",
      "print_format_type": "Jinja",
      "html": "...",
      "css": "...",
      "print_format_for": "DocType"
    }
    ```

---

## Phase 2: Logic 实现

- [ ] 2.1 创建 Print Format Logic 文件

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/print_format.go`
  - Purpose: 创建 Print Format Logic 服务文件
  - Requirements: 所有功能需求
  - Leverage: 现有 Logic: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/doctype.go`，design.md 中的完整实现代码
  - Prompt: Role: Go Developer | Task: 创建 print_format.go 文件，定义 sPrintFormat 结构体和 PrintFormat 变量 | Context: 参考 doctype.go 的实现模式 | Restrictions: 遵循 .cursor/rules/go-bmp.mdc | Success: 文件创建成功，结构体定义正确

- [ ] 2.2 实现 Meta 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/print_format.go`
  - Purpose: 实现 Print Format 元数据查询
  - Requirements: 1.1-1.4
  - Leverage: design.md 中的 Meta 实现代码
  - Prompt: Role: Go Developer | Task: 实现 Meta 方法，调用 service.Doctype().Meta | Context: DocType 为 "Print Format" | Restrictions: 错误处理使用 gerror.Wrapf | Success: Meta 方法实现正确，错误处理完善

- [ ] 2.3 实现 List 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/print_format.go`
  - Purpose: 实现 Print Format 列表查询
  - Requirements: 2.1-2.5
  - Leverage: design.md 中的 List 实现代码，参考 selling/sale_order.go 的列表查询实现
  - Prompt: Role: Go Developer | Task: 实现 List 方法，支持按 DocType 过滤和分页 | Context: 调用 service.Document().List，解析响应数据 | Restrictions: 错误处理使用 gerror.Wrapf | Success: List 方法实现正确，支持过滤和分页

- [ ] 2.4 实现 Get 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/print_format.go`
  - Purpose: 实现 Print Format 详情查询
  - Requirements: 3.1-3.4
  - Leverage: design.md 中的 Get 实现代码，参考 selling/sale_order.go 的详情查询实现
  - Prompt: Role: Go Developer | Task: 实现 Get 方法，根据名称查询 Print Format 详情 | Context: 调用 service.Document().Get，解析响应数据 | Restrictions: 参数验证，错误处理使用 gerror.Wrapf | Success: Get 方法实现正确，包含参数验证

- [ ] 2.5 实现 Create 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/print_format.go`
  - Purpose: 实现 Print Format 创建
  - Requirements: 4.1-4.5
  - Leverage: design.md 中的 Create 实现代码，参考 selling/sale_order.go 的创建实现
  - Prompt: Role: Go Developer | Task: 实现 Create 方法，创建新的 Print Format | Context: 调用 service.Document().Create，参数验证 | Restrictions: 参数验证，错误处理使用 gerror.Wrapf | Success: Create 方法实现正确，包含参数验证

- [ ] 2.6 实现 Update 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/print_format.go`
  - Purpose: 实现 Print Format 更新
  - Requirements: 4.1-4.5
  - Leverage: design.md 中的 Update 实现代码，参考 selling/sale_order.go 的更新实现
  - Prompt: Role: Go Developer | Task: 实现 Update 方法，更新现有的 Print Format | Context: 调用 service.Document().Update，然后调用 Get 获取更新后的信息 | Restrictions: 参数验证，错误处理使用 gerror.Wrapf | Success: Update 方法实现正确，包含参数验证

- [ ] 2.7 实现 Delete 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/print_format.go`
  - Purpose: 实现 Print Format 删除
  - Requirements: 5.1-5.4
  - Leverage: design.md 中的 Delete 实现代码
  - Prompt: Role: Go Developer | Task: 实现 Delete 方法，删除 Print Format | Context: 调用 service.Document().Delete | Restrictions: 参数验证，错误处理使用 gerror.Wrapf | Success: Delete 方法实现正确，包含参数验证

- [ ] 2.8 注册到 Logic

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/logic.go`
  - Purpose: 确保 print_format.go 被导入
  - Requirements: 所有功能需求
  - Leverage: 现有 logic.go 文件
  - Code: 确保 `_ "ttpos-bmp/app/ttpos-erp/internal/logic/erpnext"` 已存在（应该已经存在）
  - Success: Logic 注册成功

---

## Phase 3: 测试和优化

- [ ] 3.1 编写 Logic 单元测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/print_format_test.go`
  - Purpose: 测试 Print Format Logic 各个方法
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/doctype_test.go`（如果存在）
  - Prompt: Role: QA Engineer | Task: 为 PrintFormatLogic 编写单元测试，覆盖率 ≥ 70% | Context: 测试 Meta, List, Get, Create, Update, Delete 方法 | Restrictions: 使用 mock 或测试环境 | Success: 测试覆盖率达标，所有测试通过

- [ ] 3.2 集成测试

  - File: `ttpos-bmp/test/integration/print_format_test.go`（如果存在测试目录）
  - Purpose: 测试完整的 Print Format 流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端测试 | Context: 创建 Print Format → 查询详情 → 更新 → 查询列表 → 删除 | Success: 集成测试通过

- [ ] 3.3 文档更新

  - File: `docs/shared/api/erp_api.md`（如果存在）
  - Purpose: 更新 API 文档
  - Requirements: 文档要求
  - Leverage: 现有 API 文档
  - Success: 文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标（PrintFormatLogic ≥ 70%）
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-bmp.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/security.mdc`

---

## 进度追踪

### 执行流程

1. **选择任务**: 从 Phase 1 开始，按顺序执行
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 参考 design.md 中的实现代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

### 预计时间

- Phase 1: 1 天（8 小时，包含 CSV 解析和 JSON 生成）
- Phase 2: 2 天（16 小时）
- Phase 3: 1.5 天（12 小时）
- **总计**: 4.5 天（36 小时）= **SP 5**

---

## 附录：AI Prompt 示例

### Logic 实现

```
Role: Go Developer with ERPNext API expertise

Task: 实现 PrintFormatLogic，包含完整的 Print Format CRUD 逻辑

Context:
- File: ttpos-bmp/app/ttpos-erp/internal/logic/erpnext/print_format.go
- Leverage: design.md 中的完整实现代码
- Requirements: requirements.md 所有需求
- Dependencies: IDocument, IDoctype (通过 service.Document() 和 service.Doctype())

Implementation Steps:
1. 定义 sPrintFormat 结构体和 PrintFormat 变量
2. 实现 Meta 方法（调用 service.Doctype().Meta）
3. 实现 List 方法（调用 service.Document().List，支持过滤和分页）
4. 实现 Get 方法（调用 service.Document().Get）
5. 实现 Create 方法（调用 service.Document().Create）
6. 实现 Update 方法（调用 service.Document().Update）
7. 实现 Delete 方法（调用 service.Document().Delete）

Restrictions:
- 使用 gerror.Wrapf 包装错误
- 不使用 panic
- 参数验证
- 参考 doctype.go 和 sale_order.go 的实现模式

Success Criteria:
- 代码通过 go fmt 和 go vet
- 业务逻辑正确
- 错误处理完善
- 符合 Go BMP 规范
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 活动日志：`docs/team/activities/2025-11/2025-11-24.md`
- 当执行任务中形成复盘/优化建议时，及时沉淀 Episode 并在本节更新名称。

---

**模板版本**: v1.0.0  
**最后更新**: 2025-11-24  
**维护者**: 后端开发组

