# ERP 文档初始化支持更新模式 任务分解

> 本文档定义 ERP 文档初始化支持更新模式 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 0.5-1 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 6  
**已完成**: 3  
**进行中**: Task 2.1 (单元测试)  
**完成率**: 50%

---

## Phase 1: 代码实现

### 核心修改

- [x] 1.1 修改 initDocumentsFromDir 方法

  - **File**: `ttpos-bmp/app/ttpos-erp/internal/logic/setup/setup.go` (第 520-527 行)
  - **Purpose**: 增加智能判断逻辑，根据 name 字段决定创建或更新
  - **Requirements**: Requirement 1 (智能判断创建或更新)、Requirement 2 (错误处理和日志)、Requirement 3 (向后兼容)
  - **Leverage**: 
    - 现有方法结构: `ttpos-bmp/app/ttpos-erp/internal/logic/setup/setup.go` (第 485-538 行)
    - `service.Document().Create()`: 已有的创建方法
    - `service.Document().Update()`: 需要使用的更新方法
  - **Prompt**:
    ```
    Role: Go Developer specializing in GoFrame Logic Layer
    
    Task: 修改 initDocumentsFromDir 方法，使其能够根据 JSON 数据中的 name 字段智能判断并执行创建或更新操作
    
    Context:
    - Current file: ttpos-bmp/app/ttpos-erp/internal/logic/setup/setup.go (第 520-527 行)
    - Leverage code: 现有的 service.Document().Create() 和 service.Document().Update() 方法
    - Requirements: 
      - 1.1: 从 docData 中读取 name 字段并进行类型断言
      - 1.2: 判断 name 字段是否存在且不为空
      - 1.3: 根据判断结果调用对应的 service 方法
      - 2.3: 更新成功时记录 Info 级别日志，格式：{ItemName}更新成功: {path}
      - 2.4: 创建成功时记录 Info 级别日志，格式：{ItemName}创建成功: {path}
    - Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    
    Implementation Details:
    1. 在调用 service.Document.Create 之前，添加类型断言：
       docName, hasName := docData["name"].(string)
    2. 添加条件分支：
       if hasName && docName != "" {
           // 调用 Update 方法
       } else {
           // 调用 Create 方法（保持原有逻辑）
       }
    3. 分别为两个分支添加成功日志：
       - Update 成功: g.Log().Infof(ctx, "%s更新成功: %s", config.ItemName, path)
       - Create 成功: g.Log().Infof(ctx, "%s创建成功: %s", config.ItemName, path)
    4. 保持错误处理机制不变
    
    Restrictions:
    - 所有注释使用中文
    - 不使用 panic，返回 error
    - 遵循 GoFrame 开发规范
    - 保持方法签名不变
    - 不影响其他调用此方法的代码
    
    Success Criteria:
    - name 不为空时调用 Update 方法
    - name 为空时调用 Create 方法
    - 成功日志明确标注"创建"或"更新"
    - 错误日志保持现有格式
    - 代码通过 go fmt 和 go vet
    ```
  - **Success**: 
    - [x] name 不为空时调用 Update 方法
    - [x] name 为空时调用 Create 方法
    - [x] 成功日志明确标注"创建"或"更新"
    - [x] 错误日志保持现有格式

- [x] 1.2 更新方法注释

  - **File**: `ttpos-bmp/app/ttpos-erp/internal/logic/setup/setup.go` (第 477-484 行)
  - **Purpose**: 更新 initDocumentsFromDir 方法的注释，说明支持创建和更新两种模式
  - **Requirements**: 文档要求
  - **Leverage**: 现有注释风格
  - **Implementation**:
    ```go
    // initDocumentsFromDir 通用的文档初始化方法
    // 遍历指定目录下的所有JSON文件，解析并创建或更新对应类型的文档
    // 如果JSON数据中包含非空的name字段，则更新已有文档；否则创建新文档
    // 参数：
    //   - ctx: 上下文对象
    //   - config: 初始化配置
    //
    // 返回：
    //   - err: 错误信息
    ```
  - **Success**: 注释更新完成，说明清晰

---

## Phase 2: 测试验证

### 单元测试

