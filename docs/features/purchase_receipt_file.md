# Purchase Receipt File Service 收货单附件服务说明文档

## 📋 概述

`service/purchase_receipt_file.go` 是 TTPOS 系统的收货单附件管理服务，负责处理收货单与文件附件的关联关系。该服务提供附件的上传关联、查询、删除等功能，并实施附件数量限制和状态验证等业务规则，确保收货单附件管理的规范性和安全性。

**文件路径**: `/home/coder/workspaces/ttpos-server-go/main/app/service/purchase_receipt_file.go`  
**文件大小**: 183行  
**接口定义**: `IPurchaseReceiptFileSrv`  
**实现结构**: `purchaseReceiptFileSrv`

---

## 🏗️ 架构设计

### 接口定义 (IPurchaseReceiptFileSrv)

```go
type IPurchaseReceiptFileSrv interface {
    // 保存收货单附件关联
    SaveReceiptFiles(ctx context.Context, receiptOrderUuid uint64, fileUuids []uint64) error
    
    // 查询收货单附件列表
    GetReceiptFiles(ctx context.Context, receiptOrderUuid uint64) ([]resp.ReceiptFileInfo, error)
    
    // 删除收货单附件
    DeleteReceiptFile(ctx context.Context, fileUuid uint64, receiptOrderUuid uint64) error
    
    // 删除收货单的所有附件
    DeleteAllReceiptFiles(ctx context.Context, receiptOrderUuid uint64) error
    
    // 验证附件数量限制
    ValidateFileLimit(ctx context.Context, receiptOrderUuid uint64, newFileCount int) error
    
    // 验证收货单状态（草稿状态才能编辑）
    ValidateReceiptStatus(ctx context.Context, receiptOrderUuid uint64) error
}
```

### 依赖服务

```go
type purchaseReceiptFileSrv struct {
    dbm *database.DBManager  // 数据库管理器
}
```

### 服务初始化

```go
func NewPurchaseReceiptFileSrv(dbm *database.DBManager) IPurchaseReceiptFileSrv {
    return NewPurchaseReceiptFileSrvImpl(dbm)
}

func NewPurchaseReceiptFileSrvImpl(dbm *database.DBManager) IPurchaseReceiptFileSrv {
    return &purchaseReceiptFileSrv{
        dbm: dbm,
    }
}
```

**初始化参数**:
- `dbm`: 数据库管理器，用于获取数据库连接

---

## 🎯 核心概念

### 1. 收货单附件关系

收货单附件通过中间表 `PurchaseReceiptFile` 关联：

```
PurchaseReceiptOrder (收货单)
         ↓
PurchaseReceiptFile (关联表)
         ↓
File (文件)
```

#### 数据模型

```go
type PurchaseReceiptFile struct {
    BaseModel                          // Uuid, CreateTime, UpdateTime, DeleteTime
    ReceiptOrderUuid uint64           // 收货单UUID
    FileUuid         uint64           // 文件UUID
    SortOrder        int              // 排序顺序
    File             *model.File      // 关联的文件对象
}
```

### 2. 业务规则

| 规则 | 说明 |
|-----|------|
| 附件数量限制 | 每个收货单最多10个附件 |
| 状态限制 | 仅草稿状态（Pending）可以编辑附件 |
| 排序规则 | 按上传顺序自动设置SortOrder |
| 级联删除 | 删除收货单时可批量删除附件 |

### 3. 收货单状态

| 状态 | 常量 | 可编辑附件 |
|-----|------|-----------|
| 草稿 | `ReceiptOrderStatusPending` | ✅ 是 |
| 已提交 | 其他状态 | ❌ 否 |

---

## 🎯 核心功能

### 1. 保存收货单附件 (SaveReceiptFiles)

**功能描述**: 批量关联文件到收货单，自动设置排序顺序。

#### 方法签名

```go
func (s *purchaseReceiptFileSrv) SaveReceiptFiles(
    ctx context.Context, 
    receiptOrderUuid uint64, 
    fileUuids []uint64
) error
```

#### 参数说明

