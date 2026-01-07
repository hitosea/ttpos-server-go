# 门店多渠道激活服务 任务分解

> 本文档定义门店多渠道激活服务的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 19  
**已完成**: 13  
**进行中**: -  
**完成率**: 68%

---

## Phase 1: Protobuf 定义和代码生成（0.5 天）

- [x] 1.1 定义 Shop 服务 Protobuf

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/shop/shop.proto`
  - Purpose: 定义 Shop 服务的 gRPC 接口
  - Requirements: 1.1, 2.1
  - Leverage: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/grab/grab.proto` - 参考 Grab 服务定义
  - Prompt: Role: gRPC Developer | Task: 创建 shop.proto，定义 ActivateShop 和 GetShopProviderCfg 服务 | Context: 使用 proto3 语法，导入 takeout_api.proto，返回 takeout.ApiResponse | Restrictions: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc，字段使用 snake_case，添加中文注释 | Success: Protobuf 定义完成，编译无错误

- [x] 1.2 定义请求和响应消息

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/shop/shop.proto`
  - Purpose: 定义 ActivateShopReq/Resp, GetShopProviderCfgReq/Resp
  - Requirements: 1.1, 2.1
  - Leverage: `ttpos-bmp/app/ttpos-takeout/manifest/protobuf/grab/grab.proto` (35-36, 20-23)
  - Success: 消息定义完整，字段类型正确

- [x] 1.3 生成 gRPC 代码

  - File: -
  - Purpose: 生成 gRPC Go 代码
  - Requirements: 1.1, 2.1
  - Leverage: Task 1.1-1.2 的 Protobuf 定义
  - Command: `cd ttpos-bmp/app/ttpos-takeout && gf gen pb`
  - Success: 生成 api/shop/*.pb.go, api/shop/*_grpc.pb.go

- [x] 1.4 生成 Service 接口

  - File: -
  - Purpose: 生成 Service 接口代码
  - Requirements: 1.1, 2.1
  - Leverage: Task 1.3 生成的代码
  - Command: `cd ttpos-bmp/app/ttpos-takeout && gf gen service`
  - Success: 生成 internal/service/ 接口文件

---

## Phase 2: Logic 层实现（0.5 天）

### 常量定义

- [x] 2.1 添加 ProviderLineman 常量

  - File: `ttpos-bmp/app/ttpos-takeout/internal/consts/consts.go`
  - Purpose: 定义 LINE MAN 渠道常量
  - Requirements: 3.1
  - Leverage: 现有常量定义: `ProviderGrab`, `ProviderSkootar`
  - Prompt: Role: Go Developer | Task: 在 ProviderName 类型中添加 ProviderLineman 常量 | Context: 值为 "lineman"，保持与现有常量一致的风格 | Restrictions: 遵循 Go 常量命名规范 | Success: 常量定义完成

### ShopActivate Logic

- [x] 2.2 创建 ShopActivate Logic 文件

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/shop_activate/shop_activate.go`
  - Purpose: 实现门店激活的多渠道路由逻辑
  - Requirements: 1.1-1.8
  - Leverage: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_self_serve/grab_self_serve.go` - 参考 Logic 结构
  - Prompt: Role: Go Developer specializing in GoFrame | Task: 创建 ShopActivate Logic，实现 ActivateShop 方法 | Context: 注册到 service，使用 switch-case 实现渠道路由 | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc，使用 g.Log() 记录日志 | Success: Logic 文件创建成功

- [x] 2.3 实现 activateLineman 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/shop_activate/shop_activate.go`
  - Purpose: 实现 LINE MAN 渠道激活逻辑
  - Requirements: 1.3
  - Leverage: `service.ShopProviderCfg().UpsertShopProviderCfg` - 创建配置记录
  - Prompt: Role: Go Developer | Task: 实现 activateLineman，调用 UpsertShopProviderCfg 创建 INACTIVE 状态记录 | Context: 参数: shopUUID, ProviderLineman, "", ProviderShopStatusInactive | Restrictions: 错误处理使用 gerror.Wrap | Success: Lineman 激活逻辑正确，返回 ActivateShopResp

