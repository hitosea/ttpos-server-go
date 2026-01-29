# Lineman Menu Sync V2 任务分解

> 本文档定义 Lineman Menu Sync V2 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 21  
**已完成**: 19  
**进行中**: -  
**完成率**: 90.5%

---

## Phase 1: 代码重构（0.5 天）✅ 已完成

### 目标：将 `lineman_token` 包迁移到 `lineman` 包，统一管理

- [x] 1.1 迁移文件到新目录

  - File: `internal/logic/lineman_token/*.go` → `internal/logic/lineman/*.go`
  - Purpose: 统一 Lineman 相关逻辑到一个包
  - Requirements: 重构步骤 1
  - Leverage: 现有文件: `internal/logic/lineman_token/`
  - Command:
    ```bash
    cd ttpos-bmp/app/ttpos-takeout/internal/logic
    mkdir -p lineman
    mv lineman_token/lineman_token.go lineman/token.go
    mv lineman_token/config.go lineman/token_config.go
    mv lineman_token/partner_config_loader.go lineman/partner_config_loader.go
    ```
  - Success: ✅ 文件已移动到新目录

- [x] 1.2 更新包名

  - File: `internal/logic/lineman/*.go`
  - Purpose: 更新包声明
  - Requirements: 重构步骤 2
  - Leverage: Task 1.1 迁移的文件
  - Change: `package lineman_token` → `package lineman`
  - Success: ✅ 所有文件包名已更新

- [x] 1.3 更新服务注册

  - File: `internal/logic/lineman/token.go`
  - Purpose: 更新服务注册名称
  - Requirements: 重构步骤 2
  - Leverage: Task 1.2
  - Change:
    ```go
    func init() {
    -   service.RegisterLinemanToken(New())
    +   service.RegisterLineman(New())
    }
    ```
  - Success: ✅ 服务注册已更新

- [x] 1.4 更新服务接口

  - File: `internal/service/lineman.go`
  - Purpose: 定义统一的 Lineman 服务接口
  - Requirements: 重构步骤 3
  - Leverage: 现有接口: `internal/service/*.go`
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 更新 ILineman 接口，包含 Token 管理和菜单同步方法 | Context: 保留原 lineman_token 的所有方法，新增 SyncMenu, BuildMenuPayload 方法 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 接口定义完整，方法签名正确
  - Success: ✅ `ILinemanToken` → `ILineman`，文件重命名 lineman_token.go → lineman.go

- [x] 1.5 更新所有引用

  - File: 搜索 `service.LinemanToken()` 的所有引用
  - Purpose: 将所有引用更新为 `service.Lineman()`
  - Requirements: 重构步骤 3
  - Leverage: 全局搜索
  - Command:
    ```bash
    # 搜索所有引用
    grep -r "service.LinemanToken()" ttpos-bmp/app/ttpos-takeout/internal/ --include="*.go"
    
    # 替换为 service.Lineman()
    find ttpos-bmp/app/ttpos-takeout/internal/ -name "*.go" -exec sed -i 's/service\.LinemanToken()/service.Lineman()/g' {} +
    ```
  - Success: ✅ `service.LinemanToken()` → `service.Lineman()`，1 个文件已更新

- [x] 1.6 删除旧目录

  - File: `internal/logic/lineman_token/`
  - Purpose: 清理旧代码
  - Requirements: 重构步骤 1
  - Leverage: Task 1.1-1.5 完成后
  - Command: `rm -rf internal/logic/lineman_token/`
  - Success: ✅ lineman_token 目录已删除，测试通过

---

## Phase 2: 数据映射与 Client 封装（1 天）✅ 已完成

### 2.1 创建 DTO 定义（2 小时）

- [x] 2.1.1 创建 Menu Sync Request DTO

  - File: `internal/model/dto/lineman/menu_sync_request.go`
  - Purpose: 定义 Lineman API 请求结构
  - Requirements: Requirement 1（数据映射）, design.md - 数据模型
  - Leverage: design.md 中的 DTO 定义
  - Prompt: Role: Go Developer | Task: 创建 MenuSyncRequest 及相关结构体，定义 Lineman API 请求格式 | Context: 包含 MenuGroup, MenuItem, Property, PropValue, NameTrans 等结构体，使用 json tag | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: DTO 创建成功，结构完整，json tag 正确
  - Success: Request DTO 创建成功

