# ERPNext 条形码设置与 TTPOS 扫码收货功能实现方案

> 门店扫码收货功能完整实现指南：ERPNext 条形码设置 + TTPOS 扫码收货

---

## 📋 功能概述

门店收货时，支持两种扫码方式：
1. **扫物品条形码/二维码**：快速添加收货物品
2. **扫收货单条形码/二维码**：快速打开收货单进行收货

---

## 一、ERPNext 中设置物品条形码/二维码

### 1.1 在 ERPNext 中设置物品条形码

#### 方法一：通过 ERPNext Web 界面设置

**路径**：`库存 > 物品 > 新建/编辑物品`

**操作步骤**：

1. **打开物品编辑页面**
   - 进入 `库存 > 物品`
   - 选择要设置条形码的物品，点击编辑

2. **添加条形码**
   - 在物品详情页面，找到 **"条形码"（Barcodes）** 标签页
   - 点击 **"添加行"** 按钮
   - 在 **"条形码"** 字段中输入条形码值（支持 EAN-13、UPC-A、Code 128 等格式）
   - 可以添加多个条形码（例如：主条形码、备用条形码）

3. **保存物品**
   - 点击 **"保存"** 按钮
   - 条形码信息会保存到 ERPNext 数据库中

**示例**：
```
物品编码：ITEM-001
物品名称：面粉
条形码：
  - 主条形码：1234567890123
  - 备用条形码：1234567890124
```

#### 方法二：通过 ERPNext API 设置

**接口路径**：
```
PUT /api/resource/Item/{item_code}
```

**请求体**：
```json
{
  "barcodes": [
    {
      "barcode": "1234567890123"
    },
    {
      "barcode": "1234567890124"
    }
  ]
}
```

**说明**：
- `barcodes` 是一个数组，可以包含多个条形码
- 每个条形码对象包含 `barcode` 字段
- 支持 EAN-13（13位）、UPC-A（12位）、Code 128 等格式

### 1.2 在 ERPNext 中设置物品二维码

ERPNext 本身不直接支持二维码，但可以通过以下方式实现：

#### 方法一：使用条形码字段存储二维码内容

**原理**：将二维码内容（通常是 URL 或文本）存储在条形码字段中

**操作步骤**：
1. 在物品的 **"条形码"** 标签页添加条形码
2. 在条形码字段中输入二维码内容（例如：`ITEM-001` 或 `https://example.com/item/ITEM-001`）
3. 保存物品

**说明**：
- 二维码扫描器可以扫描二维码并获取内容
- TTPOS 端接收到二维码内容后，可以解析并匹配物品

#### 方法二：使用自定义字段存储二维码

**路径**：`设置 > 自定义 > 自定义字段`

**操作步骤**：
1. 创建自定义字段：
   - **文档类型**：Item（物品）
   - **字段标签**：二维码
   - **字段名称**：custom_qrcode
   - **字段类型**：数据（Data）
2. 在物品编辑页面，找到 **"更多信息"** 标签页
3. 在 **"二维码"** 字段中输入二维码内容
4. 保存物品

### 1.3 条形码同步到 TTPOS

当物品从 ERPNext 同步到 TTPOS 时，条形码信息会自动同步：

**同步流程**：
```
ERPNext 物品（包含条形码）
    ↓ (gRPC)
BMP 模块（ttpos-erp）
    ↓ (同步物品)
TTPOS Main 模块
    ↓ (保存到数据库)
物品表（ttpos_material.barcode_value）
```

**代码位置**：
- ERPNext → BMP：`ttpos-bmp/app/ttpos-erp/internal/logic/stock/item.go`
- BMP → TTPOS：`main/app/service/rpc/erp/item.go`
- 物品模型：`main/app/model/material.go`（`BarcodeValue` 字段）

**验证方法**：
1. 在 ERPNext 中设置物品条形码
2. 在 TTPOS 中同步物品（或等待自动同步）
3. 在 TTPOS 物品列表中查看物品的条形码字段

---

## 二、TTPOS 扫码收货功能实现

### 2.1 扫物品条形码收货

#### 功能描述

在收货单创建/编辑页面，通过扫描物品条形码快速添加收货物品。

