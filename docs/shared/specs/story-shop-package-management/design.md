# 旧管理端-商品管理-套餐 设计文档

> 本文档定义旧管理端商品管理中套餐功能的 PHP 后端 API 技术设计和实现方案。

## 📋 概述

在旧管理端的商品管理模块中，增强套餐商品管理功能的 PHP 后端 API。**注意：数据库字段已添加，本次仅实现 PHP 后端 API 功能。** 主要涉及 PHP Model 更新和业务逻辑调整，支持更灵活的分组类型、可选数量、必选/默认选项配置。

---

## 🎯 规范对齐

### PHP 规范 (php.mdc)

- 遵循 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数
- 使用软删除

### API 设计规范 (api.mdc)

- URL 使用 snake_case
- 响应格式统一
- 错误信息清晰明确

### 数据库规范 (database.mdc)

- 必需字段完整
- 时间字段使用 int
- 金额字段使用 decimal

---

## 🔄 代码复用分析

### 可复用的现有组件

- **ProductPackageGroup Model**: `admin/app/common/model/product/ProductPackageGroup.php` - 套餐组模型
- **ProductPackageGroupItem Model**: `admin/app/common/model/product/ProductPackageGroupItem.php` - 套餐组商品模型
- **ProductPackageGroup::updatePackageGroup**: 更新套餐分组方法（需要扩展）
- **ProductService**: `admin/app/shop/service/ProductService.php` - 商品服务
- **Product Controller**: `admin/app/shop/controller/product/store/Product.php` - 商品控制器

### 集成点

- **现有 API**: `/api/shop/product/*` - 商品管理相关接口
- **数据库表**: `ttpos_product_package_group`, `ttpos_product_package_group_item`
- **Model 方法**: `ProductPackageGroup::updatePackageGroup` - 需要更新以支持新字段

---

## 🏗️ 架构设计

### 分层设计原则

**PHP MVC 三层架构**:

```
Controller 层
  ↓ 依赖
Service 层
  ↓ 依赖
Model 层
```

**依赖规则**:

- ✅ 上层可依赖下层
- ❌ 禁止下层依赖上层
- ❌ Controller 不写业务逻辑
- ✅ Service 处理业务逻辑
- ✅ Model 处理数据访问

### 模块划分

#### PHP Admin 模块

- **Controller 层**: `admin/app/shop/controller/product/store/Product.php` - 商品控制器
- **Service 层**: `admin/app/shop/service/ProductService.php` - 商品业务逻辑
- **Model 层**: 
  - `admin/app/common/model/product/ProductPackageGroup.php` - 套餐组模型
  - `admin/app/common/model/product/ProductPackageGroupItem.php` - 套餐组商品模型
- **Validate 层**: `admin/app/shop/validate/ProductValidate.php` - 参数验证（如需要）

---

## 🗄️ 数据库设计

### 数据表说明

**注意：数据库字段已添加，本次无需数据库迁移。**

#### 表 1: ttpos_product_package_group_item

**已存在的字段**:

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| id | bigint unsigned | 主键 ID | AUTO_INCREMENT |
| uuid | bigint unsigned | 唯一标识 | DEFAULT 0, UNIQUE |
| product_package_group_uuid | bigint unsigned | 套餐组UUID | DEFAULT 0, INDEX |
| related_uuid | bigint unsigned | 关联商品UUID | DEFAULT 0, INDEX |
| product_bom_uuid | bigint unsigned | 商品BOM UUID | DEFAULT 0, INDEX |
| num | decimal(22,4) | 数量 | DEFAULT 1.0000 |
| sort | int | 排序 | DEFAULT 0, INDEX |
| add_price | decimal(22,4) | 加价金额 | DEFAULT 0.0000 |
| **is_required** | **tinyint(1)** | **是否必选** | **DEFAULT 0** |
| **is_default** | **tinyint(1)** | **是否默认** | **DEFAULT 0** |
| create_time | int | 创建时间 | DEFAULT 0 |
| update_time | int | 更新时间 | DEFAULT 0 |
| delete_time | int | 删除时间 | DEFAULT 0 |

#### 表 2: ttpos_product_package_group

**已存在的字段**:

- `group_type`: 分组类型（0-固定 1-可选），默认 0
- `optional_count`: 可选数量，默认 0

---

## 📊 数据模型

### PHP Model