| 参数 | 类型 | 说明 |
|-----|------|-----|
| `ctx` | context.Context | 请求上下文 |
| `receiptOrderUuid` | uint64 | 收货单UUID |
| `fileUuids` | []uint64 | 文件UUID列表 |

#### 执行流程

```
1. 验证参数
   ↓
2. 验证附件数量限制
   - 新增数量 ≤ 10
   - 当前数量 + 新增数量 ≤ 10
   ↓
3. 构建附件关联数据
   ├─ 生成UUID
   ├─ 设置ReceiptOrderUuid
   ├─ 设置FileUuid
   └─ 设置SortOrder（按索引）
   ↓
4. 批量创建关联记录
   ↓
5. 返回结果
```

#### 代码实现

```go
func (s *purchaseReceiptFileSrv) SaveReceiptFiles(ctx context.Context, receiptOrderUuid uint64, fileUuids []uint64) error {
    // 1. 空列表直接返回
    if len(fileUuids) == 0 {
        return nil
    }
    
    // 2. 验证附件数量限制
    if err := s.ValidateFileLimit(ctx, receiptOrderUuid, len(fileUuids)); err != nil {
        return err
    }
    
    db := ctx.GetDB()
    repo := repository.NewPurchaseReceiptFileRepo(db)
    
    // 3. 批量创建附件关联
    var files []model.PurchaseReceiptFile
    for idx, fileUuid := range fileUuids {
        uuid, err := utils.GetID()
        if err != nil {
            return errors.WithMessage(err, "生成UUID失败")
        }
        
        files = append(files, model.PurchaseReceiptFile{
            BaseModel: model.BaseModel{
                Uuid: uuid,
            },
            ReceiptOrderUuid: receiptOrderUuid,
            FileUuid:         fileUuid,
            SortOrder:        idx,  // 按索引设置排序
        })
    }
    
    // 4. 批量插入
    return repo.BatchCreate(files)
}
```

#### 使用示例

```go
// 上传文件并关联到收货单
fileUuids := []uint64{101, 102, 103}  // 已上传的文件UUID列表
receiptOrderUuid := uint64(12345)

err := purchaseReceiptFileSrv.SaveReceiptFiles(ctx, receiptOrderUuid, fileUuids)
if err != nil {
    if err.Error() == "最多支持10个附件" {
        // 提示用户附件数量超限
    }
}
```

---

### 2. 查询收货单附件列表 (GetReceiptFiles)

**功能描述**: 查询收货单的所有附件，返回附件详细信息。

#### 方法签名

```go
func (s *purchaseReceiptFileSrv) GetReceiptFiles(
    ctx context.Context, 
    receiptOrderUuid uint64
) ([]resp.ReceiptFileInfo, error)
```

#### 参数说明

| 参数 | 类型 | 说明 |
|-----|------|-----|
| `ctx` | context.Context | 请求上下文 |
| `receiptOrderUuid` | uint64 | 收货单UUID |

#### 返回数据结构

```go
type ReceiptFileInfo struct {
    FileUuid   uint64 `json:"file_uuid"`   // 文件UUID
    FileName   string `json:"file_name"`   // 文件名
    FileSize   int64  `json:"file_size"`   // 文件大小（字节）
    FileType   string `json:"file_type"`   // 文件MIME类型
    Extension  string `json:"extension"`   // 文件扩展名
    FilePath   string `json:"file_path"`   // 文件访问URL
    SortOrder  int    `json:"sort_order"`  // 排序顺序
    CreateTime int    `json:"create_time"` // 创建时间
}
```

#### 执行流程

```
1. 获取数据库连接
   ↓
2. 查询附件关联（预加载文件信息）
   ↓
3. 遍历结果
   ├─ 跳过File为空的记录
   ├─ 获取BaseURL
   └─ 构建文件访问URL
   ↓
4. 转换为响应格式
   ↓
5. 返回附件列表
```

#### 代码实现

