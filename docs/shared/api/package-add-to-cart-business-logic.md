# 套餐加购到购物车业务逻辑文档

## 概述

本文档详细描述了从收银端（POS）将套餐商品添加到购物车的完整业务逻辑流程，包括请求入口、参数处理、商品创建、价格计算等核心环节。

## 业务流程图

```
前端请求
  ↓
API Handler (cashier_desk.go)
  ↓
Service Layer (order_product.go)
  ├─ 检查/创建销售账单
  ├─ 加锁防并发
  ├─ 构建套餐商品参数
  └─ 调用通用加购方法
      ↓
Action Layer (order_action.go)
  ├─ 验证订单状态
  ├─ 创建订单商品
  └─ 计算并保存订单
      ↓
Order Service (order.go)
  ├─ 创建套餐主商品
  └─ 创建套餐子商品
      ↓
返回购物车信息
```

## 详细流程分析

### 1. API 入口层

**文件位置**: `main/app/api/v1/cashier/cashier_desk.go`

**方法**: `OrderCartProductPackageAdd`

```773:800:main/app/api/v1/cashier/cashier_desk.go
// OrderCartProductPackageAdd 向购物车添加套餐
// @Summary 向购物车添加套餐
// @Description 向购物车添加套餐
// @Tags 收银端.桌台
// @Accept json
// @Produce json
// @Security JwtToken
// @param data body req.OrderCartProductPackageAddReq true "套餐参数"
// @Success 200 {object} dto.Response{data=resp.ShopCart}
// @Failure 404 {object} nil "未找到"
// @Router /cashier/desk/order/cart/product_package/add [post]
func (h *DeskHandler) OrderCartProductPackageAdd(c *gin.Context) {
	ctx := helper.GetContext(c)
	// 绑定请求参数
	params := req.OrderCartProductPackageAddReq{}
	if err := c.ShouldBindJSON(&params); err != nil {
		helper.HandleValidationError(c, err, params, req.OrderReqMessage)
		return
	}
	// 向购物车添加套餐
	res, err := h.orderSrv.OrderCartProductPackageAdd(ctx, params)
	if err != nil {
		helper.ErrorWithDetail(c, constant.CodeFail, errors.WithMessage(err))
		return
	}
	// 返回结果
	helper.Success(c, res)
}
```

**请求参数结构**:

```10:35:main/app/dto/req/shop_cart.go
// OrderCartProductPackageAddReq 向购物车添加套餐请求参数
type OrderCartProductPackageAddReq struct {
	SaleBillUuid       uint64           `json:"sale_bill_uuid"`       // 销售账单UUID
	SaleOrderUuid      uint64           `json:"sale_order_uuid"`      // 销售订单UUID
	ProductPackageUuid uint64           `json:"product_package_uuid"` // 套餐UUID
	Products           []ProductRequest `json:"products"`             // 套餐商品请求列表

	isH5Product bool `json:"-"` // 是否是H5端下单的商品
}

func (req *OrderCartProductPackageAddReq) SetIsH5Product() {
	req.isH5Product = true
}

func (req *OrderCartProductPackageAddReq) IsH5Product() bool {
	return req.isH5Product
}

// ProductRequest 套餐商品请求参数
type ProductRequest struct {
	ProductPackageGroupUuid uint64 `json:"product_package_group_uuid"` // 套餐分组UUID
	// 普通商品的参数
	EditProductReq
	Num     float64 `json:"num"`      // 商品数量
	UnitNum float64 `json:"unit_num"` // 一个套餐的单个子商品的数量
}
```

**职责**:
- 接收前端请求并绑定参数
- 调用 Service 层方法
- 处理错误并返回响应

---

### 2. Service 层 - 套餐加购处理

**文件位置**: `main/app/service/order_product.go`

**方法**: `OrderCartProductPackageAdd`

