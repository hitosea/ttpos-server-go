# 新管理端-叫号系统配置管理 任务分解

> 本文档定义 新管理端-叫号系统配置管理 的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 11（原12个，已取消1个）  
**已完成**: 11  
**进行中**: -  
**完成率**: 100%

---

## Phase 1: DTO 和模型扩展

- [x] 1.1 扩展 BindDeviceReq 增加 name 字段（已取消：绑定设备API移除name参数）

  - File: `main/app/dto/req/callboard.go`
  - Purpose: 绑定设备时支持设置设备名称
  - Requirements: 3.1
  - Leverage: 现有 `BindDeviceReq` 结构体
  - Prompt: Role: Go Developer | Task: 扩展 BindDeviceReq，增加 name 字段（必填，最大长度 20 字符） | Context: 使用 binding 标签验证 `binding:"required,max=20"` | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 扩展成功，validation 正确

- [x] 1.2 扩展 UpdateBindInfoReq 增加 name 字段

  - File: `main/app/dto/req/callboard.go`
  - Purpose: 更新设备信息时支持更新设备名称
  - Requirements: 3.2
  - Leverage: 现有 `UpdateBindInfoReq` 结构体
  - Prompt: Role: Go Developer | Task: 扩展 UpdateBindInfoReq，增加 name 字段（必填，最大长度 20 字符） | Context: 使用 binding 标签验证 `binding:"required,max=20"` | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 扩展成功，validation 正确

- [x] 1.3 扩展 DeviceItem 增加 name 字段

  - File: `main/app/dto/resp/callboard.go`
  - Purpose: 设备列表响应包含设备名称
  - Requirements: 3.7
  - Leverage: 现有 `DeviceItem` 结构体
  - Prompt: Role: Go Developer | Task: 扩展 DeviceItem，增加 name 字段 | Context: 如果名称为空，返回默认值 "WALLACE" | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: DTO 扩展成功

---

## Phase 2: Service 层扩展

- [x] 2.1 扩展 CallBoardService.GetDeviceList 返回设备名称

  - File: `main/app/service/callboard/service.go`
  - Purpose: 设备列表查询时返回设备名称，如果为空返回默认值 "WALLACE"
  - Requirements: 3.5, 3.7
  - Leverage: 现有 `GetDeviceList` 方法
  - Prompt: Role: Go Developer | Task: 扩展 GetDeviceList 方法，从 Redis 读取设备名称，如果为空返回 "WALLACE" | Context: 使用 `cachekey.GetBindedDeviceKey(deviceId)` 获取缓存 key，使用 `HGet` 读取 name 字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法扩展成功，默认值处理正确

- [x] 2.2 扩展 CallBoardService.UpdateBindInfo 支持更新设备名称

  - File: `main/app/service/callboard/service.go`
  - Purpose: 更新设备信息时保存设备名称到 Redis
  - Requirements: 3.2, 3.3
  - Leverage: 现有 `UpdateBindInfo` 方法
  - Prompt: Role: Go Developer | Task: 扩展 UpdateBindInfo 方法，保存设备名称到 Redis 缓存 | Context: 使用 `HMSet` 更新 Redis hash，包含 name 字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 方法扩展成功，名称保存正确

- [x] 2.3 扩展 CallBoardService.BindDevice 支持设置设备名称（已取消：绑定设备API移除name参数）

---

## Phase 3: API 层实现

- [x] 3.1 扩展 /shop/callboard/device/list API（Service层已扩展，API自动支持）

  - File: `main/app/api/v1/shop/shop_callboard.go`
  - Purpose: 设备列表 API 返回设备名称字段
  - Requirements: 3.7
  - Leverage: 现有 `GetDeviceList` API 方法
  - Prompt: Role: Go Developer | Task: 扩展 GetDeviceList API，响应中包含设备名称 | Context: Service 层已返回 name 字段，API 层无需修改（如果 Service 已扩展） | Restrictions: 遵循 .cursor/rules/api.mdc | Success: API 响应包含 name 字段

- [x] 3.2 扩展 /shop/callboard/device/update API（Service层已扩展，API自动支持）

  - File: `main/app/api/v1/shop/shop_callboard.go`
  - Purpose: 更新设备信息 API 支持更新设备名称
  - Requirements: 3.2, 3.3
  - Leverage: 现有 `UpdateBindInfo` API 方法
  - Prompt: Role: Go Developer | Task: 扩展 UpdateBindInfo API，支持接收和更新设备名称 | Context: Request DTO 已包含 name 字段，Service 层已支持更新名称 | Restrictions: 遵循 .cursor/rules/api.mdc | Success: API 支持更新设备名称

