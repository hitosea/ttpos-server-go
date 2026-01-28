# 调拨单收货附件上传功能 需求文档

## 📋 基本信息

| 项目              | 内容                                                                 |
| ----------------- | -------------------------------------------------------------------- |
| **Spec ID**       | story-shop-transfer-receive-attachment                               |
| **来源 Proposal** | [shop-transfer-receive-attachment](../../../team/proposals/2026-01/shop-transfer-receive-attachment.md) |
| **关联任务**      | DooTask #39043                                                       |
| **创建日期**      | 2026-01-28                                                           |
| **负责人**        | xiezhihuan                                                           |
| **目标版本**      | v2.16.0                                                              |

## 📋 审核状态

| 项目         | 内容       |
| ------------ | ---------- |
| **审核状态** | 开发中     |
| **审核人**   | -          |
| **审核日期** | 2026-01-28 |

---

## 📝 用户故事

**作为** 店长/门店管理员
**我想** 在调拨单收货时上传签收单、送货单等附件
**以便于** 留存收货凭证，满足财务审计要求，减少货物纠纷

---

## 功能需求

### Requirement 1: 附件上传功能

**用户故事**: 作为店长，我想在调拨单收货时上传附件，以便于留存收货凭证

#### 验收标准

1. **WHEN** 用户在待收货详情页点击上传 **THEN** 系统 **SHALL** 弹出上传方式选择（拍照/相册/文件管理器）
2. **WHEN** 用户选择上传方式后 **THEN** 系统 **SHALL** 跳转对应手机页面（相机/相册/文件管理）
3. **WHEN** 用户选定文件后 **THEN** 系统 **SHALL** 显示上传进度条（百分比 + 进度动画）
4. **WHEN** 上传成功 **THEN** 系统 **SHALL** 实时刷新附件列表，显示新增附件
5. **WHEN** 上传失败 **THEN** 系统 **SHALL** 弹出明确提示（格式不支持/文件过大/网络异常）

### Requirement 2: 附件格式与限制

**用户故事**: 作为店长，我想上传多种格式的附件，以便于满足不同凭证类型需求

#### 验收标准

1. **WHEN** 用户上传文档类文件 **THEN** 系统 **SHALL** 支持 PDF、Word (.doc/.docx)、Excel (.xlsx/.xls) 格式
2. **WHEN** 用户上传图片类文件 **THEN** 系统 **SHALL** 支持 JPG、PNG、GIF 格式
3. **WHEN** 用户上传的单个文件超过 20MB **THEN** 系统 **SHALL** 提示"文件过大"并阻止上传
4. **WHEN** 用户已上传 10 个附件后继续上传 **THEN** 系统 **SHALL** 提示"最多支持10个附件"
5. **WHEN** 用户上传不支持的格式 **THEN** 系统 **SHALL** 提示"格式不支持"

### Requirement 3: 附件必填校验

**用户故事**: 作为财务人员，我想确保收货时必须上传凭证，以便于满足审计合规要求

#### 验收标准

1. **WHEN** 用户点击【保存】操作 **THEN** 系统 **SHALL** 允许无附件保存（附件非必填）
2. **WHEN** 用户点击【确定收货】且未上传附件 **THEN** 系统 **SHALL** toast 提示"请上传相关附件后确定收货"并阻止操作
3. **WHEN** 用户使用旧版本 App（无附件功能） **THEN** 系统 **SHALL** 提示"您的软件版本过低，请升级后再试"

### Requirement 4: 附件预览与下载

**用户故事**: 作为运营人员，我想查看和下载收货附件，以便于监督调拨执行情况

#### 验收标准

1. **WHEN** 用户点击预览图片类附件 **THEN** 系统 **SHALL** 在应用内弹框打开预览
2. **WHEN** 用户点击预览文档类附件 **THEN** 系统 **SHALL** 下载到本地后打开
3. **WHEN** 用户点击下载按钮 **THEN** 系统 **SHALL** 将附件保存到本地
4. **WHEN** 用户点击删除已上传的附件 **THEN** 系统 **SHALL** 从列表中移除该附件

### Requirement 5: 状态控制

**用户故事**: 作为店长，我想在已完成订单中查看附件但不能修改，以便于保证凭证完整性

#### 验收标准

1. **WHEN** 订单状态为待收货 **THEN** 系统 **SHALL** 允许上传、删除、预览、下载附件
2. **WHEN** 订单状态为已完成 **THEN** 系统 **SHALL** 仅允许预览和下载附件，隐藏上传和删除功能

---

## 非功能需求

### 测试要求

- [ ] 单元测试覆盖率 ≥ 80%
- [ ] Service 层测试覆盖核心逻辑
- [ ] Repository 层测试覆盖数据操作

### 平台兼容性

- [ ] Android 5.0+
- [ ] iOS 12.0+

### 国际化要求

- [ ] 所有提示文案使用多语言 key

---

## 约束条件

### 技术约束（后端）

- Go 版本: 1.23+
- 框架: Gin + GORM
- 分层架构: API → Service → Repository → Model
- 必须遵循 CLAUDE.md 和 .cursor/rules/go-main.mdc 规范
- 复用采购收货附件逻辑（PurchaseReceiptFile）模式

### 技术约束（前端）

- 复用现有附件上传组件
- 与采购收货附件上传交互保持一致

### 资源约束

- Story Point: 3 (必须 ≤ 5)

---

## 风险和缓解

### 风险 1: 调拨单状态常量不明确

**影响**: 中
**缓解措施**: 开发前确认 `constant/transfer_order.go` 中的状态定义，明确"待收货"和"已完成"对应的常量值

### 风险 2: 版本兼容性

**影响**: 低
**缓解措施**: 后端增加版本检测逻辑，旧版本 App 返回友好提示

---

## 技术参考

### 可复用代码清单

| 采购收货组件 | 调拨单对应组件 |
|-------------|---------------|
| `model/purchase_receipt_file.go` | `model/transfer_order_file.go` |
| `repository/purchase_receipt_file.go` | `repository/transfer_order_file.go` |
| `service/purchase_receipt_file.go` | `service/transfer_order_file.go` |

### 直接复用

- `service/upload_file.go` - `UploadDocument()` 文件上传接口
- `dto/resp/purchase_receipt.go` - `ReceiptFileInfo` 附件响应结构

### 新建数据库表

```sql
CREATE TABLE ttpos_transfer_order_file (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    uuid BIGINT UNSIGNED NOT NULL UNIQUE,
    transfer_order_uuid BIGINT UNSIGNED NOT NULL,
    file_uuid BIGINT UNSIGNED NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    create_time INT NOT NULL DEFAULT 0,
    update_time INT NOT NULL DEFAULT 0,
    delete_time INT NOT NULL DEFAULT 0,
    INDEX idx_transfer_order_uuid (transfer_order_uuid),
    INDEX idx_file_uuid (file_uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

---

**版本**: v1.0.0
**创建日期**: 2026-01-28