```1866:1934:main/app/service/order_product.go
// OrderCartProductPackageAdd 往购物车添加套餐
func (s *orderSrv) OrderCartProductPackageAdd(ctx context.Context, request req.OrderCartProductPackageAddReq) (*resp.ShopCart, error) {

	// 当不填销售账单ID时，表示要新建一个销售账单
	if request.SaleBillUuid == 0 {
		// 判断是否有待支付、未挂单的订单
		billInfo, hasInstantOrder, err := HasInstantOrder(ctx, s.dbm.GetDB(ctx.GetDbId()))
		if err != nil {
			return nil, err
		}
		if billInfo != nil && hasInstantOrder {
			request.SaleBillUuid = billInfo.Uuid
			request.SaleOrderUuid = billInfo.SaleOrders[0].Uuid
		} else {
			order, err := s.CreateInstantOrder(ctx)
			if err != nil {
				ctx.Log().Info("添加商品时点餐订单创建失败", zap.Any("err", err.Error()))
				return nil, errors.WithMessage(err)
			}
			ctx.Log().Debug("添加商品时点餐订单创建成功", zap.Any("order info", order))
			request.SaleBillUuid = order.SaleBillUuid
			request.SaleOrderUuid = order.SaleOrderUuid
		}
	}

	// 上锁防止并发操作
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)

	// 往销售账单里添加商品
	productParam := req.ProductParams{
		FlavorProductBomUuid: request.ProductPackageUuid,
		Num:                  1,
		Operation:            "add",
	}
	// 记录相关的子商品。
	subProducts := make([]req.ProductParams, 0)
	for _, productReq := range request.Products {
		subProduct := req.ProductParams{
			FlavorProductBomUuid:            productReq.FlavorUuid,
			Num:                             productReq.Num,
			ProductPackageAttributeUuidList: productReq.AttributeUuidList,
			ProductPackageGroupUuid:         productReq.ProductPackageGroupUuid,
			Operation:                       "add",
		}
		subProducts = append(subProducts, subProduct)
	}
	productParam.SetIsPackageProduct(subProducts) // 设置为套餐商品

	shopCart, err := s.OrderCartProductAdd(ctx, req.ProductAddReq{
		SaleBillUuid:  request.SaleBillUuid,
		SaleOrderUuid: request.SaleOrderUuid,
		Products: []req.ProductParams{
			productParam,
		},
		IsH5Product: request.IsH5Product(),
	})
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return shopCart, nil
}
```

**核心处理步骤**:

1. **销售账单检查/创建**
   - 如果 `SaleBillUuid` 为 0，表示需要创建新的销售账单
   - 先检查是否存在待支付、未挂单的订单
   - 如果不存在，则创建新的即时订单

2. **并发控制**
   - 使用分布式锁（`lock.LockUuid`）防止同一销售账单的并发操作
   - 确保数据一致性

3. **构建套餐商品参数**
   - 创建主商品参数（`ProductParams`），设置套餐 UUID
   - 遍历子商品列表，构建子商品参数数组
   - 调用 `SetIsPackageProduct` 标记为套餐商品

4. **调用通用加购方法**
   - 将套餐商品参数转换为通用的 `ProductAddReq`
   - 调用 `OrderCartProductAdd` 方法（与普通商品共用）

---

### 3. Service 层 - 通用商品加购

**文件位置**: `main/app/service/order_product.go`

**方法**: `OrderCartProductAdd`

```113:154:main/app/service/order_product.go
// OrderCartProductAdd 向购物车添加商品
func (s *orderSrv) OrderCartProductAdd(ctx context.Context, request req.ProductAddReq, opts ...repository.OrderCartInfoOptionFunc) (*resp.ShopCart, error) {
	if ctx.NoLock() {
		s.lock.LockUuid(request.SaleBillUuid)
		defer s.lock.UnlockUuid(request.SaleBillUuid)
		ctx.AddLock()
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	// 获取销售账单信息
	saleBill, errSaleBill := repository.NewOrderRepo(db).GetSaleBillAllInfo(request.SaleBillUuid)
	if errSaleBill != nil {
		return nil, errors.WithMessage(errSaleBill)
	}
	// 判断订单状态
	if ctx.GetSource() == constant.SourceAssistant {
		if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderAddProduct, request.SaleOrderUuid, model.WithIsAssistant()); err != nil {
			return nil, errors.WithMessage(err)
		}
	} else {
		if err := saleBill.ValidateOrderStatus(ctx.GetSource(), constant.OrderAddProduct, request.SaleOrderUuid); err != nil {
			return nil, errors.WithMessage(err)
		}
	}

	// 设置添加来源
	saleBill.SetOperateSource(ctx.GetSource())

	// 加购
	if err := s.ActionAdd(ctx, request, saleBill); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取新的购物车商品数据
	info, err := s.GetOrderCartInfo(ctx, request.SaleBillUuid, opts...)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	return info, nil
}
```

**核心处理步骤**:

1. **再次加锁**（双重检查）
2. **获取销售账单完整信息**
3. **验证订单状态**（确保可以添加商品）
4. **设置操作来源**
5. **调用 ActionAdd 执行加购**
6. **获取更新后的购物车信息**

---

### 4. Action 层 - 执行加购操作

**文件位置**: `main/app/service/order_action.go`

**方法**: `ActionAdd`

