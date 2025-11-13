# Warehouse Service (仓库服务) 详细说明

## 概述

`warehouse.go` 文件实现了仓库管理服务，负责处理仓库的完整生命周期管理，包括创建、查询、更新、删除、默认仓库设置、出入库记录查询以及与 ERP 系统的同步。该服务支持总部/分店模式、普通仓库和在途仓库两种类型，并提供完整的库存管理和对方机构管理功能。

## 文件位置
```
ttpos-server-go/main/app/service/warehouse.go
```

## 核心功能

### 1. 服务接口定义

#### IWarehouseSrv 接口
定义了仓库服务的所有方法：

```go
type IWarehouseSrv interface {
    GetWarehouseList(ctx, req, isHeadquarters...) (resp, error)      // 仓库列表
    GetHeadquarterWarehouseList(ctx) (resp, error)                   // 总部仓库列表
    CreateWarehouse(ctx, req) error                                   // 创建仓库
    UpdateWarehouse(ctx, req) error                                   // 更新仓库
    DeleteWarehouse(ctx, req) error                                   // 删除仓库
    SetDefaultWarehouse(ctx, req) error                               // 设置默认仓库
    GetWarehouse(ctx, req) (resp, error)                              // 获取仓库详情
    GetWarehouseInOutList(ctx, req) (resp, error)                     // 出入库明细列表
    CheckCodeExists(ctx, req) (resp, error)                           // 检查仓库编码是否存在
    GetOtherOrgList(ctx) (resp, error)                                // 对方机构列表
    GetWarehouseMaterialList(ctx, req) (resp, error)                  // 仓库物品列表
    
    SyncWarehouse(ctx) error                                          // 同步仓库
    SyncWarehouseItemStock(ctx) error                                 // 同步仓库物品库存
}
```

### 2. 仓库类型

系统支持两种仓库类型：

| 类型 | 常量值 | 说明 | 特性 |
|------|--------|------|------|
| 普通仓库 | `normal` | 日常使用的仓库 | 可设置为默认仓库 |
| 在途仓库 | `transit` | 调拨/转运中的仓库 | 不可禁用、不可删除、不可设为默认 |

**代码映射**：
```go
const (
    WarehouseTypeNormal  = "normal"
    WarehouseTypeTransit = "transit"
)

// ERP类型映射
typeMap := map[string]string{
    "normal":  "Normal",
    "transit": "Transit",
}
```

### 3. 对方机构类型

出入库记录中涉及的对方机构：

```go
const (
    OtherOrgCodeSupplierErpCode = "supplier-erp-code"  // 供应商（ERP编码）
    OtherOrgCodeCompanyUuid     = "company-uuid"       // 公司（分店UUID）
)
```

**编码格式**：
- 供应商：`supplier-erp-code:{ErpCode}`
- 分店：`company-uuid:{CompanyUuid}`

## 核心流程

### 1. 获取仓库列表 (`GetWarehouseList`)

#### 功能描述
分页查询仓库列表，支持多条件筛选。

#### 请求参数 (`req.WarehouseListReq`)
- `PageNo`: 页码
- `PageSize`: 每页数量
- `Keyword`: 关键词（名称或编码）
- `Type`: 仓库类型（`normal` / `transit`）
- `Status`: 状态（0-禁用，1-启用）

**可选参数**：
- `isHeadquarters`: 是否查询总部仓库

#### 核心流程

**步骤 1：构建查询选项**
```go
var opts []repository.DBOption

// 名称或编码筛选
if req.Keyword != "" {
    opts = append(opts, warehouseRepo.WhereNameOrCodeLike(req.Keyword))
}

// 类型筛选
if req.Type != "" && slice.Contain([]string{"normal", "transit"}, req.Type) {
    opts = append(opts, warehouseRepo.WhereType(req.Type))
}

// 状态筛选
if req.Status != nil && slice.Contain([]int{0, 1}, *req.Status) {
    opts = append(opts, warehouseRepo.WhereStatus(*req.Status))
}

// 总部/分店筛选
if isHeadquarter {
    opts = append(opts, warehouseRepo.WhereIsHeadquarter(isHeadquarter))
} else {
    opts = append(opts, warehouseRepo.WhereHeadquarterUuid(0))  // 仅本店仓库
}

// 按更新时间倒序
opts = append(opts, warehouseRepo.OrderByUpdateTime(true))
```

**步骤 2：分页查询**
```go
warehouses, total, err := warehouseRepo.GetListWithPagination(
    req.PageNo,
    req.PageSize,
    opts...)
```

**步骤 3：构建响应数据**
```go
warehouseList := make([]resp.WarehouseResp, 0, len(warehouses))
for _, warehouse := range warehouses {
    warehouseList = append(warehouseList, s.buildWarehouseResp(ctx, warehouse, isHeadquarter))
}
```

#### 返回结构 (`resp.WarehouseListResp`)
```go
{
    List: []WarehouseResp{
        Uuid,          // 仓库UUID
        LocalName,     // 多语言名称
        Type,          // 类型（normal/transit）
        Code,          // 仓库编码（总部时返回ErpCode）
        Status,        // 状态（0-禁用，1-启用）
        Contact,       // 联系人
        Phone,         // 联系电话
        Address,       // 地址
        IsDefault,     // 是否默认仓库
        IsEditable,    // 是否可编辑
        HasItem,       // 是否有物品
    },
    Meta: {PageNo, PageSize, Total}
}
```

### 2. 获取总部仓库列表 (`GetHeadquarterWarehouseList`)

#### 功能描述
快捷方法，获取总部的所有启用的普通仓库。

#### 实现
```go
return s.GetWarehouseList(ctx, req.WarehouseListReq{
    PageReq: dto.PageReq{
        PageNo:   1,
        PageSize: 999,
    },
    Type:   "normal",
    Status: &[]int{1}[0],
}, true)  // isHeadquarters = true
```

**特点**：
- 仅查询普通仓库
- 仅查询启用状态
- 一次性获取所有（999条）

### 3. 获取仓库详情 (`GetWarehouse`)

#### 功能描述
根据UUID获取单个仓库的详细信息。

#### 请求参数 (`req.WarehouseReq`)
- `Uuid`: 仓库UUID

#### 实现
```go
warehouse, err := warehouseRepo.GetByUuid(req.Uuid)
return s.buildWarehouseResp(ctx, *warehouse, false), nil
```

### 4. 创建仓库 (`CreateWarehouse`)

#### 功能描述
创建新仓库，包括多语言名称创建和ERP同步。

#### 请求参数 (`req.CreateWarehouseReq`)
- `LocaleName`: 多语言名称（9种语言）
- `Code`: 仓库编码（自动转大写）
- `Type`: 类型（`normal` / `transit`）
- `Status`: 状态（0-禁用，1-启用）
- `Contact`: 联系人
- `Phone`: 联系电话
- `Address`: 地址

#### 核心流程

**步骤 1：编码预处理**
```go
addReq.Code = strings.ToUpper(addReq.Code)  // 转大写
```

**步骤 2：编码唯一性检查**
```go
exists, err := warehouseRepo.IsCodeExists(addReq.Code, 0)
if exists {
    return errors.New("仓库编码已存在")
}
```

