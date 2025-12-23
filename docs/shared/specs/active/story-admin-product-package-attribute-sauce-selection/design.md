# 旧后台商品管理-成品/套餐可选范围调整 设计文档

> 本文档定义商品管理中属性、加料、套餐分组可选范围调整的技术设计和实现方案。

## 📋 概述

本功能将旧后台商品管理中的"必选+最大可选"模式统一调整为"最小-最大可选范围"模式，涉及数据库表结构调整、数据迁移、后端逻辑修改和前端界面调整。设计遵循向后兼容原则，确保旧数据平滑迁移。

---

## 🎯 规范对齐

### PHP 规范 (php.mdc)

- 遵循 ThinkPHP 6.0 MVC 分层
- Controller 不写业务逻辑
- 使用验证器验证参数
- 使用软删除

### 数据库规范 (database.mdc)

- 字段名使用 snake_case
- 时间字段使用 int 类型，_time 结尾
- 提供数据迁移脚本
- 支持回滚机制

---

## 🔄 代码复用分析

### 可复用的现有组件

- **商品模型**: `admin/app/shop/model/product/Product.php` - add/edit 方法需要调整
- **商品通用模型**: `admin/app/common/model/product/Product.php` - 字段定义需要调整
- **属性模型**: `admin/app/shop/model/product/ProductAttribute.php` - addAttribute/updateAttribute 方法需要调整
- **属性组模型**: `admin/app/common/model/product/AttributeGroup.php` - 字段映射需要调整
- **商品规格模型**: `admin/app/common/model/product/Product.php` - addFlavor/updateFlavor 方法（加料）
- **套餐分组模型**: `admin/app/common/model/product/ProductPackageGroup.php` - 需要添加字段

### 集成点

- **订单验证**: `admin/app/common/model/BaseModelOrder.php` - 订单提交时验证属性和加料可选范围
- **数据库迁移**: `admin/database/migrations/` - 创建迁移脚本
- **前端界面**: `admin/views/shop/pages/product/` - 调整表单字段

---

## 🏗️ 架构设计

### 分层设计原则

**PHP ThinkPHP 三层架构**:

```
Controller 层 (接收请求、参数验证)
  ↓ 调用
Model 层 (业务逻辑、数据处理)
  ↓ 调用
Database 层 (数据持久化)
```

**依赖规则**:

- ✅ Controller 调用 Model
- ❌ Model 不依赖 Controller
- ✅ Model 内部可互相调用

### 模块划分

#### PHP Admin 模块

- **Controller 层**: `admin/app/shop/controller/product/store/Product.php` - 商品管理控制器
- **Model 层**: 
  - `admin/app/shop/model/product/Product.php` - 商品模型（主要修改）
  - `admin/app/common/model/product/Product.php` - 通用商品模型
  - `admin/app/shop/model/product/ProductAttribute.php` - 属性模型
  - `admin/app/common/model/product/AttributeGroup.php` - 属性组模型
  - `admin/app/common/model/product/ProductPackageGroup.php` - 套餐分组模型
- **Database 层**: `admin/database/migrations/` - 数据迁移

---

## 🗄️ 数据库设计

### 数据表设计

#### 表 1: ttpos_product_package（商品包）

**现有字段（需要调整）**:

| 字段 | 类型 | 说明 | 现有值 | 调整后 |
|------|------|------|--------|--------|
| sauce_required | tinyint(1) | 是否必选小料 | 0-否 1-是 | **废弃** |
| sauce_max_selection | int | 小料最大选择数量 | 0-不限制 >0-具体值 | **重定义为最大可选** |

**新增字段**:

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| sauce_min_selection | int | 小料最小选择数量 | DEFAULT 0, ≥0 |

**迁移规则**:

```sql
-- 新增字段
ALTER TABLE ttpos_product_package ADD COLUMN sauce_min_selection int NOT NULL DEFAULT 0 COMMENT '小料最小选择数量' AFTER sauce_required;

-- 数据迁移
UPDATE ttpos_product_package 
SET sauce_min_selection = CASE 
    WHEN sauce_required = 1 THEN 1 
    ELSE 0 
END;

-- sauce_max_selection 保持原值（0表示不限制，>0表示具体值）
-- 如果 sauce_max_selection = 0 且有加料，则设置为加料数量
-- 这个逻辑需要在应用层处理
```

#### 表 2: ttpos_product_attribute_group（商品属性组）

**现有字段（需要调整）**:

