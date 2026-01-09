# 子店可修改总店同步物品安全库存 任务分解

> 本文档定义子店可修改总店同步物品安全库存功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 9  
**进行中**: -  
**完成率**: 60%

---

## Phase 1: DTO 和 Service 层实现

### DTO 层

- [x] 1.1 创建 Request DTO

  - File: `main/app/dto/req/material.go`
  - Purpose: 定义修改安全库存的请求参数
  - Requirements: 1.2
  - Leverage: 现有 DTO: `main/app/dto/req/material.go`，参考其他 Update 请求结构
  - Prompt: Role: Go Developer | Task: 在 material.go 中添加 MaterialUpdateSafetyStockReq 结构体 | Context: 包含 uuid (uint64, required) 和 safety_stock (*float64, 可为 null) 字段，使用 binding 标签 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 创建成功，字段定义正确

- [ ] 1.2 创建 Response DTO（可选）

  - File: `main/app/dto/resp/material.go`
  - Purpose: 定义修改安全库存的响应数据（如需要）
  - Requirements: 1.7
  - Leverage: 现有 DTO: `main/app/dto/resp/material.go`
  - Prompt: Role: Go Developer | Task: 在 material.go 中添加 MaterialUpdateSafetyStockResp 结构体（如需要） | Context: 包含 uuid 和 safety_stock 字段 | Restrictions: data 必须是对象 | Success: DTO 创建成功

### Service 层

- [x] 1.3 在 Service 接口中添加方法

  - File: `main/app/service/i_material_srv.go`
  - Purpose: 在 IMaterialSrv 接口中添加 UpdateMaterialSafetyStock 方法
  - Requirements: 1.5
  - Leverage: 现有 Service 接口: `main/app/service/i_material_srv.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 在 IMaterialSrv 接口中添加 UpdateMaterialSafetyStock(ctx context.Context, req req.MaterialUpdateSafetyStockReq) error 方法 | Context: 方法签名与现有方法保持一致 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义正确

- [x] 1.4 实现 Service 业务逻辑

  - File: `main/app/service/material.go`
  - Purpose: 实现 UpdateMaterialSafetyStock 方法，包含权限校验、业务校验和更新逻辑
  - Requirements: 1.3, 1.4, 1.5, 1.6
  - Leverage: 现有 Service 实现: `main/app/service/material.go`，参考其他 Update 方法
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 UpdateMaterialSafetyStock 方法 | Context: 1) 权限校验：只有子店账号（company.IsSubShop()）才能调用；2) 查询物品（通过 uuid）；3) 业务校验：只能修改总店同步的物品（headquarter_uuid > 0）；4) 使用 UpdateMaterialData 更新安全库存；5) 使用 errors.WithMessage 包装错误 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 实现完整，权限和业务校验正确，错误处理正确

- [ ] 1.5 编写 Service 单元测试

  - File: `main/app/service/material_srv_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 UpdateMaterialSafetyStock 方法编写单元测试，覆盖率 ≥ 70% | Context: 测试场景：1) 子店账号修改总店同步物品 - 成功；2) 非子店账号调用 - 返回错误；3) 修改非总店同步的物品 - 返回错误；4) 物品不存在 - 返回错误；5) 参数验证失败 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

---

## Phase 2: API 层实现

