# TTPOS 项目 Cursor 配置说明

本目录包含 TTPOS 项目的 Cursor AI 编辑器配置文件，用于规范开发流程、提高代码质量和团队协作效率。

## 📁 目录结构

```
.cursor/
├── README.mdc                         # 本文件，配置说明
├── rules/                             # 开发规范文件（.mdc格式）
│   ├── 00-project-overview.mdc       # 项目总览
│   ├── 01-go-main-rules.mdc          # Go Main模块开发规范
│   ├── 02-go-bmp-rules.mdc           # Go BMP模块开发规范（引用ttpos-bmp/.cursor/rules/）
│   ├── 03-php-rules.mdc              # PHP开发规范
│   ├── 04-vue-rules.mdc              # Vue前端开发规范
│   ├── 05-database-rules.mdc         # 数据库规范
│   ├── 06-api-rules.mdc              # API设计规范
│   ├── 07-testing-rules.mdc          # 测试规范
│   ├── 08-git-rules.mdc              # Git提交规范
│   ├── 09-docker-rules.mdc           # Docker与部署规范
│   ├── 10-security-rules.mdc         # 安全规范
│   └── 11-performance-rules.mdc      # 性能优化规范
├── templates/                         # 代码模板
│   ├── go/                           # Go代码模板
│   │   ├── service.go.tpl            # 服务层模板
│   │   ├── repository.go.tpl         # 仓储层模板
│   │   ├── controller.go.tpl         # 控制器模板
│   │   ├── model.go.tpl              # 模型模板
│   │   └── test.go.tpl               # 测试模板
│   ├── php/                          # PHP代码模板
│   │   ├── controller.php.tpl        # 控制器模板
│   │   ├── model.php.tpl             # 模型模板
│   │   ├── service.php.tpl           # 服务模板
│   │   └── validate.php.tpl          # 验证器模板
│   ├── vue/                          # Vue组件模板
│   │   ├── page.vue.tpl              # 页面组件模板
│   │   ├── component.vue.tpl         # 通用组件模板
│   │   └── api.ts.tpl                # API接口模板
│   ├── database/                     # 数据库模板
│   │   └── migration.php.tpl         # 迁移文件模板
│   └── docs/                         # 文档模板
│       ├── api-doc.mdc               # API文档模板
│       └── feature-doc.mdc           # 功能文档模板
├── snippets/                          # 代码片段
│   ├── go-snippets.json              # Go代码片段
│   ├── php-snippets.json             # PHP代码片段
│   └── vue-snippets.json             # Vue代码片段
├── checklists/                        # 检查清单
│   ├── code-review.mdc               # 代码审查清单
│   ├── feature-development.mdc       # 功能开发清单
│   ├── bug-fix.mdc                   # Bug修复清单
│   ├── refactor.mdc                  # 重构清单
│   └── deployment.mdc                # 部署清单
├── workflows/                         # 工作流程
│   ├── feature-workflow.mdc          # 功能开发流程
│   ├── bugfix-workflow.mdc           # Bug修复流程
│   ├── hotfix-workflow.mdc           # 紧急修复流程
│   └── release-workflow.mdc          # 版本发布流程
├── tools/                             # 工具脚本
│   ├── generate-migration.sh         # 生成迁移文件
│   ├── generate-service.sh           # 生成服务代码
│   ├── generate-api-doc.sh           # 生成API文档
│   └── check-code-quality.sh         # 代码质量检查
└── examples/                          # 示例代码
    ├── go-service-example/           # Go服务示例
    ├── php-controller-example/       # PHP控制器示例
    ├── vue-component-example/        # Vue组件示例
    └── api-design-example/           # API设计示例
```

## 🎯 配置目的

### 1. 统一开发规范
- 确保团队成员遵循一致的代码风格和最佳实践
- 减少代码审查中的争议和返工
- 提高代码可读性和可维护性

### 2. 提高开发效率
- 提供代码模板和片段，减少重复编写
- 标准化工作流程，减少决策时间
- 自动化常见任务，提升生产力

### 3. 保证代码质量
- 通过检查清单确保完整性
- 规范化测试流程
- 强化安全意识

### 4. 促进知识共享
- 文档化最佳实践
- 提供示例代码参考
- 降低新成员学习曲线

## 📋 使用指南

### 开发前
1. 阅读 `rules/00-project-overview.mdc` 了解项目架构
2. 根据开发模块查看对应的规范文件
3. 查看 `workflows/` 了解开发流程
4. 使用 `checklists/` 规划开发任务

### 开发中
1. 使用 `templates/` 快速生成代码框架
2. 参考 `examples/` 了解最佳实践
3. 使用 `snippets/` 提高编码效率
4. 遵循相应模块的开发规范

### 开发后
1. 使用 `checklists/code-review.mdc` 自查代码
2. 运行 `tools/check-code-quality.sh` 检查代码质量
3. 确保通过所有测试
4. 按照 `rules/08-git-rules.mdc` 提交代码

## 🔧 .mdc 文件格式说明

`.mdc` (Markdown Cursor) 是 Cursor 编辑器使用的规范文件格式，与普通 Markdown 相同，但专门用于 AI 辅助开发的规范定义。

### 特点
- 完全兼容 Markdown 语法
- 被 Cursor AI 识别为开发规范
- 可以包含代码示例、检查清单等
- 支持中文内容

### 最佳实践
- 使用清晰的标题层级
- 提供具体的代码示例
- 使用 ✅ ❌ 标记正确和错误的做法
- 包含详细的注释说明

## 📚 规范文件说明