```399:425:main/app/service/order_action.go
// ActionAdd 加购
func (s *orderSrv) ActionAdd(ctx context.Context, request req.ProductAddReq, saleBill *model.SaleBill) error {
	db := ctx.GetDB()

	var err error
	if request.IsMemberAdd {
		saleBill, err = s.actionAdd(ctx, request, saleBill, WithIsMemberAdd())
		if err != nil {
			return errors.WithMessage(err)
		}
	} else {
		saleBill, err = s.actionAdd(ctx, request, saleBill)
		if err != nil {
			return errors.WithMessage(err)
		}
	}

	if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
		if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
			return errors.WithMessage(err)
		}
		return nil
	}); err != nil {
		return errors.WithMessage(err)
	}
	return nil
}
```

**内部方法**: `actionAdd`

```742:788:main/app/service/order_action.go
// 加购。内部方法复用
func (s *orderSrv) actionAdd(ctx context.Context, request req.ProductAddReq, saleBill *model.SaleBill, options ...func(option *ActionAddOption)) (*model.SaleBill, error) {
	option := &ActionAddOption{}
	for _, optionFunc := range options {
		optionFunc(option)
	}

	// 检查销售订单商品数量是否超过1000项
	if err := saleBill.CheckSaleOrderProductNum(); err != nil {
		return nil, errors.WithMessage(err)
	}

	// 获取当前销售订单信息
	saleOrder := saleBill.GetSaleOrder(request.SaleOrderUuid)
	if saleOrder == nil {
		return nil, errors.New("销售订单不存在")
	}
	// 录入订单商品数据
	saleOrderProducts, err := s.newSaleOrderProduct(ctx, CreateSaleOrderProductParams{
		IsH5Product: request.IsH5Product,
		Setting:     *saleBill.SaleBillSetting,
		SaleBill:    saleBill,
		SaleOrder:   saleOrder,
		Products:    request.Products,
	}, options...)
	if err != nil {
		return nil, errors.WithMessage(err, "构建商品失败")
	}

	if !option.skipLimit {
		// 检查限购
		if request.IsH5Product == true {
			// 如果是h5端加购的话
			if err := s.checkLimitPurchase(ctx, saleBill, saleOrderProducts, model.WithH5CheckLimit()); err != nil {
				return nil, errors.WithMessage(err)
			}
		} else {
			if err := s.checkLimitPurchase(ctx, saleBill, saleOrderProducts); err != nil {
				return nil, errors.WithMessage(err)
			}
		}
	}
	// 检查超时不能加购
	if err := s.checkTimeoutAndCannotAddPurchase(ctx, saleBill, saleOrderProducts); err != nil {
		return nil, errors.WithMessage(err)
	}
	// saleBill已经加入了新的商品，并且重新计算了价格
```

**核心处理步骤**:

1. **检查商品数量限制**（不超过 1000 项）
2. **获取销售订单**
3. **创建订单商品**（调用 `newSaleOrderProduct`）
4. **检查限购规则**（如果未跳过）
5. **检查超时限制**
6. **在事务中计算并保存订单**

---

### 5. Order Service - 创建订单商品

**文件位置**: `main/app/service/order.go`

**方法**: `newSaleOrderProduct`

**核心逻辑**:

