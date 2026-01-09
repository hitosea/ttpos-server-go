# 新管理端-分类管理-增强分类 设计文档

> 本文档定义新管理端分类管理增强功能的技术设计和实现方案。

## 📋 概述

本功能仅在分类表增加两个字段（`is_display_in_store` 和 `is_display_in_takeout`），用于控制分类在店内和外卖平台的显示。同时修改对应的接口支持这两个字段。

---

## 🎯 规范对齐

### Go Main 规范 (go-main.mdc)

- ✅ Service 只依赖其他 Service 接口，不直接依赖 Repository
- ✅ Repository 只持有 db 实例，不持有 DBManager
- ✅ URL 使用 snake_case
- ✅ data 字段必须是对象，不能是 null 或数组
- ✅ 不使用 panic，返回 error

### 数据库规范 (database.mdc)

- ✅ 必需字段完整：`id`, `uuid`, `create_time`, `update_time`, `delete_time`
- ✅ 时间字段使用 int 类型，\_time 结尾，默认值 0
- ✅ 表名使用 ttpos\_ 前缀
- ✅ 字段名使用 snake_case
- ✅ 迁移前检查字段是否存在（确保幂等性）

---

## 🗄️ 数据库设计

### 数据表设计

#### 表：ttpos_product_category（扩展）

**新增字段**：

```sql
-- 检查字段是否存在，如果不存在则添加
ALTER TABLE `ttpos_product_category`
ADD COLUMN IF NOT EXISTS `is_display_in_store` tinyint(1) NOT NULL DEFAULT 1 COMMENT '是否在店内显示: 1-是 0-否' AFTER `status`,
ADD COLUMN IF NOT EXISTS `is_display_in_takeout` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否在外卖平台显示: 1-是 0-否' AFTER `is_display_in_store`;

-- 添加索引（如需要）
ALTER TABLE `ttpos_product_category`
ADD INDEX IF NOT EXISTS `idx_is_display_in_store` (`is_display_in_store`),
ADD INDEX IF NOT EXISTS `idx_is_display_in_takeout` (`is_display_in_takeout`);
```

**字段说明**:

| 字段 | 类型 | 说明 | 约束 | 默认值 |
|------|------|------|------|--------|
| is_display_in_store | tinyint(1) | 是否在店内显示 | NOT NULL | 1 |
| is_display_in_takeout | tinyint(1) | 是否在外卖平台显示 | NOT NULL | 0 |

**迁移文件**: `admin/database/migrations/{YYYYMMDDHHMMSS}_add_category_display_fields.php`

---

## 📊 数据模型

### Go Model

```go
// main/app/model/category.go
type ProductCategory struct {
	BaseModel
	Name                  string `gorm:"column:name;type:varchar(255);comment:名称;NOT NULL" json:"name"`
	MultiLanguageNameUuid uint64 `gorm:"column:multi_language_name_uuid;type:bigint(20) unsigned;default:0;comment:多语言名称ID;NOT NULL" json:"multi_language_name_uuid"`
	Status                int    `gorm:"column:status;type:tinyint(1);default:0;comment:状态, 1-开启 0-关闭;NOT NULL" json:"status"`
	ParentUuid            uint64 `gorm:"column:parent_uuid;type:bigint(20) unsigned;comment:父级ID" json:"parent_uuid"`
	IsSpecial             int    `gorm:"column:is_special;type:tinyint(1);default:0;comment:特殊分类, 1-是 0-否;NOT NULL" json:"is_special"`
	CategoryKey           string `gorm:"column:category_key;type:varchar(255);comment:关键字;NOT NULL" json:"category_key"`
	Sort                  uint   `gorm:"column:sort;type:int(11);default:0;comment:排序;NOT NULL" json:"sort"`
	Code                  string `gorm:"column:code;type:varchar(255);comment:分类编码;NOT NULL" json:"code"`
	HeadquarterUuid       uint64 `gorm:"column:headquarter_uuid;type:bigint(20) unsigned;default:0;comment:总部ID;NOT NULL" json:"headquarter_uuid"`
	
	// 新增字段
	IsDisplayInStore   int `gorm:"column:is_display_in_store;type:tinyint(1);default:1;comment:是否在店内显示;NOT NULL" json:"is_display_in_store"`
	IsDisplayInTakeout int `gorm:"column:is_display_in_takeout;type:tinyint(1);default:0;comment:是否在外卖平台显示;NOT NULL" json:"is_display_in_takeout"`

	MultiLanguageName MultiLanguageName `gorm:"foreignKey:multi_language_name_uuid;references:uuid"`
}
```

### DTO 定义

#### Request DTO

