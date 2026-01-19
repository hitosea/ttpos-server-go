# 盘点单模板功能实现方案

> 📖 **用途**: 实现在 TTPOS 系统中创建盘点单模板（日盘、周盘、月盘），配置盘点物品，并快速应用到创建盘点单中

---

## 一、功能概述

### 1.1 业务需求

- ✅ 创建盘点单模板（日盘、周盘、月盘）
- ✅ 模板可以配置盘点物品列表
- ✅ 创建盘点单时可以选择模板，快速填充物品和配置
- ✅ 模板支持按公司、仓库管理

### 1.2 功能特点

- **模板类型**：日盘、周盘、月盘
- **物品配置**：支持配置多个物品，每个物品可设置默认单位
- **快速应用**：创建盘点单时选择模板，自动填充物品列表和配置
- **灵活修改**：应用模板后仍可修改物品和数量

---

## 二、数据库设计

### 2.1 盘点单模板表

```sql
-- 盘点单模板表
CREATE TABLE `ttpos_stock_reconciliation_template` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '模板ID',
  `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '模板UUID',
  `company_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '公司UUID',
  `template_name` VARCHAR(100) NOT NULL COMMENT '模板名称',
  `template_type` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '模板类型：1-日盘 2-周盘 3-月盘',
  `warehouse_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '默认仓库UUID',
  `purpose` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '盘点目的：1-库存盘点 2-期初盘点',
  `type` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '盘点类型：1-指定物品盘点 2-全部物品盘点',
  `remark` VARCHAR(500) DEFAULT '' COMMENT '备注',
  `is_default` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否默认模板：0-否 1-是',
  `status` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '状态：0-禁用 1-启用',
  `created_at` INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间',
  `updated_at` INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间',
  `deleted_at` INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`),
  KEY `idx_company_uuid` (`company_uuid`),
  KEY `idx_template_type` (`template_type`),
  KEY `idx_status` (`status`),
  KEY `idx_warehouse_uuid` (`warehouse_uuid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='盘点单模板表';
```

### 2.2 盘点单模板物品明细表

```sql
-- 盘点单模板物品明细表
CREATE TABLE `ttpos_stock_reconciliation_template_item` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '明细ID',
  `uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '明细UUID',
  `template_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '模板UUID',
  `material_uuid` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '物品UUID',
  `sort_order` INT(11) NOT NULL DEFAULT 0 COMMENT '排序顺序',
  `created_at` INT(11) NOT NULL DEFAULT 0 COMMENT '创建时间',
  `updated_at` INT(11) NOT NULL DEFAULT 0 COMMENT '更新时间',
  `deleted_at` INT(11) NOT NULL DEFAULT 0 COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_uuid` (`uuid`),
  KEY `idx_template_uuid` (`template_uuid`),
  KEY `idx_material_uuid` (`material_uuid`),
  KEY `idx_sort_order` (`sort_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='盘点单模板物品明细表';
```

---

## 三、Model 定义

### 3.1 StockReconciliationTemplate Model

```go
// 文件：main/app/model/stock_reconciliation_template.go
package model

// StockReconciliationTemplate 盘点单模板表
type StockReconciliationTemplate struct {
	BaseModel
	CompanyUuid  uint64 `gorm:"column:company_uuid;not null;default:0;index:idx_company_uuid;comment:公司UUID" json:"company_uuid"`
	TemplateName string `gorm:"column:template_name;type:varchar(100);not null;comment:模板名称" json:"template_name"`
	TemplateType int    `gorm:"column:template_type;not null;default:1;index:idx_template_type;comment:模板类型：1-日盘 2-周盘 3-月盘" json:"template_type"`
	WarehouseUuid uint64 `gorm:"column:warehouse_uuid;not null;default:0;index:idx_warehouse_uuid;comment:默认仓库UUID" json:"warehouse_uuid"`
	Purpose      int    `gorm:"column:purpose;not null;default:1;comment:盘点目的：1-库存盘点 2-期初盘点" json:"purpose"`
	Type         int    `gorm:"column:type;not null;default:1;comment:盘点类型：1-指定物品盘点 2-全部物品盘点" json:"type"`
	Remark       string `gorm:"column:remark;type:varchar(500);default:'';comment:备注" json:"remark"`
	IsDefault    int    `gorm:"column:is_default;not null;default:0;comment:是否默认模板：0-否 1-是" json:"is_default"`
	Status       int    `gorm:"column:status;not null;default:1;index:idx_status;comment:状态：0-禁用 1-启用" json:"status"`

	// 关联仓库
	Warehouse *Warehouse `gorm:"foreignKey:WarehouseUuid;references:Uuid"`
	// 关联模板物品明细
	TemplateItems []*StockReconciliationTemplateItem `gorm:"foreignKey:TemplateUuid;references:Uuid"`
}

