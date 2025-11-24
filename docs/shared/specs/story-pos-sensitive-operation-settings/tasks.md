# 敏感操作设置（收银机）任务分解

> 本文档定义收银机和点餐助手端敏感操作（折扣/退款）权限验证的详细执行任务清单。

## 📋 任务分解原则

- **颗粒度**: 每个任务 1-4 小时（SP ≤ 1）
- **可追踪**: 使用 `- [ ]` 和 `- [x]` 标记完成状态
- **可复用**: 明确 Leverage 可复用代码
- **AI 友好**: 提供 Prompt 模板辅助执行
- **需求关联**: 每个任务关联 requirements.md 中的具体需求

## 📊 进度总览

**总任务数**: 22  
**已完成**: 0  
**进行中**: -  
**完成率**: 0%

---

## Phase 1: 接口开发

### DTO 层

- [ ] 1.1 创建检查授权请求和响应 DTO

  - File: `main/app/dto/req/order.go`, `main/app/dto/resp/order.go`
  - Purpose: 定义检查授权接口的请求和响应结构
  - Requirements: 1.1, 1.2, 1.3, 1.4
  - Leverage: 现有的 DTO 结构，参考其他请求/响应定义
  - Prompt: Role: Go Developer | Task: 创建检查授权请求和响应 DTO | Context: 需要创建 CheckAuthorizationReq (operation_type: string) 和 CheckAuthorizationResp (has_permission: bool) | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 binding tag 进行验证 | Success: DTO 创建完成，字段定义正确，验证规则正确

- [ ] 1.2 创建密码验证请求和响应 DTO

  - File: `main/app/dto/req/order.go`, `main/app/dto/resp/order.go`
  - Purpose: 定义密码验证接口的请求和响应结构
  - Requirements: 2.1, 2.2, 2.3, 2.4
  - Leverage: 现有的 DTO 结构，参考其他请求/响应定义
  - Prompt: Role: Go Developer | Task: 创建密码验证请求和响应 DTO | Context: 需要创建 VerifyPasswordReq (operation_type, authorized_staff_account, password) 和 VerifyPasswordResp (verified, authorized_staff) | Restrictions: 遵循 .cursor/rules/go-main.mdc，使用 binding tag 进行验证 | Success: DTO 创建完成，字段定义正确，验证规则正确

- [ ] 1.3 扩展退款请求 DTO，增加授权参数

  - File: `main/app/dto/req/order.go`
  - Purpose: 在 OrderReturnReq 中增加授权员工账号和密码字段
  - Requirements: 3.1
  - Leverage: 现有的 `OrderReturnReq` 结构体
  - Prompt: Role: Go Developer | Task: 在 OrderReturnReq 结构体中添加授权参数字段 | Context: 需要添加 AuthorizedStaffAccount (string, 可选) 和 AuthorizedStaffPassword (string, 可选) 两个字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc，字段为可选 | Success: DTO 扩展完成，字段定义正确

- [ ] 1.4 扩展整单改价请求 DTO，增加授权参数

  - File: `main/app/dto/req/order.go`
  - Purpose: 在 OrderAmountChangeReq 中增加授权员工账号和密码字段
  - Requirements: 4.1
  - Leverage: 现有的 `OrderAmountChangeReq` 结构体
  - Prompt: Role: Go Developer | Task: 在 OrderAmountChangeReq 结构体中添加授权参数字段 | Context: 需要添加 AuthorizedStaffAccount (string, 可选) 和 AuthorizedStaffPassword (string, 可选) 两个字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc，字段为可选 | Success: DTO 扩展完成，字段定义正确

- [ ] 1.5 扩展打折请求 DTO，增加授权参数

  - File: `main/app/dto/req/order.go`
  - Purpose: 在 OrderDiscountReq 中增加授权员工账号和密码字段
  - Requirements: 4.2
  - Leverage: 现有的 `OrderDiscountReq` 结构体
  - Prompt: Role: Go Developer | Task: 在 OrderDiscountReq 结构体中添加授权参数字段 | Context: 需要添加 AuthorizedStaffAccount (string, 可选) 和 AuthorizedStaffPassword (string, 可选) 两个字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc，字段为可选 | Success: DTO 扩展完成，字段定义正确

