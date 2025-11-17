# Supplier Service (供应商服务) 详细说明

## 概述

`supplier.go` 文件实现了供应商管理服务，负责处理供应商的完整生命周期管理，包括创建、查询、更新、删除以及与 ERP 系统的同步。该服务支持总部/分店模式，能够区分内部供应商（总部）和外部供应商，并管理供应商与物品的关联关系。

## 文件位置
```
ttpos-server-go/main/app/service/supplier.go
```

## 核心功能

### 1. 服务接口定义

#### ISupplierSrv 接口
定义了供应商服务的所有方法：

```go
type ISupplierSrv interface {
    GetSupplierList(ctx, req)      // 供应商列表（分页、过滤）
    CreateSupplier(ctx, req)        // 创建供应商（同步到ERP）
    UpdateSupplier(ctx, req)        // 更新供应商（同步到ERP）
    DeleteSupplier(ctx, req)        // 删除供应商（软删除+ERP标记）
    GetSupplierSelect(ctx, req)     // 获取供应商选择器列表
    GetSupplier(ctx, req)           // 获取单个供应商详情
    CheckNameExists(ctx, req)       // 检查名称是否存在
    CheckCodeExists(ctx, req)       // 检查编码是否存在
    SyncSupplier(ctx)               // 同步ERP供应商数据
}
```

### 2. 供应商列表查询 (`GetSupplierList`)

#### 功能描述
获取供应商列表，支持复杂过滤、排序和分页。

#### 请求参数 (`req.SupplierListReq`)
- `PageNo`: 页码
- `PageSize`: 每页数量
- `Keyword`: 关键词搜索（名称或编码）

#### 核心逻辑

**过滤规则：**
1. **排除特定数据**：排除 `is_internal_supplier=1` 且 `headquarter_uuid=0` 的数据
2. **关键词搜索**：按名称或编码模糊匹配
3. **总部特殊处理**：如果当前是总部，过滤掉 ERP 编码为 "总部-供应商" 的记录
4. **排序**：按创建时间倒序

**代码示例：**
```go
// 构建查询选项
opts := []repository.DBOption{
    supplierRepo.WhereExcludeInternalSupplierWithoutHeadquarter(),
    supplierRepo.WhereNameOrCodeLike(req.Keyword),
    supplierRepo.OrderByCreateTime(true),
}

// 总部过滤
if companySetting.IsHeadquarter() {
    opts = append(opts, supplierRepo.WhereErpCodeNot(constant.ErpHeadquartersSupplierCode))
}
```

#### 响应数据处理
- **IsEditable**: 判断供应商是否可编辑（基于 `HeadquarterUuid`）
- **IsHeadquarter**: 标识是否为总部供应商（ErpCode == 总部供应商编码）

#### 返回结构 (`resp.SupplierListResp`)
```go
{
    List: []*SupplierInfo{
        Uuid, Name, Code,
        IsEditable,      // 是否可编辑
        IsHeadquarter,   // 是否总部供应商
        // ... 其他字段
    },
    Meta: {PageNo, PageSize, Total}
}
```

### 3. 获取供应商详情 (`GetSupplier`)

#### 功能描述
根据 UUID 获取单个供应商的完整信息。

#### 请求参数 (`req.SupplierReq`)
- `Uuid`: 供应商 UUID

#### 返回数据
```go
{
    SupplierInfo: {Uuid, Name, Code, IsHeadquarter, IsEditable},
    Address,                         // 地址
    ContactName,                     // 联系人
    ContactPhone,                    // 联系电话
    Status,                          // 状态（0-禁用，1-启用）
    HasRelatedPurchaseOrder,         // 是否有关联的采购订单
}
```

#### 关联检查
- **hasRelatedPurchaseOrder**: 检查供应商是否有关联的采购订单（通过 `PurchaseOrders` 关联）

### 4. 创建供应商 (`CreateSupplier`)

#### 功能描述
创建新供应商，并同步到 ERP 系统（如果启用）。

#### 请求参数 (`req.SupplierCreateReq`)
- `Name`: 供应商名称
- `Code`: 供应商编码（自动转大写）
- `Address`: 地址
- `ContactName`: 联系人
- `ContactPhone`: 联系电话
- `Status`: 状态（0-禁用，1-启用）

#### 核心流程

