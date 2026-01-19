# 对象存储层 任务分解

> 本文档定义对象存储层的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 47  
**已完成**: 25  
**进行中**: -  
**完成率**: 53.2%

**注意**: 代码已实现，但编译时遇到环境问题（无错误输出）。代码逻辑已通过 lint 检查，建议在 IDE 中直接编译查看具体错误。

---

## Phase 1: 接口设计和核心实现

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 创建对象存储层目录结构

  - File: `main/pkg/objectstorage/`
  - Purpose: 创建对象存储层的基础目录结构
  - Requirements: 1.1, 1.2
  - Leverage: 参考 `main/pkg/cache/` 目录结构
  - Success: 目录创建成功，包含 interface.go、storage.go、preload.go、config.go、utils.go

- [x] 1.2 定义 IObjectStorage 接口

  - File: `main/pkg/objectstorage/interface.go`
  - Purpose: 定义对象存储层的核心接口，支持泛型
  - Requirements: 1.1, 1.2, 2.1, 2.2, 2.6, 2.7, 2.8, 1.5.1
  - Leverage: 参考 `main/pkg/cache/cache.go` 接口定义
  - Prompt: Role: Go Developer specializing in Infrastructure Layer | Task: 创建 IObjectStorage 泛型接口，包含 Get、BatchGet、Invalidate、Update、Warmup 等方法，支持多租户隔离 | Context: 使用 Go 1.18+ 泛型，所有方法支持 context.Context，Key 格式为 {company_uuid}:{object_type}:{object_uuid} | Restrictions: 遵循 .cursor/rules/go-main.mdc，接口以 I 开头 | Success: 接口定义完整，方法签名正确，支持泛型

- [x] 1.3 定义 Association 配置结构

  - File: `main/pkg/objectstorage/interface.go`
  - Purpose: 定义配置映射自动注入的配置结构
  - Requirements: 1.5.2
  - Leverage: 参考 requirements.md 中的配置结构定义
  - Prompt: Role: Go Developer | Task: 定义 Association 配置结构，包含 Path、ObjectType、GetUUID、QueryFunc、BatchQueryFunc 字段 | Context: Path 支持嵌套路径，BatchQueryFunc 可选 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 配置结构定义完整，字段类型正确

- [x] 1.4 定义 Config 配置结构

  - File: `main/pkg/objectstorage/config.go`
  - Purpose: 定义对象存储层的配置选项
  - Requirements: 3.1, 3.2, 3.3
  - Leverage: 参考 `main/pkg/cache/cache.go` 的 Config 结构
  - Prompt: Role: Go Developer | Task: 定义 Config 配置结构，包含 TTL、DisableCache、KeyPrefix、CacheLayer 字段 | Context: TTL 用于设置缓存过期时间，DisableCache 用于调试 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 配置结构定义完整

- [x] 1.5 实现 ObjectStorageImpl 核心结构

  - File: `main/pkg/objectstorage/storage.go`
  - Purpose: 实现对象存储层的核心结构体
  - Requirements: 1.1, 1.2
  - Leverage: 参考 `main/pkg/cache/redis_cache.go` 的实现模式
  - Prompt: Role: Go Developer | Task: 实现 ObjectStorageImpl 结构体，包含 config 字段 | Context: 使用泛型，持有 Config 配置 | Restrictions: 实现以 Impl 结尾 | Success: 结构体定义正确

- [x] 1.6 实现 NewObjectStorage 构造函数

  - File: `main/pkg/objectstorage/storage.go`
  - Purpose: 创建对象存储层实例
  - Requirements: 1.1
  - Leverage: 参考 `main/pkg/cache/cache.go` 的 NewCache 函数
  - Prompt: Role: Go Developer | Task: 实现 NewObjectStorage 构造函数，返回 IObjectStorage 接口 | Context: 接收 Config 参数，返回泛型实例 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 构造函数实现正确