- [ ] 1.6 扩展抹零请求 DTO，增加授权参数

  - File: `main/app/dto/req/order.go`
  - Purpose: 在 OrderZeroRuleReq 中增加授权员工账号和密码字段
  - Requirements: 4.3
  - Leverage: 现有的 `OrderZeroRuleReq` 结构体
  - Prompt: Role: Go Developer | Task: 在 OrderZeroRuleReq 结构体中添加授权参数字段 | Context: 需要添加 AuthorizedStaffAccount (string, 可选) 和 AuthorizedStaffPassword (string, 可选) 两个字段 | Restrictions: 遵循 .cursor/rules/go-main.mdc，字段为可选 | Success: DTO 扩展完成，字段定义正确

### API 层

- [ ] 1.7 创建检查授权接口

  - File: `main/app/api/v1/cashier/cashier_order.go` 或新建文件
  - Purpose: 创建检查授权 API 接口
  - Requirements: 1.1, 1.2, 1.3, 1.4
  - Leverage: 现有的 cashier_order.go 文件结构
  - Prompt: Role: Go Developer | Task: 创建检查授权 API 接口 | Context: 接口路径为 /cashier/order/check_authorization (POST)，调用 Service 的 CheckAuthorization 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc 和 .cursor/rules/api.mdc | Success: API 接口创建完成，路由注册正确，响应格式正确

- [ ] 1.8 创建密码验证接口

  - File: `main/app/api/v1/cashier/cashier_order.go` 或新建文件
  - Purpose: 创建密码验证 API 接口
  - Requirements: 2.1, 2.2, 2.3, 2.4
  - Leverage: 现有的 cashier_order.go 文件结构
  - Prompt: Role: Go Developer | Task: 创建密码验证 API 接口 | Context: 接口路径为 /cashier/order/verify_password (POST)，调用 Service 的 VerifyPassword 方法 | Restrictions: 遵循 .cursor/rules/go-main.mdc 和 .cursor/rules/api.mdc | Success: API 接口创建完成，路由注册正确，响应格式正确，错误处理正确

### Service 层

- [ ] 1.9 实现检查授权 Service 方法

  - File: `main/app/service/order_manage.go` 或新建文件
  - Purpose: 实现检查当前员工是否有权限操作的逻辑
  - Requirements: 1.5, 1.6, 1.7
  - Leverage: `main/app/service/setting/setting.go` - 业务设置 Service
  - Prompt: Role: Go Developer | Task: 实现检查授权 Service 方法 | Context: 需要从业务设置中读取授权员工列表和密码验证开关，判断当前员工是否在授权名单中 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Service 只依赖其他 Service 接口 | Success: Service 方法实现完成，逻辑正确，错误处理正确

- [ ] 1.10 实现密码验证 Service 方法

  - File: `main/app/service/order_manage.go` 或新建文件
  - Purpose: 实现密码验证逻辑
  - Requirements: 2.5, 2.6, 2.7, 2.8
  - Leverage: `main/app/service/staff.go` - 员工 Service，`main/app/service/setting/setting.go` - 业务设置 Service
  - Prompt: Role: Go Developer | Task: 实现密码验证 Service 方法 | Context: 需要根据账号（邮箱或手机号）查找员工，验证员工是否在授权名单中，验证密码是否正确 | Restrictions: 遵循 .cursor/rules/go-main.mdc，Service 只依赖其他 Service 接口 | Success: Service 方法实现完成，逻辑正确，错误处理正确

---

## Phase 2: 退款接口增强