- [x] 2.4 实现 activateGrab 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/shop_activate/shop_activate.go`
  - Purpose: 实现 Grab 渠道激活逻辑
  - Requirements: 1.4
  - Leverage: `service.GrabSelfServe().CreateSelfServeJourney` - 创建自助激活链接
  - Prompt: Role: Go Developer | Task: 实现 activateGrab，调用 CreateSelfServeJourney，返回激活链接 | Context: 传递 ProviderName, ShopUuid, RequestId | Restrictions: 错误处理，记录详细日志 | Success: Grab 激活逻辑正确，返回 self_serve_url

### ShopProviderCfg Logic

- [x] 2.5 扩展 GetShopProviderCfgForRPC 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/shop_provider_cfg/shop_provider_cfg.go`
  - Purpose: 支持查询单个或所有渠道配置
  - Requirements: 2.1-2.7
  - Leverage: 现有方法 `GetShopProviderCfg`, `GetShopProviderCfgByMerchantID`
  - Prompt: Role: Go Developer | Task: 实现 GetShopProviderCfgForRPC，支持 provider_name 可选参数 | Context: 为空时查询所有渠道（lineman, grab），不为空时查询指定渠道，不存在时返回 INACTIVE | Restrictions: 遵循 GoFrame 规范，使用 g.Log() 记录警告 | Success: 查询逻辑正确，返回 GetShopProviderCfgResp

---

## Phase 3: Controller 层和测试（0.5 天）

### Controller 层

- [x] 3.1 创建 Shop Controller 文件

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/shop/shop.go`
  - Purpose: 实现 Shop gRPC 接口
  - Requirements: 1.1, 2.1
  - Leverage: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/grab/grab.go` - 参考 Grab Controller
  - Prompt: Role: gRPC Developer with GoFrame expertise | Task: 创建 Shop Controller，实现 UnimplementedShopServer | Context: 定义 Register 方法注册到 gRPC Server | Restrictions: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc | Success: Controller 文件创建成功

