# 自助点餐机管理 设计文档

> 本文档定义自助点餐机管理功能的技术设计和实现方案。

## 📋 概述

在云平台中增加自助点餐机的管理功能，包括：
1. 在商家管理模块中增加【自助点餐机】开关，允许商家在新建或编辑时控制是否启用自助点餐机功能（默认关闭）
2. 在客户端管理模块中增加【自助点餐机】版本管理功能，支持版本发布、更新和回滚

**实现范围**：仅实现后端 API，参考 `enable_data_management` 的实现方式。

---

## 🎯 规范对齐

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

---

## 🔄 代码复用分析

### 可复用的现有组件

- **商家管理 Controller**: `admin/app/admin/controller/Shop.php` - 新建/编辑商家接口
- **验证器**: `admin/app/admin/validate/AppValidate.php` - 参数验证规则
- **App Model**: `admin/app/admin/model/app/App.php` - 商家数据模型
- **客户端版本管理 Controller**: `admin/app/admin/controller/client/Client.php` - 版本管理接口
- **数据库迁移模板**: `admin/database/migrations/20251120013811_add_table_map_fields_to_company_setting.php` - 参考迁移脚本

### 集成点

- **商家管理 API**: 在现有新建/编辑接口中添加 `enable_kiosk` 参数
- **数据库表**: 在 `company_setting` 表中添加 `enable_kiosk` 字段
- **客户端版本管理**: 在现有版本管理接口中支持 type=6（自助点餐机）

---

## 🏗️ 架构设计

### 分层设计原则

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

### 模块划分

#### PHP Admin 模块

- **Controller 层**: `admin/app/admin/controller/` - 控制器
  - `Shop.php` - 商家管理（新增/编辑接口）
  - `client/Client.php` - 客户端版本管理
- **Validate 层**: `admin/app/admin/validate/` - 参数验证
  - `AppValidate.php` - 商家参数验证
- **Model 层**: `admin/app/admin/model/` - 数据模型
  - `app/App.php` - 商家模型

---

## 🗄️ 数据库设计

### 数据表设计

#### 表 1: company_setting（修改现有表）

**新增字段**:

```sql
ALTER TABLE `company_setting` 
ADD COLUMN `enable_kiosk` INT(3) NOT NULL DEFAULT 0 COMMENT '是否启用自助点餐机：0-否；1-是' 
AFTER `enable_data_management`;
```

**字段说明**:
| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| enable_kiosk | INT(3) | 是否启用自助点餐机 | DEFAULT 0, NOT NULL |

**迁移文件**: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_enable_kiosk_to_company_setting.php`

**参考实现**: `admin/database/migrations/20251120013811_add_table_map_fields_to_company_setting.php`

### 数据库迁移

**迁移脚本**:

```bash
# 创建迁移文件
cd admin
php think migrate:create AddEnableKioskToCompanySetting

# 执行迁移
php think migrate:run
```

**参考**: `docs/agent/workflows/database-migration.md`

---

## 📊 数据模型

### PHP Model

```php
// admin/app/admin/model/app/App.php
// 在 getList() 方法的 $field 数组中添加：
"su.enable_kiosk",
```

**参考实现**: `admin/app/admin/model/app/App.php` (第 96 行)

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
- **Body** (新增参数):
  ```json
  {
    "enable_kiosk": 0
  }
  ```

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "app_id": 123456
  }
}
```

**参考实现**: `admin/app/admin/controller/Shop.php` (add 方法)

#### API 2: 编辑商家（修改现有接口）

**请求**:

- **URL**: `/api/admin/shop/edit`
- **Method**: `GET,POST`
- **Headers**:
  ```json
  {
    "Authorization": "Bearer {token}",
    "Content-Type": "application/json"
  }
  ```
- **Body** (新增参数):
  ```json
  {
    "app_id": 123456,
    "enable_kiosk": 1
  }
  ```

**响应**:

```json
{
  "code": 1,
  "message": "修改成功",
  "data": {
    "app_id": 123456
  }
}
```

**参考实现**: `admin/app/admin/controller/Shop.php` (edit 方法)

#### API 3: 商家列表（修改现有接口）

**请求**:

- **URL**: `/api/admin/shop/index`
- **Method**: `POST`

**响应** (新增字段):

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": {
      "data": [
        {
          "app_id": 123456,
          "name": "测试商家",
          "enable_kiosk": 0
        }
      ]
    }
  }
}
```

**参考实现**: `admin/app/admin/controller/Shop.php` (index 方法)

#### API 4: 客户端版本列表（修改现有接口）

**请求**:

- **URL**: `/api/admin/client.client/index`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "type": "6"
  }
  ```

