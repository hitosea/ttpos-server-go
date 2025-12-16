# 确保套餐分组在各个端排序一致 设计文档

> 本文档定义确保套餐分组在各个端排序一致功能的技术设计和实现方案。

## 📋 概述

修复新管理端套餐分组排序保存未生效的问题，通过在数据库表中添加排序字段，并在保存和查询时正确处理排序，确保套餐分组在所有终端显示一致的排序顺序。

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- Service 只依赖其他 Service 接口
- Repository 只持有 db 实例
- URL 使用 snake_case
- data 字段必须是对象
- 不使用 panic，返回 error

### API 设计规范 (api.mdc)

- URL 使用 snake_case
- 响应格式统一：`{code, message, data{}}`
- data 不能为 null 或数组

### 数据库规范 (database.mdc)

- 必需字段完整：`id`, `uuid`, `create_time`, `update_time`, `delete_time`
- 时间字段使用 int 类型
- 字段名使用 snake_case
- 表名使用 ttpos\_ 前缀

---

## 🔄 代码复用分析

### 可复用的现有组件

- **ProductPackageGroup Model (Go)**: `main/app/model/product_package_group.go` - 套餐分组模型
- **ProductPackageGroupRepo**: `main/app/repository/product_package_group.go` - 套餐分组仓库
- **ProductService**: `main/app/service/product.go` - 商品服务，包含 `SaveProductPackageGroup` 方法
- **ProductPackageGroup Model (PHP)**: `admin/app/common/model/product/ProductPackageGroup.php` - PHP 套餐分组模型，包含 `addPackageGroup` 和 `updatePackageGroup` 方法
- **Product Model (PHP)**: `admin/app/common/model/product/Product.php` - PHP 商品模型，包含 `productPackageGroup` 关联
- **数据库迁移模板**: `admin/database/migrations/20251131980000_add_group_type_and_optional_count_to_product_package_group_table.php` - 参考迁移文件格式

### 集成点

- **新管理端套餐保存接口**: 
  - `POST /api/v1/shop/product/add` - 添加套餐时设置排序
  - `POST /api/v1/shop/product/edit` - 编辑套餐时设置排序
- **旧管理端套餐保存接口**: 
  - `POST /api/shop/product.store.product/add` - 添加套餐时设置排序（调用 `ProductPackageGroup::addPackageGroup()`）
  - `POST /api/shop/product.store.product/edit` - 编辑套餐时设置排序（调用 `ProductPackageGroup::updatePackageGroup()`）
- **套餐查询接口**: 查询套餐分组时按 sort 字段排序（新管理端和旧管理端）
- **数据库表**: `ttpos_product_package_group` - 添加 sort 字段（新管理端和旧管理端共用）

---

## 🏗️ 架构设计

### 分层设计原则

**Go Main 三层架构**:

```
API 层 (Controller/API)
  ↓ 依赖
业务层 (Service)
  ↓ 依赖
数据层 (Repository)
```

### 模块划分

#### Go Main 模块

- **Model 层**: `main/app/model/product_package_group.go` - 添加 Sort 字段
- **Repository 层**: `main/app/repository/product_package_group.go` - 查询时按 sort 排序
- **Service 层**: `main/app/service/product.go` - 保存时设置 sort 值
- **API 层**: 无需修改，使用现有接口

#### PHP Admin 模块

- **Model 层**: `admin/app/common/model/product/ProductPackageGroup.php` - 保存时设置 sort 值
- **关联查询**: `admin/app/common/model/product/Product.php` - 查询时按 sort 排序
- **Controller 层**: 无需修改，使用现有接口

---

## 🗄️ 数据库设计

### 数据表设计

#### 表: ttpos_product_package_group

**添加字段**:

```sql
ALTER TABLE `ttpos_product_package_group` 
ADD COLUMN `sort` int NOT NULL DEFAULT 0 COMMENT '排序字段，数值越小越靠前' 
AFTER `optional_count`;
```