- [ ] 2.1 修改退款 Service，增加授权验证逻辑

  - File: `main/app/service/order_manage.go`
  - Purpose: 在退款逻辑开始前，检查授权并验证权限
  - Requirements: 3.2, 3.3, 3.4
  - Leverage: 现有的 `ReturnOrder` 方法，Phase 1 创建的检查授权和密码验证方法
  - Prompt: Role: Go Developer | Task: 在退款 Service 中增加授权验证逻辑 | Context: 如果提供了授权参数，调用密码验证方法；如果没有提供，调用检查授权方法；验证通过后记录授权员工信息 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 授权验证逻辑正确，错误处理正确

- [ ] 2.2 修改操作记录创建逻辑，记录授权员工信息

  - File: `main/app/service/order_manage.go`
  - Purpose: 在创建退款操作记录时，如果使用了授权验证，记录授权员工信息
  - Requirements: 3.5, 3.6, 6.1, 6.2, 6.4
  - Leverage: 现有的操作记录创建逻辑
  - Prompt: Role: Go Developer | Task: 修改操作记录创建逻辑，记录授权员工信息 | Context: 在创建退款操作记录时，如果使用了授权验证，在 data 字段的 JSON 中增加 authorized_staff 对象（包含 uuid, name, email） | Restrictions: 遵循 .cursor/rules/go-main.mdc 和 .cursor/rules/database.mdc | Success: 操作记录创建逻辑正确，授权员工信息正确记录

---

## Phase 3: 折扣接口增强（整单改价、打折、抹零）

- [ ] 3.1 修改整单改价 Service，增加授权验证逻辑

  - File: `main/app/service/order_discount.go`
  - Purpose: 在整单改价逻辑开始前，检查授权并验证权限
  - Requirements: 4.4, 4.7, 4.8
  - Leverage: 现有的 `OrderAmountChange` 方法，Phase 1 创建的检查授权和密码验证方法
  - Prompt: Role: Go Developer | Task: 在整单改价 Service 中增加授权验证逻辑 | Context: 如果提供了授权参数，调用密码验证方法；如果没有提供，调用检查授权方法；验证通过后记录授权员工信息 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 授权验证逻辑正确，错误处理正确

- [ ] 3.2 修改打折 Service，增加授权验证逻辑

  - File: `main/app/service/order_discount.go`
  - Purpose: 在打折逻辑开始前，检查授权并验证权限
  - Requirements: 4.5, 4.7, 4.8
  - Leverage: 现有的 `OrderDiscount` 方法，Phase 1 创建的检查授权和密码验证方法
  - Prompt: Role: Go Developer | Task: 在打折 Service 中增加授权验证逻辑 | Context: 如果提供了授权参数，调用密码验证方法；如果没有提供，调用检查授权方法；验证通过后记录授权员工信息 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 授权验证逻辑正确，错误处理正确

- [ ] 3.3 修改抹零 Service，增加授权验证逻辑

  - File: `main/app/service/order_discount.go`
  - Purpose: 在抹零逻辑开始前，检查授权并验证权限
  - Requirements: 4.6, 4.7, 4.8
  - Leverage: 现有的 `OrderZeroRule` 方法，Phase 1 创建的检查授权和密码验证方法
  - Prompt: Role: Go Developer | Task: 在抹零 Service 中增加授权验证逻辑 | Context: 如果提供了授权参数，调用密码验证方法；如果没有提供，调用检查授权方法；验证通过后记录授权员工信息 | Restrictions: 遵循 .cursor/rules/go-main.mdc | Success: 授权验证逻辑正确，错误处理正确

- [ ] 3.4 修改操作记录创建逻辑，记录授权员工信息

  - File: `main/app/service/order_discount.go`
  - Purpose: 在创建整单改价/打折/抹零操作记录时，如果使用了授权验证，记录授权员工信息
  - Requirements: 4.9, 4.10, 6.1, 6.2, 6.3
  - Leverage: 现有的操作记录创建逻辑，参考 Phase 2.2 的实现
  - Prompt: Role: Go Developer | Task: 修改操作记录创建逻辑，记录授权员工信息 | Context: 在创建整单改价/打折/抹零操作记录时，如果使用了授权验证，在 data 字段的 JSON 中增加 authorized_staff 对象（包含 uuid, name, email） | Restrictions: 遵循 .cursor/rules/go-main.mdc 和 .cursor/rules/database.mdc | Success: 操作记录创建逻辑正确，授权员工信息正确记录