- [x] 2.1.2 创建 Menu Sync Response DTO

  - File: `internal/model/dto/lineman/menu_sync_response.go`
  - Purpose: 定义 Lineman API 响应结构
  - Requirements: Requirement 1, design.md - 数据模型
  - Leverage: design.md 中的 DTO 定义
  - Success: Response DTO 创建成功

### 2.2 实现 Data Mapper（4 小时）

- [x] 2.2.1 创建 Data Mapper 基础结构

  - File: `internal/logic/lineman/data_mapper.go`
  - Purpose: 实现 TTPOS → Lineman 数据映射
  - Requirements: Requirement 1（数据映射）
  - Leverage: design.md - Data Mapper 设计
  - Prompt: Role: Go Developer with data transformation expertise | Task: 创建 DataMapper 结构体，实现 BuildMenuPayload 主方法 | Context: 需要查询 TTPOS 菜单数据并转换为 Lineman 格式 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: DataMapper 创建成功，BuildMenuPayload 方法实现完整
  - Success: Data Mapper 基础结构创建成功

- [x] 2.2.2 实现分类映射逻辑

  - File: `internal/logic/lineman/data_mapper.go`
  - Purpose: 实现 mapCategory 方法
  - Requirements: Requirement 1.1-1.3
  - Leverage: Task 2.2.1，design.md
  - Prompt: Role: Go Developer | Task: 实现 mapCategory 方法，将 TTPOS 分类转换为 Lineman MenuGroup | Context: 生成分类ID（TTPOS-CAT-{id}），映射名称，useSellingTime 固定 false | Restrictions: 跳过无效数据，记录警告日志 | Success: 分类映射成功，ID 格式正确
  - Success: 分类映射逻辑实现完成

- [x] 2.2.3 实现商品映射逻辑

  - File: `internal/logic/lineman/data_mapper.go`
  - Purpose: 实现 mapItem 方法
  - Requirements: Requirement 1.1-1.7
  - Leverage: Task 2.2.1-2.2.2，design.md
  - Prompt: Role: Go Developer | Task: 实现 mapItem 方法，将 TTPOS 商品转换为 Lineman MenuItem | Context: 生成商品ID，映射名称/描述/价格/图片/状态，价格元→分转换，salesChannelsAvailability 固定 true | Restrictions: 处理多语言字段，使用中文降级 | Success: 商品映射成功，价格转换正确，多语言处理正确
  - Success: 商品映射逻辑实现完成

- [x] 2.2.4 实现属性映射逻辑

  - File: `internal/logic/lineman/data_mapper.go`
  - Purpose: 实现 mapProperty 方法
  - Requirements: Requirement 1.1, 1.8-1.9
  - Leverage: Task 2.2.1-2.2.3，design.md
  - Prompt: Role: Go Developer | Task: 实现 mapProperty 方法，将 TTPOS 规格/加料/属性转换为 Lineman Property | Context: 生成属性ID，映射名称/min/max，type根据selectionRangeMax判断（>1为2，=1为1），映射属性值 | Restrictions: 处理多语言字段 | Success: 属性映射成功，type 判断正确
  - Success: 属性映射逻辑实现完成

- [x] 2.2.5 实现多语言处理方法

  - File: `internal/logic/lineman/data_mapper.go`
  - Purpose: 实现 buildNameTranslation 和 buildDescTranslation 方法
  - Requirements: Requirement 4（多语言处理）
  - Leverage: Task 2.2.1-2.2.4，design.md
  - Prompt: Role: Go Developer | Task: 实现多语言字段处理，优先使用泰语/英语翻译，无翻译时使用中文降级 | Context: 两个方法：buildNameTranslation(nameCN, nameTH, nameEN), buildDescTranslation(descCN, descTH, descEN) | Restrictions: 降级处理：使用中文填充 thai 和 english | Success: 多语言处理正确，降级策略生效
  - Success: 多语言处理方法实现完成

