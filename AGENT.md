---
alwaysApply: true
description: Agent 速查表（后端版）
priority: 1
---

# Agent 速查表

> 本文件包含所有核心规则的压缩映射表。遇到任何问题优先查阅此文件。

---

## 场景识别 → 执行命令

| 用户说...                      | 识别为     | 立即执行                                                               | 涉及文件/目录                  | 参考规范        |
| ------------------------------ | ---------- | ---------------------------------------------------------------------- | ------------------------------ | --------------- |
| "新人" "入职" "不熟悉"         | 新成员入职 | `/onboard quick` → 推送必读清单                                        | agent/workflows/onboarding.md  | intro.mdc       |
| "有个想法" "提需求" "能不能做" | 需求发起   | `/propose {name}` → 填写 → 评审 → `/create-spec story-{module}-{name}` | team/proposals/, shared/specs/ | specs.mdc       |
| "实现功能" "开发 XX" "新增 XX" | 功能开发   | 读 `shared/specs/{}/tasks.md` → 逐任务执行                             | shared/specs/, main/, admin/   | go-main/php.mdc |
| "报错" "bug" "崩溃" "异常"     | Bug 修复   | 搜 Graphiti → 定位 → 修复 → 测试 → 记录                                | shared/troubleshooting/, main/ | go-main.mdc     |
| "集成 XX" "对接 XX" "API"      | 第三方集成 | 查 `integrations/{service}/` → 创建 API 类 → 测试 → 文档               | integrations/, shared/api/     | api.mdc         |
| "迁移数据库" "新增表" "改字段" | 数据库迁移 | 创建迁移文件 → 更新 model → 更新 seeds                                 | admin/database/migrations/     | database.mdc    |
| "gRPC" "微服务" "ttpos-bmp"    | 微服务集成 | 查 ttpos-bmp 文档 → 定义 Protobuf → 注册服务                           | ttpos-bmp/                     | go-bmp.mdc      |
| "慢" "卡顿" "优化性能"         | 性能优化   | 分析瓶颈 → 优化 → 验证 → 记录                                          | 相关代码                       | go-main.mdc     |
| "前端开发" "Vue 组件" "页面"   | 前端开发   | 读 `shared/specs/{}/tasks.md` → 实现组件 → 测试                        | admin/views/                   | vue.mdc         |
| "安全审查" "漏洞" "SQL 注入"   | 安全检查   | 查 security.mdc → 检查代码 → 修复 → 测试                               | 所有代码                       | security.mdc    |
| "提交代码" "git commit"        | Git 提交   | 查 version.mdc → 写提交信息 → 推送                                     | .git/                          | version.mdc     |
| "新人" "入职" "不熟悉"         | 新成员     | `/onboard quick` → 推送必读清单                                        | .cursor/rules/, docs/          | intro.mdc       |

---

## 核心规范 (硬性约束)

### 命名规范

| 类型         | 格式                         | 示例                                      | 规则                                                             |
| ------------ | ---------------------------- | ----------------------------------------- | ---------------------------------------------------------------- |
| **Spec**     | `{level}-{module}-{feature}` | `story-order-quick-payment`               | level: story/task, module: order/member/..., feature: kebab-case |
| **Proposal** | `{YYYY-MM-DD}-{feature}`     | `2025-11-16-quick-payment`                | 存放: docs/team/proposals/                                       |
| **Go 类名**  | PascalCase                   | `OrderController`, `UserService`          | 大驼峰，接口以 I 开头                                            |
| **PHP 类名** | PascalCase                   | `OrderController`, `UserService`          | 大驼峰                                                           |
| **文件名**   | snake_case                   | `order_controller.go`, `user_service.php` | 小写+下划线                                                      |
| **URL**      | snake_case                   | `/api/v1/order/cart_info`                 | 蛇形命名，不用 kebab-case                                        |
| **变量**     | camelCase                    | `userName`, `isLoading`                   | 小驼峰                                                           |
| **常量**     | UPPER_SNAKE_CASE             | `API_BASE_URL`                            | 全大写+下划线                                                    |