```php
// admin/app/common/model/product/ProductPackageGroup.php
// 更新 updatePackageGroup 方法，支持新字段

public static function updatePackageGroup($data, $product)
{
    // ... 现有逻辑 ...
    
    foreach ($groupList as $item) {
        $groupData = [
            'name' => $item['group_name'],
            'product_package_uuid' => $product['uuid'],
            'group_type' => $item['group_type'] ?? 0, // 新增：分组类型
            'optional_count' => $item['optional_count'] ?? 0, // 新增：可选数量
        ];
        
        // ... 保存分组逻辑 ...
        
        foreach ($groupItemList as $item) {
            $itemData = [
                'product_package_group_uuid' => $group['uuid'],
                'related_uuid' => $productBoms[$item['product_id']],
                'product_bom_uuid' => $item['product_id'],
                'num' => $item['num'] ?? 1, // 修改：默认值为1
                'sort' => $item['sort'] ?? 0,
                'add_price' => $item['add_price'] ?? 0, // 新增：加价金额
                'is_required' => $item['is_required'] ?? 0, // 新增：是否必选
                'is_default' => $item['is_default'] ?? 0, // 新增：是否默认
            ];
            
            // ... 保存商品逻辑 ...
        }
        
        // 新增：数据校验
        if ($groupData['group_type'] == 1) { // 可选类型
            $requiredCount = 0;
            foreach ($groupItemList as $item) {
                if (($item['is_required'] ?? 0) == 1) {
                    $requiredCount++;
                }
            }
            if ($requiredCount > ($groupData['optional_count'] ?? 0)) {
                throw new \Exception('必选不可大于可选数量');
            }
        }
    }
}
```

---

## 🔌 API 设计

### RESTful API

#### API 1: 保存套餐组（更新现有接口）

**请求**:

- **URL**: `/api/shop/product/edit` (POST)
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
    "product_package_uuid": 123456,
    "package_group": [
      {
        "group_id": 789012,
        "group_name": "主菜",
        "group_type": 1,
        "optional_count": 2,
        "product_list": [
          {
            "item_id": 345678,
            "product_id": 456789,
            "num": 1,
            "add_price": 0,
            "is_required": 1,
            "is_default": 0
          }
        ]
      }
    ]
  }
  ```

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

**错误响应**:

```json
{
  "code": 0,
  "message": "必选不可大于可选数量",
  "data": {}
}
```

---

## 🧩 组件和接口

### Model 层

#### ProductPackageGroup Model

```php
// admin/app/common/model/product/ProductPackageGroup.php
// 更新 updatePackageGroup 方法

public static function updatePackageGroup($data, $product)
{
    // 1. 支持 group_type 和 optional_count 字段
    // 2. 支持 is_required 和 is_default 字段
    // 3. 支持 add_price 字段
    // 4. num 字段默认值为 1
    // 5. 数据校验：必选数量不可大于可选数量
}
```

### Service 层

无需新增 Service 方法，使用现有的 Service 即可。

### Controller 层

无需修改 Controller，使用现有的 Controller 即可。

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 必选数量大于可选数量

- **处理方式**: 在 Model 层进行校验，抛出异常
- **用户影响**: 前端显示错误提示："必选不可大于可选数量"
- **代码示例**:
  ```php
  if ($requiredCount > $optionalCount) {
      throw new \Exception('必选不可大于可选数量');
  }
  ```

#### 场景 2: 参数格式不正确

- **处理方式**: 在 Controller 或 Validate 层进行校验
- **用户影响**: 返回参数错误信息

---

## 🔒 安全设计

### 身份验证

- **JWT Token**: 所有 API 需要 Token 验证
- **权限控制**: 检查用户是否有商品管理权限

### 数据安全

- **SQL 注入防护**: 使用参数化查询（ThinkPHP ORM）
- **XSS 防护**: 输入校验

---

## 🧪 测试策略

### 单元测试

**测试内容**:

- Model 业务逻辑（数据校验）
- 数据保存逻辑

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
- 数据校验逻辑

---

## 📈 性能优化

### 优化策略

1. **数据库优化**:
   - 使用现有索引
   - 优化查询条件

2. **事务管理**:
   - 使用数据库事务保证数据一致性

### 性能指标

- 本地响应时间: < 200ms
- 数据库查询: < 50ms

---

## 📚 实现清单

### Phase 1: Model 更新

- [ ] 更新 ProductPackageGroup Model（支持 group_type 和 optional_count）
- [ ] 更新 ProductPackageGroupItem Model（支持 is_required、is_default、add_price）
- [ ] 更新 updatePackageGroup 方法

### Phase 2: 核心实现

- [ ] 实现数据校验逻辑
- [ ] 更新参数处理逻辑

### Phase 3: 测试

- [ ] 单元测试
- [ ] API 测试
- [ ] 集成测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-11-25  
**作者**: 开发组  
**审核者**: 待审核
