# 统一外卖订单数据结构字段 任务分解

> 本文档定义在 GetOrderInfo 接口中增加 order_data 字段的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 10  
**已完成**: 7  
**进行中**: -  
**完成率**: 70%

---

## Phase 1: Protobuf 定义扩展（0.5 天）

- [x] 1.1 修改 Protobuf 定义增加 order_data 字段

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/order/order.proto`
  - Purpose: 在 GetOrderInfoResp 中增加 order_data 字段（转换后的统一订单数据）
  - Requirements: Requirement 1.1, 1.2, 1.3
  - Leverage: 现有 Protobuf 定义
  - Details:
    - 在 `GetOrderInfoResp` 消息中增加 `string order_data = 6;`
    - 添加注释：`// 转换后的统一订单数据（TakeoutOrder JSON）`
    - 保留所有现有字段（向后兼容）
  - Prompt: Role: gRPC Developer | Task: 修改 order.proto，在 GetOrderInfoResp 中增加 order_data 字段 | Context: 字段编号为 6，类型为 string，用于存储 TakeoutOrder JSON | Restrictions: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc，保持向后兼容 | Success: Protobuf 定义修改完成，字段注释清晰

- [x] 1.2 重新生成 Protobuf Go 代码

  - File: -
  - Purpose: 生成包含新字段的 Go 代码
  - Requirements: Requirement 1.4
  - Leverage: Task 1.1 的 Protobuf 定义
  - Command: `cd ttpos-bmp/app/ttpos-takeout && make proto`
  - Success: 代码生成成功，编译无错误

- [x] 1.3 验证编译通过

  - File: -
  - Purpose: 确保修改不破坏现有编译
  - Requirements: Requirement 1.4
  - Leverage: Task 1.2 生成的代码
  - Command: `cd ttpos-bmp/app/ttpos-takeout && go build`
  - Success: 编译通过，无错误和警告

---

## Phase 2: Logic 层实现（0.5 天）

- [x] 2.1 修改 GetOrderInfo 方法添加数据转换逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/order/order.go`（或相应的 logic 文件）
  - Purpose: 在 GetOrderInfo 中调用 converter 转换 raw_data 为 TakeoutOrder
  - Requirements: Requirement 2.1, 2.2, 2.3, 2.4, 2.5, 2.6
  - Leverage: 
    - 现有 GetOrderInfo 实现
    - `utility.ConvertGrabToTakeoutOrder` (已实现)
    - `utility.ConvertLinemanToTakeoutOrder` (已实现)
  - Details:
    - 导入 `ttpos-bmp/app/ttpos-takeout/utility` 包
    - 导入 `ttpos-api/ttpos-takeout/message` 包
    - 根据 `provider_name` 选择对应的转换函数
    - 处理转换错误（优雅降级，返回空字符串）
    - 记录错误日志（包含 orderId、provider、error）
    - 序列化 `TakeoutOrder` 为 JSON 字符串
    - 填充到 `GetOrderInfoResp.OrderData` 字段
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 修改 GetOrderInfo 方法，增加数据转换逻辑 | Context: 使用 utility.ConvertGrabToTakeoutOrder 和 utility.ConvertLinemanToTakeoutOrder，序列化为 JSON，填充 order_data 字段 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，转换失败不影响接口返回，使用 g.Log() 记录日志 | Success: 逻辑实现完成，转换逻辑正确，错误处理合理

- [x] 2.2 添加性能监控日志

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/order/order.go`
  - Purpose: 记录数据转换耗时，便于性能分析
  - Requirements: Requirement 4.1
  - Leverage: Task 2.1 的实现
  - Details:
    - 在转换前记录开始时间
    - 在转换后计算耗时
    - 使用 `g.Log().Debugf()` 记录转换耗时
  - Success: 性能日志添加完成

---

## Phase 3: 测试和验证（0.5 天）

- [x] 3.1 编写/更新 Converter 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/utility/takeout_converter_test.go`
  - Purpose: 确保 Grab 和 Lineman 转换逻辑正确，覆盖率 100%
  - Requirements: Requirement 3.1, 3.2, 3.3, 3.4
  - Leverage: 现有 converter 测试
  - Details:
    - 测试 Grab 订单转换（完整订单、缺失字段、null 值）
    - 测试 Lineman 订单转换（完整订单、缺失字段、null 值）
    - 验证 JSON 格式可正确反序列化
    - 验证核心字段一致性（orderID、商品、价格）
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 Converter 编写/更新单元测试，覆盖率 100% | Context: 测试 Grab 和 Lineman 转换，测试边界情况，验证数据一致性 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 测试覆盖率 100%，所有测试通过

- [ ] 3.2 编写 GetOrderInfo 集成测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/order/order_test.go`（或新建测试文件）
  - Purpose: 测试 GetOrderInfo 接口完整流程
  - Requirements: Requirement 2.1, 2.2, 2.3, 2.4, 2.5
  - Leverage: Task 2.1 的实现
  - Details:
    - 测试 Grab 订单：验证 order_data 正确返回
    - 测试 Lineman 订单：验证 order_data 正确返回
    - 测试转换失败场景：验证 order_data 为空，接口仍正常返回
    - 测试不支持的平台：验证 order_data 为空
  - Prompt: Role: QA Engineer specializing in integration testing | Task: 为 GetOrderInfo 编写集成测试 | Context: 测试 Grab/Lineman 两种场景，测试转换失败场景 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 所有集成测试通过

