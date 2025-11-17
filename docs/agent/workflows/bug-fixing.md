# Bug 修复工作流（后端版）

> 本文档定义 Bug 修复的完整流程

---

## 📋 概述

### 适用场景

- 生产环境 Bug
- 测试环境缺陷
- 用户反馈问题

### 预计时间

- 简单 Bug: 0.5-1 天
- 中等 Bug: 1-2 天
- 复杂/紧急 Bug: 需要 Hotfix 流程

---

## 完整流程

```
收到 Bug → 复现问题 → 根因分析 → 搜索历史方案 →
修复代码 → 编写测试 → 提交 PR → 记录经验
```

---

## 执行流程

### Step 1: 复现问题 (10-30 分钟)

#### 收集信息

- [ ] Bug 描述
- [ ] 复现步骤
- [ ] 预期行为 vs 实际行为
- [ ] 环境信息（操作系统/浏览器/版本）
- [ ] 错误日志/截图

#### 尝试复现

```bash
# 切换到对应版本
git checkout <tag_or_commit>

# 启动服务
cd main && go run main.go  # Go服务
cd admin && php think run  # PHP服务

# 按照复现步骤操作
```

#### 输出产物

- [ ] Bug 可稳定复现
- [ ] 最小复现步骤文档

---

### Step 2: 根因分析 (20-60 分钟)

#### 分析错误日志

**Go 服务日志**:

```go
// 查看日志
logger.Logger.Error/Warn/Info
```

**PHP 服务日志**:

```bash
tail -f runtime/log/*.log
```

#### 使用调试工具

- Delve (Go 调试器)
- Xdebug (PHP 调试器)
- Chrome DevTools (Web)
- 数据库查询日志

#### 代码分析

```bash
# 查找相关代码
codebase_search "错误关键词"

# 精确定位
grep -r "错误信息" main/app/
grep -r "错误信息" admin/app/

# 查看 Git Blame
git blame <file>
```

#### 根因分类

- [ ] 逻辑错误
- [ ] 边界条件处理不当
- [ ] 并发/竞态问题
- [ ] 数据异常
- [ ] 第三方依赖问题
- [ ] 数据库问题

---

### Step 3: 搜索历史方案 (5-15 分钟)

#### 搜索 Graphiti

```
query: "{错误关键词} fix solution"
group_id: "ttpos-golang" or "ttpos-php" or "ttpos-database"
```

#### 搜索 troubleshooting 文档

```bash
grep -r "相关错误信息" docs/shared/troubleshooting/
```

#### 搜索 Git 历史

```bash
git log --all --grep="相关关键词"
git log --all -- path/to/problem/file
```

---

### Step 4: 修复代码 (30 分钟 - 数小时)

#### 创建修复分支

```bash
# 普通 Bug
git checkout develop
git checkout -b fix/issue-{number}-{description}

# 紧急 Hotfix (从 release 分支)
git checkout release
git checkout -b hotfix/v{version}/{description}
```

#### 修复代码

**Go 示例**:

```go
// ✅ 正确的错误处理
func GetOrder(ctx context.Context, id uint64) (*model.Order, error) {
    if id == 0 {
        return nil, errors.New("订单ID不能为空")
    }

    order, err := repo.FindOrder(id)
    if err != nil {
        return nil, errors.WithMessage(err, "查询订单失败")
    }

    return order, nil
}
```

**PHP 示例**:

```php
// ✅ 正确的错误处理
public function getOrder($id) {
    if (empty($id)) {
        throw new InvalidArgumentException('订单ID不能为空');
    }

    try {
        $order = $this->orderModel->find($id);
        return $order;
    } catch (Exception $e) {
        Log::error('getOrder Error: ' . $e->getMessage());
        return null;
    }
}
```

#### 验证修复

- [ ] 原问题已解决
- [ ] 没有引入新问题
- [ ] 边界条件已处理
- [ ] 相关功能未受影响

#### 代码审查自查

- [ ] 符合代码规范
- [ ] 没有调试代码
- [ ] 错误处理完整
- [ ] 注释清晰（使用中文）

参考: `.cursor/rules/go-main.mdc`, `.cursor/rules/php.mdc`

---

### Step 5: 编写测试 (30-60 分钟)

#### 编写回归测试

**Go 测试示例**:

```go
func TestOrderService_GetOrder_FixBug123(t *testing.T) {
    // Given: 触发 Bug 的前置条件

    // When: 执行触发 Bug 的操作
    order, err := service.GetOrder(ctx, 0)

    // Then: 验证问题已修复
    assert.Error(t, err)
    assert.Nil(t, order)
    assert.Contains(t, err.Error(), "订单ID不能为空")
}
```

