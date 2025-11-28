# Bug-251128-004: 手机端保存单位关联商品时提示错误，保存不了

## 基本信息

| 字段       | 值                    |
| ---------- | --------------------- |
| Bug ID     | bug-251128-004        |
| 模块       | product               |
| 严重程度   | medium                |
| 发现版本   | v2.10.9               |
| 发现日期   | 2025-11-28            |
| 发现者     | 王昱                  |
| 状态       | 🟡 规划中             |

## 问题描述

### 现象

手机端（Shop 商家管理端）在编辑商品单位并保存关联商品时，系统提示错误，无法成功保存。

### 复现步骤

1. 登录手机端商家管理后台
2. 进入商品管理 → 单位管理
3. 编辑某个商品单位
4. 尝试关联商品（选择 `ProductPackageUuids`）
5. 点击保存
6. **结果**：系统提示错误，保存失败

### 预期行为

- 手机端应该能够正常保存单位关联商品
- 系统应该在校验通过后成功更新商品单位的关联关系

### 实际行为

- 保存操作失败，提示错误信息
- 无法完成单位关联商品的更新

### 错误原因分析

根据业务规则，存在以下限制：

1. **ERP 上普通商品不可修改对应变体上的单位**
   - 当商品是普通商品（非套餐）时，如果该商品在 ERP 中已有对应的变体（Item Variant），则不允许修改变体上的单位（Stock UOM）

2. **ERP 上已销售商品单位不可修改**
   - 如果商品已经在 ERP 中有销售记录（已产生订单），则不允许修改该商品的单位

### 相关代码位置

**主要涉及文件**：
- `main/app/service/product.go` - `EditProductUnit` 方法（第 2219-2353 行）
- `main/app/api/v1/shop/shop_product.go` - `EditProductUnit` API 接口（第 278-298 行）

**问题代码段**：

```2274:2343:main/app/service/product.go
		if len(editUnitReq.ProductPackageUuids) > 0 {
			// 同步更新商品到erp
			if ctx.GetCompany().IsOpenErp() {
				productPackageRepo := repository.NewProductPackageRepo(tx)
				productPackages, errGetProductPackage := productPackageRepo.GetProductPackageList(
					repository.CommonRepo.WhereInUuids(editUnitReq.ProductPackageUuids),
					repository.CommonRepo.Preload(
						repository.WithPreload{
							Query: "ProductBoms",
						},
					),
				)
				if errGetProductPackage != nil {
					return errors.WithMessage(errGetProductPackage, "获取商品包失败")
				}
				for i := range productPackages {
					productPackage := productPackages[i]
					for _, productBom := range productPackage.ProductBoms {
						if productBom.IsFlavor() {
							if productPackage.IsPackage() {
								// 同步套餐到erp
								multiLanguageName := model.NewMultiLanguageName(productBom.Name)
								enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
								if err != nil {
									return errors.WithMessage(err, "翻译失败")
								}
								erpSrv := erp.NewIErpSrv(s.dbm)
								_, errErp := erpSrv.AddPackage(ctx, req.PackageAddErpReq{
									ItemName: enName,
									StockUom: productUnit.ErpnextUom,
									ItemCode: productBom.ErpCode,
								})
								if errErp != nil {
									return errors.WithMessage(errErp, "同步商品到erp失败")
								}
							} else if productPackage.IsProduct() {
								// 同步商品到erp
								productMultiLanguageName := model.NewMultiLanguageName(productPackage.Name)
								productEnName, err := s.getEnName(ctx, productMultiLanguageName.GetNames())
								if err != nil {
									return errors.WithMessage(err, "翻译失败")
								}
								multiLanguageName := model.NewMultiLanguageName(productBom.Name)
								enName, err := s.getEnName(ctx, multiLanguageName.GetNames())
								if err != nil {
									return errors.WithMessage(err, "翻译失败")
								}
								erpSrv := erp.NewIErpSrv(s.dbm)
								itemName := fmt.Sprintf("%s-%s", productEnName, enName)
								_, errErp := erpSrv.AddProduct(ctx, req.ProductAddErpReq{
									ItemName: itemName,
									StockUom: productUnit.ErpnextUom,
									ItemCode: productBom.ErpCode,
								})
								if errErp != nil {
									return errors.WithMessage(errErp, "同步商品到erp失败")
								}
							}
						}
					}
				}
			}
			// 修改商品的单位UUID
			err = tx.Model(&model.ProductPackage{}).Where("uuid in (?)", editUnitReq.ProductPackageUuids).Updates(map[string]any{
				"unit_uuid": productUnit.Uuid,
			}).Error
			if err != nil {
				return errors.WithMessage(errors.New("保存关联商品失败"), err.Error())
			}
		}
```