```go
func (s *purchaseReceiptFileSrv) GetReceiptFiles(ctx context.Context, receiptOrderUuid uint64) ([]resp.ReceiptFileInfo, error) {
    db := ctx.GetDB()
    repo := repository.NewPurchaseReceiptFileRepo(db)
    
    // 查询附件关联（预加载文件信息）
    files, err := repo.GetByReceiptOrderUuidWithFiles(receiptOrderUuid)
    if err != nil {
        return nil, errors.WithMessage(err, "查询收货单附件失败")
    }
    
    // 转换为响应格式
    var result = make([]resp.ReceiptFileInfo, 0)
    baseURL := utils.GetBaseURL(ctx.GetGin().Request)
    
    for _, file := range files {
        if file.File == nil {
            continue  // 跳过文件对象为空的记录
        }
        
        result = append(result, resp.ReceiptFileInfo{
            FileUuid:   file.FileUuid,
            FileName:   file.File.RealName,
            FileSize:   int64(file.File.FileSize),
            FileType:   file.File.FileType,
            Extension:  file.File.Extension,
            FilePath:   file.File.GetUrl(baseURL),  // 生成完整访问URL
            SortOrder:  file.SortOrder,
            CreateTime: int(file.CreateTime),
        })
    }
    
    return result, nil
}
```

#### 使用示例

```go
// 查询收货单附件
receiptOrderUuid := uint64(12345)
files, err := purchaseReceiptFileSrv.GetReceiptFiles(ctx, receiptOrderUuid)
if err != nil {
    // 错误处理
}

// 遍历附件
for _, file := range files {
    fmt.Printf("文件名: %s, 大小: %d, URL: %s\n", 
        file.FileName, file.FileSize, file.FilePath)
}
```

---

### 3. 删除收货单附件 (DeleteReceiptFile)

**功能描述**: 删除收货单的单个附件，需要验证收货单状态。

#### 方法签名

```go
func (s *purchaseReceiptFileSrv) DeleteReceiptFile(
    ctx context.Context, 
    fileUuid uint64, 
    receiptOrderUuid uint64
) error
```

#### 参数说明

| 参数 | 类型 | 说明 |
|-----|------|-----|
| `ctx` | context.Context | 请求上下文 |
| `fileUuid` | uint64 | 文件UUID |
| `receiptOrderUuid` | uint64 | 收货单UUID |

#### 执行流程

```
1. 验证收货单状态
   - 仅草稿状态可以删除
   ↓
2. 删除附件关联记录
   ↓
3. 返回结果
```

#### 代码实现

```go
func (s *purchaseReceiptFileSrv) DeleteReceiptFile(ctx context.Context, fileUuid uint64, receiptOrderUuid uint64) error {
    // 验证收货单状态
    if err := s.ValidateReceiptStatus(ctx, receiptOrderUuid); err != nil {
        return err
    }
    
    db := ctx.GetDB()
    repo := repository.NewPurchaseReceiptFileRepo(db)
    
    // 删除关联记录
    return repo.DeleteByFileUuidAndReceiptOrderUuid(fileUuid, receiptOrderUuid)
}
```

#### 使用示例

```go
// 删除附件
fileUuid := uint64(101)
receiptOrderUuid := uint64(12345)

err := purchaseReceiptFileSrv.DeleteReceiptFile(ctx, fileUuid, receiptOrderUuid)
if err != nil {
    if err.Error() == "仅草稿状态的收货单可以修改附件" {
        // 提示用户收货单状态不允许删除
    }
}
```

---

### 4. 删除收货单的所有附件 (DeleteAllReceiptFiles)

**功能描述**: 批量删除收货单的所有附件关联，通常在删除收货单时调用。

#### 方法签名

```go
func (s *purchaseReceiptFileSrv) DeleteAllReceiptFiles(
    ctx context.Context, 
    receiptOrderUuid uint64
) error
```

#### 参数说明

| 参数 | 类型 | 说明 |
|-----|------|-----|
| `ctx` | context.Context | 请求上下文 |
| `receiptOrderUuid` | uint64 | 收货单UUID |

#### 执行流程

```
1. 获取数据库连接
   ↓
2. 批量删除所有附件关联
   ↓
3. 返回结果
```

#### 代码实现