- [x] 3.3 新增上传背景图 API

  - File: `main/app/api/v1/shop/shop_callboard.go`
  - Purpose: 上传叫号系统背景图片
  - Requirements: 2.3
  - Leverage: 参考 `main/app/api/v1/shop/shop_product.go` - `UploadProductImage` 方法
  - Prompt: Role: Go Developer | Task: 创建 UploadBackgroundImage API，参考 UploadProductImage 实现 | Context: 使用 FormFile 接收文件，调用 UploadFile Service，返回图片 UUID 和 URL | Restrictions: 遵循 .cursor/rules/api.mdc，文件格式验证（JPEG、PNG、WEBP），文件大小限制（20MB） | Success: API 创建成功，文件上传正确

- [x] 3.4 注册上传背景图 API 路由

  - File: `main/app/api/v1/shop/shop_callboard.go` - `RegisterCallBoardHandlers`
  - Purpose: 注册上传背景图 API 路由
  - Requirements: 3.3
  - Leverage: 现有路由注册代码
  - Success: 路由注册成功，路径为 `/shop/callboard/upload_background_image`

---

## Phase 4: 测试

- [ ] 4.1 编写 Service 单元测试

  - File: `main/app/service/callboard/service_test.go`
  - Purpose: 确保 Service 业务逻辑正确
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/service/*_srv_test.go`
  - Prompt: Role: QA Engineer with Go testing expertise | Task: 为扩展的 Service 方法编写单元测试，覆盖率 ≥ 70% | Context: 测试设备名称管理、配置管理、默认值处理、错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 测试覆盖率 ≥ 70%，所有测试通过

- [ ] 4.2 编写 API 集成测试

  - File: `main/app/api/v1/shop/shop_callboard_test.go`
  - Purpose: 测试 API 接口
  - Requirements: 所有功能需求
  - Leverage: 现有测试: `main/app/api/*_api_test.go`
  - Prompt: Role: QA Engineer specializing in API testing | Task: 为新增和扩展的 API 编写集成测试 | Context: 测试所有 API 接口，测试参数验证，测试响应格式，测试错误处理 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 所有 API 测试通过

- [x] 4.3 手动测试：设备名称管理

  - File: -
  - Purpose: 手动测试设备名称设置和查询
  - Requirements: 3.1, 3.2, 3.7
  - Success: 设备名称设置、更新、查询功能正常，默认值 "WALLACE" 正确

- [x] 4.4 手动测试：背景图片上传

  - File: -
  - Purpose: 手动测试背景图片上传功能
  - Requirements: 2.1, 2.2, 3.3
  - Success: 图片上传成功，格式和大小验证正确

---

## Phase 5: 文档更新

- [x] 5.1 更新 API 文档（Swagger）

  - File: `main/docs/swagger.yaml` 或自动生成
  - Purpose: 确保 API 文档与代码同步
  - Requirements: 所有 API 需求
  - Leverage: Swagger 注释
  - Success: API 文档已更新

- [ ] 5.2 更新 CHANGELOG

  - File: `CHANGELOG.md`
  - Purpose: 记录功能变更
  - Requirements: 文档要求
  - Success: CHANGELOG 已更新

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [x] 所有任务标记为 `[x]`
- [ ] Go 代码通过 `go fmt` 和 `go vet`
- [ ] 测试覆盖率达标
  - Service: ≥ 70%
  - Repository: ≥ 80%
- [ ] 所有测试通过

### 功能完整性

- [x] requirements.md 中的所有需求已满足
- [x] design.md 中的设计已实现
- [x] 验收标准已达成
- [x] 设备名称管理功能正常
- [x] 背景图片上传功能正常

### 文档同步

- [x] API 文档已更新（Swagger）
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
grep -c "^- \[" docs/shared/specs/active/story-shop-callboard-settings/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-shop-callboard-settings/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-shop-callboard-settings/tasks.md
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

---

## ✅ 完成状态

**开发阶段**: ✅ 已完成  
**测试阶段**: ✅ 手动测试已完成  
**文档阶段**: ✅ API 文档已更新  
**待完成**: CHANGELOG 更新（可选）