**字段说明**:

| 字段 | 类型 | 说明 | 约束 |
|------|------|------|------|
| sort | int | 排序字段，数值越小越靠前 | NOT NULL DEFAULT 0 |

**迁移文件**: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_sort_to_product_package_group_table.php`

### 数据库迁移

**迁移脚本**:

```bash
# 创建迁移文件
cd admin
php think migrate:create AddSortToProductPackageGroupTable

# 执行迁移
php think migrate:run
```

**同步 Go Model**:

在 `main/app/model/product_package_group.go` 中添加 Sort 字段

---

## 📊 数据模型

### Go Model

```go
// main/app/model/product_package_group.go
type ProductPackageGroup struct {
	BaseModel
	Name                  string `json:"name" gorm:"type:text;comment:名称"`
	MultiLanguageNameUuid uint64 `json:"multi_language_name_uuid" gorm:"index:idx_multi_language_name_uuid;not null;default:0;comment:多语言名称ID"`
	ProductPackageUuid    uint64 `json:"product_package_uuid" gorm:"index:idx_product_package_uuid;not null;default:0;comment:商品套餐UUID"`
	GroupType             int    `json:"group_type" gorm:"type:tinyint;not null;default:0;comment:分组类型 0-固定 1-可选"`
	OptionalCount         int    `json:"optional_count" gorm:"type:int;not null;default:0;comment:可选数量，表示本组商品中要求选择多少个商品"`
	Sort                  int    `json:"sort" gorm:"type:int;not null;default:0;comment:排序字段，数值越小越靠前"` // ⭐ 新增字段

	ProductPackageGroupItems []ProductPackageGroupItem `gorm:"foreignKey:product_package_group_uuid;references:uuid"` // 商品套餐组商品
	MultiLanguageName        MultiLanguageName         `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`   // 多语言名称
	ProductPackage           *ProductPackage           `gorm:"foreignKey:product_package_uuid;references:uuid"`       // 商品套餐
}
```

---

## 🔌 API 设计

### RESTful API

#### API 1: 新管理端保存套餐（更新现有接口）

**请求**:

- **URL**: `/api/v1/shop/product/add` (POST) 或 `/api/v1/shop/product/edit` (POST)
- **Method**: `POST`
- **Body**:
  ```json
  {
    "package": {
      "groups": [
        {
          "uuid": 123456,
          "locale_name": {...},
          "group_type": 0,
          "optional_count": 1,
          "products": [...]
        },
        {
          "uuid": 789012,
          "locale_name": {...},
          "group_type": 1,
          "optional_count": 2,
          "products": [...]
        }
      ]
    }
  }
  ```

**变更说明**:

- 前端传递的分组数组顺序即为排序顺序
- 后端根据数组索引自动设置 sort 值（索引从 0 开始，sort = index + 1，或直接使用 index）

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

#### API 2: 旧管理端添加套餐（更新现有接口）

**请求**:

- **URL**: `/api/shop/product.store.product/add` (POST)
- **Method**: `POST`
- **Body**:
  ```json
  {
    "type": 30,
    "package_group": [
      {
        "group_name": {...},
        "group_type": 0,
        "optional_count": 1,
        "product_list": [...]
      },
      {
        "group_name": {...},
        "group_type": 1,
        "optional_count": 2,
        "product_list": [...]
      }
    ]
  }
  ```

**变更说明**:

- 前端传递的 `package_group` 数组顺序即为排序顺序
- 后端在 `ProductPackageGroup::addPackageGroup()` 方法中根据数组索引自动设置 sort 值（索引从 0 开始，sort = index + 1）

**响应**:

```json
{
  "code": 1,
  "message": "添加成功",
  "data": {}
}
```

#### API 3: 旧管理端编辑套餐（更新现有接口）

**请求**:

- **URL**: `/api/shop/product.store.product/edit` (POST)
- **Method**: `POST`
- **Body**:
  ```json
  {
    "product_id": 123456,
    "type": 30,
    "package_group": [
      {
        "group_id": 789012,
        "group_name": {...},
        "group_type": 0,
        "optional_count": 1,
        "product_list": [...]
      },
      {
        "group_id": 345678,
        "group_name": {...},
        "group_type": 1,
        "optional_count": 2,
        "product_list": [...]
      }
    ]
  }
  ```

**变更说明**:

- 前端传递的 `package_group` 数组顺序即为排序顺序
- 后端在 `ProductPackageGroup::updatePackageGroup()` 方法中根据数组索引自动设置 sort 值（索引从 0 开始，sort = index + 1）

**响应**:

```json
{
  "code": 1,
  "message": "更新成功",
  "data": {}
}
```

#### API 4: 查询套餐（更新现有接口）

**请求**:

- **URL**: `/api/v1/shop/product/{uuid}` (GET)
- **Method**: `GET`

**变更说明**:

- 查询套餐分组时，按 `sort` 字段升序排序，相同 sort 值时按 `id` 字段升序排序
- 确保所有终端显示相同的排序顺序

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "package_group_list": {
      "list": [
        {
          "uuid": 123456,
          "locale_name": {...},
          "group_type": 0,
          "optional_count": 1,
          "sort": 1,
          "products": [...]
        },
        {
          "uuid": 789012,
          "locale_name": {...},
          "group_type": 1,
          "optional_count": 2,
          "sort": 2,
          "products": [...]
        }
      ]
    }
  }
}
```

