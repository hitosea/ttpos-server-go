# Agent 速查表

> 本文件包含所有核心规则的压缩映射表。遇到任何问题优先查阅此文件。

---

## 场景识别 → 执行命令

| 用户说...                      | 识别为     | 立即执行                                                               | 涉及文件/目录                  | 参考规范        |
| ------------------------------ | ---------- | ---------------------------------------------------------------------- | ------------------------------ | --------------- |
| "新人" "入职" "不熟悉"         | 新成员入职 | `/onboard quick` → 推送必读清单                                        | agent/workflows/onboarding.md  | intro.mdc       |
| "有个想法" "提需求" "能不能做" | 需求发起   | `/spec-propose {name}` → 评审 → `/spec-create` → 审核 → `/spec-design`      | team/proposals/, shared/specs/ | specs.mdc       |
| "实现功能" "开发 XX" "新增 XX" | 功能开发   | 读 `shared/specs/active/{}/tasks.md` → 逐任务执行                      | shared/specs/, main/, admin/   | go-main/php.mdc |
| "报错" "bug" "崩溃" "异常"     | Bug 修复   | `/bug-create` → 分析 → `/bug-spec` → 修复 → `/bug-archive` → Graphiti | shared/bugs/, main/            | go-main.mdc     |
| "集成 XX" "对接 XX" "API"      | 第三方集成 | 查 `integrations/{service}/` → 创建 API 类 → 测试 → 文档               | integrations/, shared/api/     | api.mdc         |
| "迁移数据库" "新增表" "改字段" | 数据库迁移 | 创建迁移文件 → 更新 model → 更新 seeds                                 | admin/database/migrations/     | database.mdc    |
| "gRPC" "微服务" "ttpos-bmp"    | 微服务集成 | 查 ttpos-bmp 文档 → 定义 Protobuf → 注册服务                           | ttpos-bmp/                     | go-bmp.mdc      |
| "慢" "卡顿" "优化性能"         | 性能优化   | 分析瓶颈 → 优化 → 验证 → 记录                                          | 相关代码                       | go-main.mdc     |
| "前端开发" "Vue 组件" "页面"   | 前端开发   | 读 `shared/specs/active/{}/tasks.md` → 实现组件 → 测试                 | admin/views/                   | vue.mdc         |
| "安全审查" "漏洞" "SQL 注入"   | 安全检查   | 查 security.mdc → 检查代码 → 修复 → 测试                               | 所有代码                       | security.mdc    |
| "提交代码" "git commit"        | Git 提交   | 查 version.mdc → 写提交信息 → 推送                                     | .git/                          | version.mdc     |
| "新人" "入职" "不熟悉"         | 新成员     | `/onboard quick` → 推送必读清单                                        | .cursor/rules/, docs/          | intro.mdc       |

---

## 跨仓库协作

当需要查阅前端代码、接口定义、交互实现或跨仓库引用时：

1. 读取根目录下的 `.agents` 文件
2. 获取 `FRONTEND_PATH` 等配置的绝对路径
3. 使用 `read_file` 或 `list_dir` 或其他 MCP 服务的工具访问该绝对路径下的文件

---

## 规范速链

| 主题               | 查看                                                                                                             |
| ------------------ | ---------------------------------------------------------------------------------------------------------------- |
| 项目入门与结构     | `.cursor/rules/intro.mdc` · `.cursor/rules/structs.mdc`                                                          |
| 工作流与知识管理   | `.cursor/rules/workflows.mdc` · `.cursor/rules/knowledge_management.mdc`                                         |
| 命名/Go 主项目规范 | `.cursor/rules/go-main.mdc`                                                                                      |
| 打印模块规范       | `.cursor/rules/go-printer.mdc`                                                                                   |
| PHP 与数据库规范   | `.cursor/rules/php.mdc` · `.cursor/rules/database.mdc`                                                           |
| API 与安全规范     | `.cursor/rules/api.mdc` · `.cursor/rules/security.mdc`                                                           |
| 前端/Vue 规范      | `.cursor/rules/vue.mdc`                                                                                          |
| BMP/Proto 规范     | `ttpos-bmp/.cursor/rules/go-rules.mdc` · `ttpos-bmp/.cursor/rules/proto-rules.mdc` · `ttpos-bmp/.cursor/rules/*` |
| 文档与版本规范     | `.cursor/rules/documentation.mdc` · `.cursor/rules/version.mdc`                                                  |

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
| **执行任务**              | `docs/shared/specs/active/*/tasks.md`                  | 任务逐条执行                   |
| **创建需求提案**          | `docs/team/proposals/{YYYY-MM}/{name}.md`              | `/spec-propose`                     |
| **创建需求文档**          | `docs/shared/specs/active/{level}-{module}-{feature}/requirements.md` | `/spec-create`                 |
| **创建设计文档**          | `docs/shared/specs/active/{level}-{module}-{feature}/design.md + tasks.md` | `/spec-design`                 |
| **归档 Spec**             | `docs/shared/specs/archived/{version}/`                | `/spec-archive`                |
| **废弃 Spec**             | `docs/shared/specs/deprecated/`                        | `/spec-deprecate`              |
| **创建 Bug 报告**         | `docs/shared/bugs/active/bug-{id}-{module}-{brief}/bug.md` | `/bug-create`             |
| **创建 Bug 修复方案**     | `docs/shared/bugs/active/bug-{id}-{module}-{brief}/solution.md + tasks.md` | `/bug-spec` |
| **归档 Bug**              | `docs/shared/bugs/resolved/{version}/`                 | `/bug-archive`                 |

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
| **功能规格**   | `docs/shared/specs/active/story-*-*/` | 需求和设计 |
| **Bug 管理**   | `docs/shared/bugs/active/` / `resolved/{version}/` | Bug 报告   |
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
│   ├── bugs/           #    Bug管理
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
1. 本速查表 (AGENTS.md)
2. Graphiti (经验类问题: "如何" "为什么" "踩坑")
3. .cursor/rules/*.mdc (规范核心清单)
4. docs/ (详细文档、完整示例)
5. codebase_search (实现细节、代码位置)
6. web_search (最新技术、第三方文档)
```

---

## 技术栈识别优先级 (⚠️ 重要)

**在分析 Bug、功能开发、代码修改时，必须按以下顺序确定技术栈：**

```
1. 优先查找 Go Main 模块 (main/app/) - 核心业务
2. 其次查找 PHP Admin 模块 (admin/app/) - 仅当 Go 找不到
3. 最后查找 Go BMP 模块 (ttpos-bmp/) - 微服务
```

**核心原则**：
- ✅ 先搜索 Go Main，找不到再搜索 PHP
- ❌ 不要看到"管理端"就认为是 PHP
- ✅ 根据实际代码确定技术栈

**详细识别方法、常见误区和实战案例**，请查阅:
→ `docs/agent/workflows/development/bug-fixing.md` Step 2 技术栈识别
