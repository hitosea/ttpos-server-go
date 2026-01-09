# Grab 菜单获取回退 TTPOS 导出接口 任务分解

> 本文档定义 Grab 菜单获取回退 TTPOS 导出接口功能的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 8  
**已完成**: 8  
**进行中**: -  
**完成率**: 100%

---

## Phase 1: 配置准备

- [x] 1.1 添加 TTPOS endpoint 配置项

  - File: `ttpos-bmp/app/ttpos-takeout/manifest/config/config.tpl.yaml`
  - Purpose: 配置 TTPOS 主模块的 HTTP 地址，用于回退调用
  - Requirements: 2.2
  - Leverage: 现有配置项 `app.callbackSecret`
  - Changes:
    ```yaml
    app:
      ttposEndpoint: $TTPOS_MAIN_ENDPOINT  # 新增
      callbackSecret: $JWT_SECRET          # 已有
    ```
  - Success: 配置项添加成功，可通过 `g.Cfg().MustGet(ctx, "app.ttposEndpoint")` 读取

---

## Phase 2: 核心实现（Go BMP）

### Logic 层

- [x] 2.1 实现 getTtposAuth 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
  - Purpose: 生成 TTPOS 认证头，用于调用 main 模块接口
  - Requirements: 2.3
  - Leverage: `ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/skootar.go` - `getCallBackAuth` 方法
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 实现 getTtposAuth 方法，生成 MD5(shopUUID + callbackSecret) 认证头 | Context: 参考 skootar.go 中的 getCallBackAuth 方法，使用 gmd5.EncryptString | Restrictions: 使用 gerror 包装错误，不记录 secret 到日志 | Success: 方法实现完成，返回正确的 MD5 字符串

- [x] 2.2 实现 fetchMenuFromTTpos 方法

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
  - Purpose: 调用 TTPOS 主模块 `/api/v1/takeout/menu/export` 接口获取菜单数据
  - Requirements: 2.1, 2.2, 2.4, 2.5, 2.6
  - Leverage:
    - Task 2.1 的 `getTtposAuth` 方法
    - `ttpos-bmp/app/ttpos-takeout/internal/logic/skootar/estimate_distance.go` - HTTP 调用示例
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 实现 fetchMenuFromTTpos 方法，调用 TTPOS 导出接口 | Context: 使用 g.Client() 发起 HTTP POST 请求，设置 10s 超时，携带 X-TTPOS-SECRET 头 | Restrictions: 使用 gerror 包装所有错误，配置缺失返回 gcode.CodeMissingConfiguration | Success: 方法实现完成，能正确调用 TTPOS 接口并解析响应

- [x] 2.3 修改 HandleGetMenu 方法，添加回退逻辑

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu.go`
  - Purpose: 在本地菜单快照为空时，回退调用 TTPOS 导出接口
  - Requirements: 1.1, 1.2, 1.3, 2.1, 3.1, 3.2, 3.3
  - Leverage:
    - 现有 `HandleGetMenu` 方法
    - Task 2.2 的 `fetchMenuFromTTpos` 方法
  - Prompt: Role: Go Developer with GoFrame expertise | Task: 修改 HandleGetMenu 方法，在 menuJSON 为空时调用 fetchMenuFromTTpos | Context: 保持现有逻辑不变，仅在 menuJSON == "" 时添加回退分支 | Restrictions: 使用 gerror 包装错误，记录 Info/Error 日志 | Success: HandleGetMenu 支持回退调用，测试通过

---

## Phase 3: 测试

- [x] 3.1 编写 getTtposAuth 单元测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu_test.go`
  - Purpose: 测试认证头生成逻辑
  - Requirements: 2.3
  - Leverage: 现有测试文件
  - Test Cases:
    - 正常生成 MD5 字符串
    - 不同 shopUUID 生成不同结果
  - Success: 测试通过，覆盖正常场景