```go
func (s *purchaseReceiptFileSrv) DeleteAllReceiptFiles(ctx context.Context, receiptOrderUuid uint64) error {
    db := ctx.GetDB()
    repo := repository.NewPurchaseReceiptFileRepo(db)
    
    // 批量删除所有附件关联
    return repo.DeleteByReceiptOrderUuid(receiptOrderUuid)
}
```

#### 使用场景

```go
// 删除收货单时同时删除所有附件
func (s *PurchaseReceiptOrderSrv) DeleteReceiptOrder(ctx context.Context, receiptOrderUuid uint64) error {
    // 删除附件关联
    if err := s.receiptFileSrv.DeleteAllReceiptFiles(ctx, receiptOrderUuid); err != nil {
        return err
    }
    
    // 删除收货单
    // ...
}
```

---

### 5. 验证附件数量限制 (ValidateFileLimit)

**功能描述**: 验证附件数量是否超过限制（最多10个）。

#### 方法签名

```go
func (s *purchaseReceiptFileSrv) ValidateFileLimit(
    ctx context.Context, 
    receiptOrderUuid uint64, 
    newFileCount int
) error
```

#### 参数说明

| 参数 | 类型 | 说明 |
|-----|------|-----|
| `ctx` | context.Context | 请求上下文 |
| `receiptOrderUuid` | uint64 | 收货单UUID |
| `newFileCount` | int | 新增附件数量 |

#### 验证逻辑

```
1. 验证新增数量
   - newFileCount > 10 → 返回错误
   ↓
2. 查询当前附件数量
   ↓
3. 验证总数量
   - currentCount + newFileCount > 10 → 返回错误
   ↓
4. 验证通过
```

#### 代码实现

```go
func (s *purchaseReceiptFileSrv) ValidateFileLimit(ctx context.Context, receiptOrderUuid uint64, newFileCount int) error {
    // 1. 验证新增数量
    if newFileCount > 10 {
        return errors.New("最多支持10个附件")
    }
    
    db := ctx.GetDB()
    repo := repository.NewPurchaseReceiptFileRepo(db)
    
    // 2. 查询当前附件数量
    currentCount, err := repo.CountByReceiptOrderUuid(receiptOrderUuid)
    if err != nil {
        return errors.WithMessage(err, "查询附件数量失败")
    }
    
    // 3. 验证总数量
    if currentCount+int64(newFileCount) > 10 {
        return errors.New("最多支持10个附件")
    }
    
    return nil
}
```

#### 验证场景

| 场景 | 当前数量 | 新增数量 | 结果 |
|-----|---------|---------|------|
| 首次上传 | 0 | 5 | ✅ 通过 |
| 追加上传 | 8 | 2 | ✅ 通过 |
| 超出限制 | 8 | 3 | ❌ 失败 |
| 批量超限 | 0 | 11 | ❌ 失败 |

---

### 6. 验证收货单状态 (ValidateReceiptStatus)

**功能描述**: 验证收货单是否为草稿状态，仅草稿状态可以编辑附件。

#### 方法签名

```go
func (s *purchaseReceiptFileSrv) ValidateReceiptStatus(
    ctx context.Context, 
    receiptOrderUuid uint64
) error
```

#### 参数说明

| 参数 | 类型 | 说明 |
|-----|------|-----|
| `ctx` | context.Context | 请求上下文 |
| `receiptOrderUuid` | uint64 | 收货单UUID |

#### 验证逻辑

```
1. 查询收货单
   ↓
2. 验证收货单是否存在
   ↓
3. 验证收货单状态
   - status != Pending → 返回错误
   ↓
4. 验证通过
```

#### 代码实现

```go
func (s *purchaseReceiptFileSrv) ValidateReceiptStatus(ctx context.Context, receiptOrderUuid uint64) error {
    db := ctx.GetDB()
    receiptRepo := repository.NewPurchaseReceiptOrderRepo(db)
    
    // 查询收货单
    receiptOrder, err := receiptRepo.GetByUuid(receiptOrderUuid)
    if err != nil {
        return errors.WithMessage(err, "收货单不存在")
    }
    
    // 仅草稿状态可以编辑附件
    if receiptOrder.Status != constant.ReceiptOrderStatusPending {
        return errors.New("仅草稿状态的收货单可以修改附件")
    }
    
    return nil
}
```