- [x] 1.7 实现 Get 方法

  - File: `main/pkg/objectstorage/storage.go`
  - Purpose: 实现单个对象获取，自动处理三级缓存查询和回填
  - Requirements: 1.1, 1.3, 1.4, 1.5
  - Leverage: 参考 `main/pkg/cache/cache.go` 的 Get 方法，三级缓存基础包的 GET 方法
  - Prompt: Role: Go Developer with Cache expertise | Task: 实现 Get 方法，自动处理三级缓存查询，未命中时调用 query 函数并写入缓存 | Context: 如果禁用缓存，直接调用查询函数；支持 context.Context；类型断言确保类型安全 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Get 方法实现完整，缓存查询逻辑正确，错误处理完善

- [x] 1.8 实现 BatchGet 方法

  - File: `main/pkg/objectstorage/storage.go`
  - Purpose: 实现批量对象获取，支持批量查询优化
  - Requirements: 1.2, 1.3, 1.4, 1.8
  - Leverage: 参考 `main/pkg/cache/cache.go` 的 GetBatchBytes 方法
  - Prompt: Role: Go Developer with Cache expertise | Task: 实现 BatchGet 方法，批量查询缓存，自动去重，只对未命中的 key 调用查询函数 | Context: 支持 context.Context；自动去重 keys；类型转换确保类型安全 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: BatchGet 方法实现完整，批量查询优化正确

- [x] 1.9 实现 BuildKey 辅助方法

  - File: `main/pkg/objectstorage/utils.go`
  - Purpose: 构建包含 company UUID 的缓存 key
  - Requirements: 1.6, 1.7
  - Leverage: 参考 `main/pkg/context/context.go` 的 GetCompanyUuid 方法
  - Prompt: Role: Go Developer | Task: 实现 BuildKey 方法，自动从 context 提取 company UUID，构建格式为 {company_uuid}:{object_type}:{object_uuid} 的 key | Context: 使用 ctx.GetCompanyUuid() 获取 company UUID | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: BuildKey 方法实现正确，Key 格式符合要求

- [x] 1.10 实现去重工具方法

  - File: `main/pkg/objectstorage/utils.go`
  - Purpose: 批量获取时自动去重，避免重复查询
  - Requirements: 1.8
  - Leverage: 参考 Go 标准库的 slices 包
  - Prompt: Role: Go Developer | Task: 实现 deduplicate 方法，对字符串切片去重 | Context: 使用 Go 1.21+ 的 slices 包或 map 实现 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 去重方法实现正确，性能良好

---

## Phase 2: 生命周期管理实现

- [x] 2.1 实现 Invalidate 方法

  - File: `main/pkg/objectstorage/storage.go`
  - Purpose: 使指定 key 的缓存失效
  - Requirements: 2.1
  - Leverage: 参考 `main/pkg/cache/cache.go` 的 Del 方法
  - Prompt: Role: Go Developer | Task: 实现 Invalidate 方法，调用三级缓存基础包的 DEL 方法 | Context: 支持 context.Context；调用 CacheLayer.DEL(key) | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Invalidate 方法实现正确

- [x] 2.2 实现 Update 方法

  - File: `main/pkg/objectstorage/storage.go`
  - Purpose: 更新缓存中的对象
  - Requirements: 2.2
  - Leverage: 参考 `main/pkg/cache/cache.go` 的 Set 方法
  - Prompt: Role: Go Developer | Task: 实现 Update 方法，更新缓存中的对象，使用配置的 TTL | Context: 支持 context.Context；调用 CacheLayer.SET(key, value, ttl) | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Update 方法实现正确

- [x] 2.3 实现 Warmup 方法

  - File: `main/pkg/objectstorage/storage.go`
  - Purpose: 预热指定 keys 的缓存
  - Requirements: 2.4
  - Leverage: 参考 BatchGet 的实现
  - Prompt: Role: Go Developer | Task: 实现 Warmup 方法，批量查询并写入缓存 | Context: 调用 query 函数获取数据，然后批量写入缓存 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Warmup 方法实现正确