**步骤 3：名称唯一性和长度检查**
```go
checkService := NewCheckNameSrv(s.dbm)
names := checkService.MakeCheckNameList(ctx, addReq.LocaleName)

// 检查长度
for _, name := range names {
    if !checkService.CheckNameLength(ctx, name.Text, 140) {
        return errors.New("仓库名称长度不能超过140")
    }
}

// 检查唯一性
exists = checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
    Source: constant.CheckNameSourceWarehouse,
    Names:  names,
})
if exists {
    return errors.New("仓库名称已存在")
}
```

**步骤 4：生成UUID**
```go
uuid, err := utils.GetID()
```

**步骤 5：事务处理**

```go
err = db.Transaction(func(tx *gorm.DB) error {
    // 5.1 获取英文名称（用于ERP）
    warehouseName, err := GetEnName(ctx, s.settingSrv, addReq.LocaleName)
    
    // 5.2 创建多语言名称
    multiLanguageName := model.MultiLanguageName{
        ZhName:   addReq.LocaleName.ZH,
        ThName:   addReq.LocaleName.TH,
        EnName:   addReq.LocaleName.EN,
        ZhTwName: addReq.LocaleName.ZHTW,
        JaName:   addReq.LocaleName.JA,
        KoName:   addReq.LocaleName.KO,
        MyName:   addReq.LocaleName.MY,
        TrName:   addReq.LocaleName.TR,
        SvName:   addReq.LocaleName.SV,
    }
    tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName)
    
    warehouse.Name = addReq.LocaleName.ToJson()
    warehouse.MultiLanguageNameUuid = multiLanguageName.Uuid
    
    var erpCode string
    // 5.3 调用ERP接口创建仓库
    if ctx.GetCompany().IsOpenErp() {
        erpCode, err = erp.NewIErpSrv(s.dbm).CreateWarehouse(
            ctx.GetContext(), 
            req.CreateErpnextWarehouseReq{
                SiteCode:      companySetting.ErpnextSiteCode,
                WarehouseName: warehouseName,
                AliasName:     warehouseName,
                CompanyAbbr:   companySetting.ErpnextCompanyAbbr,
                Branch:        companySetting.ErpnextBranchName,
                Disabled:      addReq.Status == 0,
                WarehouseType: typeMap[addReq.Type],  // "Normal" or "Transit"
            })
    }
    warehouse.ErpCode = erpCode
    
    // 5.4 保存到数据库
    tx.Model(&model.Warehouse{}).Create(&warehouse)
    
    return nil
})
```

### 5. 更新仓库 (`UpdateWarehouse`)

#### 功能描述
更新仓库信息，包括多语言名称和ERP同步。

#### 请求参数 (`req.UpdateWarehouseReq`)
- `Uuid`: 仓库UUID
- `LocaleName`: 多语言名称
- `Code`: 仓库编码
- `Type`: 类型
- `Status`: 状态
- `Contact`: 联系人
- `Phone`: 联系电话
- `Address`: 地址

#### 核心流程

**步骤 1：存在性检查**
```go
warehouse, err := warehouseRepo.GetByUuid(updateReq.Uuid)
if err == gorm.ErrRecordNotFound {
    return errors.New("仓库不存在")
}
```

**步骤 2：可编辑性检查**
```go
if !isEditable(ctx, warehouse.HeadquarterUuid) {
    return errors.New("仓库不可编辑")
}
```

**步骤 3：业务规则检查**
```go
// 默认仓库或在途仓库不可禁用
if (warehouse.IsDefault == 1 || warehouse.Type == constant.WarehouseTypeTransit) && 
    updateReq.Status == 0 {
    return errors.New("默认仓库或在途仓库不可禁用")
}
```

**步骤 4：编码和名称唯一性检查**
```go
// 编码检查（排除自己）
exists, err := warehouseRepo.IsCodeExists(updateReq.Code, warehouse.Uuid)
if exists {
    return errors.New("仓库编码已存在")
}

// 名称检查（排除自己）
exists = checkService.InnerCheckNameExists(ctx, req.CheckNameRequest{
    Uuid:   warehouse.Uuid,
    Source: constant.CheckNameSourceCategory,
    Names:  names,
})
if exists {
    return errors.New("仓库名称已存在")
}
```

**步骤 5：事务更新**

```go
err = db.Transaction(func(tx *gorm.DB) error {
    // 5.1 更新多语言名称
    tx.Model(&model.MultiLanguageName{}).
        Where("uuid = ?", warehouse.MultiLanguageNameUuid).
        Updates(map[string]any{
            "zh_name", "th_name", "en_name", "zh_tw_name",
            "ja_name", "ko_name", "my_name", "tr_name", "sv_name",
        })
    
    // 5.2 更新仓库
    tx.Model(&model.Warehouse{}).
        Where("uuid = ?", warehouse.Uuid).
        Updates(updateData)
    
    // 5.3 同步到ERP
    if ctx.GetCompany().IsOpenErp() && warehouse.ErpCode != "" {
        erp.NewIErpSrv(s.dbm).UpdateWarehouse(ctx.GetContext(), 
            req.UpdateErpnextWarehouseReq{
                CreateErpnextWarehouseReq: req.CreateErpnextWarehouseReq{
                    SiteCode, CompanyAbbr, Branch,
                    Disabled:      updateReq.Status == 0,
                    WarehouseType: typeMap[updateReq.Type],
                },
                Name: warehouse.ErpCode,
            })
    }
    
    return nil
})
```

**步骤 6：清理翻译队列**
```go
// 手动更新后，从待翻译集合中删除
s.translateSrv.RemoveMultiLanguageNameUuidFromSet(
    ctx.GetCompanyUuid(), 
    warehouse.MultiLanguageNameUuid)
```

### 6. 删除仓库 (`DeleteWarehouse`)

#### 功能描述
软删除仓库，并在ERP中标记为禁用。

#### 请求参数 (`req.DeleteWarehouseReq`)
- `Uuid`: 仓库UUID

#### 核心流程

**步骤 1：存在性检查**
```go
existingWarehouse, err := warehouseRepo.GetByUuid(deleteWarehouseReq.Uuid)
if err == gorm.ErrRecordNotFound {
    return errors.New("仓库不存在")
}
```

**步骤 2：业务规则检查**
```go
// 默认仓库不可删除
if existingWarehouse.IsDefault == 1 {
    return errors.New("默认仓库不可删除")
}

// 总部同步的仓库不可删除
if !isEditable(ctx, existingWarehouse.HeadquarterUuid) {
    return errors.New("仓库不可删除")
}

// 在途仓库不可删除
if existingWarehouse.Type == constant.WarehouseTypeTransit {
    return errors.New("在途仓库不可删除")
}

// 有物品的仓库不可删除
if existingWarehouse.Items != nil && len(existingWarehouse.Items) > 0 {
    return errors.New("仓库存在关联的物品，不可删除")
}
```

**步骤 3：ERP处理（标记禁用）**
```go
if ctx.GetCompany().IsOpenErp() && existingWarehouse.ErpCode != "" {
    err = erp.NewIErpSrv(s.dbm).DeleteWarehouse(ctx.GetContext(), 
        req.DeleteErpnextWarehouseReq{
            SiteCode: companySetting.ErpnextSiteCode,
            Name:     existingWarehouse.ErpCode,
        })
    // 忽略 "not found" 错误
    if err != nil && !strings.Contains(err.Error(), "not found") {
        return errors.WithMessage(errors.New("删除仓库失败"), err.Error())
    }
}
```

