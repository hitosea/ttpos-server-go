# story-shop-transfer-receive-attachment 任务清单

## 📊 进度总览

| 项目 | 数值 |
|------|------|
| **总 SP** | 5 |
| **总任务数** | 12 |
| **已完成** | 8 |
| **完成率** | 67% |

---

## Phase 1: 数据层（Model + Repository + 迁移）

### 1.1 创建数据库迁移文件

| 项目 | 内容 |
|------|------|
| **File** | `admin/database/migrations/{timestamp}_create_transfer_order_file.php` |
| **Purpose** | 创建 `ttpos_transfer_order_file` 表 |
| **Requirements** | Req 1, 2 |
| **Leverage** | 参考现有迁移文件格式 |

**迁移内容**:
```php
Schema::create('transfer_order_file', function (Blueprint $table) {
    $table->id();
    $table->unsignedBigInteger('uuid')->unique();
    $table->unsignedBigInteger('transfer_order_uuid')->index();
    $table->unsignedBigInteger('file_uuid')->index();
    $table->integer('sort_order')->default(0);
    $table->integer('create_time')->default(0);
    $table->integer('update_time')->default(0);
    $table->integer('delete_time')->default(0);
});
```

- [x] 完成

### 1.2 更新 shop_01.sql 种子文件

| 项目 | 内容 |
|------|------|
| **File** | `admin/database/seeds/shop_01.sql` |
| **Purpose** | 同步新表结构到种子文件 |
| **Requirements** | 数据库规范 |

- [x] 完成

### 1.3 创建 TransferOrderFile Model

| 项目 | 内容 |
|------|------|
| **File** | `main/app/model/transfer_order_file.go` |
| **Purpose** | 调拨单附件关联模型 |
| **Requirements** | Req 1 |
| **Leverage** | 复制 `model/purchase_receipt_file.go`，修改字段名 |

**关键代码**:
```go
type TransferOrderFile struct {
    BaseModel
    TransferOrderUuid uint64 `gorm:"..."`
    FileUuid          uint64 `gorm:"..."`
    SortOrder         int    `gorm:"..."`

    TransferOrder *TransferOrder `gorm:"foreignKey:TransferOrderUuid;references:Uuid"`
    File          *File          `gorm:"foreignKey:FileUuid;references:Uuid"`
}
```

- [x] 完成

### 1.4 创建 TransferOrderFile Repository

| 项目 | 内容 |
|------|------|
| **File** | `main/app/repository/transfer_order_file.go` |
| **Purpose** | 调拨单附件数据访问层 |
| **Requirements** | Req 1, 4, 5 |
| **Leverage** | 复制 `repository/purchase_receipt_file.go`，修改方法名 |

**接口方法**:
- `BatchCreate(files)`
- `GetByTransferOrderUuidWithFiles(uuid)`
- `DeleteByTransferOrderUuid(uuid)`
- `DeleteByFileUuidAndTransferOrderUuid(fileUuid, orderUuid)`
- `CountByTransferOrderUuid(uuid)`

- [x] 完成

---

## Phase 2: 服务层（Service）

### 2.1 创建 TransferOrderFile Service

| 项目 | 内容 |
|------|------|
| **File** | `main/app/service/transfer_order_file.go` |
| **Purpose** | 调拨单附件业务逻辑 |
| **Requirements** | Req 1, 2, 3, 4, 5 |
| **Leverage** | 复制 `service/purchase_receipt_file.go`，修改状态常量 |

**接口方法**:
- `SaveTransferOrderFiles(ctx, orderUuid, fileUuids)`
- `GetTransferOrderFiles(ctx, orderUuid)`
- `DeleteTransferOrderFile(ctx, fileUuid, orderUuid)`
- `DeleteAllTransferOrderFiles(ctx, orderUuid)`
- `ValidateFileLimit(ctx, orderUuid, count)`
- `ValidateTransferOrderStatus(ctx, orderUuid)`

**关键改动**:
```go
// 状态常量从 ReceiptOrderStatusPending 改为 TransferOrderStatusReceiving
if order.Status != constant.TransferOrderStatusReceiving {
    return errors.New("仅待收货状态的调拨单可以修改附件")
}
```

- [x] 完成

### 2.2 修改 TransferOrder Service 集成附件逻辑

| 项目 | 内容 |
|------|------|
| **File** | `main/app/service/transfer_order/transfer_order.go` |
| **Purpose** | 在收货流程中集成附件保存 |
| **Requirements** | Req 1, 3 |

**集成点**:
1. **保存收货** - 调用 `SaveTransferOrderFiles`（可选）
2. **确认收货** - 校验附件必填 + 调用 `SaveTransferOrderFiles`

**关键代码**:
```go
// 确认收货时校验附件
if len(req.FileUuids) == 0 {
    return errors.New("请上传相关附件后确定收货")
}

// 保存附件（事务外）
if len(req.FileUuids) > 0 {
    err = s.transferOrderFileSrv.SaveTransferOrderFiles(ctx, orderUuid, req.FileUuids)
    if err != nil {
        logger.Logger.Warn("保存调拨单附件失败", zap.Error(err))
    }
}
```

- [x] 完成

---

## Phase 3: API 层集成

### 3.1 修改请求 DTO

