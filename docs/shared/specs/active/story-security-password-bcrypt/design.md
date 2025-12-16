# 员工密码加密升级 - 技术设计文档

**文档编号**: DESIGN-2025-12-001  
**创建日期**: 2025-12-16  
**创建人**: 曾振华  
**文档状态**: 📝 草稿  
**关联需求**: `requirements.md`

---

## 🎯 设计目标

1. 实现 MD5 到 bcrypt 的平滑迁移
2. 最小化代码侵入，保持现有架构
3. 确保向后兼容，不影响现有用户
4. 性能损耗控制在可接受范围

---

## 🏗️ 总体架构

### 核心设计原则

1. **双验证兼容**: 同时支持 MD5 和 bcrypt 两种格式
2. **自动识别**: 通过密码前缀自动判断加密方式
3. **异步升级**: 登录成功后异步升级，不阻塞主流程
4. **最小改动**: 只修改密码相关逻辑，不改变业务流程

### 识别机制

```go
// 根据密码前缀判断类型
func getPasswordType(storedPassword string) PasswordType {
    if strings.HasPrefix(storedPassword, "$2a$") ||
       strings.HasPrefix(storedPassword, "$2b$") ||
       strings.HasPrefix(storedPassword, "$2y$") {
        return PasswordTypeBcrypt
    }
    return PasswordTypeMD5
}
```

---

## 📦 模块设计

### 1. 密码工具模块（Go）

**文件**: `main/pkg/utils/password.go`

#### 核心函数

```go
// HashPasswordBcrypt 使用 bcrypt 加密密码
func HashPasswordBcrypt(password string) (string, error)

// VerifyPasswordBcrypt 验证 bcrypt 密码
func VerifyPasswordBcrypt(password, hash string) bool

// VerifyPassword 通用密码验证（自动识别格式）
// 返回：(验证是否成功, 是否需要升级)
func VerifyPassword(password, storedPassword string) (bool, bool)

// UpgradePasswordAsync 异步升级密码为 bcrypt
func UpgradePasswordAsync(db *gorm.DB, table, idField string, id uint64, password string)
```

#### 实现细节

```go
package utils

import (
	"strings"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	BcryptCost = 10 // bcrypt 成本因子
)

// HashPasswordBcrypt 使用 bcrypt 加密密码
func HashPasswordBcrypt(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPasswordBcrypt 验证 bcrypt 密码
func VerifyPasswordBcrypt(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// VerifyPassword 通用密码验证
// 返回：(验证是否成功, 是否需要升级为bcrypt)
func VerifyPassword(password, storedPassword string) (bool, bool) {
	// 判断是否为 bcrypt 格式
	if strings.HasPrefix(storedPassword, "$2a$") ||
	   strings.HasPrefix(storedPassword, "$2b$") ||
	   strings.HasPrefix(storedPassword, "$2y$") {
		// 使用 bcrypt 验证
		isValid := VerifyPasswordBcrypt(password, storedPassword)
		return isValid, false // 已经是 bcrypt，不需要升级
	}
	
	// 使用旧的 MD5 验证
	md5Password := EncryptPassword(password)
	if md5Password == storedPassword {
		return true, true // 验证成功，需要升级为 bcrypt
	}
	
	return false, false // 验证失败
}

// UpgradePasswordAsync 异步升级密码（适用于登录场景）
func UpgradePasswordAsync(db *gorm.DB, table, fieldName string, idField string, id uint64, plainPassword string) {
	go func() {
		newHash, err := HashPasswordBcrypt(plainPassword)
		if err != nil {
			// 记录日志但不影响主流程
			return
		}
		
		// 更新数据库
		db.Table(table).
			Where(idField+" = ?", id).
			Update(fieldName, newHash)
	}()
}
```

### 2. 密码工具模块（PHP）

**文件**: `admin/app/common.php`

#### 修改 salt_hash() 函数