**1. 编码预处理**
```go
createSupplierReq.Code = strings.ToUpper(createSupplierReq.Code)  // 转大写
```

**2. 编码唯一性检查**
```go
codeExists, err := supplierRepo.IsCodeExists(createSupplierReq.Code, 0)
if codeExists {
    return errors.New("供应商编码已存在")
}
```

**3. ERP 集成（如果启用）**
```go
if ctx.GetCompany().IsOpenErp() {
    erpCode, err = erp.NewIErpSrv(s.dbm).CreateSupplier(ctx.GetContext(), req.CreateSupplierReq{
        SiteCode:     companySetting.ErpnextSiteCode,
        SupplierName: createSupplierReq.Name,
        CompanyAbbr:  companySetting.ErpnextCompanyAbbr,
        Branch:       companySetting.ErpnextBranchName,
        Disabled:     createSupplierReq.Status == 0,
    })
}
```

**4. 数据库创建**
```go
err = supplierRepo.Create(&model.Supplier{
    Name, Code, Address, ContactName, ContactPhone, Status,
    ErpCode: erpCode,  // ERP返回的编码
})
```

### 5. 更新供应商 (`UpdateSupplier`)

#### 功能描述
更新供应商信息，并同步到 ERP 系统。

#### 请求参数 (`req.SupplierUpdateReq`)
- `Uuid`: 供应商 UUID
- `Name`: 供应商名称
- `Code`: 供应商编码（自动转大写）
- `Address`: 地址
- `ContactName`: 联系人
- `ContactPhone`: 联系电话
- `Status`: 状态

#### 核心流程

**1. 存在性检查**
```go
supplier, err := supplierRepo.GetByUuid(updateSupplierReq.Uuid)
if err == gorm.ErrRecordNotFound {
    return errors.New("供应商不存在")
}
```

**2. 可编辑性检查**
```go
if !isEditable(ctx, supplier.HeadquarterUuid) {
    return errors.New("供应商不可编辑")
}
```

**3. 编码唯一性检查（排除自己）**
```go
codeExists, err := supplierRepo.IsCodeExists(updateSupplierReq.Code, updateSupplierReq.Uuid)
if codeExists {
    return errors.New("供应商编码已存在")
}
```

**4. ERP 同步（仅同步自己创建的供应商）**
```go
if ctx.GetCompany().IsOpenErp() && supplier.ErpCode != "" {
    err = erp.NewIErpSrv(s.dbm).UpdateSupplier(ctx.GetContext(), req.UpdateSupplierReq{
        CreateSupplierReq: req.CreateSupplierReq{
            SiteCode, CompanyAbbr, Branch,
            Disabled: updateSupplierReq.Status == 0,
        },
        Name:        supplier.ErpCode,  // ERP中的名称
        CompanyUuid: ctx.GetCompanyUuid(),
    })
}
```

**5. 数据库更新**
```go
err = supplierRepo.Update(supplier.Uuid, map[string]any{
    "code", "name", "address", "contact_name", "contact_phone", "status",
})
```

### 6. 删除供应商 (`DeleteSupplier`)

#### 功能描述
软删除供应商，并在 ERP 中标记为禁用。

#### 请求参数 (`req.SupplierDeleteReq`)
- `Uuid`: 供应商 UUID

#### 核心流程

**1. 存在性检查**
```go
supplier, err := supplierRepo.GetByUuid(deleteSupplierReq.Uuid)
if err == gorm.ErrRecordNotFound {
    return errors.New("供应商不存在")
}
```

**2. 可编辑性检查**
```go
if !isEditable(ctx, supplier.HeadquarterUuid) {
    return errors.New("供应商不可删除")
}
```

**3. 关联检查**
```go
if s.hasRelatedPurchaseOrder(ctx, supplier) {
    return errors.New("该供应商存在关联的采购订单，无法删除")
}
```

**4. ERP 处理（标记禁用）**
```go
if ctx.GetCompany().IsOpenErp() && supplier.ErpCode != "" {
    err = erp.NewIErpSrv(s.dbm).DeleteSupplier(ctx.GetContext(), req.DeleteSupplierReq{
        SiteCode: companySetting.ErpnextSiteCode,
        Name:     supplier.ErpCode,
    })
    // 忽略 "not found" 错误
    if err != nil && !strings.Contains(err.Error(), "not found") {
        return errors.WithMessage(errors.New("删除供应商失败"), err.Error())
    }
}
```