### 2.3 实现 API Client（2 小时）

- [x] 2.3.1 创建 Menu Sync Client

  - File: `internal/client/lineman/menu_sync_client.go`
  - Purpose: 封装 Lineman Menu Sync API 调用
  - Requirements: Requirement 2（API Client）
  - Leverage: design.md - Client 层设计，现有 Token 管理
  - Prompt: Role: Go Developer with HTTP client expertise | Task: 创建 MenuSyncClient，实现 SyncMenu 方法调用 Lineman API | Context: 使用 g.Client()，PUT 请求，设置 Authorization 和 Content-Type 头，解析 JSON 响应 | Restrictions: 超时 30 秒，遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: Client 创建成功，API 调用成功，响应解析正确
  - Success: Menu Sync Client 创建成功

- [x] 2.3.2 实现重试机制

  - File: `internal/client/lineman/retry.go`
  - Purpose: 实现请求重试策略
  - Requirements: Requirement 2.4（重试）
  - Leverage: Task 2.3.1，design.md - 重试策略
  - Prompt: Role: Go Developer | Task: 实现 WithRetry 函数，支持最多 3 次重试，指数退避策略 | Context: maxRetries=3, retryDelay=2s, retryBackoff=2.0，记录重试日志 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 重试机制实现成功，指数退避正确
  - Success: 重试机制实现完成

---

## Phase 3: 同步流程与日志记录（1 天）✅ 已完成

### 3.1 实现菜单同步业务逻辑（3 小时）

- [x] 3.1.1 实现 SyncMenu 方法

  - File: `internal/logic/lineman/menu_sync.go`
  - Purpose: 实现菜单同步主流程
  - Requirements: Requirement 3（同步流程）
  - Leverage: Task 2.2, 2.3，design.md - Logic 层设计
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 SyncMenu 方法，协调 Token 获取、数据构建、API 调用、日志记录 | Context: 调用 GetAuthorizationHeader, BuildMenuPayload, Client.SyncMenu, recordMenuLog | Restrictions: 错误处理完善，记录详细日志，遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: SyncMenu 实现成功，流程正确，错误处理完善
  - Success: SyncMenu 方法实现完成

- [x] 3.1.2 实现 BuildMenuPayload 方法

  - File: `internal/logic/lineman/menu_sync.go`
  - Purpose: 构建菜单数据
  - Requirements: Requirement 1（数据映射）
  - Leverage: Task 2.2（Data Mapper），Task 3.1.1
  - Prompt: Role: Go Developer | Task: 实现 BuildMenuPayload 方法，调用 DataMapper 构建菜单数据 | Context: 创建 DataMapper 实例，调用 BuildMenuPayload | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: BuildMenuPayload 实现成功
  - Success: BuildMenuPayload 方法实现完成

- [x] 3.1.3 实现日志记录方法

  - File: `internal/logic/lineman/menu_sync.go`
  - Purpose: 记录同步日志到 takeout_menu_log 表
  - Requirements: Requirement 6（错误处理与监控）
  - Leverage: Task 3.1.1，design.md - recordMenuLog 设计，DAO: `dao.MenuLog`
  - Prompt: Role: Go Developer with DAO expertise | Task: 实现 recordMenuLog 方法，插入同步日志到 takeout_menu_log 表 | Context: 使用 dao.MenuLog.Insert，字段：uuid, merchant_id, provider_name='lineman', sync_type, request_id, status, menu_snapshot, error_code, error_msg, created_at, updated_at | Restrictions: 使用 GoFrame DAO 层，遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 日志记录成功，数据插入正确
  - Success: 日志记录方法实现完成