#### 业务流程

```
1. 门店人员打开收货单创建/编辑页面
   ↓
2. 点击"扫码添加物品"按钮
   ↓
3. 扫描物品条形码/二维码
   ↓
4. TTPOS 根据条形码查询物品
   - 在物品表（ttpos_material）中查询 barcode_value
   - 如果找到物品，显示物品信息
   ↓
5. 验证物品是否在采购单中
   - 检查物品是否属于当前采购单
   - 如果不在，提示错误
   ↓
6. 自动添加到收货明细
   - 如果物品已在明细中，增加数量
   - 如果物品不在明细中，新增一行
   ↓
7. 更新收货单界面
```

#### API 设计

**接口路径**：
```
POST /shop/purchase/receipt/scan/material
```

**请求参数**：
```json
{
  "purchase_order_uuid": "1234567890",
  "barcode": "1234567890123"
}
```

**响应数据**：
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "material": {
      "uuid": 1234567890,
      "name": "面粉",
      "code": "ITEM-001",
      "barcode_value": "1234567890123",
      "unit_name": "kg",
      "purchase_num": 100.0,
      "arrival_num": 0.0
    },
    "is_in_order": true,
    "message": "物品已添加到收货明细"
  }
}
```

**错误响应**：
```json
{
  "code": 1,
  "message": "物品不存在或不属于当前采购单",
  "data": null
}
```

#### 代码实现位置

**建议实现位置**：
- API 接口：`main/app/api/v1/shop/shop_purchase.go`
- Service 层：`main/app/service/purchase_order/receipt_order.go`
- Repository 层：`main/app/repository/material.go`

**关键代码逻辑**：

```go
// 1. 处理条形码（过滤空格、非数字字符，截取13位）
processedBarcode := utils.ProcessBarcode(req.Barcode)

// 2. 根据条形码查询物品
materialRepo := repository.NewMaterialRepo(db)
material, err := materialRepo.GetMaterialByBarcode(processedBarcode)
if err != nil {
    return errors.New("物品不存在")
}

// 3. 验证物品是否在采购单中
purchaseOrderItemRepo := repository.NewPurchaseOrderItemRepo(db)
orderItem, err := purchaseOrderItemRepo.GetByPurchaseOrderUuidAndMaterialUuid(
    req.PurchaseOrderUuid, 
    material.Uuid,
)
if err != nil {
    return errors.New("物品不属于当前采购单")
}

// 4. 返回物品信息
return material, nil
```

**需要新增的 Repository 方法**：

```go
// GetMaterialByBarcode 根据条形码查询物品
func (r *MaterialRepoImpl) GetMaterialByBarcode(barcode string, opts ...DBOption) (*model.Material, error) {
    var material model.Material
    query := r.db.Model(&model.Material{}).
        Where("barcode_value = ?", barcode).
        Where("delete_time = ?", 0)
    
    for _, opt := range opts {
        query = opt(query)
    }
    
    if err := query.First(&material).Error; err != nil {
        return nil, errors.WithMessage(err, "根据条形码查询物品失败")
    }
    
    return &material, nil
}
```

### 2.2 扫收货单条形码/二维码

#### 功能描述

通过扫描收货单条形码/二维码，快速打开收货单进行收货操作。

#### 业务流程

```
1. 门店人员打开"收货单列表"页面
   ↓
2. 点击"扫码打开收货单"按钮
   ↓
3. 扫描收货单条形码/二维码
   - 收货单编号格式：SHRK202501120001
   - 或二维码内容：receipt_uuid 或 receipt_no
   ↓
4. TTPOS 根据条形码/二维码查询收货单
   - 在收货单表（ttpos_purchase_receipt_order）中查询
   - 支持按 order_no 或 uuid 查询
   ↓
5. 跳转到收货单详情页面
   - 显示收货单信息
   - 显示收货明细列表
   - 可以进行收货操作
