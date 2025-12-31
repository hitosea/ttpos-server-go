# 密码 bcrypt 升级 - 完整影响范围分析

**分析日期**: 2025-12-16  
**分析人**: 曾振华  
**关联文档**: requirements.md, design.md, tasks.md

---

## 📊 PHP 代码影响汇总

通过全代码搜索 `salt_hash` 方法，发现：
- **15 个文件**使用了 `salt_hash`
- **35 处代码**需要修改
- **涉及 6 个数据库表**

---

## 🗄️ 数据库表影响范围

### P0 - 必须处理（核心业务）

| 数据库 | 表名 | 字段 | 用途 | 涉及文件数 | 影响代码行 |
|--------|------|------|------|------------|-----------|
| saas | ttpos_admin_user | password | 超管密码 | 2 | 5 |
| saas | ttpos_staff | password | 统一账号员工密码 | 4 | 8 |
| shop_* | ttpos_staff | password | 门店员工密码（含供应商、收银员） | 3 | 7 |
| shop_* | ttpos_staff | permission_password | 权限密码 | 2 | 3 |

**小计**: 3 个表（去重），11 个文件，23 处代码

**说明**：商家数据库中不存在独立的 `ttpos_supplier_user` 和 `ttpos_cashier_user` 表。供应商和收银员账号都存储在 `ttpos_staff` 表中，通过统一的员工管理接口处理。

### 暂不处理

| 数据库 | 表名 | 字段 | 用途 | 原因 |
|--------|------|------|------|------|
| shop_* | ttpos_member | password | 会员密码 | 涉及范围大，单独规划 |

---

## 📁 文件影响详细清单

### 1. 核心员工/用户管理（P0 - 必须）

#### admin/app/shop/model/shop/User.php
**功能**: 店铺员工登录和密码管理  
**影响代码**:
- 行 30: `checkLogin()` - 登录验证
- 行 86: `editPass()` - 验证旧密码
- 行 94: `editPass()` - 加密新密码
- 行 118: `saasEditPassword()` - 验证旧密码
- 行 131: `saasEditPassword()` - 加密新密码

**修改方案**:
```php
// 修改前：直接 MD5 比对
->where('password', salt_hash($password))

// 修改后：双验证
list($isValid, $needUpgrade) = verify_password($password, $user['password']);
if ($needUpgrade) {
    upgrade_password_async($user['uuid'], 'ttpos_staff', 'uuid', $password);
}
```

---

#### admin/app/admin/model/admin/User.php
**功能**: 超管登录和密码管理  
**影响代码**:
- 行 21: `login()` - 登录验证
- 行 58: `renew()` - 验证旧密码
- 行 63: `renew()` - 检查新旧密码相同
- 行 68: `renew()` - 加密新密码
- 行 113: `add()` - 添加新超管
- 行 147: `edit()` - 更新超管密码

**修改方案**: 同上

---

#### admin/app/shop/model/auth/User.php
**功能**: 员工认证管理（含权限密码）  
**影响代码**:
- 行 123: `add()` - 加密 password
- 行 124: `add()` - 加密 permission_password
- 行 228: `edit()` - 更新 password
- 行 238: `edit()` - 更新 permission_password

**特殊性**: 同时处理两个密码字段

---

#### admin/app/admin/model/admin/Staff.php
**功能**: 统一账号员工管理  
**影响代码**:
- 行 251: `add()` - 添加员工
- 行 330: `edit()` - 更新员工密码

---

#### admin/app/admin/model/CompanyStaff.php
**功能**: 商家员工创建（同步到 saas.ttpos_staff）  
**影响代码**:
- 行 90: `add()` - 创建员工并同步

**特殊性**: 需要确保 saas 库同步正确

---

#### admin/app/shop/controller/Passport.php
**功能**: 登录控制器  
**影响代码**:
- 行 64: `login()` - 登录前加密密码

**修改方案**: 移除控制器层的密码加密，交给 Model 层处理

---

### 2. 应用/系统管理（P1 - 重要）

#### admin/app/admin/model/app/App.php
**功能**: 应用管理员账号  
**影响代码**:
- 行 319: `add()` - 创建应用管理员
- 行 360: `edit()` - 更新应用管理员密码

---

#### admin/app/admin/controller/Erpnext.php
**功能**: ERPNext 集成密码验证  
**影响代码**:
- 行 77: 验证用户密码（用于 ERP 同步）

---

### 3. 供应商和收银员（已包含在 ttpos_staff 表处理中）

#### admin/app/shop/model/supplier/Supplier.php
**功能**: 供应商账号管理  
**影响代码**:
- 行 63: `add()` - 创建供应商账号（写入 ttpos_staff 表）
- 行 116: `edit()` - 更新供应商密码（更新 ttpos_staff 表）

**说明**：此文件操作的是 `ttpos_staff` 表，通过修改核心的 `shop/User.php` 登录验证逻辑和 `auth/User.php` 员工管理逻辑，供应商的登录和密码管理会自动覆盖。**无需单独处理**。