// StockReconciliationTemplateItem 盘点单模板物品明细表
type StockReconciliationTemplateItem struct {
	BaseModel
	TemplateUuid uint64 `gorm:"column:template_uuid;not null;default:0;index:idx_template_uuid;comment:模板UUID" json:"template_uuid"`
	MaterialUuid  uint64 `gorm:"column:material_uuid;not null;default:0;index:idx_material_uuid;comment:物品UUID" json:"material_uuid"`
	SortOrder     int    `gorm:"column:sort_order;not null;default:0;index:idx_sort_order;comment:排序顺序" json:"sort_order"`

	// 关联物品
	Material *Material `gorm:"foreignKey:MaterialUuid;references:Uuid"`
}
```

### 3.2 常量定义

```go
// 文件：main/app/constant/stock_reconciliation_template.go
package constant

const (
	// StockReconciliationTemplateTypeDaily 模板类型-日盘
	StockReconciliationTemplateTypeDaily = 1
	// StockReconciliationTemplateTypeWeekly 模板类型-周盘
	StockReconciliationTemplateTypeWeekly = 2
	// StockReconciliationTemplateTypeMonthly 模板类型-月盘
	StockReconciliationTemplateTypeMonthly = 3
)
```

---

## 四、API 接口设计

### 4.1 模板管理接口

#### 4.1.1 创建模板

**请求**：
```go
// 文件：main/app/dto/req/stock_reconciliation_template.go
type StockReconciliationTemplateCreateReq struct {
	TemplateName string   `json:"template_name" binding:"required"` // 模板名称
	TemplateType int      `json:"template_type" binding:"required,oneof=1 2 3"` // 模板类型：1-日盘 2-周盘 3-月盘
	WarehouseUuid uint64  `json:"warehouse_uuid" binding:"required"` // 默认仓库UUID
	Purpose      int      `json:"purpose" binding:"required,oneof=1 2"` // 盘点目的：1-库存盘点 2-期初盘点
	Type         int      `json:"type" binding:"required,oneof=1 2"` // 盘点类型：1-指定物品盘点 2-全部物品盘点
	MaterialUuids []uint64 `json:"material_uuids"` // 物品UUID列表（可选）
	Remark       string   `json:"remark"` // 备注
	IsDefault    bool     `json:"is_default"` // 是否默认模板
}
```

**响应**：
```go
// 文件：main/app/dto/resp/stock_reconciliation_template.go
type StockReconciliationTemplateCreateResp struct {
	Uuid uint64 `json:"uuid"` // 模板UUID
}
```

#### 4.1.2 更新模板

**请求**：
```go
type StockReconciliationTemplateUpdateReq struct {
	Uuid         uint64   `json:"uuid" binding:"required"` // 模板UUID
	TemplateName string   `json:"template_name"` // 模板名称
	TemplateType int      `json:"template_type"` // 模板类型
	WarehouseUuid uint64  `json:"warehouse_uuid"` // 默认仓库UUID
	Purpose      int      `json:"purpose"` // 盘点目的
	Type         int      `json:"type"` // 盘点类型
	MaterialUuids []uint64 `json:"material_uuids"` // 物品UUID列表
	Remark       string   `json:"remark"` // 备注
	IsDefault    bool     `json:"is_default"` // 是否默认模板
	Status       int      `json:"status"` // 状态：0-禁用 1-启用
}
```

#### 4.1.3 查询模板列表

**请求**：
```go
type StockReconciliationTemplateListReq struct {
	PageNo       int   `json:"page_no" form:"page_no" binding:"required,min=1"` // 页码
	PageSize     int   `json:"page_size" form:"page_size" binding:"required,min=1"` // 每页数量
	TemplateType int   `json:"template_type" form:"template_type"` // 模板类型筛选：1-日盘 2-周盘 3-月盘
	WarehouseUuid uint64 `json:"warehouse_uuid" form:"warehouse_uuid"` // 仓库UUID筛选
	Status       int   `json:"status" form:"status"` // 状态筛选：0-禁用 1-启用
	Keyword      string `json:"keyword" form:"keyword"` // 关键字搜索（模板名称）
}
```

**响应**：
```go
type StockReconciliationTemplateListResp struct {
	List     []*StockReconciliationTemplateDetail `json:"list"` // 模板列表
	Total    int64                                 `json:"total"` // 总数
	PageNo   int                                   `json:"page_no"` // 页码
	PageSize int                                   `json:"page_size"` // 每页数量
}