| 字段 | 类型 | 说明 | 现有值 | 调整后 |
|------|------|------|--------|--------|
| is_must | tinyint(1) | 是否必选 | 0-否 1-是 | **废弃** |
| max_selection | int | 最大选择数量 | 0-不限制 >0-具体值 | **重定义为最大可选** |

**新增字段**:

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| min_selection | int | 最小选择数量 | DEFAULT 0, ≥0 |

**迁移规则**:

```sql
-- 新增字段
ALTER TABLE ttpos_product_attribute_group ADD COLUMN min_selection int NOT NULL DEFAULT 0 COMMENT '最小选择数量' AFTER is_must;

-- 数据迁移
UPDATE ttpos_product_attribute_group 
SET min_selection = CASE 
    WHEN is_must = 1 THEN 1 
    ELSE 0 
END;

-- max_selection 保持原值（0表示不限制，>0表示具体值）
-- 如果 max_selection = 0 且有属性值，则设置为属性值数量
-- 这个逻辑需要在应用层处理
```

#### 表 3: ttpos_product_package_group（套餐分组）

**新增字段**:

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| min_selection | int | 最小选择数量 | DEFAULT 1, ≥0 |
| max_selection | int | 最大选择数量 | DEFAULT 1, ≥0 |

**迁移规则**:

```sql
-- 新增字段
ALTER TABLE ttpos_product_package_group 
ADD COLUMN min_selection int NOT NULL DEFAULT 1 COMMENT '最小选择数量',
ADD COLUMN max_selection int NOT NULL DEFAULT 1 COMMENT '最大选择数量';

-- 现有套餐分组默认设置为 1-1（可选1个）
UPDATE ttpos_product_package_group 
SET min_selection = 1, max_selection = 1
WHERE min_selection = 0 AND max_selection = 0;
```

### 数据库迁移

> ⚠️ **重要提示**：数据库字段已在任务 #37946 中添加（迁移文件：`20251222145027_add_selection_range_fields.php`），无需重复创建迁移。

**已添加的字段**：

1. **ttpos_product_package**：
   - `sauce_min_selection` (int) - 小料最小选择数量

2. **ttpos_product_package_attribute_group**：
   - `min_selection` (int) - 最小选择数量

3. **ttpos_product_package_group**：
   - `optional_min_count` (int) - 最小可选数量
   - `optional_count` 注释已更新为"最大可选数量"

**数据迁移规则**（已执行）：

- `sauce_required = 1` → `sauce_min_selection = 1`
- `is_must = 1` → `min_selection = 1`
- `group_type = 1` → `optional_min_count = 1`
- `sauce_max_selection = 0` → 自动设置为加料数量
- `max_selection = 0` → 自动设置为属性值数量

**参考文件**：`admin/database/migrations/20251222145027_add_selection_range_fields.php`

---

## 📊 数据模型

### PHP Model 调整

#### 1. 商品模型 (admin/app/shop/model/product/Product.php)

**add() 方法调整**:

```php
// 属性验证逻辑（第113-128行）
// 原逻辑：
if (isset($data['product_attr']) && is_array($data['product_attr']) && !empty($data['product_attr'])) {
    $attr = $data['product_attr'][0];
    // 最多默认勾选数量
    $defaultSelectCount = count(array_filter($attr['default_select'], function ($item) {
        return $item == 1;
    }));
    if (
        $attr['attribute_open_max_select'] == 1 &&
        isset($attr['attribute_value']) &&
        isset($attr['attribute_max_select']) &&
        $defaultSelectCount > $attr['attribute_max_select']
    ) {
        $this->error = '不能超过最多可选数量' . ' ' . $attr['attribute_max_select'];
        return false;
    }
}

// 调整为：
if (isset($data['product_attr']) && is_array($data['product_attr']) && !empty($data['product_attr'])) {
    foreach ($data['product_attr'] as $attr) {
        // 验证最小最大可选范围
        $minSelect = $attr['attribute_min_select'] ?? 0;
        $maxSelect = $attr['attribute_max_select'] ?? 0;
        
        // 验证最大可选 >= 最小可选
        if ($maxSelect > 0 && $maxSelect < $minSelect) {
            $this->error = '最大可选不可小于最小可选';
            return false;
        }
        
        // 验证默认勾选数量
        if (isset($attr['default_select']) && is_array($attr['default_select'])) {
            $defaultSelectCount = count(array_filter($attr['default_select'], function ($item) {
                return $item == 1;
            }));
            
            // 默认勾选数量必须在可选范围内
            if ($defaultSelectCount < $minSelect) {
                $this->error = '默认勾选数量不能少于最小可选数量';
                return false;
            }
            if ($maxSelect > 0 && $defaultSelectCount > $maxSelect) {
                $this->error = '默认勾选数量不能超过最大可选数量';
                return false;
            }
        }
        
        // 验证最大可选不能超过属性值数量
        if ($maxSelect > 0 && isset($attr['attribute_value'])) {
            $attrValueCount = count($attr['attribute_value']);
            if ($maxSelect > $attrValueCount) {
                $this->error = '最大可选数量不能超过属性值数量';
                return false;
            }
        }
    }
}
```

