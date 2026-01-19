# ERPNext 第三方物流扫码发货功能实现方案

> 第三方物流人员通过 ERPNext 进行扫码发货的完整实现指南

---

## 📋 功能概述

**业务需求**：
- 第三方物流人员登录 ERPNext 进行发货操作
- 通过扫码方式快速完成发货流程
- 支持扫码发货单、扫码物品条形码等多种扫码方式

**核心功能**：
1. **账号权限配置**：为第三方物流人员配置发货权限
2. **发货业务流程**：完整的发货操作流程
3. **扫码发货功能**：支持扫码发货单、扫码物品条形码

---

## 一、账号权限配置

### 1.1 创建第三方物流用户账号

**导航路径**：
```
主页 → 设置（Settings） → 用户和权限（Users and Permissions） → 用户（Users）
```

**操作步骤**：

1. **创建新用户**：
   - 点击 **"新建"** 按钮
   - 填写用户信息：
     ```
     邮箱：logistics@example.com
     全名：第三方物流人员
     语言：中文（简体）
     时区：Asia/Shanghai
     ```

2. **设置用户角色**：
   - 在 **"角色"** 标签页添加以下角色：
     - `Stock User`（库存用户）- 必需，用于查看和操作库存
     - `Stock Manager`（库存管理员）- 可选，如果需要更多权限
     - `Sales User`（销售用户）- 可选，如果需要查看销售订单
     - `Delivery Note User`（送货单用户）- 可选，如果项目中有此自定义角色

3. **保存用户**

**重要说明**：
- ✅ 第三方物流人员**不需要** `System Manager`、`Accounts Manager` 等高级权限
- ✅ 只需要基本的库存和发货相关权限
- ✅ 建议创建专门的 `Logistics User` 角色（如果项目支持）

---

### 1.2 配置用户权限（User Permission）

**导航路径**：
```
主页 → 设置（Settings） → 用户和权限（Users and Permissions） → 用户权限（User Permissions）
```

**配置步骤**：

#### 步骤 1：配置公司权限

为第三方物流用户配置可以访问的公司：

**User Permission 1（总部公司）**：
```
用户：logistics@example.com
允许：Company
适用于：CFG（或总部公司名称）
是否默认：✅ 是
Hide Descendants：❌ 不勾选
```

**User Permission 2（门店A）**（如果需要）：
```
用户：logistics@example.com
允许：Company
适用于：门店A
是否默认：❌ 否
Hide Descendants：❌ 不勾选
```

**说明**：
- ✅ 如果第三方物流需要为多个门店发货，需要为每个门店创建 User Permission
- ✅ **不勾选 Hide Descendants**：允许查看子公司的数据
- ✅ 如果只需要为总部发货，只需要配置总部公司的权限

#### 步骤 2：配置仓库权限（可选）

如果需要对第三方物流人员限制只能访问特定仓库：

**User Permission（仓库）**：
```
用户：logistics@example.com
允许：Warehouse
适用于：发货仓库名称（如：总部仓库、北京仓库）
是否默认：✅ 是
Hide Descendants：❌ 不勾选
```

**说明**：
- ⚠️ 如果配置了仓库权限，用户只能看到和操作指定仓库的数据
- ✅ 如果不配置仓库权限，用户可以看到所有有权限公司的仓库数据

---

### 1.3 配置角色权限（Role Permission）

**导航路径**：
```
主页 → 设置（Settings） → 用户和权限（Users and Permissions） → 角色权限管理器（Role Permission Manager）
```

**配置步骤**：

#### 步骤 1：配置 Delivery Note（送货单）权限

**角色**：`Stock User` 或 `Logistics User`

**权限配置**：

| 文档类型 | 权限类型 | 权限级别 | 说明 |
|---------|---------|---------|------|
| Delivery Note | 读取 | 用户 | 可以查看自己创建的送货单 |
| Delivery Note | 写入 | 用户 | 可以创建和编辑送货单 |
| Delivery Note | 提交 | 用户 | 可以提交送货单 |
| Delivery Note | 取消 | 用户 | 可以取消送货单 |
| Delivery Note | 删除 | 用户 | 可以删除草稿状态的送货单 |