**步骤 4：软删除**
```go
err = warehouseRepo.Delete(existingWarehouse.Uuid)  // 设置 delete_time
```

### 7. 设置默认仓库 (`SetDefaultWarehouse`)

#### 功能描述
将指定仓库设置为默认仓库，其他仓库自动取消默认状态。

#### 请求参数 (`req.SetDefaultWarehouseReq`)
- `Uuid`: 仓库UUID

#### 核心流程

**步骤 1：存在性检查**
```go
warehouse, err := warehouseRepo.GetByUuid(req.Uuid)
if err == gorm.ErrRecordNotFound {
    return errors.New("仓库不存在")
}
```

**步骤 2：业务规则检查**
```go
// 总部仓库不可设为默认
if !isEditable(ctx, warehouse.HeadquarterUuid) {
    return errors.New("仓库不可设置为默认仓库")
}

// 在途仓库不可设为默认
if warehouse.Type == constant.WarehouseTypeTransit {
    return errors.New("在途仓库不可设置为默认仓库")
}
```

**步骤 3：设置默认仓库**
```go
// 更新所有仓库：将原默认仓库的 is_default 设为 0，新仓库设为 1
err = warehouseRepo.UpdateIsDefault(warehouse.Uuid)
```

### 8. 出入库明细列表 (`GetWarehouseInOutList`)

#### 功能描述
查询仓库出入库记录，支持多维度筛选。

#### 请求参数 (`req.GetWarehouseInOutListReq`)
- `PageNo`: 页码
- `PageSize`: 每页数量
- `Keyword`: 关键词（物品名称）
- `StartTime`: 开始时间（时间戳）
- `EndTime`: 结束时间（时间戳）
- `Type`: 类型（逗号分隔，如："purchase,sale"）
- `MaterialCategoryUuids`: 物品分类UUID列表
- `SupplierUuids`: 供应商UUID列表
- `OrderNo`: 单据编号
- `OrgCodes`: 对方机构编码列表

#### 出入库类型

| 类型 | 常量 | 说明 |
|------|------|------|
| `purchase` | `WarehouseInOutLogScenePurchase` | 采购入库 |
| `sale` | `WarehouseInOutLogSceneSale` | 销售出库 |
| `delivery` | `WarehouseInOutLogSceneDelivery` | 配送出库 |
| `profit_in` | `WarehouseInOutLogSceneProfitIn` | 盘盈入库 |
| `loss_out` | `WarehouseInOutLogSceneLossOut` | 盘亏出库 |
| `transfer_in` | `WarehouseInOutLogSceneTransferIn` | 调拨入库 |
| `transfer_out` | `WarehouseInOutLogSceneTransferOut` | 调拨出库 |

#### 核心流程

**步骤 1：物品名称筛选**
```go
materialUuidsByKeyword := []uint64{}
filterByKeyword := false

if req.Keyword != "" {
    filterByKeyword = true
    // 从物品表模糊查询，获取物品UUID列表
    materialRepo := repository.NewMaterialRepo(db)
    materialUuids, err := materialRepo.GetMaterialUuidsByKeyword(req.Keyword)
    materialUuidsByKeyword = append(materialUuidsByKeyword, materialUuids...)
}
```

**步骤 2：时间区间筛选**
```go
if req.StartTime != 0 && req.EndTime != 0 {
    // 判断是毫秒时间戳还是秒时间戳
    if req.StartTime > 10000000000 {
        req.StartTime = req.StartTime / 1000
    }
    if req.EndTime > 10000000000 {
        req.EndTime = req.EndTime / 1000
    }
    if req.StartTime > req.EndTime {
        return errors.New("开始时间不能大于结束时间")
    }
    opts = append(opts, warehouseInOutLogRepo.WhereCreateTimeBetween(
        int(req.StartTime), int(req.EndTime)))
}
```

**步骤 3：类型筛选**
```go
if req.Type != "" {
    logTypes := []int{}
    for _, typ := range strings.Split(req.Type, ",") {
        logTypes = append(logTypes, constant.WarehouseInOutLogTypeToInt(typ))
    }
    opts = append(opts, warehouseInOutLogRepo.WhereSceneIn(logTypes))
}
```

**步骤 4：物品分类筛选**
```go
materialUuidsByCategory := []uint64{}
filterByCategory := false

if len(req.MaterialCategoryUuids) > 0 {
    filterByCategory = true
    materialRepo := repository.NewMaterialRepo(db)
    materialUuids, err := materialRepo.GetMaterialUuidsByCategoryUuids(
        req.MaterialCategoryUuids)
    materialUuidsByCategory = append(materialUuidsByCategory, materialUuids...)
}
```

**步骤 5：合并筛选条件（交集）**

如果同时有关键词和分类筛选，取交集：
```go
if filterByKeyword && filterByCategory && 
   len(materialUuidsByKeyword) > 0 && len(materialUuidsByCategory) > 0 {
    // 获取两个切片的交集
    materialUuids := utils.GetDuplicateElements(
        materialUuidsByKeyword, 
        materialUuidsByCategory)
    if len(materialUuids) > 0 {
        filterOpt = warehouseInOutLogRepo.WhereMaterialUuids(materialUuids)
    } else {
        // 没有交集，返回空列表
        return resp.WarehouseInOutListResp{List: make([]resp.WarehouseInOutResp, 0), ...}
    }
}
```

**步骤 6：供应商筛选**
```go
if len(req.SupplierUuids) > 0 {
    opts = append(opts, warehouseInOutLogRepo.WhereSupplierUuids(req.SupplierUuids))
}
```

**步骤 7：单据编号筛选**
```go
if req.OrderNo != "" {
    opts = append(opts, warehouseInOutLogRepo.WhereOrderNo(req.OrderNo))
}
```

**步骤 8：对方机构筛选**

解析对方机构编码：
```go
if len(req.OrgCodes) > 0 {
    supplierErpCodes := []string{}
    companyUuids := []uint64{}
    
    for _, orgCode := range req.OrgCodes {
        orgCodeArr := strings.Split(orgCode, ":")
        if len(orgCodeArr) < 2 {
            continue
        }
        
        if orgCodeArr[0] == OtherOrgCodeSupplierErpCode {
            // 供应商：supplier-erp-code:S001
            supplierErpCodes = append(supplierErpCodes, orgCodeArr[1])
        } else if orgCodeArr[0] == OtherOrgCodeCompanyUuid {
            // 公司：company-uuid:1001
            if companyUuid, err := strconv.ParseUint(orgCodeArr[1], 10, 64); err == nil {
                companyUuids = append(companyUuids, companyUuid)
            }
        }
    }
    
    if len(supplierErpCodes) > 0 || len(companyUuids) > 0 {
        opts = append(opts, warehouseInOutLogRepo.WhereSupplierErpCodesOrCompanyUuids(
            supplierErpCodes, companyUuids))
    }
}
```

**步骤 9：排除在途仓记录**
```go
// 排除在途仓的出入库记录（Scene 20, 21）
opts = append(opts, warehouseInOutLogRepo.WhereSceneNotIn([]int{20, 21}))
```