---

## Phase 4: 前端实现

**注意**: 前端实现需要根据实际的前端技术栈（Flutter/React Native/Vue等）来确定具体实现方式。以下任务为通用描述。

### 授权验证弹窗组件

- [ ] 4.1 创建授权验证弹窗组件（POS）

  - File: 根据前端技术栈确定
  - Purpose: 创建授权验证弹窗组件，支持输入授权员工账号和密码
  - Requirements: 5.1, 5.2, 5.3
  - Leverage: 现有的弹窗组件
  - Prompt: Role: Frontend Developer | Task: 创建授权验证弹窗组件 | Context: 弹窗包含授权员工账号输入框（支持邮箱或手机号）、权限密码输入框、确认/取消按钮 | Restrictions: 遵循前端开发规范 | Success: 弹窗组件创建完成，UI 正确，交互正常

- [ ] 4.2 创建授权验证弹窗组件（Assistant）

  - File: 根据前端技术栈确定
  - Purpose: 创建授权验证弹窗组件（可复用 POS 的组件）
  - Requirements: 5.1, 5.2, 5.3
  - Leverage: Phase 4.1 创建的弹窗组件
  - Prompt: Role: Frontend Developer | Task: 创建授权验证弹窗组件（Assistant） | Context: 可复用 POS 的弹窗组件，或创建独立的组件 | Restrictions: 遵循前端开发规范 | Success: 弹窗组件创建完成，UI 正确，交互正常

### 接口调用

- [ ] 4.3 实现检查授权接口调用

  - File: 根据前端技术栈确定
  - Purpose: 在点击折扣/退款按钮时，调用检查授权接口
  - Requirements: 5.1
  - Leverage: 现有的 API 调用封装
  - Prompt: Role: Frontend Developer | Task: 实现检查授权接口调用 | Context: 在点击折扣/退款按钮时，先调用检查授权接口，根据返回结果决定是否弹出授权弹窗 | Restrictions: 遵循前端开发规范 | Success: 接口调用正确，逻辑正确

- [ ] 4.4 实现密码验证接口调用

  - File: 根据前端技术栈确定
  - Purpose: 在授权弹窗中，调用密码验证接口
  - Requirements: 5.4, 5.5, 5.6, 5.7
  - Leverage: 现有的 API 调用封装
  - Prompt: Role: Frontend Developer | Task: 实现密码验证接口调用 | Context: 在授权弹窗中，输入账号和密码后，调用密码验证接口，处理成功和失败情况 | Restrictions: 遵循前端开发规范 | Success: 接口调用正确，错误处理正确，toast 提示正确

- [ ] 4.5 实现退款接口调用（带授权参数）

  - File: 根据前端技术栈确定
  - Purpose: 在退款时，如果使用了授权验证，传递授权参数
  - Requirements: 5.7
  - Leverage: 现有的退款接口调用
  - Prompt: Role: Frontend Developer | Task: 实现退款接口调用（带授权参数） | Context: 在调用退款接口时，如果使用了授权验证，传递 authorized_staff_account 和 authorized_staff_password 参数 | Restrictions: 遵循前端开发规范 | Success: 接口调用正确，参数传递正确

- [ ] 4.6 实现折扣接口调用（带授权参数）

  - File: 根据前端技术栈确定
  - Purpose: 在折扣时，如果使用了授权验证，传递授权参数
  - Requirements: 5.7
  - Leverage: 现有的折扣接口调用
  - Prompt: Role: Frontend Developer | Task: 实现折扣接口调用（带授权参数） | Context: 在调用折扣接口时，如果使用了授权验证，传递 authorized_staff_account 和 authorized_staff_password 参数 | Restrictions: 遵循前端开发规范 | Success: 接口调用正确，参数传递正确

### 表单验证