```1616:2009:main/app/service/order.go
func (s *orderSrv) newSaleOrderProduct(ctx context.Context, params CreateSaleOrderProductParams, options ...func(option *ActionAddOption)) ([]*model.SaleOrderProduct, error) {
	option := &ActionAddOption{}
	for _, opt := range options {
		opt(option)
	}

	innerParams := InnerParams{
		IsDeskSaleBill:         params.SaleBill.IsDeskSaleBill(),
		SaleBillUuid:           params.SaleBill.Uuid,
		SaleOrderUuid:          params.SaleOrder.Uuid,
		DeskUuid:               params.SaleBill.DeskUuid,
		DiningMethod:           params.SaleBill.DiningMethod,
		MemberDiscountRate:     params.SaleOrder.MemberDiscountRate,
		MemberCardDiscountRate: params.SaleOrder.MemberCardDiscountRate,
		CustomDiscountRate:     params.SaleOrder.CustomDiscountRate,
	}

	// 获取门店业务设置
	businessSetting, err := s.settingSrv.GetBusinessSetting(ctx)
	if err != nil {
		return nil, errors.WithMessage(err)
	}

	db := s.dbm.GetDB(ctx.GetDbId())
	ctx.SetDB(db)
	saleOrderProducts := make([]*model.SaleOrderProduct, 0)
	for _, product := range params.Products {
		// 获取商品包信息
		productBom, err := repository.NewProductPackageRepo(db).GetProductPackageBaseInfoByBomUuid(product.FlavorProductBomUuid)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		productPackage := productBom.ProductPackage
		productName := productPackage.MultiLanguageName.GetNameByLang(ctx.GetLanguage())

		if productBom.IsDelete() {
			return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "商品规格已经删除")))
		}
		// 商品已经下架
		if productBom.IsProductPackageDown() {
			return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "商品已经下架")))
		}

		// 获取某商品规格信息
		flavorProductBom, errFlavorProductBom := repository.NewProductBomRepo(db).GetFlavorProductBomByUuid(product.FlavorProductBomUuid)
		if errFlavorProductBom != nil {
			return nil, errors.WithMessage(errFlavorProductBom)
		}
		if flavorProductBom.GetStockNum() < float64(product.Num) {
			return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "库存不足")))
		}
		// 如果商品规格关联了材料，检查材料库存是否充足
		if len(flavorProductBom.FlavorMaterials) > 0 {
			for _, flavorMaterial := range flavorProductBom.FlavorMaterials {
				if flavorMaterial.IsDelete() {
					continue
				}
				materialStockNum := flavorMaterial.Material.GetStockNum()
				if materialStockNum < flavorMaterial.GetDecreaseNum(product.Num) {
					return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "库存不足")))
				}
			}
		}

		// 获取加料信息
		sauceProductBoms := make(map[uint64]*model.ProductBom)
		if len(product.SauceProductBomUuidList) > 0 {
			sauceProductBomList, errSauceProductBomList := repository.NewProductBomRepo(db).GetSauceProductBomsByUuids(product.SauceProductBomUuidList)
			if errSauceProductBomList != nil {
				return nil, errors.WithMessage(errSauceProductBomList)
			}
			if len(sauceProductBomList) != len(product.SauceProductBomUuidList) {
				sauceUuidMap := make(map[uint64]struct{})
				for _, uuid := range product.SauceProductBomUuidList {
					sauceUuidMap[uuid] = struct{}{}
				}
				for _, bom := range sauceProductBomList {
					delete(sauceUuidMap, bom.Uuid)
				}
				names := make([]string, 0)
				for uuid := range sauceUuidMap {
					bom, err := repository.NewProductBomRepo(db).GetSauceProductBomByUuid(uuid)
					if err != nil {
						return nil, errors.WithMessage(err)
					}
					sauceName := bom.ProductSauce.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
					names = append(names, sauceName)
				}
				tipStrPrefix := i18n.Translate(ctx.GetLanguage(), "加料")
				tipStr := i18n.Translate(ctx.GetLanguage(), "已下架，请重新选择其他加料")
				return nil, errors.New(tipStrPrefix + " " + strings.Join(names, ",") + " " + tipStr)
			}
			for i, bom := range sauceProductBomList {
				sauceProductBoms[bom.Uuid] = sauceProductBomList[i]
				sauceName := bom.ProductSauce.MultiLanguageName.GetNameByLang(ctx.GetLanguage())
				if bom.GetStockNum() < product.Num {
					return nil, errors.WithMessage(fmt.Errorf("%s %s", sauceName, i18n.Translate(ctx.GetLanguage(), "库存不足")))
				}
				// 检查加料材料库存是否充足
				if len(bom.ProductSauce.SauceMaterials) > 0 {
					for _, sauceMaterial := range bom.ProductSauce.SauceMaterials {
						materialStockNum := sauceMaterial.Material.GetStockNum()
						if materialStockNum < sauceMaterial.GetDecreaseNum(product.Num) {
							return nil, errors.WithMessage(fmt.Errorf("%s %s", sauceName, i18n.Translate(ctx.GetLanguage(), "加料材料库存不足")))
						}
					}
				}
			}
		}

		// 获取属性信息
		productAttributes := make(map[uint64]*model.ProductPackageAttribute)
		if len(product.ProductPackageAttributeUuidList) > 0 {
			productAttributeList, errProductAttributeList := repository.NewProductPackageAttributeRepo(db).GetProductPackageAttributesByUuids(product.ProductPackageAttributeUuidList)
			if errProductAttributeList != nil {
				return nil, errors.WithMessage(errProductAttributeList)
			}
			for i, attribute := range productAttributeList {
				productAttributes[attribute.Uuid] = productAttributeList[i]
			}
		}

		// 构建加料信息
		sauces := make([]model.Sauce, 0)
		for sauceProductBomUuid, sauceProductBom := range sauceProductBoms {
			sauce := model.Sauce{
				Name:           sauceProductBom.ProductSauce.MultiLanguageName.GetNameByLang(ctx.GetLanguage()), // 记录顾客下单时所用语言的名字
				Price:          sauceProductBom.Price,
				ProductBomUuid: sauceProductBomUuid,
			}
			sauces = append(sauces, sauce)
		}

		// 构建属性信息
		attributes := sortProductAttributes(ctx, productAttributes)

		isAcceptOrder := constant.OrderProductIsAcceptOrderAccepted // 已接单
		if params.IsH5Product {
			isAcceptOrder = constant.OrderProductIsAcceptOrderUnAccept // 未接单
		}
		deviceSn := ctx.GetDeviceSn()
		if ctx.GetSource() == jwt.SourceH5 {
			deviceSn = jwt.SourceH5 // 扫码h5订单，设备sn为h5
		}

		flavorPrice := flavorProductBom.Price
		isBatch := func() uint8 {
			if businessSetting.OpenIsBatch() {
				if productPackage.IsBatchBool() {
					return 1
				}
			}
			return 0
		}()

		// 前置模式下，处理分批类型UUID
		batchTagUuid := uint64(0)
		if businessSetting.BatchCookingMode == constant.BatchCookingModePre && isBatch == 1 {
			batchTagRepo := repository.NewBatchTagRepo(db)
			if product.BatchTagUuid > 0 {
				// 验证 batch_tag_uuid 的有效性
				_, err := batchTagRepo.GetBatchTagInfo(product.BatchTagUuid)
				if err != nil {
					return nil, errors.WithMessage(fmt.Errorf("分批类型不存在"), err.Error())
				}
				batchTagUuid = product.BatchTagUuid
			} else {
				// 如果未提供，使用默认分批类型（排序第一的类型）
				batchTags, err := batchTagRepo.GetBatchTagList()
				if err != nil {
					return nil, errors.WithMessage(err)
				}
				if len(batchTags) > 0 {
					// 按 sort 排序，获取排序第一的类型
					sort.Slice(batchTags, func(i, j int) bool {
						return batchTags[i].Sort < batchTags[j].Sort
					})
					batchTagUuid = batchTags[0].Uuid
				}
			}
		}

		saleOrderProduct := model.NewDefaultSaleOrderProduct(model.DefaultSaleOrderProduct{
			DeviceId:               deviceSn,
			Name:                   productPackage.Name,
			OpenMemberDiscount:     productPackage.OpenDiscount,
			TaxRate:                productPackage.TaxRate(innerParams.DiningMethod),
			DeductStockType:        productPackage.DeductStockType,
			MultiLanguageNameUuid:  productPackage.MultiLanguageNameUuid,
			ImageFileUuid:          productPackage.ImageFileUuid,
			ProductPackageUuid:     productPackage.Uuid,
			SaleBillUuid:           innerParams.SaleBillUuid,
			SaleOrderUuid:          innerParams.SaleOrderUuid,
			MemberDiscountRate:     innerParams.MemberDiscountRate,
			MemberCardDiscountRate: innerParams.MemberCardDiscountRate,
			CustomDiscountRate:     innerParams.CustomDiscountRate,
			Sauces:                 sauces,
			Num:                    product.Num,
			NumType:                productPackage.NumType,
			PackageSubProductParams: func() string {
				if product.GetIsPackageProduct() {
					return utils.ToJson(product.GetSubProductList())
				}
				return ""
			}(),
			ProductType: func() uint {
				if product.GetIsPackageProduct() {
					return constant.ProductTypePackage // 套餐商品
				}
				return constant.ProductTypeProduct // 普通商品
			}(),
			Flavor: model.Flavor{
				Name:           flavorProductBom.ProductFlavor.MultiLanguageName.ToJson(), // 填顾客下单时规格的名字
				Price:          flavorPrice,
				ProductBomUuid: product.FlavorProductBomUuid,
				ErpCode:        flavorProductBom.ErpCode,
			},
			Attribute:     attributes,
			IsAcceptOrder: uint(isAcceptOrder),
			Remark:        product.Remark,
			BatchTagUuid:  batchTagUuid,
		}, &productPackage, product.Operation)

		// 生成签名
		saleOrderProduct.UpdateSign()
		ctx.Log().Debug("生成商品签名", zap.Any("sign", saleOrderProduct.Sign))

		// 计算商品数据。折扣、税费、服务
		saleOrderProduct.CalcSaleOrderProduct(params.Setting)

		// 检查是否有相同签名的商品（用于合并相同商品）
		// ... 合并逻辑 ...

		// 如果该商品是套餐，则新建套餐子商品
		if saleOrderProduct.ProductType == constant.ProductTypePackage {
			subProducts, err := s.newPackageSubProducts(ctx, product.GetSubProducts(), innerParams, params, saleOrderProduct.Uuid, saleOrderProduct.DeductStockType)
			if err != nil {
				return nil, errors.WithMessage(err)
			}
			params.SaleOrder.SaleOrderProducts = append(params.SaleOrder.SaleOrderProducts, subProducts...)
			saleOrderProducts = append(saleOrderProducts, subProducts...)
		}
	}
	return saleOrderProducts, nil
}
```

