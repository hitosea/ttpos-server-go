# Opt-251223-001 优化方案

## 需求概述

优化新管理端支付管理与 ERP 系统对接，解决以下问题：

1. **ERP授权时支付方式创建范围不准确**：应仅在 ERP 中创建 Cash、Balance、Free Meal 三种基础支付方式
2. **新增支付方式时返回字段不正确**：应使用 `PaymentId` 替代 `Name` 作为唯一标识
3. **更新支付方式时参数不完整**：应同时传递 `Name` 和 `PaymentId`
4. **LIANLIANPAY支付配置成功后缺少ERP同步**：配置成功后应自动向 ERP 添加对应支付方式

## 问题分析

### 技术债务分析

**当前问题**：

1. **InitShop 支付方式同步逻辑不准确**
   - 当前：同步所有已创建的支付方式到 ERP
   - 问题：应该只同步基础支付方式（Cash、Balance、Free Meal）
   - 位置：`main/app/service/rpc/erp/setup.go:164-223`

2. **SaveModeOfPayment 返回值保存不完整**
   - 当前：只保存 `SaveModeOfPaymentResp.Name` 或 `SaveModeOfPaymentResp.PaymentId` 到 `erpnext_payment` 字段
   - 问题：应该同时保存 Name 到 `erpnext_payment` 字段，PaymentId 到新增的 `erpnext_payment_id` 字段
   - 位置：`main/app/service/rpc/erp/setup.go:220`、`main/app/service/payment_method.go:447-454`、`main/app/service/payment_method.go:534-551`

3. **SaveModeOfPayment 更新参数选择逻辑不完善**
   - 当前：更新时可能只传递 `Name` 或 `PaymentId`
   - 问题：应该优先使用 PaymentId（如果 `erpnext_payment_id` 不为空），否则使用 Name（`erpnext_payment`）
   - 位置：`main/app/service/payment_method.go:558-575`

4. **Free Meal 在支付管理列表中显示**
   - 当前：新管理端 `GetManagementList` 未过滤 Free Meal
   - 问题：Free Meal 不应该在支付管理中显示
   - 位置：`main/app/service/payment_method.go:182-245`
   - 旧后台：`admin/app/common/model/store/PayType.php:list()` 方法需要检查

5. **LIANLIANPAY 支付配置缺少 ERP 同步**
   - 当前：`UpdateLianlianPayConfig` 只保存配置，未同步支付方式到 ERP
   - 问题：配置成功后应创建对应的 Mode of Payment
   - 位置：`main/app/service/payment_method.go:605-663`

**维护成本**：

- 数据不一致风险：ERP 中存在过多支付方式，与实际业务不符
- 更新失败风险：使用 Name 更新可能导致 ERP 无法识别
- 功能不完整：LIANLIANPAY 配置后需要手动同步支付方式

**改进空间**：

- 优化 ERP 同步逻辑，只同步必要的基础支付方式
- 使用 PaymentId 作为唯一标识，提高数据一致性
- 完善 LIANLIANPAY 配置流程，自动同步支付方式

## 优化方案

### 方案对比

**方案 1: 最小改动方案**
- 优点：改动范围小，风险低
- 缺点：逻辑分散，后续维护困难
- 实施成本：低（1-2天）
- 预期收益：中等
- 风险：低

**方案 2: 完整重构方案**
- 优点：逻辑统一，易于维护
- 缺点：改动范围大，需要充分测试
- 实施成本：高（3-5天）
- 预期收益：高
- 风险：中

**✅ 最终选择: 方案 1（最小改动方案）**

理由：
- 当前问题明确，改动范围可控
- 风险低，不影响现有功能
- 可以快速上线，后续再优化

### 实施步骤

1. **优化 InitShop 支付方式同步逻辑**
   - 只同步基础支付方式（Cash、Balance、Free Meal for ERP）
   - 先判断 Free Meal for ERP（code=92000）是否存在，不存在先添加
   - 使用 PaymentId 保存到 `erpnext_payment` 字段
   - **注意**：Free Meal for ERP（code=92000）是专门用于ERP同步的，不改变原有的Free Meal（code=-1）

2. **优化 SaveModeOfPayment 方法**
   - 更新时同时传递 `Name` 和 `PaymentId`
   - 返回 `PaymentId` 供调用方保存

