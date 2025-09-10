# TTPOS 项目结构规范

## 目录组织结构

```
ttpos-server-go/
├── main/                           # 主要业务服务 (Go)
│   ├── app/                        # 应用层
│   │   ├── api/                    # API控制器层
│   │   ├── service/                # 业务逻辑层
│   │   ├── repository/             # 数据访问层
│   │   ├── model/                  # 数据模型
│   │   ├── dto/                    # 数据传输对象
│   │   │   ├── req/                # 请求参数
│   │   │   └── resp/               # 响应数据
│   │   ├── constant/               # 常量定义
│   │   ├── errors/                 # 错误定义
│   │   ├── event/                  # 事件处理器
│   │   ├── tasks/                  # 定时任务
│   │   └── printer/                # 打印服务
│   ├── pkg/                        # 基础设施包
│   │   ├── database/               # 数据库管理
│   │   ├── cache/                  # 缓存管理
│   │   ├── eventbus/               # 事件总线
│   │   ├── lock/                   # 并发锁
│   │   ├── logger/                 # 日志管理
│   │   ├── utils/                  # 工具类
│   │   ├── validator/              # 数据验证
│   │   ├── encrypt/                # 加密工具
│   │   ├── auth/                   # 认证授权
│   │   └── websocket/              # WebSocket工具
│   ├── middleware/                 # 中间件
│   ├── router/                     # 路由配置
│   ├── config/                     # 配置管理
│   ├── i18n/                       # 国际化
│   ├── public/                     # 静态资源
│   ├── docs/                       # API文档
│   ├── log/                        # 日志文件
│   └── scripts/                    # 构建脚本
├── websocket/                      # WebSocket服务 (Go)
├── ttpos-bmp/                      # 业务中台服务 (Go)
├── redis-proxy/                    # Redis代理服务 (Go)
├── admin/                          # 管理后台 (PHP)
│   ├── app/                        # 应用层
│   │   ├── admin/                  # 管理模块
│   │   ├── cashier/                # 收银模块
│   │   ├── shop/                   # 门店模块
│   │   └── common/                 # 公共模块
│   ├── config/                     # 配置文件
│   ├── database/                   # 数据库迁移
│   ├── public/                     # 静态资源
│   ├── views/                      # 前端视图
│   │   ├── admin/                  # 管理后台前端
│   │   └── shop/                   # 门店前端
│   └── extend/                     # 扩展包
├── docker/                         # Docker配置
│   ├── mysql/                      # MySQL容器配置
│   ├── redis/                      # Redis容器配置
│   ├── nginx/                      # Nginx配置
│   └── php/                        # PHP容器配置
├── scripts/                        # 项目脚本
└── makefile                        # 构建脚本
```

## 命名规范

### 文件命名

#### Go语言文件
- **包名**：使用`snake_case`格式，如`user_service`
- **文件名**：使用`snake_case`格式，如`user_service.go`
- **测试文件**：使用`[filename]_test.go`格式
- **接口文件**：使用`I[Name].go`格式，如`IUserService.go`

#### PHP文件
- **类名**：使用`PascalCase`格式，如`UserService`
- **文件名**：使用`snake_case`格式，如`user_service.php`
- **控制器**：使用`Controller`后缀，如`UserController.php`

#### 前端文件
- **组件名**：使用`PascalCase`格式，如`UserList.vue`
- **工具文件**：使用`camelCase`格式，如`dateUtils.js`
- **样式文件**：使用`kebab-case`格式，如`user-list.scss`

### 代码命名

#### Go语言
- **结构体**：`PascalCase`，ID字段使用大写，如`StaffId`
- **接口**：以`I`开头，如`IUserService`
- **方法**：`camelCase`，如`GetUserById`
- **常量**：`UPPER_SNAKE_CASE`，如`ORDER_STATUS_ACTIVE`
- **变量**：`camelCase`，如`userName`

#### PHP
- **类名**：`PascalCase`，如`UserService`
- **方法名**：`camelCase`，如`getUserById`
- **常量**：`UPPER_SNAKE_CASE`
- **变量**：`camelCase`

#### JavaScript/TypeScript
- **类名**：`PascalCase`
- **方法名**：`camelCase`
- **常量**：`UPPER_SNAKE_CASE`
- **变量**：`camelCase`

## 导入模式

### Go语言导入顺序
```go
import (
    // 标准库
    "context"
    "fmt"
    "time"

    // 第三方包
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    // 项目包（按字母顺序）
    "ttpos-server-go/app/model"
    "ttpos-server-go/app/service"
    "ttpos-server-go/pkg/database"
)
```

### PHP导入规范
```php
<?php
// 使用语句按字母顺序排列
use App\Models\User;
use App\Services\UserService;
use Illuminate\Http\Request;
```

### 前端导入规范
```javascript
// 第三方库
import Vue from 'vue'
import { ref, computed } from 'vue'

// 项目内部模块
import UserService from '@/services/userService'
import { formatDate } from '@/utils/dateUtils'
```