**步骤 10：分页查询**
```go
warehouseInOutLogs, total, err := warehouseInOutLogRepo.GetListWithPagination(
    req.PageNo, req.PageSize, opts...)
```

**步骤 11：构建响应**
```go
list := make([]resp.WarehouseInOutResp, 0, len(warehouseInOutLogs))
for _, log := range warehouseInOutLogs {
    list = append(list, s.buildWarehouseInOutResp(log))
}
```

#### 返回结构 (`resp.WarehouseInOutListResp`)
```go
{
    List: []WarehouseInOutResp{
        Uuid,                 // 记录UUID
        OrderNo,              // 单据编号
        Type,                 // 类型（purchase/sale/...）
        Date,                 // 日期（YYYY-MM-DD）
        Num,                  // 数量
        Amount,               // 金额
        MaterialUuid,         // 物品UUID
        MaterialName,         // 物品名称（多语言）
        MaterialCode,         // 物品编码
        MaterialBarcode,      // 物品条码
        MaterialCategoryUuid, // 物品分类UUID
        SupplierUuid,         // 供应商UUID
        SupplierErpCode,      // 供应商ERP编码
        SupplierName,         // 供应商名称
        WarehouseUuid,        // 仓库UUID
        WarehouseName,        // 仓库名称
        OtherOrgType,         // 对方机构类型
        OtherOrgName,         // 对方机构名称
    },
    Meta: {PageNo, PageSize, Total}
}
```

### 9. 同步仓库 (`SyncWarehouse`)

#### 功能描述
从ERP系统同步仓库数据到本地数据库，包括总部仓库同步（分店场景）。

#### 核心流程

**步骤 1：前置检查**
```go
if !company.IsOpenErp() {
    return errors.New("公司未开启erp")
}
```

**步骤 2：获取ERP仓库列表**
```go
companySetting := ctx.GetCompanySetting()
var warehouseErpCodes []string

warehouseList, err := erp.NewIErpSrv(s.dbm).GetWarehouseList(
    ctx.GetContext(), 
    req.GetErpnextWarehouseListReq{
        SiteCode:    companySetting.ErpnextSiteCode,
        CompanyAbbr: companySetting.ErpnextCompanyAbbr,
        Branch:      companySetting.ErpnextBranchName,
    })

for _, warehouse := range warehouseList {
    warehouseErpCodes = append(warehouseErpCodes, warehouse.Name)
}
```

**步骤 3：获取总部仓库（分店场景）**
```go
var headquarter model.CompanySetting
var headquarterWarehouses []model.Warehouse

if companySetting.IsSubShop() {
    // 获取总部公司信息
    s.dbm.GetDB(constant.DefaultDB).Model(&model.CompanySetting{}).
        Where("uuid = ?", companySetting.HeadquarterUuid).
        First(&headquarter)
    
    // 获取总部仓库
    s.dbm.GetDB(headquarter.Uuid).Model(&model.Warehouse{}).
        Preload("MultiLanguageName").
        Find(&headquarterWarehouses)
}
```

**步骤 4：获取本店仓库**
```go
db := s.dbm.GetDB(companySetting.CompanyUuid)
var warehouses []model.Warehouse
var existsDefaultWarehouse bool  // 是否已有默认仓库
var warehouseCodes []string
var deletingWarehouseUuids []uint64
warehouseMap := make(map[string]model.Warehouse)

db.Model(&model.Warehouse{}).
    Scopes(repository.ExcludeHeadquarter).
    Where("erp_code != ''").
    Find(&warehouses)

for _, warehouse := range warehouses {
    if warehouse.IsDefault == 1 {
        existsDefaultWarehouse = true
    }
    if !slices.Contains(warehouseErpCodes, warehouse.ErpCode) {
        deletingWarehouseUuids = append(deletingWarehouseUuids, warehouse.Uuid)
    }
    warehouseMap[warehouse.ErpCode] = warehouse
    warehouseCodes = append(warehouseCodes, warehouse.Code)
}
```

**步骤 5：事务处理**

```go
err = db.Transaction(func(tx *gorm.DB) error {
    // 5.1 软删除ERP中已不存在的仓库
    if len(deletingWarehouseUuids) > 0 {
        tx.Model(&model.Warehouse{}).
            Where("uuid IN (?)", deletingWarehouseUuids).
            Update("delete_time", time.Now().Unix())
    }
    
    var insertingWarehouses []model.Warehouse
    var defaultWarehouseErpCode string
    var multiLanguageNameUuids []uint64
    
    // 5.2 遍历ERP仓库
    for _, erpWarehouse := range warehouseList {
        warehouse := warehouseMap[erpWarehouse.Name]
        
        // 状态转换
        var status int
        if !erpWarehouse.Disabled {
            status = 1
        }
        
        // 判断仓库编码
        var code string
        isDefault := warehouse.IsDefault
        if strings.Contains(erpWarehouse.Name, constant.NormalWarehouseCodeContains) {
            code = constant.NormalWarehouseCode  // WH00001
            if !existsDefaultWarehouse {
                isDefault = 1  // 设为默认
            }
        } else if strings.Contains(erpWarehouse.Name, constant.TransitWarehouseCodeContains) {
            code = constant.TransitWarehouseCode  // WH00002
        }
        
        if isDefault == 1 {
            defaultWarehouseErpCode = erpWarehouse.Name
        }
        
        // 类型转换
        var warehouseType string
        if erpWarehouse.WarehouseType == constant.ErpWarehouseTypeNormal1 || 
           erpWarehouse.WarehouseType == constant.ErpWarehouseTypeNormal2 {
            warehouseType = constant.WarehouseTypeNormal
        } else if erpWarehouse.WarehouseType == constant.ErpWarehouseTypeTransit {
            warehouseType = constant.WarehouseTypeTransit
        }
        
        if warehouse.Uuid != 0 {  // 已存在，更新
            tx.Model(&model.Warehouse{}).
                Where("uuid = ?", warehouse.Uuid).
                Updates(map[string]any{
                    "type", "status", "is_default",
                    "contact", "phone", "address",
                    "delete_time": constant.NotDeleted,
                })
        } else {  // 新增
            // 创建多语言名称
            multiLanguageName := model.MultiLanguageName{
                ZhName:   erpWarehouse.AliasName,
                ThName:   erpWarehouse.AliasName,
                EnName:   erpWarehouse.AliasName,
                ZhTwName: erpWarehouse.AliasName,
                JaName:   erpWarehouse.AliasName,
                KoName:   erpWarehouse.AliasName,
                MyName:   erpWarehouse.AliasName,
                TrName:   erpWarehouse.AliasName,
                SvName:   erpWarehouse.AliasName,
            }
            tx.Model(&model.MultiLanguageName{}).Create(&multiLanguageName)
            
            // 生成编码（如果没有特定编码）
            if code == "" {
                maxCode := 2
                for _, warehouseCode := range warehouseCodes {
                    if after, ok := strings.CutPrefix(warehouseCode, "WH"); ok {
                        codeInt, _ := strconv.Atoi(after)
                        if codeInt > maxCode {
                            maxCode = codeInt
                        }
                    }
                }
                code = fmt.Sprintf("WH%02d", maxCode+1)
                warehouseCodes = append(warehouseCodes, code)
            }
            
            localeName := multiLanguageName.GetNames()
            insertingWarehouses = append(insertingWarehouses, model.Warehouse{
                Name:                  localeName.ToJson(),
                MultiLanguageNameUuid: multiLanguageName.Uuid,
                Type:                  warehouseType,
                Code:                  code,
                Status:                status,
                IsDefault:             isDefault,
                ErpCode:               erpWarehouse.Name,
                Contact, Phone, Address,
            })
            multiLanguageNameUuids = append(multiLanguageNameUuids, multiLanguageName.Uuid)
        }
    }
    
    // 5.3 同步总部仓库（分店场景）
    if len(headquarterWarehouses) > 0 {
        // 删除旧的总部仓库
        tx.Where("headquarter_uuid > 0").Delete(&model.Warehouse{})
        
        var insertingMultiLanguageNames []model.MultiLanguageName
        for _, hqWarehouse := range headquarterWarehouses {
            multiLanguageName := hqWarehouse.MultiLanguageName.GetNames()
            insertingWarehouses = append(insertingWarehouses, model.Warehouse{
                BaseModel: model.BaseModel{
                    Uuid:       hqWarehouse.Uuid,
                    CreateTime, UpdateTime, DeleteTime,
                },
                Name, MultiLanguageNameUuid, Type, Code, Status, IsDefault,
                ErpCode, Contact, Phone, Address,
                HeadquarterUuid: headquarter.Uuid,  // 标记来源
            })
            
            // 新增多语言名称（如果不存在）
            if hqWarehouse.MultiLanguageName != nil &&
               hqWarehouse.MultiLanguageName.Uuid != 0 &&
               !slices.Contains(existsMultiLanguageUuids, hqWarehouse.MultiLanguageName.Uuid) {
                insertingMultiLanguageNames = append(insertingMultiLanguageNames, 
                    model.MultiLanguageName{...})
            }
        }
        
        if len(insertingMultiLanguageNames) > 0 {
            tx.Model(&model.MultiLanguageName{}).Create(&insertingMultiLanguageNames)
        }
    }
    
    // 5.4 批量插入新仓库
    if len(insertingWarehouses) > 0 {
        tx.Model(&model.Warehouse{}).Create(&insertingWarehouses)
    }
    
    // 5.5 更新所有物品的 warehouse_uuid 为默认仓库
    var defaultWarehouse model.Warehouse
    tx.Model(&model.Warehouse{}).
        Where("erp_code = ?", defaultWarehouseErpCode).
        Scopes(repository.NotDeleted).
        Find(&defaultWarehouse)
    
    tx.Model(&model.Material{}).
        Where("id > 0").
        Update("warehouse_uuid", defaultWarehouse.Uuid)
    
    return nil
})
```

