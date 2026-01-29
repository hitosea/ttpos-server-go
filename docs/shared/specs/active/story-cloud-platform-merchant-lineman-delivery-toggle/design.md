# 云平台-商家管理-LINE MAN外卖控制 设计文档

> 本文档定义云平台商家管理中 LINE MAN 外卖控制功能的技术设计和实现方案。

## 📋 概述

在云平台的商家管理模块中，新增/编辑商家时增加 LINE MAN 外卖功能的开启/关闭开关配置项。商家可以通过该配置控制是否启用 LINE MAN 外卖服务，配置后系统会根据该状态控制 LINE MAN 外卖相关的业务逻辑。

**实现范围**：
1. 在商家管理模块中添加 LINE MAN 外卖开关配置（后端 API）
2. 在 Go Main 模块中添加 LINE MAN 外卖状态支持（模型、服务、权限）
3. 在 PHP Admin 模块中添加 LINE MAN 外卖状态支持（查询、授权、权限）
4. 在前端实现 LINE MAN 外卖开关显示（Vue 组件）

**参考实现**：`enable_grab_delivery` 字段的实现方式（`story-cloud-platform-merchant-grab-delivery-toggle`）

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- 遵循 Gin 框架规范
- 使用 GORM ORM
- URL 使用 snake_case
- 返回错误使用 error 类型，不使用 panic

### PHP 规范 (php.mdc)

- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数
- 使用软删除

### API 设计规范 (api.mdc)

- URL 使用 snake_case
- 响应格式统一：`{code, message, data{}}`
- data 不能为 null 或数组

### 数据库规范 (database.mdc)

- 必需字段完整（id, uuid, create_time, update_time, delete_time）
- 时间字段使用 int
- 字段名使用 snake_case

### Vue 前端规范 (vue.mdc)

- 使用 Vue 3 + TypeScript + Composition API
- 使用 Element Plus 组件库
- 遵循命名规范

---

## 🔄 代码复用分析

### 可复用的现有组件

**PHP Admin 模块**:
- **商家管理 Controller**: `admin/app/admin/controller/Shop.php` - 新建/编辑商家接口
- **验证器**: `admin/app/admin/validate/AppValidate.php` - 参数验证规则
- **App Model**: `admin/app/admin/model/app/App.php` - 商家数据模型
- **数据库迁移模板**: `admin/database/migrations/20251208191025_add_enable_grab_delivery_to_company_setting.php` - 参考迁移脚本

**Go Main 模块**:
- **CompanySetting Model**: `main/app/model/company.go` - 公司设置模型
- **BaseInfo DTO**: `main/app/dto/resp/base.go` - 基础信息 DTO
- **Auth Service**: `main/app/service/auth.go` - 认证服务
- **Product Service**: `main/app/service/product.go` - 商品服务
- **RoleAccess Service**: `main/app/service/role_access.go` - 权限服务

**Vue 前端模块**:
- **商家编辑对话框**: `admin/views/admin/src/pages/merchant/components/dialog-edit.vue` - 商家编辑页面
- **商家 API**: `admin/views/admin/src/api/merchant/index.ts` - 商家 API 定义

### 集成点

- **商家管理 API**: 在现有新建/编辑接口中添加 `enable_lineman_delivery` 参数
- **数据库表**: 在 `company_setting` 表中添加 `enable_lineman_delivery` 字段
- **Go Main 模块**: 在 Model、Service、DTO 中添加 LINE MAN 外卖状态支持
- **PHP Admin 模块**: 在 Controller、Model 中添加 LINE MAN 外卖状态支持
- **Vue 前端**: 在商家编辑页面中添加 LINE MAN 外卖开关

---

## 🏗️ 架构设计

### 分层设计原则

**Go Main 三层架构**:

```
Handler 层 (API Handler)
  ↓ 依赖
Service 层 (Business Logic)
  ↓ 依赖
Model 层 (Data Access)
```

**依赖规则**:

- ✅ Handler 依赖 Service
- ✅ Service 依赖 Model
- ❌ Handler 不直接依赖 Model
- ✅ 使用 DTO 进行数据传输

**PHP Admin 三层架构**:

```
Controller 层 (Controller)
  ↓ 依赖
Service 层 (Service)
  ↓ 依赖
Model 层 (Model)
```

**依赖规则**:

- ✅ Controller 依赖 Service
- ✅ Service 依赖 Model
- ❌ Controller 不写业务逻辑
- ✅ 使用验证器验证参数