- [x] 3.2 实现 ActivateShop Controller

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/shop/shop_v1_activate_shop.go`
  - Purpose: 实现 ActivateShop gRPC 接口
  - Requirements: 1.1-1.8
  - Leverage: `service.ShopActivate().ActivateShop` - Logic 层方法
  - Prompt: Role: gRPC Developer | Task: 实现 ActivateShop，调用 Logic 层，返回 takeout.ApiResponse | Context: 使用 anypb.New() 包装响应数据 | Restrictions: Controller 返回 ApiResponse，错误处理返回 CodeServiceError | Success: 接口实现正确，响应格式符合规范

- [x] 3.3 实现 GetShopProviderCfg Controller

  - File: `ttpos-bmp/app/ttpos-takeout/internal/controller/rpc/shop/shop_v1_get_shop_provider_cfg.go`
  - Purpose: 实现 GetShopProviderCfg gRPC 接口
  - Requirements: 2.1-2.7
  - Leverage: `service.ShopProviderCfg().GetShopProviderCfgForRPC` - Logic 层方法
  - Prompt: Role: gRPC Developer | Task: 实现 GetShopProviderCfg，调用 Logic 层，返回 takeout.ApiResponse | Context: 使用 anypb.New() 包装响应数据 | Restrictions: Controller 返回 ApiResponse，错误处理返回 CodeServiceError | Success: 接口实现正确，响应格式符合规范

- [x] 3.4 注册 Shop gRPC 服务

  - File: `ttpos-bmp/app/ttpos-takeout/internal/boot/rpc.go`
  - Purpose: 将 Shop 服务注册到 gRPC Server
  - Requirements: 1.1, 2.1
  - Leverage: 现有服务注册代码
  - Prompt: Role: Go Developer | Task: 在 gRPC Server 初始化中调用 shop.Register(s) | Context: 在 rpc.go 中的 gRPC 注册部分添加 | Restrictions: 确保在服务启动前注册 | Success: Shop 服务成功注册

### 测试

- [ ] 3.5 编写 ShopActivate Logic 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/shop_activate/shop_activate_test.go`
  - Purpose: 测试多渠道激活逻辑
  - Requirements: 1.1-1.8
  - Leverage: 现有 Logic 测试文件
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 编写 ShopActivate 单元测试，覆盖率 ≥ 70% | Context: 测试 activateLineman, activateGrab, 错误场景 | Restrictions: 使用 GoFrame 测试框架 | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 3.6 编写 GetShopProviderCfgForRPC 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/shop_provider_cfg/shop_provider_cfg_test.go`
  - Purpose: 测试查询单个/所有渠道逻辑
  - Requirements: 2.1-2.7
  - Leverage: 现有 Logic 测试文件
  - Prompt: Role: QA Engineer | Task: 编写 GetShopProviderCfgForRPC 单元测试 | Context: 测试查询单个渠道，查询所有渠道，配置不存在 | Restrictions: 使用 GoFrame 测试框架 | Success: 测试覆盖所有场景，测试通过

- [ ] 3.7 编写 gRPC 接口集成测试

  - File: `ttpos-bmp/app/ttpos-takeout/test/integration/shop_test.go`
  - Purpose: 测试 Shop gRPC 接口
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试文件
  - Prompt: Role: QA Automation Engineer | Task: 编写 Shop gRPC 接口集成测试 | Context: 测试 ActivateShop（lineman, grab），测试 GetShopProviderCfg（单个，所有） | Restrictions: 使用真实数据库，清理测试数据 | Success: 所有集成测试通过

---

## Phase 4: 联调和文档（0.5 天）

- [ ] 4.1 本地环境测试

  - File: -
  - Purpose: 在本地环境验证功能
  - Requirements: 所有功能需求
  - Leverage: Task 3.5-3.7 的测试脚本
  - Success: 本地测试通过，功能正常

- [ ] 4.2 前后端联调

  - File: -
  - Purpose: 与前端对接，验证 API 调用
  - Requirements: 所有功能需求
  - Leverage: Postman/gRPC 客户端
  - Success: 前后端联调通过，API 正常工作

- [ ] 4.3 性能测试

  - File: -
  - Purpose: 测试 gRPC 响应时间
  - Requirements: 性能要求
  - Leverage: ghz（gRPC 性能测试工具）
  - Success: 响应时间 < 200ms

- [ ] 4.4 完善技术文档

  - File: `docs/shared/specs/active/story-takeout-shop-activate/design.md`
  - Purpose: 更新技术文档，补充实际实现细节
  - Requirements: 文档要求
  - Leverage: 实际代码实现
  - Success: 文档与代码一致

- [ ] 4.5 代码审查

  - File: -
  - Purpose: 团队代码审查
  - Requirements: 代码质量要求
  - Leverage: GitHub Pull Request
  - Success: 代码审查通过，无阻塞性问题

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Logic: ≥ 70%
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] design.md 已更新（如有实现调整）
- [ ] CHANGELOG.md 已更新（如有版本发布）

### 规范遵循

- [ ] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [ ] 遵循 `ttpos-bmp/.cursor/rules/proto-rules.mdc`
- [ ] 遵循 `.cursor/rules/go-ttpos-takeout`
- [ ] 遵循 `.cursor/rules/api.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-takeout-shop-activate/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-shop-activate/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-takeout-shop-activate/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-shop-activate/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-takeout-shop-activate/tasks.md)" | bc
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

### Go BMP 开发

```
Role: Go Developer specializing in GoFrame

Task: {具体任务描述，引用 Requirements}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc, ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 禁止修改 dao/entity/do/service 目录（自动生成）
- Controller 返回 takeout.ApiResponse
- Logic 返回具体业务数据类型
- 使用 g.Log() 记录日志，日志描述使用中文
- 使用 gerror 进行错误处理
- 注册 Service 到 internal/service

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70% (Logic)
```

### Protobuf 开发

```
Role: gRPC Developer with Protobuf expertise

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/proto-rules.mdc

Restrictions:
- 使用 proto3 语法
- 字段名使用 snake_case
- 请求消息以 Req 结尾
- 响应消息以 Resp 结尾
- 添加详细的中文注释
- 服务返回 takeout.ApiResponse

Success Criteria:
- {成功标准1}
- Protobuf 定义完整
- 编译无错误
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70% (Logic)

Test Cases Required:
- 正常场景测试
- 异常场景测试（参数错误、渠道不支持）
- 边界条件测试
- 多渠道路由测试

Restrictions:
- 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc
- 使用 GoFrame 测试框架
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
- 活动日志：`docs/team/activities/rikugun/2026-01/2026-01-07.md`
- 在执行任务过程中若总结出经验或规避策略，请记录 Episode，并在 tasks.md 尾部更新 `Related Episode`。

---

**模板版本**: v1.0.0  
**最后更新**: 2026-01-07  
**维护者**: rikugun