**说明**：
- ✅ 第三方物流人员需要能够**创建、编辑、提交**送货单
- ✅ 不需要**删除已提交**的送货单权限
- ✅ 不需要**审批**权限（如果项目中有审批流程）

#### 步骤 2：配置 Item（物品）权限

**权限配置**：

| 文档类型 | 权限类型 | 权限级别 | 说明 |
|---------|---------|---------|------|
| Item | 读取 | 全部 | 可以查看所有物品信息（用于扫码匹配） |
| Item | 写入 | 无 | 不需要创建或编辑物品 |

**说明**：
- ✅ 第三方物流人员需要**读取**物品信息，用于扫码匹配物品
- ❌ 不需要**写入**权限（不能修改物品信息）

#### 步骤 3：配置 Warehouse（仓库）权限

**权限配置**：

| 文档类型 | 权限类型 | 权限级别 | 说明 |
|---------|---------|---------|------|
| Warehouse | 读取 | 全部 | 可以查看仓库信息 |
| Warehouse | 写入 | 无 | 不需要创建或编辑仓库 |

---

### 1.4 创建自定义角色（可选，推荐）

如果项目支持，建议创建专门的 `Logistics User` 角色：

**导航路径**：
```
主页 → 设置（Settings） → 用户和权限（Users and Permissions） → 角色（Roles）
```

**创建角色**：

```
角色名称：Logistics User
角色描述：第三方物流人员角色，用于发货操作
```

**权限配置**：
- 继承 `Stock User` 的基础权限
- 添加 `Delivery Note` 的创建、编辑、提交权限
- 限制其他不必要的权限（如：不能创建销售订单、不能修改物品等）

---

## 二、发货业务流程

### 2.1 业务流程概述

```
1. 总部创建销售订单（Sales Order）
   ↓
2. 总部从销售订单创建送货单（Delivery Note）
   - 选择具体发货仓库
   - 填写发货数量
   ↓
3. 第三方物流人员登录 ERPNext
   ↓
4. 查看待发货的送货单列表
   ↓
5. 选择送货单，进行扫码发货
   - 方式一：扫码送货单编号，快速打开送货单
   - 方式二：在送货单中扫码物品条形码，确认发货物品
   ↓
6. 确认发货数量，提交送货单
   ↓
7. 库存从发货仓库扣减
   ↓
8. 发货完成
```

---

### 2.2 详细操作流程

#### 步骤 1：总部创建送货单（Delivery Note）

**路径**：`销售 > 交货单 > 新建` 或 `从销售订单创建`

**操作步骤**：

1. **从销售订单创建送货单**：
   - 进入 `销售 > 销售订单`
   - 选择需要发货的销售订单
   - 点击 **"创建"** → **"交货单"**

2. **填写送货单信息**：
   ```
   客户：门店名称
   公司：CFG（或总部公司）
   交货日期：实际发货日期
   仓库：具体发货仓库（如：北京仓库）
   物品明细：
     - 物品A：数量 100
     - 物品B：数量 50
   ```

3. **保存送货单**（不提交）：
   - 点击 **"保存"** 按钮
   - 送货单状态为 **"草稿"（Draft）**
   - 此时**不会扣减库存**

4. **打印送货单**（可选）：
   - 打印送货单，包含送货单编号和二维码
   - 交给第三方物流人员

**重要说明**：
- ✅ 送货单在**草稿状态**时，第三方物流人员可以编辑
- ✅ 送货单**提交后**，库存会立即扣减，无法再编辑
- ✅ 建议在确认发货后再提交送货单

---

#### 步骤 2：第三方物流人员查看待发货列表

**路径**：`销售 > 交货单`

**操作步骤**：

1. **登录 ERPNext**：
   - 使用第三方物流人员账号登录
   - 例如：`logistics@example.com`

2. **查看送货单列表**：
   - 进入 `销售 > 交货单`
   - 列表显示所有有权限的送货单