```

#### 收货单条形码生成

**生成时机**：
- 创建收货单时自动生成收货单编号
- 格式：`SHRK + 8位日期(YYYYMMDD) + 4位序号`
- 示例：`SHRK202501120001`

**代码位置**：
- `main/app/service/purchase_order/receipt_order.go::CreatePurchaseReceiptOrder()`

**生成逻辑**：
```go
// 格式：SHRK + 8位日期(YYYYMMDD) + 4位序号
// 示例：SHRK202501120001
receiptNo := "SHRK" + datePart + serialNo
```

#### 收货单二维码生成（可选）

**生成内容**：
- 方案一：收货单编号（`SHRK202501120001`）
- 方案二：收货单 UUID（`1234567890`）
- 方案三：JSON 格式（`{"type":"receipt","no":"SHRK202501120001","uuid":1234567890}`）

**生成位置**：
- 收货单详情页面显示二维码
- 收货单打印时包含二维码
- 可以通过 API 获取二维码图片

#### API 设计

**接口路径**：
```
POST /shop/purchase/receipt/scan/receipt
```

**请求参数**：
```json
{
  "barcode": "SHRK202501120001"
}
```

**或**：
```json
{
  "barcode": "1234567890"
}
```

**响应数据**：
```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "receipt": {
      "uuid": 1234567890,
      "order_no": "SHRK202501120001",
      "status": 0,
      "purchase_order_uuid": 9876543210,
      "purchase_order_no": "CG202501120001",
      "supplier_name": "供应商A",
      "num": 10.0,
      "expect_arrival_time": 1705123200,
      "receive_time": 0,
      "items": [...]
    }
  }
}
```

**错误响应**：
```json
{
  "code": 1,
  "message": "收货单不存在",
  "data": null
}
```

#### 代码实现位置

**建议实现位置**：
- API 接口：`main/app/api/v1/shop/shop_purchase.go`
- Service 层：`main/app/service/purchase_order/receipt_order.go`
- Repository 层：`main/app/repository/purchase_order.go`

**关键代码逻辑**：

```go
// 1. 处理条形码（可能是收货单编号或 UUID）
processedBarcode := utils.ProcessBarcode(req.Barcode)

// 2. 尝试按收货单编号查询
receiptOrderRepo := repository.NewPurchaseReceiptOrderRepo(db)
receiptOrder, err := receiptOrderRepo.GetByOrderNo(processedBarcode)
if err != nil {
    // 3. 如果按编号查询失败，尝试按 UUID 查询（如果条形码是数字）
    if uuid, err := strconv.ParseUint(processedBarcode, 10, 64); err == nil {
        receiptOrder, err = receiptOrderRepo.GetByUuid(uuid)
    }
    if err != nil {
        return errors.New("收货单不存在")
    }
}

// 4. 返回收货单信息
return receiptOrder, nil
```

---

## 三、前端实现建议

### 3.1 扫码功能集成

#### 使用 HTML5 扫码 API

**方案一：使用 `<input type="file" accept="image/*" capture="environment">`**

```vue
<template>
  <div>
    <input
      ref="barcodeInput"
      type="file"
      accept="image/*"
      capture="environment"
      @change="handleBarcodeScan"
      style="display: none"
    />
    <button @click="triggerScan">扫码</button>
  </div>
</template>

<script setup>
const barcodeInput = ref(null)

const triggerScan = () => {
  barcodeInput.value?.click()
}

const handleBarcodeScan = async (event) => {
  const file = event.target.files[0]
  if (!file) return
  
  // 使用二维码识别库（如 jsQR、html5-qrcode）
  const barcode = await scanBarcodeFromImage(file)
  if (barcode) {
    // 调用后端 API
    await handleBarcodeResult(barcode)
  }
}
</script>
```

**方案二：使用摄像头实时扫码**

```vue
<template>
  <div>
    <video ref="video" autoplay playsinline></video>
    <canvas ref="canvas" style="display: none"></canvas>
    <button @click="startScan">开始扫码</button>
    <button @click="stopScan">停止扫码</button>
  </div>
</template>

<script setup>
import { Html5Qrcode } from 'html5-qrcode'

const video = ref(null)
const canvas = ref(null)
let html5Qrcode = null