**5. 软删除**
```go
err = supplierRepo.Delete(deleteSupplierReq.Uuid)  // 设置 delete_time
```

### 7. 获取供应商选择器列表 (`GetSupplierSelect`)

#### 功能描述
获取供应商选择器列表，用于下拉选择，支持根据采购类型过滤。

#### 请求参数 (`req.SupplierSelectReq`)
- `PurchaseType`: 采购类型（0-内部采购，1-外部采购）

#### 版本兼容处理

**旧版本逻辑（< 2.6.0）**
如果 ERP 已启用且版本 < 2.6.0，直接从 ERP 获取供应商列表：
```go
if ctx.GetCompany().IsOpenErp() && ctx.Version(context.LT, "2.6.0") {
    erpResp, err := erp.NewIErpSrv(s.dbm).GetSupplierList(ctx)
    // 转换为简化格式
    for _, supplier := range erpResp.SupplierList {
        supplierList = append(supplierList, &resp.SupplierSimpleInfo{
            Name: supplier.SupplierName,
            Code: supplier.Name,  // ERP编码
        })
    }
}
```

**新版本逻辑（>= 2.6.0）**

1. **查询选项**：
```go
opts := []repository.DBOption{
    supplierRepo.OrderByName(false),  // 按名称升序
    supplierRepo.WhereNotDeleted(),
    supplierRepo.WhereStatus(1),      // 仅启用的
}

// 如果启用ERP，仅查询有ERP编码的供应商
if ctx.GetCompany().IsOpenErp() {
    opts = append(opts, supplierRepo.WhereErpCodeNot(""))
}
```

2. **采购类型过滤**：
```go
for _, supplier := range suppliers {
    if req.PurchaseType == 1 {  // 外部采购
        // 去掉总部供应商
        if supplier.ErpCode == constant.ErpHeadquartersSupplierCode {
            continue
        }
    } else {  // 内部采购（从总部）
        // 仅保留总部供应商
        if supplier.ErpCode != constant.ErpHeadquartersSupplierCode {
            continue
        }
    }
    supplierList = append(supplierList, &resp.SupplierSimpleInfo{
        Uuid, Name, Code: supplier.ErpCode,
    })
}
```

#### 返回结构 (`resp.SupplierSelectResp`)
```go
{
    List: []*SupplierSimpleInfo{
        Uuid,  // 供应商UUID
        Name,  // 供应商名称
        Code,  // ERP编码
    }
}
```

### 8. 名称/编码检查 (`CheckNameExists` / `CheckCodeExists`)

#### CheckNameExists
检查供应商名称是否已存在（创建或更新时）。

**请求参数**：
- `Name`: 名称
- `Uuid`: 供应商UUID（更新时传入，用于排除自己）

#### CheckCodeExists
检查供应商编码是否已存在。

**请求参数**：
- `Code`: 编码
- `Uuid`: 供应商UUID（更新时传入，用于排除自己）

#### 返回结构
```go
{
    Exists: true/false  // 是否存在
}
```

### 9. 同步 ERP 供应商 (`SyncSupplier`)

#### 功能描述
从 ERP 系统同步供应商数据到本地数据库，包括供应商基本信息和供应商-物品关联关系。这是一个复杂的双向同步过程。

#### 核心流程

**1. 前置检查**
```go
if !company.IsOpenErp() {
    return errors.New("公司未开启erp")
}
```

**2. 获取 ERP 供应商列表**
```go
erpSuppliers, err := erp.NewIErpSrv(s.dbm).ListSuppliers(ctx, req.GetErpnextSupplierListReq{
    SiteCode:    companySetting.ErpnextSiteCode,
    CompanyAbbr: companySetting.ErpnextCompanyAbbr,
    Branch:      companySetting.ErpnextBranchName,
})

var supplierErpCodes []string
for _, erpSupplier := range erpSuppliers {
    supplierErpCodes = append(supplierErpCodes, erpSupplier.Name)
}
```