#### 状态限制

| 收货单状态 | 可编辑附件 | 说明 |
|-----------|-----------|------|
| 草稿 (Pending) | ✅ 是 | 可以增删改附件 |
| 已提交 | ❌ 否 | 不可修改附件 |
| 已审核 | ❌ 否 | 不可修改附件 |
| 已完成 | ❌ 否 | 不可修改附件 |

---

## 🔄 业务流程

### 1. 上传附件流程

```
用户选择文件
  ↓
调用文件上传接口
  ├─ 上传到文件服务器
  └─ 创建File记录
  ↓
获取FileUuid列表
  ↓
调用SaveReceiptFiles
  ├─ 验证附件数量限制
  ├─ 创建关联记录
  └─ 设置排序顺序
  ↓
返回成功
  ↓
前端显示附件列表
```

### 2. 查看附件流程

```
打开收货单详情
  ↓
调用GetReceiptFiles
  ├─ 查询附件关联
  ├─ 预加载文件信息
  └─ 构建访问URL
  ↓
返回附件列表
  ↓
前端显示附件
  ├─ 图片预览
  ├─ 文件下载链接
  └─ 文件信息展示
```

### 3. 删除附件流程

```
用户点击删除
  ↓
调用DeleteReceiptFile
  ├─ 验证收货单状态
  └─ 删除关联记录
  ↓
返回成功
  ↓
前端移除显示
```

### 4. 提交收货单流程

```
用户填写收货单信息
  ↓
上传附件（可选）
  ↓
用户点击提交
  ↓
收货单状态变更
  ├─ Pending → Submitted
  └─ 附件不可再编辑
  ↓
提交成功
```

---

## 🎨 API接口示例

### 1. 上传附件接口

#### 请求

```http
POST /api/v1/purchase/receipt_order/files/save
Authorization: Bearer {token}
Content-Type: application/json

{
  "receipt_order_uuid": 12345,
  "file_uuids": [101, 102, 103]
}
```

#### 响应

```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

#### 错误响应

```json
{
  "code": 0,
  "message": "最多支持10个附件",
  "data": {}
}
```

### 2. 查询附件列表接口

#### 请求

```http
GET /api/v1/purchase/receipt_order/files?receipt_order_uuid=12345
Authorization: Bearer {token}
```

#### 响应

```json
{
  "code": 1,
  "message": "success",
  "data": {
    "list": [
      {
        "file_uuid": 101,
        "file_name": "收货单照片.jpg",
        "file_size": 1024000,
        "file_type": "image/jpeg",
        "extension": "jpg",
        "file_path": "http://example.com/uploads/2024/11/12/abc123.jpg",
        "sort_order": 0,
        "create_time": 1699000000
      },
      {
        "file_uuid": 102,
        "file_name": "签收单.pdf",
        "file_size": 512000,
        "file_type": "application/pdf",
        "extension": "pdf",
        "file_path": "http://example.com/uploads/2024/11/12/def456.pdf",
        "sort_order": 1,
        "create_time": 1699000100
      }
    ]
  }
}
```

### 3. 删除附件接口

#### 请求

```http
DELETE /api/v1/purchase/receipt_order/files/delete
Authorization: Bearer {token}
Content-Type: application/json

{
  "receipt_order_uuid": 12345,
  "file_uuid": 101
}
```

#### 响应

```json
{
  "code": 1,
  "message": "success",
  "data": {}
}
```

#### 错误响应

```json
{
  "code": 0,
  "message": "仅草稿状态的收货单可以修改附件",
  "data": {}
}
```

### 4. Controller实现示例

```go
// SaveReceiptFiles 保存收货单附件
// @Summary 保存收货单附件
// @Description 批量关联文件到收货单
// @Tags 收货单管理
// @Accept json
// @Produce json
// @Param request body req.SaveReceiptFilesReq true "请求参数"
// @Success 200 {object} dto.Response "成功"
// @Security JwtToken
// @Router /api/v1/purchase/receipt_order/files/save [post]
func (c *PurchaseReceiptOrderController) SaveReceiptFiles(ctx *gin.Context) {
    var req req.SaveReceiptFilesReq
    if err := ctx.ShouldBindJSON(&req); err != nil {
        response.Error(ctx, "参数错误")
        return
    }
    
    err := c.receiptFileSrv.SaveReceiptFiles(ctx, req.ReceiptOrderUuid, req.FileUuids)
    if err != nil {
        response.Error(ctx, err.Error())
        return
    }
    
    response.Success(ctx, nil)
}