**步骤 6：添加到待翻译队列**
```go
if len(multiLanguageNameUuids) > 0 {
    s.translateSrv.AddMultiLanguageNameUuidToSet(
        ctx.GetCompanyUuid(), 
        multiLanguageNameUuids...)
}
```

### 10. 同步仓库物品库存 (`SyncWarehouseItemStock`)

#### 功能描述
从ERP系统同步仓库物品的库存数量。

#### 核心流程

**步骤 1：前置检查**
```go
if !company.IsOpenErp() {
    return errors.New("公司未开启erp")
}
```

**步骤 2：获取仓库列表**
```go
db := s.dbm.GetDB(ctx.GetCompanyUuid())
warehouseRepo := repository.NewWarehouseRepo(db)
warehouses, err := warehouseRepo.Get(repository.ExcludeHeadquarter)
```

**步骤 3：获取物品映射**
```go
var materials []model.Material
db.Model(&model.Material{}).Where("code != ''").Find(&materials)

materialMap := make(map[string]model.Material)
for _, material := range materials {
    materialMap[material.Code] = material
}
```

**步骤 4：遍历仓库同步库存**

```go
var insertingWarehouseItems []model.WarehouseItem

for _, warehouse := range warehouses {
    // 4.1 从ERP获取库存
    stockList, err := erp.NewIErpSrv(s.dbm).GetMaterialStockNum(ctx, warehouse.ErpCode)
    
    // 4.2 获取本地仓库物品
    var warehouseItems []model.WarehouseItem
    db.Model(&model.WarehouseItem{}).
        Where("warehouse_uuid = ? AND material_code != ''", warehouse.Uuid).
        Scopes(repository.NotDeleted).
        Find(&warehouseItems)
    
    warehouseItemMap := make(map[string]model.WarehouseItem)
    for _, warehouseItem := range warehouseItems {
        warehouseItemMap[warehouseItem.MaterialCode] = warehouseItem
    }
    
    // 4.3 获取物品消耗量
    materialConsumption, err := s.materialSrv.GetWarehouseItemConsumption(ctx, warehouse.Uuid)
    
    materialConsumptionMap := make(map[string]float64)
    for _, consumption := range materialConsumption.List {
        materialConsumptionMap[consumption.MaterialCode] = consumption.Consumption
    }
    
    // 4.4 计算并更新库存
    for _, stock := range stockList {
        material, ok := materialMap[stock.ItemCode]
        if !ok {
            continue
        }
        
        // 账面库存 = ERP实际数量 - 本地消耗量（可能为负数）
        stockNum := stock.ActualQty - materialConsumptionMap[stock.ItemCode]
        
        if warehouseItem, ok := warehouseItemMap[stock.ItemCode]; ok {
            // 更新已存在的库存
            db.Model(&model.WarehouseItem{}).
                Where("uuid = ?", warehouseItem.Uuid).
                Update("stock", stockNum)
        } else {
            // 创建新库存记录
            insertingWarehouseItems = append(insertingWarehouseItems, 
                model.WarehouseItem{
                    WarehouseUuid: warehouse.Uuid,
                    MaterialUuid:  material.Uuid,
                    MaterialCode:  material.Code,
                    Stock:         stockNum,
                    Valuation:     1.0,  // 默认估值
                })
        }
    }
}

// 批量插入新记录
if len(insertingWarehouseItems) > 0 {
    db.Model(&model.WarehouseItem{}).Create(&insertingWarehouseItems)
}
```

**库存计算公式**：
```
账面库存 = ERP实际数量 - 本地消耗量
```

**说明**：
- `ERP实际数量`：ERP系统中的物理库存
- `本地消耗量`：本地待同步到ERP的消耗（如销售、调拨出库等）
- 库存可能为负数（超卖情况）

### 11. 获取对方机构列表 (`GetOtherOrgList`)

#### 功能描述
获取出入库操作涉及的对方机构列表（供应商和分店）。

#### 核心流程

**步骤 1：获取供应商列表**
```go
db := ctx.GetDB()
supplierRepo := repository.NewSupplierRepo(db)

suppliers, err := supplierRepo.GetList(
    supplierRepo.OrderByCreateTime(true),
    supplierRepo.WhereNotDeleted(),
)

otherOrgs := make([]resp.OtherOrgResp, 0)
for _, supplier := range suppliers {
    if supplier.IsInternalSupplier == 1 {
        continue  // 跳过内部供应商（总部）
    }
    otherOrgs = append(otherOrgs, resp.OtherOrgResp{
        Name:          supplier.Name,
        Code:          fmt.Sprintf("%s:%s", OtherOrgCodeSupplierErpCode, supplier.ErpCode),
        IsHeadquarter: supplier.HeadquarterUuid > 0,
    })
}
```