- [ ] 4.7 实现授权弹窗表单验证

  - File: 根据前端技术栈确定
  - Purpose: 实现授权弹窗的表单验证逻辑
  - Requirements: 5.3
  - Leverage: 现有的表单验证逻辑
  - Prompt: Role: Frontend Developer | Task: 实现授权弹窗表单验证 | Context: 验证授权员工账号和权限密码必填，未输入时在输入框下方提示 | Restrictions: 遵循前端开发规范 | Success: 表单验证正确，提示信息正确

---

## Phase 5: 测试和优化

- [ ] 5.1 功能测试

  - File: `test/`（或相应测试目录）
  - Purpose: 测试敏感操作设置的所有功能
  - Requirements: 所有功能需求
  - Leverage: 现有测试框架
  - Prompt: Role: QA Engineer | Task: 编写功能测试用例，覆盖所有功能需求 | Context: 测试检查授权接口、密码验证接口、退款接口增强、折扣接口增强、授权验证弹窗、操作记录增强 | Restrictions: 测试覆盖完整 | Success: 所有功能测试通过

- [ ] 5.2 集成测试

  - File: `test/`（或相应测试目录）
  - Purpose: 测试完整流程
  - Requirements: 所有功能需求
  - Leverage: 现有集成测试框架
  - Prompt: Role: QA Automation Engineer | Task: 实现端到端集成测试 | Context: 测试完整流程：检查授权 → 弹出弹窗 → 密码验证 → 执行操作 → 记录操作日志 | Restrictions: 测试真实用户场景 | Success: 集成测试通过

- [ ] 5.3 边界测试

  - File: `test/`（或相应测试目录）
  - Purpose: 测试各种边界情况
  - Requirements: 所有功能需求
  - Leverage: 现有测试框架
  - Prompt: Role: QA Engineer | Task: 测试边界情况 | Context: 测试密码错误、账号不存在、账号不是权限员工、后台修改密码后再次验证等边界情况 | Restrictions: 测试覆盖完整 | Success: 边界测试通过

---

## 提交清单

完成所有任务后，请检查：

### 代码质量

- [ ] 所有任务标记为 `[x]`
- [ ] Go 代码通过 gofmt 格式化，通过 golangci-lint 检查
- [ ] 所有测试通过

### 功能完整性

- [ ] requirements.md 中的所有需求已满足
- [ ] design.md 中的设计已实现
- [ ] 验收标准已达成

### 文档同步

- [ ] API 文档已更新（如有新接口）
- [ ] CHANGELOG.md 已更新
- [ ] 用户指南已更新（如需要）

### 规范遵循

- [ ] 遵循 `.cursor/rules/go-main.mdc`
- [ ] 遵循 `.cursor/rules/api.mdc`
- [ ] 遵循 `.cursor/rules/database.mdc`
- [ ] 遵循 `.cursor/rules/security.mdc`

---

## 进度追踪

### 手动追踪命令

```bash
# 查看总任务数
grep -c "^- \[" docs/shared/specs/story-pos-sensitive-operation-settings/tasks.md

# 查看已完成任务数
grep -c "^- \[x\]" docs/shared/specs/story-pos-sensitive-operation-settings/tasks.md

# 查看未完成任务数
grep -c "^- \[ \]" docs/shared/specs/story-pos-sensitive-operation-settings/tasks.md

# 计算完成率
echo "scale=2; $(grep -c "^- \[x\]" docs/shared/specs/story-pos-sensitive-operation-settings/tasks.md) * 100 / $(grep -c "^- \[" docs/shared/specs/story-pos-sensitive-operation-settings/tasks.md)" | bc
```

### 执行流程

1. **选择任务**: 选择下一个未完成任务
2. **阅读需求**: 查看 requirements.md 中的关联需求
3. **查看复用**: 检查 Leverage 中的可复用代码
4. **使用 AI**: 复制 Prompt 模板，让 AI 生成代码
5. **实现代码**: 按照规范实现功能
6. **运行检查**: Go 代码格式化，golangci-lint 检查
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
**最后更新**: 2025-11-24  
**维护者**: 后端开发组