**加料验证逻辑（第189-214行）**:

```php
// 原逻辑：
if (isset($data['product_feed']) && is_array($data['product_feed']) && !empty($data['product_feed'])) {
    if (count($data['product_feed']) > 10) {
        $this->error = '最多可添加10个加料';
        return false;
    }
    // 最多默认勾选数量
    $defaultSelectCount = count(array_filter($data['product_feed'], function ($item) {
        return $item == 1;
    }));
    if (
        $data['feed_open_max_select'] == 1 &&
        isset($data['feed_max_select']) &&
        $defaultSelectCount > $data['feed_max_select']
    ) {
        $this->error = '不能超过最多可选数量' . ' ' . $data['feed_max_select'];
        return false;
    }
    // ...
}

// 调整为：
if (isset($data['product_feed']) && is_array($data['product_feed']) && !empty($data['product_feed'])) {
    // 调整限制数量从10改为100
    if (count($data['product_feed']) > 100) {
        $this->error = '最多可添加100个加料';
        return false;
    }
    
    // 验证最小最大可选范围
    $minSelect = $data['feed_min_select'] ?? 0;
    $maxSelect = $data['feed_max_select'] ?? 0;
    
    // 验证最大可选 >= 最小可选
    if ($maxSelect > 0 && $maxSelect < $minSelect) {
        $this->error = '最大可选不可小于最小可选';
        return false;
    }
    
    // 验证默认勾选数量
    $defaultSelectCount = count(array_filter($data['product_feed'], function ($item) {
        return isset($item['is_default']) && $item['is_default'] == 1;
    }));
    
    // 默认勾选数量必须在可选范围内
    if ($defaultSelectCount < $minSelect) {
        $this->error = '默认勾选数量不能少于最小可选数量';
        return false;
    }
    if ($maxSelect > 0 && $defaultSelectCount > $maxSelect) {
        $this->error = '默认勾选数量不能超过最大可选数量';
        return false;
    }
    
    // 验证最大可选不能超过加料数量
    if ($maxSelect > 0 && $maxSelect > count($data['product_feed'])) {
        $this->error = '最大可选数量不能超过加料数量';
        return false;
    }
    
    // ...
}
```

**套餐分组验证逻辑（第129-187行）**:

```php
// 原逻辑：
if ($isPackage) {
    // ...
    if (count($packageGroup) > 5) {
        $this->error = '套餐分组最多只能设置5个';
        return false;
    }
    foreach ($packageGroup as &$item) {
        // ...
    }
}

// 调整为：
if ($isPackage) {
    // ...
    // 调整限制数量从5改为100
    if (count($packageGroup) > 100) {
        $this->error = '套餐分组最多只能设置100个';
        return false;
    }
    foreach ($packageGroup as $groupIndex => &$item) {
        // ...
        
        // 验证分组可选范围
        $minSelect = $item['group_min_select'] ?? 1;
        $maxSelect = $item['group_max_select'] ?? 1;
        
        // 验证最大可选 >= 最小可选（允许为0）
        if ($maxSelect < $minSelect) {
            $this->error = sprintf('分组%d最大可选不可小于最小可选', $groupIndex + 1);
            return false;
        }
        
        // 验证最大可选不能超过分组商品数量
        $groupProductCount = count($item['product_list'] ?? []);
        if ($maxSelect > $groupProductCount) {
            $this->error = sprintf('分组%d最大可选数量不能超过商品数量', $groupIndex + 1);
            return false;
        }
    }
}
```

**数据映射调整（第254-257行）**:

```php
// 原逻辑：
$data['sauce_required'] = $data['feed_required'] ?? 0; // 是否必选小料, 0-否 1-是
$data['sauce_max_selection'] = $data['feed_max_select'] ?? 0; // 小料最大选择数量

// 调整为：
$data['sauce_min_selection'] = $data['feed_min_select'] ?? 0; // 小料最小选择数量
$data['sauce_max_selection'] = $data['feed_max_select'] ?? 0; // 小料最大选择数量
// sauce_required 保留但不再使用，兼容旧数据
$data['sauce_required'] = ($data['feed_min_select'] ?? 0) > 0 ? 1 : 0;
```