**步骤 2：获取分店列表（总部场景）**
```go
if companySetting.IsHeadquarter() {
    companies, err := repository.NewCompanyRepo(s.dbm.GetDB(constant.DefaultDB)).
        GetListByHeadquarterUuid(ctx.GetCompanyUuid())
    
    for _, company := range companies {
        otherOrgs = append(otherOrgs, resp.OtherOrgResp{
            Name: company.Name,
            Code: fmt.Sprintf("%s:%s", OtherOrgCodeCompanyUuid, 
                strconv.FormatUint(company.Uuid, 10)),
        })
    }
}
```

#### 返回结构 (`resp.OtherOrgListResp`)
```go
{
    List: []OtherOrgResp{
        Name,          // 机构名称
        Code,          // 机构编码（格式：类型:值）
        IsHeadquarter, // 是否总部机构（仅供应商有此字段）
    }
}
```

### 12. 获取仓库物品列表 (`GetWarehouseMaterialList`)

#### 功能描述
获取指定仓库中的所有物品及其库存信息。

#### 请求参数 (`req.WarehouseMaterialListReq`)
- `WarehouseUuid`: 仓库UUID
- `PageNo`: 页码
- `PageSize`: 每页数量

#### 核心流程

**步骤 1：查询仓库物品**
```go
warehouseItemRepo := repository.NewWarehouseItemRepo(db)
materialRepo := repository.NewMaterialRepo(db)

opts := []repository.DBOption{
    warehouseItemRepo.WhereWarehouseUuid(req.WarehouseUuid),
}

warehouseItemMap := make(map[uint64]model.WarehouseItem)
warehouseItems, err := warehouseItemRepo.GetWarehouseMaterials(opts...)

for _, warehouseItem := range warehouseItems {
    warehouseItemMap[warehouseItem.MaterialUuid] = *warehouseItem
}
```

**步骤 2：分页查询物品**
```go
paginatedMaterials, total, err := materialRepo.GetMaterialListWithPagination(
    req.PageNo,
    req.PageSize,
    repository.NewCommonRepo().WhereByStatus(uint(1)),  // 仅启用的
    repository.NotDeleted,
)
```

**步骤 3：批量查询物品详情**
```go
materialUuids := make([]uint64, 0, len(paginatedMaterials))
for _, material := range paginatedMaterials {
    materialUuids = append(materialUuids, material.Uuid)
}

var materials []*model.Material
if len(materialUuids) > 0 {
    materials, err = materialRepo.GetMaterialDetailByUuids(materialUuids)
}

materialMap := make(map[uint64]*model.Material)
for _, material := range materials {
    materialMap[material.Uuid] = material
}
```

**步骤 4：构建响应数据**

```go
list := make([]resp.WarehouseMaterialInfo, 0, len(materials))

for _, material := range materials {
    item, _ := warehouseItemMap[material.Uuid]
    
    // 构建物品单位信息
    units := make([]resp.MaterialUnitInfo, 0)
    var baseUnit resp.MaterialUnitInfo
    
    // 基准单位
    if material.Unit != nil {
        baseUnit = resp.MaterialUnitInfo{
            MaterialUnitUuid: material.Unit.Uuid,
            UnitUuid:         material.Unit.UnitUuid,
            UnitName:         material.Unit.Unit.MultiLanguageName.GetNames(),
            ConversionRate:   material.Unit.ConversionRate,
            IsDefault:        material.Unit.IsDefault,
        }
    }
    
    // 非基准单位
    if len(material.NotBaseUnitList) > 0 {
        for _, unit := range material.NotBaseUnitList {
            if unit.Unit != nil {
                unitInfo := resp.MaterialUnitInfo{
                    MaterialUnitUuid: unit.Uuid,
                    UnitUuid:         unit.UnitUuid,
                    UnitName:         unit.Unit.MultiLanguageName.GetNames(),
                    ConversionRate:   unit.ConversionRate,
                    IsDefault:        unit.IsDefault,
                }
                units = append(units, unitInfo)
            }
        }
    }
    
    // 构建响应
    info := resp.WarehouseMaterialInfo{
        MaterialUuid:    material.Uuid,
        MaterialName:    material.MultiLanguageName.GetNames(),
        MaterialCode:    material.Code,
        MaterialBarcode: material.BarcodeValue,
        InternalCode:    material.InternalCode,
        BookedQuantity:  item.Stock,  // 账面库存数量
        BaseUnit:        baseUnit,
        Units:           units,
        CategoryUuid:    material.CategoryUuid,
        Image:           material.GetImage(utils.GetBaseURL(ctx.GetGin().Request)),
        Status:          material.Status,
        DeleteTime:      material.DeleteTime,
    }
    list = append(list, info)
}
```

#### 返回结构 (`resp.WarehouseMaterialListResp`)
```go
{
    List: []WarehouseMaterialInfo{
        MaterialUuid,    // 物品UUID
        MaterialName,    // 物品名称（多语言）
        MaterialCode,    // 物品编码
        MaterialBarcode, // 物品条码
        InternalCode,    // 内部编码
        BookedQuantity,  // 账面库存数量
        BaseUnit: {      // 基准单位
            MaterialUnitUuid,
            UnitUuid,
            UnitName,
            ConversionRate,
            IsDefault,
        },
        Units: [{        // 非基准单位列表
            MaterialUnitUuid,
            UnitUuid,
            UnitName,
            ConversionRate,
            IsDefault,
        }],
        CategoryUuid,    // 分类UUID
        Image,           // 图片URL
        Status,          // 状态
        DeleteTime,      // 删除时间
    },
    Meta: {PageNo, PageSize, Total}
}
```

### 13. 检查编码是否存在 (`CheckCodeExists`)

#### 功能描述
检查仓库编码是否已存在。

#### 请求参数 (`req.CheckCodeExistsReq`)
- `Code`: 仓库编码
- `Uuid`: 仓库UUID（更新时传入，用于排除自己）

#### 实现
```go
warehouseRepo := repository.NewWarehouseRepo(db)
exists, err := warehouseRepo.IsCodeExists(req.Code, req.Uuid)

return resp.CheckNameCodeExistsResp{Exists: exists}, nil
```

## 辅助函数

### 1. buildWarehouseResp
构建仓库响应数据。

```go
func (s *warehouseSrv) buildWarehouseResp(
    ctx context.Context, 
    warehouse model.Warehouse, 
    isHeadquarter bool) resp.WarehouseResp {
    
    var localName dto.LocaleResponse
    if warehouse.MultiLanguageName != nil {
        localName = warehouse.MultiLanguageName.GetNames()
    }
    
    return resp.WarehouseResp{
        Uuid:      warehouse.Uuid,
        LocalName: localName,
        Type:      warehouse.Type,
        Code: func() string {
            if isHeadquarter {
                return warehouse.ErpCode  // 总部返回ERP编码
            }
            return warehouse.Code  // 本店返回本地编码
        }(),
        Status:     warehouse.Status,
        Contact:    warehouse.Contact,
        Phone:      warehouse.Phone,
        Address:    warehouse.Address,
        IsDefault:  warehouse.IsDefault,
        IsEditable: isEditable(ctx, warehouse.HeadquarterUuid),
        HasItem:    warehouse.Items != nil && len(warehouse.Items) > 0,
    }
}
```