**Vue 前端架构**:

```
Pages (页面组件)
  ↓ 依赖
Components (业务组件)
  ↓ 依赖
API (API 封装)
  ↓ 调用
Backend API
```

### 模块划分

#### Go Main 模块

- **Model 层**: `main/app/model/` - 数据模型
  - `company.go` - 公司设置模型（新增 `EnableLinemanDelivery` 字段和 `IsOpenLINEMANDelivery()` 方法）
- **DTO 层**: `main/app/dto/resp/` - 响应 DTO
  - `base.go` - 基础信息 DTO（新增 `IsOpenLINEMANDelivery` 字段）
- **Service 层**: `main/app/service/` - 业务服务
  - `auth.go` - 认证服务（返回 LINE MAN 外卖状态）
  - `product.go` - 商品服务（外卖类型过滤）
  - `role_access.go` - 权限服务（权限过滤）
  - `h5_service.go` - H5 服务（返回 LINE MAN 外卖状态）
- **Handler 层**: `main/app/api/v1/menu/` - API 处理器
  - `menu_handler.go` - 菜单处理器（返回 LINE MAN 外卖状态）

#### PHP Admin 模块

- **Controller 层**: `admin/app/admin/controller/` - 控制器
  - `Shop.php` - 商家管理（新增/编辑接口）
- **Validate 层**: `admin/app/admin/validate/` - 参数验证
  - `AppValidate.php` - 商家参数验证
- **Model 层**: `admin/app/admin/model/` - 数据模型
  - `app/App.php` - 商家模型
- **Common Model 层**: `admin/app/common/model/` - 公共模型
  - `app/App.php` - 授权信息模型
  - `shop/Access.php` - 权限过滤模型
- **Shop Controller 层**: `admin/app/shop/controller/` - 商家端控制器
  - `Controller.php` - 基础控制器（查询字段）

#### Vue 前端模块

- **Pages**: `admin/views/admin/src/pages/merchant/` - 商家管理页面
  - `components/dialog-edit.vue` - 商家编辑对话框
- **API**: `admin/views/admin/src/api/merchant/` - 商家 API
  - `index.ts` - API 定义和 TypeScript 类型

---

## 🗄️ 数据库设计

### 数据表设计

#### 表 1: company_setting（修改现有表）

**新增字段**:

```sql
ALTER TABLE `company_setting` 
ADD COLUMN `enable_lineman_delivery` INT(3) NOT NULL DEFAULT 0 COMMENT '是否启用LINE MAN外卖：0-否；1-是' 
AFTER `enable_grab_delivery`;
```

**字段说明**:
| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| enable_lineman_delivery | INT(3) | 是否启用LINE MAN外卖 | DEFAULT 0, NOT NULL |

**迁移文件**: `admin/database/migrations/20260112103704_add_enable_lineman_delivery_to_company_setting.php`

**参考实现**: `admin/database/migrations/20251208191025_add_enable_grab_delivery_to_company_setting.php`

### 数据库迁移

**迁移脚本**:

```bash
# 创建迁移文件
cd admin
php think migrate:create AddEnableLinemanDeliveryToCompanySetting

# 执行迁移
php think migrate:run
```

**参考**: `docs/agent/workflows/database-migration.md`

---

## 📊 数据模型

### Go Main Model

```go
// main/app/model/company.go
type CompanySetting struct {
    // ... 其他字段 ...
    EnableGrabDelivery        int    `gorm:"column:enable_grab_delivery;type:int(11);default:0;comment:是否启用Grab外卖: 0-否 1-是;NOT NULL" json:"enable_grab_delivery"`
    EnableLinemanDelivery     int    `gorm:"column:enable_lineman_delivery;type:int(11);default:0;comment:是否启用LINE MAN外卖: 0-否 1-是;NOT NULL" json:"enable_lineman_delivery"`
}

// 是否开启 Grab 外卖
func (model *CompanySetting) IsOpenGrabDelivery() bool {
    return model.EnableGrabDelivery == 1
}

// 是否开启 LINE MAN 外卖
func (model *CompanySetting) IsOpenLINEMANDelivery() bool {
    return model.EnableLinemanDelivery == 1
}
```

### Go Main DTO

```go
// main/app/dto/resp/base.go
type BaseInfo struct {
    // ... 其他字段 ...
    IsOpenGrabDelivery    bool   `json:"is_open_grab_delivery"`    // 是否开启Grab外卖功能
    IsOpenLINEMANDelivery bool   `json:"is_open_lineman_delivery"` // 是否开启LINE MAN外卖功能
}
```