3. **筛选待发货的送货单**：
   - 使用筛选器：
     ```
     状态 = 草稿（Draft）
     公司 = CFG（或总部公司）
     ```

4. **查看送货单详情**：
   - 点击送货单编号，查看详情
   - 查看物品明细、数量、仓库等信息

---

#### 步骤 3：扫码发货操作

**方式一：扫码送货单编号，快速打开送货单**

1. **点击"扫码打开送货单"按钮**：
   - 在送货单列表页面，点击扫码按钮
   - 或使用快捷键（如果配置了）

2. **扫描送货单二维码/条形码**：
   - 使用扫码枪或手机扫码
   - 扫码内容：送货单编号（如：`DN-00001`）

3. **自动打开送货单**：
   - 系统根据扫码结果自动打开对应的送货单
   - 显示送货单详情和物品明细

**方式二：在送货单中扫码物品条形码，确认发货物品**

1. **打开送货单详情页面**：
   - 从列表中选择送货单，或通过扫码打开

2. **点击"扫码添加物品"按钮**：
   - 在送货单详情页面的物品明细区域
   - 点击 **"扫码"** 按钮

3. **扫描物品条形码**：
   - 使用扫码枪扫描物品上的条形码
   - 系统自动匹配物品

4. **确认发货数量**：
   - 如果物品已在明细中，可以修改数量
   - 如果物品不在明细中，可以添加新行

5. **验证发货数量**：
   - 系统会验证发货数量是否超过订单数量
   - 如果超过，会提示错误

---

#### 步骤 4：提交送货单，完成发货

**操作步骤**：

1. **确认送货单信息**：
   - 检查客户、仓库、物品明细、数量等信息
   - 确认无误后，点击 **"提交"** 按钮

2. **提交送货单**：
   - 系统会验证：
     - 仓库库存是否充足
     - 发货数量是否超过订单数量
     - 必填字段是否完整
   - 验证通过后，送货单状态变为 **"已提交"（Submitted）**

3. **库存扣减**：
   - 提交后，库存从发货仓库扣减
   - 可以在 `库存 > 库存余额` 中查看库存变化

4. **发货完成**：
   - 送货单状态为 **"已提交"**
   - 可以打印送货单，交给司机配送
   - 可以查看送货单历史记录

---

### 2.3 业务流程注意事项

#### 重要限制

1. **库存验证**：
   - ⚠️ 提交送货单时，系统会验证仓库库存是否充足
   - ⚠️ 如果库存不足，无法提交送货单
   - ✅ 需要先补货或调整发货数量

2. **数量限制**：
   - ⚠️ 发货数量不能超过销售订单数量
   - ⚠️ 如果销售订单数量为 100，发货数量最多为 100
   - ✅ 可以部分发货（发货数量 < 订单数量）

3. **仓库选择**：
   - ⚠️ 送货单中的仓库必须是实际仓库（不能是虚拟仓库）
   - ✅ 如果使用仓库层级，必须选择具体的子仓库

4. **权限限制**：
   - ⚠️ 第三方物流人员只能操作有权限的公司和仓库的送货单
   - ⚠️ 无法查看或操作其他公司的送货单

---

## 三、扫码发货相关配置

### 3.1 ERPNext 条形码设置

#### 3.1.1 设置物品条形码

**路径**：`库存 > 物品 > 新建/编辑物品`

**操作步骤**：

1. **打开物品编辑页面**：
   - 进入 `库存 > 物品`
   - 选择要设置条形码的物品，点击编辑

2. **添加条形码**：
   - 在物品详情页面，找到 **"条形码"（Barcodes）** 标签页
   - 点击 **"添加行"** 按钮
   - 在 **"条形码"** 字段中输入条形码值
   - 支持格式：EAN-13（13位）、UPC-A（12位）、Code 128 等

3. **保存物品**：
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

**说明**：
- ✅ 每个物品可以设置多个条形码（主条形码、备用条形码）
- ✅ 条形码必须唯一（不能重复）
- ✅ 建议使用标准格式（EAN-13、UPC-A）

