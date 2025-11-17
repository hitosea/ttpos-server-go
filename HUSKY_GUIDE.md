# Husky Git Hooks 快速指南

## 🎯 什么是 Husky？

Husky 是一个 Git Hook 管理工具，用于在代码提交时自动执行检查，确保代码质量和提交规范。

## ✅ 快速开始

### 首次使用（3 步完成）

```bash
# 1. 克隆项目后，安装依赖
npm install

# 2. 验证安装（应该看到 husky - Git hooks installed）
ls -la .husky/

# 3. 开始使用（正常提交即可）
git add .
git commit -m "feat(module): 添加新功能"
```

## 📝 提交规范速查

### 提交信息格式

```
<type>(<scope>): <subject>
```

### 常用类型（type）

| 类型       | 说明        | 示例                           |
| ---------- | ----------- | ------------------------------ |
| `feat`     | ✨ 新功能   | `feat(auth): 添加用户登录`     |
| `fix`      | 🐛 修复 Bug | `fix(order): 修复订单计算错误` |
| `docs`     | 📝 文档     | `docs: 更新 README`            |
| `refactor` | ♻️ 重构     | `refactor(user): 优化用户服务` |
| `perf`     | ⚡ 性能     | `perf(query): 优化数据库查询`  |
| `style`    | 💄 格式     | `style: 统一代码格式`          |
| `test`     | ✅ 测试     | `test(api): 添加 API 测试`     |
| `chore`    | 🔧 其他     | `chore: 更新依赖包`            |

### 分支命名规范

```bash
✅ feature/add-email-service    # 新功能
✅ hotfix/fix-payment-bug       # 紧急修复
✅ release/v2.0.0               # 发布
✅ develop                      # 开发分支
✅ main                         # 主分支

❌ my-feature                   # 不符合规范
❌ test                         # 不符合规范
```

## 💡 常见示例

### 正确示例 ✅

```bash
# 新功能
git commit -m "feat(message): 添加 Mailgun 邮件发送功能"

# 修复 Bug
git commit -m "fix(order): 修复订单金额计算错误"

# 文档更新
git commit -m "docs: 更新环境变量配置说明"

# 性能优化
git commit -m "perf(cache): 添加 Redis 缓存"

# 不需要 scope
git commit -m "docs: 更新 README"
```

### 错误示例 ❌

```bash
# ❌ 缺少 type
git commit -m "添加用户登录功能"

# ❌ type 使用中文
git commit -m "新功能(auth): 添加登录"

# ❌ type 错误
git commit -m "add(auth): 添加登录"

# ❌ 冒号后缺少空格
git commit -m "feat(auth):添加登录"
```

## 🚨 提交被拒绝？

### 错误 1: 分支命名不符合规范

```bash
❌ 分支命名不符合 git flow 约定：test-branch
```

**解决方案**:

```bash
# 重命名分支
git branch -m test-branch feature/test-branch
```

### 错误 2: 提交信息格式错误

```bash
❌ 提交信息不符合 Conventional Commits 规范
```

**解决方案**:

```bash
# 使用正确格式重新提交
git commit -m "feat(module): 正确的提交信息"

# 或修改上次提交信息
git commit --amend -m "feat(module): 正确的提交信息"
```

## 🔧 紧急情况

如果遇到紧急情况需要快速提交（不推荐）：

```bash
# 跳过 hook 检查
git commit --no-verify -m "你的提交信息"

# ⚠️ 警告：仅在特殊情况下使用
```

## 📖 详细文档

- [Husky 配置说明](./.husky/README.md)
- [安装指南](./.husky/INSTALLATION.md)

## 🆘 需要帮助？

1. 查看错误提示信息
2. 参考本文档的示例
3. 查看详细文档
4. 联系项目维护者

---

**最后更新**: 2025-10-27  
**维护者**: TTPOS Team