### PHP Model

```php
// admin/app/admin/model/app/App.php
// 在 getList() 方法的 $field 数组中添加：
"su.enable_grab_delivery",
"su.enable_lineman_delivery",
```

**参考实现**: `admin/app/admin/model/app/App.php` (第 101-103 行)

---

## 🔌 API 设计

### RESTful API

#### API 1: 新建商家（修改现有接口）

**请求**:

- **URL**: `/api/admin/shop/add`
- **Method**: `POST`
- **Headers**:
  ```json
  {
    "Authorization": "Bearer {token}",
    "Content-Type": "application/json"
  }
  ```
- **Body**:
  ```json
  {
    "name": "商家名称",
    "enable_grab_delivery": 0,
    "enable_lineman_delivery": 0
  }
  ```

**响应**:

```json
{
  "code": 1,
  "message": "添加成功",
  "data": {}
}
```

**参数说明**:
- `enable_lineman_delivery`: 可选，默认 0（0-关闭，1-开启）

#### API 2: 编辑商家（修改现有接口）

**请求**:

- **URL**: `/api/admin/shop/edit`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "app_id": 1,
    "name": "商家名称",
    "enable_grab_delivery": 1,
    "enable_lineman_delivery": 1
  }
  ```

**响应**:

```json
{
  "code": 1,
  "message": "修改成功",
  "data": {
    "app_id": 1
  }
}
```

**参数说明**:
- `enable_lineman_delivery`: 可选，默认 0（0-关闭，1-开启）

#### API 3: 商家列表查询（修改现有接口）

**请求**:

- **URL**: `/api/admin/shop/index`
- **Method**: `POST`

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {
        "app_id": 1,
        "name": "商家名称",
        "enable_grab_delivery": 0,
        "enable_lineman_delivery": 0
      }
    ]
  }
}
```

**返回字段**:
- `enable_lineman_delivery`: 是否启用LINE MAN外卖（0-否，1-是）

#### API 4: 商家端基础信息（Go Main）

**请求**:

- **URL**: `/shop/base`
- **Method**: `POST`

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "base_info": {
      "is_open_grab_delivery": true,
      "is_open_lineman_delivery": false
    }
  }
}
```

**返回字段**:
- `is_open_lineman_delivery`: 是否开启LINE MAN外卖（true/false）

#### API 5: 商家端基础信息（PHP Admin）

**请求**:

- **URL**: `/shop/base`
- **Method**: `POST`

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "supplier": {
      "enable_grab_delivery": 1,
      "enable_lineman_delivery": 0
    }
  }
}
```

**返回字段**:
- `enable_lineman_delivery`: 是否启用LINE MAN外卖（0/1）

---

## 🧩 组件和接口

### Go Main 实现

#### Model 层

```go
// main/app/model/company.go
// 已实现：第 113-114 行
EnableGrabDelivery        int    `gorm:"column:enable_grab_delivery;type:int(11);default:0;comment:是否启用Grab外卖: 0-否 1-是;NOT NULL" json:"enable_grab_delivery"`
EnableLinemanDelivery     int    `gorm:"column:enable_lineman_delivery;type:int(11);default:0;comment:是否启用LINE MAN外卖: 0-否 1-是;NOT NULL" json:"enable_lineman_delivery"`

// 已实现：第 237-239 行
func (model *CompanySetting) IsOpenLINEMANDelivery() bool {
    return model.EnableLinemanDelivery == 1
}
```

#### DTO 层

```go
// main/app/dto/resp/base.go
// 已实现：第 150-151 行
IsOpenGrabDelivery    bool   `json:"is_open_grab_delivery"`    // 是否开启Grab外卖功能
IsOpenLINEMANDelivery bool   `json:"is_open_lineman_delivery"` // 是否开启LINE MAN外卖功能
```

#### Service 层

```go
// main/app/service/auth.go
// 已实现：第 699-700 行（3 处位置）
IsOpenGrabDelivery:    companySetting.IsOpenGrabDelivery(),
IsOpenLINEMANDelivery: companySetting.IsOpenLINEMANDelivery(),

// main/app/service/product.go
// 已实现：第 8533-8535 行
if companySetting.IsOpenLINEMANDelivery() && takeOutMap[constant.TakeoutTypeLINEMAN] {
    takeoutTypes = append(takeoutTypes, constant.TakeoutTypeLINEMAN)
}

// main/app/service/role_access.go
// 已实现：第 233-235 行
if slices.Contains([]uint64{2857159888896001, 2857180860416001, 2857201831936000, 2857222803456001}, permission.Uuid) && !companySetting.IsOpenLINEMANDelivery() {
    continue
}
```

