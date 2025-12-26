# Get Item Stock By Bin Service 任务分解

> 本文档定义 获取商品按货位分组的库存信息 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 7
**已完成**: 7
**进行中**: -
**完成率**: 100%

---

## Phase 1: Protobuf 定义和生成

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 定义 Protobuf 消息

  - File: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto`
  - Purpose: 定义 GetItemStockByBin 的请求和响应消息
  - Requirements: 1.1, 1.2
  - Leverage: 现有 Protobuf 定义: `ttpos-bmp/app/ttpos-erp/manifest/protobuf/stock/stock.proto`
  - Prompt: Role: gRPC Developer | Task: 在 stock.proto 中新增 GetItemStockByBin 相关的消息定义 | Context: 使用 proto3 语法，包含请求参数 Warehouse 和 item_code，响应包含指定的字段列表 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc | Success: Protobuf 消息定义完成，语法正确

- [x] 1.2 生成 gRPC 代码

  - File: -
  - Purpose: 生成 Go gRPC 代码
  - Requirements: 1.1
  - Leverage: Task 1.1 的 Protobuf 定义
  - Command: `cd ttpos-bmp/app/ttpos-erp && make dao`
  - Success: gRPC Go 代码生成成功，无编译错误

- [x] 1.3 更新服务注册

  - File: `ttpos-bmp/app/ttpos-erp/manifest/config/config.yaml`
  - Purpose: 注册新的 gRPC 服务到 Nacos
  - Requirements: 1.1
  - Leverage: 现有配置: `ttpos-bmp/app/ttpos-erp/manifest/config/config.yaml`
  - Success: 服务配置更新完成

---

## Phase 2: 核心实现

### Logic 层

- [x] 2.1 在 stock_bin.go 中新增 GetItemStockByBin 方法

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_bin.go`
  - Purpose: 在现有文件中实现获取按货位分组库存的核心业务逻辑
  - Requirements: 1.1, 1.2, 1.4
  - Leverage: 现有文件: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_bin.go`，复用 GetBin 方法逻辑
  - Prompt: Role: Go Developer with BMP microservice expertise | Task: 在 stock_bin.go 中新增 GetItemStockByBin 方法 | Context: 查询货位信息，然后根据货位查询库存台账数据，返回指定字段 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，错误处理完整 | Success: 方法实现完成，业务逻辑正确

- [x] 2.2 编写新增方法的单元测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_bin_test.go`
  - Purpose: 确保新增方法的业务逻辑正确
  - Requirements: 1.1, 1.3, 1.4
  - Leverage: 现有测试: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock_bin_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 GetItemStockByBin 方法编写单元测试，覆盖率 ≥ 70% | Context: 测试参数验证，测试货位查询，测试库存台账查询，测试错误处理 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

### Controller 层

- [x] 2.3 实现 gRPC Controller

  - File: `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock.go`
  - Purpose: 实现 gRPC 接口控制器
  - Requirements: 1.1, 1.5
  - Leverage: 现有 Controller: `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock.go`
  - Prompt: Role: Go Developer with gRPC expertise | Task: 在 stock.go 中添加 GetItemStockByBin gRPC 方法 | Context: 调用对应的 Logic 方法，处理请求和响应 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: gRPC Controller 实现完成

- [x] 2.4 编写 Controller 单元测试

  - File: `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock_test.go`
  - Purpose: 测试 gRPC 接口
  - Requirements: 1.1, 1.5
  - Leverage: 现有测试: `ttpos-bmp/app/ttpos-erp/internal/controller/rpc/stock/stock_test.go`
  - Prompt: Role: QA Engineer specializing in gRPC testing | Task: 为 GetItemStockByBin Controller 编写单元测试 | Context: 测试 gRPC 调用，测试参数传递，测试响应格式 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 所有 gRPC 测试通过

---

## Phase 3: 集成测试和优化

- [x] 3.1 集成测试

  - File: `ttpos-bmp/app/ttpos-erp/test/integration/stock_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有验收标准
  - Leverage: 现有集成测试: `ttpos-bmp/app/ttpos-erp/test/integration/`
  - Prompt: Role: QA Automation Engineer | Task: 实现 GetItemStockByBin 端到端集成测试 | Context: 测试 gRPC 调用流程，测试数据一致性，测试 Bin 服务集成 | Restrictions: 测试真实场景 | Success: 集成测试通过

- [x] 3.2 性能优化

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock.go`
  - Purpose: 优化查询性能
  - Requirements: 性能要求
  - Leverage: 现有性能优化代码
  - Prompt: Role: Go Performance Engineer | Task: 优化 GetItemStockByBin 查询性能 | Context: 添加数据库索引，优化 SQL 查询，实现缓存策略 | Restrictions: 响应时间 < 200ms | Success: 性能测试通过

- [x] 3.3 错误处理完善

  - File: `ttpos-bmp/app/ttpos-erp/internal/logic/stock/stock.go`
  - Purpose: 完善错误处理逻辑
  - Requirements: 1.3, 1.4
  - Leverage: 现有错误处理代码
  - Prompt: Role: Go Developer with error handling expertise | Task: 完善 GetItemStockByBin 的错误处理 | Context: 参数验证，Bin 服务异常处理，数据库异常处理 | Restrictions: 返回明确的错误信息，不泄露敏感信息 | Success: 错误处理完善，测试覆盖所有错误场景

---

## Phase 4: 文档更新

- [x] 4.1 更新 API 文档

  - File: `docs/shared/api/bmp-api.md`
  - Purpose: 记录新的 gRPC 接口
  - Requirements: 1.1, 1.5
  - Leverage: 现有 API 文档: `docs/shared/api/bmp-api.md`
  - Prompt: Role: Technical Writer | Task: 在 BMP API 文档中添加 GetItemStockByBin 接口说明 | Context: 包含请求参数，响应格式，使用示例 | Restrictions: 文档准确完整 | Success: API 文档更新完成

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Logic: ≥ 70%
  - Controller: ≥ 70%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新
- [ ] CHANGELOG.md 已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-bmp.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-bmp-get-item-stock-by-bin/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-bmp-get-item-stock-by-bin/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-bmp-get-item-stock-by-bin/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-bmp-get-item-stock-by-bin/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-bmp-get-item-stock-by-bin/tasks.md)" | bc
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

### BMP 微服务开发

```
Role: Go Developer specializing in BMP microservice

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc

Restrictions:
- 禁止修改 dao/entity/do/ 目录（自动生成）
- gRPC 服务必须注册到 Nacos
- 使用 GoFrame 2.x
- 遵循项目分层架构

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70%
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70%

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- 集成场景测试

Restrictions:
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
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
**最后更新**: 2025-12-26
**维护者**: BMP 开发组