### 核心规范（必读）
1. **00-project-overview.mdc** - 项目总览，了解整体架构
2. **01-go-main-rules.mdc** - Go Main模块规范（基于Gin框架）
3. **02-go-bmp-rules.mdc** - Go BMP模块规范（基于GoFrame框架）
4. **03-php-rules.mdc** - PHP开发规范（基于ThinkPHP）
5. **04-vue-rules.mdc** - Vue3前端开发规范

### 专项规范
6. **05-database-rules.mdc** - 数据库设计和迁移规范
7. **06-api-rules.mdc** - RESTful API设计规范
8. **07-testing-rules.mdc** - 单元测试和集成测试规范
9. **08-git-rules.mdc** - Git提交和分支管理规范
10. **09-docker-rules.mdc** - Docker容器化和部署规范
11. **10-security-rules.mdc** - 安全开发规范
12. **11-performance-rules.mdc** - 性能优化规范

## 🔗 与现有规范的关系

### ttpos-bmp 专用规范
`ttpos-bmp/.cursor/rules/` 目录下已有专用规范：
- `go-rules.mdc` - GoFrame开发规范
- `proto-rules.mdc` - Protobuf开发规范

这些规范专门针对 ttpos-bmp 项目，本目录的 `02-go-bmp-rules.mdc` 会引用和补充这些内容。

### 规范优先级
1. **模块专用规范** > **项目通用规范**
2. 开发 ttpos-bmp 时，优先参考 `ttpos-bmp/.cursor/rules/`
3. 开发 main 时，参考 `.cursor/rules/01-go-main-rules.mdc`
4. 开发 admin 时，参考 `.cursor/rules/03-php-rules.mdc` 和 `04-vue-rules.mdc`

## 🛠️ 工具使用

### 代码生成工具
```bash
# 生成数据库迁移文件
.cursor/tools/generate-migration.sh CreateUsersTable

# 生成服务代码
.cursor/tools/generate-service.sh UserService

# 生成API文档
.cursor/tools/generate-api-doc.sh
```

### 代码质量检查
```bash
# 检查代码质量
.cursor/tools/check-code-quality.sh

# 检查特定模块
.cursor/tools/check-code-quality.sh main
.cursor/tools/check-code-quality.sh admin
```

## 📝 更新和维护

### 更新原则
- 规范文件应保持稳定，重大变更需团队讨论
- 模板和工具可以根据实际需求灵活调整
- 所有更新都应该有明确的理由和说明
- 更新后需通知全体团队成员

### 更新流程
1. 提出更新建议（Issue或讨论会）
2. 团队评审和讨论
3. 达成共识后更新文件
4. 通知全体成员
5. 更新版本记录

### 版本管理
- 使用 Git 管理所有配置文件
- 重要更新打 tag
- 在文件底部记录更新历史

## 🤝 贡献指南

### 如何贡献
1. 发现规范不合理或有遗漏
2. 在团队会议上提出
3. 提交 PR 更新相关文件
4. 经过审核后合并

### 贡献内容
- 完善开发规范
- 添加代码模板
- 补充示例代码
- 改进工具脚本
- 更新文档说明

## 📞 相关资源

### 项目文档
- [项目README](/README.md) - 项目总体介绍
- [提交规范](/COMMIT_CONVENTION.md) - Git提交规范
- [重构方案](/docs/refactor/README.md) - 系统重构计划
- [API文档](http://localhost:8080/swagger/index.html) - Swagger API文档

### 外部资源
- [Cursor 官方文档](https://cursor.sh/docs)
- [Go语言规范](https://go.dev/doc/effective_go)
- [GoFrame框架文档](https://goframe.org.cn)
- [ThinkPHP文档](https://www.kancloud.cn/manual/thinkphp6_0)
- [Vue3文档](https://cn.vuejs.org/)

## 💡 快速开始

### 第一次使用
1. 阅读本 README
2. 查看 `rules/00-project-overview.mdc`
3. 根据你的开发任务，阅读对应的规范文件
4. 使用模板和工具开始开发

### 日常使用
1. 开发前查看相关规范
2. 使用代码模板快速生成框架
3. 开发中遵循规范要求
4. 提交前使用检查清单自查

## 📊 规范覆盖范围

| 模块 | 语言/框架 | 规范文件 | 状态 |
|------|----------|---------|------|
| Main | Go + Gin | 01-go-main-rules.mdc | ✅ |
| BMP | Go + GoFrame | 02-go-bmp-rules.mdc | ✅ |
| Admin | PHP + ThinkPHP | 03-php-rules.mdc | ✅ |
| Frontend | Vue3 + TypeScript | 04-vue-rules.mdc | ✅ |
| Database | MySQL | 05-database-rules.mdc | ✅ |
| API | RESTful | 06-api-rules.mdc | ✅ |
| Testing | Go/PHP/Vue | 07-testing-rules.mdc | ✅ |
| Git | Git Flow | 08-git-rules.mdc | ✅ |
| Docker | Docker Compose | 09-docker-rules.mdc | ✅ |
| Security | 全栈 | 10-security-rules.mdc | ✅ |
| Performance | 全栈 | 11-performance-rules.mdc | ✅ |

## 📝 版本历史

| 版本 | 日期 | 说明 | 负责人 |
|------|------|------|--------|
| v1.0.0 | 2025-11-16 | 初始版本，建立基础配置结构 | AI Assistant |

---

**最后更新**: 2025-11-16  
**维护者**: TTPOS Team  
**文件格式**: .mdc (Markdown Cursor)