type StockReconciliationTemplateDetail struct {
	Uuid         uint64   `json:"uuid"` // 模板UUID
	TemplateName string   `json:"template_name"` // 模板名称
	TemplateType int      `json:"template_type"` // 模板类型
	TemplateTypeName string `json:"template_type_name"` // 模板类型名称
	WarehouseUuid uint64  `json:"warehouse_uuid"` // 默认仓库UUID
	WarehouseName string  `json:"warehouse_name"` // 默认仓库名称
	Purpose      int      `json:"purpose"` // 盘点目的
	PurposeName  string   `json:"purpose_name"` // 盘点目的名称
	Type         int      `json:"type"` // 盘点类型
	TypeName     string   `json:"type_name"` // 盘点类型名称
	MaterialCount int     `json:"material_count"` // 物品数量
	Remark       string   `json:"remark"` // 备注
	IsDefault    bool     `json:"is_default"` // 是否默认模板
	Status       int      `json:"status"` // 状态
	CreatedAt    int64    `json:"created_at"` // 创建时间
	UpdatedAt    int64    `json:"updated_at"` // 更新时间
	Items        []*StockReconciliationTemplateItemDetail `json:"items"` // 物品明细
}

type StockReconciliationTemplateItemDetail struct {
	Uuid         uint64 `json:"uuid"` // 明细UUID
	MaterialUuid uint64 `json:"material_uuid"` // 物品UUID
	MaterialCode string `json:"material_code"` // 物品编码
	MaterialName string `json:"material_name"` // 物品名称
	SortOrder    int    `json:"sort_order"` // 排序顺序
}
```

#### 4.1.4 查询模板详情

**请求**：
```go
type StockReconciliationTemplateDetailReq struct {
	Uuid uint64 `json:"uuid" form:"uuid" binding:"required"` // 模板UUID
}
```

**响应**：
```go
type StockReconciliationTemplateDetailResp struct {
	StockReconciliationTemplateDetail
}
```

#### 4.1.5 删除模板

**请求**：
```go
type StockReconciliationTemplateDeleteReq struct {
	Uuid uint64 `json:"uuid" binding:"required"` // 模板UUID
}
```

### 4.2 应用模板创建盘点单

#### 4.2.1 修改盘点单创建请求

**修改**：
```go
// 文件：main/app/dto/req/stock_reconciliation.go
type StockReconciliationSaveReq struct {
	Uuid            uint64                        `json:"uuid"`              // 盘点单UUID，如果为0，表示新建
	TemplateUuid    uint64                        `json:"template_uuid"`    // 模板UUID（可选，如果指定则应用模板）
	IsSubmit        bool                          `json:"is_submit"`         // 是否提交
	SubmitAfterSave bool                          `json:"submit_after_save"` // 是否在保存后提交
	WarehouseUuid   uint64                        `json:"warehouse_uuid"`    // 仓库UUID（如果指定模板，可覆盖模板的仓库）
	Purpose         int                           `json:"purpose"`           // 盘点目的（如果指定模板，可覆盖模板的目的）
	Type            int                           `json:"type"`              // 盘点类型（如果指定模板，可覆盖模板的类型）
	Items           []*StockReconciliationItemReq `json:"items"`             // 盘点单物品明细（如果指定模板，可覆盖模板的物品）
}
```

---

## 五、Service 层实现

### 5.1 模板 Service

```go
// 文件：main/app/service/stock_reconciliation_template.go
package service