#### Handler 层

```go
// main/app/api/v1/menu/menu_handler.go
// 已实现：第 101-102 行
IsOpenGrabDelivery:    companySetting.IsOpenGrabDelivery(),
IsOpenLINEMANDelivery: companySetting.IsOpenLINEMANDelivery(),

// main/app/service/h5_service.go
// 已实现：第 104-105 行
IsOpenGrabDelivery:    companySetting.IsOpenGrabDelivery(),
IsOpenLINEMANDelivery: companySetting.IsOpenLINEMANDelivery(),
```

### PHP Admin 实现

#### Controller 层

```php
// admin/app/admin/controller/Shop.php
// 已实现：第 103-104 行（add 方法）
* @Apidoc\Param("enable_grab_delivery", type="int", require=false, default=0, desc="是否启用Grab外卖: 0不开启, 1开启")
* @Apidoc\Param("enable_lineman_delivery", type="int", require=false, default=0, desc="是否启用LINE MAN外卖: 0不开启, 1开启")

// 已实现：第 185-186 行（edit 方法）
* @Apidoc\Param("enable_grab_delivery", type="int", require=false, default=0, desc="是否启用Grab外卖: 0不开启, 1开启")
* @Apidoc\Param("enable_lineman_delivery", type="int", require=false, default=0, desc="是否启用LINE MAN外卖: 0不开启, 1开启")
```

#### Validate 层

```php
// admin/app/admin/validate/AppValidate.php
// 已实现：第 59-61 行（验证规则）
'enable_grab_delivery|是否启用Grab外卖' => 'in:0,1',
'enable_lineman_delivery|是否启用LINEMAN外卖' => 'in:0,1',

// 已实现：第 121-123 行（add 场景）
'enable_grab_delivery',
'enable_lineman_delivery',

// 已实现：第 168-170 行（edit 场景）
'enable_grab_delivery',
'enable_lineman_delivery',
```

#### Model 层

```php
// admin/app/admin/model/app/App.php
// 已实现：第 101-103 行
"su.enable_grab_delivery",
"su.enable_lineman_delivery",

// admin/app/shop/controller/Controller.php
// 已实现：第 162 行
enable_grab_delivery, enable_lineman_delivery,

// admin/app/common/model/app/App.php
// 已实现：第 221-222 行
'enable_grab_delivery' => $this->supplier?->enable_grab_delivery ?? 0,
'enable_lineman_delivery' => $this->supplier?->enable_lineman_delivery ?? 0,

// admin/app/common/model/shop/Access.php
// 已实现：第 386 行（权限过滤）
if (($licenses['enable_grab_delivery'] ?? 0) == 0 && ($licenses['enable_lineman_delivery'] ?? 0) == 0) {
    if ($value['uuid'] == 1734000001) { // 外卖权限UUID
        continue;
    }
}
```

### Vue 前端实现

#### 表单组件

```vue
<!-- admin/views/admin/src/pages/merchant/components/dialog-edit.vue -->
<!-- 已实现：第 162-172 行 -->
<!-- Grab外卖 -->
<el-form-item :label="$t('Grab外卖')" prop="enable_grab_delivery">
  <el-radio-group v-model="formData.enable_grab_delivery">
    <el-radio :value="1">{{ $t('开启') }}</el-radio>
    <el-radio :value="0">{{ $t('关闭') }}</el-radio>
  </el-radio-group>
</el-form-item>
<!-- LINE MAN外卖 -->
<el-form-item :label="$t('LINE MAN外卖')" prop="enable_lineman_delivery">
  <el-radio-group v-model="formData.enable_lineman_delivery">
    <el-radio :value="1">{{ $t('开启') }}</el-radio>
    <el-radio :value="0">{{ $t('关闭') }}</el-radio>
  </el-radio-group>
</el-form-item>

<!-- 已实现：第 397-398 行（默认值） -->
enable_grab_delivery: 0,
enable_lineman_delivery: 0,

<!-- 已实现：第 439-440 行（验证规则） -->
enable_grab_delivery: [{ required: true, message: $t('请选择'), trigger: 'blur' }],
enable_lineman_delivery: [{ required: true, message: $t('请选择'), trigger: 'blur' }],

<!-- 已实现：第 676-678 行（数据绑定） -->
enable_grab_delivery: props.detail?.enable_grab_delivery || 0,
enable_lineman_delivery: props.detail?.enable_lineman_delivery || 0,
```

