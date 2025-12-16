# 员工密码加密方式升级提案

**提案人**: 曾振华  
**提案日期**: 2025-12-16  
**提案类型**: 安全增强  
**优先级**: 🔴 高  
**影响范围**: 全局（员工登录、密码管理）

---

## 📋 背景

### 当前问题

目前系统中员工密码使用 **双重 MD5 + 固定盐值** 的加密方式：

```php
// PHP 实现
function salt_hash($password) {
    return md5(md5($password) . 'jjjshop_salt_2020');
}
```

```go
// Go 实现
func EncryptPassword(password string) string {
    firstMd5 := md5.Sum([]byte(password))
    return md5.Sum([]byte(hex.EncodeToString(firstMd5[:]) + "jjjshop_salt_2020"))
}
```

### 安全隐患

1. **MD5 已被证明不安全**：
   - 容易被彩虹表攻击
   - 计算速度快，易于暴力破解
   - 存在碰撞漏洞

2. **固定盐值**：
   - 所有用户使用同一个盐值 `jjjshop_salt_2020`
   - 一旦泄露，所有密码都可被批量破解

3. **合规风险**：
   - 不符合现代安全标准（OWASP、PCI DSS）
   - 可能影响系统安全审计

### 影响表和字段

| 数据库 | 表名 | 字段 | 用途 |
|--------|------|------|------|
| saas | ttpos_admin_user | password | 超管密码 |
| saas | ttpos_staff | password | 统一账号员工密码 |
| 商家库 | ttpos_staff | password | 门店员工密码 |
| 商家库 | ttpos_staff | permission_password | 权限密码 |
| 商家库 | ttpos_member | password | 会员密码（本期不处理） |

---

## 💡 解决方案

### 核心策略：**渐进式升级（原地升级）**

采用 **方案二：原地升级**，无需新增字段，通过密码前缀自动识别加密方式：

- **MD5 密码**：32 位小写十六进制字符串
- **bcrypt 密码**：以 `$2a$`、`$2b$` 或 `$2y$` 开头

### 升级流程

```
用户登录
    ↓
读取存储密码
    ↓
判断密码类型（前缀识别）
    ├─ bcrypt → 直接验证 → 登录成功
    └─ MD5 → MD5验证
         ├─ 验证成功 → 异步升级为bcrypt → 登录成功
         └─ 验证失败 → 返回密码错误
```

### 技术选型

- **Go**: `golang.org/x/crypto/bcrypt` (官方推荐)
- **PHP**: `password_hash()` / `password_verify()` (内置函数)
- **成本因子**: 默认 10（平衡安全性和性能）

### 升级范围

**本期处理**：
- ✅ saas.ttpos_admin_user.password（超管密码）
- ✅ saas.ttpos_staff.password（统一账号密码）
- ✅ 商家库.ttpos_staff.password（员工密码）
- ✅ 商家库.ttpos_staff.permission_password（权限密码）

**暂不处理**：
- ❌ 会员密码（ttpos_member.password）- 涉及范围大，单独规划

---

## 🎯 预期目标

### 安全性提升

1. **防暴力破解**：bcrypt 计算速度慢，增加破解成本
2. **自动加盐**：每个密码使用独立随机盐值
3. **抗彩虹表**：无法使用预计算表攻击
4. **符合标准**：满足 OWASP 安全要求

### 业务影响

1. **无感知升级**：用户登录时自动完成迁移
2. **零停机时间**：不需要维护窗口
3. **可回滚**：双验证逻辑保证兼容性
4. **渐进式**：按实际登录频率自然完成迁移

---

## 📊 风险评估

| 风险项 | 等级 | 缓解措施 |
|--------|------|----------|
| 性能影响 | 🟡 中 | bcrypt 比 MD5 慢，但登录频率低，影响可接受 |
| 兼容性问题 | 🟢 低 | 双验证逻辑保证平滑过渡 |
| token 生成 | 🟡 中 | PHP 部分代码使用 `md5($password)` 生成 token，需调整 |
| 长期未登录用户 | 🟢 低 | 保留双验证逻辑，随时可登录并升级 |

---

## 📅 实施计划

### 阶段一：准备（1天）
- 创建提案和技术设计文档
- 评审和确认方案

### 阶段二：开发（3-5天）
- 实现密码工具类（Go + PHP）
- 修改登录验证逻辑
- 修改密码修改逻辑
- 修改添加/更新员工逻辑
- 修改权限密码验证逻辑

### 阶段三：测试（2-3天）
- 单元测试
- 集成测试
- 登录场景测试
- 密码修改测试
- 边界情况测试

### 阶段四：上线（1天）
- 灰度发布
- 监控迁移进度
- 处理异常情况

### 阶段五：收尾（持续）
- 监控迁移完成率
- 长期未登录用户通知
- 文档归档

**预计总工期**: 7-10 个工作日

---

## 💬 讨论问题

1. **是否需要强制密码重置？**
   - 建议：不强制，自然升级即可

2. **token 生成策略如何调整？**
   - 方案 A：使用密码修改时间戳替代
   - 方案 B：保留 MD5 字段用于 token（过渡期）

3. **迁移完成标准是什么？**
   - 建议：90% 活跃用户完成迁移即可

4. **是否需要监控面板？**
   - 建议：可选，通过 SQL 查询统计即可

---

## 📚 参考资料

- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [bcrypt 原理](https://en.wikipedia.org/wiki/Bcrypt)
- [Go bcrypt 文档](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- [PHP password_hash 文档](https://www.php.net/manual/en/function.password-hash.php)

---

## ✅ 审批流程

- [ ] 技术负责人审批
- [ ] 安全团队审批
- [ ] 产品负责人确认
- [ ] 测试负责人确认

**审批结果**: 待审批

---

**提案状态**: 📝 待评审  
**下一步**: 创建详细的技术设计文档和任务分解