```go
// main/app/dto/req/product.go（扩展）
type CategoryCreateReq struct {
	Name                string `json:"name" binding:"required"`
	ParentUuid          uint64 `json:"parent_uuid"`
	IsDisplayInStore    *int   `json:"is_display_in_store"`    // 新增：默认 1
	IsDisplayInTakeout  *int   `json:"is_display_in_takeout"`  // 新增：默认 0
	// ... 其他字段
}

type CategoryUpdateReq struct {
	Uuid                uint64 `json:"uuid" binding:"required"`
	Name                string `json:"name"`
	Status              *int   `json:"status"`
	IsDisplayInStore    *int   `json:"is_display_in_store"`    // 新增
	IsDisplayInTakeout  *int   `json:"is_display_in_takeout"`  // 新增
	// ... 其他字段
}
```

#### Response DTO

```go
// main/app/dto/resp/product_resp/category.go（扩展）
type CategoryResp struct {
	Uuid                uint64 `json:"uuid"`
	Name                string `json:"name"`
	Status              int    `json:"status"`
	IsDisplayInStore    int    `json:"is_display_in_store"`    // 新增
	IsDisplayInTakeout  int    `json:"is_display_in_takeout"`  // 新增
	TakeoutProductCount int64  `json:"takeout_product_count"` // 新增：被外卖商品选中的数量
	// ... 其他字段
	CreateTime          int64  `json:"create_time"`
	UpdateTime          int64  `json:"update_time"`
}
```

---

## 🔌 API 设计

### RESTful API

#### API 1: 创建分类（增强）

**请求**:

- **URL**: `/api/v1/shop/product/category/add`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "name": "热菜",
    "is_display_in_store": 1,
    "is_display_in_takeout": 1
  }
  ```

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "uuid": 123456789,
    "name": "热菜",
    "is_display_in_store": 1,
    "is_display_in_takeout": 1
  }
}
```

#### API 2: 编辑分类（增强）

**请求**:

- **URL**: `/api/v1/shop/product/category/edit`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "uuid": 123456789,
    "name": "热菜",
    "is_display_in_store": 1,
    "is_display_in_takeout": 0
  }
  ```

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "uuid": 123456789,
    "name": "热菜",
    "is_display_in_store": 1,
    "is_display_in_takeout": 0
  }
}
```

**业务逻辑**:
1. 验证参数：
   - 店内显示不允许取消（`is_display_in_store` 不能设置为 0）
   - 被 Grab 商品勾选的分类不允许取消外卖显示（如果 `takeout_product_count > 0`，`is_display_in_takeout` 不能设置为 0）
2. 更新分类信息
3. 返回更新结果

#### API 3: 获取分类列表（增强）

**请求**:

- **URL**: `/api/v1/shop/product/category/list`
- **Method**: `GET`

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {
        "uuid": 123456789,
        "name": "热菜",
        "is_display_in_store": 1,
        "is_display_in_takeout": 1
      }
    ],
    "meta": {
      "page_no": 1,
      "page_size": 20,
      "total": 100
    }
  }
}
```

#### API 4: 获取分类详情（增强）

**请求**:

- **URL**: `/api/v1/shop/product/category`
- **Method**: `GET`
- **Query**: `?uuid=123456789`

**响应**:

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "uuid": 123456789,
    "name": "热菜",
    "is_display_in_store": 1,
    "is_display_in_takeout": 1,
    "takeout_product_count": 5
  }
}
```

**业务逻辑**:
1. 查询分类基本信息
2. 统计该分类被外卖商品选中的数量（查询 `ttpos_product_package_takeout` 表中 `category_uuid` 等于分类 `uuid` 且 `delete_time = 0` 的记录数）
3. 返回分类信息和外卖商品数量

---

## 🧩 组件和接口

### Service 层

#### Service 实现（增强现有方法）

**查询外卖商品数量逻辑**:

```go
// 统计分类被外卖商品选中的数量
func (s *categorySrv) getTakeoutProductCount(db *gorm.DB, categoryUuid uint64) (int64, error) {
	var count int64
	err := db.Model(&model.ProductPackageTakeout{}).
		Where("category_uuid = ? AND delete_time = ?", categoryUuid, 0).
		Count(&count).Error
	return count, err
}
```

**GetByUuid 方法增强**:

```go
// main/app/service/category_srv.go（增强）
func (s *categorySrv) GetByUuid(ctx *gin.Context, uuid uint64) (*dto_resp.CategoryResp, error) {
	// 获取 Repository
	categoryRepo := repository.NewCategoryRepo(s.dbm.GetDB(ctx))
	
	// 查询分类
	category, err := categoryRepo.GetByUuid(uuid)
	if err != nil {
		return nil, errors.WithMessage(err, "分类不存在")
	}
	
	// 统计被外卖商品选中的数量
	db := s.dbm.GetDB(ctx)
	takeoutProductCount, err := s.getTakeoutProductCount(db, uuid)
	if err != nil {
		// 如果查询失败，设置为 0，不影响主流程
		takeoutProductCount = 0
	}
	
	// 返回响应
	return &dto_resp.CategoryResp{
		Uuid:                category.Uuid,
		Name:                category.Name,
		IsDisplayInStore:    category.IsDisplayInStore,
		IsDisplayInTakeout:  category.IsDisplayInTakeout,
		TakeoutProductCount: takeoutProductCount, // 新增字段
		CreateTime:          category.CreateTime,
		UpdateTime:          category.UpdateTime,
	}, nil
}

func (s *categorySrv) Create(ctx *gin.Context, req *dto_req.CategoryCreateReq) (*dto_resp.CategoryResp, error) {
	// 设置默认值
	isDisplayInStore := 1
	if req.IsDisplayInStore != nil {
		isDisplayInStore = *req.IsDisplayInStore
	}
	isDisplayInTakeout := 0
	if req.IsDisplayInTakeout != nil {
		isDisplayInTakeout = *req.IsDisplayInTakeout
	}
	
	// 验证：至少开启一个显示渠道
	if isDisplayInStore == 0 && isDisplayInTakeout == 0 {
		return nil, errors.New("至少需要开启一个显示渠道")
	}
	
	// 创建分类（包含新字段）
	category := &model.ProductCategory{
		Uuid:              pkg_uuid.GenerateUuid(),
		Name:              req.Name,
		IsDisplayInStore:  isDisplayInStore,
		IsDisplayInTakeout: isDisplayInTakeout,
		// ... 其他字段
	}
	
	// 保存...
}

func (s *categorySrv) Update(ctx *gin.Context, req *dto_req.CategoryUpdateReq) (*dto_resp.CategoryResp, error) {
	// 如果修改了显示字段，验证至少开启一个
	if req.IsDisplayInStore != nil && req.IsDisplayInTakeout != nil {
		if *req.IsDisplayInStore == 0 && *req.IsDisplayInTakeout == 0 {
			return nil, errors.New("至少需要开启一个显示渠道")
		}
	}
	
	// 更新分类（包含新字段）
	updateData := map[string]interface{}{
		"update_time": time.Now().Unix(),
	}
	if req.IsDisplayInStore != nil {
		updateData["is_display_in_store"] = *req.IsDisplayInStore
	}
	if req.IsDisplayInTakeout != nil {
		updateData["is_display_in_takeout"] = *req.IsDisplayInTakeout
	}
	// ... 其他字段更新
	
	// 更新...
}
```

---

## 🚨 错误处理

### 错误场景

#### 场景 1: 尝试取消店内显示

- **处理方式**: 参数验证时检查，返回错误
- **用户影响**: 用户看到"店内显示不允许取消"
- **代码示例**:
  ```go
  if isDisplayInStore == 0 {
      return errors.New("店内显示不允许取消")
  }
  ```

#### 场景 2: 被 Grab 商品勾选的分类尝试取消外卖显示

- **处理方式**: 先统计外卖商品数量，如果数量 > 0，返回错误
- **用户影响**: 用户看到"该分类已被外卖商品使用，不允许取消外卖显示"
- **代码示例**:
  ```go
  if editReq.IsDisplayInTakeout != nil && *editReq.IsDisplayInTakeout == 0 {
      var takeoutProductCount int64
      err = db.Model(&model.ProductPackageTakeout{}).
          Where("category_uuid = ? AND delete_time = ?", productCategory.Uuid, 0).
          Count(&takeoutProductCount).Error
      if err == nil && takeoutProductCount > 0 {
          return errors.New("该分类已被外卖商品使用，不允许取消外卖显示")
      }
  }
  ```

---

## 📚 实现清单

### Phase 1: 数据库和模型

- [ ] 创建数据库迁移文件
- [ ] 执行数据库迁移
- [ ] 更新 Go Model

### Phase 2: DTO 和 Service

- [ ] 更新 Request DTO
- [ ] 更新 Response DTO
- [ ] 增强 Service 方法（Create, Update）

### Phase 3: API 层

- [ ] 增强分类创建 API
- [ ] 增强分类编辑 API
- [ ] 增强分类查询 API（列表、详情）

### Phase 4: 测试

- [ ] 单元测试
- [ ] API 测试

**详细任务**: 参见 `tasks.md`

---

**版本**: v1.0.0  
**创建日期**: 2025-12-09  
**作者**: weifashi  
**审核者**: 待确认