**核心处理步骤**:

1. **商品信息验证**
   - 检查商品是否删除
   - 检查商品是否下架
   - 检查商品库存是否充足
   - 检查材料库存是否充足

2. **加料和属性处理**
   - 获取并验证加料信息
   - 获取并验证属性信息
   - 构建加料和属性数据结构

3. **创建订单商品对象**
   - 设置商品基本信息
   - 设置价格、折扣率等
   - 如果是套餐，设置 `ProductType` 为 `ProductTypePackage`
   - 保存套餐子商品参数（JSON 格式）

4. **生成商品签名**
   - 用于识别相同商品（规格、加料、属性完全一致）

5. **计算商品价格**
   - 计算折扣、税费、服务费等

6. **创建套餐子商品**（如果是套餐）
   - 调用 `newPackageSubProducts` 创建子商品

---

### 6. 创建套餐子商品

**方法**: `newPackageSubProducts`

```2062:2074:main/app/service/order.go
// 新建套餐子商品
func (s *orderSrv) newPackageSubProducts(ctx context.Context, subProducts []req.ProductParams, innerParams InnerParams,
	params CreateSaleOrderProductParams, packageUuid uint64, deductStockType uint) ([]*model.SaleOrderProduct, error) {
	subSaleOrderProducts := make([]*model.SaleOrderProduct, 0)
	for _, subProduct := range subProducts {
		subSaleOrderProduct, err := s.newSaleOrderProductForPackageSubProduct(ctx, subProduct, innerParams, params, packageUuid, deductStockType)
		if err != nil {
			return nil, errors.WithMessage(err)
		}
		subSaleOrderProducts = append(subSaleOrderProducts, subSaleOrderProduct)
	}
	return subSaleOrderProducts, nil
}
```