**问题分析**：

1. 代码在同步到 ERP 时，直接调用 `erpSrv.AddProduct` 或 `erpSrv.AddPackage` 更新单位，但没有先检查：
   - 商品是否为普通商品且已有变体（不允许修改变体单位）
   - 商品是否已在 ERP 中有销售记录（不允许修改已销售商品单位）

2. ERP 接口可能返回错误，但错误信息可能不够明确，导致用户无法理解失败原因

## 环境信息

- **终端**: Shop 商家管理端（手机端）
- **API 路径**: `/shop/product/unit/edit`
- **服务模块**: Main 模块（Go + Gin）
- **ERP 集成**: 已开启 ERP 同步

## 影响范围

### 涉及终端

- [x] Shop 商家管理端（手机端）
- [ ] POS 收银端
- [ ] KDS 厨显端
- [ ] 其他终端

### 涉及功能

- 商品单位编辑功能
- 单位关联商品功能
- ERP 商品同步功能

### 业务影响

- **用户体验**：手机端无法正常保存单位关联商品，影响商品管理效率
- **数据一致性**：可能导致 TTPOS 与 ERP 数据不一致
- **操作限制**：商户需要切换到 PC 端或其他方式完成操作

## 初步分析

### 技术栈识别

- **模块**: Main 模块（Go + Gin）
- **服务层**: `main/app/service/product.go`
- **API 层**: `main/app/api/v1/shop/shop_product.go`
- **ERP 集成**: `ttpos-bmp` 模块

### 可能原因

1. **缺少业务规则校验**：在同步到 ERP 前，没有检查商品是否允许修改单位
2. **错误处理不完善**：ERP 返回的错误信息可能不够友好，没有转换为用户可理解的提示
3. **ERP 接口限制**：ERPNext 可能对已销售商品或变体商品的单位修改有严格限制

### 需要进一步调查

1. 确认 ERP 接口返回的具体错误信息
2. 确认如何判断商品是否已销售（查询订单记录）
3. 确认如何判断商品是否为普通商品且已有变体
4. 确认 ERPNext 对单位修改的具体限制规则

## 相关链接

### 相关提案

- [ERP UpdateProduct 增加 UOM 字段更新支持](../../team/proposals/2025-11/erp-update-product-uom.md)

### 相关代码

- `main/app/service/product.go` - `EditProductUnit` 方法
- `main/app/api/v1/shop/shop_product.go` - `EditProductUnit` API
- `ttpos-bmp/app/ttpos-erp/internal/logic/stock/product.go` - ERP 商品更新逻辑

### 相关文档

- ERP 集成文档（待补充）

### 修复方案和任务

- [修复方案](solution.md)
- [任务清单](tasks.md)

## 修复方案（简要）

### 修复步骤

#### 1. 单位关联商品/套餐时的 ERP 同步逻辑调整

**问题**：当前代码在编辑单位关联商品时，对普通商品使用 `AddProduct` 方法，对套餐使用 `AddPackage` 方法，但这些方法不适合更新已存在商品的单位。

**修复方案**：
- **普通商品**：调用 `erpSrv.UpdateProduct` 方法修改 ERP 上模板单位（Item Template 的 `stock_uom`）
- **套餐**：调用 `erpSrv.UpdateProduct` 方法更新套餐单位

**涉及代码位置**：
- `main/app/service/product.go` - `EditProductUnit` 方法（第 2274-2343 行）
- `main/app/service/product.go` - `AddProductUnit` 方法（第 2108-2147 行）

**实现要点**：
- 普通商品：使用 `productPackage.ErpCode`（模板 Item Code）调用 `UpdateProduct` 更新模板单位
- 套餐：使用 `productBom.ErpCode`（套餐 Item Code）调用 `UpdateProduct` 更新单位
- 需要先确保 Main 模块的 `UpdateProductReq` 结构体支持 `StockUom` 字段

#### 2. 涉及业务场景同步修复