---

## 🧩 组件和接口

### Service 层

#### SaveProductPackageGroup 方法更新

```go
// main/app/service/product.go
func (s *productSrv) SaveProductPackageGroup(tx *gorm.DB, groupList []CheckProductPackageGroupResult, productPackageUuid uint64) error {
	commonRepo := repository.NewCommonRepo()
	productPackageGroupRepo := repository.NewProductPackageGroupRepo(tx)
	multiLanguageNameRepo := repository.NewMultiLanguageNameRepo(tx)
	productBomRepo := repository.NewProductBomRepo(tx)

	// ⭐ 根据数组索引设置排序值
	for index, group := range groupList {
		sortValue := index + 1 // 排序从 1 开始，或使用 index（从 0 开始）
		
		if group.IsDelete {
			// ... 删除逻辑
		} else {
			if group.Uuid == 0 {
				// 新增分组
				// ... 保存多语言名称
				err = productPackageGroupRepo.CreateProductPackageGroup(&model.ProductPackageGroup{
					BaseModel:             model.BaseModel{Uuid: groupUuid},
					Name:                  group.LocaleName.ToJson(),
					MultiLanguageNameUuid: multiLanguageNameUuid,
					ProductPackageUuid:    productPackageUuid,
					GroupType:             group.GroupType,
					OptionalCount:         group.OptionalCount,
					Sort:                  sortValue, // ⭐ 设置排序值
				})
				// ... 保存分组商品
			} else {
				// 更新分组
				// ... 保存多语言名称
				err = productPackageGroupRepo.UpdateProductPackageGroup(map[string]any{
					"name":           group.LocaleName.ToJson(),
					"group_type":      group.GroupType,
					"optional_count":  group.OptionalCount,
					"sort":            sortValue, // ⭐ 更新排序值
				}, commonRepo.WhereByUuid(group.Uuid))
				// ... 更新分组商品
			}
		}
	}
	return nil
}
```

### Repository 层

#### WithProductPackageGroup 预加载方法更新

