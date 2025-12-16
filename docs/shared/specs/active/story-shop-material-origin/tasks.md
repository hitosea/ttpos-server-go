# 物品管理-原产地功能 任务分解

> 本文档定义物品管理中原产地字段的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 15  
**已完成**: 10  
**进行中**: -  
**完成率**: 67%

---

## Phase 1: 数据库设计和迁移

- [x] 1.1 创建数据库迁移文件

  - File: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_origin_country_code_to_material_table.php`
  - Purpose: 在 `ttpos_material` 表中增加 `origin_country_code` 字段
  - Requirements: R1.1, R1.2
  - Leverage: 现有迁移文件: `admin/database/migrations/`，参考模板: `docs/agent/templates/database-migration-template.md`
  - Prompt: Role: Database Engineer | Task: 创建迁移文件，在 ttpos_material 表中增加 origin_country_code 字段（varchar(10), DEFAULT ''）| Context: 字段位置在 allow_substore_visible 之后，注释为"原产地国家编码（ISO 3166-1 alpha-2，如：CN, US, TH）" | Restrictions: 遵循 .cursor/rules/database.mdc，迁移前检查字段是否存在 | Success: 迁移文件创建成功，字段定义正确

- [ ] 1.2 执行数据库迁移

  - File: -
  - Purpose: 在数据库中执行迁移，添加字段
  - Requirements: R1.2
  - Leverage: Task 1.1 的迁移文件
  - Command: `cd admin && php think migrate:run`
  - Success: 迁移执行成功，字段已添加

- [x] 1.3 修改 Go Model

  - File: `main/app/model/material.go`
  - Purpose: 在 Material 结构体中增加 OriginCountryCode 字段
  - Requirements: R1.1
  - Leverage: 现有 Model: `main/app/model/material.go`，迁移文件: Task 1.1
  - Prompt: Role: Go Developer | Task: 在 Material 结构体中增加 OriginCountryCode 字段 | Context: 字段类型为 string，gorm 标签为 `gorm:"type:varchar(10);default:'';column:origin_country_code;comment:'原产地国家编码（ISO 3166-1 alpha-2）'"` | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Model 修改成功，字段映射正确

- [x] 1.4 更新 Seeds 文件

  - File: `admin/database/seeds/shop_01.sql`
  - Purpose: 在 Seeds 文件中增加 origin_country_code 字段定义
  - Requirements: R1.2
  - Leverage: 现有 Seeds: `admin/database/seeds/shop_01.sql`
  - Prompt: Role: Database Engineer | Task: 在 shop_01.sql 的 ttpos_material 表定义中增加 origin_country_code 字段 | Context: 字段位置在 allow_substore_visible 之后，类型为 varchar(10) DEFAULT '' | Restrictions: 保持 Seeds 文件格式一致 | Success: Seeds 文件更新成功

---

## Phase 2: 常量定义

- [x] 2.1 创建国家枚举常量文件

  - File: `main/app/constant/country.go`
  - Purpose: 定义197个国家/地区的枚举常量
  - Requirements: R2.1, R2.2
  - Leverage: 现有常量文件: `main/app/constant/product.go`，参考多语言结构: `main/app/model/multi_language_name.go`
  - Prompt: Role: Go Developer | Task: 创建 country.go 常量文件，定义197个国家/地区的枚举 | Context: 每个国家包含 Code（ISO 3166-1 alpha-2）和 LocaleNames（多语言名称），支持 zh, zhtw, en, ja, ko, my, th, tr, de, sv | Restrictions: 使用 AI 查询197个国家数据，遵循 ISO 3166-1 标准 | Success: 常量文件创建成功，包含197个国家
  - Note: 当前包含47个国家，后续补充完整197个国家

- [x] 2.2 实现国家查询方法

  - File: `main/app/constant/country.go`
  - Purpose: 实现 GetAllCountries 和 GetCountryByCode 方法
  - Requirements: R2.1
  - Leverage: Task 2.1 的国家数据
  - Prompt: Role: Go Developer | Task: 实现 GetAllCountries() 和 GetCountryByCode(code string) 方法 | Context: GetAllCountries 返回所有国家列表，GetCountryByCode 根据编码查找国家 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法实现成功，查询正确

---

## Phase 3: 核心实现（Go Main）

### DTO 层

- [x] 3.1 修改 Request DTO

  - File: `main/app/dto/req/material.go`
  - Purpose: 在 MaterialAddReq 和 MaterialEditReq 中增加 OriginCountryCode 字段
  - Requirements: R3.2, R3.3
  - Leverage: 现有 DTO: `main/app/dto/req/material.go`
  - Prompt: Role: Go Developer | Task: 在 MaterialAddReq 和 MaterialEditReq 中增加 OriginCountryCode 字段（string, 可选）| Context: 字段为可选字段，无需必填验证 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 修改成功，字段定义正确

- [x] 3.2 修改 Response DTO

  - File: `main/app/dto/resp/material_resp/material.go`
  - Purpose: 在 MaterialDetailResp 中增加原产地相关字段，创建 CountryItem 和 CountryListResp
  - Requirements: R3.1, R3.4
  - Leverage: 现有 DTO: `main/app/dto/resp/material_resp/material.go`，参考多语言结构: `main/app/dto/common_resp.go`
  - Prompt: Role: Go Developer | Task: 在 MaterialDetailResp 中增加 OriginCountryCode 和 OriginCountry 字段，创建 CountryItem 和 CountryListResp 结构体 | Context: OriginCountry 为可选字段（*CountryItem），CountryItem 包含 Code 和 LocaleName | Restrictions: data 必须是对象，不能是 null 或数组 | Success: DTO 修改成功，响应格式正确

### Service 层

- [x] 3.3 创建 Country Service 接口

  - File: `main/app/service/i_country_srv.go`
  - Purpose: 定义国家服务接口
  - Requirements: R3.4
  - Leverage: 现有 Service 接口: `main/app/service/i_nationality_srv.go`
  - Prompt: Role: Go Developer specializing in Service Layer | Task: 创建 ICountrySrv 接口，定义 GetList 方法 | Context: GetList 返回国家列表，无需参数 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 接口定义完整，方法签名正确

- [x] 3.4 实现 Country Service

  - File: `main/app/service/country_srv.go`
  - Purpose: 实现国家服务业务逻辑
  - Requirements: R3.4
  - Leverage: 现有 Service 实现: `main/app/service/nationality_service.go`，Task 2.1-2.2 的常量
  - Prompt: Role: Go Developer with business logic expertise | Task: 实现 countrySrv，从常量读取国家列表并返回 | Context: 无需依赖数据库，直接从 constant.GetAllCountries() 读取 | Restrictions: 遵循 .cursor/rules/go-main.mdc，不使用 panic，返回 error | Success: Service 实现完整，业务逻辑正确

- [x] 3.5 修改 Material Service - GetMaterialDetail

  - File: `main/app/service/material.go`
  - Purpose: 在 GetMaterialDetail 方法中增加原产地信息返回
  - Requirements: R3.1
  - Leverage: 现有 Service 实现: `main/app/service/material.go`，Task 2.2 的常量方法
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 GetMaterialDetail 方法中增加原产地信息返回 | Context: 如果 OriginCountryCode 不为空，从常量获取国家信息并返回 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 修改成功，原产地信息正确返回

- [x] 3.6 修改 Material Service - AddMaterial

  - File: `main/app/service/material.go`
  - Purpose: 在 AddMaterial 方法中保存原产地信息
  - Requirements: R3.2
  - Leverage: 现有 Service 实现: `main/app/service/material.go`，Task 3.1 的 DTO
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 AddMaterial 方法中保存 OriginCountryCode 字段 | Context: 从 req.OriginCountryCode 获取并保存到 material.OriginCountryCode | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 修改成功，原产地信息正确保存

- [x] 3.7 修改 Material Service - EditMaterial

  - File: `main/app/service/material.go`
  - Purpose: 在 EditMaterial 方法中更新原产地信息
  - Requirements: R3.3
  - Leverage: 现有 Service 实现: `main/app/service/material.go`，Task 3.1 的 DTO
  - Prompt: Role: Go Developer with business logic expertise | Task: 在 EditMaterial 方法中更新 OriginCountryCode 字段 | Context: 从 req.OriginCountryCode 获取并更新到 material.OriginCountryCode | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: Service 修改成功，原产地信息正确更新

### API 层

- [x] 3.8 创建 Country API Controller

  - File: `main/app/api/v1/shop/shop_country.go`
  - Purpose: 实现国家列表 API 接口
  - Requirements: R3.4
  - Leverage: 现有 API: `main/app/api/v1/shop/shop_nationality.go`，Task 3.3-3.4 的 Service
  - Prompt: Role: Go Developer with Gin framework expertise | Task: 创建 CountryHandler，实现 GetList 接口 | Context: URL 为 /shop/country/list，使用 helper.Success() 返回响应 | Restrictions: 遵循 .cursor/rules/api.mdc，不直接使用 c.JSON() | Success: API 创建成功，响应格式正确

- [x] 3.9 注册 Country API 路由

  - File: `main/router/router.go`
  - Purpose: 注册国家列表 API 路由
  - Requirements: R3.4
  - Leverage: 现有路由: `main/router/router.go`，参考 `shop.RegisterMaterialHandlers`
  - Prompt: Role: Go Developer | Task: 在 shop 路由组中注册国家列表路由 | Context: 路由为 GET /shop/country/list，需要认证中间件 | Restrictions: 遵循现有路由注册模式 | Success: 路由注册成功

- [x] 3.10 注册 Country Service 依赖

  - File: `main/app/api/v1/shop/shop_country.go`
  - Purpose: 在 Service 依赖注入中注册 Country Service
  - Requirements: R3.4
  - Leverage: 现有 Service 注册: `main/app/api/v1/shop/shop_nationality.go`
  - Prompt: Role: Go Developer | Task: 在 Service 依赖注入中注册 Country Service | Context: 创建 CountrySrv 实例并注入到 CountryHandler | Restrictions: 遵循现有依赖注入模式 | Success: Service 注册成功

---

## Phase 4: 测试和优化

- [ ] 4.1 编写 Service 单元测试

  - File: `main/app/service/country_srv_test.go`, `main/app/service/material_srv_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为 CountrySrv 和 MaterialSrv 编写单元测试，覆盖率 ≥ 70% | Context: 测试国家列表获取、物品详情原产地返回、物品创建/编辑原产地保存 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 4.2 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_country_test.go`
  - Purpose: 测试 API 接口
  - Requirements: R3.4
  - Leverage: 现有测试: `main/app/api/v1/shop/*_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为 Country API 编写集成测试 | Context: 测试国家列表接口调用、响应格式、错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

- [ ] 4.3 性能优化（可选）

  - File: `main/app/service/country_srv.go`
  - Purpose: 实现国家列表 Redis 缓存
  - Requirements: 性能要求
  - Leverage: 现有缓存实现: `main/pkg/cache/`
  - Prompt: Role: Go Developer with Redis expertise | Task: 为国家列表实现 Redis 缓存，过期时间24小时 | Context: Key 为 ttpos:country:list，使用 Cache-Aside Pattern | Restrictions: 遵循现有缓存模式 | Success: 缓存实现完成，命中率 > 80%

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

- [ ] API 文档已更新（如有新接口）
- [ ] 数据库文档已更新（如有新表）
- [ ] CHANGELOG.md 已更新

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-shop-material-origin/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-material-origin/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-material-origin/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/active/story-shop-material-origin/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/active/story-shop-material-origin/tasks.md)" | bc
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
**最后更新**: 2025-12-05  
**维护者**: 后端开发组