**方法**: `newSaleOrderProductForPackageSubProduct`

```2076:2242:main/app/service/order.go
func (s *orderSrv) newSaleOrderProductForPackageSubProduct(ctx context.Context, product req.ProductParams, innerParams InnerParams, params CreateSaleOrderProductParams, packageUuid uint64, deductStockType uint) (*model.SaleOrderProduct, error) {
	db := ctx.GetDB()
	// 获取商品包信息
	productBom, err := repository.NewProductPackageRepo(db).GetProductPackageBaseInfoByBomUuid(product.FlavorProductBomUuid)
	if err != nil {
		return nil, errors.WithMessage(err)
	}
	productPackage := productBom.ProductPackage
	productName := productPackage.MultiLanguageName.GetNameByLang(ctx.GetLanguage())

	if productBom.IsDelete() {
		return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "商品规格已经删除")))
	}
	// 商品已经下架
	if productBom.IsProductPackageDown() {
		return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "商品已经下架")))
	}

	// 获取某商品规格信息
	flavorProductBom, errFlavorProductBom := repository.NewProductBomRepo(db).GetFlavorProductBomByUuid(product.FlavorProductBomUuid)
	if errFlavorProductBom != nil {
		return nil, errors.WithMessage(errFlavorProductBom)
	}
	if flavorProductBom.GetStockNum() < float64(product.Num) {
		return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "库存不足")))
	}
	// 如果商品规格关联了材料，检查材料库存是否充足
	if len(flavorProductBom.FlavorMaterials) > 0 {
		for _, flavorMaterial := range flavorProductBom.FlavorMaterials {
			if flavorMaterial.IsDelete() {
				continue
			}
			materialStockNum := flavorMaterial.Material.GetStockNum()
			if materialStockNum < flavorMaterial.GetDecreaseNum(product.Num) {
				return nil, errors.WithMessage(fmt.Errorf("%s %s", productName, i18n.Translate(ctx.GetLanguage(), "库存不足")))
			}
		}
	}

	// 获取加料信息
	sauceProductBoms, errSauceProductBoms :=
		// ... 加料处理逻辑 ...

	saleOrderProduct := model.NewDefaultSaleOrderProduct(model.DefaultSaleOrderProduct{
		DeviceId:               deviceSn,
		Name:                   productPackage.Name,
		OpenMemberDiscount:     productPackage.OpenDiscount,
		TaxRate:                productPackage.TaxRate(innerParams.DiningMethod),
		DeductStockType:        deductStockType,
		MultiLanguageNameUuid:  productPackage.MultiLanguageNameUuid,
		ImageFileUuid:          productPackage.ImageFileUuid,
		ProductPackageUuid:     productPackage.Uuid,
		SaleBillUuid:           innerParams.SaleBillUuid,
		SaleOrderUuid:          innerParams.SaleOrderUuid,
		MemberDiscountRate:     innerParams.MemberDiscountRate,
		MemberCardDiscountRate: innerParams.MemberCardDiscountRate,
		CustomDiscountRate:     innerParams.CustomDiscountRate,
		Sauces:                 sauces,
		Num:                    product.Num,
		UnitNum:                product.UnitNum,
		IsTabletAddAndCooking:  innerParams.IsTabletAddAndCooking,
		NumType:                productPackage.NumType,
		PackageSubProductParams: func() string {
			if product.GetIsPackageProduct() {
				return utils.ToJson(product.GetSubProductList())
			}
			return ""
		}(),
		ProductType:      constant.ProductTypePackageSubProduct, // 套餐子商品
		PackageUuid:      packageUuid,
		PackageGroupUuid: product.ProductPackageGroupUuid,
		Flavor: model.Flavor{
			Name:           flavorProductBom.ProductFlavor.MultiLanguageName.ToJson(), // 填顾客下单时规格的名字
			Price:          flavorProductBom.Price,
			ProductBomUuid: product.FlavorProductBomUuid,
			ErpCode:        flavorProductBom.ErpCode,
		},
		Attribute:     attributes,
		IsAcceptOrder: uint(isAcceptOrder),
		Remark:        product.Remark,
	}, &productPackage, product.Operation)

	// 生成签名
	saleOrderProduct.UpdateSign()
	ctx.Log().Debug("生成商品签名", zap.Any("sign", saleOrderProduct.Sign))

	// 计算商品数据。折扣、税费、服务
	saleOrderProduct.CalcSaleOrderProduct(params.Setting)

	return saleOrderProduct, nil
}
```