- [ ] 2.1 编写单元测试

  - **File**: `ttpos-bmp/app/ttpos-erp/internal/logic/setup/setup_test.go`
  - **Purpose**: 确保 initDocumentsFromDir 方法的两种模式都正确工作
  - **Requirements**: 测试要求（覆盖率 ≥ 80%）
  - **Leverage**: 现有测试结构（如有）
  - **Test Cases**:
    1. **Test_initDocumentsFromDir_CreateNewDocument**: 测试创建新文档
       - 输入: JSON 数据不包含 name 字段
       - Mock: `service.Document().Create()` 返回成功
       - 验证: Create 方法被调用，日志包含"创建成功"
    
    2. **Test_initDocumentsFromDir_UpdateExistingDocument**: 测试更新已有文档
       - 输入: JSON 数据包含 name 字段且不为空
       - Mock: `service.Document().Update()` 返回成功
       - 验证: Update 方法被调用，日志包含"更新成功"
    
    3. **Test_initDocumentsFromDir_EmptyName**: 测试 name 为空字符串
       - 输入: JSON 数据包含 name 字段但值为 ""
       - Mock: `service.Document().Create()` 返回成功
       - 验证: Create 方法被调用
    
    4. **Test_initDocumentsFromDir_UpdateFailure**: 测试更新失败
       - 输入: JSON 数据包含 name 字段
       - Mock: `service.Document().Update()` 返回错误
       - 验证: 记录错误日志，继续处理
    
    5. **Test_initDocumentsFromDir_CreateFailure**: 测试创建失败
       - 输入: JSON 数据不包含 name 字段
       - Mock: `service.Document().Create()` 返回错误
       - 验证: 记录错误日志，继续处理
    
    6. **Test_initDocumentsFromDir_WrongType**: 测试 name 字段类型错误
       - 输入: JSON 数据包含 name 字段但类型不是 string
       - 验证: 类型断言失败，调用 Create 方法
  - **Prompt**:
    ```
    Role: QA Engineer with Go testing expertise
    
    Task: 为 initDocumentsFromDir 方法编写单元测试，覆盖率 ≥ 80%
    
    Context:
    - Target file: ttpos-bmp/app/ttpos-erp/internal/logic/setup/setup.go
    - Test file: ttpos-bmp/app/ttpos-erp/internal/logic/setup/setup_test.go
    - Coverage target: ≥ 80%
    
    Test Cases Required:
    1. 创建新文档（name 字段不存在或为空）
    2. 更新已有文档（name 字段存在且不为空）
    3. 更新失败的错误处理
    4. 创建失败的错误处理
    5. name 字段类型错误的处理
    6. 混合场景（多个文件，部分创建，部分更新）
    
    Mock Requirements:
    - Mock service.Document().Create() 方法
    - Mock service.Document().Update() 方法
    - Mock 文件系统（使用临时目录）
    
    Restrictions:
    - 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
    - 使用 GoFrame 测试工具
    - 必须包含边界情况测试
    
    Success Criteria:
    - 测试覆盖率 ≥ 80%
    - 所有测试通过
    - 边界情况已覆盖
    - Mock 正确设置
    ```
  - **Success**: 
    - 测试覆盖率 ≥ 80%
    - 所有测试通过
    - 6 个测试用例全部通过

### 集成测试

- [ ] 2.2 手动集成测试

  - **File**: -
  - **Purpose**: 在开发环境验证创建和更新两种场景
  - **Requirements**: Requirement 1, 2, 3
  - **Test Environment**: 开发环境 ERP 系统
  - **Test Steps**:
    
    **步骤 1: 准备测试 JSON 文件**
    ```bash
    # 创建测试目录
    mkdir -p /tmp/erp-test/custom_fields
    
    # 创建测试 JSON 文件（不包含 name，用于创建）
    cat > /tmp/erp-test/custom_fields/test_field_1.json <<EOF
    {
        "doctype": "Custom Field",
        "label": "测试字段1",
        "fieldtype": "Data",
        "dt": "Item"
    }
    EOF
    ```
    
    **步骤 2: 首次执行初始化**
    ```bash
    # 执行初始化（会调用 Create 方法）
    cd ttpos-bmp/app/ttpos-erp
    go run cmd/cmd.go --init-custom-fields /tmp/erp-test/custom_fields
    
    # 检查日志，应包含 "创建成功"
    # 检查 ERPNext 系统，验证自定义字段已创建
    ```
    
    **步骤 3: 修改 JSON 文件，添加 name 字段**
    ```bash
    # 获取创建的文档名称（从 ERPNext 系统或日志中）
    # 假设文档名称为: "Custom Field-Item-test_field_1"
    
    # 更新 JSON 文件，添加 name 字段
    cat > /tmp/erp-test/custom_fields/test_field_1.json <<EOF
    {
        "name": "Custom Field-Item-test_field_1",
        "doctype": "Custom Field",
        "label": "测试字段1（已更新）",
        "fieldtype": "Data",
        "dt": "Item"
    }
    EOF
    ```
    
    **步骤 4: 重复执行初始化**
    ```bash
    # 再次执行初始化（会调用 Update 方法）
    go run cmd/cmd.go --init-custom-fields /tmp/erp-test/custom_fields
    
    # 检查日志，应包含 "更新成功"
    # 检查 ERPNext 系统，验证字段 label 已更新为 "测试字段1（已更新）"
    ```
    
    **步骤 5: 混合场景测试**
    ```bash
    # 创建第二个 JSON 文件（不包含 name，用于创建新文档）
    cat > /tmp/erp-test/custom_fields/test_field_2.json <<EOF
    {
        "doctype": "Custom Field",
        "label": "测试字段2",
        "fieldtype": "Int",
        "dt": "Item"
    }
    EOF
    
    # 执行初始化（应该创建 test_field_2，更新 test_field_1）
    go run cmd/cmd.go --init-custom-fields /tmp/erp-test/custom_fields
    
    # 检查日志，应包含：
    # - "test_field_1 更新成功"
    # - "test_field_2 创建成功"
    ```
  - **Verification**:
    - [ ] 首次执行：文档创建成功，日志包含"创建成功"
    - [ ] 重复执行：文档更新成功，日志包含"更新成功"
    - [ ] ERPNext 系统中数据正确
    - [ ] 混合场景：创建和更新都成功
  - **Success**: 所有验证通过，功能正常