```go
// main/app/repository/product_package_group.go
// WithProductPackageGroup 预加载商品套餐组
func (r *productPackageGroupRepoImpl) WithProductPackageGroup(opts ...DBOption) DBOption {
	return func(db *gorm.DB) *gorm.DB {
		return db.Preload("ProductPackageGroup", func(db *gorm.DB) *gorm.DB {
			// ⭐ 按 sort 字段升序排序，相同 sort 值时按 id 升序排序
			db = db.Order("sort ASC, id ASC")
			for _, opt := range opts {
				db = opt(db)
			}
			return db
		})
	}
}
```

#### 查询方法更新

所有查询套餐分组的方法都需要按 `sort` 字段排序：

```go
// main/app/repository/product_package_group.go
// GetProductPackageGroup 获取商品套餐组
func (r *productPackageGroupRepoImpl) GetProductPackageGroup(opts ...DBOption) (*model.ProductPackageGroup, error) {
	var productPackageGroup model.ProductPackageGroup
	db := r.db.Order("sort ASC, id ASC") // ⭐ 添加排序，相同 sort 值时按 id 升序排序

	for _, opt := range opts {
		db = opt(db)
	}

	err := db.First(&productPackageGroup).Error
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return &productPackageGroup, nil
}
```

### PHP Admin 模块

#### ProductPackageGroup Model 更新

##### addPackageGroup 方法（添加套餐时调用）

```php
// admin/app/common/model/product/ProductPackageGroup.php

/**
 * 添加套餐商品组
 */
public static function addPackageGroup($data, $product)
{
    $insertGroups = [];
    $insertGroupItems = [];

    $packageGroup = $data['package_group'] ?? [];
    // ⭐ 根据数组索引设置排序值
    foreach ($packageGroup as $index => $item) {
        $sortValue = $index + 1; // 排序从 1 开始
        
        // ... 数据校验逻辑
        
        $insertGroups[] = [
            'uuid' => $groupUuid,
            'name' => $item['group_name'],
            'multi_language_name_uuid' => $multiLanguageNameUuid,
            'product_package_uuid' => $product['uuid'],
            'group_type' => $groupData['group_type'],
            'optional_count' => $groupData['optional_count'],
            'sort' => $sortValue, // ⭐ 设置排序值
            'create_time' => time(),
            'update_time' => time(),
        ];
        // ... 保存分组商品
    }
    // ... 批量保存
}
```

**调用位置**: `admin/app/shop/model/product/Product.php` 的 `add()` 方法中，当 `$isPackage` 为 true 时调用。

##### updatePackageGroup 方法（编辑套餐时调用）

```php
// admin/app/common/model/product/ProductPackageGroup.php

/**
 * 更新套餐分组
 */
public static function updatePackageGroup($data, $product)
{
    $groupUuidList = [];
    $groupItemUuidList = [];
    $groupList = $data['package_group'];
    
    // ⭐ 根据数组索引设置排序值
    foreach ($groupList as $index => $item) {
        $sortValue = $index + 1; // 排序从 1 开始
        
        $groupData = [
            'name' => $item['group_name'],
            'product_package_uuid' => $product['uuid'],
            'group_type' => $item['group_type'] ?? 0,
            'optional_count' => $item['optional_count'] ?? 0,
            'sort' => $sortValue, // ⭐ 设置排序值
        ];
        
        // ... 保存或更新分组逻辑
        if ($groupUuid == 0) {
            // 新增分组
            $groupData['multi_language_name_uuid'] = $multiLanguageNameUuid;
            $group = self::create($groupData);
        } else {
            // 更新分组
            $groupData['multi_language_name_uuid'] = $multiLanguageNameUuid;
            $group->save($groupData);
        }
        // ... 保存分组商品
    }
    // ... 删除逻辑
}
```

**调用位置**: `admin/app/shop/model/product/Product.php` 的 `edit()` 方法中，当 `$isPackage` 为 true 时调用。

#### Product Model 关联查询更新