const startScan = async () => {
  html5Qrcode = new Html5Qrcode("reader")
  await html5Qrcode.start(
    { facingMode: "environment" },
    {
      fps: 10,
      qrbox: { width: 250, height: 250 }
    },
    (decodedText) => {
      handleBarcodeResult(decodedText)
      stopScan()
    },
    (errorMessage) => {
      // 忽略错误
    }
  )
}

const stopScan = () => {
  if (html5Qrcode) {
    html5Qrcode.stop().then(() => {
      html5Qrcode = null
    })
  }
}
</script>
```

#### 使用第三方扫码库

**推荐库**：
- `html5-qrcode`：支持图片和摄像头扫码
- `jsQR`：纯 JavaScript 二维码识别库
- `quaggaJS`：支持条形码和二维码识别

### 3.2 收货单页面集成扫码功能

**收货单创建/编辑页面**：

```vue
<template>
  <div class="receipt-page">
    <!-- 扫码添加物品 -->
    <el-button @click="scanMaterial" type="primary">
      <el-icon><Scan /></el-icon>
      扫码添加物品
    </el-button>
    
    <!-- 收货明细列表 -->
    <el-table :data="receiptItems">
      <el-table-column prop="material_name" label="物品名称" />
      <el-table-column prop="barcode_value" label="条形码" />
      <el-table-column prop="num" label="收货数量" />
    </el-table>
  </div>
</template>

<script setup>
const scanMaterial = async () => {
  // 打开扫码界面
  const barcode = await openBarcodeScanner()
  if (!barcode) return
  
  // 调用后端 API
  const response = await api.post('/shop/purchase/receipt/scan/material', {
    purchase_order_uuid: purchaseOrderUuid.value,
    barcode: barcode
  })
  
  if (response.code === 0) {
    // 添加物品到收货明细
    addMaterialToReceipt(response.data.material)
  } else {
    ElMessage.error(response.message)
  }
}
</script>
```

**收货单列表页面**：

```vue
<template>
  <div class="receipt-list-page">
    <!-- 扫码打开收货单 -->
    <el-button @click="scanReceipt" type="primary">
      <el-icon><Scan /></el-icon>
      扫码打开收货单
    </el-button>
    
    <!-- 收货单列表 -->
    <el-table :data="receiptList">
      <!-- ... -->
    </el-table>
  </div>
</template>

<script setup>
const scanReceipt = async () => {
  // 打开扫码界面
  const barcode = await openBarcodeScanner()
  if (!barcode) return
  
  // 调用后端 API
  const response = await api.post('/shop/purchase/receipt/scan/receipt', {
    barcode: barcode
  })
  
  if (response.code === 0) {
    // 跳转到收货单详情页面
    router.push(`/purchase/receipt/detail/${response.data.receipt.uuid}`)
  } else {
    ElMessage.error(response.message)
  }
}
</script>
```

---

## 四、ERPNext 收货单条形码（可选）

### 4.1 在 ERPNext 中生成收货单条形码

**ERPNext Purchase Receipt 编号格式**：
- 格式：`MAT-PRE-YYYY-XXXXX`
- 示例：`MAT-PRE-2025-00303`

**生成位置**：
- ERPNext 创建 Purchase Receipt 时自动生成
- 可以通过 ERPNext 的 Series 配置自定义格式

### 4.2 同步收货单编号到 TTPOS

**同步流程**：
```
TTPOS 创建收货单
    ↓ (调用 ERPNext API)
ERPNext 创建 Purchase Receipt
    ↓ (返回收货单编号)