---

#### 3.1.2 设置送货单条形码/二维码

**生成时机**：
- 创建送货单时自动生成送货单编号
- 格式：`DN-` + 序号（如：`DN-00001`）

**生成位置**：
- 送货单详情页面显示送货单编号
- 送货单打印时包含送货单编号和二维码
- 可以通过 API 获取二维码图片

**二维码生成（可选）**：

**生成内容**：
- 方案一：送货单编号（`DN-00001`）
- 方案二：JSON 格式（`{"type":"delivery_note","name":"DN-00001"}`）

**生成位置**：
- 送货单详情页面显示二维码
- 送货单打印时包含二维码
- 可以通过自定义字段或打印格式添加二维码

**实现方式**：

1. **使用 ERPNext 打印格式**：
   - 在打印格式中添加二维码字段
   - 使用 ERPNext 的二维码生成功能

2. **使用自定义字段**：
   - 创建自定义字段存储二维码内容
   - 在送货单保存时自动生成二维码

3. **使用客户端脚本**：
   - 在送货单表单中添加二维码显示
   - 使用 JavaScript 库生成二维码（如：qrcode.js）

---

### 3.2 扫码功能实现

#### 3.2.1 扫码打开送货单功能

**功能描述**：
- 通过扫描送货单条形码/二维码，快速打开送货单进行发货操作

**实现方式**：

**方案一：使用 ERPNext 客户端脚本**

**导航路径**：
```
主页 → 设置（Settings） → 自定义（Customize） → 客户端脚本（Client Script）
```

**创建客户端脚本**：

**脚本类型**：`DocType`  
**文档类型**：`Delivery Note`  
**脚本类型**：`List View`（列表视图脚本）

**脚本内容**（JavaScript）：
```javascript
// 扫码打开送货单功能
(function() {
    // 添加扫码按钮到列表视图工具栏
    frappe.listview_settings['Delivery Note'] = {
        onload: function(listview) {
            // 添加扫码按钮
            listview.page.add_inner_button(__('扫码打开送货单'), function() {
                openBarcodeScanner(listview);
            }, __('工具'));
        }
    };
    
    // 扫码功能
    function openBarcodeScanner(listview) {
        // 方案一：使用 HTML5 摄像头扫码
        if (navigator.mediaDevices && navigator.mediaDevices.getUserMedia) {
            // 打开摄像头扫码界面
            frappe.prompt([
                {
                    label: '扫码结果',
                    fieldname: 'barcode',
                    fieldtype: 'Data',
                    reqd: 1
                }
            ], function(values) {
                // 根据扫码结果查找送货单
                const deliveryNoteName = values.barcode;
                if (deliveryNoteName) {
                    // 跳转到送货单详情页面
                    window.location.href = `/app/delivery-note/${deliveryNoteName}`;
                }
            }, __('扫码打开送货单'), __('打开'));
        } else {
            // 方案二：手动输入送货单编号
            frappe.prompt([
                {
                    label: '送货单编号',
                    fieldname: 'delivery_note_name',
                    fieldtype: 'Link',
                    options: 'Delivery Note',
                    reqd: 1
                }
            ], function(values) {
                if (values.delivery_note_name) {
                    window.location.href = `/app/delivery-note/${values.delivery_note_name}`;
                }
            }, __('打开送货单'), __('打开'));
        }
    }
})();
```

**方案二：使用 ERPNext 自定义按钮**

**导航路径**：
```
主页 → 设置（Settings） → 自定义（Customize） → 自定义表单（Customize Form）
```

**操作步骤**：

1. **选择文档类型**：`Delivery Note`
2. **添加自定义按钮**：
   - 在表单中添加按钮字段
   - 按钮标签：`扫码打开`
   - 按钮操作：打开扫码界面

---

#### 3.2.2 扫码添加物品功能

**功能描述**：
- 在送货单详情页面，通过扫描物品条形码快速添加或确认发货物品