3. **过滤 Free Meal 显示**
   - 新管理端：在 `GetManagementList` 中过滤 Free Meal
   - 旧后台：在 `PayType::list()` 中过滤 Free Meal

4. **完善 LIANLIANPAY 配置流程**
   - 在 `UpdateLianlianPayConfig` 中，配置成功后同步支付方式到 ERP
   - 判断商家是否开启 ERP，未开启则跳过

### 技术方案

#### 1. InitShop 支付方式同步优化

**修改位置**：`main/app/service/rpc/erp/setup.go:164-223`

**实现逻辑**：

```go
// 1. 判断商家是否开启 ERP
if !company.IsOpenErp() || company.CompanySetting.ErpnextSiteCode == "" {
    // 跳过 ERP 同步
    return resp.InitShopResp{...}, nil
}

// 2. 定义基础支付方式列表
basePaymentCodes := []int{
    constant.PaymentMethodCodeCash,          // 40
    constant.PaymentMethodCodeBalance,       // 10
    constant.PaymentMethodCodeFreeMealForErp, // 92000（用于ERP同步的Free Meal）
}

// 3. 确保 Free Meal for ERP 存在（code=92000，不改变原有的Free Meal code=-1）
freeMealForErpPayment := paymentMethodRepo.GetPaymentMethod(
    paymentMethodRepo.WhereCode(constant.PaymentMethodCodeFreeMealForErp),
    commonRepo.WhereBySoftDelete(),
)
if freeMealForErpPayment.Uuid == 0 {
    // 创建 Free Meal for ERP 支付方式（code=92000）
    // ... 创建逻辑
}

// 4. 只同步基础支付方式
for _, code := range basePaymentCodes {
    paymentMethod := paymentMethodRepo.GetPaymentMethod(
        paymentMethodRepo.WhereCode(code),
        commonRepo.WhereBySoftDelete(),
    )
    if paymentMethod.Uuid == 0 {
        continue
    }
    
    // 如果已有 erpnext_payment，跳过
    if paymentMethod.ErpnextPayment != "" {
        continue
    }
    
    // 根据 source 获取 channel
    channel := erp.GetChannelBySource(paymentMethod.Source)
    
    // 同步到 ERP
    params := &selling.SaveModeOfPaymentReq{
        CompanyAbbr: initShopReq.CompanyAbbr,
        Branch:      response.BranchName,
        Channel:     channel,
        PayType:     paymentMethod.PaymentName,
    }
    saveResp, err := sellingClient.SaveModeOfPayment(...)
    // ... 错误处理 ...
    
    // 保存 Name 和 PaymentId
    saveModeOfPaymentRespMap[code] = repository.ErpnextPaymentInfo{
        Name:      saveModeOfPaymentResp.Name,
        PaymentId: saveModeOfPaymentResp.PaymentId,
    }
}
```

#### 2. SaveModeOfPayment 更新参数优化

**修改位置**：`main/app/service/rpc/erp/selling.go:483-496`

**实现逻辑**：

```go
params := &selling.SaveModeOfPaymentReq{
    CompanyAbbr: companySetting.ErpnextCompanyAbbr,
    Branch:      companySetting.ErpnextBranchName,
    PayType:     saveModeOfPaymentReq.PayType,
}

if saveModeOfPaymentReq.Channel != "" {
    params.Channel = saveModeOfPaymentReq.Channel
}

// 更新时优先使用 PaymentId，否则使用 Name
if saveModeOfPaymentReq.PaymentId != nil && *saveModeOfPaymentReq.PaymentId != "" {
    params.PaymentId = *saveModeOfPaymentReq.PaymentId
} else if saveModeOfPaymentReq.Name != nil {
    params.Name = saveModeOfPaymentReq.Name
}

if saveModeOfPaymentReq.Enabled != nil {
    params.Enabled = saveModeOfPaymentReq.Enabled
}
```

#### 3. Free Meal 过滤优化

**新管理端修改位置**：`main/app/service/payment_method.go:182-245`

**实现逻辑**：

```go
excludeCodes := []int{
    constant.PaymentMethodCodeGrab,
    constant.PaymentMethodCodeLineMan,
    constant.PaymentMethodCodeFreePay, // 新增：过滤 Free Meal
}
```