#### admin/app/cashier/model/cashier/User.php
**功能**: 收银员密码管理  
**影响代码**:
- 行 28: `checkLogin()` - 登录验证（继承自 UserModel，查询 ttpos_staff 表）
- 行 123: `editPass()` - 验证旧密码
- 行 127: `editPass()` - 加密新密码

**说明**：此文件继承自 `shop/User.php`，操作的是 `ttpos_staff` 表。通过修改核心的 `shop/User.php` 逻辑，收银员的登录和密码管理会自动覆盖。**无需单独处理**。

---

### 4. 旧版文件（可能废弃，低优先级）

| 文件 | 状态 | 处理建议 |
|------|------|----------|
| admin/app/shop/model_old/shop/User.php | 旧版 | 确认是否废弃，如废弃则不修改 |
| admin/app/shop/model_old/auth/User.php | 旧版 | 确认是否废弃，如废弃则不修改 |
| admin/app/shop/model_old/supplier/Supplier.php | 旧版 | 确认是否废弃，如废弃则不修改 |

---

## 🎯 任务优先级划分

### P0 - 必须完成（阻塞发布）
- ✅ Go 密码工具类
- ✅ Go 登录/修改密码/添加员工逻辑（6个任务）
- ✅ PHP 密码工具函数
- ✅ PHP 核心员工管理（6个任务）

**工期**: 4 天

### P1 - 重要但不阻塞
- ✅ PHP 供应商管理
- ✅ PHP 收银员管理
- ✅ PHP 应用管理
- ✅ PHP ERP 集成

**工期**: 1 天

### P2 - 可选（灰度后处理）
- 旧版文件（确认废弃后忽略）

---

## ⚠️ 特殊注意事项

### 1. 权限密码字段
**涉及表**: shop_*.ttpos_staff.permission_password  
**涉及文件**: admin/app/shop/model/auth/User.php  
**特殊性**: 同时处理两个密码字段（password + permission_password）

### 2. 同步到 saas 库
**涉及文件**: 
- admin/app/admin/model/CompanyStaff.php
- admin/app/shop/model/auth/User.php

**风险**: 确保商家库和 saas 库密码同步一致

### 3. 登录控制器
**涉及文件**: admin/app/shop/controller/Passport.php  
**特殊性**: 当前在控制器层预加密，需要改为 Model 层处理

### 4. Token 生成
**涉及文件**: 
- admin/app/shop/model/shop/User.php（行 67, 156, 166）
- admin/app/admin/model/admin/User.php（行 37）

**风险**: 部分代码使用 `md5($password)` 生成 token，bcrypt 后需调整

---

## 📈 工作量统计

| 类型 | 数量 | 预计工时 |
|------|------|----------|
| Go 代码文件 | 3 | 2 天 |
| PHP 代码文件（P0） | 6 | 2 天 |
| PHP 代码文件（P1） | 2 | 0.5 天 |
| 单元测试 | - | 0.5 天 |
| 集成测试 | - | 1 天 |
| 性能测试 | - | 0.5 天 |
| 安全测试 | - | 0.5 天 |
| 灰度发布 | - | 0.5 天 |
| **总计** | **11 个文件** | **8 天** |

---

## ✅ 验收标准

### 功能验收
- [ ] MD5 密码可以正常登录（Go + PHP）
- [ ] bcrypt 密码可以正常登录（Go + PHP）
- [ ] MD5 密码登录后自动升级为 bcrypt
- [ ] 修改密码时新密码使用 bcrypt
- [ ] 添加员工时密码使用 bcrypt
- [ ] 权限密码验证正常
- [ ] 应用管理功能正常
- [ ] 供应商登录和密码管理正常（通过 ttpos_staff 表）
- [ ] 收银员登录和密码管理正常（通过 ttpos_staff 表）
- [ ] ERP 集成功能正常

### 性能验收
- [ ] 登录响应时间 P95 < 300ms
- [ ] bcrypt 验证时间 < 100ms
- [ ] 密码升级不阻塞主流程

### 安全验收
- [ ] 密码不记录到日志
- [ ] bcrypt 成本因子为 10
- [ ] 错误信息不泄露敏感信息

---

## 📝 风险评估

| 风险项 | 等级 | 影响 | 缓解措施 |
|--------|------|------|----------|
| PHP 文件数量多 | 🟡 中 | 开发和测试工作量增加 | 分 P0/P1 优先级，确保核心功能先完成 |
| token 生成逻辑 | 🟡 中 | 可能影响会话管理 | 详细测试，必要时调整 token 策略 |
| saas 库同步 | 🟡 中 | 商家库和 saas 库数据不一致 | 重点测试同步逻辑 |
| 供应商/收银员继承逻辑 | 🟢 低 | 继承关系可能导致遗漏 | 在测试阶段明确测试供应商和收银员场景 |

---

## 🔄 后续优化

### 短期（1-2周）
- 监控密码迁移进度
- 收集性能数据
- 优化慢查询

### 中期（1-2月）
- P1 任务全部完成
- 90% 用户完成迁移
- 性能优化

### 长期（3-6月）
- 考虑会员密码升级
- 评估是否移除 MD5 验证逻辑
- 评估是否提高 bcrypt 成本因子

---

**文档状态**: ✅ 完整  
**最后更新**: 2025-12-16  
**审核状态**: 待审核