**实现方式**：

**创建客户端脚本**：

**脚本类型**：`DocType`  
**文档类型**：`Delivery Note`  
**脚本类型**：`Form`（表单脚本）

**脚本内容**（JavaScript）：
```javascript
// 扫码添加物品功能
frappe.ui.form.on('Delivery Note', {
    refresh: function(frm) {
        // 只在草稿状态显示扫码按钮
        if (frm.doc.docstatus === 0) {
            // 添加扫码按钮到物品明细区域
            frm.add_custom_button(__('扫码添加物品'), function() {
                scanItemBarcode(frm);
            }, __('工具'));
        }
    }
});

// 扫码物品条形码
function scanItemBarcode(frm) {
    frappe.prompt([
        {
            label: '物品条形码',
            fieldname: 'barcode',
            fieldtype: 'Data',
            reqd: 1
        }
    ], function(values) {
        const barcode = values.barcode;
        if (!barcode) return;
        
        // 调用后端 API 查询物品
        frappe.call({
            method: 'erpnext.stock.doctype.item.item.get_item_details',
            args: {
                barcode: barcode
            },
            callback: function(r) {
                if (r.message) {
                    // 物品已找到，添加到送货单明细
                    addItemToDeliveryNote(frm, r.message);
                } else {
                    frappe.show_alert({
                        message: __('物品不存在'),
                        indicator: 'red'
                    }, 3);
                }
            }
        });
    }, __('扫码添加物品'), __('添加'));
}

// 添加物品到送货单明细
function addItemToDeliveryNote(frm, item) {
    // 检查物品是否已在明细中
    const existingRow = frm.doc.items.find(row => row.item_code === item.item_code);
    
    if (existingRow) {
        // 如果已存在，增加数量
        frappe.model.set_value(existingRow.doctype, existingRow.name, 'qty', existingRow.qty + 1);
    } else {
        // 如果不存在，添加新行
        const newRow = frm.add_child('items');
        frappe.model.set_value(newRow.doctype, newRow.name, 'item_code', item.item_code);
        frappe.model.set_value(newRow.doctype, newRow.name, 'qty', 1);
        frappe.model.set_value(newRow.doctype, newRow.name, 'warehouse', frm.doc.set_warehouse || '');
    }
    
    frm.refresh_field('items');
    frappe.show_alert({
        message: __('物品已添加'),
        indicator: 'green'
    }, 3);
}
```

**说明**：
- ✅ 扫码功能需要在浏览器中启用摄像头权限
- ✅ 可以使用扫码枪（USB 或蓝牙）直接输入条形码
- ✅ 如果使用扫码枪，扫码结果会自动填充到输入框

---

### 3.3 扫码设备配置

#### 3.3.1 扫码枪配置

**USB 扫码枪**：
- 连接方式：USB 接口
- 工作模式：模拟键盘输入
- 配置：无需特殊配置，插入即可使用
- 使用：扫描条形码后，内容会自动输入到当前焦点输入框

**蓝牙扫码枪**：
- 连接方式：蓝牙连接
- 工作模式：模拟键盘输入
- 配置：需要先配对蓝牙设备
- 使用：扫描条形码后，内容会自动输入到当前焦点输入框

**说明**：
- ✅ 扫码枪通常会自动在扫码结果后添加回车（Enter）键
- ✅ 可以在客户端脚本中监听回车键，自动触发查询或添加操作

---

#### 3.3.2 手机扫码配置

**使用手机摄像头扫码**：

**方案一：使用 HTML5 摄像头 API**

```javascript
// 打开摄像头扫码
navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' } })
    .then(function(stream) {
        // 显示摄像头画面
        const video = document.createElement('video');
        video.srcObject = stream;
        video.play();
        
        // 使用二维码识别库（如：jsQR、html5-qrcode）
        // 识别二维码内容
    })
    .catch(function(err) {
        console.error('无法访问摄像头', err);
    });
```

**方案二：使用第三方扫码库**