**旧后台修改位置**：`admin/app/common/model/store/PayType.php:138-140`

**实现逻辑**：

```php
public static function list($shopSupplierId = 0, $appId = 0)
{
    $LianLian_enable = (bool)PaymentService::checkSignSalt($shopSupplierId);
    
    return self::where('shop_supplier_id', $shopSupplierId)
        ->where('app_id', $appId)
        ->where('code', '<>', OrderPayTypeEnum::FREE_PAY) // 新增：过滤免单
        ->orderRaw('CAST(sort AS UNSIGNED)')
        ->order('create_time', 'desc')
        ->select()
        ->toArray();
}
```

#### 4. LIANLIANPAY 配置 ERP 同步

**修改位置**：`main/app/service/payment_method.go:605-663`

**实现逻辑**：

```go
func (s *paymentMethodSrv) UpdateLianlianPayConfig(ctx context.Context, configReq *req.LianlianPayConfigUpdateReq) error {
    // ... 现有配置保存逻辑 ...
    
    // 配置保存成功后，同步支付方式到 ERP
    company := ctx.GetCompany()
    if company.IsOpenErp() && company.CompanySetting.ErpnextSiteCode != "" {
        erpSrv := erpService.NewIErpSrv(s.dbm)
        
        // 定义 LIANLIANPAY 支付方式 code 到 PayType 名称的映射
        lianlianPayMethodsMap := map[int]string{
            constant.PaymentMethodCodeLianLianWechatPay:   "Wechat Pay",
            constant.PaymentMethodCodeLianLianAliPay:       "Alipay",
            constant.PaymentMethodCodeLianLianQRPromptPay:  "QR PromptPay",
        }
        
        paymentMethodRepo := repository.NewPaymentMethodRepo(s.dbm.GetDB(ctx.GetCompanyUuid()))
        
        // 批量查询 LIANLIANPAY 支付方式（source=2）且未同步到 ERP 的（erpnext_payment 为空）
        paymentMethods := paymentMethodRepo.GetPaymentMethodList(
            repository.CommonRepo.WhereBySource(constant.PaymentMethodSourceLianLianPay),
            paymentMethodRepo.WhereNotExistsErpnextPayment(),
            repository.CommonRepo.WhereBySoftDelete(),
        )
        
        for _, paymentMethod := range paymentMethods {
            // 获取对应的 PayType 名称
            payTypeName, exists := lianlianPayMethodsMap[paymentMethod.Code]
            if !exists {
                continue // 不在映射表中的支付方式，跳过
            }
            
            // 根据支付方式的 source 获取 channel
            channel := erp.GetChannelBySource(paymentMethod.Source)
            
            // 同步到 ERP
            saveResp, err := erpSrv.SaveModeOfPayment(ctx, req.SaveModeOfPaymentReq{
                CompanyUuid: ctx.GetCompanyUuid(),
                Channel:     channel,
                PayType:     payTypeName,
            })
            
            if err != nil || saveResp == nil {
                logger.Logger.Error("同步 LIANLIANPAY 支付方式到 ERP 失败",
                    zap.Error(err),
                    zap.String("pay_type", payTypeName),
                    zap.Int("code", paymentMethod.Code),
                    zap.Uint64("uuid", paymentMethod.Uuid),
                )
                continue
            }
            
            // 保存 PaymentId
            if saveResp.PaymentId != "" {
                if err := paymentMethodRepo.UpdatePaymentMethod(
                    map[string]any{"erpnext_payment": saveResp.PaymentId},
                    repository.CommonRepo.WhereByUuid(paymentMethod.Uuid),
                ); err != nil {
                    logger.Logger.Error("更新 LIANLIANPAY 支付方式 ERP PaymentId 失败",
                        zap.Error(err),
                        zap.String("pay_type", payTypeName),
                        zap.Uint64("uuid", paymentMethod.Uuid),
                    )
                }
            }
        }
    }
    
    return nil
}
```

## 收益评估

### 数据准确性提升

- **ERP 支付方式数量**：从"所有支付方式" → "仅基础支付方式"（减少 70-90%）
- **数据一致性**：使用 PaymentId 作为唯一标识，避免名称变更导致的问题
- **同步准确性**：LIANLIANPAY 配置后自动同步，避免手动操作遗漏