TTPOS 更新收货单 erp_order_no
```

**代码位置**：
- `main/app/service/purchase_order/receipt_order.go::CreatePurchaseReceiptOrder()`
- `ttpos-bmp/app/ttpos-erp/internal/logic/buying/buying.go::SavePurchaseReceipt()`

**同步逻辑**：
```go
// 创建 ERPNext Purchase Receipt
erpResp, err := s.erpSrv.SavePurchaseReceipt(ctx, erpReq)
if err == nil && erpResp.PurchaseReceiptName != "" {
    // 更新收货单的 ERP 订单号
    receiptOrder.ErpOrderNo = erpResp.PurchaseReceiptName
    receiptOrderRepo.Update(receiptOrder)
}
```

### 4.3 扫 ERPNext 收货单编号

**功能描述**：
- 支持扫描 ERPNext 的收货单编号（`MAT-PRE-2025-00303`）
- 在 TTPOS 中查询对应的收货单（通过 `erp_order_no` 字段）

**API 设计**：
```
POST /shop/purchase/receipt/scan/receipt
```

**请求参数**：
```json
{
  "barcode": "MAT-PRE-2025-00303"
}
```

**查询逻辑**：
```go
// 1. 先按 TTPOS 收货单编号查询
receiptOrder, err := receiptOrderRepo.GetByOrderNo(barcode)
if err != nil {
    // 2. 如果失败，按 ERP 收货单编号查询
    receiptOrder, err = receiptOrderRepo.GetByErpOrderNo(barcode)
    if err != nil {
        return errors.New("收货单不存在")
    }
}
```

---

## 五、实现检查清单

### 5.1 ERPNext 配置

- [ ] 在 ERPNext 中为物品设置条形码
- [ ] 验证条形码格式（EAN-13、UPC-A 等）
- [ ] 确认条形码同步到 TTPOS
- [ ] 测试多个条形码支持（主条形码、备用条形码）

### 5.2 TTPOS 后端实现

- [ ] 实现 `GetMaterialByBarcode` Repository 方法
- [ ] 实现 `POST /shop/purchase/receipt/scan/material` API
- [ ] 实现 `POST /shop/purchase/receipt/scan/receipt` API
- [ ] 添加条形码处理逻辑（`utils.ProcessBarcode`）
- [ ] 添加错误处理和提示信息
- [ ] 添加单元测试

### 5.3 TTPOS 前端实现

- [ ] 集成扫码功能（html5-qrcode 或 jsQR）
- [ ] 在收货单创建/编辑页面添加"扫码添加物品"按钮
- [ ] 在收货单列表页面添加"扫码打开收货单"按钮
- [ ] 实现扫码结果处理逻辑
- [ ] 添加扫码错误提示
- [ ] 优化扫码体验（自动对焦、提示音等）

### 5.4 测试验证

- [ ] 测试扫物品条形码添加收货物品
- [ ] 测试扫收货单编号打开收货单
- [ ] 测试扫 ERPNext 收货单编号
- [ ] 测试错误场景（物品不存在、不属于采购单等）
- [ ] 测试不同条形码格式（EAN-13、UPC-A、Code 128）
- [ ] 测试二维码扫描（如果支持）

---

## 六、注意事项

### 6.1 条形码格式处理

**TTPOS 条形码处理规则**（`main/pkg/utils/barcode.go`）：
1. 过滤空格
2. 过滤非数字字符
3. 如果长度大于13位，截取前13位

**注意**：
- 此规则适用于 EAN-13 格式（13位数字）
- 如果使用其他格式（如 Code 128），可能需要调整处理逻辑

### 6.2 条形码唯一性

**要求**：
- 每个物品的条形码必须唯一
- 在物品导入时已进行条形码重复检查（`main/app/service/material.go::ImportMaterial()`）

**验证方法**：
```go
// 检查条形码是否已存在
if repository.NewMaterialRepo(db).CheckBarcodeExist(barcode, 0) {
    return errors.New("物品条码已存在")
}
```

### 6.3 性能优化

**建议**：
- 对条形码字段建立数据库索引
- 使用缓存加速条形码查询（Redis）
- 批量扫码时使用批量查询接口

**索引创建**：
```sql
CREATE INDEX idx_material_barcode_value ON ttpos_material(barcode_value);
```

### 6.4 安全性

**建议**：
- 验证扫码权限（只有有权限的用户才能扫码）
- 防止 SQL 注入（使用参数化查询）
- 限制扫码频率（防止恶意扫描）

---

## 七、相关文档

- [ERPNext 仓库层级管理](./erpnext-warehouse-hierarchy.md)
- [采购收货单功能文档](../architecture/features/purchase_order.md)
- [物品管理功能文档](../architecture/features/material.md)
- [条形码处理工具](../architecture/utils/barcode.md)

---

**最后更新**：2026-01-07  
**维护者**：TTPOS Team

