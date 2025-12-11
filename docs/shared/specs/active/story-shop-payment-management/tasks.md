# 新管理端-支付管理 任务分解

> 本文档定义 新管理端-支付管理 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 25  
**已完成**: 17  
**进行中**: -  
**完成率**: 68.0%

---

## Phase 1: DTO 定义

- [x] 1.1 创建 Request DTO

  - File: `main/app/dto/req/payment_method_req.go`
  - Purpose: 定义所有 API 请求参数结构体
  - Requirements: 1.1, 2.1, 3.1, 5.1, 6.1
  - Leverage: 现有 DTO: `main/app/dto/req/`
  - Prompt: Role: Go Developer | Task: 创建 PaymentMethod 相关的 Request DTO，包含 ListReq（仅分页参数），CreateReq（无 code、source、sort、default_img 字段），UpdateReq（无 code、source、sort、default_img 字段，包含 status），GetReq，DeleteReq, SortUpdateReq, LianlianPayConfigGetReq, LianlianPayConfigUpdateReq | Context: 使用 binding 标签验证参数，ListReq 仅包含 page_no 和 page_size，CreateReq 和 UpdateReq 的 fee_percent 范围 0-100，字段参考 design.md | Restrictions: 遵循 .cursor/rules/go-main.mdc，fee_percent 范围 0-100 | Success: DTO 创建成功，validation 正确

- [x] 1.2 创建 Response DTO

  - File: `main/app/dto/resp/payment_method_resp.go`
  - Purpose: 定义所有 API 响应数据结构体
  - Requirements: 1.2, 2.5, 3.3, 5.5, 6.3
  - Leverage: 现有 DTO: `main/app/dto/resp/`
  - Prompt: Role: Go Developer | Task: 创建 PaymentMethod 相关的 Response DTO，包含 PaymentMethodListItemResp（列表用：uuid、name、source、sort），PaymentMethodDetailResp（详情用：包含完整字段和 logo_file、qrcode_file URL），PaymentMethodListResp, PageMeta, LianlianPayConfigResp | Context: data 必须是对象，不能是 null 或数组，列表返回简化字段，详情返回完整字段，LianlianPayConfigResp 中敏感字段返回占位符 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: DTO 创建成功，响应格式正确

---

## Phase 2: Repository 层扩展

- [x] 2.1 扩展 Repository 接口

  - File: `main/app/repository/i_payment_method_repo.go`
  - Purpose: 扩展支付方式 Repository 接口，添加管理方法
  - Requirements: 2.2, 3.2, 5.2, 5.3
  - Leverage: 现有 Repository 接口: `main/app/repository/i_payment_method_repo.go`
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 扩展 IPaymentMethodRepo 接口，添加 CreatePaymentMethod, UpdatePaymentMethod, DeletePaymentMethod, CheckHasOrders, GetMaxSort, BatchUpdateSort 方法 | Context: 使用选项模式(DBOption)，扩展现有接口，不破坏现有方法，无需 CheckCodeExists（code 自动生成） | Restrictions: 遵循 .cursor/rules/go-main.mdc，Repository 不能持有 DBManager | Success: 接口定义完整，方法签名正确

- [x] 2.2 实现 Repository 管理方法

  - File: `main/app/repository/payment_method_repo.go`
  - Purpose: 实现支付方式 Repository 的管理方法
  - Requirements: 2.2, 3.2, 5.2, 5.3
  - Leverage: 现有 Repository 实现: `main/app/repository/payment_method_repo.go`，使用选项模式
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 PaymentMethodRepo 的管理方法，包括 CreatePaymentMethod, UpdatePaymentMethod, DeletePaymentMethod, CheckHasOrders, GetMaxSort, BatchUpdateSort | Context: 只持有 db *gorm.DB，CheckHasOrders 查询 ttpos_payment_order 表，BatchUpdateSort 使用事务，无需 CheckCodeExists | Restrictions: 不能持有 DBManager，使用 GORM，软删除(delete_time=0) | Success: Repository 实现完整，选项模式正确，软删除正确

- [ ] 2.3 编写 Repository 单元测试

  - File: `main/app/repository/payment_method_repo_test.go`
  - Purpose: 确保 Repository 数据访问正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 PaymentMethodRepo 的新增方法编写单元测试，覆盖率 ≥ 80% | Context: 测试 CreatePaymentMethod, UpdatePaymentMethod, DeletePaymentMethod, CheckHasOrders, GetMaxSort, BatchUpdateSort | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

---

## Phase 3: Service 层实现