### 2. buildWarehouseInOutResp
构建出入库记录响应数据。

```go
func (s *warehouseSrv) buildWarehouseInOutResp(log model.WarehouseInOutLog) resp.WarehouseInOutResp {
    // 获取物品信息
    var materialName dto.LocaleResponse
    var materialCode, materialBarcode string
    var materialCategoryUuid uint64
    
    if log.Material != nil {
        if log.Material.MultiLanguageName.Uuid != 0 {
            materialName = log.Material.MultiLanguageName.GetNames()
        }
        materialCode = log.Material.Code
        materialBarcode = log.Material.BarcodeValue
        materialCategoryUuid = log.Material.CategoryUuid
    }
    
    // 获取供应商信息
    var supplierName dto.LocaleResponse
    if log.Supplier != nil {
        supplierName = dto.LocaleResponse{
            ZH: log.Supplier.Name,
            EN: log.Supplier.Name,
            // ... 其他语言
        }
    } else {
        supplierName = dto.LocaleResponse{
            ZH: log.SupplierName,
            EN: log.SupplierName,
            // ... 其他语言
        }
    }
    
    // 获取仓库信息
    var warehouseName dto.LocaleResponse
    if log.Warehouse != nil && log.Warehouse.MultiLanguageName != nil {
        warehouseName = log.Warehouse.MultiLanguageName.GetNames()
    }
    
    // 类型映射
    typeStrMap := map[int]string{
        constant.WarehouseInOutLogScenePurchase:    "purchase",
        constant.WarehouseInOutLogSceneSale:        "sale",
        constant.WarehouseInOutLogSceneDelivery:    "delivery",
        constant.WarehouseInOutLogSceneProfitIn:    "profit_in",
        constant.WarehouseInOutLogSceneLossOut:     "loss_out",
        constant.WarehouseInOutLogSceneTransferIn:  "transfer_in",
        constant.WarehouseInOutLogSceneTransferOut: "transfer_out",
    }
    
    // 格式化日期
    date := ""
    if log.CreateTime > 0 {
        date = time.Unix(log.CreateTime, 0).Format("2006-01-02")
    }
    
    return resp.WarehouseInOutResp{
        Uuid, OrderNo, Type: typeStrMap[log.Scene], Date, Num, Amount,
        MaterialUuid, MaterialName, MaterialCode, MaterialBarcode, MaterialCategoryUuid,
        SupplierUuid, SupplierErpCode, SupplierName,
        WarehouseUuid, WarehouseName,
        OtherOrgType, OtherOrgName,
    }
}
```

### 3. isEditable
判断仓库是否可编辑（仅本店创建的可编辑）。

```go
func isEditable(ctx context.Context, headquarterUuid uint64) bool {
    return headquarterUuid == 0  // 总部UUID为0表示是本店创建的
}
```

## 数据模型

### 1. model.Warehouse (仓库表)

```go
type Warehouse struct {
    BaseModel                        // Uuid, CreateTime, UpdateTime, DeleteTime
    Name                  string     // 名称（JSON格式多语言）
    MultiLanguageNameUuid uint64     // 多语言UUID
    Type                  string     // 类型（normal/transit）
    Code                  string     // 仓库编码
    Status                int        // 状态（0-禁用，1-启用）
    IsDefault             int        // 是否默认仓库（0-否，1-是）
    ErpCode               string     // ERP编码
    Contact               string     // 联系人
    Phone                 string     // 联系电话
    Address               string     // 地址
    HeadquarterUuid       uint64     // 总部UUID（> 0表示从总部同步）
    
    MultiLanguageName     *MultiLanguageName `gorm:"foreignKey:MultiLanguageNameUuid"`
    Items                 []WarehouseItem    `gorm:"foreignKey:WarehouseUuid"`
}
```

### 2. model.WarehouseItem (仓库物品表)

```go
type WarehouseItem struct {
    BaseModel                    // Uuid, CreateTime, UpdateTime, DeleteTime
    WarehouseUuid uint64         // 仓库UUID
    MaterialUuid  uint64         // 物品UUID
    MaterialCode  string         // 物品编码
    Stock         float64        // 库存数量（可为负数）
    Valuation     float64        // 估值
}
```

### 3. model.WarehouseInOutLog (出入库记录表)

```go
type WarehouseInOutLog struct {
    BaseModel                     // Uuid, CreateTime, UpdateTime, DeleteTime
    OrderNo         string        // 单据编号
    Scene           int           // 场景类型
    Num             float64       // 数量
    Amount          float64       // 金额
    MaterialUuid    uint64        // 物品UUID
    SupplierUuid    uint64        // 供应商UUID
    SupplierErpCode string        // 供应商ERP编码
    SupplierName    string        // 供应商名称
    WarehouseUuid   uint64        // 仓库UUID
    OtherOrgType    string        // 对方机构类型
    OtherOrgName    string        // 对方机构名称
    
    Material        *Material     `gorm:"foreignKey:MaterialUuid"`
    Supplier        *Supplier     `gorm:"foreignKey:SupplierUuid"`
    Warehouse       *Warehouse    `gorm:"foreignKey:WarehouseUuid"`
}
```

## 业务规则

### 1. 仓库编码规则

**固定编码**：
- 普通仓库（第一个）：`WH00001`
- 在途仓库：`WH00002`

**自动编码**：
其他仓库按 `WH03`, `WH04`, `WH05` ... 依次递增

**编码规则**：
- 自动转大写
- 必须唯一
- 格式：`WH` + 5位数字

### 2. 默认仓库规则

**设置规则**：
- 每个公司只能有一个默认仓库
- 设置新默认仓库时，自动取消其他仓库的默认状态
- 在途仓库不能设为默认
- 总部同步的仓库不能设为默认

**自动设置**：
同步时，如果没有默认仓库，自动将第一个普通仓库设为默认。

### 3. 删除限制

**不可删除的仓库**：
- 默认仓库
- 在途仓库
- 总部同步的仓库
- 有物品的仓库

### 4. 禁用限制

**不可禁用的仓库**：
- 默认仓库
- 在途仓库

### 5. 总部/分店模式

#### 总部逻辑
- 可以创建和管理自己的仓库
- 可以查看所有分店（`GetOtherOrgList`中）

#### 分店逻辑
- 可以看到总部同步下来的仓库（`HeadquarterUuid > 0`）
- 总部仓库**不可编辑、不可删除、不可设为默认**
- 可以创建和管理自己的仓库（`HeadquarterUuid = 0`）

### 6. ERP 集成规则

#### 创建仓库
- 本地先检查编码和名称唯一性
- 调用ERP创建接口，获取 `ErpCode`
- 保存到本地数据库，记录 `ErpCode`

#### 更新仓库
- 仅更新有 `ErpCode` 的仓库（即自己创建的）
- 同步更新ERP和本地数据

#### 删除仓库
- 本地软删除（设置 `delete_time`）
- ERP中标记为禁用（Disabled）
- 忽略ERP的 "not found" 错误