**核心处理步骤**:

1. **遍历子商品列表**，为每个子商品创建订单商品对象
2. **验证子商品信息**（删除、下架、库存等）
3. **创建子商品对象**
   - 设置 `ProductType` 为 `ProductTypePackageSubProduct`
   - 设置 `PackageUuid` 关联到主套餐商品
   - 设置 `PackageGroupUuid` 关联到套餐分组
   - 设置 `UnitNum`（单位数量）
4. **生成签名并计算价格**

---

## 关键数据结构

### ProductParams（商品参数）

```116:140:main/app/dto/req/shop_cart.go
// ProductParams 商品参数
type ProductParams struct {
	FlavorProductBomUuid            uint64           `json:"flavor_product_bom_uuid"`             // 商品规格uuid
	Num                             float64          `json:"num"  binding:"required"`             // 数量数量
	UnitNum                         float64          `json:"unit_num"`                            // 单位数量. 目前只有平板的加购并送厨使用该字段
	Price                           *float64         `json:"price"`                               // 商品价格，商品单价。当商品价格与后台设置的最新价格不一致时，加购失败并返回最新价格。可选，不传时，不进行价格校验
	IsBuffet                        *bool            `json:"is_buffet"`                           // 是否是自助餐商品。可选，不填时，表示不判断是不是最新价格。该参数仅在判断价格时使用
	SauceProductBomUuidList         []uint64         `json:"sauce_product_bom_uuid_list"`         // 加料信息
	ProductPackageAttributeUuidList []uint64         `json:"product_package_attribute_uuid_list"` // 属性信息
	Operation                       string           `json:"operation"`                           // 操作类型。add: 加购，sub: 减购
	MustPlanUuid                    uint64           `json:"must_plan_uuid"`                      // 必点方案uuid. 可选，在必点方案弹窗中加购时填写
	Remark                          string           `json:"remark"`                              // 备注，平板端离线购物车提交
	ProductPackageGroupUuid         uint64           `json:"product_package_group_uuid"`          // 套餐分组uuid。可选，当商品是套餐商品时，该字段有值
	ProductType                     uint             `json:"product_type"`                        // 商品类型 0-商品 1-套餐
	Products                        []ProductRequest `json:"products"`                            // 套餐商品请求列表。当商品是套餐商品时，该字段有值
	ProductPackageUuid              uint64           `json:"product_package_uuid"`                // 套餐商品uuid。当商品是套餐商品时，该字段有值

	isPackageProduct        bool   // 是否是套餐商品
	packageSubProductParams string // 套餐子商品参数（JSON格式）

	subProducts         []ProductParams // 套餐子商品列表。当商品是套餐商品时，该字段有值
	isPackageSubProduct bool            // 是否是套餐子商品
	packageUuid         uint64          // 套餐uuid,用于标注套餐子商品的套餐商品（sale_order_product）的uuid
	BatchTagUuid        uint64          `json:"batch_tag_uuid"` // 分批类型UUID, 可选（前置模式时使用）
}
```