// GetReceiptFiles 查询收货单附件列表
// @Summary 查询收货单附件列表
// @Description 获取收货单的所有附件
// @Tags 收货单管理
// @Accept json
// @Produce json
// @Param receipt_order_uuid query uint64 true "收货单UUID"
// @Success 200 {object} resp.ReceiptFileListResp "成功"
// @Security JwtToken
// @Router /api/v1/purchase/receipt_order/files [get]
func (c *PurchaseReceiptOrderController) GetReceiptFiles(ctx *gin.Context) {
    receiptOrderUuid := utils.ParseUint64(ctx.Query("receipt_order_uuid"))
    
    files, err := c.receiptFileSrv.GetReceiptFiles(ctx, receiptOrderUuid)
    if err != nil {
        response.Error(ctx, err.Error())
        return
    }
    
    response.Success(ctx, map[string]interface{}{
        "list": files,
    })
}
```

---

## 📊 数据库表结构

### purchase_receipt_file 表

| 字段 | 类型 | 说明 |
|-----|------|------|
| uuid | bigint | 主键UUID |
| receipt_order_uuid | bigint | 收货单UUID |
| file_uuid | bigint | 文件UUID |
| sort_order | int | 排序顺序 |
| create_time | int | 创建时间 |
| update_time | int | 更新时间 |
| delete_time | int | 删除时间 |

### 索引建议

```sql
-- 收货单UUID索引（查询附件列表）
CREATE INDEX idx_receipt_order_uuid ON purchase_receipt_file(receipt_order_uuid);

-- 文件UUID索引（删除单个附件）
CREATE INDEX idx_file_uuid ON purchase_receipt_file(file_uuid);

-- 联合索引（删除附件时快速定位）
CREATE INDEX idx_receipt_file ON purchase_receipt_file(receipt_order_uuid, file_uuid);
```

---

## 🛡️ 最佳实践

### 1. 附件上传流程

```go
// ✅ 正确：先上传文件，再关联到收货单
func (c *Controller) UploadAndSaveFiles(ctx *gin.Context) {
    // 1. 上传文件
    files, _ := ctx.MultipartForm()
    var fileUuids []uint64
    for _, file := range files.File["files"] {
        fileUuid, _ := c.fileSrv.Upload(ctx, file)
        fileUuids = append(fileUuids, fileUuid)
    }
    
    // 2. 关联到收货单
    c.receiptFileSrv.SaveReceiptFiles(ctx, receiptOrderUuid, fileUuids)
}

// ❌ 错误：不验证数量就保存
func (c *Controller) SaveFiles(ctx *gin.Context) {
    // 直接保存，可能超出限制
    c.receiptFileSrv.SaveReceiptFiles(ctx, receiptOrderUuid, fileUuids)
}
```

### 2. 状态验证

```go
// ✅ 正确：删除前验证状态
err := receiptFileSrv.DeleteReceiptFile(ctx, fileUuid, receiptOrderUuid)
if err != nil {
    if err.Error() == "仅草稿状态的收货单可以修改附件" {
        // 友好提示用户
        response.Error(ctx, "收货单已提交，无法删除附件")
        return
    }
}

// ❌ 错误：不验证状态直接删除
repo.DeleteByFileUuidAndReceiptOrderUuid(fileUuid, receiptOrderUuid)
```

### 3. 查询附件列表

```go
// ✅ 正确：使用服务获取完整信息
files, _ := receiptFileSrv.GetReceiptFiles(ctx, receiptOrderUuid)
// 返回包含文件访问URL的完整信息

