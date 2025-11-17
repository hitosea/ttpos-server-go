# PHP Admin 模块架构设计

> 👤 **受众**: 人类开发者  
> 📖 **用途**: 深入理解 PHP Admin 模块（基于 ThinkPHP）的架构设计

---

## 架构概览

PHP Admin 模块基于 **ThinkPHP 6.0** 框架开发，采用经典的 **MVC 架构**，提供管理后台和店铺后台功能。

### 核心特点

1. **ThinkPHP 6.0**: 成熟稳定的 PHP 框架
2. **MVC 架构**: 清晰的三层分离
3. **多模块**: admin（管理后台）+ shop（店铺后台）+ common（共享代码）
4. **前后端分离**: 后端提供 API，前端使用 Vue3
5. **ApiDoc 文档**: 自动生成 API 文档

---

## 项目结构

```
admin/
├── app/
│   ├── admin/              # 管理后台模块
│   │   ├── controller/     # 控制器
│   │   ├── model/          # 模型
│   │   ├── service/        # 服务层
│   │   ├── validate/       # 验证器
│   │   └── middleware/     # 中间件
│   ├── shop/               # 店铺后台模块
│   │   ├── controller/     # 控制器
│   │   ├── model/          # 模型
│   │   ├── service/        # 服务层
│   │   └── validate/       # 验证器
│   └── common/             # 共享代码
│       ├── controller/     # 公共控制器
│       ├── model/          # 公共模型
│       ├── service/        # 公共服务
│       ├── library/        # 类库
│       └── middleware/     # 公共中间件
├── config/                 # 配置文件
├── database/               # 数据库
│   ├── migrations/         # 迁移文件
│   └── seeds/              # 种子文件
├── views/                  # 前端代码（Vue3）
│   ├── admin/              # 管理后台前端
│   └── shop/               # 店铺后台前端
└── public/                 # 公共资源
```

---

## MVC 架构

### 架构图

```
┌─────────────────────────────────────────┐
│         HTTP 请求                       │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│          中间件层                        │
│   认证、权限、日志、跨域等              │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│        控制器层 (Controller)            │
│   参数接收、验证、调用服务层            │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│        服务层 (Service)                 │
│   业务逻辑处理、事务管理                │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│        模型层 (Model)                   │
│   数据库操作、ORM、关联查询             │
└─────────────────────────────────────────┘
```

### 1. 控制器层（Controller）

**职责**: 接收请求、参数验证、调用服务层、返回响应

```php
<?php
namespace app\admin\controller;

use app\admin\service\UserService;
use app\admin\validate\UserValidate;
use think\Request;
use hg\apidoc\annotation as Apidoc;

/**
 * @Apidoc\Title("用户管理")
 * @Apidoc\Group("admin")
 */
class UserController extends Controller
{
    /**
     * @Apidoc\Title("获取用户列表")
     * @Apidoc\Method("GET")
     */
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
    
    /**
     * @Apidoc\Title("创建用户")
     * @Apidoc\Method("POST")
     */
    public function createUser(Request $request)
    {
        // 参数验证
        $data = $request->post();
        $validate = new UserValidate();
        if (!$validate->scene('create')->check($data)) {
            return $this->renderError($validate->getError());
        }
        
        // 调用服务层
        $userService = new UserService();
        $result = $userService->createUser($data);
        
        return $this->renderSuccess($result, '创建成功');
    }
}
```

### 2. 服务层（Service）

**职责**: 业务逻辑处理、事务管理、调用模型层

```php
<?php
namespace app\admin\service;

use app\admin\model\User;
use think\facade\Db;

class UserService
{
    /**
     * 获取用户列表
     */
    public function getUserList($page = 1, $pageSize = 20)
    {
        // 查询数据
        $list = User::where('delete_time', 0)
            ->page($page, $pageSize)
            ->order('create_time', 'desc')
            ->select()
            ->toArray();
        
        // 获取总数
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
    
    /**
     * 创建用户
     */
    public function createUser($data)
    {
        // 开启事务
        Db::startTrans();
        try {
            // 创建用户
            $user = User::create([
                'uuid' => $this->generateUuid(),
                'username' => $data['username'],
                'email' => $data['email'],
                'password' => password_hash($data['password'], PASSWORD_DEFAULT),
                'create_time' => time(),
                'update_time' => time(),
            ]);
            
            // 其他业务逻辑...
            
            // 提交事务
            Db::commit();
            
            return [
                'user_id' => $user->id,
                'uuid' => $user->uuid,
            ];
        } catch (\Exception $e) {
            // 回滚事务
            Db::rollback();
            throw $e;
        }
    }
}
```