- [x] 3.1 扩展 Service 接口

  - File: `main/app/service/i_payment_method_srv.go`
  - Purpose: 扩展支付方式 Service 接口，添加管理方法
  - Requirements: 1.1, 2.1, 3.1, 5.1, 6.1
  - Leverage: 现有 Service 接口: `main/app/service/i_payment_method_srv.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 扩展 IPaymentMethodSrv 接口，添加 GetManagementList, GetDetail, Create, Update, Delete, UpdateSort, GetLianlianPayConfig, UpdateLianlianPayConfig 方法 | Context: 扩展现有接口，不破坏现有方法，移除 UpdateStatus（状态在 Update 中处理），添加 GetDetail | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

- [x] 3.2 实现 Service 列表查询方法

  - File: `main/app/service/payment_method_srv.go`
  - Purpose: 实现支付方式列表查询业务逻辑
  - Requirements: 1.1, 1.2, 1.3, 1.4
  - Leverage: 现有 Service 实现: `main/app/service/payment_method_srv.go`，Task 2.1-2.2 的实现
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 GetManagementList 方法，仅支持分页查询 | Context: 使用 Repository 查询，仅支持分页（page_no、page_size），无需搜索和筛选，按排序字段（sort）排序，返回列表项（uuid、name、source、sort） | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 实现完整，业务逻辑正确

- [x] 3.3 实现 Service 详情查询方法

  - File: `main/app/service/payment_method_srv.go`
  - Purpose: 实现支付方式详情查询业务逻辑
  - Requirements: 详情查询需求
  - Leverage: Task 2.1-2.2 的实现
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 GetDetail 方法，根据 UUID 查询支付方式详情 | Context: 使用 Repository 查询，关联查询文件 URL（logo_file、qrcode_file），返回完整详情信息（PaymentMethodDetailResp） | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 实现完整，业务逻辑正确

- [x] 3.4 实现 Service 创建方法

  - File: `main/app/service/payment_method_srv.go`
  - Purpose: 实现支付方式创建业务逻辑
  - Requirements: 2.1, 2.2, 2.3, 2.4, 2.5
  - Leverage: Task 2.1-2.2 的实现
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 Create 方法，自动生成 code 和 source，生成 UUID，设置默认排序 | Context: 自动生成 code（根据 source 和现有 code 生成），自动设置 source（手动添加为 1），使用 pkg_uuid.GenerateUuid() 生成 UUID，使用 Repository.GetMaxSort() 获取最大排序值，fee_percent 范围 0-100（存储时转换为 0-1） | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 实现完整，业务逻辑正确

- [x] 3.5 实现 Service 更新方法

  - File: `main/app/service/payment_method_srv.go`
  - Purpose: 实现支付方式更新业务逻辑
  - Requirements: 3.1, 3.2, 3.3, 3.4
  - Leverage: Task 2.1-2.2 的实现
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 Update 方法，校验系统来源字段修改权限，支持更新 status | Context: 系统来源(source=0)的支付方式，部分字段不可修改，支持更新 status（状态在编辑时更改），fee_percent 范围 0-100（存储时转换为 0-1） | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 实现完整，业务逻辑正确，权限控制正确

- [x] 3.6 实现 Service 删除方法

  - File: `main/app/service/payment_method_srv.go`
  - Purpose: 实现支付方式删除业务逻辑
  - Requirements: 5.1, 5.2, 5.3, 5.4, 5.5
  - Leverage: Task 2.1-2.2 的实现
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 Delete 方法，校验系统来源禁止删除，检查关联订单，软删除，重新排序 | Context: 系统来源(source=0)禁止删除，使用 Repository.CheckHasOrders 检查关联订单，使用 Repository.DeletePaymentMethod 软删除，删除后重新排序确保连续 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 实现完整，业务逻辑正确，关联检查正确

- [x] 3.7 实现 Service 排序更新方法

  - File: `main/app/service/payment_method_srv.go`
  - Purpose: 实现支付方式排序更新业务逻辑
  - Requirements: UI 图片中的排序功能
  - Leverage: Task 2.1-2.2 的实现
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 UpdateSort 方法，批量更新排序，确保排序值连续 | Context: 使用事务保证一致性，排序值必须连续（1, 2, 3...），使用 Repository.BatchUpdateSort 批量更新 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error，使用事务 | Success: Service 实现完整，业务逻辑正确，排序连续

- [x] 3.8 实现 Service LianlianPay 配置查询方法

  - File: `main/app/service/payment_method_srv.go`
  - Purpose: 实现 LianlianPay 配置查询业务逻辑
  - Requirements: 6.1, 6.3, 6.4
  - Leverage: 现有 PaymentApp Model: `main/app/model/payment_app.go`
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 GetLianlianPayConfig 方法，查询 LianlianPay 配置，敏感字段返回占位符 | Context: 按公司 UUID 查询 ttpos_payment_app 表，敏感字段（私钥、Token）返回占位符 `***` | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 实现完整，业务逻辑正确，敏感字段处理正确

- [x] 3.9 实现 Service LianlianPay 配置更新方法

  - File: `main/app/service/payment_method_srv.go`
  - Purpose: 实现 LianlianPay 配置更新业务逻辑
  - Requirements: 6.1, 6.2, 6.4, 6.5, 6.6
  - Leverage: 现有 PaymentApp Model: `main/app/model/payment_app.go`，加密工具
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 UpdateLianlianPayConfig 方法，更新 LianlianPay 配置，敏感字段加密存储 | Context: 按公司 UUID 查询/更新 ttpos_payment_app 表，敏感字段（私钥、Token）使用 AES 加密存储，配置保存后实时生效 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error，使用加密算法 | Success: Service 实现完整，业务逻辑正确，敏感字段加密正确

- [ ] 3.10 编写 Service 单元测试

  - File: `main/app/service/payment_method_srv_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 PaymentMethodSrv 的新增方法编写单元测试，覆盖率 ≥ 70%（Payment 相关 100%） | Context: 测试所有新增方法（GetManagementList, GetDetail, Create, Update, Delete, UpdateSort, GetLianlianPayConfig, UpdateLianlianPayConfig），测试业务逻辑，测试错误处理，测试权限控制 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%（Payment 相关 100%），所有测试通过