#### 同步仓库
- **定时触发**或**手动触发**
- **双向同步**：
  - ERP → 本地：新增/更新仓库
  - 本地 → ERP：软删除本地不存在的仓库
- **总部分店同步**：分店同步总部仓库数据
- **默认仓库处理**：自动设置默认仓库，更新物品的默认仓库

#### 同步库存
- 从ERP获取实际库存
- 减去本地消耗量
- 计算账面库存（可能为负数）

### 7. 库存计算规则

**公式**：
```
账面库存 = ERP实际数量 - 本地消耗量
```

**说明**：
- `ERP实际数量`：ERP系统中的物理库存
- `本地消耗量`：本地待同步到ERP的消耗（如销售、调拨出库等）
- 允许为负数（超卖情况）

### 8. 出入库记录规则

**记录类型**：
- 入库：采购、盘盈、调拨入
- 出库：销售、配送、盘亏、调拨出

**查询规则**：
- 排除在途仓的记录（Scene 20, 21）
- 支持多维度筛选（时间、类型、物品、供应商、对方机构）
- 关键词和分类筛选取交集

## 依赖关系

### 外部服务依赖

1. **ERP Service** (`app/service/rpc/erp`):
   - `CreateWarehouse`: 创建仓库到ERP
   - `UpdateWarehouse`: 更新ERP中的仓库
   - `DeleteWarehouse`: 删除（禁用）ERP中的仓库
   - `GetWarehouseList`: 获取ERP仓库列表
   - `GetMaterialStockNum`: 获取物品库存数量

2. **Setting Service** (`app/service/setting`):
   - 获取公司设置信息

3. **Material Service** (`IMaterialSrv`):
   - `GetWarehouseItemConsumption`: 获取物品消耗量

4. **Translate Service** (`ITranslateSrv`):
   - `AddMultiLanguageNameUuidToSet`: 添加到待翻译队列
   - `RemoveMultiLanguageNameUuidFromSet`: 从待翻译队列移除

5. **CheckName Service** (`NewCheckNameSrv`):
   - 名称唯一性和长度检查

### Repository 依赖

1. **WarehouseRepo** (`repository.NewWarehouseRepo`):
   - CRUD操作
   - `IsCodeExists`: 编码唯一性检查
   - `UpdateIsDefault`: 设置默认仓库
   - 各种 `Where*` 条件构建器

2. **WarehouseItemRepo** (`repository.NewWarehouseItemRepo`):
   - `GetWarehouseMaterials`: 获取仓库物品

3. **WarehouseInOutLogRepo** (`repository.NewWarehouseInOutLogRepo`):
   - 出入库记录查询
   - 各种筛选条件

4. **MaterialRepo** (`repository.NewMaterialRepo`):
   - `GetMaterialUuidsByKeyword`: 按关键词搜索物品
   - `GetMaterialUuidsByCategoryUuids`: 按分类获取物品
   - `GetMaterialListWithPagination`: 分页查询物品
   - `GetMaterialDetailByUuids`: 批量获取物品详情

5. **SupplierRepo** (`repository.NewSupplierRepo`):
   - 获取供应商列表

6. **CompanyRepo** (`repository.NewCompanyRepo`):
   - `GetListByHeadquarterUuid`: 获取总部下的分店列表

### 数据库依赖

- 使用 `dbm.GetDB(ctx.GetDbId())` 获取当前公司数据库
- 使用 `dbm.GetDB(constant.DefaultDB)` 获取默认数据库（查询总部信息）
- 使用 `dbm.GetDB(headquarter.Uuid)` 获取总部数据库

## 使用示例

### 1. 创建仓库
```go
warehouseSrv := NewWarehouseSrv(dbm, settingSrv, materialSrv, translateSrv)

err := warehouseSrv.CreateWarehouse(ctx, req.CreateWarehouseReq{
    LocaleName: dto.LocaleResponse{
        ZH: "主仓库",
        EN: "Main Warehouse",
        // ... 其他语言
    },
    Code:    "WH001",
    Type:    "normal",
    Status:  1,
    Contact: "张三",
    Phone:   "13800138000",
    Address: "广州市天河区",
})
```

### 2. 获取仓库列表
```go
resp, err := warehouseSrv.GetWarehouseList(ctx, req.WarehouseListReq{
    PageReq: dto.PageReq{PageNo: 1, PageSize: 20},
    Type:    "normal",
    Status:  &[]int{1}[0],  // 仅启用的
})
```

### 3. 设置默认仓库
```go
err := warehouseSrv.SetDefaultWarehouse(ctx, req.SetDefaultWarehouseReq{
    Uuid: warehouseUuid,
})
```

### 4. 查询出入库记录
```go
resp, err := warehouseSrv.GetWarehouseInOutList(ctx, req.GetWarehouseInOutListReq{
    PageNo:   1,
    PageSize: 20,
    Type:     "purchase,sale",
    StartTime: startTimestamp,
    EndTime:   endTimestamp,
    MaterialCategoryUuids: []uint64{1001, 1002},
})
```

### 5. 同步仓库
```go
err := warehouseSrv.SyncWarehouse(ctx)
```

### 6. 同步库存
```go
err := warehouseSrv.SyncWarehouseItemStock(ctx)
```

## 性能优化

### 1. 批量操作

**同步时批量插入**：
```go
if len(insertingWarehouses) > 0 {
    tx.Model(&model.Warehouse{}).Create(&insertingWarehouses)
}
```

**批量查询物品详情**：
```go
materials, err = materialRepo.GetMaterialDetailByUuids(materialUuids)
```

### 2. 索引优化

关键字段应建立索引：
- `erp_code`：频繁用于查询和同步
- `code`：用于唯一性检查
- `headquarter_uuid`：用于区分本店/总部仓库
- `is_default`：用于查询默认仓库
- `type`：用于类型筛选

### 3. 分页查询

所有列表查询都使用分页，避免一次加载大量数据：
```go
warehouses, total, err := warehouseRepo.GetListWithPagination(pageNo, pageSize, opts...)
```

### 4. 预加载关联数据

出入库记录查询时预加载关联数据：
```go
warehouseInOutLogs, total, err := warehouseInOutLogRepo.GetListWithPagination(...)
// 自动预加载 Material, Supplier, Warehouse
```

## 总结

`warehouse.go` 实现了一个功能完整、逻辑复杂的仓库管理服务，主要特点包括：

1. **完整的仓库管理**：支持创建、查询、更新、删除、默认设置
2. **ERP 深度集成**：创建、更新、删除、同步全流程与ERP系统打通
3. **总部分店模式**：支持总部和分店的独立管理及数据同步
4. **双仓库类型**：普通仓库和在途仓库，满足不同业务场景
5. **库存管理**：同步ERP库存，计算账面库存（实际数量 - 消耗量）
6. **出入库记录**：完整的出入库记录查询，支持多维度筛选
7. **对方机构管理**：支持供应商和分店作为出入库对方机构
8. **多语言支持**：仓库名称支持9种语言
9. **完善的业务规则**：默认仓库、删除限制、禁用限制等
10. **性能优化**：批量操作、分页查询、索引优化

该服务是库存管理模块的核心基础，为采购、销售、调拨、盘点等业务提供支持。

