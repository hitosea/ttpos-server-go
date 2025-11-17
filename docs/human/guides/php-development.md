# PHP Admin 模块开发指南

> 👤 **受众**: 人类开发者  
> 📖 **用途**: PHP Admin 模块（ThinkPHP）的详细开发规范和最佳实践

---

## 命名规范

### 1. 类命名 - 大驼峰（PascalCase）

```php
class UserController extends Controller {}
class OrderService {}
class ProductModel extends BaseModel {}
```

### 2. 方法命名 - 小驼峰（camelCase）

```php
public function getUserList() {}
public function createUser() {}
public function updateUserInfo() {}
```

### 3. 变量命名 - 小驼峰（camelCase）

```php
$userId = 1;
$userName = 'test';
$orderList = [];
```

### 4. 常量命名 - 全大写+下划线

```php
const ORDER_STATUS_PENDING = 0;
const MAX_PAGE_SIZE = 100;
```

---

## MVC 架构

### 控制器层

```php
<?php
namespace app\admin\controller;

use app\admin\service\UserService;
use app\admin\validate\UserValidate;
use think\Request;

class UserController extends Controller
{
    public function getUserList(Request $request)
    {
        // 1. 获取参数
        $page = $request->param('page', 1);
        $pageSize = $request->param('page_size', 20);
        
        // 2. 调用服务层
        $userService = new UserService();
        $result = $userService->getUserList($page, $pageSize);
        
        // 3. 返回响应
        return $this->renderSuccess($result);
    }
}
```

### 服务层

```php
<?php
namespace app\admin\service;

use app\admin\model\User;
use think\facade\Db;

class UserService
{
    public function getUserList($page = 1, $pageSize = 20)
    {
        $list = User::where('delete_time', 0)
            ->page($page, $pageSize)
            ->select()
            ->toArray();
        
        $total = User::where('delete_time', 0)->count();
        
        return [
            'list' => $list,
            'meta' => [
                'page_no' => $page,
                'page_size' => $pageSize,
                'total' => $total,
            ],
        ];
    }
    
    public function createUser($data)
    {
        Db::startTrans();
        try {
            $user = User::create([
                'uuid' => $this->generateUuid(),
                'username' => $data['username'],
                'email' => $data['email'],
                'create_time' => time(),
            ]);
            
            Db::commit();
            return ['user_id' => $user->id];
        } catch (\Exception $e) {
            Db::rollback();
            throw $e;
        }
    }
}
```

### 模型层

```php
<?php
namespace app\admin\model;

use app\common\model\BaseModel;

class User extends BaseModel
{
    protected $name = 'user';
    protected $pk = 'id';
    
    protected $type = [
        'id' => 'integer',
        'status' => 'integer',
        'create_time' => 'integer',
    ];
    
    // 软删除
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    
    // 获取器
    public function getCreateTimeAttr($value)
    {
        return $value > 0 ? date('Y-m-d H:i:s', $value) : '';
    }
    
    // 关联
    public function roles()
    {
        return $this->hasMany(UserRole::class, 'user_id', 'id');
    }
}
```

---

## 数据库操作

### 查询

```php
// 单条查询
$user = User::where('id', 1)->find();

// 多条查询
$users = User::where('status', 1)->select();

// 分页
$users = User::where('status', 1)->page($page, $pageSize)->select();

// 关联查询
$user = User::with(['roles'])->find(1);

// 聚合
$count = User::count();
$sum = User::sum('amount');
```

### 插入

```php
// 单条插入
$user = User::create([
    'username' => 'test',
    'email' => 'test@example.com',
    'create_time' => time(),
]);

// 批量插入
User::insertAll([
    ['username' => 'user1'],
    ['username' => 'user2'],
]);
```

### 更新

```php
// 更新
User::where('id', 1)->update([
    'username' => 'new_name',
    'update_time' => time(),
]);

// 自增自减
User::where('id', 1)->inc('login_count')->update();
```

### 删除（软删除）

```php
// 软删除
User::where('id', 1)->update(['delete_time' => time()]);

// 使用模型软删除
$user = User::find(1);
$user->delete();
```