| 项目 | 内容 |
|------|------|
| **File** | `main/app/dto/req/transfer_order.go` |
| **Purpose** | 收货请求增加 FileUuids 字段 |
| **Requirements** | Req 1, 3 |

**新增字段**:
```go
type TransferOrderConfirmReceiveReq struct {
    // ... 现有字段
    FileUuids []uint64 `json:"file_uuids"` // 附件UUID列表
}

type TransferOrderSaveReceiveReq struct {
    // ... 现有字段
    FileUuids []uint64 `json:"file_uuids"` // 附件UUID列表（可选）
}

// 新增：删除附件请求
type DeleteTransferOrderFileReq struct {
    TransferOrderUuid uint64 `json:"transfer_order_uuid" binding:"required"`
    FileUuid          uint64 `json:"file_uuid" binding:"required"`
}
```

- [x] 完成

### 3.2 修改响应 DTO

| 项目 | 内容 |
|------|------|
| **File** | `main/app/dto/resp/transfer_order.go` |
| **Purpose** | 详情响应增加 Files 字段 |
| **Requirements** | Req 4, 5 |

**新增结构**:
```go
// TransferOrderFileInfo 调拨单附件信息
type TransferOrderFileInfo struct {
    FileUuid   uint64 `json:"file_uuid"`
    FileName   string `json:"file_name"`
    FileSize   int64  `json:"file_size"`
    FileType   string `json:"file_type"`
    Extension  string `json:"extension"`
    FilePath   string `json:"file_path"`
    SortOrder  int    `json:"sort_order"`
    CreateTime int    `json:"create_time"`
}

// 详情响应增加 Files 字段
type TransferOrderDetailResp struct {
    // ... 现有字段
    Files []TransferOrderFileInfo `json:"files"`
}
```

- [x] 完成

### 3.3 新增删除附件 API

| 项目 | 内容 |
|------|------|
| **File** | `main/app/api/v1/shop/shop_transfer.go` |
| **Purpose** | 新增 DELETE /shop/transfer/file 接口 |
| **Requirements** | Req 4 |

**路由注册**:
```go
shopTransfer.DELETE("/file", handler.DeleteTransferOrderFile)
```

**Handler 实现**:
```go
func (h *TransferHandler) DeleteTransferOrderFile(c *gin.Context) {
    var req req.DeleteTransferOrderFileReq
    if err := c.ShouldBindJSON(&req); err != nil {
        // 错误处理
    }

    err := h.transferOrderFileSrv.DeleteTransferOrderFile(ctx, req.FileUuid, req.TransferOrderUuid)
    // 响应处理
}
```

- [x] 完成

### 3.4 修改详情查询接口

| 项目 | 内容 |
|------|------|
| **File** | `main/app/api/v1/shop/shop_transfer.go` |
| **Purpose** | 详情响应中返回附件列表 |
| **Requirements** | Req 4, 5 |

**修改点**:
在构建详情响应时，调用 `GetTransferOrderFiles` 获取附件列表：
```go
files, _ := h.transferOrderFileSrv.GetTransferOrderFiles(ctx, order.Uuid)
resp.Files = files
```

- [x] 完成

---

## Phase 4: 测试

### 4.1 编写 Service 层单元测试

| 项目 | 内容 |
|------|------|
| **File** | `main/app/service/transfer_order_file_test.go` |
| **Purpose** | 单元测试覆盖核心逻辑 |
| **Requirements** | 覆盖率 ≥ 80% |

**测试用例**:
- `TestSaveTransferOrderFiles_Success`
- `TestSaveTransferOrderFiles_ExceedLimit`
- `TestValidateTransferOrderStatus_NotReceiving`
- `TestDeleteTransferOrderFile_Success`

- [ ] 完成

### 4.2 编写 Repository 层单元测试

| 项目 | 内容 |
|------|------|
| **File** | `main/app/repository/transfer_order_file_test.go` |
| **Purpose** | 数据访问层测试 |
| **Requirements** | 覆盖率 ≥ 70% |

- [ ] 完成

### 4.3 集成测试验证

| 项目 | 内容 |
|------|------|
| **Purpose** | 端到端流程验证 |
| **Requirements** | 所有验收标准 |

**测试场景**:
1. 上传附件 → 保存收货 → 附件关联成功
2. 上传附件 → 确认收货 → 状态变更 + 附件关联成功
3. 无附件确认收货 → 返回错误提示
4. 已完成订单删除附件 → 返回错误提示
5. 附件预览/下载 → 返回正确 URL

- [ ] 完成

---

## 提交清单

### 代码质量
- [x] `go mod tidy` 执行
- [x] `go fmt ./...` 执行
- [x] `go vet ./...` 通过 (pre-existing issues in other files)
- [ ] 测试通过: `go test ./...`

### 功能完整性
- [ ] 所有验收标准通过
- [ ] API 响应格式正确（data 为对象，切片使用 make 初始化）
- [ ] 错误提示使用多语言 key

### 迁移同步
- [x] 迁移文件已创建
- [x] shop_01.sql 已更新
- [ ] 迁移已执行验证

### 文档
- [ ] requirements.md 状态更新为"开发中"
- [ ] 完成后更新为"已完成"

---

**版本**: v1.0.0
**创建日期**: 2026-01-28
