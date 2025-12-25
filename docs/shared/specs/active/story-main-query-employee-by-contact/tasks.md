# 根据邮箱或手机号查询员工接口 任务分解

> 本文档定义根据邮箱或手机号查询员工接口的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 12  
**已完成**: 8  
**进行中**: -  
**完成率**: 66.7%

---

## Phase 1: DTO 定义

### 任务模板说明

每个任务包含以下信息：

- **File**: 需要修改的文件路径
- **Purpose**: 任务目的（一句话说明为什么要做）
- **Requirements**: 关联的需求编号（如: 1.1, 2.3）
- **Leverage**: 可复用的现有代码路径
- **Prompt**: AI 执行提示模板（可选）

---

- [x] 1.1 创建 Request DTO

  - File: `main/app/dto/req/staff.go`
  - Purpose: 定义查询员工接口的请求参数结构
  - Requirements: 1.1, 1.2, 1.3
  - Leverage: 现有 DTO: `main/app/dto/req/staff.go`，参考 `SearchStaffByKeywordReq`
  - Prompt: Role: Go Developer | Task: 在 staff.go 中新增 QueryStaffByContactReq 结构体，包含 keyword (string, 可选) 和 limit (int, 可选，默认 20，最大 20) | Context: 使用 form 标签，limit 使用 binding:"omitempty,min=1,max=20" | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，validation 正确

- [x] 1.2 创建 Response DTO

  - File: `main/app/dto/resp/saas_staff.go`
  - Purpose: 定义查询员工接口的响应数据结构
  - Requirements: 1.4
  - Leverage: 现有 DTO: `main/app/dto/resp/saas_staff.go`，参考 `SearchStaffResp`
  - Prompt: Role: Go Developer | Task: 在 saas_staff.go 中新增 QueryStaffByContactItem 和 QueryStaffByContactResp 结构体 | Context: QueryStaffByContactItem 包含 uuid, real_name, email, phone；QueryStaffByContactResp 包含 list 字段 | Restrictions: data 必须是对象，不能是 null 或数组 | Success: DTO 创建成功，响应格式正确

---

## Phase 2: Repository 层实现

- [x] 2.1 扩展 Repository 接口

  - File: `main/app/repository/saas_staff.go`
  - Purpose: 在 ISaasStaffRepo 接口中新增 QueryByKeyword 方法
  - Requirements: 1.2, 1.3
  - Leverage: 现有接口: `main/app/repository/saas_staff.go`，参考 `GetByEmailOrPhone`
  - Prompt: Role: Go Developer specializing in Repository Pattern | Task: 在 ISaasStaffRepo 接口中新增 QueryByKeyword 方法签名 | Context: 方法签名: QueryByKeyword(keyword string, staffUuids []uint64, limit int) ([]*model.SaasStaff, error) | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

- [x] 2.2 实现 Repository 方法（支持模糊搜索）

  - File: `main/app/repository/saas_staff.go`
  - Purpose: 实现 QueryByKeyword 方法，支持邮箱和手机号的模糊搜索
  - Requirements: 1.2, 1.3, 1.5
  - Leverage: 现有实现: `main/app/repository/saas_staff.go`，参考 `GetByEmailOrPhone`，使用 GORM LIKE 查询
  - Prompt: Role: Go Developer with GORM expertise | Task: 实现 QueryByKeyword 方法，支持模糊搜索邮箱和手机号 | Context: 判断 keyword 是否包含 @ 来区分邮箱和手机号，使用 LOWER(email) LIKE 和 phone LIKE，过滤 staffUuids，限制 limit | Restrictions: 使用参数化查询防止 SQL 注入，软删除过滤 delete_time=0 | Success: Repository 实现完整，模糊搜索正确，性能优化（使用索引）