**PHP 测试示例**:

```php
public function testGetOrderWithInvalidId() {
    // Given
    $this->expectException(InvalidArgumentException::class);

    // When
    $this->orderService->getOrder(0);

    // Then: 异常已抛出
}
```

#### 运行测试

```bash
# Go
cd main && go test ./...

# PHP
cd admin && php think test
```

---

### Step 6: 提交 PR (15-30 分钟)

#### 提交代码

```bash
# 普通 Bug 修复
git commit -m "fix(order): 修复订单金额计算错误

- 问题: 折扣计算时未考虑小数精度
- 原因: 使用了浮点运算导致精度丢失
- 方案: 改用 Decimal 类型计算

Fixes #123"

# Hotfix
git commit -m "fix(payment): 紧急修复支付接口超时

- 紧急修复生产环境支付失败
- 增加超时重试机制
- 调整超时时间为 30s

Fixes #456"
```

#### PR 检查清单

- [ ] Commit 消息符合规范
- [ ] 关联了对应的 Issue
- [ ] 所有测试通过
- [ ] 代码审查通过

#### Hotfix 特殊流程

- 需要同时合并到 `release` 和 `develop`
- 创建两个 PR
- 优先合并 release，紧急发布
- 参考: `.cursor/rules/version.mdc`

---

### Step 7: 记录经验 (10-20 分钟)

#### 更新 troubleshooting 文档

```bash
# 创建新文档（如果是新问题）
docs/shared/troubleshooting/{category}/{issue}.md
```

#### 记录到 Graphiti

```yaml
name: "qa-{issue-keyword}-{YYYY-MM}"
group_id: "ttpos-golang" or "ttpos-php" or "ttpos-database"
episode_body: |
  问题: {一句话描述}

  环境: {出问题的环境和版本}

  复现: {最小复现步骤}

  原因: {根本原因}

  解决方案:
  1. {修复步骤1}
  2. {修复步骤2}

  预防措施: {如何避免}

  相关代码: {文件路径}:{行号}
  相关 Issue: #{issue_number}
```

---

## 检查清单

### Bug 分析

- [ ] Bug 可稳定复现
- [ ] 根本原因已确认
- [ ] 搜索过历史方案

### 代码修复

- [ ] 修复分支已创建
- [ ] 代码符合规范（Go/PHP/Vue）
- [ ] 原问题已解决
- [ ] 没有引入新问题
- [ ] 边界条件已处理

### 测试

- [ ] 回归测试已编写
- [ ] 所有测试通过
- [ ] 测试覆盖率达标

### 提交

- [ ] Commit 消息规范
- [ ] PR 已创建
- [ ] Code Review 通过
- [ ] PR 已合并

### 知识沉淀

- [ ] troubleshooting 文档已更新
- [ ] Graphiti 已记录

---

## 常见问题

### Q: Bug 无法复现怎么办？

**A**:

1. 检查环境差异（版本/配置/数据状态）
2. 要求用户提供详细信息
3. 查看 Sentry 或日志系统

### Q: 什么情况需要 Hotfix？

**A**:

- 生产环境严重 Bug (阻塞主流程)
- 安全漏洞
- 数据丢失/损坏风险

参考: `.cursor/rules/version.mdc`

---

## Hotfix 特殊流程

```bash
# 1. 从 release 创建 hotfix 分支
git checkout release
git checkout -b hotfix/v{version}/{description}

# 2. 修复代码
# ... 修改代码 ...

# 3. 提交
git commit -m "fix(payment): 紧急修复支付崩溃问题"

# 4. 创建 PR 到 release 和 develop
# PR 1: hotfix → release (优先)
# PR 2: hotfix → develop (同步)

# 5. 合并到 release 后立即打 tag
git checkout release
git merge hotfix/v{version}/{description}
git tag -a v{version} -m "Hotfix: 修复支付崩溃"
git push origin v{version}

# 6. 紧急发布
cd main && go build
cd admin && composer install --no-dev
# ... 部署 ...
```

---

## 相关资源

### 规范文件

- `.cursor/rules/go-main.mdc` - Go 规范
- `.cursor/rules/php.mdc` - PHP 规范
- `.cursor/rules/version.mdc` - Git 工作流

### 文档

- `docs/shared/troubleshooting/` - 问题排查指南
- `docs/shared/troubleshooting/common-issues.md` - 常见问题

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 提醒：Bug 修复完成后务必更新 troubleshooting 文档并沉淀 Episode，保持问题-解决方案可追溯。

---

**最后更新**: 2025-11-16  
**维护者**: 后端开发组