**edit() 方法调整**: 同 add() 方法的调整逻辑

#### 2. 属性模型 (admin/app/shop/model/product/ProductAttribute.php)

**addAttribute() 方法调整（第15-31行）**:

```php
// 原逻辑：
public static function addAttribute($data, $product)
{
    // ...
    $attributeGroup->save([
        'is_must' => $item['attribute_required'] ?? 0, // 是否必选
        'max_selection' => $item['attribute_max_select'] ?? 0, // 最大选择数量
    ]);
}

// 调整为：
public static function addAttribute($data, $product)
{
    // ...
    // 计算默认最大可选（如果未设置）
    $attrValueCount = count($item['attribute_value'] ?? []);
    $maxSelect = $item['attribute_max_select'] ?? $attrValueCount;
    
    $attributeGroup->save([
        'min_selection' => $item['attribute_min_select'] ?? 0, // 最小选择数量
        'max_selection' => $maxSelect, // 最大选择数量
        // 保留 is_must 字段用于兼容旧数据
        'is_must' => ($item['attribute_min_select'] ?? 0) > 0 ? 1 : 0,
    ]);
}
```

**updateAttribute() 方法调整（第52-77行）**: 同 addAttribute() 方法的调整逻辑

#### 3. 套餐分组模型 (admin/app/common/model/product/ProductPackageGroup.php)

**addPackageGroup() 方法调整**: 添加 optional_min_count 和 optional_count 字段

```php
public static function addPackageGroup($data, $product)
{
    foreach ($data['package_group'] as $groupIndex => $group) {
        $packageGroup = new ProductPackageGroupModel();
        $packageGroup->save([
            'product_package_uuid' => $product->uuid,
            'group_name' => $group['group_name'],
            'sort' => $groupIndex,
            'optional_min_count' => $group['group_min_select'] ?? 1, // 最小可选
            'optional_count' => $group['group_max_select'] ?? 1, // 最大可选
        ]);
        
        // 添加分组商品...
    }
}
```

---

## 🔌 API 设计

### 现有 API 保持不变

API 接口保持不变，只调整请求和响应参数：

#### API 1: 添加商品

**请求**:

```json
{
  "product_name": "商品名称",
  "product_attr": [
    {
      "attribute_name": "规格",
      "attribute_value": ["大", "中", "小"],
      "attribute_min_select": 1,
      "attribute_max_select": 1,
      "default_select": [1, 0, 0]
    }
  ],
  "product_feed": [
    {
      "feed_name": "糖度",
      "feed_price": 0
    }
  ],
  "feed_min_select": 0,
  "feed_max_select": 3,
  "package_group": [
    {
      "group_name": "主食",
      "group_min_select": 1,
      "group_max_select": 1,
      "product_list": [...]
    }
  ]
}
```

**响应**:

```json
{
  "code": 1,
  "msg": "添加成功",
  "data": {}
}
```

#### API 2: 编辑商品

同添加商品，请求参数增加 `product_id`

---

## 🧩 组件和接口

### 向后兼容处理

#### 1. 读取旧数据时的兼容

```php
// 在读取商品详情时，兼容旧字段
public function getProductDetail($productId)
{
    $product = Product::find($productId);
    
    // 属性组兼容
    foreach ($product->productAttributeGroup as $group) {
        // 如果新字段为0（未迁移），则使用旧字段
        if ($group->min_selection == 0 && $group->is_must == 1) {
            $group->min_selection = 1;
        }
        if ($group->max_selection == 0) {
            // 计算属性值数量
            $attrValueCount = $group->productAttribute->count();
            $group->max_selection = $attrValueCount;
        }
    }
    
    // 加料兼容
    if ($product->sauce_min_selection == 0 && $product->sauce_required == 1) {
        $product->sauce_min_selection = 1;
    }
    if ($product->sauce_max_selection == 0) {
        // 计算加料数量
        $feedCount = $product->feed->count();
        $product->sauce_max_selection = $feedCount;
    }
    
    return $product;
}
```

#### 2. 保存时的兼容

