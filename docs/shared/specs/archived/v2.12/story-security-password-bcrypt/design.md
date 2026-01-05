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

## 📦 涉及模块

### 后端 Go 模块

| 模块 | 文件 | 功能 |
|------|------|------|
| 密码工具 | `main/pkg/utils/password.go` | 密码加密和验证（新建） |
| 登录服务 | `main/app/service/auth.go` | 登录验证逻辑 |
| 员工管理 | `main/app/service/staff.go` | 添加/更新员工 |
| 订单管理 | `main/app/service/order_manage.go` | 权限密码验证 |

### 后端 PHP 模块

| 模块 | 文件 | 功能 | 优先级 |
|------|------|------|--------|
| 公共函数 | `admin/app/common.php` | salt_hash() 函数 | P0 |
| 店铺员工 | `admin/app/shop/model/shop/User.php` | 登录和密码管理 | P0 |
| 超管用户 | `admin/app/admin/model/admin/User.php` | 超管登录和密码管理 | P0 |
| 员工认证 | `admin/app/shop/model/auth/User.php` | 员工添加/更新（含权限密码，包括供应商、收银员） | P0 |
| 统一账号员工 | `admin/app/admin/model/admin/Staff.php` | 统一账号员工管理 | P0 |
| 商家员工 | `admin/app/admin/model/CompanyStaff.php` | 商家员工创建 | P0 |
| 应用管理 | `admin/app/admin/model/app/App.php` | 应用管理员账号 | P1 |
| ERP 集成 | `admin/app/admin/controller/Erpnext.php` | ERPNext 密码验证 | P1 |
| 登录控制器 | `admin/app/shop/controller/Passport.php` | 登录密码处理 | P0 |

**说明**：供应商（supplier）和收银员（cashier）的账号都存储在 `ttpos_staff` 表中，通过统一的 `shop/User.php` 和 `auth/User.php` 模型处理登录和密码管理。虽然 `supplier/Supplier.php` 和 `cashier/User.php` 文件中使用了 `salt_hash()`，但它们操作的是同一张 `ttpos_staff` 表，修改核心的密码管理逻辑后这些功能会自动覆盖。

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
 * 注意：PHP 中的异步比较复杂，这里简化为同步执行
 * 生产环境可以考虑使用消息队列
 * @param int $id 用户ID
 * @param string $table 表名
 * @param string $idField ID字段名
 * @param string $fieldName 密码字段名
 * @param string $password 明文密码
 * @param int $appId 应用ID，0表示saas库，>0表示商家库
 */