- [x] 3.1.4 实现配置状态更新方法

  - File: `internal/logic/lineman/menu_sync.go`
  - Purpose: 更新 takeout_shop_provider_cfg 表的状态
  - Requirements: Requirement 5（配置管理）
  - Leverage: Task 3.1.1，DAO: `dao.ShopProviderCfg`
  - Prompt: Role: Go Developer with DAO expertise | Task: 实现 updateProviderStatus 方法，更新 provider_shop_status 字段 | Context: 使用 dao.ShopProviderCfg.Update，条件：shop_uuid + provider_name='lineman' | Restrictions: 使用 GoFrame DAO 层，遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 状态更新成功
  - Success: 配置状态更新方法实现完成

### 3.2 集成配置管理（1 小时）

- [x] 3.2.1 实现获取门店配置方法

  - File: `internal/logic/lineman/menu_sync.go`
  - Purpose: 从 takeout_shop_provider_cfg 表获取 Lineman 配置
  - Requirements: Requirement 5（配置管理）
  - Leverage: DAO: `dao.ShopProviderCfg`
  - Prompt: Role: Go Developer with DAO expertise | Task: 实现 getShopConfig 方法，查询 Lineman 配置 | Context: 查询条件：shop_uuid + provider_name='lineman'，解析 provider_merchant_id 获取 partnerId 和 storeId | Restrictions: 使用 GoFrame DAO 层，遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 配置查询成功，partnerId 和 storeId 解析正确
  - Success: 获取门店配置方法实现完成

- [x] 3.2.2 实现配置验证方法

  - File: `internal/logic/lineman/menu_sync.go`
  - Purpose: 验证 Lineman 配置是否完整
  - Requirements: Requirement 5（配置管理）
  - Leverage: Task 3.2.1
  - Prompt: Role: Go Developer | Task: 实现 validateConfig 方法，验证 partnerId 和 storeId 不为空，状态为 ACTIVE | Context: 检查必需字段，检查状态 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 配置验证成功
  - Success: 配置验证方法实现完成

### 3.3 实现查询历史日志方法（可选，1 小时）

- [x] 3.3.1 实现查询同步历史方法

  - File: `internal/logic/lineman/menu_sync.go`
  - Purpose: 查询 takeout_menu_log 表的同步历史
  - Requirements: Requirement 6.4（同步历史查询）
  - Leverage: DAO: `dao.MenuLog`
  - Prompt: Role: Go Developer with DAO expertise | Task: 实现 GetSyncHistory 方法，查询最近的同步记录 | Context: 查询条件：provider_name='lineman' + merchant_id，按 created_at 倒序，分页查询 | Restrictions: 使用 GoFrame DAO 层，遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 查询成功，返回同步历史列表
  - Success: 查询同步历史方法实现完成

---

## Phase 4: 测试与联调（1 天）✅ 核心完成

### 4.1 单元测试（3 小时）

- [x] 4.1.1 测试 Data Mapper

  - File: `internal/logic/lineman/data_mapper_test.go`
  - Purpose: 测试数据映射逻辑
  - Requirements: 测试要求
  - Leverage: Task 2.2，GoFrame 测试框架
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 DataMapper 编写单元测试，覆盖率 ≥ 70% | Context: 测试 buildNameTranslation, mapCategory, mapItem, mapProperty，测试多语言降级，测试价格转换，测试状态映射 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过
  - Success: Data Mapper 测试完成

- [x] 4.1.2 测试 Menu Sync Client

  - File: `internal/client/lineman/menu_sync_client_test.go`
  - Purpose: 测试 API Client
  - Requirements: 测试要求
  - Leverage: Task 2.3，Mock HTTP Server
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 MenuSyncClient 编写单元测试，覆盖率 ≥ 80% | Context: Mock Lineman API 响应，测试成功场景，测试错误场景（4xx, 5xx），测试超时场景 | Restrictions: 使用 httptest.NewServer() Mock API，遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过
  - Success: Menu Sync Client 测试完成

- [x] 4.1.3 测试重试机制

  - File: `internal/client/lineman/retry_test.go`
  - Purpose: 测试重试策略
  - Requirements: 测试要求
  - Leverage: Task 2.3.2
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 WithRetry 编写单元测试 | Context: 测试重试次数，测试指数退避，测试最终失败 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 测试覆盖率 100%，所有测试通过
  - Success: 重试机制测试完成