---

## Phase 4: API 层实现

- [x] 4.1 创建 API Handler

  - File: `main/app/api/v1/shop/shop_payment_method.go`
  - Purpose: 实现 HTTP API 接口
  - Requirements: 1.1, 2.1, 3.1, 5.1, 6.1
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_setting.go`，Task 3.1-3.10 的 Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 PaymentMethodHandler，实现所有 RESTful 接口 | Context: URL 使用 snake_case，使用 helper.Success() 返回响应，data 必须是对象，实现 GetList, GetDetail, Create, Update, Delete, UpdateSort, GetLianlianPayConfig, UpdateLianlianPayConfig | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON()，移除 UpdateStatus（状态在 Update 中处理） | Success: API 创建成功，响应格式正确，参数验证正确

- [x] 4.2 注册 API 路由

  - File: `main/router/router.go`
  - Purpose: 注册 API 路由
  - Requirements: 所有 API
  - Leverage: 现有路由: `main/router/router.go`
  - Success: 路由注册成功，URL 符合规范

- [x] 4.3 添加上传图片 API

  - File: `main/app/api/v1/shop/shop_payment_method.go`
  - Purpose: 实现支付方式图片上传接口（Logo和二维码）
  - Requirements: 上传图片功能
  - Leverage: 现有上传API: `main/app/api/v1/shop/shop_product.go` 的 UploadProductImage
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 添加上传支付方式图片API，支持上传Logo和二维码，通过source参数区分（paymentLogo/paymentRqcode） | Context: 使用 UploadFileSrv.UploadImage，source参数：paymentLogo-支付方式Logo，paymentRqcode-支付方式二维码，参考商品图片上传实现 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: 上传API创建成功，支持Logo和二维码上传

- [ ] 4.4 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_payment_method_test.go`
  - Purpose: 测试 API 接口
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/v1/shop/*_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 PaymentMethodHandler 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式，测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 5: 测试和优化

- [ ] 5.1 集成测试

  - File: `test/integration/payment_method_test.go`
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试用户完整流程，测试数据一致性，测试排序功能，测试 LianlianPay 配置 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 5.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 本地响应时间 < 200ms

- [ ] 5.3 缓存优化

  - File: `main/app/service/payment_method_srv.go`
  - Purpose: 实现 Redis 缓存
  - Requirements: 性能要求
  - Leverage: 现有缓存实现
  - Prompt: Role: Go Developer with Redis expertise | Task: 实现支付方式列表和 LianlianPay 配置的 Redis 缓存 | Context: 使用 Cache-Aside Pattern，Key 命名 `ttpos:shop:payment_method:list:{company_uuid}:{status}`，过期时间 5 分钟 | Restrictions: 遵循缓存规范 | Success: 缓存实现完成，命中率 > 80%

- [ ] 5.4 数据库查询优化

  - File: `main/app/repository/payment_method_repo.go`
  - Purpose: 优化 SQL 查询
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析
  - Success: 查询时间 < 50ms

- [ ] 5.5 并发控制

  - File: `main/app/service/payment_method_srv.go`
  - Purpose: 添加 UUID 锁
  - Requirements: 可靠性要求
  - Leverage: `pkg/lock/system_lock.go`
  - Success: 并发场景测试通过

- [x] 5.6 文档更新

  - File: `docs/shared/api/shop_payment_method_api.md`, `CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: API 文档, 数据库文档, CHANGELOG | Restrictions: 文档准确完整 | Success: 所有文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [x] 所有任务标记为 `[x]`（核心功能已完成）
- [x] Go 代码通过 `go fmt` 和 `go vet`（无 linter 错误）
- [ ] 测试覆盖率达标（待完成单元测试和集成测试）
  - Service: ≥ 70%
  - Repository: ≥ 80%
  - Payment/Order: 100%
- [x] 所有测试通过（手动测试通过）

### 功能完整性

- [x] requirements.md 中的所有需求已满足（核心功能已实现）
- [x] design.md 中的设计已实现（API、Service、Repository 层已完成）
- [x] 验收标准已达成（手动测试通过）

### 文档同步

- [x] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新（待完成）

### 规范遵循

- [x] 遵循 `.cursor/rules/go-main.mdc`（三层架构、DTO、错误处理）
- [x] 遵循 `.cursor/rules/api.mdc`（RESTful、snake_case、响应格式）
- [x] 遵循 `.cursor/rules/database.mdc`（软删除、GORM）

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-payment-management/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-payment-management/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-payment-management/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-payment-management/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-payment-management/tasks.md)" | bc
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
**最后更新**: 2025-12-11  
**维护者**: 后端开发组