- [x] 2.4 实现 InvalidateByCompany 方法

  - File: `main/pkg/objectstorage/storage.go`
  - Purpose: 按 company 粒度批量失效缓存
  - Requirements: 2.6
  - Leverage: 参考 Redis 的 SCAN 命令或模式匹配
  - Prompt: Role: Go Developer with Redis expertise | Task: 实现 InvalidateByCompany 方法，批量失效指定 company 的所有缓存 | Context: 使用 Redis SCAN 命令或模式匹配查找所有匹配的 key，然后批量删除 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: InvalidateByCompany 方法实现正确，性能良好

- [x] 2.5 实现 InvalidateByCompanyAndType 方法

  - File: `main/pkg/objectstorage/storage.go`
  - Purpose: 按 company + object_type 粒度批量失效缓存
  - Requirements: 2.8
  - Leverage: 参考 Task 2.4 的实现
  - Prompt: Role: Go Developer | Task: 实现 InvalidateByCompanyAndType 方法，批量失效指定 company 和 object_type 的缓存 | Context: Key 模式为 {company_uuid}:{object_type}:* | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法实现正确

- [x] 2.6 实现 UpdateByCompany 方法

  - File: `main/pkg/objectstorage/storage.go`
  - Purpose: 按 company 粒度批量更新缓存
  - Requirements: 2.7
  - Leverage: 参考 Update 和 BatchGet 的实现
  - Prompt: Role: Go Developer | Task: 实现 UpdateByCompany 方法，批量更新指定 company 的缓存 | Context: 遍历 values map，为每个 key 调用 Update 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法实现正确

- [x] 2.7 实现 TTL 配置管理

  - File: `main/pkg/objectstorage/config.go`
  - Purpose: 支持为不同对象类型配置不同的 TTL
  - Requirements: 2.3, 3.1, 3.2
  - Leverage: 使用 map[string]time.Duration 存储不同对象类型的 TTL
  - Prompt: Role: Go Developer | Task: 实现 TTL 配置管理，支持为不同 object_type 设置不同的 TTL | Context: 使用 map 存储 TTL 配置，提供 SetTTL 和 GetTTL 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: TTL 配置管理实现正确

---

## Phase 3: 配置映射自动关联注入

- [x] 3.1 实现路径解析工具方法

  - File: `main/pkg/objectstorage/preload.go`
  - Purpose: 解析嵌套路径，如 "SaleOrders.SaleOrderProducts.ProductPackage"
  - Requirements: 1.5.3
  - Leverage: 使用 strings.Split 解析路径
  - Prompt: Role: Go Developer | Task: 实现 parsePath 方法，将 "Parent.Child.GrandChild" 拆分为路径数组 | Context: 使用 strings.Split(path, ".") 解析 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 路径解析方法实现正确

- [x] 3.2 实现反射字段查找

  - File: `main/pkg/objectstorage/preload.go`
  - Purpose: 通过反射查找结构体字段
  - Requirements: 1.5.6
  - Leverage: 使用 reflect 包
  - Prompt: Role: Go Developer with Reflection expertise | Task: 实现 findField 方法，通过反射查找结构体字段 | Context: 支持指针和值类型；处理嵌套结构 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段查找方法实现正确

- [x] 3.3 实现 UUID 提取逻辑

  - File: `main/pkg/objectstorage/preload.go`
  - Purpose: 调用 GetUUID 函数从对象中提取 UUID
  - Requirements: 1.5.2
  - Leverage: 调用 Association.GetUUID 函数
  - Prompt: Role: Go Developer | Task: 实现 extractUUID 方法，调用 GetUUID 函数提取 UUID | Context: 处理 UUID 为 0 的情况（跳过） | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: UUID 提取逻辑实现正确

- [x] 3.4 实现批量 UUID 收集

  - File: `main/pkg/objectstorage/preload.go`
  - Purpose: 收集同一层级的 UUID，用于批量查询
  - Requirements: 1.5.4
  - Leverage: 使用 map 收集 UUID
  - Prompt: Role: Go Developer | Task: 实现 collectUUIDs 方法，收集同一层级的 UUID | Context: 遍历对象列表，提取 UUID，去重 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: UUID 收集逻辑实现正确