```php
// 保存商品时，同时更新旧字段
public function saveProduct($data)
{
    // 更新新字段
    $data['sauce_min_selection'] = $data['feed_min_select'] ?? 0;
    $data['sauce_max_selection'] = $data['feed_max_select'] ?? 0;
    
    // 同时更新旧字段（用于兼容旧版前端）
    $data['sauce_required'] = $data['sauce_min_selection'] > 0 ? 1 : 0;
    // feed_open_max_select 保留但不再使用
    $data['feed_open_max_select'] = $data['sauce_max_selection'] > 0 ? 1 : 0;
    
    // ...保存逻辑
}
```

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 最小值大于最大值

- **处理方式**: 返回错误提示
- **用户影响**: 提示"最大可选不可小于最小可选"
- **代码示例**:
  ```php
  if ($maxSelect > 0 && $maxSelect < $minSelect) {
      $this->error = '最大可选不可小于最小可选';
      return false;
  }
  ```

#### 场景 2: 最大可选超过可选项数量

- **处理方式**: 返回错误提示
- **用户影响**: 提示"最大可选数量不能超过属性值数量"
- **代码示例**:
  ```php
  if ($maxSelect > $attrValueCount) {
      $this->error = '最大可选数量不能超过属性值数量';
      return false;
  }
  ```

#### 场景 3: 数量限制超出

- **处理方式**: 返回错误提示
- **用户影响**: 提示"属性组数量不能超过100"

---

## 🔒 安全设计

### 数据验证

- **参数验证**: 所有输入参数必须验证类型和范围
- **SQL 注入防护**: 使用 ThinkPHP ORM，自动防护
- **XSS 防护**: 前端输入过滤

### 向后兼容安全

- **旧字段保留**: sauce_required, is_must 字段保留，避免旧版前端报错
- **双写机制**: 新数据同时更新新旧字段
- **读取兼容**: 读取时优先使用新字段，新字段为0时使用旧字段

---

## 🧪 测试策略

### 测试场景

#### 1. 数据迁移测试

- [ ] 测试 sauce_required = 1 迁移为 sauce_min_selection = 1
- [ ] 测试 is_must = 1 迁移为 min_selection = 1
- [ ] 测试回滚迁移脚本

#### 2. 边界值测试

- [ ] 测试最小值 = 最大值
- [ ] 测试最小值 > 最大值（应报错）
- [ ] 测试最大值 = 0（不限制）
- [ ] 测试最大值超过可选项数量（应报错）

#### 3. 兼容性测试

- [ ] 测试旧数据读取显示正确
- [ ] 测试旧版前端调用新接口
- [ ] 测试新版前端调用接口

#### 4. 数量限制测试

- [ ] 测试属性组数量 = 100（通过）
- [ ] 测试属性组数量 = 101（应报错）
- [ ] 测试加料数量 = 100（通过）
- [ ] 测试加料数量 = 101（应报错）
- [ ] 测试套餐分组数量 = 100（通过）
- [ ] 测试套餐分组数量 = 101（应报错）

---

## 📈 性能优化

### 优化策略

1. **数据库优化**:
   - 新增字段不影响现有索引
   - 数据迁移一次性执行

2. **兼容性处理优化**:
   - 读取时只在新字段为0时才兼容旧字段
   - 减少不必要的计算

---

## 📚 实现清单

### Phase 1: 数据库设计和迁移

- [ ] 创建数据库迁移文件
- [ ] 执行数据库迁移
- [ ] 测试数据迁移结果
- [ ] 测试回滚迁移

### Phase 2: 后端逻辑调整

- [ ] 调整商品模型 add() 方法
- [ ] 调整商品模型 edit() 方法
- [ ] 调整属性模型 addAttribute() 方法
- [ ] 调整属性模型 updateAttribute() 方法
- [ ] 调整套餐分组模型 addPackageGroup() 方法
- [ ] 调整套餐分组模型 updatePackageGroup() 方法
- [ ] 调整通用商品模型字段定义
- [ ] 调整订单验证逻辑（BaseModelOrder.php）

### Phase 3: 前端界面调整

- [ ] 调整属性表单字段（必选+最大可选 → 最小-最大可选）
- [ ] 调整加料表单字段（必选+最大可选 → 最小-最大可选）
- [ ] 调整套餐分组表单字段（添加可选范围）
- [ ] 调整数量限制提示（10→100, 5→100）

### Phase 4: 测试

- [ ] 数据迁移测试
- [ ] 边界值测试
- [ ] 兼容性测试
- [ ] 数量限制测试
- [ ] 手动功能测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-23  
**作者**: 后端开发组  
**审核者**: 待分配