**3. 处理总部供应商（分店场景）**
如果当前是分店，获取总部的供应商数据：
```go
var headquarterSuppliers []model.Supplier
if companySetting.IsSubShop() {
    // 获取总部公司信息
    s.dbm.GetDB(constant.DefaultDB).Model(&model.CompanySetting{}).
        Where("uuid = ?", companySetting.HeadquarterUuid).
        First(&headquarter)
    
    // 获取总部供应商
    s.dbm.GetDB(headquarter.Uuid).Model(&model.Supplier{}).
        Find(&headquarterSuppliers)
}
```

**4. 获取本店供应商并识别待删除项**
```go
var suppliers []model.Supplier
var deletingSupplierUuids []uint64
supplierMap := make(map[string]model.Supplier)

db.Model(&model.Supplier{}).
    Scopes(repository.ExcludeHeadquarter).  // 排除总部供应商
    Where("erp_code != ''").
    Find(&suppliers)

for _, supplier := range suppliers {
    if !slices.Contains(supplierErpCodes, supplier.ErpCode) {
        deletingSupplierUuids = append(deletingSupplierUuids, supplier.Uuid)
    }
    supplierMap[supplier.ErpCode] = supplier
}
```

**5. 事务处理：同步供应商数据**

```go
err = db.Transaction(func(tx *gorm.DB) error {
    // 5.1 删除ERP中已不存在的供应商
    if len(deletingSupplierUuids) > 0 {
        tx.Model(&model.Supplier{}).
            Where("uuid IN (?)", deletingSupplierUuids).
            Update("delete_time", time.Now().Unix())
    }
    
    var insertingHeadquarterSuppliers []model.Supplier
    
    // 5.2 遍历ERP供应商，创建或更新
    for _, erpSupplier := range erpSuppliers {
        // 处理名称（优先使用别名）
        name := erpSupplier.AliasName
        if name == "" {
            name = erpSupplier.SupplierName
        }
        
        // 获取现有供应商数据
        supplier := supplierMap[erpSupplier.Name]
        address := supplier.Address
        contactName := supplier.ContactName
        contactPhone := supplier.ContactPhone
        code := supplier.Code
        
        // 特殊处理：总部供应商固定编码
        if erpSupplier.Name == constant.ErpHeadquartersSupplierCode {
            code = constant.HeadquartersSupplierCode  // HSP00001
        }
        
        // 状态转换
        status := 1
        if erpSupplier.Disabled {
            status = 0
        }
        
        // 是否内部供应商
        isInternalSupplier := 0
        if erpSupplier.IsInternalSupplier {
            isInternalSupplier = 1
        }
        
        if supplier.Uuid == 0 {  // 新建供应商
            insertingHeadquarterSuppliers = append(insertingHeadquarterSuppliers, 
                model.Supplier{
                    Name, Address, ContactName, ContactPhone,
                    ErpCode:            erpSupplier.Name,
                    Status:             status,
                    Code:               code,
                    RepresentsCompany:  erpSupplier.RepresentsCompany,
                    IsInternalSupplier: isInternalSupplier,
                })
        } else {  // 更新供应商
            tx.Model(&model.Supplier{}).
                Where("uuid = ?", supplier.Uuid).
                Updates(map[string]any{
                    "name", "code", "status", "address", 
                    "contact_name", "contact_phone",
                    "represents_company", "is_internal_supplier",
                    "delete_time": 0,  // 恢复为未删除
                })
        }
    }
    
    // 5.3 处理总部供应商（分店场景）
    if len(headquarterSuppliers) > 0 {
        // 删除旧的总部供应商
        tx.Where("headquarter_uuid > 0").Delete(&model.Supplier{})
        
        // 复制总部供应商到分店
        for _, hqSupplier := range headquarterSuppliers {
            insertingHeadquarterSuppliers = append(insertingHeadquarterSuppliers,
                model.Supplier{
                    BaseModel:          hqSupplier.BaseModel,  // 保留UUID等
                    Name, Code, Status, Address, ContactName, ContactPhone,
                    ErpCode, RepresentsCompany, IsInternalSupplier,
                    HeadquarterUuid:    headquarter.Uuid,  // 标记来源
                })
        }
    }
    
    // 5.4 批量插入新供应商
    if len(insertingHeadquarterSuppliers) > 0 {
        tx.Model(&model.Supplier{}).Create(&insertingHeadquarterSuppliers)
    }
    
    return nil
})
```

**6. 同步供应商-物品关联关系**

