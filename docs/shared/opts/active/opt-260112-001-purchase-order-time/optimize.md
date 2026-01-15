# 优化文档：品采单待提交状态日期显示错误

> **DooTask**: #38775  
> **创建时间**: 2026-01-12  
> **优化类型**: 缺陷修复  
> **优先级**: P2 - 重要  
> **影响范围**: 新管理端 - 品采单列表

---

## 📋 问题描述

### 现象
待提交状态的品采单，单据日期显示为 "1970-01-01"（时间戳 0 对应的日期）。

### 影响范围
- **模块**: Go Main 模块
- **终端**: shop（店长/运营人员）
- **功能**: 品采单列表展示
- **状态**: 仅影响"待提交"状态的采购单

### 用户体验影响
- 待提交的采购单日期显示异常，影响用户判断单据创建时间
- 容易与历史单据混淆

---

## 🔍 问题分析

### 1. 数据库设计

品采单表 `ttpos_purchase_order` 的 `order_time` 字段定义：

```sql
`order_time` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '单据日期，采购单提交的时间（时间戳）'
```

**关键点**：
- `order_time` 表示"单据日期"，是采购单**提交**的时间
- 默认值为 0
- 只有在提交时才会设置为当前时间戳

### 2. 业务逻辑

根据代码分析：

```21:21:main/app/model/purchase_order.go
	OrderTime               int64   `gorm:"column:order_time;type:int(10) unsigned;not null;default:0;comment:单据日期，采购单提交的时间（时间戳）" json:"order_time"`
```

```82:95:main/app/model/purchase_order.go
// GetStatusText 获取状态文本
func (po *PurchaseOrder) GetStatusText() string {
	statusMap := map[int]string{
		0: "待提交",
		1: "待审核",
		2: "已通过",
		3: "已驳回",
		4: "部分收货",
		5: "全部收货",
	}
	if text, exists := statusMap[po.Status]; exists {
		return text
	}
	return "未知状态"
}
```

**状态流转**：
1. **Status=0（待提交）**: 创建草稿，`order_time` 为 0
2. **Status=1（待审核）**: 提交采购单，设置 `order_time` 为当前时间
3. 后续状态：`order_time` 保持不变

### 3. 响应数据结构

```18:18:main/app/dto/resp/purchase_order.go
	OrderTime         int64              `json:"order_time"`          // 单据日期
```

### 4. 数据转换逻辑

Service 层直接使用 copier 复制数据，未对 `order_time=0` 的情况做特殊处理：

```133:142:main/app/service/purchase_order/purchase_order.go
	// 转换响应数据
	listResp := make([]*resp.PurchaseOrderInfo, 0, len(purchaseOrders))
	for _, po := range purchaseOrders {
		poInfo := &resp.PurchaseOrderInfo{}
		if err := copier.Copy(poInfo, &po); err != nil {
			continue
		}
		poInfo.ReceiptProgress = fmt.Sprintf("%.0f%%", po.GetReceiptProgress())
		listResp = append(listResp, poInfo)
	}
```

### 5. 问题根源

**这不是 Bug，是业务逻辑**：
- 采购单在草稿状态时，`order_time` 理应为 0（未提交）
- 前端直接将 `order_time=0` 转换为日期，显示为 "1970-01-01"
- 缺少对"未提交"状态的特殊处理

---

## 💡 解决方案

### 方案选择

考虑三种方案：

| 方案 | 描述 | 优点 | 缺点 | 推荐 |
|------|------|------|------|------|
| **方案1** | 后端返回 `null` | 前端无需改动，兼容性好 | 需要改响应字段类型 | ⭐⭐⭐⭐⭐ |
| **方案2** | 后端返回 `-1` | 改动最小 | 前端需要特殊判断 | ⭐⭐⭐ |
| **方案3** | 前端处理 | 后端无需改动 | 需要前端改动，治标不治本 | ⭐⭐ |

**推荐方案1**：后端返回 `null`，前端显示为"未提交"或"-"。

### 方案1：后端返回 null（推荐）

#### 实现步骤

**Step 1: 修改响应结构体**

将 `order_time` 改为指针类型，允许返回 `null`：

