# Git 提交规范

本项目使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范和 Git Flow 分支管理策略。

## 📋 提交信息格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

### 必需部分

- **type**: 提交类型（必需）
- **subject**: 简短描述（必需，1-72个字符）

### 可选部分

- **scope**: 影响范围（可选）
- **body**: 详细描述（可选）
- **footer**: 备注信息（可选）

## 🏷️ 提交类型（Type）

| 类型 | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat(auth): 添加用户登录功能` |
| `fix` | 修复 Bug | `fix(order): 修复订单金额计算错误` |
| `docs` | 文档变更 | `docs: 更新 API 文档` |
| `style` | 代码格式（不影响代码运行） | `style: 统一代码缩进格式` |
| `refactor` | 重构（既不是新增功能，也不是修复bug） | `refactor(service): 重构用户服务层代码` |
| `perf` | 性能优化 | `perf(query): 优化数据库查询性能` |
| `test` | 测试相关 | `test(user): 添加用户模块单元测试` |
| `build` | 构建系统或外部依赖的变更 | `build: 升级 GoFrame 到 v2.9.0` |
| `ci` | CI 配置文件和脚本的变更 | `ci: 添加自动部署脚本` |
| `chore` | 其他不修改 src 或测试文件的变更 | `chore: 更新 .gitignore` |

## 📦 影响范围（Scope）

scope 用于说明 commit 影响的范围，建议使用项目中的模块名称：

**ttpos-server-go 项目常用 scope：**
- `auth` - 认证模块
- `order` - 订单模块
- `payment` - 支付模块
- `product` - 商品模块
- `user` - 用户模块
- `message` - 消息模块
- `erp` - ERP 模块
- `config` - 配置相关
- `db` - 数据库相关
- `api` - API 接口

## ✍️ 提交信息示例

### 基础示例

```bash
# 新功能
git commit -m "feat(auth): 添加 JWT 令牌刷新功能"

# 修复 Bug
git commit -m "fix(order): 修复订单状态更新失败的问题"

# 文档更新
git commit -m "docs: 更新 README 安装说明"

# 代码格式
git commit -m "style: 统一使用 gofmt 格式化代码"

# 重构
git commit -m "refactor(service): 优化用户服务代码结构"

# 性能优化
git commit -m "perf(cache): 添加 Redis 缓存提升查询性能"

# 测试
git commit -m "test(payment): 添加支付模块集成测试"

# 构建
git commit -m "build: 升级依赖包版本"

# 其他
git commit -m "chore: 更新 .gitignore 忽略规则"
```

### 包含详细描述的示例

```bash
git commit -m "feat(message): 添加 Mailgun 邮件发送功能

- 集成 Mailgun SDK
- 实现邮件发送服务
- 添加消息队列支持
- 完善错误处理和日志记录

相关文档: docs/MAILGUN_USAGE.md"
```

### Revert 提交

恢复之前的提交时，可以使用 `Revert` 前缀：

```bash
git revert <commit-hash>
# 自动生成: Revert "feat(auth): 添加用户登录功能"
```

## 🌿 分支命名规范

本项目使用 Git Flow 分支管理策略：

| 分支类型 | 命名格式 | 说明 | 示例 |
|---------|----------|------|------|
| 主分支 | `main` | 生产环境分支 | `main` |
| 开发分支 | `develop` | 开发主分支 | `develop` |
| 功能分支 | `feature/*` | 新功能开发 | `feature/user-login` |
| 修复分支 | `hotfix/*` | 紧急 Bug 修复 | `hotfix/order-calculation` |
| 发布分支 | `release/*` | 版本发布准备 | `release/v2.0.0` |

### 分支命名建议

```bash
# 功能分支
feature/add-user-authentication
feature/mailgun-integration
feature/erp-sync

# 修复分支
hotfix/fix-order-amount-bug
hotfix/fix-payment-timeout

# 发布分支
release/v2.0.0
release/v2.1.0-beta
```

## 🚫 错误示例

以下是一些不符合规范的提交信息：

```bash
# ❌ 缺少 type
git commit -m "添加用户登录功能"

# ❌ type 错误
git commit -m "add(auth): 添加用户登录功能"

# ❌ 缺少 subject
git commit -m "feat(auth):"

# ❌ subject 太短
git commit -m "feat: 修复"

# ❌ 使用中文 type
git commit -m "新功能(auth): 添加用户登录"
```

## 🔧 本地验证

在提交之前，可以使用以下命令验证提交信息格式：

```bash
# 查看最后一次提交信息
git log -1 --pretty=%B

# 修改最后一次提交信息
git commit --amend
```

## ⚙️ Husky Hook 说明

项目已配置 Husky Git Hook，在每次 `git commit` 时会自动检查：

1. **分支命名检查**: 确保当前分支符合 Git Flow 规范
2. **提交信息检查**: 确保提交信息符合 Conventional Commits 规范
3. **跳过检查**: Merge commit 和 Revert commit 会自动跳过检查

如果检查不通过，提交会被拒绝，并显示错误提示。

## 📚 参考资料

- [Conventional Commits 官方文档](https://www.conventionalcommits.org/)
- [Git Flow 工作流程](https://nvie.com/posts/a-successful-git-branching-model/)
- [Angular 提交信息规范](https://github.com/angular/angular/blob/master/CONTRIBUTING.md#commit)

## 💡 常见问题

### Q: 为什么需要遵循提交规范？

A: 
- 📊 **自动生成 CHANGELOG**: 可以根据提交记录自动生成版本日志
- 🔍 **快速查找**: 通过 type 快速定位特定类型的提交
- 📈 **代码审查**: 清晰的提交信息便于代码审查
- 🤝 **团队协作**: 统一的规范提高团队协作效率

### Q: 如果提交信息写错了怎么办？

A: 可以使用 `git commit --amend` 修改最后一次提交信息：

```bash
# 修改最后一次提交信息
git commit --amend -m "feat(auth): 正确的提交信息"
```

### Q: 如何临时跳过 Hook 检查？

A: 不建议跳过检查，但紧急情况下可以使用：

```bash
git commit --no-verify -m "你的提交信息"
```

**注意**: 仅在特殊情况下使用，日常开发请遵循规范。

### Q: Merge 和 Revert 提交需要遵循规范吗？

A: 不需要，Husky Hook 会自动跳过这两种类型的提交检查。

---

**最后更新**: 2025-10-27  
**维护者**: TTPOS Team