---

## 业务规则

### 1. 销售账单处理

- **如果 `SaleBillUuid` 为 0**：
  - 先检查是否存在待支付、未挂单的订单
  - 如果存在，使用现有订单
  - 如果不存在，创建新的即时订单

### 2. 并发控制

- 使用分布式锁（`lock.LockUuid`）防止同一销售账单的并发操作
- 在多个层级都有锁检查（双重检查）

### 3. 商品验证

- **商品状态检查**：删除、下架
- **库存检查**：商品库存、材料库存
- **加料验证**：加料是否存在、是否下架、库存是否充足
- **属性验证**：属性是否存在

### 4. 套餐商品处理

- **主商品**：
  - `ProductType` = `ProductTypePackage`
  - 保存子商品参数到 `PackageSubProductParams`（JSON 格式）
  
- **子商品**：
  - `ProductType` = `ProductTypePackageSubProduct`
  - `PackageUuid` 关联到主套餐商品
  - `PackageGroupUuid` 关联到套餐分组
  - `UnitNum` 记录单位数量

### 5. 商品合并

- 使用商品签名（`Sign`）识别相同商品
- 相同签名的商品会合并，数量相加
- 套餐商品合并时，子商品数量也会相应更新

### 6. 价格计算

- 在主商品和子商品创建后，都会调用 `CalcSaleOrderProduct` 计算价格
- 计算包括：折扣、税费、服务费等

### 7. 限购检查

- 检查商品限购规则（如果未跳过）
- H5 端和普通端有不同的限购检查逻辑

### 8. 超时检查

- 检查商品是否超时不能加购

---

## 错误处理

### 常见错误场景

1. **商品已删除/下架**
   - 错误信息：`"{商品名} 商品规格已经删除"` 或 `"{商品名} 商品已经下架"`

2. **库存不足**
   - 错误信息：`"{商品名} 库存不足"`

3. **加料已下架**
   - 错误信息：`"加料 {加料名} 已下架，请重新选择其他加料"`

4. **订单状态不允许加购**
   - 错误信息：由 `ValidateOrderStatus` 返回

5. **商品数量超过限制**
   - 错误信息：`"商品数量不能超过999个"` 或 `"销售订单商品数量超过1000项"`

---

## 数据库事务

整个加购过程在数据库事务中执行：

```go
if err := repository.CommonRepo.Transaction(db, func(db *gorm.DB) error {
    if err := s.CalcAndSaveSaleBill(ctx, db, saleBill); err != nil {
        return errors.WithMessage(err)
    }
    return nil
}); err != nil {
    return errors.WithMessage(err)
}
```

**事务中包含的操作**：
- 保存订单商品（主商品和子商品）
- 重新计算订单金额
- 更新订单状态

---

## 返回数据

最终返回 `resp.ShopCart` 结构，包含：
- 购物车商品列表
- 订单金额信息
- 桌台信息
- 其他购物车相关数据

---

## 相关接口

- **普通商品加购**: `/cashier/desk/order/cart/product/add`
- **套餐加购**: `/cashier/desk/order/cart/product_package/add`
- **查询购物车**: `/cashier/desk/order/cart/info`

---

## 注意事项

1. **套餐商品数量固定为 1**：在 `OrderCartProductPackageAdd` 中，套餐主商品的数量固定为 1，子商品的数量由前端传入

2. **子商品数量计算**：子商品的 `Num` 字段存储的是实际数量，`UnitNum` 存储的是单位数量（一个套餐包含多少个该子商品）

3. **商品签名**：用于识别相同商品，包括规格、加料、属性等信息

4. **库存扣减时机**：库存检查在加购时进行，实际扣减在送厨或结账时进行（根据 `DeductStockType` 配置）

5. **H5 商品标记**：H5 端加购的商品会标记 `IsH5Product = true`，影响接单状态和限购检查

---

## 更新日志

- **2025-01-XX**: 初始版本，基于 `cashier_desk.go` 分析套餐加购业务逻辑