- [x] 2.1 创建 API Controller 方法

  - File: `main/app/api/v1/shop/shop_material.go`
  - Purpose: 在 MaterialHandler 中添加 UpdateSafetyStock 方法
  - Requirements: 1.1, 1.7
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_material.go`，参考其他 POST 方法
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 在 MaterialHandler 中添加 UpdateSafetyStock 方法 | Context: 1) 使用 c.ShouldBindJSON 绑定请求参数；2) 调用 materialSrv.UpdateMaterialSafetyStock；3) 使用 helper.Success() 返回响应，data 必须是对象；4) 使用 helper.ErrorWithDetail() 返回错误；5) 添加 Swagger 注释 | Restrictions: 遵循 .cursor/rules/api.mdc，URL 使用 snake_case | Success: API 创建成功，响应格式正确，参数验证正确

- [x] 2.2 注册 API 路由

  - File: `main/router/router.go`
  - Purpose: 注册 POST /api/v1/shop/material/update_safety_stock 路由
  - Requirements: 1.1
  - Leverage: 现有路由: `main/router/router.go`，参考其他 material 路由
  - Prompt: Role: Go Developer | Task: 在 router.go 中注册新路由 | Context: 路由路径为 /api/v1/shop/material/update_safety_stock，方法为 POST，处理器为 MaterialHandler.UpdateSafetyStock | Restrictions: 遵循路由注册规范 | Success: 路由注册成功

- [ ] 2.3 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_material_test.go`（如存在）或创建新测试文件
  - Purpose: 测试 API 接口
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 UpdateSafetyStock API 编写集成测试 | Context: 测试所有 API 接口场景，测试参数验证，测试响应格式，测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

---

## Phase 3: 同步逻辑优化

- [x] 3.1 修改 SyncMaterial 方法 - 添加安全库存保护逻辑

  - File: `main/app/service/material.go`
  - Purpose: 在同步总部物品到子店时，如果子店已有该物品（通过 uuid 匹配），保留子店的安全库存
  - Requirements: 2.2, 2.3, 2.4
  - Leverage: 现有代码: `main/app/service/material.go` 第3044-3096行
  - 实现说明: 1) 同步前获取子店中已存在的总部物品的安全库存，构建 `uuid -> *safety_stock` 映射（包括 nil 值）；2) 在创建物品时，如果子店已有该物品（通过 uuid 匹配），则保留子店的安全库存（包括 nil）；否则使用总店的安全库存；3) 统一删除后重建，在重建时保留子店已调整的安全库存（包括 nil）
  - Success: ✅ 安全库存保护逻辑正确，nil 值也能正确保留

- [ ] 3.5 编写同步逻辑单元测试

  - File: `main/app/service/material_srv_test.go`
  - Purpose: 测试同步逻辑的安全库存保护功能
  - Requirements: 测试要求
  - Leverage: 现有测试: `main/app/service/material_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 SyncMaterial 方法编写单元测试，重点测试安全库存保护逻辑 | Context: 测试场景：1) 子店已有物品且安全库存不为 nil 时同步 - 保留子店的安全库存；2) 子店已有物品但安全库存为 nil 时同步 - 保留 nil，不覆盖为总店的安全库存；3) 子店没有物品时同步 - 使用总店的安全库存 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

---

## Phase 4: 集成测试和文档

- [ ] 4.1 集成测试 - 端到端流程

  - File: `test/integration/material_safety_stock_test.go`（如存在）或创建新测试文件
  - Purpose: 测试端到端功能流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试流程：1) 子店同步总店物品；2) 子店修改安全库存；3) 再次同步总店物品；4) 验证安全库存未被覆盖 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 4.2 性能测试

  - File: -
  - Purpose: 确保性能达标
  - Requirements: 性能要求
  - Leverage: 性能测试工具（如：wrk, ab）
  - Success: 接口响应时间 < 200ms

- [ ] 4.3 文档更新

  - File: `docs/shared/api/shop_material_api.md`（如存在），`CHANGELOG.md`
  - Purpose: 确保文档与代码同步
  - Requirements: 文档要求
  - Leverage: `docs/agent/templates/api-doc-template.md`
  - Prompt: Role: Technical Writer | Task: 更新相关文档 | Context: API 文档（新增接口说明），CHANGELOG（记录新功能和修改） | Restrictions: 文档准确完整 | Success: 所有文档已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%（使用现有方法，已有测试）
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（新增接口说明）
- [ ] CHANGELOG.md 已更新
- [ ] Swagger 注释已添加

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-branch-modify-headquarters-material-safety-stock/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-branch-modify-headquarters-material-safety-stock/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-branch-modify-headquarters-material-safety-stock/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-branch-modify-headquarters-material-safety-stock/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-branch-modify-headquarters-material-safety-stock/tasks.md)" | bc
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
**最后更新**: 2025-12-08  
**维护者**: 后端开发组
