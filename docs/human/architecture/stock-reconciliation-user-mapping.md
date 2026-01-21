# 盘点单用户映射与操作记录方案

> 解决 TTPOS 与 ERPNext 系统用户不一致时的操作人员记录问题

---

## 一、问题分析

### 1.1 现状

**TTPOS 用户体系**：
- 使用 `Staff`（员工）表
- 标识字段：`uuid`（员工UUID）、`username`（用户名）、`real_name`（真实姓名）
- 操作记录：当前盘点单表没有记录操作人员字段

**ERPNext 用户体系**：
- 使用 `User`（用户）表
- 标识字段：`name`（用户邮箱/用户名）、`email`（邮箱）、`full_name`（全名）
- 操作记录：文档有 `owner`（所有者）、`modified_by`（修改者）字段

**问题**：
- TTPOS 和 ERPNext 的用户体系完全独立
- 用户标识方式不同（UUID vs Email）
- 需要在两个系统中都能正确记录和显示操作人员

---

## 二、解决方案设计

### 2.1 用户映射表设计

**创建用户映射表**：`ttpos_erp_user_mapping`

```sql
CREATE TABLE IF NOT EXISTS `ttpos_erp_user_mapping` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint NOT NULL DEFAULT 0 COMMENT '映射记录ID',
  `company_uuid` bigint NOT NULL DEFAULT 0 COMMENT '公司UUID',
  `ttpos_staff_uuid` bigint NOT NULL DEFAULT 0 COMMENT 'TTPOS员工UUID',
  `ttpos_username` varchar(255) NOT NULL DEFAULT '' COMMENT 'TTPOS用户名',
  `ttpos_real_name` varchar(255) NOT NULL DEFAULT '' COMMENT 'TTPOS真实姓名',
  `erp_user_email` varchar(255) NOT NULL DEFAULT '' COMMENT 'ERPNext用户邮箱（用户名）',
  `erp_user_name` varchar(255) NOT NULL DEFAULT '' COMMENT 'ERPNext用户全名',
  `mapping_type` tinyint NOT NULL DEFAULT 1 COMMENT '映射类型 1-手动映射 2-自动映射',
  `is_active` tinyint NOT NULL DEFAULT 1 COMMENT '是否启用 0-禁用 1-启用',
  `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '创建时间(时间戳)',
  `update_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '更新时间(时间戳)',
  `delete_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '删除时间(时间戳)',
  UNIQUE KEY `unique_uuid` (`uuid`),
  UNIQUE KEY `unique_company_staff` (`company_uuid`, `ttpos_staff_uuid`),
  UNIQUE KEY `unique_company_erp_user` (`company_uuid`, `erp_user_email`),
  KEY `idx_company_uuid` (`company_uuid`),
  KEY `idx_ttpos_staff_uuid` (`ttpos_staff_uuid`),
  KEY `idx_erp_user_email` (`erp_user_email`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='ERP用户映射表';
```

**字段说明**：
- `ttpos_staff_uuid`：TTPOS 员工 UUID
- `erp_user_email`：ERPNext 用户邮箱（作为用户名）
- `mapping_type`：映射类型（手动映射/自动映射）
- `is_active`：是否启用（可以临时禁用映射）

### 2.2 盘点单表字段调整

**添加操作人员字段**：

```sql
ALTER TABLE `ttpos_stock_reconciliation` 
ADD COLUMN `created_by_staff_uuid` bigint NOT NULL DEFAULT 0 COMMENT '创建人TTPOS员工UUID' AFTER `submit_time`,
ADD COLUMN `created_by_staff_name` varchar(255) NOT NULL DEFAULT '' COMMENT '创建人姓名（备份）' AFTER `created_by_staff_uuid`,
ADD COLUMN `created_by_erp_user` varchar(255) NOT NULL DEFAULT '' COMMENT '创建人ERPNext用户邮箱' AFTER `created_by_staff_name`,
ADD COLUMN `modified_by_staff_uuid` bigint NOT NULL DEFAULT 0 COMMENT '最后修改人TTPOS员工UUID' AFTER `created_by_erp_user`,
ADD COLUMN `modified_by_staff_name` varchar(255) NOT NULL DEFAULT '' COMMENT '最后修改人姓名（备份）' AFTER `modified_by_staff_uuid`,
ADD COLUMN `modified_by_erp_user` varchar(255) NOT NULL DEFAULT '' COMMENT '最后修改人ERPNext用户邮箱' AFTER `modified_by_staff_name`,
ADD COLUMN `submitted_by_staff_uuid` bigint NOT NULL DEFAULT 0 COMMENT '提交人TTPOS员工UUID' AFTER `modified_by_erp_user`,
ADD COLUMN `submitted_by_staff_name` varchar(255) NOT NULL DEFAULT '' COMMENT '提交人姓名（备份）' AFTER `submitted_by_staff_uuid`,
ADD COLUMN `submitted_by_erp_user` varchar(255) NOT NULL DEFAULT '' COMMENT '提交人ERPNext用户邮箱' AFTER `submitted_by_staff_name`,
ADD COLUMN `approved_by_staff_uuid` bigint NOT NULL DEFAULT 0 COMMENT '审核人TTPOS员工UUID' AFTER `submitted_by_erp_user`,
ADD COLUMN `approved_by_staff_name` varchar(255) NOT NULL DEFAULT '' COMMENT '审核人姓名（备份）' AFTER `approved_by_staff_uuid`,
ADD COLUMN `approved_by_erp_user` varchar(255) NOT NULL DEFAULT '' COMMENT '审核人ERPNext用户邮箱' AFTER `approved_by_staff_name`,
ADD INDEX `idx_created_by_staff_uuid` (`created_by_staff_uuid`),
ADD INDEX `idx_modified_by_staff_uuid` (`modified_by_staff_uuid`);
```

**字段说明**：
- 每个操作都记录 TTPOS 和 ERPNext 两套用户信息
- 备份姓名字段用于显示，避免关联查询
- ERPNext 用户邮箱字段用于同步到 ERPNext

### 2.3 操作记录表设计

**创建操作记录表**：`ttpos_stock_reconciliation_operation_log`

```sql
CREATE TABLE IF NOT EXISTS `ttpos_stock_reconciliation_operation_log` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
  `uuid` bigint NOT NULL DEFAULT 0 COMMENT '操作记录ID',
  `stock_reconciliation_uuid` bigint NOT NULL DEFAULT 0 COMMENT '盘点单UUID',
  `operation_type` tinyint NOT NULL DEFAULT 0 COMMENT '操作类型 1-创建 2-修改 3-提交 4-审核 5-驳回',
  `operation_source` tinyint NOT NULL DEFAULT 1 COMMENT '操作来源 1-TTPOS 2-ERPNext',
  `staff_uuid` bigint NOT NULL DEFAULT 0 COMMENT '操作人TTPOS员工UUID',
  `staff_name` varchar(255) NOT NULL DEFAULT '' COMMENT '操作人姓名',
  `erp_user_email` varchar(255) NOT NULL DEFAULT '' COMMENT '操作人ERPNext用户邮箱',
  `erp_user_name` varchar(255) NOT NULL DEFAULT '' COMMENT '操作人ERPNext用户全名',
  `operation_desc` text COMMENT '操作描述',
  `operation_data` json COMMENT '操作数据（JSON格式）',
  `create_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '操作时间(时间戳)',
  UNIQUE KEY `unique_uuid` (`uuid`),
  KEY `idx_stock_reconciliation_uuid` (`stock_reconciliation_uuid`),
  KEY `idx_operation_type` (`operation_type`),
  KEY `idx_staff_uuid` (`staff_uuid`),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='盘点单操作记录表';
```

**字段说明**：
- `operation_type`：操作类型（创建、修改、提交、审核、驳回）
- `operation_source`：操作来源（TTPOS/ERPNext）
- 记录两套用户信息，便于追溯
- `operation_data`：记录操作的具体数据（JSON格式）

---

## 三、用户映射机制

### 3.1 映射方式

#### 方式一：手动映射（推荐）

**流程**：
1. 管理员在 TTPOS 后台配置用户映射
2. 选择 TTPOS 员工和对应的 ERPNext 用户
3. 保存映射关系

**适用场景**：
- 用户体系相对稳定
- 需要精确控制映射关系
- 用户数量较少

#### 方式二：自动映射

**规则**：
- 根据用户名或邮箱自动匹配
- 如果 TTPOS 用户名与 ERPNext 邮箱一致，自动创建映射
- 如果匹配失败，提示管理员手动映射

**适用场景**：
- 用户体系一致（用户名=邮箱）
- 需要快速建立映射关系
- 用户数量较多

### 3.2 映射查询逻辑

```go
// 获取 ERPNext 用户邮箱
func GetERPUserEmail(ctx context.Context, staffUuid uint64) (string, error) {
    // 1. 查询用户映射表
    mapping := getUserMapping(ctx, staffUuid)
    if mapping != nil && mapping.IsActive {
        return mapping.ErpUserEmail, nil
    }
    
    // 2. 如果没有映射，返回默认值或错误
    return "", errors.New("用户未映射到ERPNext")
}

// 获取 TTPOS 员工信息
func GetTTPOSStaff(ctx context.Context, erpUserEmail string) (*model.Staff, error) {
    // 1. 查询用户映射表
    mapping := getUserMappingByERPEmail(ctx, erpUserEmail)
    if mapping != nil && mapping.IsActive {
        return getStaffByUuid(ctx, mapping.TtposStaffUuid)
    }
    
    // 2. 如果没有映射，返回空或创建临时记录
    return nil, errors.New("ERPNext用户未映射到TTPOS")
}
```

---

## 四、操作记录实现

### 4.1 TTPOS 操作时记录

**创建盘点单**：
```go
func SaveStockReconciliation(ctx context.Context, req req.StockReconciliationSaveReq) {
    // 1. 获取当前登录用户
    staff := ctx.GetStaff()
    
    // 2. 获取 ERPNext 用户邮箱
    erpUserEmail, err := GetERPUserEmail(ctx, staff.Uuid)
    if err != nil {
        // 如果没有映射，使用默认用户或报错
        erpUserEmail = getDefaultERPUser(ctx)
    }
    
    // 3. 调用 ERPNext API 创建盘点单（传递 erpUserEmail）
    erpResp, err := erpSrv.CreateStockReconciliation(ctx, &stock.SaveStockReconciliationReq{
        Owner: erpUserEmail, // 设置所有者
        // ... 其他字段
    })
    
    // 4. 保存到 TTPOS 数据库
    stockReconciliation := &model.StockReconciliation{
        CreatedByStaffUuid: staff.Uuid,
        CreatedByStaffName: staff.GetUserName(),
        CreatedByErpUser: erpUserEmail,
        // ... 其他字段
    }
    
    // 5. 记录操作日志
    recordOperationLog(ctx, &OperationLog{
        StockReconciliationUuid: stockReconciliation.Uuid,
        OperationType: constant.OperationTypeCreate,
        OperationSource: constant.OperationSourceTTPOS,
        StaffUuid: staff.Uuid,
        StaffName: staff.GetUserName(),
        ErpUserEmail: erpUserEmail,
    })
}
```

**修改盘点单**：
```go
func UpdateStockReconciliation(ctx context.Context, req req.StockReconciliationSaveReq) {
    // 1. 获取当前登录用户
    staff := ctx.GetStaff()
    
    // 2. 获取 ERPNext 用户邮箱
    erpUserEmail, _ := GetERPUserEmail(ctx, staff.Uuid)
    
    // 3. 调用 ERPNext API 更新盘点单
    erpSrv.UpdateStockReconciliation(ctx, &stock.UpdateStockReconciliationReq{
        ModifiedBy: erpUserEmail, // 设置修改者
        // ... 其他字段
    })
    
    // 4. 更新 TTPOS 数据库
    stockReconciliation.ModifiedByStaffUuid = staff.Uuid
    stockReconciliation.ModifiedByStaffName = staff.GetUserName()
    stockReconciliation.ModifiedByErpUser = erpUserEmail
    
    // 5. 记录操作日志
    recordOperationLog(ctx, &OperationLog{
        OperationType: constant.OperationTypeUpdate,
        // ... 其他字段
    })
}
```

**提交盘点单**：
```go
func SubmitStockReconciliation(ctx context.Context, req req.StockReconciliationSubmitReq) {
    // 1. 获取当前登录用户
    staff := ctx.GetStaff()
    
    // 2. 获取 ERPNext 用户邮箱
    erpUserEmail, _ := GetERPUserEmail(ctx, staff.Uuid)
    
    // 3. 调用 ERPNext API 提交盘点单
    erpSrv.SubmitStockReconciliation(ctx, &stock.SubmitStockReconciliationReq{
        // ... 字段
    })
    
    // 4. 更新 TTPOS 数据库
    stockReconciliation.SubmittedByStaffUuid = staff.Uuid
    stockReconciliation.SubmittedByStaffName = staff.GetUserName()
    stockReconciliation.SubmittedByErpUser = erpUserEmail
    
    // 5. 记录操作日志
    recordOperationLog(ctx, &OperationLog{
        OperationType: constant.OperationTypeSubmit,
        // ... 其他字段
    })
}
```

**审核盘点单**：
```go
func ApproveStockReconciliation(ctx context.Context, req req.StockReconciliationApproveReq) {
    // 1. 获取当前登录用户
    staff := ctx.GetStaff()
    
    // 2. 获取 ERPNext 用户邮箱
    erpUserEmail, _ := GetERPUserEmail(ctx, staff.Uuid)
    
    // 3. 调用 ERPNext API 提交盘点单（审核）
    erpSrv.ApproveStockReconciliation(ctx, &stock.SubmitStockReconciliationReq{
        // ... 字段
    })
    
    // 4. 更新 TTPOS 数据库
    stockReconciliation.ApprovedByStaffUuid = staff.Uuid
    stockReconciliation.ApprovedByStaffName = staff.GetUserName()
    stockReconciliation.ApprovedByErpUser = erpUserEmail
    
    // 5. 记录操作日志
    recordOperationLog(ctx, &OperationLog{
        OperationType: constant.OperationTypeApprove,
        // ... 其他字段
    })
}
```

### 4.2 ERPNext 操作时记录

**Webhook 接收处理**：
```go
func HandleERPWebhook(c *gin.Context) {
    var webhookReq ERPWebhookReq
    if err := c.ShouldBindJSON(&webhookReq); err != nil {
        return err
    }
    
    // 1. 从 Webhook 数据中获取 ERPNext 用户信息
    erpUserEmail := webhookReq.Data.ModifiedBy // 或 Owner
    erpUserName := webhookReq.Data.ModifiedByName
    
    // 2. 查询 TTPOS 员工映射
    staff, err := GetTTPOSStaff(ctx, erpUserEmail)
    if err != nil {
        // 如果没有映射，创建临时记录或使用默认值
        staff = &model.Staff{
            Uuid: 0,
            RealName: erpUserName, // 使用 ERPNext 用户名
        }
    }
    
    // 3. 同步数据到 TTPOS
    syncStockReconciliationFromERP(ctx, webhookReq.Data)
    
    // 4. 更新操作人员字段
    stockReconciliation.ModifiedByStaffUuid = staff.Uuid
    stockReconciliation.ModifiedByStaffName = staff.GetUserName()
    stockReconciliation.ModifiedByErpUser = erpUserEmail
    
    // 5. 记录操作日志
    recordOperationLog(ctx, &OperationLog{
        OperationType: getOperationType(webhookReq.Event),
        OperationSource: constant.OperationSourceERPNext,
        StaffUuid: staff.Uuid,
        StaffName: staff.GetUserName(),
        ErpUserEmail: erpUserEmail,
        ErpUserName: erpUserName,
    })
}
```

---

## 五、数据展示

### 5.1 前端展示逻辑

**显示操作人员**：
```javascript
// 显示创建人
function displayCreator(stockReconciliation) {
    // 优先显示 TTPOS 员工姓名
    if (stockReconciliation.created_by_staff_name) {
        return stockReconciliation.created_by_staff_name;
    }
    
    // 如果没有，显示 ERPNext 用户名
    if (stockReconciliation.created_by_erp_user) {
        return stockReconciliation.created_by_erp_user;
    }
    
    return '未知';
}

// 显示操作来源标识
function displayOperationSource(stockReconciliation) {
    if (stockReconciliation.created_by_staff_uuid > 0) {
        return 'TTPOS';
    } else if (stockReconciliation.created_by_erp_user) {
        return 'ERPNext';
    }
    return '';
}
```

**操作记录列表**：
```javascript
// 显示操作记录
function displayOperationLog(log) {
    return {
        operationType: getOperationTypeName(log.operation_type),
        operator: log.staff_name || log.erp_user_name,
        source: log.operation_source === 1 ? 'TTPOS' : 'ERPNext',
        time: formatTime(log.create_time),
        desc: log.operation_desc
    };
}
```

### 5.2 数据同步时的处理

**从 ERPNext 同步时**：
```go
func syncStockReconciliationFromERP(ctx context.Context, erpData ERPStockReconciliationData) {
    // 1. 查询或创建盘点单记录
    stockReconciliation := getOrCreateStockReconciliation(erpData.Name)
    
    // 2. 同步操作人员信息
    // 创建人
    if erpData.Owner != "" {
        staff, _ := GetTTPOSStaff(ctx, erpData.Owner)
        stockReconciliation.CreatedByErpUser = erpData.Owner
        if staff != nil {
            stockReconciliation.CreatedByStaffUuid = staff.Uuid
            stockReconciliation.CreatedByStaffName = staff.GetUserName()
        }
    }
    
    // 修改人
    if erpData.ModifiedBy != "" {
        staff, _ := GetTTPOSStaff(ctx, erpData.ModifiedBy)
        stockReconciliation.ModifiedByErpUser = erpData.ModifiedBy
        if staff != nil {
            stockReconciliation.ModifiedByStaffUuid = staff.Uuid
            stockReconciliation.ModifiedByStaffName = staff.GetUserName()
        }
    }
    
    // 3. 保存数据
    saveStockReconciliation(stockReconciliation)
}
```

---

## 六、用户映射管理

### 6.1 映射配置界面

**功能需求**：
1. 列表展示：显示所有用户映射关系
2. 新增映射：选择 TTPOS 员工和 ERPNext 用户
3. 编辑映射：修改映射关系
4. 删除映射：删除映射关系（软删除）
5. 启用/禁用：临时禁用映射关系

**界面设计**：
```
用户映射管理
├─ TTPOS员工 | ERPNext用户 | 映射类型 | 状态 | 操作
├─ 张三      | zhang@example.com | 手动映射 | 启用 | 编辑/禁用
├─ 李四      | li@example.com | 自动映射 | 启用 | 编辑/禁用
└─ ...
```

### 6.2 映射 API 接口

**接口设计**：
```go
// 获取用户映射列表
GET /api/v1/admin/erp_user_mapping/list
Query: company_uuid, page_no, page_size

// 创建用户映射
POST /api/v1/admin/erp_user_mapping/create
Body: {
    "ttpos_staff_uuid": 123,
    "erp_user_email": "zhang@example.com",
    "mapping_type": 1
}

// 更新用户映射
POST /api/v1/admin/erp_user_mapping/update
Body: {
    "uuid": 456,
    "erp_user_email": "zhang.new@example.com"
}

// 删除用户映射
POST /api/v1/admin/erp_user_mapping/delete
Body: {
    "uuid": 456
}

// 启用/禁用映射
POST /api/v1/admin/erp_user_mapping/toggle
Body: {
    "uuid": 456,
    "is_active": 1
}
```

---

## 七、特殊情况处理

### 7.1 用户未映射

**场景**：TTPOS 用户操作时，没有对应的 ERPNext 用户映射

**处理方案**：
1. **使用默认用户**：使用系统配置的默认 ERPNext 用户
2. **自动创建映射**：如果 ERPNext 中存在同名用户，自动创建映射
3. **提示错误**：要求管理员先配置用户映射

**推荐方案**：使用默认用户 + 记录日志

```go
func GetERPUserEmail(ctx context.Context, staffUuid uint64) (string, error) {
    mapping := getUserMapping(ctx, staffUuid)
    if mapping != nil && mapping.IsActive {
        return mapping.ErpUserEmail, nil
    }
    
    // 使用默认用户
    defaultUser := getDefaultERPUser(ctx)
    if defaultUser != "" {
        logger.Logger.Warn("用户未映射，使用默认ERP用户", 
            zap.Uint64("staff_uuid", staffUuid),
            zap.String("default_user", defaultUser))
        return defaultUser, nil
    }
    
    return "", errors.New("用户未映射且无默认用户")
}
```

### 7.2 ERPNext 用户不存在

**场景**：ERPNext 操作时，用户邮箱在 TTPOS 中找不到映射

**处理方案**：
1. **创建临时记录**：使用 ERPNext 用户名创建临时显示记录
2. **提示映射**：记录日志，提示管理员配置映射
3. **忽略处理**：只记录 ERPNext 用户信息，不关联 TTPOS 员工

**推荐方案**：创建临时记录 + 记录日志

```go
func GetTTPOSStaff(ctx context.Context, erpUserEmail string) (*model.Staff, error) {
    mapping := getUserMappingByERPEmail(ctx, erpUserEmail)
    if mapping != nil && mapping.IsActive {
        return getStaffByUuid(ctx, mapping.TtposStaffUuid)
    }
    
    // 创建临时记录（仅用于显示）
    logger.Logger.Warn("ERPNext用户未映射到TTPOS", 
        zap.String("erp_user_email", erpUserEmail))
    
    return &model.Staff{
        Uuid: 0,
        RealName: erpUserEmail, // 使用邮箱作为显示名
    }, nil
}
```

### 7.3 用户映射变更

**场景**：用户映射关系变更后，历史数据的处理

**处理方案**：
1. **保留历史数据**：历史操作记录中的用户信息保持不变
2. **更新当前数据**：新操作使用新的映射关系
3. **数据迁移**：可选的数据迁移脚本，更新历史数据

**推荐方案**：保留历史数据，新操作使用新映射

---

## 八、实施步骤

### 8.1 数据库调整

- [ ] 创建用户映射表
- [ ] 添加盘点单操作人员字段
- [ ] 创建操作记录表
- [ ] 创建索引

### 8.2 代码实现

- [ ] 实现用户映射服务
- [ ] 实现操作记录服务
- [ ] 调整盘点单服务，添加操作人员记录
- [ ] 实现 Webhook 处理，记录 ERPNext 操作人员

### 8.3 管理界面

- [ ] 实现用户映射管理界面
- [ ] 实现操作记录查询界面
- [ ] 实现操作人员显示

### 8.4 数据迁移

- [ ] 迁移现有盘点单数据（补充操作人员信息）
- [ ] 创建初始用户映射关系

---

## 九、总结

### 9.1 核心方案

1. **用户映射表**：建立 TTPOS 员工与 ERPNext 用户的映射关系
2. **双重记录**：在盘点单表中同时记录 TTPOS 和 ERPNext 用户信息
3. **操作日志**：详细记录每次操作的用户信息和来源
4. **灵活展示**：前端优先显示 TTPOS 用户，无映射时显示 ERPNext 用户

### 9.2 优势

- ✅ **完整追溯**：可以追溯到具体操作人员
- ✅ **双向支持**：支持 TTPOS 和 ERPNext 两端操作
- ✅ **灵活映射**：支持手动和自动映射
- ✅ **历史保留**：历史数据不受映射变更影响

### 9.3 注意事项

- ⚠️ **映射配置**：需要管理员配置用户映射关系
- ⚠️ **默认用户**：需要配置默认 ERPNext 用户
- ⚠️ **数据一致性**：确保操作人员信息同步正确

---

**文档版本**：v1.0  
**创建时间**：2025-01-16  
**维护者**：TTPOS Team