```go
// PurchaseOrderInfo 采购订单信息
type PurchaseOrderInfo struct {
	Uuid              uint64             `json:"uuid"`                // 采购订单ID
	OrderNo           string             `json:"order_no"`            // 申请单编号（单据编号）
	ErpOrderNo        string             `json:"erp_order_no"`        // ERP申请单编号（采购单号）
	Status            int                `json:"status"`              // 状态 0-待提交 1-待审核 2-已通过 3-已驳回 4-全部收货(完成) 5-待总部审核
	HeadquarterStatus int                `json:"headquarter_status"`  // V2.6 总部状态 0-待提交 1-待审核 2-已通过 3-已驳回 4-全部收货(完成)
	OrderTime         *int64             `json:"order_time"`          // 单据日期（未提交时为null）
	Num               int                `json:"num"`                 // 物品数量
	OrderType         int                `json:"order_type"`          // 申请类型
	SupplierName      string             `json:"supplier_name"`       // 供应商名称
	SupplierErpCode   string             `json:"supplier_erp_code"`   // 供应商编码
	PurchaseType      int                `json:"purchase_type"`       // V2.6 采购类型 1-外部采购 2-内部采购
	WarehouseErpCode  string             `json:"warehouse_erp_code"`  // V2.6 仓库编码
	WarehouseName     dto.LocaleResponse `json:"warehouse_name"`      // V2.6 仓库名称
	ExpectArrivalTime int64              `json:"expect_arrival_time"` // 期望到货日期
	ReceiptProgress   string             `json:"receipt_progress"`    // 收货进度（百分比0.00%）前端直接显示
	CompanyUuid       uint64             `json:"company_uuid"`        // V2.6 公司UUID
	CompanyName       string             `json:"company_name"`        // V2.6 公司名称
	CompanyStoreCode  string             `json:"company_store_code"`  // 公司店铺编码
}
```

**Step 2: 修改 Service 层数据转换逻辑**

在 `GetPurchaseOrderList` 方法中，对 `order_time=0` 的情况返回 `nil`：

```go
// 转换响应数据
listResp := make([]*resp.PurchaseOrderInfo, 0, len(purchaseOrders))
for _, po := range purchaseOrders {
	poInfo := &resp.PurchaseOrderInfo{}
	if err := copier.Copy(poInfo, &po); err != nil {
		continue
	}
	
	// 处理 order_time：待提交状态时为 0，返回 null
	if po.OrderTime > 0 {
		poInfo.OrderTime = &po.OrderTime
	} else {
		poInfo.OrderTime = nil
	}
	
	poInfo.ReceiptProgress = fmt.Sprintf("%.0f%%", po.GetReceiptProgress())
	listResp = append(listResp, poInfo)
}
```

**Step 3: 前端显示逻辑**

前端接收到 `order_time: null` 时，显示为"未提交"或"-"：

```typescript
// 前端示例代码
const formatOrderTime = (orderTime: number | null) => {
  if (orderTime === null || orderTime === 0) {
    return '未提交';
  }
  return formatTimestamp(orderTime);
};
```

---

## 📝 实施计划

### 任务清单

- [ ] 修改响应结构体 `PurchaseOrderInfo`，将 `OrderTime` 改为指针类型
- [ ] 修改 `GetPurchaseOrderList` 方法，处理 `order_time=0` 的情况
- [ ] 检查其他使用 `PurchaseOrderInfo` 的地方（详情接口等）
- [ ] 前端配合修改：显示 `null` 时展示"未提交"
- [ ] 测试验证：
  - [ ] 待提交状态显示"未提交"或"-"
  - [ ] 已提交状态正常显示时间
  - [ ] 列表查询、详情查询正常
- [ ] 更新 API 文档

### 影响范围评估

**后端改动**：
- `main/app/dto/resp/purchase_order.go` - 修改响应结构体
- `main/app/service/purchase_order/purchase_order.go` - 修改数据转换逻辑

**前端改动**：
- 新管理端 - 品采单列表页面

**兼容性**：
- ✅ 向前兼容：前端已有的 `null` 判断逻辑仍然有效
- ⚠️ 需注意：前端直接使用 `order_time` 的地方需要加 `null` 判断

---

## 🧪 测试要点

### 1. 待提交状态测试
- 创建新的品采单（不提交）
- 查看列表，单据日期应显示"未提交"或"-"
- API 返回 `order_time: null`

### 2. 已提交状态测试
- 提交品采单
- 查看列表，单据日期正常显示提交时间
- API 返回 `order_time: 1736668800`（示例时间戳）

### 3. 边界测试
- 历史数据兼容性（已有的 `order_time=0` 数据）
- 详情接口是否也需要处理
- 导出功能是否受影响

---

## 📚 相关代码位置

| 文件 | 说明 |
|------|------|
| `main/app/model/purchase_order.go:21` | 数据库模型定义 |
| `main/app/dto/resp/purchase_order.go:18` | 响应结构体 |
| `main/app/service/purchase_order/purchase_order.go:78-152` | 列表查询逻辑 |
| `main/app/api/v1/shop/shop_purchase.go:38` | API 接口 |

---

## 🔄 后续优化建议

1. **统一时间字段处理规范**
   - 建立"未设置时间"的统一处理方式
   - 所有时间字段为 0 时，统一返回 `null`

2. **前端显示规范**
   - 建立"未设置"状态的统一显示规则
   - 使用"-"或"未设置"等友好文案

3. **文档完善**
   - API 文档明确标注字段可为 `null` 的情况
   - 前后端协作文档补充时间字段处理规范

---

## ✅ 验收标准

- [ ] 待提交状态的品采单，日期字段不显示 "1970-01-01"
- [ ] 待提交状态显示"未提交"或"-"等友好文案
- [ ] 已提交状态正常显示单据日期
- [ ] 不影响其他状态的品采单显示
- [ ] API 文档已更新
- [ ] 前后端联调通过