需要同步修复以下业务场景中的单位关联逻辑：

- **新增单位** (`AddProductUnit`)：第 2108-2147 行
  - 新增单位时关联商品，需要同步更新 ERP 模板单位

- **新增商品** (`AddProduct`)：需要检查是否有单位关联逻辑
  - 如果新增商品时指定了单位，需要同步到 ERP

- **编辑商品** (`EditProduct`)：需要检查单位修改逻辑
  - 如果编辑商品时修改了单位，需要同步更新 ERP

#### 3. ERP 错误处理

**当前问题**：ERP 接口返回错误时，错误信息可能不够明确，且没有记录详细日志。

**修复方案**：
- 当 ERP 接口返回错误时，记录详细的错误日志（包含商品信息、单位信息、错误详情）
- 根据错误类型决定是否中断事务：
  - **可恢复错误**（如网络超时）：记录日志，允许继续执行
  - **业务规则错误**（如已销售商品不允许修改单位）：记录日志，返回明确的错误提示给用户
  - **系统错误**（如 ERP 服务不可用）：记录日志，返回系统错误提示

**日志记录格式**：
```go
ctx.Log().Error("同步商品单位到ERP失败",
    zap.String("product_package_uuid", productPackage.Uuid),
    zap.String("product_name", productPackage.Name),
    zap.String("erp_code", productPackage.ErpCode),
    zap.String("unit_uuid", productUnit.Uuid),
    zap.String("unit_name", productUnit.Name),
    zap.String("erp_uom", productUnit.ErpnextUom),
    zap.Error(errErp))
```

### 技术实现要点

1. **UpdateProductReq 结构体扩展** ⚠️ **前置条件**
   - BMP 模块的 `item.UpdateProductReq`（gRPC）已支持 `StockUom` 字段 ✅
   - **需要更新** Main 模块的 `main/app/service/rpc/erp/product.go` 中的 `UpdateProductReq` 结构体：
     ```go
     type UpdateProductReq struct {
         ItemCode     string                `json:"item_code"`
         NotForSale   bool                  `json:"not_for_sale"`
         InternalCode string                `json:"internal_code"`
         Disabled     bool                  `json:"disabled"`
         Attributes   []UpdateProductFlavor `json:"attributes"`
         StockUom      string                `json:"stock_uom"` // ✅ 需要添加此字段
     }
     ```
   - **需要更新** `UpdateProduct` 方法，传递 `StockUom` 字段到 gRPC 调用：
     ```go
     result, err := client.UpdateProduct(..., &item.UpdateProductReq{
         ItemCode:     params.ItemCode,
         NotForSale:   params.NotForSale,
         InternalCode: params.InternalCode,
         Disabled:     params.Disabled,
         Attributes:   attributes,
         StockUom:     params.StockUom, // ✅ 需要添加此字段
     })
     ```

2. **商品类型判断**
   - 使用 `productPackage.IsProduct()` 判断是否为普通商品
   - 使用 `productPackage.IsPackage()` 判断是否为套餐

3. **ERP Code 获取**
   - 普通商品：使用 `productPackage.ErpCode`（模板 Item Code）
   - 套餐：使用 `productBom.ErpCode`（套餐 Item Code）

4. **事务处理**
   - 保持现有事务逻辑
   - ERP 同步失败时，根据错误类型决定是否回滚

### 测试要点

1. **普通商品单位修改**
   - 验证普通商品关联单位时，调用 `UpdateProduct` 更新模板单位
   - 验证 ERP 中模板 Item 的 `stock_uom` 字段正确更新

2. **套餐单位修改**
   - 验证套餐关联单位时，调用 `UpdateProduct` 方法更新单位
   - 验证套餐逻辑不受影响

3. **错误处理**
   - 验证 ERP 错误时，日志正确记录
   - 验证错误提示用户友好

4. **多场景覆盖**
   - 新增单位关联商品
   - 编辑单位关联商品
   - 新增商品指定单位
   - 编辑商品修改单位

---

**下一步行动**：

1. ✅ 技术分析：已完成，已明确问题原因和修复方向
2. ✅ 制定修复方案：已完成，已明确修复步骤和技术要点
3. ⏳ 实施修复：需要开发人员按照修复方案进行代码修改
4. ⏳ 测试验证：修复后需要进行完整的功能测试和回归测试