### 事务

```php
use think\facade\Db;

// 手动控制
Db::startTrans();
try {
    User::create($userData);
    Order::create($orderData);
    Db::commit();
} catch (\Exception $e) {
    Db::rollback();
    throw $e;
}

// 闭包自动控制
Db::transaction(function () {
    User::create($userData);
    Order::create($orderData);
});
```

---

## 验证器

### 定义验证器

```php
<?php
namespace app\admin\validate;

use think\Validate;

class UserValidate extends Validate
{
    protected $rule = [
        'username' => 'require|length:2,20|unique:user',
        'email' => 'require|email|unique:user',
        'password' => 'require|length:6,20',
    ];
    
    protected $message = [
        'username.require' => '用户名不能为空',
        'username.length' => '用户名长度为2-20个字符',
        'email.email' => '邮箱格式不正确',
    ];
    
    protected $scene = [
        'create' => ['username', 'email', 'password'],
        'update' => ['username', 'email'],
    ];
}
```

### 使用验证器

```php
public function createUser(Request $request)
{
    $data = $request->post();
    
    $validate = new UserValidate();
    if (!$validate->scene('create')->check($data)) {
        return $this->renderError($validate->getError());
    }
    
    // 业务逻辑
}
```

---

## API 文档（ApiDoc）

### ApiDoc 注解

```php
/**
 * @Apidoc\Title("用户管理")
 * @Apidoc\Group("admin")
 */
class UserController extends Controller
{
    /**
     * @Apidoc\Title("获取用户列表")
     * @Apidoc\Method("GET")
     * @Apidoc\Param("page", type="int", require=false, default="1")
     * @Apidoc\Returned("code", type="int", desc="状态码")
     * @Apidoc\Returned("data", type="object", desc="返回数据")
     */
    public function getUserList(Request $request) {}
}
```

---

## 错误处理

### 统一响应

```php
<?php
namespace app\common\controller;

use think\Controller;

class BaseController extends Controller
{
    protected function renderSuccess($data = [], $message = 'success')
    {
        return json([
            'code' => 1,
            'message' => $message,
            'data' => $data ?: (object)[],
        ]);
    }
    
    protected function renderError($message = 'error', $code = 0)
    {
        return json([
            'code' => $code,
            'message' => $message,
            'data' => (object)[],
        ]);
    }
}
```

---

## 中间件

### 定义中间件

```php
<?php
namespace app\admin\middleware;

use Closure;

class AuthMiddleware
{
    public function handle($request, Closure $next)
    {
        $token = $request->header('Authorization');
        
        if (empty($token)) {
            return json([
                'code' => 401,
                'message' => '未登录',
                'data' => (object)[],
            ]);
        }
        
        return $next($request);
    }
}
```

---

## 性能优化

### 1. 使用缓存

```php
// 使用缓存
$cacheKey = 'user_' . $userId;
$user = cache($cacheKey);
if (!$user) {
    $user = User::find($userId);
    cache($cacheKey, $user, 3600);
}
```

### 2. 避免 N+1 查询

```php
// ✅ 预加载
$users = User::with(['roles'])->select();

// ❌ N+1 查询
$users = User::select();
foreach ($users as $user) {
    $roles = $user->roles;
}
```

---

## 代码风格

### PSR-2 规范

- 使用 4 个空格缩进
- 左花括号另起一行
- 每行不超过 120 字符
- 方法之间空一行

```php
<?php
namespace app\admin\controller;

use think\Request;

class UserController extends Controller
{
    public function index(Request $request)
    {
        $page = $request->param('page', 1);
        
        if ($page < 1) {
            return $this->renderError('页码错误');
        }
        
        return $this->renderSuccess();
    }
}
```

---

## 相关文档

- [PHP 架构设计](../architecture/php-architecture.md) - 深入理解架构
- [数据库开发指南](./database-guide.md) - 数据库设计和操作
- [API 设计指南](./api-design-guide.md) - API 设计规范

---

**最后更新**: 2025-11-17  
**维护者**: TTPOS Team  
**版本**: v1.0