**推荐库**：
- `html5-qrcode`：支持图片和摄像头扫码
- `jsQR`：纯 JavaScript 二维码识别库
- `quaggaJS`：支持条形码和二维码识别

**安装**：
```bash
npm install html5-qrcode
```

**使用示例**：
```javascript
import { Html5Qrcode } from 'html5-qrcode';

const html5QrCode = new Html5Qrcode("reader");
html5QrCode.start(
    { facingMode: "environment" },
    {
        fps: 10,
        qrbox: { width: 250, height: 250 }
    },
    (decodedText) => {
        // 扫码成功，处理结果
        handleBarcodeResult(decodedText);
    },
    (errorMessage) => {
        // 忽略错误
    }
);
```

---

### 3.4 扫码功能优化

#### 3.4.1 自动聚焦输入框

**实现方式**：

```javascript
// 在打开扫码界面时，自动聚焦到输入框
frappe.prompt([
    {
        label: '物品条形码',
        fieldname: 'barcode',
        fieldtype: 'Data',
        reqd: 1
    }
], function(values) {
    // 处理扫码结果
}, __('扫码添加物品'), __('添加'));

// 自动聚焦
setTimeout(function() {
    const input = document.querySelector('input[data-fieldname="barcode"]');
    if (input) {
        input.focus();
    }
}, 100);
```

---

#### 3.4.2 自动触发查询

**实现方式**：

```javascript
// 监听输入框变化，自动触发查询
frappe.ui.form.on('Delivery Note', {
    refresh: function(frm) {
        // 添加扫码输入框
        const barcodeInput = frm.add_custom_button(__('扫码'), function() {
            // 打开扫码输入框
            const input = document.createElement('input');
            input.type = 'text';
            input.placeholder = '请扫描条形码';
            input.style.position = 'fixed';
            input.style.top = '10px';
            input.style.right = '10px';
            input.style.zIndex = '9999';
            document.body.appendChild(input);
            
            // 自动聚焦
            input.focus();
            
            // 监听输入变化（扫码枪会在扫码后自动输入并触发回车）
            input.addEventListener('keypress', function(e) {
                if (e.key === 'Enter') {
                    const barcode = input.value.trim();
                    if (barcode) {
                        // 处理扫码结果
                        handleBarcodeScan(frm, barcode);
                        // 清空输入框，准备下次扫码
                        input.value = '';
                        input.focus();
                    }
                }
            });
        });
    }
});
```

---

#### 3.4.3 扫码提示音

**实现方式**：

```javascript
// 扫码成功时播放提示音
function playBeepSound() {
    const audio = new Audio('/assets/erpnext/sounds/beep.mp3');
    audio.play().catch(function(err) {
        console.error('无法播放提示音', err);
    });
}

// 在扫码成功时调用
function handleBarcodeScan(frm, barcode) {
    // 处理扫码结果
    // ...
    
    // 播放提示音
    playBeepSound();
}
```

---

## 四、实现检查清单

### 4.1 账号权限配置

- [ ] 创建第三方物流用户账号
- [ ] 配置用户角色（Stock User、Stock Manager 等）
- [ ] 配置公司权限（User Permission）
- [ ] 配置仓库权限（如果需要）
- [ ] 配置角色权限（Role Permission）
- [ ] 创建自定义角色（Logistics User，可选）
- [ ] 测试用户登录和权限

---

### 4.2 发货业务流程

- [ ] 验证总部可以创建送货单
- [ ] 验证第三方物流人员可以查看送货单列表
- [ ] 验证第三方物流人员可以打开送货单详情
- [ ] 验证第三方物流人员可以编辑送货单（草稿状态）
- [ ] 验证第三方物流人员可以提交送货单
- [ ] 验证提交后库存正确扣减
- [ ] 验证权限限制（只能操作有权限的公司和仓库）

---

### 4.3 扫码发货功能