- [ ] 3.3 执行性能测试

  - File: `ttpos-bmp/app/ttpos-takeout/utility/takeout_converter_test.go`（或新建 benchmark 文件）
  - Purpose: 验证数据转换性能达标（< 50ms）
  - Requirements: Requirement 4.1, 4.2
  - Leverage: Go Benchmark
  - Details:
    - 编写 Benchmark 测试：`BenchmarkConvertGrabToTakeoutOrder`
    - 编写 Benchmark 测试：`BenchmarkConvertLinemanToTakeoutOrder`
    - 测试大订单（100+ 商品）的转换性能
  - Command: `cd ttpos-bmp/app/ttpos-takeout && go test -bench=. -benchmem ./utility/`
  - Success: 数据转换耗时 < 50ms (P99)

- [ ] 3.4 手动测试验证

  - File: -
  - Purpose: 手动验证真实订单数据转换正确性
  - Requirements: 所有功能需求
  - Leverage: 测试环境真实订单数据
  - Details:
    - 使用 gRPC 客户端调用 GetOrderInfo
    - 验证 Grab 真实订单的 order_data
    - 验证 Lineman 真实订单的 order_data
    - 对比 raw_data 和 order_data 的核心字段
  - Success: 手动测试验证通过，数据正确

---

## Phase 4: 文档和代码审查（0.5 天）

- [ ] 4.1 更新 TakeoutOrder 使用文档

  - File: `ttpos-api/ttpos-takeout/message/README.md`
  - Purpose: 更新文档说明 order_data 字段的使用方法
  - Requirements: 文档验收
  - Leverage: 现有 README 文档
  - Details:
    - 说明 order_data 字段的用途
    - 提供使用示例（如何解析 order_data JSON）
    - 说明与 raw_data 的区别
  - Success: 文档更新完成，说明清晰

- [ ] 4.2 代码注释完善

  - File: 所有修改的文件
  - Purpose: 确保代码注释完整且准确
  - Requirements: 文档验收
  - Leverage: 现有代码
  - Details:
    - Protobuf 字段注释完整
    - Logic 层关键逻辑有注释
    - 错误处理有清晰说明
  - Success: 代码注释完善，易于理解

- [ ] 4.3 Code Review

  - File: 所有修改的文件
  - Purpose: 确保代码质量和规范遵循
  - Requirements: 所有需求
  - Leverage: `.cursor/rules/code-review.mdc`
  - Details:
    - 检查规范遵循（go-rules.mdc, proto-rules.mdc）
    - 检查错误处理是否完整
    - 检查性能影响
    - 检查向后兼容性
  - Success: Code Review 通过，无遗留问题

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Converter: 100%
  - Logic: ≥ 70%
- [ ] 所有测试通过
- [ ] Benchmark 性能达标（< 50ms）

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成：
  - Protobuf 定义包含 order_data 字段
  - Grab 订单返回正确的 order_data
  - Lineman 订单返回正确的 order_data
  - 转换失败返回空字符串，接口正常
  - 现有调用方不受影响

### 文档同步

- [ ] README.md 已更新（ttpos-api/ttpos-takeout/message/）
- [ ] 代码注释完整（Protobuf、Logic、Utility）
- [ ] tasks.md 完成率 100%

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 向后兼容（保留 raw_data 字段）

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/tech-takeout-structured-order-data/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/tech-takeout-structured-order-data/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/tech-takeout-structured-order-data/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/tech-takeout-structured-order-data/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/tech-takeout-structured-order-data/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务（按 Phase 顺序）
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: `go fmt`, `go vet`, `go test`
7. **标记完成**: 将 `[ ]` 改为 `[x]`
8. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### Protobuf 开发

```
Role: gRPC Developer

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 使用 proto3 语法
- 字段编号连续递增
- 必须添加字段注释
- 保持向后兼容

Success Criteria:
- {成功标准1}
- Protobuf 定义正确
- 代码生成成功
```

### Go BMP 开发

```
Role: Go Developer with GoFrame expertise

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc

Restrictions:
- 必须使用 GoFrame 2.x
- 不修改 dao/entity/do/ 目录
- 使用 g.Log() 记录日志
- 不使用 panic，返回 error
- 使用 errors.Wrap 包装错误

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
- Coverage target: 100% (Converter) 或 ≥ 70% (Logic)

Test Cases Required:
- 正常场景测试（Grab、Lineman）
- 异常场景测试（转换失败、不支持的平台）
- 边界条件测试（空订单、缺失字段、null 值）
- 性能测试（Benchmark）

Restrictions:
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
- 必须包含边界情况测试

Success Criteria:
- 测试覆盖率达标
- 所有测试通过
- 边界情况已覆盖
- 性能达标（< 50ms）
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充 - 外卖订单统一模型设计经验]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-12.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**创建日期**: 2026-01-12  
**作者**: rikugun  
**维护者**: rikugun