function upgrade_password_async($id, $table, $idField, $fieldName, $password, $appId = 0) {
    try {
        $newHash = hash_password_bcrypt($password);
        
        // 根据 appId 确定数据库连接
        if ($appId === 0) {
            // saas 库（超管用户、统一账号员工）
            $db = \think\facade\Db::connect('saas');
        } else {
            // 商家库（门店员工）
            $db = \think\facade\Db::connect('shop' . $appId);
        }
        
        $db->table($table)
            ->where($idField, $id)
            ->update([$fieldName => $newHash]);
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

**⚠️ 多数据库兼容性说明**：

TTPOS 系统采用多数据库架构：
- **saas 库**：存储超管用户（`ttpos_admin_user`）和统一账号员工（`ttpos_staff`）
- **shop_* 库**：每个商家有独立数据库，存储门店员工（`ttpos_staff`）

因此 `upgrade_password_async` 函数必须支持多数据库：
- `$appId = 0`：操作 saas 库
- `$appId > 0`：操作商家库（`shop{$appId}`）

**调用示例**：
```php
// 超管登录升级（saas 库）
upgrade_password_async($user['admin_user_id'], 'ttpos_admin_user', 'admin_user_id', 'password', $password, 0);

// 门店员工登录升级（商家库）
upgrade_password_async($user['uuid'], 'ttpos_staff', 'uuid', 'password', $password, $user['company_uuid']);
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

### 流程 4: 登录验证（PHP - 门店员工）

**文件**: `admin/app/shop/model/shop/User.php`（P0）

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
    upgrade_password_async($user['uuid'], 'ttpos_staff', 'uuid', 'password', $password);
}

// Token 生成保持原有逻辑
$user['token'] = signToken($user['uuid'], 'shop', '', md5($user->password), $user['company_uuid']);
```

**注意**：Token 生成统一使用 `md5($user->password)`，无论密码是 MD5 还是 bcrypt 格式。

### 流程 5: 超管登录验证（PHP）

**文件**: `admin/app/admin/model/admin/User.php`（P0）

#### 修改说明

类似于门店员工登录，需要在超管用户模型中应用同样的双验证逻辑：

```php
// 查找超管用户
$user = $this->where('username', $username)->find();

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
// 注意：超管表使用 admin_user_id 作为主键
if ($needUpgrade) {
    upgrade_password_async($user['admin_user_id'], 'ttpos_admin_user', 'admin_user_id', 'password', $password);
}

// Token 生成保持原有逻辑
$user['token'] = signToken($user['admin_user_id'], 'admin', '', md5($user->password));
```

**注意**：
- 超管表（`ttpos_admin_user`）使用 `admin_user_id` 作为主键，而非 `uuid`
- Token 生成统一使用 `md5($user->password)`

### 流程 6: 员工创建/更新（PHP - 含权限密码）

**文件**: `admin/app/shop/model/auth/User.php`（P0）

#### 修改说明

在添加或更新员工（包括供应商、收银员）时，需要使用 bcrypt 加密密码和权限密码：

```php
// 添加员工
public function add($data) {
    // 使用 bcrypt 加密密码（不设置 password_change_time）
    $arr = [
        'uuid' => createUuid(),
        'phone' => trim($data['phone']),
        'username' => trim($data['user_name']),
        'password' => hash_password_bcrypt($data['password']),
        'permission_password' => hash_password_bcrypt($data['permission_password']),
        'real_name' => trim($data['real_name']),
        'user_type' => $user['user_type'],
        'company_uuid' => $appId
    ];
    
    return $this->save($arr);
}

// 更新员工密码
public function updatePassword($data) {
    $arr = [
        'phone' => $data['phone'],
        'username' => $data['user_name'],
        'real_name' => $data['real_name'],
    ];
    
    // 密码处理：如果提供了新密码，使用 bcrypt 加密
    if (!empty($data['password'])) {
        $arr['password'] = hash_password_bcrypt($data['password']);
        $arr['password_change_time'] = time();
    }
    
    // 权限密码处理：如果传了且不为空，使用 bcrypt 加密
    if (!empty($data['permission_password'])) {
        $arr['permission_password'] = hash_password_bcrypt($data['permission_password']);
    }
    
    return $this->update($arr);
}
```

**注意**：
- 创建员工时不设置 `password_change_time`
- 更新密码时，`auth/User.php` 中保留 `password_change_time = time()` 设置
- 其他模块（如 `shop/User.php` 的 `editPass`）不设置此字段

### 流程 7: 权限密码验证（Go）

**文件**: `main/app/service/order_manage.go`

#### 修改说明

在取消订单等敏感操作中验证权限密码时，使用双验证逻辑：

```go
// 验证权限密码
if staff.PermissionPassword != "" {
    isValid, needUpgrade := utils.VerifyPassword(permissionPassword, staff.PermissionPassword)
    if !isValid {
        return errors.New("权限密码错误")
    }
    
    // 如果需要升级权限密码，异步升级
    if needUpgrade {
        utils.UpgradePasswordAsync(
            s.dbm.GetDB(staff.CompanyUuid),
            "ttpos_staff",
            "permission_password",
            "uuid",
            staff.Uuid,
            permissionPassword,
        )
    }
}
```

### 流程 8: 统一账号员工管理（PHP）

**文件**: `admin/app/admin/model/admin/Staff.php`（P0）

#### 修改说明

统一账号员工（saas 数据库）的密码管理逻辑与门店员工类似：

```php
// 添加统一账号员工
public function add($data) {
    $arr = [
        'email' => trim($data['email'] ?? ''),
        'phone' => trim($data['phone'] ?? ''),
        'real_name' => trim($data['real_name'] ?? ''),
        'password' => hash_password_bcrypt($data['password']),
        'password_change_count' => 0,
        'password_change_time' => 0,  // 初始化为 0
        'is_disable' => 0,
        'last_company_uuid' => $data['company_uuid'] ?? 0,
    ];
    return $this->save($arr);
}

// 更新员工密码
public function updatePassword($data) {
    $arr = [
        'email' => $data['email'],
        'phone' => $data['phone'],
        'real_name' => $data['real_name'],
    ];
    
    // 如果提供了密码，使用 bcrypt 加密
    if (!empty($data['password'])) {
        $arr['password'] = hash_password_bcrypt($data['password']);
        $arr['password_change_count'] = Db::raw('password_change_count + 1');
        $arr['password_change_time'] = time();
    }
    
    return $this->update($arr);
}
```

**注意**：
- 创建员工时 `password_change_time` 设置为 `0`，表示从未修改过密码
- 更新密码时设置为当前时间，用于密码修改次数统计
- `password_change_count` 字段会递增，记录密码修改次数

### 流程 9: 商家员工创建（PHP）

**文件**: `admin/app/admin/model/CompanyStaff.php`（P0）

#### 修改说明

商家员工创建时使用 bcrypt 加密（不设置 password_change_time）：

```php
public function createStaff($data) {
    $uuid = createUuid();
    $password = hash_password_bcrypt($data['password']);
    
    // 同时保存到ttpos_staff表中
    // ...
    
    return $this->save($data);
}
```

### 流程 10: 应用管理员账号（PHP - P1）

**文件**: `admin/app/admin/model/app/App.php`（P1）

#### 修改说明

应用管理员账号密码处理（优先级 P1）：

```php
// 更新应用信息（含密码）
public function edit($data) {
    $user_data = [
        'username' => $data['user_name'],
        'phone' => $data['link_phone'] ?? '',
    ];
    
    if (!empty($data['password'])) {
        // 使用 bcrypt 加密密码（不设置 password_change_time）
        $user_data['password'] = hash_password_bcrypt($data['password']);
    }
    
    // 同步更新到 saas 和 shop 数据库
    // ...
}
```

**注意**：应用管理员密码更新时不设置 `password_change_time`，保持字段原有值。

### 流程 11: ERPNext 密码验证（PHP - P1）

**文件**: `admin/app/admin/controller/Erpnext.php`（P1）

#### 修改说明

ERPNext 集成相关的密码验证（优先级 P1）：

```php
// 验证 ERPNext 账号密码
public function authCompany() {
    // 验证用户名密码是否正确
    $user = User::withTrashed()
        ->whereRaw('BINARY username = :username', ['username' => $this->admin['user']['user_name']])
        ->order('admin_user_id', 'desc')
        ->order('delete_time')
        ->find();
    
    if (!$user) {
        return $this->renderError('密码错误');
    }
    
    // 验证密码（支持 MD5 和 bcrypt）
    list($isValid, $needUpgrade) = verify_password($param['password'] ?? '', $user->password);
    if (!$isValid) {
        return $this->renderError('密码错误');
    }
    
    // 如果需要升级，异步升级为 bcrypt
    // 注意：超管表使用 admin_user_id 作为主键
    if ($needUpgrade) {
        upgrade_password_async($user['admin_user_id'], 'ttpos_admin_user', 'admin_user_id', 'password', $param['password'] ?? '');
    }
    
    // ...
}
```

**注意**：ERPNext 验证使用超管账号，需要使用 `admin_user_id` 字段。

---

## 🗄️ 数据库设计

### 不需要 Schema 变更

采用原地升级方案，无需新增字段或修改表结构。

### 涉及的表和字段

| 数据库 | 表名 | 字段 | 当前格式 | 目标格式 | 优先级 | 说明 |
|--------|------|------|----------|----------|--------|------|
| saas | ttpos_admin_user | password | MD5 | bcrypt | P0 | 超管密码 |
| saas | ttpos_staff | password | MD5 | bcrypt | P0 | 统一账号员工密码 |
| shop_* | ttpos_staff | password | MD5 | bcrypt | P0 | 门店员工、供应商、收银员密码 |
| shop_* | ttpos_staff | permission_password | MD5 | bcrypt | P0 | 权限密码 |

**说明**：
- 供应商（supplier）和收银员（cashier）的账号都存储在 `ttpos_staff` 表中
- 通过统一的 `shop/User.php` 和 `auth/User.php` 模型处理登录和密码管理
- 修改核心的密码管理逻辑后，这些功能会自动覆盖

### 密码格式示例

| 格式 | 示例 | 长度 |
|------|------|------|
| MD5 | `5f4dcc3b5aa765d61d8327deb882cf99` | 32 |
| bcrypt | `$2y$10$xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx` | 60 |

### 数据迁移进度查询

```sql
-- 查询迁移进度（门店员工 - shop_* 数据库）
SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN password LIKE '$2%' THEN 1 ELSE 0 END) as bcrypt_count,
    ROUND(SUM(CASE WHEN password LIKE '$2%' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as bcrypt_percent
FROM ttpos_staff;

-- 查询权限密码迁移进度（shop_* 数据库）
SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN permission_password LIKE '$2%' THEN 1 ELSE 0 END) as bcrypt_count,
    SUM(CASE WHEN permission_password IS NOT NULL AND permission_password != '' THEN 1 ELSE 0 END) as has_permission_password,
    ROUND(SUM(CASE WHEN permission_password LIKE '$2%' THEN 1 ELSE 0 END) * 100.0 / 
          SUM(CASE WHEN permission_password IS NOT NULL AND permission_password != '' THEN 1 ELSE 0 END), 2) as bcrypt_percent
FROM ttpos_staff
WHERE permission_password IS NOT NULL AND permission_password != '';

-- 查询迁移进度（超管 - saas 数据库）
SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN password LIKE '$2%' THEN 1 ELSE 0 END) as bcrypt_count,
    ROUND(SUM(CASE WHEN password LIKE '$2%' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as bcrypt_percent
FROM ttpos_admin_user;

-- 查询迁移进度（统一账号员工 - saas 数据库）
SELECT 
    COUNT(*) as total,
    SUM(CASE WHEN password LIKE '$2%' THEN 1 ELSE 0 END) as bcrypt_count,
    ROUND(SUM(CASE WHEN password LIKE '$2%' THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as bcrypt_percent
FROM ttpos_staff;
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

### 5. Token 生成策略

**当前实现**：统一使用 `md5($user->password)` 生成 token 中的密码标识。

**实现特点**：
- ✅ 保持了原有 token 生成逻辑的完全兼容性
- ✅ 无论密码是 MD5 还是 bcrypt 格式，token 生成方式一致
- ✅ 不破坏现有 token 验证机制
- ⚠️ bcrypt 密码的 md5 值会随着密码内容变化，与 MD5 密码行为一致

**工作原理**：
- MD5 密码：`md5(md5($password))`，存储后再 md5 用于 token
- bcrypt 密码：`$2y$10$...`（60位），存储后再 md5 用于 token
- 两种格式的 md5 结果都是 32 位，格式统一

**建议优化方向**（后续版本）：
1. 使用不可变字段（如 `uuid`）作为 token 标识
2. 引入专门的 `token_salt` 字段
3. 基于 JWT 的 `exp` 时间戳作为验证依据

**当前方案的权衡**：
- 优先保证功能的平滑过渡
- 零代码侵入现有 token 验证机制
- 后续可独立优化 token 生成策略

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

#### Go 模块测试场景
- ✅ MD5 密码登录并自动升级
- ✅ bcrypt 密码登录
- ✅ 修改密码（旧密码 MD5/bcrypt）
- ✅ 添加新员工
- ✅ 权限密码验证（MD5/bcrypt）
- ✅ 权限密码自动升级

#### PHP 模块测试场景

**门店员工（shop/User.php）**
- ✅ 门店员工 MD5 密码登录并升级
- ✅ 门店员工 bcrypt 密码登录
- ✅ 供应商账号登录验证
- ✅ 收银员账号登录验证

**超管（admin/User.php）**
- ✅ 超管 MD5 密码登录并升级
- ✅ 超管 bcrypt 密码登录

**员工创建/更新（auth/User.php）**
- ✅ 创建新员工（密码 bcrypt）
- ✅ 更新员工密码（bcrypt）
- ✅ 设置权限密码（bcrypt）
- ✅ 更新权限密码（bcrypt）

**统一账号员工（admin/Staff.php）**
- ✅ 统一账号员工登录验证
- ✅ 创建统一账号员工
- ✅ 密码自动升级

**商家员工（CompanyStaff.php）**
- ✅ 创建商家员工（密码 bcrypt）

**登录控制器（Passport.php）**
- ✅ 登录流程完整测试
- ✅ 密码错误处理
- ✅ 账号不存在处理

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

### 第一阶段：代码开发 - P0 模块（4天）
- **Day 1**: 实现 Go 密码工具类和核心逻辑
  - `main/pkg/utils/password.go` 密码工具类
  - `main/app/service/auth.go` 登录和修改密码逻辑
  - `main/app/service/staff.go` 添加员工逻辑
  - `main/app/service/order_manage.go` 权限密码验证

- **Day 2**: 实现 PHP 密码工具类
  - `admin/app/common.php` 公共密码函数

- **Day 3**: 实现 PHP 核心模块
  - `admin/app/shop/model/shop/User.php` 门店员工登录
  - `admin/app/admin/model/admin/User.php` 超管登录
  - `admin/app/shop/model/auth/User.php` 员工创建/更新

- **Day 4**: 完善其他 P0 PHP 模块
  - `admin/app/admin/model/admin/Staff.php` 统一账号员工
  - `admin/app/admin/model/CompanyStaff.php` 商家员工
  - `admin/app/shop/controller/Passport.php` 登录控制器

### 第二阶段：测试（2天）
- **Day 5**: 单元测试和集成测试
  - Go 模块单元测试
  - PHP 模块单元测试
  - 登录流程集成测试
  - 密码修改流程测试
  - 权限密码验证测试

- **Day 6**: 性能测试和安全测试
  - 登录响应时间测试
  - bcrypt 验证性能测试
  - 并发压力测试
  - 安全审计（密码日志、错误信息）

### 第三阶段：灰度发布（1天）
- **Day 7**: 10% 门店灰度
  - 监控登录成功率
  - 监控响应时间
  - 监控密码升级成功率
  - 收集反馈和问题

### 第四阶段：全量发布（1天）
- **Day 8**: 100% 门店发布
  - 全量开启
  - 持续监控关键指标
  - 准备回滚方案

### 第五阶段：P1 模块开发（后期）
- 应用管理员账号支持
- ERPNext 密码验证支持

---

## 📝 待办事项

### 核心工具类（P0）
- [x] Go 密码工具类实现 (`main/pkg/utils/password.go`)
- [x] PHP 密码工具类实现 (`admin/app/common.php`)

### Go 模块修改（P0）
- [x] Go 登录验证逻辑修改 (`main/app/service/auth.go`)
- [x] Go 修改密码逻辑修改 (`main/app/service/auth.go`)
- [x] Go 添加员工逻辑修改 (`main/app/service/staff.go`)
- [x] Go 更新员工逻辑修改 (`main/app/service/staff.go`)
- [x] Go 权限密码验证逻辑修改 (`main/app/service/order_manage.go`)

### PHP 模块修改（P0）
- [x] PHP 门店员工登录验证 (`admin/app/shop/model/shop/User.php`)
- [x] PHP 超管登录验证 (`admin/app/admin/model/admin/User.php`)
- [x] PHP 员工创建/更新逻辑 (`admin/app/shop/model/auth/User.php`)
- [x] PHP 统一账号员工管理 (`admin/app/admin/model/admin/Staff.php`)
- [x] PHP 商家员工创建 (`admin/app/admin/model/CompanyStaff.php`)
- [x] PHP 登录控制器处理 (`admin/app/shop/controller/Passport.php`)

### PHP 模块修改（P1 - 已完成）
- [x] PHP 应用管理员账号 (`admin/app/admin/model/app/App.php`)
- [x] PHP ERPNext 密码验证 (`admin/app/admin/controller/Erpnext.php`)

### 测试（P0）
- [ ] 单元测试编写（Go + PHP）
- [ ] 集成测试编写（登录、修改密码、添加员工、权限密码）
- [ ] 性能测试（登录响应时间、bcrypt 验证时间）
- [ ] 安全测试（密码日志检查、错误信息检查）

### 文档和发布（P0）
- [ ] 文档更新（API 文档、操作手册）
- [ ] 监控指标配置
- [ ] 灰度发布计划
- [ ] 回滚方案准备

### 实现细节说明

**password_change_time 字段处理**：
- 创建员工时：设置为 `0`（表示从未修改）
- 更新密码时：**不再更新此字段**（保持原有值）
- 说明：该字段仅用于记录密码修改次数相关的时间戳，不影响密码验证

**Token 生成策略**：
- 统一使用 `md5($user->password)` 生成 token 标识
- 无论密码是 MD5 还是 bcrypt 格式，都使用相同逻辑
- 保持与原有逻辑的完全兼容性
- 后续可考虑使用其他不可变字段优化

**主键字段差异**：
- 超管表（`ttpos_admin_user`）使用 `admin_user_id`
- 员工表（`ttpos_staff`）使用 `uuid`
- 密码升级时需要使用正确的字段名

**多数据库支持**：
- saas 库（`appId = 0`）：超管用户、统一账号员工
- shop_* 库（`appId > 0`）：门店员工
- `upgrade_password_async` 函数根据 `appId` 自动选择正确的数据库连接

---

## 🔗 相关文档

- 需求文档: `requirements.md`
- 任务分解: `tasks.md`
- 提案文档: `../../team/proposals/2025-12/password-bcrypt-upgrade.md`

---

**文档状态**: ✅ 待评审  
**下一步**: 分解具体任务并开始开发
