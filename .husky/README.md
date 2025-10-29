# Husky Git Hooks 配置说明

本项目使用 [Husky](https://typicode.github.io/husky/) 管理 Git Hooks，自动化代码质量检查和提交规范验证。

## 📦 已安装的版本

- **Husky**: v8.0.3
- **Node.js**: >= 14.0.0

## 🎯 已配置的 Hooks

### commit-msg

在每次提交时自动验证提交信息是否符合规范。

**检查项**:
1. ✅ 分支命名检查（Git Flow 规范）
2. ✅ 提交信息格式检查（Conventional Commits 规范）
3. ✅ 自动跳过 Merge commit 和 Revert commit

**文件位置**: `.husky/commit-msg`

## 🚀 快速开始

### 首次安装

```bash
# 1. 克隆项目后，安装依赖
npm install

# 2. Husky 会自动初始化（通过 prepare 脚本）
# 输出: husky - Git hooks installed
```

### 验证安装

```bash
# 检查 .husky 目录是否存在
ls -la .husky/

# 应该看到以下文件：
# - _/              (Husky 内部文件)
# - commit-msg      (提交信息检查)
# - README.md       (本文件)
```

## 📝 使用示例

### 正确的提交流程

```bash
# 1. 确保在正确的分支上
git checkout -b feature/add-email-service

# 2. 进行代码修改
# ...

# 3. 提交代码（符合规范）
git add .
git commit -m "feat(message): 添加 Mailgun 邮件发送功能"

# ✅ 提交成功
```

### 错误示例与修正

#### 示例 1: 分支命名不符合规范

```bash
# ❌ 错误：在不符合规范的分支上提交
git checkout -b my-feature
git commit -m "feat: 添加新功能"

# 输出错误：
# ❌ 分支命名不符合 git flow 约定：my-feature
```

**解决方案**:
```bash
# 重命名分支
git branch -m my-feature feature/my-feature
```

#### 示例 2: 提交信息格式错误

```bash
# ❌ 错误：缺少 type
git commit -m "添加用户登录功能"

# 输出错误：
# ❌ 提交信息不符合 Conventional Commits 规范
```

**解决方案**:
```bash
# 使用正确的格式
git commit -m "feat(auth): 添加用户登录功能"
```

#### 示例 3: type 拼写错误

```bash
# ❌ 错误：type 使用了中文
git commit -m "新功能(auth): 添加用户登录"

# ❌ 错误：type 拼写错误
git commit -m "add(auth): 添加用户登录"
```

**解决方案**:
```bash
# 使用正确的英文 type
git commit -m "feat(auth): 添加用户登录功能"
```

## 🔍 Hook 执行流程

```
开发者执行 git commit
         ↓
   触发 commit-msg hook
         ↓
   检查当前分支命名 ──→ 不符合 ──→ 拒绝提交 ❌
         ↓ 符合
   跳过 Merge/Revert? ──→ 是 ──→ 允许提交 ✅
         ↓ 否
   检查提交信息格式 ──→ 不符合 ──→ 拒绝提交 ❌
         ↓ 符合
      允许提交 ✅
```

## ⚙️ 配置文件说明

### package.json

```json
{
  "scripts": {
    "prepare": "husky install"
  },
  "devDependencies": {
    "husky": "^8.0.3"
  }
}
```

- `prepare`: 在 `npm install` 后自动执行，初始化 Husky
- Husky 作为开发依赖，不会影响生产环境

### .husky/commit-msg

Git hook 脚本，在每次提交时执行：

```bash
#!/usr/bin/env sh
. "$(dirname -- "$0")/_/husky.sh"

# 跳过 merge commit 的检查
if git rev-parse -q --verify MERGE_HEAD >/dev/null; then
  exit 0
fi

# ... 检查逻辑
```

## 🛠️ 维护和管理

### 添加新的 Hook

```bash
# 创建新的 hook（以 pre-commit 为例）
npx husky add .husky/pre-commit "npm test"

# 或手动创建
cat > .husky/pre-commit << 'EOF'
#!/usr/bin/env sh
. "$(dirname -- "$0")/_/husky.sh"

# 你的检查脚本
npm test
EOF

# 添加可执行权限
chmod +x .husky/pre-commit
```

### 禁用某个 Hook

临时禁用（不推荐）：

```bash
# 方法 1: 使用 --no-verify
git commit --no-verify -m "临时跳过检查"

# 方法 2: 临时删除 hook
mv .husky/commit-msg .husky/commit-msg.bak
```

永久禁用：

```bash
# 删除对应的 hook 文件
rm .husky/commit-msg
```

### 更新 Husky

```bash
# 更新到最新版本
npm install husky@latest --save-dev

# 重新初始化
npm run prepare
```

## 📊 Hook 检查规则详解

### 1. 分支命名检查

**允许的分支格式**:
- `feature/*` - 新功能分支
- `hotfix/*` - 紧急修复分支
- `release/*` - 发布分支
- `develop` - 开发主分支
- `main` - 生产主分支

**示例**:
```bash
✅ feature/add-email-service
✅ hotfix/fix-payment-bug
✅ release/v2.0.0
✅ develop
✅ main

❌ my-feature
❌ bugfix/fix-error
❌ test-branch
```

### 2. 提交信息检查

**格式要求**:
```
<type>(<scope>): <subject>
```

**详细规则**:
- `type`: 必需，只能是规定的 10 种类型之一
- `scope`: 可选，用括号包裹
- `:`: 必需，冒号后必须有空格
- `subject`: 必需，1-72 个字符

**正则表达式**:
```regex
^(feat|fix|build|ci|docs|style|refactor|test|chore|perf)(\(.+\))?: .{1,72}$
```

**示例**:
```bash
✅ feat(auth): 添加用户登录功能
✅ fix: 修复订单计算错误
✅ docs(api): 更新 API 文档
✅ refactor(service): 重构用户服务

❌ 添加用户登录                    # 缺少 type
❌ add(auth): 添加用户登录         # type 错误
❌ feat(auth):添加用户登录         # 冒号后缺少空格
❌ feat(auth): 修                 # subject 太短
```

### 3. 自动跳过检查

以下情况会自动跳过检查：

**Merge Commit**:
```bash
git merge feature/some-feature
# 自动生成: Merge branch 'feature/some-feature' into develop
# ✅ 自动跳过检查
```

**Revert Commit**:
```bash
git revert abc123
# 自动生成: Revert "feat(auth): 添加用户登录功能"
# ✅ 自动跳过检查
```

## 🐛 故障排查

### 问题 1: Hook 未执行

**症状**: 提交时没有看到任何检查提示

**解决方案**:
```bash
# 1. 检查 Husky 是否已安装
ls -la .husky/

# 2. 重新安装
rm -rf node_modules .husky
npm install

# 3. 检查 hook 文件权限
chmod +x .husky/commit-msg

# 4. 检查 Git hooks 路径
git config core.hooksPath
# 应该输出: .husky
```

### 问题 2: 提交总是失败

**症状**: 每次提交都被拒绝

**解决方案**:
```bash
# 1. 查看具体错误信息
git commit -m "你的提交信息"
# 仔细阅读错误提示

# 2. 验证分支名称
git branch --show-current

# 3. 验证提交信息格式
# 确保格式为: <type>(<scope>): <subject>

# 4. 临时跳过检查（仅紧急情况）
git commit --no-verify -m "feat: 紧急修复"
```

### 问题 3: Windows 系统兼容性

**症状**: Windows 下 hook 无法执行

**解决方案**:
```bash
# 1. 确保使用 Git Bash 或 WSL

# 2. 检查文件换行符
git config core.autocrlf true

# 3. 重新安装 Husky
npm install
```

## 📖 最佳实践

1. **提交信息要清晰明了**
   - 简短描述做了什么（不超过 72 字符）
   - 必要时在 body 中补充详细说明

2. **合理使用 scope**
   - 使用项目中已有的模块名称
   - 保持一致性，便于后续查找

3. **遵循分支命名规范**
   - 使用有意义的分支名称
   - 避免使用拼音或随意命名

4. **小步提交**
   - 每个提交只做一件事
   - 避免一次提交包含多个不相关的修改

5. **及时同步**
   - 定期从 develop 分支拉取最新代码
   - 减少合并冲突

## 🔗 相关文档

- [提交规范详细说明](../COMMIT_CONVENTION.md)
- [Git Flow 工作流程](../docs/GIT_WORKFLOW.md)（如果有）

---

**配置时间**: 2025-10-27  
**维护者**: TTPOS Team  
**版本**: v1.0.0