```php
/**
 * 验证密码（支持 MD5 和 bcrypt）
 * @param string $password 明文密码
 * @param string $storedPassword 存储的密码哈希
 * @return array [是否验证成功, 是否需要升级]
 */
function verify_password($password, $storedPassword) {
    // 判断是否为 bcrypt 格式
    if (str_starts_with($storedPassword, '$2a$') || 
        str_starts_with($storedPassword, '$2b$') || 
        str_starts_with($storedPassword, '$2y$')) {
        // 使用 bcrypt 验证
        $isValid = password_verify($password, $storedPassword);
        return [$isValid, false]; // 已经是 bcrypt，不需要升级
    }
    
    // 使用旧的 MD5 验证
    $md5Password = salt_hash($password);
    if ($md5Password === $storedPassword) {
        return [true, true]; // 验证成功，需要升级
    }
    
    return [false, false]; // 验证失败
}

/**
 * 使用 bcrypt 加密密码
 * @param string $password 明文密码
 * @return string bcrypt 哈希
 */
function hash_password_bcrypt($password) {
    return password_hash($password, PASSWORD_BCRYPT, ['cost' => 10]);
}

/**
 * 异步升级密码为 bcrypt
 * @param int $id 用户ID
 * @param string $table 表名
 * @param string $idField ID字段名
 * @param string $password 明文密码
 */
function upgrade_password_async($id, $table, $idField, $password) {
    // PHP 中异步比较复杂，可以简化为同步或使用队列
    try {
        $newHash = hash_password_bcrypt($password);
        \think\facade\Db::table($table)
            ->where($idField, $id)
            ->update(['password' => $newHash]);
    } catch (\Exception $e) {
        // 记录日志但不影响主流程
        log_write('密码升级失败: ' . $e->getMessage());
    }
}

/**
 * 保留原 salt_hash 函数用于兼容
 */
function salt_hash($password) {
    return md5(md5($password) . 'jjjshop_salt_2020');
}
```

---

## 🔄 核心流程设计

### 流程 1: 登录验证（Go）

**文件**: `main/app/service/auth.go`

#### 修改前（第 153 行）

```go
if staff.Uuid == 0 || utils.EncryptPassword(loginReq.Password) != staff.Password {
    return loginResp, errors.New("账号或密码错误")
}
```

#### 修改后

```go
// 验证密码（支持 MD5 和 bcrypt）
isValid, needUpgrade := utils.VerifyPassword(loginReq.Password, staff.Password)
if !isValid {
    return loginResp, errors.New("账号或密码错误")
}

// 如果需要升级密码，异步升级为 bcrypt
if needUpgrade {
    utils.UpgradePasswordAsync(
        s.dbm.GetDB(staff.CompanyUuid),
        "ttpos_staff",
        "password",
        "uuid",
        staff.Uuid,
        loginReq.Password,
    )
}
```

### 流程 2: 修改密码（Go）

**文件**: `main/app/service/auth.go`

#### 修改前（第 1384-1390 行）

```go
if staff.Password != utils.EncryptPassword(changePasswordReq.OldPassword) {
    return errors.New("旧密码错误")
}

update := map[string]any{
    "password": utils.EncryptPassword(changePasswordReq.NewPassword),
    // ...
}
```

#### 修改后

```go
// 验证旧密码（支持 MD5 和 bcrypt）
isValid, _ := utils.VerifyPassword(changePasswordReq.OldPassword, staff.Password)
if !isValid {
    return errors.New("旧密码错误")
}

// 使用 bcrypt 加密新密码
newPasswordHash, err := utils.HashPasswordBcrypt(changePasswordReq.NewPassword)
if err != nil {
    return errors.New("密码加密失败")
}

update := map[string]any{
    "password": newPasswordHash,
    "password_change_count": staff.PasswordChangeCount + 1,
    "password_change_time": time.Now().Unix(),
}
```

### 流程 3: 添加员工（Go）

**文件**: `main/app/service/staff.go`

#### 修改前（第 493 行）

```go
Password: utils.EncryptPassword(addReq.Password),
```

#### 修改后

```go
// 使用 bcrypt 加密密码
passwordHash, err := utils.HashPasswordBcrypt(addReq.Password)
if err != nil {
    return errors.New("密码加密失败"), nil
}

Password: passwordHash,
```

### 流程 4: 登录验证（PHP）

**文件**: `admin/app/shop/model/shop/User.php`

#### 修改前（第 30 行）

```php
->where('password', $password)
```

#### 修改后

```php
$user = $this->with(['app', 'supplier'])
    ->where(function ($q) use ($username) {
        $q->whereRaw('BINARY username = :username', ['username' => $username]);
        $q->whereOr('phone', $username);
    })
    ->find();

if (!$user) {
    $this->error = '账号或密码错误';
    return false;
}

// 验证密码（支持 MD5 和 bcrypt）
list($isValid, $needUpgrade) = verify_password($password, $user['password']);
if (!$isValid) {
    $this->error = '账号或密码错误';
    return false;
}

// 如果需要升级，异步升级为 bcrypt
if ($needUpgrade) {
    upgrade_password_async($user['uuid'], 'ttpos_staff', 'uuid', $password);
}
```

---

## 🗄️ 数据库设计

### 不需要 Schema 变更

采用原地升级方案，无需新增字段或修改表结构。

### 密码格式示例

| 格式 | 示例 | 长度 |
|------|------|------|
| MD5 | `5f4dcc3b5aa765d61d8327deb882cf99` | 32 |
| bcrypt | `$2y$10$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx` | 60 |

### 数据迁移进度查询