### 可维护性提升

- **代码清晰度**：明确区分基础支付方式和自定义支付方式的处理逻辑
- **错误率降低**：使用 PaymentId 更新，降低因名称变更导致的更新失败
- **功能完整性**：LIANLIANPAY 配置流程完整，减少人工干预

### 用户体验改善

- **支付管理界面**：不显示 Free Meal，界面更清晰
- **配置流程**：LIANLIANPAY 配置后自动同步，无需手动操作

## 影响分析

### 兼容性

- **数据库兼容**：`erpnext_payment` 字段存储 PaymentId 而非 Name，需要兼容现有数据
- **API 兼容**：SaveModeOfPayment 方法参数增加 PaymentId，向后兼容（可选参数）

### 风险评估

| 风险项 | 风险等级 | 影响范围 | 应对措施 |
|--------|---------|---------|---------|
| 现有数据兼容 | 中 | ERP 同步数据 | 迁移脚本：将现有 Name 转换为 PaymentId |
| ERP 同步失败 | 低 | 新商家授权 | 错误日志记录，不影响主流程 |
| Free Meal 过滤遗漏 | 低 | 支付管理列表 | 充分测试新管理端和旧后台 |

### 回滚方案

1. **代码回滚**：Git 回滚到优化前版本
2. **数据回滚**：如有数据迁移，提供回滚脚本
3. **ERP 数据清理**：手动清理 ERP 中多余的支付方式（如需要）

## 测试计划

### 功能测试

**测试用例 1：ERP 授权时基础支付方式同步**
- Given: 商家授权 ERP
- When: InitShop 执行
- Then: 
  - 仅 Cash、Balance、Free Meal for ERP（code=92000）同步到 ERP
  - Free Meal for ERP 如果不存在，先创建再同步
  - `erpnext_payment` 字段保存 PaymentId
  - 原有的 Free Meal（code=-1）保持不变

**测试用例 2：支付方式列表不显示 Free Meal**
- Given: 新管理端/旧后台查询支付方式列表
- When: 调用列表接口
- Then: Free Meal 不在列表中

**测试用例 3：LIANLIANPAY 配置后自动同步**
- Given: 商家开启 ERP，配置 LIANLIANPAY
- When: UpdateLianlianPayConfig 执行成功
- Then: 
  - Wechat Pay、Alipay、QR PromptPay 同步到 ERP
  - `erpnext_payment` 字段保存 PaymentId

**测试用例 4：更新支付方式传递 Name 和 PaymentId**
- Given: 更新支付方式
- When: SaveModeOfPayment 执行
- Then: 同时传递 Name 和 PaymentId 参数

### 回归测试

- ERP 授权流程正常
- 支付方式列表查询正常
- LIANLIANPAY 配置流程正常
- 支付方式更新功能正常

### 灰度发布

- **灰度策略**：先发布到测试环境，验证通过后全量发布
- **监控指标**：
  - ERP 同步成功率
  - 支付方式列表查询响应时间
  - LIANLIANPAY 配置成功率

## 上线计划

### 发布时间

- **测试环境**：2025-12-24
- **生产环境**：2025-12-25

### 监控指标

- ERP 同步错误日志
- 支付方式列表查询性能
- LIANLIANPAY 配置成功率

### 应急预案

- **ERP 同步失败**：记录错误日志，不影响主流程
- **数据不一致**：提供数据修复脚本
- **性能问题**：回滚代码

## 经验沉淀

**优化方法**：
- 明确区分基础支付方式和自定义支付方式的处理逻辑
- 使用 PaymentId 作为唯一标识，提高数据一致性
- 完善配置流程，减少人工干预

**关键技术**：
- ERP 同步逻辑优化
- PaymentId 字段使用
- Free Meal 过滤逻辑
- Free Meal for ERP（code=92000）与原有 Free Meal（code=-1）分离

**注意事项**：
- 确保商家开启 ERP 后再执行同步
- 兼容现有数据（Name → PaymentId 迁移）
- 充分测试新管理端和旧后台的列表查询

**适用场景**：
- ERP 数据同步优化
- 支付方式管理优化
- 配置流程自动化

---

**创建时间**: 2025-12-23 17:10  
**方案状态**: 🟢 规划中