- [ ] 2.3 编写 Repository 单元测试

  - File: `main/app/repository/saas_staff_test.go`
  - Purpose: 确保 Repository 数据访问正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/repository/*_repo_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 QueryByKeyword 方法编写单元测试，覆盖率 ≥ 80% | Context: 测试邮箱模糊搜索、手机号模糊搜索、空关键词、结果限制、UUID 过滤 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 80%，所有测试通过

---

## Phase 3: Service 层实现

- [x] 3.1 扩展 Service 接口

  - File: `main/app/service/saas_staff.go`
  - Purpose: 在 ISaasStaffSrv 接口中新增 QueryStaffByContact 方法
  - Requirements: 1.1, 1.6, 1.7
  - Leverage: 现有接口: `main/app/service/saas_staff.go`，参考 `SearchStaff`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 在 ISaasStaffSrv 接口中新增 QueryStaffByContact 方法签名 | Context: 方法签名: QueryStaffByContact(ctx context.Context, req req.QueryStaffByContactReq) (*resp.QueryStaffByContactResp, error) | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

- [x] 3.2 实现 Service 业务逻辑（权限过滤）

  - File: `main/app/service/saas_staff.go`
  - Purpose: 实现 QueryStaffByContact 方法，包含授权员工过滤和门店可见性过滤
  - Requirements: 1.6, 1.7
  - Leverage: 现有实现: `main/app/service/saas_staff.go`，参考 `SearchStaff` 方法的权限过滤逻辑，使用 `setting.NewSrv` 获取业务设置，使用 `CompanyRepo.GetVisibleCompanyList` 获取可见门店
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 QueryStaffByContact 方法，包含授权员工过滤和门店可见性过滤 | Context: 获取业务设置中的 discount_authorized_staff_ids 和 refund_authorized_staff_ids，合并去重；获取当前门店可见范围内的门店列表；获取可见门店下的授权员工；调用 Repository.QueryByKeyword 查询 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error，查询失败返回空列表 | Success: Service 实现完整，权限过滤正确，门店可见性过滤正确

- [ ] 3.3 编写 Service 单元测试

  - File: `main/app/service/saas_staff_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 QueryStaffByContact 方法编写单元测试，覆盖率 ≥ 70% | Context: 测试授权员工过滤、门店可见性过滤、关键词搜索、空结果处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

---

## Phase 4: API 层实现

- [x] 4.1 实现 API Handler

  - File: `main/app/api/v1/cashier/cashier_order.go`, `main/app/api/v1/assistant/assistant_order.go`
  - Purpose: 实现 QueryStaffByContact HTTP 接口（收银机端和助手端）
  - Requirements: 1.1, 1.8, 1.9
  - Leverage: 现有 API: `main/app/api/v1/cashier/cashier_order.go`，参考 `VerifyPassword` 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 在 OrderHandler 中新增 QueryStaffByContact 方法，实现 RESTful 接口 | Context: URL 使用 snake_case，使用 helper.Success() 返回响应，data 必须是对象，添加 Swagger 注释 | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON() | Success: API 创建成功，响应格式正确，参数验证正确

- [x] 4.2 注册 API 路由

  - File: `main/app/api/v1/cashier/cashier_order.go`, `main/app/api/v1/assistant/assistant_order.go`
  - Purpose: 注册 QueryStaffByContact 路由（收银机端和助手端）
  - Requirements: 1.1
  - Leverage: 现有路由: `main/app/api/v1/cashier/cashier_order.go`，参考 `RegisterOrderHandlers` 函数
  - Prompt: Role: Go Developer | Task: 在 RegisterOrderHandlers 函数中注册 QueryStaffByContact 路由 | Context: 路由路径: GET /cashier/order/query-staff-by-contact 和 GET /assistant/order/query-staff-by-contact，需要认证中间件 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 路由注册成功

- [ ] 4.3 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_staff_test.go`（如不存在则创建）
  - Purpose: 测试 API 接口
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 QueryStaffByContact API 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式，测试权限控制 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 5: 测试和优化

- [ ] 5.1 集成测试

  - File: `test/integration/staff_query_test.go`（如不存在则创建）
  - Purpose: 测试端到端功能
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试用户完整流程：查询授权员工 → 返回结果 → 前端展示 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 5.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 本地响应时间 < 200ms

- [ ] 5.3 数据库查询优化

  - File: `main/app/repository/saas_staff.go`
  - Purpose: 优化 SQL 查询
  - Requirements: 性能要求
  - Leverage: EXPLAIN 分析，确保使用索引
  - Success: 查询时间 < 50ms

- [ ] 5.4 文档更新

  - File: `docs/shared/api/staff_api.md`（如存在），`CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: API 文档, CHANGELOG | Restrictions: 文档准确完整 | Success: 所有文档已更新

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

- [ ] API 文档已更新（Swagger 注释）
- [ ] CHANGELOG.md 已更新
- [ ] 设计文档已更新（如有调整）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-main-query-employee-by-contact/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-main-query-employee-by-contact/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-main-query-employee-by-contact/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-main-query-employee-by-contact/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-main-query-employee-by-contact/tasks.md)" | bc
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
- URL 使用 snake_case
- data 字段必须是对象
- 不使用 panic，返回 error
- 使用 errors.WithMessage 包装错误

Success Criteria:
- {成功标准1}
- 代码通过 go fmt 和 go vet
- 测试覆盖率 ≥ 70% (Service) 或 ≥ 80% (Repository)
```

### 测试工程师

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Coverage target: ≥ 70% (Service) 或 ≥ 80% (Repository)

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试
- 并发场景测试（如适用）

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
**最后更新**: 2025-12-23  
**维护者**: 后端开发组