#### TypeScript 类型

```typescript
// admin/views/admin/src/api/merchant/index.ts
// 已实现：第 121-122 行
export interface MerchantEditParams {
  // ... 其他字段 ...
  enable_grab_delivery?: number; // 是否开启Grab外卖: 0不开启, 1开启
  enable_lineman_delivery?: number; // 是否开启LINE MAN外卖: 0不开启, 1开启
}
```

---

## ⚡ 缓存设计

### Redis 缓存

**缓存策略**:

- **Key 命名**: `ttpos:company_setting:{company_uuid}`
- **过期时间**: 根据业务场景设置（建议 5 分钟）
- **更新策略**: Cache-Aside Pattern

**示例**:

```php
// 读取商家设置时，优先从缓存读取
$cacheKey = "ttpos:company_setting:{$companyUuid}";
$setting = Cache::get($cacheKey);
if (!$setting) {
    $setting = CompanySetting::where('company_uuid', $companyUuid)->find();
    Cache::set($cacheKey, $setting, 300); // 5分钟
}
```

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 参数验证失败

- **处理方式**: 返回参数错误提示
- **用户影响**: 显示错误信息，提示用户检查输入
- **代码示例**:
  ```php
  if (!in_array($param['enable_lineman_delivery'], [0, 1])) {
      $this->error = 'enable_lineman_delivery 参数错误';
      return false;
  }
  ```

#### 场景 2: 数据库操作失败

- **处理方式**: 记录错误日志，返回操作失败提示
- **用户影响**: 显示操作失败信息
- **代码示例**:
  ```php
  if (!$model->save()) {
      logger()->error('保存商家设置失败', ['error' => $model->getError()]);
      return false;
  }
  ```

---

## 🔒 安全设计

### 身份验证

- **JWT Token**: 所有 API 需要 Token 验证
- **权限控制**: 只有管理员可以修改商家设置

### 数据安全

- **参数验证**: 使用验证器验证参数范围（0 或 1）
- **SQL 注入防护**: 使用参数化查询
- **XSS 防护**: 前端输入校验

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**:

- PHP Service: 70%+
- PHP Model: 80%+
- Go Service: 70%+
- Go Model: 80%+

**测试内容**:

- 参数验证
- 数据库操作
- 业务逻辑

### API 测试

**测试内容**:

- API 接口调用
- 参数验证
- 响应格式
- 错误处理

### 前端测试

**测试内容**:

- 表单显示
- 开关切换
- 数据保存
- 错误提示

---

## 📈 性能优化

### 优化策略

1. **数据库优化**:
   - 字段已添加在现有表中，无需新表
   - 字段位置优化（在 enable_grab_delivery 之后）

2. **缓存优化**:
   - Redis 缓存商家设置
   - 缓存命中率 > 80%

3. **前端优化**:
   - 使用 computed 属性缓存计算结果
   - 避免不必要的重新渲染

### 性能指标

- 本地响应时间: < 200ms
- 数据库查询: < 50ms
- 缓存命中率: > 80%

---

## 🌐 浏览器兼容性

### 前端兼容性（Vue）

- Chrome 90+
- Safari 14+
- Firefox 88+
- Edge 90+

---

## 📚 实现清单

### Phase 1: 数据库和模型

- [x] 创建数据库迁移文件
- [x] 执行数据库迁移
- [x] 更新 Go Main Model
- [x] 更新 PHP Model

### Phase 2: 后端 API

- [x] 更新 PHP Controller 接口
- [x] 更新 PHP 验证器
- [x] 更新 PHP Model 字段
- [x] 更新 Go Main DTO
- [x] 更新 Go Main Service
- [x] 更新 Go Main Handler

### Phase 3: 业务逻辑

- [x] 实现外卖类型过滤
- [x] 实现权限过滤
- [x] 实现授权信息返回

### Phase 4: 前端实现

- [x] 实现表单显示
- [x] 实现验证规则
- [x] 实现数据绑定
- [x] 实现 TypeScript 类型

### Phase 5: 测试

- [ ] 单元测试
- [ ] API 测试
- [ ] 前端功能测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/曾振华/2026-01/2026-01-12.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2026-01-12  
**作者**: 曾振华  
**审核者**: 曾振华