```go
erpSrv := erp.NewIErpSrv(s.dbm)
supplierList, err := supplierRepo.GetList(supplierRepo.WhereNotDeleted())

for _, supplier := range supplierList {
    // 6.1 获取供应商物品列表
    erpSupplierItems, err := erpSrv.GetSupplierItemList(ctx, 
        req.GetErpnextSupplierItemListReq{
            SiteCode: companySetting.ErpnextSiteCode,
            Supplier: supplier.ErpCode,
        })
    
    // 6.2 删除旧的关联关系
    materialSupplierRepo.DeletePermanently(
        materialSupplierRepo.WhereSupplierErpCode(supplier.ErpCode),
    )
    
    // 6.3 添加新的关联关系
    if len(erpSupplierItems) > 0 {
        for _, supplierErp := range erpSupplierItems {
            // 根据物品编码查找物品
            material, err := materialRepo.GetMaterialByCode(supplierErp.ItemCode)
            if err != nil {
                if strings.Contains(err.Error(), "record not found") {
                    continue  // 物品不存在，跳过
                }
                return err
            }
            
            // 创建关联
            materialSupplierRepo.Create(&model.MaterialSupplier{
                MaterialUuid:    material.Uuid,
                MaterialCode:    material.Code,
                SupplierUuid:    supplier.Uuid,
                SupplierErpCode: supplier.ErpCode,
                HeadquarterUuid: supplier.HeadquarterUuid,
            })
        }
    }
}
```

## 辅助函数

### isEditable
判断供应商是否可编辑。

**逻辑**：
- 如果 `HeadquarterUuid` 为 0，表示是本店创建的，可编辑
- 如果 `HeadquarterUuid` > 0，表示是从总部同步的，不可编辑

```go
func isEditable(ctx context.Context, headquarterUuid uint64) bool {
    return headquarterUuid == 0
}
```

## 数据模型

### model.Supplier
供应商模型：
```go
type Supplier struct {
    BaseModel               // Uuid, CreateTime, UpdateTime, DeleteTime
    Name            string  // 供应商名称
    Code            string  // 供应商编码（自定义）
    Address         string  // 地址
    ContactName     string  // 联系人
    ContactPhone    string  // 联系电话
    Status          int     // 状态（0-禁用，1-启用）
    ErpCode         string  // ERP编码
    RepresentsCompany string // 代表公司
    IsInternalSupplier int  // 是否内部供应商（1-是，0-否）
    HeadquarterUuid uint64  // 总部UUID（> 0表示从总部同步）
    
    PurchaseOrders  []PurchaseOrder `gorm:"foreignKey:SupplierUuid"` // 关联采购订单
}
```

### model.MaterialSupplier
供应商-物品关联模型：
```go
type MaterialSupplier struct {
    MaterialUuid    uint64  // 物品UUID
    MaterialCode    string  // 物品编码
    SupplierUuid    uint64  // 供应商UUID
    SupplierErpCode string  // 供应商ERP编码
    HeadquarterUuid uint64  // 总部UUID
}
```

## 依赖关系

### 外部服务依赖
1. **ERP Service** (`app/service/rpc/erp`):
   - `CreateSupplier`: 创建供应商到ERP
   - `UpdateSupplier`: 更新ERP中的供应商
   - `DeleteSupplier`: 删除（禁用）ERP中的供应商
   - `ListSuppliers`: 获取ERP供应商列表
   - `GetSupplierList`: 获取供应商列表（旧版本）
   - `GetSupplierItemList`: 获取供应商物品列表

### Repository 依赖
1. **SupplierRepo** (`repository.NewSupplierRepo`):
   - `GetListWithPagination`: 分页查询
   - `GetList`: 列表查询
   - `GetByUuid`: 根据UUID查询
   - `Create`: 创建
   - `Update`: 更新
   - `Delete`: 软删除
   - `IsNameExists`: 检查名称是否存在
   - `IsCodeExists`: 检查编码是否存在
   - 各种 `Where*` 条件构建器

2. **MaterialRepo** (`repository.NewMaterialRepo`):
   - `GetMaterialByCode`: 根据编码获取物品

3. **MaterialSupplierRepo** (`repository.NewMaterialSupplierRepo`):
   - `Create`: 创建关联
   - `DeletePermanently`: 永久删除