- [ ] 在 ERPNext 中为物品设置条形码
- [ ] 验证条形码格式（EAN-13、UPC-A 等）
- [ ] 实现扫码打开送货单功能
- [ ] 实现扫码添加物品功能
- [ ] 配置扫码设备（扫码枪或手机）
- [ ] 测试扫码功能（扫码枪和手机摄像头）
- [ ] 优化扫码体验（自动聚焦、提示音等）
- [ ] 添加错误处理和提示信息

---

### 4.4 测试验证

- [ ] 测试第三方物流人员登录
- [ ] 测试查看待发货列表
- [ ] 测试扫码打开送货单
- [ ] 测试扫码添加物品
- [ ] 测试提交送货单
- [ ] 测试库存扣减
- [ ] 测试权限限制
- [ ] 测试错误场景（物品不存在、库存不足等）

---

## 五、注意事项

### 5.1 权限安全

1. **最小权限原则**：
   - ⚠️ 第三方物流人员只需要发货相关权限
   - ❌ 不要授予不必要的权限（如：创建销售订单、修改物品等）
   - ✅ 建议创建专门的 `Logistics User` 角色

2. **数据隔离**：
   - ⚠️ 确保第三方物流人员只能看到和操作有权限的公司和仓库数据
   - ✅ 通过 User Permission 和 Role Permission 控制数据访问范围

3. **操作审计**：
   - ✅ 建议启用 ERPNext 的审计日志功能
   - ✅ 记录第三方物流人员的所有操作（创建、编辑、提交送货单等）

---

### 5.2 业务流程规范

1. **送货单状态管理**：
   - ⚠️ 草稿状态的送货单可以编辑
   - ⚠️ 已提交的送货单无法再编辑
   - ✅ 建议在确认发货后再提交送货单

2. **库存验证**：
   - ⚠️ 提交送货单时会验证库存是否充足
   - ⚠️ 如果库存不足，无法提交
   - ✅ 需要提前检查库存，确保充足

3. **数量控制**：
   - ⚠️ 发货数量不能超过销售订单数量
   - ✅ 可以部分发货（发货数量 < 订单数量）

---

### 5.3 扫码功能优化

1. **条形码格式**：
   - ⚠️ 确保物品条形码格式正确（EAN-13、UPC-A 等）
   - ⚠️ 条形码必须唯一，不能重复
   - ✅ 建议使用标准格式

2. **扫码设备**：
   - ✅ 推荐使用扫码枪（USB 或蓝牙），体验更好
   - ✅ 如果使用手机扫码，需要启用摄像头权限
   - ✅ 确保扫码设备正常工作

3. **错误处理**：
   - ✅ 添加扫码错误提示（物品不存在、条形码格式错误等）
   - ✅ 提供手动输入备选方案
   - ✅ 记录扫码失败日志，便于排查问题

---

## 六、相关文档

- [ERPNext 条形码设置与 TTPOS 扫码收货功能](./erpnext-barcode-scan-receipt.md)
- [ERPNext 按文档类型设置不同权限范围配置指南](./erpnext-doctype-specific-permission.md)
- [ERPNext 仓库层级管理实现方案](./erpnext-warehouse-hierarchy.md)
- [ERPNext 用户权限文档](https://docs.erpnext.com/docs/user/manual/en/setting-up/users-and-permissions/user-permissions)
- [ERPNext Delivery Note 文档](https://docs.erpnext.com/docs/user/manual/en/stock/delivery-note)

---

## 🎯 总结

实现第三方物流接入 ERP 的扫码发货功能，需要：

1. ✅ **配置账号权限**：创建第三方物流用户，配置发货相关权限
2. ✅ **发货业务流程**：从创建送货单到提交发货的完整流程
3. ✅ **扫码发货功能**：支持扫码打开送货单、扫码添加物品等功能
4. ✅ **扫码设备配置**：配置扫码枪或手机扫码功能
5. ✅ **测试验证**：全面测试功能和权限

**关键点**：
- 最小权限原则，只授予必要的发货权限
- 完整的发货业务流程，从创建到提交
- 便捷的扫码功能，提高发货效率
- 完善的错误处理和权限控制

---

**最后更新**：2026-01-07  
**维护者**：TTPOS Team