**说明**: type=6 表示自助点餐机类型

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": {
      "data": [
        {
          "id": 1,
          "type": 6,
          "version_number": "100",
          "version_name": "1.0.0",
          "is_publish": 1
        }
      ]
    }
  }
}
```

**参考实现**: `admin/app/admin/controller/client/Client.php` (index 方法)

#### API 5: 客户端版本添加（修改现有接口）

**请求**:

- **URL**: `/api/admin/client.client/add`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "type": 6,
    "brand": 0,
    "version_number": "100",
    "package_url": "/uploads/client/kiosk/app.apk"
  }
  ```

**说明**: type=6 表示自助点餐机类型

**参考实现**: `admin/app/admin/controller/client/Client.php` (add 方法)

#### API 6: 客户端版本发布（修改现有接口）

**请求**:

- **URL**: `/api/admin/client.client/publish`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "id": 1,
    "update_log": "更新日志",
    "forced_update": 0
  }
  ```

**参考实现**: `admin/app/admin/controller/client/Client.php` (publish 方法)

#### API 7: 客户端版本查询（修改现有接口）

**请求**:

- **URL**: `/api/admin/client.client/getNewVersion`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "type": 6,
    "brand": 0
  }
  ```

**说明**: type=6 表示自助点餐机类型

**参考实现**: `admin/app/admin/controller/client/Client.php` (getNewVersion 方法)

---

## 🧩 组件和接口

### Validate 层

#### AppValidate（修改现有验证器）

```php
// admin/app/admin/validate/AppValidate.php
// 在 $rule 数组中添加：
'enable_kiosk|是否启用自助点餐机' => 'in:0,1',

// 在 $scene['add'] 数组中添加：
'enable_kiosk',

// 在 $scene['edit'] 数组中添加：
'enable_kiosk',
```

**参考实现**: `admin/app/admin/validate/AppValidate.php` (第 54-55 行，第 110-111 行，第 151-152 行)

### Controller 层

#### Shop Controller（修改现有控制器）

```php
// admin/app/admin/controller/Shop.php
// add() 方法：自动接收 enable_kiosk 参数（通过验证器）
// edit() 方法：自动接收 enable_kiosk 参数（通过验证器）
// index() 方法：在 getList() 的 $field 中添加 "su.enable_kiosk"
```

**参考实现**: `admin/app/admin/controller/Shop.php`

#### Client Controller（修改现有控制器）

```php
// admin/app/admin/controller/client/Client.php
// index() 方法：支持 type=6（自助点餐机）
// add() 方法：支持 type=6（自助点餐机）
// publish() 方法：支持自助点餐机版本发布
// getNewVersion() 方法：支持 type=6（自助点餐机）
```

**参考实现**: `admin/app/admin/controller/client/Client.php`

---

## ⚡ 缓存设计

### Redis 缓存

**缓存策略**:

- 商家配置信息缓存（如需要）
- 客户端版本信息缓存（如需要）

**参考**: 现有缓存实现

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 参数验证失败

- **处理方式**: 验证器自动返回错误信息
- **用户影响**: 返回错误提示，不保存数据
- **代码示例**:
  ```php
  // 验证器自动处理
  if (!$validate->goCheck('add')) {
      return $this->renderError($validate->getError());
  }
  ```

#### 场景 2: 数据库操作失败

- **处理方式**: Model 返回错误信息
- **用户影响**: 返回错误提示
- **代码示例**:
  ```php
  if (!$model->edit($data)) {
      return $this->renderError($model->getError() ?: '修改失败');
  }
  ```

---

## 🔒 安全设计

### 身份验证

- **JWT Token**: 所有 API 需要 Token 验证
- **权限控制**: 管理员权限验证

### 数据安全

- **SQL 注入防护**: 使用参数化查询（ThinkPHP ORM）
- **XSS 防护**: 前端输入校验
- **CSRF 防护**: Token 验证

---

## 🧪 测试策略

### 单元测试

**测试内容**:

- 验证器参数验证
- Model 数据操作

### API 测试

**测试内容**:

- API 接口调用
- 参数验证
- 响应格式
- 错误处理

### 集成测试

**测试流程**:

- 端到端业务流程
- 数据库事务
- 数据一致性

---

## 📈 性能优化

### 优化策略

1. **数据库优化**:
   - 添加索引（如需要）
   - 优化 SQL 查询

2. **缓存优化**:
   - Redis 缓存热点数据（如需要）

---

## 📚 实现清单

### Phase 1: 数据库和模型

- [ ] 创建数据库迁移文件
- [ ] 执行数据库迁移
- [ ] 更新 App Model（添加字段到查询列表）

### Phase 2: 核心实现

- [ ] 更新验证器（添加 enable_kiosk 验证规则）
- [ ] 更新 Shop Controller（支持 enable_kiosk 参数）
- [ ] 更新 Client Controller（支持 type=6）

### Phase 3: 测试

- [ ] API 测试
- [ ] 集成测试
- [ ] 手动测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-05  
**作者**: 王昱  
**审核者**: {审核者}