### 数据库依赖
- 使用 `dbm.GetDB(ctx.GetDbId())` 获取当前公司数据库
- 使用 `dbm.GetDB(constant.DefaultDB)` 获取默认数据库（查询总部信息）
- 使用 `dbm.GetDB(headquarter.Uuid)` 获取总部数据库

## 业务规则

### 1. 供应商编码规则
- **自动转大写**：创建和更新时自动将编码转为大写
- **唯一性**：同一公司下编码必须唯一
- **总部供应商固定编码**：总部供应商固定使用 `HSP00001`

### 2. 总部/分店模式

#### 总部逻辑
- 过滤掉 `ErpCode == constant.ErpHeadquartersSupplierCode` 的供应商（自己不显示）
- 可以创建和管理自己的供应商

#### 分店逻辑
- 可以看到总部同步下来的供应商（`HeadquarterUuid > 0`）
- 总部同步的供应商**不可编辑、不可删除**
- 可以创建和管理自己的供应商（`HeadquarterUuid = 0`）

### 3. 采购类型区分

#### 内部采购 (PurchaseType = 0)
- 仅显示总部供应商（`ErpCode == constant.ErpHeadquartersSupplierCode`）
- 用于从总部采购物品

#### 外部采购 (PurchaseType = 1)
- 排除总部供应商
- 显示所有外部供应商（包括自己创建的和ERP同步的）

### 4. ERP 集成规则

#### 创建供应商
- 本地先检查编码唯一性
- 调用ERP创建接口，获取 `ErpCode`
- 保存到本地数据库，记录 `ErpCode`

#### 更新供应商
- 仅更新有 `ErpCode` 的供应商（即自己创建的）
- 同步更新ERP和本地数据

#### 删除供应商
- 本地软删除（设置 `delete_time`）
- ERP中标记为禁用（Disabled）
- 忽略ERP的 "not found" 错误（可能已被ERP删除）

#### 同步供应商
- **定时触发**或**手动触发**
- **双向同步**：
  - ERP → 本地：新增/更新供应商
  - 本地 → ERP：删除本地不存在的供应商
- **总部分店同步**：分店同步总部供应商数据
- **物品关联同步**：同步供应商-物品关联关系

### 5. 内部供应商 (Internal Supplier)

#### 定义
- `IsInternalSupplier = 1` 表示内部供应商
- 通常指总部作为供应商

#### 过滤规则
在列表查询中，排除 `is_internal_supplier=1 && headquarter_uuid=0` 的数据：
```go
supplierRepo.WhereExcludeInternalSupplierWithoutHeadquarter()
```

**含义**：
- `is_internal_supplier=1`：是内部供应商
- `headquarter_uuid=0`：但不是从总部同步的
- **排除原因**：避免显示本地创建的内部供应商（数据不一致）

### 6. 删除限制

#### 关联检查
供应商有关联的采购订单时，不允许删除：
```go
if s.hasRelatedPurchaseOrder(ctx, supplier) {
    return errors.New("该供应商存在关联的采购订单，无法删除")
}
```

#### 可编辑性检查
从总部同步的供应商不允许删除：
```go
if !isEditable(ctx, supplier.HeadquarterUuid) {
    return errors.New("供应商不可删除")
}
```

## 版本兼容性

### 版本 < 2.6.0
在获取供应商选择器列表时，直接从 ERP 获取数据：
```go
if ctx.GetCompany().IsOpenErp() && ctx.Version(context.LT, "2.6.0") {
    // 调用 ERP GetSupplierList 接口
}
```

### 版本 >= 2.6.0
从本地数据库查询供应商，数据已通过 `SyncSupplier` 同步。

## 错误处理

### 常见错误
1. **"供应商编码已存在"**：创建或更新时编码重复
2. **"供应商不存在"**：UUID 不存在
3. **"供应商不可编辑"**：尝试编辑总部同步的供应商
4. **"供应商不可删除"**：尝试删除总部同步的供应商
5. **"该供应商存在关联的采购订单，无法删除"**：有采购订单关联
6. **"公司未开启erp"**：未启用ERP时调用同步
7. **ERP 接口错误**：ERP 服务不可用或数据不一致

### 错误包装
使用 `errors.WithMessage` 包装错误，提供上下文：
```go
return errors.WithMessage(err, "获取供应商列表失败")
```

## 性能考虑