### Go Service/Repository (必须遵守)

```go
// ✅ 接口以 I 开头，实现以 Impl 结尾
type IOrderSrv interface { }
type OrderSrvImpl struct { }

// ✅ Service 依赖其他 Service
memberSrv IMemberSrv

// ✅ Repository 只持有 db (不是 dbm)
db *gorm.DB
```

### API 响应规范 (必须遵守)

```json
// ✅ data 必须是对象
{"code": 1, "message": "success", "data": {}}
{"code": 1, "message": "success", "data": {"list": []}}

// ❌ data 不能是 null 或数组
{"data": null}  // 错误
{"data": []}    // 错误
```

### 代码风格 (必须遵守)

| 规则           | 正确                                       | 错误                      |
| -------------- | ------------------------------------------ | ------------------------- |
| **不用 panic** | `return nil, errors.New("错误")`           | `panic("错误")`           |
| **URL 命名**   | `/api/v1/order/cart_info`                  | `/api/v1/order/cart-info` |
| **日志输出**   | `logger.Logger.Error()` (Go) <br> 中文注释 | `print()`, 英文注释       |
| **错误处理**   | `catch (error, stackTrace)`                | `catch (error)`           |
| **接口命名**   | `IUserService`                             | `UserService` (interface) |

### 测试覆盖率 (P0)