- [x] 3.2 编写 fetchMenuFromTTpos 单元测试（Mock HTTP）

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu_test.go`
  - Purpose: 测试 TTPOS 接口调用逻辑
  - Requirements: 2.1, 2.2, 2.4, 2.5, 2.6
  - Leverage: `net/http/httptest` 包
  - Test Cases:
    - TTPOS 接口返回成功
    - TTPOS 接口返回非 200
    - TTPOS 接口超时
    - 配置缺失
  - Success: 测试通过，覆盖所有错误场景

- [x] 3.3 编写 HandleGetMenu 集成测试

  - File: `ttpos-bmp/app/ttpos-takeout/internal/logic/grab_menu/grab_menu_test.go`
  - Purpose: 测试完整的 GetMenu 流程
  - Requirements: 1.1, 1.2, 2.1, 3.1, 3.2
  - Leverage: Mock 数据库和 HTTP 服务
  - Test Cases:
    - 本地快照存在 → 直接返回
    - 本地快照为空 → 回退调用成功
    - 本地快照为空 → 回退调用失败 → 返回 CodeNotFound
    - partnerMerchantID 无效 → 返回 CodeInvalidParameter
  - Success: 测试通过，覆盖所有业务场景

---

## Phase 4: 文档和收尾

- [x] 4.1 更新配置文档

  - File: `ttpos-bmp/app/ttpos-takeout/README.MD` 或相关配置文档
  - Purpose: 说明 `app.ttposEndpoint` 配置项的用途和格式
  - Requirements: 文档验收
  - Changes:
    ```markdown
    ### 配置说明
    
    | 配置项 | 说明 | 示例 |
    |--------|------|------|
    | app.ttposEndpoint | TTPOS 主模块地址 | http://main-service:8080 |
    ```
  - Success: 配置文档更新完成

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [x] 所有任务标记为 `[x]`
- [x] Go 代码通过 `go fmt` 和 `go vet`
- [x] 所有错误使用 `gerror` 包装
- [x] 日志格式统一：`[Grab] {操作}: shopUUID=%d, {其他参数}`
- [x] 敏感信息不记录到日志

### 功能完整性

- [x] requirements.md 中的所有需求已满足
- [x] design.md 中的设计已实现
- [x] 验收标准已达成：
  - [x] 本地快照存在 → 直接返回
  - [x] 本地快照为空 → 回退调用 TTPOS 接口
  - [x] TTPOS 调用失败 → 返回 menu not found

### 测试通过

- [x] 所有单元测试通过（需在配置环境中运行）
- [ ] 所有集成测试通过（需部署环境验证）

### 规范遵循

- [x] 遵循 `ttpos-bmp/.cursor/rules/go-rules.mdc`
- [x] 遵循 `.cursor/rules/go-bmp.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/active/story-takeout-grab-menu-fallback-ttpos/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/active/story-takeout-grab-menu-fallback-ttpos/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/active/story-takeout-grab-menu-fallback-ttpos/tasks.md
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
Role: Go Developer with GoFrame 2.x expertise

Task: {具体任务描述}

Context:
- Current file: {文件路径}
- Leverage code: {可复用代码路径}
- Requirements: {需求编号和内容}
- Project specs: 遵循 ttpos-bmp/.cursor/rules/go-rules.mdc, .cursor/rules/go-bmp.mdc

Restrictions:
- 使用 GoFrame 2.x 框架
- 使用 g.Client() 发起 HTTP 请求
- 使用 gerror 包装所有错误
- 日志格式：[Grab] {操作}: shopUUID=%d
- 敏感信息不记录到日志

Success Criteria:
- {成功标准}
- 代码通过 go fmt 和 go vet
```

### Go BMP 测试

```
Role: QA Engineer with Go testing expertise

Task: {测试任务描述}

Context:
- Target file: {测试目标文件}
- Test file: {测试文件路径}
- Requirements: {需求编号}

Test Cases Required:
- 正常场景测试
- 异常场景测试
- 边界条件测试

Restrictions:
- 使用 httptest 包 Mock HTTP 服务
- 测试所有错误分支

Success Criteria:
- 所有测试通过
- 覆盖正常和异常场景
```

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/2025-12/2025-12-18.md`

---

**模板版本**: v1.0.0  
**最后更新**: 2025-12-18  
**维护者**: rikugun

