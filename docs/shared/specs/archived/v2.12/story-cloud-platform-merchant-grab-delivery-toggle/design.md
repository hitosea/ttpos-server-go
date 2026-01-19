# 云平台-商家管理-Grab外卖控制 设计文档

> 本文档定义云平台商家管理中 Grab 外卖控制功能的技术设计和实现方案。

## 📋 概述

在云平台的商家管理模块中，新增/编辑商家时增加 Grab 外卖功能的开启/关闭开关配置项。商家可以通过该配置控制是否启用 Grab 外卖服务，配置后系统会根据该状态控制 Grab 外卖相关的业务逻辑。

**实现范围**：
1. 在商家管理模块中添加 Grab 外卖开关配置（后端 API）
2. 在新管理端实现 Grab 外卖功能的可见性控制（前端）

**参考实现**：`enable_kiosk` 字段的实现方式（`story-shop-kiosk-management`）

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

### Vue 前端规范 (vue.mdc)

- 使用 Vue 3 + TypeScript + Composition API
- 使用 Element Plus 组件库
- 遵循命名规范

---

## 🔄 代码复用分析

### 可复用的现有组件

- **商家管理 Controller**: `admin/app/admin/controller/Shop.php` - 新建/编辑商家接口
- **验证器**: `admin/app/admin/validate/AppValidate.php` - 参数验证规则
- **App Model**: `admin/app/admin/model/app/App.php` - 商家数据模型
- **数据库迁移模板**: `admin/database/migrations/20251205185229_add_enable_kiosk_to_company_setting.php` - 参考迁移脚本
- **新管理端外卖页面**: `admin/views/admin/src/pages/takeout/` - 外卖相关页面

### 集成点

- **商家管理 API**: 在现有新建/编辑接口中添加 `enable_grab_delivery` 参数
- **数据库表**: 在 `company_setting` 表中添加 `enable_grab_delivery` 字段
- **新管理端前端**: 在现有外卖页面中根据 `enable_grab_delivery` 状态控制可见性

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

#### PHP Admin 模块

- **Controller 层**: `admin/app/admin/controller/` - 控制器
  - `Shop.php` - 商家管理（新增/编辑接口）
- **Validate 层**: `admin/app/admin/validate/` - 参数验证
  - `AppValidate.php` - 商家参数验证
- **Model 层**: `admin/app/admin/model/` - 数据模型
  - `app/App.php` - 商家模型

#### Vue 前端模块

- **Pages**: `admin/views/admin/src/pages/takeout/` - 外卖相关页面
  - `order.vue` - 外卖订单列表
  - `shop.vue` - 外卖商家管理
  - `setting.vue` - 外卖设置
- **Router**: `admin/views/admin/src/router/routes.ts` - 路由配置
- **Layout**: `admin/views/admin/src/layouts/components/sidebar.vue` - 侧边栏菜单

---

## 🗄️ 数据库设计

### 数据表设计

#### 表 1: company_setting（修改现有表）

**新增字段**:

```sql
ALTER TABLE `company_setting` 
ADD COLUMN `enable_grab_delivery` INT(3) NOT NULL DEFAULT 0 COMMENT '是否启用Grab外卖：0-否；1-是' 
AFTER `enable_kiosk`;
```

**字段说明**:
| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| enable_grab_delivery | INT(3) | 是否启用Grab外卖 | DEFAULT 0, NOT NULL |

**迁移文件**: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_enable_grab_delivery_to_company_setting.php`

**参考实现**: `admin/database/migrations/20251205185229_add_enable_kiosk_to_company_setting.php`

### 数据库迁移

**迁移脚本**:

```bash
# 创建迁移文件
cd admin
php think migrate:create AddEnableGrabDeliveryToCompanySetting

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
"su.enable_grab_delivery",
```

**参考实现**: `admin/app/admin/model/app/App.php` (第 98 行，enable_kiosk 字段)

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
    "enable_grab_delivery": 0
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
- `enable_grab_delivery`: 可选，默认 0（0-关闭，1-开启）

#### API 2: 编辑商家（修改现有接口）

**请求**:

- **URL**: `/api/admin/shop/edit`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "app_id": 1,
    "name": "商家名称",
    "enable_grab_delivery": 1
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
- `enable_grab_delivery`: 可选，默认 0（0-关闭，1-开启）

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
        "enable_grab_delivery": 0
      }
    ]
  }
}
```

