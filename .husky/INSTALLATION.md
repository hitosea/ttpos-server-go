# Husky 安装和配置指南

## 📋 概述

本指南帮助团队成员在本地开发环境中正确安装和配置 Husky Git Hooks。

## ✅ 前置要求

确保已安装以下软件：

- **Node.js**: >= 14.0.0
- **npm**: >= 6.0.0
- **Git**: >= 2.0.0

验证版本：

```bash
node --version   # 应输出 v14.0.0 或更高
npm --version    # 应输出 6.0.0 或更高
git --version    # 应输出 2.0.0 或更高
```

## 🚀 安装步骤

### 1. 克隆项目（新成员）

```bash
# 克隆项目
git clone <repository-url>
cd ttpos-server-go

# 安装依赖（会自动初始化 Husky）
npm install
```

**预期输出**:
```
> ttpos-server-go@1.0.0 prepare
> husky install

husky - Git hooks installed

up to date in XXXms
```

### 2. 已有项目（拉取最新代码）

```bash
# 拉取最新代码
git pull origin develop

# 安装依赖（如果 package.json 有更新）
npm install
```

### 3. 验证安装

```bash
# 检查 .husky 目录
ls -la .husky/

# 应该看到以下内容：
# drwxr-xr-x  _/
# -rwxr-xr-x  commit-msg
# -rw-r--r--  README.md

# 检查 Git hooks 路径
git config core.hooksPath
# 应输出: .husky
```

## 🧪 测试 Hook

### 测试 1: 分支命名检查

```bash
# 创建一个不符合规范的分支
git checkout -b test-branch

# 尝试提交
git commit --allow-empty -m "feat: 测试提交"

# ❌ 预期失败，输出：
# ❌ 分支命名不符合 git flow 约定：test-branch
```

### 测试 2: 提交信息格式检查

```bash
# 切换到符合规范的分支
git checkout -b feature/test

# 使用错误的提交格式
git commit --allow-empty -m "添加新功能"

# ❌ 预期失败，输出：
# ❌ 提交信息不符合 Conventional Commits 规范
```

### 测试 3: 正确的提交

```bash
# 使用正确的格式
git commit --allow-empty -m "feat(test): 测试 husky hook"

# ✅ 预期成功
```

## 🔧 常见问题

### Q1: npm install 后没有看到 "husky - Git hooks installed"

**原因**: prepare 脚本未执行

**解决方案**:
```bash
# 手动执行 prepare 脚本
npm run prepare

# 或重新安装
rm -rf node_modules
npm install
```

### Q2: Hook 不执行

**原因**: Git hooks 路径配置错误或文件权限问题

**解决方案**:
```bash
# 1. 检查 hooks 路径
git config core.hooksPath
# 如果不是 .husky，手动设置
git config core.hooksPath .husky

# 2. 检查文件权限
ls -la .husky/commit-msg
# 应该是 -rwxr-xr-x（可执行）

# 如果没有执行权限，添加
chmod +x .husky/commit-msg

# 3. 检查文件是否存在
cat .husky/commit-msg
```

### Q3: Windows 系统下 Hook 报错

**症状**: 
```
.husky/commit-msg: command not found
```

**解决方案**:

**方法 1: 使用 Git Bash**
```bash
# 确保使用 Git Bash 而不是 CMD 或 PowerShell
# 在 Git Bash 中执行 git commit
```

**方法 2: 使用 WSL**
```bash
# 在 WSL (Windows Subsystem for Linux) 中开发
# WSL 对 shell 脚本的支持更好
```

**方法 3: 检查换行符**
```bash
# 检查文件换行符格式
git config core.autocrlf true

# 重新克隆项目
```

### Q4: 我想临时跳过检查

**场景**: 紧急情况需要快速提交

**解决方案**:
```bash
# 使用 --no-verify 跳过所有 hooks
git commit --no-verify -m "你的提交信息"

# ⚠️ 警告：仅在紧急情况下使用，日常开发请遵循规范
```

### Q5: Hook 执行很慢

**原因**: Hook 脚本执行了耗时操作

**解决方案**:
```bash
# 查看 hook 脚本内容
cat .husky/commit-msg

# 如果包含耗时操作，考虑优化或移除
```

## 🔄 更新 Husky

### 更新到最新版本

```bash
# 1. 更新 package.json 中的版本
npm install husky@latest --save-dev

# 2. 重新初始化
npm run prepare

# 3. 验证更新
npm list husky
```

### 迁移到新版本（如果有 breaking changes）

参考 Husky 官方迁移指南：
- [Husky v7 迁移到 v8](https://typicode.github.io/husky/migrating-from-v7-to-v8.html)

## 📚 团队协作

### 新成员加入

1. 克隆项目后执行 `npm install`
2. 阅读 [COMMIT_CONVENTION.md](../COMMIT_CONVENTION.md)
3. 测试 hook 是否正常工作
4. 开始正常开发

### CI/CD 环境

在 CI/CD 环境中，通常不需要安装 Husky：

```bash
# 在 CI 中跳过 prepare 脚本
npm ci --ignore-scripts

# 或设置环境变量
HUSKY=0 npm install
```

### Docker 环境

Dockerfile 中通常不需要 Husky：

```dockerfile
# 跳过 prepare 脚本
RUN npm ci --ignore-scripts --only=production
```

## 🎯 最佳实践

1. **保持 Hook 简单快速**
   - Hook 应该在 1 秒内完成
   - 避免在 hook 中执行编译、测试等耗时操作

2. **提供友好的错误提示**
   - 当检查失败时，清晰说明原因
   - 提供修正建议和示例

3. **文档及时更新**
   - Hook 规则变更时，及时更新文档
   - 通知团队成员

4. **版本控制**
   - `.husky/` 目录提交到 Git
   - `node_modules/` 不提交

## 📖 参考资料

- [Husky 官方文档](https://typicode.github.io/husky/)
- [Conventional Commits 规范](https://www.conventionalcommits.org/)
- [Git Flow 工作流](https://nvie.com/posts/a-successful-git-branching-model/)
- [项目提交规范](../COMMIT_CONVENTION.md)

## 🆘 获取帮助

如果遇到问题：

1. 查看本文档的"常见问题"章节
2. 查看 [COMMIT_CONVENTION.md](../COMMIT_CONVENTION.md)
3. 联系项目维护者
4. 在团队群中提问

---

**最后更新**: 2025-10-27  
**维护者**: TTPOS Team  
**Husky 版本**: v8.0.3