```php
// admin/app/common/model/product/Product.php

/**
 * 关联套餐分组
 */
public function productPackageGroup()
{
    return $this->hasMany(ProductPackageGroup::class, 'product_package_uuid', 'uuid')
        ->order('sort', 'asc') // ⭐ 添加排序
        ->order('id', 'asc'); // ⭐ 相同 sort 值时按 id 升序排序
}
```

或者在查询时使用：

```php
// admin/app/common/model/product/Product.php
// 在 getPackageInfo 等方法中

$product = Product::with([
    'productPackageGroup' => function ($q) {
        $q->order('sort', 'asc') // ⭐ 添加排序
          ->order('id', 'asc'); // ⭐ 相同 sort 值时按 id 升序排序
    },
    'productPackageGroup.productPackageGropItem' => function ($q) {
        // ... 其他关联
    }
])->find($productId);
```

---

## ⚡ 缓存设计

### Redis 缓存

**缓存策略**:

- 套餐分组数据变更时，清除相关缓存
- 缓存 Key: `ttpos:product:package:{uuid}:groups`

**更新策略**: Cache-Aside Pattern

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 排序值未设置

- **处理方式**: 默认使用数组索引作为排序值
- **用户影响**: 无影响，排序正常工作

#### 场景 2: 数据库迁移失败

- **处理方式**: 回滚迁移，检查错误日志
- **用户影响**: 功能不可用，需要修复后重新迁移

---

## 🔒 安全设计

### 身份验证

- **JWT Token**: 所有 API 需要 Token 验证

### 权限控制

- **商家权限**: 只能修改自己店铺的套餐分组排序

### 数据安全

- **SQL 注入防护**: 使用参数化查询
- **XSS 防护**: 前端输入校验

---

## 🧪 测试策略

### 单元测试

**目标覆盖率**:

- main/app/service: 70%+
- main/app/repository: 80%+

**测试内容**:

- Service 保存分组时排序值设置正确
- Repository 查询时按 sort 字段排序
- 数组顺序与数据库排序一致

### API 测试

**测试内容**:

- 保存套餐分组时排序正确
- 查询套餐分组时排序正确
- 多个分组排序一致性

### 集成测试

**测试流程**:

- 端到端测试：保存 → 查询 → 验证排序
- 多终端测试：新管理端、POS 端排序一致

---

## 📈 性能优化

### 优化策略

1. **数据库优化**:
   - 优化 SQL 查询
   - 使用现有索引（product_package_uuid 已有索引）

2. **查询优化**:
   - 优化排序查询性能
   - 避免全表扫描

### 性能指标

- 本地响应时间: < 200ms
- 数据库查询: < 50ms

---

## 📚 实现清单

### Phase 1: 数据库和模型

- [ ] 创建数据库迁移文件
- [ ] 执行数据库迁移
- [ ] 更新 Go Model

### Phase 2: 核心实现（Go Main）

- [ ] 更新 Repository 查询方法（添加排序）
- [ ] 更新 Service 保存方法（设置排序值）
- [ ] 更新预加载方法（添加排序）

### Phase 3: PHP Admin 模块实现

- [ ] 更新 ProductPackageGroup Model（addPackageGroup 方法设置排序）
- [ ] 更新 ProductPackageGroup Model（updatePackageGroup 方法设置排序）
- [ ] 更新 Product Model（关联查询添加排序）

### Phase 4: 测试和优化

- [ ] 单元测试
- [ ] API 测试
- [ ] 集成测试
- [ ] 性能测试

**详细任务**: 参见 `tasks.md`

---

## Graphiti & 活动日志

- Related Episode: `[待补充]`
- 模板：`docs/agent/templates/graphiti-episode.md`
- 活动日志：`docs/team/activities/{YYYY-MM}/{YYYY-MM-DD}.md`
- 在技术方案评审、关键架构决策或踩坑总结后，立即补充 Episode 并在设计文档尾部互链。

---

**版本**: v1.0.0  
**创建日期**: 2025-12-02  
**作者**: 王昱  
**审核者**: 待审核