**返回字段**:
- `enable_grab_delivery`: 是否启用Grab外卖（0-否，1-是）

---

## 🧩 组件和接口

### PHP Controller

```php
// admin/app/admin/controller/Shop.php
// 在 add() 方法的 @Apidoc\Param 中添加：
* @Apidoc\Param("enable_grab_delivery", type="int", require=false, default=0, desc="是否启用Grab外卖: 0不开启, 1开启")

// 在 edit() 方法的 @Apidoc\Param 中添加：
* @Apidoc\Param("enable_grab_delivery", type="int", require=false, default=0, desc="是否启用Grab外卖: 0不开启, 1开启")
```

**参考实现**: `admin/app/admin/controller/Shop.php` (第 102 行，enable_kiosk 参数)

### PHP Validate

```php
// admin/app/admin/validate/AppValidate.php
// 在 $scene['add'] 数组中添加：
'enable_grab_delivery',

// 在 $scene['edit'] 数组中添加：
'enable_grab_delivery',

// 在 $rule 数组中添加验证规则（如需要）：
'enable_grab_delivery' => 'in:0,1',
```

**参考实现**: `admin/app/admin/validate/AppValidate.php` (第 115 行，enable_kiosk 验证)

### PHP Model

```php
// admin/app/admin/model/app/App.php
// 在 getList() 方法的 $field 数组中添加：
"su.enable_grab_delivery",
```

**参考实现**: `admin/app/admin/model/app/App.php` (第 98 行，enable_kiosk 字段)

### Vue 前端可见性控制

#### 路由控制

```typescript
// admin/views/admin/src/router/routes.ts
// 根据商家 enable_grab_delivery 状态控制 Grab 外卖相关路由的显示
// 在路由守卫或菜单配置中添加判断逻辑
```

#### 侧边栏菜单控制

```vue
<!-- admin/views/admin/src/layouts/components/sidebar.vue -->
<!-- 根据商家 enable_grab_delivery 状态控制 Grab 外卖菜单项的显示 -->
<el-menu-item v-if="companySetting?.enable_grab_delivery === 1" ...>
```

#### 页面内容过滤

```vue
<!-- admin/views/admin/src/pages/takeout/order.vue -->
<!-- 根据商家 enable_grab_delivery 状态过滤 Grab 渠道订单 -->
<script setup lang="ts">
const filteredOrders = computed(() => {
  if (companySetting.value?.enable_grab_delivery !== 1) {
    return orders.value.filter(order => order.channel !== 'grab')
  }
  return orders.value
})
</script>
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
  if (!in_array($param['enable_grab_delivery'], [0, 1])) {
      $this->error = 'enable_grab_delivery 参数错误';
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

- 可见性控制逻辑
- 路由权限控制
- 菜单显示控制
- 页面内容过滤

---

## 📈 性能优化

### 优化策略

1. **数据库优化**:
   - 添加索引（如需要）
   - 优化 SQL 查询

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

- [ ] 创建数据库迁移文件
- [ ] 执行数据库迁移
- [ ] 更新 PHP Model

### Phase 2: 后端 API

- [ ] 更新 Controller 接口
- [ ] 更新验证器
- [ ] 更新 Model 字段

### Phase 3: 前端可见性控制

- [ ] 实现路由控制
- [ ] 实现侧边栏菜单控制
- [ ] 实现页面内容过滤

### Phase 4: 测试

- [ ] 单元测试
- [ ] API 测试
- [ ] 前端功能测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-08  
**作者**: 开发团队  
**审核者**: {审核者}