import (
	"context"
	"errors"
	"ttpos-server-go/main/app/dto/req"
	"ttpos-server-go/main/app/dto/resp"
	"ttpos-server-go/main/app/model"
	"ttpos-server-go/main/app/repository"
	
	"github.com/shopspring/decimal"
)

type stockReconciliationTemplateSrv struct{}

var StockReconciliationTemplate = new(stockReconciliationTemplateSrv)

// CreateTemplate 创建模板
func (s *stockReconciliationTemplateSrv) CreateTemplate(ctx context.Context, createReq *req.StockReconciliationTemplateCreateReq) (*resp.StockReconciliationTemplateCreateResp, error) {
	companyUuid := ctx.GetCompanyUuid()
	
	// 开启事务
	var templateUuid uint64
	err := repository.DB.Transaction(func(tx *gorm.DB) error {
		templateRepo := repository.NewStockReconciliationTemplateRepo(tx)
		
		// 如果设置为默认模板，先取消其他默认模板
		if createReq.IsDefault {
			if err := templateRepo.CancelDefaultByCompanyUuid(companyUuid); err != nil {
				return errors.WithMessage(errors.New("取消其他默认模板失败"), err.Error())
			}
		}
		
		// 创建模板
		template := &model.StockReconciliationTemplate{
			CompanyUuid:  companyUuid,
			TemplateName: createReq.TemplateName,
			TemplateType: createReq.TemplateType,
			WarehouseUuid: createReq.WarehouseUuid,
			Purpose:      createReq.Purpose,
			Type:         createReq.Type,
			Remark:       createReq.Remark,
			IsDefault:    func() int {
				if createReq.IsDefault {
					return 1
				}
				return 0
			}(),
			Status: 1,
		}
		
		if err := templateRepo.Create(template); err != nil {
			return errors.WithMessage(errors.New("创建模板失败"), err.Error())
		}
		
		templateUuid = template.Uuid
		
		// 创建模板物品明细
		if len(createReq.MaterialUuids) > 0 {
			templateItemRepo := repository.NewStockReconciliationTemplateItemRepo(tx)
			for i, materialUuid := range createReq.MaterialUuids {
				item := &model.StockReconciliationTemplateItem{
					TemplateUuid: templateUuid,
					MaterialUuid: materialUuid,
					SortOrder:    i + 1,
				}
				if err := templateItemRepo.Create(item); err != nil {
					return errors.WithMessage(errors.New("创建模板物品明细失败"), err.Error())
				}
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return &resp.StockReconciliationTemplateCreateResp{
		Uuid: templateUuid,
	}, nil
}

// GetTemplateList 获取模板列表
func (s *stockReconciliationTemplateSrv) GetTemplateList(ctx context.Context, listReq *req.StockReconciliationTemplateListReq) (*resp.StockReconciliationTemplateListResp, error) {
	companyUuid := ctx.GetCompanyUuid()
	templateRepo := repository.NewStockReconciliationTemplateRepo(repository.DB)
	
	// 构建查询条件
	query := templateRepo.Query().
		Where("company_uuid = ?", companyUuid).
		Where("deleted_at = 0")
	
	if listReq.TemplateType > 0 {
		query = query.Where("template_type = ?", listReq.TemplateType)
	}
	if listReq.WarehouseUuid > 0 {
		query = query.Where("warehouse_uuid = ?", listReq.WarehouseUuid)
	}
	if listReq.Status >= 0 {
		query = query.Where("status = ?", listReq.Status)
	}
	if listReq.Keyword != "" {
		query = query.Where("template_name LIKE ?", "%"+listReq.Keyword+"%")
	}
	
	// 查询总数
	total, err := query.Count()
	if err != nil {
		return nil, errors.WithMessage(errors.New("查询模板总数失败"), err.Error())
	}
	
	// 查询列表
	var templates []*model.StockReconciliationTemplate
	offset := (listReq.PageNo - 1) * listReq.PageSize
	err = query.
		Preload("Warehouse").
		Preload("TemplateItems.Material").
		Order("created_at DESC").
		Offset(offset).
		Limit(listReq.PageSize).
		Find(&templates).Error
	
	if err != nil {
		return nil, errors.WithMessage(errors.New("查询模板列表失败"), err.Error())
	}
	
	// 转换为响应格式
	list := make([]*resp.StockReconciliationTemplateDetail, 0, len(templates))
	for _, template := range templates {
		detail := s.convertToDetail(template)
		list = append(list, detail)
	}
	
	return &resp.StockReconciliationTemplateListResp{
		List:     list,
		Total:    total,
		PageNo:   listReq.PageNo,
		PageSize: listReq.PageSize,
	}, nil
}

// GetTemplateByUuid 根据UUID获取模板
func (s *stockReconciliationTemplateSrv) GetTemplateByUuid(ctx context.Context, uuid uint64) (*model.StockReconciliationTemplate, error) {
	companyUuid := ctx.GetCompanyUuid()
	templateRepo := repository.NewStockReconciliationTemplateRepo(repository.DB)
	
	template, err := templateRepo.GetByUuidAndCompanyUuid(uuid, companyUuid)
	if err != nil {
		return nil, errors.WithMessage(errors.New("查询模板失败"), err.Error())
	}
	
	// 预加载关联数据
	if err := repository.DB.
		Preload("Warehouse").
		Preload("TemplateItems.Material").
		First(template, "uuid = ?", uuid).Error; err != nil {
		return nil, errors.WithMessage(errors.New("查询模板详情失败"), err.Error())
	}
	
	return template, nil
}

// ApplyTemplateToStockReconciliation 应用模板到盘点单
// 返回：仓库UUID、盘点目的、盘点类型、物品列表
func (s *stockReconciliationTemplateSrv) ApplyTemplateToStockReconciliation(ctx context.Context, templateUuid uint64) (
	warehouseUuid uint64,
	purpose int,
	reconciliationType int,
	items []*req.StockReconciliationItemReq,
	err error,
) {
	template, err := s.GetTemplateByUuid(ctx, templateUuid)
	if err != nil {
		return 0, 0, 0, nil, err
	}
	
	if template.Status == 0 {
		return 0, 0, 0, nil, errors.New("模板已禁用")
	}
	
	warehouseUuid = template.WarehouseUuid
	purpose = template.Purpose
	reconciliationType = template.Type
	
	// 构建物品列表
	items = make([]*req.StockReconciliationItemReq, 0, len(template.TemplateItems))
	for _, templateItem := range template.TemplateItems {
		if templateItem.Material == nil {
			continue
		}
		
		// 获取物品的默认单位
		material := templateItem.Material
		var defaultUnitUuid uint64
		if len(material.NotBaseUnitList) > 0 {
			// 使用第一个非基准单位作为默认单位
			defaultUnitUuid = material.NotBaseUnitList[0].Uuid
		} else {
			// 如果没有非基准单位，使用基准单位
			defaultUnitUuid = material.BaseUnitUuid
		}
		
		item := &req.StockReconciliationItemReq{
			MaterialUuid: material.Uuid,
			Units: []*req.StockReconciliationItemUnitReq{
				{
					MaterialUnitUuid: defaultUnitUuid,
					Quantity:         nil, // 数量为空，由用户输入
				},
			},
		}
		items = append(items, item)
	}
	
	return warehouseUuid, purpose, reconciliationType, items, nil
}

// convertToDetail 转换为详情格式
func (s *stockReconciliationTemplateSrv) convertToDetail(template *model.StockReconciliationTemplate) *resp.StockReconciliationTemplateDetail {
	detail := &resp.StockReconciliationTemplateDetail{
		Uuid:         template.Uuid,
		TemplateName: template.TemplateName,
		TemplateType: template.TemplateType,
		TemplateTypeName: s.getTemplateTypeName(template.TemplateType),
		WarehouseUuid: template.WarehouseUuid,
		WarehouseName: func() string {
			if template.Warehouse != nil {
				return template.Warehouse.Name
			}
			return ""
		}(),
		Purpose:     template.Purpose,
		PurposeName: s.getPurposeName(template.Purpose),
		Type:        template.Type,
		TypeName:    s.getTypeName(template.Type),
		MaterialCount: len(template.TemplateItems),
		Remark:      template.Remark,
		IsDefault:   template.IsDefault == 1,
		Status:      template.Status,
		CreatedAt:   template.CreatedAt,
		UpdatedAt:   template.UpdatedAt,
	}
	
	// 转换物品明细
	detail.Items = make([]*resp.StockReconciliationTemplateItemDetail, 0, len(template.TemplateItems))
	for _, templateItem := range template.TemplateItems {
		if templateItem.Material == nil {
			continue
		}
		itemDetail := &resp.StockReconciliationTemplateItemDetail{
			Uuid:         templateItem.Uuid,
			MaterialUuid: templateItem.MaterialUuid,
			MaterialCode: templateItem.Material.Code,
			MaterialName: templateItem.Material.Name,
			SortOrder:    templateItem.SortOrder,
		}
		detail.Items = append(detail.Items, itemDetail)
	}
	
	return detail
}

// getTemplateTypeName 获取模板类型名称
func (s *stockReconciliationTemplateSrv) getTemplateTypeName(templateType int) string {
	switch templateType {
	case constant.StockReconciliationTemplateTypeDaily:
		return "日盘"
	case constant.StockReconciliationTemplateTypeWeekly:
		return "周盘"
	case constant.StockReconciliationTemplateTypeMonthly:
		return "月盘"
	default:
		return "未知"
	}
}

// getPurposeName 获取盘点目的名称
func (s *stockReconciliationTemplateSrv) getPurposeName(purpose int) string {
	switch purpose {
	case constant.StockReconciliationPurposeInventory:
		return "库存盘点"
	case constant.StockReconciliationPurposeInitial:
		return "期初盘点"
	default:
		return "未知"
	}
}

// getTypeName 获取盘点类型名称
func (s *stockReconciliationTemplateSrv) getTypeName(reconciliationType int) string {
	switch reconciliationType {
	case constant.StockReconciliationTypeSpecified:
		return "指定物品盘点"
	case constant.StockReconciliationTypeAll:
		return "全部物品盘点"
	default:
		return "未知"
	}
}
```

### 5.2 修改盘点单 Service

```go
// 文件：main/app/service/stock_reconciliation.go
// 在 SaveStockReconciliation 方法中添加模板应用逻辑

func (s *stockReconciliationSrv) SaveStockReconciliation(ctx context.Context, saveReq *req.StockReconciliationSaveReq) (uint64, error) {
	// ... 现有代码 ...
	
	// 如果指定了模板，应用模板
	if saveReq.TemplateUuid > 0 {
		templateWarehouseUuid, templatePurpose, templateType, templateItems, err := 
			service.StockReconciliationTemplate.ApplyTemplateToStockReconciliation(ctx, saveReq.TemplateUuid)
		if err != nil {
			return 0, errors.WithMessage(errors.New("应用模板失败"), err.Error())
		}
		
		// 使用模板值（如果请求中没有指定）
		if saveReq.WarehouseUuid == 0 {
			saveReq.WarehouseUuid = templateWarehouseUuid
		}
		if saveReq.Purpose == 0 {
			saveReq.Purpose = templatePurpose
		}
		if saveReq.Type == 0 {
			saveReq.Type = templateType
		}
		if len(saveReq.Items) == 0 {
			saveReq.Items = templateItems
		}
	}
	
	// ... 继续现有逻辑 ...
}
```

---

## 六、Repository 层实现

### 6.1 StockReconciliationTemplateRepo

```go
// 文件：main/app/repository/stock_reconciliation_template_repo.go
package repository

import (
	"ttpos-server-go/main/app/model"
	"gorm.io/gorm"
)

type stockReconciliationTemplateRepo struct {
	db *gorm.DB
}

func NewStockReconciliationTemplateRepo(db *gorm.DB) *stockReconciliationTemplateRepo {
	return &stockReconciliationTemplateRepo{db: db}
}

func (r *stockReconciliationTemplateRepo) Create(template *model.StockReconciliationTemplate) error {
	return r.db.Create(template).Error
}

func (r *stockReconciliationTemplateRepo) GetByUuidAndCompanyUuid(uuid, companyUuid uint64) (*model.StockReconciliationTemplate, error) {
	var template model.StockReconciliationTemplate
	err := r.db.Where("uuid = ? AND company_uuid = ? AND deleted_at = 0", uuid, companyUuid).First(&template).Error
	return &template, err
}

func (r *stockReconciliationTemplateRepo) CancelDefaultByCompanyUuid(companyUuid uint64) error {
	return r.db.Model(&model.StockReconciliationTemplate{}).
		Where("company_uuid = ? AND is_default = 1", companyUuid).
		Update("is_default", 0).Error
}

func (r *stockReconciliationTemplateRepo) Query() *gorm.DB {
	return r.db.Model(&model.StockReconciliationTemplate{})
}
```

### 6.2 StockReconciliationTemplateItemRepo

```go
// 文件：main/app/repository/stock_reconciliation_template_item_repo.go
package repository

import (
	"ttpos-server-go/main/app/model"
	"gorm.io/gorm"
)

type stockReconciliationTemplateItemRepo struct {
	db *gorm.DB
}

func NewStockReconciliationTemplateItemRepo(db *gorm.DB) *stockReconciliationTemplateItemRepo {
	return &stockReconciliationTemplateItemRepo{db: db}
}

func (r *stockReconciliationTemplateItemRepo) Create(item *model.StockReconciliationTemplateItem) error {
	return r.db.Create(item).Error
}

func (r *stockReconciliationTemplateItemRepo) DeleteByTemplateUuid(templateUuid uint64) error {
	return r.db.Where("template_uuid = ?", templateUuid).
		Delete(&model.StockReconciliationTemplateItem{}).Error
}
```

---

## 七、使用流程

### 7.1 创建模板

1. **前端调用创建模板接口**
   ```
   POST /api/stock-reconciliation-template/create
   {
     "template_name": "日盘模板-主仓库",
     "template_type": 1,
     "warehouse_uuid": 123,
     "purpose": 1,
     "type": 1,
     "material_uuids": [456, 789, 101112],
     "is_default": true
   }
   ```

2. **系统创建模板和物品明细**

### 7.2 应用模板创建盘点单

1. **前端调用创建盘点单接口，指定模板UUID**
   ```
   POST /api/stock-reconciliation/save
   {
     "template_uuid": 123,
     "warehouse_uuid": 123,  // 可选，覆盖模板的仓库
     "items": []  // 可选，覆盖模板的物品
   }
   ```

2. **系统应用模板**
   - 加载模板配置
   - 填充仓库、目的、类型
   - 填充物品列表（每个物品使用默认单位）
   - 创建盘点单

3. **用户修改物品数量**
   - 前端显示模板填充的物品列表
   - 用户输入实盘数量
   - 保存盘点单

---

## 八、总结

### 8.1 实现要点

- ✅ **数据库设计**：模板表和模板物品明细表
- ✅ **Model 定义**：完整的模型和关联关系
- ✅ **API 接口**：模板的增删改查接口
- ✅ **Service 层**：模板管理和应用逻辑
- ✅ **Repository 层**：数据访问层

### 8.2 功能特点

- ✅ 支持日盘、周盘、月盘三种模板类型
- ✅ 模板可以配置多个物品
- ✅ 创建盘点单时快速应用模板
- ✅ 应用模板后仍可修改配置和物品

### 8.3 后续优化

- 🔄 支持模板复制
- 🔄 支持模板导入导出
- 🔄 支持模板版本管理
- 🔄 支持模板使用统计

---

**文档版本**：v1.0  
**创建时间**：2025-01-16  
**维护者**：TTPOS Team
