### 1. 分页查询
列表查询使用分页，避免一次加载大量数据：
```go
supplierRepo.GetListWithPagination(pageNo, pageSize, opts...)
```

### 2. 批量操作
同步时使用批量插入：
```go
if len(insertingHeadquarterSuppliers) > 0 {
    tx.Model(&model.Supplier{}).Create(&insertingHeadquarterSuppliers)
}
```

### 3. 索引优化
关键字段应建立索引：
- `erp_code`：频繁用于查询和关联
- `code`：用于唯一性检查
- `name`：用于模糊搜索
- `headquarter_uuid`：用于区分总部/分店数据

### 4. 事务处理
同步操作使用事务，保证数据一致性：
```go
err = db.Transaction(func(tx *gorm.DB) error {
    // 多个数据库操作
    return nil
})
```

## 使用示例

### 1. 创建供应商
```go
supplierSrv := NewSupplierSrv(dbm)
err := supplierSrv.CreateSupplier(ctx, req.SupplierCreateReq{
    Name:         "供应商A",
    Code:         "SUP001",
    Address:      "广州市天河区",
    ContactName:  "张三",
    ContactPhone: "13800138000",
    Status:       1,
})
```

### 2. 获取供应商列表
```go
resp, err := supplierSrv.GetSupplierList(ctx, req.SupplierListReq{
    PageReq: req.PageReq{PageNo: 1, PageSize: 20},
    Keyword: "供应商",
})
```

### 3. 获取供应商选择器（外部采购）
```go
resp, err := supplierSrv.GetSupplierSelect(ctx, req.SupplierSelectReq{
    PurchaseType: 1,  // 外部采购
})
```

### 4. 同步 ERP 供应商
```go
err := supplierSrv.SyncSupplier(ctx)
```

## 数据流图

### 创建供应商流程
```
用户请求
    ↓
编码转大写 + 唯一性检查
    ↓
[ERP启用] → 调用 ERP CreateSupplier → 获取 ErpCode
    ↓
保存到本地数据库（包含ErpCode）
    ↓
返回成功
```

### 同步供应商流程
```
触发同步
    ↓
获取 ERP 供应商列表
    ↓
[分店] → 获取总部供应商列表
    ↓
获取本地供应商列表 + 识别待删除项
    ↓
【事务开始】
    ├─ 软删除 ERP 中已不存在的供应商
    ├─ 遍历 ERP 供应商
    │   ├─ 新供应商 → 加入插入列表
    │   └─ 已存在 → 更新
    ├─ [分店] 删除旧总部供应商 + 插入新总部供应商
    └─ 批量插入新供应商
【事务提交】
    ↓
遍历所有供应商
    ├─ 获取该供应商在 ERP 中的物品列表
    ├─ 删除旧的供应商-物品关联
    └─ 创建新的供应商-物品关联
    ↓
同步完成
```

## 配置依赖

### 公司设置 (CompanySetting)
- `IsHeadquarter()`: 是否总部
- `IsSubShop()`: 是否分店
- `ErpnextSiteCode`: ERP 站点编码
- `ErpnextCompanyAbbr`: ERP 公司简称
- `ErpnextBranchName`: ERP 分支名称
- `HeadquarterUuid`: 总部UUID（分店场景）

### 公司信息 (Company)
- `IsOpenErp()`: 是否启用ERP

### 常量
- `constant.ErpHeadquartersSupplierCode`: 总部供应商ERP编码
- `constant.HeadquartersSupplierCode`: 总部供应商本地编码（HSP00001）
- `constant.DefaultDB`: 默认数据库ID

## 总结

`supplier.go` 实现了一个功能完整、逻辑复杂的供应商管理服务，主要特点包括：

1. **ERP 深度集成**：创建、更新、删除、同步全流程与 ERP 系统打通
2. **总部分店模式**：支持总部和分店的独立管理及数据同步
3. **采购类型区分**：支持内部采购（从总部）和外部采购（从外部供应商）
4. **数据一致性保障**：使用事务、唯一性检查、关联检查确保数据完整性
5. **版本兼容**：处理不同版本的逻辑差异
6. **完善的权限控制**：总部同步的供应商不可编辑删除
7. **复杂的同步逻辑**：双向同步供应商基本信息和物品关联关系

该服务是采购管理模块的核心基础，为采购订单、供应商评估、物品采购等功能提供支持。