| 模块类型                 | 覆盖率要求 | 风险等级 |
| ------------------------ | ---------- | -------- |
| **main/app/service/**    | ≥70%       | P0       |
| **main/app/repository/** | ≥80%       | P0       |
| **ttpos-bmp/logic/**     | ≥70%       | P1       |
| **payment 相关**         | **100%**   | 高风险   |
| **order 相关**           | **100%**   | 高风险   |

### 数据库规范 (必须遵守)

```php
// ✅ 必须字段
uuid, create_time, update_time, delete_time

// ✅ 字段类型
时间: int 类型, _time 结尾, 默认值 0
金额: decimal(20,8)
UUID: biginteger, 默认值 0
```

---

## 文件路径速查（按受众分类）

### 🤖 Agent 优先（执行清单）

| 我需要...                 | 文件路径                                        | 用途                           |
| ------------------------- | ----------------------------------------------- | ------------------------------ |
| **执行工作流**            | `docs/agent/workflows/*.md`                     | 步骤检查清单                   |
| **填充模板**              | `docs/agent/templates/*.md`                     | 结构化表单                     |
| **记录排查指南**          | `docs/agent/templates/troubleshooting-guide.md` | 创建/更新 troubleshooting 文档 |
| **记录 Graphiti Episode** | `docs/agent/templates/graphiti-episode.md`      | Graphiti 入库模板              |
| **Graphiti 草稿**         | `docs/agent/graphiti/`                          | Episode 草稿仓库               |
| **查看指令**              | `.cursor/commands/*.md`                         | 指令参数和用法                 |
| **查看规范**              | `.cursor/rules/*.mdc`                           | 规则速查                       |
| **执行任务**              | `docs/shared/specs/*/tasks.md`                  | 任务逐条执行                   |
| **创建需求提案**          | `docs/team/proposals/{YYYY-MM-DD}-{name}`       | `/propose`                     |
| **创建功能规格**          | `docs/shared/specs/{level}-{module}-{feature}/` | `/create-spec`                 |

### 👤 人类优先（学习资料）

| 我需要...    | 文件路径                        | 用途         |
| ------------ | ------------------------------- | ------------ |
| **学习指南** | `docs/human/guides/*.md`        | 详细教程     |
| **架构设计** | `docs/human/architecture/*.md`  | 系统设计原理 |
| **业务知识** | `docs/human/business/*.md`      | 业务理解     |
| **技术决策** | `docs/human/decisions/`         | ADR 记录     |
| **提案写作** | `docs/team/proposals/README.md` | 提案指南     |

### 📚 共用资源（Agent + 人类）

| 我需要...      | 文件路径                           | 用途       |
| -------------- | ---------------------------------- | ---------- |
| **功能规格**   | `docs/shared/specs/story-*-*/`     | 需求和设计 |
| **API 文档**   | `docs/shared/api/*.md`             | 接口查询   |
| **问题排查**   | `docs/shared/troubleshooting/*.md` | 故障处理   |
| **第三方集成** | `docs/shared/integrations/`        | 集成文档   |
| **记录经验**   | Graphiti                           | MCP add    |
| **查询历史**   | Graphiti                           | MCP search |

### 📖 目录导航索引

```yaml
docs/
├── agent/              # 🤖 Agent 专用
│   ├── workflows/      #    工作流执行清单
│   └── templates/      #    结构化模板
│
├── human/              # 👤 人类专用
│   ├── guides/         #    学习指南
│   ├── architecture/   #    架构设计
│   ├── business/       #    业务知识
│   └── decisions/      #    技术决策（ADR）
│
├── shared/             # 📚 共用资源
│   ├── specs/          #    功能规格
│   ├── api/            #    API文档
│   ├── troubleshooting/#    问题排查
│   └── integrations/   #    第三方集成
│
└── team/               # 👥 团队协作
    ├── proposals/      #    需求提案
    └── activities/     #    活动日志
```

---

## 检索优先级 (强制顺序)

```
1. 本速查表 (AGENT.md)
2. Graphiti (经验类问题: "如何" "为什么" "踩坑")
3. .cursor/rules/*.mdc (规范核心清单)
4. docs/ (详细文档、完整示例)
5. codebase_search (实现细节、代码位置)
6. web_search (最新技术、第三方文档)
```

---

## 决策树 (核心流程)

### 需求管理流程

```
想法 → /propose → 填写提案 → 需求评审 →
  ├─ 批准 → /create-spec → 填写 requirements/design/tasks → SP评估 →
  │   ├─ SP ≤ 5 → 进 Sprint → 开发
  │   └─ SP > 5 → 拆分 Spec → 重新评估
  └─ 拒绝 → 归档
```

### 功能开发流程

```
读 specs/{}/tasks.md →
选择任务 →
读 requirements.md (关联需求) →
查 design.md (可复用代码) →
确定技术栈 (Go/PHP/Vue) →
AI 辅助实现 →
自测验证 →
标记完成 (tasks.md) →
提交代码 (Git)
```

### Bug 修复流程

```
问题现象 → 搜 Graphiti (历史问题) →
  ├─ 找到 → 复用方案 → 验证 → 提交
  └─ 未找到 →
      查日志 (SkyWalking) →
      定位代码 → 修复 → 测试 →
      记录 (docs/shared/troubleshooting/ + Graphiti Episode 模板) → 提交
```

### API 对接流程

```
查 integrations/{service}/ →
选择协议 (REST/gRPC) →
  ├─ REST: 创建 Service/Controller → Swagger 文档
  └─ gRPC: 定义 Protobuf → 生成代码 → 注册服务
编写测试 →
创建文档 (docs/shared/api/) →
验证功能
```

### 数据库迁移流程

```
设计表结构 →
创建迁移文件 (PHP Phinx) →
  date +%Y%m%d%H%M%S_add_table_name.php →
编写迁移代码 →
更新 Go model (main/app/model/) →
更新 seeds 文件 (admin/database/seeds/) →
执行迁移 (php think migrate:run) →
验证数据
```

---

## 关键约束 (硬性规则)

### Scrum 规则

- **只有 SP ≤ 5 的需求才能进 Sprint** (强制)
- SP > 5 必须拆分 (通常按模块拆分)
- Spec 命名: 一个模块 + 一个功能

### Go 代码规则

- 接口以 `I` 开头 (IUserService)
- 实现以 `Impl` 结尾 (UserServiceImpl)
- Repository 简写 Repo, Service 简写 Srv
- **不使用 panic**，返回 error
- URL 使用 snake_case (不用 kebab-case)
- try-catch 必须捕获 `error` 和 `stackTrace`
- 所有注释使用中文

### PHP 代码规则

- 数据库迁移文件统一在 PHP 中管理
- 时间字段: int 类型, \_time 结尾, 默认值 0
- 金额字段: decimal(20,8)
- 必须字段: uuid, create_time, update_time, delete_time
- 迁移前检查表/字段是否已存在

### API 响应规则

- data 字段不能是 null 或数组
- 分页信息统一放在 meta 中
- 响应格式: {code, message, data{list, meta}}
- 本地响应时间 < 200ms

---

## 快速命令映射

| 用户需求       | Agent 执行命令                           | 后续操作                           |
| -------------- | ---------------------------------------- | ---------------------------------- |
| 新成员入职     | `/onboard quick`                         | 推送必读清单和入职路径             |
| 创建提案       | `/propose quick-payment`                 | 自动创建并填充基本信息             |
| 创建 Spec      | `/create-spec story-order-quick-payment` | 自动创建 requirements/design/tasks |
| 创建数据库迁移 | 创建 PHP Phinx 迁移文件                  | 同步更新 Go model 和 seeds         |
| 创建 gRPC 服务 | 定义 Protobuf → 生成代码                 | 注册到 Nacos                       |
| 创建 API 文档  | `/create-api-doc order`                  | 自动扫描代码生成文档               |

---

## 关键文件清单 (按优先级)

### P0 (必读 - 薄层规则)

1. `AGENT.md` (本文件 - 核心速查)
2. `.cursor/rules/intro.mdc` (项目快速入门)
3. `.cursor/rules/structs.mdc` (项目结构快速定位)
4. `.cursor/rules/workflows.mdc` (工作流导航)
5. `.cursor/rules/go-main.mdc` (Go Main 核心约束)
6. `.cursor/rules/php.mdc` (PHP 核心约束)
7. `.cursor/rules/api.mdc` (API 核心约束)
8. `.cursor/rules/database.mdc` (数据库开发规范)
9. `.cursor/rules/security.mdc` (安全开发规范)
10. `.cursor/rules/vue.mdc` (Vue 前端开发规范)

### P1 (按需查阅 - 详细文档)

7. `docs/human/architecture/overview.md` (系统架构总览)
8. `docs/human/guides/go-main-development.md` (Go 详细开发指南)
9. `docs/human/guides/php-development.md` (PHP 详细开发指南)
10. `docs/human/guides/api-design-guide.md` (API 设计详细指南)
11. `docs/human/guides/database-guide.md` (数据库开发详细指南)
12. `ttpos-bmp/.cursor/rules/go-rules.mdc` (GoFrame 专用规范)
13. `ttpos-bmp/.cursor/rules/proto-rules.mdc` (Protobuf 专用规范)

### P2 (必要时 - 深入学习)

14. `docs/human/architecture/*.md` (各模块架构设计)
15. `docs/human/business/glossary.md` (业务术语表)
16. `docs/shared/specs/README.md` (功能规格说明)
17. `docs/shared/api/conventions.md` (API 规范速查)
18. `docs/agent/workflows/` (待补充 - Agent 工作流)
19. `docs/agent/templates/` (待补充 - Agent 模板)

---

## 工作原则

1. 优先查本文件
2. 最小化读取，只读必要文件
3. 尽量一次工具调用解决问题
4. 自动建立 Proposal ↔ Spec 链接
5. 根据 Sprint 节点和 IF-THEN 规则主动提醒
6. 耗时 >30 分钟的问题记录到 Graphiti
7. 数据库迁移必须同步更新 Go model 和 seeds

---

## 📝 文档创建指南（重要！）

### Agent 视角 vs 人类视角

**创建文档前，必须明确受众：**

```yaml
IF 文档主要给 Agent 阅读 THEN
  受众: 🤖 Agent
  位置: docs/agent/workflows/ 或 docs/agent/templates/
  风格: Agent 视角

ELSE IF 文档主要给人类学习 THEN
  受众: 👤 人类
  位置: docs/human/guides/ 或 docs/human/architecture/ 或 docs/human/business/
  风格: 人类视角

ELSE
  受众: 📚 共用
  位置: docs/shared/specs/ 或 docs/shared/api/ 或 docs/shared/troubleshooting/
  风格: 结构化 + 简洁
```

### 🤖 Agent 视角文档特征

```yaml
目标: 快速执行，精准定位
长度: <300行
结构:
  - 步骤检查清单 ✓
  - 决策树 (IF-THEN) ✓
  - 快速命令 ✓
  - 相关资源链接 ✓
风格:
  - YAML/表格/代码块优先
  - 无冗长解释
  - 无"为什么"
  - 只有"做什么"和"怎么做"

示例: ✅ "检查 tasks.md 全部 [x]"
  ✅ "IF SP > 5 THEN 拆分"
  ❌ "为了确保代码质量，我们需要..."
  ❌ "这样做的好处是..."
```

### 👤 人类视角文档特征

```yaml
目标: 深度理解，系统学习
长度: 不限
结构:
  - WHY: 设计原因 ✓
  - HOW: 详细步骤 ✓
  - EXAMPLE: 完整示例 ✓
  - TROUBLESHOOTING: 常见问题 ✓
风格:
  - 叙述性文字
  - 详细解释
  - 设计权衡
  - 最佳实践

示例: ✅ "我们选择 Gin 框架是因为..."
  ✅ "这个设计权衡考虑了..."
  ✅ "完整的代码示例如下..."
```

### 创建新文档时的决策树

```
需要创建文档 →
  谁会读? → Agent →
    放哪? →
      执行流程? → docs/agent/workflows/
      模板? → docs/agent/templates/
      指令? → .cursor/commands/
    风格? → Agent 视角 ✓

  谁会读? → 人类 →
    放哪? →
      学习教程? → docs/human/guides/
      架构设计? → docs/human/architecture/
      业务知识? → docs/human/business/
      示例参考? → docs/human/examples/
    风格? → 人类视角 ✓

  谁会读? → 都会读 →
    放哪? →
      功能规格? → docs/shared/specs/
      API文档? → docs/shared/api/
      问题排查? → docs/shared/troubleshooting/
      需求提案? → docs/team/proposals/
    风格? → 结构化 + 简洁 ✓
```

### ⚠️ 常见错误

```yaml
错误 1: Agent 文档写成教程
  - ❌ 600+ 行详细解释
  - ✅ <300 行执行清单

错误 2: 人类文档过于简洁
  - ❌ 只有命令，无解释
  - ✅ 详细说明 WHY 和 HOW

错误 3: 受众不明确
  - ❌ 混合 Agent 和人类内容
  - ✅ 明确标注受众 (🤖/👤/📚)

错误 4: 文档孤岛
  - ❌ 创建后未被引用
  - ✅ 在 README 或本文件中建立索引
```

---

## 📋 执行清单（创建文档时）

### Before 创建

- [ ] 明确受众 (🤖 Agent / 👤 人类 / 📚 共用)
- [ ] 确定位置 (workflows/ / guides/ / templates/ / ...)
- [ ] 确定风格 (Agent 视角 / 人类视角)

### During 创建

- [ ] 遵循对应风格
- [ ] Agent 文档: <300 行 + 结构化
- [ ] 人类文档: 详细 + 包含 WHY

### After 创建

- [ ] 在目录 README.md 中添加索引
- [ ] 在本文件中添加路径（如需要）
- [ ] 在相关工作流中建立引用
- [ ] 验证无孤岛文档