### 4.2 集成测试（2 小时）

- [x] 4.2.1 端到端集成测试

  - File: `internal/logic/lineman/integration_test.go`
  - Purpose: 测试完整的同步流程
  - Requirements: 所有功能需求
  - Leverage: Task 3.1，Mock Lineman API
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试完整流程：获取配置 → 获取 Token → 构建数据 → 调用 API → 记录日志，使用 Mock API 和测试数据库 | Restrictions: 测试真实用户场景，遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: 集成测试通过
  - Success: 集成测试完成

### 4.3 真实环境联调（2 小时）

- [ ] 4.3.1 配置测试环境

  - File: `manifest/config/config.yaml`
  - Purpose: 配置 Lineman Sandbox 环境
  - Requirements: 测试要求
  - Leverage: Lineman 提供的测试环境（如有）
  - Command: 更新配置文件，设置测试环境 endpoint
  - Success: 测试环境配置完成

- [ ] 4.3.2 真实 API 测试

  - File: -
  - Purpose: 与 Lineman Sandbox 环境联调
  - Requirements: 所有功能需求
  - Leverage: Task 4.3.1，测试账号和测试门店
  - Test Cases:
    - 简单菜单同步（10 个商品）
    - 复杂菜单同步（100 个商品，含规格）
    - 边界测试（空分类，无图片商品等）
    - 错误测试（错误的 Token，错误的 partnerId）
  - Success: 所有测试用例通过，Lineman 平台显示正确

---

## Phase 5: 文档和优化（可选）

### 5.1 文档更新

- [ ] 5.1.1 更新 API 文档

  - File: `docs/shared/api/takeout_api.md`
  - Purpose: 文档化菜单同步接口（如有 HTTP 接口）
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Success: API 文档已更新

- [ ] 5.1.2 更新 CHANGELOG

  - File: `CHANGELOG.md`
  - Purpose: 记录功能变更
  - Requirements: 文档要求
  - Leverage: 参考 `.cursor/rules/version.mdc`
  - Success: CHANGELOG 已更新

### 5.2 性能优化（可选）

- [ ] 5.2.1 添加菜单数据缓存

  - File: `internal/logic/lineman/data_mapper.go`
  - Purpose: 缓存菜单数据，减少数据库查询
  - Requirements: 性能要求
  - Leverage: Redis 缓存
  - Success: 缓存实现完成，命中率 > 80%

- [ ] 5.2.2 实现增量同步（可选）

  - File: `internal/logic/lineman/menu_sync.go`
  - Purpose: 仅同步变更的数据
  - Requirements: 性能优化
  - Leverage: Task 3.1
  - Success: 增量同步实现完成

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Logic: ≥ 70%
  - Client: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新
- [ ] 代码注释完整

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`（如适用）
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-takeout-lineman-menu-sync/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-lineman-menu-sync/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-takeout-lineman-menu-sync/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-lineman-menu-sync/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-takeout-lineman-menu-sync/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看设计**: 查看 design.md 中的设计方案
4. **查看复用**: 检查 Leverage 中的可复用代码
5. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
6. **实现代码**: 按照规范实现功能
7. **运行检查**: `go fmt`, `go vet`, `go test`
8. **标记完成**: 将 `[ ]` 改为 `[x]`
9. **提交代码**: Git commit（参考 `.cursor/rules/version.mdc`）

---

## 附录：标准 Prompt 模板

### GoFrame 开发

```
Role: Go Developer with GoFrame expertise

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc

Restrictions:
- 使用 GoFrame 2.x 框架
- 禁止修改 dao/entity/do/ 目录
- Logic 层依赖 DAO 层
- 使用 g.Log() 记录日志
- 不使用 panic，返回 error

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率达标
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70% (Logic) 或 ≥ 80% (Client)

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- Mock 外部依赖

Restrictions:
- 使用 GoFrame 测试框架
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
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-08.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2026-01-08  
**维护者**: rikugun