---

## Phase 3: 文档和清理

- [x] 3.1 运行代码检查

  - **File**: -
  - **Purpose**: 确保代码符合 Go 规范
  - **Requirements**: 代码质量要求
  - **Commands**:
    ```bash
    cd ttpos-bmp/app/ttpos-erp
    
    # 格式化代码
    go fmt ./internal/logic/setup/...
    
    # 静态检查
    go vet ./internal/logic/setup/...
    
    # 运行测试
    go test ./internal/logic/setup/... -v -cover
    ```
  - **Success**: 
    - go fmt 无变更
    - go vet 无错误
    - 测试通过，覆盖率 ≥ 80%

- [ ] 3.2 更新相关文档（如需要）

  - **File**: `ttpos-bmp/README.md` 或相关文档
  - **Purpose**: 说明初始化流程的幂等性特性
  - **Requirements**: 文档要求
  - **Content**:
    ```markdown
    ## ERPNext 文档初始化
    
    初始化流程支持幂等性操作：
    - JSON 文件不包含 `name` 字段：创建新文档
    - JSON 文件包含 `name` 字段：更新已有文档
    - 可以安全地重复执行初始化流程
    
    示例：
    
    ```json
    // 创建新文档
    {
        "doctype": "Custom Field",
        "label": "我的字段"
    }
    
    // 更新已有文档
    {
        "name": "Custom Field-Item-my_field",
        "doctype": "Custom Field",
        "label": "我的字段（已更新）"
    }
    ```
    ```
  - **Success**: 文档更新完成（如需要）

- [ ] 3.3 清理测试数据

  - **File**: -
  - **Purpose**: 清理手动测试产生的测试数据
  - **Requirements**: 环境清理
  - **Steps**:
    ```bash
    # 删除测试 JSON 文件
    rm -rf /tmp/erp-test
    
    # 在 ERPNext 系统中删除测试创建的自定义字段
    # （通过 ERPNext Web 界面或 API）
    ```
  - **Success**: 测试数据已清理

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率 ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
  - [x] Requirement 1: 智能判断创建或更新
  - [x] Requirement 2: 错误处理和日志记录
  - [x] Requirement 3: 向后兼容性
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] 方法注释已更新
- [ ] README 已更新（如需要）
- [ ] 测试文档已更新

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `.cursor/rules/go-bmp.mdc`
- [ ] 所有注释使用中文
- [ ] 不使用 panic，返回 error

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-erp-init-documents-update/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-erp-init-documents-update/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-erp-init-documents-update/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-erp-init-documents-update/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-erp-init-documents-update/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: Task 1.1 - 修改 initDocumentsFromDir 方法
2. **阅读需求**: 查看 requirements.md 中的 Requirement 1, 2, 3
3. **查看复用**: 检查现有的 service.Document() 方法
4. **使用 AI**: 复制 Prompt 模板，生成代码
5. **实现代码**: 按照设计文档实现
6. **运行检查**: `go fmt`, `go vet`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：关键代码位置

### 需要修改的文件

```
ttpos-bmp/app/ttpos-erp/internal/logic/setup/setup.go
├── 第 477-484 行: initDocumentsFromDir 方法注释（需要更新）
└── 第 520-527 行: 文档创建逻辑（需要修改）
```

### 需要调用的服务方法

```go
// 创建文档
service.Document().Create(ctx context.Context, docType string, data interface{}) (*gjson.Json, error)

// 更新文档
service.Document().Update(ctx context.Context, docType string, name string, data interface{}) (*gjson.Json, error)
```

### 测试文件位置

```
ttpos-bmp/app/ttpos-erp/internal/logic/setup/setup_test.go
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2026-01/2026-01-04.md`
- 在实现过程中若遇到问题或总结出经验，请记录 Episode。

---

**模板版本**: v1.0.0  
**最后更新**: 2026-01-04  
**维护者**: rikugun