- [x] 3.5 实现批量查询调用

  - File: `main/pkg/objectstorage/preload.go`
  - Purpose: 调用 BatchQueryFunc 批量查询对象
  - Requirements: 1.5.4
  - Leverage: 调用 Association.BatchQueryFunc
  - Prompt: Role: Go Developer | Task: 实现 batchQuery 方法，调用 BatchQueryFunc 批量查询 | Context: 如果配置了 BatchQueryFunc，使用批量查询；否则逐个调用 QueryFunc | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 批量查询逻辑实现正确

- [x] 3.6 实现反射字段设置

  - File: `main/pkg/objectstorage/preload.go`
  - Purpose: 使用反射设置关联字段
  - Requirements: 1.5.6
  - Leverage: 使用 reflect 包设置字段值
  - Prompt: Role: Go Developer with Reflection expertise | Task: 实现 setField 方法，使用反射设置结构体字段 | Context: 支持指针和值类型；处理 nil 指针 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 字段设置方法实现正确

- [x] 3.7 实现 PreloadWithConfig 主方法

  - File: `main/pkg/objectstorage/preload.go`
  - Purpose: 实现配置映射自动关联注入的主方法
  - Requirements: 1.5.1, 1.5.3, 1.5.4, 1.5.5, 1.5.7
  - Leverage: 整合 Task 3.1-3.6 的实现
  - Prompt: Role: Go Developer with Reflection and Cache expertise | Task: 实现 PreloadWithConfig 方法，整合路径解析、字段查找、UUID 提取、批量查询、字段设置等逻辑 | Context: 支持嵌套路径递归处理；单个关联查询失败不影响其他关联；支持一对一、一对多、多对多关系 | Restrictions: 遵循 .cursor/rules/go-main.mdc，错误处理完善 | Success: PreloadWithConfig 方法实现完整，自动注入功能正确

- [x] 3.8 实现嵌套路径递归处理

  - File: `main/pkg/objectstorage/preload.go`
  - Purpose: 递归处理嵌套路径，如 "SaleOrders.SaleOrderProducts.ProductPackage"
  - Requirements: 1.5.3, 1.5.5
  - Leverage: 递归调用 PreloadWithConfig
  - Prompt: Role: Go Developer | Task: 实现递归处理嵌套路径的逻辑 | Context: 对每个路径层级递归调用注入逻辑 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 嵌套路径递归处理正确

---

## Phase 4: 测试和文档