## 代码结构模式

### Go文件组织
```go
package user_service

import (
    // 导入语句
)

// 常量定义
const (
    UserStatusActive   = 1
    UserStatusInactive = 2
)

// 类型定义
type User struct {
    ID       uint64 `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Status   int    `json:"status"`
    CreateAt int64  `json:"create_at"`
}

// 接口定义
type IUserService interface {
    GetUserById(id uint64) (*User, error)
    CreateUser(req CreateUserReq) (*User, error)
}

// 结构体实现
type userService struct {
    db *gorm.DB
}

// 构造函数
func NewUserService(db *gorm.DB) IUserService {
    return &userService{db: db}
}

// 方法实现
func (s *userService) GetUserById(id uint64) (*User, error) {
    // 方法实现
}

// 私有辅助方法
func (s *userService) validateUser(user *User) error {
    // 验证逻辑
}
```

### PHP文件组织
```php
<?php

namespace App\Services;

use App\Models\User;
use App\Repositories\UserRepository;

class UserService
{
    private UserRepository $userRepository;

    public function __construct(UserRepository $userRepository)
    {
        $this->userRepository = $userRepository;
    }

    public function getUserById(int $id): ?User
    {
        // 方法实现
    }

    public function createUser(array $data): User
    {
        // 方法实现
    }

    private function validateUserData(array $data): bool
    {
        // 验证逻辑
    }
}
```

## 代码组织原则

### 1. 单一职责原则
- 每个文件只负责一个明确的功能
- 每个方法只做一件事情
- 类和接口职责清晰明确

### 2. 依赖倒置原则
- 高层模块不依赖低层模块，都依赖抽象
- 抽象不依赖细节，细节依赖抽象
- 通过接口实现依赖注入

### 3. 开闭原则
- 对扩展开放，对修改关闭
- 通过继承和组合实现功能扩展
- 避免直接修改现有代码

### 4. 接口隔离原则
- 客户端不应该依赖它不需要的接口
- 建立单一接口，不要建立庞大臃肿的接口
- 接口设计应该职责单一

### 5. 里氏替换原则
- 子类可以替换父类使用
- 保持继承关系的正确性
- 避免继承关系的滥用

## 模块边界

### 服务层边界
- **Controller层**：只负责HTTP请求处理和响应格式化
- **Service层**：负责业务逻辑处理，不直接操作数据库
- **Repository层**：只负责数据访问，不包含业务逻辑
- **Model层**：只定义数据结构，不包含业务方法

### 技术栈边界
- **Go服务**：处理核心业务逻辑和API服务
- **PHP后台**：处理管理界面和复杂报表
- **前端应用**：只负责用户界面和交互逻辑
- **基础设施**：数据库、缓存、消息队列等独立管理

### 业务模块边界
- **收银模块**：订单处理、支付结算
- **桌台模块**：桌台管理、预订服务
- **会员模块**：会员管理、积分系统
- **外送模块**：配送管理、骑手调度

## 代码大小指南

### 文件大小限制
- **Go文件**：最大500行，建议200-300行
- **PHP文件**：最大800行，建议300-500行
- **前端组件**：最大300行，建议100-200行
- **配置文件**：最大100行，建议50行以内

### 方法大小限制
- **普通方法**：最大50行，建议20-30行
- **复杂业务方法**：最大100行，建议50-80行
- **工具方法**：最大20行，建议10行以内

### 类复杂度限制
- **结构体字段**：最大20个，建议10个以内
- **方法数量**：最大15个，建议8-10个
- **依赖注入**：最大8个依赖，建议3-5个

### 目录深度限制
- **最大嵌套深度**：4级目录
- **包引用深度**：3级以内
- **文件引用链**：5个文件以内

## 文档标准

### 注释规范
- **包注释**：每个包必须有包级注释
- **函数注释**：复杂函数必须有详细注释
- **结构体注释**：重要结构体必须说明用途
- **常量注释**：非常量必须说明含义

### API文档
- **Swagger注解**：所有API接口必须有Swagger文档
- **参数说明**：请求和响应参数必须详细说明
- **错误码说明**：所有错误码必须有对应说明

### 代码文档
- **README文件**：重要目录必须有README说明
- **架构文档**：系统架构必须有详细文档
- **部署文档**：部署过程必须有详细说明

## 质量保证

### 代码检查
- **Go代码**：使用go vet、gofmt进行静态检查
- **PHP代码**：使用PHPStan进行静态分析
- **前端代码**：使用ESLint、Prettier进行检查

### 测试覆盖
- **单元测试**：核心业务逻辑测试覆盖率80%以上
- **集成测试**：API接口测试覆盖率90%以上
- **端到端测试**：关键业务流程测试覆盖率70%以上

### 持续集成
- **代码审查**：所有代码必须经过审查
- **自动化测试**：提交代码必须通过所有测试
- **代码规范**：必须通过代码规范检查