```sql
-- 查询迁移进度（Go 数据库）
SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN password LIKE '$2%' THEN 1 ELSE 0 END) as bcrypt_count,
    ROUND(SUM(CASE WHEN password LIKE '$2%' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as bcrypt_percent
FROM ttpos_staff;

-- 查询迁移进度（PHP 超管）
SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN password LIKE '$2%' THEN 1 ELSE 0 END) as bcrypt_count,
    ROUND(SUM(CASE WHEN password LIKE '$2%' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as bcrypt_percent
FROM ttpos_admin_user;
```

---

## 🔐 安全考虑

### 1. 成本因子选择

bcrypt 成本因子设置为 **10**：
- 验证时间约 60-100ms
- 平衡安全性和用户体验
- 未来可根据硬件性能提升

### 2. 密码日志

严格禁止在日志中记录明文密码：
```go
// ❌ 错误示例
log.Info("用户登录", zap.String("password", password))

// ✅ 正确示例
log.Info("用户登录", zap.Uint64("staff_uuid", staffUuid))
```

### 3. 错误信息

统一返回"账号或密码错误"，不泄露具体错误类型。

### 4. 升级失败处理

密码升级失败不影响登录成功，下次登录时会重新尝试升级。

---

## ⚡ 性能优化

### 1. 异步升级

登录成功后使用 goroutine 异步升级密码，不影响登录响应时间：
```go
go func() {
    // 升级逻辑
}()
```

### 2. 批量升级（可选）

对于长期未登录的用户，可以考虑编写脚本批量升级：
```go
// 批量升级脚本（仅在必要时使用）
// 注意：此脚本需要明文密码，实际场景中不可用
// 仅作为备选方案说明
```

### 3. 性能监控

监控指标：
- 登录响应时间（P95, P99）
- bcrypt 验证时间
- 密码升级成功率

---

## 🧪 测试策略

### 单元测试

```go
// TestVerifyPassword 测试密码验证
func TestVerifyPassword(t *testing.T) {
    // 测试 MD5 密码验证
    md5Password := utils.EncryptPassword("123456")
    isValid, needUpgrade := utils.VerifyPassword("123456", md5Password)
    assert.True(t, isValid)
    assert.True(t, needUpgrade)
    
    // 测试 bcrypt 密码验证
    bcryptPassword, _ := utils.HashPasswordBcrypt("123456")
    isValid, needUpgrade = utils.VerifyPassword("123456", bcryptPassword)
    assert.True(t, isValid)
    assert.False(t, needUpgrade)
    
    // 测试错误密码
    isValid, _ = utils.VerifyPassword("wrong", md5Password)
    assert.False(t, isValid)
}
```

### 集成测试

测试场景覆盖：
- MD5 密码登录并自动升级
- bcrypt 密码登录
- 修改密码（旧密码 MD5/bcrypt）
- 添加新员工
- 权限密码验证

### 性能测试

使用 Apache Bench 或 wrk 进行压力测试：
```bash
# 测试登录接口
ab -n 1000 -c 10 -p login.json -T application/json http://api.example.com/login
```

---

## 📊 监控和告警

### 关键指标

1. **迁移进度**：bcrypt 密码占比
2. **登录成功率**：确保不低于升级前
3. **登录响应时间**：P95 不超过 300ms
4. **错误率**：密码验证失败率

### 告警规则

- 登录成功率下降 > 5%
- 登录响应时间 P95 > 500ms
- 密码升级失败率 > 10%

---

## 🚀 发布计划

### 第一阶段：代码开发（3天）
- Day 1: 实现 Go 密码工具类和核心逻辑
- Day 2: 实现 PHP 密码工具类和核心逻辑
- Day 3: 完善所有修改点，代码自测

### 第二阶段：测试（2天）
- Day 4: 单元测试和集成测试
- Day 5: 性能测试和安全测试

### 第三阶段：灰度发布（1天）
- Day 6: 10% 门店灰度，监控指标

### 第四阶段：全量发布（1天）
- Day 7: 100% 门店发布

---

## 📝 待办事项

- [ ] Go 密码工具类实现
- [ ] PHP 密码工具类实现
- [ ] Go 登录验证逻辑修改
- [ ] PHP 登录验证逻辑修改
- [ ] Go 修改密码逻辑修改
- [ ] PHP 修改密码逻辑修改
- [ ] Go 添加员工逻辑修改
- [ ] PHP 添加员工逻辑修改
- [ ] 权限密码验证逻辑修改
- [ ] 单元测试编写
- [ ] 集成测试编写
- [ ] 性能测试
- [ ] 文档更新

---

## 🔗 相关文档

- 需求文档: `requirements.md`
- 任务分解: `tasks.md`
- 提案文档: `../../team/proposals/2025-12/password-bcrypt-upgrade.md`

---

**文档状态**: ✅ 待评审  
**下一步**: 分解具体任务并开始开发