### 3. 模型层（Model）

**职责**: 数据库操作、字段映射、关联查询

```php
<?php
namespace app\admin\model;

use app\common\model\BaseModel;

class User extends BaseModel
{
    // 表名
    protected $name = 'user';
    
    // 主键
    protected $pk = 'id';
    
    // 字段类型转换
    protected $type = [
        'id' => 'integer',
        'uuid' => 'integer',
        'status' => 'integer',
        'create_time' => 'integer',
        'update_time' => 'integer',
        'delete_time' => 'integer',
    ];
    
    // 软删除
    protected $deleteTime = 'delete_time';
    protected $defaultSoftDelete = 0;
    
    /**
     * 获取器 - 格式化创建时间
     */
    public function getCreateTimeAttr($value)
    {
        return $value > 0 ? date('Y-m-d H:i:s', $value) : '';
    }
    
    /**
     * 关联 - 用户角色
     */
    public function roles()
    {
        return $this->hasMany(UserRole::class, 'user_id', 'id');
    }
}
```

---

## 数据库操作

### 1. 查询操作

```php
// 查询单条
$user = User::where('id', 1)->find();

// 查询多条
$users = User::where('status', 1)
    ->where('delete_time', 0)
    ->select();

// 分页查询
$users = User::where('status', 1)
    ->page($page, $pageSize)
    ->select();

// 关联查询
$user = User::with(['roles'])->find(1);

// 聚合查询
$count = User::where('status', 1)->count();
$sum = User::where('status', 1)->sum('amount');
```

### 2. 事务操作

```php
use think\facade\Db;

// 方式1：手动控制
Db::startTrans();
try {
    User::create($userData);
    Order::create($orderData);
    Db::commit();
} catch (\Exception $e) {
    Db::rollback();
    throw $e;
}

// 方式2：闭包自动控制
Db::transaction(function () use ($userData, $orderData) {
    User::create($userData);
    Order::create($orderData);
});
```

---

## 验证器

### 验证器定义

```php
<?php
namespace app\admin\validate;

use think\Validate;

class UserValidate extends Validate
{
    // 验证规则
    protected $rule = [
        'username' => 'require|length:2,20|unique:user',
        'email' => 'require|email|unique:user',
        'password' => 'require|length:6,20',
        'phone' => 'mobile',
        'age' => 'number|between:1,150',
    ];
    
    // 错误消息
    protected $message = [
        'username.require' => '用户名不能为空',
        'username.length' => '用户名长度为2-20个字符',
        'username.unique' => '用户名已存在',
        'email.require' => '邮箱不能为空',
        'email.email' => '邮箱格式不正确',
    ];
    
    // 验证场景
    protected $scene = [
        'create' => ['username', 'email', 'password'],
        'update' => ['username', 'email'],
    ];
}
```

---

## 中间件

### 中间件定义

```php
<?php
namespace app\admin\middleware;

use Closure;
use think\Request;
use think\Response;

class AuthMiddleware
{
    public function handle(Request $request, Closure $next)
    {
        // 获取token
        $token = $request->header('Authorization');
        
        // 验证token
        if (empty($token)) {
            return json([
                'code' => 401,
                'message' => '未登录',
                'data' => (object)[],
            ]);
        }
        
        // 继续执行
        return $next($request);
    }
}
```

---

## API 文档（ApiDoc）

### ApiDoc 注解

```php
/**
 * @Apidoc\Title("获取用户列表")
 * @Apidoc\Desc("获取所有用户的列表信息")
 * @Apidoc\Url("/api/v1/admin/user/list")
 * @Apidoc\Method("GET")
 * @Apidoc\Param("page", type="int", require=false, default="1", desc="页码")
 * @Apidoc\Param("page_size", type="int", require=false, default="20", desc="每页数量")
 * @Apidoc\Returned("code", type="int", desc="状态码")
 * @Apidoc\Returned("message", type="string", desc="提示信息")
 * @Apidoc\Returned("data", type="object", desc="返回数据")
 */
public function getUserList(Request $request)
{
    // 实现
}
```

---

## 相关文档

- [PHP 开发指南](../guides/php-development.md) - 详细的开发规范
- [数据库设计](./database-design.md) - 数据库架构设计
- [API 设计指南](../guides/api-design-guide.md) - API 设计规范

---

**最后更新**: 2025-11-17  
**维护者**: TTPOS Team  
**版本**: v1.0