// ❌ 错误：直接查询关联表
files, _ := repo.GetByReceiptOrderUuid(receiptOrderUuid)
// 缺少文件详细信息和访问URL
```

### 4. 删除收货单

```go
// ✅ 正确：先删除附件关联，再删除收货单
db.Transaction(func(tx *gorm.DB) error {
    // 删除附件关联
    receiptFileSrv.DeleteAllReceiptFiles(ctx, receiptOrderUuid)
    
    // 删除收货单
    receiptOrderRepo.Delete(receiptOrderUuid)
    
    return nil
})

// ❌ 错误：忘记删除附件关联
receiptOrderRepo.Delete(receiptOrderUuid)
// 会留下孤立的附件关联记录
```

---

## ⚠️ 注意事项

### 1. 附件数量限制

- 每个收货单最多10个附件
- 上传前需要验证
- 包括已有附件数量

### 2. 状态限制

- 仅草稿状态可以编辑附件
- 提交后不可修改
- 删除时需要验证状态

### 3. 文件生命周期

- 删除关联不会删除文件本身
- 文件由文件服务统一管理
- 可以实现文件复用

### 4. 排序顺序

- SortOrder按上传顺序自动设置
- 从0开始递增
- 前端可以按SortOrder排序显示

### 5. 并发安全

- 数量验证在事务外
- 可能存在并发问题
- 建议加锁或使用数据库约束

---

## 🎯 使用场景

### 1. 收货拍照

```go
// 收货员拍照上传
func (c *Controller) UploadReceiptPhotos(ctx *gin.Context) {
    receiptOrderUuid := utils.ParseUint64(ctx.Query("receipt_order_uuid"))
    
    // 1. 上传照片
    files, _ := ctx.MultipartForm()
    var fileUuids []uint64
    for _, file := range files.File["photos"] {
        fileUuid, _ := c.fileSrv.Upload(ctx, file)
        fileUuids = append(fileUuids, fileUuid)
    }
    
    // 2. 关联到收货单
    err := c.receiptFileSrv.SaveReceiptFiles(ctx, receiptOrderUuid, fileUuids)
    if err != nil {
        response.Error(ctx, err.Error())
        return
    }
    
    response.Success(ctx, nil)
}
```

### 2. 签收单上传

```go
// 上传签收单PDF
func (c *Controller) UploadSignatureDocument(ctx *gin.Context) {
    receiptOrderUuid := utils.ParseUint64(ctx.Param("uuid"))
    
    // 1. 上传PDF文件
    file, _ := ctx.FormFile("signature_doc")
    fileUuid, _ := c.fileSrv.Upload(ctx, file)
    
    // 2. 关联到收货单
    c.receiptFileSrv.SaveReceiptFiles(ctx, receiptOrderUuid, []uint64{fileUuid})
}
```

### 3. 附件查看

```go
// 查看收货单附件
func (c *Controller) ViewReceiptFiles(ctx *gin.Context) {
    receiptOrderUuid := utils.ParseUint64(ctx.Param("uuid"))
    
    // 获取附件列表
    files, _ := c.receiptFileSrv.GetReceiptFiles(ctx, receiptOrderUuid)
    
    // 返回给前端显示
    response.Success(ctx, map[string]interface{}{
        "files": files,
    })
}
```

---

## 📚 相关文档

- [文件上传服务](upload_file.md) - 文件上传和管理
- [采购收货单服务](purchase_receipt_order.md) - 收货单管理

---

## 📊 服务特点总结

| 特点 | 说明 |
|-----|------|
| 简洁 | 183行代码，功能清晰 |
| 规范 | 附件数量限制、状态验证 |
| 灵活 | 支持批量上传、单个删除 |
| 安全 | 状态验证、权限控制 |
| 易用 | 接口简单，易于集成 |
| 完整 | 增删查改功能齐全 |

---

## 📄 更新日志

| 日期 | 版本 | 说明 |
|-----|------|-----|
| 2025-11-12 | 1.0 | 初始文档创建 |

---

## 👥 维护者

- 开发团队：Backend Team
- 文档维护：AI Assistant

---

**注意**: 本文档基于代码自动生成，如有代码变更，请及时更新文档。收货单附件服务是采购管理的重要组成部分，修改时需确保不影响现有业务流程。