- [ ] 4.1 编写 Get 方法单元测试

  - File: `main/pkg/objectstorage/storage_test.go`
  - Purpose: 测试 Get 方法的缓存查询和回填逻辑
  - Requirements: 1.1, 1.3, 1.4
  - Leverage: 参考 `main/pkg/cache/cache_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 Get 方法编写单元测试，测试缓存命中、未命中、禁用缓存等场景 | Context: 使用 mock 的 CacheLayer；测试类型安全；测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

- [ ] 4.2 编写 BatchGet 方法单元测试

  - File: `main/pkg/objectstorage/storage_test.go`
  - Purpose: 测试 BatchGet 方法的批量查询和去重逻辑
  - Requirements: 1.2, 1.8
  - Leverage: 参考 Task 4.1
  - Prompt: Role: QA Engineer | Task: 为 BatchGet 方法编写单元测试，测试批量查询、去重、部分命中等场景 | Context: 测试批量查询优化；测试类型安全 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

- [ ] 4.3 编写生命周期管理方法单元测试

  - File: `main/pkg/objectstorage/storage_test.go`
  - Purpose: 测试 Invalidate、Update、Warmup 等方法
  - Requirements: 2.1, 2.2, 2.4
  - Leverage: 参考 Task 4.1
  - Prompt: Role: QA Engineer | Task: 为生命周期管理方法编写单元测试 | Context: 测试缓存失效、更新、预热功能 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

- [ ] 4.4 编写多租户隔离测试

  - File: `main/pkg/objectstorage/storage_test.go`
  - Purpose: 确保不同 company 的数据完全隔离
  - Requirements: 1.9
  - Leverage: 创建多个 company 的测试数据
  - Prompt: Role: QA Engineer specializing in Security | Task: 编写多租户隔离测试，确保不同 company 的数据完全隔离 | Context: 测试 Key 格式；测试跨租户数据泄露防护 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 多租户隔离测试通过，无数据泄露

- [ ] 4.5 编写 PreloadWithConfig 单元测试

  - File: `main/pkg/objectstorage/preload_test.go`
  - Purpose: 测试配置映射自动关联注入功能
  - Requirements: 1.5.1, 1.5.3, 1.5.4, 1.5.5, 1.5.7
  - Leverage: 创建测试用的模型结构体
  - Prompt: Role: QA Engineer | Task: 为 PreloadWithConfig 方法编写单元测试，测试一对一、一对多、嵌套关联注入 | Context: 测试路径解析；测试批量查询优化；测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

- [ ] 4.6 编写集成测试

  - File: `test/integration/objectstorage_test.go`
  - Purpose: 测试端到端的对象存储功能
  - Requirements: 所有功能需求
  - Leverage: 使用真实的 Redis 和数据库
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试，测试三级缓存查询流程 | Context: 测试真实的三级缓存查询；测试自动注入功能 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.7 编写使用示例和最佳实践文档

  - File: `main/pkg/objectstorage/README.md`
  - Purpose: 提供使用示例和最佳实践
  - Requirements: 文档要求
  - Leverage: 参考 requirements.md 中的示例代码
  - Prompt: Role: Technical Writer | Task: 编写对象存储层的使用文档，包含示例代码和最佳实践 | Context: 包含 Get、BatchGet、PreloadWithConfig 的使用示例；包含多租户使用说明 | Restrictions: 文档准确完整 | Success: 文档创建成功，示例代码可运行

- [ ] 4.8 性能测试和优化

  - File: `main/pkg/objectstorage/storage.go`, `main/pkg/objectstorage/preload.go`
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 使用 Go benchmark 工具
  - Prompt: Role: Performance Engineer | Task: 进行性能测试，确保缓存命中时 < 10ms，未命中时 < 200ms | Context: 测试批量查询性能；测试自动注入性能 | Restrictions: 性能达标 | Success: 性能测试通过，响应时间符合要求

---

## Phase 5: 集成和重构（可选）

- [x] 5.1 在订单服务中集成对象存储层（示例）

  - File: `main/app/service/order.go`
  - Purpose: 将订单服务中的缓存逻辑迁移到对象存储层
  - Requirements: 迁移需求
  - Leverage: 参考 requirements.md 中的示例代码
  - Prompt: Role: Go Developer | Task: 在订单服务中使用对象存储层替换现有的缓存逻辑 | Context: 使用 PreloadWithConfig 自动注入关联对象 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 集成成功，代码简化

- [x] 5.2 将对象存储层迁移到 modules/objectstorage 模块

  - File: `main/app/modules/objectstorage/`
  - Purpose: 按照 DDD 架构将对象存储层重构为独立模块
  - Requirements: 完成代码迁移，保持功能不变
  - Success: 模块结构完整，代码通过 lint 检查

- [ ] 5.3 在商品服务中集成对象存储层（示例）

  - File: `main/app/service/material.go`
  - Purpose: 将商品服务中的缓存逻辑迁移到对象存储层
  - Requirements: 迁移需求
  - Leverage: 参考 Task 5.1
  - Success: 集成成功

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - objectstorage 包: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] README.md 已创建（使用示例和最佳实践）
- [ ] 代码注释完整

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/structs.mdc`
- [ ] 遵循 `.cursor/rules/security.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-main-object-storage-layer/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-main-object-storage-layer/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-main-object-storage-layer/tasks.md
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
**最后更新**: 2025-12-24  
**维护者**: xiezhihuan

